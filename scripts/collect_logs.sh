#!/bin/bash
# ===========================================
# OdoWell Log Collector
# ===========================================
# Coleta logs de todos os services e nodes

LOG_DIR="/root/drcrwell/logs"
DATE=$(date +%Y%m%d)
RETENTION_DAYS=7

mkdir -p "$LOG_DIR/daily"

echo "[$(date)] Coletando logs..."

# Logs do Backend (últimas 24h)
docker service logs drcrwell_backend --since 24h 2>&1 | tail -10000 > "$LOG_DIR/daily/backend_${DATE}.log"

# Logs do Frontend
docker service logs drcrwell_frontend --since 24h 2>&1 | tail -5000 > "$LOG_DIR/daily/frontend_${DATE}.log"

# Logs do PostgreSQL
docker service logs drcrwell_postgres --since 24h 2>&1 | tail -5000 > "$LOG_DIR/daily/postgres_${DATE}.log"

# Logs do Redis
docker service logs drcrwell_redis --since 24h 2>&1 | tail -2000 > "$LOG_DIR/daily/redis_${DATE}.log"

# Logs do Traefik
docker service logs traefik_traefik --since 24h 2>&1 | tail -5000 > "$LOG_DIR/daily/traefik_${DATE}.log"

# Comprimir logs antigos
find "$LOG_DIR/daily" -name "*.log" -mtime +1 -exec gzip {} \; 2>/dev/null

# Remover logs muito antigos
find "$LOG_DIR/daily" -name "*.log.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null

echo "[$(date)] Logs coletados em $LOG_DIR/daily/"
ls -lh "$LOG_DIR/daily/" | tail -20
