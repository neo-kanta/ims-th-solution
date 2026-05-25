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
| Frontend  | **Nuxt 3** + Vue 3 + TypeScript  | SPA with file-based routing       |
| Backend   | **Go (Golang)** modular monolith | chi router, domain-driven modules |
| Database  | **PostgreSQL**                   | Migrations via golang-migrate     |
| Infra     | **Docker** + Docker Compose      | Single compose for dev            |
| API style | **REST** (JSON)                  | OpenAPI documented                |

**Architecture style:** Containerized Modular Monolith.
One backend binary, one frontend app, one database — but with strict module boundaries
that allow future extraction to microservices if needed.

---

## 3. Repository Layout

```
ims-th-solution/
├── api/                    # Go modular monolith
│   ├── cmd/                    # Entrypoints (server, migrate, seed, scheduler)
│   ├── internal/               # Domain modules (private to this app)
│   │   ├── workflow/           # Day-start → manager-approval → closing
│   │   ├── investment/         # Analysis reports, decisions, execution, review
│   │   ├── approval/           # Approval flow config, signing groups/teams
│   │   ├── permissions/        # Accounts, groups, function/data permissions
│   │   ├── iam/                # Login, sessions, tokens
│   │   ├── notification/       # Notification engine & config
│   │   ├── audit/              # Audit log recording & querying
│   │   ├── market_data/        # Market data integration adapter
│   │   ├── reference_data/     # Currencies, markets, instruments, holidays
│   │   ├── compliance/         # IRG / pre-trade / post-trade rule hooks
│   │   └── integration/        # ETL jobs & external system adapters
│   ├── platform/               # Cross-cutting: config, DB, middleware, logging, errors, clock
│   └── pkg/                    # Shared kernel: types, enums, inter-module contracts
│
├── web/                   # Nuxt 3 + TypeScript
│   ├── app/                    # Nuxt app directory
│   │   ├── layouts/            # default, auth, dashboard
│   │   ├── pages/              # File-based routing by domain
│   │   ├── middleware/         # auth.ts, permission.ts
│   │   └── plugins/            # API client, toast, dayjs
│   ├── modules/                # Domain feature modules (components, stores, composables, api)
│   │   ├── workflow/
│   │   ├── investment/
│   │   ├── approval/
│   │   ├── permissions/
│   │   ├── notification/
│   │   └── audit/
│   └── shared/                 # Shared UI: components, composables, stores, types, utils
│
├── database/                   # Migrations, seeds, views, functions, triggers, ERDs
├── docs/                       # ADRs, API docs, domain docs, runbooks, onboarding
├── scripts/                    # Dev automation scripts
├── infra/                      # Docker, docker-compose, nginx, env files
└── tests/                      # Cross-cutting integration & E2E tests
```

### Backend Module Internal Structure (EVERY module follows this)

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
│   └── dto/             # Application-level DTOs
├── infrastructure/
│   ├── persistence/     # Repository SQL implementations
│   └── adapter/         # External system adapters
├── transport/
│   ├── handler/         # HTTP handlers
│   ├── dto/request/     # Request structs
│   ├── dto/response/    # Response structs
│   ├── validator/       # Request validation
│   └── router.go        # Route registration
├── jobs/                # Background/scheduled tasks
├── permission/          # Module permission codes & policies
├── module.go            # Dependency wiring, route registration
└── README.md
```

### Frontend Module Internal Structure (EVERY module follows this)

```
frontend/app/features/<domain>/
├── components/          # Domain-specific UI components
├── composables/         # Domain composables (useXxx.ts)
├── stores/              # Pinia stores (useXxxStore.ts)
├── types/               # Local TypeScript types (xxx.types.ts)
└── index.ts             # Barrel exports
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
- **Notification:** Basic settings, message templates, per-contract notification config, disabled notifications
- **Reference Data:** Markets (SET, TFEX, foreign), currencies, instrument types, Thai holidays
- **Compliance/IRG:** Extension points for blacklist/whitelist, investment ratio, instrument restriction, credit rating rules (hooks only in PoC)
- **ETL/Integration:** Import/export adapters for market data, OMS, PAM — stub in PoC

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

Read these for deeper context:

- `docs/adr/001-modular-monolith.md` — Why monolith, not microservices
- `docs/adr/005-folder-structure.md` — Full folder structure rationale (D05 document)
- `docs/domain/workflow_states.md` — Workflow state machine details
- `docs/domain/stock_investment_flow.md` — Full stock investment lifecycle
- `docs/api/api_endpoints.md` — API contract listing

---

## 9. Technology Versions

| Tool            | Version | Notes                    |
| --------------- | ------- | ------------------------ |
| Go              | 1.23+   | Use latest stable        |
| Node.js         | 20 LTS+ | For Nuxt build           |
| Nuxt            | 3.x     | Latest stable            |
| Vue             | 3.x     | Composition API only     |
| TypeScript      | 5.x     | Strict mode              |
| PostgreSQL      | 16+     | With uuid-ossp extension |
| Docker          | 24+     |                          |
| Docker Compose  | v2+     |                          |
| chi (Go router) | v5      | HTTP router              |
| golang-migrate  | v4      | DB migrations            |
| Pinia           | 2.x     | Vue state management     |
| Tailwind CSS    | 3.x     | Utility-first CSS        |
| Playwright      | latest  | E2E testing              |

---

## 10. Environment Variables

Key environment variables the system expects:

```env
# Backend
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
