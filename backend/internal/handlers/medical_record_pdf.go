package handlers

import (
	"drcrwell/backend/internal/models"
	"encoding/json"
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

	// Odontogram
	if record.Odontogram != nil && *record.Odontogram != "" {
		renderOdontogram(pdf, tr, *record.Odontogram)
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

type ToothData struct {
	Status     string   `json:"status"`
	Procedures []string `json:"procedures"`
}

func renderOdontogram(pdf *gofpdf.Fpdf, tr func(string) string, odontogramJSON string) {
	// Parse odontogram JSON
	var odontogram map[string]ToothData
	if err := json.Unmarshal([]byte(odontogramJSON), &odontogram); err != nil {
		return // Skip rendering if JSON is invalid
	}

	// Skip if odontogram is empty
	if len(odontogram) == 0 {
		return
	}

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Odontograma"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	// Status labels mapping
	statusLabels := map[string]string{
		"healthy":     "Saudavel",
		"cavity":      "Carie",
		"restoration": "Restauracao",
		"missing":     "Ausente",
		"root_canal":  "Canal",
		"crown":       "Coroa",
		"implant":     "Implante",
	}

	// Status symbols for PDF
	statusSymbols := map[string]string{
		"healthy":     "[OK]",
		"cavity":      "[!!]",
		"restoration": "[R]",
		"missing":     "[X]",
		"root_canal":  "[C]",
		"crown":       "[CR]",
		"implant":     "[I]",
	}

	// Define tooth quadrants (FDI notation)
	quadrants := []struct {
		name  string
		teeth []string
	}{
		{"Superior Direito", []string{"18", "17", "16", "15", "14", "13", "12", "11"}},
		{"Superior Esquerdo", []string{"21", "22", "23", "24", "25", "26", "27", "28"}},
		{"Inferior Direito", []string{"48", "47", "46", "45", "44", "43", "42", "41"}},
		{"Inferior Esquerdo", []string{"31", "32", "33", "34", "35", "36", "37", "38"}},
	}

	pdf.SetFont("Arial", "", 8)

	for _, quadrant := range quadrants {
		hasData := false
		// Check if this quadrant has any data
		for _, toothNum := range quadrant.teeth {
			if _, exists := odontogram[toothNum]; exists {
				hasData = true
				break
			}
		}

		// Skip quadrant if no data
		if !hasData {
			continue
		}

		// Render quadrant name
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(180, 5, tr(quadrant.name+":"), "LTR", 0, "L", false, 0, "")
		pdf.Ln(-1)

		pdf.SetFont("Arial", "", 8)
		// Render each tooth in the quadrant
		for _, toothNum := range quadrant.teeth {
			if toothData, exists := odontogram[toothNum]; exists {
				statusLabel := statusLabels[toothData.Status]
				if statusLabel == "" {
					statusLabel = toothData.Status
				}
				symbol := statusSymbols[toothData.Status]
				if symbol == "" {
					symbol = "[?]"
				}

				text := fmt.Sprintf("  Dente %s %s: %s", toothNum, symbol, statusLabel)
				pdf.CellFormat(180, 4, tr(text), "LR", 0, "L", false, 0, "")
				pdf.Ln(-1)
			}
		}
	}

	// Close the odontogram box
	pdf.CellFormat(180, 0, "", "LBR", 0, "L", false, 0, "")
	pdf.Ln(-1)

	// Add legend
	pdf.Ln(2)
	pdf.SetFont("Arial", "I", 7)
	pdf.CellFormat(180, 4, tr("Legenda: [OK]=Saudavel [!!]=Carie [R]=Restauracao [X]=Ausente [C]=Canal [CR]=Coroa [I]=Implante"), "0", 0, "L", false, 0, "")
	pdf.Ln(-1)
}
