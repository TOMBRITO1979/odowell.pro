package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateWaitingListEntry adds a patient to the waiting list
func CreateWaitingListEntry(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var entry models.WaitingList
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	entry.CreatedBy = userID.(uint)

	// Set default status if not provided
	if entry.Status == "" {
		entry.Status = "waiting"
	}

	// Set default priority if not provided
	if entry.Priority == "" {
		entry.Priority = "normal"
	}

	// Validate patient exists
	var patient models.Patient
	if err := db.First(&patient, entry.PatientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
		return
	}

	// Validate dentist if provided
	if entry.DentistID != nil {
		var dentist models.User
		if err := db.First(&dentist, *entry.DentistID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Dentist not found"})
			return
		}
	}

	// Use raw SQL to create entry
	createSQL := `
		INSERT INTO waiting_lists
		(patient_id, dentist_id, procedure, preferred_dates, priority, status, notes, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	now := time.Now()
	var newID uint

	// Handle empty preferred_dates - use NULL instead of empty string for JSONB
	var preferredDates interface{}
	if entry.PreferredDates == "" {
		preferredDates = nil
	} else {
		preferredDates = entry.PreferredDates
	}

	err := db.Raw(createSQL,
		entry.PatientID,
		entry.DentistID,
		entry.Procedure,
		preferredDates,
		entry.Priority,
		entry.Status,
		entry.Notes,
		entry.CreatedBy,
		now,
		now,
	).Scan(&newID).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create waiting list entry", "details": err.Error()})
		return
	}

	// Return the entry data we already have
	entry.ID = newID
	entry.CreatedAt = now
	entry.UpdatedAt = now

	c.JSON(http.StatusCreated, entry)
}

// GetWaitingList retrieves waiting list entries with filters
func GetWaitingList(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.WaitingList{})

	// Filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if patientID := c.Query("patient_id"); patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}
	if dentistID := c.Query("dentist_id"); dentistID != "" {
		if dentistID == "any" {
			query = query.Where("dentist_id IS NULL")
		} else {
			query = query.Where("dentist_id = ?", dentistID)
		}
	}
	if procedure := c.Query("procedure"); procedure != "" {
		query = query.Where("procedure ILIKE ?", "%"+procedure+"%")
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get entries
	var entries []models.WaitingList
	if err := query.
		Order("priority DESC, created_at ASC"). // Urgent first, then oldest first
		Offset(offset).
		Limit(pageSize).
		Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch waiting list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries":   entries,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetWaitingListEntry retrieves a single waiting list entry
func GetWaitingListEntry(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id := c.Param("id")
	var entry models.WaitingList

	if err := db.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Waiting list entry not found"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// UpdateWaitingListEntry updates a waiting list entry
func UpdateWaitingListEntry(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var entry models.WaitingList
	if err := db.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Waiting list entry not found"})
		return
	}

	var updates models.WaitingList
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle empty preferred_dates - use NULL instead of empty string for JSONB
	var preferredDates interface{}
	if updates.PreferredDates == "" {
		preferredDates = nil
	} else {
		preferredDates = updates.PreferredDates
	}

	// Use raw SQL to update to avoid GORM issues
	updateSQL := `
		UPDATE waiting_lists
		SET patient_id = $1,
		    dentist_id = $2,
		    procedure = $3,
		    preferred_dates = $4,
		    priority = $5,
		    status = $6,
		    notes = $7,
		    updated_at = $8
		WHERE id = $9
	`

	if err := db.Exec(updateSQL,
		updates.PatientID,
		updates.DentistID,
		updates.Procedure,
		preferredDates,
		updates.Priority,
		updates.Status,
		updates.Notes,
		time.Now(),
		id,
	).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update waiting list entry"})
		return
	}

	// Return the updated data
	updates.ID = entry.ID
	updates.CreatedAt = entry.CreatedAt
	updates.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, updates)
}

// ContactWaitingListEntry marks an entry as contacted
func ContactWaitingListEntry(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id := c.Param("id")

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	now := time.Now()
	contactedBy := userID.(uint)

	updateSQL := `
		UPDATE waiting_lists
		SET status = 'contacted',
		    contacted_at = $1,
		    contacted_by = $2,
		    updated_at = $3
		WHERE id = $4
	`

	if err := db.Exec(updateSQL, now, contactedBy, now, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as contacted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Marked as contacted", "id": id})
}

// ScheduleWaitingListEntry marks an entry as scheduled and links to appointment
func ScheduleWaitingListEntry(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var request struct {
		AppointmentID uint `json:"appointment_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate appointment exists
	var appointment models.Appointment
	if err := db.First(&appointment, request.AppointmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	now := time.Now()

	updateSQL := `
		UPDATE waiting_lists
		SET status = 'scheduled',
		    scheduled_at = $1,
		    appointment_id = $2,
		    updated_at = $3
		WHERE id = $4
	`

	if err := db.Exec(updateSQL, now, request.AppointmentID, now, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as scheduled"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Marked as scheduled", "id": id, "appointment_id": request.AppointmentID})
}

// DeleteWaitingListEntry soft deletes a waiting list entry
func DeleteWaitingListEntry(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var entry models.WaitingList
	if err := db.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Waiting list entry not found"})
		return
	}

	if err := db.Delete(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete waiting list entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Waiting list entry deleted successfully"})
}

// GetWaitingListStats returns statistics about the waiting list
func GetWaitingListStats(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var stats struct {
		TotalWaiting  int64 `json:"total_waiting"`
		TotalUrgent   int64 `json:"total_urgent"`
		TotalNormal   int64 `json:"total_normal"`
		TotalContacted int64 `json:"total_contacted"`
		TotalScheduled int64 `json:"total_scheduled"`
	}

	db.Model(&models.WaitingList{}).Where("status = ?", "waiting").Count(&stats.TotalWaiting)
	db.Model(&models.WaitingList{}).Where("priority = ? AND status = ?", "urgent", "waiting").Count(&stats.TotalUrgent)
	db.Model(&models.WaitingList{}).Where("priority = ? AND status = ?", "normal", "waiting").Count(&stats.TotalNormal)
	db.Model(&models.WaitingList{}).Where("status = ?", "contacted").Count(&stats.TotalContacted)
	db.Model(&models.WaitingList{}).Where("status = ?", "scheduled").Count(&stats.TotalScheduled)

	c.JSON(http.StatusOK, stats)
}
