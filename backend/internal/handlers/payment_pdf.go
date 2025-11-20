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

func GeneratePaymentsPDF(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint("tenant_id")

	// Get tenant info for header
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	// Get filters from query params
	patientID := c.Query("patient_id")
	paymentType := c.Query("type")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Build query
	query := db.Session(&gorm.Session{NewDB: true}).
		Table("payments").
		Preload("Patient")

	if patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}
	if paymentType != "" {
		query = query.Where("type = ?", paymentType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	// Get payments
	var payments []models.Payment
	if err := query.Order("created_at DESC").Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}

	// Calculate totals
	var totalIncome, totalExpenses float64
	for _, payment := range payments {
		if payment.Type == "income" && payment.Status == "paid" {
			totalIncome += payment.Amount
		} else if payment.Type == "expense" && payment.Status == "paid" {
			totalExpenses += payment.Amount
		}
	}
	balance := totalIncome - totalExpenses

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
	pdf.Cell(0, 8, tr("Relatorio de Pagamentos"))
	pdf.Ln(10)

	// Filter info
	if startDate != "" || endDate != "" || patientID != "" || paymentType != "" || status != "" {
		pdf.SetFont("Arial", "I", 9)
		filterText := "Filtros aplicados: "
		if startDate != "" && endDate != "" {
			filterText += fmt.Sprintf("Periodo: %s a %s", startDate, endDate)
		}
		if paymentType != "" {
			if paymentType == "income" {
				filterText += " | Tipo: Receita"
			} else {
				filterText += " | Tipo: Despesa"
			}
		}
		if status != "" {
			filterText += " | Status: " + getPaymentStatusLabel(status)
		}
		pdf.Cell(0, 5, tr(filterText))
		pdf.Ln(8)
	}

	// Statistics
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Resumo Financeiro"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(90, 6, tr("Total de Receitas:"), "1", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 128, 0)
	pdf.CellFormat(90, 6, fmt.Sprintf("R$ %.2f", totalIncome), "1", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(-1)

	pdf.CellFormat(90, 6, tr("Total de Despesas:"), "1", 0, "L", false, 0, "")
	pdf.SetTextColor(255, 0, 0)
	pdf.CellFormat(90, 6, fmt.Sprintf("R$ %.2f", totalExpenses), "1", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(-1)

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(90, 6, tr("Saldo:"), "1", 0, "L", false, 0, "")
	if balance >= 0 {
		pdf.SetTextColor(0, 128, 0)
	} else {
		pdf.SetTextColor(255, 0, 0)
	}
	pdf.CellFormat(90, 6, fmt.Sprintf("R$ %.2f", balance), "1", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(10)

	// Payments table
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Lista de Pagamentos"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	// Table header
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(25, 6, tr("Data"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(45, 6, tr("Paciente"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 6, tr("Tipo"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 6, tr("Valor"), "1", 0, "R", false, 0, "")
	pdf.CellFormat(25, 6, tr("Metodo"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(25, 6, tr("Status"), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 8)
	for _, payment := range payments {
		// Check if we need a new page
		if pdf.GetY() > 250 {
			pdf.AddPage()
			// Repeat header
			pdf.SetFont("Arial", "B", 9)
			pdf.CellFormat(25, 6, tr("Data"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(45, 6, tr("Paciente"), "1", 0, "L", false, 0, "")
			pdf.CellFormat(30, 6, tr("Tipo"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 6, tr("Valor"), "1", 0, "R", false, 0, "")
			pdf.CellFormat(25, 6, tr("Metodo"), "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 6, tr("Status"), "1", 0, "C", false, 0, "")
			pdf.Ln(-1)
			pdf.SetFont("Arial", "", 8)
		}

		pdf.CellFormat(25, 5, payment.CreatedAt.Format("02/01/2006"), "1", 0, "C", false, 0, "")

		patientName := "-"
		if payment.Patient != nil {
			patientName = payment.Patient.Name
			if len(patientName) > 20 {
				patientName = patientName[:20] + "..."
			}
		}
		pdf.CellFormat(45, 5, tr(patientName), "1", 0, "L", false, 0, "")

		typeLabel := "Receita"
		if payment.Type == "expense" {
			typeLabel = "Despesa"
		}
		pdf.CellFormat(30, 5, tr(typeLabel), "1", 0, "C", false, 0, "")

		pdf.CellFormat(30, 5, fmt.Sprintf("R$ %.2f", payment.Amount), "1", 0, "R", false, 0, "")
		pdf.CellFormat(25, 5, tr(getPaymentMethodLabel(payment.PaymentMethod)), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 5, tr(getPaymentStatusLabel(payment.Status)), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	// Footer
	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Total de pagamentos: %d | Gerado em: %s", len(payments), time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=relatorio_pagamentos.pdf")

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

func getPaymentMethodLabel(method string) string {
	methods := map[string]string{
		"cash":        "Dinheiro",
		"credit_card": "Credito",
		"debit_card":  "Debito",
		"pix":         "PIX",
		"transfer":    "Transf.",
		"insurance":   "Convenio",
	}

	if label, ok := methods[method]; ok {
		return label
	}
	return method
}

func getPaymentStatusLabel(status string) string {
	statuses := map[string]string{
		"pending":   "Pendente",
		"paid":      "Pago",
		"overdue":   "Atrasado",
		"cancelled": "Cancelado",
	}

	if label, ok := statuses[status]; ok {
		return label
	}
	return status
}
