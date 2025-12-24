# AUDITORIA DE SEGURANCA - Itens Pendentes de Implementacao

**Data:** 2025-12-24
**Responsavel:** Claude Code
**Status:** EM IMPLEMENTACAO

---

## RESULTADO DA AUDITORIA

### Itens Pendentes (Bloqueadores)

| # | Item | Risco | Arquivo | Linha |
|---|------|-------|---------|-------|
| 1 | TLS nao obrigatorio no PostgreSQL | CRITICO | `database/database.go` | 39 |
| 2 | Fallback silencioso para DB global | ALTO | `middleware/tenant.go` | 42 |
| 3 | consent.go usa GetDBFromContext inseguro | MEDIO | `handlers/consent.go` | 452 |
| 4 | Tenant ativo nao validado no middleware | ALTO | `middleware/tenant.go` | 13-35 |

### Itens Ja Corrigidos

| # | Item | Status |
|---|------|--------|
| A | Chaves API legadas (texto plano) | CORRIGIDO - So aceita hash |
| B | Pool de conexoes configuravel | CORRIGIDO - Variaveis de ambiente |
| C | GetDBFromContextSafe | IMPLEMENTADO - 194 usos |

---

## PLANO DE IMPLEMENTACAO SEGURA

### ITEM 1: Forcar TLS no PostgreSQL

**Problema:** `sslMode = "prefer"` permite conexao sem criptografia.

**Solucao:**
1. Alterar default de `prefer` para `require` em producao
2. Adicionar validacao no startup que falha se SSL nao for seguro
3. Logar warning se usando `disable` em qualquer ambiente

**Arquivo:** `backend/internal/database/database.go`

**Codigo atual (linha 38-40):**
```go
} else {
    sslMode = "prefer"
}
```

**Codigo corrigido:**
```go
} else {
    sslMode = "require" // TLS obrigatorio em producao
}

// Validar seguranca do SSL mode
if sslMode == "disable" {
    log.Println("WARNING: SSL disabled - NOT RECOMMENDED for production!")
}
```

**Teste:** Verificar logs de startup mostram conexao com TLS.

---

### ITEM 2: Remover Fallback do GetDBFromContext

**Problema:** Se `db` nao existe no contexto, retorna `database.GetDB()` (schema public).

**Solucao:**
1. Alterar `GetDBFromContext` para retornar erro em vez de fallback
2. Forcar uso de `GetDBFromContextSafe` em todos handlers

**Arquivo:** `backend/internal/middleware/tenant.go`

**Codigo atual (linha 38-45):**
```go
func GetDBFromContext(c *gin.Context) interface{} {
    db, exists := c.Get("db")
    if !exists {
        return database.GetDB() // FALLBACK PERIGOSO
    }
    return db
}
```

**Codigo corrigido:**
```go
// GetDBFromContext - DEPRECATED: Use GetDBFromContextSafe instead
// This function now fails closed instead of falling back to global DB
func GetDBFromContext(c *gin.Context) interface{} {
    db, exists := c.Get("db")
    if !exists {
        log.Println("SECURITY ERROR: Attempted to access DB without tenant context")
        return nil // Fail closed - no fallback
    }
    return db
}
```

**Teste:** Verificar que rotas sem TenantMiddleware retornam erro.

---

### ITEM 3: Corrigir consent.go para Usar GetDBFromContextSafe

**Problema:** Linha 452 usa `GetDBFromContext` inseguro.

**Solucao:** Substituir por `GetDBFromContextSafe`.

**Arquivo:** `backend/internal/handlers/consent.go`

**Codigo atual (linha 450-454):**
```go
tenantID, _ := c.Get("tenant_id")
var tenant models.Tenant
dbPublicRaw := middleware.GetDBFromContext(c)
dbPublic, ok := dbPublicRaw.(*gorm.DB)
if !ok {
```

**Codigo corrigido:**
```go
tenantID, _ := c.Get("tenant_id")
var tenant models.Tenant
dbPublic, ok := middleware.GetDBFromContextSafe(c)
if !ok {
    return // GetDBFromContextSafe already sent error response
}
```

**Teste:** Testar endpoint de consent com e sem autenticacao.

---

### ITEM 4: Validar Tenant Ativo no Middleware

**Problema:** TenantMiddleware nao verifica se tenant esta ativo.

**Solucao:**
1. Buscar tenant no banco antes de setar search_path
2. Verificar `tenant.active == true`
3. Retornar 403 se tenant inativo

**Arquivo:** `backend/internal/middleware/tenant.go`

**Codigo atual (linha 13-35):**
```go
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID, exists := c.Get("tenant_id")
        if !exists {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID not found"})
            c.Abort()
            return
        }
        // Set schema sem validar tenant ativo
        schemaName := fmt.Sprintf("tenant_%d", tenantID)
        ...
    }
}
```

**Codigo corrigido:**
```go
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID, exists := c.Get("tenant_id")
        if !exists {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Tenant ID not found"})
            c.Abort()
            return
        }

        // Validate tenant is active before setting schema
        db := database.GetDB()
        var tenant models.Tenant
        if err := db.First(&tenant, tenantID).Error; err != nil {
            c.JSON(http.StatusForbidden, gin.H{"error": "Tenant not found"})
            c.Abort()
            return
        }
        if !tenant.Active {
            c.JSON(http.StatusForbidden, gin.H{"error": "Tenant account is inactive"})
            c.Abort()
            return
        }

        schemaName := fmt.Sprintf("tenant_%d", tenantID)
        ...
    }
}
```

**Teste:** Desativar um tenant e verificar que acesso e bloqueado.

---

## CHECKLIST DE TESTES POS-IMPLEMENTACAO

Para cada item implementado, executar:

### Banco de Dados
- [ ] Conexao estabelecida com TLS
- [ ] Todos os 15 schemas existem (public + 14 tenants)
- [ ] Todas tabelas tenant possuem colunas corretas
- [ ] Todas tabelas public possuem colunas corretas

### API
- [ ] Health check retorna OK
- [ ] Login funciona corretamente
- [ ] Rotas protegidas requerem autenticacao
- [ ] CORS configurado para app.odowell.pro
- [ ] Rate limiting funcionando

### Seguranca
- [ ] Tenant inativo nao consegue acessar
- [ ] Rotas sem middleware retornam erro (nao fallback)
- [ ] API keys hasheadas funcionam
- [ ] JWT valido corretamente

### Dados de Teste
- [ ] Inserir paciente de teste
- [ ] Inserir agendamento de teste
- [ ] Inserir pagamento de teste
- [ ] Verificar listagem funciona

---

## PROGRESSO

| Item | Status | Data | Testado |
|------|--------|------|---------|
| 1. TLS Obrigatorio | PENDENTE | - | - |
| 2. Remover Fallback | PENDENTE | - | - |
| 3. Corrigir consent.go | PENDENTE | - | - |
| 4. Validar Tenant Ativo | PENDENTE | - | - |

---

**Ultima atualizacao:** 2025-12-24
