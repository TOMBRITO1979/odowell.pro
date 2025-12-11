package scheduler

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/models"
	"fmt"
	"log"
	"time"
)

// StartSLAChecker starts the daily SLA checker for LGPD data requests
func StartSLAChecker() {
	go func() {
		// Run immediately on startup
		checkSLADeadlines()

		// Then run every day at 8:00 AM
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			checkSLADeadlines()
		}
	}()
	log.Println("SLA Checker scheduler started")
}

// checkSLADeadlines checks all pending data requests for SLA deadlines
func checkSLADeadlines() {
	log.Println("Running SLA deadline check...")

	db := database.GetDB()
	if db == nil {
		log.Println("ERROR: Database not available for SLA check")
		return
	}

	// Get all tenants
	var tenants []models.Tenant
	if err := db.Where("active = ?", true).Find(&tenants).Error; err != nil {
		log.Printf("ERROR: Failed to get tenants for SLA check: %v", err)
		return
	}

	for _, tenant := range tenants {
		checkTenantSLADeadlines(tenant)
	}
}

// checkTenantSLADeadlines checks SLA deadlines for a specific tenant
func checkTenantSLADeadlines(tenant models.Tenant) {
	db := database.GetDB()

	// Set search path to tenant schema
	if err := db.Exec(fmt.Sprintf("SET search_path TO %s", tenant.DBSchema)).Error; err != nil {
		log.Printf("ERROR: Failed to set schema for tenant %d: %v", tenant.ID, err)
		return
	}

	// Get pending/in_progress requests
	var requests []models.DataRequest
	if err := db.Where("status IN ?", []string{"pending", "in_progress"}).Find(&requests).Error; err != nil {
		log.Printf("ERROR: Failed to get data requests for tenant %d: %v", tenant.ID, err)
		return
	}

	// Get admin users for this tenant to send alerts
	var admins []models.User
	db.Table("public.users").Where("tenant_id = ? AND role = ? AND active = ?", tenant.ID, "admin", true).Find(&admins)

	for _, request := range requests {
		// Ensure deadline is set
		if request.Deadline == nil {
			deadline := request.CalculateDeadline()
			request.Deadline = &deadline
			db.Save(&request)
		}

		// Check for overdue requests
		if request.IsOverdue() {
			log.Printf("ALERT: Data request #%d for tenant %d is OVERDUE!", request.ID, tenant.ID)
			sendOverdueAlert(tenant, request, admins)
		} else if request.IsNearDeadline() {
			log.Printf("WARNING: Data request #%d for tenant %d is near deadline (%d days remaining)",
				request.ID, tenant.ID, request.DaysRemaining())
			sendNearDeadlineAlert(tenant, request, admins)
		}
	}
}

// sendOverdueAlert sends an email alert for overdue requests
func sendOverdueAlert(tenant models.Tenant, request models.DataRequest, admins []models.User) {
	subject := fmt.Sprintf("[URGENTE] Solicitacao LGPD #%d VENCIDA - %s", request.ID, tenant.Name)

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="background-color: #ff4d4f; color: white; padding: 15px; border-radius: 8px 8px 0 0;">
            <h2 style="margin: 0;">⚠️ ALERTA: Prazo LGPD Vencido</h2>
        </div>

        <div style="background-color: #fff1f0; padding: 20px; border: 1px solid #ffa39e; border-radius: 0 0 8px 8px;">
            <p><strong>A solicitacao abaixo excedeu o prazo legal de 15 dias da LGPD:</strong></p>

            <table style="width: 100%%; border-collapse: collapse; margin: 15px 0;">
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>ID:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">#%d</td>
                </tr>
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>Paciente:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>Tipo:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>Status:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>Data da Solicitacao:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td>
                </tr>
                <tr style="background-color: #ffccc7;">
                    <td style="padding: 8px;"><strong>Prazo:</strong></td>
                    <td style="padding: 8px;"><strong>%s (VENCIDO)</strong></td>
                </tr>
            </table>

            <p style="color: #cf1322; font-weight: bold;">
                ⚠️ ATENÇÃO: O descumprimento dos prazos da LGPD pode resultar em multas de ate 2%% do faturamento anual da empresa.
            </p>

            <p>Acesse o sistema para processar esta solicitacao imediatamente.</p>
        </div>

        <p style="font-size: 12px; color: #999; margin-top: 20px;">
            Este e um email automatico do sistema de gestao LGPD.
        </p>
    </div>
</body>
</html>
`, request.ID, request.PatientName, getTypeLabel(string(request.Type)),
		getStatusLabel(string(request.Status)), request.CreatedAt.Format("02/01/2006"),
		request.Deadline.Format("02/01/2006"))

	// Send to all admins
	for _, admin := range admins {
		if admin.Email != "" {
			if err := helpers.SendEmail(admin.Email, subject, body); err != nil {
				log.Printf("ERROR: Failed to send overdue alert to %s: %v", admin.Email, err)
			}
		}
	}
}

// sendNearDeadlineAlert sends an email alert for requests near deadline
func sendNearDeadlineAlert(tenant models.Tenant, request models.DataRequest, admins []models.User) {
	daysRemaining := request.DaysRemaining()
	subject := fmt.Sprintf("[ATENCAO] Solicitacao LGPD #%d vence em %d dias - %s", request.ID, daysRemaining, tenant.Name)

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <div style="background-color: #faad14; color: white; padding: 15px; border-radius: 8px 8px 0 0;">
            <h2 style="margin: 0;">⏰ ATENCAO: Prazo LGPD Proximo</h2>
        </div>

        <div style="background-color: #fffbe6; padding: 20px; border: 1px solid #ffe58f; border-radius: 0 0 8px 8px;">
            <p><strong>A solicitacao abaixo esta proxima do prazo legal de 15 dias da LGPD:</strong></p>

            <table style="width: 100%%; border-collapse: collapse; margin: 15px 0;">
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>ID:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">#%d</td>
                </tr>
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>Paciente:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>Tipo:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>Status:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td>
                </tr>
                <tr>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;"><strong>Data da Solicitacao:</strong></td>
                    <td style="padding: 8px; border-bottom: 1px solid #ddd;">%s</td>
                </tr>
                <tr style="background-color: #fff1b8;">
                    <td style="padding: 8px;"><strong>Prazo:</strong></td>
                    <td style="padding: 8px;"><strong>%s (%d dias restantes)</strong></td>
                </tr>
            </table>

            <p>Acesse o sistema para processar esta solicitacao antes do vencimento.</p>
        </div>

        <p style="font-size: 12px; color: #999; margin-top: 20px;">
            Este e um email automatico do sistema de gestao LGPD.
        </p>
    </div>
</body>
</html>
`, request.ID, request.PatientName, getTypeLabel(string(request.Type)),
		getStatusLabel(string(request.Status)), request.CreatedAt.Format("02/01/2006"),
		request.Deadline.Format("02/01/2006"), daysRemaining)

	// Send to all admins
	for _, admin := range admins {
		if admin.Email != "" {
			if err := helpers.SendEmail(admin.Email, subject, body); err != nil {
				log.Printf("ERROR: Failed to send near deadline alert to %s: %v", admin.Email, err)
			}
		}
	}
}

func getTypeLabel(typeValue string) string {
	types := map[string]string{
		"access":      "Acesso aos Dados",
		"portability": "Portabilidade",
		"correction":  "Correcao",
		"deletion":    "Exclusao",
		"revocation":  "Revogacao de Consentimento",
	}
	if label, ok := types[typeValue]; ok {
		return label
	}
	return typeValue
}

func getStatusLabel(status string) string {
	statuses := map[string]string{
		"pending":     "Pendente",
		"in_progress": "Em Andamento",
		"completed":   "Concluido",
		"rejected":    "Rejeitado",
	}
	if label, ok := statuses[status]; ok {
		return label
	}
	return status
}
