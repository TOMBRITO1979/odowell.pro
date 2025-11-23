package models

import (
	"time"

	"gorm.io/gorm"
)

type WaitingList struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	PatientID      uint           `gorm:"not null;index" json:"patient_id"`
	DentistID      *uint          `gorm:"index" json:"dentist_id,omitempty"` // Optional - any dentist if null
	Procedure      string         `gorm:"type:text" json:"procedure"`                    // Procedure needed
	PreferredDates string         `gorm:"type:jsonb" json:"preferred_dates,omitempty"`   // JSONB array of date ranges
	Priority       string         `gorm:"type:text;default:'normal'" json:"priority"`    // normal, urgent
	Status         string         `gorm:"type:text;default:'waiting'" json:"status"`     // waiting, contacted, scheduled, cancelled
	ContactedAt    *time.Time     `json:"contacted_at,omitempty"`                        // When patient was contacted
	ContactedBy    *uint          `json:"contacted_by,omitempty"`                        // User who contacted
	ScheduledAt    *time.Time     `json:"scheduled_at,omitempty"`                        // When scheduled
	AppointmentID  *uint          `json:"appointment_id,omitempty"`                      // Reference to scheduled appointment
	Notes          string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy      uint           `gorm:"not null" json:"created_by"`                    // User who added to waiting list
}

// TableName specifies the table name for WaitingList model
func (WaitingList) TableName() string {
	return "waiting_lists"
}
