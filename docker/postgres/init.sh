#!/bin/bash
set -euo pipefail

db_name="${POSTGRES_DB:-luxus_connect_dev}"
keycloak_db="luxus_kc_dev"

create_database_if_missing() {
	local name="$1"
	local exists
	exists="$(psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres \
		-tAc "SELECT 1 FROM pg_database WHERE datname = '${name}'")"
	if [[ "$exists" != "1" ]]; then
		createdb --username "$POSTGRES_USER" "$name"
	fi
}

create_database_if_missing "$db_name"
create_database_if_missing "$keycloak_db"

shopt -s nullglob
migrations=(/docker-entrypoint-initdb.d/migrations/[0-9]*.sql)
if ((${#migrations[@]} == 0)); then
	echo "WARNING: no SQL migrations found under /docker-entrypoint-initdb.d/migrations"
	exit 0
fi

for migration in "${migrations[@]}"; do
	echo "Applying $migration to $db_name"
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$db_name" -f "$migration"
done
