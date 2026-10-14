#!/bin/bash

set -euo pipefail

# Perfil "dev" inclui connect-web-dev (Vite).
# Na VPS Hostinger: ./docker-up.sh prod
# O nginx do Ubuntu fica na 80/443; os contentores usam 3005 / 8002 / 8081.
if [ "${1:-}" = "prod" ]; then
  unset COMPOSE_PROFILES
else
  export COMPOSE_PROFILES=dev
fi

if [ "${1:-}" = "prod" ]; then
  COMPOSE_FILES=(-f "docker-compose.yml" -f "docker-compose.prod.yml")
else
  if [ -n "${1:-}" ] && [ -f "docker-compose.$1.yml" ]; then
    COMPOSE_FILES=(-f "docker-compose.yml" -f "docker-compose.override.yml" -f "docker-compose.$1.yml")
  else
    COMPOSE_FILES=(-f "docker-compose.yml" -f "docker-compose.override.yml")
  fi
fi

# Keep the current stack available while images are built. Compose recreates only
# the services whose configuration/image changed, avoiding an unnecessary 502
# window caused by taking the whole stack down first.
if [ "${1:-}" = "prod" ]; then
  # Validate without printing environment values or credentials.
  images=$(docker compose "${COMPOSE_FILES[@]}" config --format json |
    python3 -c 'import json,sys; s=json.load(sys.stdin)["services"]; print("\n".join(s[n]["image"] for n in ("postgres", "rabbitmq", "minio", "minio-init", "keycloak")))')
  while IFS= read -r image; do
    if ! [[ "$image" =~ @sha256:[a-f0-9]{64}$ ]]; then
      echo "Produção exige imagens Postgres, RabbitMQ, MinIO, mc e Keycloak fixadas por digest SHA-256." >&2
      exit 1
    fi
  done <<< "$images"
fi
docker compose "${COMPOSE_FILES[@]}" up -d --build --remove-orphans --wait --wait-timeout 180
