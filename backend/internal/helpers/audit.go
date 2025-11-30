package helpers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
)

// AuditAction logs an action to the audit trail
func AuditAction(c *gin.Context, action, resource string, resourceID uint, success bool, details map[string]interface{}) {
	// Get user info from context
	userID, _ := c.Get("user_id")
	userEmail, _ := c.Get("user_email")
	userRole, _ := c.Get("user_role")

	// Convert details to JSON
	detailsJSON := ""
	if details != nil {
		if jsonBytes, err := json.Marshal(details); err == nil {
			detailsJSON = string(jsonBytes)
		}
	}

	// Create audit log entry
	auditLog := models.AuditLog{
		UserID:     toUint(userID),
		UserEmail:  toString(userEmail),
		UserRole:   toString(userRole),
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Method:     c.Request.Method,
		Path:       c.Request.URL.Path,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Details:    detailsJSON,
		Success:    success,
	}

	// Save to database asynchronously to not slow down requests
	go func() {
		db := database.GetDB()
		if err := db.Create(&auditLog).Error; err != nil {
			log.Printf("Failed to create audit log: %v", err)
		}
	}()
}

// AuditLogin logs a login attempt
func AuditLogin(c *gin.Context, email string, success bool, details map[string]interface{}) {
	auditLog := models.AuditLog{
		UserEmail:  email,
		Action:     "login",
		Resource:   "auth",
		Method:     c.Request.Method,
		Path:       c.Request.URL.Path,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Success:    success,
	}

	if details != nil {
		if jsonBytes, err := json.Marshal(details); err == nil {
			auditLog.Details = string(jsonBytes)
		}
	}

	go func() {
		db := database.GetDB()
		if err := db.Create(&auditLog).Error; err != nil {
			log.Printf("Failed to create audit log: %v", err)
		}
	}()
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
