package handlers

import (
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateMedicalRecord(c *gin.Context) {
	var record models.MedicalRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create medical record"})
		return
	}

	// Load relationships
	db.Preload("Patient").Preload("Dentist").First(&record, record.ID)

	c.JSON(http.StatusCreated, gin.H{"record": record})
}

func GetMedicalRecords(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.MedicalRecord{})

	if patientID := c.Query("patient_id"); patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}
	if recordType := c.Query("type"); recordType != "" {
		query = query.Where("type = ?", recordType)
	}

	var total int64
	query.Count(&total)

	var records []models.MedicalRecord
	if err := query.Preload("Patient").Preload("Dentist").
		Offset(offset).Limit(pageSize).Order("created_at DESC").
		Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"records":   records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetMedicalRecord(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var record models.MedicalRecord
	if err := db.Preload("Patient").Preload("Dentist").First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Medical record not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"record": record})
}

func UpdateMedicalRecord(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	// Check if record exists
	var count int64
	if err := db.Model(&models.MedicalRecord{}).Where("id = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Medical record not found"})
		return
	}

	var input models.MedicalRecord
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update using Exec to avoid the duplicate table error
	result := db.Exec(`
		UPDATE medical_records
		SET patient_id = ?, dentist_id = ?, appointment_id = ?, type = ?,
		    odontogram = ?, diagnosis = ?, treatment_plan = ?, procedure_done = ?,
		    materials = ?, prescription = ?, certificate = ?, evolution = ?,
		    notes = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, input.PatientID, input.DentistID, input.AppointmentID, input.Type,
		input.Odontogram, input.Diagnosis, input.TreatmentPlan, input.ProcedureDone,
		input.Materials, input.Prescription, input.Certificate, input.Evolution,
		input.Notes, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record"})
		return
	}

	// Load the updated record with relationships
	var record models.MedicalRecord
	db.Preload("Patient").Preload("Dentist").First(&record, id)

	c.JSON(http.StatusOK, gin.H{"record": record})
}

func DeleteMedicalRecord(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	if err := db.Delete(&models.MedicalRecord{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Medical record deleted successfully"})
}
