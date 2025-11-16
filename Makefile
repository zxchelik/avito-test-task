# Загружаем переменные из .env файла
include .env
export $(shell sed 's/=.*//' .env)

GOOSE_DRIVER = postgres
GOOSE_DBSTRING = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)
MIGRATIONS_DIR = migrations
BASE_URL = http://0.0.0.0:8080

.PHONY: migrate-up migrate-down migrate-status migrate-create

# Применение миграций
migrate-up:
	@echo "Applying migrations from $(MIGRATIONS_DIR)..."
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATIONS_DIR) up

# Откат миграций
migrate-down:
	@echo "Rolling back migrations..."
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATIONS_DIR) down

# Просмотр статуса миграций
migrate-status:
	@echo "Migration status in $(MIGRATIONS_DIR):"
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATIONS_DIR) status

# Создание миграции
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "❌ Укажите имя миграции: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)..."
	GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose -dir $(MIGRATIONS_DIR) -s create $(name) sql

run:
	docker compose up -d \

stop:
	docker compose down


lint:
	golangci-lint run --fix

load-test:
	make run
	k6 run -e BASE_URL=$(BASE_URL) load-test.js
	make stop