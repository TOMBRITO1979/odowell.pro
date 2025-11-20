.PHONY: help build push deploy remove logs

help:
	@echo "Dr. Crwell - Sistema de Gestão Odontológica"
	@echo ""
	@echo "Comandos disponíveis:"
	@echo "  make build         - Build das imagens Docker"
	@echo "  make push          - Push das imagens para Docker Hub"
	@echo "  make deploy        - Deploy no Docker Swarm"
	@echo "  make remove        - Remove a stack do Swarm"
	@echo "  make logs-backend  - Ver logs do backend"
	@echo "  make logs-frontend - Ver logs do frontend"

build:
	@echo "Building backend..."
	docker build --no-cache -t $(DOCKER_USERNAME)/drcrwell-backend:latest ./backend
	@echo "Building frontend..."
	docker build --no-cache -t $(DOCKER_USERNAME)/drcrwell-frontend:latest ./frontend
	@echo "Build completed!"

push:
	@echo "Pushing images to Docker Hub..."
	docker push $(DOCKER_USERNAME)/drcrwell-backend:latest
	docker push $(DOCKER_USERNAME)/drcrwell-frontend:latest
	@echo "Push completed!"

deploy:
	@echo "Deploying to Docker Swarm..."
	docker stack deploy -c docker-stack.yml drcrwell
	@echo "Deployment completed!"
	@echo "Frontend: https://$(FRONTEND_URL)"
	@echo "Backend: https://$(BACKEND_URL)"

remove:
	@echo "Removing stack from Swarm..."
	docker stack rm drcrwell
	@echo "Stack removed!"

logs-backend:
	docker service logs -f drcrwell_backend

logs-frontend:
	docker service logs -f drcrwell_frontend

logs-db:
	docker service logs -f drcrwell_postgres
