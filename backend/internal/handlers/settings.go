package handlers

import (
	"crypto/rand"
	"drcrwell/backend/internal/models"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

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
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint("tenant_id")

	var settings models.TenantSettings
	result := db.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Return empty settings if not found
			c.JSON(http.StatusOK, gin.H{
				"settings": models.TenantSettings{
					TenantID: tenantID,
					SMTPPort: 587,
					SMTPUseTLS: true,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func UpdateTenantSettings(c *gin.Context) {
	var input models.TenantSettings
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint("tenant_id")

	// Check if settings exist
	var count int64
	db.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).Count(&count)

	if count == 0 {
		// Create new settings
		input.TenantID = tenantID
		if err := db.Table("public.tenant_settings").Create(&input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settings", "details": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"settings": input})
		return
	}

	// Update using direct SQL to avoid GORM complexity
	result := db.Exec(`
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
		input.SMTPHost, input.SMTPPort, input.SMTPUsername,
		func() string {
			if input.SMTPPassword != "" {
				return input.SMTPPassword
			}
			return ""
		}(),
		input.SMTPFromName, input.SMTPFromEmail, input.SMTPUseTLS,
		input.WhatsAppAPIKey, input.WhatsAppNumber, input.SMSAPIKey, input.SMSProvider,
		tenantID,
	)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings", "details": result.Error.Error()})
		return
	}

	// Also update the tenant name to match the clinic name
	if input.ClinicName != "" {
		db.Exec("UPDATE public.tenants SET name = ?, updated_at = NOW() WHERE id = ?", input.ClinicName, tenantID)
	}

	// Return the input data as it was just saved successfully
	c.JSON(http.StatusOK, gin.H{"settings": input})
}

// GetEmbedToken returns the current embed token status
func GetEmbedToken(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
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
	db := c.MustGet("db").(*gorm.DB)
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
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint("tenant_id")

	if err := db.Exec("UPDATE public.tenants SET embed_token = '', updated_at = NOW() WHERE id = ?", tenantID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token revoked successfully"})
}
