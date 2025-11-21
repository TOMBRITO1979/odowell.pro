-- Migration: Create permissions system tables
-- Date: 2025-11-21
-- Description: Add modules, permissions, and user_permissions tables for RBAC

-- =============================================================================
-- 1. CREATE MODULES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS public.modules (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    active BOOLEAN DEFAULT true
);

CREATE INDEX idx_modules_code ON public.modules(code);
CREATE INDEX idx_modules_deleted_at ON public.modules(deleted_at);

-- =============================================================================
-- 2. CREATE PERMISSIONS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS public.permissions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    module_id INTEGER NOT NULL REFERENCES public.modules(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL CHECK (action IN ('view', 'create', 'edit', 'delete')),
    description VARCHAR(200),

    UNIQUE(module_id, action)
);

CREATE INDEX idx_permissions_module_id ON public.permissions(module_id);
CREATE INDEX idx_permissions_deleted_at ON public.permissions(deleted_at);

-- =============================================================================
-- 3. CREATE USER_PERMISSIONS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS public.user_permissions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    user_id INTEGER NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES public.permissions(id) ON DELETE CASCADE,
    granted_by INTEGER REFERENCES public.users(id),
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, permission_id)
);

CREATE INDEX idx_user_permissions_user_id ON public.user_permissions(user_id);
CREATE INDEX idx_user_permissions_permission_id ON public.user_permissions(permission_id);
CREATE INDEX idx_user_permissions_deleted_at ON public.user_permissions(deleted_at);

-- =============================================================================
-- COMMENTS
-- =============================================================================
COMMENT ON TABLE public.modules IS 'System modules that can have permissions assigned';
COMMENT ON TABLE public.permissions IS 'Actions that can be performed on modules';
COMMENT ON TABLE public.user_permissions IS 'Permissions granted to users';
