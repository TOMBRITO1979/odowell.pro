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
)

// ExportPaymentsCSV exports payments to CSV format
func ExportPaymentsCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Build SQL query with filters
	sqlQuery := "SELECT * FROM payments WHERE deleted_at IS NULL"
	var args []interface{}

	if patientID := c.Query("patient_id"); patientID != "" {
		sqlQuery += " AND patient_id = ?"
		args = append(args, patientID)
	}
	if budgetID := c.Query("budget_id"); budgetID != "" {
		sqlQuery += " AND budget_id = ?"
		args = append(args, budgetID)
	}
	if paymentType := c.Query("type"); paymentType != "" {
		sqlQuery += " AND type = ?"
		args = append(args, paymentType)
	}
	if category := c.Query("category"); category != "" {
		sqlQuery += " AND category = ?"
		args = append(args, category)
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

	// Load payments
	var payments []models.Payment
	if err := db.Raw(sqlQuery, args...).Scan(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}

	// Set headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=pagamentos_%s.csv", time.Now().Format("20060102_150405")))

	// Create CSV writer
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID",
		"Paciente",
		"Orçamento ID",
		"Tipo",
		"Categoria",
		"Descrição",
		"Valor",
		"Método Pagamento",
		"Parcela",
		"Data Criação",
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV header"})
		return
	}

	// Write data rows
	for _, payment := range payments {
		// Fetch patient name using raw SQL
		patientName := ""
		var tmpPatientName string
		if err := db.Raw("SELECT name FROM patients WHERE id = ? AND deleted_at IS NULL", payment.PatientID).Scan(&tmpPatientName).Error; err == nil {
			patientName = tmpPatientName
		}

		budgetID := ""
		if payment.BudgetID != nil {
			budgetID = fmt.Sprintf("%d", *payment.BudgetID)
		}

		installmentInfo := ""
		if payment.IsInstallment {
			installmentInfo = fmt.Sprintf("%d/%d", payment.InstallmentNumber, payment.TotalInstallments)
		}

		row := []string{
			fmt.Sprintf("%d", payment.ID),
			patientName,
			budgetID,
			payment.Type,
			payment.Category,
			payment.Description,
			fmt.Sprintf("%.2f", payment.Amount),
			payment.PaymentMethod,
			installmentInfo,
			payment.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if err := writer.Write(row); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write CSV row"})
			return
		}
	}
}

// ImportPaymentsCSV imports payments from CSV file
func ImportPaymentsCSV(c *gin.Context) {
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

		if len(record) < 7 {
			errors = append(errors, fmt.Sprintf("Linha %d: Dados insuficientes", lineNum))
			continue
		}

		// Parse required fields
		patientID, err := strconv.ParseUint(record[0], 10, 32)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Linha %d: ID do paciente inválido", lineNum))
			continue
		}

		amount, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Linha %d: Valor inválido", lineNum))
			continue
		}

		paymentType := record[1]
		category := record[2]
		description := record[3]
		paymentMethod := record[4]

		// Validate payment type
		validTypes := map[string]bool{
			"income":  true,
			"expense": true,
		}
		if !validTypes[paymentType] {
			paymentType = "income"
		}

		// Optional budget ID
		var budgetID *uint
		if len(record) > 6 && record[6] != "" {
			bid, err := strconv.ParseUint(record[6], 10, 32)
			if err == nil {
				budgetIDVal := uint(bid)
				budgetID = &budgetIDVal
			}
		}

		// Create payment
		payment := models.Payment{
			PatientID:     uint(patientID),
			BudgetID:      budgetID,
			Type:          paymentType,
			Category:      category,
			Description:   description,
			Amount:        amount,
			PaymentMethod: paymentMethod,
			IsInstallment: false,
		}

		if err := db.Create(&payment).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Linha %d: Erro ao criar pagamento - %v", lineNum, err))
			continue
		}

		imported++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Importação concluída: %d pagamentos importados", imported),
		"imported": imported,
		"errors":   errors,
	})
}
