package handlers

import (
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

	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }
	if err := db.Create(&patient).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar paciente"})
		return
	}

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

	// Filtro de busca
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR cpf ILIKE ?", "%"+search+"%", "%"+search+"%")
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
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	// Verificar se paciente existe
	var count int64
	if err := db.Model(&models.Patient{}).Where("id = ?", id).Count(&count).Error; err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente não encontrado"})
		return
	}

	// Fazer bind dos novos dados
	var input models.Patient
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar paciente"})
		return
	}

	// Buscar paciente atualizado
	var patient models.Patient
	db.First(&patient, id)

	c.JSON(http.StatusOK, gin.H{"patient": patient})
}

// DeletePatient - Deletar paciente (soft delete)
func DeletePatient(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	if err := db.Delete(&models.Patient{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar paciente"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Paciente deletado com sucesso"})
}
