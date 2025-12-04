# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Odowell** (formerly Dr. Crwell) is a multitenant SaaS dental clinic management system.

| Layer | Technology |
|-------|------------|
| Backend | Go 1.21 + Gin |
| Frontend | React 18 + Vite + Ant Design |
| Database | PostgreSQL 15 (schema-per-tenant) |
| Cache | Redis (optional) |
| Deployment | Docker Swarm + Traefik |

**Live URLs:** Frontend: `https://app.odowell.pro` | API: `https://api.odowell.pro`

## Commands

### Local Development

```bash
# Backend
cd backend && go run cmd/api/main.go

# Frontend
cd frontend && npm install && npm run dev
```

### Deployment

```bash
./deploy.sh                    # Full automated deploy
make build && make push && make deploy  # Manual steps
make remove                    # Remove stack
```

### Logs

```bash
make logs-backend
make logs-frontend
make logs-db
```

### Database Access

```bash
docker exec -it $(docker ps -q -f name=postgres) psql -U drcrwell_user -d drcrwell_db
\dn                    # List schemas
\dt tenant_1.*         # List tenant tables
```

## Architecture

### Multitenant: Schema-per-Tenant

Each tenant gets an isolated PostgreSQL schema (`tenant_X`).

**Public schema** (shared): `tenants`, `users`, `modules`, `permissions`, `user_permissions`

**Tenant schemas** (isolated): `patients`, `appointments`, `medical_records`, `prescriptions`, `exams`, `budgets`, `payments`, `products`, `suppliers`, `stock_movements`, `campaigns`, `attachments`, `tasks`, `settings`, `waiting_list`, `treatment_protocols`

**Schema switching**: `TenantMiddleware` executes `SET search_path TO tenant_X` based on JWT `tenant_id`.

### Backend Structure

- **Entry point**: `cmd/api/main.go` - All routes defined here
- **Handlers**: `internal/handlers/` - HTTP controllers
- **Models**: `internal/models/` - GORM models
- **Middleware chain**: `AuthMiddleware() → TenantMiddleware() → SubscriptionMiddleware() → PermissionMiddleware(module, action)`

### Frontend Structure

- **Entry point**: `src/main.jsx`
- **Routes**: `src/App.jsx`
- **Layout/Menu**: `src/components/layouts/DashboardLayout.jsx`
- **API client**: `src/services/api.js` - Axios with JWT interceptor
- **Auth/Permissions**: `src/contexts/AuthContext.jsx` - `usePermission()` hook

## RBAC (Role-Based Access Control)

**Two-layer enforcement:**

1. **Frontend** - `usePermission()` hook controls UI visibility:
   ```jsx
   const { canDelete } = usePermission();
   {canDelete('patients') && <Button>Delete</Button>}
   ```

2. **Backend** - `PermissionMiddleware` enforces on routes:
   ```go
   patients.DELETE("/:id",
     middleware.PermissionMiddleware("patients", "delete"),
     handlers.DeletePatient
   )
   ```

**Admin bypass**: Users with `role = 'admin'` skip all permission checks.

## Adding New Features

### Backend

1. Create model: `internal/models/feature.go`
2. Create handler: `internal/handlers/feature.go`
3. Register routes in `cmd/api/main.go` with appropriate middleware
4. Add module to `modules` table (if new module)

### Frontend

1. Add API methods: `src/services/api.js`
2. Create page: `src/pages/feature/Feature.jsx`
3. Add route: `src/App.jsx`
4. Add menu item: `src/components/layouts/DashboardLayout.jsx` (with `canView()`)

## Important Implementation Details

### Odontogram

Stored as JSONB in `medical_records.odontogram`:
```json
{
  "11": {"status": "healthy", "procedures": []},
  "12": {"status": "cavity", "procedures": ["restoration"]}
}
```

- **Frontend component**: `src/components/Odontogram.jsx` (interactive/read-only modes)
- **PDF generation**: `backend/internal/handlers/medical_record_pdf.go`

### GORM Limitations

For cross-schema operations, raw SQL is sometimes needed:
```go
db.Exec(`UPDATE medical_records SET ... WHERE id = ?`, ...)
```

### Migrations

Auto-migration runs on startup for tenant schemas. For public schema changes, add SQL migrations to `backend/migrations/`.

## Troubleshooting

**403 Forbidden**: Check `user_permissions` table, verify module code matches, confirm module is active in `modules` table.

**Frontend not updating**: Clear browser cache (Ctrl+Shift+R), check deploy logs.

## Environment Variables

**Backend** (`.env`): `DB_*`, `JWT_SECRET`, `CORS_ORIGINS`, `REDIS_*`, `STRIPE_*`

**Frontend**: `VITE_API_URL`

## Docker Images

- `tomautomations/drcrwell-backend:latest`
- `tomautomations/drcrwell-frontend:latest`
