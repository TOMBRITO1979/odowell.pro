# OdoWell - Historico do Projeto
**Ultima atualizacao:** 2025-12-22

## Visao Geral
Sistema de automacao WhatsApp para clinica odontologica OdoWell usando:
- **n8n** para workflows de automacao
- **WAHA** (WhatsApp HTTP API) para integracao WhatsApp
- **Redis** para memoria de conversas e bloqueios
- **Google Gemini** como modelo de IA
- **OdoWell API** para backend (pacientes, dentistas, agendamentos, leads)

---

## URLs e Credenciais

### n8n
- **URL:** https://auto1979leoale.odowell.pro
- **API Key:** eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI4N2U5YWE3OC01MjUyLTQ2MzYtOWM0OC0yYzRmMDJkZTc5ZjYiLCJpc3MiOiJuOG4iLCJhdWQiOiJwdWJsaWMtYXBpIiwiaWF0IjoxNzY2MzUzMjcyLCJleHAiOjE3Njg4NzgwMDB9.ELtLOXGM3IoDjCs5AfAvhaCXmbQMTo0n3BDEg6Ox4Mc
- **Expira:** ~Jan 2026

### WAHA
- **URL:** https://wahazap1979leoale.odowell.pro
- **API Key:** 28629dc04610b89c16baadcc4233ab6b
- **Session:** 380947105869 (telefone da clinica)

### OdoWell API
- **URL:** https://api.odowell.pro
- **API Key (workflows):** 7a4400cff68cbf66c688e18b81f73821766a314105beb4fcb3aa5008e0a233f5
- **API Key (alternativa):** 0ab88c06cf5911a0d20ea6e806a4bcb433f9bd3a804138406e0273961d3906a9

### Redis
- **Credential ID no n8n:** d03Gkt3G9WmZzcvE

### Gemini
- **Credential ID no n8n:** Wdf3JCPdnOETBHPc

---

## Workflows n8n

| ID | Nome | Funcao |
|----|------|--------|
| Yc8nqtQmbN3r2agw | OdoWell - 1. Router | Recebe webhooks WAHA, classifica intencao, roteia |
| ssRwQYiRWY2j5hfy | OdoWell - 2. Agendamento | Agente LangChain com ferramentas para agendamento |
| UOiHJu2ce8pHGEGD | OdoWell - 3. Consultas | Lista consultas do paciente |
| 1fuQzFFyaaAH8VBQ | OdoWell - 4. FAQ | Responde perguntas sobre a clinica |
| rCQZOa45YUtNHvPt | OdoWell - Limpar Memoria | Limpa memoria Redis (inativo) |

### Webhooks
- Router: `POST /webhook/odowell-router`
- Agendamento: `POST /webhook/odowell-agendamento`
- Consultas: `POST /webhook/odowell-consultas`
- FAQ: `POST /webhook/odowell-faq`

---

## Estrutura do Router

```
Webhook WAHA
    |
Extrair Dados (extrai phone, pushName, message, session, remoteJid)
    |
Msg do Bot? (filtra mensagens do proprio bot)
    |
Tem Mensagem? (verifica se tem conteudo)
    |
Verificar Bloqueio (Redis GET {phone}_block) <-- DEBOUNCE
    |
Ja Processando? (IF bloqueio existe, para)
    |
Tipo Mensagem (texto ou audio)
    |
[Texto] --> Preparar Msg Texto --> Merge
[Audio] --> Download --> Decrypt --> Transcrever --> Preparar Msg Audio --> Merge
    |
Classificar Intencao (Gemini)
    |
Rotear Intencao (Switch)
    |
[AGENDAMENTO] --> Chamar Agendamento
[CONSULTAS] --> Chamar Consultas
[FAQ] --> Chamar FAQ
[SAUDACAO] --> Enviar Saudacao
```

---

## Formato de Dados WAHA

### Webhook recebido do WAHA:
```json
{
  "session": "380947105869",
  "payload": {
    "_data": {
      "Info": {
        "Chat": "14077608242@s.whatsapp.net",
        "IsFromMe": false,
        "PushName": "Nome do Cliente",
        "MediaType": "text"
      },
      "Message": {
        "conversation": "texto da mensagem"
      }
    }
  }
}
```

### Enviar mensagem via WAHA:
```bash
curl -X POST "https://wahazap1979leoale.odowell.pro/api/sendText" \
  -H "X-Api-Key: 28629dc04610b89c16baadcc4233ab6b" \
  -H "Content-Type: application/json" \
  -d '{
    "session": "380947105869",
    "chatId": "14077608242@s.whatsapp.net",
    "text": "Mensagem aqui"
  }'
```

---

## Dentistas Cadastrados

| ID | Nome | Especialidade |
|----|------|---------------|
| 7 | Dra. Marina Santos | Ortodontia (aparelhos) |
| 8 | Dr. Roberto Lima | Implantodontia (implantes) |
| 9 | Dra. Fernanda Costa | Endodontia (canal) |

---

## Base de Conhecimento da Clinica

### Localizacao
- Endereco: Av. das Palmeiras, 1250 - Sala 302, Edificio Medical Center, Jardim Europa, SP
- CEP: 04523-010
- Referencia: Em frente ao Shopping Jardins, ao lado da Farmacia Vida
- Estacionamento: Conveniado no subsolo

### Horarios
- Seg-Sex: 8h as 20h
- Sabado: 8h as 14h
- Aberto no almoco
- Fechado domingos e feriados

### Formas de Pagamento
- Particular: Dinheiro, PIX, Debito, Credito (ate 12x), Boleto
- Convenios: Amil Dental, Bradesco Dental, SulAmerica Odonto, Odontoprev, MetLife, Porto Seguro
- Cobertura convenio: Consultas, Limpeza, Restauracoes, Extracoes, Canal, Raio-X
- NAO COBRE: Clareamento, Implantes (tem particular parcelado)

### Primeira Consulta
- Avaliacao gratuita
- Duracao: 30 minutos

### Contato
- Telefone: (11) 3456-7890

---

## Problemas Corrigidos (2025-12-22)

### 1. Template pushName no Router
**Problema:** `{{ $json.pushName }}` aparecia como texto literal
**Causa:** Formato errado - n8n precisa de `=` antes de expressoes
**Solucao:** Alterado para `={{ 'Ola ' + $json.pushName + '!...' }}`

### 2. Webhook $json.body
**Problema:** Dados do webhook estavam em `$json.body.*` nao em `$json.*`
**Solucao:** Alterado "Salvar Contexto" nos workflows FAQ e Agendamento para usar `$json.body.phone`, `$json.body.session`, etc.

### 3. FAQ nao usava base de conhecimento
**Problema:** Gemini ignorava system message e alucinava
**Solucao:** Incluir base de conhecimento diretamente no prompt text (nao so no system message)

### 4. Mensagem "nao esta cadastrado"
**Problema:** IA dizia "Parece que voce ainda nao esta cadastrado"
**Solucao:** Novo prompt que pede naturalmente: "Para prosseguir, pode me confirmar seu nome completo e data de nascimento?"

### 5. Debounce para mensagens rapidas
**Problema:** Usuario enviava 2 msgs rapidas e recebia 2 respostas
**Solucao:** Adicionado verificacao Redis no Router - se `{phone}_block` existe, nao processa

### 6. Dentista errado para canal
**Problema:** IA mencionava "Dr. Bruno Almeida" para canal
**Solucao:** Corrigido para "Dra. Fernanda Costa" - Endodontia

### 7. Lead sem data de nascimento
**Problema:** Ferramenta salvar_lead nao tinha campo birth_date
**Solucao:**
- Adicionado campo `BirthDate` no modelo Lead (backend)
- Atualizado handler `WhatsAppCreateLead` para aceitar e salvar birth_date
- Atualizado ferramenta `salvar_lead` no n8n para enviar birth_date
- Formato aceito: YYYY-MM-DD ou DD/MM/YYYY

---

## Ferramentas do Agente Agendamento

| Nome | Funcao |
|------|--------|
| verificar_paciente | Verifica se pessoa existe (lead/patient/unknown) |
| salvar_lead | Salva novo lead com nome, telefone e data nascimento |
| converter_lead_paciente | Converte lead em paciente |
| cadastrar_paciente | Cadastra novo paciente completo |
| listar_dentistas | Lista dentistas disponiveis |
| horarios_disponiveis | Busca horarios livres |
| agendar_consulta | Cria agendamento |
| listar_consultas | Lista consultas do paciente |
| cancelar_consulta | Cancela consulta |
| remarcar_consulta | Remarca consulta |

---

## Fluxo de Atendimento Ideal

### Pessoa Nova (nao cadastrada)
1. Cliente: "Oi, quero agendar"
2. Maria: "Para prosseguir, pode me confirmar seu nome completo e data de nascimento?"
3. Cliente: "Tom Brito, 9/7/79"
4. Maria: "Obrigado! Atendemos [responde duvida]. Qual data prefere?"
5. Cliente: "Dia 16 pela manha"
6. Maria: "No dia 16 temos: 8h, 9h, 10h. Qual prefere?"
7. Cliente: "10h"
8. Maria: "Antes de confirmar, preciso do seu endereco completo."
9. Cliente: "Rua Nova, 10, Pilar, Caxias, RJ"
10. Maria: "Agendamento confirmado! Codigo: #123"

### Lead (ja contatou antes)
1. Usa dados do lead
2. Pergunta se quer agendar
3. Se sim: pede endereco, converte para paciente, agenda

### Paciente (ja cadastrado)
1. Cumprimenta pelo nome
2. Vai direto para agendamento

---

## Comandos Uteis

### Testar FAQ
```bash
curl -X POST "https://auto1979leoale.odowell.pro/webhook/odowell-faq" \
  -H "Content-Type: application/json" \
  -d '{"phone": "11999998888", "pushName": "Teste", "message": "onde fica a clinica?", "session": "380947105869", "remoteJid": "11999998888@s.whatsapp.net"}'
```

### Testar Agendamento
```bash
curl -X POST "https://auto1979leoale.odowell.pro/webhook/odowell-agendamento" \
  -H "Content-Type: application/json" \
  -d '{"phone": "11999998888", "pushName": "Teste", "message": "quero agendar", "session": "380947105869", "remoteJid": "11999998888@s.whatsapp.net"}'
```

### Ver execucoes recentes
```bash
curl -s -X GET "https://auto1979leoale.odowell.pro/api/v1/executions?limit=5" \
  -H "X-N8N-API-KEY: [API_KEY]" | jq '.data[] | {id, workflowId, status}'
```

### Ver mensagens WAHA
```bash
curl -s "https://wahazap1979leoale.odowell.pro/api/380947105869/chats/14077608242@s.whatsapp.net/messages?limit=5" \
  -H "X-Api-Key: 28629dc04610b89c16baadcc4233ab6b" | jq '.[].body'
```

---

## Arquivos de Backup

Os workflows foram salvos em `/tmp/` durante a sessao:
- `flow1_router_current.json`
- `flow2_agendamento_current.json`
- `flow4_faq_current.json`

Para backup permanente, exportar via n8n UI ou API.

---

## Agregacao de Mensagens (Pendente)

### Problema
Quando usuario envia mensagens rapidas (ex: "Oi" + "Quero agendar"), cada uma dispara uma execucao separada, gerando respostas picadas.

### Solucao Atual (Parcial)
O Router tem debounce basico via `{phone}_block` no Redis - se ja esta processando uma resposta, novas mensagens sao ignoradas. Isso evita duplicatas mas nao agrega mensagens.

### Limitacao do n8n
O node "Wait" nao funciona bem com webhooks pois requer configuracao especial de resumo.

### Alternativas para Implementar
1. **Configurar no WAHA** - Verificar se WAHA tem opcao de "batching"
2. **Workflow Agendado** - Workflow separado que roda a cada X segundos, processa mensagens acumuladas
3. **Code Node com delay** - setTimeout no Code node (tem limitacoes)

---

## Proximos Passos Sugeridos

1. [ ] Testar fluxo completo via WhatsApp real
2. [ ] Implementar agregacao de mensagens (escolher abordagem)
3. [ ] Ativar workflow "Limpar Memoria" com webhook
4. [ ] Adicionar logs/monitoramento
5. [ ] Configurar alertas para erros

---

## Contato
- Usuario n8n: appodowell@hotmail.com (Wellington Rodrigo B S)
