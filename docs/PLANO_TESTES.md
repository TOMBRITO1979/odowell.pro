# Plano de Testes - OdoWell WhatsApp Automation

**Data:** 2025-12-22
**Telefone Paciente (teste):** 14077608242
**Telefone Clínica:** 380947105869

---

## Configuração de Teste

### Enviar Mensagem (Paciente → Clínica)
```bash
curl -X POST "https://wahazap1979leoale.odowell.pro/api/sendText" \
  -H "X-Api-Key: 28629dc04610b89c16baadcc4233ab6b" \
  -H "Content-Type: application/json" \
  -d '{
    "session": "14077608242",
    "chatId": "380947105869@c.us",
    "text": "MENSAGEM AQUI"
  }'
```

### Ver Resposta da Clínica
```bash
curl -s "https://wahazap1979leoale.odowell.pro/api/14077608242/chats/380947105869@c.us/messages?limit=3" \
  -H "X-Api-Key: 28629dc04610b89c16baadcc4233ab6b"
```

### Limpeza Após Cada Teste
1. Apagar Redis (memória de conversa)
2. Apagar Lead (se criado)
3. Apagar Paciente (se criado)
4. Apagar Agendamento (se criado)

---

## Testes

### TESTE 1: FAQ - Localização
**Objetivo:** Verificar se responde corretamente sobre localização
**Mensagem:** "Oi, onde fica a clínica?"
**Esperado:** Endereço: Av. das Palmeiras, 1250 - Sala 302...
**Status:** [ ] Pendente

---

### TESTE 2: FAQ - Convênios
**Objetivo:** Verificar se responde sobre convênios aceitos
**Mensagem:** "Quais convênios vocês aceitam?"
**Esperado:** Lista de convênios (Amil, Bradesco, SulAmerica, etc.)
**Status:** [ ] Pendente

---

### TESTE 3: FAQ - Horário de Funcionamento
**Objetivo:** Verificar se responde sobre horários
**Mensagem:** "Qual horário de funcionamento?"
**Esperado:** Seg-Sex 8h-20h, Sábado 8h-14h
**Status:** [ ] Pendente

---

### TESTE 4: Pessoa Nova - Apenas Informação (Salvar Lead)
**Objetivo:** Pessoa nova pede info, não quer agendar → salvar como lead com birth_date
**Fluxo:**
1. "Oi, quanto custa uma consulta?"
2. (Maria pede nome e nascimento)
3. "João Teste, 15/03/1990"
4. (Maria responde sobre consulta)
5. "Ok, obrigado, vou pensar"
**Esperado:** Lead salvo com nome, telefone, birth_date e contact_reason
**Status:** [ ] Pendente

---

### TESTE 5: Pessoa Nova - Quer Agendar (Fluxo Completo)
**Objetivo:** Pessoa nova agenda consulta completa
**Fluxo:**
1. "Oi, quero agendar uma consulta"
2. (Maria pede nome e nascimento)
3. "Maria Silva, 20/05/1985"
4. (Maria pergunta data)
5. "Amanhã de manhã"
6. (Maria mostra horários disponíveis)
7. "10h"
8. (Maria pede endereço)
9. "Rua das Flores, 100, Centro, São Paulo, SP"
10. (Maria confirma agendamento com ID)
**Esperado:** Paciente cadastrado + Consulta agendada com ID
**Status:** [ ] Pendente

---

### TESTE 6: Lead Existente - Quer Agendar
**Objetivo:** Lead volta e decide agendar → converter para paciente
**Pré-requisito:** Criar lead primeiro
**Fluxo:**
1. "Oi, agora quero agendar"
2. (Maria reconhece pelo nome)
3. (Maria pergunta data)
4. "Dia 26 às 14h"
5. (Maria pede endereço para converter)
6. "Av. Brasil, 500, Centro, Rio, RJ"
7. (Maria confirma agendamento)
**Esperado:** Lead convertido para paciente + Consulta agendada
**Status:** [ ] Pendente

---

### TESTE 7: Paciente Existente - Listar Consultas
**Objetivo:** Paciente pergunta suas consultas
**Pré-requisito:** Paciente com consulta agendada
**Mensagem:** "Quais são minhas consultas?"
**Esperado:** Lista de consultas com IDs
**Status:** [ ] Pendente

---

### TESTE 8: Paciente Existente - Cancelar Consulta
**Objetivo:** Paciente cancela uma consulta
**Pré-requisito:** Paciente com consulta agendada
**Fluxo:**
1. "Quero cancelar minha consulta"
2. (Maria lista consultas)
3. "A consulta #ID"
4. (Maria confirma cancelamento)
**Esperado:** Consulta cancelada
**Status:** [ ] Pendente

---

### TESTE 9: Paciente Existente - Remarcar Consulta
**Objetivo:** Paciente remarca consulta
**Pré-requisito:** Paciente com consulta agendada
**Fluxo:**
1. "Preciso remarcar minha consulta"
2. (Maria lista consultas)
3. "A #ID para dia 27 às 15h"
4. (Maria confirma remarcação)
**Esperado:** Consulta remarcada para nova data/hora
**Status:** [ ] Pendente

---

### TESTE 10: Tratamento de Canal (Dentista Correto)
**Objetivo:** Verificar se menciona Dra. Fernanda Costa para canal
**Mensagem:** "Preciso fazer um canal, quem faz?"
**Esperado:** Dra. Fernanda Costa - Endodontia
**Status:** [ ] Pendente

---

## Resumo de Resultados (2025-12-22 - ATUALIZADO)

| Teste | Descrição | Status | Observações |
|-------|-----------|--------|-------------|
| 1 | FAQ Localização | ✅ OK | Endereço correto |
| 2 | FAQ Convênios | ✅ OK | Lista completa |
| 3 | FAQ Horários | ✅ OK | Horários corretos |
| 4 | Lead (só info) | ⚠️ PARCIAL | Conversa OK, IA não chama salvar_lead consistentemente |
| 5 | Agendamento Completo | ✅ OK | Fluxo completo funcionando (após correção) |
| 6 | Lead → Paciente | ⏸️ | Depende de 4 funcionar |
| 7 | Listar Consultas | ⏸️ | Não testado |
| 8 | Cancelar Consulta | ⏸️ | Não testado |
| 9 | Remarcar Consulta | ⏸️ | Não testado |
| 10 | Dentista Canal | ✅ OK | Dra. Fernanda Costa - Endodontia |

## Problemas Identificados

### 1. Router não dispara com mensagens WAHA internas
**Problema:** Quando enviamos mensagem de uma sessão WAHA para outra na mesma instância, o webhook pode não disparar corretamente.
**Solução:** Testar workflows diretamente via webhook ou usar telefone real externo.

### 2. IA não chama salvar_lead automaticamente
**Problema:** Quando pessoa não quer agendar, a IA não está chamando a ferramenta salvar_lead.
**Solução Implementada:** Auto-criação de leads no backend (ver item 5).

### 3. Erro de schema nas ferramentas de agendamento - CORRIGIDO
**Problema:** `Received tool input did not match expected schema` - A IA não enviava dentist_id.
**Solução:** Simplificada ferramenta horarios_disponiveis para aceitar apenas date (obrigatório).
**Status:** ✅ Corrigido - Agendamento funcionando.

### 4. IA não salva lead automaticamente - CORRIGIDO
**Problema:** Mesmo com instruções explícitas, a IA não chama salvar_lead consistentemente.
**Causa:** Limitação do comportamento do LLM - não segue instruções de chamada de ferramenta obrigatória.
**Solução Implementada:** Auto-criação de leads no backend (ver item 5).

### 5. NOVA SOLUÇÃO: Auto-criação de Leads no Backend
**Implementação:**
- Modificado endpoint `/api/whatsapp/leads/check/:phone` para aceitar parâmetro `auto_create=true`
- Quando `auto_create=true` e telefone não existe, cria lead automaticamente com status "new"
- Novo endpoint `PUT /api/whatsapp/leads/:id` para atualizar dados do lead (nome, nascimento)
- Leads são criados com nome "Contato WhatsApp" e atualizados quando IA coleta os dados

**Workflow n8n atualizado:**
- `verificar_paciente` agora usa `?auto_create=true` na URL
- Nova ferramenta `atualizar_lead` para atualizar nome e nascimento
- Removida ferramenta `salvar_lead_inicial` (não mais necessária)

**Status:** ✅ Backend deployado - Pendente: importar workflow atualizado no n8n

**Para atualizar o workflow n8n:**
1. Acesse n8n.odowell.pro
2. Abra o workflow "OdoWell - 2. Agendamento"
3. No nó `verificar_paciente`, altere a URL para incluir `?auto_create=true`:
   ```
   https://api.odowell.pro/api/whatsapp/leads/check/{{ $fromAI("phone") }}?auto_create=true
   ```
4. Adicione novo nó HTTP Request Tool chamado `atualizar_lead`:
   - Method: PUT
   - URL: `https://api.odowell.pro/api/whatsapp/leads/{{ $fromAI("lead_id") }}`
   - Headers: X-API-Key: 7a4400cff68cbf66c688e18b81f73821766a314105beb4fcb3aa5008e0a233f5
   - Body JSON: `{"name": "{{ $fromAI("name") }}", "birth_date": "{{ $fromAI("birth_date") }}"}`
5. Conecte o novo nó ao Agente
6. Remova o nó `salvar_lead_inicial` (opcional, não afeta funcionamento)
7. Salve e ative o workflow

**Arquivo JSON do workflow atualizado:** `docs/workflows/2_agendamento_autocreate.json`

---

## Comandos de Limpeza

### Limpar Redis
```bash
# Via n8n ou diretamente no Redis
redis-cli DEL "odowell_agend_14077608242"
redis-cli DEL "14077608242_block"
```

### Apagar Lead
```bash
curl -X DELETE "https://api.odowell.pro/api/whatsapp/leads/{lead_id}" \
  -H "X-API-Key: 0ab88c06cf5911a0d20ea6e806a4bcb433f9bd3a804138406e0273961d3906a9"
```

### Apagar Paciente
```bash
curl -X DELETE "https://api.odowell.pro/api/whatsapp/patients/{patient_id}" \
  -H "X-API-Key: 0ab88c06cf5911a0d20ea6e806a4bcb433f9bd3a804138406e0273961d3906a9"
```

### Apagar Agendamento
```bash
curl -X DELETE "https://api.odowell.pro/api/whatsapp/appointments/{appointment_id}" \
  -H "X-API-Key: 0ab88c06cf5911a0d20ea6e806a4bcb433f9bd3a804138406e0273961d3906a9"
```
