# Monitoramento OdoWell

## 1. Monitoramento Interno (Automático)

O script `/root/drcrwell/scripts/health_monitor.sh` executa a cada 5 minutos e verifica:
- API Backend (https://api.odowell.pro/health)
- Frontend (https://app.odowell.pro)
- PostgreSQL
- Redis
- Disco (alerta se > 85%)
- Memória (alerta se > 90%)
- Swarm nodes

Logs: `/root/drcrwell/logs/health_monitor.log`

## 2. UptimeRobot (Monitoramento Externo)

### Configurar em https://uptimerobot.com (gratuito):

1. Criar conta em https://uptimerobot.com
2. Adicionar os seguintes monitors:

| Nome | URL | Tipo | Intervalo |
|------|-----|------|-----------|
| OdoWell API | https://api.odowell.pro/health | HTTP(s) | 5 min |
| OdoWell Frontend | https://app.odowell.pro | HTTP(s) | 5 min |

3. Configurar alertas por email/SMS/Telegram

### Alternativas ao UptimeRobot:
- Pingdom (https://pingdom.com)
- Better Uptime (https://betteruptime.com)
- Cronitor (https://cronitor.io)

## 3. Alertas

Configurar email de alerta no arquivo `.env`:
```
ALERT_EMAIL=seu-email@exemplo.com
```

## 4. Verificar Status

```bash
# Ver logs de health check
tail -50 /root/drcrwell/logs/health_monitor.log

# Executar check manual
/root/drcrwell/scripts/health_monitor.sh

# Ver status dos services
docker service ls
docker node ls
```

## 5. Backups

- Automático: 00:00 e 12:00 diariamente
- Destino: S3 (odowell-app bucket)
- Retenção local: 7 dias
- Logs: `/root/drcrwell/backups/backup.log`

```bash
# Executar backup manual
/root/drcrwell/scripts/backup.sh

# Ver backups
ls -la /root/drcrwell/backups/*.sql.gz
```
