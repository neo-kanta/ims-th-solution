# IMS Thailand — Project Makefile
# Quick commands for development, testing, and deployment.

.PHONY: dev dev-backend dev-frontend migrate-up migrate-down migrate-new seed db-reset \
        contract-check test test-unit test-integration test-e2e test-e2e-backend test-e2e-ci \
        e2e-db-setup lint build docker-build swagger api-client

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
	cd backend && go run ./cmd/seed

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
# Contracts
# =============================

contract-check: ## Verify cross-module contract bindings against the running database
	cd backend && go run cmd/contract-check/main.go

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

e2e-db-setup: ## Create/migrate/seed the dedicated ims_e2e database (safe to rerun)
	bash tests/e2e/scripts/setup-db.sh

test-e2e: ## Playwright E2E tests (requires backend+frontend running against ims_e2e)
	cd tests/e2e && npm run test:e2e

test-e2e-ci: ## Playwright E2E tests, CI mode (retries + CI reporter)
	cd tests/e2e && npm run test:e2e:ci

test-e2e-backend: ## Backend-only Go E2E tests (401/403 proof; build tag e2e)
	# Scoped to TestE2E_IAM_* — TestE2E_InvestmentOversellEnvelope in this
	# same package has pre-existing, unrelated bugs (see tests/e2e/COVERAGE.md)
	# uncovered while wiring this target up; fix tracked separately.
	cd backend && go test -tags e2e -count=1 -v -run TestE2E_IAM ./tests/e2e/...

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

swagger: ## Generate backend Swagger docs (V1, basePath /api/v1)
	cd backend && go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/server/main.go -d . -o docs --parseInternal --useStructName --tags "!Investment - Portfolios V2"

swagger-v2: ## Generate Portfolio V2 Swagger docs (basePath /api/v2 — see backend/cmd/server/swagger_v2_docs.go)
	cd backend && go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/server/swagger_v2_docs.go -d . -o docs/v2 --parseInternal --useStructName --tags "Investment - Portfolios V2" --instanceName=v2

api-client: swagger swagger-v2 ## Generate frontend API types from backend Swagger docs
	cd frontend && npm run api:generate

# =============================
# Help
# =============================
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
