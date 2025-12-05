package handlers

import (
	"database/sql"
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

	// Validate patient exists using raw SQL to avoid GORM contamination
	var patientCount int64
	if err := db.Raw("SELECT COUNT(*) FROM patients WHERE id = ? AND deleted_at IS NULL", entry.PatientID).Scan(&patientCount).Error; err != nil || patientCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
		return
	}

	// Validate dentist if provided using raw SQL
	if entry.DentistID != nil {
		var dentistCount int64
		if err := db.Raw("SELECT COUNT(*) FROM public.users WHERE id = ? AND deleted_at IS NULL", *entry.DentistID).Scan(&dentistCount).Error; err != nil || dentistCount == 0 {
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

// WaitingListEntry represents a waiting list entry with patient and dentist info
type WaitingListEntry struct {
	ID             uint       `json:"id"`
	PatientID      uint       `json:"patient_id"`
	DentistID      *uint      `json:"dentist_id,omitempty"`
	Procedure      string     `json:"procedure"`
	PreferredDates string     `json:"preferred_dates,omitempty"`
	Priority       string     `json:"priority"`
	Status         string     `json:"status"`
	ContactedAt    *time.Time `json:"contacted_at,omitempty"`
	ContactedBy    *uint      `json:"contacted_by,omitempty"`
	ScheduledAt    *time.Time `json:"scheduled_at,omitempty"`
	AppointmentID  *uint      `json:"appointment_id,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	CreatedBy      uint       `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Patient        *struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email,omitempty"`
		Phone string `json:"phone,omitempty"`
	} `json:"patient,omitempty"`
	Dentist *struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"dentist,omitempty"`
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

	// Build WHERE conditions
	whereConditions := []string{"wl.deleted_at IS NULL"}
	whereArgs := []interface{}{}

	if status := c.Query("status"); status != "" {
		whereConditions = append(whereConditions, "wl.status = ?")
		whereArgs = append(whereArgs, status)
	}
	if priority := c.Query("priority"); priority != "" {
		whereConditions = append(whereConditions, "wl.priority = ?")
		whereArgs = append(whereArgs, priority)
	}
	if patientID := c.Query("patient_id"); patientID != "" {
		whereConditions = append(whereConditions, "wl.patient_id = ?")
		whereArgs = append(whereArgs, patientID)
	}
	if dentistID := c.Query("dentist_id"); dentistID != "" {
		if dentistID == "any" {
			whereConditions = append(whereConditions, "wl.dentist_id IS NULL")
		} else {
			whereConditions = append(whereConditions, "wl.dentist_id = ?")
			whereArgs = append(whereArgs, dentistID)
		}
	}
	if procedure := c.Query("procedure"); procedure != "" {
		whereConditions = append(whereConditions, "wl.procedure ILIKE ?")
		whereArgs = append(whereArgs, "%"+procedure+"%")
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			whereClause += " AND " + whereConditions[i]
		}
	}

	// Count total
	countSQL := "SELECT COUNT(*) FROM waiting_lists wl " + whereClause
	var total int64
	db.Raw(countSQL, whereArgs...).Scan(&total)

	// Get entries with patient and dentist info
	selectSQL := `
		SELECT
			wl.id, wl.patient_id, wl.dentist_id, wl.procedure, wl.preferred_dates,
			wl.priority, wl.status, wl.contacted_at, wl.contacted_by,
			wl.scheduled_at, wl.appointment_id, wl.notes, wl.created_by,
			wl.created_at, wl.updated_at,
			p.id as patient_db_id, p.name as patient_name, p.email as patient_email, p.phone as patient_phone,
			u.id as dentist_db_id, u.name as dentist_name
		FROM waiting_lists wl
		LEFT JOIN patients p ON p.id = wl.patient_id
		LEFT JOIN public.users u ON u.id = wl.dentist_id
		` + whereClause + `
		ORDER BY
			CASE WHEN wl.priority = 'urgent' THEN 0 ELSE 1 END,
			wl.created_at ASC
		LIMIT ? OFFSET ?
	`

	queryArgs := append(whereArgs, pageSize, offset)

	rows, err := db.Raw(selectSQL, queryArgs...).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch waiting list", "details": err.Error()})
		return
	}
	defer rows.Close()

	var entries []WaitingListEntry
	for rows.Next() {
		var entry WaitingListEntry

		// All fields that can be NULL need sql.Null* types
		var id, patientID, createdBy sql.NullInt64
		var preferredDates, procedure, notes, priority, status sql.NullString
		var dentistID, contactedBy, appointmentID sql.NullInt64
		var contactedAt, scheduledAt, createdAt, updatedAt sql.NullTime

		// Patient info (from LEFT JOIN - could be null if patient deleted)
		var patientDbID sql.NullInt64
		var patientName, patientEmail, patientPhone sql.NullString

		// Dentist info (from LEFT JOIN - null if no dentist specified)
		var dentistDbID sql.NullInt64
		var dentistName sql.NullString

		err := rows.Scan(
			&id, &patientID, &dentistID, &procedure, &preferredDates,
			&priority, &status, &contactedAt, &contactedBy,
			&scheduledAt, &appointmentID, &notes, &createdBy,
			&createdAt, &updatedAt,
			&patientDbID, &patientName, &patientEmail, &patientPhone,
			&dentistDbID, &dentistName,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan waiting list entry", "details": err.Error()})
			return
		}

		// Map required fields
		if id.Valid {
			entry.ID = uint(id.Int64)
		}
		if patientID.Valid {
			entry.PatientID = uint(patientID.Int64)
		}
		if priority.Valid {
			entry.Priority = priority.String
		}
		if status.Valid {
			entry.Status = status.String
		}
		if createdBy.Valid {
			entry.CreatedBy = uint(createdBy.Int64)
		}
		if createdAt.Valid {
			entry.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			entry.UpdatedAt = updatedAt.Time
		}

		// Map optional fields
		if dentistID.Valid {
			did := uint(dentistID.Int64)
			entry.DentistID = &did
		}
		if procedure.Valid {
			entry.Procedure = procedure.String
		}
		if preferredDates.Valid {
			entry.PreferredDates = preferredDates.String
		}
		if contactedAt.Valid {
			entry.ContactedAt = &contactedAt.Time
		}
		if contactedBy.Valid {
			cb := uint(contactedBy.Int64)
			entry.ContactedBy = &cb
		}
		if scheduledAt.Valid {
			entry.ScheduledAt = &scheduledAt.Time
		}
		if appointmentID.Valid {
			aid := uint(appointmentID.Int64)
			entry.AppointmentID = &aid
		}
		if notes.Valid {
			entry.Notes = notes.String
		}

		// Attach patient info
		if patientDbID.Valid && patientName.Valid {
			entry.Patient = &struct {
				ID    uint   `json:"id"`
				Name  string `json:"name"`
				Email string `json:"email,omitempty"`
				Phone string `json:"phone,omitempty"`
			}{
				ID:    uint(patientDbID.Int64),
				Name:  patientName.String,
				Email: patientEmail.String,
				Phone: patientPhone.String,
			}
		}

		// Attach dentist info if exists
		if dentistDbID.Valid && dentistName.Valid {
			entry.Dentist = &struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
			}{
				ID:   uint(dentistDbID.Int64),
				Name: dentistName.String,
			}
		}

		entries = append(entries, entry)
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
