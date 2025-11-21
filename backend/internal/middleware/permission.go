package middleware

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PermissionMiddleware checks if user has permission for a specific action on a module
func PermissionMiddleware(moduleCode, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Safe type assertion
		userID, ok := userIDInterface.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
			c.Abort()
			return
		}

		// Get user role from context
		userRoleInterface, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		// Safe type assertion
		userRole, ok := userRoleInterface.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user role type"})
			c.Abort()
			return
		}

		// Admins bypass permission checks (superuser)
		if userRole == "admin" {
			c.Next()
			return
		}

		// Check if user has permission
		hasPermission, err := CheckUserPermission(userID, moduleCode, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check permissions", "details": err.Error()})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "Insufficient permissions",
				"module": moduleCode,
				"action": action,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckUserPermission checks if a user has a specific permission
func CheckUserPermission(userID uint, moduleCode, action string) (bool, error) {
	db := database.GetDB()

	// Query to check if user has permission
	var count int64
	err := db.Table("public.user_permissions up").
		Select("COUNT(*)").
		Joins("INNER JOIN public.permissions p ON p.id = up.permission_id").
		Joins("INNER JOIN public.modules m ON m.id = p.module_id").
		Where("up.user_id = ?", userID).
		Where("m.code = ?", moduleCode).
		Where("p.action = ?", action).
		Where("up.deleted_at IS NULL").
		Where("p.deleted_at IS NULL").
		Where("m.deleted_at IS NULL AND m.active = true").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserPermissions returns all permissions for a user as a map
func GetUserPermissions(userID uint) (map[string]map[string]bool, error) {
	db := database.GetDB()

	type PermissionRow struct {
		ModuleCode string
		Action     string
	}

	var rows []PermissionRow
	err := db.Table("public.user_permissions up").
		Select("m.code as module_code, p.action").
		Joins("INNER JOIN public.permissions p ON p.id = up.permission_id").
		Joins("INNER JOIN public.modules m ON m.id = p.module_id").
		Where("up.user_id = ?", userID).
		Where("up.deleted_at IS NULL").
		Where("p.deleted_at IS NULL").
		Where("m.deleted_at IS NULL AND m.active = true").
		Find(&rows).Error

	if err != nil {
		return nil, err
	}

	// Build permissions map
	permissions := make(map[string]map[string]bool)
	for _, row := range rows {
		if permissions[row.ModuleCode] == nil {
			permissions[row.ModuleCode] = make(map[string]bool)
		}
		permissions[row.ModuleCode][row.Action] = true
	}

	return permissions, nil
}

// GetAllUserPermissionsWithDefaults returns permissions with default false values for all modules
func GetAllUserPermissionsWithDefaults(userID uint) (map[string]map[string]bool, error) {
	db := database.GetDB()

	// Get all modules
	var modules []models.Module
	if err := db.Where("active = ?", true).Where("deleted_at IS NULL").Find(&modules).Error; err != nil {
		return nil, err
	}

	// Initialize with all modules and false permissions
	permissions := make(map[string]map[string]bool)
	for _, module := range modules {
		permissions[module.Code] = map[string]bool{
			"view":   false,
			"create": false,
			"edit":   false,
			"delete": false,
		}
	}

	// Get user's actual permissions
	userPerms, err := GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	// Merge with actual permissions
	for moduleCode, actions := range userPerms {
		for action, value := range actions {
			if permissions[moduleCode] != nil {
				permissions[moduleCode][action] = value
			}
		}
	}

	return permissions, nil
}
