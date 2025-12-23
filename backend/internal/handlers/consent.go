package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// ============================================
// CONSENT TEMPLATES CRUD
// ============================================

// CreateConsentTemplate - Criar novo template de consentimento
func CreateConsentTemplate(c *gin.Context) {
	var input struct {
		Title       string `json:"title" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Version     string `json:"version" binding:"required"`
		Description string `json:"description"`
		Active      *bool  `json:"active"`
		IsDefault   *bool  `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// If IsDefault is true, unset other defaults for this type
	if input.IsDefault != nil && *input.IsDefault {
		db.Exec("UPDATE consent_templates SET is_default = false, updated_at = NOW() WHERE type = ? AND is_default = true AND deleted_at IS NULL",
			input.Type)
	}

	template := models.ConsentTemplate{
		Title:       input.Title,
		Type:        input.Type,
		Content:     input.Content,
		Version:     input.Version,
		Description: input.Description,
	}

	if input.Active != nil {
		template.Active = *input.Active
	} else {
		template.Active = true
	}

	if input.IsDefault != nil {
		template.IsDefault = *input.IsDefault
	}

	if err := db.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar template de consentimento"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"template": template})
}

// GetConsentTemplates - Listar todos os templates
func GetConsentTemplates(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var templates []models.ConsentTemplate
	query := db.Model(&models.ConsentTemplate{})

	// Filter by type if provided
	if typeFilter := c.Query("type"); typeFilter != "" {
		query = query.Where("type = ?", typeFilter)
	}

	// Filter by active status
	if activeFilter := c.Query("active"); activeFilter != "" {
		if activeFilter == "true" {
			query = query.Where("active = ?", true)
		} else if activeFilter == "false" {
			query = query.Where("active = ?", false)
		}
	}

	// Filter by default
	if defaultFilter := c.Query("is_default"); defaultFilter == "true" {
		query = query.Where("is_default = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// GetConsentTemplate - Buscar um template específico
func GetConsentTemplate(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var template models.ConsentTemplate
	if err := db.First(&template, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template não encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": template})
}

// UpdateConsentTemplate - Atualizar template
func UpdateConsentTemplate(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var template models.ConsentTemplate
	if err := db.First(&template, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template não encontrado"})
		return
	}

	var input struct {
		Title       string `json:"title"`
		Type        string `json:"type"`
		Content     string `json:"content"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Active      *bool  `json:"active"`
		IsDefault   *bool  `json:"is_default"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If IsDefault is being set to true, unset other defaults for this type
	if input.IsDefault != nil && *input.IsDefault {
		db.Exec("UPDATE consent_templates SET is_default = false, updated_at = NOW() WHERE type = ? AND is_default = true AND id != ? AND deleted_at IS NULL",
			input.Type, uint(id))
	}

	// Update fields
	if input.Title != "" {
		template.Title = input.Title
	}
	if input.Type != "" {
		template.Type = input.Type
	}
	if input.Content != "" {
		template.Content = input.Content
	}
	if input.Version != "" {
		template.Version = input.Version
	}
	if input.Description != "" {
		template.Description = input.Description
	}
	if input.Active != nil {
		template.Active = *input.Active
	}
	if input.IsDefault != nil {
		template.IsDefault = *input.IsDefault
	}

	// Use fresh session to avoid GORM query contamination
	dbSave := db.Session(&gorm.Session{NewDB: true})
	if err := dbSave.Save(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"template": template})
}

// DeleteConsentTemplate - Deletar template
func DeleteConsentTemplate(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Check if template has associated consents
	var consentCount int64
	db.Model(&models.PatientConsent{}).Where("template_id = ?", uint(id)).Count(&consentCount)
	if consentCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Não é possível deletar template com consentimentos associados",
			"count": consentCount,
		})
		return
	}

	if err := db.Delete(&models.ConsentTemplate{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deletado com sucesso"})
}

// toISO88591 converts UTF-8 string to ISO-8859-1 (Latin-1) for PDF compatibility
func toISO88591(s string) string {
	// Convert UTF-8 to ISO-8859-1
	buf := make([]byte, len(s))
	j := 0
	for _, r := range s {
		if r < 256 {
			buf[j] = byte(r)
			j++
		} else {
			// Replace unsupported characters with '?'
			buf[j] = '?'
			j++
		}
	}
	return string(buf[:j])
}

// HTMLSection represents a parsed section from HTML content
type HTMLSection struct {
	Type    string // "title", "subtitle", "paragraph", "list", "listitem"
	Content string
	Level   int // For headers (h1=1, h2=2, etc)
}

// parseHTMLToSections converts HTML content to structured sections for PDF rendering
func parseHTMLToSections(html string) []HTMLSection {
	var sections []HTMLSection

	// Remove div and style tags but keep content
	html = regexp.MustCompile(`<div[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`</div>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<style[^>]*>.*?</style>`).ReplaceAllString(html, "")

	// Process headers (h1-h4)
	h2Re := regexp.MustCompile(`<h2[^>]*>(.*?)</h2>`)
	h3Re := regexp.MustCompile(`<h3[^>]*>(.*?)</h3>`)
	h4Re := regexp.MustCompile(`<h4[^>]*>(.*?)</h4>`)

	// Process paragraphs
	pRe := regexp.MustCompile(`<p[^>]*>(.*?)</p>`)

	// Process lists
	ulRe := regexp.MustCompile(`<ul[^>]*>(.*?)</ul>`)
	liRe := regexp.MustCompile(`<li[^>]*>(.*?)</li>`)

	// Process strong/bold
	strongRe := regexp.MustCompile(`<strong>(.*?)</strong>`)
	bRe := regexp.MustCompile(`<b>(.*?)</b>`)

	// Split by major elements and process in order
	// Use a simple approach: find all matches with positions
	type match struct {
		pos     int
		section HTMLSection
	}

	var matches []match

	// Find h2 headers
	for _, m := range h2Re.FindAllStringSubmatchIndex(html, -1) {
		content := stripHTMLTags(html[m[2]:m[3]])
		matches = append(matches, match{m[0], HTMLSection{"title", content, 2}})
	}

	// Find h3 headers
	for _, m := range h3Re.FindAllStringSubmatchIndex(html, -1) {
		content := stripHTMLTags(html[m[2]:m[3]])
		matches = append(matches, match{m[0], HTMLSection{"subtitle", content, 3}})
	}

	// Find h4 headers
	for _, m := range h4Re.FindAllStringSubmatchIndex(html, -1) {
		content := stripHTMLTags(html[m[2]:m[3]])
		matches = append(matches, match{m[0], HTMLSection{"section", content, 4}})
	}

	// Find paragraphs
	for _, m := range pRe.FindAllStringSubmatchIndex(html, -1) {
		content := html[m[2]:m[3]]
		// Check if contains strong/bold for emphasis
		if strongRe.MatchString(content) || bRe.MatchString(content) {
			content = stripHTMLTags(content)
			matches = append(matches, match{m[0], HTMLSection{"emphasis", content, 0}})
		} else {
			content = stripHTMLTags(content)
			if strings.TrimSpace(content) != "" {
				matches = append(matches, match{m[0], HTMLSection{"paragraph", content, 0}})
			}
		}
	}

	// Find list items (within ul tags)
	for _, ulMatch := range ulRe.FindAllStringSubmatchIndex(html, -1) {
		ulContent := html[ulMatch[2]:ulMatch[3]]
		for _, liMatch := range liRe.FindAllStringSubmatch(ulContent, -1) {
			content := stripHTMLTags(liMatch[1])
			// Replace checkbox markers
			content = strings.ReplaceAll(content, "?", "☐")
			if strings.TrimSpace(content) != "" {
				matches = append(matches, match{ulMatch[0], HTMLSection{"listitem", content, 0}})
			}
		}
	}

	// Sort by position
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].pos > matches[j].pos {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Extract sections
	for _, m := range matches {
		sections = append(sections, m.section)
	}

	return sections
}

// stripHTMLTags removes all HTML tags from a string
func stripHTMLTags(s string) string {
	// Remove all HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")
	// Decode HTML entities
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	// Normalize whitespace
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// renderHTMLContentToPDF renders parsed HTML content to PDF with proper formatting
func renderHTMLContentToPDF(pdf *gofpdf.Fpdf, content string, tr func(string) string) {
	sections := parseHTMLToSections(content)

	for _, section := range sections {
		switch section.Type {
		case "title":
			pdf.SetFont("Arial", "B", 14)
			pdf.SetTextColor(44, 62, 80) // Dark blue-gray
			pdf.Ln(6)
			pdf.MultiCell(0, 7, tr(section.Content), "", "C", false)
			pdf.Ln(4)

		case "subtitle":
			pdf.SetFont("Arial", "B", 12)
			pdf.SetTextColor(52, 73, 94) // Slightly lighter
			pdf.Ln(4)
			pdf.MultiCell(0, 6, tr(section.Content), "", "C", false)
			pdf.Ln(3)

		case "section":
			pdf.SetFont("Arial", "B", 11)
			pdf.SetTextColor(41, 128, 185) // Blue
			pdf.Ln(4)
			pdf.MultiCell(0, 6, tr(section.Content), "", "L", false)
			pdf.Ln(2)

		case "emphasis":
			pdf.SetFont("Arial", "B", 10)
			pdf.SetTextColor(51, 51, 51)
			pdf.MultiCell(0, 5, tr(section.Content), "", "L", false)
			pdf.Ln(1)

		case "paragraph":
			pdf.SetFont("Arial", "", 10)
			pdf.SetTextColor(51, 51, 51)
			pdf.MultiCell(0, 5, tr(section.Content), "", "J", false)
			pdf.Ln(2)

		case "listitem":
			pdf.SetFont("Arial", "", 10)
			pdf.SetTextColor(51, 51, 51)
			// Add bullet point
			pdf.SetX(20)
			pdf.MultiCell(170, 5, tr("• "+section.Content), "", "L", false)
		}
	}
}

// GenerateTemplatePDF - Gerar PDF do template
func GenerateTemplatePDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var template models.ConsentTemplate
	if err := db.First(&template, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template não encontrado"})
		return
	}

	// Get tenant info for clinic data
	tenantID, _ := c.Get("tenant_id")
	var tenant models.Tenant
	dbPublic := middleware.GetDBFromContext(c).(*gorm.DB)
	dbPublic.Exec("SET search_path TO public")
	dbPublic.First(&tenant, tenantID)

	// Create PDF with Unicode translator
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header with clinic name
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(51, 51, 51)
	if tenant.Name != "" {
		pdf.Cell(0, 10, tr(tenant.Name))
		pdf.Ln(6)
	}

	// Clinic contact info
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(100, 100, 100)
	if tenant.Address != "" && tenant.City != "" {
		pdf.Cell(0, 5, tr(tenant.Address+", "+tenant.City+" - "+tenant.State))
		pdf.Ln(5)
	}
	if tenant.Phone != "" {
		pdf.Cell(0, 5, tr("Tel: "+tenant.Phone))
		pdf.Ln(5)
	}
	pdf.Ln(5)

	// Line separator
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)

	// Template info (version and type)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(100, 100, 100)
	typeLabel := getConsentTypeLabel(template.Type)
	pdf.Cell(0, 5, tr(fmt.Sprintf("Versao: %s | Tipo: %s", template.Version, typeLabel)))
	pdf.Ln(8)

	// Render HTML content properly
	renderHTMLContentToPDF(pdf, template.Content, tr)

	// Footer with date
	pdf.Ln(10)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(150, 150, 150)
	pdf.Cell(0, 5, tr(fmt.Sprintf("Documento gerado em: %s", time.Now().Format("02/01/2006 15:04"))))

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=template_%d.pdf", id))

	err = pdf.Output(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar PDF"})
		return
	}
}

// getConsentTypeLabel returns the Portuguese label for consent type
func getConsentTypeLabel(consentType string) string {
	labels := map[string]string{
		"general":      "Geral",
		"treatment":    "Tratamento",
		"procedure":    "Procedimento",
		"anesthesia":   "Anestesia",
		"data_privacy": "Privacidade de Dados",
	}
	if label, ok := labels[consentType]; ok {
		return label
	}
	return consentType
}

// ============================================
// PATIENT CONSENTS
// ============================================

// CreatePatientConsent - Criar novo consentimento (assinar termo)
func CreatePatientConsent(c *gin.Context) {
	var input struct {
		PatientID        uint   `json:"patient_id" binding:"required"`
		TemplateID       uint   `json:"template_id" binding:"required"`
		SignatureData    string `json:"signature_data" binding:"required"`
		SignatureType    string `json:"signature_type" binding:"required"`
		SignerName       string `json:"signer_name" binding:"required"`
		SignerRelation   string `json:"signer_relation" binding:"required"`
		WitnessName      string `json:"witness_name"`
		WitnessSignature string `json:"witness_signature"`
		Notes            string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Get user ID from context with safe type assertion
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// Get template to create snapshot
	var template models.ConsentTemplate
	if err := db.First(&template, input.TemplateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template não encontrado"})
		return
	}

	// Note: Patient validation skipped - foreign key constraint will catch invalid patient IDs

	// Create consent with template snapshot
	consent := models.PatientConsent{
		PatientID:        input.PatientID,
		TemplateID:       input.TemplateID,
		TemplateTitle:    template.Title,
		TemplateVersion:  template.Version,
		TemplateContent:  template.Content,
		TemplateType:     template.Type,
		SignedAt:         time.Now(),
		SignatureData:    input.SignatureData,
		SignatureType:    input.SignatureType,
		SignerName:       input.SignerName,
		SignerRelation:   input.SignerRelation,
		WitnessName:      input.WitnessName,
		WitnessSignature: input.WitnessSignature,
		Notes:            input.Notes,
		IPAddress:        c.ClientIP(),
		UserAgent:        c.Request.UserAgent(),
		SignedByUserID:   userID,
		Status:           models.ConsentStatusActive,
	}

	// Get truly fresh DB session to avoid model contamination from template query
	dbCreate := db.Session(&gorm.Session{NewDB: true})
	if err := dbCreate.Create(&consent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar consentimento"})
		return
	}

	// Load relationships
	dbCreate.Preload("Patient").Preload("Template").First(&consent, consent.ID)

	c.JSON(http.StatusCreated, gin.H{"consent": consent})
}

// GetPatientConsents - Listar consentimentos de um paciente
func GetPatientConsents(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	patientID, err := strconv.ParseUint(c.Param("patient_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de paciente inválido"})
		return
	}

	var consents []models.PatientConsent
	query := db.Preload("Template").Where("patient_id = ?", uint(patientID))

	// Filter by type if provided
	if typeFilter := c.Query("type"); typeFilter != "" {
		query = query.Where("template_type = ?", typeFilter)
	}

	// Filter by status
	if statusFilter := c.Query("status"); statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if err := query.Order("signed_at DESC").Find(&consents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar consentimentos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"consents": consents})
}

// GetConsent - Buscar um consentimento específico
func GetConsent(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var consent models.PatientConsent
	if err := db.Preload("Patient").Preload("Template").First(&consent, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consentimento não encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"consent": consent})
}

// UpdateConsentStatus - Atualizar status do consentimento (revogar, etc)
func UpdateConsentStatus(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var consent models.PatientConsent
	if err := db.First(&consent, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consentimento não encontrado"})
		return
	}

	consent.Status = input.Status
	if input.Notes != "" {
		if consent.Notes != "" {
			consent.Notes = consent.Notes + "\n" + input.Notes
		} else {
			consent.Notes = input.Notes
		}
	}

	if err := db.Save(&consent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"consent": consent})
}

// DeleteConsent - Deletar consentimento (soft delete)
func DeleteConsent(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := db.Delete(&models.PatientConsent{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar consentimento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consentimento deletado com sucesso"})
}

// GenerateConsentPDF - Gerar PDF de um consentimento assinado
func GenerateConsentPDF(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	tenantID := c.GetUint("tenant_id")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Get tenant info
	var tenant models.Tenant
	tenantDB, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	if err := tenantDB.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar informações da clínica"})
		return
	}

	// Get consent with patient info
	var consent models.PatientConsent
	if err := db.Preload("Patient").First(&consent, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consentimento não encontrado"})
		return
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header - Clinic Info
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, tr(tenant.Name))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 9)
	if tenant.Address != "" && tenant.City != "" {
		pdf.Cell(0, 5, tr(tenant.Address+", "+tenant.City+" - "+tenant.State))
		pdf.Ln(5)
	}
	if tenant.Phone != "" {
		pdf.Cell(0, 5, tr("Tel: "+tenant.Phone))
		pdf.Ln(5)
	}
	pdf.Ln(5)

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, tr("TERMO DE CONSENTIMENTO"))
	pdf.Ln(10)

	// Consent Type and Version
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 6, tr(consent.TemplateTitle))
	pdf.Ln(6)

	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(0, 5, tr("Versao: "+consent.TemplateVersion))
	pdf.Ln(8)

	// Patient Info Section
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Informacoes do Paciente"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(60, 6, tr("Nome:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(120, 6, tr(consent.Patient.Name), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	if consent.Patient.CPF != "" {
		pdf.CellFormat(60, 6, tr("CPF:"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(120, 6, consent.Patient.CPF, "1", 0, "L", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.CellFormat(60, 6, tr("Data de Assinatura:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(120, 6, consent.SignedAt.Format("02/01/2006 15:04"), "1", 0, "L", false, 0, "")
	pdf.Ln(10)

	// Consent Content - render HTML properly
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(180, 7, tr("Conteudo do Termo"), "1", 0, "L", true, 0, "")
	pdf.Ln(8)

	// Render HTML content with proper formatting
	renderHTMLContentToPDF(pdf, consent.TemplateContent, tr)
	pdf.Ln(5)

	// Signature Section
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(180, 7, tr("Assinaturas"), "1", 0, "L", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(60, 6, tr("Assinado por:"), "1", 0, "L", false, 0, "")
	pdf.CellFormat(120, 6, tr(consent.SignerName), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Relacao:"), "1", 0, "L", false, 0, "")
	relationLabel := getSignerRelationLabel(consent.SignerRelation)
	pdf.CellFormat(120, 6, tr(relationLabel), "1", 0, "L", false, 0, "")
	pdf.Ln(-1)

	pdf.CellFormat(60, 6, tr("Tipo de Assinatura:"), "1", 0, "L", false, 0, "")
	signatureTypeLabel := getSignatureTypeLabel(consent.SignatureType)
	pdf.CellFormat(120, 6, tr(signatureTypeLabel), "1", 0, "L", false, 0, "")
	pdf.Ln(5)

	// Try to add signature image if available
	if consent.SignatureData != "" {
		if err := addBase64ImageToPDF(pdf, consent.SignatureData, 15, pdf.GetY(), 80, 30); err == nil {
			pdf.Ln(35)
		}
	}

	// Witness if present
	if consent.WitnessName != "" {
		pdf.Ln(5)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 6, tr("Testemunha"))
		pdf.Ln(6)

		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(60, 6, tr("Nome:"), "1", 0, "L", false, 0, "")
		pdf.CellFormat(120, 6, tr(consent.WitnessName), "1", 0, "L", false, 0, "")
		pdf.Ln(5)

		if consent.WitnessSignature != "" {
			if err := addBase64ImageToPDF(pdf, consent.WitnessSignature, 15, pdf.GetY(), 80, 30); err == nil {
				pdf.Ln(35)
			}
		}
	}

	// Metadata Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 4, tr("Documento gerado eletronicamente em "+time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(4)
	pdf.Cell(0, 4, tr("IP: "+consent.IPAddress))

	// Output PDF
	filename := fmt.Sprintf("termo_consentimento_%d_%s.pdf", consent.ID, time.Now().Format("20060102"))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar PDF"})
		return
	}
}

// Helper function to add base64 image to PDF
func addBase64ImageToPDF(pdf *gofpdf.Fpdf, base64Data string, x, y, w, h float64) error {
	// Decode base64 string
	// Format: data:image/png;base64,iVBORw0KG...
	// We need to extract just the base64 part
	if len(base64Data) < 22 {
		return fmt.Errorf("invalid base64 data")
	}

	// Find the comma that separates the header from the data
	dataStart := 0
	for i, c := range base64Data {
		if c == ',' {
			dataStart = i + 1
			break
		}
	}

	if dataStart == 0 {
		// If no comma found, assume it's already just base64
		dataStart = 0
	} else {
		base64Data = base64Data[dataStart:]
	}

	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return err
	}

	// Save to temporary file
	tmpFile := fmt.Sprintf("/tmp/signature_%d.png", time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, imageData, 0644); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	// Add image to PDF
	pdf.ImageOptions(tmpFile, x, y, w, h, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

	return nil
}

// Helper functions for labels
func getSignerRelationLabel(relation string) string {
	labels := map[string]string{
		models.SignerRelationPatient:        "Paciente",
		models.SignerRelationGuardian:       "Responsavel Legal",
		models.SignerRelationRepresentative: "Representante",
	}
	if label, ok := labels[relation]; ok {
		return label
	}
	return relation
}

func getSignatureTypeLabel(sigType string) string {
	labels := map[string]string{
		models.SignatureTypeDigital:     "Digital",
		models.SignatureTypeHandwritten: "Manuscrita",
		models.SignatureTypeTyped:       "Digitada",
	}
	if label, ok := labels[sigType]; ok {
		return label
	}
	return sigType
}

// GetConsentTypes - Listar tipos de consentimento disponíveis
func GetConsentTypes(c *gin.Context) {
	types := []gin.H{
		{"value": models.ConsentTypeTreatment, "label": "Tratamento"},
		{"value": models.ConsentTypeProcedure, "label": "Procedimento"},
		{"value": models.ConsentTypeAnesthesia, "label": "Anestesia"},
		{"value": models.ConsentTypeDataPrivacy, "label": "Privacidade de Dados (LGPD)"},
		{"value": models.ConsentTypeGeneral, "label": "Geral"},
	}

	c.JSON(http.StatusOK, gin.H{"types": types})
}
