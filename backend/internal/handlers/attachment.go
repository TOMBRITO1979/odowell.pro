package handlers

import (
	"drcrwell/backend/internal/models"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const uploadPath = "./uploads"

func init() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		fmt.Println("Failed to create uploads directory:", err)
	}
}

func UploadAttachment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.GetUint("user_id")

	// Get patient ID from form
	patientIDStr := c.PostForm("patient_id")
	if patientIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "patient_id is required"})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 10MB"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), patientIDStr, ext)
	filePath := filepath.Join(uploadPath, filename)

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create attachment record
	attachment := models.Attachment{
		PatientID:    parseUint(patientIDStr),
		FileName:     file.Filename,
		FilePath:     filePath,
		FileType:     c.PostForm("file_type"),
		MimeType:     file.Header.Get("Content-Type"),
		FileSize:     file.Size,
		Category:     c.PostForm("category"),
		Description:  c.PostForm("description"),
		UploadedByID: userID,
	}

	if err := db.Create(&attachment).Error; err != nil {
		os.Remove(filePath) // Clean up file if DB insert fails
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create attachment record"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"attachment": attachment})
}

func GetAttachment(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var attachment models.Attachment
	if err := db.Preload("Patient").Preload("UploadedBy").First(&attachment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attachment not found"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
		return
	}

	// If download query param is present, serve file for download
	if c.Query("download") == "true" {
		c.FileAttachment(attachment.FilePath, attachment.FileName)
		return
	}

	// Otherwise return attachment metadata
	c.JSON(http.StatusOK, gin.H{"attachment": attachment})
}

func DeleteAttachment(c *gin.Context) {
	id := c.Param("id")
	db := c.MustGet("db").(*gorm.DB)

	var attachment models.Attachment
	if err := db.First(&attachment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attachment not found"})
		return
	}

	// Delete file from disk
	if err := os.Remove(attachment.FilePath); err != nil {
		fmt.Println("Failed to delete file:", err)
	}

	// Delete from database
	if err := db.Delete(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attachment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attachment deleted successfully"})
}

func parseUint(s string) uint {
	var result uint
	fmt.Sscanf(s, "%d", &result)
	return result
}
