package models

import (
	"time"

	"gorm.io/gorm"
)

// Attachment represents files attached to patients (photos, exams, documents)
type Attachment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PatientID uint     `gorm:"not null;index" json:"patient_id"`
	Patient   *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	// File info
	FileName    string `gorm:"not null" json:"file_name"`
	FilePath    string `gorm:"not null" json:"file_path"`
	FileType    string `json:"file_type"` // image, document, exam
	MimeType    string `json:"mime_type"`
	FileSize    int64  `json:"file_size"` // in bytes

	// Classification
	Category    string `json:"category"` // photo, xray, exam, document, other
	Description string `gorm:"type:text" json:"description"`

	UploadedByID uint  `gorm:"not null;index" json:"uploaded_by_id"`
	UploadedBy   *User `gorm:"foreignKey:UploadedByID" json:"uploaded_by,omitempty"`
}
