package handlers

import (
	"drcrwell/backend/internal/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

// CreatePrescription creates a new prescription/medical report
func CreatePrescription(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetUint("user_id")
	tenantID := c.GetUint("tenant_id")

	var input models.Prescription
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user (dentist) info
	var dentist models.User
	if err := db.Table("public.users").First(&dentist, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load dentist info"})
		return
	}

	// Get tenant (clinic) info
	var tenant models.Tenant
	if err := db.Table("public.tenants").Where("id = ?", tenantID).First(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load clinic info"})
		return
	}

	// Set default values
	input.DentistID = userID
	input.Status = "draft"

	// Cache clinic and dentist info for document consistency
	input.ClinicName = tenant.Name
	input.ClinicAddress = tenant.Address + ", " + tenant.City + " - " + tenant.State
	if tenant.ZipCode != "" {
		input.ClinicAddress += " - CEP: " + tenant.ZipCode
	}
	input.ClinicPhone = tenant.Phone

	input.DentistName = dentist.Name
	input.DentistCRO = dentist.CRO

	// If signer is specified, load signer info
	if input.SignerID != nil && *input.SignerID > 0 {
		var signer models.User
		if err := db.Table("public.users").First(&signer, *input.SignerID).Error; err == nil {
			input.SignerName = signer.Name
			input.SignerCRO = signer.CRO
		}
	} else {
		// Default: dentist is the signer
		input.SignerID = &userID
		input.SignerName = dentist.Name
		input.SignerCRO = dentist.CRO
	}

	// Create prescription
	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create prescription"})
		return
	}

	// Load relationships
	db.Preload("Patient").Preload("Dentist").First(&input, input.ID)

	c.JSON(http.StatusCreated, gin.H{"prescription": input})
}

// GetPrescriptions retrieves all prescriptions
func GetPrescriptions(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.Prescription{})

	// Filters
	if patientID := c.Query("patient_id"); patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}
	if dentistID := c.Query("dentist_id"); dentistID != "" {
		query = query.Where("dentist_id = ?", dentistID)
	}
	if prescriptionType := c.Query("type"); prescriptionType != "" {
		query = query.Where("type = ?", prescriptionType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var prescriptions []models.Prescription
	if err := query.Preload("Patient").Preload("Dentist").Preload("Signer").
		Offset(offset).Limit(pageSize).Order("created_at DESC").
		Find(&prescriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prescriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prescriptions": prescriptions,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
	})
}

// GetPrescription retrieves a single prescription
func GetPrescription(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var prescription models.Prescription
	if err := db.Preload("Patient").Preload("Dentist").Preload("Signer").
		First(&prescription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prescription not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prescription": prescription})
}

// UpdatePrescription updates a prescription
func UpdatePrescription(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	// Check if prescription exists using raw SQL
	var prescription models.Prescription
	if err := db.Raw("SELECT * FROM prescriptions WHERE id = ? AND deleted_at IS NULL", id).Scan(&prescription).Error; err != nil || prescription.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prescription not found"})
		return
	}

	var input models.Prescription
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If signer is specified, load signer info using raw SQL
	var signerName, signerCRO string
	if input.SignerID != nil && *input.SignerID > 0 {
		var signer models.User
		if err := db.Raw("SELECT * FROM public.users WHERE id = ? AND deleted_at IS NULL", *input.SignerID).Scan(&signer).Error; err == nil && signer.ID > 0 {
			signerName = signer.Name
			signerCRO = signer.CRO
		}
	}

	// Update using Exec to avoid the duplicate table error
	result := db.Exec(`
		UPDATE prescriptions
		SET patient_id = ?, type = ?, title = ?, medications = ?, content = ?,
		    diagnosis = ?, valid_until = ?, notes = ?, prescription_date = ?,
		    signer_id = ?, signer_name = ?, signer_cro = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, input.PatientID, input.Type, input.Title, input.Medications, input.Content,
		input.Diagnosis, input.ValidUntil, input.Notes, input.PrescriptionDate,
		input.SignerID, signerName, signerCRO, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update prescription"})
		return
	}

	// Load the updated prescription with relationships using raw SQL
	db.Raw("SELECT * FROM prescriptions WHERE id = ? AND deleted_at IS NULL", id).Scan(&prescription)

	c.JSON(http.StatusOK, gin.H{"prescription": prescription})
}

// IssuePrescription issues a prescription (changes status from draft to issued)
func IssuePrescription(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var prescription models.Prescription
	if err := db.First(&prescription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prescription not found"})
		return
	}

	if prescription.Status != "draft" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prescription already issued or cancelled"})
		return
	}

	now := time.Now()

	// Update using Model to avoid the duplicate table error
	if err := db.Model(&models.Prescription{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "issued",
		"issued_at":  now,
		"updated_at": now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue prescription"})
		return
	}

	// Load relationships
	db.Preload("Patient").Preload("Dentist").First(&prescription, id)

	c.JSON(http.StatusOK, gin.H{"prescription": prescription})
}

// PrintPrescription marks a prescription as printed and increments print count
func PrintPrescription(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var prescription models.Prescription
	if err := db.First(&prescription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prescription not found"})
		return
	}

	now := time.Now()
	prescription.PrintedAt = &now
	prescription.PrintCount++

	if err := db.Save(&prescription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update print status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Print status updated",
		"print_count": prescription.PrintCount,
	})
}

// DeletePrescription soft deletes a prescription
func DeletePrescription(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	if err := db.Delete(&models.Prescription{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete prescription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prescription deleted successfully"})
}

// GeneratePrescriptionPDF generates a PDF for a prescription
func GeneratePrescriptionPDF(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var prescription models.Prescription
	if err := db.Preload("Patient").First(&prescription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prescription not found"})
		return
	}

	// Create PDF with UTF-8 support
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add UTF-8 font support using CP1252 translator
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Header - Clinic info
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, tr(prescription.ClinicName), "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, tr(prescription.ClinicAddress), "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, tr(fmt.Sprintf("Tel: %s", prescription.ClinicPhone)), "", 1, "C", false, 0, "")

	pdf.Ln(10)

	// Title
	pdf.SetFont("Arial", "B", 14)
	title := prescription.Title
	if title == "" {
		typeLabels := map[string]string{
			"prescription":   "Receita Médica",
			"medical_report": "Laudo Médico",
			"certificate":    "Atestado Odontológico",
			"referral":       "Encaminhamento",
		}
		if label, ok := typeLabels[prescription.Type]; ok {
			title = label
		} else {
			title = "Documento"
		}
	}
	pdf.CellFormat(0, 10, tr(title), "", 1, "L", false, 0, "")

	pdf.Ln(5)

	// Patient info
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 6, tr(fmt.Sprintf("Paciente: %s", prescription.Patient.Name)))
	pdf.Ln(8)

	// Diagnosis
	if prescription.Diagnosis != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 6, tr("Diagnóstico:"))
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 5, tr(prescription.Diagnosis), "", "L", false)
		pdf.Ln(3)
	}

	// Medications
	if prescription.Medications != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(0, 6, tr("Medicamentos:"))
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 5, tr(prescription.Medications), "", "L", false)
		pdf.Ln(3)
	}

	// Content
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 6, tr("Orientações:"))
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, tr(prescription.Content), "", "L", false)

	// Valid until
	if prescription.ValidUntil != nil {
		pdf.Ln(5)
		pdf.SetFont("Arial", "I", 9)
		pdf.Cell(0, 5, tr(fmt.Sprintf("Válido até: %s", prescription.ValidUntil.Format("02/01/2006"))))
	}

	// Signature
	pdf.Ln(20)
	pdf.SetFont("Arial", "", 10)

	// Date - use prescription_date if set, otherwise issued_at, otherwise current date
	dateStr := time.Now().Format("02/01/2006")
	if prescription.PrescriptionDate != nil {
		dateStr = prescription.PrescriptionDate.Format("02/01/2006")
	} else if prescription.IssuedAt != nil {
		dateStr = prescription.IssuedAt.Format("02/01/2006")
	}
	pdf.Cell(0, 5, dateStr)
	pdf.Ln(15)

	// Signer info - use signer if set, otherwise use dentist
	signerName := prescription.SignerName
	signerCRO := prescription.SignerCRO
	if signerName == "" {
		signerName = prescription.DentistName
		signerCRO = prescription.DentistCRO
	}

	// Signer signature line
	pdf.SetX(120)
	pdf.CellFormat(70, 0, "", "T", 1, "C", false, 0, "")
	pdf.SetX(120)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(70, 5, tr(signerName), "", 1, "C", false, 0, "")
	pdf.SetX(120)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(70, 5, tr(fmt.Sprintf("CRO: %s", signerCRO)), "", 1, "C", false, 0, "")

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=receita_%s.pdf", id))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}
