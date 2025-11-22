# Funcionalidades de Exportação e Importação de Orçamentos

## ✨ Novas Funcionalidades Implementadas

### 1. Exportar Orçamentos para CSV
**Localização:** Página de Orçamentos → Botão "Exportar CSV"

**Descrição:** Exporta a lista de orçamentos atualmente filtrada para um arquivo CSV.

**Colunas exportadas:**
- ID
- Paciente
- Profissional
- Descrição
- Valor Total
- Status
- Válido Até
- Data Criação
- Observações

**Como usar:**
1. Acesse a página de Orçamentos
2. (Opcional) Aplique filtros (paciente, status, etc.)
3. Clique em "Exportar CSV"
4. O arquivo será baixado automaticamente

**Arquivo gerado:** `orcamentos_YYYYMMDD_HHmmss.csv`

---

### 2. Importar Orçamentos via CSV
**Localização:** Página de Orçamentos → Botão "Importar CSV"

**Descrição:** Permite fazer upload de múltiplos orçamentos de uma vez através de arquivo CSV.

**Formato do CSV:**
O arquivo deve ter as seguintes colunas (SEM cabeçalho):
1. ID do Paciente (número)
2. Descrição do Orçamento (texto)
3. Valor Total (número decimal, ex: 1500.50)
4. Status (pending/approved/rejected/expired)
5. Observações (opcional, texto)

**Exemplo de linha CSV:**
```
1,"Tratamento de canal",1500.50,pending,"Inclui coroa"
5,"Implante dentário",3500.00,approved,"Aprovado pela seguradora"
```

**Como usar:**
1. Acesse a página de Orçamentos
2. Clique em "Importar CSV"
3. No modal que abrir, clique em "Selecionar arquivo CSV"
4. Escolha seu arquivo CSV
5. Aguarde a importação
6. Verifique o resultado (sucessos e erros)

**Validações:**
- Paciente deve existir no sistema
- Valor deve ser número válido
- Status deve ser um dos valores permitidos
- Arquivo deve ter extensão .csv

**Resultado:**
- Mostra quantos orçamentos foram importados com sucesso
- Lista erros encontrados (se houver)
- Atualiza a lista automaticamente

---

### 3. Gerar PDF da Lista de Orçamentos
**Localização:** Página de Orçamentos → Botão "Gerar PDF"

**Descrição:** Gera um relatório PDF com todos os orçamentos atualmente filtrados em formato de tabela.

**Informações no PDF:**
- Cabeçalho com dados da clínica
- Título: "Relatório de Orçamentos"
- Data de geração
- Tabela com:
  - ID
  - Paciente
  - Profissional
  - Descrição
  - Valor
  - Status
  - Data
- Total geral dos orçamentos
- Quantidade total de orçamentos

**Como usar:**
1. Acesse a página de Orçamentos
2. (Opcional) Aplique filtros para gerar relatório específico
3. Clique em "Gerar PDF"
4. O PDF será baixado automaticamente

**Arquivo gerado:** `orcamentos_lista_YYYYMMDD_HHmmss.pdf`

**Formato:** Paisagem (A4) para melhor visualização da tabela

---

## 🔧 Detalhes Técnicos

### Endpoints Backend Criados:
- `GET /api/budgets/export/csv` - Exportar CSV
- `POST /api/budgets/import/csv` - Importar CSV
- `GET /api/budgets/export/pdf` - Gerar PDF da lista

### Permissões Necessárias:
- **Exportar CSV/PDF:** Permissão "view" no módulo "budgets"
- **Importar CSV:** Permissão "create" no módulo "budgets"

### Filtros Suportados:
Todas as funcionalidades respeitam os filtros aplicados:
- Paciente
- Status
- Datas (se implementadas)

### Arquivos Modificados:
**Backend:**
- `/backend/internal/handlers/budget_export.go` (NOVO)
- `/backend/cmd/api/main.go` (rotas adicionadas)

**Frontend:**
- `/frontend/src/pages/financial/Budgets.jsx`
- `/frontend/src/services/api.js`

---

## 📝 Exemplo de CSV Completo

```csv
1,"Limpeza e profilaxia",250.00,pending,"Paciente com gengivite"
1,"Tratamento de canal + coroa",1800.00,approved,"Aprovado - 3x sem juros"
2,"Aparelho ortodôntico",4500.00,pending,"Aguardando aprovação"
3,"Clareamento dental",800.00,approved,"Tratamento de 2 semanas"
5,"Extração de siso",600.00,rejected,"Paciente desistiu"
```

---

## ⚠️ Observações Importantes

1. **Importação:**
   - O profissional (dentist_id) será o usuário logado
   - Datas de criação serão a data atual
   - Itens detalhados do orçamento não são importados via CSV (apenas descrição)

2. **Exportação:**
   - Exporta apenas os campos principais
   - Itens detalhados do orçamento NÃO são exportados
   - Para orçamentos completos, use o PDF individual

3. **PDF da Lista:**
   - Textos muito longos são truncados para caber na tabela
   - Pagination automática se houver muitos orçamentos

---

## 🎯 Casos de Uso

**1. Migração de Dados:**
Use a importação CSV para migrar orçamentos de outro sistema.

**2. Relatórios Gerenciais:**
Use o PDF da lista para relatórios mensais de orçamentos.

**3. Análise em Planilha:**
Use a exportação CSV para análise em Excel/Google Sheets.

**4. Backup:**
Use a exportação CSV como backup periódico dos dados.

---

**Deploy realizado em:** 2025-11-22
**Versão:** Backend sha256:8c272e7257931dec2084a76e15fd381b382e990723db8f2da1ba1c30f619a3ab
**Versão:** Frontend sha256:0ab0460a99d94fccb91bbc5118f6881192452a4186304de57e53ae188162a698
