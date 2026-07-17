# Portfolio V2 Claude Code Implementation Prompt

Use this file when context is limited. Give Claude Code the prompt below and tell it to continue from the current repository state.

## Copy/Paste Prompt

You are Claude Code working as a senior staff engineer on this repo:

```text
C:\Users\kanta\source\repos\ims-th-solution
```

Your mission:
Complete the Portfolio V2 migration safely, in phases, until the investment module is portfolio-first.

Current business decision:
This application is portfolio-management first. Fund/product is secondary/future. Portfolio is the operational center.

Current known state:
- A local commit already exists:
  `4f66918 feat(investment): add portfolio-code v2 lookup route`
- Continue from the current repo state. Do not redo that work unless tests prove it is broken.
- Do not push to remote.

Non-negotiable architecture:
- Public API/UI identity: portfolio code, not UUID.
- Internal DB/domain identity: `portfolio_id` UUID.
- `portfolio_id` is the operational/accounting source of truth.
- `fund_id` and `contract_id` are legacy compatibility fields during migration.
- New V2 request bodies must not expose or trust `fund_id` or `contract_id`.
- Frontend users must not type or see UUIDs for normal portfolio navigation.
- V1 behavior must stay intact until V2 is fully cut over.

Working rules:
- Work autonomously milestone by milestone.
- Keep each milestone PR-sized and reviewable.
- Commit locally after each completed milestone if tests pass.
- Never push.
- Do not use `git add .`; stage only files related to the current milestone.
- Do not delete unrelated untracked docs or user files.
- Before destructive DB changes, prove no code path depends on the field anymore.
- If a migration can fail because existing data is dirty, add preflight SQL and report it clearly.
- Run tests before every commit.

## Read First

Before coding, run:

```powershell
git status --short
git log -1 --oneline
```

Then read these files in order:

```text
docs/api/portfolio-v2-api-ddd.md
docs/frontend/portfolio-v2-frontend-ddd.md

backend/internal/investment/module.go
backend/internal/investment/transport/handler/investment_handler.go
backend/internal/investment/domain/portfolio_repository.go
backend/internal/investment/domain/portfolio_errors.go
backend/internal/investment/infrastructure/persistence/portfolio_repository.go
backend/internal/investment/transport/handler/portfolio_v2_handler_test.go

backend/internal/investment/domain/entity/portfolio.go
backend/internal/investment/domain/entity/portfolio_transaction.go
backend/internal/investment/transport/dto/request/requests.go
backend/internal/investment/transport/dto/response/responses.go
backend/internal/investment/transport/dto/response/mappers.go

backend/internal/investment/infrastructure/persistence/transaction_repository.go
backend/internal/investment/infrastructure/persistence/position_repository.go
backend/internal/investment/infrastructure/persistence/cash_repository.go
backend/internal/investment/infrastructure/persistence/valuation_repository.go

backend/internal/investment/domain/entity/decision.go
backend/internal/investment/domain/entity/execution.go
backend/internal/investment/domain/entity/trade_confirmation.go
backend/internal/investment/infrastructure/persistence/decision_repository.go
backend/internal/investment/infrastructure/persistence/execution_repository.go
backend/internal/investment/infrastructure/persistence/trade_confirmation_repository.go

database/migrations/20260428000002_investment__create_funds_portfolios.up.sql
database/migrations/20260428000004_investment__create_ledger.up.sql
database/migrations/20260428000005_investment__create_pricing_valuation.up.sql
database/migrations/20260601000001_investment__create_decisions_executions_confirmations.up.sql
database/table/49_investment__portfolios.sql
database/table/48_investment__portfolio_transactions.sql
database/table/101_investment__decisions.sql
database/table/102_investment__executions.sql
database/table/108_investment__trade_confirmations.sql

backend/internal/investment/infrastructure/adapter/investment_subject_access.go
backend/internal/investment/infrastructure/adapter/portfolio_scope_adapter.go
backend/internal/compliance/ports/portfolio_meta.go
backend/internal/compliance/infrastructure/adapter/investment_data_adapters.go
backend/internal/workflow/transport/router.go
docs/architecture/approval-ddd-boundary.md

frontend/app/api/openapi.ts
frontend/app/api/urls.ts
frontend/app/api/ims-api.d.ts
frontend/app/features/investment-ledger/services/investmentLedgerApi.ts
frontend/app/features/investment-ledger/composables/usePortfolioDirectory.ts
frontend/app/features/investment-ledger/composables/usePortfolioLedger.ts
frontend/app/features/my-funds/services/myFundsApi.ts
frontend/app/pages/investment/funds/index.vue
frontend/app/pages/investment/funds/[fundId]/decisions/index.vue
```

After reading, summarize:

```text
1. where portfolio_id is already the real source of truth
2. where fund_id/contract_id are still required
3. which files must change for the current milestone
4. which files must not be changed yet
```

Then implement the milestones below.

## Milestone 1: Database Hardening

Goal:
Make the current database safer without removing legacy fields.

Create migration pair:

```text
database/migrations/20260703000001_investment__portfolio_v2_hardening.up.sql
database/migrations/20260703000001_investment__portfolio_v2_hardening.down.sql
```

If those filenames already exist, use the next unused timestamp with the same migration name.

Implement:

1. Add `portfolio_type` to `investment__portfolios`.
   - `NOT NULL DEFAULT 'LIVE'`
   - CHECK values: `LIVE`, `SIMULATION`, `MODEL`
   - Existing rows become `LIVE`

2. Enforce active portfolio code uniqueness globally.
   - Add preflight query in comments or a docs/preflight file:

```sql
SELECT code, COUNT(*)
FROM investment__portfolios
WHERE deleted_at IS NULL
GROUP BY code
HAVING COUNT(*) > 1;
```

   - Add partial unique index:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS uq_inv_portfolios_code_alive
ON investment__portfolios (code)
WHERE deleted_at IS NULL;
```

3. Prevent fund/portfolio identity drift while legacy `fund_id` still exists.
   - Add unique constraint on `investment__portfolios (id, fund_id)`.
   - Add composite FK `(portfolio_id, fund_id) -> investment__portfolios(id, fund_id)` to:
     - `investment__portfolio_transactions`
     - `investment__decisions`
     - `investment__executions`
     - `investment__trade_confirmations`

4. Prevent contract drift while legacy `contract_id` still exists.
   - Add CHECK `(contract_id = fund_id)` to:
     - `investment__decisions`
     - `investment__executions`
     - `investment__trade_confirmations`

5. Add preflight SQL for drift detection before constraints:
   - portfolio transaction fund drift
   - decision fund/contract drift
   - execution fund/contract drift
   - trade confirmation fund/contract drift

6. Add down migration.
   - Drop constraints/indexes/column in reverse order.

Do not:
- Drop `fund_id`
- Drop `contract_id`
- Make `fund_id` nullable
- Change compliance/workflow schema
- Remove `SUBSCRIPTION` or `REDEMPTION`

Tests:

```powershell
cd backend
go test ./internal/investment/...
```

Also run any migration check/test command found in the repo.

Commit message:

```text
feat(investment): harden portfolio v2 database identity
```

## Milestone 2: Portfolio-Code V2 Read Endpoints

Goal:
Expand V2 portfolio-code routes beyond detail lookup.

Implement V2 endpoints:

```text
GET /api/v2/portfolios/{portfolioCode}/holdings
GET /api/v2/portfolios/{portfolioCode}/cash
GET /api/v2/portfolios/{portfolioCode}/transactions
GET /api/v2/portfolios/{portfolioCode}/valuations
GET /api/v2/portfolios/{portfolioCode}/valuations/latest
```

Rules:
- Resolve `portfolioCode -> portfolio_id` once in the handler layer.
- Application services and repositories continue using UUID `portfolio_id`.
- Keep V1 routes unchanged.
- Unknown portfolio code returns 404.
- Ambiguous portfolio code returns 409.
- Non-UUID portfolio code must work.
- Preserve existing permission checks.
- Update V2 Swagger and generated frontend API types.

Tests:
- Valid code works.
- Unknown code returns 404.
- Ambiguous code returns 409.
- Non-UUID code works.
- Existing V1 tests still pass.

Commit message:

```text
feat(investment): add portfolio-code v2 read endpoints
```

## Milestone 3: Portfolio-Code V2 Ledger Write Endpoints

Goal:
Move transaction write paths to portfolio-code V2 without exposing `fund_id`.

Implement:

```text
POST /api/v2/portfolios/{portfolioCode}/transactions/simulate
POST /api/v2/portfolios/{portfolioCode}/transactions
POST /api/v2/portfolios/{portfolioCode}/transactions/{transactionId}/reverse
```

Rules:
- Request body must not accept `fund_id`.
- Server derives any legacy `fund_id` from the resolved portfolio.
- Server never trusts client `fund_id` or `contract_id`.
- Existing V1 behavior remains unchanged.
- Preserve current ledger projection logic.
- Update V2 Swagger and generated frontend API types.

Tests:
- V2 simulate/post/reverse work with portfolio code.
- V2 request with no `fund_id` works.
- If client sends extra `fund_id`/`contract_id`, reject or ignore consistently and document the choice.
- Wrong portfolio code returns 404.
- Ambiguous code returns 409.
- Existing transaction tests pass.

Commit message:

```text
feat(investment): add portfolio-code v2 ledger writes
```

## Milestone 4: Frontend Portfolio Workspace Foundation

Goal:
Create frontend portfolio workspace using code-based routes.

Implement:

```text
/portfolios
/portfolios/:portfolioCode
/portfolios/:portfolioCode/overview
/portfolios/:portfolioCode/holdings
/portfolios/:portfolioCode/cash
/portfolios/:portfolioCode/ledger
```

Rules:
- Route param is `portfolioCode`.
- `usePortfolioContext()` loads by code and stores internal `portfolio.id` only for API/internal use.
- UI displays portfolio code/name, never raw UUID as the normal label.
- Keep existing fund pages working.
- Do not migrate decisions/executions yet.

Tests:
- V2 API client base URL test still passes.
- Add route/composable tests if the repo has an existing pattern.
- Run relevant frontend tests.
- Run `npm run build` if feasible.

Commit message:

```text
feat(frontend): add portfolio-code workspace foundation
```

## Milestone 5: Decisions, Executions, Confirmations V2

Goal:
Move investment decisions, executions, and confirmations under portfolio-code V2.

Rules:
- No `fund_id` or `contract_id` in new V2 request bodies.
- Backend resolves `portfolioCode -> portfolio_id`.
- Backend derives legacy contract/fund only inside adapters if compliance/workflow still need it.
- Do not refactor compliance/workflow schema yet.
- Add tests for portfolio ownership consistency.
- Keep V1 behavior intact.

Commit message:

```text
feat(investment): add portfolio-code decision workflow
```

## Milestone 6: Cleanup Readiness Report

Goal:
Do not delete columns yet. Produce a cleanup report showing what still depends on `fund_id` and `contract_id`.

Create or update a doc under `docs/handoff/` with:
- Remaining backend references to `fund_id`/`contract_id`
- Remaining frontend references
- Remaining schema references
- Which references are legitimate legacy modules
- Which references can be removed next
- Exact proposed final deletion migration

Commit message:

```text
docs(investment): document portfolio v2 cleanup readiness
```

## Definition Of Done

The run is done when:
- Milestones 1-6 are completed, or a milestone is blocked with clear evidence.
- All commits are local only.
- Nothing is pushed.
- Existing V1 app behavior remains intact.
- Portfolio V2 works by portfolio code.
- DB prevents new portfolio/fund/contract identity drift.
- Frontend has a code-based portfolio workspace foundation.
- A cleanup readiness report exists before destructive deletion.

When reporting back, list for each milestone:
- commit hash
- files changed summary
- tests run
- deferred risks
- next recommended action
