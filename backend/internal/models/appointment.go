package models

import (
	"time"

	"gorm.io/gorm"
)

// Appointment represents a scheduled appointment
type Appointment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PatientID uint      `gorm:"not null;index" json:"patient_id"`
	Patient   *Patient  `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	DentistID uint      `gorm:"not null;index" json:"dentist_id"`
	Dentist   *User     `gorm:"foreignKey:DentistID" json:"dentist,omitempty"`

	// Appointment details
	StartTime   time.Time `gorm:"not null;index" json:"start_time"`
	EndTime     time.Time `gorm:"not null" json:"end_time"`

	Type        string    `json:"type"` // consultation, treatment, emergency, return
	Procedure   string    `json:"procedure"`

	// Status
	Status      string    `gorm:"default:'scheduled'" json:"status"` // scheduled, confirmed, in_progress, completed, cancelled, no_show

	// Confirmation
	Confirmed   bool      `gorm:"default:false" json:"confirmed"`
	ConfirmedAt *time.Time `json:"confirmed_at"`

	// Reminder sent
	ReminderSent bool     `gorm:"default:false" json:"reminder_sent"`

	Notes       string    `gorm:"type:text" json:"notes"`

	// Room
	Room        string    `json:"room"`

	// Recurrence
	IsRecurring bool      `gorm:"default:false" json:"is_recurring"`
	RecurrenceRule string `json:"recurrence_rule,omitempty"` // JSON with recurrence config
}
