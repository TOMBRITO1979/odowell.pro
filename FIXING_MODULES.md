# Guia de Correção de Módulos - Dr. Crwell

## Problema Comum
Handlers e frontend implementados, mas rotas não funcionam (erro 404 ou falha ao criar/listar).

## Checklist de Correção

### 1️⃣ Verificar se as rotas estão registradas no `backend/cmd/api/main.go`

**Localização**: `backend/cmd/api/main.go` - seção `tenanted` (linhas ~51-75)

**Como verificar**:
```bash
grep -n "nome_do_modulo" backend/cmd/api/main.go
```

**Se NÃO existir, adicionar**:
```go
// Nome do Módulo CRUD
nome_modulo := tenanted.Group("/nome-rota")
{
    nome_modulo.POST("", handlers.CreateNomeModulo)
    nome_modulo.GET("", handlers.GetNomeModulos)
    nome_modulo.GET("/:id", handlers.GetNomeModulo)
    nome_modulo.PUT("/:id", handlers.UpdateNomeModulo)
    nome_modulo.DELETE("/:id", handlers.DeleteNomeModulo)
    // Adicionar rotas extras se necessário
}
```

---

### 2️⃣ Verificar estrutura da tabela no banco de dados

**Comando para verificar**:
```bash
docker exec -i $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -c "\d tenant_1.nome_da_tabela"
```

**Comparar com o modelo Go** em `backend/internal/models/nome_modelo.go`

**Se faltar colunas, adicionar**:
```sql
ALTER TABLE tenant_1.nome_tabela ADD COLUMN IF NOT EXISTS nome_coluna TIPO;
```

**Colunas comuns que podem estar faltando**:
- `created_at TIMESTAMP`
- `updated_at TIMESTAMP`
- `deleted_at TIMESTAMP`
- Campos específicos do modelo

---

### 3️⃣ Adicionar Preload de relacionamentos nos handlers

**Localização**: `backend/internal/handlers/nome_handler.go`

**Verificar se tem `.Preload()` nas funções**:

```go
// CreateNomeModulo - Adicionar após criar
func CreateNomeModulo(c *gin.Context) {
    // ... código de criação

    // ✅ Adicionar isso APÓS db.Create()
    db.Preload("Patient").Preload("OutroRelacionamento").First(&objeto, objeto.ID)

    c.JSON(http.StatusCreated, gin.H{"objeto": objeto})
}

// GetNomeModulos - Adicionar no query
func GetNomeModulos(c *gin.Context) {
    var objetos []models.NomeModulo

    // ✅ Adicionar .Preload() antes de .Find()
    if err := query.Preload("Patient").Preload("OutroRelacionamento").
        Offset(offset).Limit(pageSize).Find(&objetos).Error; err != nil {
        // ...
    }
}

// GetNomeModulo - Adicionar no First
func GetNomeModulo(c *gin.Context) {
    var objeto models.NomeModulo

    // ✅ Adicionar .Preload() antes de .First()
    if err := db.Preload("Patient").Preload("OutroRelacionamento").
        First(&objeto, id).Error; err != nil {
        // ...
    }
}

// UpdateNomeModulo - Adicionar após salvar
func UpdateNomeModulo(c *gin.Context) {
    // ... código de update

    // ✅ Adicionar isso APÓS db.Save()
    db.Preload("Patient").Preload("OutroRelacionamento").First(&objeto, objeto.ID)

    c.JSON(http.StatusOK, gin.H{"objeto": objeto})
}
```

**Relacionamentos comuns**:
- `Preload("Patient")` - Para módulos que referenciam pacientes
- `Preload("Dentist")` ou `Preload("User")` - Para módulos com dentista/usuário
- `Preload("Budget")` - Para pagamentos
- `Preload("Product")` - Para movimentações de estoque

---

### 4️⃣ Rebuild e Deploy do Backend

```bash
# Build
docker build --no-cache -t tomautomations/drcrwell-backend:latest ./backend

# Push
docker push tomautomations/drcrwell-backend:latest

# Deploy
docker stack deploy -c docker-stack.yml drcrwell

# Verificar status (aguardar ~10 segundos)
docker service ps drcrwell_backend --no-trunc | head -5
```

---

### 5️⃣ Testar a API

```bash
# Verificar logs
docker service logs drcrwell_backend --tail 50

# Testar no navegador ou via curl
# GET - Listar
curl -H "Authorization: Bearer SEU_TOKEN" https://drapi.crwell.pro/api/nome-rota

# POST - Criar
curl -X POST -H "Authorization: Bearer SEU_TOKEN" -H "Content-Type: application/json" \
  -d '{"campo1":"valor1"}' https://drapi.crwell.pro/api/nome-rota
```

---

## Exemplo Completo - Appointments

### Antes da Correção:
- ❌ Rotas não registradas em `main.go`
- ❌ Tabela sem colunas: `confirmed_at`, `reminder_sent`, `is_recurring`, `recurrence_rule`
- ❌ Handlers sem `Preload("Patient")`

### Após Correção:
```go
// 1. Adicionado em main.go
appointments := tenanted.Group("/appointments")
{
    appointments.POST("", handlers.CreateAppointment)
    appointments.GET("", handlers.GetAppointments)
    appointments.GET("/:id", handlers.GetAppointment)
    appointments.PUT("/:id", handlers.UpdateAppointment)
    appointments.DELETE("/:id", handlers.DeleteAppointment)
    appointments.PATCH("/:id/status", handlers.UpdateAppointmentStatus)
}

// 2. Adicionado colunas no banco
ALTER TABLE tenant_1.appointments ADD COLUMN IF NOT EXISTS confirmed_at TIMESTAMP;
ALTER TABLE tenant_1.appointments ADD COLUMN IF NOT EXISTS reminder_sent BOOLEAN DEFAULT false;
ALTER TABLE tenant_1.appointments ADD COLUMN IF NOT EXISTS is_recurring BOOLEAN DEFAULT false;
ALTER TABLE tenant_1.appointments ADD COLUMN IF NOT EXISTS recurrence_rule TEXT;

// 3. Adicionado Preload nos handlers
db.Preload("Patient").First(&appointment, appointment.ID)
query.Preload("Patient").Offset(offset).Limit(pageSize).Find(&appointments)
```

### Resultado:
✅ CRUD completo funcionando
✅ Criação, listagem, edição, exclusão, visualização
✅ Dados de relacionamento carregados corretamente

---

## Módulos Corrigidos
- [x] **Appointments** (Agendamentos) - ✅ Corrigido em 2025-11-19
- [x] **Medical Records** (Prontuários) - ✅ Corrigido em 2025-11-19
- [x] **Budgets** (Orçamentos) - ✅ Corrigido em 2025-11-19
- [x] **Payments** (Pagamentos) - ✅ Corrigido em 2025-11-19
- [x] **Products** (Produtos) - ✅ Corrigido em 2025-11-19
- [x] **Suppliers** (Fornecedores) - ✅ Corrigido em 2025-11-19
- [x] **Stock Movements** (Movimentações de Estoque) - ✅ Corrigido em 2025-11-19
- [x] **Dashboard/Reports** (Relatórios) - ✅ Corrigido em 2025-11-19
- [x] **Campaigns** (Campanhas) - ✅ Corrigido em 2025-11-19 + 4 modelos criados
- [x] **Exams** (Exames) - ✅ Corrigido em 2025-11-19 + AWS S3 integrado
- [x] **Prescriptions** (Receituário) - ✅ Criado em 2025-11-19 - Backend e Frontend completo com cache de dados da clínica e dentista

## Módulos Pendentes de Correção
- [ ] Próximo módulo aqui...

---

## ⚠️ PROBLEMA COMUM: Erro de UPDATE (tabela duplicada)

### Sintoma
```
ERROR: table name "tabela" specified more than once (SQLSTATE 42712)
UPDATE "tabela" SET ... FROM "tabela" WHERE ...
```

### Causa
O GORM gera SQL incorreto quando fazemos:
1. Carregamos um registro: `db.First(&record, id)`
2. Alteramos campos manualmente: `record.Campo = novoValor`
3. Salvamos: `db.Save(&record)`

### ❌ Código que causa o erro:
```go
func UpdateNomeModulo(c *gin.Context) {
    id := c.Param("id")
    db := c.MustGet("db").(*gorm.DB)

    var record models.NomeModulo
    if err := db.First(&record, id).Error; err != nil {
        // ...
    }

    var input models.NomeModulo
    if err := c.ShouldBindJSON(&input); err != nil {
        // ...
    }

    // ❌ Atualização campo por campo
    record.Campo1 = input.Campo1
    record.Campo2 = input.Campo2
    // ...

    // ❌ Save() gera SQL com FROM incorreto
    if err := db.Save(&record).Error; err != nil {
        // ...
    }
}
```

### ✅ Solução: Usar db.Model().Where().Updates()
```go
func UpdateNomeModulo(c *gin.Context) {
    id := c.Param("id")
    db := c.MustGet("db").(*gorm.DB)

    // ✅ Verificar se existe
    var count int64
    if err := db.Model(&models.NomeModulo{}).Where("id = ?", id).Count(&count).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        return
    }
    if count == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
        return
    }

    var input models.NomeModulo
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // ✅ Usar Updates com map para evitar o erro
    updates := map[string]interface{}{
        "campo1": input.Campo1,
        "campo2": input.Campo2,
        // ... todos os campos
    }

    if err := db.Model(&models.NomeModulo{}).Where("id = ?", id).Updates(updates).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update"})
        return
    }

    // ✅ Carregar o registro atualizado com relacionamentos
    var record models.NomeModulo
    db.Preload("Patient").Preload("OutroRelacionamento").First(&record, id)

    c.JSON(http.StatusOK, gin.H{"record": record})
}
```

### Aplicar em todos os handlers Update!

---

## Comandos Úteis

```bash
# Ver tabelas de um schema
docker exec -i $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -c "\dt tenant_1.*"

# Ver estrutura de uma tabela
docker exec -i $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -c "\d tenant_1.nome_tabela"

# Executar SQL direto
docker exec -i $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db << 'EOF'
-- Seu SQL aqui
EOF

# Ver logs do backend
docker service logs drcrwell_backend --tail 100 | grep -i erro

# Listar todos os handlers
find backend/internal/handlers -name "*.go" | xargs basename -s .go
```

---

## Notas Importantes

1. **Sempre fazer backup** antes de alterar o banco de dados
2. **Testar localmente** antes de fazer deploy em produção
3. **Verificar logs** após cada deploy para garantir que não há erros
4. **Documentar** cada correção neste arquivo
5. **Commit** das alterações após confirmar que está funcionando

---

**Última atualização**: 2025-11-19
**Responsável**: Claude Code
