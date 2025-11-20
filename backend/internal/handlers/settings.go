package handlers

import (
	"drcrwell/backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	// Return the input data as it was just saved successfully
	c.JSON(http.StatusOK, gin.H{"settings": input})
}
