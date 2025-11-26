package models

import (
	"time"

	"gorm.io/gorm"
)

// ConsentTemplate represents a consent term template
type ConsentTemplate struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Template Information
	Title       string `gorm:"not null" json:"title"`
	Type        string `gorm:"not null;index" json:"type"` // treatment, procedure, anesthesia, data_privacy, general
	Content     string `gorm:"type:text;not null" json:"content"` // HTML content
	Version     string `gorm:"not null" json:"version"` // Semantic version (e.g., "1.0.0")
	Description string `gorm:"type:text" json:"description"`

	// Status
	Active      bool   `gorm:"default:true" json:"active"`
	IsDefault   bool   `gorm:"default:false" json:"is_default"` // Default template for this type

	// Relationships
	PatientConsents []PatientConsent `gorm:"foreignKey:TemplateID" json:"consents,omitempty"`
}

// PatientConsent represents a signed consent by a patient
type PatientConsent struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Foreign Keys
	PatientID  uint `gorm:"not null;index" json:"patient_id"`
	TemplateID uint `gorm:"not null;index" json:"template_id"`

	// Template Snapshot (at time of signing)
	TemplateTitle   string `gorm:"not null" json:"template_title"`
	TemplateVersion string `gorm:"not null" json:"template_version"`
	TemplateContent string `gorm:"type:text;not null" json:"template_content"`
	TemplateType    string `gorm:"not null" json:"template_type"`

	// Signature Information
	SignedAt       time.Time `json:"signed_at"`
	SignatureData  string    `gorm:"type:text" json:"signature_data"` // Base64 encoded signature image
	SignatureType  string    `json:"signature_type"` // digital, handwritten, typed
	SignerName     string    `json:"signer_name"` // Name of person who signed
	SignerRelation string    `json:"signer_relation"` // patient, guardian, representative

	// Witness Information (optional)
	WitnessName      string `json:"witness_name"`
	WitnessSignature string `gorm:"type:text" json:"witness_signature"` // Base64 encoded

	// Metadata
	IPAddress  string `json:"ip_address"`
	UserAgent  string `gorm:"type:text" json:"user_agent"`
	SignedByUserID uint `json:"signed_by_user_id"` // User who registered the signature

	// Additional Information
	Notes      string `gorm:"type:text" json:"notes"`

	// Status
	Status     string `gorm:"default:'active'" json:"status"` // active, revoked, expired

	// Relationships
	Patient  Patient         `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Template ConsentTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

// ConsentTemplateType constants
const (
	ConsentTypeTreatment   = "treatment"
	ConsentTypeProcedure   = "procedure"
	ConsentTypeAnesthesia  = "anesthesia"
	ConsentTypeDataPrivacy = "data_privacy"
	ConsentTypeGeneral     = "general"
)

// SignatureType constants
const (
	SignatureTypeDigital     = "digital"
	SignatureTypeHandwritten = "handwritten"
	SignatureTypeTyped       = "typed"
)

// SignerRelation constants
const (
	SignerRelationPatient        = "patient"
	SignerRelationGuardian       = "guardian"
	SignerRelationRepresentative = "representative"
)

// ConsentStatus constants
const (
	ConsentStatusActive  = "active"
	ConsentStatusRevoked = "revoked"
	ConsentStatusExpired = "expired"
)

// TableName specifies the table name for PatientConsent model
func (PatientConsent) TableName() string {
	return "patient_consents"
}
