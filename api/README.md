# Luxus.Connect API (Go)

API REST do Luxus.Connect — substitui o backend ASP.NET Core mantendo o mesmo contrato HTTP (`/v1/*`, JSON snake_case).

## Executar

```bash
cd api
go run ./cmd/api
```

Variáveis obrigatórias: `DATABASE_URL`, `KEYCLOAK_REALM`, `KEYCLOAK_AUTH_SERVER_URL`, `KEYCLOAK_RESOURCE`.

Opcionais: `RABBITMQ_URL`, `OBJECT_STORAGE_*`, `CORS_ORIGINS`, `PORT` (default 8080), `ENVIRONMENT`.

## Estrutura

| Pacote | Função |
|--------|--------|
| `cmd/api` | Entry point |
| `internal/handlers` | Rotas HTTP |
| `internal/services` | Regras de negócio |
| `internal/store` | PostgreSQL (pgx) |
| `internal/auth` | JWT Keycloak |
| `internal/importservice` | Importação de faturas VIVO |
| `internal/vivo` | Parser de ficheiros VIVO (ISO-8859-1) |
| `internal/messaging` | RabbitMQ |
| `internal/storage` | S3/R2 presigned URLs |

## Docker

Build a partir da raiz do repositório via `docker compose build connect-api` (contexto `./api`).

## Testes

```bash
go test ./...
```
