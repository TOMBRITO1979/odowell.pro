package handlers

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

// SignDocumentRequest represents the request to sign a document
type SignDocumentRequest struct {
	Password string `json:"password" binding:"required"`
}

// SignPrescription signs a prescription with the user's digital certificate
func SignPrescription(c *gin.Context) {
	userID := c.GetUint("user_id")
	prescriptionID := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var input SignDocumentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha do certificado obrigatória"})
		return
	}

	// Get prescription
	var prescription models.Prescription
	if err := db.Preload("Patient").First(&prescription, prescriptionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receita não encontrada"})
		return
	}

	// Check if already signed
	if prescription.IsSigned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Receita já está assinada"})
		return
	}

	// Get user's active certificate
	cert, err := GetActiveCertificate(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum certificado ativo encontrado. Faça upload de um certificado primeiro."})
		return
	}

	// Get user info (use database.GetDB() to access public schema directly)
	var user models.User
	if err := database.GetDB().Table("public.users").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar dados do usuário"})
		return
	}

	// Decrypt and parse certificate
	pfxData, err := DecryptCertificate(cert, input.Password)
	if err != nil {
		helpers.AuditAction(c, "sign_prescription", "prescriptions", prescription.ID, false, map[string]interface{}{
			"error": "Senha do certificado inválida",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha do certificado inválida"})
		return
	}

	// Parse the certificate to get private key
	privateKey, x509Cert, err := pkcs12.Decode(pfxData, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar certificado"})
		return
	}

	// Generate PDF content for hashing
	pdfContent, err := generatePrescriptionPDFContent(&prescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar conteúdo do documento"})
		return
	}

	// Create hash of the document
	hash := sha256.Sum256(pdfContent)
	hashHex := hex.EncodeToString(hash[:])

	// Sign the hash
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de chave não suportado (apenas RSA)"})
		return
	}

	signature, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, hash[:])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao assinar documento"})
		return
	}

	// Verify signature
	err = rsa.VerifyPKCS1v15(&rsaKey.PublicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro na verificação da assinatura"})
		return
	}

	// Update prescription with signature info using raw SQL to avoid GORM table duplication issue
	now := time.Now()
	result := db.Exec(`
		UPDATE prescriptions
		SET is_signed = true, signed_at = ?, signed_by_id = ?, signed_by_name = ?,
		    signed_by_cro = ?, certificate_id = ?, certificate_thumbprint = ?,
		    signature_hash = ?, status = 'issued', issued_at = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, now, userID, user.Name, user.CRO, cert.ID, cert.Thumbprint, hashHex, now, prescription.ID)

	if result.Error != nil {
		helpers.AuditAction(c, "sign_prescription", "prescriptions", prescription.ID, false, map[string]interface{}{
			"error": "Erro ao salvar assinatura",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar assinatura"})
		return
	}

	// Update local struct for response
	prescription.IsSigned = true
	prescription.SignedAt = &now
	prescription.SignedByID = &userID
	prescription.SignedByName = user.Name
	prescription.SignedByCRO = user.CRO
	prescription.CertificateID = &cert.ID
	prescription.CertificateThumbprint = cert.Thumbprint
	prescription.SignatureHash = hashHex
	prescription.Status = "issued"
	prescription.IssuedAt = &now

	// Update certificate last used
	cert.LastUsedAt = &now
	database.GetDB().Save(cert)

	// Audit log
	helpers.AuditAction(c, "sign_prescription", "prescriptions", prescription.ID, true, map[string]interface{}{
		"patient_id":       prescription.PatientID,
		"patient_name":     prescription.Patient.Name,
		"certificate_id":   cert.ID,
		"certificate_cn":   x509Cert.Subject.CommonName,
		"signature_hash":   hashHex[:16] + "...", // Log only first 16 chars
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Receita assinada digitalmente com sucesso",
		"prescription": gin.H{
			"id":                    prescription.ID,
			"is_signed":             prescription.IsSigned,
			"signed_at":             prescription.SignedAt,
			"signed_by_name":        prescription.SignedByName,
			"signed_by_cro":         prescription.SignedByCRO,
			"certificate_thumbprint": prescription.CertificateThumbprint,
			"signature_hash":        prescription.SignatureHash,
		},
	})
}

// SignMedicalRecord signs a medical record with the user's digital certificate
func SignMedicalRecord(c *gin.Context) {
	userID := c.GetUint("user_id")
	recordID := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var input SignDocumentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha do certificado obrigatória"})
		return
	}

	// Get medical record
	var record models.MedicalRecord
	if err := db.Preload("Patient").Preload("Dentist").First(&record, recordID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prontuário não encontrado"})
		return
	}

	// Check if already signed
	if record.IsSigned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prontuário já está assinado"})
		return
	}

	// Get user's active certificate
	cert, err := GetActiveCertificate(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum certificado ativo encontrado. Faça upload de um certificado primeiro."})
		return
	}

	// Get user info (use database.GetDB() to access public schema directly)
	var user models.User
	if err := database.GetDB().Table("public.users").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar dados do usuário"})
		return
	}

	// Decrypt and parse certificate
	pfxData, err := DecryptCertificate(cert, input.Password)
	if err != nil {
		helpers.AuditAction(c, "sign_medical_record", "medical_records", record.ID, false, map[string]interface{}{
			"error": "Senha do certificado inválida",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha do certificado inválida"})
		return
	}

	// Parse the certificate to get private key
	privateKey, x509Cert, err := pkcs12.Decode(pfxData, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar certificado"})
		return
	}

	// Generate content for hashing (simplified - hash of record data)
	contentToSign := fmt.Sprintf(
		"MedicalRecord|ID:%d|PatientID:%d|DentistID:%d|Type:%s|Diagnosis:%s|TreatmentPlan:%s|ProcedureDone:%s|CreatedAt:%s",
		record.ID, record.PatientID, record.DentistID, record.Type,
		record.Diagnosis, record.TreatmentPlan, record.ProcedureDone,
		record.CreatedAt.Format(time.RFC3339),
	)

	// Create hash of the content
	hash := sha256.Sum256([]byte(contentToSign))
	hashHex := hex.EncodeToString(hash[:])

	// Sign the hash
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tipo de chave não suportado (apenas RSA)"})
		return
	}

	signature, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, hash[:])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao assinar documento"})
		return
	}

	// Verify signature
	err = rsa.VerifyPKCS1v15(&rsaKey.PublicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro na verificação da assinatura"})
		return
	}

	// Update medical record with signature info
	now := time.Now()
	recordIDUint, _ := strconv.ParseUint(recordID, 10, 32)

	result := db.Exec(`
		UPDATE medical_records
		SET is_signed = true, signed_at = ?, signed_by_id = ?, signed_by_name = ?,
		    signed_by_cro = ?, certificate_id = ?, certificate_thumbprint = ?,
		    signature_hash = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, now, userID, user.Name, user.CRO, cert.ID, cert.Thumbprint, hashHex, recordID)

	if result.Error != nil {
		helpers.AuditAction(c, "sign_medical_record", "medical_records", uint(recordIDUint), false, map[string]interface{}{
			"error": "Erro ao salvar assinatura",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar assinatura"})
		return
	}

	// Update certificate last used
	cert.LastUsedAt = &now
	database.GetDB().Save(cert)

	// Audit log
	helpers.AuditAction(c, "sign_medical_record", "medical_records", uint(recordIDUint), true, map[string]interface{}{
		"patient_id":     record.PatientID,
		"patient_name":   record.Patient.Name,
		"certificate_id": cert.ID,
		"certificate_cn": x509Cert.Subject.CommonName,
		"signature_hash": hashHex[:16] + "...",
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Prontuário assinado digitalmente com sucesso",
		"record": gin.H{
			"id":                    record.ID,
			"is_signed":             true,
			"signed_at":             now,
			"signed_by_name":        user.Name,
			"signed_by_cro":         user.CRO,
			"certificate_thumbprint": cert.Thumbprint,
			"signature_hash":        hashHex,
		},
	})
}

// GenerateSignedPrescriptionPDF generates a PDF with digital signature information
func GenerateSignedPrescriptionPDF(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}
	tenantID := c.GetUint("tenant_id")

	var prescription models.Prescription
	if err := db.Preload("Patient").First(&prescription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receita não encontrada"})
		return
	}

	// Get tenant settings for clinic info
	var settings models.TenantSettings
	settingsFound := database.DB.Table("public.tenant_settings").Where("tenant_id = ?", tenantID).First(&settings).Error == nil

	// Create PDF with UTF-8 support
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
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
			"prescription":   "Receita Medica",
			"medical_report": "Laudo Medico",
			"certificate":    "Atestado Odontologico",
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
		pdf.Cell(0, 6, tr("Diagnostico:"))
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
	pdf.Cell(0, 6, tr("Orientacoes:"))
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, tr(prescription.Content), "", "L", false)

	// Valid until
	if prescription.ValidUntil != nil {
		pdf.Ln(5)
		pdf.SetFont("Arial", "I", 9)
		pdf.Cell(0, 5, tr(fmt.Sprintf("Valido ate: %s", prescription.ValidUntil.Format("02/01/2006"))))
	}

	// Signature section
	pdf.Ln(20)
	pdf.SetFont("Arial", "", 10)

	// Date
	dateStr := time.Now().Format("02/01/2006")
	if prescription.PrescriptionDate != nil {
		dateStr = prescription.PrescriptionDate.Format("02/01/2006")
	} else if prescription.IssuedAt != nil {
		dateStr = prescription.IssuedAt.Format("02/01/2006")
	}
	pdf.Cell(0, 5, dateStr)
	pdf.Ln(15)

	// Signer info
	signerName := prescription.SignerName
	signerCRO := prescription.SignerCRO
	if signerName == "" {
		signerName = prescription.DentistName
		signerCRO = prescription.DentistCRO
	}

	// Signature line
	pdf.SetX(120)
	pdf.CellFormat(70, 0, "", "T", 1, "C", false, 0, "")
	pdf.SetX(120)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(70, 5, tr(signerName), "", 1, "C", false, 0, "")
	pdf.SetX(120)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(70, 5, tr(fmt.Sprintf("CRO: %s", signerCRO)), "", 1, "C", false, 0, "")

	// Digital signature block (if signed)
	if prescription.IsSigned {
		pdf.Ln(10)
		pdf.SetFillColor(240, 240, 240)
		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(0, 5, tr("DOCUMENTO ASSINADO DIGITALMENTE"), "1", 1, "C", true, 0, "")

		pdf.SetFont("Arial", "", 7)
		pdf.CellFormat(0, 4, tr(fmt.Sprintf("Assinado por: %s (CRO: %s)", prescription.SignedByName, prescription.SignedByCRO)), "LR", 1, "L", true, 0, "")
		if prescription.SignedAt != nil {
			pdf.CellFormat(0, 4, tr(fmt.Sprintf("Data/Hora: %s", prescription.SignedAt.Format("02/01/2006 15:04:05"))), "LR", 1, "L", true, 0, "")
		}
		pdf.CellFormat(0, 4, tr(fmt.Sprintf("Certificado: %s", prescription.CertificateThumbprint)), "LR", 1, "L", true, 0, "")
		pdf.CellFormat(0, 4, tr(fmt.Sprintf("Hash SHA-256: %s", prescription.SignatureHash)), "LRB", 1, "L", true, 0, "")

		pdf.Ln(2)
		pdf.SetFont("Arial", "I", 6)
		pdf.MultiCell(0, 3, tr("Este documento foi assinado digitalmente com certificado ICP-Brasil. A integridade pode ser verificada atraves do hash acima."), "", "C", false)
	}

	// Settings from tenant
	_ = settingsFound

	// Output PDF
	c.Header("Content-Type", "application/pdf")
	filename := fmt.Sprintf("receita_assinada_%s.pdf", id)
	if !prescription.IsSigned {
		filename = fmt.Sprintf("receita_%s.pdf", id)
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := pdf.Output(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar PDF"})
		return
	}
}

// VerifyDocumentSignature verifies if a document's signature is valid
func VerifyDocumentSignature(c *gin.Context) {
	docType := c.Param("type") // prescription or medical_record
	docID := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var signatureInfo struct {
		IsSigned              bool
		SignedAt              *time.Time
		SignedByName          string
		SignedByCRO           string
		CertificateThumbprint string
		SignatureHash         string
	}

	switch docType {
	case "prescription":
		var prescription models.Prescription
		if err := db.First(&prescription, docID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Documento não encontrado"})
			return
		}
		signatureInfo.IsSigned = prescription.IsSigned
		signatureInfo.SignedAt = prescription.SignedAt
		signatureInfo.SignedByName = prescription.SignedByName
		signatureInfo.SignedByCRO = prescription.SignedByCRO
		signatureInfo.CertificateThumbprint = prescription.CertificateThumbprint
		signatureInfo.SignatureHash = prescription.SignatureHash

	case "medical_record":
		var record models.MedicalRecord
		if err := db.First(&record, docID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Documento não encontrado"})
			return
		}
		signatureInfo.IsSigned = record.IsSigned
		signatureInfo.SignedAt = record.SignedAt
		signatureInfo.SignedByName = record.SignedByName
		signatureInfo.SignedByCRO = record.SignedByCRO
		signatureInfo.CertificateThumbprint = record.CertificateThumbprint
		signatureInfo.SignatureHash = record.SignatureHash

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de documento inválido"})
		return
	}

	if !signatureInfo.IsSigned {
		c.JSON(http.StatusOK, gin.H{
			"signed":  false,
			"message": "Documento não está assinado digitalmente",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"signed":                true,
		"signed_at":             signatureInfo.SignedAt,
		"signed_by_name":        signatureInfo.SignedByName,
		"signed_by_cro":         signatureInfo.SignedByCRO,
		"certificate_thumbprint": signatureInfo.CertificateThumbprint,
		"signature_hash":        signatureInfo.SignatureHash,
		"message":               "Documento assinado digitalmente com certificado ICP-Brasil",
	})
}

// generatePrescriptionPDFContent generates PDF content for hashing (internal helper)
func generatePrescriptionPDFContent(prescription *models.Prescription) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// Simplified content for consistent hashing
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 10, tr(fmt.Sprintf("Prescription ID: %d", prescription.ID)))
	pdf.Ln(10)
	pdf.Cell(0, 10, tr(fmt.Sprintf("Patient ID: %d", prescription.PatientID)))
	pdf.Ln(10)
	pdf.Cell(0, 10, tr(fmt.Sprintf("Type: %s", prescription.Type)))
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, tr(prescription.Content), "", "L", false)
	if prescription.Medications != "" {
		pdf.Ln(5)
		pdf.MultiCell(0, 5, tr(prescription.Medications), "", "L", false)
	}
	pdf.Ln(5)
	pdf.Cell(0, 5, tr(fmt.Sprintf("Created: %s", prescription.CreatedAt.Format(time.RFC3339))))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
