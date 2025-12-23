package models

import (
	"time"

	"gorm.io/gorm"
)

// Lead represents a potential patient/contact from WhatsApp or other sources
type Lead struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Contact Information
	Name      string     `gorm:"not null" json:"name"`
	Phone     string     `gorm:"not null;index" json:"phone"` // Primary identifier for WhatsApp lookup
	Email     string     `json:"email"`
	BirthDate *time.Time `json:"birth_date,omitempty"` // Data de nascimento do lead

	// Lead Source and Tracking
	Source        string `gorm:"default:'whatsapp'" json:"source"` // whatsapp, website, referral, instagram, facebook, other
	ContactReason string `gorm:"type:text" json:"contact_reason"`  // Summary of conversation/reason for contact

	// Status Management
	Status string `gorm:"default:'new'" json:"status"` // new, contacted, qualified, converted, lost

	// Conversion Tracking
	ConvertedToPatientID *uint      `json:"converted_to_patient_id,omitempty"` // When converted to patient
	ConvertedAt          *time.Time `json:"converted_at,omitempty"`            // When conversion happened

	// Additional Information
	Notes     string `gorm:"type:text" json:"notes"`
	CreatedBy uint   `gorm:"not null" json:"created_by"` // User who created the lead
}

// TableName specifies the table name for Lead model
func (Lead) TableName() string {
	return "leads"
}
