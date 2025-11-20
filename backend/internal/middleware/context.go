package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetDBFromContextSafe safely retrieves the database connection from context
// Returns the DB and true if successful, or responds with error and returns nil, false
func GetDBFromContextSafe(c *gin.Context) (*gorm.DB, bool) {
	dbVal, exists := c.Get("db")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
		c.Abort()
		return nil, false
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection type"})
		c.Abort()
		return nil, false
	}

	return db, true
}

// GetUserIDSafe safely retrieves the user ID from context
func GetUserIDSafe(c *gin.Context) (uint, bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		c.Abort()
		return 0, false
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		c.Abort()
		return 0, false
	}

	return userID, true
}

// GetTenantIDSafe safely retrieves the tenant ID from context
func GetTenantIDSafe(c *gin.Context) (uint, bool) {
	tenantIDVal, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant ID not found"})
		c.Abort()
		return 0, false
	}

	tenantID, ok := tenantIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid tenant ID type"})
		c.Abort()
		return 0, false
	}

	return tenantID, true
}
