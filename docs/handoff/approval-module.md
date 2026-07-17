# Approval Module Handoff

**Branch:** `neo-develop`
**As of:** 2026-06-17
**Build status:** `go test ./...` — PASS | `npm run test` — PASS (177 frontend tests, 40+ backend)

---

## 1. Purpose

The Approval module is a **generic, reusable approval engine** that can approve any IMS business object through a single, auditable workflow. Business objects are referenced via a polymorphic `(subject_type, subject_id)` pair — the approval engine itself never imports other modules' internals.

The module enforces three non-negotiable controls:

| Control | Rule |
|---|---|
| Maker-checker | The user who submits a request can never approve it |
| Backend-gated actions | `allowed_actions` is computed server-side and drives all UI button visibility — never duplicated in the frontend |
| Immutable audit trail | `approval__events` rows cannot be updated or deleted (DB trigger enforced) |

---

## 2. What Is Wired Today

### Subjects with active approval flows

| Subject Type | Process Type | Trigger | Status |
|---|---|---|---|
| `RESEARCH_REPORT` | `INVESTMENT_ANALYSIS_REPORT` | `ResearchReportCommandHandler.Submit` | **ACTIVE — demo-ready** |
| `INVESTMENT_DECISION` | `INVESTMENT_DECISION` | `DecisionCommandHandler.Submit` | **ACTIVE — demo-ready** |
| `PORTFOLIO` | `PORTFOLIO_ONBOARDING` | portfolio onboarding flow | **DEFERRED — not wired (Option A)** |

RESEARCH_REPORT and INVESTMENT_DECISION register callbacks, validators, and access ports in `main.go`. PORTFOLIO callback exists (`PortfolioApprovalCallback`) but is not registered — the subject access port does not support it yet. See `docs/handoff/approval-subject-access-port.md` for deferred implementation steps.

The `PROC_PORTFOLIO_ONBOARDING_DEFAULT` seed config is kept for reference but deactivated (`is_active = false`).

---

## 3. Architecture

### Backend layers

```
backend/internal/approval/
  domain/entity/          ApprovalRequest, ApprovalTask, ApprovalEvent,
                          ApprovalSignatureRecord, ApprovalProcessConfig, ...
  domain/valueobject/     ProcessType, SubjectType, ApproverMode, RequestStatus,
                          TaskStatus, EventType, SignatureLabel (stable string enums)
  domain/ports.go         PermissionPort, UserDirectory, LeaveChecker,
                          DelegateResolver, Notifier, AuditPort, SubjectSync
  domain/errors.go        Typed sentinel errors (DomainError with Kind + Message)
  domain/repository.go    Repository interface (all SQL behind this boundary)
  application/service/    ApprovalRuntimeService  ← main lifecycle engine
                          ApprovalConfigService   ← CRUD for groups/teams/processes
  infrastructure/adapter/ AuditLoggerAdapter, PostgresUserDirectory,
                          PostgresDelegateResolver
  infrastructure/postgres/ SQL-backed repository (split by concern: requests,
                           tasks, events, signatures, groups, teams, process)
  transport/handler/      RuntimeHandler, ConfigHandler (HTTP only — no business logic)
  transport/dto/          request/, response/ (DTO structs + mappers)
  transport/router.go     Route registration with per-route permission gates
  permission/policies.go  Permission catalog; all codes owned here
  module.go               Module struct — wires all layers, exposes contract ports
```

### Cross-module contract ports (backend/pkg/contract)

The approval module satisfies three contract interfaces used by the investment module:

| Interface | Method | Used by |
|---|---|---|
| `ApprovalSubmitter` | `SubmitForApproval` | Research report, decision, portfolio handlers |
| `ApprovalStatusProvider` | `GetApprovalStage` | Enriches read DTOs with approval stage badge |
| `ApprovalBatchActor` | `ApproveByRequest`, `RejectByRequest` | Batch approve/reject investment endpoints |

These are wired in `main.go` after both modules are constructed:

```go
investmentModule.SetApprovalSubmitter(approvalModule)
investmentModule.SetApprovalStatusProvider(approvalModule)
investmentModule.SetApprovalBatchActor(approvalModule)
approvalModule.RegisterSubjectCallback("RESEARCH_REPORT", investmentModule.ApprovalSubjectCallback())
approvalModule.RegisterSubjectCallback("INVESTMENT_DECISION", investmentModule.DecisionApprovalSubjectCallback())
approvalModule.RegisterSubjectCallback("PORTFOLIO", investmentModule.PortfolioApprovalCallback())
```

### Frontend layers

```
frontend/app/features/approval/
  types.ts                  Re-exports from generated ims-api.d.ts (NEVER hand-written)
  services/approvalApi.ts   Thin fetch wrappers for all approval endpoints
  composables/
    useApprovalInbox.ts     Paginated inbox list
    useApprovalRequest.ts   Single request detail (request + tasks + timeline + signatures)
    useApprovalActions.ts   Approve / reject / withdraw / revoke mutations
    useApprovalConfig.ts    Groups, teams, processes CRUD
    useApprovalSubjectStatus.ts  Status panel for an embedded subject (NEW)
  lib/
    approvalStatus.ts       requestStatusBadge(), requestStatusLabel(), isActiveStatus(), prettify()
    approvalMappers.ts      toTimelineEvents(), toStamps() — pure view-model mappers
  components/
    ApprovalStatusBadge.vue       Status chip with colour variant
    ApprovalTimeline.vue          Ordered event log
    ApprovalActionPanel.vue       Approve / reject form (canAct prop from parent)
    ApprovalRequestDetail.vue     Full detail page: request + tasks + timeline + stamps
    ApprovalInboxTable.vue        Paginated inbox list
    ApprovalGroupForm.vue         Group CRUD
    ApprovalGroupMemberTable.vue  Member management with reorder
    ApprovalProcessConfigForm.vue Process config + stage builder
    ApprovalStageBuilder.vue      Up to 3 stage definitions
    ApprovalTeamForm.vue          Team CRUD
    ApprovalTeamMemberTable.vue   Team member management
```

---

## 4. Request Lifecycle

```
SubmitApproval()
    ├── Validate: no active request for this subject (duplicate guard)
    ├── Resolve: ApprovalProcessConfig by (process_type, contract_id, effective_date)
    ├── Create: ApprovalRequest (PENDING_APPROVAL)
    ├── Create: ApprovalTasks for stage 1 (resolveStage → GROUP_PRIORITY / GROUP_ANY / TEAM_MINIMUM / SINGLE_USER)
    ├── Append: SUBMITTED event + TASK_CREATED event
    └── Notify: assigned approvers

ApproveTask(taskID, actorID, comment)
    ├── Lock request row (serialise concurrent actions)
    ├── authorizeAction(): maker-checker, data-scope, delegation check
    ├── Update task → APPROVED; write signature record
    ├── Append: DELEGATED event (if proxy), APPROVED event
    ├── Count approved tasks vs required threshold
    │   ├── Stage incomplete → return (other tasks still pending)
    │   └── Stage complete:
    │       ├── Skip remaining PENDING tasks in this stage
    │       ├── Append: STAGE_COMPLETED
    │       ├── Has next stage? → advance, create next stage tasks, notify
    │       └── Final stage? → APPROVED, append REQUEST_COMPLETED, SubjectSync.OnApproved()

RejectTask(taskID, actorID, reason)
    ├── authorizeAction(): same as above
    ├── Update task → REJECTED; cancel all other PENDING tasks on request
    ├── Request → REJECTED; append REJECTED + REQUEST_COMPLETED
    └── SubjectSync.OnRejected()

WithdrawRequest(requestID, actorID)    — submitter only, non-terminal requests
CancelRequest(requestID, actorID)      — privileged, non-terminal requests
RevokeRequest(requestID, actorID, reason) — APPROVED requests only; SubjectSync.OnRejected() to reopen subject
```

---

## 5. Approver Modes

| Mode | Task assignment | Stage completes when |
|---|---|---|
| `SINGLE_USER` | One configured user | That user approves |
| `GROUP_PRIORITY` | Highest-priority eligible group member (skipping submitter) | That member approves |
| `GROUP_ANY` | All eligible group members | Any `required_approval_count` approve |
| `TEAM_MINIMUM` | All eligible REVIEWER_AGENT team members (on the contract) | `min_required_stamps` approve |

Eligible = `is_active = true` AND (group: `status = APPROVED`) AND not the submitter.

---

## 6. Delegation

A delegate row in `approval__delegations` lets user **B** act as a proxy for user **A** for a time window.

```
approval__delegations
  from_user_id    UUID  — original assignee
  to_user_id      UUID  — proxy actor
  contract_id     UUID? — NULL = wildcard (all contracts)
  active_from     TIMESTAMPTZ
  active_until    TIMESTAMPTZ
  is_active       BOOLEAN
```

`PostgresDelegateResolver.ResolveDelegate()` is called from two places:

1. **`authorizeAction()`** — on every approve/reject, to accept the proxy actor.
2. **`GetApprovalRequest()`** — on detail reads, to set `ViewerTask` so the delegate sees action buttons.

Both calls use identical arguments `(assignee, contractID, now())`. Only `authorizeAction` additionally calls `ensureDataPermission`; the seed delegation uses `contract_id = NULL` so this gate is skipped in the demo.

**Demo delegation (seed 012):** ben → green, wildcard, 30 days. Green can approve any task assigned to ben, producing a `DELEGATED` signature that records both actors.

---

## 7. `allowed_actions` — Backend-Computed, Frontend-Consumed

`computeAllowedActions()` in `runtime_service.go:769` is the single source of truth for what a viewer may do:

```go
func computeAllowedActions(req, viewerTask, viewerID) []string {
    // "approve" + "reject" only when PENDING_APPROVAL AND viewer has a task
    // "withdraw" only when non-terminal AND viewer is the submitter
    // "cancel" when non-terminal
    // "revoke" when APPROVED
}
```

The HTTP response includes `allowed_actions []string` on every `GET /approvals/requests/{id}` and `GET /approvals/subjects/{subjectType}/{subjectId}/status` response. The frontend reads this list and shows/hides buttons accordingly. **No role or status logic is duplicated in the frontend.**

---

## 8. API Routes

All routes are under `/api/v1` and require a valid JWT (`AuthMiddleware`). Each route additionally requires a specific function permission enforced server-side.

### Runtime (`/approvals`)

| Method | Path | Permission | Handler |
|---|---|---|---|
| GET | `/inbox` | `APPROVAL_VIEW_INBOX` | Personal pending task list |
| GET | `/requests` | `APPROVAL_VIEW_REQUEST` | Paginated request list |
| GET | `/requests/{requestId}` | `APPROVAL_VIEW_REQUEST` | Full detail + allowed_actions |
| GET | `/requests/{requestId}/timeline` | `APPROVAL_AUDIT_VIEW` | Immutable event log |
| GET | `/subjects/{subjectType}/{subjectId}/status` | `APPROVAL_VIEW_REQUEST` | Latest request for a subject |
| POST | `/submit` | `APPROVAL_SUBMIT` | Create new approval request |
| POST | `/tasks/{taskId}/approve` | `APPROVAL_APPROVE` | Approve a task |
| POST | `/tasks/{taskId}/reject` | `APPROVAL_REJECT` | Reject a task (reason required) |
| POST | `/requests/{requestId}/withdraw` | `APPROVAL_WITHDRAW` | Submitter withdraws |
| POST | `/requests/{requestId}/cancel` | `APPROVAL_CANCEL` | Privileged cancel |
| POST | `/requests/{requestId}/revoke` | `APPROVAL_REVOKE` | Revoke an APPROVED request |

### Configuration (`/approval-config`)

| Method | Path | Permission |
|---|---|---|
| GET/POST | `/groups` | `APPROVAL_CONFIG_VIEW` / `APPROVAL_GROUP_MANAGE` |
| PUT | `/groups/{id}` | `APPROVAL_GROUP_MANAGE` |
| GET/POST | `/groups/{id}/members` | `APPROVAL_CONFIG_VIEW` / `APPROVAL_GROUP_MANAGE` |
| PUT/POST | `/groups/{id}/members/{memberId}` | `APPROVAL_GROUP_MANAGE` |
| POST | `/groups/{id}/members/{memberId}/approve` | `APPROVAL_GROUP_MANAGE` |
| POST | `/groups/{id}/members/{memberId}/revoke` | `APPROVAL_GROUP_MANAGE` |
| POST | `/groups/{id}/members/reorder` | `APPROVAL_GROUP_MANAGE` |
| GET/POST | `/teams` | `APPROVAL_CONFIG_VIEW` / `APPROVAL_TEAM_MANAGE` |
| GET/POST/PUT/DELETE | `/teams/{id}/members` | `APPROVAL_CONFIG_VIEW` / `APPROVAL_TEAM_MANAGE` |
| GET/POST | `/teams/{id}/contracts` | `APPROVAL_CONFIG_VIEW` / `APPROVAL_TEAM_MANAGE` |
| GET/POST | `/processes` | `APPROVAL_CONFIG_VIEW` / `APPROVAL_PROCESS_MANAGE` |
| GET/PUT | `/processes/{id}` | `APPROVAL_CONFIG_VIEW` / `APPROVAL_PROCESS_MANAGE` |
| POST | `/processes/{id}/activate` | `APPROVAL_PROCESS_MANAGE` |
| POST | `/processes/{id}/deactivate` | `APPROVAL_PROCESS_MANAGE` |

---

## 9. Database Schema

Migrations live in `database/migrations/` using prefix `approval__`. All tables use `approval__` double-underscore naming.

| Table | Purpose |
|---|---|
| `approval__groups` | Reusable approver pools |
| `approval__group_members` | Members; only `status=APPROVED AND is_active=true` are eligible |
| `approval__teams` | Per-contract/fund teams with stamp min/max |
| `approval__team_contracts` | Active team-to-contract assignments (one active team per contract) |
| `approval__team_members` | Team members (`ORDER_SUBMITTER` / `REVIEWER_AGENT`) |
| `approval__process_configs` | Process definitions with contract scope and effective date |
| `approval__process_stages` | Up to 3 ordered stages per config (max enforced by `CHECK`) |
| `approval__requests` | One approval instance per active business object |
| `approval__tasks` | Per-approver work items for each stage |
| `approval__events` | **Immutable** timeline; DB triggers reject UPDATE/DELETE |
| `approval__signature_records` | Digital stamp records with proxy/delegation marker |
| `approval__delegations` | Delegation grants (from → to, time-bounded, optional contract) |

Key constraints:
- One active request per `(subject_type, subject_id)` — partial unique index on non-terminal statuses.
- `approval__request_no_seq` generates `APR-000001` style human-readable numbers.
- `approval__events` immutability enforced at DB level (`BEFORE UPDATE/DELETE` trigger raises exception).

### Relevant migrations

| File | Change |
|---|---|
| `20260529000001_approval__create_tables` | Initial 11-table schema |
| `20260613000006_approval__add_portfolio_onboarding_process_type` | Adds `PORTFOLIO_ONBOARDING` process type |
| `20260616000001_approval__add_revoked_status` | Adds `REVOKED` to request status CHECK + `EventRevoked` event type |
| `20260616000002_approval__delegations` | New `approval__delegations` table |
| `20260616000003_approval__add_revoked_event_type` | Adds `REVOKED` to `approval__events` event_type CHECK |
| `20260617000001_approval__config_snapshot` | Adds `config_snapshot JSONB` to `approval__requests` (see §P0-5 below) |

---

## 10. Tests

### Backend (`backend/internal/approval/`)

`application/service/runtime_service_test.go` — 18 table-driven unit tests covering:
- `GROUP_PRIORITY` single-stage happy path
- `GROUP_ANY` multi-approval threshold
- Maker-checker rejection (submitter cannot approve own request)
- Multi-stage advancement (stage 1 → stage 2 → APPROVED)
- Delegation: delegate can approve; non-delegate cannot
- Rejection cancels all pending tasks
- Withdraw (submitter only) and Cancel (anyone with permission)

`application/command/decision_lifecycle_test.go` — 2 tests at the command-handler level:
- `TestSubmitDecision_UnapprovedReport_Rejected` — report with non-APPROVED review status blocks submit
- `TestSubmitDecision_NoReport_Skips_ReferenceCheck` — no report reference skips the check entirely

Test infrastructure: all fakes live in `application/service/fake_repo_test.go` (in-memory `fakeRepo`, `fakePermissions`, `fakeDelegateResolver`). The command-handler tests reuse `fakeDecisionRepo`, `fakeResearchReportRepo`, `recordingAudit` from the same `command` package.

### Frontend (`frontend/tests/approval-components.test.ts`)

37 pure TypeScript tests (no DOM, Vitest only) across 5 suites:

| Suite | Covers |
|---|---|
| `ApprovalStatusBadge` | `requestStatusBadge`, `requestStatusLabel`, `isActiveStatus` |
| `ApprovalTimeline` | `toTimelineEvents` mapper (actor fallbacks, delegation metadata, comments) |
| `ApprovalActionPanel` | `canAct` derivation, reject-reason validation |
| `ApprovalInboxTable` | `prettify`, `toStamps` mapper (proxy signature flag) |
| `ApprovalRequestDetail` | **Critical constraint test** — `allowed_actions` is the sole driver of button visibility |

---

## 11. Permission Codes

All defined in `backend/internal/approval/permission/policies.go`:

```
APPROVAL_VIEW_INBOX     — see personal task inbox
APPROVAL_VIEW_REQUEST   — read requests, detail, subject status
APPROVAL_SUBMIT         — submit a subject into the workflow
APPROVAL_APPROVE        — approve an assigned task
APPROVAL_REJECT         — reject an assigned task
APPROVAL_WITHDRAW       — withdraw own request
APPROVAL_CANCEL         — cancel any in-flight request (privileged)
APPROVAL_REVOKE         — revoke an APPROVED request
APPROVAL_AUDIT_VIEW     — read the immutable timeline
APPROVAL_CONFIG_VIEW    — read groups, teams, processes
APPROVAL_GROUP_MANAGE   — create/edit groups and members
APPROVAL_TEAM_MANAGE    — create/edit teams, contracts, members
APPROVAL_PROCESS_MANAGE — create/edit process configs and stages
```

Legacy codes `APPROVAL_VIEW` and `APPROVAL_CONFIG` are retained for backward compatibility with existing grants.

---

## 12. Security Fixes Applied 2026-06-17

The following P0 security issues were fixed. Each is documented in detail in
`docs/handoff/approval-permission-p0-fixes.md`.

| Fix | What changed |
|---|---|
| **P0-1 Subject access ports wired** | `main.go` registers investment `SubjectAccessor` for `RESEARCH_REPORT` and `INVESTMENT_DECISION` only. `PORTFOLIO` is deferred (Option A). Missing ports previously caused fail-open reads. |
| **P0-2 Fail-closed subject access** | `InvestmentSubjectAccessor` converts every nil-dependency or permission-error path to a deny response instead of allowing through. |
| **P0-3 UserDescriptor / SubjectDescriptor in DTOs** | `RequestResponse`, `TaskResponse`, `EventResponse`, `SignatureResponse` all carry nested `submitter`, `subject`, `actor`, `assigned_user`, `delegated_from`, `signer`, `proxy_for` descriptor objects. Raw UUID fallback fields are kept as routing IDs only and must never be rendered as labels. |
| **P0-4 Final-stage validation** | `config_service.go`: exactly one `is_final_stage=true` required, and it must be on the highest `stage_number`. Zero final stages → validation error (P0-13 updated this from auto-mark). Multiple finals → 422. Wrong position final → 422. |
| **P0-5 Config snapshot** | At submit time, stages are serialised to `config_snapshot JSONB` on `approval__requests`. `loadActionContext` and stage advancement use the snapshot (via `resolveStages()`), falling back to live config only for pre-snapshot requests. |
| **P0-6 Existence oracle** | `GetApprovalRequest`, `GetApprovalTimeline`, `GetSubjectApprovalStatus`: access-denied now returns `ErrNotFound` (or `nil`) instead of `ErrForbidden` so request existence is not revealed. |
| **P0-7 `allowed_actions` submitter/non-submitter guards** | `cancel` only shown to the submitter; `revoke` only to non-submitters. Previously both appeared for all viewers. |
| **P0-8 Permission terminal-state guard** | `permission_service.go:finish()` now rejects transitions into/out of `APPROVED`, `REJECTED`, `MERGED`, `CANCELLED`, `CLOSED`. `Reject()` now requires a non-blank comment. |
| **P0-9 Frontend UUID-free display** | All mapper fallbacks (`approval/index.vue`, `approvalMappers.ts`, `ApprovalRequestDetail.vue`, `PermissionRequestDetailScreen.vue`, `SettingsDataPermissionsPanel.vue`, config pages) use human-readable fields. UUID strings must never reach rendered text nodes. |
| **P0-10 Frontend tests** | `tests/approval-no-uuid.test.ts`, `tests/permission-no-uuid.test.ts` — regression guards that assertNoUUID on every display path. |
| **P0-11 PORTFOLIO deferred (Option A)** | `main.go` PORTFOLIO callbacks and access port registration removed. Seed config deactivated. Controlled demo: RESEARCH_REPORT and INVESTMENT_DECISION only. |
| **P0-12 Permission nil checker fail-closed** | `permission_service.go:requirePermission` returns Forbidden when checker is nil instead of allowing through. |
| **P0-13 Final-stage validation rejects missing final** | `validateStages` now returns a validation error when zero stages are marked `is_final_stage=true`, superseding the old auto-mark behaviour. |
| **P0-14 Config + stage builder UUID inputs removed** | `ApprovalProcessConfigForm.vue` and `ApprovalStageBuilder.vue` replaced free-text UUID inputs with disabled selectors ("not available"). |
| **P0-15 Permission Changes tab `target_id` removed** | `PermissionRequestDetailScreen.vue` no longer renders `item.target_id` as visible text. |
| **P0-16 `allowed_actions` revoke removed** | `computeAllowedActions` (called by `GetRequest`) no longer includes `revoke`. Note: `subjectAllowedActions` (called by `GetSubjectStatus`) still emits `"revoke"` for APPROVED requests, but `ApprovalRequestDetail.vue` derives `canRevoke` exclusively from `GetRequest`'s `allowed_actions` — not subject-status — so the revoke button is never shown. |
| **P0-17 `ListRequests` missing auth guard** | `runtime_handler.go:ListRequests` was ignoring the `ok` return of `actorID()`. A request without a valid JWT context caused `ViewerID = uuid.Nil`, which the service treats as a system bypass (returns all requests unfiltered). Fixed: returns 401 when actor context is absent. |
| **P0-18 `ApprovalGroupMemberTable` UUID display** | Member rows no longer render `m.user_id` in the table. Add-member form field changed from label "User UUID" to "User" with a disabled input ("User selector source not available"). |
| **P0-19 `ApprovalTeamMemberTable` UUID display** | Member rows no longer render `m.user_id`. Add-member input placeholder changed from "User UUID" to "User selector source not available" (disabled). |
| **P0-20 Delegation/proxy `principalUserName` UUID** | `useApprovalSession.ts` was passing raw `proxy_for_user_id` / `delegated_from_user_id` as `principalUserName`. All six assignment sites now use `derivePrincipalName(isDelegated, descriptor)` from `approvalMappers.ts`, which returns the descriptor's `display_name`, "Unknown User" when delegation exists but the name is unresolved, or `undefined` when no delegation is in effect. |
| **P0-21 Signature fallback persists raw UUID** | `writeSignature()` in `runtime_service.go` initialized `display := signer.String()` which stored the raw UUID as `SignerDisplayName` when the directory lookup failed. Fixed: `display := "Unknown User"` as the initial value. |

---

## 13. Invariants — Do Not Break

These rules are enforced in code and tested; a regression on any of them would break the approval guarantee:

1. **Never duplicate `allowed_actions` logic in the frontend.** `computeAllowedActions()` in `runtime_service.go` is the single source of truth.
2. **Never hardcode approver IDs in page components** or seeds. Approvers are resolved at runtime from process config → group → eligible members.
3. **Never delete approval events.** The DB trigger enforces this. If a scenario requires "undoing" an approval, use `RevokeRequest`.
4. **Never bypass the maker-checker check.** The submitter is stored on the request and checked in `authorizeAction()` against the acting user — not against a role.
5. **Never write API types by hand.** All frontend types in `types.ts` are aliases of `ims-api.d.ts` generated from the backend Swagger. Run `make api-client` after backend schema changes.
6. **Delegation check must appear in both `authorizeAction` AND `GetApprovalRequest`.** If the delegate check is only in one path, either action succeeds but buttons don't appear (or vice versa).

---

## 14. Known Gaps / Next Steps

| Gap | Priority | Notes |
|---|---|---|
| `LeaveChecker` is a no-op | Low | `NopLeaveChecker` always returns `false`. A real leave module can implement the interface and be wired in `NewModule` without changing the engine. |
| `TEAM_MINIMUM` stage resolution | Medium | `resolver.go` routes `TEAM_MINIMUM` to the team repo but the team resolver path has limited test coverage. |
| Delegation UI (inbox for delegates) | Medium | Backend correctly sets `ViewerTask` for delegates. No dedicated "acting as delegate" inbox filter exists yet in the frontend. |
| Revoke not surfaced in UI | P1 | `computeAllowedActions` (P0-16) omits `"revoke"` so `canRevoke` is always false in `ApprovalRequestDetail.vue`. The subject-status endpoint (`subjectAllowedActions`) still emits `"revoke"` for APPROVED requests but the detail view sources `allowed_actions` only from `GetRequest`, so the button is never rendered. The revoke route (`POST /requests/{id}/revoke`, `APPROVAL_REVOKE` permission) is wired and functional. To surface the button: inject a function-permission check result into `computeAllowedActions` at the `GetRequest` call site. |
| Process type `PORTFOLIO_ONBOARDING` | Low | Added to the enum and migration; seed does not yet include a demo process config for it. PORTFOLIO deferred (Option A). |
| User selector for group/team members | P1 | `ApprovalGroupMemberTable` and `ApprovalTeamMemberTable` add-member inputs are disabled placeholders. A real user-search autocomplete is required to re-enable member management in the UI. |
| Notification backend | Low | `NopNotifier` discards all notifications. Wire a real notifier when the notification module provides `ApprovalNotifier`. |
| E2E test for full approval flow | Medium | No Playwright tests cover the full submit → approve → callback path yet. |
