package handlers

import (
	"drcrwell/backend/internal/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// GenerateSaleReceiptPDF generates a PDF receipt for a sale movement
func GenerateSaleReceiptPDF(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	tenantID := c.GetUint("tenant_id")
	movementID := c.Param("id")

	log.Printf("GenerateSaleReceiptPDF: tenant_id=%d, movement_id=%s", tenantID, movementID)

	// Get tenant info - use fresh session to avoid GORM statement accumulation
	// The db from SetSchema has accumulated state, so we need a clean query builder
	var tenant models.Tenant
	if err := db.Session(&gorm.Session{NewDB: true}).Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		log.Printf("GenerateSaleReceiptPDF: Failed to load tenant: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	// Get movement with product - use fresh session with NewDB to get clean query builder
	// but still inherit the search_path from the connection
	var movement models.StockMovement
	if err := db.Session(&gorm.Session{NewDB: true}).Preload("Product").First(&movement, movementID).Error; err != nil {
		log.Printf("GenerateSaleReceiptPDF: Failed to find movement %s: %v", movementID, err)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movement not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Validate it's a sale
	if movement.Reason != "sale" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This movement is not a sale"})
		return
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Brand color
	brandR, brandG, brandB := 22, 163, 74 // #16a34a

	// Header with brand color
	pdf.SetFillColor(brandR, brandG, brandB)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(180, 12, tr(tenant.Name), "", 0, "C", true, 0, "")
	pdf.Ln(14)

	// Clinic info
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 10)
	if tenant.Address != "" {
		pdf.Cell(0, 5, tr(tenant.Address+", "+tenant.City+" - "+tenant.State))
		pdf.Ln(5)
	}
	if tenant.Phone != "" {
		pdf.Cell(0, 5, tr("Tel: "+tenant.Phone))
		pdf.Ln(5)
	}
	pdf.Ln(10)

	// Title
	pdf.SetFillColor(brandR, brandG, brandB)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(180, 10, tr("RECIBO DE VENDA"), "", 0, "C", true, 0, "")
	pdf.Ln(15)

	// Receipt info
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(90, 6, fmt.Sprintf("Recibo No: %d", movement.ID))
	pdf.Cell(90, 6, fmt.Sprintf("Data: %s", movement.CreatedAt.Format("02/01/2006 15:04")))
	pdf.Ln(12)

	// Buyer info section
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 8, tr("Dados do Comprador"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)

	buyerName := movement.BuyerName
	if buyerName == "" {
		buyerName = "Nao informado"
	}
	pdf.CellFormat(50, 6, tr("Nome:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(130, 6, tr(buyerName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	buyerDoc := movement.BuyerDocument
	if buyerDoc == "" {
		buyerDoc = "Nao informado"
	}
	pdf.CellFormat(50, 6, tr("CPF/CNPJ:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(130, 6, tr(buyerDoc), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	buyerPhone := movement.BuyerPhone
	if buyerPhone == "" {
		buyerPhone = "Nao informado"
	}
	pdf.CellFormat(50, 6, tr("Telefone:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(130, 6, tr(buyerPhone), "1", 0, "L", false, 0, "")
	pdf.Ln(12)

	// Product info section
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 8, tr("Produto Vendido"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)

	productName := "N/A"
	productCode := ""
	if movement.Product != nil {
		productName = movement.Product.Name
		productCode = movement.Product.Code
	}

	pdf.CellFormat(50, 6, tr("Produto:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(130, 6, tr(productName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	if productCode != "" {
		pdf.CellFormat(50, 6, tr("Codigo:"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(130, 6, tr(productCode), "1", 0, "L", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.CellFormat(50, 6, tr("Quantidade:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(130, 6, fmt.Sprintf("%d", movement.Quantity), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(50, 6, tr("Preco Unitario:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(130, 6, fmt.Sprintf("R$ %.2f", movement.UnitPrice), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	// Total with highlight
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(brandR, brandG, brandB)
	pdf.SetTextColor(255, 255, 255)
	pdf.CellFormat(50, 8, tr("TOTAL:"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(130, 8, fmt.Sprintf("R$ %.2f", movement.TotalPrice), "1", 0, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(12)

	// Notes
	if movement.Notes != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 6, tr("Observacoes:"))
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(180, 5, tr(movement.Notes), "", "L", false)
		pdf.Ln(5)
	}

	// Signature lines
	pdf.Ln(20)
	pdf.SetFont("Arial", "", 10)

	// Seller signature
	pdf.Line(15, pdf.GetY(), 90, pdf.GetY())
	pdf.Ln(3)
	pdf.Cell(75, 5, tr("Vendedor"))
	pdf.Cell(20, 5, "")

	// Buyer signature
	pdf.Line(105, pdf.GetY()-3, 195, pdf.GetY()-3)
	pdf.Cell(75, 5, tr("Comprador"))
	pdf.Ln(15)

	// Footer
	pdf.SetY(-25)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Recibo gerado em: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(4)
	pdf.Cell(0, 5, tr("Este recibo e valido como comprovante de venda."))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=recibo_venda_%d.pdf", movement.ID))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate receipt"})
		return
	}
}
