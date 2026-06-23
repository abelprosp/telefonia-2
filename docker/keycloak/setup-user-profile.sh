#!/bin/sh
set -e
/opt/keycloak/bin/kcadm.sh config credentials --server http://localhost:8080 --realm master --user admin --password "${KC_ADMIN_PWD:-admin}"

echo "Updating luxus user profile..."
/opt/keycloak/bin/kcadm.sh update realms/luxus/users/profile -r luxus -f /tmp/user-profile.json

ORG='{"luxus":{"id":"00000000-0000-0000-0000-000000000001","name":["Luxus Connect"]}}'

update_org() {
  USERNAME="$1"
  USER_ID=$(/opt/keycloak/bin/kcadm.sh get users -r luxus -q "username=${USERNAME}" --fields id | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"\([^"]*\)"$/\1/')
  if [ -z "$USER_ID" ]; then
    echo "User ${USERNAME} not found"
    return
  fi
  /opt/keycloak/bin/kcadm.sh update "users/${USER_ID}" -r luxus \
    -s "attributes.organization=[\"${ORG}\"]"
  echo "Updated organization for ${USERNAME} (${USER_ID})"
}

for USERNAME in dev parceiro funcionario financeiro; do
  update_org "$USERNAME"
done

echo "Done."
