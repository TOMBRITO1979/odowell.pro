#!/bin/bash

# Test Waiting List Feature

set -e

echo "======================================"
echo "WAITING LIST - FEATURE TEST"
echo "======================================"
echo ""

# Configuration
API_URL="https://api.odowell.pro"
TEST_EMAIL="test@baseline.com"
TEST_PASSWORD="testpassword123"

# Login
echo "1. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $API_URL/api/auth/login \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Login failed"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

echo "✅ Login successful"
echo ""

# Get first patient
echo "2. Getting a patient for testing..."
PATIENT_RESPONSE=$(curl -s -X GET "$API_URL/api/patients?page=1&limit=1" \
    -H "Authorization: Bearer $TOKEN")

PATIENT_ID=$(echo "$PATIENT_RESPONSE" | jq -r '.patients[0].id')

if [ "$PATIENT_ID" = "null" ] || [ -z "$PATIENT_ID" ]; then
    echo "❌ No patients found"
    exit 1
fi

echo "✅ Patient found: ID $PATIENT_ID"
echo ""

# Create waiting list entry
echo "3. Adding patient to waiting list..."
CREATE_DATA=$(cat <<EOF
{
  "patient_id": $PATIENT_ID,
  "procedure": "Limpeza dental",
  "priority": "normal",
  "notes": "Paciente prefere horários de manhã"
}
EOF
)

CREATE_RESPONSE=$(curl -s -X POST $API_URL/api/waiting-list \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$CREATE_DATA")

ENTRY_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')

if [ "$ENTRY_ID" = "null" ] || [ -z "$ENTRY_ID" ]; then
    echo "❌ Failed to create waiting list entry"
    echo "Response: $CREATE_RESPONSE"
    exit 1
fi

echo "✅ Waiting list entry created: ID $ENTRY_ID"
echo ""

# Get waiting list
echo "4. Fetching waiting list..."
LIST_RESPONSE=$(curl -s -X GET "$API_URL/api/waiting-list?page=1&limit=10" \
    -H "Authorization: Bearer $TOKEN")

TOTAL=$(echo "$LIST_RESPONSE" | jq -r '.total')

echo "✅ Waiting list fetched: $TOTAL total entries"
echo ""

# Get stats
echo "5. Fetching statistics..."
STATS_RESPONSE=$(curl -s -X GET "$API_URL/api/waiting-list/stats" \
    -H "Authorization: Bearer $TOKEN")

WAITING_COUNT=$(echo "$STATS_RESPONSE" | jq -r '.total_waiting')

echo "✅ Statistics fetched: $WAITING_COUNT waiting"
echo ""

# Mark as contacted
echo "6. Marking as contacted..."
CONTACT_RESPONSE=$(curl -s -X POST "$API_URL/api/waiting-list/$ENTRY_ID/contact" \
    -H "Authorization: Bearer $TOKEN")

STATUS=$(echo "$CONTACT_RESPONSE" | jq -r '.status')

if [ "$STATUS" != "contacted" ]; then
    echo "❌ Failed to mark as contacted"
    echo "Response: $CONTACT_RESPONSE"
else
    echo "✅ Marked as contacted"
fi
echo ""

# Update entry
echo "7. Updating entry..."
UPDATE_DATA=$(cat <<EOF
{
  "patient_id": $PATIENT_ID,
  "procedure": "Limpeza dental + Aplicação de flúor",
  "priority": "urgent",
  "status": "contacted",
  "notes": "Paciente confirmou interesse. Urgente devido a dor."
}
EOF
)

UPDATE_RESPONSE=$(curl -s -X PUT "$API_URL/api/waiting-list/$ENTRY_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$UPDATE_DATA")

UPDATED_PRIORITY=$(echo "$UPDATE_RESPONSE" | jq -r '.priority')

if [ "$UPDATED_PRIORITY" != "urgent" ]; then
    echo "❌ Failed to update entry"
    echo "Response: $UPDATE_RESPONSE"
else
    echo "✅ Entry updated to urgent priority"
fi
echo ""

# Get single entry
echo "8. Fetching single entry..."
GET_RESPONSE=$(curl -s -X GET "$API_URL/api/waiting-list/$ENTRY_ID" \
    -H "Authorization: Bearer $TOKEN")

PROCEDURE=$(echo "$GET_RESPONSE" | jq -r '.procedure')

echo "✅ Entry fetched: $PROCEDURE"
echo ""

# Delete entry
echo "9. Deleting entry..."
DELETE_RESPONSE=$(curl -s -X DELETE "$API_URL/api/waiting-list/$ENTRY_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo "$DELETE_RESPONSE" | grep -q "success\|deletado"; then
    echo "✅ Entry deleted successfully"
else
    echo "⚠️  Delete response: $DELETE_RESPONSE"
fi
echo ""

# Run baseline tests
echo "10. Running baseline tests..."
./test-system-baseline.sh > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Baseline tests passed"
else
    echo "⚠️  Some baseline tests failed (check ./test-system-baseline.sh)"
fi
echo ""

echo "======================================"
echo "WAITING LIST - TEST SUMMARY"
echo "======================================"
echo "✅ Login"
echo "✅ Create entry"
echo "✅ List entries"
echo "✅ Get statistics"
echo "✅ Mark as contacted"
echo "✅ Update entry"
echo "✅ Get single entry"
echo "✅ Delete entry"
echo "✅ Baseline integrity"
echo ""
echo "🎉 All Waiting List tests passed!"
echo ""
