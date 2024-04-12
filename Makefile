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

build-debug: ## Builds the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml -f .cloud/docker/docker-compose.debug.yaml build

up: ## Starts the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml up -d

up-dev: ## Starts the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml -f .cloud/docker/docker-compose.dev.yaml up -d

up-debug: ## Starts the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml -f .cloud/docker/docker-compose.debug.yaml up -d

down: ## Stops the Docker images
	@$(DOCKER_COMP) -f .cloud/docker/docker-compose.yaml down

save: ## Save image to disk
	docker save ceherzog/dispatcher -o dispatcher.tar

test: ## Start test and benchmark
	@echo "------ Run test -----"
	go test ./... -coverprofile=./tests/results/coverage.out
	@echo "\n------ Display coverage -----"
	go tool cover -html=./tests/results/coverage.out
	@echo "\n------ Start benchmark -----"
	go test -bench=Create ./tests -run=^# -benchmem -benchtime=10s -cpuprofile=./tests/results/cpu.out -memprofile=./tests/results/mem.out
	go tool pprof -http=:8080 ./tests/results/cpu.out
	go tool pprof -http=:8080 ./tests/results/mem.out


## —— Jenkins —————————————————————————————————————————————————————————————————
build-jenkins: ## Build Jenkins
	@-docker buildx create --name plugin --use --bootstrap
	@docker buildx build \
	-f ../../.cloud/docker/Dockerfile \
	--platform linux/amd64,linux/arm64 \
	--target plugin \
	--build-arg PLUGINNAME=js-license \
	--tag ceherzog/plugin-js-license:latest \
	--push ../..