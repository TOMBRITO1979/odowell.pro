package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

// HandleStripeWebhook processes Stripe webhook events for tenant subscriptions
func HandleStripeWebhook(c *gin.Context) {
	// Get tenant ID from URL parameter
	tenantIDStr := c.Param("tenant_id")
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	// Get tenant settings for webhook secret
	var settings models.TenantSettings
	if err := database.DB.Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		log.Printf("Webhook: Tenant settings not found for tenant %d", tenantID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant settings not found"})
		return
	}

	// Read request body
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Webhook: Error reading body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading body"})
		return
	}

	// Verify webhook signature if secret is configured
	var event stripe.Event
	if settings.StripeWebhookSecret != "" {
		// Decrypt webhook secret if encrypted
		webhookSecret, decryptErr := helpers.DecryptIfNeeded(settings.StripeWebhookSecret)
		if decryptErr != nil {
			log.Printf("Webhook: Could not decrypt webhook secret: %v", decryptErr)
			webhookSecret = settings.StripeWebhookSecret // Fallback to raw value
		}

		sigHeader := c.GetHeader("Stripe-Signature")
		event, err = webhook.ConstructEvent(payload, sigHeader, webhookSecret)
		if err != nil {
			log.Printf("Webhook: Signature verification failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook signature verification failed"})
			return
		}
	} else {
		// Parse without verification (not recommended for production)
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("Webhook: Error parsing event: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing event"})
			return
		}
	}

	log.Printf("Webhook: Received event %s for tenant %d", event.Type, tenantID)

	// Get tenant DB
	db := database.GetTenantDBByID(uint(tenantID))
	if db == nil {
		log.Printf("Webhook: Could not get tenant DB for tenant %d", tenantID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not get tenant DB"})
		return
	}

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		handlePatientCheckoutCompleted(db, event, uint(tenantID))

	case "customer.subscription.created":
		handlePatientSubscriptionCreated(db, event)

	case "customer.subscription.updated":
		handlePatientSubscriptionUpdated(db, event)

	case "customer.subscription.deleted":
		handlePatientSubscriptionDeleted(db, event)

	case "invoice.paid":
		handlePatientInvoicePaid(db, event)

	case "invoice.payment_failed":
		handlePatientInvoicePaymentFailed(db, event)

	case "invoice.created":
		handlePatientInvoiceCreated(db, event)

	default:
		log.Printf("Webhook: Unhandled event type: %s", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func handlePatientCheckoutCompleted(db *database.TenantDB, event stripe.Event, tenantID uint) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		log.Printf("Webhook: Error parsing checkout session: %v", err)
		return
	}

	log.Printf("Webhook: Checkout completed - Session ID: %s, Subscription: %s", session.ID, session.Subscription.ID)

	// Find subscription by checkout session ID
	var sub models.PatientSubscription
	if err := db.DB.Where("checkout_session_id = ?", session.ID).First(&sub).Error; err != nil {
		log.Printf("Webhook: Subscription not found for checkout session %s", session.ID)
		return
	}

	// Update subscription with Stripe subscription ID
	sub.StripeSubscriptionID = session.Subscription.ID
	sub.Status = models.SubscriptionStatusActive

	if err := db.DB.Save(&sub).Error; err != nil {
		log.Printf("Webhook: Error updating subscription: %v", err)
		return
	}

	log.Printf("Webhook: Subscription %d activated with Stripe subscription %s", sub.ID, session.Subscription.ID)
}

func handlePatientSubscriptionCreated(db *database.TenantDB, event stripe.Event) {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		log.Printf("Webhook: Error parsing subscription: %v", err)
		return
	}

	// Try to find existing subscription by Stripe subscription ID
	var sub models.PatientSubscription
	if err := db.DB.Where("stripe_subscription_id = ?", stripeSub.ID).First(&sub).Error; err != nil {
		// Subscription not found locally, might have been created through checkout
		log.Printf("Webhook: Subscription created in Stripe but not found locally: %s", stripeSub.ID)
		return
	}

	updateSubscriptionFromStripe(db, &sub, &stripeSub)
}

func handlePatientSubscriptionUpdated(db *database.TenantDB, event stripe.Event) {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		log.Printf("Webhook: Error parsing subscription: %v", err)
		return
	}

	var sub models.PatientSubscription
	if err := db.DB.Where("stripe_subscription_id = ?", stripeSub.ID).First(&sub).Error; err != nil {
		log.Printf("Webhook: Subscription not found: %s", stripeSub.ID)
		return
	}

	updateSubscriptionFromStripe(db, &sub, &stripeSub)
}

func handlePatientSubscriptionDeleted(db *database.TenantDB, event stripe.Event) {
	var stripeSub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSub); err != nil {
		log.Printf("Webhook: Error parsing subscription: %v", err)
		return
	}

	var sub models.PatientSubscription
	if err := db.DB.Where("stripe_subscription_id = ?", stripeSub.ID).First(&sub).Error; err != nil {
		log.Printf("Webhook: Subscription not found for deletion: %s", stripeSub.ID)
		return
	}

	now := time.Now()
	sub.Status = models.SubscriptionStatusCanceled
	sub.CanceledAt = &now

	if err := db.DB.Save(&sub).Error; err != nil {
		log.Printf("Webhook: Error updating subscription: %v", err)
		return
	}

	log.Printf("Webhook: Subscription %d marked as canceled", sub.ID)
}

func handlePatientInvoicePaid(db *database.TenantDB, event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Webhook: Error parsing invoice: %v", err)
		return
	}

	// Only process subscription invoices
	if invoice.Subscription == nil {
		return
	}

	var sub models.PatientSubscription
	if err := db.DB.Where("stripe_subscription_id = ?", invoice.Subscription.ID).First(&sub).Error; err != nil {
		log.Printf("Webhook: Subscription not found for invoice: %s", invoice.Subscription.ID)
		return
	}

	// Create or update payment record
	var payment models.PatientSubscriptionPayment
	result := db.DB.Where("stripe_invoice_id = ?", invoice.ID).First(&payment)

	periodStart := time.Unix(invoice.PeriodStart, 0)
	periodEnd := time.Unix(invoice.PeriodEnd, 0)
	paidAt := time.Now()

	if result.Error != nil {
		// Create new payment record
		payment = models.PatientSubscriptionPayment{
			PatientSubscriptionID: sub.ID,
			StripeInvoiceID:       invoice.ID,
			Amount:                invoice.AmountPaid,
			Currency:              string(invoice.Currency),
			Status:                models.PaymentStatusPaid,
			PeriodStart:           periodStart,
			PeriodEnd:             periodEnd,
			PaidAt:                &paidAt,
		}

		if invoice.PaymentIntent != nil {
			payment.StripePaymentIntentID = invoice.PaymentIntent.ID
		}
		if invoice.Charge != nil {
			payment.StripeChargeID = invoice.Charge.ID
		}
		if invoice.HostedInvoiceURL != "" {
			payment.InvoiceURL = invoice.HostedInvoiceURL
		}
		if invoice.InvoicePDF != "" {
			payment.InvoicePdfURL = invoice.InvoicePDF
		}

		if err := db.DB.Create(&payment).Error; err != nil {
			log.Printf("Webhook: Error creating payment record: %v", err)
			return
		}
	} else {
		// Update existing payment
		payment.Status = models.PaymentStatusPaid
		payment.Amount = invoice.AmountPaid
		payment.PaidAt = &paidAt
		if invoice.HostedInvoiceURL != "" {
			payment.InvoiceURL = invoice.HostedInvoiceURL
		}
		if invoice.InvoicePDF != "" {
			payment.InvoicePdfURL = invoice.InvoicePDF
		}

		if err := db.DB.Save(&payment).Error; err != nil {
			log.Printf("Webhook: Error updating payment record: %v", err)
			return
		}
	}

	log.Printf("Webhook: Payment recorded for subscription %d, amount: %d", sub.ID, invoice.AmountPaid)
}

func handlePatientInvoicePaymentFailed(db *database.TenantDB, event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Webhook: Error parsing invoice: %v", err)
		return
	}

	if invoice.Subscription == nil {
		return
	}

	var sub models.PatientSubscription
	if err := db.DB.Where("stripe_subscription_id = ?", invoice.Subscription.ID).First(&sub).Error; err != nil {
		log.Printf("Webhook: Subscription not found for failed invoice: %s", invoice.Subscription.ID)
		return
	}

	// Update subscription status
	sub.Status = models.SubscriptionStatusPastDue
	if err := db.DB.Save(&sub).Error; err != nil {
		log.Printf("Webhook: Error updating subscription status: %v", err)
	}

	// Create payment record with failure info
	periodStart := time.Unix(invoice.PeriodStart, 0)
	periodEnd := time.Unix(invoice.PeriodEnd, 0)

	var failureMessage, failureCode string
	if invoice.LastFinalizationError != nil {
		failureMessage = invoice.LastFinalizationError.Msg
		failureCode = string(invoice.LastFinalizationError.Code)
	}

	payment := models.PatientSubscriptionPayment{
		PatientSubscriptionID: sub.ID,
		StripeInvoiceID:       invoice.ID,
		Amount:                invoice.AmountDue,
		Currency:              string(invoice.Currency),
		Status:                models.PaymentStatusOpen,
		PeriodStart:           periodStart,
		PeriodEnd:             periodEnd,
		FailureMessage:        failureMessage,
		FailureCode:           failureCode,
	}

	if invoice.HostedInvoiceURL != "" {
		payment.InvoiceURL = invoice.HostedInvoiceURL
	}

	// Check if payment already exists
	var existing models.PatientSubscriptionPayment
	if db.DB.Where("stripe_invoice_id = ?", invoice.ID).First(&existing).Error == nil {
		existing.Status = models.PaymentStatusOpen
		existing.FailureMessage = failureMessage
		existing.FailureCode = failureCode
		db.DB.Save(&existing)
	} else {
		db.DB.Create(&payment)
	}

	log.Printf("Webhook: Payment failed for subscription %d, invoice: %s", sub.ID, invoice.ID)
}

func handlePatientInvoiceCreated(db *database.TenantDB, event stripe.Event) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("Webhook: Error parsing invoice: %v", err)
		return
	}

	if invoice.Subscription == nil {
		return
	}

	var sub models.PatientSubscription
	if err := db.DB.Where("stripe_subscription_id = ?", invoice.Subscription.ID).First(&sub).Error; err != nil {
		return // Subscription not found, might be from different source
	}

	// Check if payment record already exists
	var existing models.PatientSubscriptionPayment
	if db.DB.Where("stripe_invoice_id = ?", invoice.ID).First(&existing).Error == nil {
		return // Already exists
	}

	periodStart := time.Unix(invoice.PeriodStart, 0)
	periodEnd := time.Unix(invoice.PeriodEnd, 0)

	payment := models.PatientSubscriptionPayment{
		PatientSubscriptionID: sub.ID,
		StripeInvoiceID:       invoice.ID,
		Amount:                invoice.AmountDue,
		Currency:              string(invoice.Currency),
		Status:                models.PaymentStatusDraft,
		PeriodStart:           periodStart,
		PeriodEnd:             periodEnd,
	}

	if invoice.HostedInvoiceURL != "" {
		payment.InvoiceURL = invoice.HostedInvoiceURL
	}

	if err := db.DB.Create(&payment).Error; err != nil {
		log.Printf("Webhook: Error creating pending payment record: %v", err)
		return
	}

	log.Printf("Webhook: Pending payment created for subscription %d", sub.ID)
}

func updateSubscriptionFromStripe(db *database.TenantDB, sub *models.PatientSubscription, stripeSub *stripe.Subscription) {
	sub.Status = string(stripeSub.Status)
	sub.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd

	if stripeSub.CurrentPeriodStart > 0 {
		t := time.Unix(stripeSub.CurrentPeriodStart, 0)
		sub.CurrentPeriodStart = &t
	}
	if stripeSub.CurrentPeriodEnd > 0 {
		t := time.Unix(stripeSub.CurrentPeriodEnd, 0)
		sub.CurrentPeriodEnd = &t
	}
	if stripeSub.CanceledAt > 0 {
		t := time.Unix(stripeSub.CanceledAt, 0)
		sub.CanceledAt = &t
	}

	if err := db.DB.Save(sub).Error; err != nil {
		log.Printf("Webhook: Error updating subscription: %v", err)
		return
	}

	log.Printf("Webhook: Subscription %d updated to status %s", sub.ID, sub.Status)
}

// GenerateWebhookURL generates the webhook URL for a tenant
func GenerateWebhookURL(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")

	// Generate webhook URL
	webhookURL := fmt.Sprintf("https://api.odowell.pro/api/v1/webhooks/stripe/%d", tenantID)

	c.JSON(http.StatusOK, gin.H{
		"webhook_url": webhookURL,
		"events": []string{
			"checkout.session.completed",
			"customer.subscription.created",
			"customer.subscription.updated",
			"customer.subscription.deleted",
			"invoice.paid",
			"invoice.payment_failed",
			"invoice.created",
		},
	})
}
