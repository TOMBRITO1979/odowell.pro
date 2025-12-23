package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateTreatmentProtocol creates a new treatment protocol
func CreateTreatmentProtocol(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var protocol models.TreatmentProtocol
	if err := c.ShouldBindJSON(&protocol); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context with safe type assertion
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}
	protocol.CreatedBy = userID

	// Handle empty procedures - use NULL instead of empty string for JSONB
	var procedures interface{}
	if protocol.Procedures == "" {
		procedures = nil
	} else {
		procedures = protocol.Procedures
	}

	// Use raw SQL to create
	createSQL := `
		INSERT INTO treatment_protocols
		(name, description, procedures, duration, cost, active, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	now := time.Now()
	var newID uint

	err := db.Raw(createSQL,
		protocol.Name,
		protocol.Description,
		procedures,
		protocol.Duration,
		protocol.Cost,
		protocol.Active,
		protocol.CreatedBy,
		now,
		now,
	).Scan(&newID).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create protocol", "details": err.Error()})
		return
	}

	// Return the protocol data
	protocol.ID = newID
	protocol.CreatedAt = now
	protocol.UpdatedAt = now

	c.JSON(http.StatusCreated, protocol)
}

// GetTreatmentProtocols retrieves treatment protocols with filters
func GetTreatmentProtocols(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.TreatmentProtocol{})

	// Filters
	if active := c.Query("active"); active != "" {
		query = query.Where("active = ?", active == "true")
	}
	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get protocols
	var protocols []models.TreatmentProtocol
	if err := query.
		Order("name ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&protocols).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch protocols"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"protocols": protocols,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTreatmentProtocol retrieves a single treatment protocol
func GetTreatmentProtocol(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id := c.Param("id")
	var protocol models.TreatmentProtocol

	if err := db.First(&protocol, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Protocol not found"})
		return
	}

	c.JSON(http.StatusOK, protocol)
}

// UpdateTreatmentProtocol updates a treatment protocol
func UpdateTreatmentProtocol(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var protocol models.TreatmentProtocol
	if err := db.First(&protocol, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Protocol not found"})
		return
	}

	var updates models.TreatmentProtocol
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle empty procedures
	var procedures interface{}
	if updates.Procedures == "" {
		procedures = nil
	} else {
		procedures = updates.Procedures
	}

	// Use raw SQL to update
	updateSQL := `
		UPDATE treatment_protocols
		SET name = $1,
		    description = $2,
		    procedures = $3,
		    duration = $4,
		    cost = $5,
		    active = $6,
		    updated_at = $7
		WHERE id = $8
	`

	if err := db.Exec(updateSQL,
		updates.Name,
		updates.Description,
		procedures,
		updates.Duration,
		updates.Cost,
		updates.Active,
		time.Now(),
		id,
	).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update protocol"})
		return
	}

	// Return the updated data
	updates.ID = protocol.ID
	updates.CreatedAt = protocol.CreatedAt
	updates.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, updates)
}

// DeleteTreatmentProtocol soft deletes a treatment protocol
func DeleteTreatmentProtocol(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var protocol models.TreatmentProtocol
	if err := db.First(&protocol, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Protocol not found"})
		return
	}

	if err := db.Delete(&protocol).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete protocol"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Protocol deleted successfully"})
}
