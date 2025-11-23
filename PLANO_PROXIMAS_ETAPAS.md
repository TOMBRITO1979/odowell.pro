# PLANO DAS PRÓXIMAS ETAPAS - Odowell

**Ordenado do MAIS FÁCIL → MAIS DIFÍCIL**

---

## ✅ FASE 1: Relatórios Específicos (FÁCIL - 1 dia)

### 1.1 Taxa de Conversão de Orçamentos
**Dificuldade**: ⭐ Fácil
**Tempo estimado**: 2-3 horas
**Risco**: Baixo

**O que fazer:**
- Criar endpoint `/api/reports/budget-conversion`
- SQL query: contar orçamentos por status (pending, approved, rejected)
- Calcular % de conversão: (approved / total) * 100
- Frontend: Adicionar gráfico na página de Relatórios

**Arquivos a modificar:**
- `backend/internal/handlers/report.go` - adicionar função GetBudgetConversionReport
- `backend/cmd/api/main.go` - registrar rota
- `frontend/src/pages/Reports.jsx` - adicionar card com gráfico

**SQL necessário:**
```sql
SELECT
  status,
  COUNT(*) as count,
  ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
FROM budgets
WHERE created_at >= ? AND created_at <= ?
GROUP BY status
```

---

### 1.2 Controle de Inadimplência
**Dificuldade**: ⭐ Fácil
**Tempo estimado**: 2-3 horas
**Risco**: Baixo

**O que fazer:**
- Criar endpoint `/api/reports/overdue-payments`
- SQL query: pagamentos com `due_date < NOW()` e `status = 'pending'`
- Agrupar por paciente, mostrar valor total em atraso
- Frontend: Tabela com pacientes inadimplentes

**Arquivos a modificar:**
- `backend/internal/handlers/report.go` - adicionar GetOverduePaymentsReport
- `backend/cmd/api/main.go` - registrar rota
- `frontend/src/pages/Reports.jsx` - adicionar tabela de inadimplência

**SQL necessário:**
```sql
SELECT
  p.name as patient_name,
  p.id as patient_id,
  COUNT(*) as overdue_count,
  SUM(pm.amount) as total_overdue
FROM payments pm
JOIN budgets b ON pm.budget_id = b.id
JOIN patients p ON b.patient_id = p.id
WHERE pm.status = 'pending'
  AND pm.due_date < CURRENT_DATE
GROUP BY p.id, p.name
ORDER BY total_overdue DESC
```

---

## ✅ FASE 2: Documentos e Termos de Consentimento (MÉDIO - 2 dias)

**Dificuldade**: ⭐⭐ Médio
**Tempo estimado**: 1-2 dias
**Risco**: Médio

### Backend
1. **Model**: `backend/internal/models/consent_document.go`
```go
type ConsentDocument struct {
    ID          uint
    PatientID   uint
    Type        string // "implant", "orthodontics", "surgery", etc.
    Content     string // Template HTML
    SignedAt    *time.Time
    SignedBy    string // Patient name
    SignatureData string // Base64 signature image
}
```

2. **Handler**: `backend/internal/handlers/consent_document.go`
   - CRUD completo
   - Endpoint para gerar PDF do termo assinado

3. **Templates**: Criar templates padrão de termos
   - Implante
   - Ortodontia
   - Cirurgia
   - Clareamento
   - Extração

### Frontend
1. **Componente**: `frontend/src/pages/consent-documents/`
   - Lista de termos
   - Formulário de criação
   - Modal de assinatura digital (canvas)
   - Visualização e impressão

2. **Biblioteca de assinatura**: `react-signature-canvas`
```bash
npm install react-signature-canvas
```

**Complexidade adicional:**
- Biblioteca de assinatura digital
- Geração de PDF com assinatura
- Templates personalizáveis

---

## ✅ FASE 3: Confirmação Automática de Consultas (DIFÍCIL - 3-5 dias)

**Dificuldade**: ⭐⭐⭐ Difícil
**Tempo estimado**: 3-5 dias
**Risco**: Alto

### Opção A: Integração Chatwoot (RECOMENDADO)
**Vantagens:**
- Sistema já tem campanhas no código
- Chatwoot é open-source
- Suporta WhatsApp Business API

**Passos:**
1. **Configurar Chatwoot**
   - Deploy do Chatwoot (Docker)
   - Configurar WhatsApp Business API
   - Obter API credentials

2. **Backend: Integração**
   - Adicionar campos em `settings` table:
     - `chatwoot_api_url`
     - `chatwoot_api_key`
     - `chatwoot_account_id`

   - Criar `backend/internal/services/chatwoot.go`:
     ```go
     type ChatwootService struct {
         APIUrl    string
         APIKey    string
         AccountID string
     }

     func (c *ChatwootService) SendMessage(phone, message string) error
     func (c *ChatwootService) SendTemplate(phone, templateName string, params map[string]string) error
     ```

3. **Automação de Lembretes**
   - Criar worker/cron job em `backend/internal/workers/appointment_reminders.go`
   - Enviar lembretes 24h antes da consulta
   - Enviar lembretes 2h antes da consulta
   - Permitir confirmação via WhatsApp

4. **Frontend: Configuração**
   - Página de settings para configurar Chatwoot
   - Teste de envio de mensagem
   - Histórico de mensagens enviadas

**Arquivos principais:**
- `backend/internal/services/chatwoot.go` (NOVO)
- `backend/internal/workers/appointment_reminders.go` (NOVO)
- `backend/internal/handlers/campaign.go` (ATUALIZAR)
- `frontend/src/pages/settings/Notifications.jsx` (NOVO)

**Desafios:**
- Configurar WhatsApp Business API (requer aprovação do Facebook)
- Gerenciar fila de mensagens
- Tratar respostas e confirmações
- Logs e auditoria de envios

---

### Opção B: Integração Twilio (ALTERNATIVA)
**Vantagens:**
- Mais simples de configurar
- Suporta SMS e WhatsApp
- Documentação excelente

**Desvantagens:**
- Serviço pago
- Dependência de terceiro

---

## ❌ FASE 4: Emissão de Notas Fiscais (MUITO DIFÍCIL - DEIXAR POR ÚLTIMO)

**Dificuldade**: ⭐⭐⭐⭐⭐ Muito Difícil
**Tempo estimado**: 1-2 semanas
**Risco**: Muito Alto

**Por que é difícil:**
- Integração com SEFAZ (órgão governamental)
- Certificado digital A1 obrigatório
- XML complexo com validação rigorosa
- Diferentes regras por município (NFS-e)
- Ambiente de homologação vs produção
- Armazenamento seguro de XMLs
- Cancelamento e carta de correção

**Deixar para o final!**

---

## 📊 RESUMO DO PLANO

### Ordem de Implementação Sugerida:

1. ✅ **FASE 1** (1 dia - FÁCIL)
   - Taxa de conversão de orçamentos
   - Controle de inadimplência

2. ✅ **FASE 2** (2 dias - MÉDIO)
   - Documentos e termos de consentimento

3. ✅ **FASE 3** (3-5 dias - DIFÍCIL)
   - Confirmação automática com Chatwoot

4. ❌ **FASE 4** (1-2 semanas - MUITO DIFÍCIL)
   - Notas fiscais (DEIXAR POR ÚLTIMO)

---

## 🎯 RECOMENDAÇÃO

**Começar pela FASE 1** (Relatórios):
- Baixo risco
- Rápido de implementar
- Entrega valor imediato
- Não tem dependências externas
- Usa tecnologias já dominadas

**Próximos passos:**
1. Implementar taxa de conversão de orçamentos ⏱️ 2-3h
2. Implementar controle de inadimplência ⏱️ 2-3h
3. Testar e validar os relatórios
4. Deploy e documentação

**Deseja que eu comece pela FASE 1 (Relatórios)?**
