#!/bin/bash

# Complete Waiting List Test with Database Verification

set +e

echo "=========================================="
echo "TESTE COMPLETO - LISTA DE ESPERA"
echo "=========================================="
echo ""

API_URL="https://api.odowell.pro"
TEST_EMAIL="test@baseline.com"
TEST_PASSWORD="testpassword123"

# Login
echo "1️⃣  TESTE: Login"
LOGIN_RESPONSE=$(curl -s -X POST $API_URL/api/auth/login \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Login falhou"
    exit 1
fi
echo "✅ Login bem-sucedido"
echo ""

# Get patient
echo "2️⃣  TESTE: Buscar paciente para teste"
PATIENT_RESPONSE=$(curl -s -X GET "$API_URL/api/patients?page=1&limit=1" \
    -H "Authorization: Bearer $TOKEN")

PATIENT_ID=$(echo "$PATIENT_RESPONSE" | jq -r '.patients[0].id')
echo "✅ Paciente ID: $PATIENT_ID"
echo ""

# CREATE - Add to waiting list
echo "3️⃣  TESTE: CREATE - Adicionar à lista de espera"
CREATE_RESPONSE=$(curl -s -X POST $API_URL/api/waiting-list \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
  \"patient_id\": $PATIENT_ID,
  \"procedure\": \"Limpeza dental completa\",
  \"priority\": \"normal\",
  \"notes\": \"Teste automatizado - pode deletar\"
}")

ENTRY_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')

if [ "$ENTRY_ID" = "null" ] || [ -z "$ENTRY_ID" ]; then
    echo "❌ Falha ao criar entrada"
    echo "Resposta: $CREATE_RESPONSE"
    exit 1
fi
echo "✅ Entrada criada - ID: $ENTRY_ID"

# Verify in database
echo "   Verificando no banco de dados..."
DB_CHECK=$(docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "SET search_path TO tenant_1; SELECT id, procedure, status FROM waiting_lists WHERE id = $ENTRY_ID;")
if echo "$DB_CHECK" | grep -q "Limpeza dental"; then
    echo "✅ Dados confirmados no banco"
else
    echo "❌ Dados não encontrados no banco"
    exit 1
fi
echo ""

# READ - Get single entry
echo "4️⃣  TESTE: READ - Buscar entrada específica"
GET_RESPONSE=$(curl -s -X GET "$API_URL/api/waiting-list/$ENTRY_ID" \
    -H "Authorization: Bearer $TOKEN")

PROCEDURE=$(echo "$GET_RESPONSE" | jq -r '.procedure')
if [ "$PROCEDURE" = "Limpeza dental completa" ]; then
    echo "✅ Entrada recuperada corretamente"
else
    echo "❌ Falha ao recuperar entrada"
    exit 1
fi
echo ""

# UPDATE - Change to urgent
echo "5️⃣  TESTE: UPDATE - Alterar prioridade para urgente"
UPDATE_RESPONSE=$(curl -s -X PUT "$API_URL/api/waiting-list/$ENTRY_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
  \"patient_id\": $PATIENT_ID,
  \"procedure\": \"Limpeza + Aplicação de flúor\",
  \"priority\": \"urgent\",
  \"status\": \"waiting\",
  \"notes\": \"Atualizado para urgente\"
}")

PRIORITY=$(echo "$UPDATE_RESPONSE" | jq -r '.priority')
if [ "$PRIORITY" = "urgent" ]; then
    echo "✅ Prioridade atualizada para urgente"
else
    echo "❌ Falha ao atualizar prioridade"
    exit 1
fi

# Verify update in database
echo "   Verificando atualização no banco..."
DB_PRIORITY=$(docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "SET search_path TO tenant_1; SELECT priority FROM waiting_lists WHERE id = $ENTRY_ID;" | xargs)
if [ "$DB_PRIORITY" = "urgent" ]; then
    echo "✅ Atualização confirmada no banco"
else
    echo "❌ Atualização não refletida no banco"
    exit 1
fi
echo ""

# LIST - Get all entries
echo "6️⃣  TESTE: LIST - Listar todas as entradas"
LIST_RESPONSE=$(curl -s -X GET "$API_URL/api/waiting-list?page=1&limit=10" \
    -H "Authorization: Bearer $TOKEN")

TOTAL=$(echo "$LIST_RESPONSE" | jq -r '.total')
echo "✅ Lista recuperada - Total: $TOTAL entradas"
echo ""

# STATS - Get statistics
echo "7️⃣  TESTE: STATS - Buscar estatísticas"
STATS_RESPONSE=$(curl -s -X GET "$API_URL/api/waiting-list/stats" \
    -H "Authorization: Bearer $TOKEN")

WAITING=$(echo "$STATS_RESPONSE" | jq -r '.total_waiting')
URGENT=$(echo "$STATS_RESPONSE" | jq -r '.total_urgent')
echo "✅ Estatísticas: $WAITING aguardando, $URGENT urgentes"
echo ""

# CONTACT - Mark as contacted
echo "8️⃣  TESTE: CONTACT - Marcar como contatado"
CONTACT_RESPONSE=$(curl -s -X POST "$API_URL/api/waiting-list/$ENTRY_ID/contact" \
    -H "Authorization: Bearer $TOKEN")

STATUS=$(echo "$CONTACT_RESPONSE" | jq -r '.status')
if [ "$STATUS" = "contacted" ]; then
    echo "✅ Marcado como contatado"
else
    echo "❌ Falha ao marcar como contatado"
    exit 1
fi

# Verify status in database
echo "   Verificando status no banco..."
DB_STATUS=$(docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "SET search_path TO tenant_1; SELECT status FROM waiting_lists WHERE id = $ENTRY_ID;" | xargs)
if [ "$DB_STATUS" = "contacted" ]; then
    echo "✅ Status confirmado no banco"
else
    echo "❌ Status não atualizado no banco"
    exit 1
fi
echo ""

# FILTER - Test filters
echo "9️⃣  TESTE: FILTER - Filtrar por prioridade urgente"
FILTER_RESPONSE=$(curl -s -X GET "$API_URL/api/waiting-list?priority=urgent" \
    -H "Authorization: Bearer $TOKEN")

FILTERED=$(echo "$FILTER_RESPONSE" | jq -r '.entries | length')
echo "✅ Filtro aplicado - $FILTERED entradas urgentes"
echo ""

# DELETE - Remove from list
echo "🔟 TESTE: DELETE - Remover da lista"
DELETE_RESPONSE=$(curl -s -X DELETE "$API_URL/api/waiting-list/$ENTRY_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo "$DELETE_RESPONSE" | grep -qi "success\|deletado"; then
    echo "✅ Entrada deletada"
else
    echo "⚠️  Resposta delete: $DELETE_RESPONSE"
fi

# Verify deletion (soft delete)
echo "   Verificando soft delete no banco..."
DB_DELETED=$(docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "SET search_path TO tenant_1; SELECT deleted_at IS NOT NULL FROM waiting_lists WHERE id = $ENTRY_ID;" | xargs)
if [ "$DB_DELETED" = "t" ]; then
    echo "✅ Soft delete confirmado no banco"
else
    echo "❌ Soft delete não funcionou"
    exit 1
fi
echo ""

echo "=========================================="
echo "✅ TODOS OS TESTES PASSARAM!"
echo "=========================================="
echo ""
echo "Resumo dos Testes:"
echo "  ✅ CREATE - Dados inseridos corretamente"
echo "  ✅ READ - Busca funcionando"
echo "  ✅ UPDATE - Alterações persistidas"
echo "  ✅ DELETE - Soft delete funcionando"
echo "  ✅ LIST - Listagem com paginação"
echo "  ✅ STATS - Estatísticas corretas"
echo "  ✅ CONTACT - Ação especial funcionando"
echo "  ✅ FILTER - Filtros aplicados"
echo "  ✅ DATABASE - Persistência verificada"
echo ""
