package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CreateTenantRequest struct {
	Name      string `json:"name" binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`

	// Admin user info
	AdminName     string `json:"admin_name" binding:"required"`
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=6"`
}

// CreateTenant creates a new tenant with its own schema and admin user
func CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Normalize subdomain
	subdomain := strings.ToLower(strings.TrimSpace(req.Subdomain))

	// Check if subdomain already exists
	var existing models.Tenant
	if err := db.Where("subdomain = ?", subdomain).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Subdomain already exists"})
		return
	}

	// Check if admin email already exists
	var existingUser models.User
	if err := db.Where("email = ?", req.AdminEmail).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Admin email already registered"})
		return
	}

	// Start transaction
	tx := db.Begin()

	// Create tenant
	tenant := models.Tenant{
		Name:      req.Name,
		Subdomain: subdomain,
		DBSchema:  fmt.Sprintf("tenant_%s", subdomain),
		Email:     req.Email,
		Phone:     req.Phone,
		Address:   req.Address,
		City:      req.City,
		State:     req.State,
		ZipCode:   req.ZipCode,
		Active:    true,
		PlanType:  "basic",
	}

	if err := tx.Create(&tenant).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	// Update schema name with tenant ID
	tenant.DBSchema = fmt.Sprintf("tenant_%d", tenant.ID)
	if err := tx.Save(&tenant).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant schema"})
		return
	}

	// Create tenant schema in database
	if err := database.CreateSchema(tenant.DBSchema); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant schema"})
		return
	}

	// Create admin user
	adminUser := models.User{
		TenantID: tenant.ID,
		Name:     req.AdminName,
		Email:    req.AdminEmail,
		Role:     "admin",
		Active:   true,
	}

	if err := adminUser.HashPassword(req.AdminPassword); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := tx.Create(&adminUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user"})
		return
	}

	// Auto-migrate tables in tenant schema
	tenantDB := database.SetSchema(tx, tenant.DBSchema)
	if err := autoMigrateTenantTables(tenantDB); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to migrate tenant tables"})
		return
	}

	// Commit transaction
	tx.Commit()

	// Generate token for immediate login
	token, _ := generateToken(adminUser.ID, adminUser.TenantID, adminUser.Email, adminUser.Role)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tenant created successfully",
		"tenant": gin.H{
			"id":        tenant.ID,
			"name":      tenant.Name,
			"subdomain": tenant.Subdomain,
			"schema":    tenant.DBSchema,
		},
		"admin": gin.H{
			"id":    adminUser.ID,
			"name":  adminUser.Name,
			"email": adminUser.Email,
		},
		"token": token,
	})
}

// GetTenants lists all tenants (admin only)
func GetTenants(c *gin.Context) {
	var tenants []models.Tenant
	db := database.GetDB()

	if err := db.Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tenants": tenants})
}

// autoMigrateTenantTables creates all tables in tenant schema
func autoMigrateTenantTables(db interface{}) error {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		return fmt.Errorf("invalid database type")
	}

	return gormDB.AutoMigrate(
		&models.Patient{},
		&models.Appointment{},
		&models.MedicalRecord{},
		&models.Budget{},
		&models.Payment{},
		&models.Commission{},
		&models.Product{},
		&models.Supplier{},
		&models.StockMovement{},
		&models.Campaign{},
		&models.CampaignRecipient{},
		&models.Attachment{},
	)
}
