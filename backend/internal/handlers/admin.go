package handlers

import (
	"drcrwell/backend/internal/cache"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"drcrwell/backend/internal/scheduler"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminDashboardResponse represents admin dashboard data for caching
type AdminDashboardResponse struct {
	TotalTenants        int64 `json:"total_tenants"`
	ActiveTenants       int64 `json:"active_tenants"`
	TrialTenants        int64 `json:"trial_tenants"`
	TotalUsers          int64 `json:"total_users"`
	ActiveSubscriptions int64 `json:"active_subscriptions"`
	MonthlyRevenue      int64 `json:"monthly_revenue"`
	UnverifiedTenants   int64 `json:"unverified_tenants"`
	TrialStats          struct {
		Active       int64 `json:"active"`
		ExpiringSoon int64 `json:"expiring_soon"`
		Expired      int64 `json:"expired"`
	} `json:"trial_stats"`
	InactiveStats struct {
		Expired  int64 `json:"expired"`
		Canceled int64 `json:"canceled"`
		PastDue  int64 `json:"past_due"`
		Total    int64 `json:"total"`
	} `json:"inactive_stats"`
}

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

	// Batch query: Get user counts per tenant (1 query instead of N)
	type UserCountResult struct {
		TenantID uint
		Count    int64
	}
	var userCounts []UserCountResult
	database.DB.Model(&models.User{}).
		Select("tenant_id, COUNT(*) as count").
		Group("tenant_id").
		Scan(&userCounts)

	// Create map for fast lookup
	userCountMap := make(map[uint]int64)
	for _, uc := range userCounts {
		userCountMap[uc.TenantID] = uc.Count
	}

	// Batch query: Get all subscriptions (1 query instead of N)
	var subscriptions []models.Subscription
	database.DB.Find(&subscriptions)

	// Create map for fast lookup
	subscriptionMap := make(map[uint]*models.Subscription)
	for i := range subscriptions {
		subscriptionMap[subscriptions[i].TenantID] = &subscriptions[i]
	}

	// Batch query: Get patient counts using a dynamic UNION query
	// This is unavoidable due to schema-per-tenant, but we do it in one query
	patientCountMap := make(map[uint]int64)
	if len(tenants) > 0 {
		// Build UNION query for all tenant schemas
		var unionParts []string
		for _, tenant := range tenants {
			unionParts = append(unionParts, fmt.Sprintf(
				"SELECT %d as tenant_id, COUNT(*) as count FROM tenant_%d.patients",
				tenant.ID, tenant.ID,
			))
		}
		if len(unionParts) > 0 {
			unionQuery := ""
			for i, part := range unionParts {
				if i > 0 {
					unionQuery += " UNION ALL "
				}
				unionQuery += part
			}

			type PatientCountResult struct {
				TenantID uint  `gorm:"column:tenant_id"`
				Count    int64 `gorm:"column:count"`
			}
			var patientCounts []PatientCountResult
			database.DB.Raw(unionQuery).Scan(&patientCounts)

			for _, pc := range patientCounts {
				patientCountMap[pc.TenantID] = pc.Count
			}
		}
	}

	// Build result using maps (no additional queries)
	var result []TenantWithStats
	for _, tenant := range tenants {
		stats := TenantWithStats{
			Tenant:       tenant,
			UserCount:    int(userCountMap[tenant.ID]),
			PatientCount: int(patientCountMap[tenant.ID]),
		}

		// Get subscription from map
		if sub, ok := subscriptionMap[tenant.ID]; ok {
			stats.SubscriptionDetails = sub
			stats.LastPaymentDate = sub.CurrentPeriodStart
			stats.LastPaymentAmount = sub.PriceMonthly
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
	cacheKey := "admin_dashboard"

	// Use cached data with 5-minute TTL
	result, err := cache.GetOrSetTyped[AdminDashboardResponse](cacheKey, cache.TTLDashboard, func() (AdminDashboardResponse, error) {
		var data AdminDashboardResponse

		database.DB.Model(&models.Tenant{}).Count(&data.TotalTenants)
		database.DB.Model(&models.Tenant{}).Where("active = ?", true).Count(&data.ActiveTenants)
		database.DB.Model(&models.Tenant{}).Where("subscription_status = ?", "trialing").Count(&data.TrialTenants)
		database.DB.Model(&models.User{}).Count(&data.TotalUsers)
		database.DB.Model(&models.Subscription{}).Where("status = ?", "active").Count(&data.ActiveSubscriptions)
		database.DB.Model(&models.Tenant{}).Where("email_verified = ?", false).Count(&data.UnverifiedTenants)

		// Get revenue this month
		database.DB.Model(&models.Subscription{}).
			Where("status = ?", "active").
			Select("COALESCE(SUM(price_monthly), 0)").
			Scan(&data.MonthlyRevenue)

		// Get trial stats
		activeTrials, expiringSoon, expiredTrials := scheduler.GetExpiredTrialStats()
		data.TrialStats.Active = activeTrials
		data.TrialStats.ExpiringSoon = expiringSoon
		data.TrialStats.Expired = expiredTrials

		// Get inactive tenants count (expired, canceled, past_due)
		database.DB.Model(&models.Tenant{}).Where("subscription_status = ?", "expired").Count(&data.InactiveStats.Expired)
		database.DB.Model(&models.Tenant{}).Where("subscription_status = ?", "canceled").Count(&data.InactiveStats.Canceled)
		database.DB.Model(&models.Tenant{}).Where("subscription_status = ?", "past_due").Count(&data.InactiveStats.PastDue)
		data.InactiveStats.Total = data.InactiveStats.Expired + data.InactiveStats.Canceled + data.InactiveStats.PastDue

		return data, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar dashboard administrativo"})
		return
	}

	c.JSON(http.StatusOK, result)
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
