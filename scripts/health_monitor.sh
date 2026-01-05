#!/bin/bash
# ===========================================
# OdoWell Health Monitor Script
# ===========================================
# Monitora endpoints e serviços, envia alertas por email

set -e

# Carregar variáveis de ambiente
if [ -f /root/drcrwell/.env ]; then
    export $(grep -v '^#' /root/drcrwell/.env | xargs)
fi

# Configurações
LOG_FILE="/root/drcrwell/logs/health_monitor.log"
ALERT_EMAIL="${ALERT_EMAIL:-admin@odowell.pro}"
API_URL="https://api.odowell.pro/health"
FRONTEND_URL="https://app.odowell.pro"

# Criar diretório de logs se não existir
mkdir -p /root/drcrwell/logs

# Função de log
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Função de alerta
send_alert() {
    local subject="$1"
    local message="$2"
    
    log "ALERT: $subject"
    
    # Enviar email (requer mailutils ou similar)
    if command -v mail &> /dev/null; then
        echo "$message" | mail -s "[OdoWell ALERT] $subject" "$ALERT_EMAIL"
    fi
    
    # Também pode enviar via API de webhook (Slack, Discord, etc.)
}

# Verificar API Backend
check_api() {
    local response
    local http_code
    
    http_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 "$API_URL" 2>/dev/null)
    
    if [ "$http_code" == "200" ]; then
        log "✓ API OK (HTTP $http_code)"
        return 0
    else
        send_alert "API DOWN" "API endpoint $API_URL retornou HTTP $http_code"
        return 1
    fi
}

# Verificar Frontend
check_frontend() {
    local http_code
    
    http_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 "$FRONTEND_URL" 2>/dev/null)
    
    if [ "$http_code" == "200" ] || [ "$http_code" == "304" ]; then
        log "✓ Frontend OK (HTTP $http_code)"
        return 0
    else
        send_alert "Frontend DOWN" "Frontend $FRONTEND_URL retornou HTTP $http_code"
        return 1
    fi
}

# Verificar PostgreSQL
check_postgres() {
    local container_id
    container_id=$(docker ps -q -f name=postgres 2>/dev/null | head -1)
    
    if [ -z "$container_id" ]; then
        send_alert "PostgreSQL DOWN" "Container PostgreSQL não encontrado"
        return 1
    fi
    
    if docker exec "$container_id" pg_isready -U "${POSTGRES_USER:-odowell_app}" &>/dev/null; then
        log "✓ PostgreSQL OK"
        return 0
    else
        send_alert "PostgreSQL DOWN" "PostgreSQL não está respondendo"
        return 1
    fi
}

# Verificar Redis
check_redis() {
    local container_id
    container_id=$(docker ps -q -f name=redis 2>/dev/null | head -1)
    
    if [ -z "$container_id" ]; then
        send_alert "Redis DOWN" "Container Redis não encontrado"
        return 1
    fi
    
    if docker exec "$container_id" redis-cli --no-auth-warning -a "${REDIS_PASSWORD:-}" ping 2>/dev/null | grep -q "PONG"; then
        log "✓ Redis OK"
        return 0
    else
        send_alert "Redis DOWN" "Redis não está respondendo"
        return 1
    fi
}

# Verificar espaço em disco
check_disk() {
    local usage
    usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$usage" -lt 85 ]; then
        log "✓ Disco OK (${usage}% usado)"
        return 0
    else
        send_alert "Disco CRÍTICO" "Uso de disco em ${usage}% - limpar urgente!"
        return 1
    fi
}

# Verificar memória
check_memory() {
    local usage
    usage=$(free | awk 'NR==2 {printf "%.0f", $3/$2 * 100}')
    
    if [ "$usage" -lt 90 ]; then
        log "✓ Memória OK (${usage}% usado)"
        return 0
    else
        send_alert "Memória CRÍTICA" "Uso de memória em ${usage}%"
        return 1
    fi
}

# Verificar Docker Swarm nodes
check_swarm() {
    local nodes
    nodes=$(docker node ls --format "{{.Status}}" 2>/dev/null | grep -c "Ready" || echo "0")
    
    if [ "$nodes" -ge 2 ]; then
        log "✓ Swarm OK ($nodes nodes ativos)"
        return 0
    else
        send_alert "Swarm ALERTA" "Apenas $nodes node(s) ativo(s) no Swarm"
        return 1
    fi
}

# Execução principal
log "========== Health Check Started =========="

ERRORS=0

check_api || ((ERRORS++))
check_frontend || ((ERRORS++))
check_postgres || ((ERRORS++))
check_redis || ((ERRORS++))
check_disk || ((ERRORS++))
check_memory || ((ERRORS++))
check_swarm || ((ERRORS++))

if [ $ERRORS -eq 0 ]; then
    log "========== All checks PASSED =========="
else
    log "========== $ERRORS check(s) FAILED =========="
fi

exit $ERRORS
