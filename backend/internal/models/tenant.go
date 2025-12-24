package models

import (
	"time"

	"gorm.io/gorm"
)

// Tenant represents a dental clinic (multi-tenant isolation)
type Tenant struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Tenant identification
	Name       string `gorm:"not null" json:"name"`
	Subdomain  string `gorm:"unique;not null" json:"subdomain"`
	DBSchema   string `gorm:"unique;not null" json:"-"` // Hidden from API responses for security

	// Contact info
	Email      string `json:"email"`
	Phone      string `json:"phone"`

	// Address
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	ZipCode    string `json:"zip_code"`

	// Status
	Active        bool `gorm:"default:true" json:"active"`
	EmailVerified bool `gorm:"default:false" json:"email_verified"`

	// Subscription
	PlanType           string     `gorm:"default:'basic'" json:"plan_type"` // basic, professional, premium, gold
	ExpiresAt          *time.Time `json:"expires_at"`
	SubscriptionStatus string     `gorm:"default:'trialing'" json:"subscription_status"` // trialing, active, expired, cancelled
	TrialEndsAt        *time.Time `json:"trial_ends_at"`
	PatientLimit       int        `gorm:"default:1000" json:"patient_limit"` // Max patients allowed

	// WhatsApp/External API Integration
	APIKey          string     `gorm:"unique" json:"api_key,omitempty"`    // API key for external integrations (WhatsApp, AI bots)
	APIKeyActive    bool       `gorm:"default:false" json:"api_key_active"` // Whether API key is enabled
	APIKeyLastUsed  *time.Time `json:"api_key_last_used,omitempty"`        // Last time API key was used
	APIKeyExpiresAt *time.Time `json:"api_key_expires_at,omitempty"`       // API key expiration date (nil = never expires)
	APIKeyCreatedAt *time.Time `json:"api_key_created_at,omitempty"`       // When the current API key was generated

	// Embed token for external forms
	EmbedToken string `json:"embed_token,omitempty"`
}

// APIKeyMaxAgeDays is the maximum age of an API key before forced rotation
const APIKeyMaxAgeDays = 90

// IsAPIKeyExpired checks if the API key has expired (explicit expiration date)
func (t *Tenant) IsAPIKeyExpired() bool {
	if t.APIKeyExpiresAt == nil {
		return false // No expiration set
	}
	return time.Now().After(*t.APIKeyExpiresAt)
}

// NeedsAPIKeyRotation checks if the API key is older than 90 days and needs rotation
func (t *Tenant) NeedsAPIKeyRotation() bool {
	if t.APIKeyCreatedAt == nil {
		return false // No creation date, can't determine age
	}
	maxAge := time.Duration(APIKeyMaxAgeDays) * 24 * time.Hour
	return time.Since(*t.APIKeyCreatedAt) > maxAge
}

// DaysUntilAPIKeyRotation returns days until forced rotation (negative if overdue)
func (t *Tenant) DaysUntilAPIKeyRotation() int {
	if t.APIKeyCreatedAt == nil {
		return APIKeyMaxAgeDays // Assume new key
	}
	maxAge := time.Duration(APIKeyMaxAgeDays) * 24 * time.Hour
	rotationDate := t.APIKeyCreatedAt.Add(maxAge)
	daysLeft := int(time.Until(rotationDate).Hours() / 24)
	return daysLeft
}

// IsSubscriptionActive checks if tenant has an active subscription
func (t *Tenant) IsSubscriptionActive() bool {
	if t.SubscriptionStatus == "active" {
		return true
	}
	if t.SubscriptionStatus == "trialing" && t.TrialEndsAt != nil {
		return time.Now().Before(*t.TrialEndsAt)
	}
	return false
}

// TableName specifies the table name for Tenant model
func (Tenant) TableName() string {
	return "public.tenants"
}
