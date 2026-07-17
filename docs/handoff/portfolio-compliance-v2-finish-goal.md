# Portfolio Compliance V2 Finish Goal

This document is the source of truth for the Claude Code goal prompt. The prompt itself should stay short and point here, because the Claude goal field has a 4000 character limit.

## Mission

Finish the blocking Portfolio Compliance V2 work so the branch can be reviewed and committed safely.

The application is portfolio-management first. Portfolio is the primary user-facing identity. Users work by portfolio code, not UUID. Fund is secondary context. The old `ContractID` naming still exists inside cross-module compliance interfaces, but for this goal it means optional fund scope. Do not rename it in this change.

## Current State

The Portfolio Compliance V2 foundation is mostly implemented as local working-tree changes. It adds portfolio-code compliance routes, portfolio rule bindings, new compliance rule families, frontend compliance workspace files, and migrations that allow compliance audit records without contract identity.

The branch is not ready to commit because the remaining blockers are functional and build-related:

1. Execution does not run compliance between manager approval and execution.
2. The new compliance frontend likely misses i18n keys.
3. V2 pre-trade handler maps validation errors to HTTP 500 instead of 400.
4. The nullable `contract_id` down migration needs honest rollback semantics.
5. The frontend service has a small custom API error helper instead of the shared OpenAPI helper.
6. The compliance view has an open-ended effective date display bug.

## Non-Goals

Do not rename `ContractID` across compliance today.

Do not remove the existing submit-time compliance check. The requested lifecycle needs a second gate, not a move.

Do not redesign the UI in this goal.

Do not work on P2/P3 cleanup such as N+1 rule version loading, duplicated exposure math, duplicated nil UUID conversion, or row-state refactors.

Do not commit or push.

Do not stage or modify unrelated Bruno health files or unrelated docs.

## Required Fixes

### 1. Add Execution-Time Compliance Gate

Primary file:

- `backend/internal/investment/application/command/execution.go`

Wiring file:

- `backend/internal/investment/module.go`

Relevant existing pattern:

- `backend/internal/investment/application/command/submit_decision_for_execution.go`

Required behavior:

- Add a `contract.ComplianceChecker` dependency to `ExecutionCommandHandler`.
- Add a setter such as `SetComplianceChecker(checker contract.ComplianceChecker)`.
- Wire it in `backend/internal/investment/module.go` near the existing decision compliance wiring.
- In `ExecutionCommandHandler.Create()`, run compliance after:
  - decision exists
  - decision status is `APPROVED` or `READY_FOR_EXECUTION`
  - workflow transaction lock passes
- Run compliance before:
  - creating the execution row
  - changing decision status to `READY_FOR_EXECUTION`
- Use the decision as the source of truth:
  - `PortfolioID`: `d.PortfolioID`
  - legacy `ContractID`: `d.FundID`
  - `BusinessDate`: `d.BusinessDate`
  - `Actor`: `req.ActorID.String()`
  - `OrderID`: `d.ID`
  - `Ticker`: `d.InstrumentCode`
  - `Side`: map from investment side to compliance side using the existing helper if available
  - `Quantity`: requested ordered quantity if provided, otherwise decision quantity
  - `Price`: derive from decision limit price or ordered amount/quantity if a safe existing pattern exists; otherwise mirror submit-time behavior
  - `Currency`: `d.Currency`
  - `Exchange`: `d.Exchange`
  - `Fees`: zero unless the execution create request has a fee field in current code
- If compliance returns `BLOCK`, do not create execution. Return the same typed/domain error family used by existing compliance rejection where possible, ideally `domain.ErrComplianceRejected`.
- If compliance returns `PASS` or `WARN`, continue. WARN should not block unless existing product logic says otherwise.
- If compliance checker is nil, preserve existing test compatibility unless the surrounding module pattern makes nil impossible. Do not introduce panics.

Tests to add or update:

- Execution is blocked when compliance returns `BLOCK`.
- Execution proceeds when compliance returns `PASS`.
- Execution proceeds on `WARN` if current behavior treats WARN as non-blocking.
- Compliance request contains the decision portfolio ID and fund ID as legacy `ContractID`.
- No execution row is created on BLOCK.

### 2. Add Missing Frontend i18n Keys

Files:

- `frontend/app/shared/i18n/messages/en/portfolio.ts`
- `frontend/app/shared/i18n/messages/th/portfolio.ts`
- `frontend/app/shared/i18n/messages/zh/portfolio.ts`

Required behavior:

- Inspect `frontend/app/features/portfolio-workspace/PortfolioComplianceView.vue`.
- Add every `portfolio.compliance.*` key used by the view to all three locale files.
- Keep wording short and operational.
- Do not expose UUIDs, `contract_id`, or `fund_id` in normal UI labels.
- Run build/typecheck to prove the key tree compiles.

### 3. Return HTTP 400 for V2 Validation Failures

File:

- `backend/internal/investment/transport/handler/portfolio_v2_compliance_handler.go`

Required behavior:

- Bad user input from `RunPreTradeCheckByCode` should return HTTP 400.
- Real infrastructure/internal failures should remain HTTP 500.
- Reuse an existing validation error type or sentinel if the backend already has one.
- If the compliance command currently wraps validation as generic `fmt.Errorf("invalid pre-trade request: %w", err)`, add the smallest safe classifier that distinguishes validation from infrastructure.
- Add or update handler tests for invalid request causing 400.

Examples of user/input errors:

- missing ticker
- invalid side
- zero/negative quantity
- zero/negative price
- invalid date
- invalid decimal

### 4. Clarify Nullable contract_id Down Migration

File:

- `database/migrations/20260707000001_compliance__nullable_contract_id.down.sql`

Required behavior:

- Do not imply this down migration is safely reversible after portfolio-only checks have written NULL `contract_id` rows.
- Pick one of these approaches:
  - Add a deliberate backfill/sentinel strategy before `SET NOT NULL`, if that matches repo conventions.
  - Or clearly document that the down migration is not reversible after portfolio-only compliance rows exist.
- Prefer the smallest safe option for this branch.

### 5. Use Shared OpenAPI Error Handling

File:

- `frontend/app/features/portfolio-workspace/services/portfolioComplianceApi.ts`

Required behavior:

- Remove the local `assertOpenApiResponseVoid` helper if the shared helper can handle DELETE/204 responses.
- Use the same shared OpenAPI response/error mechanism used by the repo, for example `unwrapOpenApiResponse` or another helper from `~/api/openapi`.
- Preserve typed API usage and do not hand-write DTOs.

### 6. Fix Open-Ended Effective Date Display

File:

- `frontend/app/features/portfolio-workspace/PortfolioComplianceView.vue`

Required behavior:

- Open-ended bindings should show the i18n `indefinite` label.
- Do not rely on `formatDate(undefined)` returning a falsy value if it actually returns a dash placeholder.
- Check `effective_to` directly before formatting.

## Validation Commands

Run the highest-signal validation available in this repo:

```powershell
go test ./...
```

For frontend, inspect `package.json` first, then run the relevant scripts. At minimum run:

```powershell
npm run build
```

Also run:

```powershell
git diff --check
```

Run `gofmt` on edited Go files.

## Acceptance Criteria

The goal is done only when:

- Execution creation is blocked by a `BLOCK` compliance verdict.
- Execution creation still works for allowed decisions when compliance passes.
- V2 pre-trade validation errors return 400.
- Frontend build/typecheck no longer fails because of missing portfolio compliance i18n keys.
- Open-ended rule bindings display an `indefinite` label.
- The frontend compliance API service uses the shared OpenAPI error path.
- The nullable `contract_id` down migration honestly documents or handles irreversibility.
- Backend and frontend validation commands pass, or any failure is clearly unrelated and documented.

## Final Response Required From Claude

Claude should report:

- Files changed.
- Tests/build commands run and exact pass/fail result.
- Whether the branch is ready to commit.
- Anything intentionally deferred.

Claude must not commit or push.
