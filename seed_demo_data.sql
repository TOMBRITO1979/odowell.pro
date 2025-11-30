-- =============================================================================
-- SCRIPT DE DADOS DEMO - CLÍNICA ODONTOLÓGICA ODOWELL
-- Simula 30 dias de operação com 10 pacientes/dia
-- =============================================================================

-- Usar schema do tenant 1
SET search_path TO tenant_1, public;

-- =============================================================================
-- 1. CRIAR DENTISTAS ADICIONAIS (precisamos de pelo menos 3)
-- =============================================================================

-- Criar dentistas no schema public
INSERT INTO public.users (tenant_id, name, email, password, role, cro, specialty, active, created_at, updated_at)
VALUES
(1, 'Dra. Marina Santos', 'marina@clinica.com', '$2b$12$z3UdtQ9sFB06x3LuiKb/.ux6lPzGhuuPxkQx7Lu/Kcj571v9WhIpq', 'dentist', 'CRO-SP 12345', 'Ortodontia', true, NOW(), NOW()),
(1, 'Dr. Roberto Lima', 'roberto@clinica.com', '$2b$12$z3UdtQ9sFB06x3LuiKb/.ux6lPzGhuuPxkQx7Lu/Kcj571v9WhIpq', 'dentist', 'CRO-SP 54321', 'Implantodontia', true, NOW(), NOW()),
(1, 'Dra. Fernanda Costa', 'fernanda@clinica.com', '$2b$12$z3UdtQ9sFB06x3LuiKb/.ux6lPzGhuuPxkQx7Lu/Kcj571v9WhIpq', 'dentist', 'CRO-SP 98765', 'Endodontia', true, NOW(), NOW()),
(1, 'Ana Recepção', 'recepcao@clinica.com', '$2b$12$z3UdtQ9sFB06x3LuiKb/.ux6lPzGhuuPxkQx7Lu/Kcj571v9WhIpq', 'receptionist', NULL, NULL, true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 2. FORNECEDORES DE MATERIAIS ODONTOLÓGICOS
-- =============================================================================

INSERT INTO tenant_1.suppliers (name, cnpj, email, phone, address, city, state, zip_code, active, notes, created_at, updated_at) VALUES
('Dental Cremer', '12.345.678/0001-90', 'vendas@dentalcremer.com.br', '(11) 3456-7890', 'Av. Industrial, 1500', 'São Paulo', 'SP', '01310-100', true, 'Fornecedor principal de materiais', NOW(), NOW()),
('Dentsply Sirona', '98.765.432/0001-10', 'contato@dentsply.com.br', '(11) 2345-6789', 'Rua Odontológica, 200', 'Campinas', 'SP', '13000-000', true, 'Equipamentos e resinas', NOW(), NOW()),
('Angelus Indústria', '45.678.901/0001-23', 'comercial@angelus.ind.br', '(43) 3371-7000', 'Av. Piquiri, 288', 'Londrina', 'PR', '86020-190', true, 'Materiais endodônticos', NOW(), NOW()),
('3M Dental', '33.098.765/0001-45', 'dental@3m.com.br', '(11) 4002-8922', 'Rodovia Raposo Tavares, km 23', 'São Paulo', 'SP', '05543-010', true, 'Adesivos e resinas', NOW(), NOW()),
('SS White', '67.890.123/0001-67', 'vendas@sswhite.com.br', '(21) 2233-4455', 'Rua da Saúde, 500', 'Rio de Janeiro', 'RJ', '20220-100', true, 'Instrumentais', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 3. PRODUTOS (ESTOQUE)
-- =============================================================================

INSERT INTO tenant_1.products (name, code, description, category, quantity, minimum_stock, unit, cost_price, sale_price, active, created_at, updated_at) VALUES
('Resina Composta Z350 A2', 'RES-001', 'Resina composta 3M Z350 cor A2', 'material', 25, 10, 'un', 120.00, 0, true, NOW(), NOW()),
('Resina Composta Z350 A3', 'RES-002', 'Resina composta 3M Z350 cor A3', 'material', 20, 10, 'un', 120.00, 0, true, NOW(), NOW()),
('Adesivo Single Bond', 'ADE-001', 'Adesivo dental 3M Single Bond', 'material', 15, 5, 'un', 180.00, 0, true, NOW(), NOW()),
('Ácido Fosfórico 37%', 'ACI-001', 'Condicionador ácido 37%', 'material', 30, 15, 'un', 25.00, 0, true, NOW(), NOW()),
('Cimento Ionômero de Vidro', 'CIM-001', 'Cimento restaurador provisório', 'material', 12, 5, 'un', 85.00, 0, true, NOW(), NOW()),
('Anestésico Articaína', 'ANE-001', 'Anestésico com vaso 4%', 'medicine', 100, 50, 'un', 8.50, 0, true, NOW(), NOW()),
('Anestésico Lidocaína', 'ANE-002', 'Anestésico lidocaína 2%', 'medicine', 80, 40, 'un', 6.00, 0, true, NOW(), NOW()),
('Agulha Gengival Curta', 'AGU-001', 'Agulha para anestesia curta 30G', 'consumable', 200, 100, 'un', 0.80, 0, true, NOW(), NOW()),
('Agulha Gengival Longa', 'AGU-002', 'Agulha para anestesia longa 27G', 'consumable', 150, 75, 'un', 0.85, 0, true, NOW(), NOW()),
('Luva Procedimento M', 'LUV-001', 'Luva látex tamanho M', 'consumable', 500, 200, 'un', 0.35, 0, true, NOW(), NOW()),
('Luva Procedimento P', 'LUV-002', 'Luva látex tamanho P', 'consumable', 300, 150, 'un', 0.35, 0, true, NOW(), NOW()),
('Máscara Descartável', 'MAS-001', 'Máscara tripla camada', 'consumable', 400, 200, 'un', 0.25, 0, true, NOW(), NOW()),
('Sugador Descartável', 'SUG-001', 'Sugador plástico descartável', 'consumable', 500, 250, 'un', 0.15, 0, true, NOW(), NOW()),
('Gaze Estéril', 'GAZ-001', 'Gaze estéril 7,5x7,5cm', 'consumable', 300, 150, 'un', 0.20, 0, true, NOW(), NOW()),
('Algodão Rolete', 'ALG-001', 'Algodão rolete dental', 'consumable', 400, 200, 'un', 0.10, 0, true, NOW(), NOW()),
('Lima Endodôntica K-File', 'LIM-001', 'Lima endodôntica manual', 'material', 50, 20, 'un', 15.00, 0, true, NOW(), NOW()),
('Cone de Guta-Percha', 'CON-001', 'Cone principal guta-percha', 'material', 100, 50, 'un', 1.20, 0, true, NOW(), NOW()),
('Hidróxido de Cálcio', 'HID-001', 'Pasta hidróxido de cálcio', 'material', 20, 10, 'un', 45.00, 0, true, NOW(), NOW()),
('Hipoclorito de Sódio 2,5%', 'HIP-001', 'Irrigante endodôntico', 'material', 15, 8, 'un', 18.00, 0, true, NOW(), NOW()),
('Broca Diamantada', 'BRO-001', 'Broca diamantada FG', 'material', 60, 30, 'un', 8.00, 0, true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 4. PROTOCOLOS DE TRATAMENTO
-- =============================================================================

INSERT INTO tenant_1.treatment_protocols (name, description, procedures, duration, cost, active, created_by, created_at, updated_at) VALUES
('Restauração Simples', 'Protocolo para restauração em resina de 1 face', '[{"name":"Anestesia","duration":5},{"name":"Remoção de cárie","duration":15},{"name":"Aplicação de resina","duration":20}]', 40, 150.00, true, 4, NOW(), NOW()),
('Restauração Complexa', 'Restauração de múltiplas faces', '[{"name":"Anestesia","duration":5},{"name":"Isolamento absoluto","duration":10},{"name":"Remoção de cárie","duration":20},{"name":"Aplicação de resina","duration":30}]', 65, 280.00, true, 4, NOW(), NOW()),
('Canal Unirradicular', 'Tratamento endodôntico de dente com 1 canal', '[{"name":"Anestesia","duration":5},{"name":"Abertura coronária","duration":15},{"name":"Instrumentação","duration":40},{"name":"Obturação","duration":20}]', 80, 600.00, true, 4, NOW(), NOW()),
('Canal Multirradicular', 'Tratamento endodôntico de dente com múltiplos canais', '[{"name":"Anestesia","duration":5},{"name":"Abertura coronária","duration":20},{"name":"Instrumentação","duration":90},{"name":"Obturação","duration":30}]', 145, 1200.00, true, 4, NOW(), NOW()),
('Extração Simples', 'Exodontia de dente erupcionado', '[{"name":"Anestesia","duration":10},{"name":"Sindesmotomia","duration":5},{"name":"Luxação e avulsão","duration":15},{"name":"Sutura","duration":10}]', 40, 200.00, true, 4, NOW(), NOW()),
('Limpeza Completa', 'Profilaxia e raspagem supragengival', '[{"name":"Remoção de tártaro","duration":30},{"name":"Polimento","duration":15},{"name":"Aplicação de flúor","duration":5}]', 50, 180.00, true, 4, NOW(), NOW()),
('Clareamento Consultório', 'Clareamento dental de consultório', '[{"name":"Proteção gengival","duration":15},{"name":"Aplicação de gel 1","duration":15},{"name":"Aplicação de gel 2","duration":15},{"name":"Aplicação de gel 3","duration":15}]', 60, 800.00, true, 4, NOW(), NOW()),
('Instalação de Aparelho', 'Colocação de aparelho ortodôntico fixo', '[{"name":"Profilaxia","duration":15},{"name":"Colagem de brackets","duration":45},{"name":"Inserção de fio","duration":15}]', 75, 1500.00, true, 4, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 5. TERMOS DE CONSENTIMENTO (TEMPLATES)
-- =============================================================================

INSERT INTO tenant_1.consent_templates (title, type, content, version, description, active, is_default, created_at, updated_at) VALUES
('Termo de Consentimento para Tratamento Odontológico', 'treatment', '<h2>TERMO DE CONSENTIMENTO LIVRE E ESCLARECIDO</h2><p>Eu, paciente abaixo identificado, declaro que fui informado(a) sobre o tratamento odontológico proposto, incluindo seus benefícios, riscos e alternativas.</p><p>Declaro estar ciente de que:</p><ul><li>O tratamento proposto foi explicado de forma clara</li><li>Fui informado(a) sobre possíveis complicações</li><li>Tive a oportunidade de fazer perguntas</li><li>Autorizo a realização do tratamento</li></ul>', '1.0.0', 'Termo padrão para tratamentos gerais', true, true, NOW(), NOW()),
('Termo de Consentimento para Procedimentos Cirúrgicos', 'procedure', '<h2>TERMO DE CONSENTIMENTO PARA CIRURGIA</h2><p>Declaro estar ciente dos riscos inerentes ao procedimento cirúrgico proposto, incluindo mas não limitado a: sangramento, infecção, dor pós-operatória, parestesia temporária ou permanente.</p><p>Comprometo-me a seguir todas as orientações pré e pós-operatórias.</p>', '1.0.0', 'Termo para procedimentos cirúrgicos', true, false, NOW(), NOW()),
('Termo de Consentimento para Clareamento Dental', 'procedure', '<h2>TERMO DE CONSENTIMENTO PARA CLAREAMENTO</h2><p>Estou ciente de que o clareamento dental pode causar sensibilidade transitória e que os resultados variam de acordo com cada paciente.</p><p>Fui orientado(a) sobre os cuidados durante e após o tratamento.</p>', '1.0.0', 'Termo específico para clareamento', true, false, NOW(), NOW()),
('Termo de Consentimento para Uso de Dados', 'data_privacy', '<h2>TERMO DE CONSENTIMENTO PARA USO DE DADOS</h2><p>Autorizo a clínica a armazenar e processar meus dados pessoais e de saúde conforme a LGPD, para fins de:</p><ul><li>Prontuário eletrônico</li><li>Agendamento de consultas</li><li>Comunicação sobre tratamentos</li></ul>', '1.0.0', 'Termo LGPD', true, true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 6. GERAR 130 NOVOS PACIENTES (já existem 22)
-- =============================================================================

DO $$
DECLARE
    i INTEGER;
    birth_date DATE;
    gender TEXT;
    blood_types TEXT[] := ARRAY['A+', 'A-', 'B+', 'B-', 'AB+', 'AB-', 'O+', 'O-'];
    first_names_m TEXT[] := ARRAY['João', 'Pedro', 'Lucas', 'Matheus', 'Gabriel', 'Rafael', 'Bruno', 'Felipe', 'Gustavo', 'Rodrigo', 'André', 'Marcos', 'Paulo', 'Carlos', 'Daniel', 'Fernando', 'Ricardo', 'Eduardo', 'Henrique', 'Thiago', 'Leonardo', 'Marcelo', 'Alexandre', 'Vinicius', 'Diego'];
    first_names_f TEXT[] := ARRAY['Maria', 'Ana', 'Juliana', 'Fernanda', 'Patricia', 'Camila', 'Amanda', 'Larissa', 'Beatriz', 'Carolina', 'Gabriela', 'Isabela', 'Leticia', 'Mariana', 'Natalia', 'Paula', 'Renata', 'Sabrina', 'Tatiana', 'Vanessa', 'Bruna', 'Cristina', 'Daniela', 'Elaine', 'Fabiana'];
    last_names TEXT[] := ARRAY['Silva', 'Santos', 'Oliveira', 'Souza', 'Rodrigues', 'Ferreira', 'Alves', 'Pereira', 'Lima', 'Gomes', 'Costa', 'Ribeiro', 'Martins', 'Carvalho', 'Almeida', 'Lopes', 'Soares', 'Fernandes', 'Vieira', 'Barbosa', 'Rocha', 'Dias', 'Nascimento', 'Andrade', 'Moreira'];
    streets TEXT[] := ARRAY['Rua das Flores', 'Av. Brasil', 'Rua São Paulo', 'Av. Paulista', 'Rua do Comércio', 'Av. Principal', 'Rua XV de Novembro', 'Rua Tiradentes', 'Av. Independência', 'Rua da Paz'];
    cities TEXT[] := ARRAY['São Paulo', 'Campinas', 'Santos', 'Sorocaba', 'Ribeirão Preto', 'São José dos Campos', 'Piracicaba', 'Bauru', 'Jundiaí', 'Guarulhos'];
    allergies_list TEXT[] := ARRAY['', '', '', '', '', 'Penicilina', 'Dipirona', 'Látex', 'Ibuprofeno', ''];
    diseases_list TEXT[] := ARRAY['', '', '', '', '', 'Diabetes', 'Hipertensão', 'Asma', 'Diabetes, Hipertensão', ''];
    insurances TEXT[] := ARRAY['Amil Dental', 'Bradesco Dental', 'SulAmérica', 'Unimed', 'Odontoprev', 'Porto Seguro Dental', 'MetLife', ''];
    first_name TEXT;
    last_name TEXT;
    full_name TEXT;
    cpf_num TEXT;
    phone_num TEXT;
    cell_num TEXT;
    has_ins BOOLEAN;
    ins_name TEXT;
BEGIN
    FOR i IN 1..130 LOOP
        IF random() < 0.5 THEN
            gender := 'M';
            first_name := first_names_m[1 + floor(random() * array_length(first_names_m, 1))::int];
        ELSE
            gender := 'F';
            first_name := first_names_f[1 + floor(random() * array_length(first_names_f, 1))::int];
        END IF;

        last_name := last_names[1 + floor(random() * array_length(last_names, 1))::int];
        full_name := first_name || ' ' || last_name;
        birth_date := DATE '1950-01-01' + (random() * 25000)::int;
        cpf_num := lpad(floor(random() * 999999999)::text, 11, '0');
        phone_num := '(11) ' || lpad(floor(random() * 9999)::text, 4, '0') || '-' || lpad(floor(random() * 9999)::text, 4, '0');
        cell_num := '(11) 9' || lpad(floor(random() * 9999)::text, 4, '0') || '-' || lpad(floor(random() * 9999)::text, 4, '0');

        IF random() < 0.3 THEN
            has_ins := true;
            ins_name := insurances[1 + floor(random() * (array_length(insurances, 1) - 1))::int];
        ELSE
            has_ins := false;
            ins_name := '';
        END IF;

        INSERT INTO tenant_1.patients (
            name, cpf, birth_date, gender, email, phone, cell_phone,
            address, number, district, city, state, zip_code,
            allergies, systemic_diseases, blood_type,
            has_insurance, insurance_name, active, created_at, updated_at
        ) VALUES (
            full_name, cpf_num, birth_date, gender,
            lower(replace(first_name, ' ', '')) || floor(random() * 999)::text || '@email.com',
            phone_num, cell_num,
            streets[1 + floor(random() * array_length(streets, 1))::int],
            floor(random() * 2000)::text, 'Centro',
            cities[1 + floor(random() * array_length(cities, 1))::int], 'SP',
            lpad(floor(random() * 99999)::text, 5, '0') || '-' || lpad(floor(random() * 999)::text, 3, '0'),
            allergies_list[1 + floor(random() * array_length(allergies_list, 1))::int],
            diseases_list[1 + floor(random() * array_length(diseases_list, 1))::int],
            blood_types[1 + floor(random() * array_length(blood_types, 1))::int],
            has_ins, ins_name, true,
            NOW() - (random() * 365 || ' days')::interval, NOW()
        );
    END LOOP;
END $$;

-- =============================================================================
-- 7. AGENDAMENTOS (30 dias, ~10 por dia = 300 agendamentos)
-- =============================================================================

DO $$
DECLARE
    day_offset INTEGER;
    slot INTEGER;
    patient_id INTEGER;
    dentist_id INTEGER;
    dentist_ids INTEGER[];
    patient_ids INTEGER[];
    start_dt TIMESTAMP;
    end_dt TIMESTAMP;
    appt_status TEXT;
    procedures TEXT[] := ARRAY['Consulta de avaliação', 'Restauração', 'Limpeza', 'Extração', 'Canal', 'Clareamento', 'Manutenção de aparelho', 'Retorno'];
    types TEXT[] := ARRAY['consultation', 'treatment', 'treatment', 'treatment', 'treatment', 'treatment', 'treatment', 'return'];
    proc_idx INTEGER;
    base_date DATE;
BEGIN
    SELECT ARRAY(SELECT id FROM public.users WHERE tenant_id = 1 AND role IN ('dentist', 'admin') AND deleted_at IS NULL LIMIT 4) INTO dentist_ids;
    SELECT ARRAY(SELECT id FROM tenant_1.patients WHERE deleted_at IS NULL ORDER BY id) INTO patient_ids;
    base_date := CURRENT_DATE - 15;

    FOR day_offset IN 0..29 LOOP
        IF EXTRACT(DOW FROM base_date + day_offset) = 0 THEN
            CONTINUE;
        END IF;

        FOR slot IN 1..10 + floor(random() * 3)::int LOOP
            patient_id := patient_ids[1 + floor(random() * array_length(patient_ids, 1))::int];
            dentist_id := dentist_ids[1 + floor(random() * array_length(dentist_ids, 1))::int];
            proc_idx := 1 + floor(random() * array_length(procedures, 1))::int;

            start_dt := (base_date + day_offset)::timestamp +
                        (8 + floor(random() * 10))::int * interval '1 hour' +
                        (floor(random() * 2) * 30)::int * interval '1 minute';
            end_dt := start_dt + (30 + floor(random() * 3) * 30)::int * interval '1 minute';

            IF (base_date + day_offset) < CURRENT_DATE THEN
                IF random() < 0.70 THEN appt_status := 'completed';
                ELSIF random() < 0.85 THEN appt_status := 'no_show';
                ELSE appt_status := 'cancelled'; END IF;
            ELSIF (base_date + day_offset) = CURRENT_DATE THEN
                IF random() < 0.3 THEN appt_status := 'completed';
                ELSIF random() < 0.6 THEN appt_status := 'in_progress';
                ELSE appt_status := 'confirmed'; END IF;
            ELSE
                IF random() < 0.6 THEN appt_status := 'confirmed';
                ELSE appt_status := 'scheduled'; END IF;
            END IF;

            INSERT INTO tenant_1.appointments (
                patient_id, dentist_id, start_time, end_time,
                type, procedure, status, confirmed, notes, created_at, updated_at
            ) VALUES (
                patient_id, dentist_id, start_dt, end_dt,
                types[proc_idx], procedures[proc_idx], appt_status,
                appt_status IN ('confirmed', 'completed', 'in_progress'),
                CASE WHEN random() < 0.3 THEN 'Paciente solicitou horário específico' ELSE '' END,
                NOW() - (random() * 30 || ' days')::interval, NOW()
            );
        END LOOP;
    END LOOP;
END $$;

-- =============================================================================
-- 8. ORÇAMENTOS (100 orçamentos com diferentes status)
-- =============================================================================

DO $$
DECLARE
    i INTEGER;
    patient_id INTEGER;
    dentist_id INTEGER;
    dentist_ids INTEGER[];
    patient_ids INTEGER[];
    budget_status TEXT;
    budget_value NUMERIC;
    items_json TEXT;
    valid_date DATE;
BEGIN
    SELECT ARRAY(SELECT id FROM public.users WHERE tenant_id = 1 AND role IN ('dentist', 'admin') AND deleted_at IS NULL LIMIT 4) INTO dentist_ids;
    SELECT ARRAY(SELECT id FROM tenant_1.patients WHERE deleted_at IS NULL ORDER BY id) INTO patient_ids;

    FOR i IN 1..100 LOOP
        patient_id := patient_ids[1 + floor(random() * array_length(patient_ids, 1))::int];
        dentist_id := dentist_ids[1 + floor(random() * array_length(dentist_ids, 1))::int];

        IF random() < 0.30 THEN budget_status := 'pending';
        ELSIF random() < 0.80 THEN budget_status := 'approved';
        ELSIF random() < 0.95 THEN budget_status := 'rejected';
        ELSE budget_status := 'expired'; END IF;

        budget_value := 150 + floor(random() * 4850);
        items_json := '[';
        IF budget_value > 500 THEN items_json := items_json || '{"description":"Restauração em resina","quantity":2,"unit_price":150,"total":300},'; END IF;
        IF budget_value > 1000 THEN items_json := items_json || '{"description":"Tratamento de canal","quantity":1,"unit_price":800,"total":800},'; END IF;
        IF budget_value > 2000 THEN items_json := items_json || '{"description":"Coroa de porcelana","quantity":1,"unit_price":1500,"total":1500},'; END IF;
        items_json := items_json || '{"description":"Profilaxia","quantity":1,"unit_price":180,"total":180}]';
        valid_date := CURRENT_DATE + floor(random() * 30)::int;

        INSERT INTO tenant_1.budgets (
            patient_id, dentist_id, description, total_value, items,
            status, valid_until, notes, created_at, updated_at
        ) VALUES (
            patient_id, dentist_id, 'Plano de tratamento odontológico', budget_value, items_json,
            budget_status, valid_date,
            CASE WHEN random() < 0.3 THEN 'Paciente solicitou parcelamento' ELSE '' END,
            NOW() - (random() * 60 || ' days')::interval, NOW()
        );
    END LOOP;
END $$;

-- =============================================================================
-- 9. TRATAMENTOS (orçamentos aprovados viram tratamentos)
-- =============================================================================

INSERT INTO tenant_1.treatments (
    budget_id, patient_id, dentist_id, description, total_value, paid_value,
    total_installments, installment_value, status, start_date, expected_end_date,
    notes, created_at, updated_at
)
SELECT
    b.id, b.patient_id, b.dentist_id, b.description, b.total_value,
    CASE WHEN random() < 0.3 THEN b.total_value WHEN random() < 0.6 THEN b.total_value * 0.5 ELSE b.total_value * 0.3 END,
    CASE WHEN b.total_value > 1000 THEN floor(b.total_value / 500) + 1 ELSE 1 END,
    b.total_value / GREATEST(1, floor(b.total_value / 500) + 1),
    CASE WHEN random() < 0.2 THEN 'completed' ELSE 'in_progress' END,
    NOW() - (random() * 30 || ' days')::interval,
    NOW() + (random() * 60 || ' days')::interval,
    'Tratamento iniciado', NOW(), NOW()
FROM tenant_1.budgets b WHERE b.status = 'approved'
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 10. PAGAMENTOS - RECEITAS
-- =============================================================================

INSERT INTO tenant_1.payments (
    budget_id, patient_id, type, category, description, amount,
    payment_method, status, due_date, paid_date, notes, created_at, updated_at
)
SELECT
    t.budget_id, t.patient_id, 'income', 'treatment',
    'Pagamento de tratamento - ' || t.description, t.paid_value,
    CASE floor(random() * 5) WHEN 0 THEN 'cash' WHEN 1 THEN 'credit_card' WHEN 2 THEN 'debit_card' WHEN 3 THEN 'pix' ELSE 'transfer' END,
    'paid', t.start_date::date, t.start_date::date, '', NOW(), NOW()
FROM tenant_1.treatments t WHERE t.paid_value > 0
ON CONFLICT DO NOTHING;

-- Pagamentos avulsos de consultas
DO $$
DECLARE
    i INTEGER;
    patient_ids INTEGER[];
    patient_id INTEGER;
    pay_date DATE;
BEGIN
    SELECT ARRAY(SELECT id FROM tenant_1.patients WHERE deleted_at IS NULL ORDER BY id) INTO patient_ids;
    FOR i IN 1..50 LOOP
        patient_id := patient_ids[1 + floor(random() * array_length(patient_ids, 1))::int];
        pay_date := CURRENT_DATE - floor(random() * 30)::int;
        INSERT INTO tenant_1.payments (
            patient_id, type, category, description, amount, payment_method, status, due_date, paid_date, created_at, updated_at
        ) VALUES (
            patient_id, 'income', 'consultation', 'Consulta de avaliação', 150 + floor(random() * 100),
            CASE floor(random() * 5) WHEN 0 THEN 'cash' WHEN 1 THEN 'credit_card' WHEN 2 THEN 'debit_card' WHEN 3 THEN 'pix' ELSE 'transfer' END,
            'paid', pay_date, pay_date, NOW(), NOW()
        );
    END LOOP;
END $$;

-- =============================================================================
-- 11. PAGAMENTOS - DESPESAS
-- =============================================================================

INSERT INTO tenant_1.payments (patient_id, type, category, description, amount, payment_method, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(1, 'expense', 'rent', 'Aluguel do imóvel - Novembro', 5500.00, 'transfer', 'paid', '2024-11-05', '2024-11-05', 'Contrato anual', NOW(), NOW()),
(1, 'expense', 'rent', 'Aluguel do imóvel - Dezembro', 5500.00, 'transfer', 'pending', '2024-12-05', NULL, 'Contrato anual', NOW(), NOW()),
(1, 'expense', 'utilities', 'Conta de luz - Novembro', 850.00, 'pix', 'paid', '2024-11-15', '2024-11-14', '', NOW(), NOW()),
(1, 'expense', 'utilities', 'Conta de luz - Dezembro', 920.00, 'pix', 'pending', '2024-12-15', NULL, '', NOW(), NOW()),
(1, 'expense', 'utilities', 'Conta de água - Novembro', 180.00, 'pix', 'paid', '2024-11-20', '2024-11-19', '', NOW(), NOW()),
(1, 'expense', 'utilities', 'Conta de água - Dezembro', 195.00, 'pix', 'pending', '2024-12-20', NULL, '', NOW(), NOW()),
(1, 'expense', 'utilities', 'Internet e telefone - Novembro', 350.00, 'debit_card', 'paid', '2024-11-10', '2024-11-10', '', NOW(), NOW()),
(1, 'expense', 'utilities', 'Internet e telefone - Dezembro', 350.00, 'debit_card', 'pending', '2024-12-10', NULL, '', NOW(), NOW()),
(1, 'expense', 'salary', 'Salário - Recepcionista Ana', 2500.00, 'transfer', 'paid', '2024-11-05', '2024-11-05', 'Folha', NOW(), NOW()),
(1, 'expense', 'salary', 'Salário - Recepcionista Ana', 2500.00, 'transfer', 'pending', '2024-12-05', NULL, 'Folha', NOW(), NOW()),
(1, 'expense', 'salary', 'Salário - Auxiliar limpeza', 1800.00, 'transfer', 'paid', '2024-11-05', '2024-11-05', '', NOW(), NOW()),
(1, 'expense', 'salary', 'Salário - Auxiliar limpeza', 1800.00, 'transfer', 'pending', '2024-12-05', NULL, '', NOW(), NOW()),
(1, 'expense', 'commission', 'Comissão - Dra. Marina - Nov', 4500.00, 'transfer', 'paid', '2024-11-10', '2024-11-10', '30%', NOW(), NOW()),
(1, 'expense', 'commission', 'Comissão - Dr. Roberto - Nov', 5200.00, 'transfer', 'paid', '2024-11-10', '2024-11-10', '30%', NOW(), NOW()),
(1, 'expense', 'commission', 'Comissão - Dra. Fernanda - Nov', 3800.00, 'transfer', 'paid', '2024-11-10', '2024-11-10', '30%', NOW(), NOW()),
(1, 'expense', 'material', 'Compra resinas - Dental Cremer', 2800.00, 'credit_card', 'paid', '2024-11-08', '2024-11-08', 'NF 12345', NOW(), NOW()),
(1, 'expense', 'material', 'Materiais consumo - Angelus', 1500.00, 'pix', 'paid', '2024-11-12', '2024-11-12', 'NF 67890', NOW(), NOW()),
(1, 'expense', 'material', 'Reposição estoque - 3M', 3200.00, 'transfer', 'pending', '2024-12-08', NULL, '', NOW(), NOW()),
(1, 'expense', 'maintenance', 'Manutenção compressor', 450.00, 'cash', 'paid', '2024-11-18', '2024-11-18', 'Preventiva', NOW(), NOW()),
(1, 'expense', 'maintenance', 'Calibração autoclave', 280.00, 'pix', 'paid', '2024-11-22', '2024-11-22', 'Certificado', NOW(), NOW()),
(1, 'expense', 'services', 'Honorários contábeis - Nov', 800.00, 'transfer', 'paid', '2024-11-10', '2024-11-10', '', NOW(), NOW()),
(1, 'expense', 'services', 'Honorários contábeis - Dez', 800.00, 'transfer', 'pending', '2024-12-10', NULL, '', NOW(), NOW()),
(1, 'expense', 'marketing', 'Google Ads - Nov', 600.00, 'credit_card', 'paid', '2024-11-01', '2024-11-01', '', NOW(), NOW()),
(1, 'expense', 'marketing', 'Instagram Ads - Nov', 400.00, 'credit_card', 'paid', '2024-11-01', '2024-11-01', '', NOW(), NOW()),
(1, 'expense', 'insurance', 'Seguro RC', 350.00, 'debit_card', 'paid', '2024-11-15', '2024-11-15', 'Parcela 8/12', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 12. PRONTUÁRIOS MÉDICOS
-- =============================================================================

DO $$
DECLARE
    i INTEGER;
    patient_ids INTEGER[];
    dentist_ids INTEGER[];
    patient_id INTEGER;
    dentist_id INTEGER;
    rec_type TEXT;
    rec_types TEXT[] := ARRAY['anamnesis', 'treatment', 'procedure'];
    diagnoses TEXT[] := ARRAY['Cárie em dente 36', 'Gengivite leve', 'Necessidade de canal em dente 46', 'Fratura de esmalte', 'Periodontite inicial', 'Bruxismo', 'Sensibilidade dentinária'];
    procedures_done TEXT[] := ARRAY['Restauração em resina composta', 'Raspagem supragengival', 'Pulpectomia iniciada', 'Aplicação de dessensibilizante', 'Profilaxia completa', 'Extração de dente 38'];
    odontogram_sample TEXT;
BEGIN
    SELECT ARRAY(SELECT id FROM tenant_1.patients WHERE deleted_at IS NULL ORDER BY id LIMIT 100) INTO patient_ids;
    SELECT ARRAY(SELECT id FROM public.users WHERE tenant_id = 1 AND role IN ('dentist', 'admin') AND deleted_at IS NULL LIMIT 4) INTO dentist_ids;

    odontogram_sample := '{"11":{"status":"healthy"},"12":{"status":"healthy"},"13":{"status":"healthy"},"14":{"status":"restored","procedures":["restoration"]},"15":{"status":"healthy"},"16":{"status":"cavity","procedures":[]},"17":{"status":"healthy"},"18":{"status":"missing"},"21":{"status":"healthy"},"22":{"status":"healthy"},"23":{"status":"healthy"},"24":{"status":"healthy"},"25":{"status":"restored","procedures":["restoration"]},"26":{"status":"healthy"},"27":{"status":"healthy"},"28":{"status":"missing"},"31":{"status":"healthy"},"32":{"status":"healthy"},"33":{"status":"healthy"},"34":{"status":"healthy"},"35":{"status":"healthy"},"36":{"status":"root_canal","procedures":["root_canal","crown"]},"37":{"status":"healthy"},"38":{"status":"missing"},"41":{"status":"healthy"},"42":{"status":"healthy"},"43":{"status":"healthy"},"44":{"status":"healthy"},"45":{"status":"healthy"},"46":{"status":"cavity","procedures":[]},"47":{"status":"healthy"},"48":{"status":"impacted"}}';

    FOR i IN 1..150 LOOP
        patient_id := patient_ids[1 + floor(random() * array_length(patient_ids, 1))::int];
        dentist_id := dentist_ids[1 + floor(random() * array_length(dentist_ids, 1))::int];
        rec_type := rec_types[1 + floor(random() * array_length(rec_types, 1))::int];

        INSERT INTO tenant_1.medical_records (
            patient_id, dentist_id, type, odontogram, diagnosis, treatment_plan, procedure_done, materials, evolution, notes, created_at, updated_at
        ) VALUES (
            patient_id, dentist_id, rec_type,
            CASE WHEN rec_type = 'anamnesis' THEN odontogram_sample ELSE NULL END,
            diagnoses[1 + floor(random() * array_length(diagnoses, 1))::int],
            'Plano: ' || procedures_done[1 + floor(random() * array_length(procedures_done, 1))::int],
            CASE WHEN rec_type = 'procedure' THEN procedures_done[1 + floor(random() * array_length(procedures_done, 1))::int] ELSE NULL END,
            CASE WHEN rec_type = 'procedure' THEN 'Resina Z350, Adesivo Single Bond, Anestésico Articaína' ELSE NULL END,
            'Paciente evoluindo bem, sem queixas.', '',
            NOW() - (random() * 60 || ' days')::interval, NOW()
        );
    END LOOP;
END $$;

-- =============================================================================
-- 13. PRESCRIÇÕES (Receitas, Atestados, Laudos)
-- =============================================================================

DO $$
DECLARE
    i INTEGER;
    patient_ids INTEGER[];
    dentist_ids INTEGER[];
    patient_id INTEGER;
    dentist_id INTEGER;
    presc_type TEXT;
    presc_types TEXT[] := ARRAY['prescription', 'certificate', 'medical_report'];
    medications_list TEXT[] := ARRAY[
        'Amoxicilina 500mg - Tomar 1 cápsula de 8 em 8 horas por 7 dias',
        'Nimesulida 100mg - Tomar 1 comprimido de 12 em 12 horas por 5 dias',
        'Dipirona 500mg - Tomar 1 comprimido de 6 em 6 horas se dor',
        'Ibuprofeno 600mg - Tomar 1 comprimido de 8 em 8 horas por 3 dias'
    ];
    certificates_list TEXT[] := ARRAY[
        'Atesto que o(a) paciente compareceu a consulta odontológica nesta data, necessitando de afastamento por 1 dia.',
        'Atesto que o(a) paciente realizou procedimento cirúrgico odontológico, necessitando de repouso por 2 dias.',
        'Atesto para fins de comprovação que o(a) paciente está em tratamento odontológico nesta clínica.'
    ];
    reports_list TEXT[] := ARRAY[
        'Laudo Pericial: Após exame clínico e radiográfico, constata-se a necessidade de tratamento ortodôntico.',
        'Laudo Técnico: Paciente apresenta perda de elemento dental 36, indicado implante osseointegrado.',
        'Laudo de Aptidão: Paciente apto(a) para procedimentos odontológicos sob anestesia local.'
    ];
BEGIN
    SELECT ARRAY(SELECT id FROM tenant_1.patients WHERE deleted_at IS NULL ORDER BY id LIMIT 100) INTO patient_ids;
    SELECT ARRAY(SELECT id FROM public.users WHERE tenant_id = 1 AND role IN ('dentist', 'admin') AND deleted_at IS NULL LIMIT 4) INTO dentist_ids;

    FOR i IN 1..80 LOOP
        patient_id := patient_ids[1 + floor(random() * array_length(patient_ids, 1))::int];
        dentist_id := dentist_ids[1 + floor(random() * array_length(dentist_ids, 1))::int];
        presc_type := presc_types[1 + floor(random() * array_length(presc_types, 1))::int];

        INSERT INTO tenant_1.prescriptions (
            patient_id, dentist_id, type, title, medications, content, diagnosis, status,
            clinic_name, clinic_address, clinic_phone, dentist_name, dentist_cro, prescription_date, created_at, updated_at
        ) VALUES (
            patient_id, dentist_id, presc_type,
            CASE presc_type WHEN 'prescription' THEN 'Receita Médica' WHEN 'certificate' THEN 'Atestado Odontológico' ELSE 'Laudo Odontológico' END,
            CASE WHEN presc_type = 'prescription' THEN medications_list[1 + floor(random() * array_length(medications_list, 1))::int] ELSE NULL END,
            CASE presc_type
                WHEN 'prescription' THEN 'Uso interno conforme orientação médica.'
                WHEN 'certificate' THEN certificates_list[1 + floor(random() * array_length(certificates_list, 1))::int]
                ELSE reports_list[1 + floor(random() * array_length(reports_list, 1))::int]
            END,
            CASE WHEN presc_type = 'medical_report' THEN 'Avaliação odontológica completa' ELSE NULL END,
            'issued', 'Clínica Dental Sorriso Perfeito', 'Rua das Flores, 123 - Centro - São Paulo/SP', '(11) 3456-7890',
            'Dr. Carlos Silva', 'CRO-SP 12345',
            (NOW() - (random() * 30 || ' days')::interval)::date,
            NOW() - (random() * 30 || ' days')::interval, NOW()
        );
    END LOOP;
END $$;

-- =============================================================================
-- 14. LISTA DE ESPERA
-- =============================================================================

DO $$
DECLARE
    i INTEGER;
    patient_ids INTEGER[];
    dentist_ids INTEGER[];
    patient_id INTEGER;
    dentist_id INTEGER;
    wait_status TEXT;
    procedures TEXT[] := ARRAY['Implante dentário', 'Clareamento', 'Aparelho ortodôntico', 'Extração de siso', 'Prótese fixa', 'Limpeza profunda'];
BEGIN
    SELECT ARRAY(SELECT id FROM tenant_1.patients WHERE deleted_at IS NULL ORDER BY id LIMIT 100) INTO patient_ids;
    SELECT ARRAY(SELECT id FROM public.users WHERE tenant_id = 1 AND role IN ('dentist', 'admin') AND deleted_at IS NULL LIMIT 4) INTO dentist_ids;

    FOR i IN 1..25 LOOP
        patient_id := patient_ids[1 + floor(random() * array_length(patient_ids, 1))::int];
        dentist_id := dentist_ids[1 + floor(random() * array_length(dentist_ids, 1))::int];

        IF random() < 0.6 THEN wait_status := 'waiting';
        ELSIF random() < 0.8 THEN wait_status := 'contacted';
        ELSE wait_status := 'scheduled'; END IF;

        INSERT INTO tenant_1.waiting_lists (
            patient_id, dentist_id, procedure, priority, status, notes, created_by, created_at, updated_at
        ) VALUES (
            patient_id,
            CASE WHEN random() < 0.5 THEN dentist_id ELSE NULL END,
            procedures[1 + floor(random() * array_length(procedures, 1))::int],
            CASE WHEN random() < 0.8 THEN 'normal' ELSE 'urgent' END,
            wait_status,
            CASE WHEN random() < 0.3 THEN 'Paciente prefere horários da manhã' ELSE '' END,
            4, NOW() - (random() * 30 || ' days')::interval, NOW()
        );
    END LOOP;
END $$;

-- =============================================================================
-- 15. TAREFAS DA EQUIPE
-- =============================================================================

INSERT INTO tenant_1.tasks (title, description, due_date, priority, status, created_by, created_at, updated_at) VALUES
('Ligar para pacientes faltosos', 'Entrar em contato com pacientes que faltaram nas consultas da última semana', CURRENT_DATE + 1, 'high', 'pending', 4, NOW(), NOW()),
('Verificar estoque de anestésicos', 'Fazer contagem do estoque de anestésicos e verificar necessidade de compra', CURRENT_DATE + 2, 'medium', 'pending', 4, NOW(), NOW()),
('Enviar lembrete de consultas', 'Enviar WhatsApp confirmando consultas de amanhã', CURRENT_DATE, 'high', 'in_progress', 4, NOW(), NOW()),
('Organizar prontuários físicos', 'Digitalizar prontuários antigos e organizar arquivo', CURRENT_DATE + 7, 'low', 'pending', 4, NOW(), NOW()),
('Manutenção do ar condicionado', 'Agendar técnico para manutenção preventiva', CURRENT_DATE + 5, 'medium', 'pending', 4, NOW(), NOW()),
('Reunião de equipe', 'Reunião mensal para discussão de casos e planejamento', CURRENT_DATE + 3, 'medium', 'pending', 4, NOW(), NOW()),
('Atualizar tabela de preços', 'Revisar valores dos procedimentos para 2025', CURRENT_DATE + 10, 'medium', 'pending', 4, NOW(), NOW()),
('Comprar EPIs', 'Repor estoque de luvas, máscaras e gorros', CURRENT_DATE + 2, 'high', 'pending', 4, NOW(), NOW()),
('Calibrar equipamentos', 'Verificar calibração do fotopolimerizador e raio-x', CURRENT_DATE + 14, 'medium', 'pending', 4, NOW(), NOW()),
('Confirmar férias equipe', 'Verificar escala de férias de dezembro', CURRENT_DATE + 5, 'low', 'pending', 4, NOW(), NOW()),
('Pagar fornecedores', 'Processar pagamentos pendentes da Dental Cremer', CURRENT_DATE + 1, 'high', 'completed', 4, NOW() - interval '2 days', NOW()),
('Enviar relatório mensal', 'Preparar relatório financeiro de novembro', CURRENT_DATE - 5, 'high', 'completed', 4, NOW() - interval '10 days', NOW()),
('Atualizar site', 'Adicionar novos procedimentos e fotos ao site', CURRENT_DATE + 20, 'low', 'pending', 4, NOW(), NOW()),
('Treinamento nova funcionária', 'Treinar Ana no sistema de agendamento', CURRENT_DATE - 3, 'high', 'completed', 4, NOW() - interval '7 days', NOW()),
('Revisar contratos', 'Analisar renovação de contratos com convênios', CURRENT_DATE + 15, 'medium', 'pending', 4, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 16. CAMPANHAS DE MARKETING
-- =============================================================================

INSERT INTO tenant_1.campaigns (name, type, subject, message, segment_type, status, total_recipients, sent, delivered, created_by_id, created_at, updated_at) VALUES
('Lembrete de retorno', 'whatsapp', NULL, 'Olá! Notamos que faz tempo desde sua última visita. Que tal agendar uma avaliação? Ligue: (11) 3456-7890', 'all', 'sent', 50, 48, 45, 4, NOW() - interval '15 days', NOW()),
('Promoção Clareamento', 'whatsapp', NULL, 'Dezembro chegou! Aproveite 20% OFF no clareamento dental. Agende já!', 'tags', 'sent', 30, 28, 26, 4, NOW() - interval '5 days', NOW()),
('Feliz Natal', 'email', 'Boas Festas - Clínica Sorriso Perfeito', 'A equipe da Clínica Sorriso Perfeito deseja a você e sua família um Natal cheio de alegria!', 'all', 'scheduled', 150, 0, 0, 4, NOW(), NOW()),
('Aniversariantes do mês', 'whatsapp', NULL, 'Feliz aniversário! A Clínica preparou um presente especial para você. Entre em contato!', 'custom', 'draft', 0, 0, 0, 4, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =============================================================================
-- 17. MOVIMENTAÇÕES DE ESTOQUE
-- =============================================================================

DO $$
DECLARE
    prod RECORD;
    user_id INTEGER := 4;
BEGIN
    FOR prod IN SELECT id, quantity FROM tenant_1.products WHERE deleted_at IS NULL LOOP
        INSERT INTO tenant_1.stock_movements (product_id, type, quantity, reason, user_id, notes, created_at, updated_at)
        VALUES (prod.id, 'entry', prod.quantity, 'purchase', user_id, 'Estoque inicial', NOW() - interval '30 days', NOW());

        IF prod.quantity > 20 THEN
            INSERT INTO tenant_1.stock_movements (product_id, type, quantity, reason, user_id, notes, created_at, updated_at)
            VALUES (prod.id, 'exit', floor(random() * 10 + 5)::int, 'usage', user_id, 'Uso em procedimentos', NOW() - interval '15 days', NOW());
            INSERT INTO tenant_1.stock_movements (product_id, type, quantity, reason, user_id, notes, created_at, updated_at)
            VALUES (prod.id, 'exit', floor(random() * 8 + 3)::int, 'usage', user_id, 'Uso em procedimentos', NOW() - interval '7 days', NOW());
        END IF;
    END LOOP;
END $$;

-- =============================================================================
-- 18. CONSENTIMENTOS ASSINADOS
-- =============================================================================

DO $$
DECLARE
    i INTEGER;
    patient_ids INTEGER[];
    template_ids INTEGER[];
    patient_id INTEGER;
    template_id INTEGER;
    template_rec RECORD;
BEGIN
    SELECT ARRAY(SELECT id FROM tenant_1.patients WHERE deleted_at IS NULL ORDER BY id LIMIT 80) INTO patient_ids;
    SELECT ARRAY(SELECT id FROM tenant_1.consent_templates WHERE active = true AND deleted_at IS NULL) INTO template_ids;

    FOR i IN 1..60 LOOP
        patient_id := patient_ids[1 + floor(random() * array_length(patient_ids, 1))::int];
        template_id := template_ids[1 + floor(random() * array_length(template_ids, 1))::int];
        SELECT * INTO template_rec FROM tenant_1.consent_templates WHERE id = template_id;

        INSERT INTO tenant_1.patient_consents (
            patient_id, template_id, template_title, template_version, template_content, template_type,
            signed_at, signature_type, signer_name, signer_relation, signed_by_user_id, status, created_at, updated_at
        ) VALUES (
            patient_id, template_id, template_rec.title, template_rec.version, template_rec.content, template_rec.type,
            NOW() - (random() * 60 || ' days')::interval, 'digital',
            (SELECT name FROM tenant_1.patients WHERE id = patient_id), 'patient', 4, 'active',
            NOW() - (random() * 60 || ' days')::interval, NOW()
        ) ON CONFLICT DO NOTHING;
    END LOOP;
END $$;

-- =============================================================================
-- FINALIZAÇÃO - VERIFICAR TOTAIS
-- =============================================================================

DO $$
DECLARE
    total_patients INTEGER;
    total_appointments INTEGER;
    total_budgets INTEGER;
    total_treatments INTEGER;
    total_payments INTEGER;
    total_records INTEGER;
    total_prescriptions INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_patients FROM tenant_1.patients WHERE deleted_at IS NULL;
    SELECT COUNT(*) INTO total_appointments FROM tenant_1.appointments WHERE deleted_at IS NULL;
    SELECT COUNT(*) INTO total_budgets FROM tenant_1.budgets WHERE deleted_at IS NULL;
    SELECT COUNT(*) INTO total_treatments FROM tenant_1.treatments WHERE deleted_at IS NULL;
    SELECT COUNT(*) INTO total_payments FROM tenant_1.payments WHERE deleted_at IS NULL;
    SELECT COUNT(*) INTO total_records FROM tenant_1.medical_records WHERE deleted_at IS NULL;
    SELECT COUNT(*) INTO total_prescriptions FROM tenant_1.prescriptions WHERE deleted_at IS NULL;

    RAISE NOTICE '========================================';
    RAISE NOTICE 'DADOS DEMO INSERIDOS COM SUCESSO!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Pacientes: %', total_patients;
    RAISE NOTICE 'Agendamentos: %', total_appointments;
    RAISE NOTICE 'Orçamentos: %', total_budgets;
    RAISE NOTICE 'Tratamentos: %', total_treatments;
    RAISE NOTICE 'Pagamentos: %', total_payments;
    RAISE NOTICE 'Prontuários: %', total_records;
    RAISE NOTICE 'Prescrições: %', total_prescriptions;
    RAISE NOTICE '========================================';
END $$;
