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
