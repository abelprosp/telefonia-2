#!/usr/bin/env bash
# Atualiza o login próprio sem interromper banco ou storage.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
python3 scripts/configure-local-auth.py
bash docker-up.sh prod
bash scripts/setup-vps-nginx.sh
printf '\nLogin próprio atualizado. Consulte docs/AUTH_LOCAL.md para configurar ou recuperar contas.\n'
