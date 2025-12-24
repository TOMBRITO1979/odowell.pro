#!/bin/bash
# ===========================================
# OdoWell Per-Tenant Restore Script
# ===========================================
# Este script restaura apenas um tenant específico
# de um backup completo, sem afetar outros tenants

set -e

BACKUP_DIR="/root/drcrwell/backups"
TEMP_DIR="/tmp/odowell_restore_$$"

# Cores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
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

info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

cleanup() {
    rm -rf "$TEMP_DIR" 2>/dev/null || true
}

trap cleanup EXIT

# Verificar argumentos
if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Uso: $0 <arquivo_backup.sql.gz> <tenant_id>"
    echo ""
    echo "Exemplos:"
    echo "  $0 odowell_backup_20241224_120000.sql.gz 5"
    echo "  $0 /root/drcrwell/backups/odowell_backup_20241224_120000.sql.gz 12"
    echo ""
    echo "Backups disponíveis:"
    ls -lh "$BACKUP_DIR"/odowell_backup_*.sql.gz 2>/dev/null | head -10 || echo "Nenhum backup encontrado"
    echo ""

    # Listar tenants se possível
    POSTGRES_CONTAINER=$(docker ps -q -f name=drcrwell_postgres 2>/dev/null)
    if [ -n "$POSTGRES_CONTAINER" ]; then
        echo "Tenants disponíveis:"
        docker exec "$POSTGRES_CONTAINER" psql -U odowell_app -d drcrwell_db -t -c \
            "SELECT id, name FROM public.tenants WHERE deleted_at IS NULL ORDER BY id;" 2>/dev/null || true
    fi
    exit 1
fi

BACKUP_FILE="$1"
TENANT_ID="$2"
SCHEMA_NAME="tenant_$TENANT_ID"

# Validar tenant_id
if ! [[ "$TENANT_ID" =~ ^[0-9]+$ ]]; then
    error "tenant_id deve ser um número: $TENANT_ID"
    exit 1
fi

# Verificar se arquivo existe
if [ ! -f "$BACKUP_FILE" ]; then
    if [ -f "$BACKUP_DIR/$BACKUP_FILE" ]; then
        BACKUP_FILE="$BACKUP_DIR/$BACKUP_FILE"
    else
        error "Arquivo não encontrado: $BACKUP_FILE"
        exit 1
    fi
fi

log "Arquivo de backup: $BACKUP_FILE"
log "Tenant ID: $TENANT_ID"
log "Schema: $SCHEMA_NAME"

# Encontrar container do PostgreSQL
POSTGRES_CONTAINER=$(docker ps -q -f name=drcrwell_postgres 2>/dev/null)

if [ -z "$POSTGRES_CONTAINER" ]; then
    error "Container PostgreSQL não encontrado!"
    exit 1
fi

log "Container PostgreSQL: $POSTGRES_CONTAINER"

# Verificar se tenant existe
TENANT_EXISTS=$(docker exec "$POSTGRES_CONTAINER" psql -U odowell_app -d drcrwell_db -t -c \
    "SELECT COUNT(*) FROM public.tenants WHERE id = $TENANT_ID;" 2>/dev/null | tr -d ' ')

if [ "$TENANT_EXISTS" != "1" ]; then
    error "Tenant ID $TENANT_ID não encontrado no banco de dados!"
    exit 1
fi

# Buscar nome do tenant
TENANT_NAME=$(docker exec "$POSTGRES_CONTAINER" psql -U odowell_app -d drcrwell_db -t -c \
    "SELECT name FROM public.tenants WHERE id = $TENANT_ID;" 2>/dev/null | tr -d ' ')

info "Tenant encontrado: $TENANT_NAME (ID: $TENANT_ID)"

# Confirmar restauração
echo ""
warn "ATENÇÃO: Isto irá SUBSTITUIR todos os dados do tenant '$TENANT_NAME' (ID: $TENANT_ID)!"
warn "Outros tenants NÃO serão afetados."
echo ""
read -p "Tem certeza que deseja continuar? (digite 'sim' para confirmar): " CONFIRM

if [ "$CONFIRM" != "sim" ]; then
    echo "Operação cancelada."
    exit 0
fi

# Criar diretório temporário
mkdir -p "$TEMP_DIR"

# Criar backup de segurança do tenant antes de restaurar
SAFETY_BACKUP="$BACKUP_DIR/pre_restore_tenant_${TENANT_ID}_$(date +%Y%m%d_%H%M%S).sql.gz"
log "Criando backup de segurança do tenant $TENANT_ID: $SAFETY_BACKUP"

docker exec "$POSTGRES_CONTAINER" pg_dump -U odowell_app -d drcrwell_db \
    --schema="$SCHEMA_NAME" \
    --no-owner --no-acl 2>/dev/null | gzip > "$SAFETY_BACKUP"

# Extrair backup completo
log "Extraindo backup..."
gunzip -c "$BACKUP_FILE" > "$TEMP_DIR/full_backup.sql"

# Extrair apenas o schema do tenant do backup
log "Extraindo dados do tenant $TENANT_ID do backup..."

# Criar arquivo SQL para o tenant específico
TENANT_SQL="$TEMP_DIR/tenant_${TENANT_ID}.sql"

# Usar awk para extrair apenas as linhas do schema específico
# O pg_dump organiza os dados por schema, então extraímos a seção relevante
awk -v schema="$SCHEMA_NAME" '
    # Detectar início do schema
    /^CREATE SCHEMA/ && $0 ~ schema { in_schema=1 }
    /^SET search_path = / && $0 ~ schema { in_schema=1 }

    # Se estamos no schema, capturar linhas
    in_schema { print }

    # Detectar fim do schema (próximo schema ou fim de dados do schema)
    in_schema && /^SET search_path = / && !($0 ~ schema) { in_schema=0 }
    in_schema && /^\\connect/ { in_schema=0 }
' "$TEMP_DIR/full_backup.sql" > "$TENANT_SQL"

# Verificar se extraiu dados
EXTRACTED_SIZE=$(wc -c < "$TENANT_SQL")
if [ "$EXTRACTED_SIZE" -lt 100 ]; then
    # Tentar método alternativo: grep por linhas que contém o schema
    log "Tentando método alternativo de extração..."
    grep -E "(${SCHEMA_NAME}|SET search_path.*${SCHEMA_NAME})" "$TEMP_DIR/full_backup.sql" > "$TENANT_SQL" 2>/dev/null || true

    EXTRACTED_SIZE=$(wc -c < "$TENANT_SQL")
    if [ "$EXTRACTED_SIZE" -lt 100 ]; then
        error "Não foi possível extrair dados do tenant $TENANT_ID do backup!"
        error "Verifique se o backup contém dados deste tenant."
        exit 1
    fi
fi

log "Dados extraídos: $(wc -l < "$TENANT_SQL") linhas"

# Restaurar schema do tenant
log "Restaurando tenant $TENANT_ID..."

# Primeiro, dropar tabelas existentes do tenant (mas manter o schema)
docker exec "$POSTGRES_CONTAINER" psql -U odowell_app -d drcrwell_db -c \
    "DO \$\$ DECLARE r RECORD;
    BEGIN
        FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = '$SCHEMA_NAME')
        LOOP
            EXECUTE 'DROP TABLE IF EXISTS $SCHEMA_NAME.' || quote_ident(r.tablename) || ' CASCADE';
        END LOOP;
    END \$\$;" 2>/dev/null

# Restaurar dados do tenant
cat "$TENANT_SQL" | docker exec -i "$POSTGRES_CONTAINER" psql -U odowell_app -d drcrwell_db 2>/dev/null

RESTORE_STATUS=$?

if [ $RESTORE_STATUS -eq 0 ]; then
    log "Restauração do tenant $TENANT_ID concluída com sucesso!"
    echo ""
    echo "========================================="
    echo "RESTAURAÇÃO DE TENANT CONCLUÍDA"
    echo "========================================="
    echo "Tenant: $TENANT_NAME (ID: $TENANT_ID)"
    echo "Schema: $SCHEMA_NAME"
    echo "Backup fonte: $BACKUP_FILE"
    echo "Backup de segurança: $SAFETY_BACKUP"
    echo "========================================="
    echo ""
    info "Para reverter, use: ./restore-tenant-safety.sh $SAFETY_BACKUP $TENANT_ID"
else
    error "Falha na restauração do tenant!"
    warn "Use o backup de segurança para reverter: $SAFETY_BACKUP"
    exit 1
fi
