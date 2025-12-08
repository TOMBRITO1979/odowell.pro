package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"drcrwell/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateLead creates a new lead
func CreateLead(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("user_id").(uint)

	var lead models.Lead
	if err := c.ShouldBindJSON(&lead); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Clean phone number (remove spaces, dashes, etc)
	lead.Phone = cleanPhoneNumber(lead.Phone)

	// Set defaults
	lead.CreatedBy = userID
	if lead.Status == "" {
		lead.Status = "new"
	}
	if lead.Source == "" {
		lead.Source = "whatsapp"
	}

	if err := db.Create(&lead).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar lead"})
		return
	}

	c.JSON(http.StatusCreated, lead)
}

// GetLeads returns all leads with optional filtering
func GetLeads(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var leads []models.Lead
	query := db.Session(&gorm.Session{NewDB: true})

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by source
	if source := c.Query("source"); source != "" {
		query = query.Where("source = ?", source)
	}

	// Search by name or phone
	if search := c.Query("search"); search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("name ILIKE ? OR phone ILIKE ? OR email ILIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&models.Lead{}).Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&leads).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar leads"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      leads,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetLead returns a single lead by ID
func GetLead(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var lead models.Lead
	if err := db.First(&lead, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead não encontrado"})
		return
	}

	c.JSON(http.StatusOK, lead)
}

// UpdateLead updates a lead
func UpdateLead(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var lead models.Lead
	if err := db.First(&lead, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead não encontrado"})
		return
	}

	var input models.Lead
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Clean phone if provided
	if input.Phone != "" {
		input.Phone = cleanPhoneNumber(input.Phone)
	}

	// Update fields
	updates := map[string]interface{}{
		"name":           input.Name,
		"phone":          input.Phone,
		"email":          input.Email,
		"source":         input.Source,
		"contact_reason": input.ContactReason,
		"status":         input.Status,
		"notes":          input.Notes,
	}

	if err := db.Model(&lead).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar lead"})
		return
	}

	// Reload to get updated data
	db.First(&lead, id)
	c.JSON(http.StatusOK, lead)
}

// DeleteLead soft deletes a lead
func DeleteLead(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var lead models.Lead
	if err := db.First(&lead, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead não encontrado"})
		return
	}

	if err := db.Delete(&lead).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir lead"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead excluído com sucesso"})
}

// CheckLeadByPhone checks if a lead or patient exists by phone number
// This is used by WhatsApp integration to check if contact is known
func CheckLeadByPhone(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	phone := c.Param("phone")

	// Clean phone number
	phone = cleanPhoneNumber(phone)

	// Check if patient exists with this phone
	var patient models.Patient
	patientExists := db.Where("phone = ? OR cell_phone = ?", phone, phone).First(&patient).Error == nil

	if patientExists {
		c.JSON(http.StatusOK, gin.H{
			"exists":     true,
			"type":       "patient",
			"id":         patient.ID,
			"name":       patient.Name,
			"patient_id": patient.ID,
		})
		return
	}

	// Check if lead exists with this phone
	var lead models.Lead
	leadExists := db.Where("phone = ?", phone).First(&lead).Error == nil

	if leadExists {
		c.JSON(http.StatusOK, gin.H{
			"exists":  true,
			"type":    "lead",
			"id":      lead.ID,
			"name":    lead.Name,
			"lead_id": lead.ID,
			"status":  lead.Status,
		})
		return
	}

	// Not found
	c.JSON(http.StatusOK, gin.H{
		"exists": false,
		"type":   "unknown",
	})
}

// ConvertLeadToPatient converts a lead to a patient
func ConvertLeadToPatient(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	id := c.Param("id")

	var lead models.Lead
	if err := db.First(&lead, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead não encontrado"})
		return
	}

	// Check if already converted
	if lead.ConvertedToPatientID != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lead já foi convertido para paciente"})
		return
	}

	// Optional: receive additional patient data
	var additionalData struct {
		CPF       string `json:"cpf"`
		BirthDate string `json:"birth_date"`
		Address   string `json:"address"`
		City      string `json:"city"`
		State     string `json:"state"`
		ZipCode   string `json:"zip_code"`
		Notes     string `json:"notes"`
	}
	c.ShouldBindJSON(&additionalData)

	// Create patient from lead
	patient := models.Patient{
		Name:      lead.Name,
		Phone:     lead.Phone,
		CellPhone: lead.Phone,
		Email:     lead.Email,
		CPF:       additionalData.CPF,
		Address:   additionalData.Address,
		City:      additionalData.City,
		State:     additionalData.State,
		ZipCode:   additionalData.ZipCode,
		Notes:     lead.ContactReason + "\n\n" + lead.Notes + "\n\n" + additionalData.Notes,
		Active:    true,
	}

	// Parse birth date if provided
	if additionalData.BirthDate != "" {
		if t, err := time.Parse("2006-01-02", additionalData.BirthDate); err == nil {
			patient.BirthDate = &t
		}
	}

	// Create patient
	if err := db.Create(&patient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar paciente"})
		return
	}

	// Update lead status
	now := time.Now()
	lead.Status = "converted"
	lead.ConvertedToPatientID = &patient.ID
	lead.ConvertedAt = &now

	if err := db.Save(&lead).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar lead"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Lead convertido para paciente com sucesso",
		"lead":       lead,
		"patient":    patient,
		"patient_id": patient.ID,
	})
}

// GetLeadStats returns statistics about leads
func GetLeadStats(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var statusCounts []StatusCount
	db.Session(&gorm.Session{NewDB: true}).
		Table("leads").
		Select("status, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("status").
		Scan(&statusCounts)

	var total int64
	db.Session(&gorm.Session{NewDB: true}).
		Table("leads").
		Where("deleted_at IS NULL").
		Count(&total)

	var thisMonth int64
	db.Session(&gorm.Session{NewDB: true}).
		Table("leads").
		Where("deleted_at IS NULL").
		Where("EXTRACT(MONTH FROM created_at) = EXTRACT(MONTH FROM CURRENT_DATE)").
		Where("EXTRACT(YEAR FROM created_at) = EXTRACT(YEAR FROM CURRENT_DATE)").
		Count(&thisMonth)

	var converted int64
	db.Session(&gorm.Session{NewDB: true}).
		Table("leads").
		Where("deleted_at IS NULL").
		Where("status = ?", "converted").
		Count(&converted)

	c.JSON(http.StatusOK, gin.H{
		"total":         total,
		"this_month":    thisMonth,
		"converted":     converted,
		"by_status":     statusCounts,
	})
}

// Helper function to clean phone numbers
func cleanPhoneNumber(phone string) string {
	// Remove common characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "+", "")
	return phone
}
