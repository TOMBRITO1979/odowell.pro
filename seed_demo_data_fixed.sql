-- Script de População CORRIGIDO com Dados Fictícios Profissionais
-- OdoWell - Dados para demonstração comercial
-- Versão 2 - Estrutura correta conforme models

SET search_path TO tenant_1;

-- Primeiro, verificar se há dentistas (user_id) na tabela public.users
-- Vamos assumir user_id 1 (admin) como dentista principal

-- ========================================
-- 1. AGENDAMENTOS (Próximos 7 dias) - CORRIGIDO
-- ========================================

-- Hoje
INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, type, procedure, status, notes, created_at, updated_at) VALUES
(1, 1, NOW()::date + INTERVAL '9 hours', NOW()::date + INTERVAL '10 hours', 'consultation', 'Avaliação inicial', 'scheduled', 'Primeira consulta de avaliação', NOW(), NOW()),
(5, 1, NOW()::date + INTERVAL '10 hours', NOW()::date + INTERVAL '11 hours', 'treatment', 'Limpeza e profilaxia', 'scheduled', 'Limpeza semestral', NOW(), NOW()),
(8, 1, NOW()::date + INTERVAL '14 hours', NOW()::date + INTERVAL '15 hours', 'treatment', 'Tratamento endodôntico', 'scheduled', 'Segunda sessão de endodontia', NOW(), NOW()),
(12, 1, NOW()::date + INTERVAL '15 hours', NOW()::date + INTERVAL '16 hours', 'treatment', 'Restauração em resina', 'scheduled', 'Restauração molar inferior', NOW(), NOW()),

-- Amanhã
INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, type, procedure, status, notes, created_at, updated_at) VALUES
(2, 1, NOW()::date + INTERVAL '1 day 9 hours', NOW()::date + INTERVAL '1 day 10 hours', 'return', 'Retorno pós-tratamento', 'scheduled', 'Retorno para avaliação', NOW(), NOW()),
(4, 1, NOW()::date + INTERVAL '1 day 10 hours', NOW()::date + INTERVAL '1 day 11 hours', 'treatment', 'Clareamento dental', 'scheduled', 'Primeira sessão de clareamento', NOW(), NOW()),
(7, 1, NOW()::date + INTERVAL '1 day 11 hours', NOW()::date + INTERVAL '1 day 12 hours', 'treatment', 'Extração dentária', 'scheduled', 'Extração de siso superior', NOW(), NOW()),
(10, 1, NOW()::date + INTERVAL '1 day 14 hours', NOW()::date + INTERVAL '1 day 15 hours', 'treatment', 'Limpeza e flúor', 'scheduled', 'Profilaxia e aplicação de flúor', NOW(), NOW()),
(15, 1, NOW()::date + INTERVAL '1 day 16 hours', NOW()::date + INTERVAL '1 day 17 hours', 'consultation', 'Avaliação protética', 'scheduled', 'Avaliação para prótese', NOW(), NOW()),

-- Dia +2
INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, type, procedure, status, notes, created_at, updated_at) VALUES
(3, 1, NOW()::date + INTERVAL '2 days 9 hours', NOW()::date + INTERVAL '2 days 10 hours', 'treatment', 'Restauração', 'scheduled', 'Restauração em resina', NOW(), NOW()),
(6, 1, NOW()::date + INTERVAL '2 days 10 hours', NOW()::date + INTERVAL '2 days 11 hours', 'treatment', 'Endodontia', 'scheduled', 'Abertura e preparo do canal', NOW(), NOW()),
(9, 1, NOW()::date + INTERVAL '2 days 14 hours', NOW()::date + INTERVAL '2 days 15 hours', 'emergency', 'Atendimento de urgência', 'scheduled', 'Consulta de emergência - dor', NOW(), NOW()),
(13, 1, NOW()::date + INTERVAL '2 days 15 hours', NOW()::date + INTERVAL '2 days 16 hours', 'treatment', 'Limpeza', 'scheduled', 'Limpeza anual', NOW(), NOW()),

-- Dia +3
INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, type, procedure, status, notes, created_at, updated_at) VALUES
(11, 1, NOW()::date + INTERVAL '3 days 9 hours', NOW()::date + INTERVAL '3 days 10 hours', 'treatment', 'Restaurações múltiplas', 'scheduled', 'Múltiplas restaurações', NOW(), NOW()),
(14, 1, NOW()::date + INTERVAL '3 days 10 hours', NOW()::date + INTERVAL '3 days 11 hours', 'consultation', 'Avaliação ortodôntica', 'scheduled', 'Avaliação ortodôntica', NOW(), NOW()),
(16, 1, NOW()::date + INTERVAL '3 days 14 hours', NOW()::date + INTERVAL '3 days 15 hours', 'treatment', 'Endodontia - finalização', 'scheduled', 'Finalização e obturação', NOW(), NOW()),
(18, 1, NOW()::date + INTERVAL '3 days 16 hours', NOW()::date + INTERVAL '3 days 17 hours', 'treatment', 'Limpeza com ultrassom', 'scheduled', 'Limpeza com ultrassom', NOW(), NOW()),

-- Dia +4
INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, type, procedure, status, notes, created_at, updated_at) VALUES
(17, 1, NOW()::date + INTERVAL '4 days 9 hours', NOW()::date + INTERVAL '4 days 10 hours', 'consultation', 'Primeira consulta', 'scheduled', 'Primeira consulta', NOW(), NOW()),
(19, 1, NOW()::date + INTERVAL '4 days 10 hours', NOW()::date + INTERVAL '4 days 11 hours', 'treatment', 'Troca de restauração', 'scheduled', 'Troca de restauração antiga', NOW(), NOW()),
(20, 1, NOW()::date + INTERVAL '4 days 14 hours', NOW()::date + INTERVAL '4 days 15 hours', 'treatment', 'Clareamento - sessão 2', 'scheduled', 'Segunda sessão de clareamento', NOW(), NOW()),
(1, 1, NOW()::date + INTERVAL '4 days 15 hours', NOW()::date + INTERVAL '4 days 16 hours', 'return', 'Retorno', 'scheduled', 'Retorno pós-tratamento', NOW(), NOW()),

-- Dia +5
INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, type, procedure, status, notes, created_at, updated_at) VALUES
(2, 1, NOW()::date + INTERVAL '5 days 9 hours', NOW()::date + INTERVAL '5 days 10 hours', 'treatment', 'Profilaxia', 'scheduled', 'Profilaxia semestral', NOW(), NOW()),
(3, 1, NOW()::date + INTERVAL '5 days 10 hours', NOW()::date + INTERVAL '5 days 11 hours', 'treatment', 'Endodontia - início', 'scheduled', 'Início de endodontia', NOW(), NOW()),
(4, 1, NOW()::date + INTERVAL '5 days 14 hours', NOW()::date + INTERVAL '5 days 15 hours', 'treatment', 'Restauração', 'scheduled', 'Restauração pré-molar', NOW(), NOW()),

-- Dia +6
INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, type, procedure, status, notes, created_at, updated_at) VALUES
(5, 1, NOW()::date + INTERVAL '6 days 9 hours', NOW()::date + INTERVAL '6 days 10 hours', 'consultation', 'Avaliação geral', 'scheduled', 'Avaliação geral', NOW(), NOW()),
(6, 1, NOW()::date + INTERVAL '6 days 10 hours', NOW()::date + INTERVAL '6 days 11 hours', 'treatment', 'Limpeza com jateamento', 'scheduled', 'Limpeza com jateamento', NOW(), NOW()),
(7, 1, NOW()::date + INTERVAL '6 days 14 hours', NOW()::date + INTERVAL '6 days 15 hours', 'treatment', 'Restauração estética', 'scheduled', 'Restauração estética anterior', NOW(), NOW());

-- Agendamentos concluídos (semana passada)
INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, type, procedure, status, notes, created_at, updated_at) VALUES
(1, 1, NOW()::date - INTERVAL '7 days 9 hours', NOW()::date - INTERVAL '7 days 10 hours', 'consultation', 'Avaliação', 'completed', 'Consulta inicial realizada', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
(2, 1, NOW()::date - INTERVAL '6 days 10 hours', NOW()::date - INTERVAL '6 days 11 hours', 'treatment', 'Limpeza', 'completed', 'Limpeza concluída', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
(3, 1, NOW()::date - INTERVAL '5 days 14 hours', NOW()::date - INTERVAL '5 days 15 hours', 'treatment', 'Restauração', 'completed', 'Restauração finalizada', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(4, 1, NOW()::date - INTERVAL '4 days 9 hours', NOW()::date - INTERVAL '4 days 10 hours', 'consultation', 'Avaliação', 'completed', 'Avaliação concluída', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(5, 1, NOW()::date - INTERVAL '3 days 10 hours', NOW()::date - INTERVAL '3 days 11 hours', 'treatment', 'Limpeza', 'completed', 'Profilaxia realizada', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(6, 1, NOW()::date - INTERVAL '2 days 14 hours', NOW()::date - INTERVAL '2 days 15 hours', 'treatment', 'Endodontia', 'completed', 'Primeira sessão concluída', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(7, 1, NOW()::date - INTERVAL '1 day 9 hours', NOW()::date - INTERVAL '1 day 10 hours', 'treatment', 'Restauração', 'completed', 'Restauração em composite', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- Algumas faltas e cancelamentos
INSERT INTO appointments (patient_id, dentist_id, start_time, end_time, type, procedure, status, notes, created_at, updated_at) VALUES
(8, 1, NOW()::date - INTERVAL '8 days 10 hours', NOW()::date - INTERVAL '8 days 11 hours', 'consultation', 'Avaliação', 'no_show', 'Paciente não compareceu', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
(9, 1, NOW()::date - INTERVAL '10 days 14 hours', NOW()::date - INTERVAL '10 days 15 hours', 'treatment', 'Limpeza', 'cancelled', 'Cancelado pelo paciente', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
(10, 1, NOW()::date - INTERVAL '12 days 9 hours', NOW()::date - INTERVAL '12 days 10 hours', 'treatment', 'Restauração', 'no_show', 'Não compareceu', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days');

-- ========================================
-- 2. ORÇAMENTOS - CORRIGIDO
-- ========================================

-- Orçamentos Aprovados
INSERT INTO budgets (patient_id, dentist_id, description, total_value, status, valid_until, notes, created_at, updated_at) VALUES
(1, 1, 'Limpeza + Restauração (2 dentes)', 720.00, 'approved', NOW()::date + INTERVAL '30 days', 'Aprovado com 10% de desconto', NOW() - INTERVAL '15 days', NOW() - INTERVAL '14 days'),
(2, 1, 'Tratamento de Canal + Coroa', 2500.00, 'approved', NOW()::date + INTERVAL '30 days', 'Pagamento parcelado em 5x', NOW() - INTERVAL '20 days', NOW() - INTERVAL '19 days'),
(3, 1, 'Clareamento Dental (2 sessões)', 1080.00, 'approved', NOW()::date + INTERVAL '30 days', 'Aprovado - paciente convênio', NOW() - INTERVAL '10 days', NOW() - INTERVAL '9 days'),
(5, 1, 'Restaurações Estéticas (3 dentes)', 1350.00, 'approved', NOW()::date + INTERVAL '30 days', 'Parcelamento 3x sem juros', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days'),
(6, 1, 'Extração Siso + Limpeza', 900.00, 'approved', NOW()::date + INTERVAL '30 days', 'Desconto à vista', NOW() - INTERVAL '12 days', NOW() - INTERVAL '11 days'),
(10, 1, 'Implante Dentário + Coroa', 4500.00, 'approved', NOW()::date + INTERVAL '60 days', 'Parcelamento em 10x', NOW() - INTERVAL '25 days', NOW() - INTERVAL '24 days'),
(13, 1, 'Limpeza + Aplicação Flúor', 380.00, 'approved', NOW()::date + INTERVAL '30 days', 'Pagamento à vista', NOW() - INTERVAL '8 days', NOW() - INTERVAL '7 days'),
(15, 1, 'Prótese Parcial Removível', 3000.00, 'approved', NOW()::date + INTERVAL '45 days', 'Desconto 6.25%', NOW() - INTERVAL '18 days', NOW() - INTERVAL '17 days'),

-- Orçamentos Pendentes
(4, 1, 'Aparelho Ortodôntico', 5500.00, 'pending', NOW()::date + INTERVAL '30 days', 'Aguardando aprovação do paciente', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
(7, 1, 'Facetas de Porcelana (4 dentes)', 6800.00, 'pending', NOW()::date + INTERVAL '30 days', 'Orçamento enviado por email', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(9, 1, 'Limpeza Profunda + Tratamento Gengival', 1450.00, 'pending', NOW()::date + INTERVAL '30 days', 'Paciente solicitou prazo para decidir', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
(14, 1, 'Coroa de Porcelana', 1800.00, 'pending', NOW()::date + INTERVAL '30 days', 'Aguardando resposta', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(17, 1, 'Restaurações + Limpeza', 920.00, 'pending', NOW()::date + INTERVAL '30 days', 'Primeiro orçamento do paciente', NOW(), NOW()),
(19, 1, 'Clareamento + Limpeza', 1100.00, 'pending', NOW()::date + INTERVAL '30 days', 'Orçamento solicitado hoje', NOW(), NOW()),

-- Orçamentos Rejeitados
(8, 1, 'Implante + Enxerto Ósseo', 7500.00, 'rejected', NOW()::date - INTERVAL '5 days', 'Paciente achou muito caro', NOW() - INTERVAL '30 days', NOW() - INTERVAL '25 days'),
(11, 1, 'Ortodontia Estética', 6200.00, 'rejected', NOW()::date - INTERVAL '10 days', 'Optou por outra clínica', NOW() - INTERVAL '20 days', NOW() - INTERVAL '10 days'),
(16, 1, 'Prótese Total Superior', 2800.00, 'rejected', NOW()::date - INTERVAL '3 days', 'Sem condições no momento', NOW() - INTERVAL '15 days', NOW() - INTERVAL '12 days'),
(18, 1, 'Harmonização Facial', 3500.00, 'rejected', NOW()::date - INTERVAL '7 days', 'Decidiu não fazer', NOW() - INTERVAL '14 days', NOW() - INTERVAL '7 days'),

-- Orçamentos Cancelados
(12, 1, 'Lentes de Contato Dental (6 dentes)', 8400.00, 'cancelled', NOW()::date - INTERVAL '15 days', 'Cancelado a pedido do paciente', NOW() - INTERVAL '40 days', NOW() - INTERVAL '25 days'),
(20, 1, 'Aparelho Autoligado', 6800.00, 'cancelled', NOW()::date - INTERVAL '8 days', 'Mudou de cidade', NOW() - INTERVAL '35 days', NOW() - INTERVAL '27 days');

-- ========================================
-- 3. PAGAMENTOS (Receitas) - CORRIGIDO
-- ========================================

-- Pagamentos dos orçamentos aprovados
-- Orçamento 1 - Ana Paula (720.00) - À vista
INSERT INTO payments (budget_id, patient_id, type, category, description, amount, payment_method, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(1, 1, 'income', 'treatment', 'Limpeza + Restauração - Pagamento à vista', 720.00, 'debit_card', 'paid', NOW()::date - INTERVAL '14 days', NOW()::date - INTERVAL '14 days', 'Pagamento realizado', NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days');

-- Orçamento 2 - Bruno (2500.00) - 5x de 500.00
INSERT INTO payments (budget_id, patient_id, type, category, description, amount, payment_method, is_installment, installment_number, total_installments, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(2, 2, 'income', 'treatment', 'Tratamento de Canal + Coroa - Parcela 1/5', 500.00, 'credit_card', true, 1, 5, 'paid', NOW()::date - INTERVAL '19 days', NOW()::date - INTERVAL '19 days', 'Parcela 1 paga', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days'),
(2, 2, 'income', 'treatment', 'Tratamento de Canal + Coroa - Parcela 2/5', 500.00, 'credit_card', true, 2, 5, 'paid', NOW()::date + INTERVAL '11 days', NOW()::date - INTERVAL '10 days', 'Parcela 2 paga antecipadamente', NOW() - INTERVAL '19 days', NOW() - INTERVAL '10 days'),
(2, 2, 'income', 'treatment', 'Tratamento de Canal + Coroa - Parcela 3/5', 500.00, 'credit_card', true, 3, 5, 'pending', NOW()::date + INTERVAL '11 days', NULL, 'A vencer', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days'),
(2, 2, 'income', 'treatment', 'Tratamento de Canal + Coroa - Parcela 4/5', 500.00, 'credit_card', true, 4, 5, 'pending', NOW()::date + INTERVAL '41 days', NULL, 'A vencer', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days'),
(2, 2, 'income', 'treatment', 'Tratamento de Canal + Coroa - Parcela 5/5', 500.00, 'credit_card', true, 5, 5, 'pending', NOW()::date + INTERVAL '71 days', NULL, 'A vencer', NOW() - INTERVAL '19 days', NOW() - INTERVAL '19 days');

-- Orçamento 3 - Carla (1080.00) - À vista
INSERT INTO payments (budget_id, patient_id, type, category, description, amount, payment_method, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(3, 3, 'income', 'treatment', 'Clareamento Dental - Pagamento à vista', 1080.00, 'pix', 'paid', NOW()::date - INTERVAL '9 days', NOW()::date - INTERVAL '9 days', 'Transferência PIX', NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days');

-- Orçamento 4 - Eduarda (1350.00) - 3x de 450.00
INSERT INTO payments (budget_id, patient_id, type, category, description, amount, payment_method, is_installment, installment_number, total_installments, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(4, 5, 'income', 'treatment', 'Restaurações Estéticas - Parcela 1/3', 450.00, 'credit_card', true, 1, 3, 'paid', NOW()::date - INTERVAL '4 days', NOW()::date - INTERVAL '4 days', 'Parcela 1 paga', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(4, 5, 'income', 'treatment', 'Restaurações Estéticas - Parcela 2/3', 450.00, 'credit_card', true, 2, 3, 'pending', NOW()::date + INTERVAL '26 days', NULL, 'A vencer', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
(4, 5, 'income', 'treatment', 'Restaurações Estéticas - Parcela 3/3', 450.00, 'credit_card', true, 3, 3, 'pending', NOW()::date + INTERVAL '56 days', NULL, 'A vencer', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days');

-- Orçamento 5 - Fernando (900.00) - À vista
INSERT INTO payments (budget_id, patient_id, type, category, description, amount, payment_method, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(5, 6, 'income', 'treatment', 'Extração Siso + Limpeza - Pagamento à vista', 900.00, 'cash', 'paid', NOW()::date - INTERVAL '11 days', NOW()::date - INTERVAL '11 days', 'Pagamento em espécie', NOW() - INTERVAL '11 days', NOW() - INTERVAL '11 days');

-- Orçamento 7 - Mariana (380.00) - À vista
INSERT INTO payments (budget_id, patient_id, type, category, description, amount, payment_method, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(7, 13, 'income', 'treatment', 'Limpeza + Aplicação Flúor - Pagamento à vista', 380.00, 'pix', 'paid', NOW()::date - INTERVAL '7 days', NOW()::date - INTERVAL '7 days', 'PIX recebido', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days');

-- ========================================
-- 4. DESPESAS - CORRIGIDO
-- ========================================

-- Criar um paciente fictício para vincular despesas (ID 21)
INSERT INTO patients (name, cpf, phone, email, active, created_at, updated_at) VALUES
('DESPESAS GERAIS', '000.000.000-00', '(00) 00000-0000', 'despesas@sistema.com', false, NOW(), NOW());

-- Despesas Fixas Mensais
INSERT INTO payments (patient_id, type, category, description, amount, payment_method, status, due_date, paid_date, notes, created_at, updated_at) VALUES
(21, 'expense', 'rent', 'Aluguel Clínica - ' || TO_CHAR(NOW(), 'MM/YYYY'), 5500.00, 'transfer', 'paid', NOW()::date - INTERVAL '20 days', NOW()::date - INTERVAL '20 days', 'Aluguel pago', NOW() - INTERVAL '20 days', NOW() - INTERVAL '20 days'),
(21, 'expense', 'utilities', 'Energia Elétrica - ' || TO_CHAR(NOW(), 'MM/YYYY'), 890.50, 'debit_card', 'paid', NOW()::date - INTERVAL '15 days', NOW()::date - INTERVAL '15 days', 'Conta de luz', NOW() - INTERVAL '15 days', NOW() - INTERVAL '15 days'),
(21, 'expense', 'utilities', 'Água - ' || TO_CHAR(NOW(), 'MM/YYYY'), 156.30, 'debit_card', 'paid', NOW()::date - INTERVAL '12 days', NOW()::date - INTERVAL '12 days', 'Conta de água', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
(21, 'expense', 'utilities', 'Internet - ' || TO_CHAR(NOW(), 'MM/YYYY'), 199.90, 'debit_card', 'paid', NOW()::date - INTERVAL '10 days', NOW()::date - INTERVAL '10 days', 'Internet empresarial', NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
(21, 'expense', 'salary', 'Salário Dr. Carlos - ' || TO_CHAR(NOW(), 'MM/YYYY'), 12000.00, 'transfer', 'paid', NOW()::date - INTERVAL '5 days', NOW()::date - INTERVAL '5 days', 'Salário mensal', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(21, 'expense', 'salary', 'Salário Recepcionista - ' || TO_CHAR(NOW(), 'MM/YYYY'), 2800.00, 'transfer', 'paid', NOW()::date - INTERVAL '5 days', NOW()::date - INTERVAL '5 days', 'Salário mensal', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(21, 'expense', 'material', 'Compra Materiais - Dental Cremer', 3450.80, 'credit_card', 'paid', NOW()::date - INTERVAL '18 days', NOW()::date - INTERVAL '18 days', 'Reposição estoque geral', NOW() - INTERVAL '18 days', NOW() - INTERVAL '18 days'),
(21, 'expense', 'material', 'Compra Resinas - FGM', 1890.50, 'credit_card', 'paid', NOW()::date - INTERVAL '12 days', NOW()::date - INTERVAL '12 days', 'Resinas compostas', NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),
(21, 'expense', 'material', 'Compra Instrumentais - SS White', 2150.00, 'transfer', 'paid', NOW()::date - INTERVAL '8 days', NOW()::date - INTERVAL '8 days', 'Brocas e instrumentos', NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
(21, 'expense', 'material', 'Compra Materiais Endodontia', 1650.75, 'credit_card', 'pending', NOW()::date + INTERVAL '5 days', NULL, 'Vencimento próximo', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(21, 'expense', 'maintenance', 'Manutenção Equipamentos', 850.00, 'pix', 'paid', NOW()::date - INTERVAL '6 days', NOW()::date - INTERVAL '6 days', 'Manutenção preventiva', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days');

-- ========================================
-- 5. TAREFAS - CORRIGIDO
-- ========================================

INSERT INTO tasks (title, description, priority, status, due_date, created_by, created_at, updated_at) VALUES
('Confirmar agendamentos de amanhã', 'Ligar para todos os pacientes agendados para amanhã e confirmar presença', 'high', 'pending', NOW()::date + INTERVAL '1 day', 1, NOW(), NOW()),
('Realizar pedido de materiais', 'Verificar estoque e fazer pedido de materiais que estão abaixo do mínimo', 'medium', 'in_progress', NOW()::date + INTERVAL '2 days', 1, NOW() - INTERVAL '1 day', NOW()),
('Manutenção autoclave', 'Agendar manutenção preventiva da autoclave', 'high', 'pending', NOW()::date + INTERVAL '3 days', 1, NOW(), NOW()),
('Enviar orçamentos pendentes', 'Fazer follow-up dos orçamentos que estão pendentes há mais de 3 dias', 'medium', 'pending', NOW()::date + INTERVAL '1 day', 1, NOW(), NOW()),
('Atualizar prontuários', 'Digitalizar prontuários físicos pendentes', 'low', 'in_progress', NOW()::date + INTERVAL '7 days', 1, NOW() - INTERVAL '2 days', NOW()),
('Cobrar pagamentos atrasados', 'Entrar em contato com pacientes com parcelas vencidas', 'high', 'pending', NOW()::date, 1, NOW(), NOW()),
('Preparar relatório mensal', 'Compilar dados do mês para apresentação', 'medium', 'pending', NOW()::date + INTERVAL '5 days', 1, NOW(), NOW()),
('Revisar estoque mínimo', 'Ajustar níveis mínimos de estoque baseado em consumo', 'low', 'completed', NOW()::date - INTERVAL '2 days', 1, NOW() - INTERVAL '5 days', NOW() - INTERVAL '2 days');

-- ========================================
-- 6. REGISTROS MÉDICOS - CORRIGIDO
-- ========================================

INSERT INTO medical_records (patient_id, dentist_id, appointment_id, type, diagnosis, treatment_plan, procedure_done, notes, created_at, updated_at) VALUES
(1, 1, 31, 'treatment', 'Cárie profunda em elemento 36', 'Restauração em resina composta', 'Restauração classe II em resina A2 no elemento 36', 'Paciente apresentou sensibilidade ao frio. Restauração concluída com sucesso.', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
(2, 1, 32, 'treatment', 'Gengivite leve', 'Profilaxia e orientação de higiene oral', 'Realizada limpeza com remoção de cálculo supra-gengival', 'Paciente orientado sobre técnica de escovação e uso de fio dental.', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
(3, 1, 33, 'treatment', 'Fratura de restauração em elemento 15', 'Remoção de restauração antiga e confecção de nova', 'Removida restauração de amálgama e substituída por resina A3', 'Paciente assintomático ao final do procedimento.', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
(6, 1, 36, 'treatment', 'Pulpite irreversível em elemento 46', 'Tratamento endodôntico', 'Abertura coronária, acesso aos canais, instrumentação', 'Medicação intracanal com hidróxido de cálcio. Paciente retorna em 7 dias.', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
(7, 1, 37, 'treatment', 'Diastema entre elementos 11 e 21', 'Fechamento de diastema com resina composta', 'Fechamento de diastema com resina A1', 'Paciente satisfeita com resultado estético. Orientações pós-procedimento.', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

RESET search_path;

SELECT
  'População concluída com sucesso!' as status,
  COUNT(DISTINCT patient_id) || ' pacientes' as pacientes,
  (SELECT COUNT(*) FROM tenant_1.appointments) || ' agendamentos' as agendamentos,
  (SELECT COUNT(*) FROM tenant_1.budgets) || ' orçamentos' as orcamentos,
  (SELECT COUNT(*) FROM tenant_1.payments WHERE type='income') || ' receitas' as receitas,
  (SELECT COUNT(*) FROM tenant_1.payments WHERE type='expense') || ' despesas' as despesas,
  (SELECT COUNT(*) FROM tenant_1.tasks) || ' tarefas' as tarefas,
  (SELECT COUNT(*) FROM tenant_1.medical_records) || ' registros médicos' as registros
FROM tenant_1.patients;
