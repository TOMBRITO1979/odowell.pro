package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// VerifyEmail handles email verification requests
func VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	db := database.GetDB()

	// Find the verification record
	var verification models.EmailVerification
	if err := db.Where("token = ?", token).First(&verification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Check if already verified
	if verification.Verified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already verified"})
		return
	}

	// Check if expired
	if verification.IsExpired() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token has expired. Please request a new verification email."})
		return
	}

	// Mark as verified
	verification.MarkAsVerified()
	if err := db.Save(&verification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	// Update tenant as verified
	if verification.TenantID > 0 {
		if err := db.Model(&models.Tenant{}).Where("id = ?", verification.TenantID).Update("email_verified", true).Error; err != nil {
			log.Printf("Failed to update tenant email_verified: %v", err)
		}
	}

	// Log the verification
	helpers.AuditAction(c, "email_verified", "tenant", verification.TenantID, true, map[string]interface{}{
		"email": verification.Email,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully. You can now log in.",
	})
}

// ResendVerificationEmail resends the verification email
func ResendVerificationEmail(c *gin.Context) {
	type ResendRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	var req ResendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Find tenant by email
	var tenant models.Tenant
	if err := db.Where("email = ?", req.Email).First(&tenant).Error; err != nil {
		// Don't reveal if email exists or not
		c.JSON(http.StatusOK, gin.H{
			"message": "If an account exists with this email, a verification link will be sent.",
		})
		return
	}

	// Check if already verified
	if tenant.EmailVerified {
		c.JSON(http.StatusOK, gin.H{
			"message": "If an account exists with this email, a verification link will be sent.",
		})
		return
	}

	// Delete any existing verification tokens for this tenant
	db.Where("tenant_id = ? AND verified = ?", tenant.ID, false).Delete(&models.EmailVerification{})

	// Generate new token
	token, err := models.GenerateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verification token"})
		return
	}

	// Create new verification record
	verification := models.EmailVerification{
		TenantID:  tenant.ID,
		Email:     tenant.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := db.Create(&verification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create verification record"})
		return
	}

	// Send verification email
	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	if err := helpers.SendVerificationEmail(tenant.Email, tenant.Name, token, baseURL); err != nil {
		log.Printf("Failed to send verification email: %v", err)
		// Don't reveal the error to the user
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If an account exists with this email, a verification link will be sent.",
	})
}

// CreateAndSendVerification creates a verification record and sends email
// This is called internally when a new tenant is registered
func CreateAndSendVerification(tenantID uint, email, name string) error {
	db := database.GetDB()

	// Generate token
	token, err := models.GenerateVerificationToken()
	if err != nil {
		return err
	}

	// Create verification record
	verification := models.EmailVerification{
		TenantID:  tenantID,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := db.Create(&verification).Error; err != nil {
		return err
	}

	// Send verification email
	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	return helpers.SendVerificationEmail(email, name, token, baseURL)
}
