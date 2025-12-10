package handlers

import (
	"drcrwell/backend/internal/scheduler"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetRetentionStats returns data retention statistics
func GetRetentionStats(c *gin.Context) {
	// Only admins can view retention stats
	userRole, _ := c.Get("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	stats := scheduler.GetRetentionStats()
	if stats == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao obter estatisticas"})
		return
	}

	// Add retention policy info
	stats["retention_policy"] = map[string]interface{}{
		"audit_logs_days":            scheduler.DefaultRetentionConfig["audit_logs"],
		"expired_tokens_days":        scheduler.DefaultRetentionConfig["expired_tokens"],
		"old_sessions_days":          scheduler.DefaultRetentionConfig["old_sessions"],
		"medical_records_years":      20, // CFO requirement
		"fiscal_documents_years":     5,  // Tax law requirement
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// TriggerRetentionCleanup manually triggers the retention cleanup
func TriggerRetentionCleanup(c *gin.Context) {
	// Only admins can trigger cleanup
	userRole, _ := c.Get("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	results := scheduler.ForceRetentionCleanup()

	c.JSON(http.StatusOK, gin.H{
		"message": "Limpeza de retencao executada",
		"results": results,
	})
}

// GetRetentionPolicy returns the current retention policy configuration
func GetRetentionPolicy(c *gin.Context) {
	policy := map[string]interface{}{
		"audit_logs": map[string]interface{}{
			"retention_days": scheduler.DefaultRetentionConfig["audit_logs"],
			"description":    "Logs de auditoria para rastreamento de acoes",
			"legal_basis":    "LGPD Art. 37 - Registro das operacoes de tratamento",
		},
		"medical_records": map[string]interface{}{
			"retention_years": 20,
			"description":     "Prontuarios medicos e odontologicos",
			"legal_basis":     "CFO Resolucao 118/2012 - Retencao minima de 20 anos",
		},
		"fiscal_documents": map[string]interface{}{
			"retention_years": 5,
			"description":     "Documentos fiscais e notas",
			"legal_basis":     "Legislacao tributaria brasileira",
		},
		"password_resets": map[string]interface{}{
			"retention_days": 1,
			"description":    "Tokens de recuperacao de senha expirados",
			"legal_basis":    "Seguranca - limpeza automatica",
		},
		"email_verifications": map[string]interface{}{
			"retention_days": 7,
			"description":    "Verificacoes de email nao completadas",
			"legal_basis":    "Seguranca - limpeza automatica",
		},
		"consents": map[string]interface{}{
			"retention_years": 5,
			"description":     "Termos de consentimento assinados",
			"legal_basis":     "LGPD Art. 8 - Prova do consentimento",
		},
	}

	c.JSON(http.StatusOK, gin.H{"policy": policy})
}
