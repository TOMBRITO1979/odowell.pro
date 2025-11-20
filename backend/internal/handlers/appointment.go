package handlers

import (
        "drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateAppointment(c *gin.Context) {
	var appointment models.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }
	if err := db.Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create appointment"})
		return
	}

	// Load patient relationship
	db.Preload("Patient").First(&appointment, appointment.ID)

	c.JSON(http.StatusCreated, gin.H{"appointment": appointment})
}

func GetAppointments(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.Appointment{})

	// Filters
	if patientID := c.Query("patient_id"); patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}
	if dentistID := c.Query("dentist_id"); dentistID != "" {
		query = query.Where("dentist_id = ?", dentistID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("start_time >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("start_time <= ?", endDate)
	}

	var total int64
	query.Count(&total)

	var appointments []models.Appointment
	if err := query.Preload("Patient").Offset(offset).Limit(pageSize).Order("start_time ASC").
		Find(&appointments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
	})
}

func GetAppointment(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	var appointment models.Appointment
	if err := db.Preload("Patient").First(&appointment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appointment": appointment})
}

func UpdateAppointment(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	var appointment models.Appointment
	if err := db.First(&appointment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	var input models.Appointment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields directly on the loaded record
	appointment.PatientID = input.PatientID
	appointment.DentistID = input.DentistID
	appointment.StartTime = input.StartTime
	appointment.EndTime = input.EndTime
	appointment.Status = input.Status
	appointment.Procedure = input.Procedure
	appointment.Notes = input.Notes

	if err := db.Save(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment"})
		return
	}

	// Load patient relationship
	db.Preload("Patient").First(&appointment, appointment.ID)

	c.JSON(http.StatusOK, gin.H{"appointment": appointment})
}

func DeleteAppointment(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	if err := db.Delete(&models.Appointment{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete appointment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Appointment deleted successfully"})
}

func UpdateAppointmentStatus(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var appointment models.Appointment
	if err := db.First(&appointment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	appointment.Status = req.Status
	if err := db.Save(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"appointment": appointment})
}
