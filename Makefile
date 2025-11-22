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
	$(MAKE) down
	$(MAKE) up

.PHONY: logs
logs:
	@docker compose logs -f app

.PHONY: ps
ps:
	@docker compose ps

.PHONY: rebuild
rebuild:
	@echo "♻️ Rebuilding all images"
	@docker compose build --no-cache
	$(MAKE) up


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


# ===================== AIR (for dev hot reload) ======================
# Air.toml references this:
# cmd = "make air-build"

.PHONY: air-build
air-build:
	@go build -o ./tmp/main -buildvcs=false .
