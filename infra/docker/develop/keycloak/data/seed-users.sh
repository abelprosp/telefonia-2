#!/bin/sh
set -e
/opt/keycloak/bin/kcadm.sh config credentials --server http://localhost:8080 --realm master --user admin --password "${KC_ADMIN_PWD:-admin}"

ORG='{"luxus":{"id":"00000000-0000-0000-0000-000000000001","name":["Luxus Connect"]}}'

create_user() {
  USERNAME="$1"
  PASSWORD="$2"
  FIRST="$3"
  LAST="$4"
  EMAIL="$5"
  ROLES="$6"

  EXISTING=$(/opt/keycloak/bin/kcadm.sh get users -r luxus -q "username=${USERNAME}" --fields id 2>/dev/null | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)"$/\1/')
  if [ -n "$EXISTING" ]; then
    echo "User ${USERNAME} already exists (${EXISTING})"
    USER_ID="$EXISTING"
    /opt/keycloak/bin/kcadm.sh update "users/${USER_ID}" -r luxus \
      -s 'attributes.organization=["{\"luxus\":{\"id\":\"00000000-0000-0000-0000-000000000001\",\"name\":[\"Luxus Connect\"]}}"]'
  else
    /opt/keycloak/bin/kcadm.sh create users -r luxus \
      -s "username=${USERNAME}" \
      -s "enabled=true" \
      -s "emailVerified=true" \
      -s "firstName=${FIRST}" \
      -s "lastName=${LAST}" \
      -s "email=${EMAIL}" \
      -s 'attributes.organization=["{\"luxus\":{\"id\":\"00000000-0000-0000-0000-000000000001\",\"name\":[\"Luxus Connect\"]}}"]'
    USER_ID=$(/opt/keycloak/bin/kcadm.sh get users -r luxus -q "username=${USERNAME}" --fields id | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)"$/\1/')
    echo "Created user ${USERNAME} (${USER_ID})"
  fi

  /opt/keycloak/bin/kcadm.sh set-password -r luxus --username "${USERNAME}" --new-password "${PASSWORD}" >/dev/null

  for ROLE in $ROLES; do
    /opt/keycloak/bin/kcadm.sh add-roles -r luxus --uusername "${USERNAME}" --rolename "${ROLE}" 2>/dev/null || true
  done
}

create_user dev dev Dev Luxus dev@luxus.local "master admin user"
create_user parceiro parceiro Parceiro Luxus parceiro@luxus.local "partner user"
create_user funcionario funcionario Funcionario Luxus funcionario@luxus.local "employee user"
create_user financeiro financeiro Financeiro Luxus financeiro@luxus.local "financial user"

echo "Done seeding luxus users."
