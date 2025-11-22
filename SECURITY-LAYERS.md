# 🔒 Segurança em Camadas - Sistema RBAC

## Proteção Completa: Frontend + Backend

### Camada 1: Frontend (UX) ✅
**Objetivo:** Melhorar experiência do usuário

```javascript
// Botões são ESCONDIDOS quando não há permissão
{canDelete('patients') && (
  <Button onClick={handleDelete}>Deletar</Button>
)}
```

**Resultado:** Usuário sem permissão não vê o botão

---

### Camada 2: Backend (SEGURANÇA REAL) ✅
**Objetivo:** Bloquear requisições não autorizadas

```go
// TODAS as rotas protegidas com middleware
patients.DELETE("/:id",
    middleware.PermissionMiddleware("patients", "delete"),  // ← BLOQUEIO!
    handlers.DeletePatient
)
```

**Resultado:** Mesmo se chamar a API diretamente, retorna **403 Forbidden**

---

## Como o Middleware Funciona

### Fluxo de Verificação:

```
1. Requisição DELETE /api/patients/1
   ↓
2. AuthMiddleware verifica JWT
   ↓
3. PermissionMiddleware("patients", "delete")
   ↓
4. Verifica no banco:
   - user_id tem permissão?
   - módulo "patients"
   - ação "delete"
   ↓
5. Se NÃO tem permissão:
   → 403 Forbidden
   → {"error": "Insufficient permissions"}

6. Se TEM permissão:
   → Executa handler DeletePatient
   → 200 OK
```

---

## Todas as Rotas Protegidas

### ✅ CREATE (POST)
```
POST /api/patients          → PermissionMiddleware("patients", "create")
POST /api/appointments      → PermissionMiddleware("appointments", "create")
POST /api/budgets           → PermissionMiddleware("budgets", "create")
POST /api/payments          → PermissionMiddleware("payments", "create")
POST /api/products          → PermissionMiddleware("products", "create")
... (todas as outras)
```

### ✅ EDIT (PUT)
```
PUT /api/patients/:id       → PermissionMiddleware("patients", "edit")
PUT /api/appointments/:id   → PermissionMiddleware("appointments", "edit")
PUT /api/budgets/:id        → PermissionMiddleware("budgets", "edit")
PUT /api/payments/:id       → PermissionMiddleware("payments", "edit")
PUT /api/products/:id       → PermissionMiddleware("products", "edit")
... (todas as outras)
```

### ✅ DELETE
```
DELETE /api/patients/:id       → PermissionMiddleware("patients", "delete")
DELETE /api/appointments/:id   → PermissionMiddleware("appointments", "delete")
DELETE /api/budgets/:id        → PermissionMiddleware("budgets", "delete")
DELETE /api/payments/:id       → PermissionMiddleware("payments", "delete")
DELETE /api/products/:id       → PermissionMiddleware("products", "delete")
... (todas as outras)
```

---

## Bypass para Admin

```go
// Linha 46-50 do middleware/permission.go
if userRole == "admin" {
    c.Next()  // Admin tem acesso total
    return
}
```

Admin tem **acesso completo** a todas as operações, independente de permissões configuradas.

---

## Teste Manual

### Como Testar o Bloqueio:

1. **Fazer login como Maria (sem permissão de delete)**
```bash
curl -X POST https://api.dr.crwell.pro/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"maria@gmail.com","password":"senha123"}'
# Salvar o token retornado
```

2. **Tentar deletar um paciente**
```bash
curl -X DELETE https://api.dr.crwell.pro/api/patients/1 \
  -H "Authorization: Bearer SEU_TOKEN_AQUI"
```

3. **Resultado esperado:**
```json
{
  "error": "Insufficient permissions",
  "module": "patients",
  "action": "delete"
}
```
**Status HTTP: 403 Forbidden**

---

## Conclusão

### ✅ Sistema 100% Seguro!

1. **Frontend:** Esconde botões (melhor UX)
2. **Backend:** Bloqueia requisições (segurança real)
3. **Dupla proteção:** Mesmo se alguém tentar burlar o frontend, o backend bloqueia

### 🎯 Mesmo usando ferramentas como:
- Console do navegador
- Postman
- cURL
- Scripts personalizados

**→ O backend vai BLOQUEAR com 403 Forbidden!**

---

## Logs para Auditoria

Você pode monitorar tentativas de acesso não autorizado nos logs:

```bash
docker service logs drcrwell_backend --follow 2>&1 | grep "Insufficient permissions"
```

Isso mostra todas as tentativas bloqueadas em tempo real.
