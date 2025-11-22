package handlers

import (
	"drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateTask - Criar nova tarefa
func CreateTask(c *gin.Context) {
	var input struct {
		Title          string   `json:"title" binding:"required"`
		Description    string   `json:"description"`
		DueDate        *string  `json:"due_date"`
		Priority       string   `json:"priority"`
		Status         string   `json:"status"`
		ResponsibleIDs []uint   `json:"responsible_ids"`
		Assignments    []struct {
			AssignableType string `json:"assignable_type"`
			AssignableID   uint   `json:"assignable_id"`
		} `json:"assignments"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	// Create task
	task := models.Task{
		Title:       input.Title,
		Description: input.Description,
		Priority:    input.Priority,
		Status:      input.Status,
		CreatedBy:   userID.(uint),
	}

	// Parse due date if provided
	if input.DueDate != nil && *input.DueDate != "" {
		// Note: Frontend should send ISO format date string
		// GORM will handle the conversion
	}

	if err := db.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar tarefa"})
		return
	}

	// Create responsible users
	if len(input.ResponsibleIDs) > 0 {
		for _, userID := range input.ResponsibleIDs {
			// Use raw SQL to avoid GORM association issues
			if err := db.Exec("INSERT INTO task_users (task_id, user_id) VALUES (?, ?)", task.ID, userID).Error; err != nil {
				log.Printf("Error creating task user: %v", err)
			}
		}
	}

	// Create assignments
	if len(input.Assignments) > 0 {
		for _, assignment := range input.Assignments {
			// Use raw SQL to avoid GORM association issues
			if err := db.Exec("INSERT INTO task_assignments (task_id, assignable_type, assignable_id) VALUES (?, ?, ?)",
				task.ID, assignment.AssignableType, assignment.AssignableID).Error; err != nil {
				log.Printf("Error creating task assignment: %v", err)
			}
		}
	}

	// Load relationships
	if err := db.Preload("Creator").Preload("Responsibles.User").Preload("Assignments").First(&task, task.ID).Error; err != nil {
		// Task was created, just return without full relationships
		c.JSON(http.StatusCreated, gin.H{"task": task})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task": task})
}

// GetTasks - Listar todas as tarefas com filtros
func GetTasks(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	var total int64
	var tasks []models.Task

	query := db.Model(&models.Task{})

	// Filter by search (title or description)
	if search := c.Query("search"); search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Filter by status
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by priority
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// Filter by responsible user
	if responsibleID := c.Query("responsible_id"); responsibleID != "" {
		query = query.Joins("JOIN task_users ON task_users.task_id = tasks.id").
			Where("task_users.user_id = ? AND task_users.deleted_at IS NULL", responsibleID)
	}

	// Filter by created by
	if createdBy := c.Query("created_by"); createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}

	// Count total
	query.Count(&total)

	// Load with relationships
	query.Preload("Creator").
		Preload("Responsibles.User").
		Preload("Assignments").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&tasks)

	c.JSON(http.StatusOK, gin.H{
		"tasks":     tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTask - Buscar tarefa por ID
func GetTask(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var task models.Task
	if err := db.Preload("Creator").
		Preload("Responsibles.User").
		Preload("Assignments").
		First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tarefa não encontrada"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// UpdateTask - Atualizar tarefa
func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	var input struct {
		Title          string   `json:"title"`
		Description    string   `json:"description"`
		DueDate        *string  `json:"due_date"`
		Priority       string   `json:"priority"`
		Status         string   `json:"status"`
		ResponsibleIDs []uint   `json:"responsible_ids"`
		Assignments    []struct {
			AssignableType string `json:"assignable_type"`
			AssignableID   uint   `json:"assignable_id"`
		} `json:"assignments"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update task fields directly
	updates := map[string]interface{}{
		"title":       input.Title,
		"description": input.Description,
		"priority":    input.Priority,
		"status":      input.Status,
	}

	if err := db.Model(&models.Task{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar tarefa"})
		return
	}

	// Load task for response
	var task models.Task
	db.First(&task, id)

	// Update responsible users - delete old and create new using raw SQL
	db.Exec("UPDATE task_users SET deleted_at = NOW() WHERE task_id = ? AND deleted_at IS NULL", task.ID)
	if len(input.ResponsibleIDs) > 0 {
		for _, userID := range input.ResponsibleIDs {
			db.Exec("INSERT INTO task_users (task_id, user_id) VALUES (?, ?)", task.ID, userID)
		}
	}

	// Update assignments - delete old and create new using raw SQL
	db.Exec("UPDATE task_assignments SET deleted_at = NOW() WHERE task_id = ? AND deleted_at IS NULL", task.ID)
	if len(input.Assignments) > 0 {
		for _, assignment := range input.Assignments {
			db.Exec("INSERT INTO task_assignments (task_id, assignable_type, assignable_id) VALUES (?, ?, ?)",
				task.ID, assignment.AssignableType, assignment.AssignableID)
		}
	}

	// Reload with relationships
	db.Preload("Creator").
		Preload("Responsibles.User").
		Preload("Assignments").
		First(&task, task.ID)

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// DeleteTask - Excluir tarefa (soft delete)
func DeleteTask(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Soft delete task using raw SQL
	if err := db.Exec("UPDATE tasks SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir tarefa"})
		return
	}

	// Also soft delete related records using raw SQL
	db.Exec("UPDATE task_users SET deleted_at = NOW() WHERE task_id = ? AND deleted_at IS NULL", id)
	db.Exec("UPDATE task_assignments SET deleted_at = NOW() WHERE task_id = ? AND deleted_at IS NULL", id)

	c.JSON(http.StatusOK, gin.H{"message": "Tarefa excluída com sucesso"})
}

// GetPendingCount - Retorna quantidade de tarefas pendentes do usuário logado
func GetPendingCount(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c)
	if !ok {
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	var count int64
	db.Model(&models.Task{}).
		Joins("JOIN task_users ON task_users.task_id = tasks.id").
		Where("task_users.user_id = ? AND task_users.deleted_at IS NULL", userID).
		Where("tasks.status IN (?)", []string{"pending", "in_progress"}).
		Count(&count)

	c.JSON(http.StatusOK, gin.H{"count": count})
}
