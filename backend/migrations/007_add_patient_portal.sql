-- Migration: Add patient portal support
-- Adds patient_id column to users table for linking patient accounts

-- Add patient_id column to users table
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS patient_id INTEGER;

-- Create index for patient_id lookups
CREATE INDEX IF NOT EXISTS idx_users_patient_id ON public.users(patient_id) WHERE patient_id IS NOT NULL;

-- Add patient_portal module to modules table
INSERT INTO public.modules (code, name, description, icon, active, created_at, updated_at)
VALUES ('patient_portal', 'Portal do Paciente', 'Gerenciamento de acesso ao portal do paciente', 'UserOutlined', true, NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- Add permissions for patient_portal module
INSERT INTO public.permissions (module_id, action, created_at, updated_at)
SELECT m.id, action, NOW(), NOW()
FROM public.modules m
CROSS JOIN (VALUES ('view'), ('create'), ('edit'), ('delete')) AS actions(action)
WHERE m.code = 'patient_portal'
ON CONFLICT DO NOTHING;
