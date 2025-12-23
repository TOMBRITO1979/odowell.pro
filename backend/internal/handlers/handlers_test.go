package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"drcrwell/backend/internal/database"
	"drcrwell/backend/internal/models"
)

const testSchema = "test_handlers"

func init() {
	gin.SetMode(gin.TestMode)
}

// getTestDSN returns the PostgreSQL connection string for tests (without search_path)
func getTestDSN() string {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "odowell_app"
	}
	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "1rc6mGjOgGAq69kL9jKCUN3kWJ2OLd3a"
	}
	dbname := os.Getenv("TEST_DB_NAME")
	if dbname == "" {
		dbname = "drcrwell_db"
	}
	sslmode := os.Getenv("TEST_DB_SSL_MODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=America/Sao_Paulo",
		host, port, user, password, dbname, sslmode,
	)
}

// getTestDSNWithSchema returns DSN with search_path set via connection options
func getTestDSNWithSchema(schema string) string {
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "odowell_app"
	}
	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "1rc6mGjOgGAq69kL9jKCUN3kWJ2OLd3a"
	}
	dbname := os.Getenv("TEST_DB_NAME")
	if dbname == "" {
		dbname = "drcrwell_db"
	}
	sslmode := os.Getenv("TEST_DB_SSL_MODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	// Set search_path via options parameter - this is applied at connection time
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=America/Sao_Paulo options='-c search_path=%s,public'",
		host, port, user, password, dbname, sslmode, schema,
	)
}

// setupTestDB creates a PostgreSQL test schema for testing
// Uses single connection with explicit search_path
func setupTestDB() *gorm.DB {
	dsn := getTestDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}

	// Limit to single connection to ensure search_path persists
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Drop and recreate test schema for clean slate
	if err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", testSchema)).Error; err != nil {
		panic("failed to drop schema: " + err.Error())
	}
	if err := db.Exec(fmt.Sprintf("CREATE SCHEMA %s", testSchema)).Error; err != nil {
		panic("failed to create schema: " + err.Error())
	}

	// Set search_path to test schema and public (for User model)
	if err := db.Exec(fmt.Sprintf("SET search_path TO %s, public", testSchema)).Error; err != nil {
		panic("failed to set search_path: " + err.Error())
	}

	// Set global DB for helpers that use database.GetDB()
	database.DB = db

	// Create tables using raw SQL to avoid GORM foreign key issues across schemas
	// This is necessary because Appointment references User which is in public schema
	tableSQL := []string{
		`CREATE TABLE IF NOT EXISTS patients (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			name VARCHAR(255) NOT NULL,
			cpf VARCHAR(14),
			rg VARCHAR(20),
			birth_date TIMESTAMP,
			gender VARCHAR(10),
			email VARCHAR(255),
			phone VARCHAR(20),
			cell_phone VARCHAR(20),
			address TEXT,
			number VARCHAR(20),
			complement TEXT,
			district VARCHAR(100),
			city VARCHAR(100),
			state VARCHAR(2),
			zip_code VARCHAR(10),
			allergies TEXT,
			medications TEXT,
			systemic_diseases TEXT,
			blood_type VARCHAR(5),
			has_insurance BOOLEAN DEFAULT FALSE,
			insurance_name VARCHAR(255),
			insurance_number VARCHAR(50),
			tags TEXT,
			active BOOLEAN DEFAULT TRUE,
			notes TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS appointments (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			patient_id INTEGER NOT NULL,
			dentist_id INTEGER NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			type VARCHAR(50),
			procedure VARCHAR(255),
			status VARCHAR(20) DEFAULT 'scheduled',
			confirmed BOOLEAN DEFAULT FALSE,
			confirmed_at TIMESTAMP,
			reminder_sent BOOLEAN DEFAULT FALSE,
			notes TEXT,
			room VARCHAR(50),
			is_recurring BOOLEAN DEFAULT FALSE,
			recurrence_rule TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS medical_records (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			patient_id INTEGER NOT NULL,
			dentist_id INTEGER,
			appointment_id INTEGER,
			chief_complaint TEXT,
			diagnosis TEXT,
			treatment_plan TEXT,
			procedures_performed TEXT,
			notes TEXT,
			odontogram JSONB,
			anamnesis JSONB,
			is_signed BOOLEAN DEFAULT FALSE,
			signed_at TIMESTAMP,
			signature_hash VARCHAR(255),
			signed_by INTEGER,
			signature_data TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS prescriptions (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			patient_id INTEGER NOT NULL,
			dentist_id INTEGER NOT NULL,
			medical_record_id INTEGER,
			medications TEXT,
			instructions TEXT,
			notes TEXT,
			is_signed BOOLEAN DEFAULT FALSE,
			signed_at TIMESTAMP,
			signature_hash VARCHAR(255)
		)`,
		`CREATE TABLE IF NOT EXISTS exams (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			patient_id INTEGER NOT NULL,
			dentist_id INTEGER,
			medical_record_id INTEGER,
			type VARCHAR(100),
			name VARCHAR(255),
			description TEXT,
			file_url VARCHAR(500),
			file_name VARCHAR(255),
			file_type VARCHAR(100),
			result TEXT,
			notes TEXT,
			exam_date TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS budgets (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			patient_id INTEGER NOT NULL,
			dentist_id INTEGER,
			procedures JSONB,
			total_value DECIMAL(10,2),
			discount DECIMAL(10,2) DEFAULT 0,
			final_value DECIMAL(10,2),
			status VARCHAR(20) DEFAULT 'pending',
			notes TEXT,
			valid_until TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			patient_id INTEGER NOT NULL,
			budget_id INTEGER,
			amount DECIMAL(10,2) NOT NULL,
			payment_method VARCHAR(50),
			payment_date TIMESTAMP,
			notes TEXT,
			status VARCHAR(20) DEFAULT 'pending'
		)`,
		`CREATE TABLE IF NOT EXISTS treatments (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			patient_id INTEGER NOT NULL,
			dentist_id INTEGER,
			name VARCHAR(255),
			description TEXT,
			status VARCHAR(20) DEFAULT 'pending',
			start_date TIMESTAMP,
			end_date TIMESTAMP,
			notes TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS attachments (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			patient_id INTEGER NOT NULL,
			file_name VARCHAR(255),
			file_url VARCHAR(500),
			file_type VARCHAR(100),
			description TEXT,
			category VARCHAR(100)
		)`,
		`CREATE TABLE IF NOT EXISTS patient_consents (
			id SERIAL PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			patient_id INTEGER NOT NULL,
			consent_type VARCHAR(100),
			description TEXT,
			signed_at TIMESTAMP,
			signature_data TEXT,
			is_signed BOOLEAN DEFAULT FALSE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_patients_deleted_at ON patients(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_appointments_deleted_at ON appointments(deleted_at)`,
		`CREATE INDEX IF NOT EXISTS idx_appointments_patient_id ON appointments(patient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_appointments_dentist_id ON appointments(dentist_id)`,
		`CREATE INDEX IF NOT EXISTS idx_appointments_start_time ON appointments(start_time)`,
	}
	for _, sql := range tableSQL {
		if err := db.Exec(sql).Error; err != nil {
			panic(fmt.Sprintf("failed to create table: %v", err))
		}
	}

	return db
}

// createTestAppointment creates an appointment directly in the database
func createTestAppointment(db *gorm.DB, patientID, dentistID uint, startTime time.Time, status string) models.Appointment {
	endTime := startTime.Add(1 * time.Hour)
	appointment := models.Appointment{
		PatientID: patientID,
		DentistID: dentistID,
		StartTime: models.LocalTime{Time: startTime},
		EndTime:   models.LocalTime{Time: endTime},
		Status:    status,
	}
	db.Create(&appointment)
	return appointment
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

// createTestUser creates a user in the public.users table and returns it
func createTestUser(db *gorm.DB, name string, email string) models.User {
	var user models.User

	// First check if user already exists (to avoid duplicates in tests)
	if err := db.Where("email = ?", email).First(&user).Error; err == nil {
		return user
	}

	// Create user using raw SQL to avoid schema conflicts
	// User model uses public.users table explicitly
	db.Exec(`INSERT INTO public.users (name, email, password, role, active, tenant_id, created_at, updated_at)
		VALUES ($1, $2, 'test_password_hash', $3, $4, $5, NOW(), NOW())
		ON CONFLICT (email) DO NOTHING`, name, email, "dentist", true, 1)

	// Fetch the created or existing user
	db.Where("email = ?", email).First(&user)
	return user
}
