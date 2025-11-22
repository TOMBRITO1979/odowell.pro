# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Dr. Crwell is a multitenant SaaS dental clinic management system with Go backend, React frontend, and Docker Swarm deployment. Features full RBAC (Role-Based Access Control) for granular user permissions.

## Common Development Commands

### Building and Deploying

```bash
# Full deployment (builds, pushes, deploys)
./deploy.sh

# Individual steps
make build         # Build Docker images
make push          # Push to Docker Hub
make deploy        # Deploy to Swarm
make remove        # Remove stack

# View logs
make logs-backend
make logs-frontend
make logs-db
```

### Backend Development

```bash
cd backend

# Run locally (requires PostgreSQL running)
go run cmd/api/main.go

# Format code
go fmt ./...

# Update dependencies
go mod tidy
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Run dev server (port 5173)
npm run dev

# Build for production
npm run build
```

## Architecture

### Multitenant Strategy: Schema-per-Tenant

**Critical concept**: Each tenant gets an isolated PostgreSQL schema (`tenant_X`).

- **Public schema**: Contains `tenants`, `users`, `modules`, `permissions`, `user_permissions` tables
- **Tenant schemas** (e.g., `tenant_1`, `tenant_2`): All business data isolated per clinic
  - `patients`, `appointments`, `medical_records`, `prescriptions`, `exams`
  - `budgets`, `payments`, `products`, `suppliers`, `stock_movements`
  - `campaigns`, `campaign_recipients`, `attachments`, `tasks`, `settings`

**Schema switching**: Middleware executes `SET search_path TO tenant_X` per request based on JWT `tenant_id`.

### Backend Structure (Go + Gin)

- **Entry point**: `cmd/api/main.go` - Route registration and server setup
- **Handlers** (`internal/handlers/`): HTTP controllers for each resource
- **Models** (`internal/models/`): GORM database models
- **Middleware** (`internal/middleware/`):
  - `AuthMiddleware()`: JWT validation, extracts `user_id`, `tenant_id`, `user_role`
  - `TenantMiddleware()`: Sets database schema to `tenant_X`
  - `PermissionMiddleware(module, action)`: Enforces RBAC permissions
- **Database** (`internal/database/`): Connection, auto-migration, schema management

### Frontend Structure (React + Vite)

- **Entry point**: `src/main.jsx`
- **Pages** (`src/pages/`): One component per route
- **Components** (`src/components/`): Reusable UI (Ant Design based)
- **Contexts** (`src/contexts/AuthContext.jsx`): Auth state, permissions, user data
- **Hooks**: `usePermission()` exported from AuthContext for permission checks
- **API Client** (`src/services/api.js`): Axios instance with JWT interceptor

## RBAC (Role-Based Access Control)

### How RBAC Works

**Two-layer security**: Frontend (UX) + Backend (enforcement)

1. **Frontend layer**: Uses `usePermission()` hook to hide/disable UI elements
   ```jsx
   const { canDelete, canEdit, canCreate } = usePermission();

   {canDelete('patients') && <Button onClick={handleDelete}>Delete</Button>}
   ```

2. **Backend layer**: `PermissionMiddleware` validates every protected route
   ```go
   patients.DELETE("/:id",
     middleware.PermissionMiddleware("patients", "delete"),
     handlers.DeletePatient
   )
   ```

### Permission Flow

1. User logs in → JWT includes `permissions` object (all user permissions)
2. Frontend decodes JWT, stores permissions in AuthContext
3. UI conditionally renders based on `canView()`, `canCreate()`, `canEdit()`, `canDelete()`
4. API requests hit backend → `PermissionMiddleware` queries `user_permissions` table
5. If no permission → 403 Forbidden; if admin role → bypass (superuser)

### RBAC Tables (in public schema)

- `modules`: Available system modules (patients, appointments, budgets, etc.)
- `permissions`: Actions per module (view, create, edit, delete)
- `user_permissions`: Junction table linking users to permissions

**Admin bypass**: Users with `role = 'admin'` skip all permission checks (line 47 in `middleware/permission.go`).

### Adding RBAC to New Endpoints

1. Register route with `PermissionMiddleware(module, action)` in `cmd/api/main.go`
2. Use `usePermission()` hook in frontend to conditionally render UI
3. Module code must match between frontend/backend (e.g., `"patients"`, `"budgets"`)

## Adding New Features

### Backend

1. Create model in `internal/models/` (e.g., `task.go`)
2. Add handler in `internal/handlers/` (e.g., `task.go`)
3. Register routes in `cmd/api/main.go`:
   ```go
   tasks := tenanted.Group("/tasks")
   tasks.POST("", middleware.PermissionMiddleware("tasks", "create"), handlers.CreateTask)
   tasks.GET("", middleware.PermissionMiddleware("tasks", "view"), handlers.GetTasks)
   // etc.
   ```
4. If new module: Add entry to `modules` table in public schema

### Frontend

1. Create API methods in `src/services/api.js`
2. Create page component in `src/pages/` (e.g., `src/pages/tasks/Tasks.jsx`)
3. Add route in `src/App.jsx`
4. Use `usePermission()` to guard UI elements
5. Add menu item in `src/components/layouts/DashboardLayout.jsx` (guarded by `canView()`)

## Authentication & JWT

**Login flow**:
1. POST `/api/auth/login` with `{email, password}`
2. Backend validates credentials, generates JWT with claims:
   - `user_id`, `tenant_id`, `email`, `role`, `permissions` (full permission map)
3. Frontend stores JWT in localStorage, decodes permissions into AuthContext
4. All subsequent requests include `Authorization: Bearer <token>` header
5. JWT expires after 24 hours

**Profile picture upload**:
- Users can upload profile pictures (JPEG/PNG, max 5MB)
- Endpoint: `POST /api/auth/profile/picture`
- Files stored in `/root/uploads/profile-pictures/` with unique names
- Served via `/uploads` static route
- Old pictures automatically deleted on new upload
- Avatar displays profile picture in header and profile page

**Middleware chain for protected routes**:
```
Request → AuthMiddleware → TenantMiddleware → PermissionMiddleware → Handler
```

## Configuration

**Environment variables** (see `.env.example`):

Backend:
- `DB_*`: PostgreSQL connection
- `JWT_SECRET`: MUST be changed for production (256+ bits)
- `CORS_ORIGINS`: Comma-separated allowed origins
- `AWS_*`: S3-compatible storage for attachments (optional)

Frontend:
- `VITE_API_URL`: Backend API base URL

## Deployment

- **Orchestrator**: Docker Swarm
- **Network**: `network_public` (external, shared with Traefik reverse proxy)
- **Traefik**: Handles SSL/TLS via Let's Encrypt, routing to backend/frontend
- **Services**:
  - `drcrwell_postgres`: PostgreSQL 15 (1 replica, manager node)
  - `drcrwell_backend`: Go API (configurable replicas via `BACKEND_REPLICAS`)
  - `drcrwell_frontend`: Nginx + React SPA (configurable via `FRONTEND_REPLICAS`)

**Deployment labels** in `docker-stack.yml` configure Traefik routing via `FRONTEND_URL` and `BACKEND_URL`.

## Troubleshooting

**Backend won't start**:
- `make logs-backend` to check errors
- Verify `DB_*` env vars and PostgreSQL connectivity
- Ensure `JWT_SECRET` is set

**Permission errors (403 Forbidden)**:
- Check user has `user_permissions` entries for module/action
- Verify module `code` matches between frontend/backend
- Confirm `modules` table has entry for the module (and `active = true`)
- Admin users bypass all checks (verify `role = 'admin'`)

**Database schema issues**:
```bash
docker exec -it <postgres-container> psql -U drcrwell_user -d drcrwell_db
\dn                    # List schemas
\dt public.*           # List public tables
\dt tenant_1.*         # List tenant schema tables
```

## Docker Build

Both backend and frontend use multi-stage builds:

**Backend**: `golang:1.21-alpine` (build) → `alpine:latest` (runtime, ~15MB)
**Frontend**: `node:18` (build with Vite) → `nginx:alpine` (serve static files)

Images are tagged as `${DOCKER_USERNAME}/drcrwell-backend:latest` and `${DOCKER_USERNAME}/drcrwell-frontend:latest`.
