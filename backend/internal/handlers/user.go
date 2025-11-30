package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTenantUsers returns all users in the current tenant
func GetTenantUsers(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")

	db := database.GetDB()

	type UserResponse struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		Role        string `json:"role"`
		Active      bool   `json:"active"`
		Phone       string `json:"phone,omitempty"`
		CRO         string `json:"cro,omitempty"`
		Specialty   string `json:"specialty,omitempty"`
		HideSidebar bool   `json:"hide_sidebar"`
	}

	var users []UserResponse
	err := db.Table("public.users").
		Select("id, name, email, role, active, phone, cro, specialty, hide_sidebar").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("name ASC").
		Find(&users).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": len(users),
	})
}

// CreateUserRequest represents request to create a new user
type CreateUserRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Role      string `json:"role" binding:"required"`
	Phone     string `json:"phone"`
	CRO       string `json:"cro"`
	Specialty string `json:"specialty"`
}

// CreateTenantUser creates a new user in the current tenant
func CreateTenantUser(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password strength
	if valid, msg := ValidatePassword(req.Password); !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Validate role
	if req.Role != "admin" && req.Role != "dentist" && req.Role != "receptionist" && req.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	db := database.GetDB()

	// Check if email already exists
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Create user
	user := models.User{
		Name:      req.Name,
		Email:     req.Email,
		TenantID:  tenantID.(uint),
		Role:      req.Role,
		Phone:     req.Phone,
		CRO:       req.CRO,
		Specialty: req.Specialty,
		Active:    true,
	}

	if err := user.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// UpdateUserRequest represents request to update user info
type UpdateUserRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Role      string `json:"role" binding:"required"`
	Active    bool   `json:"active"`
	Phone     string `json:"phone"`
	CRO       string `json:"cro"`
	Specialty string `json:"specialty"`
}

// UpdateTenantUser updates a user in the current tenant
func UpdateTenantUser(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	userID := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	if req.Role != "admin" && req.Role != "dentist" && req.Role != "receptionist" && req.Role != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	db := database.GetDB()

	// Get user
	var user models.User
	if err := db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if email is taken by another user
	var existingUser models.User
	if err := db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	// Update user
	user.Name = req.Name
	user.Email = req.Email
	user.Role = req.Role
	user.Active = req.Active
	user.Phone = req.Phone
	user.CRO = req.CRO
	user.Specialty = req.Specialty

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user": gin.H{
			"id":     user.ID,
			"name":   user.Name,
			"email":  user.Email,
			"role":   user.Role,
			"active": user.Active,
		},
	})
}

// UpdateUserSidebarRequest represents request to update sidebar preference
type UpdateUserSidebarRequest struct {
	HideSidebar bool `json:"hide_sidebar"`
}

// UpdateUserSidebar updates the hide_sidebar preference for a user
func UpdateUserSidebar(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	userID := c.Param("id")

	var req UpdateUserSidebarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Update the user's sidebar preference
	result := db.Exec(
		"UPDATE public.users SET hide_sidebar = ?, updated_at = NOW() WHERE id = ? AND tenant_id = ?",
		req.HideSidebar, userID, tenantID,
	)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sidebar preference"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Sidebar preference updated successfully",
		"hide_sidebar": req.HideSidebar,
	})
}
