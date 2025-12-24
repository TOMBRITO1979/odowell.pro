#!/bin/bash
# ===========================================
# OdoWell Redis Backup Script
# ===========================================
# Faz backup do Redis (RDB) e envia para S3
# Executar via cron 1x/dia

set -e

BACKUP_DIR="/root/drcrwell/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/redis_backup_$TIMESTAMP.rdb"
LOG_FILE="$BACKUP_DIR/redis_backup.log"
RETENTION_DAYS=7

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

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

# Criar diretório se não existir
mkdir -p "$BACKUP_DIR"

log "Iniciando backup do Redis..."

# Encontrar container Redis
REDIS_CONTAINER=$(docker ps -q -f name=drcrwell_redis 2>/dev/null | head -1)

if [ -z "$REDIS_CONTAINER" ]; then
    error "Container Redis não encontrado!"
    exit 1
fi

log "Container Redis: $REDIS_CONTAINER"

# Forçar salvamento do RDB
log "Executando BGSAVE..."
docker exec "$REDIS_CONTAINER" redis-cli -a "$REDIS_PASSWORD" BGSAVE 2>/dev/null

# Aguardar conclusão do BGSAVE
sleep 5
LASTSAVE=$(docker exec "$REDIS_CONTAINER" redis-cli -a "$REDIS_PASSWORD" LASTSAVE 2>/dev/null)
log "Último save: $LASTSAVE"

# Copiar arquivo RDB do container
log "Copiando dump.rdb..."
docker cp "$REDIS_CONTAINER:/data/dump.rdb" "$BACKUP_FILE" 2>/dev/null

if [ ! -f "$BACKUP_FILE" ]; then
    error "Falha ao copiar dump.rdb!"
    exit 1
fi

BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
log "Backup criado: $BACKUP_FILE ($BACKUP_SIZE)"

# Comprimir backup
log "Comprimindo backup..."
gzip "$BACKUP_FILE"
BACKUP_FILE="${BACKUP_FILE}.gz"

# Upload para S3 (se configurado)
if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_ACCESS_KEY" ] && [ -n "$AWS_BUCKET_NAME" ]; then
    log "Enviando backup Redis para S3..."
    aws s3 cp "$BACKUP_FILE" "s3://$AWS_BUCKET_NAME/redis_backups/$(basename $BACKUP_FILE)" \
        --sse AES256 \
        2>> "$LOG_FILE"

    if [ $? -eq 0 ]; then
        log "Backup Redis enviado para S3 com sucesso!"
    else
        warn "Falha ao enviar para S3 (backup local preservado)"
    fi
else
    warn "AWS não configurado - backup apenas local"
fi

# Limpar backups antigos
log "Removendo backups Redis com mais de $RETENTION_DAYS dias..."
find "$BACKUP_DIR" -name "redis_backup_*.rdb.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null
BACKUP_COUNT=$(find "$BACKUP_DIR" -name "redis_backup_*.rdb.gz" | wc -l)
log "Backups Redis mantidos: $BACKUP_COUNT"

# Resumo
echo ""
echo "========================================="
echo "BACKUP REDIS CONCLUÍDO"
echo "========================================="
echo "Arquivo: $BACKUP_FILE"
echo "Tamanho: $(du -h "$BACKUP_FILE" | cut -f1)"
echo "Data: $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="
