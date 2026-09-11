#!/bin/bash
set -euo pipefail

if [ "${1:-}" = "prod" ]; then
  docker compose -f docker-compose.yml -f docker-compose.prod.yml down --remove-orphans
elif [ -n "${1:-}" ] && [ -f "docker-compose.$1.yml" ]; then
  docker compose -f docker-compose.yml -f docker-compose.override.yml -f "docker-compose.$1.yml" down --remove-orphans
else
  docker compose -f docker-compose.yml -f docker-compose.override.yml down --remove-orphans
fi
