#!/usr/bin/env bash
set -e

DB_NAME="${DB_NAME:-luxus_connect_dev}"
DB_USER="${POSTGRES_USER:-postgres}"
CONTAINER_NAME="postgres.connect.luxus"

echo "Aplicando migrações no banco $DB_NAME com o usuário $DB_USER..."
echo "(A API também aplica 021+ no arranque; este script cobre a base completa.)"

shopt -s nullglob
migrations=(db/migrations/[0-9]*.sql)
IFS=$'\n' migrations=($(printf '%s\n' "${migrations[@]}" | sort))
for f in "${migrations[@]}"; do
  echo "==> Executando $f"
  docker exec -i "$CONTAINER_NAME" psql -v ON_ERROR_STOP=1 -U "$DB_USER" -d "$DB_NAME" < "$f"
done

echo "✅ Todas as migrações foram aplicadas com sucesso!"
