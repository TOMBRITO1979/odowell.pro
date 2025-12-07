package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/account"
)

// StripeCredentialsRequest represents the request to update Stripe credentials
type StripeCredentialsRequest struct {
	StripeSecretKey      string `json:"stripe_secret_key"`
	StripePublishableKey string `json:"stripe_publishable_key"`
	StripeWebhookSecret  string `json:"stripe_webhook_secret"`
}

// GetStripeSettings returns the current Stripe configuration for the tenant
func GetStripeSettings(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")

	var settings models.TenantSettings
	// Use explicit public schema since TenantMiddleware sets search_path to tenant schema
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		// Return empty settings if not found
		c.JSON(http.StatusOK, gin.H{
			"stripe_connected":       false,
			"stripe_publishable_key": "",
			"stripe_account_name":    "",
			"has_secret_key":         false,
			"has_webhook_secret":     false,
		})
		return
	}

	// Don't expose full secret key, just indicate if it's set
	c.JSON(http.StatusOK, gin.H{
		"stripe_connected":       settings.StripeConnected,
		"stripe_publishable_key": settings.StripePublishableKey,
		"stripe_account_name":    settings.StripeAccountName,
		"has_secret_key":         settings.StripeSecretKey != "",
		"has_webhook_secret":     settings.StripeWebhookSecret != "",
	})
}

// UpdateStripeCredentials updates the Stripe credentials for the tenant
func UpdateStripeCredentials(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	log.Printf("UpdateStripeCredentials: Starting for tenant %d", tenantID)

	var req StripeCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("UpdateStripeCredentials: JSON bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("UpdateStripeCredentials: Request received - has secret: %v, has pub: %v, has webhook: %v",
		req.StripeSecretKey != "", req.StripePublishableKey != "", req.StripeWebhookSecret != "")

	// Get or create tenant settings (explicit public schema)
	var settings models.TenantSettings
	result := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings)
	if result.Error != nil {
		settings = models.TenantSettings{TenantID: tenantID}
	}

	// Update only provided fields
	if req.StripeSecretKey != "" {
		settings.StripeSecretKey = req.StripeSecretKey
	}
	if req.StripePublishableKey != "" {
		settings.StripePublishableKey = req.StripePublishableKey
	}
	if req.StripeWebhookSecret != "" {
		settings.StripeWebhookSecret = req.StripeWebhookSecret
	}

	// Validate credentials by making a test API call (use plain key for validation)
	plainSecretKey := settings.StripeSecretKey
	log.Printf("UpdateStripeCredentials: plainSecretKey length: %d", len(plainSecretKey))
	if plainSecretKey != "" {
		// Decrypt if already encrypted (for re-validation)
		decrypted, err := helpers.DecryptIfNeeded(plainSecretKey)
		if err == nil {
			plainSecretKey = decrypted
		}

		stripe.Key = plainSecretKey
		log.Printf("UpdateStripeCredentials: Validating Stripe key...")

		// Try to get account info to validate the key
		acct, err := account.Get()
		if err != nil {
			log.Printf("UpdateStripeCredentials: Stripe validation FAILED: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Stripe secret key: " + err.Error()})
			return
		}
		log.Printf("UpdateStripeCredentials: Stripe validation SUCCESS - account: %s", acct.ID)

		settings.StripeConnected = true
		if acct.BusinessProfile != nil && acct.BusinessProfile.Name != "" {
			settings.StripeAccountName = acct.BusinessProfile.Name
		} else if acct.Settings != nil && acct.Settings.Dashboard != nil {
			settings.StripeAccountName = acct.Settings.Dashboard.DisplayName
		} else {
			settings.StripeAccountName = "Stripe Account"
		}

		// Encrypt the secret key before saving
		encrypted, err := helpers.Encrypt(plainSecretKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt credentials"})
			return
		}
		settings.StripeSecretKey = encrypted
	}

	// Encrypt webhook secret if provided
	if req.StripeWebhookSecret != "" {
		encrypted, err := helpers.Encrypt(req.StripeWebhookSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt webhook secret"})
			return
		}
		settings.StripeWebhookSecret = encrypted
	}

	// Save settings using raw SQL to avoid GORM table name issues
	log.Printf("UpdateStripeCredentials: Saving settings, ID=%d", settings.ID)
	if settings.ID == 0 {
		// Insert new record
		log.Printf("UpdateStripeCredentials: Inserting new record...")
		if err := database.DB.Exec(`
			INSERT INTO public.tenant_settings
			(tenant_id, stripe_secret_key, stripe_publishable_key, stripe_webhook_secret, stripe_connected, stripe_account_name, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			tenantID, settings.StripeSecretKey, settings.StripePublishableKey, settings.StripeWebhookSecret, settings.StripeConnected, settings.StripeAccountName,
		).Error; err != nil {
			log.Printf("UpdateStripeCredentials: INSERT FAILED: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settings: " + err.Error()})
			return
		}
		log.Printf("UpdateStripeCredentials: INSERT SUCCESS")
	} else {
		// Update existing record
		log.Printf("UpdateStripeCredentials: Updating existing record...")
		if err := database.DB.Exec(`
			UPDATE public.tenant_settings
			SET stripe_secret_key = ?, stripe_publishable_key = ?, stripe_webhook_secret = ?, stripe_connected = ?, stripe_account_name = ?, updated_at = NOW()
			WHERE tenant_id = ?`,
			settings.StripeSecretKey, settings.StripePublishableKey, settings.StripeWebhookSecret, settings.StripeConnected, settings.StripeAccountName, tenantID,
		).Error; err != nil {
			log.Printf("UpdateStripeCredentials: UPDATE FAILED: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings: " + err.Error()})
			return
		}
		log.Printf("UpdateStripeCredentials: UPDATE SUCCESS")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                "Stripe credentials updated successfully",
		"stripe_connected":       settings.StripeConnected,
		"stripe_publishable_key": settings.StripePublishableKey,
		"stripe_account_name":    settings.StripeAccountName,
		"has_secret_key":         settings.StripeSecretKey != "",
		"has_webhook_secret":     settings.StripeWebhookSecret != "",
	})
}

// DisconnectStripe removes Stripe credentials for the tenant
func DisconnectStripe(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")

	// Clear Stripe fields using raw SQL
	result := database.DB.Exec(`
		UPDATE public.tenant_settings
		SET stripe_secret_key = '', stripe_publishable_key = '', stripe_webhook_secret = '', stripe_connected = false, stripe_account_name = '', updated_at = NOW()
		WHERE tenant_id = ?`,
		tenantID,
	)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Settings not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Stripe disconnected successfully"})
}

// TestStripeConnection tests if the current Stripe credentials are valid
func TestStripeConnection(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")

	var settings models.TenantSettings
	// Use explicit public schema
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"connected": false,
			"error":     "Settings not found",
		})
		return
	}

	if settings.StripeSecretKey == "" {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
			"error":     "Stripe secret key not configured",
		})
		return
	}

	// Decrypt if encrypted
	decryptedKey, err := helpers.DecryptIfNeeded(settings.StripeSecretKey)
	if err != nil {
		decryptedKey = settings.StripeSecretKey // Fallback
	}
	stripe.Key = decryptedKey

	// Test by getting account info
	acct, err := account.Get()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
			"error":     "Invalid credentials: " + err.Error(),
		})
		return
	}

	accountName := "Stripe Account"
	if acct.BusinessProfile != nil && acct.BusinessProfile.Name != "" {
		accountName = acct.BusinessProfile.Name
	} else if acct.Settings != nil && acct.Settings.Dashboard != nil && acct.Settings.Dashboard.DisplayName != "" {
		accountName = acct.Settings.Dashboard.DisplayName
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":    true,
		"account_id":   acct.ID,
		"account_name": accountName,
		"country":      acct.Country,
		"email":        acct.Email,
	})
}
