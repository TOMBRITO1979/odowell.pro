package models

import (
	"time"

	"gorm.io/gorm"
)

// Budget represents a treatment budget/quote
type Budget struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PatientID uint     `gorm:"not null;index" json:"patient_id"`
	Patient   *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	DentistID uint  `gorm:"not null;index" json:"dentist_id"`
	Dentist   *User `gorm:"foreignKey:DentistID" json:"dentist,omitempty"`

	// Budget details
	Description string  `gorm:"type:text" json:"description"`
	TotalValue  float64 `gorm:"not null" json:"total_value"`

	// Items (JSON array)
	Items       *string `gorm:"type:jsonb" json:"items,omitempty"`

	// Status
	Status      string `gorm:"default:'pending'" json:"status"` // pending, approved, rejected, expired

	ValidUntil  *time.Time `json:"valid_until"`

	Notes       string `gorm:"type:text" json:"notes"`

	// Relationships
	Payments    []Payment `gorm:"foreignKey:BudgetID" json:"payments,omitempty"`
}

// Payment represents a financial transaction
type Payment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	BudgetID  *uint   `json:"budget_id"`
	Budget    *Budget `gorm:"foreignKey:BudgetID" json:"budget,omitempty"`

	PatientID uint     `gorm:"not null;index" json:"patient_id"`
	Patient   *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	// Payment details
	Type        string  `json:"type"` // income, expense
	Category    string  `json:"category"` // treatment, material, salary, rent, etc
	Description string  `gorm:"type:text" json:"description"`
	Amount      float64 `gorm:"not null" json:"amount"`

	// Payment method
	PaymentMethod string `json:"payment_method"` // cash, credit_card, debit_card, pix, transfer, insurance

	// Installments
	IsInstallment     bool `gorm:"default:false" json:"is_installment"`
	InstallmentNumber int  `json:"installment_number"`
	TotalInstallments int  `json:"total_installments"`

	// Status
	Status      string     `gorm:"default:'pending'" json:"status"` // pending, paid, overdue, cancelled
	DueDate     *time.Time `json:"due_date"`
	PaidDate    *time.Time `json:"paid_date"`

	// Insurance
	IsInsurance bool   `gorm:"default:false" json:"is_insurance"`
	InsuranceName string `json:"insurance_name"`

	Notes       string `gorm:"type:text" json:"notes"`
}

// Commission represents professional commissions
type Commission struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	DentistID uint    `gorm:"not null;index" json:"dentist_id"`
	Dentist   *User   `gorm:"foreignKey:DentistID" json:"dentist,omitempty"`

	PaymentID uint     `gorm:"not null;index" json:"payment_id"`
	Payment   *Payment `gorm:"foreignKey:PaymentID" json:"payment,omitempty"`

	Percentage float64 `json:"percentage"`
	Amount     float64 `json:"amount"`

	Status     string     `gorm:"default:'pending'" json:"status"` // pending, paid
	PaidDate   *time.Time `json:"paid_date"`
}
