package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drcrwell/backend/internal/cache"
	"drcrwell/backend/internal/models"
	"drcrwell/backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateLead creates a new lead
func CreateLead(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
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

	// Invalidate phone cache after creating lead
	tenantID, _ := c.Get("tenant_id")
	cacheKey := fmt.Sprintf("phone_check:%v:%s", tenantID, lead.Phone)
	cache.Delete(cacheKey)

	c.JSON(http.StatusCreated, lead)
}

// GetLeads returns all leads with optional filtering
func GetLeads(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

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
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
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
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
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

	// Use fresh session + empty model + Where to avoid duplicate table error
	if err := db.Session(&gorm.Session{NewDB: true}).Model(&models.Lead{}).Where("id = ?", lead.ID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar lead"})
		return
	}

	// Reload to get updated data using fresh session
	db.Session(&gorm.Session{NewDB: true}).First(&lead, id)
	c.JSON(http.StatusOK, lead)
}

// DeleteLead soft deletes a lead
func DeleteLead(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	id := c.Param("id")

	var lead models.Lead
	if err := db.First(&lead, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead não encontrado"})
		return
	}

	// Use fresh session to avoid context contamination
	if err := db.Session(&gorm.Session{NewDB: true}).Delete(&lead).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir lead"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lead excluído com sucesso"})
}

// CheckLeadByPhone checks if a lead or patient exists by phone number
// This is used by WhatsApp integration to check if contact is known
// Uses Redis cache to reduce database queries (cache TTL: 5 minutes)
// If auto_create=true query param is set, automatically creates a lead for unknown numbers
func CheckLeadByPhone(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	phone := c.Param("phone")
	autoCreate := c.Query("auto_create") == "true"

	// Clean phone number
	phone = cleanPhoneNumber(phone)

	// Get tenant_id for cache key (avoid cross-tenant data leaks)
	tenantID, _ := c.Get("tenant_id")
	cacheKey := fmt.Sprintf("phone_check:%v:%s", tenantID, phone)

	// Try to get from cache first (skip cache if auto_create to ensure fresh check)
	if !autoCreate {
		var cachedResult map[string]interface{}
		if err := cache.Get(cacheKey, &cachedResult); err == nil {
			c.JSON(http.StatusOK, cachedResult)
			return
		}
	}

	// Optimized: Single UNION query to check both patients and leads
	// This reduces 2 database roundtrips to 1
	type PhoneCheckResult struct {
		Type   string `gorm:"column:type"`
		ID     uint   `gorm:"column:id"`
		Name   string `gorm:"column:name"`
		Status string `gorm:"column:status"`
	}

	var queryResult PhoneCheckResult
	// Use regexp_replace to normalize phone numbers in database for comparison
	// This allows matching (11) 98765-4321 with 11987654321
	phonePattern := "%" + phone + "%"
	unionQuery := `
		SELECT 'patient' as type, id, name, '' as status
		FROM patients
		WHERE (regexp_replace(COALESCE(phone, ''), '[^0-9]', '', 'g') LIKE ?
		    OR regexp_replace(COALESCE(cell_phone, ''), '[^0-9]', '', 'g') LIKE ?)
		AND deleted_at IS NULL
		UNION ALL
		SELECT 'lead' as type, id, name, status
		FROM leads
		WHERE regexp_replace(COALESCE(phone, ''), '[^0-9]', '', 'g') LIKE ?
		AND deleted_at IS NULL
		LIMIT 1
	`
	err := db.Raw(unionQuery, phonePattern, phonePattern, phonePattern).Scan(&queryResult).Error

	if err == nil && queryResult.ID > 0 {
		var result gin.H
		if queryResult.Type == "patient" {
			result = gin.H{
				"exists":     true,
				"type":       "patient",
				"id":         queryResult.ID,
				"name":       queryResult.Name,
				"patient_id": queryResult.ID,
			}
		} else {
			result = gin.H{
				"exists":  true,
				"type":    "lead",
				"id":      queryResult.ID,
				"name":    queryResult.Name,
				"lead_id": queryResult.ID,
				"status":  queryResult.Status,
			}
		}
		// Cache the result for 5 minutes
		cache.Set(cacheKey, result, 5*time.Minute)
		c.JSON(http.StatusOK, result)
		return
	}

	// Not found - auto-create lead if requested
	if autoCreate {
		lead := models.Lead{
			Name:          "Contato WhatsApp",
			Phone:         phone,
			Source:        "whatsapp",
			Status:        "new",
			ContactReason: "Primeiro contato via WhatsApp",
			Notes:         fmt.Sprintf("[Auto-criado em %s]", time.Now().Format("02/01/2006 15:04")),
			CreatedBy:     0, // System
		}

		if err := db.Session(&gorm.Session{}).Create(&lead).Error; err == nil {
			// Clear cache since we created a new lead
			cache.Delete(cacheKey)

			result := gin.H{
				"exists":       true,
				"type":         "lead",
				"id":           lead.ID,
				"name":         lead.Name,
				"lead_id":      lead.ID,
				"status":       lead.Status,
				"auto_created": true,
			}
			cache.Set(cacheKey, result, 5*time.Minute)
			c.JSON(http.StatusOK, result)
			return
		}
	}

	// Not found - cache for shorter time (2 minutes) as new contacts may register soon
	result := gin.H{
		"exists": false,
		"type":   "unknown",
	}
	cache.Set(cacheKey, result, 2*time.Minute)
	c.JSON(http.StatusOK, result)
}

// WhatsAppUpdateLead updates a lead with additional information (name, birth_date, contact_reason)
// This is called by the AI after collecting information from the user
// PUT /api/whatsapp/leads/:id
func WhatsAppUpdateLead(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	id := c.Param("id")

	var lead models.Lead
	if err := db.First(&lead, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead não encontrado"})
		return
	}

	var input struct {
		Name          string `json:"name"`
		BirthDate     string `json:"birth_date"`
		ContactReason string `json:"contact_reason"`
		Notes         string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})

	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.ContactReason != "" {
		updates["contact_reason"] = input.ContactReason
	}
	if input.Notes != "" {
		// Append to existing notes
		newNotes := lead.Notes
		if newNotes != "" {
			newNotes += "\n"
		}
		newNotes += input.Notes
		updates["notes"] = newNotes
	}

	// Parse birth date if provided
	if input.BirthDate != "" {
		formats := []string{
			"2006-01-02",  // YYYY-MM-DD
			"02/01/2006",  // DD/MM/YYYY
			"2/1/2006",    // D/M/YYYY
			"02-01-2006",  // DD-MM-YYYY
		}
		for _, format := range formats {
			if parsed, err := time.Parse(format, input.BirthDate); err == nil {
				updates["birth_date"] = parsed
				break
			}
		}
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum campo para atualizar"})
		return
	}

	if err := db.Session(&gorm.Session{NewDB: true}).Model(&models.Lead{}).Where("id = ?", lead.ID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar lead"})
		return
	}

	// Reload lead
	db.Session(&gorm.Session{NewDB: true}).First(&lead, id)

	// Invalidate cache
	tenantID, _ := c.Get("tenant_id")
	cacheKey := fmt.Sprintf("phone_check:%v:%s", tenantID, lead.Phone)
	cache.Delete(cacheKey)

	c.JSON(http.StatusOK, gin.H{
		"message": "Lead atualizado com sucesso",
		"lead":    lead,
	})
}

// ConvertLeadToPatient converts a lead to a patient
func ConvertLeadToPatient(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
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
		Name      string `json:"name"`
		CPF       string `json:"cpf"`
		BirthDate string `json:"birth_date"`
		Address   string `json:"address"`
		City      string `json:"city"`
		State     string `json:"state"`
		ZipCode   string `json:"zip_code"`
		Notes     string `json:"notes"`
		Phone     string `json:"phone"`
	}
	c.ShouldBindJSON(&additionalData)

	// Use name from additionalData if provided, otherwise use lead.Name
	patientName := lead.Name
	if additionalData.Name != "" {
		patientName = additionalData.Name
		// Also update the lead's name for consistency
		db.Session(&gorm.Session{NewDB: true}).Model(&models.Lead{}).Where("id = ?", lead.ID).Update("name", additionalData.Name)
	}

	// Create patient from lead
	patient := models.Patient{
		Name:      patientName,
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

	// Parse birth date if provided - try multiple formats
	if additionalData.BirthDate != "" {
		dateFormats := []string{
			"2006-01-02",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05",
			"02/01/2006",
			"02-01-2006",
		}
		for _, format := range dateFormats {
			if t, err := time.Parse(format, additionalData.BirthDate); err == nil {
				patient.BirthDate = &t
				break
			}
		}
	}

	// Create patient using fresh session to avoid context contamination
	if err := db.Session(&gorm.Session{NewDB: true}).Create(&patient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar paciente: " + err.Error()})
		return
	}

	// Update lead status using empty model + Where to avoid duplicate table error
	now := time.Now()
	updateData := map[string]interface{}{
		"status":                  "converted",
		"converted_to_patient_id": patient.ID,
		"converted_at":            now,
	}

	if err := db.Session(&gorm.Session{NewDB: true}).Model(&models.Lead{}).Where("id = ?", lead.ID).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar lead"})
		return
	}

	// Reload lead with updated data
	db.Session(&gorm.Session{NewDB: true}).First(&lead, lead.ID)

	// Invalidate phone cache after converting lead to patient
	tenantID, _ := c.Get("tenant_id")
	cacheKey := fmt.Sprintf("phone_check:%v:%s", tenantID, lead.Phone)
	cache.Delete(cacheKey)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Lead convertido para paciente com sucesso",
		"lead":       lead,
		"patient":    patient,
		"patient_id": patient.ID,
	})
}

// GetLeadStats returns statistics about leads
func GetLeadStats(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

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
