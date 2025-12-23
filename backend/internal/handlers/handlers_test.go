package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// testDBCounter is used to create unique database names for each test
var testDBCounter int

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() *gorm.DB {
	testDBCounter++
	dbName := "file::memory:?cache=shared&mode=memory&_journal_mode=WAL"

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}

	// Clean up existing data from previous tests (shared memory)
	db.Exec("DELETE FROM audit_logs")
	db.Exec("DELETE FROM appointments")
	db.Exec("DELETE FROM patients")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM medical_records")
	db.Exec("DELETE FROM prescriptions")
	db.Exec("DELETE FROM exams")
	db.Exec("DELETE FROM budgets")
	db.Exec("DELETE FROM payments")
	db.Exec("DELETE FROM treatments")
	db.Exec("DELETE FROM attachments")
	db.Exec("DELETE FROM patient_consents")

	// Reset auto-increment counters
	db.Exec("DELETE FROM sqlite_sequence")

	// Create audit_logs table for AuditAction helper
	db.Exec(`CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME,
		user_id INTEGER,
		user_email TEXT,
		user_role TEXT,
		action TEXT,
		resource TEXT,
		resource_id INTEGER,
		method TEXT,
		path TEXT,
		details TEXT,
		ip_address TEXT,
		user_agent TEXT,
		success INTEGER
	)`)

	// Set global DB for helpers that use database.GetDB()
	database.DB = db

	// Create tables manually to avoid SQLite compatibility issues
	db.Exec(`CREATE TABLE IF NOT EXISTS patients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		name TEXT NOT NULL,
		cpf TEXT,
		rg TEXT,
		birth_date DATETIME,
		gender TEXT,
		email TEXT,
		phone TEXT,
		cell_phone TEXT,
		address TEXT,
		number TEXT,
		complement TEXT,
		district TEXT,
		city TEXT,
		state TEXT,
		zip_code TEXT,
		allergies TEXT,
		medications TEXT,
		systemic_diseases TEXT,
		blood_type TEXT,
		has_insurance INTEGER DEFAULT 0,
		insurance_name TEXT,
		insurance_number TEXT,
		tags TEXT,
		active INTEGER DEFAULT 1,
		notes TEXT
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		tenant_id INTEGER,
		name TEXT NOT NULL,
		email TEXT,
		password TEXT,
		role TEXT,
		active INTEGER DEFAULT 1,
		phone TEXT,
		cpf TEXT,
		cro TEXT,
		specialty TEXT,
		is_super_admin INTEGER DEFAULT 0
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS appointments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		patient_id INTEGER NOT NULL,
		dentist_id INTEGER NOT NULL,
		start_time DATETIME NOT NULL,
		end_time DATETIME NOT NULL,
		type TEXT,
		procedure TEXT,
		status TEXT DEFAULT 'scheduled',
		confirmed INTEGER DEFAULT 0,
		confirmed_at DATETIME,
		reminder_sent INTEGER DEFAULT 0,
		notes TEXT,
		room TEXT,
		is_recurring INTEGER DEFAULT 0,
		recurrence_rule TEXT
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS medical_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS prescriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS exams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS budgets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS payments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS treatments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS attachments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS patient_consents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER
	)`)

	return db
}

// createTestAppointment creates an appointment directly in the database
func createTestAppointment(db *gorm.DB, patientID, dentistID uint, startTime time.Time, status string) {
	endTime := startTime.Add(1 * time.Hour)
	db.Exec(`INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		patientID, dentistID, startTime, endTime, status, time.Now(), time.Now())
}

// setupTestContext creates a gin context with the test DB and optional user context
func setupTestContext(db *gorm.DB) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("db", db)
	c.Set("user_id", uint(1))
	c.Set("tenant_id", uint(1))
	return c, w
}

// setupTestContextWithBody creates a gin context with a JSON body
func setupTestContextWithBody(db *gorm.DB, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	c.Set("db", db)
	c.Set("user_id", uint(1))
	c.Set("tenant_id", uint(1))
	return c, w
}

// parseJSONResponse parses the response body into a map
func parseJSONResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	return result
}

// createTestPatient creates a patient in the test DB and returns it
func createTestPatient(db *gorm.DB, name string, phone string) models.Patient {
	patient := models.Patient{
		Name:      name,
		Phone:     phone,
		CellPhone: phone,
		Active:    true,
	}
	db.Create(&patient)
	return patient
}

// createTestUser creates a user in the test DB and returns it
func createTestUser(db *gorm.DB, name string, email string) models.User {
	user := models.User{
		Name:     name,
		Email:    email,
		Role:     "dentist",
		Active:   true,
		TenantID: 1,
	}
	db.Create(&user)
	return user
}
