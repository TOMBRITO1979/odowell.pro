#!/bin/bash

# Login e obter token
echo "=== Fazendo login ==="
LOGIN_RESPONSE=$(curl -s -X POST https://api.odowell.pro/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"wasolutionscorp@gmail.com","password":"Senha123"}')

TOKEN=$(echo $LOGIN_RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['token'])")
echo "Token obtido: ${TOKEN:0:30}..."

# Criar paciente
echo ""
echo "=== Criando paciente ==="
PATIENT_RESPONSE=$(curl -s -X POST https://api.odowell.pro/api/patients \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Paciente Teste API",
    "email": "teste@api.com",
    "active": true,
    "cpf": "11122233344"
  }')

echo "Resposta da API:"
echo $PATIENT_RESPONSE | python3 -m json.tool

# Extrair ID
PATIENT_ID=$(echo $PATIENT_RESPONSE | python3 -c "import sys, json; data=json.load(sys.stdin); print(data.get('id', 'ERRO'))")
echo ""
echo "ID do paciente criado: $PATIENT_ID"
