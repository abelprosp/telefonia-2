#!/bin/bash

set -e

DEFAULT_DIVE_VERSION=latest
DEFAULT_DIVE_IMAGE=docker.io/wagoodman/dive
DEFAULT_DIVE_UI_CONFIG_PATH="$(pwd)/.dive-ui.yaml"
DEFAULT_DIVE_CI_CONFIG_PATH="$(pwd)/.dive-ci.yaml"
DEFAULT_DOCKER_SOCK_PATH=/var/run/docker.sock

function usage() {
    MODE=$1

    case $MODE in
        "ui")
            CONFIG_PATH=$DEFAULT_DIVE_UI_CONFIG_PATH
            WHAT_IT_DOES="Realiza análise da imagem Docker, mostrando ao final a UI.";;
        "ci")
            CONFIG_PATH=$DEFAULT_DIVE_CI_CONFIG_PATH
            WHAT_IT_DOES="Realiza análise da imagem Docker, sem mostrar a UI, i.e. modo pass/fail.";;
        *)
            echo "Modo de execução não suportado. Valores suportados: ui | ci"
            exit 1;;
    esac

    echo "----------------------------------------------------------------------------------------------------------------------------------"
    echo -e "\033[1m${MODE^^}: $WHAT_IT_DOES\033[0m"
    echo "----------------------------------------------------------------------------------------------------------------------------------"
    echo "Comando: ./docker-analysis.sh $MODE [target-image] [config-file-path] [dive-version]"
    echo "----------------------------------------------------------------------------------------------------------------------------------"
    echo "Onde:"
    echo "- target-image (string):       O repositório da imagem a ser analisada, sem a tag (e.g. companyrepo/whateverisinside)."
    echo "- config-file-path (string):   O caminho (absoluto) do arquivo de configurações."
    echo "                               Valor padrão: \"$CONFIG_PATH\""
    echo "- dive-version (string):       A versão, em formato de versionamento semântico da aplicação Dive."
    echo "                               Valor padrão: \"latest\""
    echo
    echo "Exemplo: ./docker-analysis.sh $MODE docker.io/wagoodman/dive /home/foo/.dive.yaml 0.9.2"
    echo "----------------------------------------------------------------------------------------------------------------------------------"
}

function trim() {
    echo "$(echo -e "${1}" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
}

function validate_input() {
    MODE=$1
    TARGET_IMAGE=$2

    if [ ! $TARGET_IMAGE ]
    then
        echo -e "ERRO: O repositório para a imagem Docker deve ser especificado!\n"
        usage $MODE
        exit 1
    fi
}

function ui() {
    TARGET_IMAGE=$(trim $1)
    CONFIG_PATH=$(trim ${2:-DEFAULT_DIVE_UI_CONFIG_PATH})
    DIVE_VERSION=$(trim ${3:-$DEFAULT_DIVE_VERSION})

    validate_input "ui" $TARGET_IMAGE

    DIVE_IMAGE="$DEFAULT_DIVE_IMAGE:$DIVE_VERSION"

    docker run --rm -it \
        -v $DEFAULT_DOCKER_SOCK_PATH:$DEFAULT_DOCKER_SOCK_PATH \
        -v  "$(pwd)":"$(pwd)" \
        -v $CONFIG_PATH:"$HOME/.dive.yaml" \
        $DIVE_IMAGE \
            $TARGET_IMAGE

    exit 0
}

function ci() {
    TARGET_IMAGE=$(trim $1)
    CONFIG_PATH=$(trim ${2:-DEFAULT_DIVE_CI_CONFIG_PATH})
    DIVE_VERSION=$(trim ${3:-$DEFAULT_DIVE_VERSION})

    validate_input "ci" $TARGET_IMAGE

    DIVE_IMAGE="$DEFAULT_DIVE_IMAGE:$DIVE_VERSION"

    docker run \
        --rm \
        -it \
        -e CI=true \
        -v $DEFAULT_DOCKER_SOCK_PATH:$DEFAULT_DOCKER_SOCK_PATH \
        $DIVE_IMAGE \
            --ci-config $CONFIG_PATH \
            $TARGET_IMAGE

    exit 0
}

function volumes() {
    for d in `docker ps -a | awk '{print $1}' | tail -n +2`; do
        d_name=`docker inspect -f {{.Name}} $d`
        echo "========================================================="
        echo "$d_name ($d) volumes:"

        VOLUME_IDS=$(docker inspect -f "{{.Config.Volumes}}" $d)
        VOLUME_IDS=$(echo ${VOLUME_IDS} | sed 's/map\[//' | sed 's/]//')

        array=(${VOLUME_IDS// / })

        for i in "${!array[@]}"
        do
            VOLUME_ID=$(echo ${array[i]} | sed 's/:{}//')
            VOLUME_SIZE=$(docker exec -ti $d_name du -d 0 -h ${VOLUME_ID})
            echo "$VOLUME_SIZE"
        done
    done

    exit 0
}

$@
