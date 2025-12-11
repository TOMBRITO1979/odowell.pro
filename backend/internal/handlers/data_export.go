package handlers

import (
	"bytes"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PatientExportData represents the structured data export for LGPD portability
type PatientExportData struct {
	ExportDate     string                   `json:"export_date"`
	ExportedFor    string                   `json:"exported_for"`
	RequestID      uint                     `json:"request_id"`
	Patient        PatientBasicData         `json:"patient"`
	Appointments   []AppointmentExportData  `json:"appointments"`
	MedicalRecords []MedicalRecordExportData `json:"medical_records"`
	Prescriptions  []PrescriptionExportData `json:"prescriptions"`
	Exams          []ExamExportData         `json:"exams"`
	Consents       []ConsentExportData      `json:"consents"`
}

type PatientBasicData struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	CPF         string `json:"cpf"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	BirthDate   string `json:"birth_date,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	ZipCode     string `json:"zip_code,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type AppointmentExportData struct {
	ID          uint   `json:"id"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	DentistName string `json:"dentist_name"`
	Notes       string `json:"notes,omitempty"`
}

type MedicalRecordExportData struct {
	ID            uint   `json:"id"`
	Type          string `json:"type"`
	Date          string `json:"date"`
	DentistName   string `json:"dentist_name"`
	Diagnosis     string `json:"diagnosis,omitempty"`
	TreatmentPlan string `json:"treatment_plan,omitempty"`
	ProcedureDone string `json:"procedure_done,omitempty"`
	Evolution     string `json:"evolution,omitempty"`
	Notes         string `json:"notes,omitempty"`
}

type PrescriptionExportData struct {
	ID           uint   `json:"id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Date         string `json:"date"`
	DentistName  string `json:"dentist_name"`
	Diagnosis    string `json:"diagnosis,omitempty"`
	Medications  string `json:"medications,omitempty"`
	Content      string `json:"content,omitempty"`
	ValidUntil   string `json:"valid_until,omitempty"`
}

type ExamExportData struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Date        string `json:"date"`
	Result      string `json:"result,omitempty"`
	Notes       string `json:"notes,omitempty"`
	RequestedBy string `json:"requested_by,omitempty"`
}

type ConsentExportData struct {
	ID           uint   `json:"id"`
	TemplateName string `json:"template_name"`
	SignedAt     string `json:"signed_at"`
	Status       string `json:"status"`
}

// ExportPatientData exports all patient data in JSON or CSV format for LGPD portability
func ExportPatientData(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	requestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de solicitacao invalido"})
		return
	}

	// Get the data request
	var request models.DataRequest
	if err := db.First(&request, requestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Solicitacao nao encontrada"})
		return
	}

	// Check if request type is portability or access
	if request.Type != models.DataRequestTypePortability && request.Type != models.DataRequestTypeAccess {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Este tipo de solicitacao nao permite exportacao de dados"})
		return
	}

	// Check if identity was verified for portability requests
	if request.Type == models.DataRequestTypePortability && !request.OTPVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Identidade do titular nao foi verificada. Envie o codigo OTP primeiro."})
		return
	}

	// Get patient data
	var patient models.Patient
	if err := db.First(&patient, request.PatientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente nao encontrado"})
		return
	}

	// Build export data with all related records
	export := PatientExportData{
		ExportDate:  time.Now().Format("2006-01-02 15:04:05"),
		ExportedFor: "LGPD - Lei Geral de Protecao de Dados",
		RequestID:   uint(requestID),
		Patient: PatientBasicData{
			ID:        patient.ID,
			Name:      patient.Name,
			CPF:       patient.CPF,
			Email:     patient.Email,
			Phone:     patient.Phone,
			Address:   patient.Address,
			City:      patient.City,
			State:     patient.State,
			ZipCode:   patient.ZipCode,
			CreatedAt: patient.CreatedAt.Format("2006-01-02"),
		},
	}

	if patient.BirthDate != nil {
		export.Patient.BirthDate = patient.BirthDate.Format("2006-01-02")
	}

	// Load appointments
	var appointments []models.Appointment
	db.Where("patient_id = ?", patient.ID).Preload("Dentist").Order("start_time DESC").Find(&appointments)
	for _, apt := range appointments {
		dentistName := ""
		if apt.Dentist != nil {
			dentistName = apt.Dentist.Name
		}
		export.Appointments = append(export.Appointments, AppointmentExportData{
			ID:          apt.ID,
			Date:        apt.StartTime.Format("2006-01-02"),
			Time:        apt.StartTime.Format("15:04"),
			Status:      apt.Status,
			Type:        apt.Type,
			DentistName: dentistName,
			Notes:       apt.Notes,
		})
	}

	// Load medical records
	var records []models.MedicalRecord
	db.Where("patient_id = ?", patient.ID).Preload("Dentist").Order("created_at DESC").Find(&records)
	for _, rec := range records {
		dentistName := ""
		if rec.Dentist != nil {
			dentistName = rec.Dentist.Name
		}
		export.MedicalRecords = append(export.MedicalRecords, MedicalRecordExportData{
			ID:            rec.ID,
			Type:          rec.Type,
			Date:          rec.CreatedAt.Format("2006-01-02"),
			DentistName:   dentistName,
			Diagnosis:     rec.Diagnosis,
			TreatmentPlan: rec.TreatmentPlan,
			ProcedureDone: rec.ProcedureDone,
			Evolution:     rec.Evolution,
			Notes:         rec.Notes,
		})
	}

	// Load prescriptions
	var prescriptions []models.Prescription
	db.Where("patient_id = ?", patient.ID).Preload("Dentist").Order("created_at DESC").Find(&prescriptions)
	for _, presc := range prescriptions {
		dentistName := ""
		if presc.Dentist != nil {
			dentistName = presc.Dentist.Name
		}
		validUntil := ""
		if presc.ValidUntil != nil {
			validUntil = presc.ValidUntil.Format("2006-01-02")
		}
		prescDate := ""
		if presc.PrescriptionDate != nil {
			prescDate = presc.PrescriptionDate.Format("2006-01-02")
		} else {
			prescDate = presc.CreatedAt.Format("2006-01-02")
		}
		export.Prescriptions = append(export.Prescriptions, PrescriptionExportData{
			ID:          presc.ID,
			Type:        presc.Type,
			Title:       presc.Title,
			Date:        prescDate,
			DentistName: dentistName,
			Diagnosis:   presc.Diagnosis,
			Medications: presc.Medications,
			Content:     presc.Content,
			ValidUntil:  validUntil,
		})
	}

	// Load exams
	var exams []models.Exam
	db.Where("patient_id = ?", patient.ID).Order("created_at DESC").Find(&exams)
	for _, exam := range exams {
		examDate := ""
		if exam.ExamDate != nil {
			examDate = exam.ExamDate.Format("2006-01-02")
		} else {
			examDate = exam.CreatedAt.Format("2006-01-02")
		}
		export.Exams = append(export.Exams, ExamExportData{
			ID:          exam.ID,
			Type:        exam.ExamType,
			Name:        exam.Name,
			Date:        examDate,
			Result:      exam.Description,
			Notes:       exam.Notes,
			RequestedBy: "",
		})
	}

	// Load consents
	var consents []models.PatientConsent
	db.Where("patient_id = ?", patient.ID).Preload("Template").Order("created_at DESC").Find(&consents)
	for _, consent := range consents {
		templateName := consent.TemplateTitle
		if consent.Template.Title != "" {
			templateName = consent.Template.Title
		}
		signedAt := consent.SignedAt.Format("2006-01-02 15:04:05")
		export.Consents = append(export.Consents, ConsentExportData{
			ID:           consent.ID,
			TemplateName: templateName,
			SignedAt:     signedAt,
			Status:       consent.Status,
		})
	}

	// Get format from query parameter (default: json)
	format := c.DefaultQuery("format", "json")

	switch format {
	case "json":
		exportJSON(c, export, patient.Name)
	case "csv":
		exportCSV(c, export, patient.Name)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato invalido. Use 'json' ou 'csv'"})
		return
	}

	// Audit log
	helpers.AuditAction(c, "export_data", "data_requests", uint(requestID), true, map[string]interface{}{
		"patient_id": patient.ID,
		"format":     format,
	})
}

func exportJSON(c *gin.Context, data PatientExportData, patientName string) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar JSON"})
		return
	}

	filename := fmt.Sprintf("dados_lgpd_%s_%s.json", sanitizeFilename(patientName), time.Now().Format("20060102"))
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/json", jsonData)
}

func exportCSV(c *gin.Context, data PatientExportData, patientName string) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header info
	writer.Write([]string{"EXPORTACAO DE DADOS LGPD"})
	writer.Write([]string{"Data de Exportacao", data.ExportDate})
	writer.Write([]string{"Solicitacao ID", fmt.Sprintf("%d", data.RequestID)})
	writer.Write([]string{""})

	// Patient data
	writer.Write([]string{"DADOS DO PACIENTE"})
	writer.Write([]string{"Campo", "Valor"})
	writer.Write([]string{"ID", fmt.Sprintf("%d", data.Patient.ID)})
	writer.Write([]string{"Nome", data.Patient.Name})
	writer.Write([]string{"CPF", data.Patient.CPF})
	writer.Write([]string{"Email", data.Patient.Email})
	writer.Write([]string{"Telefone", data.Patient.Phone})
	writer.Write([]string{"Data de Nascimento", data.Patient.BirthDate})
	writer.Write([]string{"Endereco", data.Patient.Address})
	writer.Write([]string{"Cidade", data.Patient.City})
	writer.Write([]string{"Estado", data.Patient.State})
	writer.Write([]string{"CEP", data.Patient.ZipCode})
	writer.Write([]string{"Cadastrado em", data.Patient.CreatedAt})
	writer.Write([]string{""})

	// Appointments
	writer.Write([]string{"CONSULTAS"})
	writer.Write([]string{"ID", "Data", "Hora", "Status", "Tipo", "Profissional", "Observacoes"})
	for _, apt := range data.Appointments {
		writer.Write([]string{
			fmt.Sprintf("%d", apt.ID),
			apt.Date,
			apt.Time,
			apt.Status,
			apt.Type,
			apt.DentistName,
			apt.Notes,
		})
	}
	writer.Write([]string{""})

	// Medical Records
	writer.Write([]string{"PRONTUARIOS"})
	writer.Write([]string{"ID", "Tipo", "Data", "Profissional", "Diagnostico", "Plano de Tratamento", "Procedimento", "Evolucao", "Observacoes"})
	for _, rec := range data.MedicalRecords {
		writer.Write([]string{
			fmt.Sprintf("%d", rec.ID),
			rec.Type,
			rec.Date,
			rec.DentistName,
			rec.Diagnosis,
			rec.TreatmentPlan,
			rec.ProcedureDone,
			rec.Evolution,
			rec.Notes,
		})
	}
	writer.Write([]string{""})

	// Prescriptions
	writer.Write([]string{"RECEITUARIOS"})
	writer.Write([]string{"ID", "Tipo", "Titulo", "Data", "Profissional", "Diagnostico", "Medicamentos", "Conteudo", "Valido ate"})
	for _, presc := range data.Prescriptions {
		writer.Write([]string{
			fmt.Sprintf("%d", presc.ID),
			presc.Type,
			presc.Title,
			presc.Date,
			presc.DentistName,
			presc.Diagnosis,
			presc.Medications,
			presc.Content,
			presc.ValidUntil,
		})
	}
	writer.Write([]string{""})

	// Exams
	writer.Write([]string{"EXAMES"})
	writer.Write([]string{"ID", "Tipo", "Nome", "Data", "Resultado", "Observacoes", "Solicitado por"})
	for _, exam := range data.Exams {
		writer.Write([]string{
			fmt.Sprintf("%d", exam.ID),
			exam.Type,
			exam.Name,
			exam.Date,
			exam.Result,
			exam.Notes,
			exam.RequestedBy,
		})
	}
	writer.Write([]string{""})

	// Consents
	writer.Write([]string{"TERMOS DE CONSENTIMENTO"})
	writer.Write([]string{"ID", "Termo", "Assinado em", "Status"})
	for _, consent := range data.Consents {
		writer.Write([]string{
			fmt.Sprintf("%d", consent.ID),
			consent.TemplateName,
			consent.SignedAt,
			consent.Status,
		})
	}

	writer.Flush()

	filename := fmt.Sprintf("dados_lgpd_%s_%s.csv", sanitizeFilename(patientName), time.Now().Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	// Add BOM for Excel compatibility
	bom := []byte{0xEF, 0xBB, 0xBF}
	c.Data(http.StatusOK, "text/csv; charset=utf-8", append(bom, buf.Bytes()...))
}

func sanitizeFilename(name string) string {
	// Remove special characters and spaces
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else if r == ' ' {
			result += "_"
		}
	}
	if len(result) > 30 {
		result = result[:30]
	}
	return result
}
