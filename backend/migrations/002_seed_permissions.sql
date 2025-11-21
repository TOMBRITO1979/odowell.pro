-- Migration: Seed default modules and permissions
-- Date: 2025-11-21
-- Description: Populate modules and permissions tables with default data

-- =============================================================================
-- 1. INSERT MODULES
-- =============================================================================
INSERT INTO public.modules (code, name, description, icon, active) VALUES
('dashboard', 'Dashboard', 'Painel principal com métricas e visão geral', 'DashboardOutlined', true),
('patients', 'Pacientes', 'Gestão de pacientes', 'UserOutlined', true),
('appointments', 'Agenda', 'Gestão de agendamentos e consultas', 'CalendarOutlined', true),
('medical_records', 'Prontuários', 'Prontuários médicos e odontológicos', 'MedicineBoxOutlined', true),
('prescriptions', 'Receituário', 'Prescrições e receitas médicas', 'FormOutlined', true),
('exams', 'Exames', 'Upload e gestão de exames', 'FileOutlined', true),
('budgets', 'Orçamentos', 'Gestão de orçamentos de tratamentos', 'DollarOutlined', true),
('payments', 'Pagamentos', 'Gestão financeira e pagamentos', 'DollarOutlined', true),
('products', 'Produtos', 'Gestão de produtos e estoque', 'ShoppingOutlined', true),
('suppliers', 'Fornecedores', 'Gestão de fornecedores', 'ShoppingOutlined', true),
('stock_movements', 'Movimentações de Estoque', 'Controle de entradas e saídas', 'ShoppingOutlined', true),
('campaigns', 'Campanhas', 'Campanhas de marketing e comunicação', 'MessageOutlined', true),
('reports', 'Relatórios', 'Relatórios gerenciais e analíticos', 'BarChartOutlined', true),
('settings', 'Configurações', 'Configurações do sistema', 'SettingOutlined', true)
ON CONFLICT (code) DO NOTHING;

-- =============================================================================
-- 2. INSERT PERMISSIONS (4 actions for each module)
-- =============================================================================

-- Helper function to create permissions for a module
DO $$
DECLARE
    module_record RECORD;
    actions TEXT[] := ARRAY['view', 'create', 'edit', 'delete'];
    action TEXT;
    action_descriptions JSONB := '{
        "view": "Visualizar",
        "create": "Criar novos registros",
        "edit": "Editar registros existentes",
        "delete": "Deletar registros"
    }'::JSONB;
BEGIN
    -- For each module
    FOR module_record IN SELECT id, code, name FROM public.modules WHERE active = true LOOP
        -- For each action
        FOREACH action IN ARRAY actions LOOP
            INSERT INTO public.permissions (module_id, action, description)
            VALUES (
                module_record.id,
                action,
                action_descriptions->>action || ' em ' || module_record.name
            )
            ON CONFLICT (module_id, action) DO NOTHING;
        END LOOP;
    END LOOP;
END $$;

-- =============================================================================
-- 3. GRANT ALL PERMISSIONS TO EXISTING ADMIN USERS
-- =============================================================================

-- Grant all permissions to users with 'admin' role
DO $$
DECLARE
    admin_user RECORD;
    permission_record RECORD;
BEGIN
    -- For each admin user
    FOR admin_user IN SELECT id FROM public.users WHERE role = 'admin' AND deleted_at IS NULL LOOP
        -- Grant all permissions
        FOR permission_record IN SELECT id FROM public.permissions WHERE deleted_at IS NULL LOOP
            INSERT INTO public.user_permissions (user_id, permission_id, granted_by, granted_at)
            VALUES (
                admin_user.id,
                permission_record.id,
                admin_user.id,  -- Self-granted
                CURRENT_TIMESTAMP
            )
            ON CONFLICT (user_id, permission_id) DO NOTHING;
        END LOOP;

        RAISE NOTICE 'Granted all permissions to admin user ID: %', admin_user.id;
    END LOOP;
END $$;

-- =============================================================================
-- 4. CREATE DEFAULT PERMISSION TEMPLATES FOR ROLES
-- =============================================================================

-- This creates a view to easily see default permissions by role
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
        -- View only for financial and inventory
        WHEN m.code IN ('budgets', 'payments', 'products', 'suppliers', 'stock_movements') AND p.action = 'view' THEN true
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
        -- Receptionists manage appointments and patients
        WHEN m.code IN ('patients', 'appointments') THEN true
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
