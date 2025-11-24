DOCKER_COMPOSE = docker-compose
APP_NAME = promo-bot
BINARY_NAME = bin/$(APP_NAME)

.PHONY: build run deps test clean docker-build docker-up docker-down docker-logs docker-clean

.DEFAULT_GOAL := build

# Local development
build: deps
	go build -o $(BINARY_NAME) .

#по сути это run не работает
run:
	go run .

deps:
	go mod download
	go mod tidy

test:
	go test -v ./...

clean:
	rm -rf bin/

# Docker commands
docker-build:
	$(DOCKER_COMPOSE) build

docker-build-nocache:
	$(DOCKER_COMPOSE) build --no-cache

docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f

docker-logs-bot:
	$(DOCKER_COMPOSE) logs -f bot

docker-restart:
	$(DOCKER_COMPOSE) restart

docker-clean:
	$(DOCKER_COMPOSE) down -v --remove-orphans

docker-ps:
	$(DOCKER_COMPOSE) ps

# Combined workflows
dev: docker-up docker-logs-bot

deploy: docker-build-nocache docker-up

stop: docker-down

status: docker-ps

reset: docker-clean docker-build-nocache docker-up

# Utility commands
env-check:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found"; \
		exit 1; \
	fi