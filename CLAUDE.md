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
docker exec -it $(docker ps -q -f name=postgres) psql -U odowell_app -d drcrwell_db
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

**Available modules** (from `public.modules` table): `dashboard`, `patients`, `appointments`, `medical_records`, `prescriptions`, `exams`, `budgets`, `payments`, `expenses`, `products`, `suppliers`, `stock_movements`, `campaigns`, `reports`, `users`, `settings`, `tasks`, `waiting_list`, `treatment_protocols`, `consents`, `treatments`, `plans`, `certificates`, `audit_logs`, `data_requests`

**Actions**: `view`, `create`, `edit`, `delete`

**User roles**: `admin` (full access), `dentist` (clinical focus), `receptionist` (front desk)

**Super Admin**: Users with `is_super_admin = true` have access to platform administration (Adm Empresas). This is the ONLY exclusive super admin feature.

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

## Digital Signature (ICP-Brasil A1)

### Overview
The system supports digital signature for prescriptions and medical records using ICP-Brasil A1 certificates (.pfx/.p12 files).

### Files
- **Backend**:
  - `internal/handlers/certificate.go` - Certificate upload, encryption, validation
  - `internal/handlers/document_signer.go` - Document signing logic
  - `internal/models/user_certificate.go` - Certificate model (stored in public schema)
- **Frontend**:
  - `src/pages/certificates/Certificates.jsx` - Certificate management page
  - `src/services/api.js` - `certificatesAPI` and `signingAPI` modules

### Security
- Certificates are encrypted with AES-256-GCM before storage
- Key derivation uses PBKDF2 with SHA-256 (100,000 iterations)
- Password is required at signing time (not stored)
- Signature uses RSA+SHA256

### Database
- `public.user_certificates` - Stores encrypted certificates (public schema, shared across tenants)
- `prescriptions.is_signed`, `signed_at`, `signature_hash`, etc. - Signature fields
- `medical_records.is_signed`, `signed_at`, `signature_hash`, etc. - Signature fields

### Flow
1. User uploads .pfx/.p12 certificate with password
2. System validates certificate, extracts metadata, encrypts with AES-256
3. User opens prescription/medical record, clicks "Assinar Digitalmente"
4. System prompts for certificate password
5. System decrypts certificate, generates SHA-256 hash of document, signs with RSA
6. Document is marked as signed (cannot be edited/deleted)

## AWS S3 Storage

### Configuration
- **Bucket**: `odowell-app`
- **Region**: `us-east-1` (N. Virginia)
- **IAM User**: `OdoWell-S3-user`

### File Structure
```
odowell-app/
├── backups/                          # Database backups (90-day retention)
│   └── odowell_backup_YYYYMMDD_HHMMSS.sql.gz
└── tenant_X_subdomain/               # Tenant files (no expiration)
    └── exams/
        └── [CPF]/
            └── [timestamp]_[filename]
```

### Environment Variables
```env
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx
AWS_BUCKET_NAME=odowell-app
AWS_REGION=us-east-1
```

## Backup System

### Automatic Backups
- **Schedule**: 2x/day (00:00 and 12:00 via cron)
- **Local retention**: 7 days
- **S3 retention**: 90 days (lifecycle rule)

### Scripts
- `scripts/backup.sh` - PostgreSQL backup + S3 upload
- `scripts/restore.sh` - Restore from backup
- `scripts/backup-redis.sh` - Redis backup

### Manual Commands
```bash
# Run backup manually
/root/drcrwell/scripts/backup.sh

# List available backups
ls -lh /root/drcrwell/backups/

# Restore from backup (creates safety backup first)
/root/drcrwell/scripts/restore.sh odowell_backup_YYYYMMDD_HHMMSS.sql.gz
```

### Logs
- `/root/drcrwell/backups/backup.log` - Backup execution logs
- `/root/drcrwell/backups/cron.log` - Cron job output

## Deployment & GitHub

### Docker Images
```bash
# Build and push backend
cd backend
docker build -t tomautomations/drcrwell-backend:latest .
docker push tomautomations/drcrwell-backend:latest
docker service update --image tomautomations/drcrwell-backend:latest drcrwell_backend --force

# Build and push frontend
cd frontend
docker build --build-arg VITE_API_URL=https://api.odowell.pro -t tomautomations/drcrwell-frontend:latest .
docker push tomautomations/drcrwell-frontend:latest
docker service update --image tomautomations/drcrwell-frontend:latest drcrwell_frontend --force
```

### GitHub Commit (Safe)
```bash
# Check for sensitive data before commit
git diff --name-only | xargs -I {} grep -l -E "(password|secret|token|api_key)" {} 2>/dev/null

# Verify .gitignore excludes sensitive files
cat .gitignore | grep -E "\.env|credentials|secrets"

# Commit and push
git add -A
git commit -m "feat: Description"
git push origin feature/rbac-permissions
```

### Important: Never commit
- `.env` files (contains DB passwords, JWT secrets, API keys)
- Any file with hardcoded passwords or tokens
- Backup files (`*.backup-*`)

### Sensitive Data Locations (NOT in git)
- Backend `.env`: Database credentials, JWT_SECRET, STRIPE_*, AWS_*, SMTP_*
- These are configured via Docker secrets or environment variables in production
