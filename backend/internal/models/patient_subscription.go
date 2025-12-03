package models

import (
	"time"

	"gorm.io/gorm"
)

// PatientSubscription represents a recurring payment plan for a patient
type PatientSubscription struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Patient reference
	PatientID uint `gorm:"not null;index" json:"patient_id"`

	// Stripe subscription data
	StripeSubscriptionID string `gorm:"index" json:"stripe_subscription_id"`
	StripeCustomerID     string `gorm:"index" json:"stripe_customer_id"`
	StripePriceID        string `json:"stripe_price_id"`
	StripeProductID      string `json:"stripe_product_id"`

	// Subscription details
	ProductName   string  `json:"product_name"`
	PriceAmount   int64   `json:"price_amount"`   // Amount in cents
	PriceCurrency string  `json:"price_currency"` // BRL, USD, etc
	Interval      string  `json:"interval"`       // month, year, week
	IntervalCount int     `json:"interval_count"` // e.g., 1 for monthly, 3 for quarterly

	// Status
	Status            string     `gorm:"default:'pending'" json:"status"` // pending, active, past_due, canceled, unpaid
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
	CanceledAt        *time.Time `json:"canceled_at,omitempty"`
	CancelAtPeriodEnd bool       `gorm:"default:false" json:"cancel_at_period_end"`

	// Checkout
	CheckoutSessionID string     `json:"checkout_session_id,omitempty"`
	CheckoutURL       string     `json:"checkout_url,omitempty"`
	CheckoutExpiresAt *time.Time `json:"checkout_expires_at,omitempty"`

	// Metadata
	Notes     string `json:"notes,omitempty"`
	CreatedBy uint   `json:"created_by"` // User who created the subscription

	// Relations (for JSON response)
	Patient *Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
}

// PatientSubscriptionPayment represents a payment event for a subscription
type PatientSubscriptionPayment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Subscription reference
	PatientSubscriptionID uint `gorm:"not null;index" json:"patient_subscription_id"`

	// Stripe invoice data
	StripeInvoiceID   string `gorm:"index" json:"stripe_invoice_id"`
	StripePaymentIntentID string `json:"stripe_payment_intent_id,omitempty"`
	StripeChargeID    string `json:"stripe_charge_id,omitempty"`

	// Payment details
	Amount        int64  `json:"amount"`         // Amount in cents
	Currency      string `json:"currency"`       // BRL, USD, etc
	Status        string `json:"status"`         // paid, open, void, uncollectible
	PaymentMethod string `json:"payment_method"` // card, boleto, pix

	// Period
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`

	// Payment dates
	DueDate  *time.Time `json:"due_date,omitempty"`
	PaidAt   *time.Time `json:"paid_at,omitempty"`

	// Invoice
	InvoiceURL     string `json:"invoice_url,omitempty"`
	InvoicePdfURL  string `json:"invoice_pdf_url,omitempty"`
	ReceiptURL     string `json:"receipt_url,omitempty"`

	// Failure info
	FailureMessage string `json:"failure_message,omitempty"`
	FailureCode    string `json:"failure_code,omitempty"`
}

// TableName specifies the table name for PatientSubscription
func (PatientSubscription) TableName() string {
	return "patient_subscriptions"
}

// TableName specifies the table name for PatientSubscriptionPayment
func (PatientSubscriptionPayment) TableName() string {
	return "patient_subscription_payments"
}

// Status constants
const (
	SubscriptionStatusPending   = "pending"
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPastDue   = "past_due"
	SubscriptionStatusCanceled  = "canceled"
	SubscriptionStatusUnpaid    = "unpaid"
	SubscriptionStatusTrialing  = "trialing"
	SubscriptionStatusIncomplete = "incomplete"
)

// Payment status constants
const (
	PaymentStatusPaid          = "paid"
	PaymentStatusOpen          = "open"
	PaymentStatusVoid          = "void"
	PaymentStatusUncollectible = "uncollectible"
	PaymentStatusDraft         = "draft"
)
