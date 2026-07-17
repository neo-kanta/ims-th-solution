# Workflow Backend Handoff

**Branch:** `neo-develop`
**As of:** 2026-06-15
**Build status:** `go build ./...` — PASS | `go test ./...` — PASS (zero failures)

---

## 1. Purpose

The Workflow module is the **daily control gate** for the entire IMS application.

- Every investment transaction requires the day to be in `INVESTMENT_DAY_STARTED` state before it is permitted.
- Manager approval locks further transactions for that business date.
- Transaction closing and Accounting closing are the final checkpoints before the day is archived.
- **Frontend and external callers must query workflow state using business-readable keys** (`businessDate`, `contractCode`) — never a UUID or `workflowInstanceId`.

---

## 2. Business Workflow Stages

The four mandatory stages flow in order:

```
NOT_STARTED
    ↓  START_INVESTMENT_DAY
INVESTMENT_DAY_STARTED
    ↓  MANAGER_APPROVE
MANAGER_APPROVED
    ↓  CLOSE_TRANSACTION
TRANSACTION_CLOSED
    ↓  CLOSE_ACCOUNTING
ACCOUNTING_CLOSED
```

Each stage has a corresponding cancel/rollback operation:

| Forward operation        | Reverse operation            |
|--------------------------|------------------------------|
| `START_INVESTMENT_DAY`   | `CANCEL_INVESTMENT_DAY`      |
| `MANAGER_APPROVE`        | `CANCEL_MANAGER_APPROVAL`    |
| `CLOSE_TRANSACTION`      | `CANCEL_TRANSACTION_CLOSE`   |
| `CLOSE_ACCOUNTING`       | `CANCEL_ACCOUNTING_CLOSE`    |

---

## 3. Completed Work (this session)

The blocker preventing business-readable API was that Workflow had no way to resolve a `contractCode` string to an internal `uuid.UUID` without importing investment internals.

**Completed:**

- Added `ContractCatalog` interface to `backend/pkg/contract/contracts.go`
- Added `CodeContractNotFound` error code to `backend/pkg/errcode/codes.go` + `coded_error.go`
- Added `GetByContractCode` method to `FundRepository` domain interface
- Implemented `GetByContractCode` SQL lookup in Postgres fund repository
- Created `ContractCatalogAdapter` in investment infrastructure adapter layer
- Added `ContractCatalog()` export method to investment `module.go`
- Added `contractCatalog` field, `SetContractCatalog()`, and `ResolveContractCode()` to workflow `module.go`
- Wired `workflowModule.SetContractCatalog(investmentModule.ContractCatalog())` in `cmd/server/main.go`
- Added 6 unit tests for `ContractCatalogAdapter` (all pass)
- Updated two existing test stubs (`fundAUMFundRepo`, `postFundRepo`) to implement the new interface method

---

## 4. Important Files Changed

| File | Change |
|------|--------|
| `backend/pkg/contract/contracts.go` | Added `ContractCatalog` interface |
| `backend/pkg/errcode/codes.go` | Added `CodeContractNotFound = "CONTRACT_NOT_FOUND"` |
| `backend/pkg/errcode/coded_error.go` | Added `CONTRACT_NOT_FOUND` → HTTP 404 in `DefaultStatus()` |
| `backend/internal/investment/domain/portfolio_repository.go` | Added `GetByContractCode(ctx, contractCode string) (*entity.Fund, error)` to `FundRepository` |
| `backend/internal/investment/infrastructure/persistence/fund_repository.go` | Implemented `GetByContractCode` via `WHERE contract_code = $1 AND deleted_at IS NULL LIMIT 1` |
| `backend/internal/investment/infrastructure/adapter/contract_catalog_adapter.go` | **NEW** — `ContractCatalogAdapter` + `ErrContractNotFound` + `ErrContractCodeEmpty` |
| `backend/internal/investment/infrastructure/adapter/contract_catalog_adapter_test.go` | **NEW** — 6 unit tests |
| `backend/internal/investment/module.go` | Added `ContractCatalog() contract.ContractCatalog` method |
| `backend/internal/workflow/module.go` | Added `contractCatalog` field, `SetContractCatalog()`, `ResolveContractCode()` |
| `backend/cmd/server/main.go` | Added `workflowModule.SetContractCatalog(investmentModule.ContractCatalog())` |
| `backend/internal/investment/application/command/compute_fund_aum_test.go` | Added `GetByContractCode` stub to `fundAUMFundRepo` |
| `backend/internal/investment/application/command/post_transaction_test.go` | Added `GetByContractCode` stub to `postFundRepo` |

---

## 5. Architecture Decision — Dependency Direction

```
Workflow ──► ContractCatalog (pkg/contract interface)
                    ▲
              Investment module (temporary implementation)
                    ↓ (future)
              ReferenceData/ContractMaster module
```

**Rules (must not be violated):**

- Workflow must **never** import `internal/investment/...`
- Workflow depends only on `contract.ContractCatalog` from `backend/pkg/contract`
- Investment's `ContractCatalogAdapter` is the **temporary** implementation — it may later be replaced by a ReferenceData module without touching Workflow code
- To migrate the implementation to ReferenceData, only one line in `cmd/server/main.go` changes:
  ```go
  // Before:
  workflowModule.SetContractCatalog(investmentModule.ContractCatalog())
  // After:
  workflowModule.SetContractCatalog(referenceDataModule.ContractCatalog())
  ```

Also note: Investment already depends on Workflow (`WorkflowStateProvider`, `WorkflowTradeDayLocker`) for transaction gating. To avoid circular dependency, Workflow must never take a return dependency on Investment internals.

---

## 6. Remaining Backend Work (ordered)

### 6a. Fix actor username (quick fix)

In `backend/internal/workflow/transport/handler/workflow_handler.go`, the `buildActorContext()` function currently sets `Username: ""`.

**Fix:** Read `claims.Username` from the JWT claims. The JWT already carries `usr` (username) and `rls` (group names) from login. Both are available via `middleware.GetUserClaims(r.Context())`.

```go
// Current (broken):
return vo.ActorContext{
    UserID:    userID,
    Username:  "",          // ← BUG
    ...
}
// Fix:
return vo.ActorContext{
    UserID:    userID,
    Username:  claims.Username,   // from JWT "usr" claim
    ...
}
```

Admin group check is also JWT-based: `slices.Contains(claims.Roles, "Admin")`. No DB lookup needed.

### 6b. Verify / rename state values

Check `backend/internal/workflow/domain/valueobject/workflow_state.go`.
Migration `20260613000007_workflow__expand_workflow_states` has expanded the DB CHECK constraint to allow both old and new names (Phase 1). Ensure Go constants use the canonical names:

| Required canonical name      | Old name (compat alias) |
|------------------------------|-------------------------|
| `INVESTMENT_DAY_STARTED`     | `DAY_OPEN`              |
| `MANAGER_APPROVED`           | (same)                  |
| `TRANSACTION_CLOSED`         | (same)                  |
| `ACCOUNTING_CLOSED`          | (same)                  |

### 6c. Verify / rename operation values

Check `backend/internal/workflow/domain/valueobject/workflow_action.go`.
Ensure operation names match:

| Required name              | Current name (may differ) |
|----------------------------|---------------------------|
| `START_INVESTMENT_DAY`     | `OPEN_DAY`                |
| `CANCEL_INVESTMENT_DAY`    | `CANCEL_DAY_START`        |
| `MANAGER_APPROVE`          | `APPROVE`                 |
| `CANCEL_MANAGER_APPROVAL`  | `CANCEL_APPROVAL`         |
| `CLOSE_TRANSACTION`        | `CLOSE_TRANSACTIONS`      |
| `CANCEL_TRANSACTION_CLOSE` | (same)                    |
| `CLOSE_ACCOUNTING`         | (same)                    |
| `CANCEL_ACCOUNTING_CLOSE`  | `ROLLBACK_ACCOUNTING_CLOSE` |

### 6d. Add `WorkflowApprovalSetting` entity, table, repository

This is a new entity internal to the workflow module. No cross-module dependency required.

Required fields:
- `operationType` — which workflow operation this setting applies to
- `approverAccountCode` / `approverUsername` — snapshot of who is allowed (IAM user)
- `approverRole` — e.g., "Admin"
- `isAdminOverride` — whether Admin group is default approver
- `approvalMode` — `ANY_OF` or `ALL_OF`

Migration: `database/migrations/<timestamp>_workflow__create_approval_settings.up.sql`

### 6e. Add `actor_account_code` to transition log

In `workflow__transition_log`, `actor_username` exists. Add `actor_account_code VARCHAR(255)` column (same value as username in this system; separate column for forward-compatibility). Also add an `is_admin_override BOOLEAN DEFAULT false` column for admin-override audit trail.

Migration: `database/migrations/<timestamp>_workflow__add_actor_account_code.up.sql`

### 6f. New endpoint: `GET /api/workflow/daily`

```
GET /api/workflow/daily?businessDate=YYYY-MM-DD&contractCode=ABC
```

- Use `Module.ResolveContractCode(ctx, contractCode)` to get the internal UUID
- Get-or-create the workflow day (return `NOT_STARTED` shape if row does not exist yet)
- Return the full aggregated daily response (see Section 7)

### 6g. New endpoint: `POST /api/workflow/daily/execute`

```
POST /api/workflow/daily/execute
Body: { "businessDate": "YYYY-MM-DD", "contractCode": "ABC", "operationType": "START_INVESTMENT_DAY", "remark": "..." }
```

- Resolve `contractCode` → UUID
- Route to the matching command handler based on `operationType`
- Require `remark` for cancel/rollback operations
- Apply admin override logic (check JWT `Roles` contains "Admin")

### 6h. New endpoints: `GET /PUT /api/workflow/settings`

```
GET  /api/workflow/settings?contractCode=ABC
PUT  /api/workflow/settings?contractCode=ABC
Body: { "defaultApproverRole": "Admin", "approvalMode": "ANY_OF", "configuredApprovers": [...] }
```

Requires `WorkflowApprovalSetting` entity (Step 6d).

### 6i. Daily aggregated response

The `GET /api/workflow/daily` response must return a single rich object (see Section 7). Key fields:

- `moduleReadiness.approval` — report `{ "ready": false, "reason": "approval module is scaffold" }` honestly; do NOT fake it
- `moduleReadiness.permission`, `moduleReadiness.investment`, `moduleReadiness.compliance` — live readiness check

### 6j. Admin override logic

- Check `slices.Contains(claims.Roles, "Admin")` from JWT
- If admin: allow any workflow operation regardless of `WorkflowApprovalSetting`
- Log `is_admin_override = true` in the transition log row

### 6k. Swagger/OpenAPI annotations

Add `@Summary`, `@Tags workflow`, `@Param`, `@Success`, `@Failure`, `@Router` annotations to all new handlers.
Run `make swagger && make api-client` after.

### 6l. Tests

See Section 9 for the required test checklist.

---

## 7. Target API Contract

### `GET /api/workflow/daily`

```
GET /api/workflow/daily?businessDate=2026-06-15&contractCode=ABC
Authorization: Bearer <token>
```

Response `200 OK`:
```json
{
  "businessDate": "2026-06-15",
  "contractCode": "ABC",
  "currentState": "INVESTMENT_DAY_STARTED",
  "isToday": true,
  "allowedOperations": ["MANAGER_APPROVE", "CANCEL_INVESTMENT_DAY"],
  "blockedReasons": [],
  "timeline": [
    {
      "operationType": "START_INVESTMENT_DAY",
      "fromState": "NOT_STARTED",
      "toState": "INVESTMENT_DAY_STARTED",
      "executedAt": "2026-06-15T08:00:00Z",
      "executedByAccountCode": "jsmith",
      "executedByUsername": "jsmith",
      "isAdminOverride": false,
      "remark": ""
    }
  ],
  "approvers": [
    {
      "operationType": "MANAGER_APPROVE",
      "approverAccountCode": "admin",
      "approverUsername": "admin",
      "role": "Admin",
      "isAdminOverride": true
    }
  ],
  "settings": {
    "defaultApproverRole": "Admin",
    "approvalMode": "ANY_OF",
    "configuredApprovers": []
  },
  "auditSummary": {
    "lastActorUsername": "jsmith",
    "lastAction": "START_INVESTMENT_DAY",
    "lastActionAt": "2026-06-15T08:00:00Z"
  },
  "moduleReadiness": {
    "investment": { "ready": true },
    "approval": { "ready": false, "reason": "approval module is scaffold" },
    "compliance": { "ready": true },
    "permission": { "ready": true }
  }
}
```

### `POST /api/workflow/daily/execute`

```
POST /api/workflow/daily/execute
Authorization: Bearer <token>
Content-Type: application/json

{
  "businessDate": "2026-06-15",
  "contractCode": "ABC",
  "operationType": "MANAGER_APPROVE",
  "remark": ""
}
```

Response `201 Created`:
```json
{
  "transitionId": "...",
  "fromState": "INVESTMENT_DAY_STARTED",
  "toState": "MANAGER_APPROVED",
  "executedAt": "2026-06-15T09:00:00Z",
  "executedByAccountCode": "admin",
  "isAdminOverride": false
}
```

Error responses use standard `errcode` envelope:
- `400 INVALID_REQUEST` — missing/invalid fields
- `404 CONTRACT_NOT_FOUND` — unknown contractCode
- `409 WORKFLOW_VIOLATION` — invalid state transition
- `403 FORBIDDEN` — user not permitted for this operation

### `GET /api/workflow/settings`

```
GET /api/workflow/settings?contractCode=ABC
Authorization: Bearer <token>
```

Response `200 OK`:
```json
{
  "contractCode": "ABC",
  "defaultApproverRole": "Admin",
  "approvalMode": "ANY_OF",
  "configuredApprovers": [
    { "accountCode": "jsmith", "username": "jsmith", "operationTypes": ["MANAGER_APPROVE"] }
  ]
}
```

### `PUT /api/workflow/settings`

```
PUT /api/workflow/settings?contractCode=ABC
Authorization: Bearer <token>
Content-Type: application/json

{
  "defaultApproverRole": "Admin",
  "approvalMode": "ANY_OF",
  "configuredApprovers": [
    { "accountCode": "jsmith", "operationTypes": ["MANAGER_APPROVE"] }
  ]
}
```

Response `200 OK` — same shape as GET.

---

## 8. Required Business Rules

1. **Get-or-create:** `GET /api/workflow/daily` always succeeds even if no DB row exists yet — return `NOT_STARTED` synthetic response.
2. **One instance per date+contract:** Unique constraint on `(contract_id, business_date)` — enforce at DB and application level.
3. **Business-readable identity:** All external API calls use `businessDate` + `contractCode`. Never expose `workflowInstanceId` or `contractId` UUID to the frontend.
4. **Username/accountCode display:** All timeline entries and response fields show `accountCode` / `username` (the IAM login name). Never show raw UUID in user-facing fields. Both are available from JWT claims at request time.
5. **Admin group is default approver:** The "Admin" group (from `permissions_groups` table, seeded in `20260301000005_iam__seed_admin`) can approve all workflow operations by default.
6. **Admin can override any restriction:** If `slices.Contains(claims.Roles, "Admin")` is true, allow the operation regardless of `WorkflowApprovalSetting`. Log `is_admin_override = true`.
7. **Admin override must be audited:** Every admin override writes `is_admin_override = true` on the transition log row and records the actor's accountCode.
8. **Non-admin must be configured:** Non-admin users must appear in `WorkflowApprovalSetting.configuredApprovers` for the operation, or have `WORKFLOW_EXECUTE` permission explicitly granted.
9. **Transaction lock after Manager Approval:** Once state reaches `MANAGER_APPROVED`, the investment module's `IsTransactionLocked()` returns `true` — no new transactions are allowed.
10. **Do not fake Approval module readiness:** The Approval module is still scaffold. `moduleReadiness.approval` must report `{ "ready": false, "reason": "approval module is scaffold" }` — never `true`.
11. **Admin is role/group, not position:** "Admin" in IMS is a `permissions_group` name, embedded in JWT `rls` claim. Do not conflate with "position" in any HR/IAM sense.

---

## 9. Tests Still Required

- [ ] Get-or-create returns `NOT_STARTED` for a new business date (no DB row)
- [ ] Get-or-create returns existing state when row exists
- [ ] Duplicate workflow day creation is prevented (unique constraint)
- [ ] `allowedOperations` matches expected state machine transitions
- [ ] Invalid state transition returns `WORKFLOW_VIOLATION` error
- [ ] Timeline entries include `executedByAccountCode` (not empty string)
- [ ] Admin user (`claims.Roles` contains `"Admin"`) can execute any operation
- [ ] Non-admin without approver configuration is rejected (`FORBIDDEN`)
- [ ] `WorkflowApprovalSetting` update is persisted and returned in next GET
- [ ] `is_admin_override = true` is written to transition log when admin executes

---

## 10. Known Risks / Notes

- **Approval module is scaffold.** `backend/internal/approval/domain/repository.go` is empty. Do not attempt to integrate workflow with the approval module in this implementation batch. Report `moduleReadiness.approval.ready = false`.
- **Admin is role/group, not position.** Checking "is admin" = `slices.Contains(claims.Roles, "Admin")`. No DB lookup needed. The group name "Admin" is seeded and stable.
- **ContractCatalog is temporary.** The investment module implements it only because the `contract_code` column lives in `investment__funds`. A future ReferenceData module will own this lookup.
- **State name migration is Phase 1 only.** Migration `20260613000007` expanded the DB CHECK to allow both `DAY_OPEN` and `INVESTMENT_DAY_STARTED`. The Go code should write canonical new names; Phase 2 migration (data backfill + constraint narrowing) is a separate PR.
- **`fundSelect` does not include `contract_code`.** The `GetByContractCode` SQL filters by `contract_code` in the WHERE clause but does not select it (the existing 20-column scan is reused). This is intentional — `contract_code` is not part of the Fund domain entity yet.
- **Frontend existing workflow page exists** at `frontend/app/pages/workflow/index.vue` and `frontend/app/features/workflow/`. It currently points users to per-fund views. The new daily API will power a redesigned version of this page.
