## ===================== UTILS ======================

.PHONY: install-tools
install-tools: ## Install required dev tools (run once after cloning)
	@echo "Installing dev tools..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/vektra/mockery/v2@latest
	@go install github.com/air-verse/air@latest
	@echo "✅ Done. Ensure $$(go env GOPATH)/bin is in your PATH."

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
	@echo "1️⃣  Starting databases..."
	@docker compose up -d --wait postgresdb mongodb
	@echo "2️⃣  Seeding data..."
	@go run cmd/seed/main.go
	@echo "3️⃣  Stopping databases..."
	@docker compose stop postgresdb mongodb
	@echo "✅ Done. Run 'make up' to restart the full stack."