#!/bin/bash
# ===========================================
# OdoWell Tenant Safety Restore Script
# ===========================================
# Restaura um tenant a partir de um backup de segurança
# criado pelo script restore-tenant.sh

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

# Verificar argumentos
if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Uso: $0 <backup_seguranca.sql.gz> <tenant_id>"
    echo ""
    echo "Backups de segurança disponíveis:"
    ls -lh "$BACKUP_DIR"/pre_restore_tenant_*.sql.gz 2>/dev/null || echo "Nenhum backup de segurança encontrado"
    exit 1
fi

BACKUP_FILE="$1"
TENANT_ID="$2"
SCHEMA_NAME="tenant_$TENANT_ID"

# Verificar se arquivo existe
if [ ! -f "$BACKUP_FILE" ]; then
    if [ -f "$BACKUP_DIR/$BACKUP_FILE" ]; then
        BACKUP_FILE="$BACKUP_DIR/$BACKUP_FILE"
    else
        error "Arquivo não encontrado: $BACKUP_FILE"
        exit 1
    fi
fi

log "Backup de segurança: $BACKUP_FILE"
log "Tenant ID: $TENANT_ID"

# Confirmar
echo ""
warn "Isto irá reverter o tenant $TENANT_ID para o estado anterior à restauração."
read -p "Continuar? (sim/não): " CONFIRM

if [ "$CONFIRM" != "sim" ]; then
    echo "Operação cancelada."
    exit 0
fi

# Encontrar container
POSTGRES_CONTAINER=$(docker ps -q -f name=drcrwell_postgres 2>/dev/null)

if [ -z "$POSTGRES_CONTAINER" ]; then
    error "Container PostgreSQL não encontrado!"
    exit 1
fi

# Dropar tabelas atuais do tenant
log "Limpando tabelas do tenant $TENANT_ID..."
docker exec "$POSTGRES_CONTAINER" psql -U odowell_app -d drcrwell_db -c \
    "DO \$\$ DECLARE r RECORD;
    BEGIN
        FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = '$SCHEMA_NAME')
        LOOP
            EXECUTE 'DROP TABLE IF EXISTS $SCHEMA_NAME.' || quote_ident(r.tablename) || ' CASCADE';
        END LOOP;
    END \$\$;" 2>/dev/null

# Restaurar backup de segurança
log "Restaurando backup de segurança..."
gunzip -c "$BACKUP_FILE" | docker exec -i "$POSTGRES_CONTAINER" psql -U odowell_app -d drcrwell_db 2>/dev/null

if [ $? -eq 0 ]; then
    log "Reversão concluída com sucesso!"
else
    error "Falha na reversão!"
    exit 1
fi
