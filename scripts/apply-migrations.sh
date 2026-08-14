#!/usr/bin/env bash
set -e

DB_NAME="luxus_connect_dev"
CONTAINER_NAME="postgres.connect.luxus"

echo "Aplicando migrações no banco $DB_NAME..."

for f in $(ls db/migrations/[0-9]*.sql | sort); do
  echo "==> Executando $f"
  docker exec -i "$CONTAINER_NAME" psql -U postgres -d "$DB_NAME" < "$f"
done

echo "✅ Todas as migrações foram aplicadas com sucesso!"
