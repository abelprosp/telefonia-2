#!/bin/bash

function yamlpath() {
    ENVIRONMENT=${1:-"develop"}
    INFRA_ROOT_PATH="infra/docker"
    COMPOSE_YAML_BASENAME="docker-compose"

    case $ENVIRONMENT in
        dev|develop|development|desenvolvimento)
            echo "$INFRA_ROOT_PATH/develop/$COMPOSE_YAML_BASENAME.yml" ;;
        prod|production|production)
            echo "$INFRA_ROOT_PATH/production/$COMPOSE_YAML_BASENAME.yml" ;;
        *) ;;
    esac 
}

$@
