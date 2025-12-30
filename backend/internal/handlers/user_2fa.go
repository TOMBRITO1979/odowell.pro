package handlers

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// Setup2FARequest is the request body for setting up 2FA
type Setup2FARequest struct {
	Password string `json:"password" binding:"required"`
}

// Verify2FARequest is the request body for verifying 2FA setup
type Verify2FARequest struct {
	Code string `json:"code" binding:"required"`
}

// Disable2FARequest is the request body for disabling 2FA
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// Setup2FAResponse contains the TOTP setup information
type Setup2FAResponse struct {
	Secret      string `json:"secret"`
	QRCodeURL   string `json:"qr_code_url"`
	OTPAuthURL  string `json:"otpauth_url"`
	BackupCodes []string `json:"backup_codes"`
}

// Get2FAStatus returns the current 2FA status for the user
func Get2FAStatus(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	c.JSON(http.StatusOK, gin.H{
		"two_factor_enabled": user.TwoFactorEnabled,
	})
}

// Setup2FA generates a new TOTP secret and returns QR code data
func Setup2FA(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var req Setup2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha é obrigatória"})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Senha incorreta"})
		return
	}

	// Check if 2FA is already enabled
	if user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA já está ativado"})
		return
	}

	// Get issuer name from environment or use default
	issuer := os.Getenv("APP_NAME")
	if issuer == "" {
		issuer = "Odowell"
	}

	// Generate TOTP key
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: user.Email,
		Period:      30,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		helpers.LogError("Failed to generate TOTP key", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar chave 2FA"})
		return
	}

	// Generate backup codes
	backupCodes := generateBackupCodes(8)

	// Encrypt secret before storing
	encryptedSecret, err := encryptSecret(key.Secret())
	if err != nil {
		helpers.LogError("Failed to encrypt 2FA secret", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar chave 2FA"})
		return
	}

	// Hash backup codes before storing
	hashedBackupCodes, err := hashBackupCodes(backupCodes)
	if err != nil {
		helpers.LogError("Failed to hash backup codes", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar códigos de backup"})
		return
	}

	backupCodesJSON, _ := json.Marshal(hashedBackupCodes)

	// Store encrypted secret temporarily (not enabled yet)
	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"two_factor_secret":       encryptedSecret,
		"two_factor_backup_codes": string(backupCodesJSON),
	}).Error; err != nil {
		helpers.LogError("Failed to save 2FA secret", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar configuração 2FA"})
		return
	}

	helpers.LogSecurityEvent("2fa_setup_initiated", user.ID, user.TenantID, true, nil)

	c.JSON(http.StatusOK, Setup2FAResponse{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		OTPAuthURL:  key.URL(),
		BackupCodes: backupCodes,
	})
}

// Verify2FA verifies the TOTP code and enables 2FA
func Verify2FA(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var req Verify2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código é obrigatório"})
		return
	}

	// Check if 2FA is already enabled
	if user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA já está ativado"})
		return
	}

	// Check if secret exists
	if user.TwoFactorSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Primeiro configure o 2FA"})
		return
	}

	// Decrypt secret
	secret, err := decryptSecret(user.TwoFactorSecret)
	if err != nil {
		helpers.LogError("Failed to decrypt 2FA secret", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar chave 2FA"})
		return
	}

	// Validate TOTP code
	valid := totp.Validate(req.Code, secret)
	if !valid {
		helpers.LogSecurityEvent("2fa_verification_failed", user.ID, user.TenantID, false, map[string]interface{}{
			"reason": "invalid_code",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Código inválido"})
		return
	}

	// Enable 2FA
	if err := database.DB.Model(&user).Update("two_factor_enabled", true).Error; err != nil {
		helpers.LogError("Failed to enable 2FA", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ativar 2FA"})
		return
	}

	helpers.LogSecurityEvent("2fa_enabled", user.ID, user.TenantID, true, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA ativado com sucesso",
	})
}

// Disable2FA disables 2FA for the user
func Disable2FA(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var req Disable2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha e código são obrigatórios"})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Senha incorreta"})
		return
	}

	// Check if 2FA is enabled
	if !user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA não está ativado"})
		return
	}

	// Decrypt secret
	secret, err := decryptSecret(user.TwoFactorSecret)
	if err != nil {
		helpers.LogError("Failed to decrypt 2FA secret", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar chave 2FA"})
		return
	}

	// Validate TOTP code
	valid := totp.Validate(req.Code, secret)
	if !valid {
		// Check if it's a backup code
		valid = validateBackupCode(req.Code, user.TwoFactorBackupCodes, user.ID)
	}

	if !valid {
		helpers.LogSecurityEvent("2fa_disable_failed", user.ID, user.TenantID, false, map[string]interface{}{
			"reason": "invalid_code",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Código inválido"})
		return
	}

	// Disable 2FA
	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"two_factor_enabled":      false,
		"two_factor_secret":       "",
		"two_factor_backup_codes": "",
	}).Error; err != nil {
		helpers.LogError("Failed to disable 2FA", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao desativar 2FA"})
		return
	}

	helpers.LogSecurityEvent("2fa_disabled", user.ID, user.TenantID, true, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "2FA desativado com sucesso",
	})
}

// RegenerateBackupCodes generates new backup codes
func RegenerateBackupCodes(c *gin.Context) {
	user := c.MustGet("user").(models.User)

	var req Setup2FARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha é obrigatória"})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Senha incorreta"})
		return
	}

	// Check if 2FA is enabled
	if !user.TwoFactorEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA não está ativado"})
		return
	}

	// Generate new backup codes
	backupCodes := generateBackupCodes(8)

	// Hash backup codes before storing
	hashedBackupCodes, err := hashBackupCodes(backupCodes)
	if err != nil {
		helpers.LogError("Failed to hash backup codes", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar códigos de backup"})
		return
	}

	backupCodesJSON, _ := json.Marshal(hashedBackupCodes)

	// Update backup codes
	if err := database.DB.Model(&user).Update("two_factor_backup_codes", string(backupCodesJSON)).Error; err != nil {
		helpers.LogError("Failed to save backup codes", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar códigos de backup"})
		return
	}

	helpers.LogSecurityEvent("2fa_backup_codes_regenerated", user.ID, user.TenantID, true, nil)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Códigos de backup regenerados com sucesso",
		"backup_codes": backupCodes,
	})
}

// Validate2FACode validates a TOTP code (used during login)
func Validate2FACode(userID uint, code string) bool {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return false
	}

	if !user.TwoFactorEnabled || user.TwoFactorSecret == "" {
		return true // 2FA not enabled, allow
	}

	// Decrypt secret
	secret, err := decryptSecret(user.TwoFactorSecret)
	if err != nil {
		helpers.LogError("Failed to decrypt 2FA secret during validation", err)
		return false
	}

	// Validate TOTP code
	valid := totp.Validate(code, secret)
	if valid {
		return true
	}

	// Check backup code
	return validateBackupCode(code, user.TwoFactorBackupCodes, user.ID)
}

// Helper functions

// generateBackupCodes generates n random backup codes
func generateBackupCodes(n int) []string {
	codes := make([]string, n)
	for i := 0; i < n; i++ {
		bytes := make([]byte, 5)
		rand.Read(bytes)
		codes[i] = strings.ToUpper(base32.StdEncoding.EncodeToString(bytes)[:8])
	}
	return codes
}

// hashBackupCodes hashes backup codes using bcrypt
func hashBackupCodes(codes []string) ([]string, error) {
	hashed := make([]string, len(codes))
	for i, code := range codes {
		hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hashed[i] = string(hash)
	}
	return hashed, nil
}

// validateBackupCode checks if the code matches any backup code
func validateBackupCode(code string, backupCodesJSON string, userID uint) bool {
	if backupCodesJSON == "" {
		return false
	}

	var hashedCodes []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &hashedCodes); err != nil {
		return false
	}

	code = strings.ToUpper(strings.TrimSpace(code))

	for i, hashedCode := range hashedCodes {
		if bcrypt.CompareHashAndPassword([]byte(hashedCode), []byte(code)) == nil {
			// Code matches, remove it from the list
			hashedCodes = append(hashedCodes[:i], hashedCodes[i+1:]...)
			newJSON, _ := json.Marshal(hashedCodes)
			database.DB.Model(&models.User{}).Where("id = ?", userID).Update("two_factor_backup_codes", string(newJSON))

			helpers.LogSecurityEvent("2fa_backup_code_used", userID, 0, true, map[string]interface{}{
				"codes_remaining": len(hashedCodes),
			})

			return true
		}
	}

	return false
}

// encryptSecret encrypts the TOTP secret using AES
func encryptSecret(secret string) (string, error) {
	key := get2FAEncryptionKey()
	encrypted, err := helpers.EncryptAES([]byte(secret), key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// decryptSecret decrypts the TOTP secret using AES
func decryptSecret(encryptedSecret string) (string, error) {
	key := get2FAEncryptionKey()
	encrypted, err := base64.StdEncoding.DecodeString(encryptedSecret)
	if err != nil {
		return "", err
	}
	decrypted, err := helpers.DecryptAES(encrypted, key)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// get2FAEncryptionKey returns the encryption key for 2FA secrets
func get2FAEncryptionKey() []byte {
	key := os.Getenv("TWO_FA_ENCRYPTION_KEY")
	if key == "" {
		// Fallback to JWT_SECRET if not set (not ideal but better than nothing)
		key = os.Getenv("JWT_SECRET")
	}
	// Ensure key is 32 bytes for AES-256
	if len(key) < 32 {
		key = fmt.Sprintf("%-32s", key) // Pad with spaces
	}
	return []byte(key[:32])
}

// Check2FARequired returns whether 2FA is required for login
func Check2FARequired(userID uint) bool {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return false
	}
	return user.TwoFactorEnabled
}

// Verify2FALogin is the endpoint for 2FA verification during login
func Verify2FALogin(c *gin.Context) {
	var req struct {
		UserID uint   `json:"user_id" binding:"required"`
		Code   string `json:"code" binding:"required"`
		Token  string `json:"token" binding:"required"` // Temporary token from first login step
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	// Validate temporary token
	claims, err := helpers.ValidateTempToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido ou expirado"})
		return
	}

	if claims.UserID != req.UserID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	// Validate 2FA code
	if !Validate2FACode(req.UserID, req.Code) {
		helpers.LogSecurityEvent("2fa_login_failed", req.UserID, claims.TenantID, false, map[string]interface{}{
			"reason": "invalid_code",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Código 2FA inválido"})
		return
	}

	// Generate full access token
	var user models.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Usuário não encontrado"})
		return
	}

	accessToken, err := helpers.GenerateToken(user.ID, user.TenantID, user.Role, user.IsSuperAdmin, user.PatientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar token"})
		return
	}

	refreshToken, err := helpers.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar refresh token"})
		return
	}

	// Set httpOnly cookies
	helpers.SetAuthCookies(c, accessToken, refreshToken)

	helpers.LogSecurityEvent("2fa_login_success", user.ID, user.TenantID, true, nil)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":               user.ID,
			"name":             user.Name,
			"email":            user.Email,
			"role":             user.Role,
			"is_super_admin":   user.IsSuperAdmin,
			"tenant_id":        user.TenantID,
			"profile_picture":  user.ProfilePicture,
			"two_factor_enabled": user.TwoFactorEnabled,
		},
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
