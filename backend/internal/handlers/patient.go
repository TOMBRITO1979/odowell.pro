package handlers

import (
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// decryptPatientFields decrypts sensitive patient fields after loading from DB
func decryptPatientFields(patient *models.Patient) {
	patient.CPF, _ = helpers.DecryptIfNeeded(patient.CPF)
	patient.RG, _ = helpers.DecryptIfNeeded(patient.RG)
	patient.InsuranceNumber, _ = helpers.DecryptIfNeeded(patient.InsuranceNumber)
}

// decryptPatientsFields decrypts sensitive fields for a slice of patients
func decryptPatientsFields(patients []models.Patient) {
	for i := range patients {
		decryptPatientFields(&patients[i])
	}
}

// CreatePatient - Criar novo paciente
func CreatePatient(c *gin.Context) {
	var patient models.Patient
	if err := c.ShouldBindJSON(&patient); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar campo obrigatório: telefone
	if patient.Phone == "" && patient.CellPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefone é obrigatório"})
		return
	}

	// Encrypt sensitive fields before saving
	patient.CPF, _ = helpers.EncryptIfNeeded(patient.CPF)
	patient.RG, _ = helpers.EncryptIfNeeded(patient.RG)
	patient.InsuranceNumber, _ = helpers.EncryptIfNeeded(patient.InsuranceNumber)

	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }
	if err := db.Create(&patient).Error; err != nil {
		helpers.AuditAction(c, "create", "patients", 0, false, map[string]interface{}{
			"error": "Erro ao criar paciente",
			"name":  patient.Name,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar paciente"})
		return
	}

	// Log detalhado da criação (use original CPF, not encrypted)
	helpers.AuditAction(c, "create", "patients", patient.ID, true, map[string]interface{}{
		"name":  patient.Name,
		"phone": patient.Phone,
	})

	// Decrypt fields before returning to client
	decryptPatientFields(&patient)

	c.JSON(http.StatusCreated, gin.H{"patient": patient})
}

// GetPatients - Listar todos os pacientes
func GetPatients(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var total int64
	var patients []models.Patient

	query := db.Model(&models.Patient{})

	// Filtro de busca (nome, telefone fixo e celular)
	// Note: CPF is encrypted so we cannot search by it with ILIKE
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR phone ILIKE ? OR cell_phone ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	query.Offset(offset).Limit(pageSize).Order("name ASC").Find(&patients)

	// Decrypt sensitive fields before returning
	decryptPatientsFields(patients)

	c.JSON(http.StatusOK, gin.H{
		"patients":  patients,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetPatient - Buscar paciente por ID
func GetPatient(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	var patient models.Patient
	if err := db.First(&patient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente não encontrado"})
		return
	}

	// Decrypt sensitive fields before returning
	decryptPatientFields(&patient)

	c.JSON(http.StatusOK, gin.H{"patient": patient})
}

// UpdatePatient - Atualizar paciente
func UpdatePatient(c *gin.Context) {
	id := c.Param("id")
	patientID, _ := strconv.ParseUint(id, 10, 32)
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	// Buscar dados anteriores para log de auditoria
	var oldPatient models.Patient
	if err := db.First(&oldPatient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente não encontrado"})
		return
	}

	// Fazer bind dos novos dados
	var input models.Patient
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar campo obrigatório: telefone
	if input.Phone == "" && input.CellPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefone é obrigatório"})
		return
	}

	// Encrypt sensitive fields before SQL update (since raw SQL bypasses GORM hooks)
	cpfEncrypted, _ := helpers.EncryptIfNeeded(input.CPF)
	rgEncrypted, _ := helpers.EncryptIfNeeded(input.RG)
	insuranceNumEncrypted, _ := helpers.EncryptIfNeeded(input.InsuranceNumber)

	// UPDATE usando SQL RAW para evitar bug do GORM com soft delete
	sql := `UPDATE patients SET
		name = ?, cpf = ?, rg = ?, birth_date = ?, gender = ?,
		email = ?, phone = ?, cell_phone = ?,
		address = ?, number = ?, complement = ?, district = ?, city = ?, state = ?, zip_code = ?,
		allergies = ?, medications = ?, systemic_diseases = ?, blood_type = ?,
		has_insurance = ?, insurance_name = ?, insurance_number = ?,
		tags = ?, active = ?, notes = ?,
		updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL`

	if err := db.Exec(sql,
		input.Name, cpfEncrypted, rgEncrypted, input.BirthDate, input.Gender,
		input.Email, input.Phone, input.CellPhone,
		input.Address, input.Number, input.Complement, input.District, input.City, input.State, input.ZipCode,
		input.Allergies, input.Medications, input.SystemicDiseases, input.BloodType,
		input.HasInsurance, input.InsuranceName, insuranceNumEncrypted,
		input.Tags, input.Active, input.Notes,
		id,
	).Error; err != nil {
		helpers.AuditAction(c, "update", "patients", uint(patientID), false, map[string]interface{}{
			"error": "Erro ao atualizar paciente",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar paciente"})
		return
	}

	// Buscar paciente atualizado
	var patient models.Patient
	db.First(&patient, id)

	// Decrypt fields before returning and logging
	decryptPatientFields(&patient)

	// Log detalhado da atualização com dados antes/depois
	helpers.AuditAction(c, "update", "patients", uint(patientID), true, map[string]interface{}{
		"before": map[string]interface{}{
			"name":  oldPatient.Name,
			"phone": oldPatient.Phone,
			"email": oldPatient.Email,
		},
		"after": map[string]interface{}{
			"name":  patient.Name,
			"phone": patient.Phone,
			"email": patient.Email,
		},
	})

	c.JSON(http.StatusOK, gin.H{"patient": patient})
}

// DeletePatient - Deletar paciente (soft delete)
func DeletePatient(c *gin.Context) {
	id := c.Param("id")
	patientID, _ := strconv.ParseUint(id, 10, 32)
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	// Buscar dados do paciente para log (usando sessão limpa)
	var patient models.Patient
	if err := db.Session(&gorm.Session{}).First(&patient, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente não encontrado"})
		return
	}

	// Verificar dependências antes de deletar
	dependencies := checkPatientDependencies(db, uint(patientID))
	if len(dependencies) > 0 {
		helpers.AuditAction(c, "delete", "patients", uint(patientID), false, map[string]interface{}{
			"error":        "Paciente possui registros relacionados",
			"dependencies": dependencies,
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        "Não é possível excluir este paciente pois existem registros relacionados",
			"dependencies": dependencies,
			"message":      "Exclua ou transfira os registros relacionados antes de excluir o paciente",
		})
		return
	}

	// Soft delete usando SQL direto para evitar bug do GORM com sessões reutilizadas
	if err := db.Exec("UPDATE patients SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id).Error; err != nil {
		helpers.AuditAction(c, "delete", "patients", uint(patientID), false, map[string]interface{}{
			"error": "Erro ao deletar paciente",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar paciente"})
		return
	}

	// Log detalhado da exclusão
	helpers.AuditAction(c, "delete", "patients", uint(patientID), true, map[string]interface{}{
		"deleted_patient": map[string]interface{}{
			"name":  patient.Name,
			"cpf":   patient.CPF,
			"phone": patient.Phone,
		},
	})

	c.JSON(http.StatusOK, gin.H{"message": "Paciente deletado com sucesso"})
}

// checkPatientDependencies verifica se o paciente possui registros relacionados
func checkPatientDependencies(db *gorm.DB, patientID uint) map[string]int64 {
	dependencies := make(map[string]int64)

	// Verificar agendamentos
	var appointmentsCount int64
	db.Model(&models.Appointment{}).Where("patient_id = ?", patientID).Count(&appointmentsCount)
	if appointmentsCount > 0 {
		dependencies["agendamentos"] = appointmentsCount
	}

	// Verificar prontuários
	var medicalRecordsCount int64
	db.Model(&models.MedicalRecord{}).Where("patient_id = ?", patientID).Count(&medicalRecordsCount)
	if medicalRecordsCount > 0 {
		dependencies["prontuários"] = medicalRecordsCount
	}

	// Verificar prescrições
	var prescriptionsCount int64
	db.Model(&models.Prescription{}).Where("patient_id = ?", patientID).Count(&prescriptionsCount)
	if prescriptionsCount > 0 {
		dependencies["prescrições"] = prescriptionsCount
	}

	// Verificar exames
	var examsCount int64
	db.Model(&models.Exam{}).Where("patient_id = ?", patientID).Count(&examsCount)
	if examsCount > 0 {
		dependencies["exames"] = examsCount
	}

	// Verificar orçamentos
	var budgetsCount int64
	db.Model(&models.Budget{}).Where("patient_id = ?", patientID).Count(&budgetsCount)
	if budgetsCount > 0 {
		dependencies["orçamentos"] = budgetsCount
	}

	// Verificar pagamentos
	var paymentsCount int64
	db.Model(&models.Payment{}).Where("patient_id = ?", patientID).Count(&paymentsCount)
	if paymentsCount > 0 {
		dependencies["pagamentos"] = paymentsCount
	}

	// Verificar tratamentos
	var treatmentsCount int64
	db.Model(&models.Treatment{}).Where("patient_id = ?", patientID).Count(&treatmentsCount)
	if treatmentsCount > 0 {
		dependencies["tratamentos"] = treatmentsCount
	}

	// Verificar anexos
	var attachmentsCount int64
	db.Model(&models.Attachment{}).Where("patient_id = ?", patientID).Count(&attachmentsCount)
	if attachmentsCount > 0 {
		dependencies["anexos"] = attachmentsCount
	}

	// Verificar consentimentos
	var consentsCount int64
	db.Model(&models.PatientConsent{}).Where("patient_id = ?", patientID).Count(&consentsCount)
	if consentsCount > 0 {
		dependencies["consentimentos"] = consentsCount
	}

	return dependencies
}
