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
// PERFORMANCE: Uses cached tenant_active from JWT when available to avoid DB lookup
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID not found in context"})
			c.Abort()
			return
		}

		// PERFORMANCE: Check if tenant_active is cached in JWT (new tokens have this)
		tenantActiveInterface, hasCachedStatus := c.Get("tenant_active")
		if hasCachedStatus {
			tenantActive, ok := tenantActiveInterface.(bool)
			if ok && tenantActive {
				// Tenant active status cached in JWT - skip DB lookup
				schemaName := fmt.Sprintf("tenant_%d", tenantID)
				db := database.GetDB()
				sessionDB := db.Session(&gorm.Session{})
				tenantDB := database.SetSchema(sessionDB, schemaName)
				c.Set("db", tenantDB)
				c.Set("schema", schemaName)
				c.Next()
				return
			}
			// If cached as false or not a bool, fall through to DB check
		}

		// FALLBACK: Validate tenant in database (for old tokens without cached status)
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
