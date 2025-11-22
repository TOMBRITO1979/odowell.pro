#!/bin/bash

# Dr. Crwell - Script de Deploy Automatizado

set -e

echo "============================================"
echo "Dr. Crwell - Deploy Automatizado"
echo "============================================"

# Check if .env exists
if [ ! -f .env ]; then
    echo "Arquivo .env não encontrado!"
    echo "Copiando .env.example para .env..."
    cp .env.example .env
    echo ""
    echo "⚠️  IMPORTANTE: Edite o arquivo .env e configure:"
    echo "  - DB_PASSWORD"
    echo "  - JWT_SECRET"
    echo "  - FRONTEND_URL"
    echo "  - BACKEND_URL"
    echo "  - DOCKER_USERNAME"
    echo ""
    echo "Depois execute novamente: ./deploy.sh"
    exit 1
fi

# Load environment variables
export $(cat .env | grep -v '^#' | xargs)

echo "Configurações:"
echo "  Frontend: https://${FRONTEND_URL}"
echo "  Backend: https://${BACKEND_URL}"
echo "  Docker User: ${DOCKER_USERNAME}"
echo ""

# Login to Docker Hub
echo "Login no Docker Hub..."
if [ -n "$DOCKER_TOKEN" ]; then
    echo "$DOCKER_TOKEN" | docker login -u "$DOCKER_USERNAME" --password-stdin
else
    docker login -u "$DOCKER_USERNAME"
fi

# Build images
echo ""
echo "Building imagens..."
make build

# Push images
echo ""
echo "Enviando imagens para Docker Hub..."
make push

# Deploy to Swarm
echo ""
echo "Deploy no Docker Swarm..."
# Export all env vars so docker stack can use them
set -a
source .env
set +a
make deploy

echo ""
echo "============================================"
echo "Deploy concluído com sucesso!"
echo "============================================"
echo ""
echo "Acesse:"
echo "  Frontend: https://${FRONTEND_URL}"
echo "  Backend API: https://${BACKEND_URL}/health"
echo ""
echo "Para ver logs:"
echo "  Backend: make logs-backend"
echo "  Frontend: make logs-frontend"
echo "  Database: make logs-db"
echo ""
