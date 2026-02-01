# Configuration
-include .env
export

COMPOSE = docker compose -f docker-compose.yml -f docker-compose.override.yml
APP = promo-bot
BIN = bin/$(APP)
DB_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

.PHONY: migrate-up migrate-down migrate-force migrate-create migrate-version


migrate-up:
	migrate -path ./db/migrations -database "$(DB_URL)" up
migrate-down:
	migrate -path ./db/migrations -database "$(DB_URL)" down 1
migrate-force:
ifeq ($(strip $(NAME)),)
	@echo Usage: make migrate-force VERSION=...
	@exit 1
endif
	migrate -path ./db/migrations -database "$(DB_URL)" force $(VERSION)
migrate-create:
ifeq ($(strip $(NAME)),)
	@echo Usage: make migrate-create NAME=...
	@exit 1
endif
	migrate create -ext sql -dir migrations -seq $(NAME)
migrate-version:
	migrate -path ./db/migrations -database "$(DB_URL)" version


# Local Development
.env:
ifeq ($(OS),Windows_NT)
	@powershell -Command "if ([string]::IsNullOrEmpty('$(TOKEN)')) { \
		$$secureToken = Read-Host -Prompt 'Please enter TOKEN' -AsSecureString; \
		$$token = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($$secureToken)); \
		if ([string]::IsNullOrEmpty($$token)) { \
			Write-Error 'TOKEN cannot be empty'; \
			exit 1 \
		}; \
		Copy-Item .env.example .env; \
		(Get-Content .env) -replace '^API_TOKEN=.*', \"API_TOKEN=$$token\" | Set-Content .env \
	} else { \
		Copy-Item .env.example .env; \
		(Get-Content .env) -replace '^API_TOKEN=.*', 'API_TOKEN=$(TOKEN)' | Set-Content .env \
	}"
else
	@if [ -z "$(TOKEN)" ]; then \
		echo -n "Please enter TOKEN: "; \
		read -s token; \
		echo ""; \
		if [ -z "$$token" ]; then \
			echo "Error: TOKEN cannot be empty"; \
			exit 1; \
		fi; \
		cp .env.example .env; \
		sed -i "s/^API_TOKEN=.*/API_TOKEN=$$token/" .env; \
	else \
		cp .env.example .env; \
		sed -i 's/^API_TOKEN=.*/API_TOKEN=$(TOKEN)/' .env; \
	fi
endif
	@echo ".env file created successfully"

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
	$(COMPOSE) up -d --build

up-infra:
	$(COMPOSE) up -d postgresql redis

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
	$(COMPOSE) exec postgresql psql -U $(POSTGRES_USER) $(POSTGRES_DB)

tables:
	$(COMPOSE) exec postgresql psql -U $(POSTGRES_USER) $(POSTGRES_DB) -c "\dt"

databases:
	$(COMPOSE) exec postgresql psql -U $(POSTGRES_USER) $(POSTGRES_DB) -c "\l"

query:
ifeq ($(strip $(NAME)),)
	@echo Usage: make query SQL="SELECT * FROM table;"
	@exit 1
endif
	$(COMPOSE) exec postgresql psql -U $(POSTGRES_USER) $(POSTGRES_DB) -c "$(SQL)"

dump:
	$(COMPOSE) exec postgresql pg_dump -U $(POSTGRES_USER) $(POSTGRES_DB) > dump.sql

# Workflows
dev: up logs-bot

deploy: compose-build-nocache up

reset: clean-volumes compose-build-nocache up

lint:
	golangci-lint run ./...

# Phony Targets
.PHONY: build run deps test clean \
        compose-build compose-build-nocache up down downfull logs logs-bot \
        restart clean-volumes ps \
        db tables databases query dump \
        dev deploy reset lint