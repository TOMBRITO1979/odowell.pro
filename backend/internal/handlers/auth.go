package handlers

import (
	"crypto/rand"
	"drcrwell/backend/internal/cache"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ValidatePassword checks if password meets security requirements:
// - Minimum 12 characters
// - At least 1 uppercase letter
// - At least 1 lowercase letter
// - At least 1 number
// - At least 1 special character
func ValidatePassword(password string) (bool, string) {
	if len(password) < 12 {
		return false, "A senha deve ter no mínimo 12 caracteres"
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return false, "A senha deve conter pelo menos 1 letra maiúscula"
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return false, "A senha deve conter pelo menos 1 letra minúscula"
	}

	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return false, "A senha deve conter pelo menos 1 número"
	}

	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]`).MatchString(password)
	if !hasSpecial {
		return false, "A senha deve conter pelo menos 1 caractere especial (!@#$%^&*)"
	}

	return true, ""
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	TenantID uint   `json:"tenant_id" binding:"required"`
}

type Claims struct {
	UserID       uint                       `json:"user_id"`
	TenantID     uint                       `json:"tenant_id"`
	Email        string                     `json:"email"`
	Role         string                     `json:"role"`
	IsSuperAdmin bool                       `json:"is_super_admin"`
	TenantActive bool                       `json:"tenant_active,omitempty"` // Cached tenant active status
	PatientID    *uint                      `json:"patient_id,omitempty"`    // For patient portal users
	Permissions  map[string]map[string]bool `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

// Login authenticates a user and returns JWT token
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	db := database.GetDB()

	// Find user by email
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		helpers.AuditLogin(c, req.Email, false, map[string]interface{}{"reason": "user_not_found"})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user is active
	if !user.Active {
		helpers.AuditLogin(c, req.Email, false, map[string]interface{}{"reason": "user_inactive"})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is inactive"})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		helpers.AuditLogin(c, req.Email, false, map[string]interface{}{"reason": "wrong_password"})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if tenant is active
	var tenant models.Tenant
	if err := db.Where("id = ? AND active = ?", user.TenantID, true).First(&tenant).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant is inactive or not found"})
		return
	}

	// Check if tenant email is verified (skip for super admins)
	if !tenant.EmailVerified && !user.IsSuperAdmin {
		helpers.AuditLogin(c, req.Email, false, map[string]interface{}{"reason": "email_not_verified"})
		c.JSON(http.StatusForbidden, gin.H{
			"error":              "Email não verificado. Por favor, verifique seu email para ativar a conta.",
			"email_not_verified": true,
			"tenant_email":       tenant.Email,
		})
		return
	}

	// Check if 2FA is enabled for this user
	if user.TwoFactorEnabled {
		// Generate temporary token for 2FA verification
		tempToken, err := helpers.GenerateTempToken(user.ID, user.TenantID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate 2FA token"})
			return
		}

		helpers.LogSecurityEvent("2fa_required", user.ID, user.TenantID, true, nil)

		c.JSON(http.StatusOK, gin.H{
			"requires_2fa": true,
			"temp_token":   tempToken,
			"user_id":      user.ID,
			"message":      "Por favor, insira o código de autenticação de dois fatores",
		})
		return
	}

	// Generate access token (short-lived)
	accessToken, err := generateAccessToken(user.ID, user.TenantID, user.Email, user.Role, user.IsSuperAdmin, user.PatientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	// Generate refresh token (long-lived, stored in Redis)
	refreshToken, err := generateSecureToken(32) // 64 hex characters
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	// Store refresh token in Redis (if Redis is available)
	if cache.GetClient() != nil {
		tokenData := cache.RefreshTokenData{
			UserID:       user.ID,
			TenantID:     user.TenantID,
			Email:        user.Email,
			Role:         user.Role,
			IsSuperAdmin: user.IsSuperAdmin,
		}
		if err := cache.StoreRefreshToken(refreshToken, tokenData); err != nil {
			log.Printf("WARNING: Failed to store refresh token in Redis: %v", err)
			// Continue without refresh token - access token will still work
		}
	}

	// Log successful login
	helpers.AuditLogin(c, req.Email, true, map[string]interface{}{
		"user_id":   user.ID,
		"tenant_id": user.TenantID,
		"role":      user.Role,
	})

	// Set HTTP-only cookies with tokens
	// Cookie settings for security:
	// - HttpOnly: prevents JavaScript access (XSS protection)
	// - Secure: only sent over HTTPS (in production)
	// - SameSite=Strict: CSRF protection while allowing normal navigation
	secure := os.Getenv("GIN_MODE") == "release"
	accessMaxAge := int(cache.AccessTokenExpiry.Seconds())
	refreshMaxAge := int(cache.RefreshTokenExpiry.Seconds())

	// Set access token cookie (short-lived)
	c.Header("Set-Cookie", fmt.Sprintf("auth_token=%s; Path=/; Max-Age=%d; HttpOnly; SameSite=Strict%s",
		accessToken, accessMaxAge, func() string {
			if secure {
				return "; Secure"
			}
			return ""
		}()))

	// Set refresh token cookie (long-lived) - only for /api/auth path
	c.Writer.Header().Add("Set-Cookie", fmt.Sprintf("refresh_token=%s; Path=/api/auth; Max-Age=%d; HttpOnly; SameSite=Strict%s",
		refreshToken, refreshMaxAge, func() string {
			if secure {
				return "; Secure"
			}
			return ""
		}()))

	c.JSON(http.StatusOK, gin.H{
		"token": accessToken, // Keep for backward compatibility during migration
		"expires_in": accessMaxAge,
		"user": gin.H{
			"id":             user.ID,
			"name":           user.Name,
			"email":          user.Email,
			"role":           user.Role,
			"tenant_id":      user.TenantID,
			"hide_sidebar":   user.HideSidebar,
			"is_super_admin": user.IsSuperAdmin,
			"patient_id":     user.PatientID,
		},
		"tenant": gin.H{
			"id":                  tenant.ID,
			"name":                tenant.Name,
			"subscription_status": tenant.SubscriptionStatus,
			"trial_ends_at":       tenant.TrialEndsAt,
			"plan_type":           tenant.PlanType,
		},
	})
}

// Logout clears auth and refresh cookies and invalidates refresh token
func Logout(c *gin.Context) {
	secure := os.Getenv("GIN_MODE") == "release"

	// Delete refresh token from Redis if present
	if refreshToken, err := c.Cookie("refresh_token"); err == nil && refreshToken != "" {
		if cache.GetClient() != nil {
			if err := cache.DeleteRefreshToken(refreshToken); err != nil {
				log.Printf("WARNING: Failed to delete refresh token from Redis: %v", err)
			}
		}
	}

	// Clear auth token cookie
	c.Header("Set-Cookie", fmt.Sprintf("auth_token=; Path=/; Max-Age=-1; HttpOnly; SameSite=Strict%s",
		func() string {
			if secure {
				return "; Secure"
			}
			return ""
		}()))

	// Clear refresh token cookie
	c.Writer.Header().Add("Set-Cookie", fmt.Sprintf("refresh_token=; Path=/api/auth; Max-Age=-1; HttpOnly; SameSite=Strict%s",
		func() string {
			if secure {
				return "; Secure"
			}
			return ""
		}()))

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// Register creates a new user account
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password strength
	if valid, msg := ValidatePassword(req.Password); !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	db := database.GetDB()

	// Check if email already exists
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Verify tenant exists and is active
	var tenant models.Tenant
	if err := db.Where("id = ? AND active = ?", req.TenantID, true).First(&tenant).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or inactive tenant"})
		return
	}

	// Create new user
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		TenantID: req.TenantID,
		Role:     "user",
		Active:   true,
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

// GetMe returns current user info
func GetMe(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	db := database.GetDB()

	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var tenant models.Tenant
	db.First(&tenant, user.TenantID)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":              user.ID,
			"name":            user.Name,
			"email":           user.Email,
			"role":            user.Role,
			"tenant_id":       user.TenantID,
			"cro":             user.CRO,
			"specialty":       user.Specialty,
			"profile_picture": user.ProfilePicture,
			"hide_sidebar":    user.HideSidebar,
			"is_super_admin":  user.IsSuperAdmin,
		},
		"tenant": gin.H{
			"id":                  tenant.ID,
			"name":                tenant.Name,
			"subdomain":           tenant.Subdomain,
			"subscription_status": tenant.SubscriptionStatus,
			"trial_ends_at":       tenant.TrialEndsAt,
			"plan_type":           tenant.PlanType,
		},
	})
}

// UpdateProfile updates user profile information
func UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	type UpdateProfileRequest struct {
		Name      string `json:"name" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Phone     string `json:"phone"`
		CRO       string `json:"cro"`
		Specialty string `json:"specialty"`
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Check if email is already taken by another user
	var existingUser models.User
	if err := db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	// Get user
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user
	user.Name = req.Name
	user.Email = req.Email
	user.Phone = req.Phone
	user.CRO = req.CRO
	user.Specialty = req.Specialty

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":              user.ID,
			"name":            user.Name,
			"email":           user.Email,
			"phone":           user.Phone,
			"cro":             user.CRO,
			"specialty":       user.Specialty,
			"role":            user.Role,
			"tenant_id":       user.TenantID,
			"profile_picture": user.ProfilePicture,
		},
	})
}

// ChangePassword changes user password
func ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")

	type ChangePasswordRequest struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate new password strength
	if valid, msg := ValidatePassword(req.NewPassword); !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	db := database.GetDB()

	// Get user
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify current password
	if !user.CheckPassword(req.CurrentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	if err := user.HashPassword(req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Save new password
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// UploadProfilePicture handles profile picture upload
func UploadProfilePicture(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size must be less than 5MB"})
		return
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPEG and PNG images are allowed"})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "/root/uploads/profile-pictures"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("user_%d_%d.jpg", userID, time.Now().Unix())
	filepath := fmt.Sprintf("%s/%s", uploadsDir, filename)

	// Save the file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Update user profile picture in database
	db := database.GetDB()
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete old profile picture if exists
	if user.ProfilePicture != "" {
		oldPath := fmt.Sprintf("/root/%s", user.ProfilePicture)
		os.Remove(oldPath)
	}

	// Update user with new profile picture path
	user.ProfilePicture = fmt.Sprintf("uploads/profile-pictures/%s", filename)
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile picture"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Profile picture uploaded successfully",
		"profile_picture": user.ProfilePicture,
	})
}

// generateSecureToken creates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateAccessToken creates a short-lived JWT access token
func generateAccessToken(userID, tenantID uint, email, role string, isSuperAdmin bool, patientID *uint) (string, error) {
	// Get user permissions (admins get all permissions in the middleware, but we still include them for consistency)
	var permissions map[string]map[string]bool
	var err error

	if role == "admin" {
		// Admins get all permissions
		permissions, err = middleware.GetAllUserPermissionsWithDefaults(userID)
		if err != nil {
			// Log error but continue - admin will have bypass anyway
			permissions = make(map[string]map[string]bool)
		}
	} else if role == "patient" {
		// Patients don't need permissions - they have fixed access to patient portal
		permissions = make(map[string]map[string]bool)
	} else {
		// Regular users get their assigned permissions
		permissions, err = middleware.GetUserPermissions(userID)
		if err != nil {
			// Log error but continue with empty permissions
			log.Printf("ERROR loading permissions for user %d: %v", userID, err)
			permissions = make(map[string]map[string]bool)
		}
	}

	claims := Claims{
		UserID:       userID,
		TenantID:     tenantID,
		Email:        email,
		Role:         role,
		IsSuperAdmin: isSuperAdmin,
		TenantActive: true, // PERFORMANCE: Cache tenant active status in JWT to skip DB lookup
		PatientID:    patientID,
		Permissions:  permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cache.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// generateToken creates a JWT token (legacy - uses 24 hour expiry for backward compatibility)
func generateToken(userID, tenantID uint, email, role string, isSuperAdmin bool) (string, error) {
	// Get user permissions (admins get all permissions in the middleware, but we still include them for consistency)
	var permissions map[string]map[string]bool
	var err error

	if role == "admin" {
		// Admins get all permissions
		permissions, err = middleware.GetAllUserPermissionsWithDefaults(userID)
		if err != nil {
			// Log error but continue - admin will have bypass anyway
			permissions = make(map[string]map[string]bool)
		}
	} else {
		// Regular users get their assigned permissions
		permissions, err = middleware.GetUserPermissions(userID)
		if err != nil {
			// Log error but continue with empty permissions
			log.Printf("ERROR loading permissions for user %d: %v", userID, err)
			permissions = make(map[string]map[string]bool)
		}
	}

	claims := Claims{
		UserID:       userID,
		TenantID:     tenantID,
		Email:        email,
		Role:         role,
		IsSuperAdmin: isSuperAdmin,
		TenantActive: true, // PERFORMANCE: Cache tenant active status in JWT
		Permissions:  permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// RefreshToken handles token refresh using refresh token cookie
func RefreshToken(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token required"})
		return
	}

	// Validate refresh token in Redis
	tokenData, err := cache.GetRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Verify user still exists and is active
	db := database.GetDB()
	var user models.User
	if err := db.First(&user, tokenData.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User account is inactive"})
		return
	}

	// Verify tenant is still active
	var tenant models.Tenant
	if err := db.Where("id = ? AND active = ?", user.TenantID, true).First(&tenant).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant is inactive"})
		return
	}

	// Generate new access token
	accessToken, err := generateAccessToken(user.ID, user.TenantID, user.Email, user.Role, user.IsSuperAdmin, user.PatientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set new access token cookie
	secure := os.Getenv("GIN_MODE") == "release"
	accessMaxAge := int(cache.AccessTokenExpiry.Seconds())

	c.Header("Set-Cookie", fmt.Sprintf("auth_token=%s; Path=/; Max-Age=%d; HttpOnly; SameSite=Strict%s",
		accessToken, accessMaxAge, func() string {
			if secure {
				return "; Secure"
			}
			return ""
		}()))

	c.JSON(http.StatusOK, gin.H{
		"token": accessToken, // For backward compatibility
		"expires_in": accessMaxAge,
	})
}
