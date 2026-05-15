# IMS Thailand Solution

IMS Thailand is a modular investment management platform for the Thailand PoC. The current codebase combines a Go API gateway, a Nuxt 4 frontend, PostgreSQL migrations and seeds, Redis-backed operational concerns, and local Docker Compose infrastructure.

## Current State

- Backend: Go modular monolith under `backend/internal`, served by Chi from `backend/cmd/server`.
- Frontend: Nuxt 4 app under `frontend/app`, with feature-owned screens and shared UI/i18n helpers.
- Database: PostgreSQL schema managed by SQL migrations in `database/migrations`.
- Seeds: Go-side permission catalog upserts plus SQL seed files in `database/seeds`.
- Local infrastructure: Docker Compose services for PostgreSQL, Redis, backend, and frontend from `infra/`.
- API documentation: Swagger output under `backend/docs` and UI at `/swagger/index.html`.

## Tech Stack

- Backend: Go 1.25, Chi, pgx, Redis, slog, Swagger.
- Frontend: Nuxt 4, Vue 3, Pinia, TypeScript, VueUse, Tailwind CSS 4, Vitest.
- Database and infra: PostgreSQL 16, Redis 7, Docker Compose.
- API tooling: openapi-typescript, openapi-fetch, Bruno collections, and a local MCP server.

## Repository Layout

```text
ims-th-solution/
|-- backend/                  # Go API, modules, platform packages, Swagger docs
|-- database/                 # SQL migrations and development/reference seeds
|-- docs/                     # Runbooks
|-- frontend/                 # Nuxt 4 web application
|-- infra/                    # Docker Compose, Dockerfiles, environment templates
|-- tests/                    # Playwright E2E scaffold
|-- tools/                    # Bruno collections and MCP server
|-- Makefile                  # Common local development commands
|-- CLAUDE.md                 # AI/dev assistant operating guide
`-- AIREAD.md                 # Business and domain context notes
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.25+
- Node.js 22 recommended for parity with CI
- npm 10+

### First-time setup

```bash
cd frontend
npm install
cd ..
```

Start PostgreSQL and Redis:

```bash
cd infra
docker compose up -d postgres redis
cd ..
```

Apply schema and seed data:

```bash
make migrate-up
make seed
```

Run the local backend and frontend:

```bash
make dev-backend
make dev-frontend
```

`make dev` also starts PostgreSQL and then launches the backend and frontend. If your local backend environment uses `RATE_LIMIT_BACKEND=redis`, start Redis first as shown above.

### Local URLs

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Health check: `http://localhost:8080/health`
- Swagger UI: `http://localhost:8080/swagger/index.html`

## Common Commands

```bash
make dev
make dev-backend
make dev-frontend

make migrate-up
make migrate-down
make migrate-new module=iam name=add_mfa_fields
make seed
make db-reset
make contract-check

make test
make test-unit
make test-integration
make test-e2e
make lint

make build
make docker-build
make swagger
make api-client
```

## Backend Modules

| Module | Current status | Notes |
| --- | --- | --- |
| `iam` | Active | Login, refresh, logout, sessions, MFA, admin user management, permission lookups, route middleware integration. |
| `audit` | Active | Audit recorder plus admin audit list/export routes mounted through IAM admin routing. |
| `workflow` | Active | Per-contract business-day state machine, transition history, scheduler, and workflow state provider contract. |
| `compliance` | Active | IRG rule registry, pre/post-trade checks, breaches, overrides, and rule instance administration. |
| `investment` | Active | Funds, portfolios, instruments, ledger posting/reversal, prices, valuation, holdings, cash, and AUM endpoints. |
| `market_data` | Active | Quote/history provider integration, import endpoint, provider health, PostgreSQL persistence, Redis quote cache. |
| `approval` | Scaffold | Module boundary and permission catalog only; routes are TODO. |
| `integration` | Scaffold | Module boundary only; external feed endpoints are represented in tools/docs, not wired in the server. |
| `notification` | Scaffold | Module boundary and permission catalog only. |
| `permissions` | Scaffold | Permission storage exists in migrations/seeds; operator-facing admin flows currently live through IAM/settings. |
| `reference_data` | Scaffold | Boundary reserved for currencies, markets, instruments, and Thai holidays. |

The current working tree does not contain `backend/internal/leave_delegation`. If a build references that package, check for stale imports before treating the module as implemented.

## API Surface

Public routes:

- `GET /health`
- `GET /swagger/*`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

Authenticated API groups:

- `/api/v1/auth/*` for profile, logout, password, MFA, and session operations.
- `/api/v1/admin/*` for IAM admin and audit list/export operations.
- `/api/v1/workflow/*` for day-state transitions, history, and scheduler runs.
- `/api/v1/compliance/*` for checks, breaches, overrides, and rules.
- `/api/v1/investment/*` for reference data, funds, portfolios, instruments, ledger, price, valuation, and AUM workflows.
- `/api/v1/market-data/*` for quotes, history, import, and provider health.

Most non-auth routes are permission-gated by backend middleware. Frontend guards are for UX only.

## Frontend Notes

Nuxt is configured with `srcDir: "app/"`.

- `frontend/app/features/` owns feature-specific UI and service clients.
- `frontend/app/shared/ui/` owns reusable UI primitives.
- `frontend/app/shared/i18n/` owns the internal translation registry and messages for `en`, `th`, and `zh`.
- `frontend/app/pages/` contains route shells and file-based routes.
- `frontend/app/api/openapi.ts` provides the typed OpenAPI client backed by generated `frontend/app/api/ims-api.d.ts`.

See [frontend/README.md](C:/Users/kanta/source/repos/ims-th-solution/frontend/README.md) for frontend-specific conventions.

## Database Notes

- Migration files live in [database/migrations](C:/Users/kanta/source/repos/ims-th-solution/database/migrations).
- Development seed SQL lives in [database/seeds](C:/Users/kanta/source/repos/ims-th-solution/database/seeds).
- `make seed` first upserts module-owned permission catalog entries from Go, then applies SQL seed files.
- Seeds include workflow settings/contracts, investment process assignment fixtures, investment permissions, and investment reference data.

See [database/migrations/README.md](C:/Users/kanta/source/repos/ims-th-solution/database/migrations/README.md) and [database/seeds/README.md](C:/Users/kanta/source/repos/ims-th-solution/database/seeds/README.md) for the detailed workflows.

## Tooling

- Bruno API collections live under [tools/bruno](C:/Users/kanta/source/repos/ims-th-solution/tools/bruno).
- The local MCP server lives under [tools/mcp/ims-mcp-server](C:/Users/kanta/source/repos/ims-th-solution/tools/mcp/ims-mcp-server).
- `make swagger` regenerates backend Swagger docs.
- `make api-client` regenerates frontend OpenAPI types from `backend/docs/swagger.json`.

## License

MIT
