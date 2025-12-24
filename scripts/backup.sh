#!/bin/bash
# ===========================================
# OdoWell Database Backup Script
# ===========================================
# Este script cria backups do banco de dados PostgreSQL
# e opcionalmente envia para S3
# INCLUI: Alertas por email em caso de falha

set -e

# Configurações
BACKUP_DIR="/root/drcrwell/backups"
RETENTION_DAYS=7
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/odowell_backup_$TIMESTAMP.sql.gz"
LOG_FILE="$BACKUP_DIR/backup.log"
MIN_BACKUP_SIZE_KB=100  # Backup mínimo esperado (100KB)

# Email para alertas (carregar de .env ou usar default)
ALERT_EMAIL="${ALERT_EMAIL:-admin@odowell.pro}"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função de log
log() {
    local message="[$(date '+%Y-%m-%d %H:%M:%S')] $1"
    echo -e "${GREEN}${message}${NC}"
    echo "$message" >> "$LOG_FILE"
}

error() {
    local message="[ERROR] $1"
    echo -e "${RED}${message}${NC}" >&2
    echo "$message" >> "$LOG_FILE"
}

warn() {
    local message="[WARN] $1"
    echo -e "${YELLOW}${message}${NC}"
    echo "$message" >> "$LOG_FILE"
}

# Função para enviar alerta de falha
send_alert() {
    local subject="$1"
    local body="$2"

    # Tentar enviar email via API do backend
    if [ -n "$SMTP_HOST" ]; then
        log "Enviando alerta: $subject"
        # Usar curl para chamar endpoint de alerta interno (se existir)
        # ou usar sendmail/mailx se disponível
        if command -v mail &> /dev/null; then
            echo "$body" | mail -s "$subject" "$ALERT_EMAIL" 2>/dev/null || true
        fi
    fi

    # Log local do alerta
    echo "=== ALERTA ===" >> "$LOG_FILE"
    echo "Assunto: $subject" >> "$LOG_FILE"
    echo "$body" >> "$LOG_FILE"
    echo "==============" >> "$LOG_FILE"
}

# Função de limpeza em caso de falha
cleanup_on_error() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        error "Backup falhou com código de saída: $exit_code"
        send_alert "[ALERTA] Backup OdoWell FALHOU" \
            "O backup do banco de dados OdoWell falhou em $(date).

Servidor: $(hostname)
Código de erro: $exit_code
Último log: $(tail -5 $LOG_FILE 2>/dev/null || echo 'Log indisponível')

Ação necessária: Verifique o status do PostgreSQL e execute o backup manualmente."
    fi
}

trap cleanup_on_error EXIT

# Criar diretório de backup se não existir
mkdir -p "$BACKUP_DIR"

log "Iniciando backup do OdoWell..."

# Encontrar container do PostgreSQL
POSTGRES_CONTAINER=$(docker ps -q -f name=drcrwell_postgres 2>/dev/null)

if [ -z "$POSTGRES_CONTAINER" ]; then
    error "Container PostgreSQL não encontrado!"
    exit 1
fi

log "Container PostgreSQL: $POSTGRES_CONTAINER"

# Criar backup
log "Criando dump do banco de dados..."
docker exec "$POSTGRES_CONTAINER" pg_dump \
    -U odowell_app \
    -d drcrwell_db \
    --no-owner \
    --no-acl \
    2>&1 | gzip > "$BACKUP_FILE"

if [ $? -eq 0 ]; then
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    BACKUP_SIZE_KB=$(du -k "$BACKUP_FILE" | cut -f1)
    log "Backup criado com sucesso: $BACKUP_FILE ($BACKUP_SIZE)"
else
    error "Falha ao criar backup!"
    exit 1
fi

# SEGURANÇA: Verificar tamanho mínimo do backup (detectar backups corrompidos/vazios)
log "Verificando tamanho do backup..."
if [ "$BACKUP_SIZE_KB" -lt "$MIN_BACKUP_SIZE_KB" ]; then
    error "Backup muito pequeno ($BACKUP_SIZE_KB KB < $MIN_BACKUP_SIZE_KB KB) - provavelmente corrompido!"
    send_alert "[ALERTA] Backup OdoWell SUSPEITO" \
        "O backup do banco de dados OdoWell pode estar corrompido.

Servidor: $(hostname)
Arquivo: $BACKUP_FILE
Tamanho: ${BACKUP_SIZE_KB} KB (esperado mínimo: ${MIN_BACKUP_SIZE_KB} KB)
Data: $(date)

Ação necessária: Verifique o status do PostgreSQL e os logs de backup."
    exit 1
fi
log "Tamanho do backup OK: ${BACKUP_SIZE_KB} KB"

# Verificar integridade do backup
log "Verificando integridade do backup (gzip test)..."
if gzip -t "$BACKUP_FILE" 2>/dev/null; then
    log "Backup verificado com sucesso!"
else
    error "Backup corrompido (falha no gzip test)!"
    send_alert "[ALERTA] Backup OdoWell CORROMPIDO" \
        "O backup do banco de dados OdoWell falhou na verificação de integridade.

Servidor: $(hostname)
Arquivo: $BACKUP_FILE
Tamanho: $BACKUP_SIZE
Data: $(date)

Ação necessária: Verifique o disco e execute o backup manualmente."
    exit 1
fi

# SEGURANÇA: Criptografar backup local com GPG (se configurado)
if [ -n "$BACKUP_GPG_PASSWORD" ]; then
    log "Criptografando backup com GPG..."
    ENCRYPTED_FILE="${BACKUP_FILE}.gpg"

    echo "$BACKUP_GPG_PASSWORD" | gpg --batch --yes --passphrase-fd 0 \
        --symmetric --cipher-algo AES256 \
        -o "$ENCRYPTED_FILE" "$BACKUP_FILE" 2>/dev/null

    if [ $? -eq 0 ]; then
        # Remover arquivo não criptografado
        rm -f "$BACKUP_FILE"
        BACKUP_FILE="$ENCRYPTED_FILE"
        log "Backup criptografado: $ENCRYPTED_FILE"
    else
        warn "Falha na criptografia GPG (mantendo backup sem criptografia)"
    fi
fi

# Upload para S3 (se configurado)
if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_ACCESS_KEY" ] && [ -n "$AWS_BUCKET_NAME" ]; then
    log "Enviando backup para S3..."
    aws s3 cp "$BACKUP_FILE" "s3://$AWS_BUCKET_NAME/backups/$(basename $BACKUP_FILE)" \
        --sse AES256 \
        2>/dev/null

    if [ $? -eq 0 ]; then
        log "Backup enviado para S3 com sucesso!"
    else
        warn "Falha ao enviar para S3 (backup local preservado)"
    fi
else
    warn "AWS não configurado - backup apenas local"
fi

# Limpar backups antigos
log "Removendo backups com mais de $RETENTION_DAYS dias..."
find "$BACKUP_DIR" -name "odowell_backup_*.sql.gz*" -mtime +$RETENTION_DAYS -delete 2>/dev/null
BACKUP_COUNT=$(find "$BACKUP_DIR" -name "odowell_backup_*.sql.gz*" | wc -l)
log "Backups locais mantidos: $BACKUP_COUNT"

# Resumo
echo ""
echo "========================================="
echo "BACKUP CONCLUÍDO COM SUCESSO"
echo "========================================="
echo "Arquivo: $BACKUP_FILE"
echo "Tamanho: $BACKUP_SIZE"
echo "Data: $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="
