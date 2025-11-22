package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetDashboard(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Total patients
	var totalPatients int64
	db.Session(&gorm.Session{NewDB: true}).Table("patients").Where("active = ?", true).Count(&totalPatients)

	// Appointments today
	var appointmentsToday int64
	db.Session(&gorm.Session{NewDB: true}).Table("appointments").
		Where("DATE(start_time) = CURRENT_DATE AND status != ?", "cancelled").
		Count(&appointmentsToday)

	// Appointments this month
	var appointmentsMonth int64
	db.Session(&gorm.Session{NewDB: true}).Table("appointments").
		Where("EXTRACT(MONTH FROM start_time) = EXTRACT(MONTH FROM CURRENT_DATE)").
		Where("EXTRACT(YEAR FROM start_time) = EXTRACT(YEAR FROM CURRENT_DATE)").
		Count(&appointmentsMonth)

	// Revenue this month
	var revenueMonth float64
	db.Session(&gorm.Session{NewDB: true}).Table("payments").
		Where("type = ? AND status = ?", "income", "paid").
		Where("EXTRACT(MONTH FROM paid_date) = EXTRACT(MONTH FROM CURRENT_DATE)").
		Where("EXTRACT(YEAR FROM paid_date) = EXTRACT(YEAR FROM CURRENT_DATE)").
		Select("COALESCE(SUM(amount), 0)").Scan(&revenueMonth)

	// Pending payments
	var pendingPayments float64
	db.Session(&gorm.Session{NewDB: true}).Table("payments").
		Where("status = ?", "pending").
		Select("COALESCE(SUM(amount), 0)").Scan(&pendingPayments)

	// Low stock products
	var lowStockCount int64
	db.Session(&gorm.Session{NewDB: true}).Table("products").
		Where("quantity <= minimum_stock AND active = ?", true).
		Count(&lowStockCount)

	// Pending budgets
	var pendingBudgets int64
	db.Session(&gorm.Session{NewDB: true}).Table("budgets").
		Where("status = ?", "pending").
		Count(&pendingBudgets)

	// Pending tasks
	var pendingTasks int64
	db.Session(&gorm.Session{NewDB: true}).Table("tasks").
		Where("status = ?", "pending").
		Count(&pendingTasks)

	c.JSON(http.StatusOK, gin.H{
		"total_patients":     totalPatients,
		"appointments_today": appointmentsToday,
		"appointments_month": appointmentsMonth,
		"revenue_month":      revenueMonth,
		"pending_payments":   pendingPayments,
		"low_stock_count":    lowStockCount,
		"pending_budgets":    pendingBudgets,
		"pending_tasks":      pendingTasks,
	})
}

func GetRevenueReport(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := db.Table("payments").Where("type = ? AND status = ?", "income", "paid")

	if startDate != "" {
		query = query.Where("DATE(paid_date) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(paid_date) <= ?", endDate)
	}

	var totalRevenue float64
	query.Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue)

	// Revenue by payment method
	type MethodRevenue struct {
		PaymentMethod string  `json:"payment_method"`
		Total         float64 `json:"total"`
		Count         int64   `json:"count"`
	}

	var byMethod []MethodRevenue
	query.Select("payment_method, SUM(amount) as total, COUNT(*) as count").
		Group("payment_method").
		Scan(&byMethod)

	// Revenue by month
	type MonthRevenue struct {
		Month string  `json:"month"`
		Total float64 `json:"total"`
		Count int64   `json:"count"`
	}

	var byMonth []MonthRevenue
	db.Table("payments").
		Where("type = ? AND status = ?", "income", "paid").
		Select("TO_CHAR(paid_date, 'YYYY-MM') as month, SUM(amount) as total, COUNT(*) as count").
		Group("month").
		Order("month DESC").
		Limit(12).
		Scan(&byMonth)

	c.JSON(http.StatusOK, gin.H{
		"total_revenue": totalRevenue,
		"by_method":     byMethod,
		"by_month":      byMonth,
	})
}

func GetProceduresReport(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type ProcedureCount struct {
		Procedure string `json:"procedure"`
		Count     int64  `json:"count"`
	}

	var procedures []ProcedureCount
	db.Table("appointments").
		Where("status = ? AND procedure != ''", "completed").
		Select("procedure, COUNT(*) as count").
		Group("procedure").
		Order("count DESC").
		Limit(10).
		Scan(&procedures)

	c.JSON(http.StatusOK, gin.H{
		"procedures": procedures,
	})
}

func GetAttendanceReport(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := db.Table("appointments")

	if startDate != "" {
		query = query.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(start_time) <= ?", endDate)
	}

	var total int64
	query.Count(&total)

	var completed int64
	query.Where("status = ?", "completed").Count(&completed)

	var cancelled int64
	db.Table("appointments").Where("status = ?", "cancelled").Count(&cancelled)

	var noShow int64
	db.Table("appointments").Where("status = ?", "no_show").Count(&noShow)

	var attendanceRate float64
	if total > 0 {
		attendanceRate = (float64(completed) / float64(total)) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total":           total,
		"completed":       completed,
		"cancelled":       cancelled,
		"no_show":         noShow,
		"attendance_rate": attendanceRate,
	})
}
