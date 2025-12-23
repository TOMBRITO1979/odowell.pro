package handlers

import (
	"drcrwell/backend/internal/models"
	"drcrwell/backend/internal/middleware"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

func GenerateMedicalRecordPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
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

	// Record details header
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Detalhes do Prontuario"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	// Render sections with smart page breaks
	renderSection(pdf, tr, "Diagnostico:", record.Diagnosis)
	renderSection(pdf, tr, "Plano de Tratamento:", record.TreatmentPlan)
	renderSection(pdf, tr, "Procedimentos Realizados:", record.ProcedureDone)
	renderSection(pdf, tr, "Materiais Utilizados:", record.Materials)
	renderSection(pdf, tr, "Prescricao:", record.Prescription)
	renderSection(pdf, tr, "Atestado:", record.Certificate)
	renderSection(pdf, tr, "Evolucao:", record.Evolution)
	renderSection(pdf, tr, "Notas Adicionais:", record.Notes)

	// Odontogram
	if record.Odontogram != nil && *record.Odontogram != "" {
		renderOdontogram(pdf, tr, *record.Odontogram)
	}

	// Digital signature block (if signed)
	if record.IsSigned {
		// Check if we need a new page for signature block (ensure at least 40mm space)
		_, pageHeight := pdf.GetPageSize()
		_, _, _, bottomMargin := pdf.GetMargins()
		currentY := pdf.GetY()
		if currentY > pageHeight-bottomMargin-45 {
			pdf.AddPage()
		}

		pdf.Ln(10)
		pdf.SetFillColor(230, 255, 230) // Light green background
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(180, 6, tr("DOCUMENTO ASSINADO DIGITALMENTE"), "1", 1, "C", true, 0, "")

		pdf.SetFont("Arial", "", 8)
		pdf.SetFillColor(245, 255, 245)
		pdf.CellFormat(180, 5, tr(fmt.Sprintf("Assinado por: %s (CRO: %s)", record.SignedByName, record.SignedByCRO)), "LR", 1, "L", true, 0, "")
		if record.SignedAt != nil {
			pdf.CellFormat(180, 5, tr(fmt.Sprintf("Data/Hora: %s", record.SignedAt.Format("02/01/2006 15:04:05"))), "LR", 1, "L", true, 0, "")
		}
		pdf.CellFormat(180, 5, tr(fmt.Sprintf("Certificado: %s", record.CertificateThumbprint)), "LR", 1, "L", true, 0, "")
		pdf.CellFormat(180, 5, tr(fmt.Sprintf("Hash SHA-256: %s", record.SignatureHash)), "LRB", 1, "L", true, 0, "")

		pdf.Ln(2)
		pdf.SetFont("Arial", "I", 7)
		pdf.MultiCell(180, 3, tr("Este documento foi assinado digitalmente com certificado ICP-Brasil. A integridade pode ser verificada atraves do hash acima."), "", "C", false)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	filename := fmt.Sprintf("prontuario_%d.pdf", record.ID)
	if record.IsSigned {
		filename = fmt.Sprintf("prontuario_assinado_%d.pdf", record.ID)
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

// checkPageBreak checks if there's enough space for content, adds new page if needed
func checkPageBreak(pdf *gofpdf.Fpdf, minSpace float64) {
	_, pageHeight := pdf.GetPageSize()
	_, _, _, bottomMargin := pdf.GetMargins()
	currentY := pdf.GetY()
	if currentY > pageHeight-bottomMargin-minSpace {
		pdf.AddPage()
	}
}

// renderSection renders a section with title and content, handling page breaks
func renderSection(pdf *gofpdf.Fpdf, tr func(string) string, title, content string) {
	if content == "" {
		return
	}

	// Estimate content height (approximately 5mm per line, 70 chars per line)
	lines := (len(content) / 70) + 1
	estimatedHeight := float64(lines*5) + 12 // content + title + margins

	// Check if we need a page break (leave at least 25mm or content height)
	minSpace := estimatedHeight
	if minSpace < 25 {
		minSpace = 25
	}
	if minSpace > 100 {
		minSpace = 100 // Cap at 100mm to avoid always breaking for long content
	}
	checkPageBreak(pdf, minSpace)

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(180, 6, tr(title), "LTR", 0, "L", false, 0, "")
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(180, 5, tr(content), "LBR", "L", false)
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

	// Check for page break before odontogram section
	checkPageBreak(pdf, 50)

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
