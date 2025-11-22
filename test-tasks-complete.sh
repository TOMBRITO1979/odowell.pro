#!/bin/bash

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

echo "========================================"
echo "  SUITE COMPLETA DE TESTES - TAREFAS  "
echo "========================================"
echo ""

# Aguardar backend iniciar
echo "Aguardando backend iniciar..."
sleep 12

# Obter token
echo "1. Fazendo login..."
LOGIN_RESPONSE=$(curl -X POST https://drapi.crwell.pro/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@teste.com","password":"admin123"}' \
  -s)

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token' 2>/dev/null)

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${RED}✗ FALHA: Não foi possível obter token${NC}"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo -e "${GREEN}✓ Login bem sucedido${NC}"
echo ""

# TESTE 1: Criar tarefa simples (sem responsáveis, sem assignments)
echo "========================================"
echo "TESTE 1: Criar tarefa simples"
echo "========================================"
T1_RESPONSE=$(curl -X POST https://drapi.crwell.pro/api/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Teste 1 - Simples","description":"Tarefa sem responsáveis","priority":"medium","status":"pending"}' \
  -s -w "\nHTTP_CODE:%{http_code}")

T1_CODE=$(echo "$T1_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
T1_BODY=$(echo "$T1_RESPONSE" | sed '/HTTP_CODE/d')

if [ "$T1_CODE" == "201" ]; then
  T1_ID=$(echo "$T1_BODY" | jq -r '.task.id')
  echo -e "${GREEN}✓ TESTE 1 PASSOU - Tarefa criada (ID: $T1_ID)${NC}"
  PASSED=$((PASSED+1))
else
  echo -e "${RED}✗ TESTE 1 FALHOU - Status: $T1_CODE${NC}"
  echo "Response: $T1_BODY"
  FAILED=$((FAILED+1))
fi
echo ""

# TESTE 2: Criar tarefa com 1 responsável
echo "========================================"
echo "TESTE 2: Criar tarefa com 1 responsável"
echo "========================================"
T2_RESPONSE=$(curl -X POST https://drapi.crwell.pro/api/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Teste 2 - Com Responsável","description":"Tarefa com 1 responsável","priority":"high","status":"pending","responsible_ids":[4]}' \
  -s -w "\nHTTP_CODE:%{http_code}")

T2_CODE=$(echo "$T2_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
T2_BODY=$(echo "$T2_RESPONSE" | sed '/HTTP_CODE/d')

if [ "$T2_CODE" == "201" ]; then
  T2_ID=$(echo "$T2_BODY" | jq -r '.task.id')
  echo -e "${GREEN}✓ TESTE 2 PASSOU - Tarefa criada (ID: $T2_ID)${NC}"
  PASSED=$((PASSED+1))
  
  # Verificar no banco se o responsável foi salvo
  RESP_COUNT=$(docker exec -i $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "SELECT COUNT(*) FROM tenant_1.task_users WHERE task_id = $T2_ID AND deleted_at IS NULL;" | tr -d ' ')
  if [ "$RESP_COUNT" == "1" ]; then
    echo -e "${GREEN}  ✓ Responsável salvo no banco${NC}"
  else
    echo -e "${RED}  ✗ Responsável NÃO foi salvo (count: $RESP_COUNT)${NC}"
    FAILED=$((FAILED+1))
  fi
else
  echo -e "${RED}✗ TESTE 2 FALHOU - Status: $T2_CODE${NC}"
  echo "Response: $T2_BODY"
  FAILED=$((FAILED+1))
fi
echo ""

# TESTE 3: Criar tarefa com 2 responsáveis
echo "========================================"
echo "TESTE 3: Criar tarefa com 2 responsáveis"
echo "========================================"
T3_RESPONSE=$(curl -X POST https://drapi.crwell.pro/api/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Teste 3 - Dois Responsáveis","description":"Tarefa com 2 responsáveis","priority":"urgent","status":"pending","responsible_ids":[4,7]}' \
  -s -w "\nHTTP_CODE:%{http_code}")

T3_CODE=$(echo "$T3_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
T3_BODY=$(echo "$T3_RESPONSE" | sed '/HTTP_CODE/d')

if [ "$T3_CODE" == "201" ]; then
  T3_ID=$(echo "$T3_BODY" | jq -r '.task.id')
  echo -e "${GREEN}✓ TESTE 3 PASSOU - Tarefa criada (ID: $T3_ID)${NC}"
  PASSED=$((PASSED+1))
  
  # Verificar no banco
  RESP_COUNT=$(docker exec -i $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "SELECT COUNT(*) FROM tenant_1.task_users WHERE task_id = $T3_ID AND deleted_at IS NULL;" | tr -d ' ')
  if [ "$RESP_COUNT" == "2" ]; then
    echo -e "${GREEN}  ✓ 2 responsáveis salvos no banco${NC}"
  else
    echo -e "${RED}  ✗ Esperava 2 responsáveis, encontrou: $RESP_COUNT${NC}"
    FAILED=$((FAILED+1))
  fi
else
  echo -e "${RED}✗ TESTE 3 FALHOU - Status: $T3_CODE${NC}"
  echo "Response: $T3_BODY"
  FAILED=$((FAILED+1))
fi
echo ""

# TESTE 4: Criar tarefa com assignment (paciente)
echo "========================================"
echo "TESTE 4: Criar tarefa com assignment"
echo "========================================"
T4_RESPONSE=$(curl -X POST https://drapi.crwell.pro/api/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Teste 4 - Com Assignment","description":"Tarefa linkada a paciente","priority":"high","status":"pending","responsible_ids":[4],"assignments":[{"assignable_type":"patient","assignable_id":4}]}' \
  -s -w "\nHTTP_CODE:%{http_code}")

T4_CODE=$(echo "$T4_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
T4_BODY=$(echo "$T4_RESPONSE" | sed '/HTTP_CODE/d')

if [ "$T4_CODE" == "201" ]; then
  T4_ID=$(echo "$T4_BODY" | jq -r '.task.id')
  echo -e "${GREEN}✓ TESTE 4 PASSOU - Tarefa criada (ID: $T4_ID)${NC}"
  PASSED=$((PASSED+1))
  
  # Verificar assignment no banco
  ASSIGN_COUNT=$(docker exec -i $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "SELECT COUNT(*) FROM tenant_1.task_assignments WHERE task_id = $T4_ID AND deleted_at IS NULL;" | tr -d ' ')
  if [ "$ASSIGN_COUNT" == "1" ]; then
    echo -e "${GREEN}  ✓ Assignment salvo no banco${NC}"
  else
    echo -e "${RED}  ✗ Assignment NÃO foi salvo (count: $ASSIGN_COUNT)${NC}"
    FAILED=$((FAILED+1))
  fi
else
  echo -e "${RED}✗ TESTE 4 FALHOU - Status: $T4_CODE${NC}"
  echo "Response: $T4_BODY"
  FAILED=$((FAILED+1))
fi
echo ""

# TESTE 5: Listar tarefas
echo "========================================"
echo "TESTE 5: Listar tarefas"
echo "========================================"
T5_RESPONSE=$(curl -X GET "https://drapi.crwell.pro/api/tasks?page=1&page_size=20" \
  -H "Authorization: Bearer $TOKEN" \
  -s -w "\nHTTP_CODE:%{http_code}")

T5_CODE=$(echo "$T5_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
T5_BODY=$(echo "$T5_RESPONSE" | sed '/HTTP_CODE/d')

if [ "$T5_CODE" == "200" ]; then
  TOTAL=$(echo "$T5_BODY" | jq -r '.total')
  echo -e "${GREEN}✓ TESTE 5 PASSOU - Listagem OK (Total: $TOTAL tarefas)${NC}"
  PASSED=$((PASSED+1))
else
  echo -e "${RED}✗ TESTE 5 FALHOU - Status: $T5_CODE${NC}"
  FAILED=$((FAILED+1))
fi
echo ""

# TESTE 6: Buscar tarefa específica (se T2 foi criada)
if [ ! -z "$T2_ID" ]; then
  echo "========================================"
  echo "TESTE 6: Buscar tarefa específica"
  echo "========================================"
  T6_RESPONSE=$(curl -X GET "https://drapi.crwell.pro/api/tasks/$T2_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -s -w "\nHTTP_CODE:%{http_code}")

  T6_CODE=$(echo "$T6_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
  T6_BODY=$(echo "$T6_RESPONSE" | sed '/HTTP_CODE/d')

  if [ "$T6_CODE" == "200" ]; then
    echo -e "${GREEN}✓ TESTE 6 PASSOU - Tarefa encontrada${NC}"
    PASSED=$((PASSED+1))
  else
    echo -e "${RED}✗ TESTE 6 FALHOU - Status: $T6_CODE${NC}"
    FAILED=$((FAILED+1))
  fi
  echo ""
fi

# TESTE 7: Atualizar tarefa (se T2 foi criada)
if [ ! -z "$T2_ID" ]; then
  echo "========================================"
  echo "TESTE 7: Atualizar tarefa"
  echo "========================================"
  T7_RESPONSE=$(curl -X PUT "https://drapi.crwell.pro/api/tasks/$T2_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"Teste 2 - ATUALIZADO","description":"Descrição atualizada","priority":"low","status":"in_progress","responsible_ids":[4,6]}' \
    -s -w "\nHTTP_CODE:%{http_code}")

  T7_CODE=$(echo "$T7_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
  T7_BODY=$(echo "$T7_RESPONSE" | sed '/HTTP_CODE/d')

  if [ "$T7_CODE" == "200" ]; then
    echo -e "${GREEN}✓ TESTE 7 PASSOU - Tarefa atualizada${NC}"
    PASSED=$((PASSED+1))
    
    # Verificar se os responsáveis foram atualizados (agora deve ter 2: IDs 4 e 6)
    RESP_COUNT=$(docker exec -i $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "SELECT COUNT(*) FROM tenant_1.task_users WHERE task_id = $T2_ID AND deleted_at IS NULL;" | tr -d ' ')
    if [ "$RESP_COUNT" == "2" ]; then
      echo -e "${GREEN}  ✓ Responsáveis atualizados (agora tem 2)${NC}"
    else
      echo -e "${RED}  ✗ Esperava 2 responsáveis após update, encontrou: $RESP_COUNT${NC}"
      FAILED=$((FAILED+1))
    fi
  else
    echo -e "${RED}✗ TESTE 7 FALHOU - Status: $T7_CODE${NC}"
    echo "Response: $T7_BODY"
    FAILED=$((FAILED+1))
  fi
  echo ""
fi

# TESTE 8: Pending count
echo "========================================"
echo "TESTE 8: Contagem de tarefas pendentes"
echo "========================================"
T8_RESPONSE=$(curl -X GET "https://drapi.crwell.pro/api/tasks/pending-count" \
  -H "Authorization: Bearer $TOKEN" \
  -s -w "\nHTTP_CODE:%{http_code}")

T8_CODE=$(echo "$T8_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
T8_BODY=$(echo "$T8_RESPONSE" | sed '/HTTP_CODE/d')

if [ "$T8_CODE" == "200" ]; then
  PENDING_COUNT=$(echo "$T8_BODY" | jq -r '.count')
  echo -e "${GREEN}✓ TESTE 8 PASSOU - Pending count: $PENDING_COUNT${NC}"
  PASSED=$((PASSED+1))
else
  echo -e "${RED}✗ TESTE 8 FALHOU - Status: $T8_CODE${NC}"
  FAILED=$((FAILED+1))
fi
echo ""

# TESTE 9: Deletar tarefa (se T4 foi criada)
if [ ! -z "$T4_ID" ]; then
  echo "========================================"
  echo "TESTE 9: Deletar tarefa"
  echo "========================================"
  T9_RESPONSE=$(curl -X DELETE "https://drapi.crwell.pro/api/tasks/$T4_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -s -w "\nHTTP_CODE:%{http_code}")

  T9_CODE=$(echo "$T9_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)

  if [ "$T9_CODE" == "200" ]; then
    echo -e "${GREEN}✓ TESTE 9 PASSOU - Tarefa deletada${NC}"
    PASSED=$((PASSED+1))
    
    # Verificar soft delete no banco
    DELETED=$(docker exec -i $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -t -c "SELECT deleted_at IS NOT NULL FROM tenant_1.tasks WHERE id = $T4_ID;" | tr -d ' ')
    if [ "$DELETED" == "t" ]; then
      echo -e "${GREEN}  ✓ Soft delete aplicado corretamente${NC}"
    else
      echo -e "${RED}  ✗ Soft delete NÃO funcionou${NC}"
      FAILED=$((FAILED+1))
    fi
  else
    echo -e "${RED}✗ TESTE 9 FALHOU - Status: $T9_CODE${NC}"
    FAILED=$((FAILED+1))
  fi
  echo ""
fi

# RESUMO FINAL
echo "========================================"
echo "           RESUMO DOS TESTES           "
echo "========================================"
echo -e "${GREEN}Testes Passados: $PASSED${NC}"
echo -e "${RED}Testes Falhados: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
  echo -e "${GREEN}========================================${NC}"
  echo -e "${GREEN}  ✓ TODOS OS TESTES PASSARAM! 🎉      ${NC}"
  echo -e "${GREEN}========================================${NC}"
  exit 0
else
  echo -e "${RED}========================================${NC}"
  echo -e "${RED}  ✗ ALGUNS TESTES FALHARAM            ${NC}"
  echo -e "${RED}========================================${NC}"
  exit 1
fi
