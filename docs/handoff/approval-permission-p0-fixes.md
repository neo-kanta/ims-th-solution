# Approval + Permission P0 Security Fixes — Handoff

**Updated:** 2026-06-17 (Option A — PORTFOLIO deferred; P0-17 through P0-21 applied)  
**Build status:** `go test ./...` PASS | `npm run test` PASS (177 frontend tests, 40+ backend) | Docker PASS

## Option A Decision

PORTFOLIO approval is deferred from the controlled demo. The `InvestmentSubjectAccessor` has no `PORTFOLIO` case in `resolveContractID` — registering the access port while the default arm denies all access would produce a registered but perpetually-broken subject. Removing the PORTFOLIO registration eliminates this P0 blocker without implementing full PORTFOLIO support. See `docs/handoff/approval-subject-access-port.md` and `docs/architecture/approval-ddd-boundary.md` for deferred implementation steps.

Controlled demo scope: **RESEARCH_REPORT approval**, **INVESTMENT_DECISION approval**, **Permission safety**.

---

## P0-1 — Subject Access Ports Wired in Production Server (original fix)

**File:** `backend/cmd/server/main.go`

**Problem:** `RegisterSubjectAccessPort` was never called after both modules were constructed. Without a registered port, `checkSubjectView` fails closed (returns `ErrForbidden`), meaning every approval read returned 403.

**Fix (original + Option A update):** Register RESEARCH_REPORT and INVESTMENT_DECISION only:
```go
investSubjectAccessor := investmentModule.SubjectAccessor(iamModule)
if investSubjectAccessor != nil {
    approvalModule.RegisterSubjectAccessPort("RESEARCH_REPORT", investSubjectAccessor)
    approvalModule.RegisterSubjectAccessPort("INVESTMENT_DECISION", investSubjectAccessor)
    // PORTFOLIO: not registered — subject access adapter does not support it (see above)
} else {
    slog.Warn("investment subject accessor is nil; approval reads will fail closed")
}
```

Also removed: `approvalModule.RegisterSubjectCallback("PORTFOLIO", ...)` — PORTFOLIO callback exists in the investment module but is not wired until the full implementation is ready.

**Why two types share one accessor:** RESEARCH_REPORT and INVESTMENT_DECISION live in the investment module and share the same IAM data-permission check path. Adding a new subject type requires both a new `RegisterSubjectAccessPort` call here AND a new case in `resolveContractID`.

---

## P0-2 — Investment Subject Access Fail-Closed

**File:** `backend/internal/investment/infrastructure/adapter/investment_subject_access.go`

**Problem:** Three fail-open paths existed:
1. `a.iam == nil` → allowed through without checking any permission.
2. `a.research == nil` or `a.decisions == nil` → returned `uuid.Nil` and skipped the data-permission check.
3. Nil subject lookup result → same as above.

**Fix:**
- All nil-dependency checks now return `errors.New("not found or not accessible")`.
- `HasDataPermission` errors (transient IAM failure) also return `"not found or not accessible"` — fail closed, never assume-allow on error.

**Invariant:** The error string is deliberately generic to avoid information leakage about whether a subject exists.

---

## P0-3 — UserDescriptor / SubjectDescriptor in Approval DTOs

**Files:**
- `backend/internal/approval/transport/dto/response/responses.go`
- `backend/internal/approval/domain/entity/request.go` (no change — entity fields were already present)

**Problem:** Approval responses only had raw UUID fields (`submitter_id`, `actor_user_id`, `signer_user_id`). The frontend had UUID-fallback display logic (`actor_user_id || "System"`) which rendered UUIDs when user resolution was incomplete.

**New types:**
```go
type UserDescriptor struct {
    ID          string `json:"id"`
    Username    string `json:"username"`
    DisplayName string `json:"display_name"`
}
type SubjectDescriptor struct {
    Type         string `json:"type"`
    ID           string `json:"id"`
    DisplayLabel string `json:"display_label"`
}
```

**Mapper helpers:** `userDescriptor(id, name)` and `unknownUserDescriptor(id)` in `responses.go`. When a name is blank or missing, `DisplayName = "Unknown User"` — never a UUID string.

**Raw UUID fields retained** (`submitter_id`, `actor_user_id`, etc.) as routing IDs. Frontend must only use them for `href` / `data-id` attributes, never as visible text.

**Rule:** Run `make api-client` after any change to these structs. Frontend types in `app/features/approval/types.ts` are aliases of the generated `ims-api.d.ts`.

---

## P0-4 — Config Final-Stage Validation (original fix; updated by P0-13)

**File:** `backend/internal/approval/application/service/config_service.go`

**Problem:** `validateStages` allowed:
- Zero `is_final_stage=true` entries (silently promoted highest stage, hiding bad config).
- Multiple stages marked as final (ambiguous terminal detection).
- A final stage at a lower `stage_number` than the max (engine would never reach it via `nextStage()`).

**Original fix (replaced by P0-13):** The zero-final case auto-marked the highest stage. This silently hid misconfigured inputs.

**Updated fix (P0-13, 2026-06-17):** Zero finals now returns a validation error. All three cases return an error. See P0-13 for details.

---

## P0-5 — Approval Config Snapshot

**Files:**
- `backend/internal/approval/domain/entity/request.go` — added `StageSnapshot`, `StageSnapshotFromConfig`, `ToProcessStage`
- `backend/internal/approval/infrastructure/postgres/requests.go` — persists/reads `config_snapshot JSONB`
- `backend/internal/approval/application/service/runtime_service.go` — populates snapshot at submit, uses `resolveStages()` during advancement
- `database/migrations/20260617000001_approval__config_snapshot.up.sql` / `.down.sql`

**Problem:** `ApproveTask` called `s.repo.GetConfig(ctx, r.ProcessConfigID)` live — meaning an admin could edit a process config while a request was in-flight, retroactively changing the stage routing.

**Fix:** At `SubmitApproval` time:
```go
req.ConfigSnapshot = entity.StageSnapshotFromConfig(cfg.Stages)
```

`resolveStages()` checks `r.ConfigSnapshot` first; falls back to live config for pre-existing requests without a snapshot (backwards compatible).

**Migration:** `ALTER TABLE approval__requests ADD COLUMN config_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb;`

---

## P0-6 — Existence Oracle Prevention

**File:** `backend/internal/approval/application/service/runtime_service.go`

**Problem:** When `checkSubjectView` denied access, `GetApprovalRequest` and `GetApprovalTimeline` returned `ErrForbidden`. An attacker who could probe both the "not found" (404) and "forbidden" (403) responses could determine whether a request ID existed even without subject access.

**Fix:**
```go
if err := s.checkSubjectView(...); err != nil {
    return nil, domain.NotFound("approval request not found") // same as "not found" case
}
```

`GetSubjectApprovalStatus` returns `(nil, nil)` on access denial — callers see "no active request" regardless of whether one exists.

**Test updates:** Two tests that previously expected `ErrForbidden` now correctly expect `ErrNotFound` or `nil`. Test comments were updated to explain the security rationale.

---

## P0-7 — `allowed_actions` Submitter / Non-Submitter Guards

**File:** `backend/internal/approval/application/service/runtime_service.go`

**Problem:** `cancel` was shown to all non-terminal viewers; `revoke` was shown to all APPROVED viewers including the submitter.

**Fix (at time of P0-7):**
```go
// cancel: submitter can cancel their own non-terminal request
if !req.Status.IsTerminal() && req.SubmitterID == viewerID {
    actions = append(actions, "cancel")
}
// revoke: non-submitter only
if req.Status == vo.RequestStatusApproved && req.SubmitterID != viewerID {
    actions = append(actions, "revoke")
}
```

**Note:** The `revoke` arm above was subsequently removed by **P0-16** (2026-06-17) because `computeAllowedActions` has no access to the function-permission checker — surfacing revoke to all non-submitters without a permission check is a violation of RBAC. See P0-16 for the current behavior and P1 remediation path.

**Rationale:** Allowing the submitter to revoke their own request would defeat the maker-checker principle.

---

## P0-8 — Permission Terminal-State Guard and Reject Reason

**File:** `backend/internal/permissions/application/service/permission_service.go`

**Problem:**
1. `finish()` would overwrite a terminal request's status (double-approve / overwrite rejection).
2. `Reject()` accepted a blank comment, leaving no audit reason.

**Fix 1 — terminal guard:**
```go
switch cr.Status {
case domain.RequestStatusApproved, domain.RequestStatusRejected,
    domain.RequestStatusMerged, domain.RequestStatusCancelled, domain.RequestStatusClosed:
    return domain.InvalidTransition("cannot " + strings.ToLower(status) + " a request that is already in a terminal state (" + cr.Status + ")")
}
```

**Fix 2 — reject requires reason:**
```go
func (s *Service) Reject(ctx context.Context, in DecisionInput) (*domain.ChangeRequest, error) {
    if strings.TrimSpace(in.Comment) == "" {
        return nil, domain.Invalid("a rejection reason is required")
    }
    return s.decide(ctx, in, domain.ReviewerStatusRejected)
}
```

---

## P0-9 — Frontend UUID-Free Display

**Files changed:**

| File | UUID removed | Replaced with |
|---|---|---|
| `app/pages/approval/index.vue:155` | `req.contract_id` as module label | `req.contract_type \|\| req.process_type \|\| "SYSTEM"` |
| `app/features/approval/lib/approvalMappers.ts:19` | `e.actor_user_id` fallback | `e.actor?.display_name \|\| e.actor_name \|\| "System"` |
| `app/features/approval/lib/approvalMappers.ts:31` | `e.delegated_from_user_id` in meta | `e.delegated_from?.display_name` |
| `app/features/approval/lib/approvalMappers.ts:49` | `s.signer_user_id` fallback | `s.signer?.display_name \|\| s.signer_display_name \|\| "Unknown User"` |
| `app/features/approval/components/ApprovalRequestDetail.vue:94-97` | raw `contract_id` block | `subject?.display_label \|\| subject_title` |
| `app/pages/approval/config/processes.vue:110` | `p.contract_id` label | `"contract-specific"` |
| `app/pages/approval/config/teams.vue:183` | `c.contract_id` label | `"Effective " + c.effective_date` |
| `app/features/permissions/components/PermissionRequestDetailScreen.vue:172` | `request.created_by` UUID fallback | `"Unknown User"` |
| `app/features/permissions/components/PermissionRequestDetailScreen.vue:173` | `request.target_entity_id` | removed (type label only) |
| `app/features/settings/components/SettingsDataPermissionsPanel.vue:53` | `contractId` sublabel | `head.scope` |
| `app/features/settings/components/SettingsDataPermissionsPanel.vue:142` | `grant.contractId` / `grant.user` secondary | `grant.scope` / `grant.userLabel` |

**Rule:** A raw UUID must never appear inside a rendered text node. UUIDs are acceptable in:
- `href` and `to` attributes (navigation routing)
- `data-*` attributes (machine-readable markers)
- Hidden `input` values (form submission)
- Developer tools / JSON responses (API clients, not end-user UI)

---

## P0-10 — Frontend Regression Tests

**New test files:**
- `frontend/tests/approval-no-uuid.test.ts` — 8 tests for `toTimelineEvents` and `toStamps` mappers. Each test calls `assertNoUUID()` which throws if a UUID pattern is returned as a display value.
- `frontend/tests/permission-no-uuid.test.ts` — 4 tests for the permission creator label and target entity display rules.

**Updated tests:**
- `frontend/tests/approval-components.test.ts` — two tests updated to reflect correct "Unknown User" / "System" fallbacks instead of UUID fallbacks.
- `backend/internal/approval/application/service/runtime_service_test.go` — four tests updated to expect `ErrNotFound` (existence oracle) instead of `ErrForbidden`.

---

---

## P0-11 — PORTFOLIO Approval Deferred (Option A, 2026-06-17)

**Files:**
- `backend/cmd/server/main.go` — removed `RegisterSubjectCallback("PORTFOLIO", ...)` and `RegisterSubjectAccessPort("PORTFOLIO", ...)`
- `database/seeds/013_approval_portfolio_onboarding_seed.sql` — added `UPDATE ... SET is_active = false` for `PROC_PORTFOLIO_ONBOARDING_DEFAULT`

**Reason:** `InvestmentSubjectAccessor.resolveContractID` has no `case "PORTFOLIO":` arm. Registering the access port while the default denies all access means every PORTFOLIO approval attempt returns "not found or not accessible." A broken registered subject type is worse than a deferred unsupported one.

See `docs/architecture/approval-ddd-boundary.md` and `docs/handoff/approval-subject-access-port.md` for full deferred implementation steps.

---

## P0-12 — Permission Nil Checker Fail-Closed (2026-06-17)

**File:** `backend/internal/permissions/application/service/permission_service.go`

**Problem:** `requirePermission` returned `nil` when `s.checker == nil`, allowing any authenticated user to invoke Merge (which applies all pending permission changes to production tables) if the checker was not wired.

**Fix:**
```go
if s.checker == nil {
    return domain.Forbidden("permission checker is not configured")
}
```

**Test added:** `TestRequirePermission_NilChecker_ReturnsForbidden` in `permission_service_test.go`.

---

## P0-13 — Approval Final-Stage Validation Rejects Missing Final Stage (2026-06-17)

**File:** `backend/internal/approval/application/service/config_service.go`

**Problem:** `validateStages` auto-promoted the highest stage to `is_final_stage = true` when no stage was explicitly marked, silently hiding misconfigured multi-stage configs.

**Fix:**
```go
case finalCount == 0:
    return nil, domain.Validation("exactly one stage must be marked as is_final_stage")
```

**Tests updated:**
- `TestValidateStages_SingleStage_NoFinalMarked_AutoMarks` → renamed to `TestValidateStages_SingleStage_NoFinalMarked_ReturnsError`, inverted to expect error
- `TestValidateStages_ThreeStage_AutoMarkOnZeroFinals` → renamed to `TestValidateStages_ThreeStage_NoFinalMarked_ReturnsError`, inverted to expect error

**Tests added:**
- `TestValidateStages_SingleStage_ExplicitFinal_Passes` — single stage with `is_final_stage: true` passes validation

**Frontend impact:** `ApprovalStageBuilder.vue` already defaults the first added stage to `is_final_stage: true` (line 32: `is_final_stage: stages.value.length === 0`) and auto-reassigns on remove, so valid form submissions are not affected.

---

## P0-14 — Approval Config + Stage Builder UUID Inputs Removed (2026-06-17)

**Files:**
- `frontend/app/features/approval/components/ApprovalProcessConfigForm.vue`
- `frontend/app/features/approval/components/ApprovalStageBuilder.vue`

**Problem:** Operators were asked to type raw UUIDs to configure approval scope (contract UUID) and approvers (user UUID).

**Fix:**
- `ApprovalProcessConfigForm.vue`: replaced `"Applicable contract UUID"` free-text input with a disabled "Scope selector source not available" field under label "Applicable Scope"
- `ApprovalStageBuilder.vue`: replaced `"Approver user UUID"` free-text input with a disabled "User selector source not available" field under label "Approver"

These fields are disabled until a contract/user autocomplete selector is implemented (P1).

---

## P0-15 — Permission Changes Tab target_id Removed (2026-06-17)

**File:** `frontend/app/features/permissions/components/PermissionRequestDetailScreen.vue`

**Problem:** Changes tab rendered `{{ item.target_table }} {{ item.target_id }}`, exposing the raw UUID as visible text.

**Fix:** Removed `{{ item.target_id }}`. Now renders only `{{ item.target_table }}`.

**Test added:** Two tests in `frontend/tests/permission-no-uuid.test.ts`:
- `does not render item.target_id as visible text`
- `renders target_table as the type label`

Future (P1): Add `target_display_label` to the backend DTO and render it with "Unknown Subject" fallback.

---

## P0-16 — allowed_actions Revoke Removed from Computed Set (2026-06-17)

**File:** `backend/internal/approval/application/service/runtime_service.go`

**Problem:** `computeAllowedActions` showed `revoke` to any non-submitter on an APPROVED request, regardless of whether the viewer had the function permission to revoke.

**Fix:** Removed the revoke arm from `computeAllowedActions`. Revoke is enforced at the route middleware level only. The UI does not surface a revoke button via `allowed_actions`.

**Test updated:** `TestAllowedActions_ComputedByStatus` — previously asserted revoke was present for APPROVED + non-submitter. Now asserts revoke is NOT in `allowed_actions`.

**Two-function clarification:** `subjectAllowedActions` (called by `GetSubjectStatus`) is a separate function that still emits `"revoke"` for APPROVED requests. This is NOT what `ApprovalRequestDetail.vue` uses — the detail view reads `allowed_actions` exclusively from the `GetRequest` endpoint (which calls `computeAllowedActions`). Therefore `canRevoke` is always false and the revoke button is never rendered, even though the subject-status DTO lists it. The subject-status `allowed_actions` field is currently not consumed by any UI surface.

**Remaining limitation (P1):** To surface a revoke button for authorized users, `computeAllowedActions` needs a `viewerCanRevoke bool` parameter populated by a function-permission check at the call site in `GetApprovalRequest`.

---

## P0-17 — `ListRequests` Missing Auth Guard (2026-06-17)

**File:** `backend/internal/approval/transport/handler/runtime_handler.go`

**Problem:** `ListRequests` used `actor, _ := actorID(r)` — discarding the `ok` boolean. When no JWT claims were present in the context (unauthenticated request that bypassed the auth middleware), `actor` became `uuid.Nil`. The service's `ListRequests` treats `ViewerID == uuid.Nil` as a system/admin context and returns all requests unfiltered, leaking all approval requests to unauthenticated callers.

**Fix:**
```go
actor, ok := actorID(r)
if !ok {
    httputil.Unauthorized(w, "not authenticated")
    return
}
```

**Test added:** `TestListRequests_NoActor_Returns401` in `transport/handler/runtime_handler_test.go`.

**Note:** `uuid.Nil` bypass in the service is intentional for internal/system callers that call the service layer directly (e.g., batch jobs). The handler guard ensures the bypass is only reachable from internal Go code, never from an HTTP request without a valid actor.

---

## P0-18 — `ApprovalGroupMemberTable` Raw UUID Display (2026-06-17)

**File:** `frontend/app/features/approval/components/ApprovalGroupMemberTable.vue`

**Problem:**
1. Member rows rendered `{{ m.user_id }}` as a visible secondary label below the display name.
2. The add-member form used `<AppFormField label="User UUID">` with `placeholder="e.g. 00000000-0000..."` — prompting operators to enter raw UUIDs.

**Fix:**
1. Removed `<span class="member-table__uid">{{ m.user_id }}</span>` from member rows.
2. Changed label to `"User"` and replaced the free-text input with a disabled input (`placeholder="User selector source not available"`).

**Demo note:** Group members are seeded; the disabled input does not block the demo flow.

---

## P0-19 — `ApprovalTeamMemberTable` Raw UUID Display (2026-06-17)

**File:** `frontend/app/features/approval/components/ApprovalTeamMemberTable.vue`

**Problem:**
1. Member rows rendered `{{ m.user_id }}` as a secondary span below the display name.
2. Add-member input used `placeholder="User UUID"`.

**Fix:**
1. Removed `<span class="tm-table__uid">{{ m.user_id }}</span>` from member rows.
2. Changed placeholder to `"User selector source not available"` and set `disabled`.

---

## P0-20 — Delegation/Proxy `principalUserName` Raw UUID (2026-06-17)

**File:** `frontend/app/features/approval/composables/useApprovalSession.ts`

**Problem:** Six `principalUserName` assignments in the stage-building loop used raw UUID fields as display values:
- `sig.proxy_for_user_id` (3 sites: SINGLE/GROUP/TEAM_STAMP signature cases)
- `tsk.delegated_from_user_id` (2 sites: SINGLE/TEAM_STAMP task cases)
- `evt.delegated_from_user_id` (1 site: history event mapper)

`principalUserName` is rendered in `ApprovalStageCard.vue` as `for {{ approver.principalUserName }}`.

**Fix:** Extracted `derivePrincipalName(isDelegated, descriptor)` from `approvalMappers.ts`:
```ts
export function derivePrincipalName(
  isDelegated: boolean | undefined,
  descriptor: { display_name?: string } | null | undefined,
): string | undefined {
  if (!isDelegated) return undefined;
  return descriptor?.display_name || "Unknown User";
}
```
All six assignment sites now use `derivePrincipalName(flag, descriptor)`. Returns `undefined` when no delegation is in effect — preventing "for Unknown User" on non-delegated rows.

**Key invariant:** `principalUserName` is `undefined` (not `"Unknown User"`) when the approver is not acting as a delegate/proxy. The `v-if="approver.principalUserName"` guard in `ApprovalStageCard.vue` hides the "for" tag when `principalUserName` is falsy.

**Tests added:** `derivePrincipalName – delegation display is UUID-free` suite in `tests/approval-no-uuid.test.ts` (6 tests).

---

## P0-21 — Signature Fallback Persists Raw UUID (2026-06-17)

**File:** `backend/internal/approval/application/service/runtime_service.go`

**Problem:** `writeSignature()` used `display := signer.String()` as the initial value for `SignerDisplayName`. When the directory lookup failed (directory is nil, returns error, or returns no DisplayName), the raw UUID was stored in `approval__signature_records.signer_display_name` and subsequently included in `SignatureResponse.signer_display_name`.

Because `userDescriptor(id, name)` in `responses.go` uses `name` as-is when non-empty (no "Unknown User" fallback for non-empty names), the UUID propagated into the API response even though `userDescriptor` has a guard.

**Fix:**
```go
// was: display := signer.String()
display := "Unknown User"
if s.directory != nil {
    if info, err := s.directory.GetUser(ctx, signer); err == nil && info != nil && info.DisplayName != "" {
        display = info.DisplayName
    }
}
```

**Tests added:**
- `TestWriteSignature_NilDirectory_SignerDisplayNameIsUnknownUser` — approves a request with no directory wired; asserts `SignerDisplayName == "Unknown User"` and `!= signer.String()`.
- `TestWriteSignature_ErrorDirectory_SignerDisplayNameIsUnknownUser` — same with `alwaysErrorDirectory{}` that always returns error.

**Note on pre-existing rows:** Old rows in `approval__signature_records` with a UUID already persisted as `signer_display_name` will still be served via `fromSignature()` → `userDescriptor(id, uuid-string)` where the non-empty UUID passes through as display name. This is a fresh-demo-DB only risk; no migration is planned.

---

## Verification Checklist (updated 2026-06-17)

```bash
# Backend: must all pass
cd backend && go build ./...
cd backend && go test ./internal/approval/... ./internal/permissions/... ./internal/investment/...
cd backend && go test ./...

# Frontend: must all pass
cd frontend && npm run test    # 177 tests total
cd frontend && npm run build

# Docker
cd infra && docker-compose up -d --build
```

**2026-06-17 results (P0-17 through P0-21):** All backend tests pass (approval service: P0-5 signature tests added; handler: P0-1 401 test added). Frontend 177 tests pass. Frontend build clean. Docker: all 4 containers started healthy.

---

## Remaining Limitations

| Priority | Item | Status |
|---|---|---|
| P1 | PORTFOLIO full implementation | Deferred — see approval-subject-access-port.md |
| P1 | Revoke button in allowed_actions | Not surfaced; route gate only |
| P1 | Action-specific function rights in fetcher SQL | any-flag SQL, not per-action |
| P1 | Signature stamp immutability triggers | `approval__signature_records` has no UPDATE/DELETE triggers |
| P1 | DelegatedFrom/Actor display name | SQL query lacks display name join — shows "Unknown User" |
| P1 | target_display_label in permission DTO | Only target_table shown in Changes tab |
| P1 | User selector for group/team members | Add-member inputs disabled; real user-search autocomplete required (P0-18/19) |
| P1 | Old signature rows with persisted UUID display names | Pre-fix rows in `approval__signature_records` retain UUID in `signer_display_name`; no migration planned |
| P2 | Disabled action reasons | `AllowedActions []string` only; no reason structure |
| P2 | Config snapshot isolation test | No A→B mutation regression test |
