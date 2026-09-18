#!/bin/sh
# Garante no Keycloak em produção o mapper organization_id + user profile.
# Uso (dentro do host com container Keycloak rodando):
#   bash docker/keycloak/ensure-org-id-mapper.sh
set -e

KC_CONTAINER="${KC_CONTAINER:-$(docker ps --format '{{.Names}}' | grep -i keycloak | head -1)}"
if [ -z "$KC_CONTAINER" ]; then
  echo "Keycloak container not found"
  exit 1
fi

ADMIN_USER="${KC_ADMIN_USERNAME:-admin}"
ADMIN_PASS="${KC_ADMIN_PWD:-admin}"

docker exec "$KC_CONTAINER" /opt/keycloak/bin/kcadm.sh config credentials \
  --server http://localhost:8080/auth --realm master --user "$ADMIN_USER" --password "$ADMIN_PASS" || \
docker exec "$KC_CONTAINER" /opt/keycloak/bin/kcadm.sh config credentials \
  --server http://localhost:8080 --realm master --user "$ADMIN_USER" --password "$ADMIN_PASS"

echo "Ensuring organization_id mapper on client scope 'organization'..."
SCOPE_ID=$(docker exec "$KC_CONTAINER" /opt/keycloak/bin/kcadm.sh get client-scopes -r luxus --fields id,name \
  | sed -n 's/.*"name"[[:space:]]*:[[:space:]]*"organization".*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
  | head -1)
if [ -z "$SCOPE_ID" ]; then
  SCOPE_ID=$(docker exec "$KC_CONTAINER" /opt/keycloak/bin/kcadm.sh get client-scopes -r luxus --fields id,name \
    | awk '/"name" : "organization"/{found=1} found && /"id"/{gsub(/[",]/,"",$3); print $3; exit}')
fi

if [ -z "$SCOPE_ID" ]; then
  echo "client scope organization not found — skip mapper"
else
  EXISTS=$(docker exec "$KC_CONTAINER" /opt/keycloak/bin/kcadm.sh get "client-scopes/${SCOPE_ID}/protocol-mappers/models" -r luxus \
    | grep -c '"name"[[:space:]]*:[[:space:]]*"organization-id-mapper"' || true)
  if [ "$EXISTS" = "0" ]; then
    docker exec "$KC_CONTAINER" /opt/keycloak/bin/kcadm.sh create "client-scopes/${SCOPE_ID}/protocol-mappers/models" -r luxus \
      -s name=organization-id-mapper \
      -s protocol=openid-connect \
      -s protocolMapper=oidc-usermodel-attribute-mapper \
      -s 'config."user.attribute"=organization_id' \
      -s 'config."claim.name"=organization_id' \
      -s 'config."jsonType.label"=String' \
      -s 'config."id.token.claim"=true' \
      -s 'config."access.token.claim"=true' \
      -s 'config."userinfo.token.claim"=true' \
      -s 'config.multivalued=false'
    echo "Created organization-id-mapper"
  else
    echo "organization-id-mapper already exists"
  fi
fi

echo "Done."
