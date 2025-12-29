#!/bin/bash

# =============================================================================
# Script de Teste de Produção - Odowell
# =============================================================================
# Uso: ./test_production.sh [email] [senha]
# Exemplo: ./test_production.sh admin@clinica.com minhasenha123
# =============================================================================

API_URL="https://api.odowell.pro"
FRONTEND_URL="https://app.odowell.pro"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Contadores
PASSED=0
FAILED=0
WARNINGS=0

# Funções auxiliares
pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((PASSED++))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((FAILED++))
}

warn() {
    echo -e "${YELLOW}⚠ WARN${NC}: $1"
    ((WARNINGS++))
}

# =============================================================================
# 1. TESTES DE INFRAESTRUTURA
# =============================================================================
echo ""
echo "=============================================="
echo "1. TESTES DE INFRAESTRUTURA"
echo "=============================================="

# Health check
HEALTH=$(curl -s $API_URL/health)
if echo "$HEALTH" | grep -q '"status":"ok"'; then
    pass "Backend health check"
else
    fail "Backend health check"
fi

# Frontend
FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" $FRONTEND_URL)
if [ "$FRONTEND_STATUS" = "200" ]; then
    pass "Frontend acessível (HTTP $FRONTEND_STATUS)"
else
    fail "Frontend inacessível (HTTP $FRONTEND_STATUS)"
fi

# SSL Backend (verifica se usa HTTPS e retorna resposta válida)
SSL_CHECK=$(curl -s -o /dev/null -w "%{http_code}:%{ssl_verify_result}" $API_URL/health 2>/dev/null)
HTTP_CODE=$(echo $SSL_CHECK | cut -d: -f1)
SSL_RESULT=$(echo $SSL_CHECK | cut -d: -f2)
if [ "$HTTP_CODE" = "200" ] && [ "$SSL_RESULT" = "0" ]; then
    pass "SSL Backend válido"
else
    fail "SSL Backend com problema (HTTP: $HTTP_CODE, SSL: $SSL_RESULT)"
fi

# Redis
if echo "$HEALTH" | grep -q '"redis":"healthy"'; then
    pass "Redis funcionando"
else
    fail "Redis com problema"
fi

# PostgreSQL
if echo "$HEALTH" | grep -q '"postgres":"healthy"'; then
    pass "PostgreSQL funcionando"
else
    fail "PostgreSQL com problema"
fi

# =============================================================================
# 2. TESTES DE AUTENTICAÇÃO
# =============================================================================
echo ""
echo "=============================================="
echo "2. TESTES DE AUTENTICAÇÃO"
echo "=============================================="

EMAIL=${1:-""}
PASSWORD=${2:-""}

if [ -z "$EMAIL" ] || [ -z "$PASSWORD" ]; then
    warn "Credenciais não fornecidas. Pulando testes de autenticação."
    warn "Uso: $0 email senha"
else
    # Login
    LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

    TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // empty')

    if [ -n "$TOKEN" ]; then
        pass "Login bem-sucedido"

        # Teste de acesso com token
        ME_RESPONSE=$(curl -s "$API_URL/api/auth/me" -H "Authorization: Bearer $TOKEN")
        if echo "$ME_RESPONSE" | grep -q '"id"'; then
            pass "Token válido - /api/auth/me"
        else
            fail "Token inválido ou endpoint /api/auth/me com problema"
        fi

        # =============================================================================
        # 3. TESTES DE CRUD - PACIENTES
        # =============================================================================
        echo ""
        echo "=============================================="
        echo "3. TESTES DE CRUD - PACIENTES"
        echo "=============================================="

        # Listar pacientes
        PATIENTS=$(curl -s "$API_URL/api/patients" -H "Authorization: Bearer $TOKEN")
        if echo "$PATIENTS" | grep -q '"patients"'; then
            TOTAL=$(echo "$PATIENTS" | jq '.total // 0')
            pass "Listar pacientes (Total: $TOTAL)"
        else
            fail "Listar pacientes"
        fi

        # Criar paciente de teste
        TEST_PATIENT=$(curl -s -X POST "$API_URL/api/patients" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d '{"name": "TESTE_AUTOMATIZADO_DELETAR", "cell_phone": "11999990000"}')

        PATIENT_ID=$(echo "$TEST_PATIENT" | jq -r '.patient.id // empty')
        if [ -n "$PATIENT_ID" ]; then
            pass "Criar paciente (ID: $PATIENT_ID)"

            # Buscar paciente
            GET_PATIENT=$(curl -s "$API_URL/api/patients/$PATIENT_ID" -H "Authorization: Bearer $TOKEN")
            if echo "$GET_PATIENT" | grep -q "TESTE_AUTOMATIZADO"; then
                pass "Buscar paciente por ID"
            else
                fail "Buscar paciente por ID"
            fi

            # Atualizar paciente
            UPDATE=$(curl -s -X PUT "$API_URL/api/patients/$PATIENT_ID" \
                -H "Authorization: Bearer $TOKEN" \
                -H "Content-Type: application/json" \
                -d '{"name": "TESTE_AUTOMATIZADO_ATUALIZADO", "cell_phone": "11999990000"}')
            if echo "$UPDATE" | grep -q "ATUALIZADO"; then
                pass "Atualizar paciente"
            else
                fail "Atualizar paciente"
            fi

            # Deletar paciente
            DELETE=$(curl -s -X DELETE "$API_URL/api/patients/$PATIENT_ID" \
                -H "Authorization: Bearer $TOKEN")
            if echo "$DELETE" | grep -qi "success\|deleted\|ok" || [ "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$API_URL/api/patients/$PATIENT_ID" -H "Authorization: Bearer $TOKEN")" = "200" ]; then
                pass "Deletar paciente"
            else
                fail "Deletar paciente"
            fi
        else
            fail "Criar paciente"
        fi

        # =============================================================================
        # 4. TESTES DE CRUD - AGENDAMENTOS
        # =============================================================================
        echo ""
        echo "=============================================="
        echo "4. TESTES DE CRUD - AGENDAMENTOS"
        echo "=============================================="

        # Listar agendamentos
        APPOINTMENTS=$(curl -s "$API_URL/api/appointments" -H "Authorization: Bearer $TOKEN")
        if echo "$APPOINTMENTS" | grep -qE '"appointments"|"data"|\[\]'; then
            pass "Listar agendamentos"
        else
            fail "Listar agendamentos"
        fi

        # =============================================================================
        # 5. TESTES DE CRUD - USUÁRIOS
        # =============================================================================
        echo ""
        echo "=============================================="
        echo "5. TESTES DE CRUD - USUÁRIOS"
        echo "=============================================="

        # Listar usuários
        USERS=$(curl -s "$API_URL/api/users" -H "Authorization: Bearer $TOKEN")
        if echo "$USERS" | grep -q '"users"'; then
            TOTAL_USERS=$(echo "$USERS" | jq '.total // 0')
            pass "Listar usuários (Total: $TOTAL_USERS)"
        else
            fail "Listar usuários"
        fi

        # =============================================================================
        # 6. TESTES DE MÓDULOS FINANCEIROS
        # =============================================================================
        echo ""
        echo "=============================================="
        echo "6. TESTES DE MÓDULOS FINANCEIROS"
        echo "=============================================="

        # Orçamentos
        BUDGETS=$(curl -s "$API_URL/api/budgets" -H "Authorization: Bearer $TOKEN")
        if echo "$BUDGETS" | grep -qE '"budgets"|"data"|\[\]'; then
            pass "Listar orçamentos"
        else
            fail "Listar orçamentos"
        fi

        # Pagamentos
        PAYMENTS=$(curl -s "$API_URL/api/payments" -H "Authorization: Bearer $TOKEN")
        if echo "$PAYMENTS" | grep -qE '"payments"|"data"|\[\]'; then
            pass "Listar pagamentos"
        else
            fail "Listar pagamentos"
        fi

        # Despesas
        EXPENSES=$(curl -s "$API_URL/api/expenses" -H "Authorization: Bearer $TOKEN")
        if echo "$EXPENSES" | grep -qE '"expenses"|"data"|\[\]'; then
            pass "Listar despesas"
        else
            fail "Listar despesas"
        fi

        # =============================================================================
        # 7. TESTES DE MÓDULOS CLÍNICOS
        # =============================================================================
        echo ""
        echo "=============================================="
        echo "7. TESTES DE MÓDULOS CLÍNICOS"
        echo "=============================================="

        # Prontuários
        RECORDS=$(curl -s "$API_URL/api/medical-records" -H "Authorization: Bearer $TOKEN")
        if echo "$RECORDS" | grep -qE '"records"|"data"|\[\]'; then
            pass "Listar prontuários"
        else
            fail "Listar prontuários"
        fi

        # Receitas
        PRESCRIPTIONS=$(curl -s "$API_URL/api/prescriptions" -H "Authorization: Bearer $TOKEN")
        if echo "$PRESCRIPTIONS" | grep -qE '"prescriptions"|"data"|\[\]'; then
            pass "Listar receitas"
        else
            fail "Listar receitas"
        fi

        # Exames
        EXAMS=$(curl -s "$API_URL/api/exams" -H "Authorization: Bearer $TOKEN")
        if echo "$EXAMS" | grep -qE '"exams"|"data"|\[\]'; then
            pass "Listar exames"
        else
            fail "Listar exames"
        fi

    else
        fail "Login falhou: $(echo "$LOGIN_RESPONSE" | jq -r '.error // .message // "Erro desconhecido"')"
        warn "Pulando demais testes que requerem autenticação"
    fi
fi

# =============================================================================
# 8. TESTES DE API WHATSAPP (Usa API Key)
# =============================================================================
echo ""
echo "=============================================="
echo "8. TESTES DE API WHATSAPP"
echo "=============================================="

API_KEY="99bcd1d380511d6bbd5b3fd6d6b0c66890a71aef253a5460c5aacd507cdefe42"

# Listar procedimentos
PROCS=$(curl -s "$API_URL/api/whatsapp/procedures" -H "X-API-Key: $API_KEY")
if echo "$PROCS" | grep -q '"procedures"'; then
    pass "WhatsApp API - Listar procedimentos"
else
    fail "WhatsApp API - Listar procedimentos"
fi

# Listar dentistas
DENTISTS=$(curl -s "$API_URL/api/whatsapp/dentists" -H "X-API-Key: $API_KEY")
if echo "$DENTISTS" | grep -q '"dentists"'; then
    pass "WhatsApp API - Listar dentistas"
else
    fail "WhatsApp API - Listar dentistas"
fi

# =============================================================================
# RESUMO
# =============================================================================
echo ""
echo "=============================================="
echo "RESUMO DOS TESTES"
echo "=============================================="
echo -e "${GREEN}Passou: $PASSED${NC}"
echo -e "${RED}Falhou: $FAILED${NC}"
echo -e "${YELLOW}Avisos: $WARNINGS${NC}"
echo ""

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}⚠ ATENÇÃO: Existem falhas que precisam ser corrigidas antes do lançamento!${NC}"
    exit 1
else
    echo -e "${GREEN}✓ Todos os testes passaram!${NC}"
    exit 0
fi
