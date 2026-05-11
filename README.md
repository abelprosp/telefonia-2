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

O backend segue **Clean Architecture** com **DDD**, **CQRS** e integração assíncrona onde faz sentido.

| Camada             | Projetos (principais)                                  | Função                                                       |
| ------------------ | ------------------------------------------------------ | ------------------------------------------------------------ |
| **API**            | `Luxus.Connect.Api`                                    | Controllers, OpenAPI, consumidores RabbitMQ                  |
| **Application**    | `Luxus.Connect.Application`, `Luxus.Connect.Contracts` | Comandos, validações, orquestração, contratos                |
| **Domain**         | `Luxus.Connect.Domain`                                 | Agregados, entidades, interfaces de repositório              |
| **Infraestrutura** | `Luxus.Connect.Infra.*`                                | EF Core (PostgreSQL), queries de leitura, HTTP, IoC, storage |

### Frontend (SPA)

- **Em produção:** [`src/Luxus.Connect.Web`](src/Luxus.Connect.Web) — Vite, React, TypeScript, TanStack Router, React Query, Tailwind, shadcn/ui. Detalhes: [`src/Luxus.Connect.Web/README.md`](src/Luxus.Connect.Web/README.md).
- **Referência legada:** [`src/old`](src/old) — Next.js anterior; **não** é destino de novas funcionalidades, apenas consulta de UX/fluxos.

### Fluxo de dados (resumo)

- **Escrita:** Controller → MediatR → _CommandHandler_ → EF Core (PostgreSQL) → publicação de eventos (MassTransit / RabbitMQ) quando aplicável.
- **Leitura:** Controller → repositórios de query em `Luxus.Connect.Infra.Data.Query` (EF Core / projeções) → modelos em Contracts.

**Padrões:** repositório, unit of work, mediator (MediatR), validação com FluentValidation, autenticação/autorização com pacotes **Keycloak.AuthServices**.

**Solução .NET:** ficheiro na raiz [`Luxus.Connect.slnx`](Luxus.Connect.slnx).

## Stack tecnológica

| Área                | Tecnologias                                                                                                               |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **Backend**         | ASP.NET Core 10, C#, EF Core, PostgreSQL, MediatR, FluentValidation, MassTransit + RabbitMQ                               |
| **Auth**            | Keycloak (OIDC), `Keycloak.AuthServices.Authentication` / `Authorization`                                                 |
| **Frontend**        | Vite, React, TypeScript, TanStack Router, TanStack Query, Tailwind, shadcn/ui, Zod (`VITE_*` em build)                    |
| **Object storage**  | API compatível com S3 (URLs pré-assinadas, importação de faturas)                                                         |
| **Observabilidade** | Serilog → Seq                                                                                                             |
| **Containerização** | Docker, Docker Compose; em produção na VPS: **nginx** no serviço `connect-web` (TLS, reverse proxy para `/auth` e `/api`) |

## Início rápido (desenvolvimento)

### Pré-requisitos

- [.NET 10 SDK](https://dotnet.microsoft.com/download)
- [Docker + Docker Compose](https://www.docker.com/products/docker-desktop/) (plugin v2)
- [Git](https://git-scm.com/downloads)
- [Node.js 22](https://nodejs.org/) (para o frontend local; o Docker do web usa a mesma major)
- PowerShell (Windows) ou Bash para scripts

### 1. Clonar e restaurar

```bash
git clone <url-do-repositório>
cd <pasta-do-clone>
dotnet restore
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

**Seq:** gerar hash em [Seq password hashing](https://blog.datalust.co/setting-an-initial-password-when-deploying-seq-to-docker/).

### 3. Pacotes privados (Azure Artifacts / `Goal.*`)

- **Docker (`docker compose ... --build`)**: o build injeta **`docker/build-secrets/nuget.config`** como segredo BuildKit — pode ser (1) substituído pelo `nuget.config` completo da equipa, ou (2) usar **`docker/build-secrets/nuget.credentials.config`** (no `.gitignore`) + **`NUGET_CONFIG_FILE`** no `.env`, ou (3) **`NUGET_FEED_URL`** + **`NUGET_PAT`** no `.env`.
- **SDK local (Visual Studio / `dotnet restore`)**: mantenha `nuget.config` na raiz (geralmente no `.gitignore` — copie da equipa). Detalhes de deploy: [`docs/DEPLOY_VPS.md`](docs/DEPLOY_VPS.md).

### 4. Certificados HTTPS (Windows, API local)

```powershell
mkdir $Env:USERPROFILE/.aspnet/https -Force
dotnet dev-certs https -ep $Env:USERPROFILE/.aspnet/https/Development.pfx -p c1bc6816-f70f-42e3-a71f-4ab75a294755
dotnet dev-certs https --trust
dotnet dev-certs https --check
```

### 5. Subir dependências com Docker

```bash
chmod +x docker-up.sh
./docker-up.sh
```

- Em desenvolvimento, o script define `COMPOSE_PROFILES=dev` e usa `docker-compose.yml` + `docker-compose.override.yml` (inclui `connect-web-dev` com Vite em `http://localhost:5173` quando o serviço está ativo).
- Se existir `docker-compose.<nome>.yml`, pode passar `./docker-up.sh <nome>` para acrescentar esse ficheiro (ex.: `docker-compose.vscode.yml`).

### 6. Migrações PostgreSQL

Com o Postgres acessível:

```bash
./ef-database-update.sh
```

Scripts adicionais: `./ef-migrations-add.sh`, `./ef-migrations-remove.sh`, `./ef-migrations-script.sh` (ver comentários nos scripts).

### Parar serviços

```bash
./docker-down.sh
```

## Serviços Docker (desenvolvimento)

| Serviço             | Portas (host)             | Função                                                    |
| ------------------- | ------------------------- | --------------------------------------------------------- |
| **connect-api**     | 4432 (HTTPS), 8002 (HTTP) | API principal                                             |
| **connect-web**     | 3000 → 80 (override)      | SPA estática ou base nginx                                |
| **connect-web-dev** | 5173 (perfil `dev`)       | Vite HMR                                                  |
| **postgres**        | 5432                      | Dados transacionais (`luxus_connect_dev`, `luxus_kc_dev`) |
| **rabbitmq**        | 5672, 15672 (UI)          | Mensagens                                                 |
| **seq**             | 81 (UI), 5341 (ingestão)  | Logs                                                      |
| **keycloak**        | 8081, 8443                | SSO / admin                                               |

## URLs úteis (local)

| Recurso             | URL                                                                   |
| ------------------- | --------------------------------------------------------------------- |
| API (container)     | `https://localhost:4432`                                              |
| OpenAPI             | `https://localhost:4432/api-docs`                                     |
| Seq                 | `http://localhost:81`                                                 |
| RabbitMQ Management | `http://localhost:15672` (credenciais do `.env`)                      |
| Keycloak Admin      | `http://localhost:8081` — utilizador `admin`, password `KC_ADMIN_PWD` |
| Vite (perfil dev)   | `http://localhost:5173`                                               |

**Importação de faturas (async):** upload para storage S3-compatível (URL pré-assinada), depois `POST /v1/invoice-imports` com `storage_bucket` e `storage_object_key`; estado em `GET /v1/invoice-imports/{id}`.

## Configuração Keycloak

Os valores **realm**, **client** e **URL do servidor** têm de estar alinhados entre **API**, **frontend** (`VITE_*`) e **realm no Keycloak**.

### API via Docker Compose (`docker-compose.override.yml`)

Variáveis injetadas no contentor `connect-api` (exemplo):

- `ASPNETCORE_Keycloak__Realm=luxus`
- `ASPNETCORE_Keycloak__AuthServerUrl=http://host.docker.internal:8081`
- `ASPNETCORE_Keycloak__Resource=connect-cli`
- `ASPNETCORE_Keycloak__Scopes=email profile connect-cli-scope organization`
- `ASPNETCORE_Keycloak__SslRequired=none`

### API via perfil VS Code (`src/Luxus.Connect.Api/Properties/launchSettings.json`)

O perfil **VSCode** usa outro conjunto (ex.: realm `luxus.connect`, client `luxus.connect-ui`, scopes `luxus.connect-ui-client-scope`). Ajuste o realm no Keycloak **ou** alinhe estas variáveis ao que importou no ambiente.

### Frontend (`src/Luxus.Connect.Web`)

Variáveis obrigatórias: `VITE_API_URL`, `VITE_AUTH_URL`, `VITE_CLIENT_ID`, `VITE_CLIENT_SECRET`, `VITE_STORAGE_BUCKET_NAME` (ver [`src/Luxus.Connect.Web/src/env.ts`](src/Luxus.Connect.Web/src/env.ts)). O serviço `connect-web-dev` no Compose define exemplos para apontar para a API e Keycloak no host.

### Produção (VPS)

Keycloak fica atrás do nginx em **`/auth`**; a API em **`/api`**. Variáveis típicas estão em [`docker-compose.prod.yml`](docker-compose.prod.yml) (ex.: `ASPNETCORE_Keycloak__AuthServerUrl=https://<domínio>/auth`). O client **`connect-cli`** deve ter **secret** igual a `KC_CLIENT_SECRET` no `.env`; **Redirect URIs** e **Web origins** devem incluir a origem HTTPS do front. O access token deve expor o claim **`organization`** (mapper no client), tal como descrito em [`docs/DEPLOY_VPS.md`](docs/DEPLOY_VPS.md).

**Temas Keycloak:** volume `./resources/theme/` → `/opt/keycloak/themes/` (ver imagem no `docker-compose.yml`).

## Deploy em VPS

Guia passo a passo (TLS Let’s Encrypt, `.env`, `nuget.config`, migrações, Keycloak, nginx): **[`docs/DEPLOY_VPS.md`](docs/DEPLOY_VPS.md)**.

Resumo:

1. Copiar [`docker/env.deploy.example`](docker/env.deploy.example) para `.env` e preencher (incl. `KC_CLIENT_SECRET` e object storage).
2. Colocar `fullchain.pem` e `privkey.pem` em `docker/ssl/<FQDN>/` (coerente com `docker-compose.prod.yml` e [`docker/nginx/connect-web.vps.conf`](docker/nginx/connect-web.vps.conf)).
3. Garantir `nuget.config` na raiz se o build da API precisar do feed privado.
4. Na raiz: `./docker-up.sh prod` (usa apenas `docker-compose.yml` + `docker-compose.prod.yml`, **sem** perfil `dev`).
5. Executar `./ef-database-update.sh` com Postgres acessível.

Se alterar o domínio, atualize em conjunto: `docker-compose.prod.yml`, `connect-web.vps.conf`, pasta `docker/ssl/...` e rebuild do `connect-web` (args `VITE_*` são embutidos no build).

## Contribuição e licença

- [CONTRIBUTING](CONTRIBUTING.md)
- Licença MIT — ver [LICENSE](LICENSE)

## Contato

- Anderson Ritter de Souza — [@ritter.ander](https://www.instagram.com/ritter.ander) — anderdsouza@gmail.com
#   t e l e f o n i a 2  
 