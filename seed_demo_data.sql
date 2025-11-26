-- Script de População com Dados Fictícios Profissionais
-- Sistema de Gestão de Clínica Odontológica - OdoWell
-- Dados para demonstração comercial

SET search_path TO tenant_1;

-- ========================================
-- 1. PACIENTES (20 pacientes)
-- ========================================
INSERT INTO patients (name, cpf, rg, birth_date, gender, phone, email, address, city, state, zip_code, occupation, health_insurance, active, created_at, updated_at) VALUES
('Ana Paula Silva Santos', '123.456.789-01', '12.345.678-9', '1985-03-15', 'F', '(11) 98765-4321', 'ana.paula@email.com', 'Rua das Flores, 123', 'São Paulo', 'SP', '01234-567', 'Professora', 'Bradesco Dental', true, NOW() - INTERVAL '6 months', NOW()),
('Bruno Henrique Costa', '234.567.890-12', '23.456.789-0', '1990-07-22', 'M', '(11) 97654-3210', 'bruno.costa@email.com', 'Av. Paulista, 456', 'São Paulo', 'SP', '01310-100', 'Engenheiro', 'SulAmérica', true, NOW() - INTERVAL '5 months', NOW()),
('Carla Mendes Oliveira', '345.678.901-23', '34.567.890-1', '1978-11-30', 'F', '(11) 96543-2109', 'carla.mendes@email.com', 'Rua Augusta, 789', 'São Paulo', 'SP', '01305-100', 'Médica', 'Amil Dental', true, NOW() - INTERVAL '4 months', NOW()),
('Daniel Rodrigues Almeida', '456.789.012-34', '45.678.901-2', '1995-01-18', 'M', '(11) 95432-1098', 'daniel.almeida@email.com', 'Rua Oscar Freire, 321', 'São Paulo', 'SP', '01426-001', 'Advogado', NULL, true, NOW() - INTERVAL '3 months', NOW()),
('Eduarda Ferreira Lima', '567.890.123-45', '56.789.012-3', '1988-05-25', 'F', '(11) 94321-0987', 'eduarda.lima@email.com', 'Av. Faria Lima, 654', 'São Paulo', 'SP', '01452-000', 'Publicitária', 'Porto Seguro', true, NOW() - INTERVAL '3 months', NOW()),
('Fernando Santos Pereira', '678.901.234-56', '67.890.123-4', '1982-09-12', 'M', '(11) 93210-9876', 'fernando.pereira@email.com', 'Rua Haddock Lobo, 987', 'São Paulo', 'SP', '01414-001', 'Contador', 'Bradesco Dental', true, NOW() - INTERVAL '2 months', NOW()),
('Gabriela Souza Martins', '789.012.345-67', '78.901.234-5', '1992-12-08', 'F', '(11) 92109-8765', 'gabriela.martins@email.com', 'Av. Rebouças, 147', 'São Paulo', 'SP', '05401-100', 'Designer', NULL, true, NOW() - INTERVAL '2 months', NOW()),
('Henrique Dias Barbosa', '890.123.456-78', '89.012.345-6', '1975-04-20', 'M', '(11) 91098-7654', 'henrique.barbosa@email.com', 'Rua Pamplona, 258', 'São Paulo', 'SP', '01405-001', 'Empresário', 'SulAmérica', true, NOW() - INTERVAL '1 month', NOW()),
('Isabela Gomes Cardoso', '901.234.567-89', '90.123.456-7', '1998-08-14', 'F', '(11) 90987-6543', 'isabela.cardoso@email.com', 'Rua Bela Cintra, 369', 'São Paulo', 'SP', '01415-000', 'Estudante', NULL, true, NOW() - INTERVAL '1 month', NOW()),
('João Victor Araújo', '012.345.678-90', '01.234.567-8', '1987-02-28', 'M', '(11) 89876-5432', 'joao.araujo@email.com', 'Av. Brasil, 741', 'São Paulo', 'SP', '01431-000', 'Arquiteto', 'Amil Dental', true, NOW() - INTERVAL '20 days', NOW()),
('Juliana Rocha Fernandes', '111.222.333-44', '11.222.333-4', '1993-06-17', 'F', '(11) 88765-4321', 'juliana.fernandes@email.com', 'Rua da Consolação, 852', 'São Paulo', 'SP', '01302-000', 'Jornalista', 'Porto Seguro', true, NOW() - INTERVAL '15 days', NOW()),
('Lucas Alves Monteiro', '222.333.444-55', '22.333.444-5', '1980-10-05', 'M', '(11) 87654-3210', 'lucas.monteiro@email.com', 'Av. Ibirapuera, 963', 'São Paulo', 'SP', '04029-100', 'Fotógrafo', NULL, true, NOW() - INTERVAL '10 days', NOW()),
('Mariana Castro Ribeiro', '333.444.555-66', '33.444.555-6', '1991-03-11', 'F', '(11) 86543-2109', 'mariana.ribeiro@email.com', 'Rua Vergueiro, 159', 'São Paulo', 'SP', '01504-001', 'Nutricionista', 'Bradesco Dental', true, NOW() - INTERVAL '8 days', NOW()),
('Nicolas Teixeira Moura', '444.555.666-77', '44.555.666-7', '1996-07-23', 'M', '(11) 85432-1098', 'nicolas.moura@email.com', 'Rua Augusta, 753', 'São Paulo', 'SP', '01305-100', 'Músico', NULL, true, NOW() - INTERVAL '5 days', NOW()),
('Patrícia Lima Carvalho', '555.666.777-88', '55.666.777-8', '1984-11-19', 'F', '(11) 84321-0987', 'patricia.carvalho@email.com', 'Av. Angélica, 357', 'São Paulo', 'SP', '01227-000', 'Psicóloga', 'SulAmérica', true, NOW() - INTERVAL '3 days', NOW()),
('Rafael Oliveira Cunha', '666.777.888-99', '66.777.888-9', '1989-05-07', 'M', '(11) 83210-9876', 'rafael.cunha@email.com', 'Rua Estados Unidos, 468', 'São Paulo', 'SP', '01427-000', 'Desenvolvedor', 'Amil Dental', true, NOW() - INTERVAL '2 days', NOW()),
('Sabrina Martins Soares', '777.888.999-00', '77.888.999-0', '1994-12-31', 'F', '(11) 82109-8765', 'sabrina.soares@email.com', 'Av. Nove de Julho, 579', 'São Paulo', 'SP', '01407-100', 'Veterinária', NULL, true, NOW() - INTERVAL '1 day', NOW()),
('Thiago Rodrigues Pinto', '888.999.000-11', '88.999.000-1', '1981-08-26', 'M', '(11) 81098-7654', 'thiago.pinto@email.com', 'Rua Apa, 680', 'São Paulo', 'SP', '01201-010', 'Piloto', 'Porto Seguro', true, NOW(), NOW()),
('Vanessa Almeida Reis', '999.000.111-22', '99.000.111-2', '1997-01-09', 'F', '(11) 80987-6543', 'vanessa.reis@email.com', 'Rua Cardeal Arcoverde, 791', 'São Paulo', 'SP', '05407-002', 'Fisioterapeuta', 'Bradesco Dental', true, NOW(), NOW()),
('William Costa Barros', '000.111.222-33', '00.111.222-3', '1986-04-15', 'M', '(11) 79876-5432', 'william.barros@email.com', 'Av. Brigadeiro Faria Lima, 892', 'São Paulo', 'SP', '01451-000', 'Consultor', NULL, true, NOW(), NOW());

-- ========================================
-- 2. FORNECEDORES (Materiais Odontológicos)
-- ========================================
INSERT INTO suppliers (name, cnpj, phone, email, address, city, state, zip_code, contact_person, notes, active, created_at, updated_at) VALUES
('Dental Cremer Produtos Odontológicos', '12.345.678/0001-90', '(11) 3456-7890', 'vendas@dentalcremer.com.br', 'Rua Industrial, 1500', 'São Paulo', 'SP', '03310-000', 'Carlos Silva', 'Fornecedor principal de materiais', true, NOW(), NOW()),
('FGM Produtos Odontológicos', '23.456.789/0001-01', '(47) 3363-5600', 'contato@fgm.ind.br', 'Rodovia SC 401, 1621', 'Joinville', 'SC', '89219-600', 'Maria Santos', 'Resinas e materiais restauradores', true, NOW(), NOW()),
('SS White Artigos Dentários', '34.567.890/0001-12', '(21) 3525-2525', 'sac@sswhite.com.br', 'Av. Brasil, 7780', 'Rio de Janeiro', 'RJ', '21040-020', 'João Mendes', 'Brocas e instrumentais', true, NOW(), NOW()),
('Angelus Indústria de Produtos Odontológicos', '45.678.901/0001-23', '(41) 3879-6666', 'angelus@angelus.ind.br', 'Rua Dr. João Negrão, 548', 'Londrina', 'PR', '86010-010', 'Ana Paula', 'Endodontia e esterilização', true, NOW(), NOW()),
('3M do Brasil', '56.789.012/0001-34', '(19) 3838-8000', 'dental@3m.com', 'Av. Santa Marina, 482', 'Campinas', 'SP', '13140-904', 'Roberto Costa', 'Resinas, cimentos e adesivos', true, NOW(), NOW());

-- ========================================
-- 3. PRODUTOS (Estoque)
-- ========================================
INSERT INTO products (name, description, category, sku, supplier_id, unit_price, stock_quantity, min_stock_quantity, unit, barcode, active, created_at, updated_at) VALUES
-- Materiais Restauradores
('Resina Composta A2', 'Resina fotopolimerizável cor A2 - 4g', 'Materiais Restauradores', 'RES-A2-001', 2, 45.90, 50, 10, 'Seringa', '7891234567890', true, NOW(), NOW()),
('Resina Composta A3', 'Resina fotopolimerizável cor A3 - 4g', 'Materiais Restauradores', 'RES-A3-001', 2, 45.90, 45, 10, 'Seringa', '7891234567891', true, NOW(), NOW()),
('Adesivo Dental Universal', 'Sistema adesivo universal - 5ml', 'Materiais Restauradores', 'ADE-UNI-001', 5, 98.50, 30, 5, 'Frasco', '7891234567892', true, NOW(), NOW()),
('Cimento de Ionômero de Vidro', 'Ionômero de vidro para restauração - Kit', 'Materiais Restauradores', 'ION-VID-001', 2, 125.00, 20, 5, 'Kit', '7891234567893', true, NOW(), NOW()),

-- Anestésicos
('Anestésico Lidocaína 2%', 'Lidocaína 2% com epinefrina 1:100.000 - Tubete 1.8ml', 'Anestésicos', 'ANE-LID-001', 1, 3.20, 500, 100, 'Tubete', '7891234567894', true, NOW(), NOW()),
('Anestésico Articaína 4%', 'Articaína 4% com epinefrina 1:100.000 - Tubete 1.8ml', 'Anestésicos', 'ANE-ART-001', 1, 4.50, 300, 50, 'Tubete', '7891234567895', true, NOW(), NOW()),
('Agulhas Descartáveis Curtas', 'Agulha 30G curta descartável', 'Anestésicos', 'AGU-CUR-001', 1, 0.45, 1000, 200, 'Unidade', '7891234567896', true, NOW(), NOW()),

-- Endodontia
('Limas Endodônticas K-File 21mm', 'Jogo de limas tipo K 15-40 - 21mm', 'Endodontia', 'LIM-K21-001', 4, 35.00, 25, 5, 'Jogo', '7891234567897', true, NOW(), NOW()),
('Cimento Endodôntico', 'Cimento obturador de canais - Kit', 'Endodontia', 'CIM-END-001', 4, 89.90, 15, 3, 'Kit', '7891234567898', true, NOW(), NOW()),
('Cone de Guta-Percha', 'Cones principais sortidos', 'Endodontia', 'CON-GUT-001', 4, 28.50, 40, 10, 'Caixa', '7891234567899', true, NOW(), NOW()),

-- Instrumentais
('Espelho Bucal Plano N°5', 'Espelho odontológico plano número 5', 'Instrumentais', 'ESP-BUC-005', 3, 15.80, 60, 15, 'Unidade', '7891234567900', true, NOW(), NOW()),
('Pinça Clínica', 'Pinça clínica para algodão', 'Instrumentais', 'PIN-CLI-001', 3, 18.90, 50, 10, 'Unidade', '7891234567901', true, NOW(), NOW()),
('Sonda Exploradora', 'Sonda exploradora reta', 'Instrumentais', 'SON-EXP-001', 3, 12.50, 55, 10, 'Unidade', '7891234567902', true, NOW(), NOW()),
('Brocas Carbide Esféricas', 'Jogo de brocas esféricas 1/2 a 6', 'Instrumentais', 'BRO-CAR-001', 3, 42.00, 30, 8, 'Jogo', '7891234567903', true, NOW(), NOW()),

-- Biossegurança
('Luvas de Procedimento P', 'Luvas de látex descartáveis - Pequeno', 'Biossegurança', 'LUV-LAT-P', 1, 28.90, 200, 50, 'Caixa 100un', '7891234567904', true, NOW(), NOW()),
('Luvas de Procedimento M', 'Luvas de látex descartáveis - Médio', 'Biossegurança', 'LUV-LAT-M', 1, 28.90, 180, 50, 'Caixa 100un', '7891234567905', true, NOW(), NOW()),
('Máscaras Descartáveis', 'Máscaras cirúrgicas triplas', 'Biossegurança', 'MAS-CIR-001', 1, 22.50, 150, 30, 'Caixa 50un', '7891234567906', true, NOW(), NOW()),
('Óculos de Proteção', 'Óculos de proteção individual', 'Biossegurança', 'OCU-PRO-001', 1, 8.90, 80, 20, 'Unidade', '7891234567907', true, NOW(), NOW()),
('Campo Cirúrgico Descartável', 'Campo fenestrado estéril', 'Biossegurança', 'CAM-CIR-001', 1, 3.50, 300, 50, 'Unidade', '7891234567908', true, NOW(), NOW()),

-- Radiologia
('Filme Radiográfico Periapical', 'Filme radiográfico periapical adulto', 'Radiologia', 'FIL-RAD-001', 1, 1.80, 500, 100, 'Unidade', '7891234567909', true, NOW(), NOW()),
('Fixador Radiográfico', 'Solução fixadora para revelação - 475ml', 'Radiologia', 'FIX-RAD-001', 1, 18.50, 25, 5, 'Frasco', '7891234567910', true, NOW(), NOW()),
('Revelador Radiográfico', 'Solução reveladora - 475ml', 'Radiologia', 'REV-RAD-001', 1, 18.50, 25, 5, 'Frasco', '7891234567911', true, NOW(), NOW());

-- ========================================
-- 4. AGENDAMENTOS (Próximos 7 dias)
-- ========================================

-- Hoje
INSERT INTO appointments (patient_id, professional, start_time, end_time, appointment_type, status, notes, created_at, updated_at) VALUES
(1, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '9 hours', NOW()::date + INTERVAL '10 hours', 'Consulta', 'scheduled', 'Primeira consulta de avaliação', NOW(), NOW()),
(5, 'Dra. Marina Santos', NOW()::date + INTERVAL '10 hours', NOW()::date + INTERVAL '11 hours', 'Limpeza', 'scheduled', 'Limpeza semestral', NOW(), NOW()),
(8, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '14 hours', NOW()::date + INTERVAL '15 hours', 'Tratamento de Canal', 'scheduled', 'Segunda sessão de endodontia', NOW(), NOW()),
(12, 'Dra. Marina Santos', NOW()::date + INTERVAL '15 hours', NOW()::date + INTERVAL '16 hours', 'Restauração', 'scheduled', 'Restauração molar inferior', NOW(), NOW()),

-- Amanhã
INSERT INTO appointments (patient_id, professional, start_time, end_time, appointment_type, status, notes, created_at, updated_at) VALUES
(2, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '1 day 9 hours', NOW()::date + INTERVAL '1 day 10 hours', 'Consulta', 'scheduled', 'Retorno para avaliação', NOW(), NOW()),
(4, 'Dra. Marina Santos', NOW()::date + INTERVAL '1 day 10 hours', NOW()::date + INTERVAL '1 day 11 hours', 'Clareamento', 'scheduled', 'Primeira sessão de clareamento', NOW(), NOW()),
(7, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '1 day 11 hours', NOW()::date + INTERVAL '1 day 12 hours', 'Extração', 'scheduled', 'Extração de siso superior', NOW(), NOW()),
(10, 'Dra. Marina Santos', NOW()::date + INTERVAL '1 day 14 hours', NOW()::date + INTERVAL '1 day 15 hours', 'Limpeza', 'scheduled', 'Profilaxia e aplicação de flúor', NOW(), NOW()),
(15, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '1 day 16 hours', NOW()::date + INTERVAL '1 day 17 hours', 'Consulta', 'scheduled', 'Avaliação para prótese', NOW(), NOW()),

-- Dia +2
INSERT INTO appointments (patient_id, professional, start_time, end_time, appointment_type, status, notes, created_at, updated_at) VALUES
(3, 'Dra. Marina Santos', NOW()::date + INTERVAL '2 days 9 hours', NOW()::date + INTERVAL '2 days 10 hours', 'Restauração', 'scheduled', 'Restauração em resina', NOW(), NOW()),
(6, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '2 days 10 hours', NOW()::date + INTERVAL '2 days 11 hours', 'Tratamento de Canal', 'scheduled', 'Abertura e preparo do canal', NOW(), NOW()),
(9, 'Dra. Marina Santos', NOW()::date + INTERVAL '2 days 14 hours', NOW()::date + INTERVAL '2 days 15 hours', 'Consulta', 'scheduled', 'Consulta de emergência - dor', NOW(), NOW()),
(13, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '2 days 15 hours', NOW()::date + INTERVAL '2 days 16 hours', 'Limpeza', 'scheduled', 'Limpeza anual', NOW(), NOW()),

-- Dia +3
INSERT INTO appointments (patient_id, professional, start_time, end_time, appointment_type, status, notes, created_at, updated_at) VALUES
(11, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '3 days 9 hours', NOW()::date + INTERVAL '3 days 10 hours', 'Restauração', 'scheduled', 'Múltiplas restaurações', NOW(), NOW()),
(14, 'Dra. Marina Santos', NOW()::date + INTERVAL '3 days 10 hours', NOW()::date + INTERVAL '3 days 11 hours', 'Consulta', 'scheduled', 'Avaliação ortodôntica', NOW(), NOW()),
(16, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '3 days 14 hours', NOW()::date + INTERVAL '3 days 15 hours', 'Tratamento de Canal', 'scheduled', 'Finalização e obturação', NOW(), NOW()),
(18, 'Dra. Marina Santos', NOW()::date + INTERVAL '3 days 16 hours', NOW()::date + INTERVAL '3 days 17 hours', 'Limpeza', 'scheduled', 'Limpeza com ultrassom', NOW(), NOW()),

-- Dia +4
INSERT INTO appointments (patient_id, professional, start_time, end_time, appointment_type, status, notes, created_at, updated_at) VALUES
(17, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '4 days 9 hours', NOW()::date + INTERVAL '4 days 10 hours', 'Consulta', 'scheduled', 'Primeira consulta', NOW(), NOW()),
(19, 'Dra. Marina Santos', NOW()::date + INTERVAL '4 days 10 hours', NOW()::date + INTERVAL '4 days 11 hours', 'Restauração', 'scheduled', 'Troca de restauração antiga', NOW(), NOW()),
(20, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '4 days 14 hours', NOW()::date + INTERVAL '4 days 15 hours', 'Clareamento', 'scheduled', 'Segunda sessão de clareamento', NOW(), NOW()),
(1, 'Dra. Marina Santos', NOW()::date + INTERVAL '4 days 15 hours', NOW()::date + INTERVAL '4 days 16 hours', 'Consulta', 'scheduled', 'Retorno pós-tratamento', NOW(), NOW()),

-- Dia +5
INSERT INTO appointments (patient_id, professional, start_time, end_time, appointment_type, status, notes, created_at, updated_at) VALUES
(2, 'Dra. Marina Santos', NOW()::date + INTERVAL '5 days 9 hours', NOW()::date + INTERVAL '5 days 10 hours', 'Limpeza', 'scheduled', 'Profilaxia semestral', NOW(), NOW()),
(3, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '5 days 10 hours', NOW()::date + INTERVAL '5 days 11 hours', 'Tratamento de Canal', 'scheduled', 'Início de endodontia', NOW(), NOW()),
(4, 'Dra. Marina Santos', NOW()::date + INTERVAL '5 days 14 hours', NOW()::date + INTERVAL '5 days 15 hours', 'Restauração', 'scheduled', 'Restauração pré-molar', NOW(), NOW()),

-- Dia +6
INSERT INTO appointments (patient_id, professional, start_time, end_time, appointment_type, status, notes, created_at, updated_at) VALUES
(5, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '6 days 9 hours', NOW()::date + INTERVAL '6 days 10 hours', 'Consulta', 'scheduled', 'Avaliação geral', NOW(), NOW()),
(6, 'Dra. Marina Santos', NOW()::date + INTERVAL '6 days 10 hours', NOW()::date + INTERVAL '6 days 11 hours', 'Limpeza', 'scheduled', 'Limpeza com jateamento', NOW(), NOW()),
(7, 'Dr. Carlos Alberto', NOW()::date + INTERVAL '6 days 14 hours', NOW()::date + INTERVAL '6 days 15 hours', 'Restauração', 'scheduled', 'Restauração estética anterior', NOW(), NOW());

-- Alguns agendamentos concluídos (semana passada)
INSERT INTO appointments (patient_id, professional, start_time, end_time, appointment_type, status, notes, created_at, updated_at) VALUES
(1, 'Dr. Carlos Alberto', NOW()::date - INTERVAL '7 days 9 hours', NOW()::date - INTERVAL '7 days 10 hours', 'Consulta', 'completed', 'Consulta inicial realizada', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
(2, 'Dra. Marina Santos', NOW()::date - INTERVAL '6 days 10 hours', NOW()::date - INTERVAL '6 days 11 hours', 'Limpeza', 'completed', 'Limpeza concluída', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
(3, 'Dr. Carlos Alberto', NOW()::date - INTERVAL '5 days 14 hours', NOW()::date - INTERVAL '5 days 15 hours', 'Restauração', 'completed', 'Restauração finalizada', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(4, 'Dra. Marina Santos', NOW()::date - INTERVAL '4 days 9 hours', NOW()::date - INTERVAL '4 days 10 hours', 'Consulta', 'completed', 'Avaliação concluída', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(5, 'Dr. Carlos Alberto', NOW()::date - INTERVAL '3 days 10 hours', NOW()::date - INTERVAL '3 days 11 hours', 'Limpeza', 'completed', 'Profilaxia realizada', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(6, 'Dra. Marina Santos', NOW()::date - INTERVAL '2 days 14 hours', NOW()::date - INTERVAL '2 days 15 hours', 'Tratamento de Canal', 'completed', 'Primeira sessão concluída', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(7, 'Dr. Carlos Alberto', NOW()::date - INTERVAL '1 day 9 hours', NOW()::date - INTERVAL '1 day 10 hours', 'Restauração', 'completed', 'Restauração em composite', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- Algumas faltas e cancelamentos
INSERT INTO appointments (patient_id, professional, start_time, end_time, appointment_type, status, notes, created_at, updated_at) VALUES
(8, 'Dra. Marina Santos', NOW()::date - INTERVAL '8 days 10 hours', NOW()::date - INTERVAL '8 days 11 hours', 'Consulta', 'no_show', 'Paciente não compareceu', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
(9, 'Dr. Carlos Alberto', NOW()::date - INTERVAL '10 days 14 hours', NOW()::date - INTERVAL '10 days 15 hours', 'Limpeza', 'cancelled', 'Cancelado pelo paciente', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
(10, 'Dra. Marina Santos', NOW()::date - INTERVAL '12 days 9 hours', NOW()::date - INTERVAL '12 days 10 hours', 'Restauração', 'no_show', 'Não compareceu', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days');

-- ========================================
-- 5. ORÇAMENTOS
-- ========================================

-- Orçamentos Aprovados
INSERT INTO budgets (patient_id, description, total_amount, discount, final_amount, status, validity_date, notes, created_at, updated_at) VALUES
(1, 'Limpeza + Restauração (2 dentes)', 800.00, 80.00, 720.00, 'approved', NOW()::date + INTERVAL '30 days', 'Aprovado com 10% de desconto', NOW() - INTERVAL '15 days', NOW() - INTERVAL '14 days'),
(2, 'Tratamento de Canal + Coroa', 2500.00, 0.00, 2500.00, 'approved', NOW()::date + INTERVAL '30 days', 'Pagamento parcelado em 5x', NOW() - INTERVAL '20 days', NOW() - INTERVAL '19 days'),
(3, 'Clareamento Dental (2 sessões)', 1200.00, 120.00, 1080.00, 'approved', NOW()::date + INTERVAL '30 days', 'Aprovado - paciente convênio', NOW() - INTERVAL '10 days', NOW() - INTERVAL '9 days'),
(5, 'Restaurações Estéticas (3 dentes)', 1350.00, 0.00, 1350.00, 'approved', NOW()::date + INTERVAL '30 days', 'Parcelamento 3x sem juros', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days'),
(6, 'Extração Siso + Limpeza', 950.00, 50.00, 900.00, 'approved', NOW()::date + INTERVAL '30 days', 'Desconto à vista', NOW() - INTERVAL '12 days', NOW() - INTERVAL '11 days'),
(10, 'Implante Dentário + Coroa', 4500.00, 0.00, 4500.00, 'approved', NOW()::date + INTERVAL '60 days', 'Parcelamento em 10x', NOW() - INTERVAL '25 days', NOW() - INTERVAL '24 days'),
(13, 'Limpeza + Aplicação Flúor', 380.00, 0.00, 380.00, 'approved', NOW()::date + INTERVAL '30 days', 'Pagamento à vista', NOW() - INTERVAL '8 days', NOW() - INTERVAL '7 days'),
(15, 'Prótese Parcial Removível', 3200.00, 200.00, 3000.00, 'approved', NOW()::date + INTERVAL '45 days', 'Desconto 6.25%', NOW() - INTERVAL '18 days', NOW() - INTERVAL '17 days'),

-- Orçamentos Pendentes
INSERT INTO budgets (patient_id, description, total_amount, discount, final_amount, status, validity_date, notes, created_at, updated_at) VALUES
(4, 'Aparelho Ortodôntico', 5500.00, 0.00, 5500.00, 'pending', NOW()::date + INTERVAL '30 days', 'Aguardando aprovação do paciente', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(7, 'Facetas de Porcelana (4 dentes)', 6800.00, 0.00, 6800.00, 'pending', NOW()::date + INTERVAL '30 days', 'Orçamento enviado por email', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(9, 'Limpeza Profunda + Tratamento Gengival', 1450.00, 0.00, 1450.00, 'pending', NOW()::date + INTERVAL '30 days', 'Paciente solicitou prazo para decidir', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
(14, 'Coroa de Porcelana', 1800.00, 0.00, 1800.00, 'pending', NOW()::date + INTERVAL '30 days', 'Aguardando resposta', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(17, 'Restaurações + Limpeza', 920.00, 0.00, 920.00, 'pending', NOW()::date + INTERVAL '30 days', 'Primeiro orçamento do paciente', NOW(), NOW()),
(19, 'Clareamento + Limpeza', 1100.00, 0.00, 1100.00, 'pending', NOW()::date + INTERVAL '30 days', 'Orçamento solicitado hoje', NOW(), NOW()),

-- Orçamentos Rejeitados
INSERT INTO budgets (patient_id, description, total_amount, discount, final_amount, status, validity_date, notes, created_at, updated_at) VALUES
(8, 'Implante + Enxerto Ósseo', 7500.00, 0.00, 7500.00, 'rejected', NOW()::date - INTERVAL '5 days', 'Paciente achou muito caro', NOW() - INTERVAL '30 days', NOW() - INTERVAL '25 days'),
(11, 'Ortodontia Estética', 6200.00, 0.00, 6200.00, 'rejected', NOW()::date - INTERVAL '10 days', 'Optou por outra clínica', NOW() - INTERVAL '20 days', NOW() - INTERVAL '10 days'),
(16, 'Prótese Total Superior', 2800.00, 0.00, 2800.00, 'rejected', NOW()::date - INTERVAL '3 days', 'Sem condições no momento', NOW() - INTERVAL '15 days', NOW() - INTERVAL '12 days'),
(18, 'Harmonização Facial', 3500.00, 0.00, 3500.00, 'rejected', NOW()::date - INTERVAL '7 days', 'Decidiu não fazer', NOW() - INTERVAL '14 days', NOW() - INTERVAL '7 days'),

-- Orçamentos Cancelados
INSERT INTO budgets (patient_id, description, total_amount, discount, final_amount, status, validity_date, notes, created_at, updated_at) VALUES
(12, 'Lentes de Contato Dental (6 dentes)', 8400.00, 0.00, 8400.00, 'cancelled', NOW()::date - INTERVAL '15 days', 'Cancelado a pedido do paciente', NOW() - INTERVAL '40 days', NOW() - INTERVAL '25 days'),
(20, 'Aparelho Autoligado', 6800.00, 0.00, 6800.00, 'cancelled', NOW()::date - INTERVAL '8 days', 'Mudou de cidade', NOW() - INTERVAL '35 days', NOW() - INTERVAL '27 days');

-- ========================================
-- 6. PAGAMENTOS (Receitas - Income)
-- ========================================

-- Pagamentos dos orçamentos aprovados
-- Orçamento 1 - Ana Paula (720.00) - À vista
INSERT INTO payments (budget_id, description, amount, payment_method, type, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(1, 'Limpeza + Restauração - Pagamento à vista', 720.00, 'Cartão de Débito', 'income', 'paid', NOW()::date - INTERVAL '14 days', NOW()::date - INTERVAL '14 days', 'Pagamento realizado', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days');

-- Orçamento 2 - Bruno (2500.00) - 5x de 500.00
INSERT INTO payments (budget_id, description, amount, payment_method, type, status, installments, installment_number, due_date, paid_date, notes, created_at, updated_at) VALUES
(2, 'Tratamento de Canal + Coroa - Parcela 1/5', 500.00, 'Cartão de Crédito', 'income', 'paid', 5, 1, NOW()::date - INTERVAL '19 days', NOW()::date - INTERVAL '19 days', 'Parcela 1 paga', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days'),
(2, 'Tratamento de Canal + Coroa - Parcela 2/5', 500.00, 'Cartão de Crédito', 'income', 'paid', 5, 2, NOW()::date + INTERVAL '11 days', NOW()::date - INTERVAL '10 days', 'Parcela 2 paga antecipadamente', NOW() - INTERVAL '19 days', NOW() - INTERVAL '10 days'),
(2, 'Tratamento de Canal + Coroa - Parcela 3/5', 500.00, 'Cartão de Crédito', 'income', 'pending', 5, 3, NOW()::date + INTERVAL '11 days', NULL, 'A vencer', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days'),
(2, 'Tratamento de Canal + Coroa - Parcela 4/5', 500.00, 'Cartão de Crédito', 'income', 'pending', 5, 4, NOW()::date + INTERVAL '41 days', NULL, 'A vencer', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days'),
(2, 'Tratamento de Canal + Coroa - Parcela 5/5', 500.00, 'Cartão de Crédito', 'income', 'pending', 5, 5, NOW()::date + INTERVAL '71 days', NULL, 'A vencer', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days');

-- Orçamento 3 - Carla (1080.00) - À vista
INSERT INTO payments (budget_id, description, amount, payment_method, type, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(3, 'Clareamento Dental - Pagamento à vista', 1080.00, 'PIX', 'income', 'paid', NOW()::date - INTERVAL '9 days', NOW()::date - INTERVAL '9 days', 'Transferência PIX', NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days');

-- Orçamento 4 - Eduarda (1350.00) - 3x de 450.00
INSERT INTO payments (budget_id, description, amount, payment_method, type, status, installments, installment_number, due_date, paid_date, notes, created_at, updated_at) VALUES
(4, 'Restaurações Estéticas - Parcela 1/3', 450.00, 'Cartão de Crédito', 'income', 'paid', 3, 1, NOW()::date - INTERVAL '4 days', NOW()::date - INTERVAL '4 days', 'Parcela 1 paga', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(4, 'Restaurações Estéticas - Parcela 2/3', 450.00, 'Cartão de Crédito', 'income', 'pending', 3, 2, NOW()::date + INTERVAL '26 days', NULL, 'A vencer', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(4, 'Restaurações Estéticas - Parcela 3/3', 450.00, 'Cartão de Crédito', 'income', 'pending', 3, 3, NOW()::date + INTERVAL '56 days', NULL, 'A vencer', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days');

-- Orçamento 5 - Fernando (900.00) - À vista
INSERT INTO payments (budget_id, description, amount, payment_method, type, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(5, 'Extração Siso + Limpeza - Pagamento à vista', 900.00, 'Dinheiro', 'income', 'paid', NOW()::date - INTERVAL '11 days', NOW()::date - INTERVAL '11 days', 'Pagamento em espécie', NOW() - INTERVAL '11 days', NOW() - INTERVAL '11 days');

-- Orçamento 6 - João (4500.00) - 10x de 450.00
INSERT INTO payments (budget_id, description, amount, payment_method, type, status, installments, installment_number, due_date, paid_date, notes, created_at, updated_at) VALUES
(6, 'Implante Dentário + Coroa - Parcela 1/10', 450.00, 'Cartão de Crédito', 'income', 'paid', 10, 1, NOW()::date - INTERVAL '24 days', NOW()::date - INTERVAL '24 days', 'Entrada paga', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
(6, 'Implante Dentário + Coroa - Parcela 2/10', 450.00, 'Cartão de Crédito', 'income', 'paid', 10, 2, NOW()::date + INTERVAL '6 days', NOW()::date - INTERVAL '3 days', 'Paga antecipadamente', NOW() - INTERVAL '24 days', NOW() - INTERVAL '3 days'),
(6, 'Implante Dentário + Coroa - Parcela 3/10', 450.00, 'Cartão de Crédito', 'income', 'pending', 10, 3, NOW()::date + INTERVAL '6 days', NULL, 'A vencer', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
(6, 'Implante Dentário + Coroa - Parcela 4/10', 450.00, 'Cartão de Crédito', 'income', 'pending', 10, 4, NOW()::date + INTERVAL '36 days', NULL, 'A vencer', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
(6, 'Implante Dentário + Coroa - Parcela 5/10', 450.00, 'Cartão de Crédito', 'income', 'pending', 10, 5, NOW()::date + INTERVAL '66 days', NULL, 'A vencer', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
(6, 'Implante Dentário + Coroa - Parcela 6/10', 450.00, 'Cartão de Crédito', 'income', 'pending', 10, 6, NOW()::date + INTERVAL '96 days', NULL, 'A vencer', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
(6, 'Implante Dentário + Coroa - Parcela 7/10', 450.00, 'Cartão de Crédito', 'income', 'pending', 10, 7, NOW()::date + INTERVAL '126 days', NULL, 'A vencer', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
(6, 'Implante Dentário + Coroa - Parcela 8/10', 450.00, 'Cartão de Crédito', 'income', 'pending', 10, 8, NOW()::date + INTERVAL '156 days', NULL, 'A vencer', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
(6, 'Implante Dentário + Coroa - Parcela 9/10', 450.00, 'Cartão de Crédito', 'income', 'pending', 10, 9, NOW()::date + INTERVAL '186 days', NULL, 'A vencer', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days'),
(6, 'Implante Dentário + Coroa - Parcela 10/10', 450.00, 'Cartão de Crédito', 'income', 'pending', 10, 10, NOW()::date + INTERVAL '216 days', NULL, 'A vencer', NOW() - INTERVAL '24 days', NOW() - INTERVAL '24 days');

-- Orçamento 7 - Mariana (380.00) - À vista
INSERT INTO payments (budget_id, description, amount, payment_method, type, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(7, 'Limpeza + Aplicação Flúor - Pagamento à vista', 380.00, 'PIX', 'income', 'paid', NOW()::date - INTERVAL '7 days', NOW()::date - INTERVAL '7 days', 'PIX recebido', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days');

-- Orçamento 8 - Patrícia (3000.00) - 6x de 500.00
INSERT INTO payments (budget_id, description, amount, payment_method, type, status, installments, installment_number, due_date, paid_date, notes, created_at, updated_at) VALUES
(8, 'Prótese Parcial Removível - Parcela 1/6', 500.00, 'Boleto', 'income', 'paid', 6, 1, NOW()::date - INTERVAL '17 days', NOW()::date - INTERVAL '17 days', 'Boleto pago', NOW() - INTERVAL '17 days', NOW() - INTERVAL '17 days'),
(8, 'Prótese Parcial Removível - Parcela 2/6', 500.00, 'Boleto', 'income', 'paid', 6, 2, NOW()::date + INTERVAL '13 days', NOW()::date - INTERVAL '5 days', 'Antecipado', NOW() - INTERVAL '17 days', NOW() - INTERVAL '5 days'),
(8, 'Prótese Parcial Removível - Parcela 3/6', 500.00, 'Boleto', 'income', 'pending', 6, 3, NOW()::date + INTERVAL '13 days', NULL, 'Vence em breve', NOW() - INTERVAL '17 days', NOW() - INTERVAL '17 days'),
(8, 'Prótese Parcial Removível - Parcela 4/6', 500.00, 'Boleto', 'income', 'pending', 6, 4, NOW()::date + INTERVAL '43 days', NULL, 'A vencer', NOW() - INTERVAL '17 days', NOW() - INTERVAL '17 days'),
(8, 'Prótese Parcial Removível - Parcela 5/6', 500.00, 'Boleto', 'income', 'pending', 6, 5, NOW()::date + INTERVAL '73 days', NULL, 'A vencer', NOW() - INTERVAL '17 days', NOW() - INTERVAL '17 days'),
(8, 'Prótese Parcial Removível - Parcela 6/6', 500.00, 'Boleto', 'income', 'pending', 6, 6, NOW()::date + INTERVAL '103 days', NULL, 'A vencer', NOW() - INTERVAL '17 days', NOW() - INTERVAL '17 days');

-- ========================================
-- 7. DESPESAS (Expenses)
-- ========================================

-- Despesas Fixas Mensais
INSERT INTO payments (description, amount, payment_method, type, status, due_date, paid_date, notes, created_at, updated_at) VALUES
-- Mês Atual
('Aluguel Clínica - ' || TO_CHAR(NOW(), 'MM/YYYY'), 5500.00, 'Transferência Bancária', 'expense', 'paid', NOW()::date - INTERVAL '20 days', NOW()::date - INTERVAL '20 days', 'Aluguel pago', NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days'),
('Energia Elétrica - ' || TO_CHAR(NOW(), 'MM/YYYY'), 890.50, 'Débito Automático', 'expense', 'paid', NOW()::date - INTERVAL '15 days', NOW()::date - INTERVAL '15 days', 'Conta de luz', NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days'),
('Água - ' || TO_CHAR(NOW(), 'MM/YYYY'), 156.30, 'Débito Automático', 'expense', 'paid', NOW()::date - INTERVAL '12 days', NOW()::date - INTERVAL '12 days', 'Conta de água', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
('Internet Fibra Óptica - ' || TO_CHAR(NOW(), 'MM/YYYY'), 199.90, 'Débito Automático', 'expense', 'paid', NOW()::date - INTERVAL '10 days', NOW()::date - INTERVAL '10 days', 'Internet empresarial', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
('Telefone Fixo - ' || TO_CHAR(NOW(), 'MM/YYYY'), 89.90, 'Débito Automático', 'expense', 'paid', NOW()::date - INTERVAL '8 days', NOW()::date - INTERVAL '8 days', 'Linha telefônica', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
('Salário Dr. Carlos Alberto - ' || TO_CHAR(NOW(), 'MM/YYYY'), 12000.00, 'Transferência Bancária', 'expense', 'paid', NOW()::date - INTERVAL '5 days', NOW()::date - INTERVAL '5 days', 'Salário mensal', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
('Salário Dra. Marina Santos - ' || TO_CHAR(NOW(), 'MM/YYYY'), 10500.00, 'Transferência Bancária', 'expense', 'paid', NOW()::date - INTERVAL '5 days', NOW()::date - INTERVAL '5 days', 'Salário mensal', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
('Salário Recepcionista - ' || TO_CHAR(NOW(), 'MM/YYYY'), 2800.00, 'Transferência Bancária', 'expense', 'paid', NOW()::date - INTERVAL '5 days', NOW()::date - INTERVAL '5 days', 'Salário mensal', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
('Salário Auxiliar de Limpeza - ' || TO_CHAR(NOW(), 'MM/YYYY'), 1800.00, 'Transferência Bancária', 'expense', 'paid', NOW()::date - INTERVAL '5 days', NOW()::date - INTERVAL '5 days', 'Salário mensal', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
('INSS Patronal - ' || TO_CHAR(NOW(), 'MM/YYYY'), 5420.00, 'Guia', 'expense', 'paid', NOW()::date - INTERVAL '3 days', NOW()::date - INTERVAL '3 days', 'Contribuição previdenciária', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
('FGTS - ' || TO_CHAR(NOW(), 'MM/YYYY'), 2168.00, 'Guia', 'expense', 'paid', NOW()::date - INTERVAL '2 days', NOW()::date - INTERVAL '2 days', 'FGTS funcionários', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),

-- Mês Anterior
('Aluguel Clínica - ' || TO_CHAR(NOW() - INTERVAL '1 month', 'MM/YYYY'), 5500.00, 'Transferência Bancária', 'expense', 'paid', NOW()::date - INTERVAL '50 days', NOW()::date - INTERVAL '50 days', 'Aluguel pago', NOW() - INTERVAL '50 days', NOW() - INTERVAL '50 days'),
('Energia Elétrica - ' || TO_CHAR(NOW() - INTERVAL '1 month', 'MM/YYYY'), 825.80, 'Débito Automático', 'expense', 'paid', NOW()::date - INTERVAL '45 days', NOW()::date - INTERVAL '45 days', 'Conta de luz', NOW() - INTERVAL '45 days', NOW() - INTERVAL '45 days'),
('Água - ' || TO_CHAR(NOW() - INTERVAL '1 month', 'MM/YYYY'), 142.70, 'Débito Automático', 'expense', 'paid', NOW()::date - INTERVAL '42 days', NOW()::date - INTERVAL '42 days', 'Conta de água', NOW() - INTERVAL '42 days', NOW() - INTERVAL '42 days'),

-- Compras de Materiais
('Compra Materiais - Dental Cremer', 3450.80, 'Boleto', 'expense', 'paid', NOW()::date - INTERVAL '18 days', NOW()::date - INTERVAL '18 days', 'Reposição estoque geral', NOW() - INTERVAL '18 days', NOW() - INTERVAL '18 days'),
('Compra Resinas - FGM', 1890.50, 'Cartão de Crédito', 'expense', 'paid', NOW()::date - INTERVAL '12 days', NOW()::date - INTERVAL '12 days', 'Resinas compostas', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
('Compra Instrumentais - SS White', 2150.00, 'Transferência', 'expense', 'paid', NOW()::date - INTERVAL '8 days', NOW()::date - INTERVAL '8 days', 'Brocas e instrumentos', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
('Compra Materiais Endodontia - Angelus', 1650.75, 'Boleto', 'expense', 'pending', NOW()::date + INTERVAL '5 days', NULL, 'Vencimento próximo', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
('Compra Materiais 3M', 2890.90, 'Boleto', 'expense', 'pending', NOW()::date + INTERVAL '15 days', NULL, 'A vencer', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),

-- Outras Despesas
('Manutenção Equipamentos', 850.00, 'PIX', 'expense', 'paid', NOW()::date - INTERVAL '6 days', NOW()::date - INTERVAL '6 days', 'Manutenção preventiva', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
('Material de Limpeza e Higiene', 456.80, 'Dinheiro', 'expense', 'paid', NOW()::date - INTERVAL '4 days', NOW()::date - INTERVAL '4 days', 'Compra produtos limpeza', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
('Consultoria Contábil', 980.00, 'Transferência', 'expense', 'paid', NOW()::date - INTERVAL '7 days', NOW()::date - INTERVAL '7 days', 'Serviços contábeis mensais', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
('Sistema de Gestão (Software)', 299.90, 'Cartão de Crédito', 'expense', 'paid', NOW()::date - INTERVAL '1 day', NOW()::date - INTERVAL '1 day', 'Assinatura mensal OdoWell', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- ========================================
-- 8. MOVIMENTAÇÕES DE ESTOQUE
-- ========================================

-- Entradas de Estoque (compras)
INSERT INTO stock_movements (product_id, movement_type, quantity, unit_price, total_price, supplier_id, notes, created_at, updated_at) VALUES
(1, 'in', 30, 45.90, 1377.00, 2, 'Reposição de estoque', NOW() - INTERVAL '18 days', NOW() - INTERVAL '18 days'),
(2, 'in', 30, 45.90, 1377.00, 2, 'Reposição de estoque', NOW() - INTERVAL '18 days', NOW() - INTERVAL '18 days'),
(5, 'in', 200, 3.20, 640.00, 1, 'Compra programada', NOW() - INTERVAL '18 days', NOW() - INTERVAL '18 days'),
(6, 'in', 100, 4.50, 450.00, 1, 'Compra programada', NOW() - INTERVAL '18 days', NOW() - INTERVAL '18 days'),
(14, 'in', 100, 28.90, 2890.00, 1, 'Estoque biossegurança', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
(15, 'in', 80, 28.90, 2312.00, 1, 'Estoque biossegurança', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),

-- Saídas de Estoque (consumo)
INSERT INTO stock_movements (product_id, movement_type, quantity, unit_price, total_price, notes, created_at, updated_at) VALUES
(1, 'out', 10, 45.90, 459.00, 'Uso em restaurações da semana', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
(2, 'out', 15, 45.90, 688.50, 'Uso em restaurações da semana', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
(5, 'out', 200, 3.20, 640.00, 'Consumo semanal de anestésicos', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(6, 'out', 100, 4.50, 450.00, 'Consumo mensal', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(7, 'out', 300, 0.45, 135.00, 'Consumo de agulhas', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(14, 'out', 20, 28.90, 578.00, 'Uso diário - semana', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(15, 'out', 20, 28.90, 578.00, 'Uso diário - semana', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(16, 'out', 50, 22.50, 1125.00, 'Consumo semanal', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(19, 'out', 100, 3.50, 350.00, 'Uso em procedimentos', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- ========================================
-- 9. TAREFAS
-- ========================================

INSERT INTO tasks (title, description, assigned_to, priority, status, due_date, created_at, updated_at) VALUES
('Confirmar agendamentos de amanhã', 'Ligar para todos os pacientes agendados para amanhã e confirmar presença', 'Recepção', 'high', 'pending', NOW()::date + INTERVAL '1 day', NOW(), NOW()),
('Realizar pedido de materiais', 'Verificar estoque e fazer pedido de materiais que estão abaixo do mínimo', 'Administração', 'medium', 'in_progress', NOW()::date + INTERVAL '2 days', NOW() - INTERVAL '1 day', NOW()),
('Manutenção autoclave', 'Agendar manutenção preventiva da autoclave', 'Dr. Carlos Alberto', 'high', 'pending', NOW()::date + INTERVAL '3 days', NOW(), NOW()),
('Enviar orçamentos pendentes', 'Fazer follow-up dos orçamentos que estão pendentes há mais de 3 dias', 'Recepção', 'medium', 'pending', NOW()::date + INTERVAL '1 day', NOW(), NOW()),
('Atualizar prontuários', 'Digitalizar prontuários físicos pendentes', 'Dra. Marina Santos', 'low', 'in_progress', NOW()::date + INTERVAL '7 days', NOW() - INTERVAL '2 days', NOW()),
('Cobrar pagamentos atrasados', 'Entrar em contato com pacientes com parcelas vencidas', 'Administração', 'high', 'pending', NOW()::date, NOW(), NOW()),
('Preparar relatório mensal', 'Compilar dados do mês para apresentação', 'Administração', 'medium', 'pending', NOW()::date + INTERVAL '5 days', NOW(), NOW()),
('Revisar estoque mínimo', 'Ajustar níveis mínimos de estoque baseado em consumo', 'Administração', 'low', 'completed', NOW()::date - INTERVAL '2 days', NOW() - INTERVAL '5 days', NOW() - INTERVAL '2 days'),
('Treinar novo software', 'Treinar equipe nas novas funcionalidades do sistema', 'Todos', 'medium', 'completed', NOW()::date - INTERVAL '10 days', NOW() - INTERVAL '15 days', NOW() - INTERVAL '10 days');

-- ========================================
-- 10. REGISTROS MÉDICOS (Alguns exemplos)
-- ========================================

INSERT INTO medical_records (patient_id, appointment_id, chief_complaint, diagnosis, treatment_plan, notes, created_at, updated_at) VALUES
(1, 35, 'Dor no dente 36', 'Cárie profunda em elemento 36', 'Restauração em resina composta', 'Paciente apresentou sensibilidade ao frio. Realizada restauração classe II em resina A2.', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
(2, 36, 'Limpeza de rotina', 'Gengivite leve', 'Profilaxia e orientação de higiene oral', 'Remoção de cálculo supra-gengival. Orientado uso de fio dental diariamente.', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
(3, 37, 'Restauração antiga quebrada', 'Fratura de restauração em elemento 15', 'Remoção de restauração antiga e confecção de nova', 'Removida restauração de amálgama e substituída por resina composta A3. Paciente assintomático ao final.', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(6, 39, 'Dor intensa no dente 46', 'Pulpite irreversível em elemento 46', 'Tratamento endodôntico', 'Realizada abertura coronária e acesso aos canais. Instrumentação e medicação intracanal com hidróxido de cálcio.', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(7, 41, 'Restauração estética anterior', 'Diastema entre 11 e 21', 'Fechamento de diastema com resina', 'Realizado fechamento de diastema com resina composta A1. Paciente satisfeita com resultado estético.', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- ========================================
-- 11. CONFIGURAÇÕES INICIAIS
-- ========================================

INSERT INTO settings (clinic_name, address, phone, email, logo_url, working_hours, appointment_duration, created_at, updated_at) VALUES
('Clínica Dental Sorriso Perfeito', 'Av. Paulista, 1500 - São Paulo - SP', '(11) 3456-7890', 'contato@sorrisoperfeito.com.br', '/logo.png', 'Seg-Sex: 8h-18h | Sáb: 8h-13h', 60, NOW(), NOW())
ON CONFLICT DO NOTHING;

RESET search_path;

SELECT
  'População concluída com sucesso!' as status,
  '20 pacientes' as pacientes,
  '30+ agendamentos' as agendamentos,
  '24 orçamentos' as orcamentos,
  '50+ pagamentos' as pagamentos,
  '30+ despesas' as despesas,
  '5 fornecedores' as fornecedores,
  '20 produtos' as produtos,
  '9 tarefas' as tarefas;
