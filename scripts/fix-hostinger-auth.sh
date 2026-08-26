#!/usr/bin/env bash
# Repara autenticação OIDC na VPS Hostinger (telefonia.redobrai.online).
# Uso, na raiz do repo: sudo bash scripts/fix-hostinger-auth.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "=== 1. Quem escuta 80/443/3005/8081/8002 ==="
ss -tlnp | grep -E ':80 |:443 |:3005 |:8081 |:8002 ' || true
echo

echo "=== 2. Containers Luxus ==="
docker ps -a --filter name=connect.luxus --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' || true
echo

echo "=== 3. Remover web/keycloak órfãos (nome fixo) ==="
docker rm -f web.connect.luxus keycloak.connect.luxus 2>/dev/null || true
docker compose down --remove-orphans || true
docker rm -f web.connect.luxus keycloak.connect.luxus 2>/dev/null || true

echo "=== 4. Subir stack (override: web :3005, keycloak :8081, api :8002) ==="
docker compose build connect-web --no-cache
docker compose up -d --force-recreate --remove-orphans keycloak connect-web connect-api

echo "=== 5. Nginx do host (HTTPS → 3005 / 8081 / 8002) ==="
if [ -d /etc/nginx/sites-available ] || [ -d /etc/nginx/conf.d ]; then
  bash "$ROOT/scripts/setup-vps-nginx.sh"
else
  echo "Nginx do host não encontrado. Se a Hostinger usa Apache na 443, o /auth não chega ao Keycloak."
  echo "Instale nginx ou encaminhe /auth/ para 127.0.0.1:8081/auth/"
fi

echo "=== 6. Esperar Keycloak ==="
ok=0
for i in $(seq 1 40); do
  if curl -sf "http://127.0.0.1:8081/auth/realms/luxus/.well-known/openid-configuration" \
      | grep -q '"issuer"'; then
    ok=1
    break
  fi
  echo "  tentativa $i/40..."
  sleep 3
done

echo
echo "=== 7. Testes locais (não passam pela internet) ==="
echo "-- Keycloak direto :8081/auth"
curl -sI "http://127.0.0.1:8081/auth/realms/luxus/.well-known/openid-configuration" | head -n 8 || true
echo
echo "-- via Nginx HTTPS"
curl -skI "https://127.0.0.1/auth/realms/luxus/.well-known/openid-configuration" \
  --resolve telefonia.redobrai.online:443:127.0.0.1 | head -n 8 || true
echo
echo "-- API health"
curl -sI "http://127.0.0.1:8002/health" | head -n 6 || true

if [ "$ok" -ne 1 ]; then
  echo
  echo "Keycloak ainda não responde em /auth. Logs:"
  docker logs keycloak.connect.luxus --tail 80 || true
  echo
  echo "Se o log não mostrar http-relative-path=/auth, recrie:"
  echo "  docker compose up -d --force-recreate keycloak"
  exit 1
fi

echo
echo "Pronto. No browser, o OIDC tem de ser application/json:"
echo "  curl -sI https://telefonia.redobrai.online/auth/realms/luxus/.well-known/openid-configuration"
echo "Faça hard refresh (Ctrl+Shift+R) no site."
