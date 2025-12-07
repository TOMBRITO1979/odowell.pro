package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/subscription"
)

// getDecryptedStripeKey retrieves and decrypts the Stripe secret key for a tenant
func getDecryptedStripeKey(tenantID uint) (string, error) {
	var settings models.TenantSettings
	// Use explicit public schema since TenantMiddleware sets search_path to tenant schema
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		return "", err
	}

	decrypted, err := helpers.DecryptIfNeeded(settings.StripeSecretKey)
	if err != nil {
		log.Printf("Warning: Could not decrypt Stripe key for tenant %d: %v", tenantID, err)
		return settings.StripeSecretKey, nil // Return as-is if decryption fails (backwards compatibility)
	}
	return decrypted, nil
}

// ListPatientSubscriptions returns all patient subscriptions for the tenant
func ListPatientSubscriptions(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var subscriptions []models.PatientSubscription
	query := db.Preload("Patient").Order("created_at DESC")

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by patient if provided
	if patientID := c.Query("patient_id"); patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}

	if err := query.Find(&subscriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// GetPatientSubscription returns a single subscription with payment history
func GetPatientSubscription(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	id := c.Param("id")

	var sub models.PatientSubscription
	if err := db.Preload("Patient").First(&sub, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Get last 12 payments
	var payments []models.PatientSubscriptionPayment
	db.Where("patient_subscription_id = ?", sub.ID).
		Order("created_at DESC").
		Limit(12).
		Find(&payments)

	c.JSON(http.StatusOK, gin.H{
		"subscription": sub,
		"payments":     payments,
	})
}

// CreatePatientSubscriptionRequest represents the request to create a subscription
type CreatePatientSubscriptionRequest struct {
	PatientID     uint   `json:"patient_id" binding:"required"`
	StripePriceID string `json:"stripe_price_id" binding:"required"`
	Notes         string `json:"notes"`
	SuccessURL    string `json:"success_url"`
	CancelURL     string `json:"cancel_url"`
}

// CreatePatientSubscription creates a new subscription for a patient
func CreatePatientSubscription(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")
	userID := c.GetUint("user_id")

	var req CreatePatientSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant settings for Stripe credentials (explicit public schema)
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant settings not found"})
		return
	}

	if settings.StripeSecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stripe not configured. Please add your Stripe credentials in Settings."})
		return
	}

	// Get patient info
	var patient models.Patient
	if err := db.First(&patient, req.PatientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
		return
	}

	// Set Stripe API key for this tenant (decrypt if encrypted)
	decryptedKey, err := helpers.DecryptIfNeeded(settings.StripeSecretKey)
	if err != nil {
		log.Printf("Warning: Could not decrypt Stripe key: %v", err)
		decryptedKey = settings.StripeSecretKey // Fallback to raw value
	}
	stripe.Key = decryptedKey

	// Get price details from Stripe
	priceObj, err := price.Get(req.StripePriceID, &stripe.PriceParams{
		Expand: []*string{stripe.String("product")},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Stripe price: " + err.Error()})
		return
	}

	// Check if patient already has a Stripe customer ID for this price's product
	var existingSub models.PatientSubscription
	hasExisting := db.Where("patient_id = ? AND stripe_product_id = ? AND status IN ?",
		req.PatientID, priceObj.Product.ID, []string{"active", "trialing", "past_due"}).
		First(&existingSub).Error == nil

	if hasExisting {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Patient already has an active subscription for this product"})
		return
	}

	// Create or get Stripe customer
	var stripeCustomerID string
	var existingSubWithCustomer models.PatientSubscription
	if db.Where("patient_id = ? AND stripe_customer_id != ''", req.PatientID).
		First(&existingSubWithCustomer).Error == nil {
		stripeCustomerID = existingSubWithCustomer.StripeCustomerID
	} else {
		// Create new Stripe customer
		customerParams := &stripe.CustomerParams{
			Email: stripe.String(patient.Email),
			Name:  stripe.String(patient.Name),
			Phone: stripe.String(patient.Phone),
			Metadata: map[string]string{
				"patient_id": strconv.Itoa(int(patient.ID)),
				"tenant_id":  strconv.Itoa(int(tenantID)),
			},
		}
		cust, err := customer.New(customerParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe customer: " + err.Error()})
			return
		}
		stripeCustomerID = cust.ID
	}

	// Determine URLs
	successURL := req.SuccessURL
	if successURL == "" {
		successURL = fmt.Sprintf("https://app.odowell.pro/plans?success=true&patient_id=%d", patient.ID)
	}
	cancelURL := req.CancelURL
	if cancelURL == "" {
		cancelURL = fmt.Sprintf("https://app.odowell.pro/plans?canceled=true&patient_id=%d", patient.ID)
	}

	// Create Stripe Checkout Session
	checkoutParams := &stripe.CheckoutSessionParams{
		Customer: stripe.String(stripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.StripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"patient_id": strconv.Itoa(int(patient.ID)),
			"tenant_id":  strconv.Itoa(int(tenantID)),
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"patient_id": strconv.Itoa(int(patient.ID)),
				"tenant_id":  strconv.Itoa(int(tenantID)),
			},
		},
	}

	// Add customer email if not already set
	if patient.Email != "" {
		checkoutParams.CustomerEmail = stripe.String(patient.Email)
	}

	sess, err := session.New(checkoutParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session: " + err.Error()})
		return
	}

	// Get product name
	productName := "Unknown Product"
	if priceObj.Product != nil {
		productName = priceObj.Product.Name
	}

	// Calculate checkout expiration (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create local subscription record
	sub := models.PatientSubscription{
		PatientID:         req.PatientID,
		StripeCustomerID:  stripeCustomerID,
		StripePriceID:     req.StripePriceID,
		StripeProductID:   priceObj.Product.ID,
		ProductName:       productName,
		PriceAmount:       priceObj.UnitAmount,
		PriceCurrency:     string(priceObj.Currency),
		Interval:          string(priceObj.Recurring.Interval),
		IntervalCount:     int(priceObj.Recurring.IntervalCount),
		Status:            models.SubscriptionStatusPending,
		CheckoutSessionID: sess.ID,
		CheckoutURL:       sess.URL,
		CheckoutExpiresAt: &expiresAt,
		Notes:             req.Notes,
		CreatedBy:         userID,
	}

	if err := db.Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription record"})
		return
	}

	// Load patient for response
	db.Preload("Patient").First(&sub, sub.ID)

	c.JSON(http.StatusCreated, gin.H{
		"subscription": sub,
		"checkout_url": sess.URL,
	})
}

// CancelPatientSubscription cancels a subscription
func CancelPatientSubscription(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")
	id := c.Param("id")

	var sub models.PatientSubscription
	if err := db.First(&sub, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Get tenant settings for Stripe credentials (explicit public schema)
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant settings not found"})
		return
	}

	if sub.StripeSubscriptionID != "" && settings.StripeSecretKey != "" {
		decryptedKey, _ := helpers.DecryptIfNeeded(settings.StripeSecretKey)
		stripe.Key = decryptedKey

		// Cancel at period end (graceful cancellation)
		_, err := subscription.Update(sub.StripeSubscriptionID, &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel Stripe subscription: " + err.Error()})
			return
		}

		sub.CancelAtPeriodEnd = true
	} else {
		// No Stripe subscription yet, just mark as canceled locally
		now := time.Now()
		sub.Status = models.SubscriptionStatusCanceled
		sub.CanceledAt = &now
	}

	if err := db.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// CancelPatientSubscriptionImmediately cancels a subscription immediately
func CancelPatientSubscriptionImmediately(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")
	id := c.Param("id")

	var sub models.PatientSubscription
	if err := db.First(&sub, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Get tenant settings for Stripe credentials (explicit public schema)
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant settings not found"})
		return
	}

	if sub.StripeSubscriptionID != "" && settings.StripeSecretKey != "" {
		decryptedKey, _ := helpers.DecryptIfNeeded(settings.StripeSecretKey)
		stripe.Key = decryptedKey

		// Cancel immediately
		_, err := subscription.Cancel(sub.StripeSubscriptionID, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel Stripe subscription: " + err.Error()})
			return
		}
	}

	now := time.Now()
	sub.Status = models.SubscriptionStatusCanceled
	sub.CanceledAt = &now

	if err := db.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// GetStripeProducts returns available Stripe products/prices for the tenant
func GetStripeProducts(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")

	// Get tenant settings for Stripe credentials (explicit public schema)
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant settings not found"})
		return
	}

	if settings.StripeSecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stripe not configured"})
		return
	}

	decryptedKey, _ := helpers.DecryptIfNeeded(settings.StripeSecretKey)
	stripe.Key = decryptedKey

	// List active products
	productParams := &stripe.ProductListParams{
		Active: stripe.Bool(true),
	}
	productParams.Limit = stripe.Int64(100)

	productsIter := product.List(productParams)
	var products []map[string]interface{}

	for productsIter.Next() {
		prod := productsIter.Product()

		// Get prices for this product
		priceParams := &stripe.PriceListParams{
			Product: stripe.String(prod.ID),
			Active:  stripe.Bool(true),
		}
		priceParams.Limit = stripe.Int64(10)

		pricesIter := price.List(priceParams)
		var prices []map[string]interface{}

		for pricesIter.Next() {
			p := pricesIter.Price()
			if p.Recurring != nil {
				prices = append(prices, map[string]interface{}{
					"id":             p.ID,
					"unit_amount":    p.UnitAmount,
					"currency":       p.Currency,
					"interval":       p.Recurring.Interval,
					"interval_count": p.Recurring.IntervalCount,
				})
			}
		}

		if len(prices) > 0 {
			products = append(products, map[string]interface{}{
				"id":          prod.ID,
				"name":        prod.Name,
				"description": prod.Description,
				"images":      prod.Images,
				"prices":      prices,
			})
		}
	}

	if err := productsIter.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Stripe products: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

// RefreshSubscriptionStatus syncs a subscription status with Stripe
func RefreshSubscriptionStatus(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")
	id := c.Param("id")

	var sub models.PatientSubscription
	if err := db.First(&sub, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	if sub.StripeSubscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No Stripe subscription linked"})
		return
	}

	// Get tenant settings for Stripe credentials (explicit public schema)
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant settings not found"})
		return
	}

	decryptedKey, _ := helpers.DecryptIfNeeded(settings.StripeSecretKey)
	stripe.Key = decryptedKey

	// Get subscription from Stripe
	stripeSub, err := subscription.Get(sub.StripeSubscriptionID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Stripe subscription: " + err.Error()})
		return
	}

	// Update local record
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

	if err := db.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	db.Preload("Patient").First(&sub, sub.ID)
	c.JSON(http.StatusOK, sub)
}

// GetSubscriptionPayments returns payment history for a subscription
func GetSubscriptionPayments(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	id := c.Param("id")

	// Verify subscription exists
	var sub models.PatientSubscription
	if err := db.First(&sub, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Get payments
	limit := 12
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var payments []models.PatientSubscriptionPayment
	db.Where("patient_subscription_id = ?", sub.ID).
		Order("created_at DESC").
		Limit(limit).
		Find(&payments)

	c.JSON(http.StatusOK, payments)
}

// ResendCheckoutLink generates a new checkout session for a pending subscription
func ResendCheckoutLink(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")
	id := c.Param("id")

	var sub models.PatientSubscription
	if err := db.Preload("Patient").First(&sub, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	if sub.Status != models.SubscriptionStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only resend link for pending subscriptions"})
		return
	}

	// Get tenant settings for Stripe credentials (explicit public schema)
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant settings not found"})
		return
	}

	decryptedKey, _ := helpers.DecryptIfNeeded(settings.StripeSecretKey)
	stripe.Key = decryptedKey

	// Create new checkout session
	successURL := fmt.Sprintf("https://app.odowell.pro/plans?success=true&patient_id=%d", sub.PatientID)
	cancelURL := fmt.Sprintf("https://app.odowell.pro/plans?canceled=true&patient_id=%d", sub.PatientID)

	checkoutParams := &stripe.CheckoutSessionParams{
		Customer: stripe.String(sub.StripeCustomerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(sub.StripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"patient_id":      strconv.Itoa(int(sub.PatientID)),
			"tenant_id":       strconv.Itoa(int(tenantID)),
			"subscription_id": strconv.Itoa(int(sub.ID)),
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"patient_id":      strconv.Itoa(int(sub.PatientID)),
				"tenant_id":       strconv.Itoa(int(tenantID)),
				"subscription_id": strconv.Itoa(int(sub.ID)),
			},
		},
	}

	sess, err := session.New(checkoutParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session: " + err.Error()})
		return
	}

	// Update subscription with new checkout info
	expiresAt := time.Now().Add(24 * time.Hour)
	sub.CheckoutSessionID = sess.ID
	sub.CheckoutURL = sess.URL
	sub.CheckoutExpiresAt = &expiresAt

	if err := db.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": sub,
		"checkout_url": sess.URL,
	})
}
