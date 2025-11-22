# Plano de Testes - Sistema de Permissões RBAC

## Objetivo
Verificar se o menu do frontend exibe apenas os módulos para os quais o usuário tem permissões.

## Usuário de Teste
- **Email:** maria@gmail.com
- **Senha:** senha123
- **Role:** receptionist

## Cenários de Teste

### Cenário 1: Apenas Dashboard (✓ CONFIGURADO)
**Permissões:**
- dashboard: view

**Resultado Esperado:**
- Menu deve mostrar APENAS "Dashboard"
- Total de itens: 1

### Cenário 2: Dashboard + Agenda
**Permissões:**
- dashboard: view
- appointments: view

**Resultado Esperado:**
- Menu deve mostrar: "Dashboard", "Agenda"
- Total de itens: 2

### Cenário 3: Módulos Financeiros
**Permissões:**
- dashboard: view
- budgets: view, create, edit, delete
- payments: view, create, edit, delete

**Resultado Esperado:**
- Menu deve mostrar: "Dashboard", "Financeiro" (com "Orçamentos" e "Pagamentos")
- Total de itens: 2

### Cenário 4: Recepcionista Completo
**Permissões:**
- dashboard: view
- appointments: view, create, edit, delete
- patients: view, create, edit, delete
- budgets: view, create, edit, delete
- payments: view, create, edit, delete
- exams: view
- medical_records: view

**Resultado Esperado:**
- Menu deve mostrar: Dashboard, Agenda, Pacientes, Prontuários, Exames, Financeiro
- Total de itens: 6

### Cenário 5: Sem Permissões
**Permissões:**
- (nenhuma)

**Resultado Esperado:**
- Menu deve estar vazio ou mostrar apenas mensagem
- Total de itens: 0

### Cenário 6: Apenas Estoque
**Permissões:**
- dashboard: view
- products: view
- suppliers: view
- stock_movements: view

**Resultado Esperado:**
- Menu deve mostrar: "Dashboard", "Estoque" (com "Produtos", "Fornecedores", "Movimentações")
- Total de itens: 2

## Comandos Rápidos

### Aplicar Cenário 1 (Dashboard)
```bash
/root/drcrwell/set-user-permissions.sh maria@gmail.com clear
docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -c "
INSERT INTO public.user_permissions (user_id, permission_id, granted_by, created_at, updated_at)
SELECT 7, p.id, 4, NOW(), NOW()
FROM public.permissions p INNER JOIN public.modules m ON m.id = p.module_id
WHERE m.code = 'dashboard' AND p.action = 'view';
"
```

### Aplicar Cenário 4 (Recepcionista Completo)
```bash
/root/drcrwell/set-user-permissions.sh maria@gmail.com reception
```

### Verificar Permissões Atuais
```bash
docker exec $(docker ps -q -f name=drcrwell_postgres) psql -U drcrwell_user -d drcrwell_db -c "
SELECT m.code, p.action
FROM public.user_permissions up
INNER JOIN public.permissions p ON p.id = up.permission_id
INNER JOIN public.modules m ON m.id = p.module_id
WHERE up.user_id = 7 AND up.deleted_at IS NULL
ORDER BY m.code, p.action;
"
```

### Monitorar Logs (Debug JWT)
```bash
docker service logs drcrwell_backend --follow 2>&1 | grep -E "DEBUG:|maria@gmail.com|Loaded.*permissions"
```

## Status Atual

- [x] Lógica do menu testada e funcionando corretamente (test-menu-logic.js)
- [x] Logs de debug adicionados ao backend (auth.go)
- [x] Backend deployed com logs
- [x] Frontend deployed com correção de filtragem
- [x] Cenário 1 configurado (Dashboard only)
- [ ] Aguardando teste do usuário
- [ ] Executar demais cenários
- [ ] Identificar e corrigir problema (se existir)
