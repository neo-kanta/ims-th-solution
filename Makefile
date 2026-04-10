# IMS Thailand — Project Makefile
# Quick commands for development, testing, and deployment.

.PHONY: dev dev-backend dev-frontend migrate-up migrate-down migrate-new seed db-reset \
        test test-unit test-integration test-e2e lint build docker-build swagger

# =============================
# Development
# =============================

dev: ## Start everything (docker-compose up + frontend dev)
	cd infra && docker compose up -d postgres
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	$(MAKE) dev-backend &
	$(MAKE) dev-frontend

dev-backend: ## Backend only
	cd backend && go run cmd/server/main.go

dev-frontend: ## Frontend only
	cd frontend && npm run dev

# =============================
# Database
# =============================

migrate-up: ## Run pending migrations
	cd backend && go run cmd/migrate/main.go up

migrate-down: ## Rollback last migration
	cd backend && go run cmd/migrate/main.go down

migrate-new: ## Create new migration pair (usage: make migrate-new module=workflow name=add_table)
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	touch database/migrations/$${timestamp}_$(module)__$(name).up.sql; \
	touch database/migrations/$${timestamp}_$(module)__$(name).down.sql; \
	echo "Created: database/migrations/$${timestamp}_$(module)__$(name).{up,down}.sql"

seed: ## Load seed data
	cd backend && go run cmd/seed/main.go

db-reset: ## Drop + recreate + migrate + seed
	@echo "Resetting database..."
	cd infra && docker compose stop postgres
	cd infra && docker compose rm -f postgres
	docker volume rm ims-th-solution_pgdata || true
	cd infra && docker compose up -d postgres
	@sleep 3
	$(MAKE) migrate-up
	$(MAKE) seed

# =============================
# Testing
# =============================

test: ## Run all tests
	$(MAKE) test-unit
	$(MAKE) test-integration

test-unit: ## Unit tests only (backend)
	cd backend && go test -v -short ./...

test-integration: ## Integration tests
	cd backend && go test -v -run Integration ./...

test-e2e: ## Playwright E2E tests
	cd tests/e2e && npx playwright test

lint: ## Lint both frontend and backend
	cd backend && go vet ./...
	cd frontend && npx nuxi typecheck

# =============================
# Build
# =============================

build: ## Build backend binary + frontend static
	cd backend && CGO_ENABLED=0 go build -o bin/server ./cmd/server
	cd frontend && npm run build

docker-build: ## Build Docker images
	cd infra && docker compose build

swagger: ## Generate backend Swagger docs
	cd backend && swag init -g main.go -d cmd/server,internal/iam -o docs --parseInternal --useStructName

# =============================
# Help
# =============================

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
