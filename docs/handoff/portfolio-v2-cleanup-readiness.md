# Portfolio V2 — Cleanup Readiness Report

**As of:** 2026-07-03
**Milestone:** 6 of `docs/handoff/portfolio-v2-claude-implementation-prompt.md`
**Status:** Survey only. No columns, constraints, or code paths are removed by this document.

This report answers: what still depends on `fund_id` and `contract_id` on the four
investment operational tables (`investment__decisions`, `investment__executions`,
`investment__trade_confirmations`, `investment__portfolio_transactions`), which of
that is legitimate to keep, which is safe to remove next, and what the eventual
deletion migration should look like.

It maps onto the phase plan already defined in `docs/api/portfolio-v2-api-ddd.md`
section 12: Milestones 1–3 and 5 delivered **Phase 0 (Hardening)** and **Phase 1
(API V2 Additive)**; Milestone 4 delivered only the *foundation* of **Phase 2
(Frontend Cutover)**, not the cutover itself. **Phase 3 (Schema Simplification)** —
dropping `contract_id` then `fund_id` — is not ready to start; this report is the
readiness check for it.

---

## 1. Bottom line

- **Not ready to drop any column yet.** V1 fund-first screens and V1
  fund/portfolio/contract-scoped write endpoints are still the primary, actively
  used surface. Phase 2 (frontend cutover) has not happened — Milestone 4 only
  added the `/portfolios` workspace *alongside* the existing `/investment/funds/[fundId]/*`
  tree, it did not replace anything.
- **`contract_id` is the nearer-term removal target**, not `fund_id`. On these four
  tables `contract_id` has no `REFERENCES` clause of its own — it has always been a
  bare mirror of `fund_id`, now pinned equal by the Milestone 1
  `chk_inv_*_contract_is_fund` CHECK constraints. `fund_id` is a real `NOT NULL` FK
  into `investment__funds`, used in FK chains, indexes, list filters, and response
  DTOs — a materially bigger removal.
- **The blocking work for `contract_id` removal is in Go, not SQL.** ~40 call sites
  across the investment module read `.ContractID` off decision/execution/confirmation
  entities and pass it into `workflow.IsTransactionLocked`, `workflow.IsTradeAllowed`,
  `permissions.HasDataPermission`, compliance breach records, and response DTOs.
  Because `contract_id = fund_id` is now DB-enforced, every one of these can be
  repointed to `.FundID` with no behavior change — but each site has to be touched
  and the CHECK constraint gives good confidence the substitution is currently
  safe, not that it always will be.
- **No destructive action is proposed in this milestone.** Section 6 is a draft for
  when Phase 3 is actually scheduled, not a migration to run now.

---

## 2. Backend references

### 2a. Legitimate — Fund entity, not legacy drift

These are a genuine `Portfolio belongs to Fund` relationship or genuine Fund-level
read paths. Do not touch as part of Portfolio V2 cleanup:

| File | What it does |
|---|---|
| `application/command/portfolio_crud.go` | `CreatePortfolioRequest.FundID` — creating a portfolio requires naming its owning fund. |
| `transport/dto/request/requests.go:54` (`CreatePortfolioRequest.FundID`) | Same, at the transport boundary. |
| `infrastructure/persistence/fund_repository.go`, `application/command/fund_crud.go`, `application/query/get_fund_nav_history.go`, `application/query/get_latest_fund_nav.go`, `application/query/get_fund_allocation.go`, `application/command/compute_fund_aum.go`, `domain/entity/fund.go` | Fund-level CRUD, NAV, AUM, allocation. Operate on `investment__funds` directly; `fund_id` here identifies the Fund itself, not a legacy duplicate on another table. |
| `investment__portfolios.fund_id` (schema) | Real FK, `NOT NULL REFERENCES investment__funds(id)`. This is the anchor Milestone 1's composite FK chain is built on top of — it stays permanently. |

### 2b. V1 legacy compat surface — superseded by V2, not yet removable

Only **decision creation** in V1 actually requires the *client* to supply
`fund_id`/`portfolio_id`/`contract_id` directly:

- `transport/dto/request/requests.go:208-224` — `CreateDecisionRequest`:
  ```go
  type CreateDecisionRequest struct {
      FundID           uuid.UUID  `json:"fund_id"      validate:"required"`
      PortfolioID      uuid.UUID  `json:"portfolio_id" validate:"required"`
      ContractID       uuid.UUID  `json:"contract_id"  validate:"required"`
      ...
  }
  ```
  This is exactly what V2's `CreateDecisionByCode` (Milestone 5) replaces — V2
  derives all three from the resolved portfolio and the request body has none of
  them (`transport/dto/request/requests_v2.go`, `CreateDecisionV2Request`).

**Finding, worth flagging explicitly:** execution and confirmation creation in V1
were **already** deriving `fund_id`/`portfolio_id`/`contract_id` internally, not
trusting the client, before this migration touched them:
- `application/command/execution.go:139-141` — `CreateExecutionRequest` has no
  identity fields at all; `Create()` copies `FundID`/`PortfolioID`/`ContractID` from
  the loaded parent decision (`d.FundID`, `d.PortfolioID`, `d.ContractID`).
- `application/command/trade_confirmation.go:100-102` — same pattern, copied from
  the loaded parent execution.
- `application/command/post_transaction.go:174,476` and
  `application/command/reverse_transaction.go:112` — `FundID`/`ContractID` derived
  from the resolved portfolio/fund, not client input; `request.PostTransactionRequest`
  and `request.ReverseTransactionRequest` never had a `fund_id`/`contract_id` field
  (confirmed in Milestone 3).

So V2's Milestone 5 work on executions/confirmations was really about
**portfolio-code resolution and cross-portfolio ownership verification**, not about
closing a client-trust gap — that gap only ever existed for decision creation.

**Removable next:** once the V1 `/investment/decisions` POST route is deprecated or
retired in favor of the V2 portfolio-code route, `CreateDecisionRequest.FundID` /
`.PortfolioID` / `.ContractID` and their `validate:"required"` tags can be deleted
and the handler simplified to only accept what V2 accepts. **Not yet** — V1 must stay
intact per the migration's non-negotiable rules, and the frontend (`[fundId]/operation/new.vue`,
see §3) is still the only decision-creation UI in production.

### 2c. Internal derivation/storage — needed until the schema itself changes

`fund_id`/`contract_id`/`portfolio_id` are entity fields on `entity.Decision`,
`entity.Execution`, `entity.TradeConfirmation` (all three), and `entity.PortfolioTransaction`
(`fund_id` only — this table has no `contract_id` column, see §4), persisted through
every repository's SELECT/INSERT/UPDATE column list
(`infrastructure/persistence/decision_repository.go`,
`execution_repository.go`, `trade_confirmation_repository.go`,
`transaction_repository.go`). These are load-bearing for as long as the columns
exist — they are not "extra" code to remove, they are the ORM mapping. They will
shrink automatically once §6's migration runs; no separate Go cleanup is needed
beyond the call sites in §2d.

`domain.DecisionListFilter.ContractID`/`.FundID` (`decision_repository.go:196,202`,
consumed by `decision_handler.go:110`) are query filters — legitimate as long as
the columns exist, removable in the same pass as the columns.

### 2d. Cross-module `contract_id` fan-out (the real blocker for Phase 3, part 1)

`contract_id` is not investment-local. Grepping `.ContractID` across
`backend/internal/investment` surfaces roughly 40 call sites, and 14 migration files
outside `investment__decisions/executions/trade_confirmations` reference `contract_id`
on their own tables (`workflow__*`, `permissions_data_rights`, `compliance__*`,
`approval__*`, `audit__financial_query_view`). Representative investment-module call
sites that must be repointed to `.FundID` before `contract_id` can be dropped from
these three tables:

| File:line | Usage |
|---|---|
| `application/command/execution.go:115` | `h.workflow.IsTransactionLocked(ctx, d.ContractID, d.BusinessDate)` |
| `application/command/decision_lifecycle.go:366,732` | `h.workflow.IsTradeAllowed(ctx, d.ContractID, d.BusinessDate)` |
| `application/query/can_execute_investment_process.go:67,72,77,83,90,118` | `req.ContractID` used as the workflow/process-guard key throughout |
| `application/command/decision_batch_approval.go:103,151` | `h.perms.HasDataPermission(ctx, actorID, ref.ContractID.String())` — object-level data-scope check |
| `infrastructure/adapter/investment_subject_access.go:140` | Returns `d.ContractID` as the subject's data-scope key for the Approval module |
| `domain/policy/report_reference.go:102-103` | `ApplicableContractID` match against `in.ContractID` for research-report applicability |
| `domain/errors.go:172,185`, `domain/decision_lifecycle_errors.go:164,171` | `ContractID` embedded in compliance-breach / batch-approval error payloads |
| `transport/dto/response/mappers.go:495,563,596` | `ContractID` exposed in `DecisionResponse`/`ExecutionResponse`/`TradeConfirmationResponse` (V1 responses only — V2 response DTOs from Milestone 2/5 do not expose it) |

Because Milestone 1's `chk_inv_*_contract_is_fund` CHECK constraints now guarantee
`contract_id = fund_id` for every row, every one of these reads can be mechanically
changed to `.FundID` with identical runtime behavior **today**. That substitution
itself is not part of this milestone (it's a code change, this milestone is docs
only) but it is the concrete, scoped unit of work that unblocks dropping the column.

The `permissions_data_rights` and workflow/compliance tables have their **own**
`contract_id` columns — they are not FK'd to `investment__decisions.contract_id`.
Repointing investment call sites to pass `FundID` instead of `ContractID` into those
modules' functions does not require changing those other modules' schemas, since
`contract_id` and `fund_id` are already the same UUID value.

---

## 3. Frontend references

Grepping `frontend/app` for `fund_id`/`fundId`/`contract_id`/`contractId` returns 71
files. The overwhelming majority are the `[fundId]` dynamic route segment under
`frontend/app/pages/investment/funds/[fundId]/**` — the entire V1 fund-first
workspace (holdings, decisions, operation, stages, settings, reviewers, compliance,
audit tabs). This is **legitimate legacy** in the sense that it is the currently
shipping UI, not drift — Milestone 4 built `/portfolios/*` as a parallel foundation,
it did not migrate these routes. Phase 2 (frontend cutover, per the DDD doc) is the
milestone that would retire or redirect this tree, and it has not run.

**Confirmed legacy-but-still-required (V1 decision create UI):**
`frontend/app/pages/investment/funds/[fundId]/operation/new.vue:78-79,97` sends
`fund_id: fundId.value` and `contract_id: fundId.value` in the create-decision
request body, matching V1's `CreateDecisionRequest` (§2b). This is the frontend
counterpart of the backend legacy surface — it stays until this screen is replaced
by a portfolio-code equivalent under `/portfolios/[portfolioCode]/decisions/new`
(not yet built; out of scope for Milestone 4, which only built read views —
Overview/Holdings/Cash/Ledger — not a decision-create form).

**Leave alone (derived, not hand-written):** `frontend/app/api/ims-api.d.ts` — generated
from swagger, mirrors whatever the backend DTOs expose. Its `fund_id`/`contract_id`
fields will disappear automatically the day the backend DTOs stop exposing them; no
manual frontend type edit is ever needed for this.

**Dead code — no live UUID-display risk, but worth flagging for removal or fixing before
reuse:** `frontend/app/shared/ui/IMSContractSelector.vue` renders `{{ opt.contractId }}`
directly in a visible `<span>` text node (lines 72-73). Grepping the entire frontend
app for `IMSContractSelector` finds **zero import sites** — this component is not
wired into any page. It currently violates nothing live, but if `opt.contractId` were
ever a raw UUID and someone imported this component, it would violate the documented
rule in `docs/handoff/frontend-uuid-free-policy.md` ("A raw UUID string must never be
placed in a rendered text node that is visible to the user"). Recommend either
deleting it as unused, or — if it's meant to be revived — resolving `contractId` to a
fund code/name before display, consistent with every other selector in the app.

No other frontend file showed a raw UUID being rendered as visible text for
`fund_id`/`contract_id`; the rest of the 71 matches are route params used purely for
navigation/API-path construction (an approved pattern per the UUID-free policy) or
list-filter query parameters (`decisionApi.ts:28-29,55-56`).

---

## 4. Schema references

All four tables plus the legitimate `investment__portfolios` anchor:

| Table | Column | Nullable | Constraints | Introduced in |
|---|---|---|---|---|
| `investment__decisions` | `fund_id` | `NOT NULL` | `REFERENCES investment__funds(id) ON DELETE RESTRICT`; composite `fk_inv_decision_portfolio_fund FOREIGN KEY (portfolio_id, fund_id) REFERENCES investment__portfolios(id, fund_id)` | `20260601000001` (base FK); `20260703000001` (composite FK) |
| `investment__decisions` | `contract_id` | `NOT NULL` | No direct FK. `chk_inv_decisions_contract_is_fund CHECK (contract_id = fund_id)` | `20260601000001` (column); `20260703000001` (CHECK) |
| `investment__executions` | `fund_id` | `NOT NULL` | `REFERENCES investment__funds(id) ON DELETE RESTRICT`; composite `fk_inv_execution_portfolio_fund` | `20260601000001`; `20260703000001` |
| `investment__executions` | `contract_id` | `NOT NULL` | No direct FK. `chk_inv_executions_contract_is_fund CHECK (contract_id = fund_id)` | `20260601000001`; `20260703000001` |
| `investment__trade_confirmations` | `fund_id` | `NOT NULL` | `REFERENCES investment__funds(id) ON DELETE RESTRICT`; composite `fk_inv_confirmation_portfolio_fund` | `20260601000001`; `20260703000001` |
| `investment__trade_confirmations` | `contract_id` | `NOT NULL` | No direct FK. `chk_inv_confirmations_contract_is_fund CHECK (contract_id = fund_id)`. Also indexed: `idx_inv_confirmation_contract_date (contract_id, business_date DESC)` | `20260601000001`; `20260703000001` |
| `investment__portfolio_transactions` | `fund_id` | `NOT NULL` | `REFERENCES investment__funds(id) ON DELETE RESTRICT`; composite `fk_inv_txn_portfolio_fund`. Also indexed: `idx_inv_txn_fund_date (fund_id, business_date DESC)` | `20260428000004`; `20260703000001` |
| `investment__portfolio_transactions` | *(no `contract_id` column)* | — | — | — |
| `investment__portfolios` | `fund_id` | `NOT NULL` | `REFERENCES investment__funds(id)`; `UNIQUE (id, fund_id)` (`uq_inv_portfolios_id_fund`, added `20260703000001` as the composite-FK anchor); indexed `idx_inv_portfolios_fund` | `20260428000002`; `20260703000001` |

Note `investment__portfolio_transactions` never had a `contract_id` column — the
"drop `contract_id`" step in §6 only touches decisions/executions/trade_confirmations.

Seed files under `database/seeds/investment/` contain no `fund_id`/`contract_id`
references — decisions/executions/confirmations are not seed data, they're created
only through application flows or tests, so no seed migration work is needed.

E2E test seeding (`backend/tests/e2e/investment_test.go`, `seedInvestmentPrereqs`)
inserts `fund_id` directly when creating a portfolio row — legitimate use (the same
category as §2a), not something the deletion migration touches.

---

## 5. What can be removed next vs. what's legitimate

| Item | Verdict |
|---|---|
| `investment__decisions/executions/trade_confirmations.contract_id` + their CHECK constraints + `idx_inv_confirmation_contract_date` | **Removable next**, once the ~40 Go call sites in §2d are repointed to `.FundID`. This is the smaller, more contained half of Phase 3. |
| V1 `CreateDecisionRequest.FundID/.PortfolioID/.ContractID` (backend) and `[fundId]/operation/new.vue`'s fund_id/contract_id body fields (frontend) | **Removable next, but blocked on Phase 2**: needs a V2 decision-create screen to replace this UI first, then the V1 route can be deprecated. |
| `fund_id` on all four tables, and the composite FK chain built on it | **Not removable yet.** Real FK, `NOT NULL`, backs list filters (`fund_id` query param on `/investment/decisions`, `/investment/executions`), response DTOs, and reporting indexes. Needs every read path to be proven derivable via `portfolio_id → investment__portfolios.fund_id` join before it can go — materially bigger than the `contract_id` step. Matches the DDD doc's own Phase 3 ordering ("drop `contract_id`... drop `fund_id`... once all reads derive context from portfolio"). |
| `investment__portfolios.fund_id` | **Never removable** under the current architecture — this is the genuine Portfolio-belongs-to-Fund relationship, not legacy drift. |
| `IMSContractSelector.vue` | **Removable now** (unused), independent of the schema migration — a housekeeping item, not a Portfolio V2 dependency. |

---

## 6. Proposed final deletion migration (draft — do not run yet)

This is a draft for when Phase 3's `contract_id` step is actually scheduled, i.e.
after the §2d Go call sites are repointed to `.FundID` and a build/test pass confirms
no remaining `.ContractID` reference on these three entities. **Do not apply this
migration now** — no code has been changed to stop depending on these columns yet,
consistent with the working rule "before destructive DB changes, prove no code path
depends on the field anymore."

```sql
-- database/migrations/<timestamp>_investment__drop_decision_contract_id.up.sql
--
-- Preflight (run first — must return zero rows, or the equality this migration
-- assumes has already been violated and needs investigating before dropping the
-- column that currently enforces it):
--   SELECT id FROM investment__decisions          WHERE contract_id <> fund_id;
--   SELECT id FROM investment__executions         WHERE contract_id <> fund_id;
--   SELECT id FROM investment__trade_confirmations WHERE contract_id <> fund_id;
--
-- Prerequisite: every backend .ContractID read on Decision/Execution/
-- TradeConfirmation entities has been repointed to .FundID (see cleanup
-- readiness report section 2d for the call-site list) and V1 response DTOs
-- (transport/dto/response/mappers.go) no longer serialize contract_id for
-- these three resources.

DROP INDEX IF EXISTS idx_inv_confirmation_contract_date;

ALTER TABLE investment__decisions
    DROP CONSTRAINT IF EXISTS chk_inv_decisions_contract_is_fund,
    DROP COLUMN IF EXISTS contract_id;

ALTER TABLE investment__executions
    DROP CONSTRAINT IF EXISTS chk_inv_executions_contract_is_fund,
    DROP COLUMN IF EXISTS contract_id;

ALTER TABLE investment__trade_confirmations
    DROP CONSTRAINT IF EXISTS chk_inv_confirmations_contract_is_fund,
    DROP COLUMN IF EXISTS contract_id;
```

```sql
-- database/migrations/<timestamp>_investment__drop_decision_contract_id.down.sql

ALTER TABLE investment__decisions
    ADD COLUMN IF NOT EXISTS contract_id UUID;
UPDATE investment__decisions SET contract_id = fund_id WHERE contract_id IS NULL;
ALTER TABLE investment__decisions
    ALTER COLUMN contract_id SET NOT NULL,
    ADD CONSTRAINT chk_inv_decisions_contract_is_fund CHECK (contract_id = fund_id);

ALTER TABLE investment__executions
    ADD COLUMN IF NOT EXISTS contract_id UUID;
UPDATE investment__executions SET contract_id = fund_id WHERE contract_id IS NULL;
ALTER TABLE investment__executions
    ALTER COLUMN contract_id SET NOT NULL,
    ADD CONSTRAINT chk_inv_executions_contract_is_fund CHECK (contract_id = fund_id);

ALTER TABLE investment__trade_confirmations
    ADD COLUMN IF NOT EXISTS contract_id UUID;
UPDATE investment__trade_confirmations SET contract_id = fund_id WHERE contract_id IS NULL;
ALTER TABLE investment__trade_confirmations
    ALTER COLUMN contract_id SET NOT NULL,
    ADD CONSTRAINT chk_inv_confirmations_contract_is_fund CHECK (contract_id = fund_id);

CREATE INDEX IF NOT EXISTS idx_inv_confirmation_contract_date
    ON investment__trade_confirmations (contract_id, business_date DESC);
```

**`fund_id` removal is deliberately not drafted here.** It requires: (1) Phase 2
frontend cutover complete (V1 fund-first routes retired or fully redirected to
portfolio-code routes), (2) `/investment/decisions|executions|confirmations`
list-filter query params (`fund_id=`) either removed or reimplemented as a join
through `portfolio_id`, (3) response DTOs no longer serializing `fund_id` directly
off these rows, and (4) the four composite FKs (`fk_inv_txn_portfolio_fund` etc.)
redesigned or dropped, since they're defined against `(portfolio_id, fund_id)` on
these very columns. That is a separate, larger migration to draft once (1)-(3) are
actually true — premature to draft its SQL now.

---

## 7. Recommended next action

Not part of this milestone, listed for the next planning cycle:
1. Repoint the §2d Go call sites from `.ContractID` to `.FundID` (mechanical,
   low-risk given the CHECK constraint guarantee), then apply §6's migration.
2. Build a portfolio-code decision-create screen under `/portfolios/[portfolioCode]/decisions`
   to give `[fundId]/operation/new.vue` (§3) a V2 replacement, enabling V1's
   `CreateDecisionRequest` identity fields to be deprecated.
3. Delete `IMSContractSelector.vue` (dead code, independent of the above).
4. Only after (1)-(2) are live in production for a full cycle: begin planning the
   `fund_id` removal, scoped separately per §6.
