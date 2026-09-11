# Deploy em VPS (Docker Compose — produção)

Guia consolidado para publicar **Luxus Connect** na VPS Hostinger com
`docker-compose.yml` + `docker-compose.prod.yml`. O domínio público é
**`telefonia.redobrai.online`**. O nginx do Ubuntu escuta 80/443 e encaminha
para frontend `:3005`, API `:8002` e Keycloak `:8081`; publicamente, Keycloak
fica em **`/auth`** e a API em **`/api`**.

---

## Pré-requisitos na VPS

- **Docker** e **Docker Compose** (plugin v2).
- **Git** (ou outro meio de levar o código).
- Portas **80** e **443** abertas no firewall (e **22** para SSH).
- **DNS**: `telefonia.redobrai.online` deve resolver para o IP público desta máquina.

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

## Certificados TLS (Let’s Encrypt + Certbot)

O nginx do Ubuntu usa diretamente:

`/etc/letsencrypt/live/telefonia.redobrai.online/`

O nginx do host usa diretamente os certificados do Certbot nesse caminho.

### Emitir certificado (modo `standalone`)

1. **Liberta a porta 80** — nada pode estar à escuta no 80 (para o Certbot levantar o servidor temporário). Para a stack Docker:

   ```bash
   docker compose -f docker-compose.yml -f docker-compose.prod.yml down
   ```

2. Garante que o **DNS** (`A` para o IP da VPS) está correto.

3. **IPv6 (AAAA):** se existir registo **AAAA** a apontar para **outro** servidor, o Let’s Encrypt pode validar por IPv6 e receber **404**. Corrige o `AAAA` para esta VPS, remove-o se só usas IPv4, ou usa desafio **DNS-01**.

4. Emite o certificado (ajusta o domínio e inclui `www` só se tiveres DNS para isso):

   ```bash
   sudo certbot certonly --standalone -d telefonia.redobrai.online
   ```

5. Copia (ou cria **symlinks**) para a pasta que o Docker monta:

   ```bash
   sudo nginx -t && sudo systemctl reload nginx
   ```

6. Volta a subir a stack (secção seguinte).

### Renovação

Teste: `sudo certbot renew --dry-run`. Após renovação real, execute
`sudo nginx -t && sudo systemctl reload nginx`.

---

## Subir a stack em produção

Na **raiz do repositório**:

```bash
chmod +x docker-up.sh
./docker-up.sh prod
```

- Usa **apenas** `docker-compose.yml` + `docker-compose.prod.yml`.
- **Não** ativa o perfil `dev`: o serviço **`connect-web-dev`** (Vite) **não** sobe em produção.

Para desenvolvimento local, o script usa `COMPOSE_PROFILES=dev` (exceto com alvo
`prod`) e o `docker-compose.override.yml` define as portas locais.

---

## Migrações (PostgreSQL)

Com o Postgres acessível (container em execução ou connection string correta):

```bash
./scripts/apply-migrations.sh
```

Ajusta ambiente/connection string conforme o README principal do repositório se o script assumir `localhost`.

---

## Keycloak (primeira vez / após mudar domínio)

- Admin: URL base com path `/auth`: `https://telefonia.redobrai.online/auth`.
- Realm **`luxus`**, client público **`connect-cli`**.
- **Valid redirect URIs** / **Web origins** devem incluir `https://telefonia.redobrai.online/*`.
- O access token deve incluir o claim **`organization`** (estrutura esperada pela SPA). Sem isso, o login no Keycloak até funciona, mas a app não marca sessão e não redireciona para a home. Configura um **protocol mapper** (ou similar) no client `connect-cli` para emitir esse claim.

Se o volume do Keycloak foi criado **antes** de configurar o path `/auth`, pode ser necessário rever o realm ou o volume na primeira subida com o novo esquema.

---

## Domínio e ficheiros a manter coerentes

Se alterares o FQDN, atualiza de forma consistente:

- `docker-compose.prod.yml` — `KC_HOSTNAME`, `CORS_ORIGINS`, args de build `VITE_API_URL` / `VITE_AUTH_URL`
- `docker/nginx/telefonia.redobrai.online.conf` — `server_name`, certificados e upstreams
- Certificados em `/etc/letsencrypt/live/<hostname>/`
- Rebuild obrigatório do **`connect-web`** após mudar `VITE_*`

---

## Verificação rápida

- Front: `https://telefonia.redobrai.online`
- API: rotas sob `https://telefonia.redobrai.online/api/v1/...`
- Logs: `docker compose -f docker-compose.yml -f docker-compose.prod.yml logs connect-api --tail=100`

---

## Arquitetura resumida (produção)

- **nginx do Ubuntu** expõe **80** (redirect para HTTPS) e **443** (TLS); faz proxy de `/auth/` e `/realms/` para Keycloak, de **`/api/`** para a API e serve o frontend em `:3005`.
- **connect-api** é publicado em `127.0.0.1:8002`; o nginx remove o prefixo `/api` ao encaminhar.
- **Keycloak** é publicado em `127.0.0.1:8081` e fica atrás do nginx em `/auth`.

---

## Problemas frequentes

| Sintoma                         | O que verificar                                                                                                    |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Certbot **404** no desafio ACME | Porta **80** ocupada; registo **AAAA** a apontar para outro sítio; domínio ainda a apontar para hospedagem antiga. |
| nginx não arranca               | Execute `sudo nginx -t`; confirme certificados e upstreams `:3005`, `:8002`, `:8081`.                              |
| CORS / login                    | `CORS_ORIGINS`, URLs `VITE_*` no build do web, realm Keycloak e redirects.                                         |

Para detalhes funcionais, ver [Documento de Especificação Funcional-v2.md](./Documento%20de%20Especificação%20Funcional-v2.md).
