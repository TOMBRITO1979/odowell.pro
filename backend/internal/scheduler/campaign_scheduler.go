package scheduler

import (
	"drcrwell/backend/internal/cache"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/models"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// StartCampaignScheduler starts the campaign scheduler that processes scheduled campaigns
// Uses distributed lock to prevent duplicate campaign sends across multiple instances
func StartCampaignScheduler() {
	log.Println("Campaign Scheduler started - checking for scheduled campaigns every minute (with distributed lock)")

	// Run immediately on start (with lock)
	if cache.AcquireSchedulerLock(LockCampaign, 55*time.Second) {
		processScheduledCampaigns()
	}

	// Then run every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Try to acquire lock before running
		// Lock TTL is 55 seconds (slightly less than 1 minute interval)
		if cache.AcquireSchedulerLock(LockCampaign, 55*time.Second) {
			processScheduledCampaigns()
		} else {
			log.Println("Campaign Scheduler: Skipping - another instance holds the lock")
		}
	}
}

// processScheduledCampaigns finds and processes campaigns that are ready to be sent
func processScheduledCampaigns() {
	log.Printf("Campaign Scheduler: Running check at %s", time.Now().Format("15:04:05"))

	db := database.GetDB()
	if db == nil {
		log.Println("Campaign Scheduler: Database not initialized")
		return
	}

	now := time.Now()

	// Find all tenants with scheduled campaigns
	var tenantIDs []uint
	err := db.Raw(`
		SELECT DISTINCT t.id
		FROM public.tenants t
		WHERE t.active = true
		AND EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'tenant_' || t.id
			AND table_name = 'campaigns'
		)
	`).Scan(&tenantIDs).Error

	if err != nil {
		log.Printf("Campaign Scheduler: Error finding tenants: %v", err)
		return
	}

	log.Printf("Campaign Scheduler: Found %d active tenants to check", len(tenantIDs))

	for _, tenantID := range tenantIDs {
		processTenantCampaigns(db, tenantID, now)
	}
}

// processTenantCampaigns processes scheduled campaigns for a specific tenant
func processTenantCampaigns(db *gorm.DB, tenantID uint, now time.Time) {
	schemaName := fmt.Sprintf("tenant_%d", tenantID)

	// Use fresh session with schema set - execute SET and query in same transaction
	tenantDB := db.Session(&gorm.Session{PrepareStmt: false})

	// Find scheduled campaigns that are ready to send using explicit schema
	var campaigns []models.Campaign
	err := tenantDB.Raw(fmt.Sprintf(`
		SELECT id, created_at, updated_at, name, type, subject, message,
		       segment_type, tags, filters, status, scheduled_at, sent_at,
		       total_recipients, sent, delivered, failed, opened, clicked, created_by_id
		FROM %s.campaigns
		WHERE status = 'scheduled'
		AND scheduled_at <= ?
		AND deleted_at IS NULL
	`, schemaName), now).Scan(&campaigns).Error

	if err != nil {
		log.Printf("Campaign Scheduler: Error finding campaigns for tenant %d: %v", tenantID, err)
		return
	}

	if len(campaigns) == 0 {
		return
	}

	log.Printf("Campaign Scheduler: Found %d scheduled campaign(s) for tenant %d", len(campaigns), tenantID)

	// Get tenant SMTP settings
	var settings models.TenantSettings
	if err := db.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		log.Printf("Campaign Scheduler: SMTP settings not found for tenant %d: %v", tenantID, err)
		return
	}

	// Validate SMTP configuration
	if settings.SMTPHost == "" || settings.SMTPUsername == "" || settings.SMTPPassword == "" || settings.SMTPFromEmail == "" {
		log.Printf("Campaign Scheduler: Incomplete SMTP settings for tenant %d", tenantID)
		return
	}

	// Decrypt password
	password, err := helpers.DecryptIfNeeded(settings.SMTPPassword)
	if err != nil {
		log.Printf("Campaign Scheduler: Error decrypting SMTP password for tenant %d: %v", tenantID, err)
		return
	}

	// Get clinic name for emails
	clinicName := settings.ClinicName
	if clinicName == "" {
		clinicName = "Clínica"
	}

	// Build email config
	emailConfig := helpers.TenantEmailConfig{
		Host:      settings.SMTPHost,
		Port:      settings.SMTPPort,
		Username:  settings.SMTPUsername,
		Password:  password,
		FromName:  settings.SMTPFromName,
		FromEmail: settings.SMTPFromEmail,
		UseTLS:    settings.SMTPUseTLS,
	}

	log.Printf("Campaign Scheduler: SMTP config for tenant %d - Host: %s, Port: %d, Username: %s, FromEmail: %s, UseTLS: %v",
		tenantID, settings.SMTPHost, settings.SMTPPort, settings.SMTPUsername, settings.SMTPFromEmail, settings.SMTPUseTLS)

	// Process each campaign
	for _, campaign := range campaigns {
		processCampaign(tenantDB, schemaName, campaign, emailConfig, clinicName)
	}
}

// processCampaign sends a single scheduled campaign
func processCampaign(db *gorm.DB, schemaName string, campaign models.Campaign, emailConfig helpers.TenantEmailConfig, clinicName string) {
	log.Printf("Campaign Scheduler: Processing campaign %d (%s)", campaign.ID, campaign.Name)

	// Only process email campaigns for now
	if campaign.Type != "email" {
		log.Printf("Campaign Scheduler: Skipping non-email campaign %d (type: %s)", campaign.ID, campaign.Type)
		return
	}

	// Get recipients based on segmentation using raw SQL with explicit schema
	var patients []models.Patient
	baseQuery := fmt.Sprintf(`
		SELECT id, name, email, tags, active
		FROM %s.patients
		WHERE deleted_at IS NULL
		AND (active = true OR active IS NULL)
		AND email IS NOT NULL AND email != ''
	`, schemaName)

	if campaign.SegmentType == "tags" && campaign.Tags != "" {
		tags := strings.Split(campaign.Tags, ",")
		for _, tag := range tags {
			baseQuery += " AND tags ILIKE '%" + strings.TrimSpace(tag) + "%'"
		}
	}

	if err := db.Session(&gorm.Session{PrepareStmt: false}).Raw(baseQuery).Scan(&patients).Error; err != nil {
		log.Printf("Campaign Scheduler: Error fetching patients for campaign %d: %v", campaign.ID, err)
		return
	}

	if len(patients) == 0 {
		log.Printf("Campaign Scheduler: No patients found for campaign %d", campaign.ID)
		// Mark as sent with 0 recipients
		now := time.Now()
		db.Session(&gorm.Session{PrepareStmt: false}).Exec(fmt.Sprintf(`
			UPDATE %s.campaigns
			SET status = 'sent', total_recipients = 0, sent = 0, failed = 0, sent_at = $1, updated_at = $2
			WHERE id = $3 AND deleted_at IS NULL
		`, schemaName), now, now, campaign.ID)
		return
	}

	log.Printf("Campaign Scheduler: Found %d patients for campaign %d", len(patients), campaign.ID)

	// Update campaign status to sending
	now := time.Now()
	db.Session(&gorm.Session{PrepareStmt: false}).Exec(fmt.Sprintf(`
		UPDATE %s.campaigns
		SET status = 'sending', total_recipients = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`, schemaName), len(patients), now, campaign.ID)

	// Send emails
	sentCount := 0
	failedCount := 0

	for _, patient := range patients {
		// Create recipient record using raw SQL with explicit schema
		var recipientID uint
		recipientNow := time.Now()
		err := db.Session(&gorm.Session{PrepareStmt: false}).Raw(fmt.Sprintf(`
			INSERT INTO %s.campaign_recipients (created_at, updated_at, campaign_id, patient_id, status)
			VALUES ($1, $2, $3, $4, 'pending')
			RETURNING id
		`, schemaName), recipientNow, recipientNow, campaign.ID, patient.ID).Scan(&recipientID).Error
		if err != nil {
			log.Printf("Campaign Scheduler: Error creating recipient for patient %d: %v", patient.ID, err)
			continue
		}

		// Build email body
		patientName := patient.Name
		if patientName == "" {
			patientName = "Cliente"
		}

		body := helpers.BuildCampaignEmailBody(clinicName, patientName, campaign.Message)

		// Send email
		log.Printf("Campaign Scheduler: Sending email to %s, Subject: %s, From: %s", patient.Email, campaign.Subject, emailConfig.FromEmail)
		err = helpers.SendTenantEmail(emailConfig, patient.Email, campaign.Subject, body)

		if err != nil {
			// Update recipient as failed
			db.Session(&gorm.Session{PrepareStmt: false}).Exec(fmt.Sprintf(`
				UPDATE %s.campaign_recipients
				SET status = 'failed', error_message = $1, updated_at = $2
				WHERE id = $3
			`, schemaName), err.Error(), time.Now(), recipientID)
			failedCount++
			log.Printf("Campaign Scheduler: Failed to send email to %s: %v", patient.Email, err)
		} else {
			// Update recipient as sent
			sentTime := time.Now()
			db.Session(&gorm.Session{PrepareStmt: false}).Exec(fmt.Sprintf(`
				UPDATE %s.campaign_recipients
				SET status = 'sent', sent_at = $1, updated_at = $2
				WHERE id = $3
			`, schemaName), sentTime, sentTime, recipientID)
			sentCount++
			log.Printf("Campaign Scheduler: Email sent successfully to %s", patient.Email)
		}
	}

	// Update campaign statistics
	completedAt := time.Now()
	db.Session(&gorm.Session{PrepareStmt: false}).Exec(fmt.Sprintf(`
		UPDATE %s.campaigns
		SET status = 'sent', sent = $1, failed = $2, sent_at = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`, schemaName), sentCount, failedCount, completedAt, completedAt, campaign.ID)

	log.Printf("Campaign Scheduler: Campaign %d completed: %d sent, %d failed", campaign.ID, sentCount, failedCount)
}
