package models

import (
	"time"

	"gorm.io/gorm"
)

// Campaign represents marketing campaigns (WhatsApp/Email)
type Campaign struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Campaign details
	Name        string `gorm:"not null" json:"name"`
	Type        string `json:"type"` // whatsapp, email, sms
	Subject     string `json:"subject"` // For email
	Message     string `gorm:"type:text;not null" json:"message"`

	// Segmentation
	SegmentType string `json:"segment_type"` // all, tags, custom
	Tags        string `gorm:"type:text" json:"tags"` // comma-separated tags
	Filters     string `gorm:"type:jsonb" json:"filters"` // JSON with custom filters

	// Scheduling
	Status      string     `gorm:"default:'draft'" json:"status"` // draft, scheduled, sending, sent, failed
	ScheduledAt *time.Time `json:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at"`

	// Statistics
	TotalRecipients int `gorm:"default:0" json:"total_recipients"`
	Sent            int `gorm:"default:0" json:"sent"`
	Delivered       int `gorm:"default:0" json:"delivered"`
	Failed          int `gorm:"default:0" json:"failed"`
	Opened          int `gorm:"default:0" json:"opened"`
	Clicked         int `gorm:"default:0" json:"clicked"`

	CreatedByID uint  `gorm:"not null;index" json:"created_by_id"`
	CreatedBy   *User `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`

	// Relationships
	Recipients  []CampaignRecipient `gorm:"foreignKey:CampaignID" json:"recipients,omitempty"`
}

// CampaignRecipient represents individual campaign send status
type CampaignRecipient struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	CampaignID uint      `gorm:"not null;index" json:"campaign_id"`
	Campaign   *Campaign `gorm:"foreignKey:CampaignID" json:"campaign,omitempty"`

	PatientID  uint     `gorm:"not null;index" json:"patient_id"`
	Patient    *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	Status     string     `json:"status"` // pending, sent, delivered, failed, opened, clicked
	SentAt     *time.Time `json:"sent_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
	OpenedAt   *time.Time `json:"opened_at"`
	ClickedAt  *time.Time `json:"clicked_at"`

	ErrorMessage string `json:"error_message"`
}
