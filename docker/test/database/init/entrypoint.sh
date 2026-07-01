#!/bin/bash

# Immediately exits if any error occurs during the script
# execution. If not set, an error could occur and the
# script would continue its execution.
set -o errexit


# Creating an array that defines the environment variables
# that must be set. This can be consumed later via array
# variable expansion ${REQUIRED_ENV_VARS[@]}.
readonly REQUIRED_ENV_VARS=(
    "DB_NAME"
    "DB_USERNAME"
    "DB_PASSWORD"
    "POSTGRES_USER"
    "POSTGRES_PORT"
)

# Main execution:
# - verifies if all environment variables are set
# - runs the SQL code to create user and database
main() {
    check_environment
    create_empty_db
    configure_db_port
}

# Checks if all of the required environment
# variables are set. If one of them isn't,
# echoes a text explaining which one isn't
# and the name of the ones that need to be
check_environment() {
    for required_env_var in ${REQUIRED_ENV_VARS[@]}; do
        if [[ -z "${!required_env_var}" ]]; then
            echo "Error:
        Environment variable '$required_env_var' not set.
        Make sure you have the following environment variables set:
            ${REQUIRED_ENV_VARS[@]}
        Aborting."
            exit 1
        fi
    done
}

# Performs the initialization in the already-started PostgreSQL
# using the preconfigured POSTGRE_USER user.
create_empty_db() {
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    CREATE USER $DB_USERNAME WITH PASSWORD '$DB_PASSWORD' NOSUPERUSER CREATEDB CREATEROLE INHERIT;
    GRANT CONNECT ON DATABASE postgres TO $DB_USERNAME;
    GRANT USAGE ON SCHEMA public TO $DB_USERNAME;
    GRANT CREATE ON SCHEMA public TO $DB_USERNAME;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USERNAME;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USERNAME;
    SELECT 'CREATE DATABASE "luxus_connect_dev_db"' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'luxus_connect_dev_db');\gexec
    SELECT 'CREATE DATABASE "luxus_keycloak_dev_db"' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'luxus_keycloak_dev_db');\gexec
    ALTER DATABASE "luxus_connect_dev_db" OWNER TO $DB_USERNAME; GRANT ALL PRIVILEGES ON DATABASE "luxus_connect_dev_db" TO $DB_USERNAME;
    ALTER DATABASE "luxus_keycloak_dev_db" OWNER TO $DB_USERNAME; GRANT ALL PRIVILEGES ON DATABASE "luxus_keycloak_dev_db" TO $DB_USERNAME;
EOSQL
}

configure_db_port() {
    if [[ -n "$POSTGRES_PORT" ]]; then
        POSTGRES_CONFIGURATION_FILE=$PGDATA/postgresql.conf
        POSTGRES_CONFIGURATION_MARKER="## PostgreSQL port configuration"
        echo "Configuring PostgreSQL port"

        if grep -Fxq "$POSTGRES_CONFIGURATION_MARKER" $POSTGRES_CONFIGURATION_FILE
        then
            # configuration file already written
            echo "PostgreSQL port already written, skipping"
        else
            # write configuration file
            pg_ctl -D "$PGDATA" -m fast -w stop
            echo "PostgreSQL configuration port update being written: $POSTGRES_PORT"
            echo "$POSTGRES_CONFIGURATION_MARKER" >> $POSTGRES_CONFIGURATION_FILE
            echo "port = $POSTGRES_PORT" >> $POSTGRES_CONFIGURATION_FILE

            pg_ctl -D "$PGDATA" -w start
        fi
    fi
}

# Executes the main routine with environment variables
# passed through the command line. We don't use them in
# this script but now you know ;)
main "$@"
