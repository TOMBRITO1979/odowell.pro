package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// PasswordReset stores password reset tokens
type PasswordReset struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	// The user requesting reset
	UserID uint   `json:"user_id" gorm:"index"`
	Email  string `json:"email" gorm:"index"`

	// Token info
	Token     string    `json:"token" gorm:"uniqueIndex;size:64"`
	ExpiresAt time.Time `json:"expires_at"`

	// Status
	Used   bool       `json:"used"`
	UsedAt *time.Time `json:"used_at,omitempty"`
}

// TableName specifies the table name
func (PasswordReset) TableName() string {
	return "public.password_resets"
}

// GenerateResetToken creates a new secure random token
func GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsExpired checks if the token has expired
func (pr *PasswordReset) IsExpired() bool {
	return time.Now().After(pr.ExpiresAt)
}

// MarkAsUsed marks the reset token as used
func (pr *PasswordReset) MarkAsUsed() {
	now := time.Now()
	pr.Used = true
	pr.UsedAt = &now
}
