package handlers

import (
	"drcrwell/backend/internal/helpers"
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateDataRequest creates a new LGPD data request
func CreateDataRequest(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var input struct {
		PatientID   uint   `json:"patient_id" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados invalidos: " + err.Error()})
		return
	}

	// Validate request type
	validTypes := map[string]bool{
		"access":      true,
		"portability": true,
		"correction":  true,
		"deletion":    true,
		"revocation":  true,
	}
	if !validTypes[input.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de solicitacao invalido"})
		return
	}

	// Get patient info
	var patient models.Patient
	if err := db.First(&patient, input.PatientID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Paciente nao encontrado"})
		return
	}

	request := models.DataRequest{
		PatientID:    input.PatientID,
		PatientName:  patient.Name,
		PatientCPF:   patient.CPF,
		Email:        patient.Email,
		Phone:        patient.Phone,
		Type:         models.DataRequestType(input.Type),
		Status:       models.DataRequestStatusPending,
		Description:  input.Description,
		RequestIP:    c.ClientIP(),
		RequestAgent: c.Request.UserAgent(),
	}

	if err := db.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar solicitacao"})
		return
	}

	// Audit log
	helpers.AuditAction(c, "create", "data_requests", request.ID, true, map[string]interface{}{
		"patient_id":   input.PatientID,
		"request_type": input.Type,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Solicitacao criada com sucesso",
		"request": request,
	})
}

// GetDataRequests returns paginated data requests
func GetDataRequests(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Filters
	status := c.Query("status")
	requestType := c.Query("type")
	patientID := c.Query("patient_id")

	query := db.Model(&models.DataRequest{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if requestType != "" {
		query = query.Where("type = ?", requestType)
	}
	if patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}

	var total int64
	query.Count(&total)

	var requests []models.DataRequest
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar solicitacoes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requests":  requests,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetDataRequest returns a single data request
func GetDataRequest(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	var request models.DataRequest
	if err := db.First(&request, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Solicitacao nao encontrada"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"request": request})
}

// UpdateDataRequestStatus updates the status of a data request
func UpdateDataRequestStatus(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalido"})
		return
	}

	var input struct {
		Status          string `json:"status" binding:"required"`
		ResponseNotes   string `json:"response_notes"`
		RejectionReason string `json:"rejection_reason"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados invalidos"})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"completed":   true,
		"rejected":    true,
	}
	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status invalido"})
		return
	}

	var request models.DataRequest
	if err := db.First(&request, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Solicitacao nao encontrada"})
		return
	}

	// Update fields
	request.Status = models.DataRequestStatus(input.Status)
	request.ResponseNotes = input.ResponseNotes

	if input.Status == "rejected" {
		request.RejectionReason = input.RejectionReason
	}

	if input.Status == "completed" || input.Status == "rejected" {
		userID, _ := c.Get("user_id")
		now := time.Now()
		request.ProcessedAt = &now
		if uid, ok := userID.(uint); ok {
			request.ProcessedBy = &uid
		}
	}

	if err := db.Save(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar solicitacao"})
		return
	}

	// Audit log
	helpers.AuditAction(c, "update", "data_requests", uint(id), true, map[string]interface{}{
		"new_status": input.Status,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Solicitacao atualizada com sucesso",
		"request": request,
	})
}

// GetDataRequestStats returns statistics about data requests
func GetDataRequestStats(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusCounts []StatusCount
	db.Model(&models.DataRequest{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&statusCounts)

	type TypeCount struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}
	var typeCounts []TypeCount
	db.Model(&models.DataRequest{}).
		Select("type, count(*) as count").
		Group("type").
		Find(&typeCounts)

	var totalCount int64
	db.Model(&models.DataRequest{}).Count(&totalCount)

	var pendingCount int64
	db.Model(&models.DataRequest{}).Where("status = ?", "pending").Count(&pendingCount)

	c.JSON(http.StatusOK, gin.H{
		"total":      totalCount,
		"pending":    pendingCount,
		"by_status":  statusCounts,
		"by_type":    typeCounts,
	})
}

// GetPatientDataRequests returns data requests for a specific patient
func GetPatientDataRequests(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	patientID, err := strconv.ParseUint(c.Param("patient_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de paciente invalido"})
		return
	}

	var requests []models.DataRequest
	if err := db.Where("patient_id = ?", patientID).Order("created_at DESC").Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar solicitacoes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}
