## ===================== TESTING ======================

.PHONY: test
test: ## Run all tests
	@echo "==> Running all tests..."
	@go test -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage overview
	@echo "==> Running tests with coverage overview..."
	@go test -cover ./...

.PHONY: test-cover-html
test-cover-html: ## Generate HTML coverage report and open in browser
	@echo "==> Generating HTML coverage report..."
	@mkdir -p tmp
	@go test -coverprofile=tmp/coverage.out ./...
	@go tool cover -func=tmp/coverage.out
	@echo "==> Opening coverage report in browser..."
	@go tool cover -html=tmp/coverage.out

.PHONY: test-cover-func
test-cover-func: ## Generate function coverage report (terminal)
	@echo "==> Generating function coverage report..."
	@mkdir -p tmp
	@go test -coverprofile=tmp/coverage.out ./...
	@go tool cover -func=tmp/coverage.out

.PHONY: test-bench
test-bench: ## Run benchmarks
	@echo "==> Running benchmarks..."
	@go test -bench=. -benchmem ./...

.PHONY: test-pkg
test-pkg: ## Test specific package (Usage: make test-pkg pkg=./mypackage)
	@echo "==> Testing package: $(pkg)"
	@go test -v $(pkg)

.PHONY: test-pkg-cover
test-pkg-cover: ## Generate HTML coverage for specific package (Usage: make test-pkg-cover pkg=./internal/...)
	@echo "==> Generating HTML coverage report for package: $(pkg)"
	@mkdir -p tmp
	@go test -coverprofile=tmp/coverage.out $(pkg)
	@echo "==> Opening coverage report in browser..."
	@go tool cover -html=tmp/coverage.out

.PHONY: test-bench-pkg
test-bench-pkg: ## Benchmark specific package (Usage: make bench-pkg pkg=./mypackage)
	@echo "==> Benchmarking package: $(pkg)"
	@go test -bench=. -benchmem $(pkg)