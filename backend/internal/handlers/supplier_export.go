package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func ExportSuppliersCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	sqlQuery := "SELECT * FROM suppliers WHERE deleted_at IS NULL ORDER BY name ASC"
	var suppliers []models.Supplier
	if err := db.Raw(sqlQuery).Scan(&suppliers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suppliers"})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=fornecedores_%s.csv", time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer.Write([]string{"Nome", "CNPJ", "Email", "Telefone", "Endereço", "Cidade", "Estado", "CEP", "Ativo", "Observações"})

	for _, supplier := range suppliers {
		active := "Não"
		if supplier.Active {
			active = "Sim"
		}
		writer.Write([]string{
			supplier.Name,
			supplier.CNPJ,
			supplier.Email,
			supplier.Phone,
			supplier.Address,
			supplier.City,
			supplier.State,
			supplier.ZipCode,
			active,
			supplier.Notes,
		})
	}
}

func ImportSuppliersCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only CSV files are allowed"})
		return
	}

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse CSV file"})
		return
	}

	imported := 0
	errors := []string{}

	for i, record := range records[1:] {
		lineNum := i + 2
		if len(record) < 2 {
			errors = append(errors, fmt.Sprintf("Linha %d: Dados insuficientes", lineNum))
			continue
		}

		active := true
		if len(record) > 8 && (record[8] == "Não" || record[8] == "false" || record[8] == "0") {
			active = false
		}

		supplier := models.Supplier{
			Name:    record[0],
			CNPJ:    record[1],
			Email:   record[2],
			Phone:   record[3],
			Address: record[4],
			City:    record[5],
			State:   record[6],
			ZipCode: record[7],
			Active:  active,
		}
		if len(record) > 9 {
			supplier.Notes = record[9]
		}

		if err := db.Create(&supplier).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Linha %d: %v", lineNum, err))
			continue
		}
		imported++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("%d fornecedores importados", imported),
		"imported": imported,
		"errors":   errors,
	})
}

func GenerateSuppliersListPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	tenantID := c.GetUint("tenant_id")
	var tenant models.Tenant
	db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant)

	sqlQuery := "SELECT * FROM suppliers WHERE deleted_at IS NULL ORDER BY name ASC LIMIT 50"
	var suppliers []models.Supplier
	db.Raw(sqlQuery).Scan(&suppliers)

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
	pdf.Cell(0, 8, tr("Lista de Fornecedores"))
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
	pdf.CellFormat(60, 7, tr("Nome"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 7, "CNPJ", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 7, "Email", "1", 0, "L", true, 0, "")
	pdf.CellFormat(30, 7, "Telefone", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Cidade", "1", 0, "L", true, 0, "")
	pdf.CellFormat(15, 7, "Ativo", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 7)
	for _, supplier := range suppliers {
		active := "Sim"
		if !supplier.Active {
			active = "Não"
		}

		pdf.CellFormat(60, 6, tr(supplier.Name), "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, supplier.CNPJ, "1", 0, "L", false, 0, "")
		pdf.CellFormat(50, 6, supplier.Email, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, supplier.Phone, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, supplier.City, "1", 0, "L", false, 0, "")
		pdf.CellFormat(15, 6, active, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=fornecedores_%s.pdf", time.Now().Format("20060102_150405")))
	pdf.Output(c.Writer)
}
