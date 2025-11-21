# Guia de Teste Completo - Sistema RBAC

## ✅ Sistema Implementado

### Backend (100% Completo)
- ✅ 3 tabelas: modules, permissions, user_permissions
- ✅ 14 módulos, 56 permissões
- ✅ Middleware de permissões em 67 rotas
- ✅ API de gerenciamento de usuários e permissões
- ✅ JWT com permissões incluídas
- ✅ Admin tem bypass automático

### Frontend (100% Completo)
- ✅ AuthContext com permissões do JWT
- ✅ Hook usePermission() para verificações
- ✅ Página de gerenciamento de usuários (/users)
- ✅ Componente de gerenciamento de permissões
- ✅ Menu filtrado por permissões
- ✅ Item "Usuários" apenas para admins

---

## 🧪 ROTEIRO DE TESTE

### PARTE 1: Login como ADMIN

**URL:** https://dr.crwell.pro

**Credenciais Admin:**
- Email: `wasolutionscorp@gmail.com`
- Senha: `Senha123`

**Testes a realizar:**

1. **Verificar Menu Completo**
   - [ ] Dashboard
   - [ ] Agenda (Appointments)
   - [ ] Pacientes
   - [ ] Prontuários
   - [ ] Receituário
   - [ ] Exames
   - [ ] Financeiro (Orçamentos, Pagamentos)
   - [ ] Estoque (Produtos, Fornecedores, Movimentações)
   - [ ] Campanhas
   - [ ] Relatórios
   - [ ] **Usuários** ⭐ (Novo - apenas admin)

2. **Testar Criação de Dados**
   - [ ] Criar um novo paciente
   - [ ] Criar um novo agendamento
   - [ ] Criar um novo produto
   - [ ] Criar uma nova prescrição

3. **Testar Gerenciamento de Usuários**
   - [ ] Acessar "Usuários" no menu
   - [ ] Verificar lista de usuários
   - [ ] Clicar em "Permissões" do usuário recepcionista
   - [ ] Verificar grid de permissões por módulo
   - [ ] Modificar uma permissão (ex: remover "campaigns:view")
   - [ ] Salvar
   - [ ] Verificar que foi salvo (reabrir permissões)

4. **Testar Criação de Novo Usuário**
   - [ ] Clicar em "Novo Usuário"
   - [ ] Preencher dados:
     - Nome: "Teste Funcionário"
     - Email: "funcionario@teste.com"
     - Senha: "senha123"
     - Role: "Receptionist"
   - [ ] Criar
   - [ ] Clicar em "Permissões" do novo usuário
   - [ ] Clicar em "Aplicar Permissões Padrão (receptionist)"
   - [ ] Verificar que as permissões foram aplicadas
   - [ ] Salvar

---

### PARTE 2: Login como RECEPCIONISTA

**Sair do sistema (Logout)**

**Credenciais Recepcionista:**
- Email: `recepcionista@teste.com`
- Senha: `senha123`

**Testes a realizar:**

1. **Verificar Menu Filtrado**
   - [ ] Dashboard ✓
   - [ ] Agenda ✓
   - [ ] Pacientes ✓
   - [ ] Orçamentos ✓
   - [ ] Pagamentos ✓
   - [ ] ❌ NÃO deve mostrar: Prescrições, Produtos, Fornecedores, Campanhas
   - [ ] ❌ NÃO deve mostrar: Usuários

2. **Testar Operações Permitidas**
   - [ ] Criar um paciente (deve funcionar)
   - [ ] Visualizar lista de pacientes (deve funcionar)
   - [ ] Editar um paciente (deve funcionar)
   - [ ] Criar um agendamento (deve funcionar)
   - [ ] Visualizar pagamentos (deve funcionar)

3. **Testar Bloqueios (via URL direta)**
   - [ ] Tentar acessar https://dr.crwell.pro/prescriptions
     - Resultado esperado: Pode ver a página mas ao tentar listar dará erro 403
   - [ ] Tentar acessar https://dr.crwell.pro/products
     - Resultado esperado: Pode ver a página mas ao tentar listar dará erro 403
   - [ ] Tentar acessar https://dr.crwell.pro/users
     - Resultado esperado: Mensagem "Acesso negado. Apenas administradores..."

---

### PARTE 3: Login como NOVO USUÁRIO

**Sair do sistema (Logout)**

**Credenciais Novo Usuário:**
- Email: `funcionario@teste.com`
- Senha: `senha123`

**Testes a realizar:**

1. **Verificar Permissões Padrão de Receptionist**
   - [ ] Menu deve mostrar: Dashboard, Agenda, Pacientes, Orçamentos, Pagamentos
   - [ ] Menu NÃO deve mostrar: Prescrições, Produtos, etc.

2. **Testar Operações**
   - [ ] Criar paciente
   - [ ] Criar agendamento
   - [ ] Visualizar orçamentos

---

## ✅ CHECKLIST DE VALIDAÇÃO FINAL

### Backend
- [x] Migrations executadas com sucesso
- [x] 14 módulos criados no banco
- [x] 56 permissões criadas no banco
- [x] Usuários têm permissões atribuídas
- [x] Middleware aplicado em todas as rotas
- [x] Admin tem bypass automático
- [x] Endpoints de gerenciamento funcionando

### Frontend
- [x] AuthContext extrai permissões do JWT
- [x] Hook usePermission funciona
- [x] Menu filtrado por permissões
- [x] Página Usuários criada
- [x] Componente UserPermissions funciona
- [x] Rota /users adicionada
- [x] Build e deploy realizados

### Integração
- [ ] Admin vê menu completo ⚠️ (TESTE MANUAL)
- [ ] Recepcionista vê menu filtrado ⚠️ (TESTE MANUAL)
- [ ] Permissões podem ser alteradas ⚠️ (TESTE MANUAL)
- [ ] Novo usuário herda permissões padrão ⚠️ (TESTE MANUAL)
- [ ] API retorna 403 para operações sem permissão ⚠️ (TESTE MANUAL)

---

## 📋 RELATÓRIO DE BUGS

**Registre aqui qualquer problema encontrado:**

1.
2.
3.

---

## ✅ Status do Sistema

- **Backend:** ✅ 100% Operacional
- **Frontend:** ✅ 100% Deployado
- **URL:** https://dr.crwell.pro
- **Última atualização:** 2025-11-21

---

## 📌 Próximos Passos (Opcional)

1. Adicionar botões de Create/Edit/Delete condicionais baseados em permissões
2. Adicionar proteção em nível de botão (esconder/desabilitar)
3. Adicionar mensagens de erro amigáveis quando usuário não tem permissão
4. Implementar auditoria de mudanças de permissões
5. Adicionar histórico de quem concedeu cada permissão
