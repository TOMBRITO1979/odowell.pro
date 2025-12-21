package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PermanentDeletePatient permanently deletes a patient and all related data
// This is the LGPD "Right to be Forgotten" implementation
// IMPORTANT: This is a destructive operation and cannot be undone
func PermanentDeletePatient(c *gin.Context) {
	// Only admins can perform permanent deletion
	userRole, _ := c.Get("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Apenas administradores podem executar exclusao permanente"})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	patientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de paciente invalido"})
		return
	}

	// Require confirmation token
	var input struct {
		ConfirmationToken string `json:"confirmation_token" binding:"required"`
		Reason            string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token de confirmacao e motivo sao obrigatorios"})
		return
	}

	// Validate confirmation token (should be "DELETE-{patient_id}")
	expectedToken := fmt.Sprintf("DELETE-%d", patientID)
	if input.ConfirmationToken != expectedToken {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token de confirmacao invalido"})
		return
	}

	// Check if patient exists
	var patient models.Patient
	if err := db.Unscoped().First(&patient, patientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente nao encontrado"})
		return
	}

	// Store patient info for audit log before deletion
	patientInfo := map[string]interface{}{
		"id":         patient.ID,
		"name":       patient.Name,
		"cpf":        patient.CPF,
		"email":      patient.Email,
		"phone":      patient.Phone,
		"reason":     input.Reason,
		"deleted_by": c.GetUint("user_id"),
		"deleted_at": time.Now().Format(time.RFC3339),
	}

	// Start transaction for atomic deletion
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao iniciar transacao"})
		return
	}

	// Delete in correct order to avoid FK constraint violations
	// Using Unscoped() to perform hard delete instead of soft delete

	// 1. Delete attachments
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.Attachment{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir anexos"})
		return
	}

	// 2. Delete patient consents
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.PatientConsent{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir consentimentos"})
		return
	}

	// 3. Delete tasks
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.Task{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir tarefas"})
		return
	}

	// 4. Delete waiting list entries
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.WaitingList{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir lista de espera"})
		return
	}

	// 5. Delete prescriptions (linked to medical records)
	var medicalRecordIDs []uint
	tx.Model(&models.MedicalRecord{}).Where("patient_id = ?", patientID).Pluck("id", &medicalRecordIDs)
	if len(medicalRecordIDs) > 0 {
		if err := tx.Unscoped().Where("medical_record_id IN ?", medicalRecordIDs).Delete(&models.Prescription{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir receitas"})
			return
		}
	}

	// 6. Delete medical records
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.MedicalRecord{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir prontuarios"})
		return
	}

	// 7. Delete treatment payments (linked to treatments via budgets)
	var budgetIDs []uint
	tx.Model(&models.Budget{}).Where("patient_id = ?", patientID).Pluck("id", &budgetIDs)
	if len(budgetIDs) > 0 {
		var treatmentIDs []uint
		tx.Model(&models.Treatment{}).Where("budget_id IN ?", budgetIDs).Pluck("id", &treatmentIDs)
		if len(treatmentIDs) > 0 {
			if err := tx.Unscoped().Where("treatment_id IN ?", treatmentIDs).Delete(&models.TreatmentPayment{}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir pagamentos de tratamento"})
				return
			}
		}
		// Delete treatments
		if err := tx.Unscoped().Where("budget_id IN ?", budgetIDs).Delete(&models.Treatment{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir tratamentos"})
			return
		}
	}

	// 8. Delete budgets
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.Budget{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir orcamentos"})
		return
	}

	// 9. Delete payments
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.Payment{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir pagamentos"})
		return
	}

	// 10. Delete exams
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.Exam{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir exames"})
		return
	}

	// 11. Delete appointments
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.Appointment{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir agendamentos"})
		return
	}

	// 12. Delete data requests related to this patient
	if err := tx.Unscoped().Where("patient_id = ?", patientID).Delete(&models.DataRequest{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir solicitacoes LGPD"})
		return
	}

	// 13. Finally, delete the patient
	if err := tx.Unscoped().Delete(&patient).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir paciente"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao confirmar exclusao"})
		return
	}

	// Create audit log entry (in public schema, not affected by tenant transaction)
	patientInfoJSON, _ := json.Marshal(patientInfo)
	auditLog := models.AuditLog{
		UserID:     c.GetUint("user_id"),
		UserEmail:  c.GetString("user_email"),
		UserRole:   userRole.(string),
		Action:     "permanent_delete",
		Resource:   "patients",
		ResourceID: uint(patientID),
		Method:     c.Request.Method,
		Path:       c.Request.URL.Path,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Details:    string(patientInfoJSON),
		Success:    true,
	}

	// Save audit log to public schema
	publicDB := database.GetDB()
	publicDB.Create(&auditLog)

	c.JSON(http.StatusOK, gin.H{
		"message": "Paciente e todos os dados relacionados foram excluidos permanentemente",
		"patient": map[string]interface{}{
			"id":   patient.ID,
			"name": patient.Name,
		},
	})
}

// GetPatientDeletionPreview returns a summary of data that will be deleted
func GetPatientDeletionPreview(c *gin.Context) {
	// Only admins can view deletion preview
	userRole, _ := c.Get("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	patientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de paciente invalido"})
		return
	}

	// Check if patient exists
	var patient models.Patient
	if err := db.Unscoped().First(&patient, patientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente nao encontrado"})
		return
	}

	// Count related records
	var counts struct {
		Appointments   int64 `json:"appointments"`
		MedicalRecords int64 `json:"medical_records"`
		Prescriptions  int64 `json:"prescriptions"`
		Budgets        int64 `json:"budgets"`
		Payments       int64 `json:"payments"`
		Exams          int64 `json:"exams"`
		Consents       int64 `json:"consents"`
		Tasks          int64 `json:"tasks"`
		Attachments    int64 `json:"attachments"`
		WaitingList    int64 `json:"waiting_list"`
		DataRequests   int64 `json:"data_requests"`
	}

	db.Model(&models.Appointment{}).Where("patient_id = ?", patientID).Count(&counts.Appointments)
	db.Model(&models.MedicalRecord{}).Where("patient_id = ?", patientID).Count(&counts.MedicalRecords)
	db.Model(&models.Budget{}).Where("patient_id = ?", patientID).Count(&counts.Budgets)
	db.Model(&models.Payment{}).Where("patient_id = ?", patientID).Count(&counts.Payments)
	db.Model(&models.Exam{}).Where("patient_id = ?", patientID).Count(&counts.Exams)
	db.Model(&models.PatientConsent{}).Where("patient_id = ?", patientID).Count(&counts.Consents)
	db.Model(&models.Task{}).Where("patient_id = ?", patientID).Count(&counts.Tasks)
	db.Model(&models.Attachment{}).Where("patient_id = ?", patientID).Count(&counts.Attachments)
	db.Model(&models.WaitingList{}).Where("patient_id = ?", patientID).Count(&counts.WaitingList)
	db.Model(&models.DataRequest{}).Where("patient_id = ?", patientID).Count(&counts.DataRequests)

	// Count prescriptions through medical records
	var medicalRecordIDs []uint
	db.Model(&models.MedicalRecord{}).Where("patient_id = ?", patientID).Pluck("id", &medicalRecordIDs)
	if len(medicalRecordIDs) > 0 {
		db.Model(&models.Prescription{}).Where("medical_record_id IN ?", medicalRecordIDs).Count(&counts.Prescriptions)
	}

	// Generate confirmation token
	confirmationToken := fmt.Sprintf("DELETE-%d", patientID)

	c.JSON(http.StatusOK, gin.H{
		"patient": map[string]interface{}{
			"id":    patient.ID,
			"name":  patient.Name,
			"cpf":   patient.CPF,
			"email": patient.Email,
			"phone": patient.Phone,
		},
		"data_to_delete": counts,
		"confirmation_token": confirmationToken,
		"warning": "ATENCAO: Esta acao e irreversivel. Todos os dados do paciente serao excluidos permanentemente.",
	})
}

// AnonymizePatient anonymizes patient data instead of deleting
// This keeps the records but removes personally identifiable information
func AnonymizePatient(c *gin.Context) {
	// Only admins can anonymize
	userRole, _ := c.Get("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	patientID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de paciente invalido"})
		return
	}

	var input struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Motivo e obrigatorio"})
		return
	}

	var patient models.Patient
	if err := db.First(&patient, patientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente nao encontrado"})
		return
	}

	// Store original info for audit
	originalInfo := map[string]interface{}{
		"name":   patient.Name,
		"cpf":    patient.CPF,
		"email":  patient.Email,
		"phone":  patient.Phone,
		"reason": input.Reason,
	}

	// Anonymize the patient data
	anonymizedID := fmt.Sprintf("ANON-%d-%d", patientID, time.Now().Unix())

	// Clear all personally identifiable information
	patient.Name = "Paciente Anonimizado"
	patient.CPF = anonymizedID
	patient.RG = ""
	patient.Email = ""
	patient.Phone = ""
	patient.CellPhone = ""       // Also clear cell phone
	patient.Address = ""
	patient.Number = ""          // House/apt number
	patient.Complement = ""
	patient.District = ""
	patient.City = ""
	patient.State = ""
	patient.ZipCode = ""
	patient.InsuranceName = ""   // Insurance info is PII
	patient.InsuranceNumber = ""
	patient.Notes = "Dados anonimizados conforme LGPD em " + time.Now().Format("02/01/2006")
	patient.Tags = ""

	if err := db.Save(&patient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao anonimizar paciente"})
		return
	}

	// Also anonymize patient consents (they contain signature data)
	db.Model(&models.PatientConsent{}).
		Where("patient_id = ?", patientID).
		Updates(map[string]interface{}{
			"signature_data":   nil,
			"signature_ip":     "",
			"legal_guardian":   "",
			"witness_name":     "",
			"witness_document": "",
		})

	// Audit log
	helpers.AuditAction(c, "anonymize", "patients", uint(patientID), true, originalInfo)

	c.JSON(http.StatusOK, gin.H{
		"message": "Paciente anonimizado com sucesso",
		"patient_id": patientID,
	})
}
