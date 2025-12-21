package handlers

import (
	"crypto/rand"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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

	// Validate admin password strength
	if valid, msg := ValidatePassword(req.AdminPassword); !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
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

	// Generate unique API key for tenant
	apiKeyBytes := make([]byte, 32)
	if _, err := rand.Read(apiKeyBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API key"})
		return
	}
	apiKey := hex.EncodeToString(apiKeyBytes)

	// Create tenant with 7-day trial
	trialEndsAt := time.Now().AddDate(0, 0, 7) // 7 days trial
	tenant := models.Tenant{
		Name:               req.Name,
		Subdomain:          subdomain,
		DBSchema:           fmt.Sprintf("tenant_%s", subdomain),
		Email:              req.Email,
		Phone:              req.Phone,
		Address:            req.Address,
		City:               req.City,
		State:              req.State,
		ZipCode:            req.ZipCode,
		Active:             true,
		PlanType:           "basic",
		APIKey:             apiKey,
		SubscriptionStatus: "trialing",
		TrialEndsAt:        &trialEndsAt,
		PatientLimit:       1000, // Default limit for trial
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

	// Send verification email asynchronously
	go func() {
		if err := CreateAndSendVerification(tenant.ID, tenant.Email, tenant.Name); err != nil {
			log.Printf("Failed to send verification email to %s: %v", tenant.Email, err)
		} else {
			log.Printf("Verification email sent to %s", tenant.Email)
		}
	}()

	// NOTE: Token is NOT generated here - user must verify email and login separately
	c.JSON(http.StatusCreated, gin.H{
		"message":            "Consultório criado com sucesso! Verifique seu email para ativar a conta.",
		"verification_sent":  true,
		"verification_email": tenant.Email,
		"tenant": gin.H{
			"id":             tenant.ID,
			"name":           tenant.Name,
			"subdomain":      tenant.Subdomain,
			"email_verified": false,
		},
		"admin": gin.H{
			"id":    adminUser.ID,
			"name":  adminUser.Name,
			"email": adminUser.Email,
		},
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
		&models.Treatment{},
		&models.TreatmentPayment{},
		&models.Product{},
		&models.Supplier{},
		&models.StockMovement{},
		&models.Campaign{},
		&models.CampaignRecipient{},
		&models.Attachment{},
		&models.Exam{},
		&models.Prescription{},
		&models.Task{},
		&models.TenantSettings{},
		&models.WaitingList{},
		&models.TreatmentProtocol{},
		&models.ConsentTemplate{},
		&models.PatientConsent{},
	)
}

// generateAPIKey generates a secure random API key
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateAPIKey generates a new API key for the tenant
// POST /api/settings/api-key/generate
func GenerateAPIKey(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	userRole := c.GetString("user_role")

	// Only admins can generate API keys
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Apenas administradores podem gerar chaves de API",
		})
		return
	}

	db := database.GetDB()

	// Find the tenant
	var tenant models.Tenant
	if err := db.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Clínica não encontrada",
		})
		return
	}

	// Generate new API key
	apiKey, err := generateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao gerar chave de API",
		})
		return
	}

	// Update tenant with new API key
	if err := db.Model(&tenant).Updates(map[string]interface{}{
		"api_key":        apiKey,
		"api_key_active": true,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao salvar chave de API",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Chave de API gerada com sucesso",
		"api_key": apiKey,
		"active":  true,
		"note":    "IMPORTANTE: Guarde esta chave em local seguro. Ela só será exibida uma vez.",
	})
}

// GetAPIKeyStatus returns the API key status (but not the key itself for security)
// GET /api/settings/api-key/status
func GetAPIKeyStatus(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")

	db := database.GetDB()

	var tenant models.Tenant
	if err := db.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Clínica não encontrada",
		})
		return
	}

	hasKey := tenant.APIKey != ""
	isActive := tenant.APIKeyActive

	// Show masked version of key if it exists
	maskedKey := ""
	if hasKey {
		maskedKey = tenant.APIKey[:8] + "..." + tenant.APIKey[len(tenant.APIKey)-4:]
	}

	c.JSON(http.StatusOK, gin.H{
		"error":      false,
		"has_key":    hasKey,
		"active":     isActive,
		"masked_key": maskedKey,
	})
}

// ToggleAPIKey enables or disables the API key
// PATCH /api/settings/api-key/toggle
func ToggleAPIKey(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	userRole := c.GetString("user_role")

	// Only admins can toggle API keys
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Apenas administradores podem gerenciar chaves de API",
		})
		return
	}

	type ToggleRequest struct {
		Active bool `json:"active"`
	}

	var req ToggleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Dados inválidos",
		})
		return
	}

	db := database.GetDB()

	var tenant models.Tenant
	if err := db.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Clínica não encontrada",
		})
		return
	}

	if tenant.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Nenhuma chave de API foi gerada ainda",
		})
		return
	}

	if err := db.Model(&tenant).Update("api_key_active", req.Active).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao atualizar status da chave de API",
		})
		return
	}

	statusMsg := "desativada"
	if req.Active {
		statusMsg = "ativada"
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": fmt.Sprintf("Chave de API %s com sucesso", statusMsg),
		"active":  req.Active,
	})
}

// RevokeAPIKey revokes and deletes the API key
// DELETE /api/settings/api-key
func RevokeAPIKey(c *gin.Context) {
	tenantID := c.GetUint("tenant_id")
	userRole := c.GetString("user_role")

	// Only admins can revoke API keys
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   true,
			"message": "Apenas administradores podem revogar chaves de API",
		})
		return
	}

	db := database.GetDB()

	var tenant models.Tenant
	if err := db.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Clínica não encontrada",
		})
		return
	}

	if tenant.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Nenhuma chave de API para revogar",
		})
		return
	}

	if err := db.Model(&tenant).Updates(map[string]interface{}{
		"api_key":        "",
		"api_key_active": false,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Erro ao revogar chave de API",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"message": "Chave de API revogada com sucesso",
	})
}

// WhatsAppAPIDocumentation returns documentation for the WhatsApp API
// GET /api/settings/api-key/docs
func WhatsAppAPIDocumentation(c *gin.Context) {
	baseURL := c.Request.Host

	docs := gin.H{
		"description": "API para integração com WhatsApp e assistentes de IA",
		"authentication": gin.H{
			"type":   "API Key",
			"header": "X-API-Key",
			"description": "Inclua sua chave de API no header X-API-Key de todas as requisições",
		},
		"base_url": fmt.Sprintf("https://%s/api/whatsapp", baseURL),
		"endpoints": []gin.H{
			{
				"method":      "POST",
				"path":        "/verify",
				"description": "Verifica identidade do paciente por CPF e data de nascimento",
				"body": gin.H{
					"cpf":        "string (obrigatório) - CPF do paciente (com ou sem pontuação)",
					"birth_date": "string (obrigatório) - Data de nascimento (DD/MM/AAAA ou AAAA-MM-DD)",
				},
				"response": gin.H{
					"valid":      "boolean - Se a identidade foi verificada",
					"patient_id": "number - ID do paciente (para uso em outras chamadas)",
					"name":       "string - Nome do paciente",
					"message":    "string - Mensagem para exibir ao usuário",
				},
			},
			{
				"method":      "GET",
				"path":        "/appointments?patient_id=X",
				"description": "Lista consultas agendadas do paciente",
				"params":      "patient_id (obrigatório) - ID do paciente obtido na verificação",
			},
			{
				"method":      "GET",
				"path":        "/appointments/history?patient_id=X&limit=10",
				"description": "Lista histórico de consultas do paciente",
			},
			{
				"method":      "POST",
				"path":        "/appointments/cancel?patient_id=X",
				"description": "Cancela uma consulta",
				"body": gin.H{
					"appointment_id": "number (obrigatório) - ID da consulta",
					"reason":         "string (opcional) - Motivo do cancelamento",
				},
			},
			{
				"method":      "POST",
				"path":        "/appointments/reschedule?patient_id=X",
				"description": "Remarca uma consulta",
				"body": gin.H{
					"appointment_id": "number (obrigatório) - ID da consulta",
					"new_date":       "string (obrigatório) - Nova data (AAAA-MM-DD)",
					"new_time":       "string (obrigatório) - Novo horário (HH:MM)",
				},
			},
			{
				"method":      "GET",
				"path":        "/slots?date=AAAA-MM-DD&dentist_id=X",
				"description": "Lista horários disponíveis para uma data",
			},
			{
				"method":      "POST",
				"path":        "/waiting-list?patient_id=X",
				"description": "Adiciona paciente à lista de espera",
				"body": gin.H{
					"procedure": "string (obrigatório) - Tipo de procedimento",
					"priority":  "string (opcional) - 'normal' ou 'urgent'",
					"notes":     "string (opcional) - Observações",
				},
			},
			{
				"method":      "GET",
				"path":        "/waiting-list?patient_id=X",
				"description": "Lista entradas do paciente na lista de espera",
			},
			{
				"method":      "DELETE",
				"path":        "/waiting-list/:id?patient_id=X",
				"description": "Remove paciente da lista de espera",
			},
			{
				"method":      "GET",
				"path":        "/procedures",
				"description": "Lista tipos de procedimentos disponíveis",
			},
			{
				"method":      "GET",
				"path":        "/dentists",
				"description": "Lista profissionais disponíveis",
			},
		},
		"example_flow": []string{
			"1. Chamar POST /verify com CPF e data de nascimento do paciente",
			"2. Se válido, usar o patient_id retornado nas demais chamadas",
			"3. GET /appointments para ver consultas agendadas",
			"4. POST /appointments/cancel para cancelar, ou",
			"5. GET /slots para ver horários disponíveis",
			"6. POST /appointments/reschedule para remarcar",
			"7. POST /waiting-list para entrar na lista de espera",
		},
		"generated_at": time.Now().Format("2006-01-02 15:04:05"),
	}

	c.JSON(http.StatusOK, docs)
}
