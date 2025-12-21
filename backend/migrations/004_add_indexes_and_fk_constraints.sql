-- Migration: 004_add_indexes_and_fk_constraints.sql
-- Description: Add performance indexes and foreign key constraints
-- Author: Claude Code
-- Date: 2025-12-21

-- ============================================================================
-- PUBLIC SCHEMA INDEXES
-- ============================================================================

-- Audit Logs - frequently queried by date range and user
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON public.audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action ON public.audit_logs(user_id, action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON public.audit_logs(resource, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_success ON public.audit_logs(success) WHERE success = false;

-- Users - login queries
CREATE INDEX IF NOT EXISTS idx_users_email_active ON public.users(email, active);
CREATE INDEX IF NOT EXISTS idx_users_tenant_role ON public.users(tenant_id, role);

-- Tenants - API key lookup and active filter
CREATE INDEX IF NOT EXISTS idx_tenants_active ON public.tenants(active) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_tenants_subscription_status ON public.tenants(subscription_status);
CREATE INDEX IF NOT EXISTS idx_tenants_expires_at ON public.tenants(expires_at) WHERE expires_at IS NOT NULL;

-- Password Resets - cleanup queries
CREATE INDEX IF NOT EXISTS idx_password_resets_expires_at ON public.password_resets(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_resets_used ON public.password_resets(used);

-- Email Verifications - cleanup queries
CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at ON public.email_verifications(expires_at);
CREATE INDEX IF NOT EXISTS idx_email_verifications_verified ON public.email_verifications(verified);

-- User Permissions - permission checks
CREATE INDEX IF NOT EXISTS idx_user_permissions_user_module ON public.user_permissions(user_id, module_id);

-- User Certificates - certificate lookup
CREATE INDEX IF NOT EXISTS idx_user_certificates_user_active ON public.user_certificates(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_user_certificates_expires_at ON public.user_certificates(expires_at);

-- ============================================================================
-- PUBLIC SCHEMA FOREIGN KEYS
-- ============================================================================

-- Users -> Tenants
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_tenant') THEN
        ALTER TABLE public.users
        ADD CONSTRAINT fk_users_tenant
        FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- User Permissions -> Users
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_permissions_user') THEN
        ALTER TABLE public.user_permissions
        ADD CONSTRAINT fk_user_permissions_user
        FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- User Permissions -> Modules
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_permissions_module') THEN
        ALTER TABLE public.user_permissions
        ADD CONSTRAINT fk_user_permissions_module
        FOREIGN KEY (module_id) REFERENCES public.modules(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Permissions -> Modules
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_permissions_module') THEN
        ALTER TABLE public.permissions
        ADD CONSTRAINT fk_permissions_module
        FOREIGN KEY (module_id) REFERENCES public.modules(id) ON DELETE CASCADE;
    END IF;
END $$;

-- User Certificates -> Users
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_certificates_user') THEN
        ALTER TABLE public.user_certificates
        ADD CONSTRAINT fk_user_certificates_user
        FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Tenant Settings -> Tenants
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_tenant_settings_tenant') THEN
        ALTER TABLE public.tenant_settings
        ADD CONSTRAINT fk_tenant_settings_tenant
        FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Password Resets -> Users
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_password_resets_user') THEN
        ALTER TABLE public.password_resets
        ADD CONSTRAINT fk_password_resets_user
        FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Email Verifications -> Users
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_email_verifications_user') THEN
        ALTER TABLE public.email_verifications
        ADD CONSTRAINT fk_email_verifications_user
        FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;
    END IF;
END $$;
