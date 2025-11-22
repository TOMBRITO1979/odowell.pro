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

// ExportPatientsCSV exports all patients to CSV format (no limit)
func ExportPatientsCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Build SQL query - get ALL patients (no limit)
	sqlQuery := "SELECT * FROM patients WHERE deleted_at IS NULL ORDER BY name ASC"

	// Load all patients
	var patients []models.Patient
	if err := db.Raw(sqlQuery).Scan(&patients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patients"})
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=pacientes_%s.csv", time.Now().Format("20060102_150405")))

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write BOM for UTF-8
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	// Write header
	header := []string{
		"Nome",
		"CPF",
		"RG",
		"Data Nascimento",
		"Gênero",
		"Email",
		"Telefone",
		"Celular",
		"Endereço",
		"Número",
		"Complemento",
		"Bairro",
		"Cidade",
		"Estado",
		"CEP",
		"Alergias",
		"Medicamentos",
		"Doenças Sistêmicas",
		"Observações",
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
		return
	}

	// Write data rows
	for _, patient := range patients {
		birthDate := ""
		if patient.BirthDate != nil {
			birthDate = patient.BirthDate.Format("2006-01-02")
		}

		row := []string{
			patient.Name,
			patient.CPF,
			patient.RG,
			birthDate,
			patient.Gender,
			patient.Email,
			patient.Phone,
			patient.CellPhone,
			patient.Address,
			patient.Number,
			patient.Complement,
			patient.District,
			patient.City,
			patient.State,
			patient.ZipCode,
			patient.Allergies,
			patient.Medications,
			patient.SystemicDiseases,
			patient.Notes,
		}

		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV row"})
			return
		}
	}
}

// ImportPatientsCSV imports patients from CSV file
func ImportPatientsCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

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

		if len(record) < 3 {
			errors = append(errors, fmt.Sprintf("Linha %d: Dados insuficientes (mínimo: nome, CPF, celular)", lineNum))
			continue
		}

		name := strings.TrimSpace(record[0])
		cpf := strings.TrimSpace(record[1])
		cellPhone := strings.TrimSpace(record[2])

		if name == "" {
			errors = append(errors, fmt.Sprintf("Linha %d: Nome é obrigatório", lineNum))
			continue
		}

		// Optional fields
		rg := ""
		if len(record) > 3 {
			rg = strings.TrimSpace(record[3])
		}

		var birthDate *time.Time
		if len(record) > 4 && record[4] != "" {
			bd, err := time.Parse("2006-01-02", strings.TrimSpace(record[4]))
			if err == nil {
				birthDate = &bd
			}
		}

		gender := ""
		if len(record) > 5 {
			gender = strings.TrimSpace(record[5])
		}

		email := ""
		if len(record) > 6 {
			email = strings.TrimSpace(record[6])
		}

		phone := ""
		if len(record) > 7 {
			phone = strings.TrimSpace(record[7])
		}

		address := ""
		if len(record) > 8 {
			address = strings.TrimSpace(record[8])
		}

		number := ""
		if len(record) > 9 {
			number = strings.TrimSpace(record[9])
		}

		complement := ""
		if len(record) > 10 {
			complement = strings.TrimSpace(record[10])
		}

		district := ""
		if len(record) > 11 {
			district = strings.TrimSpace(record[11])
		}

		city := ""
		if len(record) > 12 {
			city = strings.TrimSpace(record[12])
		}

		state := ""
		if len(record) > 13 {
			state = strings.TrimSpace(record[13])
		}

		zipCode := ""
		if len(record) > 14 {
			zipCode = strings.TrimSpace(record[14])
		}

		allergies := ""
		if len(record) > 15 {
			allergies = strings.TrimSpace(record[15])
		}

		medications := ""
		if len(record) > 16 {
			medications = strings.TrimSpace(record[16])
		}

		systemicDiseases := ""
		if len(record) > 17 {
			systemicDiseases = strings.TrimSpace(record[17])
		}

		notes := ""
		if len(record) > 18 {
			notes = strings.TrimSpace(record[18])
		}

		// Create patient
		patient := models.Patient{
			Name:             name,
			CPF:              cpf,
			RG:               rg,
			BirthDate:        birthDate,
			Gender:           gender,
			Email:            email,
			Phone:            phone,
			CellPhone:        cellPhone,
			Address:          address,
			Number:           number,
			Complement:       complement,
			District:         district,
			City:             city,
			State:            state,
			ZipCode:          zipCode,
			Allergies:        allergies,
			Medications:      medications,
			SystemicDiseases: systemicDiseases,
			Notes:            notes,
		}

		if err := db.Create(&patient).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Linha %d: Erro ao criar paciente - %v", lineNum, err))
			continue
		}

		imported++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Importação concluída: %d pacientes importados", imported),
		"imported": imported,
		"errors":   errors,
	})
}

// GeneratePatientsListPDF generates a PDF with the current filtered list of patients
func GeneratePatientsListPDF(c *gin.Context) {
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

	// Build SQL query with filters (only current page/filters from frontend)
	sqlQuery := "SELECT * FROM patients WHERE deleted_at IS NULL"
	var args []interface{}

	if searchTerm := c.Query("search"); searchTerm != "" {
		sqlQuery += " AND (name ILIKE ? OR cpf ILIKE ? OR email ILIKE ?)"
		searchPattern := "%" + searchTerm + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	sqlQuery += " ORDER BY name ASC"

	// Add pagination to match what's shown on screen
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}
	if ps := c.Query("page_size"); ps != "" {
		pageSize, _ = strconv.Atoi(ps)
	}
	offset := (page - 1) * pageSize
	sqlQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	// Load patients
	var patients []models.Patient
	if err := db.Raw(sqlQuery, args...).Scan(&patients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patients"})
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
	pdf.Cell(0, 8, tr("Lista de Pacientes"))
	pdf.Ln(6)

	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(8)

	// Table header with brand color
	pdf.SetFillColor(22, 163, 74) // #16a34a
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(10, 7, tr("ID"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 7, tr("Nome"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 7, tr("CPF"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 7, tr("Celular"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(55, 7, tr("Email"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 7, tr("Cidade"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 7, tr("Data Cadastro"), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	// Table rows
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 7)

	for _, patient := range patients {
		// Check if need new page
		if pdf.GetY() > 180 {
			pdf.AddPage()
			// Repeat header with brand color
			pdf.SetFillColor(22, 163, 74) // #16a34a
			pdf.SetTextColor(255, 255, 255)
			pdf.SetFont("Arial", "B", 8)
			pdf.CellFormat(10, 7, tr("ID"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(60, 7, tr("Nome"), "1", 0, "L", true, 0, "")
			pdf.CellFormat(30, 7, tr("CPF"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 7, tr("Celular"), "1", 0, "C", true, 0, "")
			pdf.CellFormat(55, 7, tr("Email"), "1", 0, "L", true, 0, "")
			pdf.CellFormat(50, 7, tr("Cidade"), "1", 0, "L", true, 0, "")
			pdf.CellFormat(35, 7, tr("Data Cadastro"), "1", 0, "C", true, 0, "")
			pdf.Ln(-1)
			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont("Arial", "", 7)
		}

		name := patient.Name
		if len(name) > 35 {
			name = name[:32] + "..."
		}

		email := patient.Email
		if len(email) > 30 {
			email = email[:27] + "..."
		}

		city := patient.City
		if len(city) > 25 {
			city = city[:22] + "..."
		}

		pdf.CellFormat(10, 6, fmt.Sprintf("%d", patient.ID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(60, 6, tr(name), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, patient.CPF, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 6, patient.CellPhone, "1", 0, "C", false, 0, "")
		pdf.CellFormat(55, 6, tr(email), "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 6, tr(city), "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, patient.CreatedAt.Format("02/01/2006"), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	// Summary
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(0, 5, fmt.Sprintf("Total de pacientes nesta pagina: %d", len(patients)))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=pacientes_lista_%s.pdf", time.Now().Format("20060102_150405")))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}
