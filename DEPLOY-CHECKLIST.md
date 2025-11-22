# ✅ Checklist de Deploy - Dr. Crwell

## Antes de fazer qualquer deploy:

### 1. ✅ Verificar CORS
```bash
# Verificar .env
grep CORS_ORIGINS .env

# Verificar no container rodando
docker exec $(docker ps -q -f name=drcrwell_backend) sh -c "env | grep CORS"

# Deve mostrar:
# CORS_ORIGINS=https://dr.crwell.pro,http://localhost:3000
```

### 2. ✅ Verificar variáveis de ambiente exportadas
```bash
# Antes do deploy, sempre fazer:
source .env
export CORS_ORIGINS
export FRONTEND_URL
export BACKEND_URL
export DOCKER_USERNAME
export AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY
export AWS_BUCKET_NAME
export AWS_REGION
```

### 3. ✅ Build e Deploy Completo
```bash
# Opção 1: Script automático (recomendado)
./deploy.sh

# Opção 2: Manual
source .env
make build
make push
make deploy
```

### 4. ✅ Após Deploy - Verificações

**Backend:**
```bash
# Logs
docker service logs drcrwell_backend --tail 50

# Verificar CORS no container
docker exec $(docker ps -q -f name=drcrwell_backend) env | grep CORS

# Testar API
curl https://drapi.crwell.pro/health
```

**Frontend:**
```bash
# Logs
docker service logs drcrwell_frontend --tail 50

# Verificar URL da API no build
docker exec $(docker ps -q -f name=drcrwell_frontend) grep -r "drapi.crwell.pro" /usr/share/nginx/html/assets/ | head -1
```

### 5. ✅ Teste Funcional
- [ ] Login funciona
- [ ] Dashboard carrega
- [ ] Edição de registros funciona (pacientes, agendamentos, etc.)
- [ ] Não há erros de CORS no console F12
- [ ] Upload e download de exames funcionam (se usando S3)

## ⚠️ Problemas Comuns

### CORS Error
**Sintoma:** Erro 403 ou erros de CORS no console do navegador

**Solução:**
```bash
source .env
docker service update --env-add CORS_ORIGINS="${CORS_ORIGINS}" drcrwell_backend
```

### Frontend não conecta ao Backend
**Sintoma:** Chamadas API vão para localhost:8080

**Solução:**
```bash
# Rebuild frontend com VITE_API_URL correto
source .env
docker build --build-arg VITE_API_URL=https://${BACKEND_URL} -t ${DOCKER_USERNAME}/drcrwell-frontend:latest ./frontend
docker push ${DOCKER_USERNAME}/drcrwell-frontend:latest
docker service update --force --image ${DOCKER_USERNAME}/drcrwell-frontend:latest drcrwell_frontend
```

### Erro 500 em Updates
**Sintoma:** PUT/PATCH retorna erro 500

**Solução:** Verificar logs do backend:
```bash
docker service logs drcrwell_backend --tail 100 | grep ERROR
```

### Exames retornam AccessDenied (403)
**Sintoma:** Erro XML "AccessDenied" ao visualizar/baixar exames

**Causa:** Credenciais AWS não configuradas ou sem permissões

**Solução:**
```bash
# 1. Verificar se credenciais estão no container
docker exec $(docker ps -q -f name=drcrwell_backend) env | grep AWS

# 2. Se vazias, atualizar serviço:
source .env
docker service update \
  --env-add AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
  --env-add AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
  --env-add AWS_BUCKET_NAME="${AWS_BUCKET_NAME}" \
  --env-add AWS_REGION="${AWS_REGION}" \
  drcrwell_backend

# 3. Verificar política IAM no console AWS
# O usuário precisa ter s3:GetObject, s3:PutObject, s3:DeleteObject
```

## 📝 Nota Importante

**SEMPRE** que fizer mudanças no código, siga este checklist completo para evitar problemas de CORS ou configuração!
