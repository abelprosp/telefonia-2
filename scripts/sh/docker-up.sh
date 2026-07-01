#!/bin/bash

set -euo pipefail

# Perfil "dev" inclui connect-web-dev (Vite). Em produção use o alvo "prod" sem este perfil.
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

docker compose "${COMPOSE_FILES[@]}" down --remove-orphans
docker compose "${COMPOSE_FILES[@]}" up -d --build --force-recreate --remove-orphans
