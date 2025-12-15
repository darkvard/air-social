include .env	

# ===================== DOCKER COMPOSE ======================

.PHONY: up
up:
	@echo "🚀 Starting all services (app + db + redis)"
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
	@docker compose ps

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

# ===================== MIGRATIONS ======================


# HOST_UID and HOST_GID store the current user's UID and GID on the host machine.
# These are used when running commands inside the container (via `docker compose exec -u`)
# to ensure that files created inside the container are owned by the host user, not by root.
# This avoids permission issues where VS Code cannot edit or save files created by the container.
HOST_UID := $(shell id -u)
HOST_GID := $(shell id -g)


# Path to migrations inside the container
MIGRATIONS_PATH = /app/internal/repository/migrations


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


# ===================== AIR =====================

# Air.toml references this:
.PHONY: air-build
air-build:
	@go build -buildvcs=false -o ./tmp/main ./cmd/api


# ===================== RABBIT MQ ======================

.PHONY: rabbitmq-ui
rabbitmq-ui:
	@echo "🐰 RabbitMQ Management UI"
	@echo "----------------------------------"
	@echo "URL      : http://localhost:$(RABBITMQ_UI_PORT)"
	@echo "Username : $(RABBITMQ_USER)"
	@echo "Password : $(RABBITMQ_PASS)"
	@echo "----------------------------------"
	@xdg-open http://localhost:$(RABBITMQ_UI_PORT) >/dev/null 2>&1 &

