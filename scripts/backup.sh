#!/usr/bin/env bash
set -euo pipefail

# Backup do PostgreSQL e dos volumes de dados (MinIO, Keycloak).
#
# Uso:
#   ./scripts/backup.sh [destino]
#
# Variáveis (com valores por omissão):
#   COMPOSE_PROJECT_NAME  prefixo dos volumes  (padrão: luxus-connect)
#   PG_CONTAINER          contentor Postgres   (padrão: postgres.connect.luxus)
#   POSTGRES_USER         utilizador           (padrão: postgres)
#   POSTGRES_DB           base de dados        (padrão: luxus_connect_dev)
#
# O dump é lido do contentor em execução (sem parar o serviço). Para uma cópia
# 100% consistente dos volumes, pare os serviços antes de correr este script.

DEST="${1:-./backups}"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT="$DEST/$STAMP"
mkdir -p "$OUT"

PROJECT="${COMPOSE_PROJECT_NAME:-luxus-connect}"
PG_CONTAINER="${PG_CONTAINER:-postgres.connect.luxus}"
DB_USER="${POSTGRES_USER:-postgres}"
DB_NAME="${POSTGRES_DB:-luxus_connect_dev}"

echo "→ pg_dump $DB_NAME ($PG_CONTAINER)"
docker exec "$PG_CONTAINER" pg_dump -U "$DB_USER" -d "$DB_NAME" --format=custom > "$OUT/db.dump"

# Copia cada volume através de um contentor auxiliar (evita problemas de permissões do host).
backup_volume() {
  local volume="$1" name="$2"
  if ! docker volume inspect "$volume" >/dev/null 2>&1; then
    echo "→ volume $volume não existe, ignorado"
    return 0
  fi
  echo "→ volume $volume"
  docker run --rm \
    -v "${volume}:/data:ro" \
    -v "$(pwd)/$OUT:/backup" \
    alpine sh -c "cd /data && tar czf /backup/${name}.tar.gz ."
}

backup_volume "${PROJECT}_minio-data" "minio-data"
backup_volume "${PROJECT}_keycloak-data" "keycloak-data"

echo "Backup concluído em $OUT"
