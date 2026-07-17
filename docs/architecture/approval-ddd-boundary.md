# Approval DDD Boundary — Architecture Reference

**As of:** 2026-06-17  
**Status:** Controlled demo ready (RESEARCH_REPORT, INVESTMENT_DECISION)

---

## Bounded Context Responsibilities

### Approval bounded context

Owns:
- The approval workflow state machine (submit → pending → approved/rejected)
- Stage routing (which approver handles which stage)
- Task assignment, delegation, and escalation
- Immutable event log (`approval__events`)
- Signature/stamp records (`approval__signature_records`)
- `allowed_actions` computation (server-side, frontend renders)

Does NOT own:
- The identity or meaning of the subjects it governs
- Contract IDs, fund IDs, portfolio IDs, or any business-specific keys
- IAM data-permission checks

The approval engine operates entirely on `(subject_type, subject_id)` pairs. It never inspects the business meaning of those values.

### Permission bounded context

Owns:
- Access policy: who can read, write, approve, configure which resources
- Function permissions (code-based gates checked at route middleware)
- Data permissions (contract/scope-based gates checked via IAM port)
- Change request lifecycle (draft → ready → approved/merged)
- Merge: applying approved permission changes to production tables

Does NOT own:
- Approval workflow state (delegates to Approval module when `ApprovalEnabled` is set)
- Subject meaning

### Investment module

Owns:
- Subject meaning: which research reports, decisions, portfolios exist
- Subject access: resolving `(subject_type, subject_id)` to a contract UUID, then delegating to IAM
- Subject callbacks: updating business objects when an approval completes

---

## Subject Access Port Pattern

The approval engine cannot import investment internals. Authorization is injected via the `SubjectAccessPort` interface:

```go
// backend/pkg/contract/approval.go
type SubjectAccessor interface {
    CanViewApprovalSubject(ctx, actorID, subjectType, subjectID) error
    CanSubmitApprovalSubject(ctx, actorID, subjectType, subjectID) error
    CanActOnApprovalSubject(ctx, actorID, subjectType, subjectID, action) error
}
```

At startup (`main.go`), each supported subject type is registered:

```go
approvalModule.RegisterSubjectAccessPort("RESEARCH_REPORT", investSubjectAccessor)
approvalModule.RegisterSubjectAccessPort("INVESTMENT_DECISION", investSubjectAccessor)
// PORTFOLIO: deferred — not registered until the accessor supports it
```

If no port is registered for a subject type, all access fails closed with `ErrForbidden`.

---

## Subject Descriptor Pattern

The approval engine returns human-readable metadata about subjects via a `SubjectDescriptor`:

```json
{
  "subject_type": "RESEARCH_REPORT",
  "subject_id": "<uuid>",
  "subject_number": "RPT-2026-042",
  "subject_title": "Q2 Equity Analysis",
  "display_label": "RPT-2026-042 — Q2 Equity Analysis"
}
```

Subject descriptors are populated by subject callbacks. The frontend renders `display_label` instead of the raw UUID.

---

## User Descriptor Pattern

Any reference to a user in an approval response is returned as a `UserDescriptor`:

```json
{
  "user_id": "<uuid>",
  "username": "alice.nakamura",
  "display_name": "Alice Nakamura",
  "account_code": "U001"
}
```

The frontend renders `display_name` or `username`. Raw UUIDs are never shown in normal UI.

---

## allowed_actions Ownership

`allowed_actions` is always computed server-side in `computeAllowedActions` (runtime_service.go). Frontend renders buttons based on this list and never duplicates the logic.

Current actions surfaced via `allowed_actions`:
- `approve` / `reject` — viewer has a pending task at the current stage
- `withdraw` — submitter only, non-terminal
- `cancel` — submitter only, non-terminal

Actions NOT in `allowed_actions` (route-level gate only):
- `revoke` — requires function-permission check not available in `computeAllowedActions` (P1 improvement)

---

## Why Approval Must Not Depend on contract_id/fund_id/portfolio_id

The approval engine is a generic bounded context. If it depended on `contract_id` or `portfolio_id`:
- Every new business module would require approval module changes
- The approval module would import business module internals (forbidden in modular monolith)
- Compliance audits would conflate workflow logic with business logic

The `(subject_type, subject_id)` pattern keeps the boundary clean. The business module owns the mapping from its identity to the contract used for data-permission checks.

---

## No Raw UUID in Frontend (Policy)

Allowed visible values in Approval/Permission normal UI:
- username, accountCode, displayName
- group code/name, team code/name
- subject number, subject title
- business code/name, displayLabel
- "Unknown User", "Deleted User", "Unknown Subject"

Forbidden visible values:
- Any raw UUID (user, actor, signer, target, contract, fund, portfolio)

See `docs/handoff/frontend-uuid-free-policy.md` for enforcement rules.

---

## PORTFOLIO Approval — Deferred (Option A)

PORTFOLIO approval is intentionally excluded from the controlled demo.

Reason: The `InvestmentSubjectAccessor` has no `PORTFOLIO` case in `resolveContractID`. Registering the access port while the switch hits the default `return error` arm means all PORTFOLIO approval access would be denied — a registered but broken subject type is worse than a deferred one.

To implement PORTFOLIO approval later:
1. Add `portfolios domain.PortfolioRepository` field to `InvestmentSubjectAccessor`
2. Wire the field in `NewInvestmentSubjectAccessor`
3. Add `case "PORTFOLIO":` to `resolveContractID` resolving the portfolio's associated contract
4. Add `PortfolioSubjectValidator` implementing `ApprovalSubjectValidator`
5. In `main.go`: register callback, validator, and access port for PORTFOLIO
6. Reactivate `PROC_PORTFOLIO_ONBOARDING_DEFAULT` in the seed

See `docs/handoff/approval-subject-access-port.md` for the full registration pattern.
