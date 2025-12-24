package middleware

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TenantMiddleware sets the database schema based on tenant
// SECURITY: Validates that tenant exists and is active before granting access
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID not found in context"})
			c.Abort()
			return
		}

		// SECURITY: Validate tenant exists and is active before setting schema
		db := database.GetDB()
		var tenant models.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		// Check if tenant is active
		if !tenant.Active {
			c.JSON(http.StatusForbidden, gin.H{"error": "Tenant account is inactive"})
			c.Abort()
			return
		}

		// Set tenant-specific schema using a new session to ensure search_path persists
		schemaName := fmt.Sprintf("tenant_%d", tenantID)

		// Create a new DB session with tenant schema
		sessionDB := db.Session(&gorm.Session{})
		tenantDB := database.SetSchema(sessionDB, schemaName)

		// Store tenant DB in context
		c.Set("db", tenantDB)
		c.Set("schema", schemaName)

		c.Next()
	}
}

// GetDBFromContext retrieves the tenant-specific database from context
// DEPRECATED: Use GetDBFromContextSafe from middleware/context.go instead
// This function now fails closed (returns nil) instead of falling back to global DB
// to prevent accidental cross-tenant data access
func GetDBFromContext(c *gin.Context) interface{} {
	db, exists := c.Get("db")
	if !exists {
		// SECURITY: Fail closed - do not fallback to global DB
		// This prevents accidental access to wrong tenant data
		return nil
	}
	return db
}
