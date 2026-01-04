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

func ExportStockMovementsCSV(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	sqlQuery := "SELECT * FROM stock_movements WHERE deleted_at IS NULL ORDER BY created_at DESC"
	var movements []models.StockMovement
	if err := db.Raw(sqlQuery).Scan(&movements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock movements"})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=movimentacoes_%s.csv", time.Now().Format("20060102_150405")))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	writer.Write([]string{"ID", "Produto ID", "Tipo", "Quantidade", "Motivo", "Usuário ID", "Data", "Observações"})

	for _, movement := range movements {
		writer.Write([]string{
			fmt.Sprintf("%d", movement.ID),
			fmt.Sprintf("%d", movement.ProductID),
			movement.Type,
			fmt.Sprintf("%d", movement.Quantity),
			movement.Reason,
			fmt.Sprintf("%d", movement.UserID),
			movement.CreatedAt.Format("2006-01-02 15:04"),
			movement.Notes,
		})
	}
}

func GenerateStockMovementsListPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	tenantID := c.GetUint("tenant_id")
	var tenant models.Tenant
	db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant)

	sqlQuery := "SELECT * FROM stock_movements WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT 50"
	var movements []models.StockMovement
	db.Raw(sqlQuery).Scan(&movements)

	// Collect unique product and user IDs to avoid N+1 queries
	productIDs := make([]uint, 0, len(movements))
	userIDs := make([]uint, 0, len(movements))
	productIDMap := make(map[uint]bool)
	userIDMap := make(map[uint]bool)

	for _, mov := range movements {
		if !productIDMap[mov.ProductID] {
			productIDs = append(productIDs, mov.ProductID)
			productIDMap[mov.ProductID] = true
		}
		if !userIDMap[mov.UserID] {
			userIDs = append(userIDs, mov.UserID)
			userIDMap[mov.UserID] = true
		}
	}

	// Fetch all products in a single query
	productNames := make(map[uint]string)
	if len(productIDs) > 0 {
		var products []struct {
			ID   uint
			Name string
		}
		db.Raw("SELECT id, name FROM products WHERE id IN (?)", productIDs).Scan(&products)
		for _, p := range products {
			productNames[p.ID] = p.Name
		}
	}

	// Fetch all users in a single query
	userNames := make(map[uint]string)
	if len(userIDs) > 0 {
		var users []struct {
			ID   uint
			Name string
		}
		db.Raw("SELECT id, name FROM public.users WHERE id IN (?)", userIDs).Scan(&users)
		for _, u := range users {
			userNames[u.ID] = u.Name
		}
	}

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
	pdf.Cell(0, 8, tr("Lista de Movimentações de Estoque"))
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
	pdf.CellFormat(50, 7, tr("Produto"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 7, "Tipo", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 7, "Qtd", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 7, "Motivo", "1", 0, "L", true, 0, "")
	pdf.CellFormat(45, 7, tr("Usuário"), "1", 0, "L", true, 0, "")
	pdf.CellFormat(35, 7, "Data", "1", 0, "C", true, 0, "")
	pdf.Ln(-1)

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 7)
	for _, movement := range movements {
		// Use pre-fetched names from maps (avoids N+1 queries)
		productName := productNames[movement.ProductID]
		userName := userNames[movement.UserID]

		if len(productName) > 25 {
			productName = productName[:22] + "..."
		}
		if len(userName) > 22 {
			userName = userName[:19] + "..."
		}

		typeLabel := movement.Type
		if typeLabel == "entry" {
			typeLabel = "Entrada"
		} else if typeLabel == "exit" {
			typeLabel = "Saída"
		} else if typeLabel == "adjustment" {
			typeLabel = "Ajuste"
		}

		pdf.CellFormat(15, 6, fmt.Sprintf("%d", movement.ID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 6, tr(productName), "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 6, typeLabel, "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 6, fmt.Sprintf("%d", movement.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 6, movement.Reason, "1", 0, "L", false, 0, "")
		pdf.CellFormat(45, 6, tr(userName), "1", 0, "L", false, 0, "")
		pdf.CellFormat(35, 6, movement.CreatedAt.Format("02/01 15:04"), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=movimentacoes_%s.pdf", time.Now().Format("20060102_150405")))
	pdf.Output(c.Writer)
}
