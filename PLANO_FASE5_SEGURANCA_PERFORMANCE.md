# Plano Fase 5: Segurança, Performance e Confiabilidade

**Data de Criação:** 2025-12-21
**Status:** ✅ CONCLUÍDO (2025-12-21)

---

## Resumo do Plano

| Etapa | Descrição | Status | Testado |
|-------|-----------|--------|---------|
| 5.1 | Security Headers | ✅ Concluído | ✅ |
| 5.2 | Graceful Shutdown | ✅ Concluído | ✅ |
| 5.3 | Backup Automático (Cron) | ✅ Concluído | ✅ |
| 5.4 | Rate Limit com Redis | ✅ Concluído | ✅ |
| 5.5 | Query Slow Log | ✅ Concluído | ✅ |
| 5.6 | Health Check Avançado | ✅ Concluído | ✅ |

---

## Etapa 5.1: Security Headers

### Objetivo
Adicionar headers de segurança HTTP para proteger contra ataques comuns.

### Arquivos a Modificar
- `backend/cmd/api/main.go` - Adicionar middleware de security headers

### Headers a Implementar
```go
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### Teste
1. Build do backend
2. Deploy
3. Verificar headers com: `curl -I https://api.odowell.pro/health`

### Critério de Sucesso
- Todos os headers aparecem na resposta
- API continua funcionando normalmente

---

## Etapa 5.2: Graceful Shutdown

### Objetivo
Implementar shutdown gracioso para não perder requisições durante deploys.

### Arquivos a Modificar
- `backend/cmd/api/main.go` - Usar http.Server com graceful shutdown

### Implementação
```go
// Criar servidor HTTP
srv := &http.Server{
    Addr:    ":8080",
    Handler: r,
}

// Goroutine para escutar sinais
go func() {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    srv.Shutdown(ctx)
}()

// Iniciar servidor
srv.ListenAndServe()
```

### Teste
1. Build do backend
2. Deploy
3. Verificar logs de shutdown
4. API continua funcionando

### Critério de Sucesso
- Logs mostram "Shutting down gracefully..."
- Sem erros 502 durante deploy

---

## Etapa 5.3: Backup Automático (Cron)

### Objetivo
Agendar backups automáticos diários do banco de dados.

### Arquivos a Criar/Modificar
- Adicionar cron job no sistema

### Implementação
```bash
# Adicionar ao crontab
0 3 * * * /root/drcrwell/scripts/backup.sh >> /var/log/odowell-backup.log 2>&1
```

### Teste
1. Executar backup manual: `./scripts/backup.sh`
2. Verificar arquivo criado em `/root/drcrwell/backups/`
3. Adicionar cron job
4. Verificar log no dia seguinte

### Critério de Sucesso
- Backup diário às 3h da manhã
- Arquivos com retenção de 7 dias
- Log de execução disponível

---

## Etapa 5.4: Rate Limit com Redis

### Objetivo
Migrar rate limiting de memória para Redis (persistente entre restarts).

### Arquivos a Modificar
- `backend/internal/middleware/ratelimit_redis.go` - Já existe, verificar se está em uso

### Teste
1. Verificar se ratelimit_redis.go está sendo usado
2. Se não, integrar no main.go
3. Testar com múltiplas requisições rápidas
4. Verificar chaves no Redis

### Critério de Sucesso
- Rate limit persiste após restart do backend
- Chaves visíveis no Redis

---

## Etapa 5.5: Query Slow Log

### Objetivo
Logar queries que demoram mais de 200ms para identificar gargalos.

### Arquivos a Modificar
- `backend/internal/database/database.go` - Configurar GORM logger

### Implementação
```go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold: 200 * time.Millisecond,
            LogLevel:      logger.Warn,
            Colorful:      true,
        },
    ),
})
```

### Teste
1. Build do backend
2. Deploy
3. Fazer consultas pesadas
4. Verificar logs de queries lentas

### Critério de Sucesso
- Queries > 200ms aparecem nos logs com WARN

---

## Etapa 5.6: Health Check Avançado

### Objetivo
Expandir health check para incluir mais métricas.

### Arquivos a Modificar
- `backend/cmd/api/main.go` - Endpoint /health

### Métricas a Adicionar
- Uptime do servidor
- Número de goroutines
- Uso de memória
- Latência do banco (ping time)
- Conexões ativas do pool

### Teste
1. Build do backend
2. Deploy
3. Verificar resposta de `curl https://api.odowell.pro/health`

### Critério de Sucesso
- Resposta inclui todas as métricas
- Valores são precisos

---

## Procedimento de Execução

Para cada etapa:

1. **Ler** esta seção do plano
2. **Implementar** as mudanças descritas
3. **Build** do Docker: `docker build -t tomautomations/drcrwell-backend:latest .`
4. **Push**: `docker push tomautomations/drcrwell-backend:latest`
5. **Deploy**: `docker service update --image tomautomations/drcrwell-backend:latest drcrwell_backend --force`
6. **Testar** conforme descrito na seção de teste
7. **Atualizar** este arquivo marcando ✅ Concluído
8. **Prosseguir** para próxima etapa

---

## Rollback

Se algo quebrar:
```bash
# Ver imagens anteriores
docker image ls tomautomations/drcrwell-backend

# Reverter para versão anterior
docker service update --image tomautomations/drcrwell-backend:previous drcrwell_backend --force
```

---

## Log de Execução

### Etapa 5.1 - Security Headers
- **Início:** 2025-12-21 06:50
- **Fim:** 2025-12-21 06:55
- **Observações:** Criado middleware security_headers.go. Todos os 7 headers funcionando: X-Frame-Options, X-Content-Type-Options, X-XSS-Protection, HSTS, Referrer-Policy, Permissions-Policy, CSP.

### Etapa 5.2 - Graceful Shutdown
- **Início:** 2025-12-21 06:58
- **Fim:** 2025-12-21 07:03
- **Observações:** Implementado http.Server com shutdown gracioso. Timeouts configurados: Read 30s, Write 30s, Idle 60s. Shutdown timeout: 30s.

### Etapa 5.3 - Backup Automático
- **Início:** 2025-12-21 07:03
- **Fim:** 2025-12-21 07:05
- **Observações:** Já estava configurado! Cron às 3h e 15h. Upload S3 funcionando. Retenção 7 dias. Logs em /root/drcrwell/backups/backup.log.

### Etapa 5.4 - Rate Limit Redis
- **Início:** 2025-12-21 07:05
- **Fim:** 2025-12-21 07:08
- **Observações:** Já estava implementado! ratelimit_redis.go com fallback para memória. Login: 5/min, WhatsApp: 200/min. Bloqueia 15min após exceder.

### Etapa 5.5 - Query Slow Log
- **Início:** 2025-12-21 07:08
- **Fim:** 2025-12-21 07:12
- **Observações:** Configurado GORM logger com SlowThreshold=200ms. Queries lentas aparecem nos logs como WARN. IgnoreRecordNotFoundError ativado.

### Etapa 5.6 - Health Check Avançado
- **Início:** 2025-12-21 07:12
- **Fim:** 2025-12-21 07:18
- **Observações:** Adicionado métricas: uptime, goroutines, memória, latência PG, pool DB (open/idle/in_use/max/wait), versão Go. Resposta JSON estruturada.
