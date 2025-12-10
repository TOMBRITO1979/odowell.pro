# Plano de Implementacao LGPD - Odowell

## FASE 1: ANALISE DE VIABILIDADE (CONCLUIDA)

### 1.1 Conclusoes da Analise

| Aspecto | Status | Risco | Observacoes |
|---------|--------|-------|-------------|
| **Frontend** | Compativel | Baixo | Estrutura modular, facil adicionar paginas publicas |
| **Backend** | Compativel | Baixo | Ja tem grupo de rotas publicas, pattern claro |
| **Banco de Dados** | Compativel | Baixo | Ja existe tabela audit_logs, migrations funcionando |
| **i18n** | NAO EXISTE | Alto | Todos os textos hardcoded em portugues |
| **Rotas** | Compativel | Baixo | Sistema de rotas bem definido (publicas vs protegidas) |

### 1.2 Decisao sobre Internacionalizacao (i18n)

**PROBLEMA IDENTIFICADO:** O sistema NAO possui sistema de traducao. Todos os 71 arquivos JSX tem textos hardcoded em portugues.

**OPCOES:**
1. **Implementar i18n completo** - Risco ALTO, requer refatorar 71+ arquivos
2. **Implementar i18n apenas para novas paginas LGPD** - Risco MEDIO, inconsistente
3. **Manter portugues (como resto do sistema)** - Risco ZERO, consistente

**RECOMENDACAO:** Opcao 3 - Manter portugues para as paginas LGPD, igual ao resto do sistema. Implementar i18n seria um projeto separado muito maior.

---

## FASE 2: PLANO DE IMPLEMENTACAO DETALHADO

### ETAPA 2.1: Paginas Legais Publicas (Risco: ZERO)

**Arquivos a CRIAR (novos):**
```
frontend/src/pages/legal/
├── TermsOfService.jsx      # Termos de Uso
├── PrivacyPolicy.jsx       # Politica de Privacidade
├── LGPDRights.jsx          # Direitos do Titular LGPD
└── CookiePolicy.jsx        # Politica de Cookies
```

**Arquivos a MODIFICAR:**
```
frontend/src/App.jsx        # Adicionar 4 rotas publicas (linhas ~30-40)
```

**Backend - NAO NECESSARIO** (paginas estaticas no frontend)

**Verificacoes:**
- [ ] Rotas nao conflitam com existentes
- [ ] Paginas seguem mesmo estilo visual (Login.jsx como referencia)
- [ ] Links de volta para login funcionam
- [ ] Build do frontend compila sem erros

---

### ETAPA 2.2: Interface de Audit Logs (Risco: MUITO BAIXO)

**Arquivos a CRIAR (novos):**
```
frontend/src/pages/admin/AuditLogs.jsx     # Pagina de visualizacao
```

**Arquivos a MODIFICAR:**
```
frontend/src/App.jsx                        # Adicionar 1 rota protegida
frontend/src/services/api.js                # Adicionar auditAPI
frontend/src/components/layouts/DashboardLayout.jsx  # Adicionar menu (admin only)
backend/cmd/api/main.go                     # Adicionar endpoints (2 linhas)
backend/internal/handlers/audit.go          # CRIAR - handler de listagem
```

**Banco de Dados:** NAO REQUER MUDANCA (tabela audit_logs ja existe)

**Verificacoes:**
- [ ] Endpoint retorna dados corretamente
- [ ] Filtros funcionam (usuario, acao, periodo)
- [ ] Paginacao funciona
- [ ] Apenas admin ve o menu
- [ ] Export CSV funciona

---

### ETAPA 2.3: Portal do Paciente - Solicitacoes LGPD (Risco: BAIXO)

**Arquivos a CRIAR (novos):**
```
frontend/src/pages/patients/PatientDataRequest.jsx  # Modal/pagina de solicitacao
backend/internal/models/data_request.go             # Model para solicitacoes
backend/internal/handlers/data_request.go           # Handler CRUD
backend/migrations/004_create_data_requests.sql     # Tabela nova
```

**Arquivos a MODIFICAR:**
```
frontend/src/App.jsx                    # Adicionar rotas
frontend/src/services/api.js            # Adicionar dataRequestAPI
frontend/src/pages/patients/Patients.jsx  # Adicionar botao "Solicitacao LGPD"
backend/cmd/api/main.go                 # Registrar rotas
```

**Banco de Dados:** NOVA TABELA `data_requests` (solicitacoes de acesso/exclusao)

**Verificacoes:**
- [ ] Paciente pode solicitar acesso aos dados
- [ ] Paciente pode solicitar exclusao
- [ ] Admin recebe notificacao
- [ ] Status da solicitacao e rastreavel
- [ ] Historico de solicitacoes funciona

---

### ETAPA 2.4: Exclusao Permanente de Dados (Risco: MEDIO)

**Arquivos a CRIAR (novos):**
```
backend/internal/handlers/data_deletion.go   # Handler de exclusao permanente
```

**Arquivos a MODIFICAR:**
```
backend/cmd/api/main.go                      # Registrar endpoint
frontend/src/pages/patients/Patients.jsx     # Adicionar opcao (admin only)
```

**Banco de Dados:** NAO REQUER MUDANCA (usa DELETE em vez de soft delete)

**LOGICA DE EXCLUSAO (ordem obrigatoria para evitar FK errors):**
```sql
BEGIN TRANSACTION;
-- 1. Deletar dependencias primeiro
DELETE FROM attachments WHERE patient_id = ?;
DELETE FROM patient_consents WHERE patient_id = ?;
DELETE FROM tasks WHERE patient_id = ?;
DELETE FROM waiting_list WHERE patient_id = ?;
DELETE FROM prescriptions WHERE medical_record_id IN (SELECT id FROM medical_records WHERE patient_id = ?);
DELETE FROM medical_records WHERE patient_id = ?;
DELETE FROM budget_items WHERE budget_id IN (SELECT id FROM budgets WHERE patient_id = ?);
DELETE FROM budgets WHERE patient_id = ?;
DELETE FROM payments WHERE patient_id = ?;
DELETE FROM exams WHERE patient_id = ?;
DELETE FROM appointments WHERE patient_id = ?;
-- 2. Deletar paciente
DELETE FROM patients WHERE id = ?;
-- 3. Registrar em audit_log (mantido para compliance)
INSERT INTO public.audit_logs (...) VALUES (...);
COMMIT;
```

**Verificacoes:**
- [ ] Transacao e atomica (tudo ou nada)
- [ ] Confirmacao dupla obrigatoria
- [ ] Apenas admin pode executar
- [ ] Audit log registra a exclusao
- [ ] Dados relacionados sao excluidos
- [ ] Backup foi feito antes do teste

---

### ETAPA 2.5: Politicas de Retencao Automatica (Risco: BAIXO)

**Arquivos a CRIAR (novos):**
```
backend/internal/scheduler/retention.go      # Job de limpeza automatica
backend/internal/handlers/retention.go       # Handler de configuracao
frontend/src/pages/Settings.jsx              # Adicionar aba de retencao (MODIFICAR)
```

**Arquivos a MODIFICAR:**
```
backend/internal/scheduler/scheduler.go      # Registrar novo job
backend/cmd/api/main.go                      # Registrar endpoints
```

**Banco de Dados:** Configuracoes em `settings` (JSONB existente)

**Verificacoes:**
- [ ] Scheduler roda no horario configurado
- [ ] Logs antigos sao limpos
- [ ] Consentimentos expirados sao marcados
- [ ] Configuracoes sao salvas por tenant

---

## FASE 3: CHECAGEM E TESTES

### 3.1 Testes de Build

```bash
# Backend - verificar compilacao
cd backend && go build ./...

# Frontend - verificar build
cd frontend && npm run build
```

### 3.2 Testes Funcionais por Etapa

**Etapa 2.1 - Paginas Legais:**
- [ ] Acessar /terms sem login
- [ ] Acessar /privacy sem login
- [ ] Acessar /lgpd sem login
- [ ] Acessar /cookies sem login
- [ ] Links de navegacao funcionam
- [ ] Estilo consistente com login

**Etapa 2.2 - Audit Logs:**
- [ ] Admin consegue acessar /audit-logs
- [ ] Nao-admin NAO consegue acessar
- [ ] Filtro por usuario funciona
- [ ] Filtro por acao funciona
- [ ] Filtro por periodo funciona
- [ ] Paginacao funciona
- [ ] Export CSV baixa arquivo

**Etapa 2.3 - Solicitacoes LGPD:**
- [ ] Criar solicitacao de acesso
- [ ] Criar solicitacao de exclusao
- [ ] Listar solicitacoes
- [ ] Atualizar status
- [ ] Notificacao para admin

**Etapa 2.4 - Exclusao Permanente:**
- [ ] BACKUP do banco ANTES de testar
- [ ] Confirmacao dupla aparece
- [ ] Apenas admin pode executar
- [ ] Dados relacionados sao excluidos
- [ ] Audit log registra exclusao
- [ ] Outras tabelas nao sao afetadas

**Etapa 2.5 - Retencao:**
- [ ] Configuracao e salva
- [ ] Job roda no horario
- [ ] Logs antigos sao limpos
- [ ] Dados ativos NAO sao afetados

### 3.3 Testes de Regressao (Nada Quebrou)

```bash
# Verificar todas as abas existentes ainda funcionam
- [ ] Login/Logout funciona
- [ ] Dashboard carrega
- [ ] Pacientes - CRUD completo
- [ ] Agenda - CRUD completo
- [ ] Prontuarios - visualizacao
- [ ] Financeiro - pagamentos/orcamentos
- [ ] Configuracoes - todas as abas
- [ ] Usuarios - listagem e edicao
```

### 3.4 Verificacoes de Banco de Dados

```sql
-- Verificar tabelas nao foram alteradas
\dt tenant_1.*

-- Verificar nova tabela existe (se aplicavel)
\d tenant_1.data_requests

-- Verificar integridade referencial
SELECT COUNT(*) FROM patients WHERE deleted_at IS NULL;
SELECT COUNT(*) FROM appointments WHERE deleted_at IS NULL;
```

### 3.5 Verificacoes de CORS e API

```bash
# Testar endpoints publicos
curl -X GET https://api.odowell.pro/api/health
curl -X GET https://api.odowell.pro/api/terms

# Testar CORS (deve retornar headers corretos)
curl -I -X OPTIONS https://api.odowell.pro/api/terms \
  -H "Origin: https://app.odowell.pro"
```

---

## FASE 4: DEPLOY

### 4.1 Pre-Deploy Checklist

- [ ] Backup do banco de dados feito
- [ ] `.env` e `.env.example` NAO contem senhas reais
- [ ] Arquivos com segredos estao no `.gitignore`
- [ ] Build local passou sem erros
- [ ] Testes manuais passaram

### 4.2 Arquivos Sensiveis (NAO COMMITAR)

```
# Ja no .gitignore - verificar:
.env
*.pem
*.key
credentials.json
```

### 4.3 Comandos de Deploy

```bash
# 1. Commit e push (sem segredos)
git add .
git status  # VERIFICAR que nao ha .env ou senhas
git commit -m "feat: Add LGPD compliance features"
git push origin feature/rbac-permissions

# 2. Build das imagens Docker
make build

# 3. Push das imagens
make push

# 4. Deploy no Swarm
make deploy

# 5. Verificar logs
make logs-backend
make logs-frontend
```

### 4.4 Verificacao Pos-Deploy

```bash
# Verificar servicos rodando
docker service ls

# Verificar health
curl https://api.odowell.pro/api/health

# Testar paginas novas
curl https://app.odowell.pro/terms
curl https://app.odowell.pro/privacy
```

---

## FASE 5: VALIDACAO FINAL

### 5.1 Checklist de Completude

| Feature | Implementada | Testada | Deployada |
|---------|--------------|---------|-----------|
| Termos de Uso | [ ] | [ ] | [ ] |
| Politica de Privacidade | [ ] | [ ] | [ ] |
| Pagina LGPD | [ ] | [ ] | [ ] |
| Politica de Cookies | [ ] | [ ] | [ ] |
| Interface Audit Logs | [ ] | [ ] | [ ] |
| Solicitacoes LGPD | [ ] | [ ] | [ ] |
| Exclusao Permanente | [ ] | [ ] | [ ] |
| Politicas de Retencao | [ ] | [ ] | [ ] |

### 5.2 Verificacao de Nao-Regressao

| Modulo Existente | Funciona Apos Deploy |
|------------------|---------------------|
| Login/Logout | [ ] |
| Dashboard | [ ] |
| Pacientes | [ ] |
| Agenda | [ ] |
| Prontuarios | [ ] |
| Receituario | [ ] |
| Exames | [ ] |
| Orcamentos | [ ] |
| Pagamentos | [ ] |
| Produtos | [ ] |
| Fornecedores | [ ] |
| Campanhas | [ ] |
| Relatorios | [ ] |
| Configuracoes | [ ] |
| Usuarios | [ ] |
| Tarefas | [ ] |
| Lista de Espera | [ ] |

---

## RESUMO DOS RISCOS

| Etapa | Risco | Mitigacao |
|-------|-------|-----------|
| 2.1 Paginas Legais | ZERO | Arquivos novos, sem dependencias |
| 2.2 Audit Logs | Muito Baixo | Leitura de dados existentes |
| 2.3 Solicitacoes LGPD | Baixo | Novo modulo isolado |
| 2.4 Exclusao Permanente | MEDIO | Transacao atomica, backup obrigatorio |
| 2.5 Retencao | Baixo | Scheduler separado |

---

## NOTA SOBRE TRADUCAO

**Status atual:** Sistema 100% em portugues, sem i18n.

**Recomendacao:** Manter paginas LGPD em portugues (consistente com sistema).

**Para implementar i18n no futuro:** Seria necessario:
1. Instalar react-i18next
2. Criar arquivos de traducao (pt-BR, en-US, es-ES)
3. Refatorar 71+ arquivos JSX
4. Adicionar seletor de idioma
5. **Estimativa:** Projeto separado de grande escala

---

*Documento criado em: 2024-12-10*
*Aguardando autorizacao para iniciar implementacao*
