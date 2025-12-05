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
		"total_patients":      totalPatients,
		"appointments_today":  appointmentsToday,
		"appointments_month":  appointmentsMonth,
		"total_appointments":  appointmentsMonth, // Alias for frontend compatibility
		"revenue_month":       revenueMonth,
		"total_revenue":       revenueMonth, // Alias for frontend compatibility
		"pending_payments":    pendingPayments,
		"low_stock_count":     lowStockCount,
		"pending_budgets":     pendingBudgets,
		"pending_tasks":       pendingTasks,
	})
}

func GetRevenueReport(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Total revenue
	var totalRevenue float64
	totalQuery := db.Session(&gorm.Session{NewDB: true}).Table("payments").Where("type = ? AND status = ?", "income", "paid")
	if startDate != "" {
		totalQuery = totalQuery.Where("DATE(paid_date) >= ?", startDate)
	}
	if endDate != "" {
		totalQuery = totalQuery.Where("DATE(paid_date) <= ?", endDate)
	}
	totalQuery.Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue)

	// Revenue by payment method
	type MethodRevenue struct {
		PaymentMethod string  `json:"payment_method"`
		Total         float64 `json:"total"`
		Count         int64   `json:"count"`
	}

	var byMethod []MethodRevenue
	methodQuery := db.Session(&gorm.Session{NewDB: true}).Table("payments").Where("type = ? AND status = ?", "income", "paid")
	if startDate != "" {
		methodQuery = methodQuery.Where("DATE(paid_date) >= ?", startDate)
	}
	if endDate != "" {
		methodQuery = methodQuery.Where("DATE(paid_date) <= ?", endDate)
	}
	methodQuery.Select("payment_method, SUM(amount) as total, COUNT(*) as count").
		Group("payment_method").
		Scan(&byMethod)

	// Revenue by month
	type MonthRevenue struct {
		Month string  `json:"month"`
		Total float64 `json:"total"`
		Count int64   `json:"count"`
	}

	var byMonth []MonthRevenue
	monthQuery := db.Session(&gorm.Session{NewDB: true}).Table("payments").
		Where("type = ? AND status = ?", "income", "paid")
	if startDate != "" {
		monthQuery = monthQuery.Where("DATE(paid_date) >= ?", startDate)
	}
	if endDate != "" {
		monthQuery = monthQuery.Where("DATE(paid_date) <= ?", endDate)
	}
	monthQuery.Select("TO_CHAR(paid_date, 'YYYY-MM') as month, SUM(amount) as total, COUNT(*) as count").
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

	// Total appointments
	var total int64
	totalQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments")
	if startDate != "" {
		totalQuery = totalQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		totalQuery = totalQuery.Where("DATE(start_time) <= ?", endDate)
	}
	totalQuery.Count(&total)

	// Completed appointments
	var completed int64
	completedQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments").Where("status = ?", "completed")
	if startDate != "" {
		completedQuery = completedQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		completedQuery = completedQuery.Where("DATE(start_time) <= ?", endDate)
	}
	completedQuery.Count(&completed)

	// Cancelled appointments
	var cancelled int64
	cancelledQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments").Where("status = ?", "cancelled")
	if startDate != "" {
		cancelledQuery = cancelledQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		cancelledQuery = cancelledQuery.Where("DATE(start_time) <= ?", endDate)
	}
	cancelledQuery.Count(&cancelled)

	// No-show appointments
	var noShow int64
	noShowQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments").Where("status = ?", "no_show")
	if startDate != "" {
		noShowQuery = noShowQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		noShowQuery = noShowQuery.Where("DATE(start_time) <= ?", endDate)
	}
	noShowQuery.Count(&noShow)

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

func GetBudgetConversionReport(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Total budgets
	var totalBudgets int64
	totalQuery := db.Session(&gorm.Session{NewDB: true}).Table("budgets")
	if startDate != "" {
		totalQuery = totalQuery.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		totalQuery = totalQuery.Where("DATE(created_at) <= ?", endDate)
	}
	totalQuery.Count(&totalBudgets)

	// Budget counts by status (simplified query)
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var statusCounts []StatusCount
	statusQuery := db.Session(&gorm.Session{NewDB: true}).Table("budgets")
	if startDate != "" {
		statusQuery = statusQuery.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		statusQuery = statusQuery.Where("DATE(created_at) <= ?", endDate)
	}
	statusQuery.Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)

	// Approved budgets
	var approvedBudgets int64
	approvedQuery := db.Session(&gorm.Session{NewDB: true}).Table("budgets").Where("status = ?", "approved")
	if startDate != "" {
		approvedQuery = approvedQuery.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		approvedQuery = approvedQuery.Where("DATE(created_at) <= ?", endDate)
	}
	approvedQuery.Count(&approvedBudgets)

	// Conversion rate
	var conversionRate float64
	if totalBudgets > 0 {
		conversionRate = (float64(approvedBudgets) / float64(totalBudgets)) * 100
	}

	// Total amount approved
	var totalApproved float64
	approvedAmountQuery := db.Session(&gorm.Session{NewDB: true}).Table("budgets").
		Where("status = ?", "approved")
	if startDate != "" {
		approvedAmountQuery = approvedAmountQuery.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		approvedAmountQuery = approvedAmountQuery.Where("DATE(created_at) <= ?", endDate)
	}
	approvedAmountQuery.Select("COALESCE(SUM(total_amount), 0)").Scan(&totalApproved)

	c.JSON(http.StatusOK, gin.H{
		"total_budgets":    totalBudgets,
		"approved_budgets": approvedBudgets,
		"conversion_rate":  conversionRate,
		"by_status":        statusCounts,
		"total_approved":   totalApproved,
	})
}

func GetOverduePaymentsReport(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Overdue payments by patient
	type OverduePatient struct {
		PatientID     uint    `json:"patient_id"`
		PatientName   string  `json:"patient_name"`
		OverdueCount  int64   `json:"overdue_count"`
		TotalOverdue  float64 `json:"total_overdue"`
		OldestDueDate string  `json:"oldest_due_date"`
	}

	var overduePatients []OverduePatient
	db.Session(&gorm.Session{NewDB: true}).
		Table("payments pm").
		Select(`
			p.id as patient_id,
			p.name as patient_name,
			COUNT(*) as overdue_count,
			COALESCE(SUM(pm.amount), 0) as total_overdue,
			TO_CHAR(MIN(pm.due_date), 'DD/MM/YYYY') as oldest_due_date
		`).
		Joins("JOIN budgets b ON pm.budget_id = b.id").
		Joins("JOIN patients p ON b.patient_id = p.id").
		Where("pm.status = ? AND pm.due_date < CURRENT_DATE", "pending").
		Group("p.id, p.name").
		Order("total_overdue DESC").
		Scan(&overduePatients)

	// Summary statistics
	var totalOverdue float64
	var overdueCount int64

	db.Session(&gorm.Session{NewDB: true}).Table("payments").
		Where("status = ? AND due_date < CURRENT_DATE", "pending").
		Count(&overdueCount)

	db.Session(&gorm.Session{NewDB: true}).Table("payments").
		Where("status = ? AND due_date < CURRENT_DATE", "pending").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalOverdue)

	// Overdue by age ranges
	type OverdueByAge struct {
		AgeRange string  `json:"age_range"`
		Count    int64   `json:"count"`
		Total    float64 `json:"total"`
	}

	var overdueByAge []OverdueByAge
	db.Session(&gorm.Session{NewDB: true}).
		Table("payments").
		Select(`
			CASE
				WHEN CURRENT_DATE - due_date <= 30 THEN '0-30 dias'
				WHEN CURRENT_DATE - due_date <= 60 THEN '31-60 dias'
				WHEN CURRENT_DATE - due_date <= 90 THEN '61-90 dias'
				ELSE 'Mais de 90 dias'
			END as age_range,
			COUNT(*) as count,
			COALESCE(SUM(amount), 0) as total
		`).
		Where("status = ? AND due_date < CURRENT_DATE", "pending").
		Group("age_range").
		Scan(&overdueByAge)

	c.JSON(http.StatusOK, gin.H{
		"total_overdue":      totalOverdue,
		"overdue_count":      overdueCount,
		"overdue_patients":   overduePatients,
		"overdue_by_age":     overdueByAge,
	})
}

func GetAdvancedDashboard(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Appointments by day
	type DailyAppointments struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	appointmentsQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments")
	if startDate != "" {
		appointmentsQuery = appointmentsQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		appointmentsQuery = appointmentsQuery.Where("DATE(start_time) <= ?", endDate)
	}

	var dailyAppointments []DailyAppointments
	appointmentsQuery.
		Select("TO_CHAR(start_time, 'DD/MM/YYYY') as date, COUNT(*) as count").
		Group("date").
		Order("date ASC").
		Scan(&dailyAppointments)

	// Appointments by professional (dentist)
	type ProfessionalAppointments struct {
		Professional string `json:"professional"`
		Count        int64  `json:"count"`
	}

	var professionalAppointments []ProfessionalAppointments
	professionalQuery := db.Session(&gorm.Session{NewDB: true}).
		Table("appointments a").
		Select("u.name as professional, COUNT(*) as count").
		Joins("JOIN public.users u ON a.dentist_id = u.id").
		Where("a.status = ?", "completed")

	if startDate != "" {
		professionalQuery = professionalQuery.Where("DATE(a.start_time) >= ?", startDate)
	}
	if endDate != "" {
		professionalQuery = professionalQuery.Where("DATE(a.start_time) <= ?", endDate)
	}

	professionalQuery.
		Group("u.name").
		Order("count DESC").
		Scan(&professionalAppointments)

	// Procedures by dentist (for detailed breakdown)
	type ProceduresByDentist struct {
		Professional string `json:"professional"`
		Procedure    string `json:"procedure"`
		Count        int64  `json:"count"`
	}

	var proceduresByDentist []ProceduresByDentist
	proceduresQuery := db.Session(&gorm.Session{NewDB: true}).
		Table("appointments a").
		Select("u.name as professional, a.procedure, COUNT(*) as count").
		Joins("JOIN public.users u ON a.dentist_id = u.id").
		Where("a.status = ? AND a.procedure != ''", "completed")

	if startDate != "" {
		proceduresQuery = proceduresQuery.Where("DATE(a.start_time) >= ?", startDate)
	}
	if endDate != "" {
		proceduresQuery = proceduresQuery.Where("DATE(a.start_time) <= ?", endDate)
	}

	proceduresQuery.
		Group("u.name, a.procedure").
		Order("u.name, count DESC").
		Scan(&proceduresByDentist)

	// Reschedules count - counting appointments with notes containing "remarcad" or similar
	// Since there's no rescheduled field, we'll set this to 0 for now
	var reschedulesCount int64 = 0

	// No-shows count
	var noShowsCount int64
	noShowsQuery := db.Session(&gorm.Session{NewDB: true}).
		Table("appointments").
		Where("status = ?", "no_show")

	if startDate != "" {
		noShowsQuery = noShowsQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		noShowsQuery = noShowsQuery.Where("DATE(start_time) <= ?", endDate)
	}
	noShowsQuery.Count(&noShowsCount)

	// Cancelled appointments
	var cancelledCount int64
	cancelledQuery := db.Session(&gorm.Session{NewDB: true}).
		Table("appointments").
		Where("status = ?", "cancelled")

	if startDate != "" {
		cancelledQuery = cancelledQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		cancelledQuery = cancelledQuery.Where("DATE(start_time) <= ?", endDate)
	}
	cancelledQuery.Count(&cancelledCount)

	// Completed appointments
	var completedCount int64
	completedQuery := db.Session(&gorm.Session{NewDB: true}).
		Table("appointments").
		Where("status = ?", "completed")

	if startDate != "" {
		completedQuery = completedQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		completedQuery = completedQuery.Where("DATE(start_time) <= ?", endDate)
	}
	completedQuery.Count(&completedCount)

	// Total appointments
	var totalAppointments int64
	totalQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments")
	if startDate != "" {
		totalQuery = totalQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		totalQuery = totalQuery.Where("DATE(start_time) <= ?", endDate)
	}
	totalQuery.Count(&totalAppointments)

	// Budgets by day
	type DailyBudgets struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var dailyBudgets []DailyBudgets
	budgetsQuery := db.Session(&gorm.Session{NewDB: true}).Table("budgets")
	if startDate != "" {
		budgetsQuery = budgetsQuery.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		budgetsQuery = budgetsQuery.Where("DATE(created_at) <= ?", endDate)
	}

	budgetsQuery.
		Select("TO_CHAR(created_at, 'DD/MM/YYYY') as date, COUNT(*) as count").
		Group("date").
		Order("date ASC").
		Scan(&dailyBudgets)

	// Budgets by status
	type BudgetStatus struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	var budgetStatus []BudgetStatus
	statusQuery := db.Session(&gorm.Session{NewDB: true}).Table("budgets")
	if startDate != "" {
		statusQuery = statusQuery.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		statusQuery = statusQuery.Where("DATE(created_at) <= ?", endDate)
	}

	statusQuery.
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&budgetStatus)

	// Attendance rate
	var attendanceRate float64
	if totalAppointments > 0 {
		attendanceRate = (float64(completedCount) / float64(totalAppointments)) * 100
	}

	// Total patients
	var totalPatients int64
	db.Session(&gorm.Session{NewDB: true}).Table("patients").Where("active = ?", true).Count(&totalPatients)

	// New patients in period
	var newPatients int64
	newPatientsQuery := db.Session(&gorm.Session{NewDB: true}).Table("patients")
	if startDate != "" {
		newPatientsQuery = newPatientsQuery.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		newPatientsQuery = newPatientsQuery.Where("DATE(created_at) <= ?", endDate)
	}
	newPatientsQuery.Count(&newPatients)

	// Revenue by dentist (from approved budgets)
	type RevenueByDentist struct {
		Professional   string  `json:"professional"`
		TotalBudgets   int64   `json:"total_budgets"`
		TotalRevenue   float64 `json:"total_revenue"`
		PaidRevenue    float64 `json:"paid_revenue"`
		PendingRevenue float64 `json:"pending_revenue"`
	}

	var revenueByDentist []RevenueByDentist
	revenueQuery := db.Session(&gorm.Session{NewDB: true}).
		Table("budgets b").
		Select(`
			u.name as professional,
			COUNT(DISTINCT b.id) as total_budgets,
			COALESCE(SUM(b.total_value), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN p.status = 'paid' THEN p.amount ELSE 0 END), 0) as paid_revenue,
			COALESCE(SUM(b.total_value), 0) - COALESCE(SUM(CASE WHEN p.status = 'paid' THEN p.amount ELSE 0 END), 0) as pending_revenue
		`).
		Joins("JOIN public.users u ON b.dentist_id = u.id").
		Joins("LEFT JOIN payments p ON p.budget_id = b.id").
		Where("b.status = ?", "approved")

	if startDate != "" {
		revenueQuery = revenueQuery.Where("DATE(b.created_at) >= ?", startDate)
	}
	if endDate != "" {
		revenueQuery = revenueQuery.Where("DATE(b.created_at) <= ?", endDate)
	}

	revenueQuery.
		Group("u.name").
		Order("total_revenue DESC").
		Scan(&revenueByDentist)

	c.JSON(http.StatusOK, gin.H{
		"daily_appointments":        dailyAppointments,
		"professional_appointments": professionalAppointments,
		"procedures_by_dentist":     proceduresByDentist,
		"revenue_by_dentist":        revenueByDentist,
		"total_appointments":        totalAppointments,
		"completed_appointments":    completedCount,
		"cancelled_appointments":    cancelledCount,
		"no_shows":                  noShowsCount,
		"reschedules":               reschedulesCount,
		"attendance_rate":           attendanceRate,
		"daily_budgets":             dailyBudgets,
		"budget_status":             budgetStatus,
		"total_patients":            totalPatients,
		"new_patients":              newPatients,
	})
}
