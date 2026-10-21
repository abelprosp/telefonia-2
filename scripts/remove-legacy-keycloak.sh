#!/usr/bin/env bash
# Remove o contêiner e, opcionalmente, os volumes do Keycloak legado.
# O Telefonia atual não usa Keycloak; este script é apenas para limpeza da VPS.
set -euo pipefail

assume_yes=0
purge_volumes=0

usage() {
  cat <<'EOF'
Uso: bash scripts/remove-legacy-keycloak.sh [--yes] [--purge-volumes]

  --yes             não pede confirmação
  --purge-volumes   remove também os volumes cujo nome contém "keycloak"
  -h, --help        mostra esta ajuda

Sem --purge-volumes, os volumes são preservados para backup/validação.
EOF
}

for arg in "$@"; do
  case "$arg" in
    --yes|-y) assume_yes=1 ;;
    --purge-volumes) purge_volumes=1 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Opção desconhecida: $arg" >&2; usage >&2; exit 1 ;;
  esac
done

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker não encontrado. Execute este script na VPS com Docker Compose v2." >&2
  exit 1
fi

if [ -n "${KEYCLOAK_CONTAINER:-}" ]; then
  containers="$KEYCLOAK_CONTAINER"
else
  containers="$(docker ps -a --filter 'name=keycloak' --format '{{.Names}}' 2>/dev/null || true)"
fi
if [ -n "$containers" ]; then
  echo "Contêineres encontrados:"
  printf '  %s\n' $containers
fi

volumes="$(docker volume ls -q --filter 'name=keycloak' 2>/dev/null || true)"
if [ -n "$volumes" ]; then
  echo "Volumes encontrados:"
  printf '  %s\n' $volumes
fi

if [ -z "$containers" ] && [ -z "$volumes" ]; then
  echo "Nenhum contêiner ou volume Keycloak encontrado. Nada a fazer."
  exit 0
fi

if [ "$assume_yes" -ne 1 ]; then
  read -r -p "Remover o contêiner legado? [s/N] " answer
  case "$answer" in
    s|S|sim|SIM|y|Y) ;;
    *) echo "Cancelado. Nada foi removido."; exit 0 ;;
  esac
fi

if [ -n "$containers" ]; then
  # shellcheck disable=SC2086
  docker rm -f $containers
  echo "Contêineres removidos."
fi

if [ "$purge_volumes" -eq 1 ] && [ -n "$volumes" ]; then
  # shellcheck disable=SC2086
  docker volume rm $volumes
  echo "Volumes removidos."
elif [ -n "$volumes" ]; then
  echo "Volumes preservados. Use --purge-volumes após validar o backup."
fi

echo "Limpeza concluída."
