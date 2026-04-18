## ===================== UTILS ======================

.PHONY: air-build
air-build: ## Build app with Air (hot reload)
	@go build -buildvcs=false -o ./tmp/main ./cmd/api

SWAGGER_MAIN_FILE := cmd/api/main.go
.PHONY: docs
docs: ## Generate Swagger documentation
	@echo "1. Formatting Swagger annotations..."
	@swag fmt
	@echo "2. Generating Swagger files..."
	@swag init -g $(SWAGGER_MAIN_FILE) --output docs/swagger --parseDependency --parseInternal
	@echo "Done"

.PHONY: mocks
mocks: ## Generate mocks using mockery
	@echo "Generating mocks..."
	@mockery
	@echo "Mocks generated successfully!"	

.PHONY: seed
seed: ## Seed database
	@echo "1️⃣  Starting database..."
	docker compose up -d postgresdb mongodb
	@echo "⏳ Waiting 3s for databases to be ready..."
	@sleep 3
	@echo "2️⃣  Seeding data..."
	go run cmd/seed/main.go
	@echo "3️⃣  Stopping databases..."
	docker compose stop postgresdb mongodb
	@echo "✅ Done"