package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

func ExportProductsCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	sqlQuery := "SELECT * FROM products WHERE deleted_at IS NULL ORDER BY name ASC"
	var products []models.Product
	if err := db.Raw(sqlQuery).Scan(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=produtos_%s.csv", time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer.Write([]string{"Nome", "Código", "Descrição", "Categoria", "Quantidade", "Estoque Mínimo", "Unidade", "Preço Custo", "Preço Venda"})

	for _, product := range products {
		writer.Write([]string{
			product.Name,
			product.Code,
			product.Description,
			product.Category,
			fmt.Sprintf("%d", product.Quantity),
			fmt.Sprintf("%d", product.MinimumStock),
			product.Unit,
			fmt.Sprintf("%.2f", product.CostPrice),
			fmt.Sprintf("%.2f", product.SalePrice),
		})
	}
}

func ImportProductsCSV(c *gin.Context) {
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

		quantity, _ := strconv.Atoi(record[4])
		minStock, _ := strconv.Atoi(record[5])
		costPrice, _ := strconv.ParseFloat(record[7], 64)
		salePrice, _ := strconv.ParseFloat(record[8], 64)

		product := models.Product{
			Name:         record[0],
			Code:         record[1],
			Description:  record[2],
			Category:     record[3],
			Quantity:     quantity,
			MinimumStock: minStock,
			Unit:         record[6],
			CostPrice:    costPrice,
			SalePrice:    salePrice,
		}

		if err := db.Create(&product).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Linha %d: %v", lineNum, err))
			continue
		}
		imported++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("%d produtos importados", imported),
		"imported": imported,
		"errors":   errors,
	})
}

func GenerateProductsListPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	tenantID := c.GetUint("tenant_id")
	var tenant models.Tenant
	db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant)

	sqlQuery := "SELECT * FROM products WHERE deleted_at IS NULL ORDER BY name ASC LIMIT 50"
	var products []models.Product
	db.Raw(sqlQuery).Scan(&products)

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
	pdf.Cell(0, 8, tr("Lista de Produtos"))
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
	pdf.CellFormat(30, 7, tr("Codigo"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Categoria", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 7, "Qtd", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Unidade", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, tr("Preco Venda"), "1", 0, "R", true, 0, "")
	pdf.Ln(-1)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 7)
	for _, product := range products {
		pdf.CellFormat(60, 6, tr(product.Name), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, product.Code, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, product.Category, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%d", product.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, product.Unit, "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, fmt.Sprintf("R$ %.2f", product.SalePrice), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=produtos_%s.pdf", time.Now().Format("20060102_150405")))
	pdf.Output(c.Writer)
}
