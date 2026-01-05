#!/bin/bash
# ===========================================
# OdoWell Log Viewer
# ===========================================
# Uso: ./view_logs.sh [service] [lines]
# Exemplo: ./view_logs.sh backend 100

SERVICE=${1:-backend}
LINES=${2:-50}

echo "=== Últimas $LINES linhas do $SERVICE ==="
docker service logs "drcrwell_${SERVICE}" --tail "$LINES" 2>&1

echo ""
echo "=== Opções disponíveis ==="
echo "  ./view_logs.sh backend 100"
echo "  ./view_logs.sh frontend 50"
echo "  ./view_logs.sh postgres 50"
echo "  ./view_logs.sh redis 50"
