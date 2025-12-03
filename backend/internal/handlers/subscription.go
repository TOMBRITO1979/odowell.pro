package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	portalsession "github.com/stripe/stripe-go/v76/billingportal/session"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/webhook"
	"gorm.io/gorm"
)

func init() {
	// Initialize Stripe with API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

// Plan info for frontend
type PlanInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	PriceID      string `json:"price_id"`
	PriceMonthly int    `json:"price_monthly"` // in cents
	PatientLimit int    `json:"patient_limit"`
}

// GetPlans returns available subscription plans
func GetPlans(c *gin.Context) {
	plans := []PlanInfo{
		{
			ID:           models.PlanBronze,
			Name:         models.PlanNames[models.PlanBronze],
			PriceID:      os.Getenv("STRIPE_PRICE_BRONZE"),
			PriceMonthly: models.PlanPrices[models.PlanBronze],
			PatientLimit: models.PlanLimits[models.PlanBronze],
		},
		{
			ID:           models.PlanSilver,
			Name:         models.PlanNames[models.PlanSilver],
			PriceID:      os.Getenv("STRIPE_PRICE_SILVER"),
			PriceMonthly: models.PlanPrices[models.PlanSilver],
			PatientLimit: models.PlanLimits[models.PlanSilver],
		},
		{
			ID:           models.PlanGold,
			Name:         models.PlanNames[models.PlanGold],
			PriceID:      os.Getenv("STRIPE_PRICE_GOLD"),
			PriceMonthly: models.PlanPrices[models.PlanGold],
			PatientLimit: models.PlanLimits[models.PlanGold],
		},
	}

	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// GetSubscriptionStatus returns current subscription status for the tenant
func GetSubscriptionStatus(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")

	db := database.GetDB()

	// Get tenant
	var tenant models.Tenant
	if err := db.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get subscription if exists
	var sub models.Subscription
	db.Where("tenant_id = ?", tenantID).First(&sub)

	// Calculate days remaining in trial
	daysRemaining := 0
	if tenant.TrialEndsAt != nil && tenant.SubscriptionStatus == "trialing" {
		remaining := time.Until(*tenant.TrialEndsAt)
		if remaining > 0 {
			daysRemaining = int(remaining.Hours() / 24)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":              tenant.SubscriptionStatus,
		"plan_type":           tenant.PlanType,
		"patient_limit":       tenant.PatientLimit,
		"trial_ends_at":       tenant.TrialEndsAt,
		"days_remaining":      daysRemaining,
		"is_active":           tenant.IsSubscriptionActive(),
		"stripe_customer_id":  sub.StripeCustomerID,
		"cancel_at_period_end": sub.CancelAtPeriodEnd,
		"current_period_end":  sub.CurrentPeriodEnd,
	})
}

// CreateCheckoutRequest represents the request to create a checkout session
type CreateCheckoutRequest struct {
	PlanID string `json:"plan_id" binding:"required"` // bronze, silver, gold
}

// CreateCheckoutSession creates a Stripe Checkout session
func CreateCheckoutSession(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	userEmail := c.GetString("user_email")

	var req CreateCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate plan
	priceID := ""
	switch req.PlanID {
	case models.PlanBronze:
		priceID = os.Getenv("STRIPE_PRICE_BRONZE")
	case models.PlanSilver:
		priceID = os.Getenv("STRIPE_PRICE_SILVER")
	case models.PlanGold:
		priceID = os.Getenv("STRIPE_PRICE_GOLD")
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan"})
		return
	}

	if priceID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Stripe price not configured"})
		return
	}

	db := database.GetDB()

	// Get tenant
	var tenant models.Tenant
	if err := db.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	// Get or create Stripe customer
	var sub models.Subscription
	db.Where("tenant_id = ?", tenantID).First(&sub)

	customerID := sub.StripeCustomerID
	if customerID == "" {
		// Create new Stripe customer
		params := &stripe.CustomerParams{
			Email: stripe.String(userEmail),
			Name:  stripe.String(tenant.Name),
			Metadata: map[string]string{
				"tenant_id": fmt.Sprintf("%d", tenantID),
			},
		}
		cust, err := customer.New(params)
		if err != nil {
			log.Printf("ERROR creating Stripe customer: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
			return
		}
		customerID = cust.ID

		// Save customer ID
		if sub.ID == 0 {
			sub = models.Subscription{
				TenantID:         tenantID,
				StripeCustomerID: customerID,
				Status:           "incomplete",
			}
			db.Create(&sub)
		} else {
			sub.StripeCustomerID = customerID
			db.Save(&sub)
		}
	}

	// Create Checkout Session
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://app.odowell.pro"
	}

	checkoutParams := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(frontendURL + "/subscription/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(frontendURL + "/subscription/cancel"),
		Metadata: map[string]string{
			"tenant_id": fmt.Sprintf("%d", tenantID),
			"plan_id":   req.PlanID,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"tenant_id": fmt.Sprintf("%d", tenantID),
				"plan_id":   req.PlanID,
			},
		},
	}

	sess, err := session.New(checkoutParams)
	if err != nil {
		log.Printf("ERROR creating checkout session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkout_url": sess.URL,
		"session_id":   sess.ID,
	})
}

// CreatePortalSession creates a Stripe Customer Portal session
func CreatePortalSession(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")

	db := database.GetDB()

	// Get subscription
	var sub models.Subscription
	if err := db.Where("tenant_id = ?", tenantID).First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No subscription found"})
		return
	}

	if sub.StripeCustomerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Stripe customer found"})
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://app.odowell.pro"
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(sub.StripeCustomerID),
		ReturnURL: stripe.String(frontendURL + "/subscription"),
	}

	sess, err := portalsession.New(params)
	if err != nil {
		log.Printf("ERROR creating portal session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create portal session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"portal_url": sess.URL,
	})
}

// CancelSubscription cancels the subscription at period end
func CancelSubscription(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")

	db := database.GetDB()

	// Get subscription
	var sub models.Subscription
	if err := db.Where("tenant_id = ?", tenantID).First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No subscription found"})
		return
	}

	if sub.StripeSubscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active subscription"})
		return
	}

	// Cancel at period end (don't cancel immediately)
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	_, err := subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		log.Printf("ERROR canceling subscription: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}

	// Update local record
	sub.CancelAtPeriodEnd = true
	db.Save(&sub)

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription will be canceled at the end of the billing period",
	})
}

// StripeWebhook handles Stripe webhook events
func StripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("ERROR reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	// Verify webhook signature
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)
	if err != nil {
		log.Printf("ERROR verifying webhook signature: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	log.Printf("Stripe webhook received: %s", event.Type)

	db := database.GetDB()

	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			log.Printf("ERROR parsing checkout session: %v", err)
			break
		}
		handleCheckoutCompleted(db, &sess)

	case "customer.subscription.created", "customer.subscription.updated":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("ERROR parsing subscription: %v", err)
			break
		}
		handleSubscriptionUpdated(db, &sub)

	case "customer.subscription.deleted":
		var sub stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			log.Printf("ERROR parsing subscription: %v", err)
			break
		}
		handleSubscriptionDeleted(db, &sub)

	case "invoice.payment_failed":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			log.Printf("ERROR parsing invoice: %v", err)
			break
		}
		handlePaymentFailed(db, &inv)

	case "invoice.paid":
		var inv stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
			log.Printf("ERROR parsing invoice: %v", err)
			break
		}
		handlePaymentSucceeded(db, &inv)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func handleCheckoutCompleted(db *gorm.DB, sess *stripe.CheckoutSession) {
	tenantIDStr := sess.Metadata["tenant_id"]
	planID := sess.Metadata["plan_id"]

	log.Printf("Checkout completed for tenant %s, plan %s", tenantIDStr, planID)

	var tenantID uint
	fmt.Sscanf(tenantIDStr, "%d", &tenantID)

	// Update subscription record
	var sub models.Subscription
	if err := db.Where("tenant_id = ?", tenantID).First(&sub).Error; err != nil {
		sub = models.Subscription{TenantID: tenantID}
	}

	sub.StripeCustomerID = sess.Customer.ID
	sub.StripeSubscriptionID = sess.Subscription.ID
	sub.PlanName = planID
	sub.PatientLimit = models.PlanLimits[planID]
	sub.PriceMonthly = models.PlanPrices[planID]
	sub.Status = "active"

	if sub.ID == 0 {
		db.Create(&sub)
	} else {
		db.Save(&sub)
	}

	// Update tenant
	var tenant models.Tenant
	if err := db.First(&tenant, tenantID).Error; err == nil {
		tenant.SubscriptionStatus = "active"
		tenant.PlanType = planID
		tenant.PatientLimit = models.PlanLimits[planID]
		db.Save(&tenant)
	}
}

func handleSubscriptionUpdated(db *gorm.DB, stripeSub *stripe.Subscription) {
	tenantIDStr := stripeSub.Metadata["tenant_id"]
	planID := stripeSub.Metadata["plan_id"]

	var tenantID uint
	fmt.Sscanf(tenantIDStr, "%d", &tenantID)

	log.Printf("Subscription updated for tenant %d: status=%s", tenantID, stripeSub.Status)

	// Update subscription record
	var sub models.Subscription
	if err := db.Where("stripe_subscription_id = ?", stripeSub.ID).First(&sub).Error; err != nil {
		// Try by tenant_id
		if err := db.Where("tenant_id = ?", tenantID).First(&sub).Error; err != nil {
			log.Printf("Subscription not found for tenant %d", tenantID)
			return
		}
	}

	sub.StripeSubscriptionID = stripeSub.ID
	sub.Status = string(stripeSub.Status)
	sub.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd

	if stripeSub.CurrentPeriodStart > 0 {
		start := time.Unix(stripeSub.CurrentPeriodStart, 0)
		sub.CurrentPeriodStart = &start
	}
	if stripeSub.CurrentPeriodEnd > 0 {
		end := time.Unix(stripeSub.CurrentPeriodEnd, 0)
		sub.CurrentPeriodEnd = &end
	}
	if stripeSub.CanceledAt > 0 {
		canceled := time.Unix(stripeSub.CanceledAt, 0)
		sub.CanceledAt = &canceled
	}

	db.Save(&sub)

	// Update tenant status
	var tenant models.Tenant
	if err := db.First(&tenant, sub.TenantID).Error; err == nil {
		tenant.SubscriptionStatus = string(stripeSub.Status)
		if planID != "" {
			tenant.PlanType = planID
			tenant.PatientLimit = models.PlanLimits[planID]
		}
		db.Save(&tenant)
	}
}

func handleSubscriptionDeleted(db *gorm.DB, stripeSub *stripe.Subscription) {
	log.Printf("Subscription deleted: %s", stripeSub.ID)

	var sub models.Subscription
	if err := db.Where("stripe_subscription_id = ?", stripeSub.ID).First(&sub).Error; err != nil {
		log.Printf("Subscription not found: %s", stripeSub.ID)
		return
	}

	sub.Status = "canceled"
	now := time.Now()
	sub.CanceledAt = &now
	db.Save(&sub)

	// Update tenant
	var tenant models.Tenant
	if err := db.First(&tenant, sub.TenantID).Error; err == nil {
		tenant.SubscriptionStatus = "expired"
		db.Save(&tenant)
	}
}

func handlePaymentFailed(db *gorm.DB, inv *stripe.Invoice) {
	log.Printf("Payment failed for subscription: %s", inv.Subscription.ID)

	var sub models.Subscription
	if err := db.Where("stripe_subscription_id = ?", inv.Subscription.ID).First(&sub).Error; err != nil {
		return
	}

	// Update tenant status
	var tenant models.Tenant
	if err := db.First(&tenant, sub.TenantID).Error; err == nil {
		tenant.SubscriptionStatus = "past_due"
		db.Save(&tenant)
	}
}

func handlePaymentSucceeded(db *gorm.DB, inv *stripe.Invoice) {
	log.Printf("Payment succeeded for subscription: %s", inv.Subscription.ID)

	var sub models.Subscription
	if err := db.Where("stripe_subscription_id = ?", inv.Subscription.ID).First(&sub).Error; err != nil {
		return
	}

	sub.Status = "active"
	db.Save(&sub)

	// Update tenant status
	var tenant models.Tenant
	if err := db.First(&tenant, sub.TenantID).Error; err == nil {
		tenant.SubscriptionStatus = "active"
		db.Save(&tenant)
	}
}
