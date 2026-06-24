#!/usr/bin/env sh
# Aplica migrações SQL em ordem (requer psql e DATABASE_URL).
set -e
ROOT="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL não definida" >&2
  exit 1
fi
for f in "$ROOT"/db/migrations/[0-9]*.sql; do
  [ -f "$f" ] || continue
  echo "Applying $f"
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
done
echo "Migrações concluídas."
