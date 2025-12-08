# Configuration
COMPOSE = docker-compose
APP = promo-bot
BIN = bin/$(APP)
PG_CONTAINER = promo-postgresql
PG_USER = test
PG_DB = test

# Local Development 
build: deps
	go build -o $(BIN) .

run:
	go run .

deps:
	go mod download
	go mod tidy

test:
	go test -v ./...

clean:
	rm -rf bin/

# Docker
compose-build:
	$(COMPOSE) build

compose-build-nocache:
	$(COMPOSE) build --no-cache

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down
downfull:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f

logs-bot:
	$(COMPOSE) logs -f bot

restart:
	$(COMPOSE) restart

clean-volumes:
	$(COMPOSE) down -v --remove-orphans

ps:
	$(COMPOSE) ps

# Database
db:
	docker exec -it $(PG_CONTAINER) psql -U $(PG_USER) -d $(PG_DB)

tables:
	docker exec $(PG_CONTAINER) psql -U $(PG_USER) -d $(PG_DB) -c "\dt"

databases:
	docker exec $(PG_CONTAINER) psql -U $(PG_USER) -d $(PG_DB) -c "\l"

query:
	@if [ -z "$(SQL)" ]; then \
		echo "Usage: make query SQL=\"SELECT * FROM table;\""; \
		exit 1; \
	fi
	docker exec $(PG_CONTAINER) psql -U $(PG_USER) -d $(PG_DB) -c "$(SQL)"

# Workflows
dev: up logs-bot

deploy: compose-build-nocache up

reset: clean-volumes compose-build-nocache up

# Phony Targets
.PHONY: build run deps test clean \
        compose-build compose-build-nocache up down downfull logs logs-bot \
        restart clean-volumes ps \
        db tables databases query exec-sql dump \
        dev deploy reset