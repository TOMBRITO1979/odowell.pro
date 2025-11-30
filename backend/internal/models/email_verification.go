package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// EmailVerification stores email verification tokens
type EmailVerification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	// The user/tenant being verified
	TenantID uint `json:"tenant_id" gorm:"index"`
	UserID   uint `json:"user_id" gorm:"index"`
	Email    string `json:"email" gorm:"index"`

	// Token info
	Token     string    `json:"token" gorm:"uniqueIndex;size:64"`
	ExpiresAt time.Time `json:"expires_at"`

	// Status
	Verified   bool       `json:"verified"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}

// TableName specifies the table name
func (EmailVerification) TableName() string {
	return "public.email_verifications"
}

// GenerateVerificationToken creates a new secure random token
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsExpired checks if the token has expired
func (ev *EmailVerification) IsExpired() bool {
	return time.Now().After(ev.ExpiresAt)
}

// MarkAsVerified marks the verification as complete
func (ev *EmailVerification) MarkAsVerified() {
	now := time.Now()
	ev.Verified = true
	ev.VerifiedAt = &now
}
