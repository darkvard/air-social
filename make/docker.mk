## ===================== DOCKER COMPOSE ======================

.PHONY: up
up: ## Start services in detached mode
	@echo "🚀 Starting services"
	@docker compose up -d

.PHONY: down
down: ## Stop services
	@echo "🛑 Stopping services"
	@docker compose down

.PHONY: restart
restart: ## Restart app container
	@echo "🔁 Restarting app"
	@docker compose restart app

.PHONY: rebuild
rebuild: ## Rebuild all images and start services
	@echo "♻️ Rebuilding all images"
	@docker compose build --no-cache
	$(MAKE) up

.PHONY: logs
logs: ## Follow app logs
	@docker compose logs -f app

.PHONY: ps
ps: ## List containers
	@docker compose ps -a 

.PHONY: debug
debug: ## Start services in debug mode
	docker compose -f docker-compose.yml -f docker-compose.debug.yml up -d	

## ===================== DEV WORKFLOW ======================

.PHONY: infra
infra: ## Start only dependencies (DB, Redis, MQ, Nginx, MinIO)
	@echo "🏗️  Starting infrastructure..."
	@docker compose up -d db redis rabbitmq nginx minio

.PHONY: app-up
app-up: ## Start/Restart App container only
	@echo "🚀 Starting App container..."
	@docker compose up -d app

.PHONY: app-down
app-down: ## Stop App container only
	@echo "🛑 Stopping App container..."
	@docker compose stop app

## ===================== Utils ======================

.PHONY: sh-app
sh-app: ## Open shell in app container
	@docker compose exec app sh

.PHONY: sh-db
sh-db: ## Open shell in db container
	@docker compose exec db sh

.PHONY: sh-redis
sh-redis: ## Open shell in redis container
	@docker compose exec redis sh

.PHONY: sh-rabbitmq
sh-rabbitmq: ## Open shell in rabbitmq container
	@docker compose exec rabbitmq sh

.PHONY: sh-nginx
sh-nginx: ## Open shell in nginx container
	@docker compose exec nginx sh