# CLAUDE.md — Instructions for Claude Code

> **Read `AIREAD.md` first.** It contains the full project context, business rules,
> and domain knowledge. This file contains only operational instructions for working in this codebase.

---

## Project Identity

- **Name:** IMS (Investment Management System) — Thailand PoC
- **Repo:** `ims-th-solution`
- **License:** MIT
- **Language:** Go 1.23+ (backend), TypeScript + Nuxt 3 (frontend), PostgreSQL (database)

---

## Quick Commands

```bash
# === Development ===
make dev                    # Start everything (docker-compose up + frontend dev)
make dev-backend            # Backend only (go run cmd/server/main.go)
make dev-frontend           # Frontend only (cd frontend && npm run dev)

# === Database ===
make migrate-up             # Run pending migrations
make migrate-down           # Rollback last migration
make migrate-new module=workflow name=add_carry_forward  # Create new migration pair
make seed                   # Load seed data
make db-reset               # Drop + recreate + migrate + seed

# === Testing ===
make test                   # All tests
make test-unit              # Unit tests only (backend)
make test-integration       # Integration tests
make test-e2e               # Playwright E2E tests
make lint                   # Lint both frontend and backend

# === Build ===
make build                  # Build backend binary + frontend static
make docker-build           # Build Docker images
```

---

## Folder Structure — Where To Put Things

### This is a modular monolith. Every file has ONE correct location.

**If you are adding a new Go file, ask: "Which module does this belong to?"**
Then place it in the correct sublayer within that module.

```
backend/internal/<MODULE>/
  domain/entity/         → Business entity structs + methods
  domain/valueobject/    → Immutable value types
  domain/event/          → Domain events
  domain/policy/         → Business validation rules
  domain/repository.go   → Repository INTERFACE only
  application/command/   → Write use cases
  application/query/     → Read use cases
  application/dto/       → Application DTOs
  infrastructure/persistence/  → SQL repository implementations
  infrastructure/adapter/      → External system adapters
  transport/handler/     → HTTP handlers
  transport/dto/request/ → Request structs
  transport/dto/response/→ Response structs
  transport/validator/   → Request validators
  transport/router.go    → Route definitions
  jobs/                  → Scheduled/background tasks
  permission/policies.go → Permission code declarations
  module.go              → Module wire-up (constructor, route registration)
```

**If you are adding a new Vue file, ask: "Which domain module and component type?"**

```
frontend/modules/<DOMAIN>/
  components/forms/      → Input forms
  components/tables/     → Data tables
  components/dialogs/    → Modal dialogs
  components/cards/      → Info cards
  components/widgets/    → Complex composite widgets
  composables/           → useXxx.ts composables
  stores/                → useXxxStore.ts Pinia stores
  api/                   → xxxApi.ts API client functions
  types/                 → xxx.types.ts TypeScript interfaces
```

**If you are adding a page (route):**

```
frontend/app/pages/<domain>/<feature>/index.vue
frontend/app/pages/<domain>/<feature>/[id].vue
frontend/app/pages/<domain>/<feature>/create.vue
```

**If you are adding shared UI:**

```
frontend/shared/components/ui/      → Base UI primitives (AppButton, AppInput, etc.)
frontend/shared/components/layout/  → Layout parts (Sidebar, TopBar, etc.)
frontend/shared/composables/        → Cross-domain composables (useAuth, usePermissionGuard)
frontend/shared/stores/             → Global stores (useAuthStore, useGlobalStore)
frontend/shared/types/              → Shared TypeScript types
frontend/shared/utils/              → Formatters, validators, constants
```

**If you are adding a database migration:**

```
database/migrations/<YYYYMMDDHHMMSS>_<module>__<description>.up.sql
database/migrations/<YYYYMMDDHHMMSS>_<module>__<description>.down.sql
```

**If you are adding cross-module infrastructure:**

```
backend/platform/       → middleware, config, DB helpers, logging, errors, clock
backend/pkg/types/      → Shared value types (Money, DateRange, Pagination)
backend/pkg/enum/       → Shared enums
backend/pkg/contract/   → Inter-module interfaces ONLY
```

---

## Module List

| Module              | Backend Path               | Frontend Path                     | Database Prefix  |
| ------------------- | -------------------------- | --------------------------------- | ---------------- |
| Workflow Management | `internal/workflow/`       | `modules/workflow/`               | `workflow__`     |
| Stock Investment    | `internal/investment/`     | `modules/investment/`             | `investment__`   |
| Approval Workflow   | `internal/approval/`       | `modules/approval/`               | `approval__`     |
| Permissions         | `internal/permissions/`    | `modules/permissions/`            | `permissions__`  |
| Identity & Access   | `internal/iam/`            | (uses shared/stores/useAuthStore) | `iam__`          |
| Notification        | `internal/notification/`   | `modules/notification/`           | `notification__` |
| Audit               | `internal/audit/`          | `modules/audit/`                  | `audit__`        |
| Market Data         | `internal/market_data/`    | —                                 | `market_data__`  |
| Reference Data      | `internal/reference_data/` | —                                 | `reference__`    |
| Compliance/IRG      | `internal/compliance/`     | —                                 | `compliance__`   |
| ETL/Integration     | `internal/integration/`    | —                                 | `integration__`  |

---

## STRICT Rules — Do NOT Violate These

### Architecture

1. **Never import one module's internal code from another module.**
   Cross-module → use `pkg/contract/` interfaces or domain events.
2. **Never put business logic in HTTP handlers.** Handlers only: parse request → call application service → format response.
3. **Never put business logic in Vue components.** Components only: display data, capture user input, call store actions.
4. **Never skip server-side permission checks.** Frontend guards are UX only. Backend middleware enforces.
5. **Never use domain entities in transport layer.** Always convert to/from DTOs.
6. **Never create circular module dependencies.** If A needs B and B needs A → extract to shared_kernel or events.

### Code Style

7. **Go files:** `snake_case.go`. Structs: `PascalCase`. No exported global variables.
8. **Vue files:** `PascalCase.vue`. Composition API + `<script setup lang="ts">` only. No Options API.
9. **One file, one responsibility.** No god files. Max ~300 lines per file as a guideline.
10. **No magic strings.** Use constants or enums for status codes, permission codes, error codes.
11. **All timestamps in UTC on backend.** Display in `Asia/Bangkok` on frontend.
12. **Every migration `.up.sql` must have a matching `.down.sql`.** Down scripts must use `IF EXISTS`.

### Testing

13. **Domain layer:** Unit test business rules with table-driven tests.
14. **Application layer:** Unit test use cases with mocked repositories.
15. **Transport layer:** Test handlers with httptest.
16. **Frontend:** Component tests with Vitest, E2E with Playwright.

---

## Initialization Guide

When initializing this project from scratch, follow this order:

### Phase 1: Repository Skeleton

```bash
# 1. Create top-level structure
mkdir -p backend/{cmd/{server,migrate,seed,scheduler},internal,platform,pkg,api/openapi}
mkdir -p frontend/{app/{layouts,pages,middleware,plugins},modules,shared,assets,public}
mkdir -p database/{migrations,seeds,reference,views,functions,triggers,test_data,erd}
mkdir -p docs/{adr,api,domain,runbook,onboarding}
mkdir -p scripts
mkdir -p infra/{docker,nginx,env}
mkdir -p tests/{integration,e2e/specs}
```

### Phase 2: Frontend Init (Nuxt 3)

```bash
cd frontend
npx nuxi@latest init . --force --packageManager npm
npm install
npm install -D @nuxtjs/tailwindcss @pinia/nuxt @vueuse/nuxt
npm install pinia @vueuse/core dayjs
npm install -D typescript @types/node
```

Configure `nuxt.config.ts`:

```typescript
export default defineNuxtConfig({
  devtools: { enabled: true },
  srcDir: "app/",
  modules: ["@nuxtjs/tailwindcss", "@pinia/nuxt", "@vueuse/nuxt"],
  css: ["~/assets/css/main.css"],
  runtimeConfig: {
    public: {
      apiBaseUrl:
        process.env.NUXT_PUBLIC_API_BASE_URL || "http://localhost:8080/api/v1",
      appName: process.env.NUXT_PUBLIC_APP_NAME || "IMS Thailand",
    },
  },
  typescript: {
    strict: true,
  },
  tailwindcss: {
    cssPath: "~/assets/css/main.css",
  },
});
```

Create frontend domain module folders:

```bash
cd frontend
for mod in workflow investment leave-delegation approval permissions notification audit; do
  mkdir -p modules/$mod/{components/{forms,tables,dialogs,cards,widgets},composables,stores,api,types}
  echo "export {}" > modules/$mod/index.ts
done

mkdir -p shared/{components/{ui,layout,data-display},composables,stores,types,utils}
mkdir -p app/pages/{workflow,investment/{analysis,decision,execution,review},leave/agents,approval/{config},permissions/{accounts,groups},settings/{notifications,audit}}
```

### Phase 3: Backend Init (Go)

```bash
cd backend
go mod init github.com/neo-kanta/ims-th-solution/backend
```

Install core dependencies:

```bash
go get github.com/go-chi/chi/v5
go get github.com/go-chi/cors
go get github.com/jackc/pgx/v5
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-jwt/jwt/v5
go get github.com/go-playground/validator/v10
go get github.com/google/uuid
go get go.uber.org/zap                       # or use slog (stdlib)
go get github.com/joho/godotenv
```

Create backend module skeletons:

```bash
cd backend
for mod in workflow investment leave_delegation approval permissions iam notification audit market_data reference_data compliance integration; do
  mkdir -p internal/$mod/{domain/{entity,valueobject,event,policy},application/{command,query,dto},infrastructure/{persistence,adapter},transport/{handler,dto/{request,response},validator},jobs,permission}
  cat > internal/$mod/module.go << 'GOEOF'
package $(echo $mod | tr '/' '_')

// Module wire-up — register routes and dependencies here.
GOEOF
  cat > internal/$mod/README.md << 'EOF'
# Module: $mod

TODO: describe this module's purpose and key domain rules.
EOF
done

# Create platform packages
for pkg in config database middleware httputil logging errors clock validation testutil; do
  mkdir -p platform/$pkg
done

# Create shared kernel packages
mkdir -p pkg/{types,enum,contract}
```

### Phase 4: Database Init

```bash
cd database
# First migration: users table
cat > migrations/20260301000001_iam__create_users.up.sql << 'SQL'
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE iam_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_on_leave BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID,
    updated_by UUID
);

CREATE INDEX idx_iam_users_username ON iam_users(username);
CREATE INDEX idx_iam_users_is_active ON iam_users(is_active);
SQL

cat > migrations/20260301000001_iam__create_users.down.sql << 'SQL'
DROP TABLE IF EXISTS iam_users;
SQL
```

### Phase 5: Infrastructure

Create `infra/docker-compose.yml`:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    ports:
      - "5437:5437"
    command: -p 5437
    environment:
      POSTGRES_DB: ims_dev
      POSTGRES_USER: ims_app
      POSTGRES_PASSWORD: ims_dev_password
    volumes:
      - pgdata:/var/lib/postgresql/data

  backend:
    build:
      context: ../backend
      dockerfile: ../infra/docker/Dockerfile.backend
    ports:
      - "8080:8080"
    env_file:
      - ./env/.env.development
    depends_on:
      - postgres

  frontend:
    build:
      context: ../frontend
      dockerfile: ../infra/docker/Dockerfile.frontend
    ports:
      - "3000:3000"
    depends_on:
      - backend

volumes:
  pgdata:
```

Create `infra/env/.env.development`:

```env
APP_ENV=development
APP_PORT=8080
APP_LOG_LEVEL=debug
APP_JWT_SECRET=dev-secret-change-in-production

DB_HOST=postgres
DB_PORT=5437
DB_NAME=ims_dev
DB_USER=ims_app
DB_PASSWORD=ims_dev_password
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25

NUXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NUXT_PUBLIC_APP_NAME=IMS Thailand
```

---

## Current Implementation Priority

The PoC is being built in this order (from D01 Project Plan):

| Priority | What                                                             | Target    |
| -------- | ---------------------------------------------------------------- | --------- |
| **P0**   | Frontend skeleton (Nuxt 3 + routing + layouts)                   | Month 1-2 |
| **P0**   | Backend API gateway (chi router + middleware + health check)     | Month 1-2 |
| **P0**   | Database schema design + initial migrations                      | Month 1-2 |
| **P1**   | Permissions module (accounts, groups, function/data permissions) | Month 2-3 |
| **P1**   | IAM module (login, session, JWT)                                 | Month 2-3 |
| **P2**   | Workflow module (day-start through closing)                      | Month 3-4 |
| **P2**   | ETL / Data integration PoC                                       | Month 2-3 |
| **P3**   | Investment module (4-step flow)                                  | Month 4-6 |
| **P3**   | Leave & delegation module                                        | Month 4-5 |
| **P4**   | Approval workflow                                                | Month 5-6 |
| **P4**   | IRG / compliance hooks                                           | Month 5-6 |
| **P5**   | Notification, audit UI, dashboard polish                         | Month 6-7 |
| **P6**   | E2E testing, UAT, production deployment                          | Month 7-8 |

**When asked to "initialize the project", focus on P0:**

1. Frontend with Nuxt 3, Tailwind, Pinia, layouts, empty pages for all domains
2. Backend with chi router, health endpoint, CORS, structured logging, config loading
3. Docker Compose with PostgreSQL
4. First migrations (iam_users, permissions tables)

---

## API Gateway / Backend Platform Layer

The backend acts as a single API gateway. All requests go through middleware in this order:

```
Request → RequestID → Logger → CORS → Recovery → Auth → PermissionCheck → DataScope → Handler
```

**Key platform packages to build first:**

```
platform/config/config.go        → Load .env, expose AppConfig struct
platform/database/connection.go  → pgx connection pool
platform/database/tx.go          → Transaction helper (begin/commit/rollback)
platform/database/health.go      → Ping check for health endpoint
platform/middleware/request_id.go→ Inject X-Request-ID header
platform/middleware/cors.go      → CORS configuration
platform/middleware/recovery.go  → Panic recovery with logging
platform/middleware/auth.go      → JWT validation, extract user context
platform/middleware/permission.go→ Function permission enforcement
platform/middleware/data_scope.go→ Data permission scoping (contract visibility)
platform/middleware/audit.go     → Audit log side-effect
platform/httputil/response.go   → Standard JSON response helpers
platform/httputil/pagination.go → Parse ?page=&limit= params
platform/httputil/error_response.go → Business-safe error formatting
platform/logging/logger.go      → Structured logger (slog or zap)
platform/errors/business.go     → BusinessError, ValidationError types
platform/errors/codes.go        → Error code constants
platform/clock/clock.go         → Clock interface for testability
platform/clock/business_date.go → Thai business day calculator
```

**Minimal `cmd/server/main.go` pattern:**

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    chimw "github.com/go-chi/chi/v5/middleware"
    // import platform packages
    // import module packages
)

func main() {
    // 1. Load config
    // 2. Connect database
    // 3. Create logger
    // 4. Wire modules
    // 5. Build router

    r := chi.NewRouter()
    r.Use(chimw.RequestID)
    r.Use(chimw.RealIP)
    r.Use(customLogger)
    r.Use(chimw.Recoverer)
    r.Use(corsMiddleware)

    // Health check (no auth required)
    r.Get("/health", healthHandler)

    // API v1 routes (auth required)
    r.Route("/api/v1", func(r chi.Router) {
        r.Use(authMiddleware)
        r.Use(auditMiddleware)

        // Mount each module's routes
        // workflowModule.RegisterRoutes(r)
        // investmentModule.RegisterRoutes(r)
        // etc.
    })

    http.ListenAndServe(":"+cfg.Port, r)
}
```

---

## Frontend Patterns

### API Client Pattern

Every module has an `api/xxxApi.ts` file. Use `$fetch` (Nuxt built-in, based on ofetch):

```typescript
// modules/workflow/api/workflowApi.ts
const BASE = "/api/v1";

export const workflowApi = {
  getStatus(contractId: string) {
    return $fetch(`${BASE}/workflow/${contractId}/status`);
  },
  startDay(contractId: string, date: string) {
    return $fetch(`${BASE}/workflow/${contractId}/start-day`, {
      method: "POST",
      body: { date },
    });
  },
};
```

### Pinia Store Pattern

```typescript
// modules/workflow/stores/useWorkflowStore.ts
import { defineStore } from "pinia";
import { workflowApi } from "../api/workflowApi";
import type { WorkflowStatus } from "../types/workflow.types";

export const useWorkflowStore = defineStore("workflow", () => {
  const status = ref<WorkflowStatus | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchStatus(contractId: string) {
    loading.value = true;
    error.value = null;
    try {
      status.value = await workflowApi.getStatus(contractId);
    } catch (e: any) {
      error.value = e.data?.message || "Failed to fetch workflow status";
    } finally {
      loading.value = false;
    }
  }

  return { status, loading, error, fetchStatus };
});
```

### Permission Guard Pattern

```typescript
// shared/composables/usePermissionGuard.ts
export function usePermissionGuard() {
  const authStore = useAuthStore();

  function hasFunction(code: string): boolean {
    return authStore.permissions.functions.includes(code);
  }

  function hasContract(contractId: string): boolean {
    return authStore.permissions.contracts.includes(contractId);
  }

  return { hasFunction, hasContract };
}
```

### Page Template

```vue
<!-- app/pages/workflow/index.vue -->
<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  meta: { permission: "WORKFLOW_VIEW" },
});

const workflowStore = useWorkflowStore();
const globalStore = useGlobalStore();

onMounted(() => {
  if (globalStore.activeContractId) {
    workflowStore.fetchStatus(globalStore.activeContractId);
  }
});
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">Workflow Operations</h1>
    <!-- Use module components here -->
  </div>
</template>
```

---

## Git Conventions

```
feat(workflow): add day-start operation handler
fix(investment): correct sell quantity validation
refactor(permissions): extract permission checker interface
docs(adr): add decision record for workflow state machine
test(approval): add approval routing policy unit tests
chore(infra): update Docker base image
```

Branch naming: `feat/<module>/<short-description>`, `fix/<module>/<short-description>`

---

## Common Mistakes to Avoid

1. **Don't create files outside the module structure.** If you're unsure where a file goes, check the module table above.
2. **Don't add npm packages without checking if Nuxt already provides it.** Nuxt auto-imports composables, `$fetch`, etc.
3. **Don't write SQL in Go handler files.** SQL lives in `infrastructure/persistence/` only.
4. **Don't hardcode contract IDs, user IDs, or permission codes.** Use constants and lookup from DB/config.
5. **Don't forget the `.down.sql` migration.** Every up has a down. No exceptions.
6. **Don't use `localStorage` in Nuxt.** Use Pinia stores with `useState` for SSR-safe state, or `useCookie` for persistence.
7. **Don't put API base URL in component files.** Use `useRuntimeConfig().public.apiBaseUrl`.

---

## When You're Unsure

1. Check `AIREAD.md` for business context
2. Check the original `.docx` spec documents for detailed field-level requirements
3. Check `docs/adr/` for past architectural decisions
4. If a requirement seems ambiguous, make the safest assumption and add a `// ASSUMPTION:` comment
5. When in doubt, keep the domain module boundary strict — it's easier to relax later than to tighten
