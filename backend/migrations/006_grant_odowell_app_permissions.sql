-- Migration: 006_grant_odowell_app_permissions.sql
-- Description: Grant permissions to odowell_app user for all schemas
-- This ensures the backend can access tables regardless of which user owns them
-- Author: Claude Code
-- Date: 2025-12-23

-- ============================================================================
-- PUBLIC SCHEMA PERMISSIONS
-- ============================================================================

-- Grant usage on public schema
GRANT USAGE ON SCHEMA public TO odowell_app;

-- Grant all privileges on existing tables
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO odowell_app;

-- Grant all privileges on existing sequences
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO odowell_app;

-- Set default privileges for future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO odowell_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO odowell_app;

-- ============================================================================
-- TENANT SCHEMAS PERMISSIONS (applied dynamically)
-- ============================================================================

-- This will be applied per-tenant via Go code in ApplyTenantPermissions()
