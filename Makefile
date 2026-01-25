include .env	

# ===================== DOCKER COMPOSE ======================

.PHONY: up
up:
	@echo "🚀 Starting all services"
	@docker compose up -d

.PHONY: down
down:
	@echo "🛑 Stopping all services"
	@docker compose down

.PHONY: restart
restart:
	@echo "🔁 Restarting app"
	@docker compose restart app

.PHONY: rebuild
rebuild:
	@echo "♻️ Rebuilding all images"
	@docker compose build --no-cache
	$(MAKE) up

.PHONY: logs
logs:
	@docker compose logs -f app

.PHONY: ps
ps:
	@docker compose ps -a

.PHONY: debug
debug:
	@echo "Starting debug dependencies..."
	@docker compose up -d nginx db redis rabbitmq minio 
	@echo "🛑 Killing Docker App container to use Local Debugger..."
	@docker compose stop app

## Utils

.PHONY: sh-app
sh-app:
	@docker compose exec app sh

.PHONY: sh-db
sh-db:
	@docker compose exec db sh

.PHONY: sh-redis
sh-redis:
	@docker compose exec redis sh

.PHONY: sh-rabbitmq
sh-rabbitmq:
	@docker compose exec rabbitmq sh

.PHONY: sh-nginx
sh-nginx:
	@docker compose exec nginx sh	

# ===================== MIGRATIONS ======================


# HOST_UID and HOST_GID store the current user's UID and GID on the host machine.
# These are used when running commands inside the container (via `docker compose exec -u`)
# to ensure that files created inside the container are owned by the host user, not by root.
# This avoids permission issues where VS Code cannot edit or save files created by the container.
HOST_UID := $(shell id -u)
HOST_GID := $(shell id -g)


# Path to migrations inside the container
MIGRATIONS_PATH = /app/internal/infrastructure/postgres/migrations


# Base migrate command (runs inside app container)
MIGRATE_CMD = docker compose exec app migrate -path=$(MIGRATIONS_PATH) -database "$(DB_DSN)"


# ---- Show current version ----
.PHONY: migrate-version
migrate-version:
	@echo "🔎  Checking current migration version..."
	-@$(MIGRATE_CMD) version || echo "No migrations applied yet."


# ---- Create new migration ----
# Usage:
#   make migrate-create name=create_users_table
.PHONY: migrate-create
migrate-create:
ifndef name
	$(error Usage: make migrate-create name=<migration_name>)
endif
	@echo "🆕  Creating migration: $(name)"
	docker compose exec -u $(HOST_UID):$(HOST_GID) app \
		migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(name)


# ---- Apply migrations (up) ----
# Usage:
#   make migrate-up
#   make migrate-up n=1
.PHONY: migrate-up
migrate-up:
	@echo "⬆️  Applying migrations..."
	@$(MIGRATE_CMD) up $(n)


# ---- Rollback (down) ----
# Usage:
#   make migrate-down
#   make migrate-down n=1
.PHONY: migrate-down
migrate-down:
	@echo "⬇️  Rolling back migrations..."
	@$(MIGRATE_CMD) down $(n)


# ---- Force migration version (dirty state fix) ----
# Usage:
#   make migrate-force version=12
.PHONY: migrate-force
migrate-force:
ifndef version
	$(error Usage: make migrate-force version=<version_number>)
endif
	@echo "⚠️  Forcing migration version to $(version)..."
	@$(MIGRATE_CMD) force $(version)


# ---- Reset Database (Drop & Create only) ----
# Usage:
#   make reset-db
# Note: 
#   1. Kills all active connections (DataGrip, App...)
#   2. Drops the database (Data loss!)
#   3. Creates a new EMPTY database.
#   4. DOES NOT run migrations (You must run 'make migrate-up' manually).
.PHONY: reset-db
reset-db:
	@echo "🔄 Starting database container..."
	@docker compose up -d db
	@sleep 2

	@echo "🛑 Killing connections to $(DB_NAME)..."
	@docker compose exec db psql -U $(DB_USER) -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$(DB_NAME)' AND pid <> pg_backend_pid();"
	
	@echo "🗑️ Dropping database $(DB_NAME)..."
	@docker compose exec db psql -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"

	@echo "✨ Creating empty database $(DB_NAME)..."
	@docker compose exec db psql -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_NAME);"
	
	@echo "✅ Database reset done! (Empty DB created)"
	@echo "👉 Next step: Run 'make migrate-up' to create tables."


# ===================== TESTING ======================


.PHONY: test
test:
	@echo "==> Running all tests..."
	@go test -v ./...

.PHONY: test-cover
test-cover:
	@echo "==> Running tests with coverage..."
	@go test -cover -v ./...

.PHONY: test-bench
test-bench:
	@echo "==> Running benchmarks..."
	@go test -bench=. -benchmem ./...
 
# Usage: make test-pkg pkg=./mypackage
.PHONY: test-pkg
test-pkg:
	@echo "==> Testing package: $(pkg)"
	@go test -v $(pkg)

# Usage: make test-pkg-cover pkg=./mypackage
.PHONY: test-pkg-cover
test-pkg-cover:
	@echo "==> Testing package with coverage: $(pkg)"
	@go test -cover -v $(pkg)

# Usage: make bench-pkg pkg=./mypackage
.PHONY: test-bench-pkg
test-bench-pkg:
	@echo "==> Benchmarking package: $(pkg)"
	@go test -bench=. -benchmem $(pkg)	


# ===================== Other =====================

# Air reload, air.toml references this:
.PHONY: air-build
air-build:
	@go build -buildvcs=false -o ./tmp/main ./cmd/api


# Swagger api documentation
SWAGGER_MAIN_FILE := cmd/api/main.go
.PHONY: docs
docs:
	@echo "1. Formatting Swagger annotations..."
	@swag fmt
	@echo "2. Generating Swagger files..."
	@swag init -g $(SWAGGER_MAIN_FILE) --output docs/swagger
	@echo "Done"

# Generate mocks
.PHONY: mocks
mocks:
	@echo "Generating mocks..."
	@rm -rf internal/mocks/*
	@mockery
	@echo "Mocks generated successfully!"	