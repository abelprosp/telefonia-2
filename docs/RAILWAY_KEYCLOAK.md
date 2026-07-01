# Keycloak no Railway — Setup de Produção

Guia completo para importar o realm **luxus** no Keycloak em produção no Railway e conectá-lo ao `connect-api` e ao `connect-web`.

## URLs de referência

| Serviço       | URL de produção                                          |
|---------------|----------------------------------------------------------|
| Keycloak      | https://keycloak-production-734c.up.railway.app          |
| connect-web   | https://connect-web-production-e247.up.railway.app       |
| connect-api   | https://telefonica-2-production.up.railway.app           |

---

## 1. Importar o realm via Admin UI

1. Acesse o Keycloak Admin Console:
   `https://keycloak-production-734c.up.railway.app/admin`

2. Faça login com as credenciais de administrador (`KC_BOOTSTRAP_ADMIN_USERNAME` / `KC_BOOTSTRAP_ADMIN_PASSWORD` definidas no serviço Railway).

3. No menu lateral, clique em **Keycloak** (dropdown do realm atual, canto superior esquerdo) → **Create realm**.

4. Clique em **Browse** e selecione o arquivo `docker/keycloak/luxus-realm.json` deste repositório.

5. Clique em **Create**. O realm `luxus` será criado com:
   - 6 roles: `admin`, `user`, `partner`, `master`, `employee`, `financial`
   - 4 usuários de teste: `dev/dev`, `parceiro/parceiro`, `funcionario/funcionario`, `financeiro/financeiro`
   - 2 client scopes: `organization`, `luxus-roles`
   - 1 cliente: `connect-cli` (public, OIDC)
   - Redirect URIs e Web Origins já incluindo a URL de produção do `connect-web`

---

## 2. Verificar o client connect-cli

Após a importação, confirme que o client está configurado corretamente:

1. No realm `luxus`, vá em **Clients** → `connect-cli`.
2. Na aba **Settings**, verifique:
   - **Valid redirect URIs** deve conter `https://connect-web-production-e247.up.railway.app/*`
   - **Web origins** deve conter `https://connect-web-production-e247.up.railway.app`
3. Se precisar adicionar manualmente, clique em **Add valid redirect URIs** / **Add web origins** e salve.

> O arquivo `luxus-realm.json` já inclui essas URLs, então a importação deve configurá-las automaticamente.

---

## 3. Variáveis de ambiente — connect-api

No serviço `connect-api` do Railway, configure as seguintes variáveis:

```env
KEYCLOAK_REALM=luxus
KEYCLOAK_RESOURCE=connect-cli
KEYCLOAK_AUTH_SERVER_URL=https://keycloak-production-734c.up.railway.app
CORS_ORIGINS=https://connect-web-production-e247.up.railway.app
```

> `KEYCLOAK_AUTH_SERVER_URL` aponta para a raiz do Keycloak (sem `/auth` — Keycloak 17+ não usa mais esse prefixo por padrão). Ajuste se o seu serviço usar `KC_HTTP_RELATIVE_PATH=/auth`.

---

## 4. Variáveis de ambiente — connect-web

No serviço `connect-web` do Railway, configure as seguintes variáveis e marque todas como **disponíveis no build** (☑ Build):

```env
VITE_API_URL=https://telefonica-2-production.up.railway.app
VITE_AUTH_URL=https://keycloak-production-734c.up.railway.app
VITE_CLIENT_ID=connect-cli
```

Após alterar variáveis de build, faça **redeploy** do serviço `connect-web` para que o Vite as incorpore no bundle.

---

## 5. Testar se está funcionando

### 5.1 Verificar o realm via OIDC discovery

```bash
curl https://keycloak-production-734c.up.railway.app/realms/luxus/.well-known/openid-configuration
```

A resposta deve ser um JSON com `issuer`, `authorization_endpoint`, `token_endpoint`, etc.

### 5.2 Obter token de acesso (Resource Owner Password — apenas para testes)

```bash
curl -s -X POST \
  https://keycloak-production-734c.up.railway.app/realms/luxus/protocol/openid-connect/token \
  -d "client_id=connect-cli" \
  -d "grant_type=password" \
  -d "username=dev" \
  -d "password=dev" \
  | jq .
```

A resposta deve conter `access_token`. Inspecione o token em [jwt.io](https://jwt.io) e confirme que os claims `roles` e `organization` estão presentes.

### 5.3 Verificar CORS

Abra o `connect-web` em produção e tente fazer login. Se o browser reportar erro de CORS, confirme que:
- `Web origins` no client `connect-cli` inclui `https://connect-web-production-e247.up.railway.app` (sem barra final).
- `CORS_ORIGINS` na `connect-api` inclui a mesma origem.

### 5.4 Verificar a API

```bash
TOKEN=$(curl -s -X POST \
  https://keycloak-production-734c.up.railway.app/realms/luxus/protocol/openid-connect/token \
  -d "client_id=connect-cli" \
  -d "grant_type=password" \
  -d "username=dev" \
  -d "password=dev" \
  | jq -r .access_token)

curl -H "Authorization: Bearer $TOKEN" \
  https://telefonica-2-production.up.railway.app/health
```

---

## 6. Alternativa: importar via --import-realm (docker-compose)

Se você estiver rodando o Keycloak localmente com Docker Compose, pode importar o realm automaticamente na inicialização:

```yaml
# docker-compose.yml (trecho)
keycloak:
  image: quay.io/keycloak/keycloak:latest
  command: start-dev --import-realm
  environment:
    KC_BOOTSTRAP_ADMIN_USERNAME: admin
    KC_BOOTSTRAP_ADMIN_PASSWORD: admin
  volumes:
    - ./docker/keycloak/luxus-realm.json:/opt/keycloak/data/import/luxus-realm.json
```

O Keycloak importa automaticamente todos os arquivos `.json` em `/opt/keycloak/data/import/` na inicialização quando a flag `--import-realm` está presente.

> Em produção no Railway, use a importação via Admin UI (seção 1) ou via `kcadm.sh` em um job de inicialização, pois o serviço Railway não monta volumes locais.

---

## 7. Troubleshooting

| Sintoma | O que verificar |
|---------|-----------------|
| Login redireciona mas volta com erro | `Valid redirect URIs` no client `connect-cli` não inclui a URL do `connect-web` |
| Erro de CORS no browser | `Web origins` no client não inclui a origem exata (sem barra final) |
| Token sem claim `roles` | Client scope `luxus-roles` não está em `Default client scopes` do `connect-cli` |
| Token sem claim `organization` | Client scope `organization` não está em `Default client scopes` do `connect-cli` |
| API retorna 401 | `KEYCLOAK_AUTH_SERVER_URL` ou `KEYCLOAK_REALM` incorretos na `connect-api` |
| Web em branco após login | `VITE_AUTH_URL` ou `VITE_API_URL` não marcados como build variables; redeploy necessário |
| Realm não aparece após importação | JSON inválido ou realm com mesmo nome já existe; delete o realm existente e reimporte |
