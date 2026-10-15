#!/bin/sh
set -e
export PORT="${PORT:-80}"
export API_PROXY_PASS="${API_PROXY_PASS:-http://connect-api:80/}"

DEST=/etc/nginx/conf.d/default.conf
if [ -w "$DEST" ] || [ ! -f "$DEST" ]; then
  envsubst '${PORT} ${API_PROXY_PASS}' \
    < /etc/nginx/conf.d/default.conf.template \
    > "$DEST"
fi

exec "$@"
