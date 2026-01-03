# Guia de Deploy - Odowell em Nova VPS

## Requisitos da VPS

- **OS**: Ubuntu 22.04 LTS (recomendado)
- **RAM**: Mínimo 8GB (recomendado 16GB)
- **CPU**: Mínimo 4 vCPUs
- **Disco**: Mínimo 100GB SSD
- **Portas abertas**: 80, 443, 22

---

## Passo 1: Preparar a VPS

```bash
# Atualizar sistema
apt update && apt upgrade -y

# Instalar dependências
apt install -y curl wget git htop nano ufw

# Configurar firewall
ufw allow 22
ufw allow 80
ufw allow 443
ufw --force enable
```

---

## Passo 2: Instalar Docker

```bash
# Instalar Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Adicionar usuário ao grupo docker (se não for root)
usermod -aG docker $USER

# Iniciar Docker Swarm
docker swarm init

# Verificar
docker info | grep Swarm
```

---

## Passo 3: Instalar Traefik (Proxy Reverso + SSL)

```bash
# Criar rede pública
docker network create --driver overlay --attachable network_public

# Criar diretório para Traefik
mkdir -p /opt/traefik
cd /opt/traefik

# Criar arquivo de configuração do Traefik
cat > traefik.yml << 'EOF'
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    command:
      - "--api.dashboard=true"
      - "--providers.docker=true"
      - "--providers.docker.swarmMode=true"
      - "--providers.docker.exposedbydefault=false"
      - "--providers.docker.network=network_public"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.web.http.redirections.entrypoint.to=websecure"
      - "--entrypoints.web.http.redirections.entrypoint.scheme=https"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.letsencryptresolver.acme.httpchallenge=true"
      - "--certificatesresolvers.letsencryptresolver.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.letsencryptresolver.acme.email=seu-email@exemplo.com"
      - "--certificatesresolvers.letsencryptresolver.acme.storage=/letsencrypt/acme.json"
      - "--log.level=INFO"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik_letsencrypt:/letsencrypt
    networks:
      - network_public
    deploy:
      placement:
        constraints:
          - node.role == manager
      labels:
        - "traefik.enable=true"
        - "traefik.http.routers.traefik.rule=Host(`traefik.odowell.pro`)"
        - "traefik.http.routers.traefik.entrypoints=websecure"
        - "traefik.http.routers.traefik.tls.certresolver=letsencryptresolver"
        - "traefik.http.routers.traefik.service=api@internal"
        - "traefik.http.services.traefik.loadbalancer.server.port=8080"

networks:
  network_public:
    external: true

volumes:
  traefik_letsencrypt:
EOF

# IMPORTANTE: Alterar o email no arquivo acima para seu email real

# Deploy do Traefik
docker stack deploy -c traefik.yml traefik
```

---

## Passo 4: Criar Certificados SSL para PostgreSQL

```bash
# Criar diretório para certificados
mkdir -p /opt/ssl
cd /opt/ssl

# Gerar certificado autoassinado para PostgreSQL
openssl req -new -x509 -days 3650 -nodes -text \
  -out server.crt \
  -keyout server.key \
  -subj "/CN=postgres"

# Ajustar permissões
chmod 600 server.key
chmod 644 server.crt

# Criar Docker secrets
docker secret create pg_ssl_cert server.crt
docker secret create pg_ssl_key server.key
```

---

## Passo 5: Configurar o Aplicativo

```bash
# Criar diretório do app
mkdir -p /opt/odowell
cd /opt/odowell

# Baixar arquivos necessários (ou copiar da VPS antiga)
# Você precisa do docker-stack.yml e .env
```

### Criar arquivo .env

```bash
cat > .env << 'EOF'
# URLs Configuration
FRONTEND_URL=app.odowell.pro
BACKEND_URL=api.odowell.pro

# Docker Configuration
DOCKER_USERNAME=tomautomations

# Replicas
FRONTEND_REPLICAS=1
BACKEND_REPLICAS=4

# Database Configuration
DB_NAME=drcrwell_db
DB_USER=odowell_app
DB_PASSWORD=SUA_SENHA_DO_BANCO_AQUI
DB_SSL_MODE=require

# Database Connection Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=3600

# Redis
REDIS_PASSWORD=SUA_SENHA_REDIS_AQUI

# JWT Secret (gere com: openssl rand -base64 64)
JWT_SECRET=SEU_JWT_SECRET_AQUI

# Encryption Key (gere com: openssl rand -base64 32)
ENCRYPTION_KEY=SUA_ENCRYPTION_KEY_AQUI

# CORS Origins
CORS_ORIGINS=https://app.odowell.pro,http://localhost:3000

# AWS S3 Configuration
AWS_ACCESS_KEY_ID=SUA_AWS_KEY_AQUI
AWS_SECRET_ACCESS_KEY=SUA_AWS_SECRET_AQUI
AWS_BUCKET_NAME=drcrwell-app
AWS_REGION=sa-east-1
S3_BASE_FOLDER=exams

# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=appodowellpro@gmail.com
SMTP_PASSWORD=SUA_SENHA_SMTP_AQUI
SMTP_FROM=appodowellpro@gmail.com

# App Branding
APP_NAME=OdoWell.Pro

# Stripe Configuration
STRIPE_SECRET_KEY=SUA_STRIPE_SECRET_KEY
STRIPE_WEBHOOK_SECRET=SEU_STRIPE_WEBHOOK_SECRET
STRIPE_PRICE_BRONZE=price_1S43CNFZD94PE51sZ6rdeiZA
STRIPE_PRICE_SILVER=price_1OYG4DFZD94PE51sKuwJYTE0
STRIPE_PRICE_GOLD=price_1SZYpOFZD94PE51sxdrLUHP8

# Timezone
TZ=America/Sao_Paulo

# Environment
ENV=production
GIN_MODE=release
EOF
```

**IMPORTANTE**: Substitua todos os valores `SUA_*_AQUI` pelos valores reais da VPS antiga.

### Copiar docker-stack.yml

Copie o arquivo `docker-stack.yml` da VPS antiga para `/opt/odowell/docker-stack.yml`

---

## Passo 6: Deploy da Aplicação

```bash
cd /opt/odowell

# Fazer pull das imagens
docker pull tomautomations/drcrwell-backend:latest
docker pull tomautomations/drcrwell-frontend:latest

# Deploy do stack
docker stack deploy -c docker-stack.yml drcrwell

# Verificar status
docker service ls
```

---

## Passo 7: Configurar DNS

No seu provedor de DNS (Cloudflare, Route53, etc.), aponte os domínios para o **novo IP da VPS**:

| Tipo | Nome | Valor |
|------|------|-------|
| A | app.odowell.pro | NOVO_IP_DA_VPS |
| A | api.odowell.pro | NOVO_IP_DA_VPS |
| A | *.odowell.pro | NOVO_IP_DA_VPS |

**Nota**: O wildcard `*.odowell.pro` é necessário para o portal do paciente por subdomínio.

---

## Passo 8: Migrar Banco de Dados (IMPORTANTE!)

### Opção A: Dump e Restore

Na VPS antiga:
```bash
# Fazer backup do banco
docker exec $(docker ps -q -f name=postgres) \
  pg_dump -U odowell_app -d drcrwell_db -F c -f /tmp/backup.dump

# Copiar para fora do container
docker cp $(docker ps -q -f name=postgres):/tmp/backup.dump ./backup.dump

# Transferir para nova VPS
scp backup.dump root@NOVA_VPS_IP:/opt/odowell/
```

Na VPS nova (após o banco estar rodando):
```bash
# Aguardar o banco iniciar
sleep 60

# Copiar backup para o container
docker cp /opt/odowell/backup.dump $(docker ps -q -f name=postgres):/tmp/

# Restaurar
docker exec $(docker ps -q -f name=postgres) \
  pg_restore -U odowell_app -d drcrwell_db -c /tmp/backup.dump
```

### Opção B: pg_dump direto (se as VPS tiverem conectividade)

```bash
# Na VPS nova, após o banco estar vazio e rodando
docker exec $(docker ps -q -f name=postgres) \
  pg_dump -h IP_VPS_ANTIGA -U odowell_app -d drcrwell_db | \
  docker exec -i $(docker ps -q -f name=postgres) \
  psql -U odowell_app -d drcrwell_db
```

---

## Passo 9: Verificar Deploy

```bash
# Ver status dos serviços
docker service ls

# Ver logs do backend
docker service logs drcrwell_backend --tail 100 -f

# Ver logs do frontend
docker service logs drcrwell_frontend --tail 100 -f

# Ver logs do banco
docker service logs drcrwell_postgres --tail 100 -f

# Testar health check
curl -k https://api.odowell.pro/health
curl -k https://app.odowell.pro
```

---

## Comandos Úteis

```bash
# Atualizar imagens
docker pull tomautomations/drcrwell-backend:latest
docker pull tomautomations/drcrwell-frontend:latest
docker service update --image tomautomations/drcrwell-backend:latest drcrwell_backend --force
docker service update --image tomautomations/drcrwell-frontend:latest drcrwell_frontend --force

# Reiniciar serviço
docker service update --force drcrwell_backend

# Ver logs em tempo real
docker service logs -f drcrwell_backend

# Escalar replicas
docker service scale drcrwell_backend=4

# Acessar banco
docker exec -it $(docker ps -q -f name=postgres) psql -U odowell_app -d drcrwell_db

# Acessar Redis
docker exec -it $(docker ps -q -f name=redis) redis-cli -a SUA_SENHA_REDIS

# Remover stack completamente
docker stack rm drcrwell
```

---

## Checklist Final

- [ ] Docker instalado e Swarm inicializado
- [ ] Rede `network_public` criada
- [ ] Traefik rodando
- [ ] Secrets `pg_ssl_cert` e `pg_ssl_key` criados
- [ ] Arquivo `.env` configurado com todas as senhas
- [ ] Stack `drcrwell` deployado
- [ ] DNS apontando para novo IP
- [ ] Certificados SSL gerados pelo Let's Encrypt
- [ ] Banco de dados migrado
- [ ] Teste de login funcionando
- [ ] Portal do paciente funcionando

---

## Troubleshooting

### Erro: "network not found"
```bash
docker network create --driver overlay --attachable network_public
```

### Erro: "secret not found"
```bash
# Recriar secrets
docker secret rm pg_ssl_cert pg_ssl_key
docker secret create pg_ssl_cert /opt/ssl/server.crt
docker secret create pg_ssl_key /opt/ssl/server.key
```

### Banco não inicia
```bash
# Ver logs
docker service logs drcrwell_postgres

# Verificar se volumes existem
docker volume ls | grep postgres
```

### SSL não funciona
```bash
# Ver logs do Traefik
docker service logs traefik_traefik

# Verificar se portas 80/443 estão abertas
ufw status
```

---

## Contato

Se precisar de ajuda, verifique os logs primeiro:
```bash
docker service logs drcrwell_backend --tail 200
```
