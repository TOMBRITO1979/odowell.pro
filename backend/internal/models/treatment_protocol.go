package models

import (
	"time"

	"gorm.io/gorm"
)

type TreatmentProtocol struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"type:text;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Procedures  string         `gorm:"type:jsonb" json:"procedures"` // Array of procedure objects
	Duration    int            `json:"duration"`                     // Estimated duration in minutes
	Cost        float64        `json:"cost"`                         // Estimated cost
	Active      bool           `gorm:"default:true" json:"active"`
	CreatedBy   uint           `gorm:"not null" json:"created_by"`
}

// TableName specifies the table name for TreatmentProtocol model
func (TreatmentProtocol) TableName() string {
	return "treatment_protocols"
}
