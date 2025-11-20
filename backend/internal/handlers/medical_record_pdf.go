package handlers

import (
	"drcrwell/backend/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

func GenerateMedicalRecordPDF(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint("tenant_id")
	recordID := c.Param("id")

	// Get tenant info for header
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	// Get medical record with patient and dentist info
	var record models.MedicalRecord
	if err := db.Session(&gorm.Session{NewDB: true}).
		Preload("Patient").
		Preload("Dentist").
		Where("id = ?", recordID).
		First(&record).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Medical record not found"})
		return
	}

	// Create PDF with proper margins for A4
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, tr(tenant.Name))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 5, tr(tenant.Address+", "+tenant.City+" - "+tenant.State))
	pdf.Ln(5)
	pdf.Cell(0, 5, tr("Tel: "+tenant.Phone))
	pdf.Ln(10)

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, tr("Prontuario Medico"))
	pdf.Ln(10)

	// Patient info
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Informacoes do Paciente"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(60, 6, tr("Paciente:"), "1", 0, "L", false, 0, "")
	patientName := record.Patient.Name
	if record.Patient == nil {
		patientName = "N/A"
	}
	pdf.CellFormat(120, 6, tr(patientName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Profissional:"), "1", 0, "L", false, 0, "")
	dentistName := "N/A"
	if record.Dentist != nil {
		dentistName = record.Dentist.Name
	}
	pdf.CellFormat(120, 6, tr(dentistName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Tipo:"), "1", 0, "L", false, 0, "")
	typeLabel := getTypeLabel(record.Type)
	pdf.CellFormat(120, 6, tr(typeLabel), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Data:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(120, 6, record.CreatedAt.Format("02/01/2006 15:04"), "1", 0, "L", false, 0, "")
	pdf.Ln(10)

	// Record details
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Detalhes do Prontuario"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)

	// Diagnosis
	if record.Diagnosis != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(180, 6, tr("Diagnostico:"), "LTR", 0, "L", false, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(record.Diagnosis), "LBR", "L", false)
	}

	// Treatment Plan
	if record.TreatmentPlan != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(180, 6, tr("Plano de Tratamento:"), "LTR", 0, "L", false, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(record.TreatmentPlan), "LBR", "L", false)
	}

	// Procedure Done
	if record.ProcedureDone != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(180, 6, tr("Procedimentos Realizados:"), "LTR", 0, "L", false, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(record.ProcedureDone), "LBR", "L", false)
	}

	// Materials
	if record.Materials != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(180, 6, tr("Materiais Utilizados:"), "LTR", 0, "L", false, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(record.Materials), "LBR", "L", false)
	}

	// Prescription
	if record.Prescription != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(180, 6, tr("Prescricao:"), "LTR", 0, "L", false, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(record.Prescription), "LBR", "L", false)
	}

	// Certificate
	if record.Certificate != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(180, 6, tr("Atestado:"), "LTR", 0, "L", false, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(record.Certificate), "LBR", "L", false)
	}

	// Evolution
	if record.Evolution != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(180, 6, tr("Evolucao:"), "LTR", 0, "L", false, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(record.Evolution), "LBR", "L", false)
	}

	// Notes
	if record.Notes != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(180, 6, tr("Notas Adicionais:"), "LTR", 0, "L", false, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(record.Notes), "LBR", "L", false)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=prontuario_%d.pdf", record.ID))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

func getTypeLabel(typeValue string) string {
	types := map[string]string{
		"anamnesis":    "Anamnese",
		"treatment":    "Tratamento",
		"procedure":    "Procedimento",
		"prescription": "Receita",
		"certificate":  "Atestado",
	}

	if label, ok := types[typeValue]; ok {
		return label
	}
	return typeValue
}
