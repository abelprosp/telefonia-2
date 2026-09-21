#!/bin/sh
# Aplica no Keycloak em execução o client scope tenant-organization
# (evita colisão com o scope built-in "organization" do Keycloak Organizations)
# e o mapper organization_id.
#
# Uso na VPS:
#   bash docker/keycloak/ensure-org-id-mapper.sh
set -e

KC_CONTAINER="${KC_CONTAINER:-$(docker ps --format '{{.Names}}' | grep -i keycloak | head -1)}"
if [ -z "$KC_CONTAINER" ]; then
  echo "Keycloak container not found"
  exit 1
fi

ADMIN_USER="${KC_ADMIN_USERNAME:-admin}"
ADMIN_PASS="${KC_ADMIN_PWD:-admin}"
REALM="${KEYCLOAK_REALM:-luxus}"
CLIENT_ID="${KEYCLOAK_RESOURCE:-connect-cli}"

kcadm() {
  docker exec "$KC_CONTAINER" /opt/keycloak/bin/kcadm.sh "$@"
}

echo "Configuring kcadm credentials..."
if ! kcadm config credentials --server http://localhost:8080/auth --realm master --user "$ADMIN_USER" --password "$ADMIN_PASS" 2>/dev/null; then
  kcadm config credentials --server http://localhost:8080 --realm master --user "$ADMIN_USER" --password "$ADMIN_PASS"
fi

SCOPE_NAME="tenant-organization"
echo "Ensuring client scope '${SCOPE_NAME}'..."

SCOPE_ID=$(kcadm get client-scopes -r "$REALM" --fields id,name 2>/dev/null \
  | tr -d '\n' \
  | sed -n "s/.*\"name\"[[:space:]]*:[[:space:]]*\"${SCOPE_NAME}\"[[:space:]]*,[[:space:]]*\"id\"[[:space:]]*:[[:space:]]*\"\\([^\"]*\\)\".*/\\1/p" \
  | head -1)

if [ -z "$SCOPE_ID" ]; then
  SCOPE_ID=$(kcadm get client-scopes -r "$REALM" --fields id,name 2>/dev/null \
    | awk -v n="$SCOPE_NAME" '
      $0 ~ "\"name\"" && $0 ~ "\"" n "\"" { want=1 }
      want && /"id"/ {
        gsub(/[",]/, "", $3);
        print $3;
        exit
      }')
fi

if [ -z "$SCOPE_ID" ]; then
  echo "Creating client scope ${SCOPE_NAME}..."
  kcadm create client-scopes -r "$REALM" \
    -s "name=${SCOPE_NAME}" \
    -s protocol=openid-connect \
    -s 'attributes."include.in.token.scope"=true' \
    -s 'attributes."display.on.consent.screen"=false'
  SCOPE_ID=$(kcadm get client-scopes -r "$REALM" --fields id,name 2>/dev/null \
    | tr -d '\n' \
    | sed -n "s/.*\"name\"[[:space:]]*:[[:space:]]*\"${SCOPE_NAME}\"[[:space:]]*,[[:space:]]*\"id\"[[:space:]]*:[[:space:]]*\"\\([^\"]*\\)\".*/\\1/p" \
    | head -1)
fi

if [ -z "$SCOPE_ID" ]; then
  echo "ERROR: could not resolve client scope id for ${SCOPE_NAME}"
  exit 1
fi
echo "Scope id: ${SCOPE_ID}"

ensure_mapper() {
  NAME="$1"
  ATTR="$2"
  CLAIM="$3"
  JSON_TYPE="$4"
  EXISTS=$(kcadm get "client-scopes/${SCOPE_ID}/protocol-mappers/models" -r "$REALM" 2>/dev/null \
    | grep -c "\"name\"[[:space:]]*:[[:space:]]*\"${NAME}\"" || true)
  if [ "$EXISTS" = "0" ]; then
    echo "Creating mapper ${NAME}..."
    kcadm create "client-scopes/${SCOPE_ID}/protocol-mappers/models" -r "$REALM" \
      -s "name=${NAME}" \
      -s protocol=openid-connect \
      -s protocolMapper=oidc-usermodel-attribute-mapper \
      -s "config.\"user.attribute\"=${ATTR}" \
      -s "config.\"claim.name\"=${CLAIM}" \
      -s "config.\"jsonType.label\"=${JSON_TYPE}" \
      -s 'config."id.token.claim"=true' \
      -s 'config."access.token.claim"=true' \
      -s 'config."userinfo.token.claim"=true' \
      -s 'config.multivalued=false'
  else
    echo "Mapper ${NAME} already exists"
  fi
}

ensure_mapper "organization-mapper" "organization" "organization" "JSON"
ensure_mapper "organization-id-mapper" "organization_id" "organization_id" "String"

echo "Attaching ${SCOPE_NAME} as default scope on client ${CLIENT_ID}..."
CID=$(kcadm get clients -r "$REALM" -q "clientId=${CLIENT_ID}" --fields id,clientId 2>/dev/null \
  | tr -d '\n' \
  | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
  | head -1)

if [ -n "$CID" ]; then
  # Ignore errors if already assigned.
  kcadm update "clients/${CID}/default-client-scopes/${SCOPE_ID}" -r "$REALM" 2>/dev/null || \
  kcadm create "clients/${CID}/default-client-scopes/${SCOPE_ID}" -r "$REALM" 2>/dev/null || true
  echo "Client ${CLIENT_ID} linked to ${SCOPE_NAME}"
else
  echo "WARN: client ${CLIENT_ID} not found — attach scope manually in Keycloak admin"
fi

echo "Done. Users must log out/in to refresh tokens."
