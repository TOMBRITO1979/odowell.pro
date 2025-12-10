package scheduler

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"log"
	"time"
)

// StartScheduler starts background jobs
func StartScheduler() {
	go runTrialExpirationChecker()
	log.Println("Scheduler started - Trial expiration checker running every hour")

	go StartRetentionScheduler()
}

// runTrialExpirationChecker runs every hour to check and deactivate expired trials
func runTrialExpirationChecker() {
	// Run immediately on start
	checkExpiredTrials()

	// Then run every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		checkExpiredTrials()
	}
}

// checkExpiredTrials finds and deactivates tenants with expired trials
func checkExpiredTrials() {
	db := database.GetDB()
	if db == nil {
		log.Println("Scheduler: Database not initialized")
		return
	}

	now := time.Now()
	var expiredTenants []models.Tenant

	// Find tenants that:
	// 1. Are in trial status
	// 2. Trial has ended
	// 3. Are still active
	err := db.Where(
		"subscription_status = ? AND trial_ends_at < ? AND active = ?",
		"trialing",
		now,
		true,
	).Find(&expiredTenants).Error

	if err != nil {
		log.Printf("Scheduler: Error finding expired trials: %v", err)
		return
	}

	if len(expiredTenants) == 0 {
		return
	}

	log.Printf("Scheduler: Found %d expired trial(s) to deactivate", len(expiredTenants))

	for _, tenant := range expiredTenants {
		// Update tenant status
		err := db.Model(&tenant).Updates(map[string]interface{}{
			"active":              false,
			"subscription_status": "expired",
		}).Error

		if err != nil {
			log.Printf("Scheduler: Error deactivating tenant %d (%s): %v", tenant.ID, tenant.Name, err)
			continue
		}

		log.Printf("Scheduler: Deactivated expired trial for tenant %d (%s)", tenant.ID, tenant.Name)

		// TODO: Send email notification about trial expiration
	}
}

// GetExpiredTrialStats returns statistics about trial expirations
func GetExpiredTrialStats() (active int64, expiringSoon int64, expired int64) {
	db := database.GetDB()
	if db == nil {
		return 0, 0, 0
	}

	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	// Active trials
	db.Model(&models.Tenant{}).Where(
		"subscription_status = ? AND trial_ends_at > ? AND active = ?",
		"trialing",
		now,
		true,
	).Count(&active)

	// Expiring in 24 hours
	db.Model(&models.Tenant{}).Where(
		"subscription_status = ? AND trial_ends_at > ? AND trial_ends_at <= ? AND active = ?",
		"trialing",
		now,
		tomorrow,
		true,
	).Count(&expiringSoon)

	// Already expired (deactivated)
	db.Model(&models.Tenant{}).Where(
		"subscription_status = ?",
		"expired",
	).Count(&expired)

	return
}
