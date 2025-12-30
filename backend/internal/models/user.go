package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TenantID  uint   `gorm:"not null;index" json:"tenant_id"`

	Name      string `gorm:"not null" json:"name"`
	Email     string `gorm:"not null;index" json:"email"`
	Password  string `gorm:"not null" json:"-"`
	Phone     string `json:"phone,omitempty"`

	Role         string `gorm:"default:'user'" json:"role"` // admin, dentist, receptionist, user, patient
	Active       bool   `gorm:"default:true" json:"active"`
	IsSuperAdmin bool   `gorm:"default:false" json:"is_super_admin"` // Platform-level admin

	// Patient portal - links user to patient record
	PatientID *uint `gorm:"index" json:"patient_id,omitempty"` // Only for role="patient"

	// Professional info (for dentists)
	CRO       string `json:"cro,omitempty"`
	Specialty string `json:"specialty,omitempty"`

	// Profile picture
	ProfilePicture string `json:"profile_picture,omitempty"`

	// UI preferences
	HideSidebar bool `gorm:"default:false" json:"hide_sidebar"`

	// Two-Factor Authentication (2FA) - TOTP
	TwoFactorEnabled     bool   `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret      string `json:"-"` // Encrypted TOTP secret, never exposed in JSON
	TwoFactorBackupCodes string `json:"-"` // Encrypted backup codes (JSON array), never exposed
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "public.users"
}

// HashPassword generates bcrypt hash of the password
func (u *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

// CheckPassword compares password with hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
