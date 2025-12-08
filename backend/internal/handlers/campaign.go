package handlers

import (
	"drcrwell/backend/internal/models"
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
	campaign.Status = "draft"

	// Set default empty JSON for JSONB field if empty (PostgreSQL requires valid JSON)
	if campaign.Filters == "" {
		campaign.Filters = "{}"
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create campaign"})
		return
	}

	// Load relationships
	db.Preload("CreatedBy").First(&campaign, campaign.ID)

	c.JSON(http.StatusCreated, gin.H{"campaign": campaign})
}

func GetCampaigns(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

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
	db := c.MustGet("db").(*gorm.DB)

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
	db := c.MustGet("db").(*gorm.DB)

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

	// Update using Exec to avoid the duplicate table error
	result := db.Exec(`
		UPDATE campaigns
		SET name = ?, type = ?, subject = ?, message = ?, segment_type = ?,
		    tags = ?, filters = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, input.Name, input.Type, input.Subject, input.Message, input.SegmentType,
		input.Tags, input.Filters, id)

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
	db := c.MustGet("db").(*gorm.DB)

	if err := db.Delete(&models.Campaign{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete campaign"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Campaign deleted successfully"})
}

func SendCampaign(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var campaign models.Campaign
	if err := db.First(&campaign, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	if campaign.Status != "draft" && campaign.Status != "scheduled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Campaign already sent"})
		return
	}

	// Get recipients based on segmentation
	var patients []models.Patient
	query := db.Model(&models.Patient{}).Where("active = ?", true)

	if campaign.SegmentType == "tags" && campaign.Tags != "" {
		tags := strings.Split(campaign.Tags, ",")
		for _, tag := range tags {
			query = query.Where("tags ILIKE ?", "%"+strings.TrimSpace(tag)+"%")
		}
	}

	if err := query.Find(&patients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipients"})
		return
	}

	// Create campaign recipients
	tx := db.Begin()

	for _, patient := range patients {
		recipient := models.CampaignRecipient{
			CampaignID: campaign.ID,
			PatientID:  patient.ID,
			Status:     "pending",
		}
		if err := tx.Create(&recipient).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recipients"})
			return
		}
	}

	// Update campaign
	campaign.TotalRecipients = len(patients)
	campaign.Status = "scheduled"
	now := time.Now()
	campaign.ScheduledAt = &now

	if err := tx.Save(&campaign).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update campaign"})
		return
	}

	tx.Commit()

	// TODO: Implement actual sending logic (WhatsApp, Email)
	// This would be done asynchronously via queue/worker

	c.JSON(http.StatusOK, gin.H{
		"message":    "Campaign scheduled for sending",
		"campaign":   campaign,
		"recipients": len(patients),
	})
}
