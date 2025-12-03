package models

import (
	"time"

	"gorm.io/gorm"
)

// Subscription represents a Stripe subscription for a tenant
type Subscription struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Tenant association
	TenantID uint `gorm:"not null;uniqueIndex" json:"tenant_id"`

	// Stripe IDs
	StripeCustomerID     string `gorm:"index" json:"stripe_customer_id"`
	StripeSubscriptionID string `gorm:"index" json:"stripe_subscription_id"`
	StripePriceID        string `json:"stripe_price_id"`

	// Plan details
	PlanName       string `json:"plan_name"`        // bronze, silver, gold
	PatientLimit   int    `json:"patient_limit"`    // 1000, 2500, 5000
	PriceMonthly   int    `json:"price_monthly"`    // in cents: 9900, 15900, 21900

	// Subscription status
	// trialing, active, canceled, past_due, unpaid, incomplete
	Status string `gorm:"default:'trialing'" json:"status"`

	// Trial period
	TrialStart *time.Time `json:"trial_start"`
	TrialEnd   *time.Time `json:"trial_end"`

	// Billing period
	CurrentPeriodStart *time.Time `json:"current_period_start"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end"`

	// Cancellation
	CanceledAt         *time.Time `json:"canceled_at"`
	CancelAtPeriodEnd  bool       `gorm:"default:false" json:"cancel_at_period_end"`
}

// TableName specifies the table name for Subscription model
func (Subscription) TableName() string {
	return "public.subscriptions"
}

// Plan constants
const (
	PlanBronze = "bronze"
	PlanSilver = "silver"
	PlanGold   = "gold"
)

// Plan limits
var PlanLimits = map[string]int{
	PlanBronze: 1000,
	PlanSilver: 2500,
	PlanGold:   5000,
}

// Plan prices in cents
var PlanPrices = map[string]int{
	PlanBronze: 9900,  // $99.00
	PlanSilver: 15900, // $159.00
	PlanGold:   21900, // $219.00
}

// Plan display names
var PlanNames = map[string]string{
	PlanBronze: "Bronze",
	PlanSilver: "Prata",
	PlanGold:   "Ouro",
}

// IsTrialActive checks if the trial period is still active
func (s *Subscription) IsTrialActive() bool {
	if s.Status != "trialing" {
		return false
	}
	if s.TrialEnd == nil {
		return false
	}
	return time.Now().Before(*s.TrialEnd)
}

// IsActive checks if the subscription is active (trialing or active)
func (s *Subscription) IsActive() bool {
	return s.Status == "active" || s.IsTrialActive()
}

// DaysRemainingInTrial returns the number of days remaining in trial
func (s *Subscription) DaysRemainingInTrial() int {
	if s.TrialEnd == nil {
		return 0
	}
	remaining := time.Until(*s.TrialEnd)
	if remaining < 0 {
		return 0
	}
	return int(remaining.Hours() / 24)
}
