package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/models"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/pbkdf2"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

const (
	// Encryption parameters
	aesKeySize   = 32 // AES-256
	saltSize     = 32
	pbkdf2Iter   = 100000
)

// UploadCertificateRequest represents the request to upload a certificate
type UploadCertificateRequest struct {
	Name     string `form:"name" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// UploadCertificate handles the upload and validation of an A1 certificate
func UploadCertificate(c *gin.Context) {
	userID := c.GetUint("user_id")
	db := database.GetDB()

	// Get the uploaded file
	file, err := c.FormFile("certificate")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo de certificado obrigatório"})
		return
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pfx") &&
	   !strings.HasSuffix(strings.ToLower(file.Filename), ".p12") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo deve ser .pfx ou .p12"})
		return
	}

	// Get form values
	name := c.PostForm("name")
	if name == "" {
		name = "Certificado A1"
	}
	password := c.PostForm("password")
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha do certificado obrigatória"})
		return
	}

	// Read file content
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler arquivo"})
		return
	}
	defer f.Close()

	pfxData, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler conteúdo do arquivo"})
		return
	}

	// Parse PKCS#12 to validate and extract info
	privateKey, cert, err := pkcs12.Decode(pfxData, password)
	if err != nil {
		helpers.AuditAction(c, "upload_certificate", "user_certificates", 0, false, map[string]interface{}{
			"error": "Senha inválida ou certificado corrompido",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha inválida ou certificado corrompido"})
		return
	}

	if privateKey == nil || cert == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Certificado não contém chave privada válida"})
		return
	}

	// Validate certificate dates
	now := time.Now()
	if now.Before(cert.NotBefore) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Certificado ainda não é válido (válido a partir de %s)", cert.NotBefore.Format("02/01/2006"))})
		return
	}
	if now.After(cert.NotAfter) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Certificado expirado em %s", cert.NotAfter.Format("02/01/2006"))})
		return
	}

	// Calculate thumbprint (SHA-1 of DER certificate)
	thumbprint := sha1.Sum(cert.Raw)
	thumbprintHex := hex.EncodeToString(thumbprint[:])

	// Check if certificate already exists for this user
	var existing models.UserCertificate
	if err := db.Where("user_id = ? AND thumbprint = ?", userID, thumbprintHex).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Este certificado já está cadastrado"})
		return
	}

	// Check if serial number already exists (global uniqueness)
	serialNumber := cert.SerialNumber.String()
	if err := db.Where("serial_number = ?", serialNumber).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Este certificado já está cadastrado por outro usuário"})
		return
	}

	// Extract CPF from ICP-Brasil certificate (if present)
	cpf := extractCPFFromCert(cert)

	// Encrypt the PFX data
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar salt de criptografia"})
		return
	}

	encryptedPFX, err := encryptPFX(pfxData, password, salt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criptografar certificado"})
		return
	}

	// Deactivate any other active certificates for this user
	db.Model(&models.UserCertificate{}).
		Where("user_id = ? AND active = ?", userID, true).
		Update("active", false)

	// Create certificate record
	userCert := models.UserCertificate{
		UserID:          userID,
		Name:            name,
		SubjectCN:       cert.Subject.CommonName,
		SubjectCPF:      cpf,
		IssuerCN:        cert.Issuer.CommonName,
		SerialNumber:    serialNumber,
		Thumbprint:      thumbprintHex,
		NotBefore:       cert.NotBefore,
		NotAfter:        cert.NotAfter,
		EncryptedPFX:    encryptedPFX,
		EncryptionSalt:  salt,
		Active:          true,
		IsValid:         true,
	}

	if err := db.Create(&userCert).Error; err != nil {
		helpers.AuditAction(c, "upload_certificate", "user_certificates", 0, false, map[string]interface{}{
			"error": "Erro ao salvar certificado",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar certificado"})
		return
	}

	// Audit log
	helpers.AuditAction(c, "upload_certificate", "user_certificates", userCert.ID, true, map[string]interface{}{
		"certificate_name":   name,
		"subject_cn":         cert.Subject.CommonName,
		"issuer_cn":          cert.Issuer.CommonName,
		"valid_until":        cert.NotAfter.Format("02/01/2006"),
		"days_until_expiry":  userCert.DaysUntilExpiry(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Certificado cadastrado com sucesso",
		"certificate": gin.H{
			"id":           userCert.ID,
			"name":         userCert.Name,
			"subject_cn":   userCert.SubjectCN,
			"subject_cpf":  userCert.SubjectCPF,
			"issuer_cn":    userCert.IssuerCN,
			"thumbprint":   userCert.Thumbprint,
			"not_before":   userCert.NotBefore,
			"not_after":    userCert.NotAfter,
			"active":       userCert.Active,
			"days_until_expiry": userCert.DaysUntilExpiry(),
		},
	})
}

// GetUserCertificates returns all certificates for the current user
func GetUserCertificates(c *gin.Context) {
	userID := c.GetUint("user_id")
	db := database.GetDB()

	var certificates []models.UserCertificate
	if err := db.Where("user_id = ?", userID).Order("active DESC, created_at DESC").Find(&certificates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar certificados"})
		return
	}

	// Build response without encrypted data
	result := make([]gin.H, len(certificates))
	for i, cert := range certificates {
		result[i] = gin.H{
			"id":                cert.ID,
			"name":              cert.Name,
			"subject_cn":        cert.SubjectCN,
			"subject_cpf":       cert.SubjectCPF,
			"issuer_cn":         cert.IssuerCN,
			"thumbprint":        cert.Thumbprint,
			"not_before":        cert.NotBefore,
			"not_after":         cert.NotAfter,
			"active":            cert.Active,
			"is_valid":          cert.IsValid && !cert.IsExpired(),
			"is_expired":        cert.IsExpired(),
			"days_until_expiry": cert.DaysUntilExpiry(),
			"last_used_at":      cert.LastUsedAt,
			"created_at":        cert.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"certificates": result,
		"total":        len(result),
	})
}

// ActivateCertificate sets a certificate as the active one for signing
func ActivateCertificate(c *gin.Context) {
	userID := c.GetUint("user_id")
	certID := c.Param("id")
	db := database.GetDB()

	// Find the certificate
	var cert models.UserCertificate
	if err := db.Where("id = ? AND user_id = ?", certID, userID).First(&cert).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificado não encontrado"})
		return
	}

	// Check if certificate is valid
	if cert.IsExpired() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Certificado expirado"})
		return
	}

	// Deactivate all other certificates
	db.Model(&models.UserCertificate{}).
		Where("user_id = ? AND id != ?", userID, certID).
		Update("active", false)

	// Activate this certificate
	cert.Active = true
	if err := db.Save(&cert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ativar certificado"})
		return
	}

	helpers.AuditAction(c, "activate_certificate", "user_certificates", cert.ID, true, map[string]interface{}{
		"certificate_name": cert.Name,
		"thumbprint":       cert.Thumbprint,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificado ativado com sucesso",
		"certificate": gin.H{
			"id":         cert.ID,
			"name":       cert.Name,
			"subject_cn": cert.SubjectCN,
			"active":     cert.Active,
		},
	})
}

// DeleteCertificate removes a certificate
func DeleteCertificate(c *gin.Context) {
	userID := c.GetUint("user_id")
	certID := c.Param("id")
	db := database.GetDB()

	// Find the certificate
	var cert models.UserCertificate
	if err := db.Where("id = ? AND user_id = ?", certID, userID).First(&cert).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificado não encontrado"})
		return
	}

	// Delete (soft delete)
	if err := db.Delete(&cert).Error; err != nil {
		helpers.AuditAction(c, "delete_certificate", "user_certificates", cert.ID, false, map[string]interface{}{
			"error": "Erro ao excluir certificado",
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir certificado"})
		return
	}

	helpers.AuditAction(c, "delete_certificate", "user_certificates", cert.ID, true, map[string]interface{}{
		"certificate_name": cert.Name,
		"thumbprint":       cert.Thumbprint,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Certificado excluído com sucesso"})
}

// ValidateCertificatePassword validates the password for a certificate
func ValidateCertificatePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	certID := c.Param("id")
	db := database.GetDB()

	var input struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha obrigatória"})
		return
	}

	// Find the certificate
	var cert models.UserCertificate
	if err := db.Where("id = ? AND user_id = ?", certID, userID).First(&cert).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificado não encontrado"})
		return
	}

	// Try to decrypt and parse the certificate
	pfxData, err := decryptPFX(cert.EncryptedPFX, input.Password, cert.EncryptionSalt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha inválida", "valid": false})
		return
	}

	// Validate the decrypted data
	_, _, err = pkcs12.Decode(pfxData, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Senha inválida", "valid": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Senha válida",
	})
}

// Helper functions

// extractCPFFromCert extracts CPF from ICP-Brasil certificate OID or Subject
func extractCPFFromCert(cert *x509.Certificate) string {
	// ICP-Brasil certificates may have CPF in the Subject CN or in OID 2.16.76.1.3.1

	// Try to extract from CN (common format: "NAME:CPF")
	cn := cert.Subject.CommonName
	cpfRegex := regexp.MustCompile(`\d{11}`)
	if match := cpfRegex.FindString(cn); match != "" {
		return formatCPF(match)
	}

	// Try to extract from Subject Alternative Names extensions
	for _, ext := range cert.Extensions {
		// OID 2.16.76.1.3.1 is used by ICP-Brasil for CPF
		if ext.Id.String() == "2.16.76.1.3.1" {
			if match := cpfRegex.FindString(string(ext.Value)); match != "" {
				return formatCPF(match)
			}
		}
	}

	return ""
}

// formatCPF formats a CPF number with dots and dash
func formatCPF(cpf string) string {
	if len(cpf) != 11 {
		return cpf
	}
	return fmt.Sprintf("%s.%s.%s-%s", cpf[0:3], cpf[3:6], cpf[6:9], cpf[9:11])
}

// encryptPFX encrypts PFX data using AES-256-GCM
func encryptPFX(pfxData []byte, password string, salt []byte) ([]byte, error) {
	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iter, aesKeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, pfxData, nil)
	return ciphertext, nil
}

// decryptPFX decrypts PFX data encrypted with AES-256-GCM
func decryptPFX(encryptedData []byte, password string, salt []byte) ([]byte, error) {
	// Derive key from password using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, pbkdf2Iter, aesKeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GetActiveCertificate returns the active certificate for a user (internal helper)
func GetActiveCertificate(userID uint) (*models.UserCertificate, error) {
	db := database.GetDB()
	var cert models.UserCertificate
	if err := db.Where("user_id = ? AND active = ?", userID, true).First(&cert).Error; err != nil {
		return nil, err
	}
	if cert.IsExpired() {
		return nil, fmt.Errorf("certificado expirado")
	}
	return &cert, nil
}

// DecryptCertificate decrypts a certificate for signing (internal helper)
func DecryptCertificate(cert *models.UserCertificate, password string) ([]byte, error) {
	return decryptPFX(cert.EncryptedPFX, password, cert.EncryptionSalt)
}
