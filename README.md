# Dr. Crwell - Sistema de Gestão para Consultório Dentário

Sistema SaaS Multitenant completo para gestão de consultórios odontológicos, desenvolvido com Go (backend) e React (frontend).

## Características

### Backend
- **Linguagem**: Go 1.21+ (Golang)
- **Framework**: Gin (ultra-rápido e leve)
- **Banco de Dados**: PostgreSQL com arquitetura multitenant (schema por tenant)
- **Autenticação**: JWT
- **Performance**: Baixo consumo de memória (~10-20MB), ideal para alta carga

### Frontend
- **Framework**: React 18 + Vite
- **UI Library**: Ant Design
- **Gráficos**: Recharts
- **Build otimizado**: Nginx + Multi-stage Docker build

### Infraestrutura
- **Orquestração**: Docker Swarm
- **Rede**: network_public (compatível com Traefik)
- **SSL/TLS**: Certificados automáticos via Let's Encrypt
- **Réplicas**: Configurável via variáveis de ambiente

## Módulos Implementados

1. **Gestão de Pacientes**: Cadastro completo, histórico, tags, anexos
2. **Agenda**: Agendamentos, confirmações, lista de espera
3. **Prontuário Eletrônico**: Odontograma, planos de tratamento, prescrições
4. **Financeiro**: Orçamentos, pagamentos, fluxo de caixa, comissões
5. **Estoque**: Produtos, fornecedores, movimentações, alertas
6. **Campanhas**: Marketing via WhatsApp/Email, segmentação
7. **Relatórios**: Dashboard, faturamento, procedimentos, attendance

## Instalação e Deploy

### Pré-requisitos

- Docker Engine 20.10+
- Docker Swarm inicializado
- Rede `network_public` criada
- Traefik configurado (para SSL automático)

### Passo 1: Configuração

```bash
# Clone o repositório ou descompacte os arquivos
cd drcrwell

# Copie o arquivo de exemplo e edite com suas configurações
cp .env.example .env
nano .env  # ou vim .env
```

**Configurações obrigatórias no .env:**

```bash
# URLs do seu domínio
FRONTEND_URL=dr.crwell.pro
BACKEND_URL=drapi.crwell.pro

# Usuário do Docker Hub
DOCKER_USERNAME=seuusuario

# Senhas (MUDE ESTAS!)
DB_PASSWORD=SuaSenhaSeguraAqui
JWT_SECRET=SeuSegredoJWTMuitoSeguroEAleatorioAqui

# Token do Docker Hub (opcional, para login automático)
DOCKER_TOKEN=dckr_pat_XXXXXXXXXXXX
```

### Passo 2: Deploy Automático

```bash
# Execute o script de deploy
./deploy.sh
```

O script irá:
1. Fazer login no Docker Hub
2. Build das imagens (backend e frontend)
3. Push para o Docker Hub
4. Deploy no Docker Swarm

### Passo 3: Verificação

```bash
# Verificar status dos serviços
docker stack ps drcrwell

# Ver logs
make logs-backend
make logs-frontend
make logs-db
```

### Acesso

- **Frontend**: https://dr.crwell.pro (ou sua URL configurada)
- **Backend API**: https://drapi.crwell.pro (ou sua URL configurada)
- **Health Check**: https://drapi.crwell.pro/health

## Deploy Manual

Se preferir fazer deploy manual:

```bash
# 1. Login no Docker Hub
docker login -u seuusuario

# 2. Build
make build

# 3. Push
make push

# 4. Deploy
make deploy
```

## Comandos Úteis

```bash
# Ver logs
make logs-backend    # Logs do backend
make logs-frontend   # Logs do frontend
make logs-db         # Logs do PostgreSQL

# Remover stack
make remove

# Recriar (atualizar)
make remove
make deploy
```

## Estrutura do Projeto

```
drcrwell/
├── backend/              # API em Go
│   ├── cmd/api/          # Entry point
│   ├── internal/         # Código interno
│   │   ├── models/       # Modelos do banco
│   │   ├── handlers/     # Controllers/Handlers
│   │   ├── middleware/   # JWT, Tenant, etc
│   │   ├── database/     # Conexão DB
│   │   └── services/     # Lógica de negócio
│   ├── Dockerfile
│   └── go.mod
├── frontend/             # App React
│   ├── src/
│   │   ├── pages/        # Páginas
│   │   ├── components/   # Componentes
│   │   ├── services/     # API calls
│   │   └── contexts/     # React Context
│   ├── Dockerfile
│   ├── nginx.conf
│   └── package.json
├── docker-stack.yml      # Stack do Swarm
├── .env.example          # Exemplo de configuração
├── Makefile              # Comandos úteis
├── deploy.sh             # Script de deploy automático
└── README.md             # Este arquivo
```

## Primeiro Acesso

1. Acesse a URL do frontend: https://dr.crwell.pro
2. Clique em "Cadastrar consultório"
3. Preencha os dados do consultório e do administrador
4. Faça login com as credenciais criadas

## Desenvolvimento Local

### Backend

```bash
cd backend

# Instalar dependências
go mod download

# Criar .env local
cp .env.example .env

# Rodar
go run cmd/api/main.go
```

### Frontend

```bash
cd frontend

# Instalar dependências
npm install

# Criar .env local
cp .env.example .env

# Rodar dev server
npm run dev
```

## Arquitetura Multitenant

O sistema usa **schema por tenant** no PostgreSQL:

- Schema `public`: Contém tabelas de tenants e usuários
- Schema `tenant_X`: Um schema por consultório, com todas as tabelas isoladas

Isso garante:
- **Isolamento total** dos dados
- **Performance** superior a row-level security
- **Facilidade** de backup por cliente
- **Escalabilidade** horizontal

## Segurança

- Autenticação via JWT com expiração de 24h
- Senhas com bcrypt (hash seguro)
- Isolamento completo entre tenants
- SQL injection protegido (GORM com prepared statements)
- CORS configurável
- HTTPS obrigatório em produção (via Traefik)

## Replicação para Outras Empresas

Este sistema é **totalmente replicável**. Para usar em outra empresa:

1. Clone/copie o projeto
2. Edite apenas o arquivo `.env`:
   - Mude `FRONTEND_URL` e `BACKEND_URL`
   - Defina novas senhas em `DB_PASSWORD` e `JWT_SECRET`
   - Configure seu `DOCKER_USERNAME`
3. Execute `./deploy.sh`
4. Pronto! Sistema rodando na nova infraestrutura

## Licença

Proprietário - Uso interno

## Suporte

Para dúvidas e suporte, entre em contato com o administrador do sistema.

---

**Desenvolvido com Go + React + Docker Swarm**
