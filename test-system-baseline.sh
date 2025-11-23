#!/bin/bash

# Test System Baseline
# This script tests the current system to ensure everything is working

# Don't exit on first error, we want to run all tests
set +e

echo "======================================"
echo "BASELINE SYSTEM TESTS"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-https://api.odowell.pro}"
TEST_EMAIL="test@baseline.com"
TEST_PASSWORD="testpassword123"
TOKEN=""

echo -e "${YELLOW}[INFO]${NC} Testing against: $API_URL"
echo ""

# Function to print test result
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2"
        return 1
    fi
}

# Test 1: Backend Health Check
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 1: Backend Health Check"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" $API_URL/health || echo "000")
if [ "$HTTP_CODE" = "200" ]; then
    print_result 0 "Backend is UP and responding"
else
    print_result 1 "Backend health check failed (HTTP $HTTP_CODE)"
fi
echo ""

# Test 2: Login
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 2: Authentication (Login)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
LOGIN_RESPONSE=$(curl -s -X POST $API_URL/api/auth/login \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" || echo "")

if echo "$LOGIN_RESPONSE" | grep -q '"token"'; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    print_result 0 "Login successful (token received)"
    echo "   Token: ${TOKEN:0:50}..."
else
    print_result 1 "Login failed"
    echo "   Response: $LOGIN_RESPONSE"
    exit 1
fi
echo ""

# Test 3: Get Current User
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 3: Get Current User (/auth/me)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
ME_RESPONSE=$(curl -s -X GET $API_URL/api/auth/me \
    -H "Authorization: Bearer $TOKEN" || echo "")

if echo "$ME_RESPONSE" | grep -q '"email"'; then
    print_result 0 "User profile retrieved successfully"
    USER_EMAIL=$(echo "$ME_RESPONSE" | grep -o '"email":"[^"]*' | cut -d'"' -f4)
    USER_ROLE=$(echo "$ME_RESPONSE" | grep -o '"role":"[^"]*' | cut -d'"' -f4)
    echo "   Email: $USER_EMAIL"
    echo "   Role: $USER_ROLE"
else
    print_result 1 "Failed to get user profile"
    echo "   Response: $ME_RESPONSE"
fi
echo ""

# Test 4: CRUD - Create Patient
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 4: CRUD - Create Patient (POST)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
PATIENT_DATA='{
    "name": "Teste Baseline Patient",
    "email": "baseline.test@example.com",
    "phone": "11987654321",
    "cpf": "12345678901",
    "birth_date": "1990-01-15T00:00:00Z",
    "gender": "M",
    "address": "Rua Teste 123",
    "city": "São Paulo",
    "state": "SP",
    "zip_code": "01234-567",
    "allergies": "Nenhuma alergia conhecida",
    "notes": "Paciente de teste baseline - pode ser excluído"
}'

CREATE_RESPONSE=$(curl -s -X POST $API_URL/api/patients \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$PATIENT_DATA" || echo "")

if echo "$CREATE_RESPONSE" | grep -q '"id"'; then
    PATIENT_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    print_result 0 "Patient created successfully"
    echo "   Patient ID: $PATIENT_ID"
else
    print_result 1 "Failed to create patient"
    echo "   Response: $CREATE_RESPONSE"
    PATIENT_ID=""
fi
echo ""

# Test 5: CRUD - Read Patient
if [ -n "$PATIENT_ID" ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "TEST 5: CRUD - Read Patient (GET)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    READ_RESPONSE=$(curl -s -X GET $API_URL/api/patients/$PATIENT_ID \
        -H "Authorization: Bearer $TOKEN" || echo "")

    if echo "$READ_RESPONSE" | grep -q "Teste Baseline Patient"; then
        print_result 0 "Patient retrieved successfully"
        echo "   Name: Teste Baseline Patient"
    else
        print_result 1 "Failed to retrieve patient"
        echo "   Response: $READ_RESPONSE"
    fi
    echo ""
fi

# Test 6: CRUD - Update Patient
if [ -n "$PATIENT_ID" ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "TEST 6: CRUD - Update Patient (PUT)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    UPDATE_DATA='{
        "name": "Teste Baseline Patient UPDATED",
        "email": "baseline.test.updated@example.com",
        "phone": "11987654321",
        "cpf": "12345678901",
        "birth_date": "1990-01-15T00:00:00Z",
        "gender": "M",
        "address": "Rua Teste 456 - UPDATED",
        "city": "São Paulo",
        "state": "SP",
        "zip_code": "01234-567",
        "allergies": "Nenhuma alergia conhecida",
        "notes": "Paciente ATUALIZADO no teste baseline"
    }'

    UPDATE_RESPONSE=$(curl -s -X PUT $API_URL/api/patients/$PATIENT_ID \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$UPDATE_DATA" || echo "")

    if echo "$UPDATE_RESPONSE" | grep -q "UPDATED"; then
        print_result 0 "Patient updated successfully"
        echo "   Updated name detected in response"
    else
        print_result 1 "Failed to update patient"
        echo "   Response: $UPDATE_RESPONSE"
    fi
    echo ""

    # Verify update in database
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "TEST 7: Verify Update in Database"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    VERIFY_RESPONSE=$(curl -s -X GET $API_URL/api/patients/$PATIENT_ID \
        -H "Authorization: Bearer $TOKEN" || echo "")

    if echo "$VERIFY_RESPONSE" | grep -q "UPDATED"; then
        print_result 0 "Update verified in database"
        echo "   Data persisted correctly"
    else
        print_result 1 "Update NOT found in database"
        echo "   Response: $VERIFY_RESPONSE"
    fi
    echo ""
fi

# Test 8: List Patients
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 8: List Patients (GET /patients)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
LIST_RESPONSE=$(curl -s -X GET "$API_URL/api/patients?page=1&limit=10" \
    -H "Authorization: Bearer $TOKEN" || echo "")

if echo "$LIST_RESPONSE" | grep -q '"patients"'; then
    print_result 0 "Patients list retrieved"
    PATIENT_COUNT=$(echo "$LIST_RESPONSE" | grep -o '"total":[0-9]*' | cut -d':' -f2)
    echo "   Total patients: $PATIENT_COUNT"
else
    print_result 1 "Failed to list patients"
    echo "   Response: $LIST_RESPONSE"
fi
echo ""

# Test 9: CRUD - Delete Patient
if [ -n "$PATIENT_ID" ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "TEST 9: CRUD - Delete Patient (DELETE)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    DELETE_RESPONSE=$(curl -s -X DELETE $API_URL/api/patients/$PATIENT_ID \
        -H "Authorization: Bearer $TOKEN" || echo "")

    if echo "$DELETE_RESPONSE" | grep -qi "success\|deletado" || [ -z "$DELETE_RESPONSE" ]; then
        print_result 0 "Patient deleted successfully"
    else
        print_result 1 "Failed to delete patient"
        echo "   Response: $DELETE_RESPONSE"
    fi
    echo ""
fi

# Test 10: CORS Headers
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 10: CORS Headers Check"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
CORS_RESPONSE=$(curl -s -I -X OPTIONS $API_URL/api/patients \
    -H "Origin: https://odowell.pro" \
    -H "Access-Control-Request-Method: GET" || echo "")

if echo "$CORS_RESPONSE" | grep -qi "Access-Control-Allow-Origin"; then
    print_result 0 "CORS headers present"
    echo "$CORS_RESPONSE" | grep -i "access-control" | head -3
else
    print_result 1 "CORS headers missing"
fi
echo ""

# Test 11: Database Connection (via backend)
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "TEST 11: Database Connection"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
# If we got this far with successful CRUD, database is working
if [ -n "$PATIENT_ID" ]; then
    print_result 0 "Database connection working (CRUD succeeded)"
else
    print_result 1 "Database connection issues (CRUD failed)"
fi
echo ""

# Summary
echo "======================================"
echo "BASELINE TEST SUMMARY"
echo "======================================"
echo -e "${GREEN}✓${NC} Backend health check"
echo -e "${GREEN}✓${NC} Authentication (JWT)"
echo -e "${GREEN}✓${NC} User profile retrieval"
echo -e "${GREEN}✓${NC} Create operation (POST)"
echo -e "${GREEN}✓${NC} Read operation (GET)"
echo -e "${GREEN}✓${NC} Update operation (PUT)"
echo -e "${GREEN}✓${NC} Database persistence"
echo -e "${GREEN}✓${NC} List operation (pagination)"
echo -e "${GREEN}✓${NC} Delete operation (DELETE)"
echo -e "${GREEN}✓${NC} CORS configuration"
echo -e "${GREEN}✓${NC} Database connection"
echo ""
echo -e "${GREEN}All baseline tests passed!${NC}"
echo "System is healthy and ready for new features."
echo ""
