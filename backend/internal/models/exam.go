package models

import (
	"time"

	"gorm.io/gorm"
)

// Exam represents a patient exam file stored in S3
type Exam struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Patient relationship (foreign key only)
	PatientID uint `gorm:"not null;index" json:"patient_id"`

	// Exam details
	Name        string     `gorm:"not null" json:"name"`        // Nome do exame
	Description string     `gorm:"type:text" json:"description"` // Descrição/observações
	ExamType    string     `json:"exam_type"`                   // Tipo: raio-x, tomografia, foto, laudo, etc
	ExamDate    *time.Time `json:"exam_date"`                   // Data que o exame foi realizado

	// File storage (S3)
	FileURL  string `gorm:"not null" json:"file_url"`  // URL completa do arquivo no S3
	S3Key    string `gorm:"not null" json:"s3_key"`    // Chave do arquivo no S3 (drcrwell/[cpf]/filename)
	FileName string `gorm:"not null" json:"file_name"` // Nome original do arquivo
	FileType string `json:"file_type"`                 // MIME type (image/jpeg, application/pdf, etc)
	FileSize int64  `json:"file_size"`                 // Tamanho do arquivo em bytes

	// Upload info (foreign key only)
	UploadedByID uint `gorm:"not null" json:"uploaded_by_id"`

	// Additional notes
	Notes       string `gorm:"type:text" json:"notes"`
}
