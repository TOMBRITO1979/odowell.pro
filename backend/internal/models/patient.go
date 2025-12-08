package models

import (
	"time"

	"gorm.io/gorm"
)

// Patient represents a dental clinic patient
type Patient struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Personal Information
	Name      string     `gorm:"not null" json:"name"`
	CPF       string     `gorm:"index" json:"cpf"`
	RG        string     `json:"rg"`
	BirthDate *time.Time `json:"birth_date"`
	Gender    string     `json:"gender"` // M, F, Other

	// Contact
	Email     string `gorm:"index" json:"email"`
	Phone     string `gorm:"index" json:"phone"`
	CellPhone string `gorm:"index" json:"cell_phone"`

	// Address
	Address   string `json:"address"`
	Number    string `json:"number"`
	Complement string `json:"complement"`
	District  string `json:"district"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`

	// Medical Information
	Allergies        string `gorm:"type:text" json:"allergies"`
	Medications      string `gorm:"type:text" json:"medications"`
	SystemicDiseases string `gorm:"type:text" json:"systemic_diseases"`
	BloodType        string `json:"blood_type"`

	// Insurance
	HasInsurance     bool   `gorm:"default:false" json:"has_insurance"`
	InsuranceName    string `json:"insurance_name"`
	InsuranceNumber  string `json:"insurance_number"`

	// Tags for segmentation
	Tags             string `gorm:"type:text" json:"tags"` // comma-separated tags

	// Status
	Active           bool   `gorm:"default:true" json:"active"`

	// Additional notes
	Notes            string `gorm:"type:text" json:"notes"`

	// Relationships
	Appointments     []Appointment `gorm:"foreignKey:PatientID" json:"appointments,omitempty"`
	MedicalRecords   []MedicalRecord `gorm:"foreignKey:PatientID" json:"medical_records,omitempty"`
	Attachments      []Attachment `gorm:"foreignKey:PatientID" json:"attachments,omitempty"`
}
