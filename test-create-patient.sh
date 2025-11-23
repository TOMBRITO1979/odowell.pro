#!/bin/bash

# Get token
TOKEN=$(curl -s -X POST https://api.odowell.pro/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@baseline.com","password":"testpassword123"}' | jq -r '.token')

echo "Token: ${TOKEN:0:50}..."
echo ""

# Create patient
echo "Creating patient..."
curl -s -X POST https://api.odowell.pro/api/patients \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
    "name": "Teste Baseline Patient",
    "email": "baseline.test@example.com",
    "phone": "11987654321",
    "cpf": "12345678901",
    "birth_date": "1990-01-15",
    "gender": "M",
    "address": "Rua Teste 123",
    "city": "São Paulo",
    "state": "SP",
    "zip_code": "01234-567",
    "allergies": "Nenhuma alergia conhecida",
    "notes": "Paciente de teste baseline - pode ser excluído"
}' | jq .
