package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"drcrwell/backend/internal/helpers"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// SentryMiddleware captures panics and sends them to Sentry
func SentryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		hub := sentry.CurrentHub().Clone()
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				stack := string(debug.Stack())
				helpers.LogError(fmt.Sprintf("Panic recovered: %v", err), nil, map[string]interface{}{
					"stack": stack,
				})

				// Capture to Sentry if configured
				if hub.Client() != nil {
					hub.WithScope(func(scope *sentry.Scope) {
						// Add user context
						if userID, exists := c.Get("user_id"); exists {
							scope.SetUser(sentry.User{
								ID: fmt.Sprintf("%d", userID),
							})
						}

						// Add tenant context
						if tenantID, exists := c.Get("tenant_id"); exists {
							scope.SetTag("tenant_id", fmt.Sprintf("%d", tenantID))
						}

						// Add request context
						if requestID, exists := c.Get("request_id"); exists {
							scope.SetTag("request_id", fmt.Sprintf("%v", requestID))
						}

						// Add request information
						scope.SetRequest(c.Request)
						scope.SetExtra("stack_trace", stack)

						hub.RecoverWithContext(c.Request.Context(), err)
					})

					// Flush events
					hub.Flush(2 * time.Second)
				}

				// Return 500 error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal server error",
					"request_id": c.GetString("request_id"),
				})
			}
		}()

		c.Next()
	}
}

// CaptureGinErrors captures Gin errors that were added via c.Error()
func CaptureGinErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// After request processing, check for errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				helpers.CaptureError(err.Err, c)
			}
		}
	}
}
