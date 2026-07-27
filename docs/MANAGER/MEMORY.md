---
type: manager-memory
project: IMS Thailand
owner: Kanta
status: active
last_verified: 2026-07-24
stability: durable
---

# Manager Memory

This file contains durable product decisions, architecture rules, and operating
constraints for any AI acting as the engineering manager. Read this file before
`HANDOFF.md` and `TASKS.md` when starting a new instance.

## Source-of-Truth Order

When sources disagree, use this order and report the conflict:

1. The owner's latest explicit decision.
2. Executable code, current migrations, and tests.
3. `docs/MANAGER/MEMORY.md`.
4. Current module and DDD documentation.
5. Historical handoff documents and generated API documentation.

Never treat a target-design document as proof that a feature is implemented.
`docs/AI_CONTEXT.md` and `docs/AI_HANDOFF.md` are incomplete templates and are
not authoritative while they still contain TODO placeholders.

## Product Direction

- IMS is a portfolio-management application. Fund issuance and fund-product
  administration are secondary bounded contexts, not the center of the user
  workflow.
- Portfolio is the primary business identity for investment operations.
- Users work with a human-readable `portfolioCode` in URLs and V2 endpoints.
  UUIDs remain internal database identities and must not be normal UI labels.
- Fund policy: `FUND_OPTIONAL` (owner decision D1, 2026-07-24 — "what I want is
  fundless"). This supersedes the earlier 2026-07-15 fund-required decision.
  Portfolios and their financial activity may exist with no fund association;
  `fund_id` is nullable across portfolio and trading tables. A fund-less
  portfolio's own id is its data-permission/workflow/compliance scope key.
  Hard constraint: fund-less financial activity must NEVER skip EOD close,
  compliance evaluation, approval, valuation, reporting, permission, or audit
  controls merely because no fund id exists — "no fund" means "use the
  portfolio as scope", not "skip the gate". All layers
  (workflow/compliance/permissions/reporting/migrations/tests/UI) must express
  this single policy; do not reintroduce a fund-required assumption anywhere.
- Portfolio types are `LIVE`, `SIMULATION`, and `MODEL`:
  - `LIVE`: real managed portfolio, official ledger/valuation, approvals, and
    compliance apply.
  - `SIMULATION`: paper or what-if portfolio; results are non-official.
  - `MODEL`: target allocation template; no real ledger or execution.
- New V2 request bodies must not accept or trust client-supplied `portfolio_id`,
  `fund_id`, or `contract_id` when portfolio code can resolve the context.

## Identity Decisions

- `contract_id` was removed from the investment operational tables
  `investment__decisions`, `investment__executions`, and
  `investment__trade_confirmations` in commit `91dab64`.
- Do not interpret every remaining `contract_id` as an investment-table defect.
  Approval, workflow, compliance, and research may use contract identity as a
  generic cross-module scope.
- In `contract.ProposedOrderCheck`, `ContractID` is currently legacy naming for
  optional fund scope. It is populated from the decision's `FundID` when one
  exists. Renaming that cross-module field is a separate migration and must not
  be mixed into Portfolio Compliance V2.
- Operational database links use `portfolio_id`. While legacy operational
  `fund_id` columns remain, composite foreign keys must prevent a row from
  claiming a fund different from its portfolio's fund.
- Public V2 routes resolve portfolio code inside the investment module. The
  compliance module must not independently resolve portfolio codes.

## Compliance Business Flow

The intended lifecycle is:

```text
Create rule and parameters
  -> bind active rule to portfolio
  -> create and submit investment decision
  -> manager approves decision
  -> execution creation re-runs pre-trade compliance
  -> PASS/WARN permits execution; BLOCK prevents execution
```

Important details:

- The existing submit-time compliance check remains. Execution-time checking is
  a second gate using the actual ordered quantity and amount.
- A `BLOCK` verdict is a business result, not an infrastructure failure. The
  investment command translates it to `domain.ErrComplianceRejected` and must
  create no execution row.
- `WARN` currently does not block execution.
- Compliance check records and breaches are an audit trail. Do not invent a
  sentinel fund/contract UUID for a portfolio-only check.
- Supported Portfolio Compliance V2 rule families currently include asset-class
  minimum/maximum allocation, maximum order percent of AUM, minimum NAV, and the
  existing minimum cash-buffer rule.

## Regulatory Program Decisions

- On 2026-07-15 the owner made Thailand, under Thai SEC rules, the first
  regulatory jurisdiction and moved Real Portfolio Regulation Rules ahead of
  Architecture Cleanup.
- Authoritative regulatory behavior must be traceable to the Thai SEC rulebook,
  the effective consolidated notification and appendices for the applicable
  product regime, and any approved portfolio/fund mandate. Summaries, demo
  seeds, and hard-coded historical percentages are not legal authority.
- Do not replace `regulatory.thai_sec` with enforcing thresholds until the owner
  or compliance expert approves the applicable regime (retail mutual fund, AI,
  UI, private fund, or provident fund) and a source-to-rule matrix with effective
  dates.
- Architecture Cleanup follows the regulatory program. Its legacy
  `ContractID` rename may clarify optional compliance/audit scope, but it must
  not be used to reintroduce fund-less portfolio development.

## DDD Boundaries

- `investment` owns portfolios, portfolio-code resolution, decisions,
  executions, transactions, holdings, cash, and valuation orchestration.
- `compliance` owns rule definitions, bindings, evaluation, check records, and
  breaches.
- Cross-module calls go through interfaces in `backend/pkg/contract`. A module
  must not import another module's `internal` packages.
- `approval` is a generic approval engine. It should operate on subject type and
  subject identity rather than becoming coupled to portfolio/fund database
  models.
- Backend permissions and data-scope checks are authoritative. Frontend guards
  are usability controls only.
- Nuxt route files should remain thin; portfolio workspace behavior belongs in
  `frontend/app/features/portfolio-workspace`.

## Financial and Safety Rules

- Preserve an immutable, attributable audit trail for financial decisions,
  compliance checks, overrides, executions, and state changes.
- Company-level AUM and P&L use a configuration-controlled reporting currency;
  the current configured value is `THB`. Application/domain aggregation and
  frontend formatting must not hard-code that value. Production startup must
  reject missing or invalid reporting-currency configuration rather than
  silently choosing a currency.
- Entire-company AUM and P&L are visible to every authenticated dashboard user,
  independent of function and fund/portfolio data-scope permissions. This
  policy grants only the company aggregate and aggregate coverage: company
  responses must not expose item-level fund or portfolio exclusion identities,
  and all drill-down screens remain data-scope protected. The `mine` scope
  includes only portfolios whose `manager_user_id` is the authenticated user,
  still bounded by the user's accessible funds.
- Company and mine AUM/P&L use each in-scope portfolio's latest available
  valuation and the latest valid authoritative executable-code FX quote into
  the configured reporting currency. A valuation older than the aggregate's
  newest valuation date remains included, with the count and oldest included
  valuation date disclosed in aggregate coverage. Never fabricate an FX rate,
  assume 1:1 across currencies, omit a currency silently, or add unlike
  currencies. A missing valuation or missing/invalid FX quote still makes the
  aggregate unavailable/incomplete with explicit coverage evidence.
- A `LIVE` portfolio with no active and effective compliance rule binding on
  the business date is not compliant. Return a typed not-configured result and
  block decision submission and execution creation without a false PASS record.
- Missing required asset classification for an existing holding or proposed
  instrument is a typed compliance-unavailable result for a `LIVE` portfolio;
  asset-allocation rules must not silently skip it. Do not change the policy for
  `SIMULATION` or `MODEL` portfolios without a separate approved decision.
- Fail closed when a mandatory production compliance or workflow dependency is
  unavailable. Nil dependencies may be tolerated only where existing unit-test
  construction explicitly relies on them.
- Bad client input returns a 4xx response. Infrastructure and unexpected errors
  return 5xx. A compliance verdict is returned or translated as a domain result.
- Never expose raw UUIDs, `fund_id`, or `contract_id` as user-facing names.
- Never put production customer data, account numbers, credentials, tokens,
  private keys, or secrets into an external AI prompt.
- Use sanitized or synthetic financial data in tests and AI sessions.
- Production migration/bootstrap paths must never assign privileges or data
  scope to named demo identities. Demo users, memberships, grants, and datasets
  must be behind an enforced development/test-only seed boundary; a directory
  name or SQL comment is not an environment control. When an unsafe historical
  migration may already have run, a new forward corrective migration must
  remove only the exact migration-owned demo grants while preserving legitimate
  production role catalogs and administrator-created assignments.

## Engineering Operating Rules

- Inspect `git status`, branch attachment, staged files, and relevant diffs
  before changing or committing anything.
- Preserve all pre-existing worktree changes. Do not reset, discard, or rewrite
  unrelated user work.
- Work on one explicit goal at a time. Keep fixes scoped to its acceptance
  criteria and record deferred work in `TASKS.md`.
- Generated Swagger and OpenAPI client files must be regenerated when their
  source API contract changes; do not hand-edit generated output.
- Before declaring backend/frontend work ready, run the relevant tests, build,
  formatting, and `git diff --check`. Record exact commands and results in
  `HANDOFF.md`.
- Do not commit or push unless the owner explicitly requests it. The current
  delivery preference is to prepare and review a commit locally, with no push.
- At the end of every manager session, update `HANDOFF.md` and `TASKS.md`.
  Change `MEMORY.md` only for a confirmed durable decision.

## Known Strategic Gaps

- Fund-less portfolios are SUPPORTED by explicit owner decision D1 (2026-07-24,
  `FUND_OPTIONAL`). The remaining work is not "should we do it" but ensuring
  every module (workflow/compliance/permissions/reporting/audit) applies the
  same controls to fund-less activity as to fund-bound activity.
- Thai SEC/BOT regulatory logic is not complete; `regulatory.thai_sec` is still
  a warning stub.
- Basket, rebalance, and switch flows still need per-line compliance checks.
- Full TypeScript checking has existing debt beyond the normal Nuxt build.
