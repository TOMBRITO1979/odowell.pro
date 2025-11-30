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

type BudgetItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

func GenerateBudgetPDF(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint("tenant_id")
	budgetID := c.Param("id")

	// Get clinic info from settings (dynamic) or tenant (fallback)
	clinic := GetClinicInfo(db, tenantID)

	// Get budget with patient and dentist info
	var budget models.Budget
	if err := db.Session(&gorm.Session{NewDB: true}).
		Preload("Patient").
		Preload("Dentist").
		Where("id = ?", budgetID).
		First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	// Parse items from JSON
	var items []BudgetItem
	if budget.Items != nil && *budget.Items != "" {
		if err := json.Unmarshal([]byte(*budget.Items), &items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse budget items"})
			return
		}
	}

	// Create PDF with proper margins for A4
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, tr(clinic.Name))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 9)
	if clinic.Address != "" {
		pdf.Cell(0, 5, tr(clinic.Address))
		pdf.Ln(5)
	}
	if clinic.Phone != "" {
		pdf.Cell(0, 5, tr("Tel: "+clinic.Phone))
		pdf.Ln(5)
	}
	pdf.Ln(5)

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, tr("Orcamento"))
	pdf.Ln(10)

	// Budget info
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Informacoes do Orcamento"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(60, 6, tr("Numero:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(120, 6, fmt.Sprintf("#%d", budget.ID), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Paciente:"), "1", 0, "L", false, 0, "")
	patientName := "N/A"
	if budget.Patient != nil {
		patientName = budget.Patient.Name
	}
	pdf.CellFormat(120, 6, tr(patientName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Profissional:"), "1", 0, "L", false, 0, "")
	dentistName := "N/A"
	if budget.Dentist != nil {
		dentistName = budget.Dentist.Name
	}
	pdf.CellFormat(120, 6, tr(dentistName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Data de Criacao:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(120, 6, budget.CreatedAt.Format("02/01/2006 15:04"), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Status:"), "1", 0, "L", false, 0, "")
	statusLabel := getStatusLabel(budget.Status)
	pdf.CellFormat(120, 6, tr(statusLabel), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	if budget.ValidUntil != nil {
		pdf.CellFormat(60, 6, tr("Valido Ate:"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(120, 6, budget.ValidUntil.Format("02/01/2006"), "1", 0, "L", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(5)

	// Description
	if budget.Description != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(180, 7, tr("Descricao"), "1", 0, "L", true, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(budget.Description), "1", "L", false)
		pdf.Ln(5)
	}

	// Items table
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Itens do Orcamento"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(80, 6, tr("Descricao"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 6, tr("Qtd"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 6, tr("Valor Unit."), "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 6, tr("Total"), "1", 0, "R", false, 0, "")
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 10)
	for _, item := range items {
		pdf.CellFormat(80, 6, tr(item.Description), "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("R$ %.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("R$ %.2f", item.Total), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	// Total
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(140, 7, tr("VALOR TOTAL"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 7, fmt.Sprintf("R$ %.2f", budget.TotalValue), "1", 0, "R", true, 0, "")
	pdf.Ln(10)

	// Notes
	if budget.Notes != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(180, 7, tr("Observacoes"), "1", 0, "L", true, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(budget.Notes), "1", "L", false)
		pdf.Ln(5)
	}

	// Footer
	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=orcamento_%d.pdf", budget.ID))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

func getStatusLabel(status string) string {
	labels := map[string]string{
		"pending":   "Pendente",
		"approved":  "Aprovado",
		"rejected":  "Rejeitado",
		"expired":   "Expirado",
		"cancelled": "Cancelado",
	}

	if label, ok := labels[status]; ok {
		return label
	}
	return status
}

func GenerateBudgetPaymentsPDF(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint("tenant_id")
	budgetID := c.Param("id")

	// Get clinic info from settings (dynamic) or tenant (fallback)
	clinic := GetClinicInfo(db, tenantID)

	// Get budget with patient, dentist and payments info
	var budget models.Budget
	if err := db.Session(&gorm.Session{NewDB: true}).
		Preload("Patient").
		Preload("Dentist").
		Preload("Payments").
		Where("id = ?", budgetID).
		First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	// Parse items from JSON
	var items []BudgetItem
	if budget.Items != nil && *budget.Items != "" {
		if err := json.Unmarshal([]byte(*budget.Items), &items); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse budget items"})
			return
		}
	}

	// Calculate payment totals
	var paidAmount, dueAmount float64
	paidAmount = 0
	for _, payment := range budget.Payments {
		if payment.Status == "paid" {
			paidAmount += payment.Amount
		}
	}
	dueAmount = budget.TotalValue - paidAmount

	// Create PDF with proper margins for A4
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, tr(clinic.Name))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 9)
	if clinic.Address != "" {
		pdf.Cell(0, 5, tr(clinic.Address))
		pdf.Ln(5)
	}
	if clinic.Phone != "" {
		pdf.Cell(0, 5, tr("Tel: "+clinic.Phone))
		pdf.Ln(5)
	}
	pdf.Ln(5)

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, tr("Orcamento e Gerenciamento de Pagamentos"))
	pdf.Ln(10)

	// Budget info
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Informacoes do Orcamento"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(60, 6, tr("Numero:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(120, 6, fmt.Sprintf("#%d", budget.ID), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Paciente:"), "1", 0, "L", false, 0, "")
	patientName := "N/A"
	if budget.Patient != nil {
		patientName = budget.Patient.Name
	}
	pdf.CellFormat(120, 6, tr(patientName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Profissional:"), "1", 0, "L", false, 0, "")
	dentistName := "N/A"
	if budget.Dentist != nil {
		dentistName = budget.Dentist.Name
	}
	pdf.CellFormat(120, 6, tr(dentistName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Data de Criacao:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(120, 6, budget.CreatedAt.Format("02/01/2006 15:04"), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Status:"), "1", 0, "L", false, 0, "")
	statusLabel := getStatusLabel(budget.Status)
	pdf.CellFormat(120, 6, tr(statusLabel), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	if budget.ValidUntil != nil {
		pdf.CellFormat(60, 6, tr("Valido Ate:"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(120, 6, budget.ValidUntil.Format("02/01/2006"), "1", 0, "L", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(5)

	// Description
	if budget.Description != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(180, 7, tr("Descricao"), "1", 0, "L", true, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(budget.Description), "1", "L", false)
		pdf.Ln(5)
	}

	// Items table
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Itens do Orcamento"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(80, 6, tr("Descricao"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 6, tr("Qtd"), "1", 0, "C", false, 0, "")
	pdf.CellFormat(40, 6, tr("Valor Unit."), "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 6, tr("Total"), "1", 0, "R", false, 0, "")
	pdf.Ln(-1)

	// Table rows
	pdf.SetFont("Arial", "", 10)
	for _, item := range items {
		pdf.CellFormat(80, 6, tr(item.Description), "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 6, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("R$ %.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("R$ %.2f", item.Total), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	// Total
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(140, 7, tr("VALOR TOTAL"), "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 7, fmt.Sprintf("R$ %.2f", budget.TotalValue), "1", 0, "R", true, 0, "")
	pdf.Ln(10)

	// Payment Management Section
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(24, 144, 255)
	pdf.Cell(0, 8, tr("Gerenciamento de Pagamentos"))
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(10)

	// Payment summary
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Resumo Financeiro"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(90, 6, tr("Valor Total:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, fmt.Sprintf("R$ %.2f", budget.TotalValue), "1", 0, "R", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(90, 6, tr("Valor Pago:"), "1", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 128, 0)
	pdf.CellFormat(90, 6, fmt.Sprintf("R$ %.2f", paidAmount), "1", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(-1)

	pdf.CellFormat(90, 6, tr("Valor Devido:"), "1", 0, "L", false, 0, "")
	if dueAmount > 0 {
		pdf.SetTextColor(255, 0, 0)
	} else {
		pdf.SetTextColor(0, 128, 0)
	}
	pdf.CellFormat(90, 6, fmt.Sprintf("R$ %.2f", dueAmount), "1", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(10)

	// Payments list
	if len(budget.Payments) > 0 {
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(180, 7, tr("Lista de Pagamentos"), "1", 0, "L", true, 0, "")
		pdf.Ln(-1)

		// Table header
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(25, 6, tr("Data"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 6, tr("Descricao"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, tr("Metodo"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, tr("Valor"), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 6, tr("Status"), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)

		// Table rows
		pdf.SetFont("Arial", "", 9)
		for _, payment := range budget.Payments {
			// Check if we need a new page
			if pdf.GetY() > 250 {
				pdf.AddPage()
				// Repeat header
				pdf.SetFont("Arial", "B", 10)
				pdf.CellFormat(25, 6, tr("Data"), "1", 0, "C", false, 0, "")
				pdf.CellFormat(50, 6, tr("Descricao"), "1", 0, "L", false, 0, "")
				pdf.CellFormat(30, 6, tr("Metodo"), "1", 0, "C", false, 0, "")
				pdf.CellFormat(35, 6, tr("Valor"), "1", 0, "R", false, 0, "")
				pdf.CellFormat(40, 6, tr("Status"), "1", 0, "C", false, 0, "")
				pdf.Ln(-1)
				pdf.SetFont("Arial", "", 9)
			}

			dateStr := ""
			if payment.PaidDate != nil {
				dateStr = payment.PaidDate.Format("02/01/2006")
			} else if payment.DueDate != nil {
				dateStr = payment.DueDate.Format("02/01/2006")
			} else {
				dateStr = payment.CreatedAt.Format("02/01/2006")
			}

			pdf.CellFormat(25, 5, dateStr, "1", 0, "C", false, 0, "")

			description := payment.Description
			if len(description) > 25 {
				description = description[:25] + "..."
			}
			pdf.CellFormat(50, 5, tr(description), "1", 0, "L", false, 0, "")

			method := getBudgetPaymentMethodLabel(payment.PaymentMethod)
			pdf.CellFormat(30, 5, tr(method), "1", 0, "C", false, 0, "")

			pdf.CellFormat(35, 5, fmt.Sprintf("R$ %.2f", payment.Amount), "1", 0, "R", false, 0, "")

			status := getBudgetPaymentStatusLabel(payment.Status)
			pdf.CellFormat(40, 5, tr(status), "1", 0, "C", false, 0, "")

			pdf.Ln(-1)
		}
	} else {
		pdf.SetFont("Arial", "I", 10)
		pdf.Cell(0, 6, tr("Nenhum pagamento registrado"))
		pdf.Ln(6)
	}

	// Notes
	if budget.Notes != "" {
		pdf.Ln(5)
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(180, 7, tr("Observacoes"), "1", 0, "L", true, 0, "")
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(budget.Notes), "1", "L", false)
	}

	// Footer
	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=orcamento_pagamentos_%d.pdf", budget.ID))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

func getBudgetPaymentMethodLabel(method string) string {
	methods := map[string]string{
		"cash":        "Dinheiro",
		"credit_card": "Cartao Credito",
		"debit_card":  "Cartao Debito",
		"pix":         "PIX",
		"transfer":    "Transferencia",
		"insurance":   "Convenio",
	}

	if label, ok := methods[method]; ok {
		return label
	}
	return method
}

func getBudgetPaymentStatusLabel(status string) string {
	statuses := map[string]string{
		"pending":   "Pendente",
		"paid":      "Pago",
		"overdue":   "Atrasado",
		"cancelled": "Cancelado",
		"refunded":  "Reembolsado",
	}

	if label, ok := statuses[status]; ok {
		return label
	}
	return status
}
