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

### Build & Test

```bash
# Backend - verify compilation
cd backend && go build ./...

# Frontend - build for production
cd frontend && npm run build
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

**Tenant schemas** (isolated): `patients`, `appointments`, `medical_records`, `prescriptions`, `exams`, `budgets`, `payments`, `expenses`, `products`, `suppliers`, `stock_movements`, `campaigns`, `attachments`, `tasks`, `settings`, `waiting_list`, `treatment_protocols`, `consent_templates`, `patient_consents`, `patient_subscriptions`, `audit_logs`

**Schema switching**: `TenantMiddleware` executes `SET search_path TO tenant_X` based on JWT `tenant_id`.

### Backend Structure

- **Entry point**: `cmd/api/main.go` - All routes defined here (~536 lines)
- **Handlers**: `internal/handlers/` - HTTP controllers (44 files)
- **Models**: `internal/models/` - GORM models (24 files)
- **Middleware**: `internal/middleware/` - Auth, Tenant, Permission, RateLimit, Subscription
- **Helpers**: `internal/helpers/` - Email, crypto, validation, audit
- **Middleware chain**: `AuthMiddleware() → TenantMiddleware() → SubscriptionMiddleware() → PermissionMiddleware(module, action)`

### Key Backend Patterns

- Business logic is in handlers (no separate service layer)
- GORM with PostgreSQL, soft deletes via `deleted_at`
- JSONB fields for flexible data: `odontogram`, `settings`, `procedures`
- S3 for file uploads (exams, attachments)
- Stripe webhooks for subscription billing

### Frontend Structure

- **Entry point**: `src/main.jsx`
- **Routes**: `src/App.jsx`
- **Layout/Menu**: `src/components/layouts/DashboardLayout.jsx` (~422 lines, permission-filtered menu)
- **API client**: `src/services/api.js` - Axios with JWT interceptor (20+ API modules)
- **Auth/Permissions**: `src/contexts/AuthContext.jsx` - `usePermission()` hook
- **Pages**: `src/pages/` - Organized by domain (patients, appointments, financial, etc.)
- **Locale**: Portuguese (pt_BR) via Ant Design ConfigProvider

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

**Available modules** (from `002_seed_permissions.sql`): `dashboard`, `patients`, `appointments`, `medical_records`, `prescriptions`, `exams`, `budgets`, `payments`, `expenses`, `products`, `suppliers`, `stock_movements`, `campaigns`, `reports`, `users`, `settings`, `tasks`, `waiting_list`, `treatment_protocols`, `consents`, `treatments`, `plans`

**Actions**: `view`, `create`, `edit`, `delete`

**User roles**: `admin` (full access), `dentist` (clinical focus), `receptionist` (front desk)

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
4. Add menu item: `src/components/layouts/DashboardLayout.jsx` (wrap with `canView('module_code')`)

### Adding New Permission Module

1. Insert into `modules` table: `INSERT INTO public.modules (code, name, description, icon, active) VALUES ('new_module', 'Display Name', 'Description', 'IconOutlined', true);`
2. Permissions auto-created via migration script or manually add to `permissions` table
3. Assign to users via `user_permissions` table

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

**Frontend not updating**: Clear browser cache (Ctrl+Shift+R), check deploy logs. Note: Service Worker is disabled to prevent caching issues.

**402 Payment Required**: Tenant subscription expired. Check `tenants.subscription_status` and `tenants.expires_at`.

**Tenant context issues**: Ensure request has valid JWT with `tenant_id` claim.

## Environment Variables

**Backend** (`.env`): `DB_*`, `JWT_SECRET`, `CORS_ORIGINS`, `REDIS_*`, `STRIPE_*`, `AWS_*` (S3), `SMTP_*`

**Frontend**: `VITE_API_URL`

See `.env.example` for complete list with descriptions.

## Docker Images

- `tomautomations/drcrwell-backend:latest`
- `tomautomations/drcrwell-frontend:latest`

## API Response Patterns

- JWT auth via `Authorization: Bearer <token>` header
- 401 triggers automatic logout on frontend
- 402 redirects to subscription page
- Axios interceptors handle token refresh in `src/services/api.js`

## Language

- All user-facing strings are in Portuguese (pt_BR)
- Error messages, labels, and notifications are hardcoded Portuguese
