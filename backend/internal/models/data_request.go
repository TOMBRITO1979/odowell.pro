package models

import (
	"time"

	"gorm.io/gorm"
)

// DataRequestType defines the type of LGPD data request
type DataRequestType string

const (
	DataRequestTypeAccess     DataRequestType = "access"      // Request to access personal data
	DataRequestTypePortability DataRequestType = "portability" // Request to export data
	DataRequestTypeCorrection DataRequestType = "correction"  // Request to correct data
	DataRequestTypeDeletion   DataRequestType = "deletion"    // Request to delete data
	DataRequestTypeRevocation DataRequestType = "revocation"  // Revoke consent
)

// DataRequestStatus defines the status of a data request
type DataRequestStatus string

const (
	DataRequestStatusPending    DataRequestStatus = "pending"     // Waiting to be processed
	DataRequestStatusInProgress DataRequestStatus = "in_progress" // Being processed
	DataRequestStatusCompleted  DataRequestStatus = "completed"   // Request fulfilled
	DataRequestStatusRejected   DataRequestStatus = "rejected"    // Request denied (with reason)
)

// DataRequest represents an LGPD data subject request
type DataRequest struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Patient info
	PatientID   uint   `json:"patient_id" gorm:"index"`
	PatientName string `json:"patient_name"`
	PatientCPF  string `json:"patient_cpf"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`

	// Request details
	Type        DataRequestType   `json:"type" gorm:"type:varchar(20);index"`
	Status      DataRequestStatus `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	Description string            `json:"description" gorm:"type:text"`

	// Processing info
	ProcessedBy   *uint      `json:"processed_by"`
	ProcessedAt   *time.Time `json:"processed_at"`
	ResponseNotes string     `json:"response_notes" gorm:"type:text"`
	RejectionReason string   `json:"rejection_reason" gorm:"type:text"`

	// For data export requests
	ExportFileURL string `json:"export_file_url"`

	// Audit trail
	RequestIP    string `json:"request_ip"`
	RequestAgent string `json:"request_agent"`

	// OTP Verification (LGPD identity verification)
	OTPCode       string     `json:"-" gorm:"type:varchar(6)"` // Hidden from JSON
	OTPExpiresAt  *time.Time `json:"otp_expires_at"`
	OTPAttempts   int        `json:"otp_attempts" gorm:"default:0"`
	OTPVerified   bool       `json:"otp_verified" gorm:"default:false"`
	OTPVerifiedAt *time.Time `json:"otp_verified_at"`

	// SLA Tracking (15 days per LGPD)
	Deadline *time.Time `json:"deadline"` // Auto-calculated: created_at + 15 days
}

// TableName returns the table name for DataRequest
func (DataRequest) TableName() string {
	return "data_requests"
}

// Patient relation for preloading
type DataRequestWithPatient struct {
	DataRequest
	Patient *Patient `json:"patient" gorm:"foreignKey:PatientID"`
}

// SLA Constants
const (
	LGPDDeadlineDays = 15 // LGPD requires response within 15 days
	SLAWarningDays   = 3  // Alert when less than 3 days remaining
)

// CalculateDeadline returns the LGPD deadline (15 days from creation)
func (d *DataRequest) CalculateDeadline() time.Time {
	return d.CreatedAt.AddDate(0, 0, LGPDDeadlineDays)
}

// DaysRemaining returns the number of days until deadline
func (d *DataRequest) DaysRemaining() int {
	if d.Deadline == nil {
		deadline := d.CalculateDeadline()
		d.Deadline = &deadline
	}
	remaining := time.Until(*d.Deadline).Hours() / 24
	if remaining < 0 {
		return 0
	}
	return int(remaining)
}

// IsOverdue returns true if the request is past its deadline
func (d *DataRequest) IsOverdue() bool {
	if d.Deadline == nil {
		deadline := d.CalculateDeadline()
		d.Deadline = &deadline
	}
	return time.Now().After(*d.Deadline) && d.Status != DataRequestStatusCompleted && d.Status != DataRequestStatusRejected
}

// IsNearDeadline returns true if less than 3 days remaining
func (d *DataRequest) IsNearDeadline() bool {
	return d.DaysRemaining() <= SLAWarningDays && !d.IsOverdue() && d.Status != DataRequestStatusCompleted && d.Status != DataRequestStatusRejected
}

// RequiresVerification returns true if this request type requires OTP verification
func (d *DataRequest) RequiresVerification() bool {
	// Deletion and portability require identity verification for security
	return d.Type == DataRequestTypeDeletion || d.Type == DataRequestTypePortability
}
