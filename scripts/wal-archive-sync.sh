#!/bin/bash
# ===========================================
# OdoWell WAL Archive Sync Script
# ===========================================
# Sincroniza WAL archives para S3 para PITR
# Executar via cron a cada 5 minutos

set -e

LOG_FILE="/root/drcrwell/backups/wal_sync.log"
WAL_VOLUME="drcrwell_wal_archive"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

# Verificar se AWS está configurado
if [ -z "$AWS_ACCESS_KEY_ID" ] || [ -z "$AWS_SECRET_ACCESS_KEY" ] || [ -z "$AWS_BUCKET_NAME" ]; then
    # Tentar carregar do .env
    if [ -f /root/drcrwell/.env ]; then
        source /root/drcrwell/.env
    fi
fi

if [ -z "$AWS_BUCKET_NAME" ]; then
    log "ERROR: AWS não configurado, pulando sync"
    exit 0
fi

# Encontrar caminho do volume WAL archive
WAL_PATH=$(docker volume inspect "$WAL_VOLUME" --format '{{.Mountpoint}}' 2>/dev/null)

if [ -z "$WAL_PATH" ] || [ ! -d "$WAL_PATH" ]; then
    log "WAL volume não encontrado ou vazio"
    exit 0
fi

# Contar arquivos WAL
WAL_COUNT=$(find "$WAL_PATH" -name "0*" -type f 2>/dev/null | wc -l)

if [ "$WAL_COUNT" -eq 0 ]; then
    exit 0  # Nada para sincronizar
fi

log "Sincronizando $WAL_COUNT arquivos WAL para S3..."

# Sincronizar para S3
aws s3 sync "$WAL_PATH" "s3://$AWS_BUCKET_NAME/wal_archive/" \
    --sse AES256 \
    --exclude "*" \
    --include "0*" \
    2>> "$LOG_FILE"

if [ $? -eq 0 ]; then
    log "WAL sync concluído: $WAL_COUNT arquivos"

    # Limpar WAL files antigos (mais de 7 dias) após upload bem sucedido
    find "$WAL_PATH" -name "0*" -type f -mtime +7 -delete 2>/dev/null
    DELETED=$(find "$WAL_PATH" -name "0*" -type f -mtime +7 2>/dev/null | wc -l)
    if [ "$DELETED" -gt 0 ]; then
        log "Removidos $DELETED WAL files antigos"
    fi
else
    log "ERROR: Falha no sync WAL para S3"
fi
