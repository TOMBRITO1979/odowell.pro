# Checklist de Produção - Odowell

## Status Atual do Sistema
- **188 endpoints de API** no backend
- **25+ páginas** no frontend
- **Erros encontrados:** Permissões de banco de dados em algumas tabelas

---

## 1. CRÍTICO - Corrigir Antes do Lançamento

### 1.1 Erros de Banco de Dados (CORRIGIDO ✅)
- [x] Corrigir permissões de owner nas tabelas:
  - `consent_templates` ✅
  - `patient_consents` ✅
  - `campaigns` ✅
  - `campaign_recipients` ✅
  - `leads` ✅
  - `data_requests` ✅
  - `treatment_protocols` ✅

> **Nota:** Warnings sobre tabelas inexistentes em tenants inativos (11-20) são esperados e não afetam a produção.

### 1.2 APIs Testadas
- [x] WhatsApp API - 18 endpoints testados e funcionando
- [ ] Patients API - Testar CRUD completo
- [ ] Appointments API - Testar CRUD completo
- [ ] Medical Records API - Testar CRUD completo
- [ ] Prescriptions API - Testar CRUD completo
- [ ] Budgets API - Testar CRUD completo
- [ ] Payments API - Testar CRUD completo
- [ ] Users API - Testar CRUD completo
- [ ] Settings API - Testar configurações

---

## 2. Fluxos Críticos de Usuário

### 2.1 Autenticação
- [ ] Login com email/senha
- [ ] Logout (invalidação de token)
- [ ] Recuperação de senha
- [ ] Troca de senha
- [ ] Renovação automática de token (refresh)
- [ ] Login em múltiplos dispositivos

### 2.2 Cadastro de Paciente
- [ ] Criar paciente com dados mínimos
- [ ] Criar paciente com dados completos
- [ ] Editar paciente
- [ ] Buscar paciente por nome/CPF/telefone
- [ ] Desativar paciente
- [ ] Upload de foto do paciente

### 2.3 Agendamento
- [ ] Criar agendamento
- [ ] Visualizar agenda do dia/semana/mês
- [ ] Remarcar agendamento
- [ ] Cancelar agendamento
- [ ] Confirmar agendamento
- [ ] Marcar como concluído
- [ ] Marcar como faltou
- [ ] Conflito de horários (deve bloquear)

### 2.4 Prontuário/Atendimento
- [ ] Abrir atendimento
- [ ] Preencher odontograma
- [ ] Salvar evolução
- [ ] Adicionar procedimentos
- [ ] Assinar digitalmente (se configurado)
- [ ] Visualizar histórico

### 2.5 Financeiro
- [ ] Criar orçamento
- [ ] Aprovar orçamento
- [ ] Registrar pagamento
- [ ] Registrar despesa
- [ ] Visualizar relatórios

### 2.6 Receitas/Prescrições
- [ ] Criar receita
- [ ] Imprimir receita
- [ ] Assinar digitalmente

---

## 3. Testes de Segurança

### 3.1 Autenticação/Autorização
- [ ] Acesso sem token retorna 401
- [ ] Token expirado retorna 401
- [ ] Usuário só acessa dados do próprio tenant
- [ ] Permissões RBAC funcionando (admin/dentist/receptionist)
- [ ] API Key do WhatsApp isolada por tenant

### 3.2 Validação de Entrada
- [ ] SQL Injection bloqueado
- [ ] XSS bloqueado
- [ ] CORS configurado corretamente
- [ ] Rate limiting funcionando
- [ ] Upload de arquivos validado (tipo/tamanho)

### 3.3 Dados Sensíveis
- [ ] Senhas criptografadas (bcrypt)
- [ ] CPF/RG criptografados
- [ ] Certificados digitais criptografados
- [ ] Logs não expõem dados sensíveis

---

## 4. Performance

### 4.1 Backend
- [ ] Tempo de resposta < 500ms para operações normais
- [ ] Queries otimizadas (verificar N+1)
- [ ] Índices criados nas tabelas principais
- [ ] Connection pooling configurado

### 4.2 Frontend
- [ ] Tempo de carregamento inicial < 3s
- [ ] Lazy loading de rotas funcionando
- [ ] Imagens otimizadas
- [ ] Cache do navegador configurado

---

## 5. Mobile/Responsividade

### 5.1 Páginas Principais
- [ ] Dashboard - visualização mobile
- [ ] Pacientes - cards mobile
- [ ] Agendamentos - cards mobile
- [ ] Prontuários - visualização mobile
- [ ] Menu lateral - drawer mobile

### 5.2 Funcionalidades
- [ ] Touch scroll funcionando
- [ ] Botões com tamanho adequado (44px+)
- [ ] Formulários usáveis em mobile
- [ ] Modais responsivos

---

## 6. Integrações Externas

### 6.1 Stripe (Pagamentos/Assinaturas)
- [ ] Checkout de assinatura funciona
- [ ] Webhooks recebidos corretamente
- [ ] Cancelamento de assinatura
- [ ] Upgrade/downgrade de plano
- [ ] Tenant bloqueado quando assinatura expira

### 6.2 Email (SMTP)
- [ ] Recuperação de senha envia email
- [ ] Notificações de agendamento
- [ ] Campanhas de marketing

### 6.3 S3 (Armazenamento)
- [ ] Upload de exames funciona
- [ ] Upload de anexos funciona
- [ ] Download de arquivos funciona
- [ ] Arquivos isolados por tenant

---

## 7. Infraestrutura

### 7.1 Docker/Swarm
- [ ] Todos os serviços rodando (4 réplicas backend)
- [ ] Health checks configurados
- [ ] Restart automático funcionando
- [ ] Logs centralizados

### 7.2 Banco de Dados
- [ ] Backup automático configurado
- [ ] Conexões suficientes no pool
- [ ] Schemas de tenant isolados
- [ ] Migrations aplicadas

### 7.3 SSL/HTTPS
- [ ] Certificados válidos
- [ ] Redirecionamento HTTP -> HTTPS
- [ ] Headers de segurança (HSTS, CSP, etc.)

---

## 8. Monitoramento

### 8.1 Logs
- [ ] Logs estruturados (JSON)
- [ ] Erros 5xx alertam
- [ ] Acesso a logs fácil

### 8.2 Métricas
- [ ] Prometheus/Grafana configurado
- [ ] Alertas de uso de CPU/memória
- [ ] Alertas de erros frequentes

### 8.3 Uptime
- [ ] Monitoramento externo (UptimeRobot, etc.)
- [ ] Página de status

---

## 9. Documentação

### 9.1 Para Usuários
- [ ] Manual de uso básico
- [ ] FAQ
- [ ] Vídeos tutoriais

### 9.2 Para Desenvolvedores
- [ ] CLAUDE.md atualizado
- [ ] API documentada
- [ ] Variáveis de ambiente documentadas

---

## 10. Pré-Lançamento

### 10.1 Dados de Teste
- [ ] Remover usuários de teste
- [ ] Limpar dados de desenvolvimento
- [ ] Verificar tenants de produção

### 10.2 Comunicação
- [ ] Email de boas-vindas configurado
- [ ] Suporte definido (email/WhatsApp)
- [ ] Termos de uso atualizados
- [ ] Política de privacidade atualizada

---

## Comandos Úteis para Testes

```bash
# Verificar serviços
docker service ls

# Logs do backend
docker service logs drcrwell_backend --tail 100 -f

# Logs do frontend
docker service logs drcrwell_frontend --tail 100 -f

# Acessar banco de dados
docker exec -it $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db

# Testar endpoint (exemplo)
curl -s https://api.odowell.pro/health | jq .

# Verificar SSL
curl -I https://app.odowell.pro
```

---

## Histórico de Bugs Corrigidos

| Data | Bug | Correção |
|------|-----|----------|
| 28/12/2024 | WhatsApp API - cancel/reschedule retornando 500 | Adicionado Session() para evitar contaminação GORM |
| 28/12/2024 | WhatsApp API - waiting-list POST retornando 500 | Substituído Create() por raw SQL |
| 28/12/2024 | Service Worker causando cache infinito | Adicionado código para limpar SW automaticamente |
| 29/12/2024 | Delete de agendamento mostrava 2 mensagens (sucesso + erro) | fetchAppointments() estava indefinida - refatorado para useCallback |

