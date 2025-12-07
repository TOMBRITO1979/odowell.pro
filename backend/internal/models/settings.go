package models

import (
	"time"

	"gorm.io/gorm"
)

type TenantSettings struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TenantID uint `gorm:"not null;uniqueIndex" json:"tenant_id"`

	// Clinic Information
	ClinicName    string `json:"clinic_name"`
	ClinicCNPJ    string `json:"clinic_cnpj"`
	ClinicAddress string `json:"clinic_address"`
	ClinicCity    string `json:"clinic_city"`
	ClinicState   string `json:"clinic_state"`
	ClinicZip     string `json:"clinic_zip"`
	ClinicPhone   string `json:"clinic_phone"`
	ClinicEmail   string `json:"clinic_email"`

	// Working Hours
	WorkingHoursStart          string `json:"working_hours_start"` // Format: "HH:MM"
	WorkingHoursEnd            string `json:"working_hours_end"`   // Format: "HH:MM"
	DefaultAppointmentDuration int    `json:"default_appointment_duration" gorm:"default:30"`

	// Payment Methods
	PaymentCashEnabled         bool `json:"payment_cash_enabled" gorm:"default:true"`
	PaymentCreditCardEnabled   bool `json:"payment_credit_card_enabled" gorm:"default:true"`
	PaymentDebitCardEnabled    bool `json:"payment_debit_card_enabled" gorm:"default:true"`
	PaymentPixEnabled          bool `json:"payment_pix_enabled" gorm:"default:true"`
	PaymentTransferEnabled     bool `json:"payment_transfer_enabled" gorm:"default:false"`
	PaymentInsuranceEnabled    bool `json:"payment_insurance_enabled" gorm:"default:false"`

	// SMTP Settings for sending campaign emails
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"` // Should be encrypted in production
	SMTPFromName string `json:"smtp_from_name"`
	SMTPFromEmail string `json:"smtp_from_email"`
	SMTPUseTLS   bool   `json:"smtp_use_tls"`

	// WhatsApp Settings (future use)
	WhatsAppAPIKey string `json:"whatsapp_api_key,omitempty"`
	WhatsAppNumber string `json:"whatsapp_number,omitempty"`

	// SMS Settings (future use)
	SMSAPIKey   string `json:"sms_api_key,omitempty"`
	SMSProvider string `json:"sms_provider,omitempty"`

	// Stripe Settings (for patient subscriptions)
	StripeSecretKey      string `json:"stripe_secret_key,omitempty"`      // Encrypted in DB
	StripePublishableKey string `json:"stripe_publishable_key,omitempty"`
	StripeWebhookSecret  string `json:"stripe_webhook_secret,omitempty"`  // Encrypted in DB
	StripeConnected      bool   `json:"stripe_connected" gorm:"default:false"`
	StripeAccountName    string `json:"stripe_account_name,omitempty"`
}

// TableName specifies the table name for TenantSettings model
// Settings are stored in public schema, not tenant schemas
func (TenantSettings) TableName() string {
	return "public.tenant_settings"
}
