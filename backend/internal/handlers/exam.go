package handlers

import (
        "drcrwell/backend/internal/middleware"
	"drcrwell/backend/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ExamResponse represents exam with related data
type ExamResponse struct {
	models.Exam
	PatientName    string `json:"patient_name"`
	UploadedByName string `json:"uploaded_by_name"`
}

// MarshalJSON implements custom JSON marshaling to include both embedded and additional fields
func (e ExamResponse) MarshalJSON() ([]byte, error) {
	type Alias ExamResponse
	return json.Marshal(&struct {
		*Alias
		PatientName    string `json:"patient_name"`
		UploadedByName string `json:"uploaded_by_name"`
	}{
		Alias:          (*Alias)(&e),
		PatientName:    e.PatientName,
		UploadedByName: e.UploadedByName,
	})
}

// enrichExamWithRelatedData adds patient and uploaded_by names to exam
func enrichExamWithRelatedData(db *gorm.DB, exam *models.Exam) ExamResponse {
	fmt.Printf("DEBUG enrichExam: Starting enrichment for exam ID %d, patient_id=%d, uploaded_by_id=%d\n",
		exam.ID, exam.PatientID, exam.UploadedByID)

	response := ExamResponse{Exam: *exam}

	// Get patient name
	var patient models.Patient
	fmt.Printf("DEBUG enrichExam: Querying patient with ID %d\n", exam.PatientID)
	if err := db.Select("name").First(&patient, exam.PatientID).Error; err == nil {
		response.PatientName = patient.Name
		fmt.Printf("DEBUG enrichExam: Found patient name: %s\n", patient.Name)
	} else {
		fmt.Printf("DEBUG enrichExam: Error getting patient: %v\n", err)
	}

	// Get uploaded_by user name (from public schema)
	var user models.User
	fmt.Printf("DEBUG enrichExam: Querying user with ID %d from public.users\n", exam.UploadedByID)
	if err := db.Raw("SELECT name FROM public.users WHERE id = ? AND deleted_at IS NULL", exam.UploadedByID).Scan(&user).Error; err == nil {
		response.UploadedByName = user.Name
		fmt.Printf("DEBUG enrichExam: Found user name: %s\n", user.Name)
	} else {
		fmt.Printf("DEBUG enrichExam: Error getting user: %v\n", err)
	}

	fmt.Printf("DEBUG enrichExam: Response PatientName=%s, UploadedByName=%s\n",
		response.PatientName, response.UploadedByName)

	return response
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// S3 configuration - uses environment variables with fallback defaults
func getS3Config() (bucket, region, baseFolder string) {
	bucket = getEnvOrDefault("AWS_BUCKET_NAME", "drcrwell-app")
	region = getEnvOrDefault("AWS_REGION", "sa-east-1")
	baseFolder = getEnvOrDefault("S3_BASE_FOLDER", "exams") // exams/[cpf]/filename
	return
}

// getS3Client creates and returns an S3 client
func getS3Client() (*s3.S3, error) {
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// If credentials are not set, return error (will be configured later)
	if awsAccessKey == "" || awsSecretKey == "" {
		return nil, fmt.Errorf("AWS credentials not configured")
	}

	_, region, _ := getS3Config()

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
	})
	if err != nil {
		return nil, err
	}

	return s3.New(sess), nil
}

// sanitizeCPF removes special characters from CPF
func sanitizeCPF(cpf string) string {
	cpf = strings.ReplaceAll(cpf, ".", "")
	cpf = strings.ReplaceAll(cpf, "-", "")
	cpf = strings.ReplaceAll(cpf, "/", "")
	return cpf
}

// CreateExam uploads an exam file to S3 and creates database record
func CreateExam(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }
	userID := c.GetUint("user_id")

	// Log request details for debugging
	fmt.Printf("DEBUG: Content-Type: %s\n", c.ContentType())
	fmt.Printf("DEBUG: Request method: %s\n", c.Request.Method)

	// Parse form data
	patientIDStr := c.PostForm("patient_id")
	fmt.Printf("DEBUG: patient_id from form: '%s'\n", patientIDStr)
	if patientIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "patient_id is required"})
		return
	}

	patientID, err := strconv.ParseUint(patientIDStr, 10, 32)
	if err != nil {
		fmt.Printf("DEBUG: Error parsing patient_id: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient_id format"})
		return
	}

	name := c.PostForm("name")
	fmt.Printf("DEBUG: name from form: '%s'\n", name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exam name is required"})
		return
	}

	// Get patient to obtain CPF for folder structure
	fmt.Printf("DEBUG: About to fetch patient with ID: %d\n", patientID)
	var patient models.Patient
	if err := db.First(&patient, patientID).Error; err != nil {
		fmt.Printf("DEBUG: Patient not found, error: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
		return
	}
	fmt.Printf("DEBUG: Patient found: %s (CPF: %s)\n", patient.Name, patient.CPF)

	// Get uploaded file
	fmt.Printf("DEBUG: About to get file from form\n")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fmt.Printf("DEBUG: Error getting file: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required", "details": err.Error()})
		return
	}
	defer file.Close()
	fmt.Printf("DEBUG: File received: %s (size: %d bytes)\n", header.Filename, header.Size)

	// Validate file size (50MB max)
	const maxFileSize = 50 * 1024 * 1024 // 50MB
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 50MB limit"})
		return
	}

	// Validate file type (basic check)
	contentType := header.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"application/pdf":  true,
		"image/jpeg":       true,
		"image/jpg":        true,
		"image/png":        true,
		"image/gif":        true,
		"application/zip":  true,
		"application/x-zip-compressed": true,
	}
	if contentType != "" && !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: PDF, JPEG, PNG, GIF, ZIP"})
		return
	}

	// Prepare S3 upload
	s3Client, err := getS3Client()
	if err != nil {
		// If AWS not configured, save locally as fallback
		uploadExamLocal(c, db, userID, uint(patientID), name, file, header)
		return
	}

	// Get S3 configuration
	bucket, region, baseFolder := getS3Config()

	// Generate S3 key: exams/[cpf]/[timestamp]_[filename]
	cpf := sanitizeCPF(patient.CPF)
	timestamp := time.Now().Unix()
	s3Key := fmt.Sprintf("%s/%s/%d_%s", baseFolder, cpf, timestamp, header.Filename)

	// Reset file pointer to beginning before reading
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset file pointer"})
			return
		}
	}

	// Upload to S3 - use the file directly as io.Reader (don't convert to string)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(s3Key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to upload to S3: %v", err)})
		return
	}

	// Generate S3 URL
	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, s3Key)

	// Prepare exam data
	description := c.PostForm("description")
	examType := c.PostForm("exam_type")
	notes := c.PostForm("notes")

	// Parse exam_date if provided
	var examDate *time.Time
	if examDateStr := c.PostForm("exam_date"); examDateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", examDateStr)
		if err == nil {
			examDate = &parsedDate
		}
	}

	// Insert using raw SQL to avoid GORM reflection issues
	var examID uint
	err = db.Raw(`
		INSERT INTO exams (patient_id, name, description, exam_type, exam_date, file_url, s3_key,
			file_name, file_type, file_size, uploaded_by_id, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		RETURNING id
	`, patientID, name, description, examType, examDate, fileURL, s3Key,
		header.Filename, contentType, header.Size, userID, notes).Scan(&examID).Error

	if err != nil {
		// If database insert fails, try to delete from S3
		s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3Key),
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create exam record",
			"details": err.Error(),
		})
		return
	}

	// Fetch the created exam
	var exam models.Exam
	db.First(&exam, examID)

	// Enrich with related data
	examResponse := enrichExamWithRelatedData(db, &exam)

	c.JSON(http.StatusCreated, gin.H{"exam": examResponse})
}

// uploadExamLocal is a fallback when S3 is not configured
func uploadExamLocal(c *gin.Context, db *gorm.DB, userID uint, patientID uint, name string, file multipart.File, header *multipart.FileHeader) {
	// Reset file pointer to beginning before reading
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset file pointer"})
			return
		}
	}

	// Create uploads directory structure
	uploadsDir := "./uploads/exams"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, header.Filename)
	filePath := filepath.Join(uploadsDir, filename)

	// Save file locally
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer out.Close()

	fileSize, err := io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file"})
		os.Remove(filePath) // Clean up partial file
		return
	}

	// Create database record
	exam := models.Exam{
		PatientID:    patientID,
		Name:         name,
		Description:  c.PostForm("description"),
		ExamType:     c.PostForm("exam_type"),
		FileURL:      fmt.Sprintf("/uploads/exams/%s", filename),
		S3Key:        filename, // Using filename as key for local storage
		FileName:     header.Filename,
		FileType:     header.Header.Get("Content-Type"),
		FileSize:     fileSize,
		UploadedByID: userID,
		Notes:        c.PostForm("notes"),
	}

	if err := db.Create(&exam).Error; err != nil {
		os.Remove(filePath) // Clean up file if database insert fails
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create exam record",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"exam": exam,
		"message": "File uploaded locally (AWS S3 not configured)",
	})
}

// GetExams retrieves all exams for a patient
func GetExams(c *gin.Context) {
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	offset := (page - 1) * pageSize

	query := db.Model(&models.Exam{})

	// Filter by patient_id if provided (optional)
	if patientID := c.Query("patient_id"); patientID != "" {
		query = query.Where("patient_id = ?", patientID)
	}

	// Filters
	if examType := c.Query("exam_type"); examType != "" {
		query = query.Where("exam_type = ?", examType)
	}

	var total int64
	query.Count(&total)

	var exams []models.Exam
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").
		Find(&exams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exams"})
		return
	}

	// Enrich exams with related data
	enrichedExams := make([]ExamResponse, len(exams))
	for i, exam := range exams {
		enrichedExams[i] = enrichExamWithRelatedData(db, &exam)
	}

	c.JSON(http.StatusOK, gin.H{
		"exams":     enrichedExams,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetExam retrieves a single exam
func GetExam(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	var exam models.Exam
	if err := db.First(&exam, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	// Get patient name (from current tenant schema)
	var patientName string
	if err := db.Raw("SELECT name FROM patients WHERE id = ? AND deleted_at IS NULL", exam.PatientID).Scan(&patientName).Error; err == nil {
		fmt.Printf("DEBUG GetExam: Found patient: %s (ID: %d)\n", patientName, exam.PatientID)
	} else {
		fmt.Printf("DEBUG GetExam: Error finding patient ID %d: %v\n", exam.PatientID, err)
	}

	// Get uploaded_by user name (from public schema)
	var uploadedByName string
	if err := db.Raw("SELECT name FROM public.users WHERE id = ? AND deleted_at IS NULL", exam.UploadedByID).Scan(&uploadedByName).Error; err == nil {
		fmt.Printf("DEBUG GetExam: Found user: %s (ID: %d)\n", uploadedByName, exam.UploadedByID)
	} else {
		fmt.Printf("DEBUG GetExam: Error finding user ID %d: %v\n", exam.UploadedByID, err)
	}

	// Build response manually
	response := gin.H{
		"id":               exam.ID,
		"created_at":       exam.CreatedAt,
		"updated_at":       exam.UpdatedAt,
		"patient_id":       exam.PatientID,
		"name":             exam.Name,
		"description":      exam.Description,
		"exam_type":        exam.ExamType,
		"exam_date":        exam.ExamDate,
		"file_url":         exam.FileURL,
		"s3_key":           exam.S3Key,
		"file_name":        exam.FileName,
		"file_type":        exam.FileType,
		"file_size":        exam.FileSize,
		"uploaded_by_id":   exam.UploadedByID,
		"notes":            exam.Notes,
		"patient_name":     patientName,
		"uploaded_by_name": uploadedByName,
	}

	fmt.Printf("DEBUG GetExam: Returning patient_name=%s, uploaded_by_name=%s\n", patientName, uploadedByName)

	c.JSON(http.StatusOK, gin.H{"exam": response})
}

// UpdateExam updates exam metadata (name, description, type, notes, exam_date)
func UpdateExam(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	// Check if exam exists
	var count int64
	if err := db.Model(&models.Exam{}).Where("id = ?", id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	var input struct {
		Name        string     `json:"name"`
		Description string     `json:"description"`
		ExamType    string     `json:"exam_type"`
		Notes       string     `json:"notes"`
		ExamDate    *time.Time `json:"exam_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update using Exec to avoid the duplicate table error
	result := db.Exec(`
		UPDATE exams
		SET name = ?, description = ?, exam_type = ?, notes = ?, exam_date = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, input.Name, input.Description, input.ExamType, input.Notes, input.ExamDate, id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exam"})
		return
	}

	// Load the updated exam
	var exam models.Exam
	db.First(&exam, id)

	c.JSON(http.StatusOK, gin.H{"exam": exam})
}

// DeleteExam deletes an exam and its file from S3
func DeleteExam(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	var exam models.Exam
	if err := db.First(&exam, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	// Try to delete from S3
	s3Client, err := getS3Client()
	if err == nil && exam.S3Key != "" {
		bucket, _, _ := getS3Config()
		s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(exam.S3Key),
		})
	} else {
		// Delete local file if S3 not configured
		if strings.HasPrefix(exam.FileURL, "/uploads/") {
			os.Remove(filepath.Join(".", exam.FileURL))
		}
	}

	// Delete database record
	if err := db.Delete(&exam).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exam"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exam deleted successfully"})
}

// GetExamDownloadURL generates a presigned URL for downloading the exam file
func GetExamDownloadURL(c *gin.Context) {
	id := c.Param("id")
	db, ok := middleware.GetDBFromContextSafe(c); if !ok { return }

	var exam models.Exam
	if err := db.First(&exam, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exam not found"})
		return
	}

	s3Client, err := getS3Client()
	if err != nil {
		// If S3 not configured, return the local file URL
		c.JSON(http.StatusOK, gin.H{
			"download_url": exam.FileURL,
			"expires_in":   3600,
		})
		return
	}

	// Generate presigned URL valid for 1 hour
	bucket, _, _ := getS3Config()
	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(exam.S3Key),
	})

	url, err := req.Presign(1 * time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate download URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": url,
		"expires_in":   3600, // 1 hour in seconds
	})
}
