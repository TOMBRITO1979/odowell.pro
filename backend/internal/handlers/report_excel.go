package handlers

import (
	"drcrwell/backend/internal/models"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func GenerateRevenueExcel(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint("tenant_id")

	// Get tenant info for header
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

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

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Relatório"
	f.SetSheetName("Sheet1", sheet)

	// Header style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	cellStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Set column widths
	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 20)

	row := 1

	// Clinic info
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), tenant.Name)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), headerStyle)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), tenant.Address+", "+tenant.City+" - "+tenant.State)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Tel: "+tenant.Phone)
	row += 2

	// Title
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Relatório de Receitas")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), headerStyle)
	row++

	// Period
	periodText := "Período: "
	if startDate != "" && endDate != "" {
		periodText += startDate + " a " + endDate
	} else {
		periodText += "Todos os registros"
	}
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), periodText)
	row += 2

	// Total Revenue
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Receita Total")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), titleStyle)
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), totalRevenue)
	f.SetCellStyle(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), titleStyle)
	row += 2

	// By Payment Method
	if len(byMethod) > 0 {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Receitas por Método de Pagamento")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), titleStyle)
		row++

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Método")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "Quantidade")
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "Total")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), tableHeaderStyle)
		row++

		for _, m := range byMethod {
			method := m.PaymentMethod
			if method == "" {
				method = "Não especificado"
			}
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), method)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), m.Count)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), m.Total)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), cellStyle)
			row++
		}
		row++
	}

	// By Month
	if len(byMonth) > 0 {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Receitas por Mês")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), titleStyle)
		row++

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Mês")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "Quantidade")
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "Total")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), tableHeaderStyle)
		row++

		for _, m := range byMonth {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), m.Month)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), m.Count)
			f.SetCellValue(sheet, fmt.Sprintf("C%d", row), m.Total)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("C%d", row), cellStyle)
			row++
		}
	}

	// Footer
	row += 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Gerado em: "+time.Now().Format("02/01/2006 15:04"))

	// Output Excel
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=relatorio_receitas.xlsx")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel"})
		return
	}
}

func GenerateAttendanceExcel(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
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

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Relatório"
	f.SetSheetName("Sheet1", sheet)

	// Styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	cellStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Set column widths
	f.SetColWidth(sheet, "A", "A", 35)
	f.SetColWidth(sheet, "B", "B", 20)

	row := 1

	// Clinic info
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), tenant.Name)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), tenant.Address+", "+tenant.City+" - "+tenant.State)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Tel: "+tenant.Phone)
	row += 2

	// Title
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Relatório de Atendimentos")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++

	// Period
	periodText := "Período: "
	if startDate != "" && endDate != "" {
		periodText += startDate + " a " + endDate
	} else {
		periodText += "Todos os registros"
	}
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), periodText)
	row += 2

	// Statistics
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Indicador")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "Valor")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), tableHeaderStyle)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Total de Agendamentos")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), total)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), cellStyle)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Atendimentos Concluídos")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), completed)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), cellStyle)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Cancelados")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), cancelled)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), cellStyle)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Faltas (No-Show)")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), noShow)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), cellStyle)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Taxa de Comparecimento (%)")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", row), fmt.Sprintf("%.2f%%", attendanceRate))
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), cellStyle)
	row++

	// Footer
	row += 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Gerado em: "+time.Now().Format("02/01/2006 15:04"))

	// Output Excel
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=relatorio_atendimentos.xlsx")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel"})
		return
	}
}

func GenerateProceduresExcel(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
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

	// Create Excel file
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Relatório"
	f.SetSheetName("Sheet1", sheet)

	// Styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	cellStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Set column widths
	f.SetColWidth(sheet, "A", "A", 40)
	f.SetColWidth(sheet, "B", "B", 15)

	row := 1

	// Clinic info
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), tenant.Name)
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), tenant.Address+", "+tenant.City+" - "+tenant.State)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Tel: "+tenant.Phone)
	row += 2

	// Title
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Relatório de Procedimentos")
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++

	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Top 20 Procedimentos Mais Realizados")
	row += 2

	// Table
	if len(procedures) > 0 {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Procedimento")
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), "Quantidade")
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), tableHeaderStyle)
		row++

		for _, p := range procedures {
			f.SetCellValue(sheet, fmt.Sprintf("A%d", row), p.Procedure)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", row), p.Count)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), cellStyle)
			row++
		}
	} else {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Nenhum procedimento concluído encontrado.")
		row++
	}

	// Footer
	row += 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", row), "Gerado em: "+time.Now().Format("02/01/2006 15:04"))

	// Output Excel
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=relatorio_procedimentos.xlsx")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel"})
		return
	}
}
