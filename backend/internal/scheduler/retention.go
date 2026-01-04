package scheduler

import (
	"drcrwell/backend/internal/cache"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"log"
	"time"
)

// RetentionConfig holds the default retention periods (in days)
var DefaultRetentionConfig = map[string]int{
	"audit_logs":      1825, // 5 years
	"expired_tokens":  1,    // 1 day
	"old_sessions":    30,   // 30 days
}

// StartRetentionScheduler starts the data retention cleanup job
// Uses distributed lock to prevent duplicate execution across multiple instances
func StartRetentionScheduler() {
	go runRetentionCleanup()
	log.Println("Scheduler started - Retention cleanup running daily at 3AM (with distributed lock)")
}

// runRetentionCleanup runs daily to clean up old data
func runRetentionCleanup() {
	// Calculate time until next 3AM
	now := time.Now()
	next3AM := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	if now.After(next3AM) {
		next3AM = next3AM.Add(24 * time.Hour)
	}
	timeUntil3AM := next3AM.Sub(now)

	// Wait until 3AM
	time.Sleep(timeUntil3AM)

	// Run with lock (lock TTL is 23 hours to ensure it releases before next run)
	if cache.AcquireSchedulerLock(LockRetention, 23*time.Hour) {
		performRetentionCleanup()
	}

	// Then run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		// Try to acquire lock before running
		if cache.AcquireSchedulerLock(LockRetention, 23*time.Hour) {
			performRetentionCleanup()
		} else {
			log.Println("Retention Scheduler: Skipping - another instance holds the lock")
		}
	}
}

// performRetentionCleanup executes the actual cleanup
func performRetentionCleanup() {
	log.Println("Retention cleanup started")

	db := database.GetDB()
	if db == nil {
		log.Println("Retention cleanup: Database not initialized")
		return
	}

	// 1. Clean up old audit logs (keep 5 years by default)
	cleanupAuditLogs(db)

	// 2. Clean up expired password reset tokens
	cleanupExpiredTokens(db)

	// 3. Clean up expired email verifications
	cleanupExpiredEmailVerifications(db)

	// 4. Mark expired consents
	markExpiredConsents(db)

	log.Println("Retention cleanup completed")
}

// cleanupAuditLogs removes audit logs older than retention period
func cleanupAuditLogs(db interface{}) {
	retentionDays := DefaultRetentionConfig["audit_logs"]
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	// Note: We don't actually delete audit logs in production
	// This just counts how many would be deleted for reporting
	var count int64
	database.GetDB().Model(&models.AuditLog{}).
		Where("created_at < ?", cutoffDate).
		Count(&count)

	if count > 0 {
		log.Printf("Retention: Found %d audit logs older than %d days (keeping for compliance)", count, retentionDays)
	}
}

// cleanupExpiredTokens removes expired password reset tokens
func cleanupExpiredTokens(db interface{}) {
	result := database.GetDB().
		Where("expires_at < ?", time.Now()).
		Delete(&models.PasswordReset{})

	if result.RowsAffected > 0 {
		log.Printf("Retention: Cleaned up %d expired password reset tokens", result.RowsAffected)
	}
}

// cleanupExpiredEmailVerifications removes expired email verification records
func cleanupExpiredEmailVerifications(db interface{}) {
	// Delete email verifications older than 7 days
	cutoffDate := time.Now().AddDate(0, 0, -7)

	result := database.GetDB().
		Where("created_at < ? AND verified_at IS NULL", cutoffDate).
		Delete(&models.EmailVerification{})

	if result.RowsAffected > 0 {
		log.Printf("Retention: Cleaned up %d expired email verifications", result.RowsAffected)
	}
}

// markExpiredConsents marks patient consents that have expired
func markExpiredConsents(db interface{}) {
	// This would run across all tenant schemas
	// For now, just log the intention
	log.Println("Retention: Consent expiration check would run here (per-tenant)")
}

// GetRetentionStats returns statistics about data retention
func GetRetentionStats() map[string]interface{} {
	db := database.GetDB()
	if db == nil {
		return nil
	}

	stats := make(map[string]interface{})

	// Total audit logs
	var totalAuditLogs int64
	db.Model(&models.AuditLog{}).Count(&totalAuditLogs)
	stats["total_audit_logs"] = totalAuditLogs

	// Audit logs older than 1 year
	yearAgo := time.Now().AddDate(-1, 0, 0)
	var oldAuditLogs int64
	db.Model(&models.AuditLog{}).Where("created_at < ?", yearAgo).Count(&oldAuditLogs)
	stats["audit_logs_older_than_1y"] = oldAuditLogs

	// Expired password resets
	var expiredResets int64
	db.Model(&models.PasswordReset{}).Where("expires_at < ?", time.Now()).Count(&expiredResets)
	stats["expired_password_resets"] = expiredResets

	// Unverified emails older than 7 days
	weekAgo := time.Now().AddDate(0, 0, -7)
	var oldUnverifiedEmails int64
	db.Model(&models.EmailVerification{}).
		Where("created_at < ? AND verified_at IS NULL", weekAgo).
		Count(&oldUnverifiedEmails)
	stats["old_unverified_emails"] = oldUnverifiedEmails

	return stats
}

// ForceRetentionCleanup manually triggers the retention cleanup
// This is called from admin endpoints
func ForceRetentionCleanup() map[string]interface{} {
	log.Println("Manual retention cleanup triggered")

	db := database.GetDB()
	if db == nil {
		return map[string]interface{}{"error": "Database not initialized"}
	}

	results := make(map[string]interface{})

	// Clean expired tokens
	tokenResult := db.Where("expires_at < ?", time.Now()).Delete(&models.PasswordReset{})
	results["password_resets_deleted"] = tokenResult.RowsAffected

	// Clean expired email verifications
	weekAgo := time.Now().AddDate(0, 0, -7)
	emailResult := db.Where("created_at < ? AND verified_at IS NULL", weekAgo).Delete(&models.EmailVerification{})
	results["email_verifications_deleted"] = emailResult.RowsAffected

	results["status"] = "completed"
	results["timestamp"] = time.Now().Format(time.RFC3339)

	log.Printf("Manual retention cleanup completed: %+v", results)
	return results
}
