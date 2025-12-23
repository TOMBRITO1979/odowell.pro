package helpers

import (
	"bytes"
	"io"
	"mime/multipart"
	"regexp"
	"strings"
	"unicode"
)

// Magic number signatures for file type validation
var (
	magicJPEG = []byte{0xFF, 0xD8, 0xFF}
	magicPNG  = []byte{0x89, 0x50, 0x4E, 0x47}
	magicGIF  = []byte{0x47, 0x49, 0x46}
	magicPDF  = []byte{0x25, 0x50, 0x44, 0x46}
	magicZIP  = []byte{0x50, 0x4B} // DOCX, XLSX, ODT, etc
	magicWebP = []byte{0x52, 0x49, 0x46, 0x46} // RIFF header
)

// FileType represents validated file types
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypePDF      FileType = "pdf"
	FileTypeDocument FileType = "document"
	FileTypeUnknown  FileType = "unknown"
)

// ValidateFileMagicNumber reads the first bytes of a file and validates it against known magic numbers
// Returns the detected file type and whether it's valid for the expected types
func ValidateFileMagicNumber(file multipart.File, allowedTypes []FileType) (FileType, bool, error) {
	// Read first 12 bytes for magic number detection
	header := make([]byte, 12)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return FileTypeUnknown, false, err
	}
	header = header[:n]

	// Reset file position
	if seeker, ok := file.(io.Seeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
	}

	detectedType := detectFileType(header)

	// Check if detected type is in allowed list
	for _, allowed := range allowedTypes {
		if detectedType == allowed {
			return detectedType, true, nil
		}
	}

	return detectedType, false, nil
}

// detectFileType determines file type from magic bytes
func detectFileType(header []byte) FileType {
	if len(header) < 3 {
		return FileTypeUnknown
	}

	// Check for image types
	if bytes.HasPrefix(header, magicJPEG) {
		return FileTypeImage
	}
	if bytes.HasPrefix(header, magicPNG) {
		return FileTypeImage
	}
	if bytes.HasPrefix(header, magicGIF) {
		return FileTypeImage
	}
	if bytes.HasPrefix(header, magicWebP) && len(header) >= 12 {
		// WebP has RIFF header, check for WEBP at offset 8
		if bytes.Equal(header[8:12], []byte("WEBP")) {
			return FileTypeImage
		}
	}

	// Check for PDF
	if bytes.HasPrefix(header, magicPDF) {
		return FileTypePDF
	}

	// Check for ZIP-based documents (DOCX, XLSX, etc.)
	if bytes.HasPrefix(header, magicZIP) {
		return FileTypeDocument
	}

	return FileTypeUnknown
}

// ValidateImageFile validates that a file is a valid image (JPEG, PNG, GIF, WebP)
func ValidateImageFile(file multipart.File) (bool, error) {
	_, valid, err := ValidateFileMagicNumber(file, []FileType{FileTypeImage})
	return valid, err
}

// ValidatePDFFile validates that a file is a valid PDF
func ValidatePDFFile(file multipart.File) (bool, error) {
	_, valid, err := ValidateFileMagicNumber(file, []FileType{FileTypePDF})
	return valid, err
}

// ValidateMedicalFile validates that a file is a valid medical document (PDF, image, or document)
func ValidateMedicalFile(file multipart.File) (bool, error) {
	_, valid, err := ValidateFileMagicNumber(file, []FileType{FileTypeImage, FileTypePDF, FileTypeDocument})
	return valid, err
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(s string) string {
	// Trim whitespace
	s = strings.TrimSpace(s)
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	return s
}

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}
	// Basic email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateCPF validates Brazilian CPF format (11 digits)
func ValidateCPF(cpf string) bool {
	// Remove non-digits
	cpf = regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")
	return len(cpf) == 11
}

// ValidatePhone validates phone format
func ValidatePhone(phone string) bool {
	if phone == "" {
		return true // Phone is optional
	}
	// Remove non-digits
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	// Brazilian phone: 10-11 digits
	return len(digits) >= 10 && len(digits) <= 11
}

// ValidatePassword checks password complexity
// Returns (isValid bool, message string)
func ValidatePassword(password string) (bool, string) {
	if len(password) < 12 {
		return false, "A senha deve ter no mínimo 12 caracteres"
	}
	if len(password) > 128 {
		return false, "A senha deve ter no máximo 128 caracteres"
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?~`"
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "A senha deve conter pelo menos uma letra maiúscula"
	}
	if !hasLower {
		return false, "A senha deve conter pelo menos uma letra minúscula"
	}
	if !hasNumber {
		return false, "A senha deve conter pelo menos um número"
	}
	if !hasSpecial {
		return false, "A senha deve conter pelo menos um caractere especial (!@#$%^&*)"
	}

	return true, ""
}

// ValidateName checks if name is valid
func ValidateName(name string) bool {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return false
	}
	if len(name) > 255 {
		return false
	}
	return true
}

// SanitizeInput sanitizes common input fields
func SanitizeInput(input map[string]interface{}) map[string]interface{} {
	for key, value := range input {
		if str, ok := value.(string); ok {
			input[key] = SanitizeString(str)
		}
	}
	return input
}
