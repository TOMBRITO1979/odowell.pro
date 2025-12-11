package models

import (
	"time"

	"gorm.io/gorm"
)

// UserCertificate stores encrypted digital certificates for users
type UserCertificate struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// User association
	UserID uint  `gorm:"not null;index" json:"user_id"`
	User   *User `gorm:"foreignKey:UserID" json:"user,omitempty"`

	// Certificate metadata (extracted from certificate)
	Name             string     `json:"name"`                         // Friendly name for the certificate
	SubjectCN        string     `json:"subject_cn"`                   // Common Name from certificate
	SubjectCPF       string     `json:"subject_cpf,omitempty"`        // CPF extracted from ICP-Brasil cert
	IssuerCN         string     `json:"issuer_cn"`                    // Certificate issuer name
	SerialNumber     string     `gorm:"uniqueIndex" json:"serial_number"` // Unique serial number
	Thumbprint       string     `gorm:"index" json:"thumbprint"`      // SHA-1 fingerprint for identification
	NotBefore        time.Time  `json:"not_before"`                   // Certificate validity start
	NotAfter         time.Time  `json:"not_after"`                    // Certificate validity end

	// Encrypted certificate storage
	EncryptedPFX     []byte     `json:"-"`                            // AES-256-GCM encrypted PFX data
	EncryptionSalt   []byte     `json:"-"`                            // Salt for password-based key derivation

	// Status
	Active           bool       `gorm:"default:false" json:"active"`  // Is this the active certificate for the user
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`       // Last time used for signing

	// Validation
	IsValid          bool       `json:"is_valid"`                     // Is the certificate currently valid
	ValidationError  string     `json:"validation_error,omitempty"`   // Error message if validation failed
}

// TableName specifies the table name for GORM (public schema)
func (UserCertificate) TableName() string {
	return "public.user_certificates"
}

// IsExpired checks if the certificate has expired
func (c *UserCertificate) IsExpired() bool {
	return time.Now().After(c.NotAfter)
}

// IsNotYetValid checks if the certificate is not yet valid
func (c *UserCertificate) IsNotYetValid() bool {
	return time.Now().Before(c.NotBefore)
}

// DaysUntilExpiry returns the number of days until the certificate expires
func (c *UserCertificate) DaysUntilExpiry() int {
	duration := time.Until(c.NotAfter)
	return int(duration.Hours() / 24)
}
