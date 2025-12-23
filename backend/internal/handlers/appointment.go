package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


// checkAppointmentConflict verifica se existe conflito de horário para o profissional
// Retorna true se existe conflito, false se o horário está livre
func checkAppointmentConflict(db *gorm.DB, dentistID uint, startTime, endTime models.LocalTime, excludeAppointmentID uint) (bool, error) {
	var count int64

	// Usa uma nova sessão para evitar contaminação de condições anteriores
	cleanDB := db.Session(&gorm.Session{NewDB: true})

	// Busca agendamentos que:
	// 1. São do mesmo dentista
	// 2. Não estão cancelados ou no_show
	// 3. Têm sobreposição de horário (start < existingEnd AND end > existingStart)
	// 4. Não é o próprio agendamento (para updates)
	query := cleanDB.Model(&models.Appointment{}).
		Where("dentist_id = ?", dentistID).
		Where("status NOT IN ?", []string{"cancelled", "no_show"}).
		Where("start_time < ? AND end_time > ?", endTime, startTime)

	// Exclui o próprio agendamento (para caso de update)
	if excludeAppointmentID > 0 {
		query = query.Where("id != ?", excludeAppointmentID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func CreateAppointment(c *gin.Context) {
	var appointment models.Appointment
	if err := c.ShouldBindJSON(&appointment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	// Validar conflito de horário para o profissional
	hasConflict, err := checkAppointmentConflict(db, appointment.DentistID, appointment.StartTime, appointment.EndTime, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar disponibilidade"})
		return
	}
	if hasConflict {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Conflito de horário",
			"message": "Já existe um agendamento para este profissional neste horário. Por favor, escolha outro horário.",
		})
		return
	}

	if err := db.Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create appointment"})
		return
	}

	// Load patient and dentist relationships
	db.Preload("Patient").Preload("Dentist").First(&appointment, appointment.ID)

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
	if procedure := c.Query("procedure"); procedure != "" {
		query = query.Where("procedure = ?", procedure)
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
	if err := query.Preload("Patient").Preload("Dentist").Offset(offset).Limit(pageSize).Order("start_time ASC").
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
	if err := db.Preload("Patient").Preload("Dentist").First(&appointment, id).Error; err != nil {
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

	// Validar conflito de horário para o profissional (excluindo o próprio agendamento)
	hasConflict, err := checkAppointmentConflict(db, input.DentistID, input.StartTime, input.EndTime, appointment.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar disponibilidade"})
		return
	}
	if hasConflict {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Conflito de horário",
			"message": "Já existe um agendamento para este profissional neste horário. Por favor, escolha outro horário.",
		})
		return
	}

	// Update fields directly (avoid GORM FROM clause issue)
	appointment.PatientID = input.PatientID
	appointment.DentistID = input.DentistID
	appointment.StartTime = input.StartTime
	appointment.EndTime = input.EndTime
	appointment.Status = input.Status
	appointment.Type = input.Type
	appointment.Procedure = input.Procedure
	appointment.Notes = input.Notes
	appointment.Room = input.Room
	appointment.Confirmed = input.Confirmed
	appointment.IsRecurring = input.IsRecurring
	appointment.RecurrenceRule = input.RecurrenceRule

	// Use raw SQL to avoid GORM's FROM clause bug
	sql := `UPDATE appointments
		SET patient_id = ?, dentist_id = ?, start_time = ?, end_time = ?,
			status = ?, type = ?, procedure = ?, notes = ?, room = ?, confirmed = ?,
			is_recurring = ?, recurrence_rule = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL`

	if err := db.Exec(sql,
		appointment.PatientID, appointment.DentistID,
		appointment.StartTime, appointment.EndTime,
		appointment.Status, appointment.Type,
		appointment.Procedure, appointment.Notes,
		appointment.Room, appointment.Confirmed, appointment.IsRecurring,
		appointment.RecurrenceRule, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment"})
		return
	}

	// Reload with patient and dentist relationships
	db.Preload("Patient").Preload("Dentist").First(&appointment, id)

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
