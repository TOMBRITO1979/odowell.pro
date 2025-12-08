package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter stores request counts per IP
type RateLimiter struct {
	requests map[string]*requestInfo
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type requestInfo struct {
	count     int
	firstReq  time.Time
	blocked   bool
	blockTime time.Time
}

// NewRateLimiter creates a rate limiter
// limit: max requests per window
// window: time window duration
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*requestInfo),
		limit:    limit,
		window:   window,
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
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, info := range rl.requests {
		// Remove entries older than 2x window
		if now.Sub(info.firstReq) > rl.window*2 {
			delete(rl.requests, ip)
		}
	}
}

// RateLimitMiddleware returns a rate limiting middleware
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		info, exists := rl.requests[ip]
		now := time.Now()

		if !exists {
			rl.requests[ip] = &requestInfo{
				count:    1,
				firstReq: now,
			}
			rl.mu.Unlock()
			c.Next()
			return
		}

		// Check if blocked
		if info.blocked {
			// Block for 15 minutes after too many attempts
			remaining := 15*time.Minute - now.Sub(info.blockTime)
			if remaining > 0 {
				rl.mu.Unlock()
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Too many attempts. Please try again later.",
					"retry_after": int(remaining.Minutes()),
				})
				c.Abort()
				return
			}
			// Unblock after 15 minutes
			info.blocked = false
			info.count = 0
			info.firstReq = now
		}

		// Reset count if window expired
		if now.Sub(info.firstReq) > rl.window {
			info.count = 1
			info.firstReq = now
			rl.mu.Unlock()
			c.Next()
			return
		}

		// Increment count
		info.count++

		// Check limit
		if info.count > rl.limit {
			info.blocked = true
			info.blockTime = now
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many login attempts. Please try again in 15 minutes.",
				"retry_after": 15,
			})
			c.Abort()
			return
		}

		rl.mu.Unlock()
		c.Next()
	}
}

// LoginRateLimiter is a pre-configured rate limiter for login endpoints
// Allows 5 attempts per minute, blocks for 15 minutes if exceeded
var LoginRateLimiter = NewRateLimiter(5, time.Minute)

// WhatsAppRateLimiter is a pre-configured rate limiter for WhatsApp API endpoints
// Allows 200 requests per minute per API key (higher limit for bot integrations)
// This protects against abuse while allowing normal WhatsApp bot traffic
var WhatsAppRateLimiter = NewRateLimiter(200, time.Minute)

// GlobalAPIRateLimiter is a pre-configured rate limiter for general API endpoints
// Allows 100 requests per minute per IP
var GlobalAPIRateLimiter = NewRateLimiter(100, time.Minute)
