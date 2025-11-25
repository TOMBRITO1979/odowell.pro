package models

import (
	"time"

	"gorm.io/gorm"
)

// MedicalRecord represents a patient's medical/dental record
type MedicalRecord struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PatientID uint     `gorm:"not null;index" json:"patient_id"`
	Patient   *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	DentistID uint     `gorm:"not null;index" json:"dentist_id"`
	Dentist   *User    `gorm:"foreignKey:DentistID" json:"dentist,omitempty"`

	AppointmentID *uint `json:"appointment_id"`

	// Record type
	Type          string `json:"type"` // anamnesis, treatment, procedure, prescription, certificate

	// Odontogram (JSON with tooth status)
	Odontogram    *string `gorm:"type:jsonb" json:"odontogram,omitempty"`

	// Treatment plan
	Diagnosis     string `gorm:"type:text" json:"diagnosis"`
	TreatmentPlan string `gorm:"type:text" json:"treatment_plan"`

	// Procedure details
	ProcedureDone string `gorm:"type:text" json:"procedure_done"`
	Materials     string `gorm:"type:text" json:"materials"`

	// Prescription
	Prescription  string `gorm:"type:text" json:"prescription"`

	// Certificate/Attestation
	Certificate   string `gorm:"type:text" json:"certificate"`

	// Evolution
	Evolution     string `gorm:"type:text" json:"evolution"`

	// Allergies
	Arlegis       string `gorm:"type:text" json:"arlegis"`

	Notes         string `gorm:"type:text" json:"notes"`
}
