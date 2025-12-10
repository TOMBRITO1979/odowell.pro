package handlers

import (
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }
	if err := db.Create(&patient).Error; err != nil {
		helpers.AuditAction(c, "create", "patients", 0, false, map[string]interface{}{
			"error": "Erro ao criar paciente",
			"name":  patient.Name,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar paciente"})
		return
	}

	// Log detalhado da criação
	helpers.AuditAction(c, "create", "patients", patient.ID, true, map[string]interface{}{
		"name":  patient.Name,
		"cpf":   patient.CPF,
		"phone": patient.Phone,
	})

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

	// Filtro de busca (nome, CPF, telefone fixo e celular)
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR cpf ILIKE ? OR phone ILIKE ? OR cell_phone ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)
	query.Offset(offset).Limit(pageSize).Order("name ASC").Find(&patients)

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
		input.Name, input.CPF, input.RG, input.BirthDate, input.Gender,
		input.Email, input.Phone, input.CellPhone,
		input.Address, input.Number, input.Complement, input.District, input.City, input.State, input.ZipCode,
		input.Allergies, input.Medications, input.SystemicDiseases, input.BloodType,
		input.HasInsurance, input.InsuranceName, input.InsuranceNumber,
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

	// Buscar dados do paciente para log
	var patient models.Patient
	db.First(&patient, id)

	if err := db.Delete(&models.Patient{}, id).Error; err != nil {
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
