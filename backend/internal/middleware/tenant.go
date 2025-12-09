package middleware

import (
	"drcrwell/backend/internal/database"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TenantMiddleware sets the database schema based on tenant
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID not found in context"})
			c.Abort()
			return
		}

		// Set tenant-specific schema using a new session to ensure search_path persists
		schemaName := fmt.Sprintf("tenant_%d", tenantID)

		// Create a new DB session with tenant schema
		db := database.GetDB()
		sessionDB := db.Session(&gorm.Session{})
		tenantDB := database.SetSchema(sessionDB, schemaName)

		// Store tenant DB in context
		c.Set("db", tenantDB)
		c.Set("schema", schemaName)

		c.Next()
	}
}

// GetDBFromContext retrieves the tenant-specific database from context
func GetDBFromContext(c *gin.Context) interface{} {
	db, exists := c.Get("db")
	if !exists {
		return database.GetDB()
	}
	return db
}
