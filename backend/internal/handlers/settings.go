package handlers

import (
	"crypto/rand"
	"crypto/tls"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/models"
	"drcrwell/backend/internal/middleware"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ClinicInfo holds the clinic information for PDFs and documents
type ClinicInfo struct {
	Name    string
	Address string
	Phone   string
	Email   string
	CNPJ    string
}

// GetClinicInfo returns clinic info from settings (preferred) or tenant (fallback)
func GetClinicInfo(db *gorm.DB, tenantID uint) ClinicInfo {
	// First try to get from settings
	var settings models.TenantSettings
	if err := db.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err == nil {
		if settings.ClinicName != "" {
			// Build address
			address := settings.ClinicAddress
			if settings.ClinicCity != "" {
				if address != "" {
					address += ", "
				}
				address += settings.ClinicCity
			}
			if settings.ClinicState != "" {
				address += " - " + settings.ClinicState
			}
			if settings.ClinicZip != "" {
				address += " - CEP: " + settings.ClinicZip
			}

			return ClinicInfo{
				Name:    settings.ClinicName,
				Address: address,
				Phone:   settings.ClinicPhone,
				Email:   settings.ClinicEmail,
				CNPJ:    settings.ClinicCNPJ,
			}
		}
	}

	// Fallback to tenant info
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err == nil {
		address := tenant.Address
		if tenant.City != "" {
			if address != "" {
				address += ", "
			}
			address += tenant.City
		}
		if tenant.State != "" {
			address += " - " + tenant.State
		}
		if tenant.ZipCode != "" {
			address += " - CEP: " + tenant.ZipCode
		}

		return ClinicInfo{
			Name:    tenant.Name,
			Address: address,
			Phone:   tenant.Phone,
			Email:   tenant.Email,
			CNPJ:    "", // Tenant model doesn't have CNPJ
		}
	}

	return ClinicInfo{Name: "Clínica"}
}

func GetTenantSettings(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	var settings models.TenantSettings
	result := db.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Return empty settings if not found
			c.JSON(http.StatusOK, gin.H{
				"settings": models.TenantSettings{
					TenantID:   tenantID,
					SMTPPort:   587,
					SMTPUseTLS: true,
				},
				"has_smtp_password": false,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	// Check if SMTP password exists before clearing it
	hasSMTPPassword := settings.SMTPPassword != ""

	// Don't return sensitive data
	settings.SMTPPassword = ""

	c.JSON(http.StatusOK, gin.H{
		"settings":          settings,
		"has_smtp_password": hasSMTPPassword,
	})
}

func UpdateTenantSettings(c *gin.Context) {
	var input models.TenantSettings
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	// Encrypt SMTP password if provided
	smtpPassword := ""
	if input.SMTPPassword != "" {
		encrypted, err := helpers.EncryptIfNeeded(input.SMTPPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt SMTP password"})
			return
		}
		smtpPassword = encrypted
	}

	// Check if settings exist
	var count int64
	db.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).Count(&count)

	if count == 0 {
		// Create new settings with encrypted password
		input.TenantID = tenantID
		input.SMTPPassword = smtpPassword
		if err := db.Table("public.tenant_settings").Create(&input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settings", "details": err.Error()})
			return
		}
		// Don't return password in response
		input.SMTPPassword = ""
		c.JSON(http.StatusCreated, gin.H{"settings": input, "has_smtp_password": smtpPassword != ""})
		return
	}

	// Build update query - only update password if provided
	var result *gorm.DB
	if smtpPassword != "" {
		// Update including password
		result = db.Exec(`
			UPDATE public.tenant_settings
			SET
				clinic_name = ?, clinic_cnpj = ?, clinic_address = ?, clinic_city = ?,
				clinic_state = ?, clinic_zip = ?, clinic_phone = ?, clinic_email = ?,
				working_hours_start = ?, working_hours_end = ?, default_appointment_duration = ?,
				payment_cash_enabled = ?, payment_credit_card_enabled = ?, payment_debit_card_enabled = ?,
				payment_pix_enabled = ?, payment_transfer_enabled = ?, payment_insurance_enabled = ?,
				smtp_host = ?, smtp_port = ?, smtp_username = ?, smtp_password = ?,
				smtp_from_name = ?, smtp_from_email = ?, smtp_use_tls = ?,
				whatsapp_api_key = ?, whatsapp_number = ?, sms_api_key = ?, sms_provider = ?,
				updated_at = NOW()
			WHERE tenant_id = ?
		`,
			input.ClinicName, input.ClinicCNPJ, input.ClinicAddress, input.ClinicCity,
			input.ClinicState, input.ClinicZip, input.ClinicPhone, input.ClinicEmail,
			input.WorkingHoursStart, input.WorkingHoursEnd, input.DefaultAppointmentDuration,
			input.PaymentCashEnabled, input.PaymentCreditCardEnabled, input.PaymentDebitCardEnabled,
			input.PaymentPixEnabled, input.PaymentTransferEnabled, input.PaymentInsuranceEnabled,
			input.SMTPHost, input.SMTPPort, input.SMTPUsername, smtpPassword,
			input.SMTPFromName, input.SMTPFromEmail, input.SMTPUseTLS,
			input.WhatsAppAPIKey, input.WhatsAppNumber, input.SMSAPIKey, input.SMSProvider,
			tenantID,
		)
	} else {
		// Update without changing password
		result = db.Exec(`
			UPDATE public.tenant_settings
			SET
				clinic_name = ?, clinic_cnpj = ?, clinic_address = ?, clinic_city = ?,
				clinic_state = ?, clinic_zip = ?, clinic_phone = ?, clinic_email = ?,
				working_hours_start = ?, working_hours_end = ?, default_appointment_duration = ?,
				payment_cash_enabled = ?, payment_credit_card_enabled = ?, payment_debit_card_enabled = ?,
				payment_pix_enabled = ?, payment_transfer_enabled = ?, payment_insurance_enabled = ?,
				smtp_host = ?, smtp_port = ?, smtp_username = ?,
				smtp_from_name = ?, smtp_from_email = ?, smtp_use_tls = ?,
				whatsapp_api_key = ?, whatsapp_number = ?, sms_api_key = ?, sms_provider = ?,
				updated_at = NOW()
			WHERE tenant_id = ?
		`,
			input.ClinicName, input.ClinicCNPJ, input.ClinicAddress, input.ClinicCity,
			input.ClinicState, input.ClinicZip, input.ClinicPhone, input.ClinicEmail,
			input.WorkingHoursStart, input.WorkingHoursEnd, input.DefaultAppointmentDuration,
			input.PaymentCashEnabled, input.PaymentCreditCardEnabled, input.PaymentDebitCardEnabled,
			input.PaymentPixEnabled, input.PaymentTransferEnabled, input.PaymentInsuranceEnabled,
			input.SMTPHost, input.SMTPPort, input.SMTPUsername,
			input.SMTPFromName, input.SMTPFromEmail, input.SMTPUseTLS,
			input.WhatsAppAPIKey, input.WhatsAppNumber, input.SMSAPIKey, input.SMSProvider,
			tenantID,
		)
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings", "details": result.Error.Error()})
		return
	}

	// Also update the tenant name to match the clinic name
	if input.ClinicName != "" {
		db.Exec("UPDATE public.tenants SET name = ?, updated_at = NOW() WHERE id = ?", input.ClinicName, tenantID)
	}

	// Check if password exists in DB (for has_smtp_password flag)
	var existingSettings models.TenantSettings
	db.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&existingSettings)
	hasSMTPPassword := existingSettings.SMTPPassword != ""

	// Don't return password in response
	input.SMTPPassword = ""
	c.JSON(http.StatusOK, gin.H{"settings": input, "has_smtp_password": hasSMTPPassword})
}

// GetEmbedToken returns the current embed token status
func GetEmbedToken(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	if tenant.EmbedToken == "" {
		c.JSON(http.StatusOK, gin.H{
			"token":     "",
			"embed_url": "",
		})
		return
	}

	// Build embed URL
	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		baseURL = "https://app.odowell.pro"
	}
	embedURL := fmt.Sprintf("%s/embed?token=%s", baseURL, tenant.EmbedToken)

	c.JSON(http.StatusOK, gin.H{
		"token":     tenant.EmbedToken,
		"embed_url": embedURL,
	})
}

// GenerateEmbedToken generates a new embed token for the tenant
func GenerateEmbedToken(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	// Generate random token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	token := hex.EncodeToString(bytes)

	// Update tenant with new token
	if err := db.Exec("UPDATE public.tenants SET embed_token = ?, updated_at = NOW() WHERE id = ?", token, tenantID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	// Build embed URL
	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		baseURL = "https://app.odowell.pro"
	}
	embedURL := fmt.Sprintf("%s/embed?token=%s", baseURL, token)

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"embed_url": embedURL,
	})
}

// RevokeEmbedToken removes the embed token for the tenant
func RevokeEmbedToken(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	if err := db.Exec("UPDATE public.tenants SET embed_token = '', updated_at = NOW() WHERE id = ?", tenantID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token revoked successfully"})
}

// TestSMTPConnection tests the SMTP connection with tenant's settings
func TestSMTPConnection(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	// Get tenant settings
	var settings models.TenantSettings
	if err := db.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configurações SMTP não encontradas"})
		return
	}

	// Validate required fields
	if settings.SMTPHost == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host SMTP não configurado"})
		return
	}
	if settings.SMTPUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Usuário SMTP não configurado"})
		return
	}
	if settings.SMTPPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha SMTP não configurada"})
		return
	}
	if settings.SMTPFromEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email de origem não configurado"})
		return
	}

	// Decrypt password
	password, err := helpers.DecryptIfNeeded(settings.SMTPPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao descriptografar senha"})
		return
	}

	// Set default port if not specified
	port := settings.SMTPPort
	if port == 0 {
		port = 587
	}

	// Test SMTP connection
	addr := settings.SMTPHost + ":" + strconv.Itoa(port)

	// Try to connect and authenticate
	var client *smtp.Client
	var connErr error

	if settings.SMTPUseTLS && port == 465 {
		// Direct TLS connection (port 465)
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         settings.SMTPHost,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Falha na conexão TLS",
				"details": err.Error(),
			})
			return
		}
		defer conn.Close()

		client, connErr = smtp.NewClient(conn, settings.SMTPHost)
	} else {
		// STARTTLS connection (port 587)
		client, connErr = smtp.Dial(addr)
		if connErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Falha na conexão SMTP",
				"details": connErr.Error(),
			})
			return
		}

		if settings.SMTPUseTLS {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: false,
				ServerName:         settings.SMTPHost,
			}
			if err := client.StartTLS(tlsConfig); err != nil {
				client.Close()
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Falha ao iniciar TLS",
					"details": err.Error(),
				})
				return
			}
		}
	}

	if connErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Falha na conexão",
			"details": connErr.Error(),
		})
		return
	}
	defer client.Close()

	// Authenticate
	auth := helpers.GetSMTPAuth(settings.SMTPUsername, password, settings.SMTPHost)
	if err := client.Auth(auth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Falha na autenticação SMTP",
			"details": err.Error(),
		})
		return
	}

	// Test MAIL FROM
	if err := client.Mail(settings.SMTPFromEmail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Falha ao definir remetente",
			"details": err.Error(),
		})
		return
	}

	client.Quit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Conexão SMTP testada com sucesso!",
		"details": gin.H{
			"host": settings.SMTPHost,
			"port": port,
			"tls":  settings.SMTPUseTLS,
			"from": settings.SMTPFromEmail,
		},
	})
}

// DeleteOwnTenant allows an admin to soft delete their own tenant/company
func DeleteOwnTenant(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	userID := c.GetUint("user_id")

	// Use database.DB directly to avoid tenant scope conflicts
	// Get user from database to check role
	var user models.User
	if err := database.DB.Table("public.users").Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Verify user is admin
	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas administradores podem deletar a empresa"})
		return
	}

	// Get tenant to verify it exists
	var tenant models.Tenant
	if err := database.DB.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Empresa não encontrada"})
		return
	}

	// Soft delete the tenant
	if err := database.DB.Table("public.tenants").Where("id = ?", tenantID).Delete(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar empresa"})
		return
	}

	// Deactivate all users of this tenant
	if err := database.DB.Table("public.users").
		Where("tenant_id = ?", tenantID).
		Update("active", false).Error; err != nil {
		// Log but don't fail - tenant is already deleted
		fmt.Printf("Warning: failed to deactivate users for tenant %d: %v\n", tenantID, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Empresa deletada com sucesso",
	})
}
