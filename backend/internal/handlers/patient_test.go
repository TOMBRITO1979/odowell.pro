package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreatePatient_Success(t *testing.T) {
	db := setupTestDB()

	body := map[string]interface{}{
		"name":       "João Silva",
		"phone":      "11999999999",
		"cell_phone": "11999999999",
		"email":      "joao@example.com",
		"active":     true,
	}

	c, w := setupTestContextWithBody(db, body)

	CreatePatient(c)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	patient, ok := result["patient"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected patient in response")
	}

	if patient["name"] != "João Silva" {
		t.Errorf("Expected name 'João Silva', got '%v'", patient["name"])
	}
}

func TestCreatePatient_MissingPhone(t *testing.T) {
	db := setupTestDB()

	body := map[string]interface{}{
		"name":  "João Silva",
		"email": "joao@example.com",
	}

	c, w := setupTestContextWithBody(db, body)

	CreatePatient(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	result := parseJSONResponse(w)
	if result["error"] != "Telefone é obrigatório" {
		t.Errorf("Expected phone required error, got '%v'", result["error"])
	}
}

func TestCreatePatient_InvalidJSON(t *testing.T) {
	db := setupTestDB()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("db", db)

	CreatePatient(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetPatients_Success(t *testing.T) {
	db := setupTestDB()

	// Create test patients
	createTestPatient(db, "Maria Santos", "11988888888")
	createTestPatient(db, "José Oliveira", "11977777777")

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&page_size=10", nil)

	GetPatients(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	patients, ok := result["patients"].([]interface{})
	if !ok {
		t.Fatal("Expected patients array in response")
	}

	if len(patients) != 2 {
		t.Errorf("Expected 2 patients, got %d", len(patients))
	}

	total := result["total"].(float64)
	if total != 2 {
		t.Errorf("Expected total 2, got %v", total)
	}
}

func TestGetPatients_WithSearch(t *testing.T) {
	db := setupTestDB()

	createTestPatient(db, "Maria Santos", "11988888888")
	createTestPatient(db, "José Oliveira", "11977777777")

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodGet, "/?search=Maria", nil)

	GetPatients(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	patients, ok := result["patients"].([]interface{})
	if !ok {
		t.Fatal("Expected patients array in response")
	}

	if len(patients) != 1 {
		t.Errorf("Expected 1 patient, got %d", len(patients))
	}
}

func TestGetPatient_Success(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Carlos Lima", "11966666666")

	if patient.ID == 0 {
		t.Fatal("Patient was not created properly")
	}

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", patient.ID)}}

	GetPatient(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	patientResult, ok := result["patient"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected patient in response")
	}

	if patientResult["name"] != "Carlos Lima" {
		t.Errorf("Expected name 'Carlos Lima', got '%v'", patientResult["name"])
	}
}

func TestGetPatient_NotFound(t *testing.T) {
	db := setupTestDB()

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	GetPatient(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	result := parseJSONResponse(w)
	if result["error"] != "Paciente não encontrado" {
		t.Errorf("Expected not found error, got '%v'", result["error"])
	}
}

func TestUpdatePatient_Success(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Ana Costa", "11955555555")

	body := map[string]interface{}{
		"name":       "Ana Costa Silva",
		"phone":      "11955555555",
		"cell_phone": "11955555555",
		"email":      "ana.updated@example.com",
		"active":     true,
	}

	jsonBody, _ := json.Marshal(body)
	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", patient.ID)}}

	UpdatePatient(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	updatedPatient, ok := result["patient"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected patient in response")
	}

	if updatedPatient["name"] != "Ana Costa Silva" {
		t.Errorf("Expected name 'Ana Costa Silva', got '%v'", updatedPatient["name"])
	}
}

func TestUpdatePatient_NotFound(t *testing.T) {
	db := setupTestDB()

	body := map[string]interface{}{
		"name":  "Test",
		"phone": "11999999999",
	}

	jsonBody, _ := json.Marshal(body)
	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	UpdatePatient(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdatePatient_MissingPhone(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Pedro Souza", "11944444444")

	body := map[string]interface{}{
		"name": "Pedro Souza Updated",
	}

	jsonBody, _ := json.Marshal(body)
	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodPut, "/", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", patient.ID)}}

	UpdatePatient(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	result := parseJSONResponse(w)
	if result["error"] != "Telefone é obrigatório" {
		t.Errorf("Expected phone required error, got '%v'", result["error"])
	}
}

func TestDeletePatient_Success(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Lucas Alves", "11933333333")

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", patient.ID)}}

	DeletePatient(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	if result["message"] != "Paciente deletado com sucesso" {
		t.Errorf("Expected success message, got '%v'", result["message"])
	}
}

func TestDeletePatient_NotFound(t *testing.T) {
	db := setupTestDB()

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: "999"}}

	DeletePatient(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeletePatient_WithDependencies(t *testing.T) {
	db := setupTestDB()

	patient := createTestPatient(db, "Julia Ferreira", "11922222222")
	user := createTestUser(db, "Dr. Test", "dr@test.com")

	// Create an appointment for the patient using PostgreSQL syntax
	db.Exec(`INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, status, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW() + INTERVAL '1 hour', 'scheduled', NOW(), NOW())`, patient.ID, user.ID)

	c, w := setupTestContext(db)
	c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", patient.ID)}}

	DeletePatient(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}

	result := parseJSONResponse(w)
	if _, ok := result["dependencies"]; !ok {
		t.Error("Expected dependencies in response")
	}
}

func TestCreatePatient_NoDB(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Provide valid patient data so JSON binding succeeds
	body := `{"name":"Test","phone":"11999999999"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	// Don't set "db" in context

	CreatePatient(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
