#!/usr/bin/env bash
# Passwords are read silently and piped over stdin, never passed as arguments.
set -euo pipefail
cd "$(dirname "$0")/.."
python3 scripts/configure-local-auth.py
case "${1:-}" in
  init-admin)
    [ "$#" -eq 3 ] || { echo "Uso: $0 init-admin USUARIO EMAIL" >&2; exit 1; }
    ;;
  reset-password|reset-mfa)
    [ "$#" -eq 2 ] || { echo "Uso: $0 $1 USUARIO" >&2; exit 1; }
    ;;
  *) echo "Comandos: init-admin, reset-password, reset-mfa" >&2; exit 1 ;;
esac
compose=(docker compose -f docker-compose.yml -f docker-compose.prod.yml)
if [ "$1" = reset-mfa ]; then
  "${compose[@]}" run --rm --no-deps -T --pull never connect-api auth "$@"
else
  read -r -s -p "Nova senha (mínimo 12 caracteres): " auth_password
  printf '\n'
  read -r -s -p "Confirme a senha: " auth_confirmation
  printf '\n'
  [ "$auth_password" = "$auth_confirmation" ] || { echo "As senhas não coincidem." >&2; exit 1; }
  printf '%s' "$auth_password" | "${compose[@]}" run --rm --no-deps -T --pull never connect-api auth "$@"
  unset auth_password auth_confirmation
fi
