#!/bin/sh
set -e
/opt/keycloak/bin/kcadm.sh config credentials --server http://localhost:8080 --realm master --user admin --password "${KC_ADMIN_PWD:-admin}"

update_org() {
  USERNAME="$1"
  USER_ID=$(/opt/keycloak/bin/kcadm.sh get users -r luxus -q "username=${USERNAME}" --fields id | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)"$/\1/')
  if [ -z "$USER_ID" ]; then
    echo "User ${USERNAME} not found"
    return
  fi
  /opt/keycloak/bin/kcadm.sh update "users/${USER_ID}" -r luxus \
    -s 'attributes.organization=["{\"luxus\":{\"id\":\"00000000-0000-0000-0000-000000000001\",\"name\":[\"Luxus Connect\"]}}"]'
  echo "Updated organization for ${USERNAME} (${USER_ID})"
  /opt/keycloak/bin/kcadm.sh get "users/${USER_ID}" -r luxus --fields username,attributes
}

update_org dev
update_org parceiro
update_org funcionario
update_org financeiro
