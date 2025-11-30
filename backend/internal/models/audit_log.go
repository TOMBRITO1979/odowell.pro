package models

import (
	"time"
)

// AuditLog stores user actions for security auditing
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	// Who performed the action
	UserID    uint   `json:"user_id" gorm:"index"`
	UserEmail string `json:"user_email"`
	UserRole  string `json:"user_role"`

	// What action was performed
	Action     string `json:"action" gorm:"index"` // create, update, delete, login, logout, view
	Resource   string `json:"resource"`            // patients, appointments, etc.
	ResourceID uint   `json:"resource_id"`

	// Request details
	Method    string `json:"method"`     // GET, POST, PUT, DELETE
	Path      string `json:"path"`       // /api/patients/123
	IPAddress string `json:"ip_address"` // Client IP
	UserAgent string `json:"user_agent"` // Browser/client info

	// Additional context
	Details string `json:"details" gorm:"type:text"` // JSON with additional info
	Success bool   `json:"success"`                  // Whether action succeeded
}

// TableName specifies the table name for audit logs
// Audit logs are stored in public schema (shared across tenants)
func (AuditLog) TableName() string {
	return "public.audit_logs"
}
