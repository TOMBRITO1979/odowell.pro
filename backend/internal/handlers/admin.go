package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"drcrwell/backend/internal/scheduler"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// TenantWithStats represents tenant with subscription and payment info
type TenantWithStats struct {
	models.Tenant
	UserCount           int        `json:"user_count"`
	PatientCount        int        `json:"patient_count"`
	LastPaymentDate     *time.Time `json:"last_payment_date"`
	LastPaymentAmount   int        `json:"last_payment_amount"`
	SubscriptionDetails *models.Subscription `json:"subscription_details"`
}

// GetAllTenants returns all tenants with subscription info (super admin only)
func GetAllTenants(c *gin.Context) {
	var tenants []models.Tenant
	if err := database.DB.Order("created_at DESC").Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar clínicas"})
		return
	}

	var result []TenantWithStats
	for _, tenant := range tenants {
		stats := TenantWithStats{
			Tenant: tenant,
		}

		// Get user count
		var userCount int64
		database.DB.Model(&models.User{}).Where("tenant_id = ?", tenant.ID).Count(&userCount)
		stats.UserCount = int(userCount)

		// Get patient count from tenant schema
		var patientCount int64
		tableName := fmt.Sprintf("tenant_%d.patients", tenant.ID)
		database.DB.Raw("SELECT COUNT(*) FROM " + tableName).Scan(&patientCount)
		stats.PatientCount = int(patientCount)

		// Get subscription details
		var subscription models.Subscription
		if err := database.DB.Where("tenant_id = ?", tenant.ID).First(&subscription).Error; err == nil {
			stats.SubscriptionDetails = &subscription
			stats.LastPaymentDate = subscription.CurrentPeriodStart
			stats.LastPaymentAmount = subscription.PriceMonthly
		}

		result = append(result, stats)
	}

	c.JSON(http.StatusOK, gin.H{"tenants": result})
}

// GetTenantDetails returns detailed info for a specific tenant
func GetTenantDetails(c *gin.Context) {
	tenantID := c.Param("id")

	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clínica não encontrada"})
		return
	}

	stats := TenantWithStats{
		Tenant: tenant,
	}

	// Get user count
	var userCount int64
	database.DB.Model(&models.User{}).Where("tenant_id = ?", tenant.ID).Count(&userCount)
	stats.UserCount = int(userCount)

	// Get patient count from tenant schema
	var patientCount int64
	tableName := fmt.Sprintf("tenant_%d.patients", tenant.ID)
	database.DB.Raw("SELECT COUNT(*) FROM " + tableName).Scan(&patientCount)
	stats.PatientCount = int(patientCount)

	// Get subscription details
	var subscription models.Subscription
	if err := database.DB.Where("tenant_id = ?", tenant.ID).First(&subscription).Error; err == nil {
		stats.SubscriptionDetails = &subscription
		stats.LastPaymentDate = subscription.CurrentPeriodStart
		stats.LastPaymentAmount = subscription.PriceMonthly
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateTenantStatus activates or deactivates a tenant
func UpdateTenantStatus(c *gin.Context) {
	tenantID := c.Param("id")

	var input struct {
		Active             *bool   `json:"active"`
		SubscriptionStatus string  `json:"subscription_status"`
		PlanType           string  `json:"plan_type"`
		PatientLimit       int     `json:"patient_limit"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clínica não encontrada"})
		return
	}

	// Update fields
	if input.Active != nil {
		tenant.Active = *input.Active
	}
	if input.SubscriptionStatus != "" {
		tenant.SubscriptionStatus = input.SubscriptionStatus
	}
	if input.PlanType != "" {
		tenant.PlanType = input.PlanType
	}
	if input.PatientLimit > 0 {
		tenant.PatientLimit = input.PatientLimit
	}

	if err := database.DB.Save(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar clínica"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Clínica atualizada com sucesso",
		"tenant":  tenant,
	})
}

// GetAdminTenantUsers returns all users for a specific tenant (super admin)
func GetAdminTenantUsers(c *gin.Context) {
	tenantID := c.Param("id")

	var users []models.User
	if err := database.DB.Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar usuários"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// UpdateTenantUserStatus activates or deactivates a user in a tenant
func UpdateTenantUserStatus(c *gin.Context) {
	tenantID := c.Param("id")
	userID := c.Param("userId")

	var input struct {
		Active bool `json:"active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tid, _ := strconv.Atoi(tenantID)
	uid, _ := strconv.Atoi(userID)

	var user models.User
	if err := database.DB.Where("id = ? AND tenant_id = ?", uid, tid).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	user.Active = input.Active

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar usuário"})
		return
	}

	status := "ativado"
	if !input.Active {
		status = "desativado"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Usuário " + status + " com sucesso",
		"user":    user,
	})
}

// GetAdminDashboard returns overall system statistics
func GetAdminDashboard(c *gin.Context) {
	var totalTenants int64
	var activeTenants int64
	var trialTenants int64
	var totalUsers int64
	var activeSubscriptions int64
	var unverifiedTenants int64

	database.DB.Model(&models.Tenant{}).Count(&totalTenants)
	database.DB.Model(&models.Tenant{}).Where("active = ?", true).Count(&activeTenants)
	database.DB.Model(&models.Tenant{}).Where("subscription_status = ?", "trialing").Count(&trialTenants)
	database.DB.Model(&models.User{}).Count(&totalUsers)
	database.DB.Model(&models.Subscription{}).Where("status = ?", "active").Count(&activeSubscriptions)
	database.DB.Model(&models.Tenant{}).Where("email_verified = ?", false).Count(&unverifiedTenants)

	// Get revenue this month
	var monthlyRevenue int64
	database.DB.Model(&models.Subscription{}).
		Where("status = ?", "active").
		Select("COALESCE(SUM(price_monthly), 0)").
		Scan(&monthlyRevenue)

	// Get trial stats
	activeTrials, expiringSoon, expiredTrials := scheduler.GetExpiredTrialStats()

	// Get inactive tenants count (expired, canceled, past_due)
	var expiredCount, canceledCount, pastDueCount int64
	database.DB.Model(&models.Tenant{}).Where("subscription_status = ?", "expired").Count(&expiredCount)
	database.DB.Model(&models.Tenant{}).Where("subscription_status = ?", "canceled").Count(&canceledCount)
	database.DB.Model(&models.Tenant{}).Where("subscription_status = ?", "past_due").Count(&pastDueCount)

	c.JSON(http.StatusOK, gin.H{
		"total_tenants":        totalTenants,
		"active_tenants":       activeTenants,
		"trial_tenants":        trialTenants,
		"total_users":          totalUsers,
		"active_subscriptions": activeSubscriptions,
		"monthly_revenue":      monthlyRevenue,
		"unverified_tenants":   unverifiedTenants,
		"trial_stats": gin.H{
			"active":        activeTrials,
			"expiring_soon": expiringSoon,
			"expired":       expiredTrials,
		},
		"inactive_stats": gin.H{
			"expired":  expiredCount,
			"canceled": canceledCount,
			"past_due": pastDueCount,
			"total":    expiredCount + canceledCount + pastDueCount,
		},
	})
}

// GetUnverifiedTenants returns all tenants that haven't verified their email
func GetUnverifiedTenants(c *gin.Context) {
	var tenants []models.Tenant
	if err := database.DB.Where("email_verified = ?", false).Order("created_at DESC").Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar clínicas não verificadas"})
		return
	}

	type UnverifiedTenant struct {
		ID        uint      `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		HoursSinceCreation float64 `json:"hours_since_creation"`
	}

	var result []UnverifiedTenant
	for _, tenant := range tenants {
		hoursSince := time.Since(tenant.CreatedAt).Hours()
		result = append(result, UnverifiedTenant{
			ID:                 tenant.ID,
			Name:               tenant.Name,
			Email:              tenant.Email,
			CreatedAt:          tenant.CreatedAt,
			HoursSinceCreation: hoursSince,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": result,
		"count":   len(result),
	})
}

// GetExpiringTrials returns tenants with trials expiring soon (within 24 hours)
func GetExpiringTrials(c *gin.Context) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	var tenants []models.Tenant
	if err := database.DB.Where(
		"subscription_status = ? AND trial_ends_at > ? AND trial_ends_at <= ? AND active = ?",
		"trialing",
		now,
		tomorrow,
		true,
	).Order("trial_ends_at ASC").Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar trials expirando"})
		return
	}

	type ExpiringTrial struct {
		ID          uint      `json:"id"`
		Name        string    `json:"name"`
		Email       string    `json:"email"`
		TrialEndsAt time.Time `json:"trial_ends_at"`
		HoursLeft   float64   `json:"hours_left"`
	}

	var result []ExpiringTrial
	for _, tenant := range tenants {
		hoursLeft := time.Until(*tenant.TrialEndsAt).Hours()
		result = append(result, ExpiringTrial{
			ID:          tenant.ID,
			Name:        tenant.Name,
			Email:       tenant.Email,
			TrialEndsAt: *tenant.TrialEndsAt,
			HoursLeft:   hoursLeft,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": result,
		"count":   len(result),
	})
}

// GetInactiveTenants returns tenants that are expired, canceled, or past_due (didn't activate or renew)
func GetInactiveTenants(c *gin.Context) {
	var tenants []models.Tenant

	// Get tenants with subscription issues: expired, canceled, past_due, or inactive
	if err := database.DB.Where(
		"subscription_status IN (?, ?, ?) OR active = ?",
		"expired",
		"canceled",
		"past_due",
		false,
	).Order("updated_at DESC").Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar clínicas inativas"})
		return
	}

	type InactiveTenant struct {
		ID                 uint       `json:"id"`
		Name               string     `json:"name"`
		Email              string     `json:"email"`
		SubscriptionStatus string     `json:"subscription_status"`
		Active             bool       `json:"active"`
		TrialEndsAt        *time.Time `json:"trial_ends_at"`
		CreatedAt          time.Time  `json:"created_at"`
		UpdatedAt          time.Time  `json:"updated_at"`
		Reason             string     `json:"reason"`
	}

	var result []InactiveTenant
	for _, tenant := range tenants {
		reason := ""
		switch tenant.SubscriptionStatus {
		case "expired":
			reason = "Trial expirado - não assinou"
		case "canceled":
			reason = "Assinatura cancelada"
		case "past_due":
			reason = "Pagamento pendente"
		default:
			if !tenant.Active {
				reason = "Desativado manualmente"
			}
		}

		result = append(result, InactiveTenant{
			ID:                 tenant.ID,
			Name:               tenant.Name,
			Email:              tenant.Email,
			SubscriptionStatus: tenant.SubscriptionStatus,
			Active:             tenant.Active,
			TrialEndsAt:        tenant.TrialEndsAt,
			CreatedAt:          tenant.CreatedAt,
			UpdatedAt:          tenant.UpdatedAt,
			Reason:             reason,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": result,
		"count":   len(result),
	})
}

// DeleteTenant soft deletes a tenant and deactivates all its users (super admin only)
func DeleteTenant(c *gin.Context) {
	tenantID := c.Param("id")

	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clínica não encontrada"})
		return
	}

	// Soft delete the tenant - GORM will set deleted_at automatically
	if err := database.DB.Delete(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar clínica"})
		return
	}

	// Deactivate all users of this tenant
	if err := database.DB.Model(&models.User{}).
		Where("tenant_id = ?", tenantID).
		Update("active", false).Error; err != nil {
		// Log but don't fail - tenant is already deleted
		fmt.Printf("Warning: failed to deactivate users for tenant %s: %v\n", tenantID, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Clínica deletada com sucesso",
	})
}
