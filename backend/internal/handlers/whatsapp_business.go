package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	MetaGraphAPIURL     = "https://graph.facebook.com/v18.0"
	MetaGraphAPIVersion = "v18.0"
)

// WhatsApp Message Types
type WhatsAppMessageType string

const (
	MessageTypeTemplate WhatsAppMessageType = "template"
	MessageTypeText     WhatsAppMessageType = "text"
)

// Meta API Request/Response structures
type MetaTemplateMessage struct {
	MessagingProduct string                 `json:"messaging_product"`
	RecipientType    string                 `json:"recipient_type"`
	To               string                 `json:"to"`
	Type             string                 `json:"type"`
	Template         *MetaTemplateComponent `json:"template,omitempty"`
}

type MetaTemplateComponent struct {
	Name       string                   `json:"name"`
	Language   MetaTemplateLanguage     `json:"language"`
	Components []MetaTemplateCompParams `json:"components,omitempty"`
}

type MetaTemplateLanguage struct {
	Code string `json:"code"`
}

type MetaTemplateCompParams struct {
	Type       string               `json:"type"`
	Parameters []MetaTemplateParam  `json:"parameters,omitempty"`
}

type MetaTemplateParam struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type MetaSendMessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

type MetaErrorResponse struct {
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FBTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}

// Template structures from Meta API
type MetaTemplate struct {
	Name       string              `json:"name"`
	Status     string              `json:"status"`
	Category   string              `json:"category"`
	Language   string              `json:"language"`
	ID         string              `json:"id"`
	Components []MetaTemplateComp  `json:"components"`
}

type MetaTemplateComp struct {
	Type    string `json:"type"`
	Format  string `json:"format,omitempty"`
	Text    string `json:"text,omitempty"`
	Example *struct {
		BodyText [][]string `json:"body_text,omitempty"`
	} `json:"example,omitempty"`
}

type MetaTemplatesResponse struct {
	Data   []MetaTemplate `json:"data"`
	Paging struct {
		Cursors struct {
			Before string `json:"before"`
			After  string `json:"after"`
		} `json:"cursors"`
	} `json:"paging"`
}

// Request/Response structures for our API
type SendWhatsAppRequest struct {
	Phone        string                 `json:"phone" binding:"required"`
	TemplateName string                 `json:"template_name" binding:"required"`
	LanguageCode string                 `json:"language_code"` // default: pt_BR
	Parameters   map[string]string      `json:"parameters"`    // key-value for template variables
}

type SendAppointmentConfirmationRequest struct {
	AppointmentID uint `json:"appointment_id" binding:"required"`
}

type WhatsAppTemplateResponse struct {
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	Category   string   `json:"category"`
	Language   string   `json:"language"`
	ID         string   `json:"id"`
	Preview    string   `json:"preview"`
	Parameters []string `json:"parameters"`
}

// GetWhatsAppTemplates fetches approved templates from Meta API
func GetWhatsAppTemplates(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)

	// Get tenant settings from public schema
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configurações não encontradas"})
		return
	}

	// Check if WhatsApp is configured
	if settings.WhatsAppBusinessAccountID == "" || settings.WhatsAppAccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "WhatsApp Business não configurado. Configure o Phone Number ID, Access Token e Business Account ID."})
		return
	}

	// Fetch templates from Meta API
	url := fmt.Sprintf("%s/%s/message_templates?fields=name,status,category,language,components",
		MetaGraphAPIURL, settings.WhatsAppBusinessAccountID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar requisição"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+settings.WhatsAppAccessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao conectar com Meta API: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var metaErr MetaErrorResponse
		json.Unmarshal(body, &metaErr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Erro da Meta API",
			"meta_error": metaErr.Error.Message,
			"meta_code":  metaErr.Error.Code,
		})
		return
	}

	var templatesResp MetaTemplatesResponse
	if err := json.Unmarshal(body, &templatesResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar resposta da Meta"})
		return
	}

	// Transform to our response format
	templates := make([]WhatsAppTemplateResponse, 0)
	for _, t := range templatesResp.Data {
		// Only include approved templates
		if t.Status != "APPROVED" {
			continue
		}

		template := WhatsAppTemplateResponse{
			Name:     t.Name,
			Status:   t.Status,
			Category: t.Category,
			Language: t.Language,
			ID:       t.ID,
		}

		// Extract preview and parameters from components
		for _, comp := range t.Components {
			if comp.Type == "BODY" && comp.Text != "" {
				template.Preview = comp.Text
				// Extract parameters like {{1}}, {{2}}, etc.
				re := regexp.MustCompile(`\{\{(\d+)\}\}`)
				matches := re.FindAllStringSubmatch(comp.Text, -1)
				for _, match := range matches {
					if len(match) > 1 {
						template.Parameters = append(template.Parameters, match[1])
					}
				}
			}
		}

		templates = append(templates, template)
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
		"count":     len(templates),
	})
}

// SendWhatsAppMessage sends a template message via Meta API
func SendWhatsAppMessage(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)

	var req SendWhatsAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant settings from public schema
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configurações não encontradas"})
		return
	}

	// Check if WhatsApp is enabled and configured
	if !settings.WhatsAppEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "WhatsApp não está habilitado para esta clínica"})
		return
	}

	if settings.WhatsAppPhoneNumberID == "" || settings.WhatsAppAccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "WhatsApp Business não configurado corretamente"})
		return
	}

	// Default language
	languageCode := req.LanguageCode
	if languageCode == "" {
		languageCode = "pt_BR"
	}

	// Normalize phone number (remove non-digits, add country code if needed)
	phone := normalizePhoneNumber(req.Phone)

	// Build message payload
	message := MetaTemplateMessage{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               phone,
		Type:             "template",
		Template: &MetaTemplateComponent{
			Name: req.TemplateName,
			Language: MetaTemplateLanguage{
				Code: languageCode,
			},
		},
	}

	// Add parameters if provided
	if len(req.Parameters) > 0 {
		params := make([]MetaTemplateParam, 0)
		// Parameters should be ordered by key (1, 2, 3, etc.)
		for i := 1; i <= len(req.Parameters); i++ {
			key := fmt.Sprintf("%d", i)
			if val, ok := req.Parameters[key]; ok {
				params = append(params, MetaTemplateParam{
					Type: "text",
					Text: val,
				})
			}
		}
		if len(params) > 0 {
			message.Template.Components = []MetaTemplateCompParams{
				{
					Type:       "body",
					Parameters: params,
				},
			}
		}
	}

	// Send message
	result, err := sendMetaMessage(settings.WhatsAppPhoneNumberID, settings.WhatsAppAccessToken, message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message_id": result.Messages[0].ID,
		"to":         phone,
	})
}

// SendAppointmentConfirmation sends a WhatsApp confirmation for a specific appointment
func SendAppointmentConfirmation(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB) // Tenant-scoped db for appointments query
	tenantID := c.MustGet("tenant_id").(uint)
	schemaName := fmt.Sprintf("tenant_%d", tenantID)

	var req SendAppointmentConfirmationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant settings from public schema
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configurações não encontradas"})
		return
	}

	// Check if WhatsApp is enabled and configured
	if !settings.WhatsAppEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "WhatsApp não está habilitado"})
		return
	}

	if settings.WhatsAppTemplateConfirmation == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template de confirmação não configurado"})
		return
	}

	// Get appointment with patient info
	var appointment struct {
		ID          uint      `json:"id"`
		StartTime   time.Time `json:"start_time"`
		PatientName string    `json:"patient_name"`
		PatientPhone string   `json:"patient_phone"`
		DentistName string    `json:"dentist_name"`
		Procedure   string    `json:"procedure"`
	}

	query := fmt.Sprintf(`
		SELECT
			a.id,
			a.start_time,
			p.name as patient_name,
			COALESCE(NULLIF(p.cell_phone, ''), NULLIF(p.phone, ''), '') as patient_phone,
			u.name as dentist_name,
			COALESCE(a.procedure, 'Consulta') as procedure
		FROM %s.appointments a
		JOIN %s.patients p ON a.patient_id = p.id
		JOIN public.users u ON a.dentist_id = u.id
		WHERE a.id = ? AND a.deleted_at IS NULL
	`, schemaName, schemaName)

	if err := db.Raw(query, req.AppointmentID).Scan(&appointment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento não encontrado"})
		return
	}

	if appointment.PatientPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paciente não possui telefone cadastrado"})
		return
	}

	// Format date and time for Brazilian format
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	appointmentTime := appointment.StartTime.In(loc)
	dateStr := appointmentTime.Format("02/01/2006")
	timeStr := appointmentTime.Format("15:04")

	// Normalize phone number
	phone := normalizePhoneNumber(appointment.PatientPhone)

	// Build template message with parameters
	// Common parameters for appointment confirmation:
	// {{1}} = patient name
	// {{2}} = date
	// {{3}} = time
	// {{4}} = dentist name (optional)
	// {{5}} = clinic name (optional)
	message := MetaTemplateMessage{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               phone,
		Type:             "template",
		Template: &MetaTemplateComponent{
			Name: settings.WhatsAppTemplateConfirmation,
			Language: MetaTemplateLanguage{
				Code: "pt_BR",
			},
			Components: []MetaTemplateCompParams{
				{
					Type: "body",
					Parameters: []MetaTemplateParam{
						{Type: "text", Text: appointment.PatientName},
						{Type: "text", Text: dateStr},
						{Type: "text", Text: timeStr},
						{Type: "text", Text: appointment.DentistName},
						{Type: "text", Text: settings.ClinicName},
					},
				},
			},
		},
	}

	// Send message
	result, err := sendMetaMessage(settings.WhatsAppPhoneNumberID, settings.WhatsAppAccessToken, message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log the message sent (optional: save to database for tracking)
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message_id":   result.Messages[0].ID,
		"to":           phone,
		"patient_name": appointment.PatientName,
		"appointment":  fmt.Sprintf("%s às %s", dateStr, timeStr),
	})
}

// TestWhatsAppConnection tests the WhatsApp Business API connection
func TestWhatsAppConnection(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)

	// Get tenant settings from public schema
	var settings models.TenantSettings
	if err := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		log.Printf("[WhatsApp Debug] Error fetching settings for tenant %d: %v", tenantID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Configurações não encontradas"})
		return
	}

	log.Printf("[WhatsApp Debug] Tenant %d - BusinessAccountID: '%s', AccessToken length: %d",
		tenantID, settings.WhatsAppBusinessAccountID, len(settings.WhatsAppAccessToken))

	if settings.WhatsAppBusinessAccountID == "" || settings.WhatsAppAccessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "WhatsApp Business não configurado. Preencha o Business Account ID e Access Token.",
		})
		return
	}

	// Test connection by fetching phone numbers
	url := fmt.Sprintf("%s/%s/phone_numbers", MetaGraphAPIURL, settings.WhatsAppBusinessAccountID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Erro ao criar requisição"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+settings.WhatsAppAccessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Erro ao conectar: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var metaErr MetaErrorResponse
		json.Unmarshal(body, &metaErr)
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      metaErr.Error.Message,
			"error_code": metaErr.Error.Code,
		})
		return
	}

	// Parse phone numbers response
	var phoneResp struct {
		Data []struct {
			ID                  string `json:"id"`
			DisplayPhoneNumber  string `json:"display_phone_number"`
			VerifiedName        string `json:"verified_name"`
			QualityRating       string `json:"quality_rating"`
		} `json:"data"`
	}
	json.Unmarshal(body, &phoneResp)

	phoneNumbers := make([]gin.H, 0)
	for _, p := range phoneResp.Data {
		phoneNumbers = append(phoneNumbers, gin.H{
			"id":             p.ID,
			"phone":          p.DisplayPhoneNumber,
			"verified_name":  p.VerifiedName,
			"quality_rating": p.QualityRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Conexão estabelecida com sucesso!",
		"phone_numbers": phoneNumbers,
	})
}

// WhatsAppWebhookVerify handles the webhook verification challenge from Meta
func WhatsAppWebhookVerify(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	// For webhook verification, we need to check against a stored token
	// Since this is a multi-tenant system, we use a global webhook token
	// that routes to tenant-specific handling
	verifyToken := c.Query("verify_token") // Our custom verify token

	if mode == "subscribe" && token != "" {
		// In production, validate against stored token
		// For now, just return the challenge
		if verifyToken != "" && token == verifyToken {
			c.String(http.StatusOK, challenge)
			return
		}
		// If no custom token, accept (for development)
		c.String(http.StatusOK, challenge)
		return
	}

	c.JSON(http.StatusForbidden, gin.H{"error": "Verification failed"})
}

// WhatsAppWebhookHandler handles incoming webhooks from Meta (message status updates, etc.)
func WhatsAppWebhookHandler(c *gin.Context) {
	// Note: This is a public endpoint, db is accessed directly
	db := database.GetDB()

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log webhook for debugging (in production, save to database)
	fmt.Printf("[WhatsApp Webhook] Received: %+v\n", payload)

	// Process webhook entries
	if entries, ok := payload["entry"].([]interface{}); ok {
		for _, entry := range entries {
			if entryMap, ok := entry.(map[string]interface{}); ok {
				if changes, ok := entryMap["changes"].([]interface{}); ok {
					for _, change := range changes {
						if changeMap, ok := change.(map[string]interface{}); ok {
							processWhatsAppWebhookChange(db, changeMap)
						}
					}
				}
			}
		}
	}

	// Always return 200 OK to acknowledge receipt
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// processWhatsAppWebhookChange processes a single webhook change event
func processWhatsAppWebhookChange(db *gorm.DB, change map[string]interface{}) {
	value, ok := change["value"].(map[string]interface{})
	if !ok {
		return
	}

	// Handle message status updates
	if statuses, ok := value["statuses"].([]interface{}); ok {
		for _, status := range statuses {
			if statusMap, ok := status.(map[string]interface{}); ok {
				messageID := statusMap["id"].(string)
				statusType := statusMap["status"].(string)
				timestamp := statusMap["timestamp"].(string)

				fmt.Printf("[WhatsApp Status] Message %s: %s at %s\n", messageID, statusType, timestamp)

				// TODO: Update message status in database
				// This would update campaign_recipients or a messages table
			}
		}
	}

	// Handle incoming messages (for two-way communication)
	if messages, ok := value["messages"].([]interface{}); ok {
		for _, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				from := msgMap["from"].(string)
				msgType := msgMap["type"].(string)

				fmt.Printf("[WhatsApp Message] From %s, type: %s\n", from, msgType)

				// TODO: Handle incoming messages
				// This could trigger automated responses or notifications
			}
		}
	}
}

// Helper functions

// sendMetaMessage sends a message via Meta Graph API
func sendMetaMessage(phoneNumberID, accessToken string, message MetaTemplateMessage) (*MetaSendMessageResponse, error) {
	url := fmt.Sprintf("%s/%s/messages", MetaGraphAPIURL, phoneNumberID)

	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar mensagem: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar mensagem: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var metaErr MetaErrorResponse
		json.Unmarshal(body, &metaErr)
		return nil, fmt.Errorf("erro da Meta API (%d): %s", metaErr.Error.Code, metaErr.Error.Message)
	}

	var result MetaSendMessageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("erro ao processar resposta: %v", err)
	}

	return &result, nil
}

// normalizePhoneNumber normalizes a Brazilian phone number for WhatsApp
func normalizePhoneNumber(phone string) string {
	// Remove all non-digit characters
	re := regexp.MustCompile(`\D`)
	phone = re.ReplaceAllString(phone, "")

	// If starts with 0, remove it
	if strings.HasPrefix(phone, "0") {
		phone = phone[1:]
	}

	// Add Brazil country code if not present
	if !strings.HasPrefix(phone, "55") {
		phone = "55" + phone
	}

	return phone
}
