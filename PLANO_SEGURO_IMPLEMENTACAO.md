# 🛡️ PLANO SEGURO DE IMPLEMENTAÇÃO - ODOWELL

**Princípio:** Segurança acima de velocidade. Cada fase é testada completamente antes de prosseguir.

---

## 📋 FASES DE IMPLEMENTAÇÃO

### ✅ FASE 0: PREPARAÇÃO (30 min)
**Status:** 🔄 Em andamento
**Objetivo:** Estabelecer baseline e estrutura de testes

#### Tarefas:
- [x] Backup completo do banco de dados
- [ ] Criar branch: feature/safe-improvements
- [ ] Criar script de teste CRUD completo
- [ ] Executar baseline de testes
- [ ] Documentar estado atual

#### Critérios de Sucesso:
✓ Backup criado com sucesso
✓ Script de teste rodando sem erros
✓ Baseline documentado

---

### 🎯 FASE 1: CAMPOS DE AGENDA (2-3h)
**Status:** ⏳ Aguardando Fase 0
**Objetivo:** Adicionar Room, Chair e Color sem quebrar nada

#### 1.1 - Backend: Adicionar campos ao modelo
**Arquivo:** `/root/drcrwell/backend/internal/models/appointment.go`

```go
// ADICIONAR após linha 27:
Room        string    `json:"room"`              // Nova: sala/consultório
ChairNumber *int      `json:"chair_number"`      // Nova: número da cadeira
Color       string    `json:"color"`             // Nova: cor para calendário
```

**Segurança:**
- ✅ Campos nullable (não quebra registros antigos)
- ✅ Sem alterar campos existentes
- ✅ Backward compatible

#### 1.2 - Testar Migrations
```bash
# Rodar migrations
docker exec -it $(docker ps -q -f name=drcrwell_backend) /app/api

# Verificar campos adicionados
docker exec -it $(docker ps -q -f name=drcrwell_postgres) \
  psql -U drcrwell_user -d drcrwell_db -c "\d tenant_1.appointments"
```

#### 1.3 - Frontend: Adicionar campos ao formulário
**Arquivo:** `/root/drcrwell/frontend/src/pages/appointments/AppointmentForm.jsx`

Adicionar:
- Select de sala
- Input de cadeira
- Color picker para cor do evento

#### 1.4 - Testes CRUD Completos

**Teste de Criação:**
```json
{
  "patient_id": 1,
  "dentist_id": 4,
  "start_time": "2025-11-25T10:00:00Z",
  "end_time": "2025-11-25T11:00:00Z",
  "type": "consultation",
  "room": "Sala 1",
  "chair_number": 2,
  "color": "#4CAF50"
}
```

**Teste de Leitura:**
- GET /api/appointments
- Verificar se room, chair_number, color aparecem

**Teste de Edição:**
- PUT /api/appointments/:id
- Mudar room de "Sala 1" para "Sala 2"
- Verificar no banco

**Teste de Compatibilidade:**
- Criar appointment SEM os campos novos
- Deve funcionar normalmente (campos nullable)

#### 1.5 - Validação no Banco
```sql
-- Verificar estrutura
\d tenant_1.appointments;

-- Verificar dados novos
SELECT id, room, chair_number, color FROM tenant_1.appointments
WHERE room IS NOT NULL;

-- Verificar appointments antigos ainda funcionam
SELECT COUNT(*) FROM tenant_1.appointments WHERE room IS NULL;
```

#### Critérios de Sucesso:
✓ Migrations rodaram sem erro
✓ Novos appointments têm os campos
✓ Appointments antigos ainda funcionam
✓ Frontend mostra campos novos
✓ Edições refletem no banco
✓ NENHUM erro em produção

#### Rollback (se necessário):
```sql
-- Remover campos (não afeta dados antigos)
ALTER TABLE tenant_1.appointments DROP COLUMN room;
ALTER TABLE tenant_1.appointments DROP COLUMN chair_number;
ALTER TABLE tenant_1.appointments DROP COLUMN color;
```

---

### 🎯 FASE 2: BLOQUEIO DE HORÁRIOS (2h)
**Status:** ⏳ Aguardando Fase 1
**Objetivo:** Permitir bloquear horários na agenda

#### 2.1 - Usar campo Type existente
**Segurança:** NÃO adiciona campos, apenas usa o que existe

Valores novos para `Type`:
- "blocked" (novo)
- "maintenance" (novo)
- "meeting" (novo)

#### 2.2 - Frontend: Tipo "Bloqueio"
- Adicionar opção no select de tipo
- Appointments bloqueados aparecem diferente no calendário
- Não permitem paciente_id (opcional)

#### 2.3 - Testes
- Criar appointment tipo "blocked"
- Verificar que não precisa de patient_id
- Validar que aparece na agenda

#### Critérios de Sucesso:
✓ Bloqueios criados com sucesso
✓ Aparecem diferente no calendário
✓ Não quebra appointments normais

---

### 🎯 FASE 3: VISUALIZAÇÃO POR DENTISTA (3h)
**Status:** ⏳ Aguardando Fase 2
**Objetivo:** Filtrar agenda por dentista

#### 3.1 - Frontend: Componente de Filtro
**Arquivo:** Novo componente `DentistFilter.jsx`

- Lista todos dentistas do tenant
- Filtro por dentista (já existe dentist_id!)
- Visualização por tabs ou dropdown

#### 3.2 - Backend: Endpoint já existe!
**Segurança:** NÃO precisa alterar backend

```bash
GET /api/appointments?dentist_id=4
```

#### 3.3 - Testes
- Filtrar appointments por dentist_id
- Verificar que mostra apenas do dentista selecionado
- Trocar filtro e verificar mudança

#### Critérios de Sucesso:
✓ Filtro funciona
✓ Performance OK (index em dentist_id já existe)
✓ Sem alteração no banco

---

### 🎯 FASE 4: DASHBOARD DO DENTISTA (2h)
**Status:** ⏳ Aguardando Fase 3
**Objetivo:** Página inicial com indicadores

#### 4.1 - Backend: Endpoints de Estatísticas
**Arquivo:** Novo `internal/handlers/dentist_stats.go`

```go
// Todos os dados JÁ EXISTEM, apenas agregar!
GET /api/dentists/:id/stats
GET /api/dentists/:id/today-appointments
```

#### 4.2 - Frontend: Dashboard
- Card: Atendimentos hoje
- Card: Próximo paciente
- Card: Total de pacientes
- Lista: Agenda do dia

#### 4.3 - Testes
- Verificar contadores corretos
- Comparar com COUNT no banco
- Validar performance

#### Critérios de Sucesso:
✓ Estatísticas corretas
✓ Dashboard carrega rápido (<2s)
✓ Sem impacto nas outras páginas

---

### 🎯 FASE 5: ODONTOGRAMA VISUAL (4-5h)
**Status:** ⏳ Aguardando Fase 4
**Objetivo:** Interface visual para odontograma

#### 5.1 - Análise: Campo JÁ EXISTE!
**Segurança:** Backend já tem campo JSONB em medical_records

```go
Odontogram *string `gorm:"type:jsonb" json:"odontogram,omitempty"`
```

#### 5.2 - Frontend: Componente Visual
**Arquivo:** Novo `components/Odontogram.jsx`

Estrutura JSON:
```json
{
  "11": {"status": "healthy", "procedures": []},
  "12": {"status": "cavity", "procedures": ["restoration"]},
  "21": {"status": "missing", "procedures": []}
}
```

#### 5.3 - Testes
- Criar prontuário com odontograma
- Verificar JSON no banco
- Editar dente e salvar
- Verificar que JSON foi atualizado

#### Critérios de Sucesso:
✓ Odontograma salva corretamente
✓ JSON válido no banco
✓ Carregamento e edição funcionam
✓ Não quebra prontuários antigos (campo nullable)

---

## 🧪 PROCEDIMENTO DE TESTE PADRÃO

### Para cada fase, executar:

#### 1. Teste de Criação (CREATE)
```bash
curl -X POST https://api.odowell.pro/api/[endpoint] \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

Validar:
- ✓ Status 200/201
- ✓ ID retornado
- ✓ Campos presentes

#### 2. Teste de Leitura (READ)
```bash
curl https://api.odowell.pro/api/[endpoint]/[id] \
  -H "Authorization: Bearer $TOKEN"
```

Validar:
- ✓ Status 200
- ✓ Dados corretos
- ✓ Campos novos presentes

#### 3. Validação no Banco
```sql
-- Conectar no banco
docker exec -it $(docker ps -q -f name=drcrwell_postgres) \
  psql -U drcrwell_user -d drcrwell_db

-- Verificar registro
SELECT * FROM tenant_1.[tabela] WHERE id = [id];
```

Validar:
- ✓ Registro existe
- ✓ Campos corretos
- ✓ Timestamps atualizados

#### 4. Teste de Edição (UPDATE)
```bash
curl -X PUT https://api.odowell.pro/api/[endpoint]/[id] \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ "campo": "novo_valor" }'
```

Validar:
- ✓ Status 200
- ✓ Campos atualizados

#### 5. Validação de Edição no Banco
```sql
SELECT * FROM tenant_1.[tabela] WHERE id = [id];
```

Validar:
- ✓ Valor atualizado
- ✓ updated_at mudou
- ✓ Outros campos intactos

#### 6. Teste de Compatibilidade
```bash
# Criar registro SEM campos novos
curl -X POST ... -d '{ ... }' # Sem campos novos
```

Validar:
- ✓ Funciona normalmente
- ✓ Campos novos ficam NULL/default
- ✓ Não há erros

#### 7. Teste Frontend
- Abrir página no navegador
- Criar novo registro
- F12 > Network > Verificar requisição
- Editar registro
- Verificar que mudanças aparecem

#### 8. Teste de Performance
```sql
-- Verificar que índices existem
\d tenant_1.[tabela]

-- Testar query com EXPLAIN
EXPLAIN ANALYZE SELECT * FROM tenant_1.appointments
WHERE dentist_id = 4 AND start_time >= NOW();
```

Validar:
- ✓ Usa índices corretos
- ✓ Query < 50ms
- ✓ Sem table scans

---

## 🚨 CRITÉRIOS DE BLOQUEIO

**PARAR imediatamente se:**
- ❌ Erro ao rodar migrations
- ❌ Perda de dados existentes
- ❌ Erro 500 em endpoints antigos
- ❌ Frontend não carrega
- ❌ Testes falharem
- ❌ Performance cair >50%

**Ação:** Rollback imediato e investigar

---

## 📊 CHECKLIST DE CONCLUSÃO DE FASE

Antes de marcar fase como completa:

- [ ] Todos os testes CRUD passaram
- [ ] Validação no banco confirmada
- [ ] Frontend funciona sem erros
- [ ] Endpoints antigos ainda funcionam
- [ ] Performance OK
- [ ] Commit feito com mensagem clara
- [ ] Backup pós-fase criado
- [ ] Documentação atualizada

---

## 🔄 PLANO DE ROLLBACK

### Se algo der errado em QUALQUER fase:

#### 1. Rollback de Código
```bash
git checkout main
git branch -D feature/safe-improvements
docker-compose restart
```

#### 2. Rollback de Banco (se migrations rodaram)
```bash
# Restaurar backup
pg_restore -U drcrwell_user -d drcrwell_db backup_pre_upgrade_*.sql

# Ou remover campos manualmente
docker exec -it $(docker ps -q -f name=drcrwell_postgres) \
  psql -U drcrwell_user -d drcrwell_db

ALTER TABLE tenant_1.appointments DROP COLUMN IF EXISTS room;
ALTER TABLE tenant_1.appointments DROP COLUMN IF EXISTS chair_number;
ALTER TABLE tenant_1.appointments DROP COLUMN IF EXISTS color;
```

#### 3. Verificar Rollback
- Testar endpoints antigos
- Verificar frontend carrega
- Confirmar dados preservados

---

## 📈 PROGRESSO

| Fase | Status | Testes | Deploy |
|------|--------|--------|--------|
| 0 - Preparação | 🔄 Em andamento | ⏳ | ⏳ |
| 1 - Campos Agenda | ⏳ Pendente | ⏳ | ⏳ |
| 2 - Bloqueios | ⏳ Pendente | ⏳ | ⏳ |
| 3 - Visualização Dentista | ⏳ Pendente | ⏳ | ⏳ |
| 4 - Dashboard | ⏳ Pendente | ⏳ | ⏳ |
| 5 - Odontograma | ⏳ Pendente | ⏳ | ⏳ |

---

## ✅ APROVAÇÃO PARA PRÓXIMA FASE

**Processo:**
1. Fase concluída ✓
2. Todos os testes passaram ✓
3. Validação no banco OK ✓
4. Usuário aprova para continuar ✓

**Aguardar aprovação explícita antes de prosseguir!**

---

**Última atualização:** 2025-11-23
**Responsável:** Claude Code + Usuário
**Backup atual:** /root/drcrwell/backups/
