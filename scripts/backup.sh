#!/usr/bin/env bash
set -euo pipefail

# Backup do PostgreSQL e dos volumes de dados (MinIO e anexos legados de Parceiros).
#
# Uso:
#   ./scripts/backup.sh [destino]
#
# Variáveis (com valores por omissão):
#   MINIO_VOLUME           volume do MinIO      (padrão: connect.luxus.minio-data)
#   PARCEIROS_UPLOAD_VOLUME volume legado       (padrão: connect.luxus.parceiros-uploads)
#   PG_CONTAINER           contentor Postgres   (padrão: postgres.connect.luxus)
#   POSTGRES_USER          utilizador           (padrão: postgres)
#   POSTGRES_DB            base de dados        (padrão: luxus_connect_dev)
#
# Os nomes vêm das chaves `name:` do docker-compose.prod.yml. Não são derivados
# de COMPOSE_PROJECT_NAME: o projeto chama-se luxus-connect, mas os volumes
# chamam-se connect.luxus.*.
#
# O dump é lido do contentor em execução (sem parar o serviço). Para uma cópia
# 100% consistente dos volumes, pare os serviços antes de correr este script.

DEST="${1:-./backups}"
STAMP="$(date +%Y%m%d-%H%M%S)"
mkdir -p "$DEST/$STAMP"
OUT="$(cd "$DEST/$STAMP" && pwd)"

MINIO_VOLUME="${MINIO_VOLUME:-connect.luxus.minio-data}"
PARCEIROS_UPLOAD_VOLUME="${PARCEIROS_UPLOAD_VOLUME:-connect.luxus.parceiros-uploads}"
PG_CONTAINER="${PG_CONTAINER:-postgres.connect.luxus}"
DB_USER="${POSTGRES_USER:-postgres}"
DB_NAME="${POSTGRES_DB:-luxus_connect_dev}"

echo "→ pg_dump $DB_NAME ($PG_CONTAINER)"
docker exec "$PG_CONTAINER" pg_dump -U "$DB_USER" -d "$DB_NAME" --format=custom > "$OUT/db.dump"
if [ ! -s "$OUT/db.dump" ]; then
  echo "ERRO: dump do PostgreSQL vazio em $OUT/db.dump" >&2
  exit 1
fi

# Copia cada volume através de um contentor auxiliar (evita problemas de permissões do host).
# Um volume obrigatório ausente interrompe o backup em vez de terminar com sucesso.
backup_volume() {
  local volume="$1" name="$2" mode="${3:-required}"
  if ! docker volume inspect "$volume" >/dev/null 2>&1; then
    if [ "$mode" = "required" ]; then
      echo "ERRO: volume obrigatório $volume não encontrado" >&2
      return 1
    fi
    echo "→ volume legado $volume não existe, ignorado"
    return 0
  fi
  echo "→ volume $volume"
  docker run --rm \
    -v "${volume}:/data:ro" \
    -v "$OUT:/backup" \
    alpine sh -c "cd /data && tar czf /backup/${name}.tar.gz ."
  if [ ! -s "$OUT/${name}.tar.gz" ]; then
    echo "ERRO: arquivo $OUT/${name}.tar.gz não foi criado" >&2
    return 1
  fi
}

backup_volume "$MINIO_VOLUME" "minio-data" required
# Anexos gravados antes da unificação no MinIO. Se o volume não existir, os
# anexos novos já estão no MinIO; o backup avisa mas não falha.
backup_volume "$PARCEIROS_UPLOAD_VOLUME" "parceiros-uploads" optional

echo "Backup concluído em $OUT"
