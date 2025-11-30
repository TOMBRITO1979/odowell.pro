#!/bin/bash
# ===========================================
# OdoWell Database Backup Script
# ===========================================
# Este script cria backups do banco de dados PostgreSQL
# e opcionalmente envia para S3

set -e

# Configurações
BACKUP_DIR="/root/drcrwell/backups"
RETENTION_DAYS=7
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/odowell_backup_$TIMESTAMP.sql.gz"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função de log
log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

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
    -U drcrwell_user \
    -d drcrwell_db \
    --no-owner \
    --no-acl \
    2>/dev/null | gzip > "$BACKUP_FILE"

if [ $? -eq 0 ]; then
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    log "Backup criado com sucesso: $BACKUP_FILE ($BACKUP_SIZE)"
else
    error "Falha ao criar backup!"
    exit 1
fi

# Verificar integridade do backup
log "Verificando integridade do backup..."
if gzip -t "$BACKUP_FILE" 2>/dev/null; then
    log "Backup verificado com sucesso!"
else
    error "Backup corrompido!"
    exit 1
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
find "$BACKUP_DIR" -name "odowell_backup_*.sql.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null
BACKUP_COUNT=$(find "$BACKUP_DIR" -name "odowell_backup_*.sql.gz" | wc -l)
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
