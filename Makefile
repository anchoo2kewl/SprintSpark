.PHONY: help dev dev-docker stop clean test lint build install-tools

help: ## Show this help message
	@echo 'SprintSpark - Project Management Tool'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

dev: ## Start both API and Web in development mode (local)
	@echo "Starting SprintSpark in development mode..."
	@echo ""
	@echo "🚀 Starting API server on http://localhost:8080"
	@echo "🚀 Starting Web server on http://localhost:5173"
	@echo ""
	@trap 'kill 0' EXIT; \
		(cd api && make dev) & \
		(cd web && npm run dev) & \
		wait

dev-docker: ## Start both services using Docker Compose
	@echo "Starting SprintSpark with Docker Compose..."
	docker-compose up --build

stop: ## Stop Docker Compose services
	@echo "Stopping services..."
	docker-compose down

clean: ## Clean all build artifacts and caches
	@echo "Cleaning project..."
	cd api && make clean
	cd web && rm -rf node_modules dist
	docker-compose down -v
	@echo "✓ Clean complete"

test: ## Run all tests
	@echo "Running API tests..."
	cd api && make test
	@echo ""
	@echo "Running Web tests..."
	cd web && npm run test

lint: ## Run all linters
	@echo "Linting API..."
	cd api && make lint
	@echo ""
	@echo "Linting Web..."
	cd web && npm run lint

build: ## Build production bundles
	@echo "Building API..."
	cd api && make build
	@echo ""
	@echo "Building Web..."
	cd web && npm run build

install-tools: ## Install development tools
	@echo "Installing API tools..."
	cd api && make install-tools
	@echo ""
	@echo "Installing Web dependencies..."
	cd web && npm install
	@echo ""
	@echo "✓ All tools installed"

setup: ## Initial project setup
	@echo "Setting up SprintSpark..."
	@echo ""
	@echo "1. Creating .env files from examples..."
	@if [ ! -f api/.env ]; then cp api/.env.example api/.env; echo "   ✓ Created api/.env"; fi
	@if [ ! -f web/.env ]; then cp web/.env.example web/.env; echo "   ✓ Created web/.env"; fi
	@echo ""
	@echo "2. Installing dependencies..."
	cd web && npm install
	@echo ""
	@echo "3. Installing development tools..."
	cd api && make install-tools
	@echo ""
	@echo "4. Creating database..."
	cd api && make migrate
	@echo ""
	@echo "✓ Setup complete! Run 'make dev' to start development servers"

db-reset: ## Reset database (WARNING: deletes all data)
	cd api && make db-reset

migrate: ## Run database migrations
	cd api && make migrate
