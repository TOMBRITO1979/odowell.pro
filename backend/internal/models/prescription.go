package models

import (
	"time"

	"gorm.io/gorm"
)

// Prescription represents medical prescriptions and reports (receituário)
type Prescription struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Patient and Dentist
	PatientID uint     `gorm:"not null;index" json:"patient_id"`
	Patient   *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	DentistID uint  `gorm:"not null;index" json:"dentist_id"`
	Dentist   *User `gorm:"foreignKey:DentistID" json:"dentist,omitempty"`

	// Type of document
	Type string `json:"type"` // prescription, medical_report, certificate, referral

	// Prescription content
	Title       string `json:"title"`        // e.g., "Receita Médica", "Atestado Odontológico"
	Medications string `gorm:"type:text" json:"medications"` // For prescriptions
	Content     string `gorm:"type:text;not null" json:"content"` // Main content/instructions
	Diagnosis   string `gorm:"type:text" json:"diagnosis"`

	// Additional info
	ValidUntil *time.Time `json:"valid_until"` // Prescription expiration
	Notes      string     `gorm:"type:text" json:"notes"`

	// Clinic info (cached for document consistency)
	ClinicName    string `json:"clinic_name"`
	ClinicAddress string `json:"clinic_address"`
	ClinicPhone   string `json:"clinic_phone"`

	// Dentist info (cached for document consistency)
	DentistName string `json:"dentist_name"`
	DentistCRO  string `json:"dentist_cro"`

	// Status
	Status     string `gorm:"default:'draft'" json:"status"` // draft, issued, cancelled
	IssuedAt   *time.Time `json:"issued_at"`
	PrintedAt  *time.Time `json:"printed_at"`
	PrintCount int    `gorm:"default:0" json:"print_count"`
}
