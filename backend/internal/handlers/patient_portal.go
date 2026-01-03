package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/models"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// PatientPortalGetProfile returns the patient's own profile data
func PatientPortalGetProfile(c *gin.Context) {
	patientID, _ := c.Get("patient_id")
	tenantID, _ := c.Get("tenant_id")

	db := database.GetDB()
	schemaName := fmt.Sprintf("tenant_%d", tenantID.(uint))
	tenantDB := database.SetSchema(db, schemaName)

	var patient models.Patient
	if err := tenantDB.First(&patient, patientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente não encontrado"})
		return
	}

	// Decrypt sensitive fields
	if patient.CPF != "" {
		if decrypted, err := helpers.Decrypt(patient.CPF); err == nil {
			patient.CPF = decrypted
		}
	}
	if patient.RG != "" {
		if decrypted, err := helpers.Decrypt(patient.RG); err == nil {
			patient.RG = decrypted
		}
	}
	if patient.InsuranceNumber != "" {
		if decrypted, err := helpers.Decrypt(patient.InsuranceNumber); err == nil {
			patient.InsuranceNumber = decrypted
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"patient": patient,
	})
}

// PatientPortalGetClinic returns the clinic information for the patient
func PatientPortalGetClinic(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")

	db := database.GetDB()

	// Get tenant info
	var tenant models.Tenant
	if err := db.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clínica não encontrada"})
		return
	}

	// Get tenant settings
	var settings models.TenantSettings
	db.Where("tenant_id = ?", tenantID).First(&settings)

	// Get dentists (professionals)
	var dentists []struct {
		ID        uint   `json:"id"`
		Name      string `json:"name"`
		CRO       string `json:"cro"`
		Specialty string `json:"specialty"`
	}
	db.Table("public.users").
		Select("id, name, cro, specialty").
		Where("tenant_id = ? AND role IN ('dentist', 'admin') AND active = true AND deleted_at IS NULL", tenantID).
		Order("name ASC").
		Find(&dentists)

	c.JSON(http.StatusOK, gin.H{
		"clinic": gin.H{
			"name":    settings.ClinicName,
			"address": settings.ClinicAddress,
			"city":    settings.ClinicCity,
			"state":   settings.ClinicState,
			"zip":     settings.ClinicZip,
			"phone":   settings.ClinicPhone,
			"email":   settings.ClinicEmail,
		},
		"working_hours": gin.H{
			"start":                settings.WorkingHoursStart,
			"end":                  settings.WorkingHoursEnd,
			"appointment_duration": settings.DefaultAppointmentDuration,
			"lunch_break_enabled":  settings.LunchBreakEnabled,
			"lunch_break_start":    settings.LunchBreakStart,
			"lunch_break_end":      settings.LunchBreakEnd,
		},
		"dentists": dentists,
	})
}

// PatientPortalGetAppointments returns the patient's appointments
func PatientPortalGetAppointments(c *gin.Context) {
	patientID, _ := c.Get("patient_id")
	tenantID, _ := c.Get("tenant_id")

	db := database.GetDB()
	schemaName := fmt.Sprintf("tenant_%d", tenantID.(uint))
	tenantDB := database.SetSchema(db, schemaName)

	status := c.Query("status") // upcoming, past, all

	var appointments []models.Appointment
	query := tenantDB.Where("patient_id = ?", patientID).
		Preload("Dentist").
		Order("start_time DESC")

	now := time.Now()
	switch status {
	case "upcoming":
		query = query.Where("start_time >= ? AND status IN ('scheduled', 'confirmed')", now)
	case "past":
		query = query.Where("start_time < ? OR status IN ('completed', 'cancelled', 'no_show')", now)
	}

	if err := query.Find(&appointments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar agendamentos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": appointments,
		"total":        len(appointments),
	})
}

// PatientPortalCreateAppointmentRequest represents the request to create an appointment
type PatientPortalCreateAppointmentRequest struct {
	DentistID uint      `json:"dentist_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Procedure string    `json:"procedure"`
	Notes     string    `json:"notes"`
}

// PatientPortalCreateAppointment creates a new appointment for the patient
// Validates that the patient doesn't have another pending appointment
func PatientPortalCreateAppointment(c *gin.Context) {
	patientID, _ := c.Get("patient_id")
	tenantID, _ := c.Get("tenant_id")

	var req PatientPortalCreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	schemaName := fmt.Sprintf("tenant_%d", tenantID.(uint))
	tenantDB := database.SetSchema(db, schemaName)

	// Check if patient already has a pending appointment
	var pendingCount int64
	tenantDB.Model(&models.Appointment{}).
		Where("patient_id = ? AND status IN ('scheduled', 'confirmed')", patientID).
		Count(&pendingCount)

	if pendingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Você já possui uma consulta agendada. Aguarde a conclusão ou cancele a consulta atual para agendar outra.",
		})
		return
	}

	// Validate dentist exists and is active
	var dentist models.User
	if err := db.Where("id = ? AND tenant_id = ? AND role IN ('dentist', 'admin') AND active = true", req.DentistID, tenantID).First(&dentist).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Profissional não encontrado ou inativo"})
		return
	}

	// Check for time conflicts with dentist's schedule
	var conflictCount int64
	tenantDB.Model(&models.Appointment{}).
		Where("dentist_id = ? AND status NOT IN ('cancelled', 'no_show')", req.DentistID).
		Where("(start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?) OR (start_time >= ? AND end_time <= ?)",
			req.EndTime, req.StartTime,
			req.EndTime, req.StartTime,
			req.StartTime, req.EndTime).
		Count(&conflictCount)

	if conflictCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Horário indisponível. Por favor, escolha outro horário."})
		return
	}

	// Create appointment
	appointment := models.Appointment{
		PatientID: patientID.(uint),
		DentistID: req.DentistID,
		StartTime: models.LocalTime{Time: req.StartTime},
		EndTime:   models.LocalTime{Time: req.EndTime},
		Type:      "consultation",
		Procedure: req.Procedure,
		Status:    "scheduled",
		Notes:     req.Notes,
	}

	if err := tenantDB.Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar agendamento"})
		return
	}

	// Load dentist data for response
	appointment.Dentist = &dentist

	helpers.AuditAction(c, "create", "appointments", appointment.ID, true, map[string]interface{}{
		"patient_portal":  true,
		"patient_id":      patientID,
		"dentist_id":      req.DentistID,
		"start_time":      req.StartTime,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Agendamento criado com sucesso",
		"appointment": appointment,
	})
}

// PatientPortalCancelAppointment cancels the patient's own appointment
func PatientPortalCancelAppointment(c *gin.Context) {
	patientID, _ := c.Get("patient_id")
	tenantID, _ := c.Get("tenant_id")
	appointmentID := c.Param("id")

	db := database.GetDB()
	schemaName := fmt.Sprintf("tenant_%d", tenantID.(uint))
	tenantDB := database.SetSchema(db, schemaName)

	var appointment models.Appointment
	if err := tenantDB.Where("id = ? AND patient_id = ?", appointmentID, patientID).First(&appointment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento não encontrado"})
		return
	}

	// Only allow cancellation of scheduled or confirmed appointments
	if appointment.Status != "scheduled" && appointment.Status != "confirmed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Este agendamento não pode ser cancelado"})
		return
	}

	// Update status to cancelled using fresh DB connection to avoid GORM state issues
	updateDB := database.SetSchema(db, schemaName)
	if err := updateDB.Model(&models.Appointment{}).Where("id = ?", appointment.ID).Update("status", "cancelled").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao cancelar agendamento"})
		return
	}

	helpers.AuditAction(c, "cancel", "appointments", appointment.ID, true, map[string]interface{}{
		"patient_portal":   true,
		"patient_id":       patientID,
		"cancelled_by":     "patient",
		"appointment_date": appointment.StartTime,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Agendamento cancelado com sucesso",
	})
}

// PatientPortalGetMedicalRecords returns the patient's medical records
func PatientPortalGetMedicalRecords(c *gin.Context) {
	patientID, _ := c.Get("patient_id")
	tenantID, _ := c.Get("tenant_id")

	db := database.GetDB()
	schemaName := fmt.Sprintf("tenant_%d", tenantID.(uint))
	tenantDB := database.SetSchema(db, schemaName)

	var records []models.MedicalRecord
	query := tenantDB.Where("patient_id = ?", patientID).
		Preload("Dentist").
		Order("created_at DESC")

	// Optional type filter
	recordType := c.Query("type")
	if recordType != "" {
		query = query.Where("type = ?", recordType)
	}

	if err := query.Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar prontuários"})
		return
	}

	// Prepare response with sanitized data (remove sensitive internal fields)
	var response []gin.H
	for _, record := range records {
		item := gin.H{
			"id":             record.ID,
			"created_at":     record.CreatedAt,
			"type":           record.Type,
			"diagnosis":      record.Diagnosis,
			"treatment_plan": record.TreatmentPlan,
			"procedure_done": record.ProcedureDone,
			"evolution":      record.Evolution,
			"notes":          record.Notes,
			"is_signed":      record.IsSigned,
			"signed_at":      record.SignedAt,
			"signed_by_name": record.SignedByName,
			"signed_by_cro":  record.SignedByCRO,
		}

		// Include dentist name if available
		if record.Dentist != nil {
			item["dentist_name"] = record.Dentist.Name
			item["dentist_cro"] = record.Dentist.CRO
		}

		response = append(response, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"medical_records": response,
		"total":           len(response),
	})
}

// PatientPortalGetMedicalRecordDetail returns a specific medical record
func PatientPortalGetMedicalRecordDetail(c *gin.Context) {
	patientID, _ := c.Get("patient_id")
	tenantID, _ := c.Get("tenant_id")
	recordID := c.Param("id")

	db := database.GetDB()
	schemaName := fmt.Sprintf("tenant_%d", tenantID.(uint))
	tenantDB := database.SetSchema(db, schemaName)

	var record models.MedicalRecord
	if err := tenantDB.Where("id = ? AND patient_id = ?", recordID, patientID).
		Preload("Dentist").
		First(&record).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prontuário não encontrado"})
		return
	}

	response := gin.H{
		"id":             record.ID,
		"created_at":     record.CreatedAt,
		"type":           record.Type,
		"diagnosis":      record.Diagnosis,
		"treatment_plan": record.TreatmentPlan,
		"procedure_done": record.ProcedureDone,
		"materials":      record.Materials,
		"evolution":      record.Evolution,
		"notes":          record.Notes,
		"odontogram":     record.Odontogram,
		"is_signed":      record.IsSigned,
		"signed_at":      record.SignedAt,
		"signed_by_name": record.SignedByName,
		"signed_by_cro":  record.SignedByCRO,
	}

	// Include dentist info if available
	if record.Dentist != nil {
		response["dentist_name"] = record.Dentist.Name
		response["dentist_cro"] = record.Dentist.CRO
		response["dentist_specialty"] = record.Dentist.Specialty
	}

	c.JSON(http.StatusOK, gin.H{
		"medical_record": response,
	})
}

// PatientPortalGetAvailableSlots returns available time slots for a dentist on a specific date
func PatientPortalGetAvailableSlots(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	dentistIDStr := c.Query("dentist_id")
	dateStr := c.Query("date") // Format: YYYY-MM-DD

	if dentistIDStr == "" || dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dentist_id e date são obrigatórios"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use YYYY-MM-DD"})
		return
	}

	db := database.GetDB()
	schemaName := fmt.Sprintf("tenant_%d", tenantID.(uint))
	tenantDB := database.SetSchema(db, schemaName)

	// Get clinic settings
	var settings models.TenantSettings
	db.Where("tenant_id = ?", tenantID).First(&settings)

	// Get existing appointments for the day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var appointments []models.Appointment
	tenantDB.Where("dentist_id = ? AND start_time >= ? AND start_time < ? AND status NOT IN ('cancelled', 'no_show')",
		dentistIDStr, startOfDay, endOfDay).
		Order("start_time ASC").
		Find(&appointments)

	// Generate available slots based on working hours
	duration := settings.DefaultAppointmentDuration
	if duration == 0 {
		duration = 30 // default 30 minutes
	}

	// Parse working hours
	workStart := settings.WorkingHoursStart
	workEnd := settings.WorkingHoursEnd
	if workStart == "" {
		workStart = "08:00"
	}
	if workEnd == "" {
		workEnd = "18:00"
	}

	// Build list of taken time ranges
	takenSlots := make(map[string]bool)
	for _, apt := range appointments {
		takenSlots[apt.StartTime.Format("15:04")] = true
	}

	// Generate available slots
	var availableSlots []gin.H
	startHour, startMin := parseTime(workStart)
	endHour, endMin := parseTime(workEnd)
	lunchStartHour, lunchStartMin := parseTime(settings.LunchBreakStart)
	lunchEndHour, lunchEndMin := parseTime(settings.LunchBreakEnd)

	current := time.Date(date.Year(), date.Month(), date.Day(), startHour, startMin, 0, 0, date.Location())
	endTime := time.Date(date.Year(), date.Month(), date.Day(), endHour, endMin, 0, 0, date.Location())
	lunchStart := time.Date(date.Year(), date.Month(), date.Day(), lunchStartHour, lunchStartMin, 0, 0, date.Location())
	lunchEnd := time.Date(date.Year(), date.Month(), date.Day(), lunchEndHour, lunchEndMin, 0, 0, date.Location())

	for current.Before(endTime) {
		slotEnd := current.Add(time.Duration(duration) * time.Minute)
		timeStr := current.Format("15:04")

		// Skip if slot is taken
		if takenSlots[timeStr] {
			current = slotEnd
			continue
		}

		// Skip if slot is during lunch break
		if settings.LunchBreakEnabled && !current.Before(lunchStart) && current.Before(lunchEnd) {
			current = slotEnd
			continue
		}

		// Skip past times for today
		if date.Day() == time.Now().Day() && current.Before(time.Now()) {
			current = slotEnd
			continue
		}

		availableSlots = append(availableSlots, gin.H{
			"start_time": current.Format("15:04"),
			"end_time":   slotEnd.Format("15:04"),
		})

		current = slotEnd
	}

	c.JSON(http.StatusOK, gin.H{
		"date":            dateStr,
		"dentist_id":      dentistIDStr,
		"available_slots": availableSlots,
		"duration":        duration,
	})
}

// parseTime parses time string "HH:MM" and returns hour and minute
func parseTime(timeStr string) (int, int) {
	if timeStr == "" {
		return 0, 0
	}
	var hour, min int
	fmt.Sscanf(timeStr, "%d:%d", &hour, &min)
	return hour, min
}

// ========================================
// Admin endpoints for managing patient portal access
// ========================================

// CreatePatientPortalAccessRequest represents the request to create portal access
type CreatePatientPortalAccessRequest struct {
	PatientID uint   `json:"patient_id" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
}

// CreatePatientPortalAccess creates a user account for a patient to access the portal
func CreatePatientPortalAccess(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")

	var req CreatePatientPortalAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	schemaName := fmt.Sprintf("tenant_%d", tenantID.(uint))
	tenantDB := database.SetSchema(db, schemaName)

	// Verify patient exists
	var patient models.Patient
	if err := tenantDB.First(&patient, req.PatientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente não encontrado"})
		return
	}

	// Check if email is already in use
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Este email já está em uso"})
		return
	}

	// Check if patient already has portal access
	var existingPatientUser models.User
	if err := db.Where("patient_id = ? AND tenant_id = ?", req.PatientID, tenantID).First(&existingPatientUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Este paciente já possui acesso ao portal"})
		return
	}

	// Create user with patient role
	patientIDPtr := req.PatientID
	user := models.User{
		TenantID:  tenantID.(uint),
		Name:      patient.Name,
		Email:     req.Email,
		Role:      "patient",
		Active:    true,
		PatientID: &patientIDPtr,
	}

	if err := user.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar senha"})
		return
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar acesso"})
		return
	}

	// Update patient email if different
	if patient.Email != req.Email {
		tenantDB.Model(&patient).Update("email", req.Email)
	}

	helpers.AuditAction(c, "create", "patient_portal_access", user.ID, true, map[string]interface{}{
		"patient_id":    req.PatientID,
		"patient_name":  patient.Name,
		"portal_email":  req.Email,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Acesso ao portal criado com sucesso",
		"user": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"patient_id": user.PatientID,
		},
	})
}

// UpdatePatientPortalPasswordRequest represents the request to update patient password
type UpdatePatientPortalPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

// UpdatePatientPortalPassword updates the password for a patient portal user
func UpdatePatientPortalPassword(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	patientID := c.Param("patient_id")

	var req UpdatePatientPortalPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Find user by patient_id
	var user models.User
	if err := db.Where("patient_id = ? AND tenant_id = ? AND role = 'patient'", patientID, tenantID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Acesso ao portal não encontrado para este paciente"})
		return
	}

	// Update password
	if err := user.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar senha"})
		return
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar senha"})
		return
	}

	helpers.AuditAction(c, "update", "patient_portal_password", user.ID, true, map[string]interface{}{
		"patient_id": patientID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Senha atualizada com sucesso",
	})
}

// DeletePatientPortalAccess removes portal access for a patient
func DeletePatientPortalAccess(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	patientID := c.Param("patient_id")

	db := database.GetDB()

	// Find and delete user
	result := db.Where("patient_id = ? AND tenant_id = ? AND role = 'patient'", patientID, tenantID).Delete(&models.User{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover acesso"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Acesso ao portal não encontrado"})
		return
	}

	helpers.AuditAction(c, "delete", "patient_portal_access", 0, true, map[string]interface{}{
		"patient_id": patientID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Acesso ao portal removido com sucesso",
	})
}

// GetPatientPortalAccess checks if a patient has portal access
func GetPatientPortalAccess(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")
	patientID := c.Param("patient_id")

	db := database.GetDB()

	var user models.User
	if err := db.Where("patient_id = ? AND tenant_id = ? AND role = 'patient'", patientID, tenantID).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"has_access": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_access": true,
		"user": gin.H{
			"id":     user.ID,
			"email":  user.Email,
			"active": user.Active,
		},
	})
}

// ========================================
// Public endpoints for patient portal login (no auth required)
// ========================================

// PatientPortalPublicClinicInfo returns public clinic info by subdomain slug
// This is used on the login page to display the clinic name
func PatientPortalPublicClinicInfo(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug é obrigatório"})
		return
	}

	// Validate slug format (alphanumeric and hyphens only)
	for _, r := range slug {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Slug inválido"})
			return
		}
	}

	db := database.GetDB()

	// Find tenant by subdomain
	var tenant models.Tenant
	if err := db.Where("subdomain = ? AND active = true", slug).First(&tenant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Clínica não encontrada"})
		return
	}

	// Get tenant settings for clinic name
	var settings models.TenantSettings
	db.Where("tenant_id = ?", tenant.ID).First(&settings)

	clinicName := settings.ClinicName
	if clinicName == "" {
		clinicName = tenant.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"clinic_name": clinicName,
		"slug":        slug,
	})
}

// PatientPortalLoginRequest represents the login request for patient portal
type PatientPortalLoginRequest struct {
	Slug     string `json:"slug" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// PatientPortalLogin authenticates a patient and returns JWT token
// Only allows users with role='patient' from the specified tenant
func PatientPortalLogin(c *gin.Context) {
	var req PatientPortalLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos"})
		return
	}

	db := database.GetDB()

	// Find tenant by subdomain
	var tenant models.Tenant
	if err := db.Where("subdomain = ? AND active = true", req.Slug).First(&tenant).Error; err != nil {
		helpers.AuditLogin(c, req.Email, false, map[string]interface{}{
			"reason":       "tenant_not_found",
			"slug":         req.Slug,
			"patient_portal": true,
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciais inválidas"})
		return
	}

	// Check if tenant subscription is active
	if !tenant.IsSubscriptionActive() {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Clínica com assinatura inativa"})
		return
	}

	// Find user by email AND tenant_id AND role='patient'
	var user models.User
	if err := db.Where("email = ? AND tenant_id = ? AND role = 'patient'", req.Email, tenant.ID).First(&user).Error; err != nil {
		helpers.AuditLogin(c, req.Email, false, map[string]interface{}{
			"reason":         "patient_not_found",
			"tenant_id":      tenant.ID,
			"slug":           req.Slug,
			"patient_portal": true,
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciais inválidas"})
		return
	}

	// Check if user is active
	if !user.Active {
		helpers.AuditLogin(c, req.Email, false, map[string]interface{}{
			"reason":         "user_inactive",
			"patient_portal": true,
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Conta inativa"})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		helpers.AuditLogin(c, req.Email, false, map[string]interface{}{
			"reason":         "wrong_password",
			"patient_portal": true,
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciais inválidas"})
		return
	}

	// Generate JWT token
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	// Create claims
	claims := Claims{
		UserID:       user.ID,
		TenantID:     user.TenantID,
		Email:        user.Email,
		Role:         user.Role,
		IsSuperAdmin: false,
		TenantActive: true,
		PatientID:    user.PatientID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar token"})
		return
	}

	// Audit successful login
	helpers.AuditLogin(c, req.Email, true, map[string]interface{}{
		"patient_portal": true,
		"tenant_id":      tenant.ID,
		"patient_id":     user.PatientID,
	})

	// Get clinic name for response
	var settings models.TenantSettings
	db.Where("tenant_id = ?", tenant.ID).First(&settings)
	clinicName := settings.ClinicName
	if clinicName == "" {
		clinicName = tenant.Name
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"patient_id": user.PatientID,
			"tenant_id":  user.TenantID,
		},
		"clinic_name": clinicName,
	})
}
