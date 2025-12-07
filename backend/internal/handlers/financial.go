package handlers

import (
	"drcrwell/backend/internal/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Budgets
func CreateBudget(c *gin.Context) {
	var budget models.Budget
	if err := c.ShouldBindJSON(&budget); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&budget).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create budget"})
		return
	}

	// Load relationships
	db.Preload("Patient").Preload("Dentist").Preload("Payments").First(&budget, budget.ID)

	c.JSON(http.StatusCreated, gin.H{"budget": budget})
}

func GetBudgets(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.Budget{})

	if patientID := c.Query("patient_id"); patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var budgets []models.Budget
	if err := query.Preload("Patient").Preload("Dentist").Preload("Payments").
		Offset(offset).Limit(pageSize).Order("created_at DESC").
		Find(&budgets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budgets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"budgets":   budgets,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetBudget(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var budget models.Budget
	if err := db.Preload("Patient").Preload("Dentist").Preload("Payments").
		First(&budget, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"budget": budget})
}

func UpdateBudget(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	// Get current budget to check status change
	var currentBudget models.Budget
	if err := db.First(&currentBudget, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	var input struct {
		PatientID         uint       `json:"patient_id"`
		DentistID         uint       `json:"dentist_id"`
		Description       string     `json:"description"`
		TotalValue        float64    `json:"total_value"`
		Items             *string    `json:"items"`
		Status            string     `json:"status"`
		ValidUntil        *time.Time `json:"valid_until"`
		Notes             string     `json:"notes"`
		TotalInstallments int        `json:"total_installments"` // For treatment creation
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update using Exec to avoid the duplicate table error
	result := db.Exec(`
		UPDATE budgets
		SET patient_id = ?, dentist_id = ?, description = ?, total_value = ?,
		    items = ?, status = ?, valid_until = ?, notes = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, input.PatientID, input.DentistID, input.Description, input.TotalValue,
		input.Items, input.Status, input.ValidUntil, input.Notes, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update budget"})
		return
	}

	// Load the updated budget with relationships
	var budget models.Budget
	db.Preload("Patient").Preload("Dentist").Preload("Payments").First(&budget, id)

	// If status changed to approved, auto-create treatment
	var treatment *models.Treatment
	if currentBudget.Status != "approved" && input.Status == "approved" {
		log.Printf("DEBUG: Budget %d status changed to approved, checking for existing treatment...", budget.ID)

		// Check if treatment already exists using raw SQL to avoid GORM model contamination
		var existingTreatmentID uint
		err := db.Raw("SELECT id FROM treatments WHERE budget_id = ? AND deleted_at IS NULL LIMIT 1", budget.ID).Scan(&existingTreatmentID).Error

		if err != nil || existingTreatmentID == 0 {
			log.Printf("DEBUG: No existing treatment found for budget %d, creating new treatment...", budget.ID)

			// Treatment doesn't exist, create it
			totalInstallments := input.TotalInstallments
			if totalInstallments <= 0 {
				totalInstallments = 1
			}

			newTreatment, createErr := CreateTreatmentFromBudgetRaw(db, &budget, totalInstallments)
			if createErr != nil {
				log.Printf("ERROR: Failed to create treatment for budget %d: %v", budget.ID, createErr)
			} else {
				log.Printf("DEBUG: Treatment %d created successfully for budget %d", newTreatment.ID, budget.ID)
				treatment = newTreatment
			}
		} else {
			log.Printf("DEBUG: Treatment %d already exists for budget %d", existingTreatmentID, budget.ID)
		}
	}

	response := gin.H{"budget": budget}
	if treatment != nil {
		response["treatment"] = treatment
		response["message"] = "Orçamento aprovado e tratamento criado automaticamente"
	}

	c.JSON(http.StatusOK, response)
}

func DeleteBudget(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	if err := db.Delete(&models.Budget{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete budget"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Budget deleted successfully"})
}

// Payments
func CreatePayment(c *gin.Context) {
	var payment models.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	// Load relationships
	db.Preload("Patient").Preload("Budget").First(&payment, payment.ID)

	c.JSON(http.StatusCreated, gin.H{"payment": payment})
}

func GetPayments(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.Payment{})

	if paymentType := c.Query("type"); paymentType != "" {
		query = query.Where("type = ?", paymentType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if patientID := c.Query("patient_id"); patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}

	var total int64
	query.Count(&total)

	var payments []models.Payment
	if err := query.Preload("Patient").Preload("Budget").
		Offset(offset).Limit(pageSize).Order("created_at DESC").
		Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments":  payments,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetPayment(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var payment models.Payment
	if err := db.Preload("Patient").Preload("Budget").First(&payment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

func UpdatePayment(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	// Check if payment exists
	var count int64
	if err := db.Model(&models.Payment{}).Where("id = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	var input models.Payment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update using Exec to avoid the duplicate table error
	result := db.Exec(`
		UPDATE payments
		SET budget_id = ?, patient_id = ?, type = ?, category = ?, description = ?,
		    amount = ?, payment_method = ?, is_installment = ?, installment_number = ?,
		    total_installments = ?, status = ?, due_date = ?, paid_date = ?,
		    is_insurance = ?, insurance_name = ?, notes = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, input.BudgetID, input.PatientID, input.Type, input.Category, input.Description,
		input.Amount, input.PaymentMethod, input.IsInstallment, input.InstallmentNumber,
		input.TotalInstallments, input.Status, input.DueDate, input.PaidDate,
		input.IsInsurance, input.InsuranceName, input.Notes, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}

	// Load the updated payment with relationships
	var payment models.Payment
	db.Preload("Patient").Preload("Budget").First(&payment, id)

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

func DeletePayment(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	if err := db.Delete(&models.Payment{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment deleted successfully"})
}

func GetCashFlow(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	log.Printf("GetCashFlow: startDate=%s, endDate=%s", startDate, endDate)

	// Result struct for aggregation
	type SumResult struct {
		Total float64
	}

	// Build date filter conditions
	dateConditions := ""
	dateArgs := []interface{}{}
	if startDate != "" {
		dateConditions += " AND created_at >= ?"
		dateArgs = append(dateArgs, startDate)
	}
	if endDate != "" {
		dateConditions += " AND created_at <= ?"
		dateArgs = append(dateArgs, endDate+" 23:59:59")
	}

	// Calculate income (paid income payments)
	// Use Session({}) to clone session (keeps search_path) with fresh statement builder
	var incomeResult SumResult
	incomeArgs := append([]interface{}{"income", "paid"}, dateArgs...)
	if err := db.Session(&gorm.Session{}).Table("payments").
		Where("type = ? AND status = ? AND deleted_at IS NULL"+dateConditions, incomeArgs...).
		Select("COALESCE(SUM(amount), 0) as total").
		Scan(&incomeResult).Error; err != nil {
		log.Printf("GetCashFlow: income query error: %v", err)
	}
	income := incomeResult.Total
	log.Printf("GetCashFlow: income=%f", income)

	// Calculate expenses (paid expense payments)
	var expensesResult SumResult
	expensesArgs := append([]interface{}{"expense", "paid"}, dateArgs...)
	if err := db.Session(&gorm.Session{}).Table("payments").
		Where("type = ? AND status = ? AND deleted_at IS NULL"+dateConditions, expensesArgs...).
		Select("COALESCE(SUM(amount), 0) as total").
		Scan(&expensesResult).Error; err != nil {
		log.Printf("GetCashFlow: expenses query error: %v", err)
	}
	expenses := expensesResult.Total
	log.Printf("GetCashFlow: expenses=%f", expenses)

	// Calculate pending (pending income payments - receivables)
	var pendingResult SumResult
	pendingArgs := append([]interface{}{"income", "pending"}, dateArgs...)
	if err := db.Session(&gorm.Session{}).Table("payments").
		Where("type = ? AND status = ? AND deleted_at IS NULL"+dateConditions, pendingArgs...).
		Select("COALESCE(SUM(amount), 0) as total").
		Scan(&pendingResult).Error; err != nil {
		log.Printf("GetCashFlow: pending query error: %v", err)
	}
	pending := pendingResult.Total
	log.Printf("GetCashFlow: pending=%f", pending)

	log.Printf("GetCashFlow: returning income=%f, expenses=%f, balance=%f, pending=%f", income, expenses, income-expenses, pending)

	c.JSON(http.StatusOK, gin.H{
		"income":   income,
		"expenses": expenses,
		"balance":  income - expenses,
		"pending":  pending,
	})
}

// CancelBudget cancels an approved budget
func CancelBudget(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var budget models.Budget
	if err := db.First(&budget, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	// Only allow cancellation of approved budgets
	if budget.Status != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only approved budgets can be cancelled"})
		return
	}

	// Update status to cancelled using raw SQL
	result := db.Exec("UPDATE budgets SET status = 'cancelled', updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel budget"})
		return
	}

	// Reload budget with updated status
	db.Preload("Patient").Preload("Dentist").Preload("Payments").First(&budget, id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Budget cancelled successfully",
		"budget":  budget,
	})
}

// RefundPayment refunds a paid payment
func RefundPayment(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var input struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payment models.Payment
	if err := db.First(&payment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Only allow refund of paid payments
	if payment.Status != "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only paid payments can be refunded"})
		return
	}

	// Update payment status
	now := time.Now()
	payment.Status = "refunded"
	payment.RefundedDate = &now
	payment.RefundReason = input.Reason

	if err := db.Save(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refund payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment refunded successfully",
		"payment": payment,
	})
}
