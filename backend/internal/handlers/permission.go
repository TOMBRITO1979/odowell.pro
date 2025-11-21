package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetModules returns all system modules
func GetModules(c *gin.Context) {
	db := database.GetDB()

	var modules []models.Module
	if err := db.Where("active = ?", true).Order("name ASC").Find(&modules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch modules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"modules": modules,
		"total":   len(modules),
	})
}

// GetAllPermissions returns all permissions grouped by module
func GetAllPermissions(c *gin.Context) {
	db := database.GetDB()

	type PermissionWithModule struct {
		ID          uint   `json:"id"`
		ModuleID    uint   `json:"module_id"`
		ModuleCode  string `json:"module_code"`
		ModuleName  string `json:"module_name"`
		Action      string `json:"action"`
		Description string `json:"description"`
	}

	var permissions []PermissionWithModule
	err := db.Table("public.permissions p").
		Select("p.id, p.module_id, m.code as module_code, m.name as module_name, p.action, p.description").
		Joins("INNER JOIN public.modules m ON m.id = p.module_id").
		Where("p.deleted_at IS NULL AND m.deleted_at IS NULL AND m.active = true").
		Order("m.name ASC, p.action ASC").
		Find(&permissions).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": permissions,
		"total":       len(permissions),
	})
}

// GetUserPermissions returns permissions for a specific user
func GetUserPermissions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	db := database.GetDB()

	// Check if user exists
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get user's permissions
	permissions, err := middleware.GetAllUserPermissionsWithDefaults(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
		"permissions": permissions,
	})
}

// UpdateUserPermissionsRequest represents the request body for updating permissions
type UpdateUserPermissionsRequest struct {
	Permissions map[string]map[string]bool `json:"permissions" binding:"required"`
}

// UpdateUserPermissions updates permissions for a specific user
func UpdateUserPermissions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Check if user exists
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get the admin user who is granting permissions
	grantedBy, _ := c.Get("user_id")
	grantedByID := grantedBy.(uint)

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all existing permissions for this user
	if err := tx.Where("user_id = ?", userID).Delete(&models.UserPermission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing permissions"})
		return
	}

	// Insert new permissions
	permissionsGranted := 0
	for moduleCode, actions := range req.Permissions {
		// Get module ID
		var module models.Module
		if err := tx.Where("code = ?", moduleCode).First(&module).Error; err != nil {
			continue // Skip if module not found
		}

		for action, granted := range actions {
			if !granted {
				continue // Skip if permission not granted
			}

			// Get permission ID
			var permission models.Permission
			if err := tx.Where("module_id = ? AND action = ?", module.ID, action).First(&permission).Error; err != nil {
				continue // Skip if permission not found
			}

			// Create user permission
			userPermission := models.UserPermission{
				UserID:       uint(userID),
				PermissionID: permission.ID,
				GrantedBy:    &grantedByID,
			}

			if err := tx.Create(&userPermission).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grant permission"})
				return
			}

			permissionsGranted++
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Permissions updated successfully",
		"permissions_granted": permissionsGranted,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// BulkUpdateUserPermissionsRequest represents bulk permission update request
type BulkUpdateUserPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// BulkUpdateUserPermissions updates user permissions using permission IDs (faster)
func BulkUpdateUserPermissions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req BulkUpdateUserPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Check if user exists
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get the admin user who is granting permissions
	grantedBy, _ := c.Get("user_id")
	grantedByID := grantedBy.(uint)

	// Start transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete all existing permissions for this user
	if err := tx.Where("user_id = ?", userID).Delete(&models.UserPermission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing permissions"})
		return
	}

	// Insert new permissions
	for _, permissionID := range req.PermissionIDs {
		userPermission := models.UserPermission{
			UserID:       uint(userID),
			PermissionID: permissionID,
			GrantedBy:    &grantedByID,
		}

		if err := tx.Create(&userPermission).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grant permission"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Permissions updated successfully",
		"permissions_granted": len(req.PermissionIDs),
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// GetDefaultRolePermissions returns default permissions for a role
func GetDefaultRolePermissions(c *gin.Context) {
	role := c.Param("role")

	if role != "admin" && role != "dentist" && role != "receptionist" && role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	db := database.GetDB()

	// Get default permissions view
	type DefaultPermission struct {
		ModuleCode    string `json:"module_code"`
		Action        string `json:"action"`
		HasPermission bool   `json:"has_permission"`
	}

	var defaultPerms []DefaultPermission
	err := db.Raw(`
		SELECT module_code, action, has_permission
		FROM public.default_role_permissions
		WHERE role = ?
		ORDER BY module_code, action
	`, role).Scan(&defaultPerms).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch default permissions"})
		return
	}

	// Convert to map structure
	permissions := make(map[string]map[string]bool)
	for _, perm := range defaultPerms {
		if permissions[perm.ModuleCode] == nil {
			permissions[perm.ModuleCode] = make(map[string]bool)
		}
		permissions[perm.ModuleCode][perm.Action] = perm.HasPermission
	}

	c.JSON(http.StatusOK, gin.H{
		"role":        role,
		"permissions": permissions,
	})
}
