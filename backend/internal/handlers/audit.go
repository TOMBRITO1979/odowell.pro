package handlers

import (
	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetAuditLogs returns paginated audit logs with filters
// Only accessible by admin users
func GetAuditLogs(c *gin.Context) {
	// Verify admin role
	userRole, _ := c.Get("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Apenas administradores podem visualizar logs de auditoria."})
		return
	}

	// Get pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get filter params
	userEmail := c.Query("user_email")
	action := c.Query("action")
	resource := c.Query("resource")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	success := c.Query("success")

	// Build query
	db := database.GetDB()
	query := db.Model(&models.AuditLog{})

	// Apply filters
	if userEmail != "" {
		query = query.Where("user_email ILIKE ?", "%"+userEmail+"%")
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			// Add 1 day to include the entire end date
			query = query.Where("created_at < ?", t.AddDate(0, 0, 1))
		}
	}
	if success != "" {
		successBool := success == "true"
		query = query.Where("success = ?", successBool)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get paginated results
	var logs []models.AuditLog
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar logs de auditoria"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetAuditLogStats returns statistics about audit logs
func GetAuditLogStats(c *gin.Context) {
	// Verify admin role
	userRole, _ := c.Get("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	db := database.GetDB()

	// Get counts by action
	type ActionCount struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	var actionCounts []ActionCount
	db.Model(&models.AuditLog{}).
		Select("action, count(*) as count").
		Group("action").
		Order("count DESC").
		Find(&actionCounts)

	// Get counts by resource
	type ResourceCount struct {
		Resource string `json:"resource"`
		Count    int64  `json:"count"`
	}
	var resourceCounts []ResourceCount
	db.Model(&models.AuditLog{}).
		Select("resource, count(*) as count").
		Group("resource").
		Order("count DESC").
		Find(&resourceCounts)

	// Get total count
	var totalCount int64
	db.Model(&models.AuditLog{}).Count(&totalCount)

	// Get failed actions count
	var failedCount int64
	db.Model(&models.AuditLog{}).Where("success = ?", false).Count(&failedCount)

	// Get unique users count
	var uniqueUsers int64
	db.Model(&models.AuditLog{}).Distinct("user_email").Count(&uniqueUsers)

	// Get logs from last 24 hours
	var last24h int64
	db.Model(&models.AuditLog{}).Where("created_at >= ?", time.Now().Add(-24*time.Hour)).Count(&last24h)

	c.JSON(http.StatusOK, gin.H{
		"total_logs":      totalCount,
		"failed_actions":  failedCount,
		"unique_users":    uniqueUsers,
		"last_24h":        last24h,
		"by_action":       actionCounts,
		"by_resource":     resourceCounts,
	})
}

// GetPortalNotificationsCount returns the count of recent portal notifications (last 24h)
func GetPortalNotificationsCount(c *gin.Context) {
	db := database.GetDB()

	var count int64
	db.Model(&models.AuditLog{}).
		Where("resource = ?", "appointments").
		Where("details LIKE ?", "%patient_portal%").
		Where("path LIKE ?", "%/patient/%").
		Where("created_at >= ?", time.Now().Add(-24*time.Hour)).
		Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// GetPortalNotifications returns patient portal activities (bookings/cancellations)
func GetPortalNotifications(c *gin.Context) {
	tenantID, _ := c.Get("tenant_id")

	// Get pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get filter params
	action := c.Query("action") // create, cancel
	unreadOnly := c.Query("unread") == "true"
	_ = unreadOnly // TODO: implement read/unread tracking

	db := database.GetDB()
	query := db.Model(&models.AuditLog{}).
		Where("resource = ?", "appointments").
		Where("details LIKE ?", "%patient_portal%")

	// Filter by tenant - the details JSON contains patient_id which we can use
	// For now, we filter by checking the path contains /patient/
	query = query.Where("path LIKE ?", "%/patient/%")

	if action != "" {
		query = query.Where("action = ?", action)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get paginated results with patient info
	type PortalNotification struct {
		models.AuditLog
		PatientName  string `json:"patient_name"`
		DentistName  string `json:"dentist_name"`
		StartTime    string `json:"start_time"`
	}

	var logs []models.AuditLog
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar notificações"})
		return
	}

	// Enrich with patient and appointment data
	schemaName := "tenant_" + strconv.FormatUint(uint64(tenantID.(uint)), 10)
	notifications := make([]map[string]interface{}, 0)

	for _, log := range logs {
		notification := map[string]interface{}{
			"id":         log.ID,
			"created_at": log.CreatedAt,
			"action":     log.Action,
			"resource":   log.Resource,
			"details":    log.Details,
			"success":    log.Success,
		}

		// Try to get appointment details
		if log.ResourceID > 0 {
			var appointment struct {
				ID          uint      `json:"id"`
				StartTime   time.Time `json:"start_time"`
				EndTime     time.Time `json:"end_time"`
				Procedure   string    `json:"procedure"`
				Status      string    `json:"status"`
				PatientName string    `json:"patient_name"`
				DentistName string    `json:"dentist_name"`
			}

			err := db.Raw(`
				SELECT a.id, a.start_time, a.end_time, a.procedure, a.status,
					   p.name as patient_name, u.name as dentist_name
				FROM `+schemaName+`.appointments a
				LEFT JOIN `+schemaName+`.patients p ON a.patient_id = p.id
				LEFT JOIN public.users u ON a.dentist_id = u.id
				WHERE a.id = ?
			`, log.ResourceID).Scan(&appointment).Error

			if err == nil && appointment.ID > 0 {
				notification["appointment"] = appointment
			}
		}

		notifications = append(notifications, notification)
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"page_size":     pageSize,
		"pages":         (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ExportAuditLogsCSV exports audit logs to CSV
func ExportAuditLogsCSV(c *gin.Context) {
	// Verify admin role
	userRole, _ := c.Get("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	// Get filter params
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	db := database.GetDB()
	query := db.Model(&models.AuditLog{})

	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at < ?", t.AddDate(0, 0, 1))
		}
	}

	var logs []models.AuditLog
	if err := query.Order("created_at DESC").Limit(10000).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar logs"})
		return
	}

	// Build CSV
	csv := "Data/Hora,Usuario,Email,Funcao,Acao,Recurso,ID Recurso,Metodo,Caminho,IP,Sucesso\n"
	for _, log := range logs {
		successStr := "Sim"
		if !log.Success {
			successStr = "Nao"
		}
		csv += log.CreatedAt.Format("2006-01-02 15:04:05") + ","
		csv += strconv.FormatUint(uint64(log.UserID), 10) + ","
		csv += "\"" + log.UserEmail + "\","
		csv += "\"" + log.UserRole + "\","
		csv += "\"" + log.Action + "\","
		csv += "\"" + log.Resource + "\","
		csv += strconv.FormatUint(uint64(log.ResourceID), 10) + ","
		csv += "\"" + log.Method + "\","
		csv += "\"" + log.Path + "\","
		csv += "\"" + log.IPAddress + "\","
		csv += successStr + "\n"
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=audit_logs_"+time.Now().Format("2006-01-02")+".csv")
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.String(http.StatusOK, csv)
}
