.PHONY: run test lint fmt migrate db-reset clean help openapi-check

# Variables
DB_PATH ?= ./data/sprintspark.db
MIGRATIONS_PATH ?= ./internal/db/migrations

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

run: ## Start development server
	@echo "Starting SprintSpark API..."
	go run cmd/api/main.go

test: ## Run tests with coverage
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "\nCoverage summary:"
	go tool cover -func=coverage.out | tail -1

lint: ## Run linters
	@echo "Running linters..."
	go fmt ./...
	go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

migrate: ## Run database migrations
	@echo "Running migrations..."
	@echo "Migrations will be applied automatically when the server starts"
	@echo "Or you can start the server briefly to run migrations:"
	@echo "  DB_PATH=$(DB_PATH) go run cmd/api/main.go"
	@echo ""
	@echo "To run migrations now:"
	@DB_PATH=$(DB_PATH) timeout 2s go run cmd/api/main.go 2>&1 | grep -i migration || true
	@if [ -f "$(DB_PATH)" ]; then \
		echo "✓ Database created at $(DB_PATH)"; \
		sqlite3 $(DB_PATH) "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;" | sed 's/^/  - /'; \
	fi

db-reset: ## Reset database (WARNING: deletes all data)
	@echo "Resetting database..."
	@read -p "This will delete all data. Continue? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		rm -f $(DB_PATH) $(DB_PATH)-shm $(DB_PATH)-wal; \
		echo "Database deleted. Run 'make migrate' to recreate."; \
	else \
		echo "Cancelled."; \
	fi

clean: ## Clean build artifacts and cache
	@echo "Cleaning..."
	go clean -cache -testcache
	rm -f coverage.out
	rm -f tmp/*

dev: ## Run with auto-reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/air-verse/air@latest"; \
		echo "Falling back to regular run..."; \
		make run; \
	fi

build: ## Build production binary
	@echo "Building..."
	go build -o bin/sprintspark cmd/api/main.go

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Done! Tools installed to $(shell go env GOPATH)/bin"

openapi-check: ## Verify OpenAPI spec covers all routes
	go test -run TestOpenAPIRoutesCoverage ./internal/api/

openapi-validate: ## Validate OpenAPI specification
	@echo "Validating OpenAPI specification..."
	@if command -v docker > /dev/null; then \
		docker run --rm -v $(PWD):/spec redocly/cli lint /spec/openapi.yaml; \
	else \
		echo "Docker not available. Install Docker or use: npx @redocly/cli lint openapi.yaml"; \
	fi

openapi-docs: ## Generate OpenAPI documentation preview
	@echo "Generating OpenAPI documentation..."
	@echo "Open http://localhost:8081 in your browser"
	@docker run --rm -p 8081:8080 -v $(PWD)/openapi.yaml:/usr/share/nginx/html/openapi.yaml \
		-e SPEC_URL=openapi.yaml redocly/redoc
