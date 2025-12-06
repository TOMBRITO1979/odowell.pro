package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// ============================================
// TREATMENTS CRUD
// ============================================

// CreateTreatment - Criar novo tratamento (normalmente a partir de um orçamento aprovado)
func CreateTreatment(c *gin.Context) {
	var input struct {
		BudgetID          uint    `json:"budget_id" binding:"required"`
		TotalInstallments int     `json:"total_installments"`
		Notes             string  `json:"notes"`
		ExpectedEndDate   *string `json:"expected_end_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Check if budget exists and is approved
	var budget models.Budget
	if err := db.Preload("Patient").First(&budget, input.BudgetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Orçamento não encontrado"})
		return
	}

	if budget.Status != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Apenas orçamentos aprovados podem virar tratamentos"})
		return
	}

	// Check if treatment already exists for this budget
	var existingTreatment models.Treatment
	if err := db.Where("budget_id = ?", input.BudgetID).First(&existingTreatment).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Já existe um tratamento para este orçamento"})
		return
	}

	// Calculate installment value
	totalInstallments := input.TotalInstallments
	if totalInstallments <= 0 {
		totalInstallments = 1
	}
	installmentValue := budget.TotalValue / float64(totalInstallments)

	// Parse expected end date
	var expectedEndDate *time.Time
	if input.ExpectedEndDate != nil && *input.ExpectedEndDate != "" {
		parsed, err := time.Parse("2006-01-02", *input.ExpectedEndDate)
		if err == nil {
			expectedEndDate = &parsed
		}
	}

	treatment := models.Treatment{
		BudgetID:          input.BudgetID,
		PatientID:         budget.PatientID,
		DentistID:         budget.DentistID,
		Description:       budget.Description,
		TotalValue:        budget.TotalValue,
		PaidValue:         0,
		TotalInstallments: totalInstallments,
		InstallmentValue:  installmentValue,
		Status:            models.TreatmentStatusInProgress,
		StartDate:         time.Now(),
		ExpectedEndDate:   expectedEndDate,
		Notes:             input.Notes,
	}

	if err := db.Create(&treatment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar tratamento"})
		return
	}

	// Load relationships
	db.Preload("Patient").Preload("Dentist").Preload("Budget").First(&treatment, treatment.ID)

	c.JSON(http.StatusCreated, gin.H{"treatment": treatment})
}

// GetTreatments - Listar todos os tratamentos
func GetTreatments(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var treatments []models.Treatment
	query := db.Preload("Patient").Preload("Dentist").Preload("Budget")

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by patient
	if patientID := c.Query("patient_id"); patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}

	// Filter by dentist
	if dentistID := c.Query("dentist_id"); dentistID != "" {
		query = query.Where("dentist_id = ?", dentistID)
	}

	// Filter by budget_id
	if budgetID := c.Query("budget_id"); budgetID != "" {
		query = query.Where("budget_id = ?", budgetID)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var total int64
	query.Model(&models.Treatment{}).Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&treatments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar tratamentos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"treatments": treatments,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// GetTreatment - Buscar um tratamento específico
func GetTreatment(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var treatment models.Treatment
	if err := db.Preload("Patient").Preload("Dentist").Preload("Budget").Preload("TreatmentPayments", func(db *gorm.DB) *gorm.DB {
		return db.Order("paid_date DESC")
	}).Preload("TreatmentPayments.ReceivedBy").First(&treatment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tratamento não encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"treatment": treatment})
}

// UpdateTreatment - Atualizar tratamento
func UpdateTreatment(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var treatment models.Treatment
	if err := db.First(&treatment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tratamento não encontrado"})
		return
	}

	var input struct {
		TotalInstallments int     `json:"total_installments"`
		Status            string  `json:"status"`
		Notes             string  `json:"notes"`
		ExpectedEndDate   *string `json:"expected_end_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if input.TotalInstallments > 0 {
		treatment.TotalInstallments = input.TotalInstallments
		treatment.InstallmentValue = treatment.TotalValue / float64(input.TotalInstallments)
	}

	if input.Status != "" {
		treatment.Status = input.Status
		if input.Status == models.TreatmentStatusCompleted {
			now := time.Now()
			treatment.CompletedDate = &now
		}
	}

	if input.Notes != "" {
		treatment.Notes = input.Notes
	}

	if input.ExpectedEndDate != nil {
		if *input.ExpectedEndDate == "" {
			treatment.ExpectedEndDate = nil
		} else {
			parsed, err := time.Parse("2006-01-02", *input.ExpectedEndDate)
			if err == nil {
				treatment.ExpectedEndDate = &parsed
			}
		}
	}

	// Use raw SQL to avoid GORM model contamination issue
	result := db.Exec(`
		UPDATE treatments SET
			updated_at = NOW(),
			total_installments = ?,
			installment_value = ?,
			status = ?,
			completed_date = ?,
			notes = ?,
			expected_end_date = ?
		WHERE id = ? AND deleted_at IS NULL
	`, treatment.TotalInstallments, treatment.InstallmentValue, treatment.Status,
		treatment.CompletedDate, treatment.Notes, treatment.ExpectedEndDate, treatment.ID)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar tratamento"})
		return
	}

	// Reload with relationships
	db.Preload("Patient").Preload("Dentist").Preload("Budget").First(&treatment, treatment.ID)

	c.JSON(http.StatusOK, gin.H{"treatment": treatment})
}

// DeleteTreatment - Deletar tratamento
func DeleteTreatment(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Check for existing payments
	var paymentCount int64
	db.Model(&models.TreatmentPayment{}).Where("treatment_id = ?", uint(id)).Count(&paymentCount)
	if paymentCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Não é possível deletar tratamento com pagamentos registrados",
			"count": paymentCount,
		})
		return
	}

	if err := db.Delete(&models.Treatment{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar tratamento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tratamento deletado com sucesso"})
}

// ============================================
// TREATMENT PAYMENTS
// ============================================

// generateReceiptNumber generates a unique receipt number
func generateReceiptNumber(db *gorm.DB) string {
	now := time.Now()
	prefix := now.Format("200601")

	var count int64
	// Use raw SQL to avoid GORM model contamination issues
	db.Raw("SELECT COUNT(*) FROM treatment_payments WHERE receipt_number LIKE ? AND deleted_at IS NULL", prefix+"%").Scan(&count)

	return fmt.Sprintf("%s%04d", prefix, count+1)
}

// CreateTreatmentPayment - Registrar um pagamento de tratamento
func CreateTreatmentPayment(c *gin.Context) {
	var input struct {
		TreatmentID       uint    `json:"treatment_id" binding:"required"`
		Amount            float64 `json:"amount" binding:"required"`
		PaymentMethod     string  `json:"payment_method" binding:"required"`
		InstallmentNumber int     `json:"installment_number"`
		Notes             string  `json:"notes"`
		PaidDate          *string `json:"paid_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Get user ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Check if treatment exists
	var treatment models.Treatment
	if err := db.First(&treatment, input.TreatmentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tratamento não encontrado"})
		return
	}

	if treatment.Status == models.TreatmentStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tratamento já foi finalizado"})
		return
	}

	if treatment.Status == models.TreatmentStatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tratamento foi cancelado"})
		return
	}

	// Parse paid date
	paidDate := time.Now()
	if input.PaidDate != nil && *input.PaidDate != "" {
		parsed, err := time.Parse("2006-01-02", *input.PaidDate)
		if err == nil {
			paidDate = parsed
		}
	}

	// Get next installment number if not provided
	installmentNumber := input.InstallmentNumber
	if installmentNumber <= 0 {
		var maxInstallment int
		db.Model(&models.TreatmentPayment{}).Where("treatment_id = ?", input.TreatmentID).Select("COALESCE(MAX(installment_number), 0)").Scan(&maxInstallment)
		installmentNumber = maxInstallment + 1
	}

	payment := models.TreatmentPayment{
		TreatmentID:       input.TreatmentID,
		Amount:            input.Amount,
		PaymentMethod:     input.PaymentMethod,
		InstallmentNumber: installmentNumber,
		ReceiptNumber:     generateReceiptNumber(db),
		Status:            models.TreatmentPaymentStatusPaid,
		PaidDate:          paidDate,
		ReceivedByID:      userID.(uint),
		Notes:             input.Notes,
	}

	// Use raw SQL to avoid foreign key issues with User table in public schema
	result := db.Exec(`
		INSERT INTO treatment_payments
		(created_at, updated_at, treatment_id, amount, payment_method, installment_number, receipt_number, status, paid_date, received_by_id, notes)
		VALUES (NOW(), NOW(), ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`, payment.TreatmentID, payment.Amount, payment.PaymentMethod, payment.InstallmentNumber,
		payment.ReceiptNumber, payment.Status, payment.PaidDate, payment.ReceivedByID, payment.Notes)

	if result.Error != nil {
		log.Printf("ERROR creating payment: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao registrar pagamento: " + result.Error.Error()})
		return
	}

	// Get the created payment ID
	var paymentID uint
	db.Raw("SELECT id FROM treatment_payments WHERE receipt_number = ? AND deleted_at IS NULL", payment.ReceiptNumber).Scan(&paymentID)
	payment.ID = paymentID

	// Update treatment paid value
	treatment.PaidValue += input.Amount

	// Check if treatment is fully paid
	if treatment.PaidValue >= treatment.TotalValue {
		treatment.Status = models.TreatmentStatusCompleted
		now := time.Now()
		treatment.CompletedDate = &now
		// Update with completed status
		db.Exec(`UPDATE treatments SET paid_value = ?, status = ?, completed_date = ?, updated_at = NOW() WHERE id = ?`,
			treatment.PaidValue, treatment.Status, treatment.CompletedDate, treatment.ID)
	} else {
		// Update only paid value
		db.Exec(`UPDATE treatments SET paid_value = ?, updated_at = NOW() WHERE id = ?`,
			treatment.PaidValue, treatment.ID)
	}

	// Load payment with created data
	db.Raw("SELECT * FROM treatment_payments WHERE id = ?", payment.ID).Scan(&payment)

	c.JSON(http.StatusCreated, gin.H{
		"payment":   payment,
		"treatment": treatment,
	})
}

// GetTreatmentPayments - Listar pagamentos de um tratamento
func GetTreatmentPayments(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	treatmentID, err := strconv.ParseUint(c.Param("treatment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de tratamento inválido"})
		return
	}

	var payments []models.TreatmentPayment
	if err := db.Preload("ReceivedBy").Where("treatment_id = ?", uint(treatmentID)).Order("paid_date DESC").Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar pagamentos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

// GetTreatmentPayment - Buscar um pagamento específico
func GetTreatmentPayment(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var payment models.TreatmentPayment
	if err := db.Preload("Treatment").Preload("Treatment.Patient").Preload("ReceivedBy").First(&payment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pagamento não encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

// UpdateTreatmentPayment - Atualizar pagamento
func UpdateTreatmentPayment(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var payment models.TreatmentPayment
	if err := db.First(&payment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pagamento não encontrado"})
		return
	}

	var input struct {
		PaymentMethod string  `json:"payment_method"`
		Notes         string  `json:"notes"`
		Status        string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldAmount := payment.Amount
	oldStatus := payment.Status

	if input.PaymentMethod != "" {
		payment.PaymentMethod = input.PaymentMethod
	}
	if input.Notes != "" {
		payment.Notes = input.Notes
	}
	if input.Status != "" {
		payment.Status = input.Status
	}

	// Use raw SQL to avoid GORM model contamination issue
	result := db.Exec(`
		UPDATE treatment_payments SET
			updated_at = NOW(),
			payment_method = ?,
			notes = ?,
			status = ?
		WHERE id = ? AND deleted_at IS NULL
	`, payment.PaymentMethod, payment.Notes, payment.Status, payment.ID)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar pagamento"})
		return
	}

	// Update treatment paid value if status changed
	if oldStatus != payment.Status {
		var treatment models.Treatment
		if err := db.Raw("SELECT * FROM treatments WHERE id = ? AND deleted_at IS NULL", payment.TreatmentID).Scan(&treatment).Error; err == nil && treatment.ID != 0 {
			if payment.Status == models.TreatmentPaymentStatusCancelled || payment.Status == models.TreatmentPaymentStatusRefunded {
				treatment.PaidValue -= oldAmount
				if treatment.Status == models.TreatmentStatusCompleted {
					treatment.Status = models.TreatmentStatusInProgress
					treatment.CompletedDate = nil
				}
			} else if oldStatus == models.TreatmentPaymentStatusCancelled || oldStatus == models.TreatmentPaymentStatusRefunded {
				treatment.PaidValue += payment.Amount
				if treatment.PaidValue >= treatment.TotalValue {
					treatment.Status = models.TreatmentStatusCompleted
					now := time.Now()
					treatment.CompletedDate = &now
				}
			}
			// Use raw SQL to update treatment
			db.Exec(`UPDATE treatments SET updated_at = NOW(), paid_value = ?, status = ?, completed_date = ? WHERE id = ? AND deleted_at IS NULL`,
				treatment.PaidValue, treatment.Status, treatment.CompletedDate, treatment.ID)
		}
	}

	db.Raw("SELECT * FROM treatment_payments WHERE id = ?", payment.ID).Scan(&payment)

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

// DeleteTreatmentPayment - Deletar pagamento
func DeleteTreatmentPayment(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var payment models.TreatmentPayment
	if err := db.Raw("SELECT * FROM treatment_payments WHERE id = ? AND deleted_at IS NULL", uint(id)).Scan(&payment).Error; err != nil || payment.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pagamento não encontrado"})
		return
	}

	// Update treatment paid value
	var treatment models.Treatment
	if err := db.Raw("SELECT * FROM treatments WHERE id = ? AND deleted_at IS NULL", payment.TreatmentID).Scan(&treatment).Error; err == nil && treatment.ID != 0 {
		if payment.Status == models.TreatmentPaymentStatusPaid {
			treatment.PaidValue -= payment.Amount
			if treatment.Status == models.TreatmentStatusCompleted {
				treatment.Status = models.TreatmentStatusInProgress
				treatment.CompletedDate = nil
			}
			// Use raw SQL to update treatment
			db.Exec(`UPDATE treatments SET updated_at = NOW(), paid_value = ?, status = ?, completed_date = ? WHERE id = ? AND deleted_at IS NULL`,
				treatment.PaidValue, treatment.Status, treatment.CompletedDate, treatment.ID)
		}
	}

	// Soft delete the payment
	if err := db.Exec("UPDATE treatment_payments SET deleted_at = NOW() WHERE id = ?", payment.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar pagamento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pagamento deletado com sucesso"})
}

// ============================================
// RECEIPT PDF
// ============================================

// GenerateReceiptPDF - Gerar recibo em PDF
func GenerateReceiptPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	tenantID := c.GetUint("tenant_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Get payment using raw SQL to avoid GORM model contamination
	var payment models.TreatmentPayment
	if err := db.Raw("SELECT * FROM treatment_payments WHERE id = ? AND deleted_at IS NULL", uint(id)).Scan(&payment).Error; err != nil || payment.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pagamento não encontrado"})
		return
	}

	// Get treatment
	var treatment models.Treatment
	if err := db.Raw("SELECT * FROM treatments WHERE id = ? AND deleted_at IS NULL", payment.TreatmentID).Scan(&treatment).Error; err != nil || treatment.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tratamento não encontrado"})
		return
	}
	payment.Treatment = &treatment

	// Get patient
	var patient models.Patient
	if err := db.Raw("SELECT * FROM patients WHERE id = ? AND deleted_at IS NULL", treatment.PatientID).Scan(&patient).Error; err != nil || patient.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente não encontrado"})
		return
	}
	payment.Treatment.Patient = &patient

	// Get tenant info (from public schema)
	var tenant models.Tenant
	if err := db.Raw("SELECT * FROM public.tenants WHERE id = ? AND deleted_at IS NULL", tenantID).Scan(&tenant).Error; err != nil || tenant.ID == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar informações da clínica"})
		return
	}

	// Get user who received (from public schema)
	var receivedBy models.User
	db.Raw("SELECT * FROM public.users WHERE id = ?", payment.ReceivedByID).Scan(&receivedBy)
	payment.ReceivedBy = &receivedBy

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header - Clinic Info
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(51, 51, 51)
	pdf.Cell(0, 10, tr(tenant.Name))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	if tenant.Address != "" {
		pdf.Cell(0, 5, tr(tenant.Address+", "+tenant.City+" - "+tenant.State))
		pdf.Ln(5)
	}
	if tenant.Phone != "" {
		pdf.Cell(0, 5, tr("Tel: "+tenant.Phone))
		pdf.Ln(5)
	}
	if tenant.Email != "" {
		pdf.Cell(0, 5, tr("Email: "+tenant.Email))
		pdf.Ln(5)
	}
	pdf.Ln(8)

	// Title
	pdf.SetFillColor(41, 128, 185)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(180, 12, tr("RECIBO DE PAGAMENTO"), "0", 0, "C", true, 0, "")
	pdf.Ln(20)

	// Receipt Number and Date
	pdf.SetTextColor(51, 51, 51)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(90, 8, tr("Recibo No: "+payment.ReceiptNumber))
	pdf.Cell(90, 8, tr("Data: "+payment.PaidDate.Format("02/01/2006")))
	pdf.Ln(15)

	// Patient Info
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 8, tr("DADOS DO PACIENTE"), "0", 0, "L", true, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, tr("Nome:"))
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(140, 6, tr(payment.Treatment.Patient.Name))
	pdf.Ln(6)

	if payment.Treatment.Patient.CPF != "" {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(40, 6, tr("CPF:"))
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(140, 6, payment.Treatment.Patient.CPF)
		pdf.Ln(6)
	}
	pdf.Ln(8)

	// Payment Info
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 8, tr("DADOS DO PAGAMENTO"), "0", 0, "L", true, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, tr("Tratamento:"))
	pdf.SetFont("Arial", "", 10)
	desc := payment.Treatment.Description
	if len(desc) > 60 {
		desc = desc[:60] + "..."
	}
	pdf.Cell(140, 6, tr(desc))
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, tr("Parcela:"))
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(140, 6, tr(fmt.Sprintf("%d de %d", payment.InstallmentNumber, payment.Treatment.TotalInstallments)))
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, tr("Forma de Pagamento:"))
	pdf.SetFont("Arial", "B", 10)
	paymentMethodLabel := getTreatmentPaymentMethodLabel(payment.PaymentMethod)
	pdf.Cell(140, 6, tr(paymentMethodLabel))
	pdf.Ln(10)

	// Amount Box
	pdf.SetFillColor(39, 174, 96)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(180, 15, tr(fmt.Sprintf("VALOR RECEBIDO: R$ %.2f", payment.Amount)), "0", 0, "C", true, 0, "")
	pdf.Ln(20)

	// Treatment Summary
	pdf.SetTextColor(51, 51, 51)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 8, tr("RESUMO DO TRATAMENTO"), "0", 0, "L", true, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(60, 6, tr("Valor Total do Tratamento:"))
	pdf.Cell(60, 6, tr(fmt.Sprintf("R$ %.2f", payment.Treatment.TotalValue)))
	pdf.Ln(6)

	pdf.Cell(60, 6, tr("Valor Pago ate o momento:"))
	pdf.Cell(60, 6, tr(fmt.Sprintf("R$ %.2f", payment.Treatment.PaidValue)))
	pdf.Ln(6)

	remaining := payment.Treatment.TotalValue - payment.Treatment.PaidValue
	pdf.Cell(60, 6, tr("Saldo Restante:"))
	pdf.SetFont("Arial", "B", 10)
	if remaining <= 0 {
		pdf.SetTextColor(39, 174, 96)
		pdf.Cell(60, 6, tr("QUITADO"))
	} else {
		pdf.SetTextColor(192, 57, 43)
		pdf.Cell(60, 6, tr(fmt.Sprintf("R$ %.2f", remaining)))
	}
	pdf.Ln(15)

	// Notes
	if payment.Notes != "" {
		pdf.SetTextColor(51, 51, 51)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 6, tr("Observacoes:"))
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(payment.Notes), "", "L", false)
		pdf.Ln(10)
	}

	// Signature
	pdf.Ln(15)
	pdf.SetTextColor(51, 51, 51)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(90, 6, tr("_________________________________"))
	pdf.Cell(90, 6, tr("_________________________________"))
	pdf.Ln(5)
	pdf.Cell(90, 6, tr("Assinatura do Paciente"))
	pdf.Cell(90, 6, tr("Assinatura do Responsavel"))
	pdf.Ln(10)

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(150, 150, 150)
	pdf.Cell(0, 5, tr(fmt.Sprintf("Recebido por: %s", payment.ReceivedBy.Name)))
	pdf.Ln(5)
	pdf.Cell(0, 5, tr(fmt.Sprintf("Documento gerado em: %s", time.Now().Format("02/01/2006 15:04"))))

	// Output PDF
	filename := fmt.Sprintf("recibo_%s.pdf", payment.ReceiptNumber)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar PDF"})
		return
	}
}

// getTreatmentPaymentMethodLabel returns Portuguese label for payment method
func getTreatmentPaymentMethodLabel(method string) string {
	labels := map[string]string{
		"cash":        "Dinheiro",
		"credit_card": "Cartao de Credito",
		"debit_card":  "Cartao de Debito",
		"pix":         "PIX",
		"transfer":    "Transferencia",
		"check":       "Cheque",
	}
	if label, ok := labels[method]; ok {
		return label
	}
	return method
}

// ============================================
// AUTO CREATE TREATMENT FROM BUDGET
// ============================================

// CreateTreatmentFromBudget - Called when budget status changes to approved
func CreateTreatmentFromBudget(db *gorm.DB, budget *models.Budget, totalInstallments int) (*models.Treatment, error) {
	if totalInstallments <= 0 {
		totalInstallments = 1
	}

	installmentValue := budget.TotalValue / float64(totalInstallments)

	treatment := models.Treatment{
		BudgetID:          budget.ID,
		PatientID:         budget.PatientID,
		DentistID:         budget.DentistID,
		Description:       budget.Description,
		TotalValue:        budget.TotalValue,
		PaidValue:         0,
		TotalInstallments: totalInstallments,
		InstallmentValue:  installmentValue,
		Status:            models.TreatmentStatusInProgress,
		StartDate:         time.Now(),
	}

	if err := db.Create(&treatment).Error; err != nil {
		return nil, err
	}

	return &treatment, nil
}

// CreateTreatmentFromBudgetRaw - Uses raw SQL to avoid GORM model contamination issues
func CreateTreatmentFromBudgetRaw(db *gorm.DB, budget *models.Budget, totalInstallments int) (*models.Treatment, error) {
	if totalInstallments <= 0 {
		totalInstallments = 1
	}

	installmentValue := budget.TotalValue / float64(totalInstallments)
	startDate := time.Now()
	status := models.TreatmentStatusInProgress

	// Insert using raw SQL to avoid GORM contamination
	result := db.Exec(`
		INSERT INTO treatments
		(created_at, updated_at, budget_id, patient_id, dentist_id, description,
		 total_value, paid_value, total_installments, installment_value, status, start_date)
		VALUES (NOW(), NOW(), ?, ?, ?, ?, ?, 0, ?, ?, ?, ?)
	`, budget.ID, budget.PatientID, budget.DentistID, budget.Description,
		budget.TotalValue, totalInstallments, installmentValue, status, startDate)

	if result.Error != nil {
		return nil, result.Error
	}

	// Get the created treatment ID
	var treatmentID uint
	db.Raw("SELECT id FROM treatments WHERE budget_id = ? AND deleted_at IS NULL ORDER BY id DESC LIMIT 1", budget.ID).Scan(&treatmentID)

	// Load the treatment with relationships
	var treatment models.Treatment
	db.Raw("SELECT * FROM treatments WHERE id = ?", treatmentID).Scan(&treatment)

	// Load patient
	var patient models.Patient
	db.Raw("SELECT * FROM patients WHERE id = ?", treatment.PatientID).Scan(&patient)
	treatment.Patient = &patient

	return &treatment, nil
}
