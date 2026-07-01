#!/bin/bash
set -eo pipefail

HOST="$(hostname -s)"
USER="${POSTGRES_USER:-postgres}"
DB="${POSTGRES_DB:-$POSTGRES_USER}"
export PGPASSWORD="${POSTGRES_PASSWORD:-}"

args=(
  --host="$HOST"
  --username="$USER"
  --dbname="$DB"
  --quiet --no-align --tuples-only
)

SELECT="$(echo 'SELECT 1' | psql "${args[@]}")"

if  [ "$SELECT" = '1' ]
then
  exit 0
fi

exit 1
