# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Dr. Crwell is a multitenant SaaS system for dental clinic management, built with Go (backend) and React (frontend), designed to run on Docker Swarm.

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

# Run tests (when implemented)
go test ./...

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

# Run dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Architecture

### Backend (Go + Gin)

- **Entry Point**: `cmd/api/main.go`
- **Structure**:
  - `internal/handlers/`: HTTP request handlers (controllers)
  - `internal/models/`: Database models (GORM)
  - `internal/middleware/`: Auth (JWT) and tenant isolation
  - `internal/database/`: Database connection and schema management
  - `internal/services/`: Business logic (if complex)

### Multitenant Architecture

- **Strategy**: Schema per tenant in PostgreSQL
- **Public Schema**: Contains `tenants` and `users` tables
- **Tenant Schemas**: Each tenant gets `tenant_X` schema with all app tables
- **Isolation**: Middleware sets schema context based on JWT token
- **Schema Switch**: `SET search_path TO tenant_X` per request

### Frontend (React + Vite)

- **Entry Point**: `src/main.jsx`
- **Structure**:
  - `src/pages/`: Page components (one per route)
  - `src/components/`: Reusable UI components
  - `src/services/api.js`: Axios API client with interceptors
  - `src/contexts/`: React Context (Auth, etc.)
- **UI Library**: Ant Design
- **Routing**: React Router v6
- **State**: React Context + hooks (no Redux)

## Database Schema

### Public Schema Tables

- `tenants`: Clinic/tenant metadata
- `users`: System users (linked to tenant_id)

### Tenant Schema Tables (in each `tenant_X` schema)

- `patients`: Patient records
- `appointments`: Scheduled appointments
- `medical_records`: Clinical records, odontogram
- `budgets`: Treatment quotes
- `payments`: Financial transactions
- `products`: Inventory items
- `suppliers`: Product suppliers
- `stock_movements`: Inventory movements
- `campaigns`: Marketing campaigns
- `campaign_recipients`: Campaign recipients
- `attachments`: Patient files (photos, exams)

## Key Implementation Details

### Authentication Flow

1. User logs in via `/api/auth/login`
2. Backend validates credentials, returns JWT with `tenant_id`
3. Frontend stores JWT in localStorage
4. Subsequent requests include JWT in `Authorization: Bearer` header
5. Auth middleware extracts `tenant_id` from JWT
6. Tenant middleware sets database schema to `tenant_X`

### Adding New Endpoints

1. Define model in `internal/models/` (if new entity)
2. Create handler in `internal/handlers/`
3. Register route in `cmd/api/main.go`
4. Add API call in `frontend/src/services/api.js`
5. Create page/component in `frontend/src/pages/`

### Docker Build Process

Backend uses multi-stage build:
1. Build stage: Go 1.21-alpine, compiles binary
2. Runtime stage: Alpine, only binary (~15MB image)

Frontend uses multi-stage build:
1. Build stage: Node 18, runs `npm run build`
2. Runtime stage: Nginx, serves static files

## Important Files

- `.env`: Production configuration (NOT in git)
- `.env.example`: Template for configuration
- `docker-stack.yml`: Swarm stack definition
- `Makefile`: Build and deploy commands
- `deploy.sh`: Automated deployment script

## Configuration

All configuration via environment variables (12-factor app):

**Backend**:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_SECRET`: MUST be changed for production
- `PORT`: API server port (default: 8080)
- `CORS_ORIGINS`: Allowed frontend origins

**Frontend**:
- `VITE_API_URL`: Backend API base URL

## Deployment Architecture

- **Orchestrator**: Docker Swarm
- **Network**: `network_public` (external, shared with Traefik)
- **Reverse Proxy**: Traefik (handles SSL/TLS via Let's Encrypt)
- **Services**:
  - `drcrwell_postgres`: PostgreSQL 15
  - `drcrwell_backend`: Go API (1+ replicas)
  - `drcrwell_frontend`: Nginx + React SPA (1+ replicas)

## Troubleshooting

### Backend won't start
- Check logs: `make logs-backend`
- Verify DB connection env vars
- Ensure PostgreSQL is running and accessible
- Check JWT_SECRET is set

### Frontend shows blank page
- Check browser console for errors
- Verify `VITE_API_URL` is correct
- Check CORS settings in backend

### Database migration errors
- Manually connect: `docker exec -it <postgres-container> psql -U drcrwell_user -d drcrwell_db`
- Check schemas: `\dn`
- Verify tables: `\dt public.*` or `\dt tenant_1.*`

## Testing

To test the system end-to-end:

1. Access frontend URL (e.g., https://dr.crwell.pro)
2. Click "Cadastrar consultório" to create a tenant
3. Fill in clinic and admin info
4. Login with created credentials
5. Test CRUD operations:
   - Create a patient
   - Create an appointment
   - View dashboard statistics

## Security Considerations

- All passwords are bcrypt hashed (cost 14)
- JWT tokens expire after 24 hours
- Database credentials should be strong and unique
- HTTPS is mandatory in production (handled by Traefik)
- Each tenant's data is isolated in separate schemas
- SQL injection is prevented via GORM prepared statements

## Performance

- Go backend: ~10-20MB RAM per instance
- Scales horizontally via Docker Swarm replicas
- Database connection pooling via GORM
- Frontend: Static files served by Nginx (fast)
- API responses are typically < 100ms

## Code Style

**Go**: Follow standard Go conventions
- Use `gofmt` for formatting
- Error handling: always check and handle errors
- Naming: camelCase for private, PascalCase for exported

**JavaScript/React**:
- ES6+ features
- Functional components + hooks (no class components)
- Destructuring where appropriate
- Async/await for promises

## Future Enhancements

Areas to expand (not yet implemented in full):
- WhatsApp Business API integration for campaigns
- Email sending service integration
- Advanced reporting with chart visualizations
- Role-based access control (RBAC) within tenants
- Audit logging
- File storage in S3-compatible service
- Automated backups
- Health check endpoints with metrics
