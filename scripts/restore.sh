#!/bin/bash
# ===========================================
# OdoWell Database Restore Script
# ===========================================
# Este script restaura um backup do banco de dados

set -e

BACKUP_DIR="/root/drcrwell/backups"

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Verificar se arquivo foi passado como argumento
if [ -z "$1" ]; then
    echo "Uso: $0 <arquivo_backup.sql.gz>"
    echo ""
    echo "Backups disponíveis:"
    ls -lh "$BACKUP_DIR"/odowell_backup_*.sql.gz 2>/dev/null || echo "Nenhum backup encontrado"
    exit 1
fi

BACKUP_FILE="$1"

# Verificar se arquivo existe
if [ ! -f "$BACKUP_FILE" ]; then
    # Tentar no diretório de backups
    if [ -f "$BACKUP_DIR/$BACKUP_FILE" ]; then
        BACKUP_FILE="$BACKUP_DIR/$BACKUP_FILE"
    else
        error "Arquivo não encontrado: $BACKUP_FILE"
        exit 1
    fi
fi

log "Arquivo de backup: $BACKUP_FILE"

# Confirmar restauração
echo ""
warn "ATENÇÃO: Isto irá SUBSTITUIR todos os dados atuais!"
read -p "Tem certeza que deseja continuar? (digite 'sim' para confirmar): " CONFIRM

if [ "$CONFIRM" != "sim" ]; then
    echo "Operação cancelada."
    exit 0
fi

# Encontrar container do PostgreSQL
POSTGRES_CONTAINER=$(docker ps -q -f name=drcrwell_postgres 2>/dev/null)

if [ -z "$POSTGRES_CONTAINER" ]; then
    error "Container PostgreSQL não encontrado!"
    exit 1
fi

log "Container PostgreSQL: $POSTGRES_CONTAINER"

# Criar backup de segurança antes de restaurar
SAFETY_BACKUP="$BACKUP_DIR/pre_restore_$(date +%Y%m%d_%H%M%S).sql.gz"
log "Criando backup de segurança: $SAFETY_BACKUP"
docker exec "$POSTGRES_CONTAINER" pg_dump -U odowell_app -d drcrwell_db 2>/dev/null | gzip > "$SAFETY_BACKUP"

# Restaurar
log "Restaurando backup..."
gunzip -c "$BACKUP_FILE" | docker exec -i "$POSTGRES_CONTAINER" psql -U odowell_app -d drcrwell_db 2>/dev/null

if [ $? -eq 0 ]; then
    log "Restauração concluída com sucesso!"
    echo ""
    echo "========================================="
    echo "RESTAURAÇÃO CONCLUÍDA"
    echo "========================================="
    echo "Backup restaurado: $BACKUP_FILE"
    echo "Backup de segurança: $SAFETY_BACKUP"
    echo "========================================="
else
    error "Falha na restauração!"
    warn "Use o backup de segurança para reverter: $SAFETY_BACKUP"
    exit 1
fi
