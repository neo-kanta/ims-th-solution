# Workflow Current State Snapshot

**Snapshot date:** 2026-06-15
**Branch:** `neo-develop`
**Latest commit:** `04f04db Merge branch 'ai-chatbot' into neo-develop`

---

## 1. Branch and Commit

```
branch:  neo-develop
commit:  04f04db  Merge branch 'ai-chatbot' into neo-develop
date:    2026-06-15
```

---

## 2. Commands Run and Results

```bash
# Build
cd backend && go build ./...
# Result: (no output) — PASS

# Full test suite
cd backend && go test ./...
# Result: all packages ok — PASS (zero failures)

# Focused adapter tests
go test ./internal/investment/infrastructure/adapter/... -v -run TestContractCatalog
# Results:
#   PASS: TestContractCatalogAdapter_ResolvesActiveContractCode
#   PASS: TestContractCatalogAdapter_EmptyCodeReturnsValidationError
#   PASS: TestContractCatalogAdapter_UnknownCodeReturnsNotFound
#   PASS: TestContractCatalogAdapter_NeverReturnsNilUUIDOnSuccess
#   PASS: TestContractCatalogAdapter_RepositoryErrorPropagates
#   PASS: TestContractCatalogAdapter_ImplementsContractCatalog
#   ok  github.com/neo-kanta/ims-th-solution/backend/internal/investment/infrastructure/adapter
```

---

## 3. Completed Work Summary

- **Stop Condition analysis:** Investigated Permissions, Approval, Investment, Compliance modules against Workflow redesign requirements. Found 2 hard blockers and 3 resolvable issues.
- **Blocker resolved — ContractCatalog:** Added `contract.ContractCatalog` interface to `pkg/contract/contracts.go` so Workflow can resolve business-readable `contractCode` strings to internal UUIDs without importing investment internals.
- **Repository method:** Added `GetByContractCode` to `FundRepository` interface and implemented SQL lookup using the `contract_code` column added in migration `20260613000008`.
- **Investment adapter:** Created `ContractCatalogAdapter` in `investment/infrastructure/adapter/` with typed errors (`ErrContractNotFound`, `ErrContractCodeEmpty`).
- **Investment module export:** Added `ContractCatalog()` method to investment `module.go`.
- **Workflow wiring:** Added `contractCatalog` field, `SetContractCatalog()`, and `ResolveContractCode()` helper to workflow `module.go`.
- **Server wiring:** Added `workflowModule.SetContractCatalog(investmentModule.ContractCatalog())` to `cmd/server/main.go`.
- **Tests fixed:** Updated `fundAUMFundRepo` and `postFundRepo` test stubs to satisfy the expanded `FundRepository` interface.
- **6 new tests:** All pass. Interface compliance verified at compile time.
- **Handoff docs created:** `docs/handoff/workflow-backend.md`, `docs/handoff/workflow-frontend.md`, `docs/handoff/workflow-current-state.md`.

---

## 4. Remaining Work Summary

**Backend (ordered):**

1. Fix `buildActorContext()` — populate `Username` from `claims.Username` (one-line fix)
2. Verify state value names in `workflow_state.go` — canonical name is `INVESTMENT_DAY_STARTED`
3. Verify operation names in `workflow_action.go` — rename to `START_INVESTMENT_DAY`, `MANAGER_APPROVE`, etc.
4. Create `WorkflowApprovalSetting` entity + repository interface + Postgres impl + migration
5. Add `actor_account_code` and `is_admin_override` columns to `workflow__transition_log` migration
6. Implement `GET /api/workflow/daily?businessDate=...&contractCode=...`
7. Implement `POST /api/workflow/daily/execute`
8. Implement `GET /api/workflow/settings` and `PUT /api/workflow/settings`
9. Build aggregated daily response with `moduleReadiness` object (report approval as `ready: false`)
10. Implement admin override logic (JWT `Roles` contains `"Admin"` check)
11. Add Swagger annotations and run `make swagger && make api-client`
12. Write required tests (10 cases listed in `workflow-backend.md` Section 9)

**Frontend (blocked until backend endpoints exist):**

- Redesign `frontend/app/pages/workflow/index.vue` using new daily API
- Add `WorkflowOverviewTab.vue`, `WorkflowAuditTab.vue`, `WorkflowSettingsTab.vue`
- Register workflow tabs via dashboard tab registry (do NOT edit `dashboard.vue`)
- Update existing composables to use new business-readable API

---

## 5. Files Changed in This Session

Files modified by the ContractCatalog blocker fix:

```
 M backend/cmd/server/main.go
 M backend/internal/investment/application/command/compute_fund_aum_test.go
 M backend/internal/investment/application/command/post_transaction_test.go
 M backend/internal/investment/domain/portfolio_repository.go
 M backend/internal/investment/infrastructure/persistence/fund_repository.go
 M backend/internal/investment/module.go
 M backend/internal/workflow/module.go
 M backend/pkg/contract/contracts.go
 M backend/pkg/errcode/coded_error.go
 M backend/pkg/errcode/codes.go
?? backend/internal/investment/infrastructure/adapter/contract_catalog_adapter.go
?? backend/internal/investment/infrastructure/adapter/contract_catalog_adapter_test.go
?? docs/handoff/workflow-backend.md
?? docs/handoff/workflow-current-state.md
?? docs/handoff/workflow-frontend.md
```

---

## 6. Important Architecture Decisions

### ContractCatalog dependency direction

```
Workflow ──► contract.ContractCatalog  (pkg/contract interface — stable)
                        ▲
           investment.ContractCatalogAdapter  (temporary implementation)
                        ↓ (future replacement)
           referenceData.ContractCatalogAdapter
```

**Invariants that must not be violated:**

1. Workflow must never import `backend/internal/investment/...`
2. Investment depends on Workflow (`WorkflowStateProvider`, `WorkflowTradeDayLocker`) — never the reverse
3. The `ContractCatalog` interface lives in `pkg/contract/contracts.go` — never inside any module's internal package
4. To replace the investment implementation with ReferenceData, change only one line in `cmd/server/main.go`

### Admin is group, not position

"Admin" in this system is a `permissions_group` (table `permissions_groups`, seeded in migration `20260301000005_iam__seed_admin`). Admin group members have `"Admin"` in their JWT `rls` claim. Check for admin status: `slices.Contains(claims.Roles, "Admin")`. Never conflate with HR positions or IAM roles in a different sense.

### Username is in JWT

`claims.Username` (from JWT `usr` claim) is populated at login. The `buildActorContext()` bug sets `Username: ""` — a one-line fix that unblocks username display in all workflow responses.

---

## 7. Next Recommended Prompt Title

```
Continue Workflow Backend Implementation from Handoff Docs
```

Use this title when starting a new session to continue the backend implementation. The full context is in the handoff docs.

---

## 8. New Instance Instructions

When starting a fresh session to continue this work:

1. **Read all three handoff docs first:**
   - `docs/handoff/workflow-backend.md` — architecture, remaining work, API contract
   - `docs/handoff/workflow-frontend.md` — frontend structure and requirements
   - `docs/handoff/workflow-current-state.md` — this file, for orientation

2. **Do NOT redo the ContractCatalog work.** It is complete. `go build ./...` and `go test ./...` both pass. Do not re-add the interface, re-create the adapter, or re-wire `main.go`.

3. **Start from remaining backend work** (Section 4 above, in order):
   - First fix `buildActorContext()` username (trivial, high impact)
   - Then check state/operation names
   - Then add `WorkflowApprovalSetting` entity
   - Then implement the new API endpoints

4. **Stop and report if blocked.** If the state machine, value objects, or any required domain type is missing or broken, stop and report before implementing transport layer code. Do not paper over missing domain logic with handler workarounds.

5. **Do not build the frontend yet.** The backend `GET /api/workflow/daily` and `POST /api/workflow/daily/execute` endpoints must exist and pass `make swagger && make api-client` before any frontend implementation begins.
