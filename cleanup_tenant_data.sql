-- Script de Limpeza de Dados do Tenant
-- Mantém apenas usuários e permissões
-- Remove todos os dados operacionais do tenant_1

SET search_path TO tenant_1;

-- Desabilitar triggers e constraints temporariamente
SET session_replication_role = 'replica';

-- Limpar dados operacionais (ordem importa devido a FKs)
TRUNCATE TABLE campaign_recipients CASCADE;
TRUNCATE TABLE campaigns CASCADE;
TRUNCATE TABLE attachments CASCADE;
TRUNCATE TABLE consent_templates CASCADE;
TRUNCATE TABLE consents CASCADE;
TRUNCATE TABLE exams CASCADE;
TRUNCATE TABLE prescriptions CASCADE;
TRUNCATE TABLE medical_records CASCADE;
TRUNCATE TABLE tasks CASCADE;
TRUNCATE TABLE waiting_list CASCADE;
TRUNCATE TABLE payments CASCADE;
TRUNCATE TABLE budgets CASCADE;
TRUNCATE TABLE appointments CASCADE;
TRUNCATE TABLE patients CASCADE;
TRUNCATE TABLE stock_movements CASCADE;
TRUNCATE TABLE products CASCADE;
TRUNCATE TABLE suppliers CASCADE;
TRUNCATE TABLE settings CASCADE;

-- Reabilitar triggers
SET session_replication_role = 'origin';

-- Resetar sequences
ALTER SEQUENCE patients_id_seq RESTART WITH 1;
ALTER SEQUENCE appointments_id_seq RESTART WITH 1;
ALTER SEQUENCE budgets_id_seq RESTART WITH 1;
ALTER SEQUENCE payments_id_seq RESTART WITH 1;
ALTER SEQUENCE medical_records_id_seq RESTART WITH 1;
ALTER SEQUENCE products_id_seq RESTART WITH 1;
ALTER SEQUENCE suppliers_id_seq RESTART WITH 1;
ALTER SEQUENCE stock_movements_id_seq RESTART WITH 1;
ALTER SEQUENCE campaigns_id_seq RESTART WITH 1;
ALTER SEQUENCE campaign_recipients_id_seq RESTART WITH 1;
ALTER SEQUENCE attachments_id_seq RESTART WITH 1;
ALTER SEQUENCE exams_id_seq RESTART WITH 1;
ALTER SEQUENCE prescriptions_id_seq RESTART WITH 1;
ALTER SEQUENCE tasks_id_seq RESTART WITH 1;
ALTER SEQUENCE waiting_list_id_seq RESTART WITH 1;
ALTER SEQUENCE consent_templates_id_seq RESTART WITH 1;
ALTER SEQUENCE consents_id_seq RESTART WITH 1;

RESET search_path;

SELECT 'Limpeza concluída! Todos os dados operacionais foram removidos.' AS status;
