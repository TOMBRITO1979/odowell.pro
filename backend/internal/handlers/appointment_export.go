package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func ExportAppointmentsCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	sqlQuery := "SELECT * FROM appointments WHERE deleted_at IS NULL ORDER BY start_time DESC"
	var appointments []models.Appointment
	if err := db.Raw(sqlQuery).Scan(&appointments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=agendamentos_%s.csv", time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer.Write([]string{"ID", "Paciente ID", "Profissional ID", "Data/Hora Início", "Data/Hora Fim", "Tipo", "Procedimento", "Status", "Observações"})

	for _, apt := range appointments {
		writer.Write([]string{
			fmt.Sprintf("%d", apt.ID),
			fmt.Sprintf("%d", apt.PatientID),
			fmt.Sprintf("%d", apt.DentistID),
			apt.StartTime.Format("2006-01-02 15:04"),
			apt.EndTime.Format("2006-01-02 15:04"),
			apt.Type,
			apt.Procedure,
			apt.Status,
			apt.Notes,
		})
	}
}

func GenerateAppointmentsListPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	tenantID := c.GetUint("tenant_id")
	var tenant models.Tenant
	db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant)

	sqlQuery := "SELECT * FROM appointments WHERE deleted_at IS NULL ORDER BY start_time DESC LIMIT 50"
	var appointments []models.Appointment
	db.Raw(sqlQuery).Scan(&appointments)

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 10)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Modern header with brand color (#16a34a)
	pdf.SetFillColor(22, 163, 74) // #16a34a
	pdf.Rect(0, 0, 297, 25, "F")

	pdf.SetTextColor(255, 255, 255) // White text
	pdf.SetFont("Arial", "B", 18)
	pdf.SetY(8)
	pdf.Cell(0, 8, tr(tenant.Name))
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, tr(tenant.Address+", "+tenant.City+" - "+tenant.State))
	pdf.Ln(10)

	// Reset text color to black
	pdf.SetTextColor(0, 0, 0)

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, tr("Lista de Agendamentos"))
	pdf.Ln(6)

	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 5, fmt.Sprintf("Gerado em: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(8)

	// Table header with brand color
	pdf.SetFillColor(22, 163, 74) // #16a34a
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(15, 7, "ID", "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 7, tr("Paciente"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(40, 7, tr("Profissional"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 7, tr("Data/Hora"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(50, 7, tr("Procedimento"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 7, "Status", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 7)
	for _, apt := range appointments {
		var patientName, dentistName string
		db.Raw("SELECT name FROM patients WHERE id = ?", apt.PatientID).Scan(&patientName)
		db.Raw("SELECT name FROM public.users WHERE id = ?", apt.DentistID).Scan(&dentistName)

		if len(patientName) > 25 {
			patientName = patientName[:22] + "..."
		}
		if len(dentistName) > 20 {
			dentistName = dentistName[:17] + "..."
		}

		pdf.CellFormat(15, 6, fmt.Sprintf("%d", apt.ID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 6, tr(patientName), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, tr(dentistName), "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, apt.StartTime.Format("02/01 15:04"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 6, tr(apt.Procedure), "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, apt.Status, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=agendamentos_%s.pdf", time.Now().Format("20060102_150405")))
	pdf.Output(c.Writer)
}
