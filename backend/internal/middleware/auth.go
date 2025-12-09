package middleware

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Claims represents JWT claims
type Claims struct {
	UserID      uint                       `json:"user_id"`
	TenantID    uint                       `json:"tenant_id"`
	Email       string                     `json:"email"`
	Role        string                     `json:"role"`
	Permissions map[string]map[string]bool `json:"permissions,omitempty"` // Optional for backward compatibility
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		// Set permissions in context (may be nil for old tokens - backward compatibility)
		if claims.Permissions != nil {
			c.Set("user_permissions", claims.Permissions)
		}

		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// Safe type assertion
		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid role type"})
			c.Abort()
			return
		}

		allowed := false
		for _, role := range allowedRoles {
			if roleStr == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuperAdminMiddleware checks if user is a super admin
func SuperAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check if user is super admin
		db := database.GetDB()
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if !user.IsSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
			c.Abort()
			return
		}

		c.Set("is_super_admin", true)
		c.Next()
	}
}

// APIKeyMiddleware validates API key for external integrations (WhatsApp, AI bots)
// The API key should be passed in the X-API-Key header
func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   true,
				"message": "API key is required. Use X-API-Key header.",
			})
			c.Abort()
			return
		}

		// Find tenant by API key
		db := database.GetDB()
		var tenant models.Tenant
		result := db.Where("api_key = ? AND api_key_active = ? AND active = ?", apiKey, true, true).First(&tenant)

		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   true,
				"message": "Invalid or inactive API key",
			})
			c.Abort()
			return
		}

		// Set tenant info in context
		c.Set("tenant_id", tenant.ID)
		c.Set("tenant_name", tenant.Name)
		c.Set("api_access", true) // Flag to identify API access vs user access

		// Set tenant-specific schema using a new session to ensure search_path persists
		schemaName := fmt.Sprintf("tenant_%d", tenant.ID)
		sessionDB := db.Session(&gorm.Session{})
		tenantDB := database.SetSchema(sessionDB, schemaName)

		// Store tenant DB in context
		c.Set("db", tenantDB)
		c.Set("schema", schemaName)

		c.Next()
	}
}
