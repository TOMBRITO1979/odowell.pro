package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

// ExportBudgetsCSV exports budgets to CSV format
func ExportBudgetsCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Build SQL query with filters
	sqlQuery := "SELECT * FROM budgets WHERE deleted_at IS NULL"
	var args []interface{}

	if patientID := c.Query("patient_id"); patientID != "" {
		sqlQuery += " AND patient_id = ?"
		args = append(args, patientID)
	}
	if dentistID := c.Query("dentist_id"); dentistID != "" {
		sqlQuery += " AND dentist_id = ?"
		args = append(args, dentistID)
	}
	if status := c.Query("status"); status != "" {
		sqlQuery += " AND status = ?"
		args = append(args, status)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		sqlQuery += " AND created_at >= ?"
		args = append(args, startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		sqlQuery += " AND created_at <= ?"
		args = append(args, endDate)
	}

	sqlQuery += " ORDER BY created_at DESC"

	// Load budgets
	var budgets []models.Budget
	if err := db.Raw(sqlQuery, args...).Scan(&budgets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budgets"})
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=orcamentos_%s.csv", time.Now().Format("20060102_150405")))

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID",
		"Paciente",
		"Profissional",
		"Descrição",
		"Valor Total",
		"Status",
		"Válido Até",
		"Data Criação",
		"Observações",
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
		return
	}

	// Write data rows
	for _, budget := range budgets {
		// Fetch patient name using raw SQL
		patientName := ""
		var tmpPatientName string
		if err := db.Raw("SELECT name FROM patients WHERE id = ? AND deleted_at IS NULL", budget.PatientID).Scan(&tmpPatientName).Error; err == nil {
			patientName = tmpPatientName
		}

		// Fetch dentist name using raw SQL from public schema
		dentistName := ""
		var tmpDentistName string
		if err := db.Raw("SELECT name FROM public.users WHERE id = ? AND deleted_at IS NULL", budget.DentistID).Scan(&tmpDentistName).Error; err == nil {
			dentistName = tmpDentistName
		}

		validUntil := ""
		if budget.ValidUntil != nil {
			validUntil = budget.ValidUntil.Format("2006-01-02")
		}

		row := []string{
			fmt.Sprintf("%d", budget.ID),
			patientName,
			dentistName,
			budget.Description,
			fmt.Sprintf("%.2f", budget.TotalValue),
			budget.Status,
			validUntil,
			budget.CreatedAt.Format("2006-01-02 15:04:05"),
			budget.Notes,
		}

		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV row"})
			return
		}
	}
}

// ImportBudgetsCSV imports budgets from CSV file
func ImportBudgetsCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	userID := c.GetUint("user_id")

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(header.Filename, ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only CSV files are allowed"})
		return
	}

	// Parse CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse CSV file"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file must contain at least a header and one data row"})
		return
	}

	// Process records (skip header)
	imported := 0
	errors := []string{}

	for i, record := range records[1:] {
		lineNum := i + 2 // Line number in file (1-indexed, accounting for header)

		if len(record) < 5 {
			errors = append(errors, fmt.Sprintf("Linha %d: Dados insuficientes", lineNum))
			continue
		}

		// Parse required fields
		patientID, err := strconv.ParseUint(record[0], 10, 32)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Linha %d: ID do paciente inválido", lineNum))
			continue
		}

		totalValue, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Linha %d: Valor total inválido", lineNum))
			continue
		}

		description := record[1]
		status := record[3]
		notes := ""
		if len(record) > 4 {
			notes = record[4]
		}

		// Validate status
		validStatuses := map[string]bool{
			"pending":  true,
			"approved": true,
			"rejected": true,
			"expired":  true,
		}
		if !validStatuses[status] {
			status = "pending"
		}

		// Create budget
		budget := models.Budget{
			PatientID:   uint(patientID),
			DentistID:   userID,
			Description: description,
			TotalValue:  totalValue,
			Status:      status,
			Notes:       notes,
		}

		if err := db.Create(&budget).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Linha %d: Erro ao criar orçamento - %v", lineNum, err))
			continue
		}

		imported++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Importação concluída: %d orçamentos importados", imported),
		"imported": imported,
		"errors":   errors,
	})
}

// GenerateBudgetsListPDF generates a PDF with the list of budgets
func GenerateBudgetsListPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	tenantID := c.GetUint("tenant_id")

	// Get tenant info
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	// Build SQL query with filters
	sqlQuery := "SELECT * FROM budgets WHERE deleted_at IS NULL"
	var args []interface{}

	if patientID := c.Query("patient_id"); patientID != "" {
		sqlQuery += " AND patient_id = ?"
		args = append(args, patientID)
	}
	if dentistID := c.Query("dentist_id"); dentistID != "" {
		sqlQuery += " AND dentist_id = ?"
		args = append(args, dentistID)
	}
	if status := c.Query("status"); status != "" {
		sqlQuery += " AND status = ?"
		args = append(args, status)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		sqlQuery += " AND created_at >= ?"
		args = append(args, startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		sqlQuery += " AND created_at <= ?"
		args = append(args, endDate)
	}

	sqlQuery += " ORDER BY created_at DESC"

	// Load budgets
	var budgets []models.Budget
	if err := db.Raw(sqlQuery, args...).Scan(&budgets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budgets"})
		return
	}

	// Create PDF
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape for better table fit
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 10)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Modern header with brand color (#16a34a)
	pdf.SetFillColor(22, 163, 74) // #16a34a
	pdf.Rect(0, 0, 297, 25, "F")

	pdf.SetTextColor(255, 255, 255) // White text
	pdf.SetFont("Arial", "B", 18)
	pdf.SetY(8)
	pdf.Cell(0, 8, tr(tenant.Name))
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, tr(tenant.Address+", "+tenant.City+" - "+tenant.State))
	pdf.Ln(10)

	// Reset text color to black
	pdf.SetTextColor(0, 0, 0)

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, tr("Relatorio de Orcamentos"))
	pdf.Ln(6)

	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(8)

	// Table header with brand color
	pdf.SetFillColor(22, 163, 74) // #16a34a
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(15, 7, tr("ID"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, tr("Paciente"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 7, tr("Profissional"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 7, tr("Descricao"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 7, tr("Valor"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(25, 7, tr("Status"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, tr("Data"), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	// Table rows
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 8)
	totalValue := 0.0

	for _, budget := range budgets {
		// Check if need new page
		if pdf.GetY() > 180 {
			pdf.AddPage()
			// Repeat header with brand color
			pdf.SetFillColor(22, 163, 74) // #16a34a
			pdf.SetTextColor(255, 255, 255)
			pdf.SetFont("Arial", "B", 9)
			pdf.CellFormat(15, 7, tr("ID"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(60, 7, tr("Paciente"), "1", 0, "L", true, 0, "")
			pdf.CellFormat(50, 7, tr("Profissional"), "1", 0, "L", true, 0, "")
			pdf.CellFormat(70, 7, tr("Descricao"), "1", 0, "L", true, 0, "")
			pdf.CellFormat(25, 7, tr("Valor"), "1", 0, "R", true, 0, "")
			pdf.CellFormat(25, 7, tr("Status"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(25, 7, tr("Data"), "1", 0, "C", true, 0, "")
			pdf.Ln(-1)
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont("Arial", "", 8)
		}

		// Fetch patient name using raw SQL
		patientName := "N/A"
		var tmpPatientName string
		if err := db.Raw("SELECT name FROM patients WHERE id = ? AND deleted_at IS NULL", budget.PatientID).Scan(&tmpPatientName).Error; err == nil && tmpPatientName != "" {
			patientName = tmpPatientName
			if len(patientName) > 30 {
				patientName = patientName[:27] + "..."
			}
		}

		// Fetch dentist name using raw SQL from public schema
		dentistName := "N/A"
		var tmpDentistName string
		if err := db.Raw("SELECT name FROM public.users WHERE id = ? AND deleted_at IS NULL", budget.DentistID).Scan(&tmpDentistName).Error; err == nil && tmpDentistName != "" {
			dentistName = tmpDentistName
			if len(dentistName) > 25 {
				dentistName = dentistName[:22] + "..."
			}
		}

		description := budget.Description
		if len(description) > 35 {
			description = description[:32] + "..."
		}

		statusLabel := getStatusLabel(budget.Status)

		pdf.CellFormat(15, 6, fmt.Sprintf("%d", budget.ID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(60, 6, tr(patientName), "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 6, tr(dentistName), "1", 0, "L", false, 0, "")
		pdf.CellFormat(70, 6, tr(description), "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("R$ %.2f", budget.TotalValue), "1", 0, "R", false, 0, "")
		pdf.CellFormat(25, 6, tr(statusLabel), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, budget.CreatedAt.Format("02/01/2006"), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)

		totalValue += budget.TotalValue
	}

	// Total row
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(245, 7, tr("TOTAL"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(25, 7, fmt.Sprintf("R$ %.2f", totalValue), "1", 0, "R", true, 0, "")
	pdf.Ln(10)

	// Summary
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 5, fmt.Sprintf("Total de orcamentos: %d", len(budgets)))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=orcamentos_lista_%s.pdf", time.Now().Format("20060102_150405")))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}
