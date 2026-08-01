.PHONY: help run build test clean setup install-deps swagger fmt hooks lint lint-code lint-arch lint-fix check

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: install-deps hooks ## Install dependencies and enable git hooks
	@echo "Dependencies installed"

install-deps: ## Install Go dependencies
	go mod download
	go mod tidy

run: ## Run the server in development mode
	go run ./cmd/server -config config.yaml -env development

build: ## Build the binary
	go build -o bin/golang-service-template ./cmd/server

test: ## Run tests
	go test -count=1 ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf tmp/

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@if ! command -v swag &> /dev/null; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g cmd/server/main.go -o swagger --parseInternal --parseDependency

lint: ## Run all linters (code quality + DDD architectural rules)
	@echo "Running all linting checks..."
	@$(MAKE) lint-code; CODE_QUALITY=$$?; \
	$(MAKE) lint-arch; ARCH_QUALITY=$$?; \
	if [ $$CODE_QUALITY -ne 0 ] || [ $$ARCH_QUALITY -ne 0 ]; then \
		echo "Linting checks failed"; \
		exit 1; \
	else \
		echo "All linting checks passed"; \
	fi

lint-code: ## Run golangci-lint (code quality checks)
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0; \
	fi
	@echo "Running code quality checks (golangci-lint)..."
	@golangci-lint run --timeout=5m

lint-arch: ## Run go-arch-lint (enforces DDD architectural rules)
	@if ! command -v go-arch-lint &> /dev/null; then \
		echo "Installing go-arch-lint..."; \
		go install github.com/fe3dback/go-arch-lint@latest; \
	fi
	@echo "Running DDD architectural checks (go-arch-lint)..."
	@go-arch-lint check --project-path .

lint-fix: ## Run linter with auto-fix
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.8.0; \
	fi
	@golangci-lint run --fix

fmt: ## Format Go sources (gofmt)
	gofmt -w .

hooks: ## Enable repo git hooks (runs make fmt on commit)
	git config core.hooksPath .githooks
	@echo "core.hooksPath set to .githooks"

check: lint test ## Run all checks (lint + test)
