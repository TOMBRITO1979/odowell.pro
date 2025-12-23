package handlers

import (
	"drcrwell/backend/internal/cache"
	"drcrwell/backend/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardBasicResponse represents basic dashboard data for caching
type DashboardBasicResponse struct {
	TotalPatients      int64   `json:"total_patients"`
	AppointmentsToday  int64   `json:"appointments_today"`
	AppointmentsMonth  int64   `json:"appointments_month"`
	TotalAppointments  int64   `json:"total_appointments"`
	RevenueMonth       float64 `json:"revenue_month"`
	TotalRevenue       float64 `json:"total_revenue"`
	PendingPayments    float64 `json:"pending_payments"`
	LowStockCount      int64   `json:"low_stock_count"`
	PendingBudgets     int64   `json:"pending_budgets"`
	PendingTasks       int64   `json:"pending_tasks"`
}

// AdvancedDashboardResponse represents advanced dashboard data for caching
type AdvancedDashboardResponse struct {
	DailyAppointments        []DailyAppointments        `json:"daily_appointments"`
	ProfessionalAppointments []ProfessionalAppointments `json:"professional_appointments"`
	ProceduresByDentist      []ProceduresByDentist      `json:"procedures_by_dentist"`
	RevenueByDentist         []RevenueByDentist         `json:"revenue_by_dentist"`
	TotalAppointments        int64                      `json:"total_appointments"`
	CompletedAppointments    int64                      `json:"completed_appointments"`
	CancelledAppointments    int64                      `json:"cancelled_appointments"`
	NoShows                  int64                      `json:"no_shows"`
	Reschedules              int64                      `json:"reschedules"`
	AttendanceRate           float64                    `json:"attendance_rate"`
	DailyBudgets             []DailyBudgets             `json:"daily_budgets"`
	BudgetStatus             []BudgetStatus             `json:"budget_status"`
	TotalPatients            int64                      `json:"total_patients"`
	NewPatients              int64                      `json:"new_patients"`
}

// Types for advanced dashboard
type DailyAppointments struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type ProfessionalAppointments struct {
	Professional string `json:"professional"`
	Count        int64  `json:"count"`
}

type ProceduresByDentist struct {
	Professional string `json:"professional"`
	Procedure    string `json:"procedure"`
	Count        int64  `json:"count"`
}

type RevenueByDentist struct {
	Professional   string  `json:"professional"`
	TotalBudgets   int64   `json:"total_budgets"`
	TotalRevenue   float64 `json:"total_revenue"`
	PaidRevenue    float64 `json:"paid_revenue"`
	PendingRevenue float64 `json:"pending_revenue"`
}

type DailyBudgets struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type BudgetStatus struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

func GetDashboard(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Get tenant ID for cache key
	tenantID := c.GetUint("tenant_id")
	cacheKey := cache.DashboardBasicKey(tenantID)

	// Use cached data with 5-minute TTL
	result, err := cache.GetOrSetTyped[DashboardBasicResponse](cacheKey, cache.TTLDashboard, func() (DashboardBasicResponse, error) {
		var data DashboardBasicResponse

		// Total patients
		db.Session(&gorm.Session{NewDB: true}).Table("patients").Where("active = ?", true).Count(&data.TotalPatients)

		// Appointments today
		db.Session(&gorm.Session{NewDB: true}).Table("appointments").
			Where("DATE(start_time) = CURRENT_DATE AND status != ?", "cancelled").
			Count(&data.AppointmentsToday)

		// Appointments this month
		db.Session(&gorm.Session{NewDB: true}).Table("appointments").
			Where("EXTRACT(MONTH FROM start_time) = EXTRACT(MONTH FROM CURRENT_DATE)").
			Where("EXTRACT(YEAR FROM start_time) = EXTRACT(YEAR FROM CURRENT_DATE)").
			Count(&data.AppointmentsMonth)

		// Revenue this month
		db.Session(&gorm.Session{NewDB: true}).Table("payments").
			Where("type = ? AND status = ?", "income", "paid").
			Where("EXTRACT(MONTH FROM paid_date) = EXTRACT(MONTH FROM CURRENT_DATE)").
			Where("EXTRACT(YEAR FROM paid_date) = EXTRACT(YEAR FROM CURRENT_DATE)").
			Select("COALESCE(SUM(amount), 0)").Scan(&data.RevenueMonth)

		// Pending payments
		db.Session(&gorm.Session{NewDB: true}).Table("payments").
			Where("status = ?", "pending").
			Select("COALESCE(SUM(amount), 0)").Scan(&data.PendingPayments)

		// Low stock products
		db.Session(&gorm.Session{NewDB: true}).Table("products").
			Where("quantity <= minimum_stock AND active = ?", true).
			Count(&data.LowStockCount)

		// Pending budgets
		db.Session(&gorm.Session{NewDB: true}).Table("budgets").
			Where("status = ?", "pending").
			Count(&data.PendingBudgets)

		// Pending tasks
		db.Session(&gorm.Session{NewDB: true}).Table("tasks").
			Where("status = ?", "pending").
			Count(&data.PendingTasks)

		// Set aliases
		data.TotalAppointments = data.AppointmentsMonth
		data.TotalRevenue = data.RevenueMonth

		return data, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar dashboard"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func GetRevenueReport(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

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
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	type ProcedureCount struct {
		Procedure string `json:"procedure"`
		Count     int64  `json:"count"`
	}

	// Query for procedures list
	var procedures []ProcedureCount
	proceduresQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments").
		Where("status = ? AND procedure != ''", "completed")

	if startDate != "" {
		proceduresQuery = proceduresQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		proceduresQuery = proceduresQuery.Where("DATE(start_time) <= ?", endDate)
	}

	proceduresQuery.Select("procedure, COUNT(*) as count").
		Group("procedure").
		Order("count DESC").
		Limit(20).
		Scan(&procedures)

	// Total count of all procedures (not grouped)
	var totalProcedures int64
	totalQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments").
		Where("status = ? AND procedure != ''", "completed")

	if startDate != "" {
		totalQuery = totalQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		totalQuery = totalQuery.Where("DATE(start_time) <= ?", endDate)
	}
	totalQuery.Count(&totalProcedures)

	// Count distinct procedures
	var distinctProcedures int64
	distinctQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments").
		Where("status = ? AND procedure != ''", "completed")

	if startDate != "" {
		distinctQuery = distinctQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		distinctQuery = distinctQuery.Where("DATE(start_time) <= ?", endDate)
	}
	distinctQuery.Select("COUNT(DISTINCT procedure)").Scan(&distinctProcedures)

	c.JSON(http.StatusOK, gin.H{
		"procedures":          procedures,
		"total_procedures":    totalProcedures,
		"distinct_procedures": distinctProcedures,
	})
}

func GetAttendanceReport(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

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

	// Scheduled appointments (agendados)
	var scheduled int64
	scheduledQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments").Where("status = ?", "scheduled")
	if startDate != "" {
		scheduledQuery = scheduledQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		scheduledQuery = scheduledQuery.Where("DATE(start_time) <= ?", endDate)
	}
	scheduledQuery.Count(&scheduled)

	// Confirmed appointments (confirmados)
	var confirmed int64
	confirmedQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments").Where("status = ?", "confirmed")
	if startDate != "" {
		confirmedQuery = confirmedQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		confirmedQuery = confirmedQuery.Where("DATE(start_time) <= ?", endDate)
	}
	confirmedQuery.Count(&confirmed)

	// In progress appointments (em atendimento)
	var inProgress int64
	inProgressQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments").Where("status = ?", "in_progress")
	if startDate != "" {
		inProgressQuery = inProgressQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		inProgressQuery = inProgressQuery.Where("DATE(start_time) <= ?", endDate)
	}
	inProgressQuery.Count(&inProgress)

	// Taxa de comparecimento: dos que deveriam comparecer (completed + no_show), quantos vieram?
	var attendanceRate float64
	attendanceBase := completed + noShow
	if attendanceBase > 0 {
		attendanceRate = (float64(completed) / float64(attendanceBase)) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"total":           total,
		"completed":       completed,
		"cancelled":       cancelled,
		"no_show":         noShow,
		"scheduled":       scheduled,
		"confirmed":       confirmed,
		"in_progress":     inProgress,
		"attendance_rate": attendanceRate,
	})
}

func GetBudgetConversionReport(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

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
	approvedAmountQuery.Select("COALESCE(SUM(total_value), 0)").Scan(&totalApproved)

	c.JSON(http.StatusOK, gin.H{
		"total_budgets":    totalBudgets,
		"approved_budgets": approvedBudgets,
		"conversion_rate":  conversionRate,
		"by_status":        statusCounts,
		"total_approved":   totalApproved,
	})
}

func GetOverduePaymentsReport(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

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
				WHEN (CURRENT_DATE - due_date::date) <= 30 THEN '0-30 dias'
				WHEN (CURRENT_DATE - due_date::date) <= 60 THEN '31-60 dias'
				WHEN (CURRENT_DATE - due_date::date) <= 90 THEN '61-90 dias'
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
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Get tenant ID for cache key
	tenantID := c.GetUint("tenant_id")
	cacheKey := cache.DashboardKey(tenantID, startDate, endDate)

	// Use cached data with 5-minute TTL
	result, err := cache.GetOrSetTyped[AdvancedDashboardResponse](cacheKey, cache.TTLDashboard, func() (AdvancedDashboardResponse, error) {
		var data AdvancedDashboardResponse

		// Appointments by day
		appointmentsQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments")
		if startDate != "" {
			appointmentsQuery = appointmentsQuery.Where("DATE(start_time) >= ?", startDate)
		}
		if endDate != "" {
			appointmentsQuery = appointmentsQuery.Where("DATE(start_time) <= ?", endDate)
		}

		appointmentsQuery.
			Select("TO_CHAR(start_time, 'DD/MM/YYYY') as date, COUNT(*) as count").
			Group("date").
			Order("date ASC").
			Scan(&data.DailyAppointments)

		// Appointments by professional (dentist)
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
			Scan(&data.ProfessionalAppointments)

		// Procedures by dentist (for detailed breakdown)
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
			Scan(&data.ProceduresByDentist)

		// Reschedules count - set to 0 for now
		data.Reschedules = 0

		// No-shows count
		noShowsQuery := db.Session(&gorm.Session{NewDB: true}).
			Table("appointments").
			Where("status = ?", "no_show")

		if startDate != "" {
			noShowsQuery = noShowsQuery.Where("DATE(start_time) >= ?", startDate)
		}
		if endDate != "" {
			noShowsQuery = noShowsQuery.Where("DATE(start_time) <= ?", endDate)
		}
		noShowsQuery.Count(&data.NoShows)

		// Cancelled appointments
		cancelledQuery := db.Session(&gorm.Session{NewDB: true}).
			Table("appointments").
			Where("status = ?", "cancelled")

		if startDate != "" {
			cancelledQuery = cancelledQuery.Where("DATE(start_time) >= ?", startDate)
		}
		if endDate != "" {
			cancelledQuery = cancelledQuery.Where("DATE(start_time) <= ?", endDate)
		}
		cancelledQuery.Count(&data.CancelledAppointments)

		// Completed appointments
		completedQuery := db.Session(&gorm.Session{NewDB: true}).
			Table("appointments").
			Where("status = ?", "completed")

		if startDate != "" {
			completedQuery = completedQuery.Where("DATE(start_time) >= ?", startDate)
		}
		if endDate != "" {
			completedQuery = completedQuery.Where("DATE(start_time) <= ?", endDate)
		}
		completedQuery.Count(&data.CompletedAppointments)

		// Total appointments
		totalQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments")
		if startDate != "" {
			totalQuery = totalQuery.Where("DATE(start_time) >= ?", startDate)
		}
		if endDate != "" {
			totalQuery = totalQuery.Where("DATE(start_time) <= ?", endDate)
		}
		totalQuery.Count(&data.TotalAppointments)

		// Budgets by day
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
			Scan(&data.DailyBudgets)

		// Budgets by status
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
			Scan(&data.BudgetStatus)

		// Attendance rate
		if data.TotalAppointments > 0 {
			data.AttendanceRate = (float64(data.CompletedAppointments) / float64(data.TotalAppointments)) * 100
		}

		// Total patients
		db.Session(&gorm.Session{NewDB: true}).Table("patients").Where("active = ?", true).Count(&data.TotalPatients)

		// New patients in period
		newPatientsQuery := db.Session(&gorm.Session{NewDB: true}).Table("patients")
		if startDate != "" {
			newPatientsQuery = newPatientsQuery.Where("DATE(created_at) >= ?", startDate)
		}
		if endDate != "" {
			newPatientsQuery = newPatientsQuery.Where("DATE(created_at) <= ?", endDate)
		}
		newPatientsQuery.Count(&data.NewPatients)

		// Revenue by dentist (from approved budgets)
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
			Scan(&data.RevenueByDentist)

		return data, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar dashboard avançado"})
		return
	}

	c.JSON(http.StatusOK, result)
}
