package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const (
	OTPLength     = 6
	OTPExpiration = 15 * time.Minute
	MaxOTPAttempts = 3
)

// GenerateOTP generates a random 6-digit OTP code
func GenerateOTP() (string, error) {
	// Generate a random 6-digit number
	max := big.NewInt(1000000) // 10^6 for 6 digits
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Format with leading zeros if necessary
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// ValidateOTP checks if the provided OTP matches the stored one
func ValidateOTP(storedOTP, providedOTP string, expiresAt *time.Time, attempts int) (bool, string) {
	// Check if OTP exists
	if storedOTP == "" {
		return false, "Código de verificação não foi gerado"
	}

	// Check if OTP has expired
	if expiresAt == nil || time.Now().After(*expiresAt) {
		return false, "Código de verificação expirado. Solicite um novo código"
	}

	// Check attempts limit
	if attempts >= MaxOTPAttempts {
		return false, "Número máximo de tentativas excedido. Solicite um novo código"
	}

	// Validate OTP
	if storedOTP != providedOTP {
		return false, "Código de verificação incorreto"
	}

	return true, ""
}

// GetOTPExpirationTime returns the expiration time for a new OTP
func GetOTPExpirationTime() time.Time {
	return time.Now().Add(OTPExpiration)
}

// SendOTPEmail sends the OTP code via email
func SendOTPEmail(toEmail, toName, otpCode string) error {
	subject := "Código de Verificação - Solicitação LGPD"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #66BB6A;">Código de Verificação</h2>

        <p>Olá %s,</p>

        <p>Recebemos uma solicitação de dados pessoais (LGPD) em seu nome. Para garantir a segurança dos seus dados, precisamos verificar sua identidade.</p>

        <div style="background-color: #f5f5f5; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px;">
            <p style="margin: 0; font-size: 14px; color: #666;">Seu código de verificação é:</p>
            <p style="margin: 10px 0; font-size: 32px; font-weight: bold; color: #66BB6A; letter-spacing: 5px;">%s</p>
            <p style="margin: 0; font-size: 12px; color: #999;">Este código expira em 15 minutos</p>
        </div>

        <p><strong>Importante:</strong></p>
        <ul>
            <li>Não compartilhe este código com ninguém</li>
            <li>Se você não fez esta solicitação, ignore este email</li>
            <li>Você tem até 3 tentativas para inserir o código correto</li>
        </ul>

        <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">

        <p style="font-size: 12px; color: #999;">
            Este é um email automático enviado em conformidade com a Lei Geral de Proteção de Dados (LGPD).<br>
            Em caso de dúvidas, entre em contato com nossa equipe.
        </p>
    </div>
</body>
</html>
`, toName, otpCode)

	return SendEmail(toEmail, subject, body)
}
