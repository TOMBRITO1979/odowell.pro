package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestCreateAppointment_Success(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(1 * time.Hour)

	body := map[string]interface{}{
		"patient_id": patient.ID,
		"dentist_id": user.ID,
		"start_time": startTime.Format("2006-01-02T15:04:05"),
		"end_time":   endTime.Format("2006-01-02T15:04:05"),
		"type":       "consultation",
		"procedure":  "Consulta inicial",
		"status":     "scheduled",
	}

	c, w := setupTestContextWithBody(db, body)

	CreateAppointment(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	appointment, ok := result["appointment"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected appointment in response")
	}

	if appointment["procedure"] != "Consulta inicial" {
		t.Errorf("Expected procedure 'Consulta inicial', got '%v'", appointment["procedure"])
	}
}

func TestCreateAppointment_Conflict(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	patient2 := createTestPatient(db, "Test Patient 2", "11888888888")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	startTime := time.Now().Add(24 * time.Hour)

	// Create first appointment
	createTestAppointment(db, patient.ID, user.ID, startTime, "scheduled")

	// Try to create conflicting appointment (overlapping time)
	body := map[string]interface{}{
		"patient_id": patient2.ID,
		"dentist_id": user.ID,
		"start_time": startTime.Add(30 * time.Minute).Format("2006-01-02T15:04:05"),
		"end_time":   startTime.Add(90 * time.Minute).Format("2006-01-02T15:04:05"),
		"status":     "scheduled",
	}

	c, w := setupTestContextWithBody(db, body)

	CreateAppointment(c)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d (conflict), got %d. Body: %s", http.StatusConflict, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	if result["error"] != "Conflito de horário" {
		t.Errorf("Expected conflict error, got '%v'", result["error"])
	}
}

func TestCreateAppointment_InvalidJSON(t *testing.T) {
	db := setupTestDB()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("db", db)

	CreateAppointment(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetAppointments_Success(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	// Create test appointments
	for i := 0; i < 3; i++ {
		startTime := time.Now().Add(time.Duration(i+1) * 24 * time.Hour)
		createTestAppointment(db, patient.ID, user.ID, startTime, "scheduled")
	}

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&page_size=10", nil)

	GetAppointments(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	appointments, ok := result["appointments"].([]interface{})
	if !ok {
		t.Fatal("Expected appointments array in response")
	}

	if len(appointments) != 3 {
		t.Errorf("Expected 3 appointments, got %d", len(appointments))
	}
}

func TestGetAppointments_FilterByDentist(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	user1 := createTestUser(db, "Dr. First", "dr1@test.com")
	user2 := createTestUser(db, "Dr. Second", "dr2@test.com")

	startTime := time.Now().Add(24 * time.Hour)

	// Create appointment for user1
	createTestAppointment(db, patient.ID, user1.ID, startTime, "scheduled")

	// Create appointment for user2
	createTestAppointment(db, patient.ID, user2.ID, startTime.Add(2*time.Hour), "scheduled")

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/?dentist_id=%d", user1.ID), nil)

	GetAppointments(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	appointments, ok := result["appointments"].([]interface{})
	if !ok {
		t.Fatal("Expected appointments array in response")
	}

	if len(appointments) != 1 {
		t.Errorf("Expected 1 appointment for dentist %d, got %d", user1.ID, len(appointments))
	}
}

func TestGetAppointments_FilterByStatus(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	startTime := time.Now().Add(24 * time.Hour)

	// Scheduled appointment
	createTestAppointment(db, patient.ID, user.ID, startTime, "scheduled")

	// Completed appointment
	createTestAppointment(db, patient.ID, user.ID, startTime.Add(2*time.Hour), "completed")

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodGet, "/?status=completed", nil)

	GetAppointments(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	appointments, ok := result["appointments"].([]interface{})
	if !ok {
		t.Fatal("Expected appointments array in response")
	}

	if len(appointments) != 1 {
		t.Errorf("Expected 1 completed appointment, got %d", len(appointments))
	}
}

func TestGetAppointment_Success(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	startTime := time.Now().Add(24 * time.Hour)
	appointment := createTestAppointment(db, patient.ID, user.ID, startTime, "scheduled")

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", appointment.ID)}}

	GetAppointment(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	_, ok := result["appointment"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected appointment in response")
	}
}

func TestGetAppointment_NotFound(t *testing.T) {
	db := setupTestDB()

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	GetAppointment(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	result := parseJSONResponse(w)
	if result["error"] != "Agendamento não encontrado" {
		t.Errorf("Expected not found error, got '%v'", result["error"])
	}
}

func TestUpdateAppointment_Success(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	startTime := time.Now().Add(24 * time.Hour)
	appointment := createTestAppointment(db, patient.ID, user.ID, startTime, "scheduled")

	// Update to a different time slot
	newStartTime := startTime.Add(7 * 24 * time.Hour)
	body := map[string]interface{}{
		"patient_id": patient.ID,
		"dentist_id": user.ID,
		"start_time": newStartTime.Format("2006-01-02T15:04:05"),
		"end_time":   newStartTime.Add(1 * time.Hour).Format("2006-01-02T15:04:05"),
		"status":     "confirmed",
		"procedure":  "Restauração",
	}

	jsonBody, _ := json.Marshal(body)
	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", appointment.ID)}}

	UpdateAppointment(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	updatedAppointment, ok := result["appointment"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected appointment in response")
	}

	if updatedAppointment["status"] != "confirmed" {
		t.Errorf("Expected status 'confirmed', got '%v'", updatedAppointment["status"])
	}
}

func TestUpdateAppointment_NotFound(t *testing.T) {
	db := setupTestDB()

	body := map[string]interface{}{
		"patient_id": 1,
		"dentist_id": 1,
		"start_time": time.Now().Format("2006-01-02T15:04:05"),
		"end_time":   time.Now().Add(1 * time.Hour).Format("2006-01-02T15:04:05"),
	}

	jsonBody, _ := json.Marshal(body)
	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	UpdateAppointment(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteAppointment_Success(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	startTime := time.Now().Add(24 * time.Hour)
	appointment := createTestAppointment(db, patient.ID, user.ID, startTime, "scheduled")

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", appointment.ID)}}

	DeleteAppointment(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	if result["message"] != "Agendamento excluído com sucesso" {
		t.Errorf("Expected success message, got '%v'", result["message"])
	}
}

func TestUpdateAppointmentStatus_Success(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	startTime := time.Now().Add(24 * time.Hour)
	appointment := createTestAppointment(db, patient.ID, user.ID, startTime, "scheduled")

	body := map[string]interface{}{
		"status": "completed",
	}

	jsonBody, _ := json.Marshal(body)
	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", appointment.ID)}}

	UpdateAppointmentStatus(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	updatedAppointment, ok := result["appointment"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected appointment in response")
	}

	if updatedAppointment["status"] != "completed" {
		t.Errorf("Expected status 'completed', got '%v'", updatedAppointment["status"])
	}
}

func TestUpdateAppointmentStatus_NotFound(t *testing.T) {
	db := setupTestDB()

	body := map[string]interface{}{
		"status": "completed",
	}

	jsonBody, _ := json.Marshal(body)
	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	UpdateAppointmentStatus(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateAppointmentStatus_MissingStatus(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Test Patient", "11999999999")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	startTime := time.Now().Add(24 * time.Hour)
	appointment := createTestAppointment(db, patient.ID, user.ID, startTime, "scheduled")

	body := map[string]interface{}{}

	jsonBody, _ := json.Marshal(body)
	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", appointment.ID)}}

	UpdateAppointmentStatus(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateAppointment_NoDB(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("{}"))
	c.Request.Header.Set("Content-Type", "application/json")
	// Don't set "db" in context

	CreateAppointment(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
