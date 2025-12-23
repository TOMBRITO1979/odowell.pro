package handlers

import (
	"drcrwell/backend/internal/models"
	"drcrwell/backend/internal/middleware"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

func GeneratePaymentReceipt(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")
	budgetID := c.Param("id")
	paymentID := c.Param("payment_id")

	// Get tenant info
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	// Get budget
	var budget models.Budget
	if err := db.Session(&gorm.Session{NewDB: true}).
		Preload("Patient").
		Where("id = ?", budgetID).
		First(&budget).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	// Get payment
	var payment models.Payment
	if err := db.Session(&gorm.Session{NewDB: true}).
		Where("id = ? AND budget_id = ?", paymentID, budgetID).
		First(&payment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Get all payments for this budget to calculate totals
	var allPayments []models.Payment
	db.Session(&gorm.Session{NewDB: true}).
		Where("budget_id = ? AND status = ?", budgetID, "paid").
		Find(&allPayments)

	// Calculate totals
	var totalPaid float64
	for _, p := range allPayments {
		totalPaid += p.Amount
	}
	remainingBalance := budget.TotalValue - totalPaid

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, tr(tenant.Name))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, tr(tenant.Address+", "+tenant.City+" - "+tenant.State))
	pdf.Ln(5)
	pdf.Cell(0, 5, tr("Tel: "+tenant.Phone))
	pdf.Ln(15)

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, tr("RECIBO DE PAGAMENTO"))
	pdf.Ln(12)

	// Receipt number and date
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Recibo No: %d", payment.ID))
	pdf.Ln(5)
	receiptDate := time.Now()
	if payment.PaidDate != nil {
		receiptDate = *payment.PaidDate
	}
	pdf.Cell(0, 6, fmt.Sprintf("Data: %s", receiptDate.Format("02/01/2006")))
	pdf.Ln(10)

	// Received from section
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Dados do Pagamento"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(60, 6, tr("Recebemos de:"), "1", 0, "L", false, 0, "")
	patientName := "N/A"
	if budget.Patient != nil {
		patientName = budget.Patient.Name
	}
	pdf.CellFormat(120, 6, tr(patientName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Referente a:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(120, 6, tr(fmt.Sprintf("Orcamento No %s", budgetID)), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Forma de Pagamento:"), "1", 0, "L", false, 0, "")
	paymentMethodLabel := getPaymentMethodLabel(payment.PaymentMethod)
	pdf.CellFormat(120, 6, tr(paymentMethodLabel), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	if payment.IsInstallment {
		pdf.CellFormat(60, 6, tr("Parcela:"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(120, 6, fmt.Sprintf("%d/%d", payment.InstallmentNumber, payment.TotalInstallments), "1", 0, "L", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(60, 8, tr("Valor Recebido:"), "1", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 128, 0)
	pdf.CellFormat(120, 8, fmt.Sprintf("R$ %.2f", payment.Amount), "1", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(12)

	// Budget summary
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Resumo do Orcamento"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(90, 6, tr("Valor Total do Tratamento:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, fmt.Sprintf("R$ %.2f", budget.TotalValue), "1", 0, "R", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(90, 6, tr("Total Pago ate o momento:"), "1", 0, "L", false, 0, "")
	pdf.SetTextColor(0, 128, 0)
	pdf.CellFormat(90, 6, fmt.Sprintf("R$ %.2f", totalPaid), "1", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(-1)

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(90, 6, tr("Saldo Restante:"), "1", 0, "L", false, 0, "")
	if remainingBalance > 0 {
		pdf.SetTextColor(255, 0, 0)
	} else {
		pdf.SetTextColor(0, 128, 0)
	}
	pdf.CellFormat(90, 6, fmt.Sprintf("R$ %.2f", remainingBalance), "1", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(15)

	// Notes
	if payment.Notes != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 6, tr("Observacoes:"))
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(payment.Notes), "", "L", false)
		pdf.Ln(5)
	}

	// Signature
	pdf.Ln(20)
	pdf.Line(15, pdf.GetY(), 100, pdf.GetY())
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 5, tr("Assinatura do Responsavel"))

	// Footer
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Recibo gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=recibo_%d.pdf", payment.ID))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate receipt"})
		return
	}
}
