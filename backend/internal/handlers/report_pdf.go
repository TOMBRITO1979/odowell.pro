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

func GenerateRevenuePDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	// Get tenant info for header
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Create a fresh query session
	query := db.Session(&gorm.Session{NewDB: true}).Table("payments").Where("type = ? AND status = ?", "income", "paid")

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
	db.Session(&gorm.Session{NewDB: true}).Table("payments").
		Where("type = ? AND status = ?", "income", "paid").
		Select("TO_CHAR(paid_date, 'YYYY-MM') as month, SUM(amount) as total, COUNT(*) as count").
		Group("month").
		Order("month DESC").
		Limit(12).
		Scan(&byMonth)

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
	pdf.Cell(0, 8, tr("Relatorio de Receitas"))
	pdf.Ln(10)

	// Period
	pdf.SetFont("Arial", "", 10)
	periodText := "Periodo: "
	if startDate != "" && endDate != "" {
		periodText += startDate + " a " + endDate
	} else {
		periodText += "Todos os registros"
	}
	pdf.Cell(0, 6, periodText)
	pdf.Ln(8)

	// Total Revenue
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(180, 8, tr("Receita Total"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(180, 8, fmt.Sprintf("R$ %.2f", totalRevenue), "1", 0, "R", false, 0, "")
	pdf.Ln(12)

	// By Payment Method
	if len(byMethod) > 0 {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 7, tr("Receitas por Metodo de Pagamento"))
		pdf.Ln(8)

		pdf.SetFillColor(220, 220, 220)
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(90, 7, tr("Metodo"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(45, 7, tr("Quantidade"), "1", 0, "C", true, 0, "")
		pdf.CellFormat(45, 7, tr("Total"), "1", 0, "R", true, 0, "")
		pdf.Ln(-1)

		pdf.SetFont("Arial", "", 9)
		for _, m := range byMethod {
			method := m.PaymentMethod
			if method == "" {
				method = "Nao especificado"
			}
			pdf.CellFormat(90, 6, tr(method), "1", 0, "L", false, 0, "")
			pdf.CellFormat(45, 6, fmt.Sprintf("%d", m.Count), "1", 0, "C", false, 0, "")
			pdf.CellFormat(45, 6, fmt.Sprintf("R$ %.2f", m.Total), "1", 0, "R", false, 0, "")
			pdf.Ln(-1)
		}
		pdf.Ln(8)
	}

	// By Month
	if len(byMonth) > 0 {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 7, tr("Receitas por Mes (Ultimos 12 meses)"))
		pdf.Ln(8)

		pdf.SetFillColor(220, 220, 220)
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(90, 7, tr("Mes"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(45, 7, tr("Quantidade"), "1", 0, "C", true, 0, "")
		pdf.CellFormat(45, 7, tr("Total"), "1", 0, "R", true, 0, "")
		pdf.Ln(-1)

		pdf.SetFont("Arial", "", 9)
		for _, m := range byMonth {
			pdf.CellFormat(90, 6, m.Month, "1", 0, "L", false, 0, "")
			pdf.CellFormat(45, 6, fmt.Sprintf("%d", m.Count), "1", 0, "C", false, 0, "")
			pdf.CellFormat(45, 6, fmt.Sprintf("R$ %.2f", m.Total), "1", 0, "R", false, 0, "")
			pdf.Ln(-1)
		}
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=relatorio_receitas.pdf")

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

func GenerateAttendancePDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	// Get tenant info for header
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := db.Session(&gorm.Session{NewDB: true}).Table("appointments")

	if startDate != "" {
		query = query.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(start_time) <= ?", endDate)
	}

	var total int64
	query.Count(&total)

	var completed int64
	db.Session(&gorm.Session{NewDB: true}).Table("appointments").Where("status = ?", "completed").Count(&completed)

	var cancelled int64
	db.Session(&gorm.Session{NewDB: true}).Table("appointments").Where("status = ?", "cancelled").Count(&cancelled)

	var noShow int64
	db.Session(&gorm.Session{NewDB: true}).Table("appointments").Where("status = ?", "no_show").Count(&noShow)

	var attendanceRate float64
	if total > 0 {
		attendanceRate = (float64(completed) / float64(total)) * 100
	}

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
	pdf.Cell(0, 8, tr("Relatorio de Atendimentos"))
	pdf.Ln(10)

	// Period
	pdf.SetFont("Arial", "", 10)
	periodText := "Periodo: "
	if startDate != "" && endDate != "" {
		periodText += startDate + " a " + endDate
	} else {
		periodText += "Todos os registros"
	}
	pdf.Cell(0, 6, periodText)
	pdf.Ln(12)

	// Statistics table
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 7, tr("Estatisticas de Atendimento"))
	pdf.Ln(8)

	pdf.SetFillColor(220, 220, 220)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(120, 8, tr("Indicador"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 8, tr("Valor"), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(120, 7, tr("Total de Agendamentos"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%d", total), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(120, 7, tr("Atendimentos Concluidos"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%d", completed), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(120, 7, tr("Cancelados"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%d", cancelled), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(120, 7, tr("Faltas (No-Show)"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%d", noShow), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(120, 7, tr("Taxa de Comparecimento"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%.2f%%", attendanceRate), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=relatorio_atendimentos.pdf")

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

func GenerateProceduresPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	// Get tenant info for header
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	type ProcedureCount struct {
		Procedure string `json:"procedure"`
		Count     int64  `json:"count"`
	}

	var procedures []ProcedureCount
	db.Session(&gorm.Session{NewDB: true}).Table("appointments").
		Where("status = ? AND procedure != ''", "completed").
		Select("procedure, COUNT(*) as count").
		Group("procedure").
		Order("count DESC").
		Limit(20).
		Scan(&procedures)

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
	pdf.Cell(0, 8, tr("Relatorio de Procedimentos"))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, tr("Top 20 Procedimentos Mais Realizados"))
	pdf.Ln(12)

	// Table
	if len(procedures) > 0 {
		pdf.SetFillColor(220, 220, 220)
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(140, 8, tr("Procedimento"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(40, 8, tr("Quantidade"), "1", 0, "C", true, 0, "")
		pdf.Ln(-1)

		pdf.SetFont("Arial", "", 9)
		for _, p := range procedures {
			pdf.CellFormat(140, 7, tr(p.Procedure), "1", 0, "L", false, 0, "")
			pdf.CellFormat(40, 7, fmt.Sprintf("%d", p.Count), "1", 0, "C", false, 0, "")
			pdf.Ln(-1)
		}
	} else {
		pdf.SetFont("Arial", "I", 10)
		pdf.Cell(0, 7, tr("Nenhum procedimento concluido encontrado."))
		pdf.Ln(-1)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=relatorio_procedimentos.pdf")

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

func GenerateBudgetConversionPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	// Get tenant info for header
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := db.Session(&gorm.Session{NewDB: true}).Table("budgets")

	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	var totalBudgets int64
	query.Count(&totalBudgets)

	var approvedBudgets int64
	db.Session(&gorm.Session{NewDB: true}).Table("budgets").Where("status = ?", "approved").Count(&approvedBudgets)

	var conversionRate float64
	if totalBudgets > 0 {
		conversionRate = (float64(approvedBudgets) / float64(totalBudgets)) * 100
	}

	var totalApproved float64
	db.Session(&gorm.Session{NewDB: true}).Table("budgets").Where("status = ?", "approved").
		Select("COALESCE(SUM(total_amount), 0)").Scan(&totalApproved)

	// Create PDF
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
	pdf.Cell(0, 8, tr("Taxa de Conversao de Orcamentos"))
	pdf.Ln(10)

	// Period
	pdf.SetFont("Arial", "", 10)
	periodText := "Periodo: "
	if startDate != "" && endDate != "" {
		periodText += startDate + " a " + endDate
	} else {
		periodText += "Todos os registros"
	}
	pdf.Cell(0, 6, periodText)
	pdf.Ln(12)

	// Statistics table
	pdf.SetFillColor(220, 220, 220)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(120, 8, tr("Indicador"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 8, tr("Valor"), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(120, 7, tr("Total de Orcamentos"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%d", totalBudgets), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(120, 7, tr("Orcamentos Aprovados"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%d", approvedBudgets), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	pdf.SetFillColor(144, 238, 144)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(120, 7, tr("Taxa de Conversao"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%.2f%%", conversionRate), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFillColor(220, 220, 220)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(120, 7, tr("Valor Total Aprovado"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("R$ %.2f", totalApproved), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=conversao_orcamentos.pdf")

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

func GenerateOverduePaymentsPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	// Get tenant info for header
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	var totalOverdue float64
	db.Session(&gorm.Session{NewDB: true}).Table("payments").
		Where("status = ? AND type = ? AND due_date < CURRENT_DATE", "pending", "expense").
		Select("COALESCE(SUM(amount), 0)").Scan(&totalOverdue)

	var overdueCount int64
	db.Session(&gorm.Session{NewDB: true}).Table("payments").
		Where("status = ? AND type = ? AND due_date < CURRENT_DATE", "pending", "expense").
		Count(&overdueCount)

	// Create PDF
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
	pdf.Cell(0, 8, tr("Controle de Inadimplencia"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006")))
	pdf.Ln(12)

	// Statistics table
	pdf.SetFillColor(255, 200, 200)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(120, 8, tr("Total em Atraso"), "1", 0, "L", true, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(60, 8, fmt.Sprintf("R$ %.2f", totalOverdue), "1", 0, "R", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFillColor(220, 220, 220)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(120, 7, tr("Quantidade de Pagamentos Atrasados"), "1", 0, "L", true, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(60, 7, fmt.Sprintf("%d", overdueCount), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	if overdueCount == 0 {
		pdf.Ln(10)
		pdf.SetFont("Arial", "BI", 11)
		pdf.SetTextColor(0, 150, 0)
		pdf.Cell(0, 7, tr("Parabens! Nenhum pagamento em atraso."))
		pdf.SetTextColor(0, 0, 0)
	}

	// Footer
	pdf.Ln(15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Relatorio gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=inadimplencia.pdf")

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}

func GenerateDashboardPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	// Get tenant info for header
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// KPIs - Appointments
	appointmentsQuery := db.Session(&gorm.Session{NewDB: true}).Table("appointments")
	if startDate != "" {
		appointmentsQuery = appointmentsQuery.Where("DATE(start_time) >= ?", startDate)
	}
	if endDate != "" {
		appointmentsQuery = appointmentsQuery.Where("DATE(start_time) <= ?", endDate)
	}

	var totalAppointments, completedCount, cancelledCount, noShowsCount, reschedulesCount int64
	appointmentsQuery.Count(&totalAppointments)
	appointmentsQuery.Where("status = ?", "completed").Count(&completedCount)
	appointmentsQuery.Where("status = ?", "cancelled").Count(&cancelledCount)
	appointmentsQuery.Where("status = ?", "no_show").Count(&noShowsCount)

	// Count reschedules
	db.Session(&gorm.Session{NewDB: true}).Table("appointments").
		Where("rescheduled_from IS NOT NULL").Count(&reschedulesCount)

	// Attendance Rate
	var attendanceRate float64
	if totalAppointments > 0 {
		attendanceRate = (float64(completedCount) / float64(totalAppointments)) * 100
	}

	// Patients
	var totalPatients, newPatients int64
	db.Session(&gorm.Session{NewDB: true}).Table("patients").Where("active = ?", true).Count(&totalPatients)

	newPatientsQuery := db.Session(&gorm.Session{NewDB: true}).Table("patients")
	if startDate != "" {
		newPatientsQuery = newPatientsQuery.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		newPatientsQuery = newPatientsQuery.Where("DATE(created_at) <= ?", endDate)
	}
	newPatientsQuery.Count(&newPatients)

	// Budget Status
	type BudgetStatus struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var budgetStatus []BudgetStatus
	budgetQuery := db.Session(&gorm.Session{NewDB: true}).Table("budgets")
	if startDate != "" {
		budgetQuery = budgetQuery.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		budgetQuery = budgetQuery.Where("DATE(created_at) <= ?", endDate)
	}
	budgetQuery.Select("status, COUNT(*) as count").Group("status").Scan(&budgetStatus)

	// Create PDF
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
	pdf.Cell(0, 8, tr("Dashboard Executivo"))
	pdf.Ln(10)

	// Period
	pdf.SetFont("Arial", "", 10)
	periodText := "Periodo: "
	if startDate != "" && endDate != "" {
		periodText += startDate + " a " + endDate
	} else {
		periodText += "Ultimos 30 dias"
	}
	pdf.Cell(0, 6, periodText)
	pdf.Ln(10)

	// KPIs Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 7, tr("Indicadores Principais"))
	pdf.Ln(8)

	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(120, 7, tr("Indicador"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 7, tr("Valor"), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	indicators := []struct {
		label string
		value string
	}{
		{"Total de Consultas", fmt.Sprintf("%d", totalAppointments)},
		{"Consultas Concluidas", fmt.Sprintf("%d", completedCount)},
		{"Consultas Canceladas", fmt.Sprintf("%d", cancelledCount)},
		{"Faltas (No-Show)", fmt.Sprintf("%d", noShowsCount)},
		{"Remarcacoes", fmt.Sprintf("%d", reschedulesCount)},
		{"Taxa de Comparecimento", fmt.Sprintf("%.2f%%", attendanceRate)},
	}

	for _, ind := range indicators {
		pdf.CellFormat(120, 6, tr(ind.label), "1", 0, "L", false, 0, "")
		pdf.CellFormat(60, 6, ind.value, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}
	pdf.Ln(8)

	// Patients Section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 7, tr("Pacientes"))
	pdf.Ln(8)

	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(120, 7, tr("Metrica"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 7, tr("Valor"), "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(120, 6, tr("Total de Pacientes Ativos"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 6, fmt.Sprintf("%d", totalPatients), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)
	pdf.CellFormat(120, 6, tr("Novos Pacientes no Periodo"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 6, fmt.Sprintf("%d", newPatients), "1", 0, "C", false, 0, "")
	pdf.Ln(-1)
	pdf.Ln(8)

	// Budget Status Section
	if len(budgetStatus) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 7, tr("Orcamentos por Status"))
		pdf.Ln(8)

		pdf.SetFillColor(240, 240, 240)
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(120, 7, tr("Status"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(60, 7, tr("Quantidade"), "1", 0, "C", true, 0, "")
		pdf.Ln(-1)

		pdf.SetFont("Arial", "", 10)
		statusTranslations := map[string]string{
			"approved":  "Aprovado",
			"cancelled": "Cancelado",
			"pending":   "Pendente",
			"rejected":  "Rejeitado",
		}

		for _, bs := range budgetStatus {
			translatedStatus := statusTranslations[bs.Status]
			if translatedStatus == "" {
				translatedStatus = bs.Status
			}
			pdf.CellFormat(120, 6, tr(translatedStatus), "1", 0, "L", false, 0, "")
			pdf.CellFormat(60, 6, fmt.Sprintf("%d", bs.Count), "1", 0, "C", false, 0, "")
			pdf.Ln(-1)
		}
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=dashboard_executivo.pdf")

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}
