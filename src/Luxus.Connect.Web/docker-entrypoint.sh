#!/bin/sh
set -e
export PORT="${PORT:-80}"
export AUTH_PROXY_PASS="${AUTH_PROXY_PASS:-http://keycloak:8080/auth/}"
export API_PROXY_PASS="${API_PROXY_PASS:-http://connect-api:80/}"

DEST=/etc/nginx/conf.d/default.conf
if [ -w "$DEST" ] || [ ! -f "$DEST" ]; then
  envsubst '${PORT} ${AUTH_PROXY_PASS} ${API_PROXY_PASS}' \
    < /etc/nginx/conf.d/default.conf.template \
    > "$DEST"
fi

exec "$@"
