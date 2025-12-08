package middleware

import (
	"context"
	"drcrwell/backend/internal/cache"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RedisRateLimiter implements distributed rate limiting using Redis
// Falls back to in-memory rate limiting if Redis is unavailable
type RedisRateLimiter struct {
	prefix   string
	limit    int
	window   time.Duration
	fallback *RateLimiter // Fallback to memory-based limiter
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
// prefix: key prefix for Redis (e.g., "login", "api", "whatsapp")
// limit: max requests per window
// window: time window duration
func NewRedisRateLimiter(prefix string, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		prefix:   prefix,
		limit:    limit,
		window:   window,
		fallback: NewRateLimiter(limit, window),
	}
}

// redisKey generates the Redis key for rate limiting
func (rl *RedisRateLimiter) redisKey(identifier string) string {
	return fmt.Sprintf("ratelimit:%s:%s", rl.prefix, identifier)
}

// blockKey generates the Redis key for blocked IPs
func (rl *RedisRateLimiter) blockKey(identifier string) string {
	return fmt.Sprintf("ratelimit:blocked:%s:%s", rl.prefix, identifier)
}

// isBlocked checks if an identifier is currently blocked
func (rl *RedisRateLimiter) isBlocked(ctx context.Context, identifier string) (bool, time.Duration) {
	client := cache.GetClient()
	if client == nil {
		return false, 0
	}

	ttl, err := client.TTL(ctx, rl.blockKey(identifier)).Result()
	if err != nil || ttl <= 0 {
		return false, 0
	}

	return true, ttl
}

// block blocks an identifier for 15 minutes
func (rl *RedisRateLimiter) block(ctx context.Context, identifier string) {
	client := cache.GetClient()
	if client == nil {
		return
	}

	client.Set(ctx, rl.blockKey(identifier), "1", 15*time.Minute)
}

// incrementAndCheck increments the counter and returns if request is allowed
func (rl *RedisRateLimiter) incrementAndCheck(ctx context.Context, identifier string) (bool, int64) {
	client := cache.GetClient()
	if client == nil {
		return true, 0 // Allow if Redis unavailable, fallback will handle
	}

	key := rl.redisKey(identifier)

	// Use Redis INCR with EXPIRE for atomic counter
	pipe := client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.window)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return true, 0 // Allow on error, don't block legitimate users
	}

	count := incr.Val()
	return count <= int64(rl.limit), count
}

// RateLimitMiddleware returns a Gin middleware for rate limiting
func (rl *RedisRateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		identifier := c.ClientIP()

		// Check if Redis is available
		client := cache.GetClient()
		if client == nil {
			// Fallback to memory-based rate limiting
			rl.fallback.RateLimitMiddleware()(c)
			return
		}

		// Check if blocked
		if blocked, remaining := rl.isBlocked(ctx, identifier); blocked {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Muitas tentativas. Tente novamente mais tarde.",
				"retry_after": int(remaining.Minutes()),
			})
			c.Abort()
			return
		}

		// Check rate limit
		allowed, count := rl.incrementAndCheck(ctx, identifier)
		if !allowed {
			// Block for 15 minutes after exceeding limit
			rl.block(ctx, identifier)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Muitas tentativas. Tente novamente em 15 minutos.",
				"retry_after": 15,
			})
			c.Abort()
			return
		}

		// Add rate limit headers for debugging
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(rl.limit)-count))

		c.Next()
	}
}

// Pre-configured Redis rate limiters with fallback

// RedisLoginRateLimiter for login endpoints
// 5 attempts per minute, blocks for 15 minutes if exceeded
// Distributed across all backend replicas
var RedisLoginRateLimiter = NewRedisRateLimiter("login", 5, time.Minute)

// RedisWhatsAppRateLimiter for WhatsApp API endpoints
// 200 requests per minute per IP (higher limit for bot integrations)
// Distributed across all backend replicas
var RedisWhatsAppRateLimiter = NewRedisRateLimiter("whatsapp", 200, time.Minute)

// RedisGlobalAPIRateLimiter for general API endpoints
// 100 requests per minute per IP
// Distributed across all backend replicas
var RedisGlobalAPIRateLimiter = NewRedisRateLimiter("api", 100, time.Minute)
