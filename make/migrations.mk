## ===================== MIGRATIONS ======================

# HOST_UID and HOST_GID store the current user's UID and GID on the host machine.
HOST_UID := $(shell id -u)
HOST_GID := $(shell id -g)

# Path to migrations inside the container
MIGRATIONS_PATH = /app/cmd/migrations

# Base migrate command (runs inside app container)
MIGRATE_CMD = docker compose exec app migrate -path=$(MIGRATIONS_PATH) -database "$(DB_DSN)"

.PHONY: migrate-version
migrate-version: ## Check current migration version
	@echo "🔎  Checking current migration version..."
	-@$(MIGRATE_CMD) version || echo "No migrations applied yet."

.PHONY: migrate-create
migrate-create: ## Create new migration (Usage: make migrate-create name=create_users_table)
ifndef name
	$(error Usage: make migrate-create name=<migration_name>)
endif
	@echo "🆕  Creating migration: $(name)"
	docker compose exec -u $(HOST_UID):$(HOST_GID) app \
		migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(name)

.PHONY: migrate-up
migrate-up: ## Apply migrations (Usage: make migrate-up [n=1])
	@echo "⬆️  Applying migrations..."
	@$(MIGRATE_CMD) up $(n)

.PHONY: migrate-down
migrate-down: ## Rollback migrations (Usage: make migrate-down [n=1])
	@echo "⬇️  Rolling back migrations..."
	@$(MIGRATE_CMD) down $(n)

.PHONY: migrate-force
migrate-force: ## Force migration version (Usage: make migrate-force version=12)
ifndef version
	$(error Usage: make migrate-force version=<version_number>)
endif
	@echo "⚠️  Forcing migration version to $(version)..."
	@$(MIGRATE_CMD) force $(version)

.PHONY: reset-postgresdb
reset-postgresdb: ## Reset Database (Drop & Create only) - Data Loss!
	@echo "🔄 Starting database container..."
	@docker compose up -d postgresdb
	@sleep 2
	@echo "🛑 Killing connections to $(DB_NAME)..."
	@docker compose exec postgresdb psql -U $(DB_USER) -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$(DB_NAME)' AND pid <> pg_backend_pid();"
	@echo "🗑️ Dropping database $(DB_NAME)..."
	@docker compose exec postgresdb psql -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "✨ Creating empty database $(DB_NAME)..."
	@docker compose exec postgresdb psql -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_NAME);"
	@echo "✅ Database reset done! (Empty DB created)"
	@echo "👉 Next step: Run 'make migrate-up' to create tables."