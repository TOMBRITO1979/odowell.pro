package middleware

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SubscriptionMiddleware checks if tenant has an active subscription or valid trial
func SubscriptionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		db := database.GetDB()

		var tenant models.Tenant
		if err := db.First(&tenant, tenantID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			c.Abort()
			return
		}

		// Check subscription status
		if !tenant.IsSubscriptionActive() {
			// Calculate days since trial expired
			daysExpired := 0
			if tenant.TrialEndsAt != nil {
				elapsed := time.Since(*tenant.TrialEndsAt)
				if elapsed > 0 {
					daysExpired = int(elapsed.Hours() / 24)
				}
			}

			c.JSON(http.StatusPaymentRequired, gin.H{
				"error":             "subscription_required",
				"message":           "Sua assinatura expirou. Por favor, assine para continuar usando o sistema.",
				"subscription_status": tenant.SubscriptionStatus,
				"trial_expired":     true,
				"days_expired":      daysExpired,
			})
			c.Abort()
			return
		}

		// Add subscription info to context
		c.Set("subscription_status", tenant.SubscriptionStatus)
		c.Set("plan_type", tenant.PlanType)
		c.Set("patient_limit", tenant.PatientLimit)

		// Add trial info if applicable
		if tenant.SubscriptionStatus == "trialing" && tenant.TrialEndsAt != nil {
			daysRemaining := int(time.Until(*tenant.TrialEndsAt).Hours() / 24)
			c.Set("trial_days_remaining", daysRemaining)
		}

		c.Next()
	}
}

// PatientLimitMiddleware checks if tenant has reached patient limit
// Use this on patient creation endpoints
func PatientLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.Next()
			return
		}

		patientLimit, exists := c.Get("patient_limit")
		if !exists {
			c.Next()
			return
		}

		limit, ok := patientLimit.(int)
		if !ok || limit == 0 {
			c.Next()
			return
		}

		// Count current patients
		db := database.GetDB()
		schemaName := c.GetString("schema")
		if schemaName == "" {
			// SECURITY: Fail closed - do not fallback to tenant_1
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "schema_not_found",
				"message": "Erro interno: schema não encontrado no contexto",
			})
			c.Abort()
			return
		}

		var count int64
		db.Table(schemaName + ".patients").Count(&count)

		if int(count) >= limit {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "patient_limit_reached",
				"message": "Você atingiu o limite de pacientes do seu plano. Faça upgrade para adicionar mais pacientes.",
				"current": count,
				"limit":   limit,
				"tenant_id": tenantID,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
