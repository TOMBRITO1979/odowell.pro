package middleware

import (
	"bytes"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware logs all API requests to the audit trail
// This provides comprehensive logging of user actions for LGPD compliance
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip health check and static files
		path := c.Request.URL.Path
		if path == "/health" || path == "/ping" || strings.HasPrefix(path, "/static") {
			c.Next()
			return
		}

		// Skip if not authenticated (login attempts are logged separately)
		userID, exists := c.Get("user_id")
		if !exists || userID == nil {
			c.Next()
			return
		}

		// Capture start time
		startTime := time.Now()

		// Read request body for logging (if applicable)
		var requestBody string
		if c.Request.Method != "GET" && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			// Restore body for handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			// Sanitize sensitive fields
			requestBody = sanitizeBody(string(bodyBytes))
		}

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Determine action from method
		action := methodToAction(c.Request.Method)

		// Determine resource from path
		resource := extractResource(path)

		// Get user info
		userEmail, _ := c.Get("user_email")
		userRole, _ := c.Get("user_role")

		// Build details
		details := map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
			"status_code": c.Writer.Status(),
			"query":       c.Request.URL.RawQuery,
		}

		// Add request body for write operations (sanitized)
		if requestBody != "" && len(requestBody) < 5000 {
			details["request_summary"] = truncateString(requestBody, 1000)
		}

		// Add any errors
		if len(c.Errors) > 0 {
			details["errors"] = c.Errors.String()
		}

		detailsJSON, _ := json.Marshal(details)

		// Determine success based on status code
		success := c.Writer.Status() >= 200 && c.Writer.Status() < 400

		// Create audit log
		auditLog := models.AuditLog{
			UserID:     toUint(userID),
			UserEmail:  toString(userEmail),
			UserRole:   toString(userRole),
			Action:     action,
			Resource:   resource,
			Method:     c.Request.Method,
			Path:       path,
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Details:    string(detailsJSON),
			Success:    success,
		}

		// Save asynchronously
		go func() {
			db := database.GetDB()
			if db == nil {
				return
			}
			if err := db.Create(&auditLog).Error; err != nil {
				log.Printf("Audit middleware: failed to create log: %v", err)
			}
		}()
	}
}

// methodToAction converts HTTP method to action name
func methodToAction(method string) string {
	switch method {
	case "GET":
		return "view"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "access"
	}
}

// extractResource extracts the resource name from the path
func extractResource(path string) string {
	// Remove /api/v1/ prefix
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/")

	// Split by / and get first segment
	parts := strings.Split(path, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "unknown"
}

// sanitizeBody removes sensitive fields from request body
func sanitizeBody(body string) string {
	// List of sensitive field names to redact
	sensitiveFields := []string{
		"password", "senha", "token", "secret", "api_key",
		"credit_card", "cartao", "cvv", "cpf", "rg",
	}

	result := body
	for _, field := range sensitiveFields {
		// Simple pattern matching - could be improved with regex
		if strings.Contains(strings.ToLower(result), field) {
			// Just note that sensitive data was present
			result = strings.ReplaceAll(result, field, field+":[REDACTED]")
		}
	}
	return result
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}

// Helper functions
func toUint(v interface{}) uint {
	if v == nil {
		return 0
	}
	if u, ok := v.(uint); ok {
		return u
	}
	return 0
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
