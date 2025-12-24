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
	Status      string `gorm:"default:'pending'" json:"status"` // pending, approved, rejected, expired, cancelled

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

	PatientID *uint    `gorm:"index" json:"patient_id"`
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
	Status       string     `gorm:"default:'pending'" json:"status"` // pending, paid, overdue, cancelled, refunded
	DueDate      *time.Time `json:"due_date"`
	PaidDate     *time.Time `json:"paid_date"`
	RefundedDate *time.Time `json:"refunded_date"`
	RefundReason string     `gorm:"type:text" json:"refund_reason"`

	// Insurance
	IsInsurance   bool   `gorm:"default:false" json:"is_insurance"`
	InsuranceName string `json:"insurance_name"`

	// Recurrence (for expenses)
	IsRecurring    bool `gorm:"default:false" json:"is_recurring"`
	RecurrenceDays int  `gorm:"default:0" json:"recurrence_days"` // 7, 15, 30, 180, 360

	Notes string `gorm:"type:text" json:"notes"`
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

// Treatment represents an approved budget being treated/paid
type Treatment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Origin budget
	BudgetID uint    `gorm:"not null;uniqueIndex" json:"budget_id"`
	Budget   *Budget `gorm:"foreignKey:BudgetID" json:"budget,omitempty"`

	PatientID uint     `gorm:"not null;index" json:"patient_id"`
	Patient   *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`

	DentistID uint  `gorm:"not null;index" json:"dentist_id"`
	Dentist   *User `gorm:"foreignKey:DentistID" json:"dentist,omitempty"`

	// Treatment details
	Description string  `gorm:"type:text" json:"description"`
	TotalValue  float64 `gorm:"not null" json:"total_value"`
	PaidValue   float64 `gorm:"default:0" json:"paid_value"`

	// Installment plan
	TotalInstallments int     `gorm:"default:1" json:"total_installments"`
	InstallmentValue  float64 `json:"installment_value"`

	// Status: in_progress, completed, cancelled
	Status string `gorm:"default:'in_progress'" json:"status"`

	// Dates
	StartDate      time.Time  `json:"start_date"`
	ExpectedEndDate *time.Time `json:"expected_end_date"`
	CompletedDate  *time.Time `json:"completed_date"`

	Notes string `gorm:"type:text" json:"notes"`

	// Relationships
	TreatmentPayments []TreatmentPayment `gorm:"foreignKey:TreatmentID" json:"treatment_payments,omitempty"`
}

// TreatmentPayment represents a payment entry for a treatment
type TreatmentPayment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TreatmentID uint       `gorm:"not null;index" json:"treatment_id"`
	Treatment   *Treatment `gorm:"foreignKey:TreatmentID" json:"treatment,omitempty"`

	// Payment details
	Amount        float64 `gorm:"not null" json:"amount"`
	PaymentMethod string  `json:"payment_method"` // cash, credit_card, debit_card, pix, transfer

	// Installment info
	InstallmentNumber int `json:"installment_number"`

	// Receipt - unique per active (non-deleted) record, handled by partial index in DB
	ReceiptNumber string `gorm:"index" json:"receipt_number"`

	// Status
	Status   string     `gorm:"default:'paid'" json:"status"` // paid, cancelled, refunded
	PaidDate time.Time  `json:"paid_date"`

	// Who received
	ReceivedByID uint  `gorm:"not null" json:"received_by_id"`
	ReceivedBy   *User `gorm:"foreignKey:ReceivedByID" json:"received_by,omitempty"`

	Notes string `gorm:"type:text" json:"notes"`
}

// Treatment status constants
const (
	TreatmentStatusInProgress = "in_progress"
	TreatmentStatusCompleted  = "completed"
	TreatmentStatusCancelled  = "cancelled"
)

// TreatmentPayment status constants
const (
	TreatmentPaymentStatusPaid      = "paid"
	TreatmentPaymentStatusCancelled = "cancelled"
	TreatmentPaymentStatusRefunded  = "refunded"
)
