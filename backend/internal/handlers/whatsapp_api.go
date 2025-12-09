package handlers

import (
	"drcrwell/backend/internal/models"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getTimezone returns the application timezone location
// Falls back to America/Sao_Paulo if TZ env var is not set
func getTimezone() *time.Location {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "America/Sao_Paulo"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Local // Fallback to system local
	}
	return loc
}

// WhatsAppVerifyRequest represents the identity verification request
type WhatsAppVerifyRequest struct {
	CPF       string `json:"cpf" binding:"required"`
	BirthDate string `json:"birth_date" binding:"required"` // Format: YYYY-MM-DD or DD/MM/YYYY
}

// WhatsAppVerifyResponse represents the identity verification response
type WhatsAppVerifyResponse struct {
	Valid     bool   `json:"valid"`
	PatientID uint   `json:"patient_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Token     string `json:"token,omitempty"` // Session token for subsequent requests
	Message   string `json:"message"`
}

// WhatsAppAppointmentResponse represents appointment data for WhatsApp
type WhatsAppAppointmentResponse struct {
	ID          uint   `json:"id"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	EndTime     string `json:"end_time"`
	Procedure   string `json:"procedure"`
	Status      string `json:"status"`
	StatusLabel string `json:"status_label"`
	DentistName string `json:"dentist_name"`
	Room        string `json:"room,omitempty"`
	CanCancel   bool   `json:"can_cancel"`
	CanReschedule bool `json:"can_reschedule"`
}

// WhatsAppCancelRequest represents a cancellation request
type WhatsAppCancelRequest struct {
	AppointmentID uint   `json:"appointment_id" binding:"required"`
	Reason        string `json:"reason"`
}

// WhatsAppRescheduleRequest represents a reschedule request
type WhatsAppRescheduleRequest struct {
	AppointmentID   uint   `json:"appointment_id" binding:"required"`
	NewDate         string `json:"new_date" binding:"required"`         // Format: YYYY-MM-DD
	NewTime         string `json:"new_time" binding:"required"`         // Format: HH:MM
	PreferredDentist *uint `json:"preferred_dentist_id,omitempty"`
}

// WhatsAppWaitingListRequest represents a waiting list entry request
type WhatsAppWaitingListRequest struct {
	Procedure      string   `json:"procedure" binding:"required"`
	PreferredDates []string `json:"preferred_dates,omitempty"` // Array of dates in YYYY-MM-DD format
	Priority       string   `json:"priority,omitempty"`        // normal, urgent
	Notes          string   `json:"notes,omitempty"`
	DentistID      *uint    `json:"dentist_id,omitempty"`
}

// WhatsAppAvailableSlotsRequest represents a request for available slots
type WhatsAppAvailableSlotsRequest struct {
	Date      string `json:"date" binding:"required"` // Format: YYYY-MM-DD
	DentistID *uint  `json:"dentist_id,omitempty"`
	Procedure string `json:"procedure,omitempty"`
}

// WhatsAppAvailableSlot represents an available time slot
type WhatsAppAvailableSlot struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	DentistID uint   `json:"dentist_id"`
	DentistName string `json:"dentist_name"`
}

// Helper function to normalize CPF (remove dots, dashes, spaces)
func normalizeCPF(cpf string) string {
	cpf = strings.ReplaceAll(cpf, ".", "")
	cpf = strings.ReplaceAll(cpf, "-", "")
	cpf = strings.ReplaceAll(cpf, " ", "")
	return strings.TrimSpace(cpf)
}

// Helper function to parse birth date from multiple formats
func parseBirthDate(dateStr string) (*time.Time, error) {
	formats := []string{
		"2006-01-02",  // YYYY-MM-DD
		"02/01/2006",  // DD/MM/YYYY
		"02-01-2006",  // DD-MM-YYYY
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("invalid date format")
}

// Helper function to get status label in Portuguese (for WhatsApp API)
func getAppointmentStatusLabel(status string) string {
	labels := map[string]string{
		"scheduled":   "Agendado",
		"confirmed":   "Confirmado",
		"in_progress": "Em Atendimento",
		"completed":   "Concluído",
		"cancelled":   "Cancelado",
		"no_show":     "Faltou",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return status
}

// Helper function to get procedure label in Portuguese
func getProcedureLabel(procedure string) string {
	labels := map[string]string{
		"consultation": "Consulta",
		"cleaning":     "Limpeza",
		"filling":      "Restauração",
		"extraction":   "Extração",
		"root_canal":   "Canal",
		"orthodontics": "Ortodontia",
		"whitening":    "Clareamento",
		"prosthesis":   "Prótese",
		"implant":      "Implante",
		"emergency":    "Emergência",
		"other":        "Outro",
	}
	if label, ok := labels[procedure]; ok {
		return label
	}
	return procedure
}

// WhatsAppVerifyIdentity verifies patient identity by CPF and birth date
// POST /api/whatsapp/verify
func WhatsAppVerifyIdentity(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req WhatsAppVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, WhatsAppVerifyResponse{
			Valid:   false,
			Message: "CPF e data de nascimento são obrigatórios",
		})
		return
	}

	// Normalize CPF
	cpf := normalizeCPF(req.CPF)
	if len(cpf) != 11 {
		c.JSON(http.StatusBadRequest, WhatsAppVerifyResponse{
			Valid:   false,
			Message: "CPF inválido. O CPF deve ter 11 dígitos.",
		})
		return
	}

	// Parse birth date
	birthDate, err := parseBirthDate(req.BirthDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, WhatsAppVerifyResponse{
			Valid:   false,
			Message: "Formato de data de nascimento inválido. Use DD/MM/AAAA ou AAAA-MM-DD.",
		})
		return
	}

	// Find patient by CPF
	var patient models.Patient
	result := db.Where("REPLACE(REPLACE(cpf, '.', ''), '-', '') = ? AND active = ?", cpf, true).First(&patient)

	if result.Error != nil {
		c.JSON(http.StatusOK, WhatsAppVerifyResponse{
			Valid:   false,
			Message: "Paciente não encontrado. Verifique o CPF informado.",
		})
		return
	}

	// Verify birth date matches
	if patient.BirthDate == nil {
		c.JSON(http.StatusOK, WhatsAppVerifyResponse{
			Valid:   false,
			Message: "Data de nascimento não cadastrada. Entre em contato com a clínica.",
		})
		return
	}

	// Compare only date part (ignore time)
	patientBirthDate := patient.BirthDate.Format("2006-01-02")
	requestBirthDate := birthDate.Format("2006-01-02")

	if patientBirthDate != requestBirthDate {
		c.JSON(http.StatusOK, WhatsAppVerifyResponse{
			Valid:   false,
			Message: "Data de nascimento não confere. Verifique os dados informados.",
		})
		return
	}

	// Generate a simple session token (in production, use JWT or similar)
	token := fmt.Sprintf("%d-%d", patient.ID, time.Now().Unix())

	c.JSON(http.StatusOK, WhatsAppVerifyResponse{
		Valid:     true,
		PatientID: patient.ID,
		Name:      patient.Name,
		Token:     token,
		Message:   fmt.Sprintf("Olá, %s! Identidade verificada com sucesso.", patient.Name),
	})
}

// normalizePhone removes formatting from phone number
func normalizePhone(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "+", "")
	return strings.TrimSpace(phone)
}

// WhatsAppGetAppointments returns patient's upcoming appointments
// GET /api/whatsapp/appointments?patient_id=X
// GET /api/whatsapp/appointments?phone=11999998888
func WhatsAppGetAppointments(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	patientID := c.Query("patient_id")
	phone := c.Query("phone")

	// If phone is provided, find patient by phone first
	if patientID == "" && phone != "" {
		normalizedPhone := normalizePhone(phone)
		if normalizedPhone == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   true,
				"message": "Telefone inválido",
			})
			return
		}

		// Get schema name for explicit table reference
		schemaName, _ := c.Get("schema")
		patientsTable := "patients"
		if schemaName != nil {
			patientsTable = fmt.Sprintf("%s.patients", schemaName)
		}

		// Search patient by phone or cell_phone using normalized comparison
		var patient models.Patient
		// Use simple LIKE search for better compatibility
		phonePattern := "%" + normalizedPhone + "%"
		err := db.Table(patientsTable).
			Where("phone LIKE ? OR cell_phone LIKE ?", phonePattern, phonePattern).
			Where("active = ?", true).
			First(&patient).Error

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   true,
				"message": "Paciente não encontrado com este telefone",
			})
			return
		}

		patientID = fmt.Sprintf("%d", patient.ID)
	}

	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Informe patient_id ou phone para buscar agendamentos",
		})
		return
	}

	// Get upcoming appointments (today onwards, excluding cancelled)
	var appointments []models.Appointment
	today := time.Now().Truncate(24 * time.Hour)

	// Get schema name for explicit table reference
	schemaName, _ := c.Get("schema")
	tableName := "appointments"
	if schemaName != nil {
		tableName = fmt.Sprintf("%s.appointments", schemaName)
	}

	err := db.Table(tableName).
		Where("patient_id = ? AND start_time >= ? AND status NOT IN (?)",
			patientID, today, []string{"cancelled"}).
		Preload("Dentist").
		Order("start_time ASC").
		Find(&appointments).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao buscar agendamentos",
		})
		return
	}

	// Convert to response format
	response := make([]WhatsAppAppointmentResponse, 0)
	for _, apt := range appointments {
		dentistName := "Não definido"
		if apt.Dentist != nil {
			dentistName = apt.Dentist.Name
		}

		// Can cancel/reschedule only if appointment is in the future and not completed
		canModify := apt.Status == "scheduled" || apt.Status == "confirmed"

		response = append(response, WhatsAppAppointmentResponse{
			ID:          apt.ID,
			Date:        apt.StartTime.Format("02/01/2006"),
			Time:        apt.StartTime.Format("15:04"),
			EndTime:     apt.EndTime.Format("15:04"),
			Procedure:   getProcedureLabel(apt.Procedure),
			Status:      apt.Status,
			StatusLabel: getAppointmentStatusLabel(apt.Status),
			DentistName: dentistName,
			Room:        apt.Room,
			CanCancel:   canModify,
			CanReschedule: canModify,
		})
	}

	if len(response) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"error":        false,
			"appointments": response,
			"message":      "Você não possui consultas agendadas.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":        false,
		"appointments": response,
		"total":        len(response),
		"message":      fmt.Sprintf("Você possui %d consulta(s) agendada(s).", len(response)),
	})
}

// WhatsAppGetAppointmentHistory returns patient's past appointments
// GET /api/whatsapp/appointments/history?patient_id=X&limit=10
func WhatsAppGetAppointmentHistory(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	patientID := c.Query("patient_id")
	limit := c.DefaultQuery("limit", "10")

	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID do paciente é obrigatório",
		})
		return
	}

	// Get past appointments
	var appointments []models.Appointment
	today := time.Now().Truncate(24 * time.Hour)

	err := db.Where("patient_id = ? AND start_time < ?", patientID, today).
		Preload("Dentist").
		Order("start_time DESC").
		Limit(10).
		Find(&appointments).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao buscar histórico",
		})
		return
	}

	// Convert to response format
	response := make([]WhatsAppAppointmentResponse, 0)
	for _, apt := range appointments {
		dentistName := "Não definido"
		if apt.Dentist != nil {
			dentistName = apt.Dentist.Name
		}

		response = append(response, WhatsAppAppointmentResponse{
			ID:          apt.ID,
			Date:        apt.StartTime.Format("02/01/2006"),
			Time:        apt.StartTime.Format("15:04"),
			EndTime:     apt.EndTime.Format("15:04"),
			Procedure:   getProcedureLabel(apt.Procedure),
			Status:      apt.Status,
			StatusLabel: getAppointmentStatusLabel(apt.Status),
			DentistName: dentistName,
			Room:        apt.Room,
			CanCancel:   false,
			CanReschedule: false,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"error":        false,
		"appointments": response,
		"total":        len(response),
		"limit":        limit,
	})
}

// WhatsAppCancelAppointment cancels an appointment
// POST /api/whatsapp/appointments/cancel
func WhatsAppCancelAppointment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req WhatsAppCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID do agendamento é obrigatório",
		})
		return
	}

	patientID := c.Query("patient_id")
	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID do paciente é obrigatório",
		})
		return
	}

	// Get appointment and verify it belongs to the patient
	var appointment models.Appointment
	err := db.Where("id = ? AND patient_id = ?", req.AppointmentID, patientID).First(&appointment).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Agendamento não encontrado ou não pertence a este paciente",
		})
		return
	}

	// Check if can be cancelled
	if appointment.Status != "scheduled" && appointment.Status != "confirmed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": fmt.Sprintf("Este agendamento não pode ser cancelado. Status atual: %s", getAppointmentStatusLabel(appointment.Status)),
		})
		return
	}

	// Check if appointment is in the future
	if appointment.StartTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Não é possível cancelar agendamentos passados",
		})
		return
	}

	// Update appointment status
	notes := appointment.Notes
	if req.Reason != "" {
		if notes != "" {
			notes += "\n"
		}
		notes += fmt.Sprintf("[Cancelado via WhatsApp em %s] Motivo: %s", time.Now().Format("02/01/2006 15:04"), req.Reason)
	} else {
		if notes != "" {
			notes += "\n"
		}
		notes += fmt.Sprintf("[Cancelado via WhatsApp em %s]", time.Now().Format("02/01/2006 15:04"))
	}

	err = db.Model(&appointment).Updates(map[string]interface{}{
		"status": "cancelled",
		"notes":  notes,
	}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao cancelar agendamento",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": fmt.Sprintf("Agendamento do dia %s às %s foi cancelado com sucesso.",
			appointment.StartTime.Format("02/01/2006"),
			appointment.StartTime.Format("15:04")),
	})
}

// WhatsAppGetAvailableSlots returns available time slots for a specific date
// GET /api/whatsapp/slots?date=YYYY-MM-DD&dentist_id=X
func WhatsAppGetAvailableSlots(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	dateStr := c.Query("date")
	dentistID := c.Query("dentist_id")

	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Data é obrigatória",
		})
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Formato de data inválido. Use AAAA-MM-DD",
		})
		return
	}

	// Check if date is in the future
	if date.Before(time.Now().Truncate(24 * time.Hour)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Não é possível verificar horários de datas passadas",
		})
		return
	}

	// Get clinic settings for working hours
	var settings models.TenantSettings
	db.Table("public.tenant_settings").First(&settings)

	workingStart := "08:00"
	workingEnd := "18:00"
	slotDuration := 30 // minutes

	if settings.WorkingHoursStart != "" {
		workingStart = settings.WorkingHoursStart
	}
	if settings.WorkingHoursEnd != "" {
		workingEnd = settings.WorkingHoursEnd
	}
	if settings.DefaultAppointmentDuration > 0 {
		slotDuration = settings.DefaultAppointmentDuration
	}

	// Get all dentists or specific one
	var dentists []models.User
	dentistQuery := db.Table("public.users").Where("role IN (?) AND active = ?", []string{"dentist", "admin"}, true)
	if dentistID != "" {
		dentistQuery = dentistQuery.Where("id = ?", dentistID)
	}
	dentistQuery.Find(&dentists)

	if len(dentists) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"error":   false,
			"slots":   []WhatsAppAvailableSlot{},
			"message": "Nenhum profissional disponível",
		})
		return
	}

	// Get existing appointments for the date
	startOfDay := date
	endOfDay := date.Add(24 * time.Hour)

	var existingAppointments []models.Appointment
	db.Where("start_time >= ? AND start_time < ? AND status NOT IN (?)",
		startOfDay, endOfDay, []string{"cancelled"}).
		Find(&existingAppointments)

	// Create a map of busy slots per dentist
	busySlots := make(map[uint]map[string]bool)
	for _, apt := range existingAppointments {
		if busySlots[apt.DentistID] == nil {
			busySlots[apt.DentistID] = make(map[string]bool)
		}
		// Mark all slots during the appointment as busy
		current := apt.StartTime
		for current.Before(apt.EndTime) {
			busySlots[apt.DentistID][current.Format("15:04")] = true
			current = current.Add(time.Duration(slotDuration) * time.Minute)
		}
	}

	// Parse working hours
	startTime, _ := time.Parse("15:04", workingStart)
	endTime, _ := time.Parse("15:04", workingEnd)

	// Generate available slots
	availableSlots := make([]WhatsAppAvailableSlot, 0)

	loc := getTimezone()
	for _, dentist := range dentists {
		current := time.Date(date.Year(), date.Month(), date.Day(),
			startTime.Hour(), startTime.Minute(), 0, 0, loc)
		endOfWork := time.Date(date.Year(), date.Month(), date.Day(),
			endTime.Hour(), endTime.Minute(), 0, 0, loc)

		for current.Before(endOfWork) {
			slotTime := current.Format("15:04")

			// Check if slot is not busy and is in the future
			if !busySlots[dentist.ID][slotTime] && current.After(time.Now()) {
				slotEnd := current.Add(time.Duration(slotDuration) * time.Minute)
				availableSlots = append(availableSlots, WhatsAppAvailableSlot{
					Date:        date.Format("02/01/2006"),
					StartTime:   slotTime,
					EndTime:     slotEnd.Format("15:04"),
					DentistID:   dentist.ID,
					DentistName: dentist.Name,
				})
			}

			current = current.Add(time.Duration(slotDuration) * time.Minute)
		}
	}

	if len(availableSlots) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"error":   false,
			"slots":   availableSlots,
			"message": fmt.Sprintf("Não há horários disponíveis para %s", date.Format("02/01/2006")),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"slots":   availableSlots,
		"total":   len(availableSlots),
		"date":    date.Format("02/01/2006"),
		"message": fmt.Sprintf("%d horário(s) disponível(is) para %s", len(availableSlots), date.Format("02/01/2006")),
	})
}

// WhatsAppRescheduleAppointment reschedules an appointment to a new date/time
// POST /api/whatsapp/appointments/reschedule
func WhatsAppRescheduleAppointment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req WhatsAppRescheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Dados incompletos. Informe o ID do agendamento, nova data e horário.",
		})
		return
	}

	patientID := c.Query("patient_id")
	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID do paciente é obrigatório",
		})
		return
	}

	// Get appointment and verify it belongs to the patient
	var appointment models.Appointment
	err := db.Where("id = ? AND patient_id = ?", req.AppointmentID, patientID).
		Preload("Dentist").
		First(&appointment).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Agendamento não encontrado ou não pertence a este paciente",
		})
		return
	}

	// Check if can be rescheduled
	if appointment.Status != "scheduled" && appointment.Status != "confirmed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": fmt.Sprintf("Este agendamento não pode ser remarcado. Status atual: %s", getAppointmentStatusLabel(appointment.Status)),
		})
		return
	}

	// Parse new date and time
	newDate, err := time.Parse("2006-01-02", req.NewDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Formato de data inválido. Use AAAA-MM-DD",
		})
		return
	}

	newTime, err := time.Parse("15:04", req.NewTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Formato de horário inválido. Use HH:MM",
		})
		return
	}

	// Combine date and time using application timezone
	loc := getTimezone()
	newStartTime := time.Date(newDate.Year(), newDate.Month(), newDate.Day(),
		newTime.Hour(), newTime.Minute(), 0, 0, loc)

	// Check if new time is in the future
	if newStartTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Não é possível remarcar para uma data/hora passada",
		})
		return
	}

	// Calculate duration from original appointment
	duration := appointment.EndTime.Sub(appointment.StartTime)
	newEndTime := newStartTime.Add(duration)

	// Determine dentist for new appointment
	dentistID := appointment.DentistID
	if req.PreferredDentist != nil {
		dentistID = *req.PreferredDentist
	}

	// Check if slot is available
	var conflictCount int64
	db.Model(&models.Appointment{}).
		Where("dentist_id = ? AND id != ? AND status NOT IN (?)", dentistID, appointment.ID, []string{"cancelled"}).
		Where("(start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?)",
			newEndTime, newStartTime, newEndTime, newStartTime, newStartTime, newEndTime).
		Count(&conflictCount)

	if conflictCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   true,
			"message": "O horário selecionado não está disponível. Por favor, escolha outro horário.",
		})
		return
	}

	// Store old time for the message
	oldDate := appointment.StartTime.Format("02/01/2006")
	oldTime := appointment.StartTime.Format("15:04")

	// Update appointment
	notes := appointment.Notes
	if notes != "" {
		notes += "\n"
	}
	notes += fmt.Sprintf("[Remarcado via WhatsApp em %s] De %s %s para %s %s",
		time.Now().Format("02/01/2006 15:04"),
		oldDate, oldTime,
		newStartTime.Format("02/01/2006"), newStartTime.Format("15:04"))

	err = db.Model(&appointment).Updates(map[string]interface{}{
		"start_time": newStartTime,
		"end_time":   newEndTime,
		"dentist_id": dentistID,
		"status":     "scheduled",
		"confirmed":  false,
		"notes":      notes,
	}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao remarcar agendamento",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": fmt.Sprintf("Agendamento remarcado com sucesso! Nova data: %s às %s",
			newStartTime.Format("02/01/2006"),
			newStartTime.Format("15:04")),
		"new_date": newStartTime.Format("02/01/2006"),
		"new_time": newStartTime.Format("15:04"),
	})
}

// WhatsAppAddToWaitingList adds patient to waiting list
// POST /api/whatsapp/waiting-list
func WhatsAppAddToWaitingList(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req WhatsAppWaitingListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Procedimento é obrigatório",
		})
		return
	}

	patientID := c.Query("patient_id")
	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID do paciente é obrigatório",
		})
		return
	}

	// Check if patient exists (using fresh session to avoid GORM contamination)
	var patient models.Patient
	if err := db.Session(&gorm.Session{}).First(&patient, patientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Paciente não encontrado",
		})
		return
	}

	// Check if patient is already on waiting list for the same procedure (using fresh session)
	var existingEntry models.WaitingList
	result := db.Session(&gorm.Session{}).Where("patient_id = ? AND procedure = ? AND status = ?", patientID, req.Procedure, "waiting").First(&existingEntry)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   true,
			"message": fmt.Sprintf("Você já está na lista de espera para %s", getProcedureLabel(req.Procedure)),
		})
		return
	}

	// Set default priority
	priority := "normal"
	if req.Priority == "urgent" {
		priority = "urgent"
	}

	// Create waiting list entry
	entry := models.WaitingList{
		PatientID:  patient.ID,
		Procedure:  req.Procedure,
		Priority:   priority,
		Status:     "waiting",
		Notes:      req.Notes,
		CreatedBy:  0, // System/WhatsApp
	}

	if req.DentistID != nil {
		entry.DentistID = req.DentistID
	}

	// Convert preferred dates to valid JSON (PostgreSQL JSONB requires valid JSON, not empty string)
	if len(req.PreferredDates) > 0 {
		// Convert array to JSON format
		jsonDates := "["
		for i, date := range req.PreferredDates {
			if i > 0 {
				jsonDates += ","
			}
			jsonDates += fmt.Sprintf("\"%s\"", date)
		}
		jsonDates += "]"
		entry.PreferredDates = jsonDates
	} else {
		entry.PreferredDates = "[]"
	}

	// Add note about WhatsApp origin
	if entry.Notes != "" {
		entry.Notes += "\n"
	}
	entry.Notes += fmt.Sprintf("[Adicionado via WhatsApp em %s]", time.Now().Format("02/01/2006 15:04"))

	// Use fresh session for create to avoid GORM contamination
	if err := db.Session(&gorm.Session{}).Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao adicionar à lista de espera",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": fmt.Sprintf("Você foi adicionado à lista de espera para %s. Entraremos em contato assim que houver disponibilidade.",
			getProcedureLabel(req.Procedure)),
		"entry_id": entry.ID,
		"priority": priority,
	})
}

// WhatsAppGetWaitingListStatus returns patient's waiting list entries
// GET /api/whatsapp/waiting-list?patient_id=X
func WhatsAppGetWaitingListStatus(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	patientID := c.Query("patient_id")

	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID do paciente é obrigatório",
		})
		return
	}

	var entries []models.WaitingList
	err := db.Where("patient_id = ? AND status IN (?)", patientID, []string{"waiting", "contacted"}).
		Order("created_at DESC").
		Find(&entries).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao buscar lista de espera",
		})
		return
	}

	type WaitingListResponse struct {
		ID           uint   `json:"id"`
		Procedure    string `json:"procedure"`
		Priority     string `json:"priority"`
		PriorityLabel string `json:"priority_label"`
		Status       string `json:"status"`
		StatusLabel  string `json:"status_label"`
		CreatedAt    string `json:"created_at"`
	}

	response := make([]WaitingListResponse, 0)
	for _, entry := range entries {
		priorityLabel := "Normal"
		if entry.Priority == "urgent" {
			priorityLabel = "Urgente"
		}

		statusLabel := "Aguardando"
		if entry.Status == "contacted" {
			statusLabel = "Contatado"
		}

		response = append(response, WaitingListResponse{
			ID:           entry.ID,
			Procedure:    getProcedureLabel(entry.Procedure),
			Priority:     entry.Priority,
			PriorityLabel: priorityLabel,
			Status:       entry.Status,
			StatusLabel:  statusLabel,
			CreatedAt:    entry.CreatedAt.Format("02/01/2006"),
		})
	}

	if len(response) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"error":   false,
			"entries": response,
			"message": "Você não está em nenhuma lista de espera.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"entries": response,
		"total":   len(response),
	})
}

// WhatsAppRemoveFromWaitingList removes patient from waiting list
// DELETE /api/whatsapp/waiting-list/:id?patient_id=X
func WhatsAppRemoveFromWaitingList(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	entryID := c.Param("id")
	patientID := c.Query("patient_id")

	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "ID do paciente é obrigatório",
		})
		return
	}

	// Find and verify entry belongs to patient
	var entry models.WaitingList
	err := db.Where("id = ? AND patient_id = ?", entryID, patientID).First(&entry).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Entrada não encontrada na lista de espera",
		})
		return
	}

	// Update status to cancelled
	err = db.Model(&entry).Update("status", "cancelled").Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao remover da lista de espera",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": fmt.Sprintf("Você foi removido da lista de espera para %s", getProcedureLabel(entry.Procedure)),
	})
}

// WhatsAppGetProcedures returns available procedures list
// GET /api/whatsapp/procedures
func WhatsAppGetProcedures(c *gin.Context) {
	procedures := []map[string]string{
		{"value": "consultation", "label": "Consulta"},
		{"value": "cleaning", "label": "Limpeza"},
		{"value": "filling", "label": "Restauração"},
		{"value": "extraction", "label": "Extração"},
		{"value": "root_canal", "label": "Canal"},
		{"value": "orthodontics", "label": "Ortodontia"},
		{"value": "whitening", "label": "Clareamento"},
		{"value": "prosthesis", "label": "Prótese"},
		{"value": "implant", "label": "Implante"},
		{"value": "emergency", "label": "Emergência"},
		{"value": "other", "label": "Outro"},
	}

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"procedures": procedures,
	})
}

// WhatsAppGetDentists returns available dentists list
// GET /api/whatsapp/dentists
func WhatsAppGetDentists(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type DentistResponse struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		Specialty string `json:"specialty,omitempty"`
	}

	var dentists []DentistResponse
	err := db.Table("public.users").
		Select("id, name, specialty").
		Where("role IN (?) AND active = ?", []string{"dentist", "admin"}, true).
		Order("name ASC").
		Find(&dentists).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao buscar profissionais",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":    false,
		"dentists": dentists,
		"total":    len(dentists),
	})
}

// WhatsAppCreateAppointmentRequest represents an appointment creation request via WhatsApp API
type WhatsAppCreateAppointmentRequest struct {
	PatientID uint   `json:"patient_id" binding:"required"`
	DentistID uint   `json:"dentist_id" binding:"required"`
	Procedure string `json:"procedure"`
	Date      string `json:"date" binding:"required"` // Format: YYYY-MM-DD
	Time      string `json:"time" binding:"required"` // Format: HH:MM
	Notes     string `json:"notes"`
	Duration  int    `json:"duration"` // Duration in minutes (optional, default 30)
}

// WhatsAppCreateAppointmentResponse represents the response for appointment creation
type WhatsAppCreateAppointmentResponse struct {
	ID          uint   `json:"id"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Procedure   string `json:"procedure"`
	DentistName string `json:"dentist_name"`
	Status      string `json:"status"`
}

// WhatsAppCreateLeadRequest represents a lead creation request via WhatsApp API
type WhatsAppCreateLeadRequest struct {
	Name          string `json:"name" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	Email         string `json:"email"`
	Source        string `json:"source"`
	ContactReason string `json:"contact_reason"`
	Notes         string `json:"notes"`
}

// WhatsAppCreateLead creates a new lead via WhatsApp API (without user authentication)
// POST /api/whatsapp/leads
func WhatsAppCreateLead(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req WhatsAppCreateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Nome e telefone são obrigatórios",
		})
		return
	}

	// Clean phone number
	phone := strings.ReplaceAll(req.Phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, "+", "")

	// Check if lead already exists with this phone (using fresh session)
	var existingLead models.Lead
	checkResult := db.Session(&gorm.Session{}).Where("phone = ? AND status NOT IN (?)", phone, []string{"converted", "lost"}).First(&existingLead)
	if checkResult.Error == nil {
		// Lead already exists, return it
		c.JSON(http.StatusOK, gin.H{
			"error":    false,
			"message":  "Lead já existe com este telefone",
			"lead":     existingLead,
			"existing": true,
		})
		return
	}

	// Set defaults
	source := "whatsapp"
	if req.Source != "" {
		source = req.Source
	}

	// Create lead with CreatedBy = 0 (system/API)
	lead := models.Lead{
		Name:          req.Name,
		Phone:         phone,
		Email:         req.Email,
		Source:        source,
		ContactReason: req.ContactReason,
		Notes:         req.Notes,
		Status:        "new",
		CreatedBy:     0, // System/WhatsApp API
	}

	// Add note about WhatsApp origin
	if lead.Notes != "" {
		lead.Notes += "\n"
	}
	lead.Notes += fmt.Sprintf("[Criado via WhatsApp API em %s]", time.Now().Format("02/01/2006 15:04"))

	// Create lead using fresh session to avoid state issues
	if err := db.Session(&gorm.Session{}).Create(&lead).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao criar lead: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"error":    false,
		"message":  "Lead criado com sucesso",
		"lead":     lead,
		"existing": false,
	})
}

// WhatsAppCreateAppointment creates a new appointment via WhatsApp API
// POST /api/whatsapp/appointments
func WhatsAppCreateAppointment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var req WhatsAppCreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Dados incompletos. Informe patient_id, dentist_id, date e time.",
		})
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Formato de data inválido. Use AAAA-MM-DD (ex: 2025-12-10)",
		})
		return
	}

	// Parse time
	timeVal, err := time.Parse("15:04", req.Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Formato de horário inválido. Use HH:MM (ex: 09:00)",
		})
		return
	}

	// Combine date and time for start_time using application timezone
	loc := getTimezone()
	startTime := time.Date(date.Year(), date.Month(), date.Day(),
		timeVal.Hour(), timeVal.Minute(), 0, 0, loc)

	// Check if appointment is in the future
	if startTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Não é possível agendar para uma data/hora passada",
		})
		return
	}

	// Get duration (default 30 minutes)
	duration := 30
	if req.Duration > 0 {
		duration = req.Duration
	}
	endTime := startTime.Add(time.Duration(duration) * time.Minute)

	// Use fresh sessions to avoid GORM state contamination
	freshDB := db.Session(&gorm.Session{})

	// Verify patient exists
	var patient models.Patient
	if err := freshDB.First(&patient, req.PatientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Paciente não encontrado",
		})
		return
	}

	// Verify dentist exists (using fresh session)
	var dentist models.User
	if err := db.Session(&gorm.Session{}).Table("public.users").Where("id = ? AND active = ?", req.DentistID, true).First(&dentist).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Profissional não encontrado",
		})
		return
	}

	// Check for time conflicts (using fresh session)
	var conflictCount int64
	db.Session(&gorm.Session{}).Model(&models.Appointment{}).
		Where("dentist_id = ? AND status NOT IN (?)", req.DentistID, []string{"cancelled", "no_show"}).
		Where("start_time < ? AND end_time > ?", endTime, startTime).
		Count(&conflictCount)

	if conflictCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   true,
			"message": "Conflito de horário. Já existe um agendamento para este profissional neste horário.",
		})
		return
	}

	// Set default procedure
	procedure := req.Procedure
	if procedure == "" {
		procedure = "consultation"
	}

	// Build notes
	notes := req.Notes
	if notes != "" {
		notes += "\n"
	}
	notes += fmt.Sprintf("[Agendado via WhatsApp API em %s]", time.Now().Format("02/01/2006 15:04"))

	// Create appointment
	appointment := models.Appointment{
		PatientID: req.PatientID,
		DentistID: req.DentistID,
		StartTime: startTime,
		EndTime:   endTime,
		Procedure: procedure,
		Status:    "scheduled",
		Notes:     notes,
	}

	if err := db.Session(&gorm.Session{}).Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao criar agendamento",
		})
		return
	}

	// Return response in expected format
	c.JSON(http.StatusCreated, gin.H{
		"error":   false,
		"message": "Consulta agendada com sucesso",
		"appointment": WhatsAppCreateAppointmentResponse{
			ID:          appointment.ID,
			Date:        startTime.Format("02/01/2006"),
			Time:        startTime.Format("15:04"),
			Procedure:   getProcedureLabel(procedure),
			DentistName: dentist.Name,
			Status:      "scheduled",
		},
	})
}
