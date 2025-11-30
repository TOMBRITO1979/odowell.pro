package helpers

import (
	"regexp"
	"strings"
	"unicode"
)

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
