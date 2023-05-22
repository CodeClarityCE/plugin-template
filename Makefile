# Executables (local)
DOCKER_COMP = docker compose

# Misc
.DEFAULT_GOAL = help
.PHONY        = help build up start down logs sh composer vendor sf cc

## —— 🐳 The PHP pipeline Makefile 🐳 ——————————————————————————————————
help: ## Outputs this help screen
	@grep -E '(^[a-zA-Z0-9_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

## —— Docker 🐳 ————————————————————————————————————————————————————————————————
build: ## Builds the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml build --pull --no-cache

build-dev: ## Builds the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml -f .cloud/docker/docker-compose.dev.yaml build

up: ## Starts the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml up -d

up-dev: ## Starts the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml -f .cloud/docker/docker-compose.dev.yaml up -d

down: ## Stops the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml down

save: ## Save image to disk
	docker save ceherzog/dispatcher -o dispatcher.tar

tests: ## Start test and benchmark
	go test -bench=.