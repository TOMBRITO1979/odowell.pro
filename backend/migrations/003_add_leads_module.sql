-- Migration: Add Leads module for CRM
-- Date: 2025-12-08
-- Description: Add leads module and permissions for WhatsApp integration

-- =============================================================================
-- 1. INSERT LEADS MODULE
-- =============================================================================
INSERT INTO public.modules (code, name, description, icon, active) VALUES
('leads', 'Leads', 'Gestão de leads e CRM para WhatsApp', 'UsergroupAddOutlined', true)
ON CONFLICT (code) DO NOTHING;

-- =============================================================================
-- 2. CREATE PERMISSIONS FOR LEADS MODULE
-- =============================================================================
DO $$
DECLARE
    module_id_val INTEGER;
    actions TEXT[] := ARRAY['view', 'create', 'edit', 'delete'];
    action TEXT;
    action_descriptions JSONB := '{
        "view": "Visualizar",
        "create": "Criar novos registros",
        "edit": "Editar registros existentes",
        "delete": "Deletar registros"
    }'::JSONB;
BEGIN
    -- Get the leads module ID
    SELECT id INTO module_id_val FROM public.modules WHERE code = 'leads';

    IF module_id_val IS NOT NULL THEN
        -- For each action
        FOREACH action IN ARRAY actions LOOP
            INSERT INTO public.permissions (module_id, action, description)
            VALUES (
                module_id_val,
                action,
                action_descriptions->>action || ' em Leads'
            )
            ON CONFLICT (module_id, action) DO NOTHING;
        END LOOP;

        RAISE NOTICE 'Created permissions for leads module (ID: %)', module_id_val;
    END IF;
END $$;

-- =============================================================================
-- 3. GRANT LEADS PERMISSIONS TO EXISTING ADMIN USERS
-- =============================================================================
DO $$
DECLARE
    admin_user RECORD;
    permission_record RECORD;
BEGIN
    -- For each admin user
    FOR admin_user IN SELECT id FROM public.users WHERE role = 'admin' AND deleted_at IS NULL LOOP
        -- Grant leads permissions
        FOR permission_record IN
            SELECT p.id
            FROM public.permissions p
            JOIN public.modules m ON m.id = p.module_id
            WHERE m.code = 'leads' AND p.deleted_at IS NULL
        LOOP
            INSERT INTO public.user_permissions (user_id, permission_id, granted_by, granted_at)
            VALUES (
                admin_user.id,
                permission_record.id,
                admin_user.id,
                CURRENT_TIMESTAMP
            )
            ON CONFLICT (user_id, permission_id) DO NOTHING;
        END LOOP;

        RAISE NOTICE 'Granted leads permissions to admin user ID: %', admin_user.id;
    END LOOP;
END $$;

-- =============================================================================
-- 4. UPDATE DEFAULT ROLE PERMISSIONS VIEW
-- =============================================================================
-- Receptionists should have full access to leads (they handle incoming inquiries)
-- Dentists should have view access

CREATE OR REPLACE VIEW public.default_role_permissions AS
SELECT
    'admin' as role,
    m.code as module_code,
    p.action,
    true as has_permission
FROM public.modules m
CROSS JOIN (VALUES ('view'), ('create'), ('edit'), ('delete')) AS actions(action)
LEFT JOIN public.permissions p ON p.module_id = m.id AND p.action = actions.action

UNION ALL

SELECT
    'dentist' as role,
    m.code as module_code,
    p.action,
    CASE
        -- Dentists have full access to clinical modules
        WHEN m.code IN ('patients', 'appointments', 'medical_records', 'prescriptions', 'exams') THEN true
        -- View only for financial, inventory, and leads
        WHEN m.code IN ('budgets', 'payments', 'products', 'suppliers', 'stock_movements', 'leads') AND p.action = 'view' THEN true
        -- View dashboard and reports
        WHEN m.code IN ('dashboard', 'reports') AND p.action = 'view' THEN true
        ELSE false
    END as has_permission
FROM public.modules m
CROSS JOIN (VALUES ('view'), ('create'), ('edit'), ('delete')) AS actions(action)
LEFT JOIN public.permissions p ON p.module_id = m.id AND p.action = actions.action

UNION ALL

SELECT
    'receptionist' as role,
    m.code as module_code,
    p.action,
    CASE
        -- Receptionists manage appointments, patients, and leads
        WHEN m.code IN ('patients', 'appointments', 'leads') THEN true
        -- View only for clinical records
        WHEN m.code IN ('medical_records', 'exams') AND p.action = 'view' THEN true
        -- Full access to financial
        WHEN m.code IN ('budgets', 'payments') THEN true
        -- View dashboard
        WHEN m.code = 'dashboard' AND p.action = 'view' THEN true
        ELSE false
    END as has_permission
FROM public.modules m
CROSS JOIN (VALUES ('view'), ('create'), ('edit'), ('delete')) AS actions(action)
LEFT JOIN public.permissions p ON p.module_id = m.id AND p.action = actions.action;

COMMENT ON VIEW public.default_role_permissions IS 'Default permission templates for each role (reference only)';
