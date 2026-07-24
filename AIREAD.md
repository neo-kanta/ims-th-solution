# AIREAD.md — Project Context for AI Assistants

> **Purpose:** This file gives AI assistants (Claude, Copilot, Cursor, etc.)
> the full context needed to work safely and correctly in this repository.
> Read this file FIRST before touching any code.

---

## 1. What Is This Project?

**IMS (Investment Management System)** — a Proof of Concept for an asset management company in Thailand.

It digitizes the core investment lifecycle:

```
Investment Research → Investment Decision → Investment Execution → Investment Review
```

The system enforces pre-trade compliance, approval workflows, delegation rules,
and permission-based access control across all operations.

**This is NOT:**

- An accounting system (PAM)
- An order management system (OMS)
- A portfolio accounting module
- A settlement or custodian system

Those are treated as **external systems** with clean integration points.

---

## 2. Architecture Overview

| Layer     | Technology                       | Notes                             |
| --------- | -------------------------------- | --------------------------------- |
| Frontend  | **Nuxt 4** + Vue 3 + TypeScript  | SPA with file-based routing       |
| Backend   | **Go (Golang)** modular monolith | chi router, domain-driven modules |
| Database  | **PostgreSQL**                   | Migrations via golang-migrate     |
| Infra     | **Docker** + Docker Compose      | Single compose for dev            |
| API style | **REST** (JSON)                  | OpenAPI documented                |
| AI Chat   | **Anthropic Claude** + MCP       | Optional; see §4.6                |

**Architecture style:** Containerized Modular Monolith.
One backend binary, one frontend app, one database — but with strict module boundaries
that allow future extraction to microservices if needed.

---

## 3. Repository Layout

```
ims-th-solution/
├── backend/                # Go modular monolith
│   ├── cmd/                    # Entrypoints: server, migrate, seed, scheduler, contract-check
│   ├── internal/               # Domain modules (private to this app)
│   │   ├── workflow/           # Day-start → manager-approval → closing
│   │   ├── investment/         # Analysis reports, decisions, execution, review
│   │   ├── approval/           # Approval flows, groups/teams, inbox, digital signatures
│   │   ├── permissions/        # Accounts, groups, function/data permissions, role hierarchy
│   │   ├── iam/                # Login, sessions, tokens, MFA
│   │   ├── notification/       # Notification engine & config
│   │   ├── audit/              # Audit log recording & querying
│   │   ├── market_data/        # Quote/history providers, Redis cache, PostgreSQL persistence
│   │   ├── reference_data/     # Thai holidays, currencies, markets, instruments
│   │   ├── compliance/         # IRG / pre-trade / post-trade rule hooks
│   │   ├── integration/        # User dashboard snapshot + task summary endpoints
│   │   ├── watchlist/          # Personal/portfolio watchlists, price-threshold alerts
│   │   └── chat/               # AI financial assistant (Anthropic LLM + MCP)
│   ├── platform/               # Cross-cutting: config, DB, middleware, logging, errors, clock, metrics
│   └── pkg/                    # Shared kernel: types, enums, inter-module contracts
│
├── frontend/               # Nuxt 4 + TypeScript
│   ├── app/                    # Nuxt srcDir
│   │   ├── api/                # Generated OpenAPI types (ims-api.d.ts) + typed client helpers
│   │   ├── assets/css/         # Global CSS
│   │   ├── composables/        # Thin Nuxt composables (useApi, useI18n)
│   │   ├── features/           # Feature-owned components, composables, stores, services, types
│   │   │   ├── approval/
│   │   │   ├── auth/
│   │   │   ├── chat/
│   │   │   ├── compliance/
│   │   │   ├── dashboard/
│   │   │   ├── investment-decision/
│   │   │   ├── investment-ledger/
│   │   │   ├── investment-research/
│   │   │   ├── investment-workspace/
│   │   │   ├── market-data/
│   │   │   ├── my-funds/
│   │   │   ├── notifications/
│   │   │   ├── operator/       # Portfolio V2 operator screens (decisions/execution)
│   │   │   ├── permissions/
│   │   │   ├── portfolio-decision/  # Portfolio V2 (portfolio-code) decision workflow
│   │   │   ├── portfolio-workspace/ # Portfolio V2 (portfolio-code) directory/holdings/cash
│   │   │   ├── settings/
│   │   │   ├── shell/          # App shell: navigation, dashboard tab registry
│   │   │   ├── watchlist/
│   │   │   └── workflow/
│   │   ├── layouts/            # default, auth, dashboard
│   │   ├── middleware/         # auth.ts, permission.ts
│   │   ├── pages/              # File-based route shells
│   │   ├── plugins/            # API client, toast, dayjs
│   │   ├── shared/             # i18n messages, routing helpers, UI primitives
│   │   ├── stores/             # Cross-feature Pinia stores
│   │   └── types/              # App-level shared types
│   └── scripts/                # generate-openapi-types.mjs and dev helpers
│
├── tools/
│   └── mcp/
│       └── ims-mcp-server/     # Node.js stdio MCP server for the chat module
│
├── database/                   # Migrations and seeds
├── docs/                       # Chat-assistant docs, API docs, runbooks, onboarding, handoff
├── infra/                      # Docker Compose, nginx, env files
└── tests/                      # Cross-cutting integration & E2E tests
```

### Backend Module Internal Structure (every module follows this)

```
internal/<module>/
├── domain/
│   ├── entity/          # Aggregate roots, entities (structs + business methods)
│   ├── valueobject/     # Immutable value types
│   ├── event/           # Domain events
│   ├── policy/          # Business validation rules
│   └── repository.go    # Repository INTERFACE (port, no implementation)
├── application/
│   ├── command/         # Write use cases (state-changing operations)
│   ├── query/           # Read use cases
│   └── service/         # Application services
├── infrastructure/
│   ├── persistence/     # Repository SQL implementations
│   └── adapter/         # External system adapters
├── transport/
│   ├── handler/         # HTTP handlers
│   ├── dto/request/     # Request structs
│   ├── dto/response/    # Response structs
│   └── router.go        # Route registration
├── jobs/                # Background/scheduled tasks
├── permission/          # Module permission codes & policies
└── module.go            # Dependency wiring, route registration
```

### Frontend Feature Internal Structure (every feature follows this)

```
frontend/app/features/<feature>/
├── components/          # Feature UI components
├── composables/         # Domain composables (useXxx.ts)
├── services/            # Service layer (API calls, business logic)
├── stores/              # Pinia stores (useXxxStore.ts)
├── types/               # Local TypeScript types
└── navigation/          # dashboardTabs.ts (if feature adds header tabs)
```

> [!WARNING]
> **API client types and routes MUST NOT be written by hand.** Feature components and stores must interact with the backend API exclusively through the typed client generated in `frontend/app/api/ims-api.d.ts` (Absolute path: `C:\Users\kanta\source\repos\ims-th-solution\frontend\app\api\ims-api.d.ts`). Do not write API clients manually.

---

## 4. Business Domain Summary

### 4.1 Workflow Management

The daily investment workflow has four sequential states per contract:

```
Investment Day Start → Manager Approval → Transaction Closing → Accounting Closing
```

**Critical rules:**

- No investment transactions allowed before Investment Day Start
- Manager Approval requires all transactions/reviews for that day to be completed first
- After Manager Approval, fund managers are LOCKED from further transactions
- Transaction Closing and Accounting Closing are separate states (inventory daily, NAV not daily)
- Each state has a matching Cancel operation that reverts to the previous state
- Day Start auto-advances after Transaction Closing (to next Thai business day)
- IRG post-trade check runs at Accounting Closing
- All workflow transitions are audited (actor, time, contract, action, result)

**Confirmation Operations:** First approval stamp — carry-forward, posting, discrepancy capture
**Review Operations:** Second approval stamp — supervisory review of transactions

### 4.2 Stock Investment Management

Full investment lifecycle for domestic and foreign stocks:

1. **Analysis Report** — research analyst creates report with recommendation (Buy/Sell/Hold)
2. **Investment Decision** — fund manager makes decision linked to approved analysis report
3. **Execution Order** — trader executes the order
4. **Trade Confirmation** — confirms executed trades, captures discrepancies
5. **Investment Review** — post-trade review

**Critical validations:**

- Minimum trading unit validation (per market)
- Price tick validation (per exchange rules)
- Buy/Sell recommendation must link to an approved analysis report
- If report approval is required, decision cannot bypass it
- Sell quantity ≤ allowed inventory (cannot oversell)
- Stock must be in investment scope / stock pool
- Cannot trade before Day Start or after Manager Approval
- Contract/fund visibility restrictions based on data permissions
- Submission must go through approval workflow
- Support for: single-target order, multi-account single-target order, future batch import

**Approval Workflow:**

- Approval groups: named groups with member lists
- Approval teams: named teams with member lists
- Approval flows: configurable sequential or group-based approval chains
- Digital signature support for approval records
- Delegation-aware reassignment (if approver on leave → route to delegate)

### 4.4 Permissions Management

Enterprise-grade, server-side enforced:

- **Account management:** Create/update user accounts
- **Group management:** Create/manage groups (Manager, Trader, Risk Control, Admin, etc.)
- **Account-Group mapping:** Assign users to groups
- **Function permissions:** Per-function access rights (e.g., INVESTMENT_VIEW, WORKFLOW_EXECUTE)
- **Data permissions:** Per-contract/fund visibility (user X can only see contracts A, B, C)
- Permission checks are ALWAYS server-side — never frontend-only
- Effective permissions = own permissions + delegated permissions from active leave

### 4.5 Supporting Modules

- **Audit / Change Log:** Configuration for which tables to audit, query interface for audit records
- **Notification:** Approval notifier, workflow stuck-day operator notifier, and watchlist alert notifier; no dedicated user nav
- **Reference Data:** Markets (SET, TFEX, foreign), currencies, instrument types, Thai holidays; exposes a `Resolver()` used by market data
- **Compliance/IRG:** Extension points for blacklist/whitelist, investment ratio, instrument restriction, credit rating rules
- **Integration:** User dashboard snapshot and personal task-summary endpoints
- **Watchlist:** Personal/portfolio watchlist items with market-price threshold rules, alert events, acknowledgement, and manual evaluation; depends on reference-data's resolver, market-data's quote provider, investment's portfolio-scope resolver, and notification's alert notifier

### 4.6 AI Chat Assistant

An AI financial assistant backed by the Anthropic Claude API (or a compatible LLM) and an MCP server.

**Architecture:**
- `backend/internal/chat/` — session management, message persistence, LLM provider adapter, MCP client
- `tools/mcp/ims-mcp-server/` — Node.js stdio MCP server; all business data is accessed through it (never by the chat module importing other modules' internals)
- Frontend: `frontend/app/features/chat/` — chat UI components and composables

**Critical rules:**
- The chat module MUST NOT import other modules' `internal/` packages; it reaches business data only via MCP tools
- `CHAT_WRITE_ENABLED` gates mutating tools; default is read-only
- If the LLM provider is misconfigured (missing API key), the `/chat` route is not mounted; the rest of the API stays up
- Chat sessions and messages are persisted in PostgreSQL; audit events are recorded via the audit module's Recorder

**Configuration:**
- `LLM_PROVIDER=anthropic` + `ANTHROPIC_API_KEY` — required for chat to activate
- `ANTHROPIC_MODEL` — model selection (defaults to a configured default)
- `CHAT_MAX_TOKENS_PER_TURN` — per-turn token budget
- `CHAT_MCP_SERVERS_CONFIG` — optional path to mcp-servers.yaml; when absent, spawns the bundled ims-mcp binary
- `CHAT_MCP_IMS_BIN` — override path to the ims-mcp binary
- `IMS_API_BASE_URL` — forwarded to the MCP server so its tools reach the IMS REST API

### 4.7 Portfolio V2 (portfolio-code identity)

An additive API generation that identifies portfolios by business `portfolioCode`
instead of internal UUIDs, mounted alongside V1 rather than replacing it.

- Routes: `/api/v2/portfolios/*`, separate from `/api/v1/investment/*`.
- Own Swagger spec/basePath at `/swagger/v2/*` (Swagger 2.0 only supports one
  basePath per spec, so V2 cannot share the V1 document — see
  `backend/cmd/server/swagger_v2_docs.go`).
- Frontend consumes it via `useOpenApiClientV2()` in `frontend/app/api/openapi.ts`,
  called from the `portfolio-decision/` and `portfolio-workspace/` feature service
  layers; the `operator/` feature builds on top of those.
- Only the `investment` module exposes V2 routes so far; this is expected to grow
  as other modules migrate to portfolio-code identity.
- Design/contract reference: `docs/api/portfolio-v2-api-ddd.md`.

---

## 5. Naming Conventions

| Artifact               | Convention                               | Example                                   |
| ---------------------- | ---------------------------------------- | ----------------------------------------- |
| Backend module folder  | `snake_case`                             | `leave_delegation/`                       |
| Go files               | `snake_case.go`                          | `analysis_report.go`                      |
| Go structs             | `PascalCase`                             | `AnalysisReport`                          |
| Go interfaces          | `PascalCase` + suffix                    | `AnalysisReportRepository`                |
| API endpoints          | `kebab-case`, plural                     | `GET /api/v1/analysis-reports`            |
| SQL migrations         | `<timestamp>_<module>__<desc>.<dir>.sql` | `20260301000001_iam__create_users.up.sql` |
| SQL tables             | `snake_case`, plural, module-prefixed    | `investment_analysis_reports`             |
| Frontend module folder | `kebab-case`                             | `leave-delegation/`                       |
| Vue components         | `PascalCase.vue`                         | `AnalysisReportForm.vue`                  |
| Pinia stores           | `use<Name>Store.ts`                      | `useInvestmentDecisionStore.ts`           |
| Composables            | `use<Name>.ts`                           | `useWorkflow.ts`                          |
| TypeScript types       | `<domain>.types.ts`                      | `investment.types.ts`                     |

Note: **NO manual API client files (like `<domain>Api.ts`) are allowed.** All API interfaces are automatically generated into `C:\Users\kanta\source\repos\ims-th-solution\frontend\app\api\ims-api.d.ts`.

---

## 6. Module Boundary Rules

**ALLOWED imports:**

- Any module → `pkg/types/`, `pkg/enum/`, `pkg/contract/` (shared kernel)
- Any module → `platform/*` (cross-cutting utilities)

**FORBIDDEN imports:**

- Module A → Module B's `domain/`, `application/`, `infrastructure/`, `transport/` — NEVER
- Cross-module communication → use `pkg/contract/` interfaces or domain events only

**NEVER do these:**

- Put business logic in HTTP handlers or Vue components
- Import domain entities into transport layer directly (use DTOs)
- Skip server-side permission checks
- Use `any` / `interface{}` for cross-module data passing
- Create circular module dependencies
- **Write custom API fetch clients, request/response models, or route fetch utilities by hand.** All frontend API interactions must use the generated schema/types from `C:\Users\kanta\source\repos\ims-th-solution\frontend\app\api\ims-api.d.ts` (built via `make api-client`). Do not fucking dare write API code by hand.

---

## 7. Business Date vs System Date

This system operates on **Thai market business days**, not naive UTC timestamps.

- `business_date` is a first-class concept used in workflow state, trading windows, and report validity
- Use the `platform/clock/` abstraction for all date operations
- Thai market holidays are stored in `reference_data` and must be loaded for business day calculations
- Backend always stores timestamps in **UTC**
- Frontend displays in **Asia/Bangkok (UTC+7)** timezone

---

## 8. Design Documents (in `docs/` directory)

The `docs/` directory contains:

- `docs/api/` — API contract documentation
- `docs/chat-assistant/` — AI chat feature design and MCP tool specs
- `docs/onboarding/` — Developer onboarding guides
- `docs/runbook/` — Operational runbooks
- `docs/handoff/` — Handoff notes
- `docs/IMS_Chat_AI_Financial_Assistant_Project_Documentation.docx` — Full chat-assistant project spec
- `docs/IMS_Chat_AI_Financial_Assistant_Project_Documentation_Summary.md` — Chat-assistant spec summary

For ADRs and domain state machine docs, refer to the original business requirement documents listed in §11.

---

## 9. Technology Versions

| Tool            | Version | Notes                                       |
| --------------- | ------- | ------------------------------------------- |
| Go              | 1.25    | `go.mod` declared version                   |
| Node.js         | 20 LTS+ | For Nuxt build                              |
| Nuxt            | 4.x     | `nuxt ^4.3.1` in package.json               |
| Vue             | 3.x     | Composition API only                        |
| TypeScript      | 5.x     | Strict mode                                 |
| PostgreSQL      | 16+     | With uuid-ossp extension                    |
| Docker          | 24+     |                                             |
| Docker Compose  | v2+     |                                             |
| chi (Go router) | v5      | HTTP router                                 |
| golang-migrate  | v4      | DB migrations                               |
| Pinia           | 3.x     | Vue state management (`pinia ^3.0.4`)        |
| Tailwind CSS    | 4.x     | Utility-first CSS (`@tailwindcss/postcss` v4)|
| openapi-fetch   | 0.17+   | Typed OpenAPI client in frontend            |
| Playwright      | latest  | E2E testing                                 |

---

## 10. Environment Variables

Key environment variables the system expects:

```env
# Backend — core
APP_ENV=development
APP_PORT=8080
APP_LOG_LEVEL=debug
APP_JWT_SECRET=<secret>

# Database
DB_HOST=localhost
DB_PORT=5437
DB_NAME=ims_dev
DB_USER=ims_app
DB_PASSWORD=<password>
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25

# Reporting currency (required in EVERY environment, no dev/test default)
REPORTING_CURRENCY=THB

# Rate limiting (optional Redis backend)
RATE_LIMIT_BACKEND=memory   # or "redis"
REDIS_ADDR=localhost:6379

# Market data
ALPHA_VANTAGE_API_KEY=<key>  # Required in non-development

# Chat / AI assistant (optional — omitting disables /chat)
LLM_PROVIDER=anthropic
ANTHROPIC_API_KEY=<key>
ANTHROPIC_MODEL=claude-sonnet-4-6
CHAT_WRITE_ENABLED=false
IMS_API_BASE_URL=http://localhost:8080

# Frontend
NUXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NUXT_PUBLIC_APP_NAME=IMS Thailand
```

---

## 11. Source Documents

The original business requirements are in these files (in the repo root or `docs/` folder):

- `00_IMSFunctionDescription_WorkflowManagement_V1_1.docx`
- `01_IMSFunctionDescription_StockInvestmentManagementV1_1.docx`
- `02_IMSFunctionDescription_Leave_ExpenseApprovalWorkflowManagement_V1_1.docx`
- `03_IMSFunctionDescription-PermissionsManagementV1_0.docx`
- `00_Workflow_Management_Flowchart_v1_0_Eng_Ver.xml` (draw.io)
- `01_StockInvestmentManagementFlowchart_v1_0_Eng_Ver.xml` (draw.io)
- `D01Project_Plan.xlsx`

When in doubt about business rules, these documents are the source of truth.
