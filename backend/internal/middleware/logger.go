package middleware

import (
	"drcrwell/backend/internal/helpers"
	"time"

	"github.com/gin-gonic/gin"
)

// JSONLoggerMiddleware logs all requests in structured JSON format
func JSONLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get request ID (set by RequestIDMiddleware)
		requestID, _ := c.Get("request_id")
		requestIDStr, _ := requestID.(string)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get user info from context (set by AuthMiddleware)
		userID, _ := c.Get("user_id")
		userIDUint, _ := userID.(uint)
		tenantID, _ := c.Get("tenant_id")
		tenantIDUint, _ := tenantID.(uint)

		// Skip logging for health checks to reduce noise
		if c.Request.URL.Path == "/health" {
			return
		}

		// Determine log level based on status code
		level := helpers.LogLevelInfo
		if c.Writer.Status() >= 400 && c.Writer.Status() < 500 {
			level = helpers.LogLevelWarn
		} else if c.Writer.Status() >= 500 {
			level = helpers.LogLevelError
		}

		// Get error if any
		var errorMsg string
		if len(c.Errors) > 0 {
			errorMsg = c.Errors.String()
		}

		// Log the request
		helpers.LogRequest(helpers.LogEntry{
			Level:      level,
			Message:    "HTTP Request",
			RequestID:  requestIDStr,
			UserID:     userIDUint,
			TenantID:   tenantIDUint,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			DurationMs: float64(duration.Microseconds()) / 1000.0,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Error:      errorMsg,
		})
	}
}
