<!-- PROJECT LOGO -->
<br />
<p align="center">
  <img src="docs/images/logo.png" alt="Logo" width="80" height="80">
  <h3 align="center">Luxus.Connect</h3>
  <p align="center">
    Plataforma operacional e financeira entre operadoras (VIVO, TIM) e a Luxus — gestão de faturas origem, linhas, importações e cadastros alinhada à especificação funcional.
    <br />
    <a href="#início-rápido-desenvolvimento"><strong>Guia de instalação »</strong></a>
    ·
    <a href="docs/DEPLOY_VPS.md">Deploy em VPS</a>
  </p>
</p>

## Sobre o sistema

**Luxus.Connect** é o sistema que apoia a **Luxus Gestão**: um intermediário entre **operadoras** e **clientes Luxus**, com foco na entidade **linha** (número de celular). O produto organiza **faturas origem** (`ProviderInvoice`), **importação em lote** (armazenamento S3-compatível + processamento assíncrono via RabbitMQ), **contas na nuvem**, **ciclos de faturamento**, **clientes** e evolução da **composição financeira** por linha, conforme a especificação v2.

- **Especificação normativa:** [`docs/Documento de Especificação Funcional-v2.md`](docs/Documento%20de%20Especificação%20Funcional-v2.md)
- **Backlog de implementação:** [`docs/BACKLOG_IMPLEMENTACAO_SPEC_V2.md`](docs/BACKLOG_IMPLEMENTACAO_SPEC_V2.md)

### Identidade (Keycloak)

Utilizadores e organizações vivem no **Keycloak** (SSO). O PostgreSQL da aplicação **não** duplica tabelas de utilizadores/organizações; campos de auditoria e contexto multi-tenant usam identificadores do token (ex.: _subject_, claims) como texto, com regras na API.

## Arquitetura

O backend é uma **API REST em Go** (binário único, ~15 MB) com o mesmo contrato HTTP que o frontend já consome. O schema PostgreSQL existente é reutilizado.

| Camada | Localização | Função |
| ------ | ----------- | ------ |
| **API** | [`api/`](api/) | HTTP (chi), auth JWT Keycloak, handlers REST, consumer RabbitMQ |
| **Store** | `api/internal/store` | PostgreSQL via pgx (queries e comandos) |
| **Serviços** | `api/internal/services` | Regras de negócio (clientes, linhas, faturas, etc.) |
| **Importação** | `api/internal/importservice` + `api/internal/vivo` | Parser VIVO + pipeline assíncrono |
| **Frontend** | [`src/Luxus.Connect.Web`](src/Luxus.Connect.Web) | Vite, React, TypeScript, TanStack Router/Query |

### Fluxo de dados (resumo)

- **Escrita:** Handler → Service → PostgreSQL; eventos de importação via RabbitMQ.
- **Leitura:** Handler → Store (SQL) → DTOs JSON snake_case.
- **Importação de faturas:** upload S3/R2 → `POST /v1/provider-invoices` → fila `luxus-connect-events` → parser VIVO.

**Solução:** [`Luxus.Connect.slnx`](Luxus.Connect.slnx) (frontend + Docker). Backend Go em [`api/go.mod`](api/go.mod).

## Stack tecnológica

| Área | Tecnologias |
| ---- | ----------- |
| **Backend** | Go 1.22, chi, pgx, RabbitMQ (amqp), AWS SDK v2 (S3/R2) |
| **Auth** | Keycloak (OIDC), validação JWT + JWKS |
| **Frontend** | Vite, React, TypeScript, TanStack Router, TanStack Query, Tailwind, shadcn/ui |
| **Object storage** | API compatível com S3 (URLs pré-assinadas, importação de faturas) |
| **Containerização** | Docker, Docker Compose; produção: nginx no `connect-web` (TLS, `/auth`, `/api`) |

## Início rápido (desenvolvimento)

### Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)
- [Docker + Docker Compose](https://www.docker.com/products/docker-desktop/) (plugin v2)
- [Git](https://git-scm.com/downloads)
- [Node.js 22](https://nodejs.org/) (para o frontend local)
- PowerShell (Windows) ou Bash para scripts

### 1. Clonar

```bash
git clone <url-do-repositório>
cd <pasta-do-clone>
```

### 2. Ficheiro `.env` na raiz

Crie `.env` com variáveis usadas pelo Compose (ajuste segredos):

```env
# PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

# Seq (hash da password inicial — não é texto plano)
SEQ_PWD_HASH=YourPasswordHashHere

# Keycloak
KC_ADMIN_PWD=admin
KC_DB_USERNAME=postgres
KC_DB_PASSWORD=postgres

# RabbitMQ
RMQ_USER=guest
RMQ_PWD=guest

# Object storage (S3-compatible) — necessário para importação de faturas
OBJECT_STORAGE_SERVICE_URL=
OBJECT_STORAGE_ACCESS_KEY_ID=
OBJECT_STORAGE_SECRET_ACCESS_KEY=
```

**Seq:** gerar hash em [Seq password hashing](https://blog.datalust.co/setting-an-initial-password-when-deploying-seq-to-docker/) (opcional; o serviço Seq permanece no Compose para logs legados).

### 3. Subir dependências com Docker

```bash
chmod +x docker-up.sh
./docker-up.sh
```

- Em desenvolvimento, o script define `COMPOSE_PROFILES=dev` e usa `docker-compose.yml` + `docker-compose.override.yml` (inclui `connect-web-dev` com Vite em `http://localhost:5173` quando o serviço está ativo).
- Se existir `docker-compose.<nome>.yml`, pode passar `./docker-up.sh <nome>` para acrescentar esse ficheiro (ex.: `docker-compose.vscode.yml`).

### 4. Schema PostgreSQL

O schema deve existir antes da API subir (bases criadas pelo `docker/postgres/init.sql` no primeiro boot). Migrações históricas EF estão em [`db/migrations/ef/`](db/migrations/ef/) — ver [`db/migrations/README.md`](db/migrations/README.md).

### API local (sem Docker)

```bash
cd api
export DATABASE_URL="Host=localhost;Username=postgres;Password=postgres;Database=luxus_connect_dev"
export KEYCLOAK_REALM=luxus
export KEYCLOAK_AUTH_SERVER_URL=http://localhost:8081
export KEYCLOAK_RESOURCE=connect-cli
go run ./cmd/api
```

### Parar serviços

```bash
./docker-down.sh
```

## Serviços Docker (desenvolvimento)

| Serviço             | Portas (host)             | Função                                                    |
| ------------------- | ------------------------- | --------------------------------------------------------- |
| **connect-api**     | 8002 (HTTP)               | API Go                                                    |
| **connect-web**     | 3000 → 80 (override)      | SPA estática ou base nginx                                |
| **connect-web-dev** | 5173 (perfil `dev`)       | Vite HMR                                                  |
| **postgres**        | 5432                      | Dados transacionais (`luxus_connect_dev`, `luxus_kc_dev`) |
| **rabbitmq**        | 5672, 15672 (UI)          | Mensagens                                                 |
| **seq**             | 81 (UI), 5341 (ingestão)  | Logs                                                      |
| **keycloak**        | 8081, 8443                | SSO / admin                                               |

## URLs úteis (local)

| Recurso             | URL                                                                   |
| ------------------- | --------------------------------------------------------------------- |
| API (container)     | `http://localhost:8002`                                               |
| Health              | `http://localhost:8002/health`                                        |
| Seq                 | `http://localhost:81`                                                 |
| RabbitMQ Management | `http://localhost:15672` (credenciais do `.env`)                      |
| Keycloak Admin      | `http://localhost:8081` — utilizador `admin`, password `KC_ADMIN_PWD` |
| Vite (perfil dev)   | `http://localhost:5173`                                               |

**Importação de faturas (async):** upload para storage S3-compatível (URL pré-assinada), depois `POST /v1/provider-invoices` com `storage_bucket` e `storage_object_key`; estado em `GET /v1/provider-invoices/{id}`.

## Configuração Keycloak

Os valores **realm**, **client** e **URL do servidor** têm de estar alinhados entre **API**, **frontend** (`VITE_*`) e **realm no Keycloak**.

### API via Docker Compose (`docker-compose.override.yml`)

Variáveis injetadas no contentor `connect-api`:

- `KEYCLOAK_REALM=luxus`
- `KEYCLOAK_AUTH_SERVER_URL=http://host.docker.internal:8081`
- `KEYCLOAK_RESOURCE=connect-cli`

### Frontend (`src/Luxus.Connect.Web`)

Variáveis obrigatórias: `VITE_API_URL`, `VITE_AUTH_URL`, `VITE_CLIENT_ID`, `VITE_CLIENT_SECRET`, `VITE_STORAGE_BUCKET_NAME` (ver [`src/Luxus.Connect.Web/src/env.ts`](src/Luxus.Connect.Web/src/env.ts)). O serviço `connect-web-dev` no Compose define exemplos para apontar para a API e Keycloak no host.

### Produção (VPS)

Keycloak fica atrás do nginx em **`/auth`**; a API em **`/api`**. Variáveis em [`docker-compose.prod.yml`](docker-compose.prod.yml) (ex.: `KEYCLOAK_AUTH_SERVER_URL=https://<domínio>/auth`).

**Temas Keycloak:** volume `./resources/theme/` → `/opt/keycloak/themes/` (ver imagem no `docker-compose.yml`).

## Deploy em VPS

Guia passo a passo (TLS Let’s Encrypt, `.env`, Keycloak, nginx): **[`docs/DEPLOY_VPS.md`](docs/DEPLOY_VPS.md)**.

Resumo:

1. Copiar [`docker/env.deploy.example`](docker/env.deploy.example) para `.env` e preencher (incl. `KC_CLIENT_SECRET` e object storage).
2. Colocar `fullchain.pem` e `privkey.pem` em `docker/ssl/<FQDN>/`.
3. Na raiz: `./docker-up.sh prod` (usa `docker-compose.yml` + `docker-compose.prod.yml`).
4. Garantir schema PostgreSQL aplicado (ver `db/migrations/`).

Se alterar o domínio, atualize em conjunto: `docker-compose.prod.yml`, `connect-web.vps.conf`, pasta `docker/ssl/...` e rebuild do `connect-web` (args `VITE_*` são embutidos no build).

## Contribuição e licença

- [CONTRIBUTING](CONTRIBUTING.md)
- Licença MIT — ver [LICENSE](LICENSE)

## Contato

- Anderson Ritter de Souza — [@ritter.ander](https://www.instagram.com/ritter.ander) — anderdsouza@gmail.com