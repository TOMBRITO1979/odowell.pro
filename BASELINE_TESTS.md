# 📊 BASELINE - Estado Inicial do Sistema

**Data:** 2025-11-23 16:51:38
**Ambiente:** Produção (api.odowell.pro)

---

## ✅ TESTES EXECUTADOS

### Resumo
- **Total de testes:** 10
- **Passou:** 17 asserções
- **Falhou:** 0
- **Taxa de sucesso:** 100%

---

## 📋 DETALHAMENTO POR MÓDULO

### 1. PACIENTES (Patients)
✅ CREATE - Paciente criado com ID 17
✅ READ - Leitura bem-sucedida
✅ UPDATE - Atualização de nome e email
✅ VERIFY - Persistência confirmada no banco

**Campos testados:**
- name: ✓
- cpf: ✓
- email: ✓
- active: ✓

---

### 2. AGENDAMENTOS (Appointments)
✅ CREATE - Agendamento criado com ID 69
✅ READ - Leitura bem-sucedida
✅ UPDATE - Mudança de status e notes
✅ VERIFY - Persistência confirmada

**Campos testados:**
- patient_id: ✓
- dentist_id: ✓
- start_time: ✓
- end_time: ✓
- type: ✓
- status: ✓
- notes: ✓

---

### 3. ORÇAMENTOS (Budgets)
✅ CREATE - Orçamento criado com ID 32
✅ UPDATE - Status e valor atualizados
✅ VERIFY - Mudanças persistidas

**Campos testados:**
- patient_id: ✓
- dentist_id: ✓
- total_value: ✓
- status: ✓
- description: ✓

---

## 🗄️ BACKUP CRIADO

**Arquivo:** `/root/drcrwell/backups/backup_pre_upgrade_20251123_164713.sql`
**Tamanho:** 98 KB
**Status:** ✓ Backup completo criado com sucesso

---

## 📝 ESTRUTURA ATUAL DO BANCO

### Tabelas Existentes:
- tenant_1.appointments
- tenant_1.attachments
- tenant_1.budgets
- tenant_1.campaigns
- tenant_1.campaign_recipients
- tenant_1.exams
- tenant_1.medical_records
- tenant_1.patients
- tenant_1.payments
- tenant_1.prescriptions
- tenant_1.products
- tenant_1.settings
- tenant_1.stock_movements
- tenant_1.suppliers
- tenant_1.tasks
- public.users
- public.tenants
- public.modules
- public.permissions
- public.user_permissions

### Campos appointments (Baseline):
```sql
\d tenant_1.appointments

Columns:
- id (bigint, PK)
- created_at (timestamp)
- updated_at (timestamp)
- deleted_at (timestamp, nullable)
- patient_id (bigint, FK)
- dentist_id (bigint, FK)
- start_time (timestamp)
- end_time (timestamp)
- type (varchar)
- procedure (varchar)
- status (varchar, default: 'scheduled')
- confirmed (boolean, default: false)
- confirmed_at (timestamp, nullable)
- reminder_sent (boolean, default: false)
- notes (text)
- is_recurring (boolean, default: false)
- recurrence_rule (varchar, nullable)
```

---

## 🔍 VALIDAÇÕES

### Integridade Referencial
✅ Foreign keys funcionando (patient_id, dentist_id)
✅ Cascade delete não afetou dados relacionados
✅ Soft delete (deleted_at) funcionando

### API Response Format
✅ Formato: `{"patient": {...}}`, `{"appointment": {...}}`, `{"budget": {...}}`
✅ Status codes corretos (200, 201, 204)
✅ Headers de autenticação funcionando

### Persistência
✅ CREATE persiste no banco
✅ UPDATE reflete no banco
✅ DELETE remove do banco
✅ Timestamps atualizados corretamente

---

## 🚀 PRÓXIMOS PASSOS

**FASE 1** está pronta para iniciar:
- Adicionar campos: `room`, `chair_number`, `color`
- Testar compatibilidade com registros antigos
- Validar que nada quebra

**Critério de aprovação:**
- Todos os testes desta baseline devem continuar passando
- Novos campos devem ser testados
- Zero erros em produção

---

## 📌 NOTAS IMPORTANTES

1. **Todos os testes são idempotentes** - Podem ser executados múltiplas vezes
2. **Cleanup automático** - Dados de teste são removidos após execução
3. **Formato de resposta** - API retorna objetos wrapeados (ex: `{"patient": {...}}`)
4. **Backward compatibility** - Sistema atual funciona 100%

---

**Status:** ✅ BASELINE ESTABELECIDO
**Pronto para:** FASE 1 - Melhorias de Agenda

---

**Comando para reexecutar baseline:**
```bash
python3 /root/drcrwell/test-crud-complete.py
```

**Comando para restaurar backup:**
```bash
CONTAINER_ID=$(docker ps -q -f name=drcrwell_postgres)
docker cp /root/drcrwell/backups/backup_pre_upgrade_20251123_164713.sql $CONTAINER_ID:/tmp/
docker exec $CONTAINER_ID psql -U drcrwell_user -d drcrwell_db < /tmp/backup_pre_upgrade_20251123_164713.sql
```
