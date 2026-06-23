# Migrações de schema PostgreSQL

O schema da aplicação foi originalmente gerido com **Entity Framework Core** (C#).
As migrações históricas estão em `ef/` para referência e para bases novas.

## Aplicar schema em base vazia

Com PostgreSQL acessível e variáveis `POSTGRES_USER` / `POSTGRES_PASSWORD` no `.env`:

1. Crie a base `luxus_connect_dev` (o `docker/postgres/init.sql` já cria bases no primeiro boot).
2. Aplique o script consolidado (recomendado, sem .NET SDK):

   ```bash
   docker exec -i postgres.connect.luxus psql -U postgres -d luxus_connect_dev < db/migrations/001_initial_schema.sql
   ```

   Alternativa: execute as migrações EF na ordem dos ficheiros `ef/2026*.cs` (Up), ou restaure um dump existente.

A API Go **não** executa migrações no startup — assume schema já aplicado.

## Evolução futura

Para novas alterações de schema, adicione ficheiros SQL versionados nesta pasta ou adopte uma ferramenta como `golang-migrate`.
