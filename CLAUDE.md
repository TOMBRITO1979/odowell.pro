# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

**Odowell** (formerly Dr. Crwell) is a multitenant SaaS dental clinic management system with:
- **Backend**: Go + Gin framework
- **Frontend**: React + Vite + Ant Design
- **Database**: PostgreSQL 15 (schema-per-tenant isolation)
- **Deployment**: Docker Swarm + Traefik (SSL/TLS)
- **Security**: Full RBAC (Role-Based Access Control)

**Live URLs:**
- Frontend: https://app.odowell.pro
- Backend API: https://api.odowell.pro

## Quick Commands

```bash
# Full deployment
./deploy.sh

# Build & deploy separately
make build && make push && make deploy

# View logs
make logs-backend
make logs-frontend
make logs-db

# Remove stack
make remove
```

## Architecture

### Multitenant: Schema-per-Tenant

Each tenant gets an isolated PostgreSQL schema (`tenant_X`):

**Public schema** (shared):
- `tenants`, `users`, `modules`, `permissions`, `user_permissions`

**Tenant schemas** (isolated per clinic):
- `patients`, `appointments`, `medical_records`, `prescriptions`, `exams`
- `budgets`, `payments`, `products`, `suppliers`, `stock_movements`
- `campaigns`, `attachments`, `tasks`, `settings`, `waiting_list`, `treatment_protocols`

**Schema switching**: Middleware executes `SET search_path TO tenant_X` based on JWT `tenant_id`.

### Backend (Go + Gin)

- **Entry point**: `cmd/api/main.go`
- **Handlers**: `internal/handlers/` - HTTP controllers
- **Models**: `internal/models/` - GORM models
- **Middleware**:
  - `AuthMiddleware()`: JWT validation
  - `TenantMiddleware()`: Schema switching
  - `PermissionMiddleware(module, action)`: RBAC enforcement

### Frontend (React + Vite)

- **Entry point**: `src/main.jsx`
- **Pages**: `src/pages/` - Route components
- **Components**: `src/components/` - Reusable UI (Ant Design)
- **API**: `src/services/api.js` - Axios with JWT interceptor
- **Auth**: `usePermission()` hook from AuthContext

## RBAC (Role-Based Access Control)

**Two-layer security**:

1. **Frontend**: `usePermission()` hook hides/disables UI
   ```jsx
   const { canDelete } = usePermission();
   {canDelete('patients') && <Button>Delete</Button>}
   ```

2. **Backend**: `PermissionMiddleware` enforces on routes
   ```go
   patients.DELETE("/:id",
     middleware.PermissionMiddleware("patients", "delete"),
     handlers.DeletePatient
   )
   ```

**Admin bypass**: Users with `role = 'admin'` skip all permission checks.

## Adding New Features

### Backend Steps

1. Create model: `internal/models/feature.go`
2. Create handler: `internal/handlers/feature.go`
3. Register routes in `cmd/api/main.go` with `PermissionMiddleware`
4. Add module to `modules` table (if new module)

### Frontend Steps

1. Add API methods: `src/services/api.js`
2. Create page: `src/pages/feature/Feature.jsx`
3. Add route: `src/App.jsx`
4. Add menu item: `src/components/layouts/DashboardLayout.jsx` (with `canView()`)

## Key Features Implemented

### Gestão de Pacientes
- Cadastro completo com histórico
- **Prontuário eletrônico com odontograma digital** (visual component)
- Armazenamento de exames e documentos
- Controle de alergias e medicações

### Agendamento
- Agenda por profissional/cadeira
- **Lista de espera** (waiting list)
- Controle de faltas e reagendamentos
- Bloqueio de horários

### Financeiro
- Orçamentos e planos de tratamento
- Contas a receber/pagar
- Gestão de convênios
- Relatórios e fluxo de caixa

### Gestão Clínica
- Registro de procedimentos
- Controle de estoque
- **Protocolos de atendimento** (treatment protocols)
- Prescrições e atestados

## Important Notes

### Odontogram Implementation

The odontogram is stored as JSONB in `medical_records.odontogram`:
```json
{
  "11": {"status": "healthy", "procedures": []},
  "12": {"status": "cavity", "procedures": ["restoration"]}
}
```

**Frontend component**: `src/components/Odontogram.jsx`
- Interactive mode (edit)
- Read-only mode (display)
- Responsive design with multiple breakpoints

**PDF generation**: `backend/internal/handlers/medical_record_pdf.go`
- Renders odontogram in PDF with quadrants and symbols

### CRUD Operations with GORM Issues

For cross-schema operations, sometimes raw SQL is needed instead of GORM:
```go
db.Exec(`UPDATE medical_records SET ... WHERE id = ?`, ...)
```

### Database Migrations

Auto-migration runs on startup for tenant schemas. For public schema changes, use SQL migrations in `backend/migrations/`.

## Troubleshooting

**403 Forbidden errors**:
- Check `user_permissions` table
- Verify module code matches frontend/backend
- Confirm module is active in `modules` table

**Database issues**:
```bash
docker exec -it $(docker ps -q -f name=postgres) psql -U drcrwell_user -d drcrwell_db
\dn                    # List schemas
\dt tenant_1.*         # List tables
```

**Frontend not updating**:
- Clear browser cache: Ctrl+Shift+R
- Check deploy logs: `make logs-frontend`

## Environment Variables

**Backend** (`.env`):
- `DB_*`: PostgreSQL connection
- `JWT_SECRET`: 256+ bits for production
- `CORS_ORIGINS`: Comma-separated allowed origins

**Frontend**:
- `VITE_API_URL`: Backend API URL (https://api.odowell.pro)

## Docker Images

- `tomautomations/drcrwell-backend:latest`
- `tomautomations/drcrwell-frontend:latest`

Multi-stage builds:
- Backend: `golang:1.21-alpine` → `alpine:latest` (~15MB)
- Frontend: `node:18-alpine` → `nginx:alpine`
