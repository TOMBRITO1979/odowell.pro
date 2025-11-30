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
	DBSchema   string `gorm:"unique;not null" json:"db_schema"`

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

	// Subscription (for future SaaS features)
	PlanType   string `gorm:"default:'basic'" json:"plan_type"` // basic, professional, premium
	ExpiresAt  *time.Time `json:"expires_at"`

	// WhatsApp/External API Integration
	APIKey     string `gorm:"unique" json:"api_key,omitempty"` // API key for external integrations (WhatsApp, AI bots)
	APIKeyActive bool `gorm:"default:false" json:"api_key_active"` // Whether API key is enabled
}

// TableName specifies the table name for Tenant model
func (Tenant) TableName() string {
	return "public.tenants"
}
