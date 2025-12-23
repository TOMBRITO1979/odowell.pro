package middleware

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter stores request counts per IP using sync.Map for better concurrency
type RateLimiter struct {
	requests sync.Map // map[string]*requestInfo - optimized for concurrent access
	limit    int
	window   time.Duration
}

type requestInfo struct {
	count     int64 // Using int64 for atomic operations
	firstReq  int64 // Unix nano timestamp for atomic operations
	blocked   int32 // 1 = blocked, 0 = not blocked (for atomic)
	blockTime int64 // Unix nano timestamp
	mu        sync.Mutex // Per-entry mutex for complex operations
}

// NewRateLimiter creates a rate limiter
// limit: max requests per window
// window: time window duration
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limit:  limit,
		window: window,
	}

	// Cleanup old entries every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	now := time.Now().UnixNano()
	windowNano := rl.window.Nanoseconds() * 2

	// Use Range to iterate safely over sync.Map
	rl.requests.Range(func(key, value interface{}) bool {
		info := value.(*requestInfo)
		firstReq := atomic.LoadInt64(&info.firstReq)

		// Remove entries older than 2x window
		if now-firstReq > windowNano {
			rl.requests.Delete(key)
		}
		return true
	})
}

// getOrCreateInfo gets existing info or creates new one atomically
func (rl *RateLimiter) getOrCreateInfo(ip string, now time.Time) (*requestInfo, bool) {
	nowNano := now.UnixNano()

	// Try to load existing
	if existing, ok := rl.requests.Load(ip); ok {
		return existing.(*requestInfo), false
	}

	// Create new entry
	newInfo := &requestInfo{
		count:    1,
		firstReq: nowNano,
	}

	// LoadOrStore ensures atomic insertion
	actual, loaded := rl.requests.LoadOrStore(ip, newInfo)
	if loaded {
		// Another goroutine inserted first, use that entry
		return actual.(*requestInfo), false
	}

	// We successfully inserted the new entry
	return newInfo, true
}

// RateLimitMiddleware returns a rate limiting middleware
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		info, isNew := rl.getOrCreateInfo(ip, now)

		// If we just created a new entry with count=1, allow request
		if isNew {
			c.Next()
			return
		}

		// Lock this specific entry for complex operations
		info.mu.Lock()
		defer info.mu.Unlock()

		nowNano := now.UnixNano()

		// Check if blocked
		if atomic.LoadInt32(&info.blocked) == 1 {
			blockTime := atomic.LoadInt64(&info.blockTime)
			// Block for 15 minutes after too many attempts
			remaining := 15*time.Minute - time.Duration(nowNano-blockTime)
			if remaining > 0 {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Too many attempts. Please try again later.",
					"retry_after": int(remaining.Minutes()),
				})
				c.Abort()
				return
			}
			// Unblock after 15 minutes
			atomic.StoreInt32(&info.blocked, 0)
			atomic.StoreInt64(&info.count, 0)
			atomic.StoreInt64(&info.firstReq, nowNano)
		}

		// Reset count if window expired
		firstReq := atomic.LoadInt64(&info.firstReq)
		if nowNano-firstReq > rl.window.Nanoseconds() {
			atomic.StoreInt64(&info.count, 1)
			atomic.StoreInt64(&info.firstReq, nowNano)
			c.Next()
			return
		}

		// Increment count
		count := atomic.AddInt64(&info.count, 1)

		// Check limit
		if count > int64(rl.limit) {
			atomic.StoreInt32(&info.blocked, 1)
			atomic.StoreInt64(&info.blockTime, nowNano)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many login attempts. Please try again in 15 minutes.",
				"retry_after": 15,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LogRateLimiterFallback logs when Redis rate limiter falls back to in-memory
func LogRateLimiterFallback(reason string) {
	log.Printf("[RateLimiter] Fallback to in-memory: %s", reason)
}

// LoginRateLimiter is a pre-configured rate limiter for login endpoints
// Allows 5 attempts per minute, blocks for 15 minutes if exceeded
var LoginRateLimiter = NewRateLimiter(5, time.Minute)

// ForgotPasswordRateLimiter is a strict rate limiter for password reset requests
// Allows 3 attempts per 15 minutes to prevent abuse
var ForgotPasswordRateLimiter = NewRateLimiter(3, 15*time.Minute)

// TwoFARateLimiter is a rate limiter for 2FA verification attempts
// Allows 5 attempts per minute to prevent brute force
var TwoFARateLimiter = NewRateLimiter(5, time.Minute)

// TenantRegistrationRateLimiter is a very strict rate limiter for tenant registration
// Allows 3 registrations per hour per IP to prevent spam
var TenantRegistrationRateLimiter = NewRateLimiter(3, time.Hour)

// WhatsAppRateLimiter is a pre-configured rate limiter for WhatsApp API endpoints
// Allows 200 requests per minute per API key (higher limit for bot integrations)
// This protects against abuse while allowing normal WhatsApp bot traffic
var WhatsAppRateLimiter = NewRateLimiter(200, time.Minute)

// GlobalAPIRateLimiter is a pre-configured rate limiter for general API endpoints
// Allows 100 requests per minute per IP
var GlobalAPIRateLimiter = NewRateLimiter(100, time.Minute)
