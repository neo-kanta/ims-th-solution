# CLAUDE.md - Instructions for AI/dev assistants

Read `AIREAD.md` when you need business context, domain terminology, or project rationale. This file is the current operational guide for working in the codebase.

## Project Snapshot

- Name: IMS Thailand Solution
- Repo: `ims-th-solution`
- License: MIT
- Backend: Go 1.25 modular monolith, Chi router, pgx, Redis, Swagger
- Frontend: Nuxt 4, Vue 3, Pinia, TypeScript, VueUse, Tailwind CSS 4
- Database: PostgreSQL migrations and seeds
- Local infra: Docker Compose with PostgreSQL, Redis, backend, and frontend services

## Quick Commands

Run these from the repository root unless noted.

```bash
# Development
make dev
make dev-backend
make dev-frontend

# Database
make migrate-up
make migrate-down
make migrate-new module=workflow name=add_carry_forward
make seed
make db-reset
make contract-check

# Tests and quality
make test
make test-unit
make test-integration
make test-e2e
make lint

# Build and generated clients
make build
make docker-build
make swagger
make api-client
```

Useful direct commands:

```bash
cd infra && docker compose up -d postgres redis
cd backend && go test ./...
cd frontend && npm run test
cd frontend && npm run build
```

`make dev` starts PostgreSQL and then launches local backend/frontend processes. If the local backend `.env` uses `RATE_LIMIT_BACKEND=redis`, start Redis first.

## Current Runtime Wiring

`backend/cmd/server/main.go` is the source of truth for mounted modules:

1. Load config and logging.
2. Connect PostgreSQL; run secure bootstrap check.
3. Optionally connect Redis when `RATE_LIMIT_BACKEND=redis`.
4. Wire audit, IAM, compliance, workflow, investment, reference data, market data, integration, permissions, approval, notification, and chat modules.
5. Wire cross-module callbacks: investment ↔ approval subject callbacks, workflow confirmation gate, contract catalog, and market-data quote provider.
6. Wire notification as the workflow stuck-day operator notifier.
7. Start the workflow scheduler on an hourly loop.
8. Mount `/health`, `/metrics`, `/swagger/*`, and `/api/v1`.

Public API:

- `GET /health`
- `GET /metrics` (Prometheus scrape endpoint)
- `GET /swagger/*`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

Authenticated API groups:

- IAM/auth/session/MFA/admin under `/api/v1/auth/*` and `/api/v1/admin/*`
- Audit admin list/export mounted by IAM under `/api/v1/admin/audit`
- Workflow under `/api/v1/workflow/*`
- Compliance under `/api/v1/compliance/*`
- Investment under `/api/v1/investment/*`
- Market data under `/api/v1/market-data/*`
- Reference data under `/api/v1/reference-data/*`
- Integration (user dashboard/tasks) under `/api/v1/integration/*`
- Permissions under `/api/v1/permissions/*`
- Approval under `/api/v1/approval/*`
- Notification under `/api/v1/notifications/*`
- Chat (conditionally mounted when LLM provider is configured) under `/api/v1/chat/*`

## Module Status

| Module           | Status | Notes                                                                                                      |
| ---------------- | ------ | ---------------------------------------------------------------------------------------------------------- |
| `iam`            | Active | Auth, sessions, MFA, admin user operations, permission and data-scope checks.                              |
| `audit`          | Active | Audit recorder plus admin audit list/export routes.                                                        |
| `workflow`       | Active | Business-day state machine, transition history, scheduler, workflow state contract.                        |
| `compliance`     | Active | IRG rule registry, checks, breaches, overrides, rule instances, contract adapter.                          |
| `investment`     | Active | Funds, portfolios, instruments, ledger, price snapshots, valuations, AUM, holdings and cash reads.         |
| `market_data`    | Active | Quote/history providers, import, provider health, Redis cache, PostgreSQL persistence.                     |
| `approval`       | Active | Approval flows, groups, teams, inbox, subject callbacks for research reports, decisions, and portfolios.   |
| `permissions`    | Active | Account/group management, function permissions, data permissions, effective-permissions view, role hierarchy.|
| `notification`   | Active | Notification engine; provides `ApprovalNotifier` and workflow stuck-day `OperatorNotifier`. No UI nav.     |
| `reference_data` | Active | Thai market holidays, currencies, markets, instruments; provides `Resolver()` to market data. No UI nav.   |
| `integration`    | Active | User dashboard snapshot and task summary endpoints (`/integration/dashboard/me`, `/integration/tasks/my`). |
| `chat`           | Active | AI financial assistant via Anthropic LLM + MCP. Conditionally mounted; rest of API unaffected if disabled. |

## Backend Placement Rules

This is a modular monolith. Start by choosing the owning module under `backend/internal/<module>/`.

```text
backend/internal/<module>/
  domain/entity/              business entities
  domain/valueobject/         immutable value types
  domain/event/               domain events
  domain/policy/              pure business rules
  domain/repository.go        repository interfaces
  application/command/        write use cases
  application/query/          read use cases
  application/service/        application services
  infrastructure/persistence/ SQL-backed repositories
  infrastructure/adapter/     external or cross-module adapters
  transport/handler/          HTTP handlers
  transport/dto/request/      request structs
  transport/dto/response/     response structs
  transport/router.go         route definitions
  jobs/                       scheduled/background tasks
  permission/policies.go      permission catalog provider
  module.go                   dependency wiring and route registration
```

Shared backend code belongs in:

```text
backend/platform/       config, database, middleware, logging, errors, clock, health, validation
backend/pkg/types/      shared value types
backend/pkg/enum/       shared enums
backend/pkg/contract/   cross-module interfaces only
```

Backend rules:

- Do not import another module's `internal` package. Use `backend/pkg/contract` interfaces or adapters.
- Keep business logic out of HTTP handlers. Handlers parse, call application code, and format responses.
- Keep SQL in `infrastructure/persistence` or migration/seed files.
- Keep transaction boundaries in application command handlers or repository helpers, not transport code.
- Enforce permissions on the backend with middleware or application checks. Frontend route guards are UX only.
- Use DTOs at transport boundaries; do not expose domain entities directly from handlers.
- Keep timestamps UTC in backend storage and logic; format for `Asia/Bangkok` in the frontend.
- Every `.up.sql` migration needs a matching `.down.sql`.

## Frontend Placement Rules

Nuxt uses `srcDir: "app/"`, so all app code is under `frontend/app`.

```text
frontend/app/
  api/                 generated OpenAPI types and typed client helpers
  assets/css/          global CSS
  composables/         thin Nuxt composables such as useApi/useI18n
  features/            feature-owned components, services, types, helpers
  layouts/             default, auth, dashboard layouts
  middleware/          auth and permission route middleware
  pages/               file-based route shells
  shared/i18n/         internal translations and formatting helpers
  shared/routing/      route meta helpers
  shared/ui/           shared UI primitives
  stores/              cross-feature Pinia stores
  types/               app-level shared types
```

Frontend rules:

- Do not create new `frontend/modules/*` code; that is outdated guidance.
- Keep route files thin and move real UI or service logic into `app/features/<feature>`.
- Put reusable primitives in `app/shared/ui` only when they are truly shared.
- Use `useApi()` or `useOpenApiClient()` for backend calls so auth and base URLs stay centralized.
- Put user-facing copy in `app/shared/i18n/messages/{en,th,zh}` and consume it through `useI18n().t(...)`.
- Use `definePageMeta` plus `auth`/`permission` middleware for protected pages.

## Dashboard Tab Registry

The dashboard header tabs are driven by a provider registry — `dashboard.vue` never needs to change when adding tabs for a new feature.

- Registry file: `frontend/app/features/shell/tabs/dashboardTabRegistry.ts`
- Each entry implements `DashboardTabProvider` (type in `frontend/app/features/shell/tabs/types.ts`)
- To add tabs: create a `navigation/dashboardTabs.ts` in your feature and append it to the registry array.

## MCP Server (tools/mcp)

`tools/mcp/ims-mcp-server/` is a standalone Node.js stdio MCP server that exposes IMS read-only (and optionally mutating) tools to the chat module.

- Built separately; output binary is `backend/ims-mcp.exe` (compiled alongside the server binary).
- The server binary auto-discovers it via `filepath.Dir(executable) + "/ims-mcp"`.
- An override path can be set with `CHAT_MCP_IMS_BIN`.

## Database And Seeds

Migrations live in `database/migrations` and use:

```text
<timestamp>_<module>__<description>.up.sql
<timestamp>_<module>__<description>.down.sql
```

Seed behavior:

- `make seed` runs `backend/cmd/seed/main.go`.
- The seeder upserts Go permission catalogs first.
- SQL files under `database/seeds` then run in lexicographic order.
- Investment reference seeds live under `database/seeds/investment`.
- Keep seeds idempotent with `ON CONFLICT`.

## API And Generated Types

- Backend Swagger source is generated into `backend/docs`.
- `make swagger` refreshes Swagger.
- `make api-client` runs Swagger generation and then `frontend/scripts/generate-openapi-types.mjs`.
- Generated frontend OpenAPI types are written to `frontend/app/api/ims-api.d.ts` (Absolute path: `C:\Users\kanta\source\repos\ims-th-solution\frontend\app\api\ims-api.d.ts`).
- Typed frontend access should go through `frontend/app/api/openapi.ts`.
- **CRITICAL**: **DO NOT write API types, requests/responses, or client fetch code manually by hand.** All endpoints must rely on the generated typed client from `ims-api.d.ts`.

## Environment Notes

- Compose env files live under `infra/env`.
- Local backend runs from `backend` and can load `backend/.env` if present.
- Do not commit real secrets.
- `APP_JWT_SECRET`, `MFA_ENCRYPTION_KEY`, and `ALPHA_VANTAGE_API_KEY` are required outside development/test according to `backend/platform/config/config.go`. `ANTHROPIC_API_KEY` is also required when the chat module is active — `LLM_PROVIDER` defaults to `anthropic`, so the key is effectively required in most deployments unless you explicitly set a different provider.
- Redis is optional only when `RATE_LIMIT_BACKEND=memory`; if set to `redis`, backend startup requires a reachable Redis instance.
- Chat module: `LLM_PROVIDER` defaults to `anthropic`; without a matching `ANTHROPIC_API_KEY` the `/chat` route is silently not mounted — no startup error, but all chat endpoints 404.
- `CHAT_WRITE_ENABLED=true` enables mutating MCP tools; default is read-only.
- The chat module spawns `ims-mcp` (compiled to `backend/ims-mcp.exe`) as a stdio MCP server. `IMS_API_BASE_URL` is forwarded to it.

## Git And Review Hygiene

- Keep changes scoped to the requested behavior.
- Do not revert user changes or unrelated dirty files.
- Prefer small, focused tests near the layer being changed.
- For backend changes, run `go test ./...` from `backend` when feasible.
- For frontend changes, run `npm run test` or `npm run build` from `frontend` when relevant.
- For docs-only changes, no build is required, but call out any known compile blockers you discover.

## Common Mistakes To Avoid

- Using outdated Nuxt 3 or `frontend/modules` instructions.
- Adding a module import from another module's `internal` package.
- Adding SQL directly in handlers.
- Trusting frontend permissions without backend enforcement.
- Forgetting to update the permission catalog provider and SQL grants together.
- Adding migration `up` files without matching `down` files.
- Regenerating OpenAPI types without first refreshing Swagger when backend routes changed.
- Writing custom frontend API fetch code, mock fetch utilities, or manual type definitions instead of using the generated types in `ims-api.d.ts`.
