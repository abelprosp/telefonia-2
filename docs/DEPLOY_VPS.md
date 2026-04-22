# Deploy em VPS (Docker Compose — produção)

Guia consolidado para publicar **Luxus Connect** numa VPS com `docker-compose.yml` + `docker-compose.prod.yml`. O domínio público usado no repositório é **`luxus.your-domain-here.com.br`** (nginx, TLS; Keycloak em **`/auth`**; API em **`/api`** — URLs públicas tipo `/api/v1/...`, o Kestrel continua a servir `/v1/...` por dentro). Se mudares o domínio, tens de alinhar **todos** os sítios indicados na secção [Domínio e ficheiros a manter coerentes](#domínio-e-ficheiros-a-manter-coerentes).

---

## Pré-requisitos na VPS

- **Docker** e **Docker Compose** (plugin v2).
- **Git** (ou outro meio de levar o código).
- Portas **80** e **443** abertas no firewall (e **22** para SSH).
- **DNS**: o nome usado na app (ex.: `luxus.your-domain-here.com.br`) deve resolver para o **IP público** desta máquina.

---

## Variáveis de ambiente (`.env` na raiz do repositório)

Copia `docker/env.deploy.example` para `.env` na raiz e preenche. Variáveis usadas pelo compose de produção incluem:

| Área                           | Variáveis (exemplos)                                                                             |
| ------------------------------ | ------------------------------------------------------------------------------------------------ |
| PostgreSQL                     | `POSTGRES_USER`, `POSTGRES_PASSWORD`                                                             |
| Seq                            | `SEQ_PWD_HASH` (hash inicial do Seq, não a senha em texto)                                       |
| Keycloak                       | `KC_ADMIN_PWD`, `KC_DB_USERNAME`, `KC_DB_PASSWORD`                                               |
| RabbitMQ                       | `RMQ_USER`, `RMQ_PWD`                                                                            |
| API / Keycloak client          | `KC_CLIENT_SECRET` (igual ao secret do client `connect-cli` no realm)                            |
| Object storage (S3-compatible) | `OBJECT_STORAGE_SERVICE_URL`, `OBJECT_STORAGE_ACCESS_KEY_ID`, `OBJECT_STORAGE_SECRET_ACCESS_KEY` |

**Não commits** o `.env` nem segredos no Git.

---

## `nuget.config` (feed privado Azure DevOps)

O ficheiro **`nuget.config` na raiz do repositório** está no **`.gitignore`**: não acompanha `git clone`. Para builds na VPS (Dockerfile da API faz `COPY nuget.config`):

1. Copia o ficheiro da tua máquina local, por exemplo:

   ```bash
   scp nuget.config usuario@servidor:/caminho/do/luxus-connect/nuget.config
   ```

2. Confirma na VPS: `ls -la nuget.config` na raiz do clone.

3. Restringe permissões se quiseres: `chmod 600 nuget.config`.

---

## Certificados TLS (Let’s Encrypt + Certbot)

O container **connect-web** (nginx) monta TLS a partir de:

`docker/ssl/<nome-do-host>/`

No estado atual do repositório, `<nome-do-host>` corresponde ao FQDN em **`docker-compose.prod.yml`** e no **nginx** (ex.: pasta `docker/ssl/luxus.your-domain-here.com.br/` com `fullchain.pem` e `privkey.pem`).

### Emitir certificado (modo `standalone`)

1. **Liberta a porta 80** — nada pode estar à escuta no 80 (para o Certbot levantar o servidor temporário). Para a stack Docker:

   ```bash
   docker compose -f docker-compose.yml -f docker-compose.prod.yml down
   ```

2. Garante que o **DNS** (`A` para o IP da VPS) está correto.

3. **IPv6 (AAAA):** se existir registo **AAAA** a apontar para **outro** servidor, o Let’s Encrypt pode validar por IPv6 e receber **404**. Corrige o `AAAA` para esta VPS, remove-o se só usas IPv4, ou usa desafio **DNS-01**.

4. Emite o certificado (ajusta o domínio e inclui `www` só se tiveres DNS para isso):

   ```bash
   sudo certbot certonly --standalone -d luxus.your-domain-here.com.br
   ```

5. Copia (ou cria **symlinks**) para a pasta que o Docker monta:

   ```bash
   sudo mkdir -p docker/ssl/luxus.your-domain-here.com.br
   sudo cp /etc/letsencrypt/live/luxus.your-domain-here.com.br/fullchain.pem docker/ssl/luxus.your-domain-here.com.br/
   sudo cp /etc/letsencrypt/live/luxus.your-domain-here.com.br/privkey.pem docker/ssl/luxus.your-domain-here.com.br/
   sudo chown -R "$USER:$USER" docker/ssl/luxus.your-domain-here.com.br
   chmod 644 docker/ssl/luxus.your-domain-here.com.br/fullchain.pem
   chmod 600 docker/ssl/luxus.your-domain-here.com.br/privkey.pem
   ```

6. Volta a subir a stack (secção seguinte).

### Renovação

Teste: `sudo certbot renew --dry-run`. Após renovação real, **reinicia** o serviço `connect-web` (ou a stack) para o nginx voltar a carregar os certificados.

---

## Subir a stack em produção

Na **raiz do repositório**:

```bash
chmod +x docker-up.sh
./docker-up.sh prod
```

- Usa **apenas** `docker-compose.yml` + `docker-compose.prod.yml`.
- **Não** ativa o perfil `dev`: o serviço **`connect-web-dev`** (Vite) **não** sobe em produção.

Para desenvolvimento local, o script usa `COMPOSE_PROFILES=dev` (exceto com alvo `prod`) e, quando existir, `docker-compose.<nome>.yml` extra; o ficheiro **`docker-compose.override.yml`** define portas locais do `connect-web` (ex.: `3000:80`) e garante o `connect-web-dev` quando usas `docker compose up` com override.

---

## Migrações (PostgreSQL)

Com o Postgres acessível (container em execução ou connection string correta):

```bash
./ef-database-update.sh
```

Ajusta ambiente/connection string conforme o README principal do repositório se o script assumir `localhost`.

---

## Keycloak (primeira vez / após mudar domínio)

- Admin: URL base com path `/auth` (ex.: `https://luxus.your-domain-here.com.br/auth`), conforme `KC_HTTP_RELATIVE_PATH` no compose de produção.
- Realm **`luxus`**, client **`connect-cli`**, secret alinhado com `KC_CLIENT_SECRET` no `.env`.
- **Valid redirect URIs** / **Web origins** devem incluir a origem HTTPS do front (ex.: `https://luxus.your-domain-here.com.br/*`).
- O access token deve incluir o claim **`organization`** (estrutura esperada pela SPA). Sem isso, o login no Keycloak até funciona, mas a app não marca sessão e não redireciona para a home. Configura um **protocol mapper** (ou similar) no client `connect-cli` para emitir esse claim.

Se o volume do Keycloak foi criado **antes** de configurar o path `/auth`, pode ser necessário rever o realm ou o volume na primeira subida com o novo esquema.

---

## Domínio e ficheiros a manter coerentes

Se alterares o FQDN (ex.: de `luxus.your-domain-here.com.br` para outro), atualiza de forma consistente:

- `docker-compose.prod.yml` — `KC_HOSTNAME`, `ASPNETCORE_Keycloak__AuthServerUrl`, `ASPNETCORE_Cors__Origins`, args de build `VITE_API_URL` / `VITE_AUTH_URL`, volume `docker/ssl/...`
- `docker/nginx/connect-web.vps.conf` — `server_name`, redirecionamentos
- Pasta **`docker/ssl/<hostname>/`** — certificados e montagem no compose
- Rebuild obrigatório do **`connect-web`** após mudar `VITE_*`

---

## Verificação rápida

- Front: `https://luxus.your-domain-here.com.br`
- API (exemplo): rotas sob `https://luxus.your-domain-here.com.br/api/v1/...`
- Logs: `docker compose -f docker-compose.yml -f docker-compose.prod.yml logs connect-api --tail=100`

---

## Arquitetura resumida (produção)

- **nginx** no `connect-web` expõe **80** (redirect para HTTPS) e **443** (TLS); faz proxy de `/auth/` para Keycloak, de **`/api/`** para a API (remove o prefixo `/api` ao encaminhar), e serve o SPA no restante.
- **connect-api** escuta só na rede Docker (sem portas publicadas no host); confia em cabeçalhos `X-Forwarded-*` (middleware de forwarded headers na API).
- **Keycloak** não expõe porta no host em produção; fica atrás do nginx em `/auth`.

---

## Problemas frequentes

| Sintoma                         | O que verificar                                                                                                    |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Certbot **404** no desafio ACME | Porta **80** ocupada; registo **AAAA** a apontar para outro sítio; domínio ainda a apontar para hospedagem antiga. |
| nginx não arranca               | `fullchain.pem` / `privkey.pem` em falta ou caminho do volume errado.                                              |
| CORS / login                    | `ASPNETCORE_Cors__Origins`, URLs `VITE_*` no build do web, realm Keycloak e redirects.                             |
| Restore NuGet no build          | `nuget.config` na raiz da VPS com credenciais do feed.                                                             |

Para mais contexto de produto e roadmap, ver [PRODUTO_E_ROADMAP.md](./PRODUTO_E_ROADMAP.md).
