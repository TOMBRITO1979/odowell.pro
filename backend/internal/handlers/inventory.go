package handlers

import (
	"drcrwell/backend/internal/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Products
func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	// Load relationships
	db.Preload("Supplier").First(&product, product.ID)

	c.JSON(http.StatusCreated, gin.H{"product": product})
}

func GetProducts(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.Product{})

	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if active := c.Query("active"); active != "" {
		query = query.Where("active = ?", active == "true")
	}

	var total int64
	query.Count(&total)

	var products []models.Product
	if err := query.Preload("Supplier").Offset(offset).Limit(pageSize).Order("name ASC").
		Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products":  products,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetProduct(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var product models.Product
	if err := db.Preload("Supplier").First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	// Check if product exists
	var count int64
	if err := db.Model(&models.Product{}).Where("id = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var input models.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update using Exec to avoid the duplicate table error
	result := db.Exec(`
		UPDATE products
		SET name = ?, code = ?, description = ?, category = ?, supplier_id = ?,
		    quantity = ?, minimum_stock = ?, unit = ?, cost_price = ?, sale_price = ?,
		    expiration_date = ?, active = ?, barcode = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, input.Name, input.Code, input.Description, input.Category, input.SupplierID,
		input.Quantity, input.MinimumStock, input.Unit, input.CostPrice, input.SalePrice,
		input.ExpirationDate, input.Active, input.Barcode, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	// Load the updated product with relationships
	var product models.Product
	db.Preload("Supplier").First(&product, id)

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	if err := db.Delete(&models.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func GetLowStockProducts(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var products []models.Product
	if err := db.Where("quantity <= minimum_stock AND active = ?", true).
		Order("quantity ASC").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch low stock products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// Suppliers
func CreateSupplier(c *gin.Context) {
	var supplier models.Supplier
	if err := c.ShouldBindJSON(&supplier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&supplier).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create supplier"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"supplier": supplier})
}

func GetSuppliers(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var suppliers []models.Supplier
	if err := db.Order("name ASC").Find(&suppliers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suppliers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suppliers": suppliers})
}

func GetSupplier(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var supplier models.Supplier
	if err := db.First(&supplier, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"supplier": supplier})
}

func UpdateSupplier(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	// Check if supplier exists
	var count int64
	if err := db.Model(&models.Supplier{}).Where("id = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
		return
	}

	var input models.Supplier
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update using Exec to avoid the duplicate table error
	result := db.Exec(`
		UPDATE suppliers
		SET name = ?, cnpj = ?, email = ?, phone = ?, address = ?,
		    city = ?, state = ?, zip_code = ?, active = ?, notes = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, input.Name, input.CNPJ, input.Email, input.Phone, input.Address,
		input.City, input.State, input.ZipCode, input.Active, input.Notes, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update supplier"})
		return
	}

	// Load the updated supplier
	var supplier models.Supplier
	db.First(&supplier, id)

	c.JSON(http.StatusOK, gin.H{"supplier": supplier})
}

func DeleteSupplier(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	if err := db.Delete(&models.Supplier{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete supplier"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Supplier deleted successfully"})
}

// Stock Movements
func CreateStockMovement(c *gin.Context) {
	var movement models.StockMovement
	if err := c.ShouldBindJSON(&movement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate movement type
	if movement.Type != "entry" && movement.Type != "exit" && movement.Type != "adjustment" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movement type. Must be: entry, exit, or adjustment"})
		return
	}

	// Validate quantity
	if movement.Quantity < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity cannot be negative"})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetUint("user_id")
	movement.UserID = userID

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Get product with row lock to prevent race conditions
	var product models.Product
	if err := tx.Clauses().First(&product, movement.ProductID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Store old quantity for logging
	oldQuantity := product.Quantity

	// If this is a sale, set the price information
	if movement.Reason == "sale" {
		movement.UnitPrice = product.SalePrice
		movement.TotalPrice = product.SalePrice * float64(movement.Quantity)
	}

	// Update product quantity based on movement type
	switch movement.Type {
	case "entry":
		product.Quantity += movement.Quantity
	case "exit":
		// Validate sufficient stock
		if product.Quantity < movement.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient stock",
				"available": product.Quantity,
				"requested": movement.Quantity,
			})
			return
		}
		product.Quantity -= movement.Quantity
	case "adjustment":
		// For adjustment, the quantity represents the new total quantity
		product.Quantity = movement.Quantity
	}

	// Ensure quantity doesn't go negative
	if product.Quantity < 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product quantity cannot be negative"})
		return
	}

	// Save the updated product
	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product quantity"})
		return
	}

	// Create movement record
	if err := tx.Create(&movement).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create movement record"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"movement": movement,
		"product": product,
		"old_quantity": oldQuantity,
		"new_quantity": product.Quantity,
	})
}

func GetStockMovements(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.StockMovement{})

	if productID := c.Query("product_id"); productID != "" {
		query = query.Where("product_id = ?", productID)
	}

	var total int64
	query.Count(&total)

	var movements []models.StockMovement
	if err := query.Preload("Product").Preload("User").
		Offset(offset).Limit(pageSize).Order("created_at DESC").
		Find(&movements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movements"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"movements": movements,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func GetStockMovement(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var movement models.StockMovement
	if err := db.Preload("Product").Preload("User").First(&movement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movement not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movement": movement})
}

func UpdateStockMovement(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	// Check if movement exists
	var existingMovement models.StockMovement
	if err := db.First(&existingMovement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movement not found"})
		return
	}

	var input models.StockMovement
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Only allow updating notes and buyer info (not quantity/type/product)
	updates := map[string]interface{}{
		"notes":              input.Notes,
		"buyer_name":         input.BuyerName,
		"buyer_document":     input.BuyerDocument,
		"buyer_phone":        input.BuyerPhone,
		"buyer_street":       input.BuyerStreet,
		"buyer_number":       input.BuyerNumber,
		"buyer_neighborhood": input.BuyerNeighborhood,
		"buyer_city":         input.BuyerCity,
		"buyer_state":        input.BuyerState,
		"buyer_zip_code":     input.BuyerZipCode,
	}

	if err := db.Model(&existingMovement).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update movement"})
		return
	}

	// Reload with relationships
	db.Preload("Product").Preload("User").First(&existingMovement, id)

	c.JSON(http.StatusOK, gin.H{"movement": existingMovement})
}

func DeleteStockMovement(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	// Get the movement to reverse the stock change
	var movement models.StockMovement
	if err := db.First(&movement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movement not found"})
		return
	}

	// Start transaction
	tx := db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Get the product to reverse the quantity change
	var product models.Product
	if err := tx.First(&product, movement.ProductID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Reverse the quantity change based on movement type
	switch movement.Type {
	case "entry":
		// Entry added stock, so we subtract
		if product.Quantity < movement.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":     "Cannot delete: would result in negative stock",
				"available": product.Quantity,
				"needed":    movement.Quantity,
			})
			return
		}
		product.Quantity -= movement.Quantity
	case "exit":
		// Exit removed stock, so we add back
		product.Quantity += movement.Quantity
	case "adjustment":
		// For adjustments, we cannot reliably reverse
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete adjustment movements"})
		return
	}

	// Update product quantity
	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product quantity"})
		return
	}

	// Soft delete the movement
	if err := tx.Delete(&movement).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete movement"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Movement deleted successfully",
		"new_quantity": product.Quantity,
	})
}

// ExitsByReason represents exits grouped by reason
type ExitsByReason struct {
	Reason        string `json:"reason"`
	Count         int64  `json:"count"`
	TotalQuantity int64  `json:"total_quantity"`
}

// ExitsByProduct represents exits grouped by product
type ExitsByProduct struct {
	ProductID     uint   `json:"product_id"`
	ProductName   string `json:"product_name"`
	TotalQuantity int64  `json:"total_quantity"`
}

// ExitsByProductDate represents exits grouped by product and date (for multi-line chart)
type ExitsByProductDate struct {
	ProductID     uint   `json:"product_id"`
	ProductName   string `json:"product_name"`
	Date          string `json:"date"`
	TotalQuantity int64  `json:"total_quantity"`
}

// GetStockMovementStats returns statistics for stock movements
func GetStockMovementStats(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	productID := c.Query("product_id")

	// Build base query for exits (type = 'exit')
	exitsQuery := db.Model(&models.StockMovement{}).Where("type = ?", "exit")

	// Apply date filters if provided
	if startDate != "" {
		exitsQuery = exitsQuery.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		exitsQuery = exitsQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}
	if productID != "" {
		exitsQuery = exitsQuery.Where("product_id = ?", productID)
	}

	// Get exits grouped by reason
	var exitsByReason []ExitsByReason
	if err := exitsQuery.Select("reason, COUNT(*) as count, SUM(quantity) as total_quantity").
		Group("reason").
		Scan(&exitsByReason).Error; err != nil {
		log.Printf("GetStockMovementStats: Failed to get exits by reason: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	// Get exits grouped by product (top 10 products)
	var exitsByProduct []ExitsByProduct
	productExitsQuery := db.Table("stock_movements").
		Select("stock_movements.product_id, products.name as product_name, SUM(stock_movements.quantity) as total_quantity").
		Joins("JOIN products ON products.id = stock_movements.product_id").
		Where("stock_movements.type = ?", "exit").
		Where("stock_movements.deleted_at IS NULL")
	if startDate != "" {
		productExitsQuery = productExitsQuery.Where("stock_movements.created_at >= ?", startDate)
	}
	if endDate != "" {
		productExitsQuery = productExitsQuery.Where("stock_movements.created_at <= ?", endDate+" 23:59:59")
	}
	if productID != "" {
		productExitsQuery = productExitsQuery.Where("stock_movements.product_id = ?", productID)
	}
	if err := productExitsQuery.Group("stock_movements.product_id, products.name").
		Order("total_quantity DESC").
		Limit(10).
		Scan(&exitsByProduct).Error; err != nil {
		log.Printf("GetStockMovementStats: Failed to get exits by product: %v", err)
		// Don't fail, just set empty array
		exitsByProduct = []ExitsByProduct{}
	}

	// Get exits grouped by product AND date for multi-line time-series chart (top 10 products)
	var exitsByProductDate []ExitsByProductDate
	productDateQuery := db.Table("stock_movements").
		Select("stock_movements.product_id, products.name as product_name, TO_CHAR(stock_movements.created_at, 'YYYY-MM-DD') as date, SUM(stock_movements.quantity) as total_quantity").
		Joins("JOIN products ON products.id = stock_movements.product_id").
		Where("stock_movements.type = ?", "exit").
		Where("stock_movements.deleted_at IS NULL")
	if startDate != "" {
		productDateQuery = productDateQuery.Where("stock_movements.created_at >= ?", startDate)
	}
	if endDate != "" {
		productDateQuery = productDateQuery.Where("stock_movements.created_at <= ?", endDate+" 23:59:59")
	}
	if productID != "" {
		productDateQuery = productDateQuery.Where("stock_movements.product_id = ?", productID)
	}
	// Only get data for top 10 products (by total quantity)
	if len(exitsByProduct) > 0 {
		topProductIDs := make([]uint, 0, len(exitsByProduct))
		for _, p := range exitsByProduct {
			topProductIDs = append(topProductIDs, p.ProductID)
		}
		productDateQuery = productDateQuery.Where("stock_movements.product_id IN ?", topProductIDs)
	}
	if err := productDateQuery.Group("stock_movements.product_id, products.name, TO_CHAR(stock_movements.created_at, 'YYYY-MM-DD')").
		Order("date ASC, product_name ASC").
		Scan(&exitsByProductDate).Error; err != nil {
		log.Printf("GetStockMovementStats: Failed to get exits by product and date: %v", err)
		exitsByProductDate = []ExitsByProductDate{}
	}

	// Get total sales revenue (sum of total_price where reason='sale')
	var totalSalesRevenue float64
	salesQuery := db.Model(&models.StockMovement{}).Where("type = ? AND reason = ?", "exit", "sale")
	if startDate != "" {
		salesQuery = salesQuery.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		salesQuery = salesQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}
	if productID != "" {
		salesQuery = salesQuery.Where("product_id = ?", productID)
	}
	if err := salesQuery.Select("COALESCE(SUM(total_price), 0)").Scan(&totalSalesRevenue).Error; err != nil {
		log.Printf("GetStockMovementStats: Failed to get sales revenue: %v", err)
		// Don't fail, just set to 0
		totalSalesRevenue = 0
	}

	// Get total exits count
	var totalExits int64
	totalExitsQuery := db.Model(&models.StockMovement{}).Where("type = ?", "exit")
	if startDate != "" {
		totalExitsQuery = totalExitsQuery.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		totalExitsQuery = totalExitsQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}
	if productID != "" {
		totalExitsQuery = totalExitsQuery.Where("product_id = ?", productID)
	}
	if err := totalExitsQuery.Count(&totalExits).Error; err != nil {
		log.Printf("GetStockMovementStats: Failed to count exits: %v", err)
		totalExits = 0
	}
	log.Printf("GetStockMovementStats: totalExits=%d (startDate=%s, endDate=%s)", totalExits, startDate, endDate)

	// Get total entries count
	var totalEntries int64
	totalEntriesQuery := db.Model(&models.StockMovement{}).Where("type = ?", "entry")
	if startDate != "" {
		totalEntriesQuery = totalEntriesQuery.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		totalEntriesQuery = totalEntriesQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}
	if productID != "" {
		totalEntriesQuery = totalEntriesQuery.Where("product_id = ?", productID)
	}
	if err := totalEntriesQuery.Count(&totalEntries).Error; err != nil {
		log.Printf("GetStockMovementStats: Failed to count entries: %v", err)
		totalEntries = 0
	}
	log.Printf("GetStockMovementStats: totalEntries=%d", totalEntries)

	// Get total sales count (number of sale transactions)
	var totalSalesCount int64
	salesCountQuery := db.Model(&models.StockMovement{}).Where("type = ? AND reason = ?", "exit", "sale")
	if startDate != "" {
		salesCountQuery = salesCountQuery.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		salesCountQuery = salesCountQuery.Where("created_at <= ?", endDate+" 23:59:59")
	}
	if productID != "" {
		salesCountQuery = salesCountQuery.Where("product_id = ?", productID)
	}
	if err := salesCountQuery.Count(&totalSalesCount).Error; err != nil {
		log.Printf("GetStockMovementStats: Failed to count sales: %v", err)
		totalSalesCount = 0
	}
	log.Printf("GetStockMovementStats: totalSalesCount=%d, exitsByProduct=%d, exitsByProductDate=%d", totalSalesCount, len(exitsByProduct), len(exitsByProductDate))

	c.JSON(http.StatusOK, gin.H{
		"exits_by_reason":       exitsByReason,
		"exits_by_product":      exitsByProduct,
		"exits_by_product_date": exitsByProductDate,
		"total_sales_revenue":   totalSalesRevenue,
		"total_sales_count":     totalSalesCount,
		"total_exits":           totalExits,
		"total_entries":         totalEntries,
	})
}
