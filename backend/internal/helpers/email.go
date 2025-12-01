package helpers

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// GetEmailConfig returns email configuration from environment
func GetEmailConfig() EmailConfig {
	return EmailConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
}

// GetAppName returns the application name from environment or default
func GetAppName() string {
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "Sistema Odontológico"
	}
	return appName
}

// SendEmail sends an email using SMTP
func SendEmail(to, subject, body string) error {
	config := GetEmailConfig()

	if config.Host == "" {
		return fmt.Errorf("SMTP not configured")
	}

	// Build the message
	msg := buildMessage(config.From, to, subject, body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	// Create TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         config.Host,
	}

	// Connect
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		// Try without TLS for port 587
		return sendWithStartTLS(addr, config, to, msg)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Close()

	// Authenticate
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %v", err)
	}

	// Set sender and recipient
	if err := client.Mail(config.From); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send the email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %v", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email body: %v", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data connection: %v", err)
	}

	return client.Quit()
}

// sendWithStartTLS sends email using STARTTLS (port 587)
func sendWithStartTLS(addr string, config EmailConfig, to, msg string) error {
	// Connect without TLS first
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer client.Close()

	// Start TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         config.Host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %v", err)
	}

	// Authenticate
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %v", err)
	}

	// Set sender and recipient
	if err := client.Mail(config.From); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send the email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %v", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email body: %v", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data connection: %v", err)
	}

	return client.Quit()
}

// buildMessage builds the email message
func buildMessage(from, to, subject, body string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return sb.String()
}

// SendVerificationEmail sends account verification email
func SendVerificationEmail(to, name, token, baseURL string) error {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)
	appName := GetAppName()

	subject := fmt.Sprintf("Verifique sua conta - %s", appName)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #1890ff; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #1890ff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #999; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s</h1>
        </div>
        <div class="content">
            <h2>Olá, %s!</h2>
            <p>Bem-vindo ao %s! Para ativar sua conta e começar a usar nosso sistema de gestão odontológica, clique no botão abaixo:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Verificar Minha Conta</a>
            </p>
            <p>Ou copie e cole este link no seu navegador:</p>
            <p style="word-break: break-all; background: #eee; padding: 10px; border-radius: 4px;">%s</p>
            <p><strong>Este link expira em 24 horas.</strong></p>
            <p>Se você não criou uma conta no %s, ignore este email.</p>
        </div>
        <div class="footer">
            <p>Este é um email automático, por favor não responda.</p>
            <p>&copy; 2024 %s - Sistema de Gestão Odontológica</p>
        </div>
    </div>
</body>
</html>
`, appName, name, appName, verifyURL, verifyURL, appName, appName)

	return SendEmail(to, subject, body)
}

// SendPasswordResetEmail sends password reset email
func SendPasswordResetEmail(to, name, token, baseURL string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)
	appName := GetAppName()

	subject := fmt.Sprintf("Redefinir sua senha - %s", appName)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #1890ff; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #1890ff; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #999; }
        .warning { background: #fff3cd; border: 1px solid #ffc107; padding: 10px; border-radius: 4px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s</h1>
        </div>
        <div class="content">
            <h2>Olá, %s!</h2>
            <p>Recebemos uma solicitação para redefinir a senha da sua conta no %s.</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Redefinir Minha Senha</a>
            </p>
            <p>Ou copie e cole este link no seu navegador:</p>
            <p style="word-break: break-all; background: #eee; padding: 10px; border-radius: 4px;">%s</p>
            <div class="warning">
                <strong>Atenção:</strong> Este link expira em 1 hora por motivos de segurança.
            </div>
            <p>Se você não solicitou a redefinição de senha, ignore este email. Sua senha permanecerá inalterada.</p>
        </div>
        <div class="footer">
            <p>Este é um email automático, por favor não responda.</p>
            <p>&copy; 2024 %s - Sistema de Gestão Odontológica</p>
        </div>
    </div>
</body>
</html>
`, appName, name, appName, resetURL, resetURL, appName)

	return SendEmail(to, subject, body)
}
