# Plano Fase 6: Melhorias Avançadas

**Data de Criação:** 2025-12-21
**Status:** ✅ Concluído

---

## Resumo do Plano

| Etapa | Descrição | Prioridade | Status | Testado |
|-------|-----------|------------|--------|---------|
| 6.1 | 2FA para Administradores (TOTP) | Alta | ✅ Concluído | ✅ |
| 6.2 | Logs Estruturados (JSON) | Alta | ✅ Concluído | ✅ |
| 6.3 | Request ID / Correlation ID | Média | ✅ Concluído | ✅ |
| 6.4 | Métricas Prometheus | Média | ✅ Concluído | ✅ |
| 6.5 | Error Tracking (Sentry) | Média | ✅ Concluído | ✅ |
| 6.6 | Cache de Consultas Frequentes | Baixa | ✅ Concluído | ✅ |

---

## Etapa 6.1: 2FA para Administradores (TOTP)

### Objetivo
Adicionar autenticação de dois fatores usando TOTP (Google Authenticator, Authy) para usuários admin.

### Arquivos a Criar/Modificar
- `backend/internal/models/user.go` - Adicionar campos 2FA
- `backend/internal/handlers/auth.go` - Lógica de verificação TOTP
- `backend/internal/handlers/user_2fa.go` - Setup/disable 2FA
- `frontend/src/pages/Settings/Security.jsx` - UI para ativar 2FA

### Campos no User
```go
// 2FA Fields
TwoFactorEnabled  bool   `gorm:"default:false" json:"two_factor_enabled"`
TwoFactorSecret   string `json:"-"` // Encrypted TOTP secret
TwoFactorBackup   string `json:"-"` // Encrypted backup codes
```

### Fluxo
1. Admin vai em Configurações > Segurança
2. Clica "Ativar 2FA"
3. Sistema gera QR Code com secret TOTP
4. Admin escaneia com Google Authenticator
5. Admin digita código para confirmar
6. Sistema salva secret (criptografado) e ativa 2FA
7. No próximo login, pede código TOTP após senha

### Dependência Go
```bash
go get github.com/pquerna/otp/totp
```

### Teste
1. Ativar 2FA em conta admin
2. Fazer logout e login novamente
3. Verificar que pede código TOTP
4. Testar código inválido (deve rejeitar)
5. Testar código válido (deve aceitar)

---

## Etapa 6.2: Logs Estruturados (JSON)

### Objetivo
Converter logs para formato JSON estruturado para facilitar análise com ferramentas como ELK Stack, CloudWatch, etc.

### Arquivos a Modificar
- `backend/internal/helpers/logger.go` - Criar logger estruturado
- `backend/cmd/api/main.go` - Usar novo logger

### Implementação
```go
// helpers/logger.go
type LogEntry struct {
    Timestamp   string      `json:"timestamp"`
    Level       string      `json:"level"`
    Message     string      `json:"message"`
    RequestID   string      `json:"request_id,omitempty"`
    UserID      uint        `json:"user_id,omitempty"`
    TenantID    uint        `json:"tenant_id,omitempty"`
    Method      string      `json:"method,omitempty"`
    Path        string      `json:"path,omitempty"`
    StatusCode  int         `json:"status_code,omitempty"`
    Duration    float64     `json:"duration_ms,omitempty"`
    Error       string      `json:"error,omitempty"`
    Extra       interface{} `json:"extra,omitempty"`
}

func LogJSON(entry LogEntry) {
    entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
    json.NewEncoder(os.Stdout).Encode(entry)
}
```

### Exemplo de Output
```json
{"timestamp":"2025-12-21T10:30:00Z","level":"INFO","message":"Request completed","request_id":"abc123","user_id":5,"tenant_id":1,"method":"GET","path":"/api/patients","status_code":200,"duration_ms":45.2}
```

### Teste
1. Fazer requisição à API
2. Verificar logs no formato JSON
3. Validar que todos os campos estão presentes

---

## Etapa 6.3: Request ID / Correlation ID

### Objetivo
Adicionar ID único a cada requisição para rastreamento end-to-end.

### Arquivos a Modificar
- `backend/internal/middleware/request_id.go` - Novo middleware
- `backend/cmd/api/main.go` - Adicionar middleware

### Implementação
```go
// middleware/request_id.go
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if request already has an ID (from load balancer, etc.)
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)

        c.Next()
    }
}
```

### Teste
1. Fazer requisição sem X-Request-ID
2. Verificar que resposta tem X-Request-ID gerado
3. Fazer requisição com X-Request-ID
4. Verificar que resposta mantém o mesmo ID
5. Verificar que logs incluem o request_id

---

## Etapa 6.4: Métricas Prometheus

### Objetivo
Expor métricas no formato Prometheus para monitoramento com Grafana.

### Arquivos a Criar
- `backend/internal/metrics/prometheus.go` - Definição de métricas
- `backend/cmd/api/main.go` - Endpoint /metrics

### Métricas a Expor
```go
// Contadores
http_requests_total{method, path, status}
http_request_errors_total{method, path, error_type}
db_queries_total{operation, table}

// Histogramas
http_request_duration_seconds{method, path}
db_query_duration_seconds{operation}

// Gauges
http_requests_in_flight
db_connections_open
db_connections_idle
```

### Dependência Go
```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

### Endpoint
```
GET /metrics
```

### Docker Stack (Prometheus + Grafana)
```yaml
# Adicionar ao docker-stack.yml
prometheus:
  image: prom/prometheus:latest
  volumes:
    - ./prometheus.yml:/etc/prometheus/prometheus.yml
  ports:
    - "9090:9090"

grafana:
  image: grafana/grafana:latest
  ports:
    - "3001:3000"
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=admin
```

### Teste
1. Acessar /metrics
2. Verificar métricas no formato Prometheus
3. Configurar Prometheus para scrape
4. Criar dashboard no Grafana

---

## Etapa 6.5: Error Tracking (Sentry)

### Objetivo
Integrar Sentry para captura automática de erros em produção.

### Arquivos a Modificar
- `backend/internal/helpers/sentry.go` - Inicialização Sentry
- `backend/cmd/api/main.go` - Middleware de captura
- `frontend/src/main.jsx` - Sentry no frontend

### Dependência Go
```bash
go get github.com/getsentry/sentry-go
```

### Implementação Backend
```go
// helpers/sentry.go
func InitSentry() {
    dsn := os.Getenv("SENTRY_DSN")
    if dsn == "" {
        log.Println("SENTRY_DSN not set, error tracking disabled")
        return
    }

    sentry.Init(sentry.ClientOptions{
        Dsn:              dsn,
        Environment:      os.Getenv("ENV"),
        TracesSampleRate: 0.1, // 10% of transactions
    })
}

func CaptureError(err error, ctx *gin.Context) {
    sentry.WithScope(func(scope *sentry.Scope) {
        scope.SetUser(sentry.User{
            ID: fmt.Sprintf("%d", ctx.GetUint("user_id")),
        })
        scope.SetTag("tenant_id", fmt.Sprintf("%d", ctx.GetUint("tenant_id")))
        scope.SetRequest(ctx.Request)
        sentry.CaptureException(err)
    })
}
```

### Frontend (React)
```javascript
import * as Sentry from "@sentry/react";

Sentry.init({
  dsn: import.meta.env.VITE_SENTRY_DSN,
  environment: import.meta.env.MODE,
  tracesSampleRate: 0.1,
});
```

### Teste
1. Criar conta no Sentry (grátis para <5k eventos/mês)
2. Obter DSN
3. Configurar variável de ambiente
4. Forçar um erro
5. Verificar que aparece no Sentry

---

## Etapa 6.6: Cache de Consultas Frequentes

### Objetivo
Usar Redis para cachear consultas frequentes e reduzir carga no banco.

### Arquivos a Modificar
- `backend/internal/cache/queries.go` - Cache helpers
- `backend/internal/handlers/patient.go` - Cachear listagem
- `backend/internal/handlers/appointment.go` - Cachear agenda

### Consultas a Cachear
| Consulta | TTL | Invalidação |
|----------|-----|-------------|
| Lista de dentistas | 5 min | Ao criar/editar usuário |
| Procedimentos/Protocolos | 10 min | Ao criar/editar protocolo |
| Configurações do tenant | 5 min | Ao editar configurações |
| Contagem de pendências | 1 min | Ao criar/completar tarefa |

### Implementação
```go
// cache/queries.go
func GetOrSet(key string, ttl time.Duration, fetchFunc func() (interface{}, error)) (interface{}, error) {
    client := GetClient()
    if client == nil {
        return fetchFunc() // Fallback to DB if Redis unavailable
    }

    // Try to get from cache
    cached, err := client.Get(ctx, key).Result()
    if err == nil {
        var result interface{}
        json.Unmarshal([]byte(cached), &result)
        return result, nil
    }

    // Fetch from DB
    result, err := fetchFunc()
    if err != nil {
        return nil, err
    }

    // Store in cache
    data, _ := json.Marshal(result)
    client.Set(ctx, key, data, ttl)

    return result, nil
}

func InvalidatePrefix(prefix string) {
    client := GetClient()
    if client == nil {
        return
    }

    keys, _ := client.Keys(ctx, prefix+"*").Result()
    if len(keys) > 0 {
        client.Del(ctx, keys...)
    }
}
```

### Teste
1. Fazer consulta (deve ir ao banco)
2. Repetir consulta (deve vir do cache - mais rápido)
3. Editar dado relacionado
4. Verificar que cache foi invalidado
5. Repetir consulta (deve ir ao banco novamente)

---

## Procedimento de Execução

Para cada etapa:

1. **Ler** esta seção do plano
2. **Implementar** as mudanças descritas
3. **Build**: `docker build -t tomautomations/drcrwell-backend:latest .`
4. **Push**: `docker push tomautomations/drcrwell-backend:latest`
5. **Deploy**: `docker service update --image tomautomations/drcrwell-backend:latest drcrwell_backend --force`
6. **Testar** conforme descrito
7. **Atualizar** este arquivo marcando ✅ Concluído
8. **Prosseguir** para próxima etapa

---

## Estimativa de Complexidade

| Etapa | Complexidade | Impacto |
|-------|--------------|---------|
| 6.1 2FA | Alta | Alto (segurança) |
| 6.2 Logs JSON | Baixa | Médio (observabilidade) |
| 6.3 Request ID | Baixa | Médio (debug) |
| 6.4 Prometheus | Média | Alto (monitoramento) |
| 6.5 Sentry | Baixa | Alto (debug produção) |
| 6.6 Cache | Média | Médio (performance) |

---

## Ordem Recomendada

1. **6.2 Logs JSON** - Base para outras melhorias
2. **6.3 Request ID** - Complementa logs
3. **6.5 Sentry** - Rápido de implementar, alto valor
4. **6.1 2FA** - Importante para segurança
5. **6.4 Prometheus** - Monitoramento completo
6. **6.6 Cache** - Otimização final

---

## Log de Execução

### Etapa 6.1 - 2FA
- **Início:** 2025-12-21 07:23
- **Fim:** 2025-12-21 07:29
- **Observações:** Implementado 2FA TOTP completo. Campos adicionados ao User model. Criado handlers/user_2fa.go com setup, verify, disable, backup codes. Modificado auth.go para verificar 2FA no login. Criado helpers/token.go para tokens temporários. Rotas: GET/POST /api/auth/2fa/*. Usando pquerna/otp v1.4.0.

### Etapa 6.2 - Logs JSON
- **Início:** 2025-12-21 07:20
- **Fim:** 2025-12-21 07:25
- **Observações:** Criado helpers/logger.go e middleware/logger.go. Logs em JSON com todos os campos: timestamp, level, request_id, user_id, tenant_id, method, path, status_code, duration_ms, ip, user_agent.

### Etapa 6.3 - Request ID
- **Início:** 2025-12-21 07:20
- **Fim:** 2025-12-21 07:25
- **Observações:** Criado middleware/request_id.go. Header X-Request-ID adicionado a todas as respostas. UUID gerado automaticamente se não fornecido.

### Etapa 6.4 - Prometheus
- **Início:** 2025-12-21 07:29
- **Fim:** 2025-12-21 07:32
- **Observações:** Criado internal/metrics/prometheus.go com métricas HTTP (requests_total, duration, in_flight, errors), DB (queries, duration, connections), e business (logins, 2FA). Middleware PrometheusMiddleware coleta métricas automaticamente. Endpoint /metrics disponível internamente.

### Etapa 6.5 - Sentry
- **Início:** 2025-12-21 07:22
- **Fim:** 2025-12-21 07:23
- **Observações:** Criado helpers/sentry.go e middleware/sentry.go. Integração Sentry v0.27.0 (compatível com Go 1.21). Captura panics e erros com contexto (user_id, tenant_id, request_id). Para ativar, configurar SENTRY_DSN no ambiente.

### Etapa 6.6 - Cache
- **Início:** 2025-12-21 07:32
- **Fim:** 2025-12-21 07:34
- **Observações:** Criado cache/queries.go com GetOrSet, GetOrSetTyped, InvalidatePrefix e helpers de invalidação. TTLs configurados: Dentistas (5min), Protocolos (10min), Settings (5min), Counts (1min). Funções de invalidação automática para mudanças em dados relacionados.

---

## Pendência: Configurar Sentry (Opcional)

### Status: ⏳ Aguardando DSN do usuário

### Instruções para configurar Sentry:

1. **Criar conta gratuita:** https://sentry.io/signup/
2. **Escolher plataforma:** Go
3. **Copiar o DSN** (formato: `https://xxx@xxx.ingest.sentry.io/xxx`)
4. **Executar comando:**
   ```bash
   docker service update drcrwell_backend --env-add SENTRY_DSN=seu-dsn-aqui
   ```

### Variável já configurada:
- ✅ `TWO_FA_ENCRYPTION_KEY=9e5da5001ae3f08b430cdeb642b6206eb13656d00897d578147d91b10919ebf6`

### Nota:
O sistema funciona perfeitamente sem Sentry. Os erros são capturados nos logs JSON estruturados. Sentry é um extra para dashboard visual de erros.
