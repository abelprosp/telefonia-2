# Diretório-raíz
ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

# Cores
GREEN   := $(shell tput -Txterm setaf 2)
WHITE   := $(shell tput -Txterm setaf 7)
YELLOW  := $(shell tput -Txterm setaf 3)
RED     := $(shell tput -Txterm setaf 1)
RESET   := $(shell tput -Txterm sgr0)

# Constrói a documentação de cada script, visualizável via 'make' ou 'make help'
# A documentação dee cada script é feita através de uma string começando por '\#\#'
# Uma categoria de comandos pode ser adicionada em uma string iniciando a mesma com @category
HELP_FUN = \
    %help; \
    while(<>) { \
		push @{$$help{$$2 // 'Comandos'}}, [$$1, $$3] if /^([a-zA-Z\-]+)\s*:.*\#\#(?:@([a-zA-Z\-]+))?\s(.*)$$/ \
	}; \
	print "Utilização: make [comando]\n\n"; \
	for (sort keys %help) { \
		print "${WHITE}$$_:${RESET}\n"; \
		for (@{$$help{$$_}}) { \
				$$sep = " " x (32 - length $$_->[0]); \
				print "  ${YELLOW}$$_->[0]${RESET}$$sep${GREEN}$$_->[1]${RESET}\n"; \
		}; \
		print "\n"; \
	}

.PHONY: help \
		header \
		format \
		lint \
		test \
		listsrvs \
		build \
		confirm \
		clean \
		destroy \
		logs \
		restart \
		start \
		init \
		status \
		stop \
		up \
		run \
		exec \
		ps \
		imganalysisci \
		imganalysisui \
		topology

.DEFAULT_GOAL := help

info: header

# Constrói o comando do Docker Compose, dados o ambiente e os argumentos necessários
define compose_cmd
	@$(eval ENV := $(strip $(1)))
	@$(eval ARGS := $(strip $(2)))
	@echo "call_compose_cmd @ [ENV($(ENV))] & [ARGS($(ARGS))]"
	@echo "---------------------------------------------------------------------------------------------"
	@docker compose -f $(ROOT_DIR)/$(shell ./scripts/docker-compose.sh yamlpath $(ENV)) $(ARGS)
endef

define HEADER
+---------------------------------------------------------------------------------------------+
  _                       ___                      _     _    __  __      _        __ _ _     
 | |  _  ___ ___  _ ___  / __|___ _ _  _ _  ___ __| |_  | |  |  \/  |__ _| |_____ / _(_) |___ 
 | |_| || \ \ / || (_-< | (__/ _ \ ' \| ' \/ -_) _|  _| | |  | |\/| / _` | / / -_)  _| | / -_)
 |____\_,_/_\_\\_,_/__/  \___\___/_||_|_||_\___\__|\__| | |  |_|  |_\__,_|_\_\___|_| |_|_\___|
                                                        |_|
+---------------------------------------------------------------------------------------------+
endef
export HEADER

header: ##@Outros Mostra o header deste help, formado com caracteres ASCII.
	clear
	@echo "$$HEADER"

help: ##@Outros Mostra esta documentação.
	clear
	@echo "$$HEADER"
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

listsrvs: ## Lista todos os nomes de serviços declarados no YAML do Docker Compose, dado um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), config --services)

build: ## Realiza a build de todas as imagens Docker, ou para um c=<node de serviço> específico, dado um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), build $(c))

confirm:
	@( read -p "$(RED)Tem certeza? [y/N]$(RESET): " sure && case "$$sure" in [sSyY]) true;; *) false;; esac )

clean: confirm ## Realiza a limpeza de todos os dados associados aos conteineres, dado um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), down)

destroy: confirm ## Remove todas as imagens, volumes, networks e conteineres não utilizados. Use com cautela!
	@docker system prune --all --volumes --force
	@docker volume prune --all --force
	@docker network prune --force
	@docker image prune --all --force

logs: ## Adiciona captura de logs para todos os conteineres ou para um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), logs --follow $(c))

restart: ## Reinicia todos os conteineres ou apenas um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), stop $(c))
	@make init c=$(c)

start: ## Inicia todos os conteineres em background (detached mode) ou apenas um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), up -d $(c))

init: ## Inicia um conteiner em detached mode, com captura de logs, dado um env=<dev | prod> ambiente de infra
	@make start env=$(env) c=$(c) && make logs env=$(env) c=$(c)

status: ## Lista os status dos conteineres em execução, dado um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), ps)

stop: ## Encerra a execução de todos os conteineres ou de apenas um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), stop $(c))

up: ## Inicia todos os conteineres em modo "attached" ou apenas um c=<nome de serviço>, dado um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), up $(c))

run: ## Roda um comando (o que seria especificado em 'CMD' na imagem), dado um c=<nome de serviço> e um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), run --rm $(c) $(cmd))

exec: ## Executa um comando em um container já iniciado, dado um c=<nome de serviço> e um s=<script> e um env=<dev | prod> ambiente de infra
	$(call compose_cmd, $(env), exec -it $(c) $(s))

ps: status ## Alias do comando 'status'

imganalysisui: ## Executa a análise de uma imagem Docker, em modo UI, dado uma img=<imagem Docker>
	@./scripts/docker-analysis.sh ui $(img)

imganalysisci: ## Executa a análise de uma imagem Docker, em modo CI, dado uma img=<imagem Docker>
	@./scripts/docker-analysis.sh ci $(img)

topology: ## Gera um diagrama dos serviços listados no arquivo YML do Docker Compose
	@./scripts/generate-topology.sh topology $(env)