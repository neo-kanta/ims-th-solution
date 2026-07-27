# LIVE Cash-Transaction Approval Gate — Design & Implementation Spec

**Task:** IMS-PORTFOLIO-FUND-OPTIONAL Stage 2
**Status:** Spec approved by owner (2026-07-24); backend implementation dispatched.
**Author:** manager session (Opus 4.8), grounded in live source verification.

> This is the single source of truth for Stage 2. It doubles as (a) the design
> doc, (b) the backend subagent's implementation spec, and (c) the basis for the
> cross-account handoff prompt. Do not re-derive policy — the three policy
> questions below were confirmed by the owner via AskUserQuestion on 2026-07-24.

---

## 1. Confirmed policy (owner, 2026-07-24)

| Portfolio type | Cash movement behavior (CASH_IN / CASH_OUT / FEE / DIVIDEND) |
|----------------|-------------------------------------------------------------|
| `SIMULATION`   | Post **immediately** — current behavior, unchanged.          |
| `MODEL`        | **Blocked** from the ledger entirely — unchanged.            |
| `LIVE`         | **Full approval workflow required** before the real transaction posts. |

Three confirmed answers:

1. **Gate scope = all cash movements**: `CASH_IN`, `CASH_OUT`, `FEE`, `DIVIDEND`.
   - `BUY` / `SELL` / `SUBSCRIPTION` / `REDEMPTION` / `REVERSAL` are **out of scope** —
     they keep posting via the existing path. (BUY/SELL gating would require the
     missing decision→execution settlement bridge "G1", which the owner did not
     select.)
2. **Submitter visibility = yes**: the submitter sees their own pending request
   (status Pending / Approved / Rejected) in the ledger view.
3. **Submitter cancel = yes**: the submitter can cancel a `PENDING` request before
   an approver acts; after cancel, approvers can no longer act on it.

---

## 2. Why a new table is required (append-only constraint)

`investment__portfolio_transactions` is **append-only**, enforced by DB triggers
`trg_inv_portfolio_transactions_no_update` and `trg_inv_portfolio_transactions_no_delete`.
Therefore a pending LIVE cash request **cannot** be inserted as a real transaction
row and mutated after approval. It needs its **own** pending-request entity/table
that supports status transitions, and the real transaction row is materialized
**only after approval**, inside the approval callback.

This mirrors `INVESTMENT_DECISION`: `decision_lifecycle.go`'s `Submit()` submits a
subject to the approval engine, and `DecisionApprovalCallback.OnApprovalDecision`
(`infrastructure/adapter/decision_approval_adapter.go`) advances the business
object when the decision is final. **`decision_approval_adapter.go` is the closest
template.**

---

## 3. Verified codebase facts (do not re-derive)

- **Approval contract** (`backend/pkg/contract/approval.go`): all needed interfaces
  already exist — `ApprovalSubmitter.SubmitForApproval`, `ApprovalSubjectCallback.OnApprovalDecision`,
  `ApprovalCanceller.CancelApprovalBySubject`, `SubjectAccessor` (object-level access),
  `ApprovalStatusProvider.GetApprovalStage`, `ApprovalSubjectValidator.ValidateSubjectApprovable`.
- **Process/subject type registration**:
  - `vo.ProcessType` + `ValidProcessType` and `vo.SubjectType` + `ValidSubjectType`
    in `backend/internal/approval/domain/valueobject/enums.go` are Go whitelists —
    submitting an unlisted type is rejected.
  - `approval__process_configs.process_type` has a DB CHECK constraint
    (`chk_approval_process_type`) that must be extended. **Template migration:
    `20260613000006_approval__add_portfolio_onboarding_process_type.{up,down}.sql`.**
  - `approval__requests` has **no** `subject_type` CHECK (domain validates) — no
    schema change needed for subject type.
  - A process **config must be seeded** for the new process type before any submit
    can resolve a config (`runtime_service.go` → `repo.Resolve(processType, ...)`).
    Seed template: `database/seeds/004_investment_process_assignment_seed.sql`,
    `015_approval_compliance_release_seed.sql`.
- **Callback + access registration** happens in `backend/cmd/server/main.go`:
  - `approvalModule.RegisterSubjectCallback("<SUBJECT_TYPE>", ...)` (lines ~176-178)
  - `approvalModule.RegisterSubjectAccessPort("<SUBJECT_TYPE>", investSubjectAccessor)` (lines ~192-194)
- **Existing invest subject accessor**: `internal/investment/infrastructure/adapter/investment_subject_access.go`
  already handles data-scope for approval subjects; extend it (or add a sibling) for `CASH_TRANSACTION`.
- **The gate point**: `PostTransactionHandler.Handle` in
  `backend/internal/investment/application/command/post_transaction.go` currently
  posts unconditionally for every portfolio type. This is where the LIVE-cash
  branch is inserted.
- **Fund-optional context**: `fund_id` is nullable on all trading tables
  (migrations `20260723000001`, `20260723000003`). The new pending-cash table must
  also carry a **nullable** `fund_id` and use the portfolio's own id as the
  data-scope key when there is no fund (mirror `portfolioScopeID`/`transactionScopeID`).
- **DB CHECK constraints on `investment__portfolio_transactions`** (must be honored
  when materializing): `chk_inv_txn_type` (allowed types), `chk_inv_txn_cash_only_fields`
  (cash types must have NULL instrument_id/quantity/price), `chk_inv_txn_currency`
  (`^[A-Z]{3}$`), `chk_inv_txn_fees_non_negative`, `chk_inv_txn_status`
  (`POSTED`/`REVERSED`).

---

## 4. Backend design

### 4.1 New migrations (matching up + down)

Pick the next sequential numbers after `20260723000003` (use today's date prefix,
e.g. `20260724000001`, verify no collision in `database/migrations/`):

1. **`..._investment__portfolio_cash_requests.{up,down}.sql`** — new table:

   ```
   investment__portfolio_cash_requests
     id                     uuid PK default gen_random_uuid()
     portfolio_id           uuid NOT NULL  FK → investment__portfolios(id)
     fund_id                uuid NULL      -- mirrors fund-optional; scope fallback = portfolio_id
     transaction_type       varchar NOT NULL CHECK IN ('CASH_IN','CASH_OUT','FEE','DIVIDEND')
     amount                 numeric NOT NULL CHECK (amount > 0)   -- sign conveyed by type
     currency               varchar NOT NULL CHECK (~ '^[A-Z]{3}$')
     fees                   numeric NOT NULL DEFAULT 0 CHECK (fees >= 0)
     value_date             date NOT NULL
     memo / description     text NULL
     status                 varchar NOT NULL CHECK IN
                              ('PENDING','APPROVED','REJECTED','CANCELLED')  default 'PENDING'
     approval_request_id    uuid NULL   -- returned by SubmitForApproval
     resulting_txn_id       uuid NULL   -- set once materialized (idempotency key)
     submitted_by           uuid NOT NULL
     submitted_at           timestamptz NOT NULL default now()
     decided_by             uuid NULL
     decided_at             timestamptz NULL
     version                int NOT NULL default 1
     created_at/updated_at/created_by/updated_by  (standard audit cols)
   ```
   - This table is **mutable** (status transitions) — do NOT add append-only triggers.
   - Index on `(portfolio_id, status)` for the submitter read path.
   - `down.sql` drops the table.

2. **`..._approval__add_cash_transaction_process_type.{up,down}.sql`** — extend
   `chk_approval_process_type` to add `'PORTFOLIO_CASH_TRANSACTION'` (copy the exact
   template migration; `up` drops + re-adds the constraint with the new value list,
   `down` restores the prior list).

### 4.2 New seed (idempotent)

`database/seeds/0XX_cash_transaction_process_seed.sql` — an
`approval__process_configs` row (+ stage/approver config) of type
`PORTFOLIO_CASH_TRANSACTION`, `ContractType = COMPANY` (portfolio-scoped, not
fund-scoped, so it resolves for fund-less portfolios too). Use `ON CONFLICT` for
idempotency. Model on `004`/`015`. Route to the same approver group used by
`INVESTMENT_DECISION` (ben→green) unless the owner specifies otherwise — a single
`GROUP_ANY` stage is acceptable for v1; note the choice in the seed comment.

### 4.3 New enum values

In `enums.go`:
- `ProcessPortfolioCashTransaction ProcessType = "PORTFOLIO_CASH_TRANSACTION"` +
  add to `ValidProcessType`.
- `SubjectCashTransaction SubjectType = "CASH_TRANSACTION"` + add to `ValidSubjectType`.

### 4.4 Domain + application (investment module)

- New entity `entity.PortfolioCashRequest` (`domain/entity/`), value object for
  status if warranted.
- New repository interface method(s) in `domain/repository.go` +
  Postgres impl in `infrastructure/persistence/`: `Create`, `GetByID`,
  `ListByPortfolio(portfolioID, statusFilter)`, `UpdateStatus`/`MarkApproved`/
  `MarkRejected`/`MarkCancelled` (with `resulting_txn_id` + decided_by/at),
  `GetBySubjectID` (for the callback).
- **Gate in `PostTransactionHandler.Handle`** (`post_transaction.go`):
  1. Resolve portfolio type (reuse `resolveCompliancePortfolioType` or portfolio load).
  2. If `transaction_type ∈ {CASH_IN,CASH_OUT,FEE,DIVIDEND}` **and** type == `LIVE`:
     - Run the same pre-post compliance/validation the immediate path runs (so a
       request that would be rejected is caught at submit time, not approval time —
       but the authoritative re-check must ALSO run at materialization).
     - Create a `PortfolioCashRequest` (status `PENDING`).
     - Call `approval.SubmitForApproval` with
       `ProcessType="PORTFOLIO_CASH_TRANSACTION"`, `SubjectType="CASH_TRANSACTION"`,
       `SubjectID=<cash request id>`, `PortfolioID`, `SubmitterID`,
       `ContractType="COMPANY"` (or `FUND` + `ContractID=fund_id` when a fund
       exists — match how decisions choose contract scope).
     - Persist `approval_request_id`.
     - Return a **distinct result** the transport layer maps to a "submitted for
       approval / PENDING" response (NOT a posted-transaction response).
  3. Else (SIMULATION cash, or any BUY/SELL/etc., or LIVE non-cash): **unchanged**
     immediate post.
  - **MODEL** stays blocked exactly as today (`canEnterLedgerTransaction` on FE +
     whatever backend guard exists — verify the backend also refuses, don't rely on FE).
- **Materialization** — new callback adapter
  `infrastructure/adapter/cash_request_approval_adapter.go` implementing
  `contract.ApprovalSubjectCallback` (template: `decision_approval_adapter.go`):
  - `OnApprovalDecision`:
    - Load the cash request by `decision.SubjectID`.
    - **Idempotency guard**: if status already terminal (`APPROVED`/`REJECTED`/
      `CANCELLED`) or `resulting_txn_id` set → no-op return nil (duplicate callback safe).
    - If `decision.Approved`:
      - **Re-run** the authoritative compliance/cash validation now (state may have
        changed since submit).
      - Materialize the real `investment__portfolio_transactions` row via the shared
        post logic (extract the core of `PostTransactionHandler` into a reusable
        internal method so both the immediate path and the callback use identical
        posting/ledger/position logic — do NOT duplicate).
      - Set request `status=APPROVED`, `resulting_txn_id`, `decided_by/at`, in the
        **same DB transaction** as the transaction insert (all-or-nothing).
    - Else (rejected): set `status=REJECTED`, `decided_by/at`, post nothing.
  - Register in `main.go`: `RegisterSubjectCallback("CASH_TRANSACTION", investmentModule.CashRequestApprovalSubjectCallback())`.
- **Subject access**: extend `investment_subject_access.go` (or add a sibling)
  to answer `CanView/CanSubmit/CanActOn` for `CASH_TRANSACTION` using the cash
  request's portfolio → fund/portfolio data-scope (reuse the portfolio scope
  helper; fund-less falls back to portfolio id). Register via
  `RegisterSubjectAccessPort("CASH_TRANSACTION", ...)` in `main.go`.
- **Cancel path**: new command `CancelCashRequest(requestID, actorID)`:
  - Verify actor == submitter (or has an admin/override permission — v1: submitter only).
  - Verify status == `PENDING`.
  - Call `approvalCanceller.CancelApprovalBySubject("CASH_TRANSACTION", requestID, actorID)`.
  - Set request `status=CANCELLED`, `decided_by/at`.
- **Validator** (optional but recommended): implement `ApprovalSubjectValidator`
  so an approver can't approve a request the submitter already cancelled.

### 4.5 Transport (V2)

- Extend the existing V2 portfolio routes under
  `/api/v2/portfolios/{code}/...`. Suggested:
  - `POST /api/v2/portfolios/{code}/transactions` — **existing** route; the handler
    now returns a "pending approval" response body for LIVE cash instead of a posted txn.
    (Keep the response shape backward-compatible; add a `status`/`pending_request`
    discriminator so the FE can tell posted vs pending.)
  - `GET  /api/v2/portfolios/{code}/cash-requests?status=` — submitter's pending list.
  - `POST /api/v2/portfolios/{code}/cash-requests/{id}/cancel` — cancel.
  - All routes enforce data-scope via `h.pc` (the P0-A `PermissionChecker` field) —
    do NOT pass literal `nil` (that was the P0-A vulnerability).
- New request/response DTOs in `transport/dto/request|response`. No domain entities
  exposed directly.
- Error mapping: bad input → 4xx; missing portfolio scope → **403** (fix the
  known fund-less 500→403 mapping gap while here if it touches this path);
  infra → 5xx.
- **Regenerate Swagger + OpenAPI client** (`make swagger` → `make api-client`) —
  never hand-edit generated files, never hand-write FE types.

### 4.6 Audit

Preserve attributable audit history: emit audit events for cash-request
submitted / approved / rejected / cancelled / materialized, mirroring the
`INVESTMENT_DECISION_*` audit actions in `decision_lifecycle.go`.

---

## 5. Frontend design (second subagent, after backend + regen)

Depends on the regenerated `ims-api.d.ts`. Do NOT start until backend API + swagger
+ `make api-client` are done (otherwise it would hand-write types — forbidden).

- **Ledger view** (`frontend/app/features/portfolio-workspace/PortfolioLedgerNewView.vue`,
  `useOrderTicket.ts`): when portfolio is `LIVE` and the movement is a cash type,
  the submit action posts and the UI shows "Submitted for approval — pending"
  instead of "Posted". SIMULATION unchanged; MODEL still blocked
  (`canEnterLedgerTransaction`).
- **Pending requests panel**: list the submitter's `PENDING` cash requests with
  status + a **Cancel** button (calls the cancel endpoint). Show Approved/Rejected
  outcomes too.
- **Cash-movement UI gap (G2)**: the current ledger form only emits BUY/SELL. To
  test cash at all, the FE must expose CASH_IN/CASH_OUT/FEE/DIVIDEND. Confirm scope
  with owner — this may be a prerequisite the FE agent must also build.
- **i18n**: EN/TH/ZH copy for all new labels/states/errors
  (`shared/i18n/messages/{en,th,zh}/portfolio.ts`), zh = Traditional.
- Use `useApi()`/`useOpenApiClient()` + generated types only.

---

## 6. Test matrix (required)

Backend (`go test`), near the layer changed:

- LIVE cash → creates PENDING request, submits to approval, posts **no** txn.
- SIMULATION cash → posts immediately (regression — unchanged).
- MODEL cash → blocked (regression).
- LIVE BUY/SELL → posts immediately (out of gate scope — regression).
- Approval APPROVED callback → materializes exactly one real txn, status APPROVED,
  cash/positions updated, `resulting_txn_id` set.
- Approval REJECTED callback → no txn, status REJECTED.
- **Duplicate callback** (approved twice) → idempotent, single txn.
- Cancel PENDING by submitter → status CANCELLED, approval cancelled, no txn.
- Cancel non-PENDING / by non-submitter → rejected.
- Cross-fund / missing-portfolio-scope access on all new routes → **403**.
- Materialization compliance re-check BLOCK → no txn, request marked appropriately,
  transaction rolled back (all-or-nothing).
- Fund-less LIVE portfolio → whole flow works (scope key = portfolio id).

Frontend (Vitest): pending-state rendering, cancel action, MODEL still blocked,
i18n parity.

---

## 7. Hard constraints (from owner + MEMORY.md)

- Preserve DDD boundaries; cross-module only via `backend/pkg/contract`.
- Business logic out of HTTP handlers.
- Every `.up.sql` has a matching `.down.sql`.
- Do NOT weaken the append-only guarantee on `investment__portfolio_transactions`.
- Backend-authoritative permissions/data-scope; FE guards are UX only.
- Attributable, immutable audit trail.
- **Do NOT commit or push** — owner is the sole committer.
- If Docker images must be rebuilt, run `docker-compose build` in the **foreground**
  (backgrounded builds were silently killed twice in a prior session).
- Regenerate Swagger + OpenAPI client if the API changes; never hand-edit generated
  output or hand-write FE API types.

---

## 8. Sequencing

1. **Backend** (this dispatch): migrations + enums + entity/repo + gate + callback +
   subject access + cancel + transport + audit + tests + swagger/api-client regen.
2. **Frontend** (after backend + regen): ledger pending UI + cancel + cash-movement
   inputs + i18n + tests.
3. Owner reviews, runs live browser UAT, and commits.

Stage 1 (fund-optional verified work, ~30 backend files + FE + migrations
`20260723000001`/`20260723000003`) is **still uncommitted** and must be committed by
the owner independently — Stage 2 builds on top of it.
