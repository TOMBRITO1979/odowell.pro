package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateCampaign(c *gin.Context) {
	var campaign models.Campaign
	if err := c.ShouldBindJSON(&campaign); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	campaign.CreatedByID = c.GetUint("user_id")

	// Se tem scheduled_at, definir status como scheduled, senão draft
	if campaign.ScheduledAt != nil && !campaign.ScheduledAt.IsZero() {
		campaign.Status = "scheduled"
	} else if campaign.Status == "" {
		campaign.Status = "draft"
	}

	// Set default empty JSON for JSONB field if empty (PostgreSQL requires valid JSON)
	if campaign.Filters == "" {
		campaign.Filters = "{}"
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	if err := db.Create(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create campaign"})
		return
	}

	// Load relationships
	db.Preload("CreatedBy").First(&campaign, campaign.ID)

	c.JSON(http.StatusCreated, gin.H{"campaign": campaign})
}

func GetCampaigns(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.Campaign{})

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if campaignType := c.Query("type"); campaignType != "" {
		query = query.Where("type = ?", campaignType)
	}

	var total int64
	query.Count(&total)

	var campaigns []models.Campaign
	if err := query.Preload("CreatedBy").
		Offset(offset).Limit(pageSize).Order("created_at DESC").
		Find(&campaigns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch campaigns"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"campaigns": campaigns,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetCampaign(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var campaign models.Campaign
	if err := db.Preload("CreatedBy").Preload("Recipients").Preload("Recipients.Patient").
		First(&campaign, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"campaign": campaign})
}

func UpdateCampaign(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Check if campaign exists and is draft
	var campaign models.Campaign
	if err := db.First(&campaign, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	if campaign.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only update draft campaigns"})
		return
	}

	var input models.Campaign
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure filters is valid JSON (empty object if not provided)
	filters := input.Filters
	if filters == "" {
		filters = "{}"
	}

	// Update using Exec to avoid the duplicate table error
	result := db.Exec(`
		UPDATE campaigns
		SET name = ?, type = ?, subject = ?, message = ?, segment_type = ?,
		    tags = ?, filters = ?::jsonb, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, input.Name, input.Type, input.Subject, input.Message, input.SegmentType,
		input.Tags, filters, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update campaign"})
		return
	}

	// Load the updated campaign with relationships
	db.Preload("CreatedBy").First(&campaign, id)

	c.JSON(http.StatusOK, gin.H{"campaign": campaign})
}

func DeleteCampaign(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	if err := db.Delete(&models.Campaign{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete campaign"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Campaign deleted successfully"})
}

func SendCampaign(c *gin.Context) {
	id := c.Param("id")
	log.Printf("SendCampaign: Iniciando envio da campanha %s", id)

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		log.Printf("SendCampaign: Falha ao obter DB do contexto")
		return
	}
	tenantID := c.GetUint("tenant_id")
	log.Printf("SendCampaign: tenant_id=%d", tenantID)

	var campaign models.Campaign
	if err := db.First(&campaign, id).Error; err != nil {
		log.Printf("SendCampaign: Campanha não encontrada: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Campanha não encontrada"})
		return
	}
	log.Printf("SendCampaign: Campanha encontrada, status=%s, type=%s", campaign.Status, campaign.Type)

	if campaign.Status != "draft" && campaign.Status != "scheduled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campanha já foi enviada"})
		return
	}

	// Only process email campaigns for now
	if campaign.Type != "email" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Apenas campanhas de email são suportadas no momento"})
		return
	}

	// Get tenant SMTP settings (use database.DB for public schema)
	log.Printf("SendCampaign: Buscando configurações SMTP para tenant %d", tenantID)
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		log.Printf("SendCampaign: Configurações SMTP não encontradas: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configurações SMTP não encontradas. Configure o SMTP em Configurações."})
		return
	}
	log.Printf("SendCampaign: SMTP Host=%s, Port=%d, Username=%s", settings.SMTPHost, settings.SMTPPort, settings.SMTPUsername)

	// Validate SMTP configuration
	if settings.SMTPHost == "" || settings.SMTPUsername == "" || settings.SMTPPassword == "" || settings.SMTPFromEmail == "" {
		log.Printf("SendCampaign: Configurações SMTP incompletas")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configurações SMTP incompletas. Configure o SMTP em Configurações."})
		return
	}

	// Decrypt password
	log.Printf("SendCampaign: Descriptografando senha SMTP")
	password, err := helpers.DecryptIfNeeded(settings.SMTPPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao descriptografar senha SMTP"})
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

	// Get recipients based on segmentation using raw SQL to avoid prepared statement issues
	var patients []models.Patient
	baseQuery := "SELECT * FROM patients WHERE deleted_at IS NULL AND (active = true OR active IS NULL) AND email IS NOT NULL AND email != ''"

	if campaign.SegmentType == "tags" && campaign.Tags != "" {
		tags := strings.Split(campaign.Tags, ",")
		for _, tag := range tags {
			baseQuery += " AND tags ILIKE '%" + strings.TrimSpace(tag) + "%'"
		}
	}

	log.Printf("SendCampaign: Executando query: %s", baseQuery)
	// Use fresh session to avoid prepared statement cache issues
	if err := db.Session(&gorm.Session{PrepareStmt: false}).Raw(baseQuery).Scan(&patients).Error; err != nil {
		log.Printf("SendCampaign: Erro ao buscar pacientes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar destinatários"})
		return
	}
	log.Printf("SendCampaign: Encontrados %d pacientes com email", len(patients))

	if len(patients) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum paciente com email encontrado para esta campanha"})
		return
	}

	// Update campaign status to sending using raw SQL with fresh session
	now := time.Now()
	log.Printf("SendCampaign: Atualizando status da campanha para 'sending'")
	if err := db.Session(&gorm.Session{PrepareStmt: false}).Exec(`
		UPDATE campaigns
		SET status = 'sending', total_recipients = ?, scheduled_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, len(patients), now, now, campaign.ID).Error; err != nil {
		log.Printf("SendCampaign: Erro ao atualizar campanha: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar campanha"})
		return
	}
	log.Printf("SendCampaign: Campanha atualizada com sucesso")

	// Reload campaign with updated values using fresh session
	db.Session(&gorm.Session{PrepareStmt: false}).First(&campaign, campaign.ID)

	// Send emails in a goroutine to avoid blocking the request
	go func() {
		sentCount := 0
		failedCount := 0

		for _, patient := range patients {
			// Create recipient record using raw SQL to avoid GORM association issues
			var recipientID uint
			now := time.Now()
			err := db.Session(&gorm.Session{PrepareStmt: false}).Raw(`
				INSERT INTO campaign_recipients (created_at, updated_at, campaign_id, patient_id, status)
				VALUES (?, ?, ?, ?, 'pending')
				RETURNING id
			`, now, now, campaign.ID, patient.ID).Scan(&recipientID).Error
			if err != nil {
				log.Printf("Erro ao criar recipient para paciente %d: %v", patient.ID, err)
				continue
			}

			// Build email body
			patientName := patient.Name
			if patientName == "" {
				patientName = "Cliente"
			}

			body := helpers.BuildCampaignEmailBody(clinicName, patientName, campaign.Message)

			// Send email
			err = helpers.SendTenantEmail(emailConfig, patient.Email, campaign.Subject, body)

			if err != nil {
				// Update recipient as failed
				db.Session(&gorm.Session{PrepareStmt: false}).Exec(`
					UPDATE campaign_recipients
					SET status = 'failed', error_message = ?, updated_at = ?
					WHERE id = ?
				`, err.Error(), time.Now(), recipientID)
				failedCount++
				log.Printf("Falha ao enviar email para %s: %v", patient.Email, err)
			} else {
				// Update recipient as sent
				sentTime := time.Now()
				db.Session(&gorm.Session{PrepareStmt: false}).Exec(`
					UPDATE campaign_recipients
					SET status = 'sent', sent_at = ?, updated_at = ?
					WHERE id = ?
				`, sentTime, sentTime, recipientID)
				sentCount++
				log.Printf("Email enviado com sucesso para %s", patient.Email)
			}
		}

		// Update campaign statistics using raw SQL with fresh session
		completedAt := time.Now()
		db.Session(&gorm.Session{PrepareStmt: false}).Exec(`
			UPDATE campaigns
			SET status = 'sent', sent = ?, failed = ?, sent_at = ?, updated_at = ?
			WHERE id = ? AND deleted_at IS NULL
		`, sentCount, failedCount, completedAt, completedAt, campaign.ID)

		log.Printf("Campanha %d concluída: %d enviados, %d falhados", campaign.ID, sentCount, failedCount)
	}()

	log.Printf("SendCampaign: Respondendo sucesso - campanha iniciada")
	c.JSON(http.StatusOK, gin.H{
		"message":    "Campanha iniciada. Os emails estão sendo enviados.",
		"campaign":   campaign,
		"recipients": len(patients),
	})
}
