# P0 Delivery Report — Investment Decision Foundation Fixes

**Date:** 2026-06-18  
**Scope:** 9 P0 control gaps in the Investment Decision submit and execution paths  
**Status:** ALL P0s FIXED, TESTED — READY FOR REVIEW

---

## Summary

All 9 P0 control gaps are resolved. Build is clean. 17 new tests added.

| Test Suite | Count | Status |
|---|---|---|
| `go test ./internal/investment/...` | All packages | PASS |
| `go test ./internal/compliance/...` | All packages | PASS |
| `go test ./internal/approval/...` | All packages | PASS |
| `go build ./...` | — | CLEAN |

---

## Required Flow (after fixes)

```
Investment Decision Draft
  → validate workflow gate (IsTradeAllowed)
  → validate fund report-required gate (Fund.RequirePretradePreview)
  → validate optional report reference only if provided
  → run Compliance / IRG pre-trade check
  → if PASS or WARN:    submit to INVESTMENT_DECISION approval engine
  → if BLOCK + all-releasable:
       transition to PENDING_COMPLIANCE_RELEASE
       create COMPLIANCE_RELEASE approval request
       return (pending; awaiting compliance officer approval)
  → if BLOCK + non-releasable: return ErrComplianceRejected (terminal)

After compliance officer approves COMPLIANCE_RELEASE:
  → ApplyComplianceReleaseDecision(approved=true)
  → auto-submit to INVESTMENT_DECISION approval engine
  → transition to PENDING_APPROVAL

After portfolio manager approves INVESTMENT_DECISION:
  → transition to APPROVED / READY_FOR_EXECUTION

Execution:
  → gate on IsTransactionLocked (workflow state)
  → gate on decision status == APPROVED or READY_FOR_EXECUTION
```

---

## P0 Fix Summary

### P0-1 — `_ = compliance` discarded the ComplianceChecker port

**Root cause:** `investment/module.go` received a `contract.ComplianceChecker` in `NewModule()` and immediately discarded it with `_ = compliance`. No pre-trade IRG check ever ran on a live decision submission.

**Fix:**
```go
if compliance != nil {
    m.decisionCmd.SetComplianceChecker(compliance)
}
m.decisionCmd.SetFundRepository(m.funds)
```

**File:** `backend/internal/investment/module.go`

---

### P0-2 — No compliance pre-trade check in Submit()

**Root cause:** `DecisionCommandHandler.Submit()` had no `ComplianceChecker` field and never called the IRG gate.

**Fix:** Added `compliance contract.ComplianceChecker` and `funds domain.FundRepository` fields to `DecisionCommandHandler`, with `SetComplianceChecker()` and `SetFundRepository()` post-construction setters. Inserted compliance check in `Submit()` after the workflow gate, before report-reference validation.

- PASS / WARN: stores `ComplianceCheckGroupID` on decision, falls through to approval submission.
- BLOCK + all `Overridable=true`: transitions decision to `PENDING_COMPLIANCE_RELEASE`, creates a `COMPLIANCE_RELEASE` approval request, returns early.
- BLOCK + any `Overridable=false`: returns `ErrComplianceRejected` (terminal, HTTP 422).

**File:** `backend/internal/investment/application/command/decision_lifecycle.go`

---

### P0-3 — No config-driven pre-trade report gate

**Root cause:** `Fund.RequirePretradePreview` was respected only in the UI. The backend Submit() path never checked it — any decision from a report-required fund could be submitted without a research report.

**Fix:** Block A, inserted immediately after the workflow gate:
```go
if h.funds != nil && d.FundID != uuid.Nil {
    fund, _ := h.funds.GetByID(ctx, d.FundID)
    if fund != nil && fund.RequirePretradePreview && d.ResearchReportID == nil {
        return nil, &domain.ErrDecisionLifecycle{...}
    }
}
```

No new config field needed — reuses the existing `Fund.RequirePretradePreview bool` entity field.

**File:** `backend/internal/investment/application/command/decision_lifecycle.go`

---

### P0-4 — BLOCK-but-releasable had no release path (new COMPLIANCE_RELEASE flow)

**Root cause:** There was no `PENDING_COMPLIANCE_RELEASE` lifecycle status, no `COMPLIANCE_RELEASE` approval process type, and no callback to resubmit the decision after a compliance officer approved the release.

**Fixes across multiple files:**

**New lifecycle status:**
```go
DecisionLifecyclePendingComplianceRelease DecisionLifecycleStatus = "PENDING_COMPLIANCE_RELEASE"
```
`CanCancel()` returns `true` for this status. Not included in `IsTerminal()`.

**New approval enum values:**
```go
ProcessComplianceRelease ProcessType = "COMPLIANCE_RELEASE"
SubjectComplianceRelease SubjectType = "COMPLIANCE_RELEASE"
```

**New `ApplyComplianceReleaseDecision()` method on `DecisionCommandHandler`:**
- Approved → re-submits to `INVESTMENT_DECISION` approval engine, transitions to `PENDING_APPROVAL`.
- Rejected → transitions to `CANCELLED` with cancellation reason.

**New callback adapter** `ComplianceReleaseCallback` in `infrastructure/adapter/compliance_release_adapter.go`:
Implements `contract.ApprovalSubjectCallback`, routes `COMPLIANCE_RELEASE` approval outcomes to `ApplyComplianceReleaseDecision`.

**Extended `InvestmentSubjectAccessor`** (`infrastructure/adapter/investment_subject_access.go`):
`resolveContractID()` handles `"COMPLIANCE_RELEASE"` subject type by looking up the decision (same physical object as `"INVESTMENT_DECISION"`).

**`main.go` registrations:**
```go
approvalModule.RegisterSubjectCallback("COMPLIANCE_RELEASE", investmentModule.ComplianceReleaseSubjectCallback())
approvalModule.RegisterSubjectValidator("COMPLIANCE_RELEASE", investmentModule.DecisionSubjectValidator())
approvalModule.RegisterSubjectAccessPort("COMPLIANCE_RELEASE", investSubjectAccessor)
```

**DB migrations:**
- `20260618000001_approval__add_compliance_release_process_type.up/down.sql` — extends `chk_approval_process_type` CHECK constraint.
- `20260618000002_investment__add_compliance_release_status.up/down.sql` — extends `chk_inv_decision_status` CHECK constraint.

**Files:**
- `backend/internal/investment/domain/valueobject/decision_lifecycle.go`
- `backend/internal/investment/domain/entity/decision.go`
- `backend/internal/approval/domain/valueobject/enums.go`
- `backend/internal/investment/application/command/decision_lifecycle.go`
- `backend/internal/investment/infrastructure/adapter/compliance_release_adapter.go` *(new)*
- `backend/internal/investment/infrastructure/adapter/investment_subject_access.go`
- `backend/internal/investment/module.go`
- `backend/cmd/server/main.go`
- `database/migrations/20260618000001_approval__add_compliance_release_process_type.up.sql` *(new)*
- `database/migrations/20260618000001_approval__add_compliance_release_process_type.down.sql` *(new)*
- `database/migrations/20260618000002_investment__add_compliance_release_status.up.sql` *(new)*
- `database/migrations/20260618000002_investment__add_compliance_release_status.down.sql` *(new)*

---

### P0-5 — Execution lacked workflow transaction-lock gate

**Root cause:** `ExecutionCommandHandler.Create()` checked decision status but never called `IsTransactionLocked()`. Executions could be opened against a locked trading day (e.g. during EOD processing or while an approval was in flight).

**Fix:** Added `workflow contract.WorkflowStateProvider` field to `ExecutionCommandHandler` with a `SetWorkflowStateProvider()` setter. In `Create()`, immediately after the decision status check:
```go
if h.workflow != nil {
    locked, err := h.workflow.IsTransactionLocked(ctx, d.ContractID, d.BusinessDate)
    if locked {
        return nil, &domain.ErrDecisionLifecycle{...}
    }
}
```
`IsTransactionLocked` is used here (not `IsTradeAllowed`) — execution is an operational act, not a new trade submission.

**Files:**
- `backend/internal/investment/application/command/execution.go`
- `backend/internal/investment/module.go`

---

### P0-6 — Compliance override accepted `approved_by` from request body

**Root cause:** `OverrideRequest` in the compliance HTTP handler had an `ApprovedBy string` field that was parsed from the request body and used to populate the command. A caller could supply any UUID as the approver — authorization by request payload.

**Fix:** Removed `ApprovedBy string` from `OverrideRequest`. The `OverriddenBy` field (already derived from JWT claims) is the only identity source. `OverrideBreachRequest.ApprovedBy *uuid.UUID` in the command layer remains available for future approval-engine integration but is nil for now.

**Files:**
- `backend/internal/compliance/transport/handler/compliance_handler.go`
- `frontend/app/features/compliance/types.ts` — removed `approved_by?: string` from `ComplianceOverrideRequest`
- `frontend/app/features/compliance/components/ComplianceBreachOverrideDialog.vue` — removed emit type field, reactive form field, `approverError` computed, `canSubmit` dependency, submit payload, and template form block

---

### P0-7 — IRG stub silently passed unimplemented regulatory rules

**Root cause:** `ThaiSECRule.Evaluate()` returned `VerdictPass` with a PoC comment. After wiring the compliance check (P0-1 + P0-2), this meant every decision would silently pass the Thai SEC regulatory rule regardless of actual checks.

**Fix:**
```go
return spi.EvalResult{
    Verdict: vo.VerdictWarn,
    Message: "regulatory.thai_sec: STUB — NOT_CONFIGURED; emitting WARN...",
    Evidence: vo.Evidence{
        Metrics: map[string]string{"status": "NOT_CONFIGURED", "stub": "true"},
    },
}, nil
```
`VerdictWarn` is used, not a new `VerdictNotConfigured` constant, because the compliance aggregator's verdict ranking only knows PASS / WARN / BLOCK.

**File:** `backend/internal/compliance/rules/stub/irg_regulatory.go`

---

### P0-8 — `ErrComplianceRejected` not handled in HTTP error writer

**Root cause:** `writeDecisionError()` in the decision HTTP handler had no case for `*domain.ErrComplianceRejected`. The error fell through to a generic 500.

**Fix:** Added case returning HTTP 422 Unprocessable Entity with structured error body.

**File:** `backend/internal/investment/transport/handler/decision_handler.go`

---

### P0-9 — Frontend override dialog sent `approved_by` in request body

**Root cause:** `ComplianceBreachOverrideDialog.vue` emitted a payload with `{ reason, approved_by }` and built a request body from that payload. Any user could set an arbitrary UUID as the approving officer.

**Fix:** Removed all `approved_by` traces:
- Emit type: `submit: [payload: { reason: string }]`
- Reactive form, watch reset, `approverError` computed, `canSubmit` reference, submit payload builder, and template `<label>` block all removed.
- `isUuid` import removed (no longer used after `approverError` deletion).

**File:** `frontend/app/features/compliance/components/ComplianceBreachOverrideDialog.vue`

---

## Files Changed

### Backend

| File | Change |
|---|---|
| `backend/internal/approval/domain/valueobject/enums.go` | Added `ProcessComplianceRelease`, `SubjectComplianceRelease`; updated validators |
| `backend/internal/investment/domain/valueobject/decision_lifecycle.go` | Added `PENDING_COMPLIANCE_RELEASE`; updated `IsValid()` |
| `backend/internal/investment/domain/entity/decision.go` | `CanCancel()` returns true for `PENDING_COMPLIANCE_RELEASE` |
| `backend/internal/investment/application/command/decision_lifecycle.go` | `compliance`/`funds` fields + setters; Block A (report gate) + Block B (compliance check) in `Submit()`; new `ApplyComplianceReleaseDecision()` |
| `backend/internal/investment/application/command/execution.go` | `workflow` field + setter; `IsTransactionLocked` gate in `Create()` |
| `backend/internal/investment/infrastructure/adapter/compliance_release_adapter.go` | **NEW** — `ComplianceReleaseCallback` adapter |
| `backend/internal/investment/infrastructure/adapter/investment_subject_access.go` | Extended `resolveContractID()` to handle `COMPLIANCE_RELEASE` |
| `backend/internal/investment/module.go` | Wire `SetComplianceChecker()`, `SetFundRepository()`, `SetWorkflowStateProvider()`; added `ComplianceReleaseSubjectCallback()` |
| `backend/cmd/server/main.go` | Register COMPLIANCE_RELEASE callback, validator, access port |
| `backend/internal/compliance/transport/handler/compliance_handler.go` | Removed `approved_by` from `OverrideRequest` |
| `backend/internal/compliance/rules/stub/irg_regulatory.go` | `VerdictPass` → `VerdictWarn` with `NOT_CONFIGURED` evidence |
| `backend/internal/investment/transport/handler/decision_handler.go` | HTTP 422 case for `ErrComplianceRejected` |

### Database Migrations

| File | Change |
|---|---|
| `database/migrations/20260618000001_approval__add_compliance_release_process_type.up.sql` | **NEW** — extends `chk_approval_process_type` to include `COMPLIANCE_RELEASE` |
| `database/migrations/20260618000001_approval__add_compliance_release_process_type.down.sql` | **NEW** — reverts constraint |
| `database/migrations/20260618000002_investment__add_compliance_release_status.up.sql` | **NEW** — extends `chk_inv_decision_status` to include `PENDING_COMPLIANCE_RELEASE` |
| `database/migrations/20260618000002_investment__add_compliance_release_status.down.sql` | **NEW** — reverts constraint |

### Frontend

| File | Change |
|---|---|
| `frontend/app/features/compliance/types.ts` | Removed `approved_by?: string` from `ComplianceOverrideRequest` |
| `frontend/app/features/compliance/components/ComplianceBreachOverrideDialog.vue` | Removed all `approved_by` references (emit, form, computed, template) |

### Tests

| File | Tests Added |
|---|---|
| `backend/internal/investment/application/command/decision_compliance_test.go` *(new)* | `TestSubmit_FundRequiresReport_NoReport_Rejected` |
| | `TestSubmit_FundRequiresReport_ReportLinked_Proceeds` |
| | `TestSubmit_FundNotRequireReport_NoReport_Proceeds` |
| | `TestSubmit_ComplianceNil_ContinuesToApproval` |
| | `TestSubmit_CompliancePass_ContinuesToApproval` |
| | `TestSubmit_ComplianceWarn_ContinuesToApproval` |
| | `TestSubmit_ComplianceBlock_NonReleasable_ReturnsError` |
| | `TestSubmit_ComplianceBlock_AllReleasable_CreatesComplianceReleaseApproval` |
| | `TestApplyComplianceReleaseDecision_Rejected_CancelsDecision` |
| | `TestApplyComplianceReleaseDecision_Approved_ResubmitsToInvestmentDecisionApproval` |
| | `TestApplyComplianceReleaseDecision_Approved_NoApprovalEngine_ReturnsError` |
| `backend/internal/investment/application/command/execution_workflow_test.go` *(new)* | `TestCreateExecution_WorkflowNil_Proceeds` |
| | `TestCreateExecution_WorkflowNotLocked_Proceeds` |
| | `TestCreateExecution_WorkflowLocked_Blocked` |
| | `TestCreateExecution_DecisionNotApproved_Blocked` |
| `backend/internal/compliance/rules/stub/irg_regulatory_test.go` *(new)* | `TestThaiSECRule_ReturnsWarn` |
| | `TestThaiSECRule_EvidenceStatus_NotConfigured` |

Total new tests: **17**

---

## Decision Lifecycle States

```
DRAFT
  │
  ▼ Submit()
  ├─ [fund RequirePretradePreview=true, no report] → ErrDecisionLifecycle (rejected)
  ├─ [compliance BLOCK, non-overridable] → ErrComplianceRejected (rejected, HTTP 422)
  │
PENDING_COMPLIANCE_RELEASE   ← [compliance BLOCK, all-overridable]
  │
  ├─ [compliance officer rejects] → CANCELLED
  │
  ▼ [compliance officer approves]
PENDING_APPROVAL             ← [compliance PASS/WARN or released]
  │
  ├─ [manager rejects] → REJECTED
  │
  ▼ [manager approves]
APPROVED
  │
  ▼ [execution created]
READY_FOR_EXECUTION
  │
  ▼ [trade confirmed]
EXECUTED

CANCELLED  ← from DRAFT, PENDING_APPROVAL, PENDING_COMPLIANCE_RELEASE
```

---

## Architecture Constraints (do not violate)

- **Never trust identity from request body.** `approvedBy`, `actorId`, `username`, `workflowStatus`, `approvalStatus`, `complianceStatus`, `permissionHints` are all attacker-controlled when sourced from the request.
- **Never authorize by `username` or `displayName`.** Only UUIDs from validated JWT claims are authoritative identity.
- **Never duplicate approval logic** inside Investment or Compliance. Route through `contract.ApprovalSubmitter`.
- **Never import another module's `internal/`** from a sibling module. Use `pkg/contract` ports only.
- **Every `.up.sql` needs a matching `.down.sql`.**
- **Never hand-edit `frontend/app/api/ims-api.d.ts`.** Run `make swagger && make api-client` after any backend response shape change.
- **`Fund.RequirePretradePreview`** is the server-side gate for the research report requirement. The frontend gate is UX-only and is not a security control.
- **IRG stubs must not return `VerdictPass`** in any code path wired into an authoritative investment submit flow.

---

## Verification Commands

```bash
# Build check
cd backend && go build ./...

# Affected package tests
cd backend && go test ./internal/investment/... ./internal/compliance/... ./internal/approval/...

# Full backend suite
cd backend && go test ./...

# Migrations (dev environment)
make migrate-up

# Frontend type check (after make swagger && make api-client if backend response shapes changed)
cd frontend && npm run build
```

**Verified 2026-06-18:**
- `go build ./...` — CLEAN
- `go test ./internal/investment/... ./internal/compliance/... ./internal/approval/...` — PASS (17 new tests)
- `cd frontend && npm run build` — CLEAN (zero TypeScript errors)
- `go test ./...` (full suite) and `make migrate-up` — pending dev environment; run before merge
- `make swagger && make api-client` — pending; `ims-api.d.ts` has a stale `approved_by` field in `OverrideRequest` (compliance override API uses hand-written `types.ts` types so this does not affect the build, but regenerate before the next API release)

---

## Deferred Items (Phase 2)

| Item | Notes |
|---|---|
| `ComplianceStatus` field on `entity.Decision` | A typed `PENDING/PASS/WARN/BREACH/RELEASED` enum on the decision would make the compliance state visible in read-path responses without requiring a join to the check group. `ComplianceCheckGroupID` is sufficient for P0 traceability. Requires DB migration. |
| Per-line compliance checks for basket/rebalance/switch orders | P0 runs the compliance check against the decision header with an empty Ticker for these order types. Full per-instrument line checking is Phase 2. |
| Compliance status enrichment in read path | `WorkflowTradeAllowed` and `ComplianceVerdict` summary fields in list/detail responses require additional port injection into query handlers. |
| Real Thai SEC regulatory rule | `ThaiSECRule` is a stub emitting `VerdictWarn`. The actual Thai SEC / BOT regulatory checks are Phase 2 work. |
| User selector for compliance officer approval | The `COMPLIANCE_RELEASE` approval flow uses the same approval group/team mechanism as `INVESTMENT_DECISION`. A dedicated compliance officer group must be configured in the approval process config for the `COMPLIANCE_RELEASE` process type. |
| Compliance override approval integration | `OverrideBreachRequest.ApprovedBy *uuid.UUID` is nil for now. When a second-level approval requirement is introduced for overrides, this should be populated from the approval engine callback, not from a request body field. |
