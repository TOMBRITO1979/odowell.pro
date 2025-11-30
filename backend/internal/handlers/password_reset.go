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

// ForgotPassword handles password reset requests
// POST /api/auth/forgot-password
func ForgotPassword(c *gin.Context) {
	type ForgotRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	var req ForgotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Find user by email
	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists - always return success message
		c.JSON(http.StatusOK, gin.H{
			"message": "Se o email estiver cadastrado, você receberá instruções para redefinir sua senha.",
		})
		return
	}

	// Check if user is active
	if !user.Active {
		c.JSON(http.StatusOK, gin.H{
			"message": "Se o email estiver cadastrado, você receberá instruções para redefinir sua senha.",
		})
		return
	}

	// Delete any existing unused reset tokens for this user
	db.Where("user_id = ? AND used = ?", user.ID, false).Delete(&models.PasswordReset{})

	// Generate new token
	token, err := models.GenerateResetToken()
	if err != nil {
		log.Printf("Failed to generate reset token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar solicitação"})
		return
	}

	// Create reset record (expires in 1 hour)
	resetRecord := models.PasswordReset{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := db.Create(&resetRecord).Error; err != nil {
		log.Printf("Failed to create reset record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar solicitação"})
		return
	}

	// Send reset email asynchronously
	go func() {
		baseURL := os.Getenv("FRONTEND_URL")
		if baseURL == "" {
			baseURL = "https://app.odowell.pro"
		}

		if err := helpers.SendPasswordResetEmail(user.Email, user.Name, token, baseURL); err != nil {
			log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
		} else {
			log.Printf("Password reset email sent to %s", user.Email)
		}
	}()

	// Log the action
	helpers.AuditAction(c, "password_reset_requested", "user", user.ID, true, map[string]interface{}{
		"email": user.Email,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Se o email estiver cadastrado, você receberá instruções para redefinir sua senha.",
	})
}

// ResetPassword handles the actual password reset
// POST /api/auth/reset-password
func ResetPassword(c *gin.Context) {
	type ResetRequest struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	var req ResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate password strength
	if valid, msg := ValidatePassword(req.NewPassword); !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	db := database.GetDB()

	// Find the reset record
	var resetRecord models.PasswordReset
	if err := db.Where("token = ?", req.Token).First(&resetRecord).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token inválido ou expirado"})
		return
	}

	// Check if already used
	if resetRecord.Used {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Este link já foi utilizado"})
		return
	}

	// Check if expired
	if resetRecord.IsExpired() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link expirado. Solicite um novo link de redefinição."})
		return
	}

	// Find the user
	var user models.User
	if err := db.First(&user, resetRecord.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Hash new password
	if err := user.HashPassword(req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar senha"})
		return
	}

	// Start transaction
	tx := db.Begin()

	// Update user password
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar senha"})
		return
	}

	// Mark reset token as used
	resetRecord.MarkAsUsed()
	if err := tx.Save(&resetRecord).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar token"})
		return
	}

	tx.Commit()

	// Log the action
	helpers.AuditAction(c, "password_reset_completed", "user", user.ID, true, map[string]interface{}{
		"email": user.Email,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Senha redefinida com sucesso! Você já pode fazer login.",
	})
}

// ValidateResetToken checks if a reset token is valid
// GET /api/auth/validate-reset-token?token=xxx
func ValidateResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": "Token não fornecido"})
		return
	}

	db := database.GetDB()

	var resetRecord models.PasswordReset
	if err := db.Where("token = ?", token).First(&resetRecord).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false, "error": "Token inválido"})
		return
	}

	if resetRecord.Used {
		c.JSON(http.StatusOK, gin.H{"valid": false, "error": "Este link já foi utilizado"})
		return
	}

	if resetRecord.IsExpired() {
		c.JSON(http.StatusOK, gin.H{"valid": false, "error": "Link expirado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"email": resetRecord.Email,
	})
}
