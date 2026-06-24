# Deploy no Railway

Guia para publicar **Luxus.Connect** no [Railway](https://railway.app) com três serviços principais: **PostgreSQL**, **connect-api** e **connect-web**.

## Arquitetura recomendada

| Serviço Railway | Root Directory | Descrição |
|-----------------|----------------|-----------|
| **PostgreSQL** | (plugin) | Banco `luxus_connect` — variável `DATABASE_URL` injetada na API |
| **connect-api** | `api` | API Go (`Dockerfile` + `railway.toml`) |
| **connect-web** | `src/Luxus.Connect.Web` | SPA React (`Dockerfile` + `railway.toml`) |

Serviços adicionais (fora deste repo ou em serviços separados):

- **Keycloak** — autenticação OIDC (imagem Docker ou instância gerida)
- **RabbitMQ** — fila de importação (CloudAMQP ou plugin)
- **Object storage** — S3/R2/MinIO para anexos e faturas operadora

## 1. Criar projeto

1. [railway.app](https://railway.app) → **New Project**
2. **Add PostgreSQL**
3. **Empty Service** → conectar ao GitHub `telefonia-2` → renomear para `connect-api`
   - Settings → **Root Directory**: `api`
   - Settings → **Networking** → Generate Domain
4. **Empty Service** → mesmo repo → renomear para `connect-web`
   - Root Directory: `src/Luxus.Connect.Web`
   - Generate Domain

## 2. Variáveis da API (`connect-api`)

Referência completa: [`.env.railway.example`](../.env.railway.example)

Obrigatórias:

```env
ENVIRONMENT=Production
DATABASE_URL=${{Postgres.DATABASE_URL}}
KEYCLOAK_REALM=luxus
KEYCLOAK_RESOURCE=connect-cli
KEYCLOAK_AUTH_SERVER_URL=https://<seu-keycloak>
KEYCLOAK_PUBLIC_AUTH_SERVER_URL=https://<dominio-publico>/auth
CORS_ORIGINS=https://<dominio-do-connect-web>
```

Opcionais (Sicredi, fila, storage, SMTP): ver `.env.railway.example`.

**Railway:** se `SICREDI_PUBLIC_API_URL` e `CORS_ORIGINS` estiverem vazios, a API usa automaticamente `https://$RAILWAY_PUBLIC_DOMAIN` do serviço.

Na API → **Variables** → referencie o Postgres:

- `DATABASE_URL` = `${{Postgres.DATABASE_URL}}` (nome do plugin pode variar)

## 3. Variáveis do frontend (`connect-web`)

Marcar como **disponíveis no build** (Railway → Variables → ☑ Build):

```env
VITE_API_URL=https://<dominio-publico-da-api>
VITE_AUTH_URL=https://<dominio-keycloak-ou-proxy>/auth
VITE_CLIENT_ID=connect-cli
VITE_CLIENT_SECRET=<secret-do-realm>
VITE_STORAGE_BUCKET_NAME=luxus-connect
```

## 4. Schema do banco

A API **não** roda migrações no startup. Após o primeiro deploy da API:

```bash
# Com Railway CLI instalado e projeto linkado
railway link
railway run --service connect-api -- psql "$DATABASE_URL" -f db/migrations/001_initial_schema.sql
```

Ou execute os scripts `db/migrations/*.sql` em ordem numérica (001 … 013) via cliente SQL conectado ao Postgres do Railway.

## 5. Keycloak

O realm de desenvolvimento está em `docker/keycloak/luxus-realm.json`. Em produção:

1. Suba Keycloak (serviço Railway com imagem `quay.io/keycloak/keycloak` ou instância externa)
2. Importe o realm e ajuste **redirect URIs** para o domínio do `connect-web`
3. Aponte `KEYCLOAK_*` na API e `VITE_AUTH_*` no web

## 6. Sicredi (webhook)

1. Deploy da API com domínio público Railway
2. `SICREDI_PUBLIC_API_URL` = `https://<api>.up.railway.app` (ou deixe vazio para auto)
3. Na UI: **Faturas → Testar conexão → Registrar webhook**

## 7. Health checks

| Serviço | Path |
|---------|------|
| connect-api | `/health` |
| connect-web | `/` |

Configurados em `api/railway.toml` e `src/Luxus.Connect.Web/railway.toml`.

## 8. Deploy contínuo

Push na branch `main` → Railway reconstrói os serviços cujo root directory mudou.

```bash
git push origin main
```

## Troubleshooting

| Problema | Solução |
|----------|---------|
| API não conecta ao Postgres | Verifique `DATABASE_URL` e SSL (`sslmode=require` é suportado) |
| CORS no browser | Inclua o domínio exato do web em `CORS_ORIGINS` |
| Web em branco / 502 | Confirme `VITE_*` no build e regenere deploy do web |
| Sicredi 403 | `SICREDI_SANDBOX=false` para credenciais de produção |
| Auth falha | `KEYCLOAK_PUBLIC_AUTH_SERVER_URL` deve ser a URL que o browser alcança |
