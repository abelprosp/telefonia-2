# Migrações de schema PostgreSQL

O schema da aplicação foi originalmente gerido com **Entity Framework Core** (C#).
As migrações históricas estão em `ef/` para referência e para bases novas.

## Aplicar schema em base vazia

Com PostgreSQL acessível e variáveis `POSTGRES_USER` / `POSTGRES_PASSWORD` no `.env`:

1. Suba o Compose: `docker/postgres/init.sh` cria as bases e aplica as migrações no primeiro boot.
2. Aplique o script consolidado (recomendado, sem .NET SDK):

   ```bash
   docker exec -i postgres.connect.luxus psql -U postgres -d luxus_connect_dev < db/migrations/001_initial_schema.sql
   ```

   Alternativa: execute as migrações EF na ordem dos ficheiros `ef/2026*.cs` (Up), ou restaure um dump existente.

Em produção, a API Go reaplica as migrações incrementais 021–025 e um
bootstrap idempotente no startup. Isso evita que uma base existente fique sem
colunas/tabelas novas durante um deploy. A migração inicial 001 continua sendo
obrigatória para uma base vazia e deve ser aplicada antes de subir a API.

## Evolução futura

Para novas alterações de schema, adicione ficheiros SQL versionados nesta pasta
e copie o novo ficheiro para `api/internal/dbmigrate/sql/`, ou adopte uma
ferramenta como `golang-migrate`.
