# IMS Thailand Solution

IMS Thailand is a modular investment-management platform for the Thailand PoC. The repository combines a Go API gateway/modular monolith, a Nuxt 4 frontend, PostgreSQL migrations and seed data, Redis-backed operational capabilities, Mailpit-backed local email testing, and Docker Compose infrastructure.

## Current State

- **Backend:** Go modular monolith under `backend/internal`, served by Chi from `backend/cmd/server`.
- **Frontend:** Nuxt 4 app under `frontend/app`, with thin route shells, feature-owned screens, shared UI, and internal EN/TH/ZH i18n.
- **Database:** PostgreSQL schema managed by SQL migrations in `database/migrations`.
- **Seeds:** Go-side permission catalog upserts plus SQL seed files in `database/seeds`.
- **Local infrastructure:** Docker Compose services for PostgreSQL, Redis, backend, frontend, and Mailpit from `infra/`.
- **API documentation:** Swagger output under `backend/docs`, JSON at `/swagger/doc.json`, and UI at `/swagger/index.html`.
- **Portfolio V2:** Additive, portfolio-code-identity API mounted at `/api/v2/portfolios/*`, documented separately at `/swagger/v2/*` (own Swagger 2.0 spec/basePath). Only the `investment` module exposes V2 routes so far; the V1 API keeps serving unchanged. See [docs/api/portfolio-v2-api-ddd.md](docs/api/portfolio-v2-api-ddd.md).
- **AI assistant:** Optional chat module that uses an LLM provider plus MCP-grounded IMS tools. If provider configuration is missing, `/chat` is disabled while the rest of the API stays up.

## Tech Stack

- **Backend:** Go 1.25, Chi, pgx, Redis, slog, Swagger, Prometheus metrics, MCP Go SDK.
- **Frontend:** Nuxt 4, Vue 3, Pinia, TypeScript, VueUse, Tailwind CSS 4, Day.js, Vitest.
- **Database and infra:** PostgreSQL 16, Redis 7, Docker Compose, Mailpit.
- **API tooling:** `swaggo/swag`, `openapi-typescript`, `openapi-fetch`, Bruno collections, and local MCP tooling.

## Repository Layout

```text
ims-th-solution/
|-- backend/                  # Go API, modules, platform packages, Swagger docs, cmd tools
|-- database/                 # SQL migrations and development/reference seeds
|-- docs/                     # Architecture notes, API contracts, handoffs, runbooks
|-- frontend/                 # Nuxt 4 web application
|-- infra/                    # Docker Compose, Dockerfiles, env templates, MCP config templates
|-- tests/                    # E2E scaffold/reserved test area
|-- tools/                    # Bruno collections and Node MCP server tooling
|-- Makefile                  # Common local development commands
|-- CLAUDE.md                 # AI/dev assistant operating guide
`-- AIREAD.md                 # Business and domain context notes
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.25+
- Node.js 22 recommended for parity with CI/local frontend dependencies
- npm 10+

### First-time setup

Install frontend dependencies:

```bash
cd frontend
npm install
cd ..
```

Start local infrastructure. PostgreSQL is required. Redis is used by the Docker backend profile and by host-run backends only when `RATE_LIMIT_BACKEND=redis`; Mailpit is useful for notification email demos.

```bash
cd infra
docker compose up -d postgres redis mailpit
cd ..
```

Apply schema and seed data:

```bash
make migrate-up
make seed
```

Run the local backend and frontend in separate terminals:

```bash
make dev-backend
make dev-frontend
```

`make dev` starts PostgreSQL and then launches the backend and frontend. It does **not** start every optional service, so start Redis/Mailpit separately when your local config needs them. When running the backend directly on the host, keep rate limiting on the default in-memory backend unless your Redis address is reachable from the host.

### Optional chat assistant setup

The chat endpoint is mounted only when the configured provider can initialize. For Anthropic-backed local testing, set at least:

```bash
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=<your-key>
ANTHROPIC_MODEL=claude-haiku-4-5-20251001
CHAT_MAX_TOKENS_PER_TURN=1024
```

The chat module is read-only by default. Mutating MCP tools remain blocked unless `CHAT_WRITE_ENABLED=true` **and** server-side permission checks pass.

### Local URLs

- Frontend: `http://localhost:3000`
- Backend API root: `http://localhost:8080/api/v1`
- Health check: `http://localhost:8080/health`
- Metrics: `http://localhost:8080/metrics`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Swagger JSON: `http://localhost:8080/swagger/doc.json`
- Mailpit UI: `http://localhost:8025`

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
make e2e-db-setup
make test-e2e
make test-e2e-ci
make test-e2e-backend
make lint

make build
make docker-build
make swagger
make api-client
```

Notes:

- `make test` runs backend unit and integration targets.
- `make test-e2e` runs the Playwright project under `tests/e2e` (requires backend + frontend running against the dedicated `ims_e2e` database — see [IAM E2E Tests](#iam-e2e-tests) below).
- `make api-client` regenerates Swagger first and then regenerates the frontend OpenAPI types.

## Backend Modules

| Module | Status | Notes |
| --- | --- | --- |
| `iam` | Active | Login, refresh, logout, sessions, MFA, admin user management, permission lookups, rate-limit middleware integration. |
| `audit` | Active | Audit recorder plus admin/user-facing audit list and export routes. |
| `workflow` | Active | Per-contract business-day state machine, transition history, scheduler, daily workflow endpoints, and contract catalog integration. |
| `compliance` | Active | IRG rule registry, pre/post-trade checks, breaches, overrides, and rule instance administration. |
| `investment` | Active | Funds, portfolios, instruments, ledger posting/reversal, prices, valuations, holdings, cash, NAV/AUM, decisions, executions, confirmations, and research reports. |
| `market_data` | Active | Quote/history provider integration, import batches, provider health, PostgreSQL persistence, and quote cache/provider ports. |
| `reference_data` | Active | Canonical securities, security mappings, unmapped-candidate review, and resolver ports used by market data/watchlist. |
| `approval` | Active | Approval runtime inbox/requests/tasks, configurable groups/teams/processes, subject callbacks, and subject access ports. |
| `permissions` | Active | Permission change requests, roles, users/groups, effective permissions, data rights, labels, checks, approval settings, notification settings, and audit routes. |
| `notification` | Active | In-app notifications, Mailpit/Gmail-ready SMTP outbox, email health/test/retry APIs, worker, approval/workflow/watchlist notifiers. |
| `watchlist` | Active | Personal/portfolio watchlist items, market-price threshold rules, alert events, acknowledgement, manual evaluation, notification integration. |
| `chat` | Active/optional | SSE chat endpoint, persisted sessions/messages/tool calls, Anthropic adapter, MCP client, read-only IMS tool path, audit/provenance hooks. |
| `integration` | Active | Dashboard snapshot and personal task summary endpoints. |

The current working tree does not contain `backend/internal/leave_delegation`. If a build references that package, check for stale imports before treating the module as implemented. Some frontend leave routes may still be route shells/placeholders unless a backend module is added later.

## API Surface

Public or unauthenticated routes:

- `GET /health`
- `GET /metrics`
- `GET /swagger/*` and `GET /swagger/v2/*`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

Authenticated API groups:

- `/api/v1/auth/*` for profile, logout, password, MFA, and session operations.
- `/api/v1/admin/*` for IAM admin operations and admin audit list/export surfaces.
- `/api/v1/audit/*` for permission-module audit logs and exports.
- `/api/v1/workflow/*` for day-state transitions, scheduler runs, daily workflow, transition rules, and settings.
- `/api/v1/compliance/*` for checks, breaches, overrides, and rule instances.
- `/api/v1/investment/*` for reference data, funds, portfolios, instruments, ledger, valuation, NAV/AUM, decisions, executions, confirmations, and research reports.
- `/api/v1/market-data/*` for quotes, history, imports/import batches, provider health, and market-data screen feeds.
- `/api/v1/reference-data/*` for canonical securities, mappings, and unmapped-candidate workflows.
- `/api/v1/integration/*` for dashboard and personal task snapshots.
- `/api/v1/permissions/*` for permission governance, effective access, roles/groups/users, settings, labels, and checks.
- `/api/v1/approvals/*` for approval runtime inbox, requests, tasks, timelines, and subject status.
- `/api/v1/approval-config/*` for approval groups, teams, processes, memberships, and activation state.
- `/api/v1/notifications/*` for notification center, email outbox, health, test email, and retry operations.
- `/api/v1/watchlists/*` for watchlist items, alerts, acknowledgement, and manual evaluation.
- `/api/v1/chat*` for streaming chat and chat session/message history when the chat module is configured.
- `/api/v2/portfolios/*` for the additive Portfolio V2 portfolio-code API (investment module only, documented at `/swagger/v2/*`).

Most non-auth routes are permission-gated by backend middleware. Frontend guards are for UX only; backend permissions and data-scope checks are the source of truth.

## Frontend Notes

Nuxt is configured with `srcDir: "app/"`.

- `frontend/app/features/` owns feature-specific UI, composables, services, and helpers.
- `frontend/app/shared/ui/` owns reusable UI primitives.
- `frontend/app/shared/i18n/` owns the internal translation registry and messages for `en`, `th`, and `zh`.
- `frontend/app/pages/` contains file-based route shells; keep new route files thin wherever practical.
- `frontend/app/api/openapi.ts` provides the typed OpenAPI client backed by generated `frontend/app/api/ims-api.d.ts`.

> [!IMPORTANT]
> Do **not** write API types or fetch calls manually by hand for Swagger-backed endpoints. Regenerate the schema with `make api-client` and consume routes through the typed client from `ims-api.d.ts`.

See [frontend/README.md](frontend/README.md) for frontend-specific conventions.

## Database Notes

- Migration files live in [database/migrations](database/migrations).
- Development seed SQL lives in [database/seeds](database/seeds).
- `make seed` first upserts module-owned permission catalog entries from Go, then applies SQL seed files.
- Seeds include workflow settings/contracts, IAM/permission fixtures, approval/notification/watchlist permissions, investment process assignments, investment reference data, and market/reference data fixtures as available.

See [database/migrations/README.md](database/migrations/README.md) and [database/seeds/README.md](database/seeds/README.md) for detailed workflows.

## IAM E2E Tests

End-to-end tests for the IAM module (authentication, RBAC/function permissions, data permissions, audit trail) run against a **dedicated `ims_e2e` database** — never `ims_dev` or production. They prove IAM works from the browser through the API to the database: backend authorization (401/403) is asserted directly against the API, not just inferred from hidden frontend menus.

Two suites cover it:

- **Playwright** (`tests/e2e/specs/iam/*.spec.ts`) — browser-driven UI flows plus direct API assertions.
- **Go** (`backend/tests/e2e/iam_test.go`, build tag `e2e`) — backend-only proof that authorization is enforced independent of any frontend.

### Local setup

```bash
cp infra/env/.env.e2e.example infra/env/.env.e2e   # values are synthetic; safe defaults out of the box
make e2e-db-setup                                   # creates/migrates/seeds a dedicated ims_e2e database (idempotent)
```

`make e2e-db-setup` creates the `ims_e2e` database on the same local Postgres container used for dev, runs migrations, runs the standard `make seed` (permission catalog + reference/demo data), then runs `backend/cmd/seed-e2e`, which seeds 7 deterministic fixtures: `e2e_admin`, `e2e_manager`, `e2e_trader`, `e2e_auditor`, `e2e_disabled`, `e2e_locked`, `e2e_no_permission` (password: `E2E_USER_PASSWORD` from `infra/env/.env.e2e`). `seed-e2e` refuses to run against anything that isn't clearly the E2E database (`APP_ENV=test` and `DB_NAME` containing `e2e`).

Then, in separate terminals with `infra/env/.env.e2e` sourced into the environment:

```bash
cd backend && go run cmd/server/main.go     # backend against ims_e2e
cd frontend && npm run dev                  # frontend, NUXT_PUBLIC_API_BASE_URL pointed at that backend
```

Run the suites:

```bash
make test-e2e-backend   # Go, backend-only, no browser needed
make test-e2e           # Playwright, requires the backend + frontend above to be running
make test-e2e:ui        # Playwright UI mode, for authoring/debugging (cd tests/e2e && npm run test:e2e:ui)
```

Coverage, known gaps, and flaky-risk areas are documented in [tests/e2e/COVERAGE.md](tests/e2e/COVERAGE.md).

### CI

The `e2e-iam` job in `.github/workflows/ci.yml` runs on every PR: a `postgres:16-alpine` service container stands in for `ims_e2e`, migrations/seeds run via env vars (no docker compose in CI), the backend and frontend start in the background, then both suites run. A Playwright HTML report is uploaded as a build artifact on every run (pass or fail).

## Configuration Notes

Backend configuration is read from environment variables, with local defaults for development. Keep secrets in gitignored env files or an external secret store.

Important local configuration areas:

- Core app/database: `APP_ENV`, `APP_PORT`, `APP_JWT_SECRET`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`.
- Reporting currency: `REPORTING_CURRENCY` (e.g. `THB`) — required in **every** environment, including development, test, and CI; an uppercase, recognized ISO 4217 code. Unlike `APP_JWT_SECRET`/`ALPHA_VANTAGE_API_KEY`, there is no development/test default — omitting it fails config load.
- Redis/rate limiting: `RATE_LIMIT_BACKEND`, `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`.
- Market data: `MARKET_DATA_PROVIDER` or `MARKET_DATA_PRIMARY_PROVIDER`, `MARKET_DATA_FALLBACK_PROVIDER`, `ALPHA_VANTAGE_API_KEY`, quote staleness/refresh variables.
- Notification email: `NOTIFICATION_EMAIL_ENABLED`, `NOTIFICATION_EMAIL_WORKER_ENABLED`, `SMTP_HOST`, `SMTP_PORT`, `SMTP_TLS_MODE`, `SMTP_FROM_ADDRESS`, `APP_PUBLIC_BASE_URL`.
- Chat/MCP: `LLM_PROVIDER`, `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL`, `CHAT_WRITE_ENABLED`, `CHAT_MCP_SERVERS_CONFIG`, `CHAT_MCP_IMS_BIN`, `IMS_API_BASE_URL`.
- Frontend runtime: `NUXT_PUBLIC_API_BASE_URL`, `NUXT_API_BASE_URL`, `NUXT_PUBLIC_APP_NAME`.

Useful templates live under [infra/env](infra/env) and [infra/config](infra/config).

## Tooling And Docs

- Bruno API collections live under [tools/bruno](tools/bruno).
- Node MCP tooling lives under [tools/mcp/ims-mcp-server](tools/mcp/ims-mcp-server).
- Go IMS MCP server source lives under `backend/cmd/ims-mcp`.
- `make swagger` regenerates backend Swagger docs under `backend/docs`.
- `make api-client` regenerates frontend OpenAPI types from `backend/docs/swagger.json` to `frontend/app/api/ims-api.d.ts`.
- Architecture, API contracts, module handoffs, and runbooks live under [docs](docs).

Generated files should not be edited manually:

- `backend/docs/docs.go`
- `backend/docs/swagger.json`
- `backend/docs/swagger.yaml`
- `frontend/app/api/ims-api.d.ts`

## License

MIT
