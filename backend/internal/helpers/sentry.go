package helpers

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// SentryConfig holds configuration for Sentry initialization
type SentryConfig struct {
	DSN              string
	Environment      string
	Release          string
	TracesSampleRate float64
}

// InitSentry initializes the Sentry SDK for error tracking
func InitSentry() error {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		log.Println("SENTRY_DSN not set, error tracking disabled")
		return nil
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	release := os.Getenv("SENTRY_RELEASE")
	if release == "" {
		release = "drcrwell-backend@1.0.0"
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      env,
		Release:          release,
		TracesSampleRate: 0.1, // 10% of transactions for performance monitoring
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Filter out sensitive data
			for k := range event.Extra {
				if k == "password" || k == "token" || k == "secret" {
					delete(event.Extra, k)
				}
			}
			return event
		},
	})
	if err != nil {
		return fmt.Errorf("sentry initialization failed: %w", err)
	}

	log.Println("Sentry initialized successfully")
	return nil
}

// CloseSentry flushes Sentry events before shutdown
func CloseSentry() {
	sentry.Flush(2 * time.Second)
}

// CaptureError captures an error and sends it to Sentry with context
func CaptureError(err error, c *gin.Context) {
	if sentry.CurrentHub().Client() == nil {
		return // Sentry not initialized
	}

	sentry.WithScope(func(scope *sentry.Scope) {
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
		scope.SetTag("method", c.Request.Method)
		scope.SetTag("path", c.Request.URL.Path)

		sentry.CaptureException(err)
	})
}

// CaptureMessage captures a message and sends it to Sentry
func CaptureMessage(message string, level sentry.Level) {
	if sentry.CurrentHub().Client() == nil {
		return // Sentry not initialized
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		sentry.CaptureMessage(message)
	})
}

// CaptureErrorWithExtra captures an error with additional context data
func CaptureErrorWithExtra(err error, extra map[string]interface{}, c *gin.Context) {
	if sentry.CurrentHub().Client() == nil {
		return // Sentry not initialized
	}

	sentry.WithScope(func(scope *sentry.Scope) {
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

		// Add extra context
		for k, v := range extra {
			scope.SetExtra(k, v)
		}

		// Add request information
		scope.SetRequest(c.Request)

		sentry.CaptureException(err)
	})
}
