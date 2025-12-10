package handlers

import (
	"drcrwell/backend/internal/helpers"
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
		helpers.AuditAction(c, "create", "medical_records", 0, false, map[string]interface{}{
			"error":      "Failed to create medical record",
			"patient_id": record.PatientID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create medical record"})
		return
	}

	// Load relationships
	db.Preload("Patient").Preload("Dentist").First(&record, record.ID)

	// Log detalhado da criação do prontuário
	helpers.AuditAction(c, "create", "medical_records", record.ID, true, map[string]interface{}{
		"patient_id":   record.PatientID,
		"patient_name": record.Patient.Name,
		"dentist_id":   record.DentistID,
		"type":         record.Type,
		"has_diagnosis": record.Diagnosis != "",
		"has_treatment_plan": record.TreatmentPlan != "",
		"has_procedure": record.ProcedureDone != "",
	})

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
	recordID, _ := strconv.ParseUint(id, 10, 32)
	db := c.MustGet("db").(*gorm.DB)

	var record models.MedicalRecord
	if err := db.Preload("Patient").Preload("Dentist").First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Medical record not found"})
		return
	}

	// Log de acesso ao prontuário (crítico para LGPD - rastrear quem viu dados sensíveis)
	helpers.AuditAction(c, "view", "medical_records", uint(recordID), true, map[string]interface{}{
		"patient_id":   record.PatientID,
		"patient_name": record.Patient.Name,
		"record_type":  record.Type,
	})

	c.JSON(http.StatusOK, gin.H{"record": record})
}

func UpdateMedicalRecord(c *gin.Context) {
	id := c.Param("id")
	recordID, _ := strconv.ParseUint(id, 10, 32)
	db := c.MustGet("db").(*gorm.DB)

	// Buscar prontuário existente para log de antes/depois
	var oldRecord models.MedicalRecord
	if err := db.Preload("Patient").First(&oldRecord, id).Error; err != nil {
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
		helpers.AuditAction(c, "update", "medical_records", uint(recordID), false, map[string]interface{}{
			"error":      "Failed to update record",
			"patient_id": oldRecord.PatientID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record"})
		return
	}

	// Load the updated record with relationships
	var record models.MedicalRecord
	db.Preload("Patient").Preload("Dentist").First(&record, id)

	// Log detalhado da atualização do prontuário
	helpers.AuditAction(c, "update", "medical_records", uint(recordID), true, map[string]interface{}{
		"patient_id":   record.PatientID,
		"patient_name": record.Patient.Name,
		"record_type":  record.Type,
		"changes": map[string]interface{}{
			"diagnosis_changed":       oldRecord.Diagnosis != record.Diagnosis,
			"treatment_plan_changed":  oldRecord.TreatmentPlan != record.TreatmentPlan,
			"procedure_done_changed":  oldRecord.ProcedureDone != record.ProcedureDone,
			"odontogram_changed":      oldRecord.Odontogram != record.Odontogram,
		},
	})

	c.JSON(http.StatusOK, gin.H{"record": record})
}

func DeleteMedicalRecord(c *gin.Context) {
	id := c.Param("id")
	recordID, _ := strconv.ParseUint(id, 10, 32)
	db := c.MustGet("db").(*gorm.DB)

	// Buscar dados do prontuário antes de deletar para log
	var record models.MedicalRecord
	db.Preload("Patient").First(&record, id)

	if err := db.Delete(&models.MedicalRecord{}, id).Error; err != nil {
		helpers.AuditAction(c, "delete", "medical_records", uint(recordID), false, map[string]interface{}{
			"error": "Failed to delete record",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
		return
	}

	// Log detalhado da exclusão do prontuário (crítico - dados médicos)
	helpers.AuditAction(c, "delete", "medical_records", uint(recordID), true, map[string]interface{}{
		"patient_id":   record.PatientID,
		"patient_name": record.Patient.Name,
		"record_type":  record.Type,
		"had_diagnosis": record.Diagnosis != "",
		"had_treatment_plan": record.TreatmentPlan != "",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Medical record deleted successfully"})
}
