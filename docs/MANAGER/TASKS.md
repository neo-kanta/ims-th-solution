---
type: manager-task-board
project: IMS Thailand
owner: Kanta
status: active
last_updated: 2026-07-17
---

# Manager Tasks

Only one task may be `IN PROGRESS`. A new AI instance must verify the repository
before changing a status. Completed work belongs in the completion log, not in
the active queue.

## P0 - COMPLETED LOCALLY - Three Production-Readiness Defects (IMS-PR3-20260717)

Goal: fix exactly three defects on `feature/investment` without merging or
pushing: configured reporting-currency AUM/P&L, LIVE compliance fail-closed
behavior for missing classification/rules, and complete EN/TH/ZH plus
branch-introduced TypeScript cleanup. This task does not make the branch
merge-ready; authorization, demo-permission migration, fill-validation, audit,
and other documented blockers remain outside scope.

Manager checklist:

- [x] Read manager state, repository instructions, relevant domain/frontend/API
      documentation, and verify live Git/stash state before edits.
- [x] Create three isolated worktrees from `ebef93b` with non-overlapping
      backend AUM, backend compliance, and frontend ownership.
- [x] Accept an executable business-date-aware FX source or document a precise
      blocker; never invent FX or omit currencies.
- [x] Integrate reporting-currency config, honest coverage semantics, and
      authoritative backend totals for `company` and `mine`.
- [x] Integrate typed LIVE compliance-not-configured and missing-classification
      failure behavior through decision/execution and real handler boundaries.
- [x] Establish `feature/investment` versus isolated `neo-develop` typecheck
      baselines; remove all branch/contract regressions.
- [x] Prove genuine EN/TH/ZH parity, interpolation, no fallback/raw keys, and no
      UUID business labels for the new workflows.
- [x] Regenerate Swagger and `frontend/app/api/ims-api.d.ts` from authoritative
      source when the accepted API contract changes; never hand-edit generated
      files.
- [x] Obtain independent backend financial/security and frontend/i18n reviews;
      resolve all P0/P1 findings and record lower-priority residuals.
- [x] Run focused plus full backend/frontend gates, E2E-tag compile, complete
      diff/whitespace/generated/secret checks, then create scoped local commits.
- [x] Update this task and `HANDOFF.md` with exact evidence and safety state.

Worker tracking:

- `IMS-PR3-AUM` — completed and integrated as `216c600`; independent financial
  review has no remaining P0/P1.
- `IMS-PR3-COMPLIANCE` — completed and integrated as `e80f086`; independent
  backend/security review has no remaining P0/P1.
- `IMS-PR3-FRONTEND` — completed and integrated as `350fe19`; all independent
  P2 findings were fixed and narrowly re-reviewed. Generated contracts are
  `790a2f7`.

Acceptance criteria are the observable requirements in the owner's 2026-07-17
brief: honest reporting-currency conversion/coverage; LIVE fail-closed behavior
with no financial persistence; exact EN/TH/ZH proof; zero branch-introduced or
generated-contract TypeScript diagnostics; full validation evidence; local-only
commits; unchanged `neo-develop`, remotes, live databases, and existing stashes.

Completion evidence and the exact 123-file implementation inventory (124 commit
entries because `backend/internal/investment/module.go` appears in two scoped
commits) are recorded in `HANDOFF.md`. This completion does not clear the
overall DO NOT MERGE verdict.
Residual work includes Portfolio V2 data-scope enforcement, production demo
membership cleanup, fill validation, binding audit, authenticated responsive
locale/theme UAT, 97 unrelated TypeScript diagnostics, a source-filtered prior
INTERNAL valuation query, and higher-precision transaction-time FX if required.

## P0 - GATED - Production Readiness and UAT

### Documentation-First Overnight Production-Readiness Program (IMS-UAT-20260717)

Goal: complete a documentation-first traceability and UAT assessment for the
portfolio transaction lifecycle, compliance controls, UI/UX, and production
engineering; then implement only the highest-priority safe gaps that can be
independently reviewed and fully verified before the 2026-07-17 06:25 Bangkok
hard stop. This task does not authorize a claim that IMS is production-ready.

Current evidence and safety state:

- Started 2026-07-16 23:40 Asia/Bangkok under the existing Goal.
- Original branch/HEAD verified as `feature/investment` at
  `0362d00bfad71feb16e624f133557447b9c37f52`.
- Permanent safety stash: stable commit
  `bd589febc6d0885796694112bbf6b8f94be163cc`, message
  `overnight-uat-safety-20260716T234123+0700`.
- Overnight review branch: `codex/overnight-uat-20260717-20260716T234125`.
  The current local cleanup is on `feature/investment`; verify live HEAD rather
  than relying on this historical snapshot.
- Safety stash was applied cleanly by hash on the overnight review branch; no
  conflicts; index cleared without
  changing working-tree content so any later staging can be explicit.
- External documentation rescan found 48 files. Documentation inspection is
  complete at the maximum safe level: 36 fully covered, 11 valid DOCX packages
  structurally covered but visual/page QA blocked, and one zero-byte invalid
  DOCX fully blocked. Application source work was prohibited and not started.
- Sanitized artifacts exist at `docs/uat/2026-07-17-document-coverage.md`,
  `docs/uat/2026-07-17-production-readiness-matrix.md`, and
  `docs/uat/2026-07-17-overnight-report.md`.
- Independent read-only review reconciled the 48-path manifest, 48-row matrix,
  manager state, overnight-only file scope, branch/HEAD/stash, redaction scan,
  and recovery procedure. It found no critical issue; its sole high finding
  was four unfinished final-report fields, which were resolved and confirmed
  closed by a narrow re-review with no remaining critical/high finding.
- On 2026-07-17, Kanta authorized a local-only cleanup on
  `feature/investment`. Existing code was separated into nine scoped commits
  (`47312fa` through `c8144fa`); no push, deployment, live migration, or stash
  deletion occurred. Backend tests/build and frontend 488-test suite/build
  passed. This checkpoint does not clear the documentation gate.
- The local edit that removes required `fund_id` validation from
  `portfolio_crud.go` is deliberately excluded because fund-less portfolios
  are an explicit owner non-goal. Bruno health URL edits remain uncommitted
  pending environment-intent confirmation.
- Independent backend, frontend, and integration reviews of the proposed
  `feature/investment` -> `neo-develop` fast-forward returned DO NOT MERGE.
  Source-proven P0/P1 blockers include Portfolio V2 data-scope bypass,
  production migrations granting demo identities, compliance fail-open/fill
  gaps, missing binding audit attribution, misleading cross-currency totals,
  false `Clear` compliance UX, stale portfolio-route races, UUID exposure, and
  incomplete EN/TH/ZH/typecheck coverage. `neo-develop` remains unchanged.

Merge-readiness remediation checklist (not started; requires owner approval):

- [ ] Enforce fund data scope on every Portfolio V2 decision, execution, and
      confirmation read/write route, fail closed, and prove cross-scope denial.
- [ ] Move `ben`/`green` memberships out of production migrations into
      environment-specific seeds or explicit provisioning.
- [ ] Fail closed for missing asset classification and missing mandatory LIVE
      portfolio rule configuration; validate/recheck actual fills.
- [ ] Add actor-attributed immutable audit events for compliance binding create
      and deactivate operations.
- [ ] Approve and implement a reporting-currency/FX or per-currency display
      policy; never present omitted/mixed currency subtotals as company AUM.
- [ ] Preserve compliance-unavailable state, prevent cross-portfolio response
      races, remove raw UUID presentation, and complete typed EN/TH/ZH copy.
- [ ] Restore legacy-route compatibility where required, correct Swagger/client
      drift, and add frontend plus Portfolio V2 lifecycle tests to CI.
- [ ] Rerun backend tests/build/vet, frontend tests/build/typecheck, migrations
      on a disposable database, and authenticated responsive/theme/locale UAT
      before reconsidering the fast-forward.

Checklist:

- [ ] Cover all 48 files under `C:\Users\kanta\Desktop\TH IMS` in a sanitized
      documentation manifest. All paths are inventoried and requirements are
      summarized, but the zero-byte DOCX and missing page renders for 11 valid
      DOCX files keep this unchecked and block the gate.
- [ ] Reconcile external requirements with executable source, tests, manager
      memory, and current domain documentation.
- [x] Create a production-readiness/UAT traceability matrix with exact evidence
      and PASS/FAIL/BLOCKED/NOT TESTED results.
- [ ] Select only high-severity, high-confidence, reversible gaps that fit the
      time and verification gates; do not invent Thai SEC policy.
- [ ] Delegate any backend/frontend implementation to one bounded writer at a
      time with exclusive ownership and the appropriate frontier skill.
- [ ] Obtain at least one independent read-only review of every material source
      change and resolve all critical/high findings before a local commit.
- [ ] Run focused and proportional broad checks, inspect complete diffs, scan
      for secrets/noise, and stage only explicit files.
- [x] Update manager handoff and produce the sanitized overnight report with
      exact Git recovery instructions and explicit confirmation that nothing
      was pushed, deployed, published, or sent.

Acceptance criteria:

- Documentation gate is complete before any application source edit.
- Every user-facing and production-readiness claim is tied to executable or
  inspected evidence; unsupported claims remain FAIL/BLOCKED/NOT TESTED.
- Portfolio lifecycle, compliance failure behavior, permissions/data scope,
  auditability, UI identity/accessibility/localization, and engineering controls
  are represented in the matrix.
- Thai SEC enforcement remains human-gated until product regime and the exact
  source-to-rule matrix are approved.
- No unresolved critical/high finding exists in any locally committed scope.
- Nothing is pushed, deployed, published, or sent.

## P0 - Completed (pending Kanta's review)

### Compliance Dashboard UX and Typed API Integration (IMS-COMPLIANCE-DASHBOARD-UX)

Goal: replace the compliance landing page with an operational control-center
layout and wire every dashboard read through the generated OpenAPI client.
Implemented on `feature/investment` at HEAD `0362d00` in the existing dirty
worktree; not staged, committed, or pushed. Full detail is in `HANDOFF.md`'s
"Session: Compliance Dashboard UX and Typed API Integration (2026-07-16)".

Checklist:

- [x] Replace the generic card grid with a responsive control center: honest
      definition/breach KPIs, open-breach work queue, portfolio finder,
      category distribution, refresh action, permission-aware actions, and
      explicit loading/error/empty states.
- [x] Use `useOpenApiClient` plus generated `ims-api.d.ts` contracts for
      `GET /compliance/rules`, `GET /compliance/breaches`, and
      `GET /investment/portfolios`; validate/normalize the generated DTOs at
      the feature boundary.
- [x] Follow every rules/portfolio page so summary counts are exact rather
      than silently capped at the endpoint's 200-record page limit.
- [x] Keep breach claims honest: the V1 breach endpoint is a role-visible
      compliance-admin queue, not a portfolio-data-scoped enterprise metric.
- [x] Preserve EN/TH/ZH copy parity and keyboard/mobile behavior.
- [x] Add pagination, API-normalization, date-boundary, finder, queue, and
      formatter regression coverage.
- [x] Validate focused compliance Vitest (12 files / 87 tests), full frontend
      Vitest (45 files / 488 tests), and production `npm run build`.

Acceptance criteria:

- Dashboard values remain unavailable while loading/on error; no fake zeroes.
- "Active" and "scheduled" mean configured definitions evaluated against the
  explicitly labelled device-local date, not proven applied controls.
- Portfolio search is hidden with an explanatory note without portfolio-view
  permission; rule creation is permission-gated.
- Existing backend endpoints are reused. A future applied-control/data-scoped
  analytics surface requires a dedicated backend read model and is not
  fabricated client-side.
- No unrelated worktree changes are modified; no commit or push occurs.

Known verification limit:

- Protected-route browser proof is blocked by the expired local app session;
  Chrome connector state was unavailable. Automated tests and production
  build are green, but authenticated desktop/mobile visual proof remains for
  Kanta's review session.

### Dashboard AUM Today / Today's P&L Live Integration (IMS-DASHBOARD-AUM-PNL)

Goal: finish and verify the live API integration for the "AUM Today" and
"Today's P&L" dashboard cards. A partial, uncommitted implementation already
existed (endpoint, adapter, handler, frontend composable/service/lib) from an
earlier session; this task audited the backend contract, fixed two
correctness bugs found in the audit, added missing test coverage, and
completed the frontend (Bangkok as-of time, race-safe scope switching,
removed the `metric as any` cast, fixed a Vitest-blocking barrel import).
Full detail in `HANDOFF.md`'s "Session: Dashboard AUM/P&L Live Integration
(2026-07-15)". Not committed — Kanta reviews before deciding.

Checklist:

- [x] Audit `ValuationSummaryAdapter` against the 16-point backend contract
      (auth, scope validation, no client-supplied identity, data-scope
      enforcement, decimal-as-string, `data_available` semantics, error
      classification).
- [x] Fix: `previousInternalSnapshot` swallowed valuation-repository errors
      as "no previous snapshot", which could attribute a portfolio's entire
      cumulative P&L to today whenever the lookup merely failed. Now
      propagates the error.
- [x] Fix: portfolios valued on different business dates were summed under
      one reported `business_date`. Now only portfolios whose latest
      snapshot matches the scope's max business date are aggregated; older
      ("stale") portfolios are excluded from that day's total.
- [x] Confirmed (not changed): minority-currency funds are excluded from
      "Entire company AUM" rather than converted — no reporting-currency/FX
      policy exists anywhere in the codebase (`docs/investment-module.md`
      lists multi-currency NAV as an explicit Phase 2 gap). This is the only
      safe behavior absent an owner-approved FX policy; flagged for Kanta,
      not silently shipped.
- [x] Added adapter tests: previous-snapshot error propagation, stale
      (inconsistent-date) portfolio exclusion, negative P&L, zero P&L.
- [x] Added `dashboard_handler_test.go` (did not exist): auth-required,
      invalid-scope 400, success envelope, provider-failure 500.
- [x] Verified Swagger/`ims-api.d.ts` already match the (unchanged) handler
      contract — no regeneration needed.
- [x] Fixed the Vitest-blocking import: `dashboard.ts` imported pure
      formatting helpers through the `my-funds` barrel, which pulls in a
      Nuxt-only `~/api/openapi` import via `myFundsApi.ts`. Now imports
      directly from `my-funds/lib/format.ts`.
- [x] Found and fixed a second instance of the same class of bug: the test
      file imported `normalizeValuationSummary` straight from
      `dashboardApi.ts`, whose top-level `~/api/openapi` import fails to
      resolve in plain Vitest regardless of the barrel fix. Extracted the
      pure mapper into `dashboard/lib/valuationSummaryMapping.ts`.
- [x] `as_of` now renders in Asia/Bangkok for the AUM/P&L cards specifically
      (`formatDashboardTime` gained an optional `timeZone` param, default
      `"UTC"` preserved for the unrelated header "last updated (UTC)"
      label, which was deliberately left untouched).
- [x] Scope-switch race safety: `useDashboardValuationSummary` now tracks a
      request id so a slow, superseded request cannot overwrite a newer
      scope's result.
- [x] Removed `:metric="metric as any"` in `DashboardOverviewScreen.vue` by
      typing `displayMetrics` as `(DashboardOverviewMetric & { loading:
      boolean })[]` instead.
- [x] Full validation run (see HANDOFF for exact commands/results): backend
      `go build`/`go test ./...`/`gofmt -l` clean; frontend targeted Vitest,
      full `npm run test` (391 tests), `npm run build`, and `npx nuxi
      typecheck` all clean of new diagnostics in touched files.

Acceptance criteria:

- AUM Today / Today's P&L display real API values, loading/no-data/error
  states are distinct, and negative/positive/zero P&L render correctly —
  verified by unit/handler/adapter tests. **Live browser proof not run —
  see "Remaining blocker" in HANDOFF.**
- Backend permissions and data scope remain authoritative; no financial
  aggregation happens in Vue/TypeScript.
- Generated OpenAPI types used end to end; no manual API types.
- No unrelated worktree changes touched.

Non-goals (explicitly deferred, not invented):

- No FX/reporting-currency conversion policy — flagged as an owner decision.
- No live three-locale browser proof (blocked; see HANDOFF).

## P0 - Completed

### Correct the Compliance Frontend (Catalog, i18n, Parameters, Effective State)

Goal: make the rule catalog accurate, type-safe, multilingual (en-US/th-TH/
zh-TW), and consistent with the backend rule registry; ensure Portfolio
Compliance V2 never sends internal identity fields; stop the bound-rules
table from overclaiming enforcement. Frontend-only; committed locally as
`2593c24` on `feature/investment` on top of `e8b2a6d`; not pushed. Full
detail in `HANDOFF.md`'s "Session: Compliance Frontend Correction
(2026-07-13)".

Checklist:

- [x] Verify Portfolio Compliance V2 bind payload carries no `fund_id`,
      `FundID`, `contract_id`, `ContractID`, `portfolio_id`, `scope_id`, or
      `scope_type` — confirmed unchanged-but-correct; added a regression test.
- [x] Rebuild `RULE_CATALOG` around `labelKey`/`explanationKey`/
      `suggestedCorrectionKey`, resolved through the existing `t()`.
- [x] Add `compliance.catalog.*` to `en`/`th`/`zh` `compliance.ts` for all
      16 backend-registered rule types.
- [x] Extend coverage from 9/16 to 16/16 backend rule types; fix category
      drift on `credit.min_rating`, `credit_rating.minimum`,
      `ratio.sector_exposure` (all backend `MANDATE`, not
      `RESTRICTION`/`RATIO`).
- [x] Mark both backend stubs (`credit_rating.minimum`, `regulatory.thai_sec`)
      `selectable: false` and disable them in the builder `<select>`.
- [x] Fix `SAMPLE_PARAMS` drift (`min_cash_buffer_pct`, `sector`,
      `default_lot_size`, `enforced_types`) and add samples for the six
      newly-covered production rule types; no sample for either stub.
- [x] Replace `PortfolioComplianceView.vue`'s `is_active`-only "currently
      enforced" claim with a Scheduled/Effective/Expired/Deactivated
      derivation against an explicitly-labelled local "as of" date.
- [x] Add a safe `unknown`-accepting boundary normalizer for
      `entry.parameters` instead of casting the generated
      `Record<string, never>` type at the render call site.
- [x] Add an explicit duplicate-submission guard to `submitBind`/`deactivate`.
- [x] Add/extend tests for all of the above; full suite passes (23 files,
      325 tests); `nuxi typecheck` output is byte-identical before/after.
- [x] Stage the explicit reviewed frontend file list; preserve all unrelated
      dirty/untracked files.

Acceptance criteria:

- Rule catalog covers every backend-registered `rule_type_id` with the
  correct category.
- Catalog label/explanation/suggestedCorrection resolve in en-US, th-TH,
  and zh-TW through `t()`; unknown types fall back to the raw id and
  backend message.
- Stub rule types are visibly marked and not selectable as production rules.
- Rule Builder parameter samples match each rule's real
  `spi.ParameterSchema()`.
- Portfolio Compliance V2 payloads exclude all internal identity/scope
  fields (verified by test, not just inspection).
- The bound-rules table never implies a scheduled or expired binding is
  "currently enforced."
- No new `any`, suppression comment, or unsafe cast introduced.
- `npm run test`, `npm run build` pass; `npx nuxi typecheck` introduces zero
  new diagnostics versus the pre-edit baseline.

Non-goals (unchanged from the compliance V2 P0 task, still respected):

- No backend rule-catalog endpoint (recorded as a P2 architecture item,
  already present below).
- No change to `PortfolioRuleCatalogEntry.parameters`'s generated
  `Record<string, never>` type or the `contracts.go` Swagger annotation
  that causes it (P2 item below, updated to reflect the frontend-side
  mitigation now in place).
- No redesign of `PortfolioComplianceView.vue` beyond the restrained
  per-row state indicator.

### Prepare and Commit Portfolio Compliance V2 Locally

Goal: turn the current uncommitted Portfolio Compliance V2 implementation into a
reviewed, reproducible, single local commit. Do not push.

Verified state on 2026-07-13: review, scoped fixes, explicit staging, and all
required build/test checks completed before the 47-file feature was committed
locally as `e8b2a6d`. The index is empty and no push was performed.

Checklist:

- [x] Attach detached HEAD to `feature/investment` without losing work.
- [x] Review the execution-time compliance gate and its tests.
- [x] Review module wiring and prove production receives a non-nil checker.
- [x] Review both compliance migration pairs, including rollback semantics.
- [x] Review portfolio-code access control and HTTP error classification.
- [x] Review the portfolio compliance UI for business clarity and hidden UUIDs.
- [x] Assess and fix the three `bindDrafts[...]` typecheck errors in
      the new compliance view without expanding into branch-wide TS cleanup.
- [x] Stage only the feature's backend, frontend, generated API, migrations, and
      required Swagger-generation fix. Exclude Bruno health changes and
      unrelated docs/refactors.
- [x] Inspect the complete staged diff for accidental scope or secrets.
- [x] Run required validation and record exact results in `HANDOFF.md`.
- [x] Commit locally with a focused message (`e8b2a6d`).
- [x] Confirm the branch remains unpushed.

Acceptance criteria:

- An approved decision is checked immediately before execution creation.
- A `BLOCK` verdict creates no execution row and returns the typed rejection.
- `PASS` and `WARN` follow the agreed behavior and are tested.
- Invalid V2 pre-trade input returns HTTP 400; internal failures remain 500.
- Portfolio rule binding and deactivation work through portfolio-code routes.
- No normal compliance UI label exposes UUID, `fund_id`, or `contract_id`.
- Backend build/tests, frontend build/tests, formatting, and diff checks pass, or
  unrelated failures are documented with evidence.
- The staged diff excludes unrelated files.
- One local commit exists on `feature/investment`; nothing is pushed.

Non-goals for this task:

- Renaming `contract.ProposedOrderCheck.ContractID`.
- Removing the existing submit-time compliance gate.
- Making `investment__portfolios.fund_id` nullable.
- Implementing Thai SEC/BOT regulations.
- Fixing all historical frontend typecheck debt.
- Redesigning the compliance UI.

## P1 - Completed

### Prove Portfolio Compliance V2 End to End

Goal: demonstrate the manager-approval-to-execution compliance lifecycle
against a disposable `ims_e2e` PostgreSQL database (never `ims_dev`), backed
by a new Go E2E test. Full detail in `HANDOFF.md`'s "Session: P1 Portfolio
Compliance V2 End-to-End Proof (2026-07-14)".

- [x] Apply migrations to a disposable or dedicated test database — `ims_e2e`
      on the shared local Postgres container, created and migrated fresh via
      `make e2e-db-setup`'s constituent commands.
- [x] Seed or create a rule instance and portfolio binding — done through the
      real `POST /api/v1/compliance/rules` and V2
      `POST /api/v2/portfolios/{code}/compliance/rules/{id}/bindings` routes,
      not direct SQL.
- [x] Create a decision, complete manager approval, and attempt execution —
      a real two-stage approval (`ben` then `green`, the seeded
      `PROC_DECISION_DEFAULT` process) via the actual `/approvals/inbox` and
      `/approvals/tasks/{id}/approve` routes.
- [x] Prove `BLOCK` prevents execution in the API. UI proof not run — see
      "Remaining blocker" in `HANDOFF.md`.
- [x] Prove `PASS` permits execution and produces an auditable check record.
- [x] Prove `WARN` does not block execution (severity-cap policy).
- [x] Prove deactivation changes which rule applies.
- [x] Prove a future `effective_from` binding does not apply before its date.
- [x] Verify portfolio-code compliance routes are locale-invariant (pure
      JSON, no `Accept-Language` handling) and that the existing frontend
      `compliance.catalog.*` i18n resolves in en/th/zh. Live three-locale
      browser proof not run — see "Remaining blocker" in `HANDOFF.md`.
- [x] Add an E2E test for the manager-approval-to-execution gate —
      `backend/tests/e2e/portfolio_v2_compliance_lifecycle_test.go`
      (`TestE2E_PortfolioComplianceV2_ManagerApprovalToExecution`).

**Bug found and fixed in the same session (in scope — directly on the tested
path):** `execution_handler.go`'s `writeExecutionError` had no case for
`*domain.ErrComplianceRejected`, so an execution-time `BLOCK` fell through to
`httputil.InternalError` (500) instead of the typed 422 rejection the
submit-time path (`writeDecisionError`) already returns. Fixed to mirror
`writeDecisionError`'s handling exactly. Flagged for owner review in
`HANDOFF.md`'s "Highest-Risk Human Review" — it is a production API
contract change (500 → 422), not test scaffolding.

**Found, not fixed (pre-existing, unrelated, out of scope):**
`TestE2E_PortfolioV2_DecisionOwnership` fails against a fresh `ims_e2e`
database — `column "contract_id" does not exist` — because it still queries
`investment__decisions.contract_id`, a column removed by commit `91dab64`
(before this session). `TestE2E_InvestmentOversellEnvelope` still fails with
the `seedOpeningPosition` type-inference bug already documented in
`tests/e2e/COVERAGE.md`. Neither touches Portfolio Compliance V2.

Definition of done: the intended lifecycle is demonstrated against PostgreSQL,
not only unit-test fakes, and the evidence is recorded in `HANDOFF.md`. Met
for every backend/API item; UI and live-locale browser proof are documented
blockers, not silently claimed.

**Independent review (2026-07-15):** a read-only second-opinion review of
`execution_handler.go` and `portfolio_v2_compliance_lifecycle_test.go`
returned "approve with concerns" — the 422 fix and the E2E test are both
correct; two non-blocking findings were logged below under P2 - Architecture
Cleanup. Full detail in `HANDOFF.md`'s "Session: Independent Review of P1
Execution Handler Fix and E2E Test (2026-07-15)".

## Owner Decision - Not Planned

### Fund-Optional Portfolio Development

Closed by Kanta on 2026-07-15: do not develop the optional-fund feature. Keep
the current required fund association. This is an explicit non-goal and its
former seven unchecked items are no longer part of the backlog.

## P2 - GATED - Regulation and Restriction

### Build Real Portfolio Regulation Rules (IMS-REG-TH-SEC)

Status: **GATED; NOT COMPLETE**. Thailand / Thai SEC is the confirmed first
jurisdiction. Official sources are mapped in
`docs/compliance/thai-sec-regulatory-source-map.md`. The remaining human gate
is selection of the first applicable product regime and approval of its exact
source-to-rule matrix; no legal threshold may be invented or enforced before
that approval. The overnight production-readiness/UAT program is the sole
active task; it must report this regulatory gap honestly and must not close it.

- [ ] Approve the first Thai SEC product regime and exact source-to-rule matrix
      with the owner/compliance expert. Jurisdiction and official source family
      are confirmed; retail MF, AI, UI, private fund, or PVD applicability is
      not yet recorded in the domain.
- [ ] Replace the `regulatory.thai_sec` warning stub with versioned rules and
      effective dates.
- [ ] Add issuer, sector, asset-class, liquidity, concentration, restricted-list,
      NAV/AUM, and cash controls only from approved requirements.
- [ ] Add rule parameter validation, evidence payloads, severity policy, and
      override approval requirements.
- [ ] Add per-line checks for basket, rebalance, and switch operations.
- [ ] Add post-trade monitoring and breach resolution behavior.
- [ ] Prove every rule with boundary tests and auditable source/version metadata.

Definition of done: implemented rules are traceable to an approved regulation
or mandate, reproducible by business date, and test both pass and breach edges.

## P2 - QUEUED AFTER REGULATIONS - Architecture Cleanup

- [ ] Plan a separate rename from legacy compliance `ContractID` to explicit
      optional fund scope without breaking approval/workflow contracts. This
      cleanup must not make portfolio fund association optional.
- [ ] Add a live rule-type catalog endpoint or formalize generation of the
      frontend catalog from the backend registry.
- [ ] Resolve duplicated exposure math and N+1 rule-version loading.
- [ ] Address branch-wide TypeScript typecheck debt as a dedicated quality task.
- [ ] Review stale module/runbook claims after Portfolio Compliance V2 commits.
- [x] Add a per-row effective-state indicator (scheduled/effective/expired)
      to `PortfolioComplianceView.vue`'s bound-rules table so `is_active=true`
      is not read as "currently enforced" when `effective_from`/`effective_to`
      place the binding outside today's reference date. Found 2026-07-13;
      fixed 2026-07-13 in the frontend-correction session (uses the
      browser's local date, explicitly labelled "as of", since the
      workspace exposes no authoritative backend business date).
- [ ] Fix the `PortfolioRuleCatalogEntry.parameters` generated-type drift:
      `backend/pkg/contract/contracts.go`'s `swaggertype:"object"` annotation
      produces `Record<string, never>` in `ims-api.d.ts`. Found 2026-07-13.
      2026-07-13: frontend now routes through a safe `unknown`-accepting
      `asParameterRecord` boundary normalizer in `PortfolioComplianceView.vue`
      instead of casting at the render call site, but the root cause (the
      Swagger annotation) is still backend-only and unfixed — this remains
      open until `contracts.go` is corrected and the client regenerated.
- [ ] `zh/portfolio.ts` uses Simplified Chinese for its pre-existing keys,
      inconsistent with `zh/compliance.ts`'s Traditional (zh-TW) convention
      used everywhere else. Found 2026-07-13 while adding new zh-TW content
      to `portfolio.ts`; the new keys added that session use Traditional
      characters, but the surrounding pre-existing keys were left as-is
      (out of scope). Recommend a dedicated zh-TW normalization pass over
      `portfolio.ts`.
- [ ] `tests/e2e/investment_test.go`'s shared `cleanupInvestmentPrereqs`
      helper (lines 394-415) never deletes `investment__decisions` or
      `investment__executions`, both `ON DELETE RESTRICT` on
      `portfolio_id`/`decision_id`. `portfolio_v2_compliance_lifecycle_test.go`
      is the first caller that creates a decision + executions against the
      seeded portfolio, so the helper's portfolio/fund deletes silently fail
      (errors discarded via `_, _ = pool.Exec(...)`), orphaning a
      fund/portfolio/decision/3-executions/compliance-records set in
      `ims_e2e` on every run. Doesn't affect correctness or repeatability
      (fresh UUIDs per run), but contradicts the "tears down afterward"
      claim and will bloat a long-lived `ims_e2e`. Found 2026-07-15 during
      independent review; recommend the helper also delete
      `investment__executions`/`investment__decisions` (and the test add
      cleanup for its own compliance rule instance/bindings) before this
      test runs routinely in CI.
- [ ] `CreateExecution` (`execution_handler.go:112`) and
      `CreateExecutionByCode` (`portfolio_v2_execution_handler.go:56-72`)
      Swagger annotations still list only `@Failure 400/401/404/409/500` —
      no `@Failure 422` — even though the 2026-07-14 session's fix means
      both can now return 422 on a compliance BLOCK.
      `SubmitDecision` already documents 422 for the equivalent submit-time
      path. Found 2026-07-15 during independent review; no current frontend
      or Bruno caller is affected today, but `make swagger && make
      api-client` should be run to close the drift per `CLAUDE.md`'s
      generated-file rule.

## Completion Log

- (uncommitted, live local dev) - `IMS-OP01-PRICE`: fixed OP-01 decision
  submission so quantity + amount derives compliance price as
  `amount / quantity`; added fail-closed invalid-input handling and backend/
  frontend regression coverage; full backend tests, 478 frontend tests, and
  frontend production build passed. Rebuilt/restarted only local backend and
  frontend containers; health verified. See `HANDOFF.md` session dated
  2026-07-16.
- (uncommitted) - P1 Portfolio Compliance V2 end-to-end proof against a
  disposable `ims_e2e` database: new Go E2E test
  (`backend/tests/e2e/portfolio_v2_compliance_lifecycle_test.go`) proving the
  full manager-approval-to-execution lifecycle, plus a bug fix
  (`writeExecutionError` now returns 422 for a compliance `BLOCK` instead of
  500). Not committed — awaiting Kanta's review per `HANDOFF.md`.
- `e8b2a6d` - Enforced Portfolio Compliance V2 before execution, including
  portfolio-code APIs, rule binding, migrations, generated API artifacts, and
  the portfolio compliance UI. Committed locally; not pushed.
- `91dab64` - Removed duplicate `contract_id` from investment decisions,
  executions, and trade confirmations.
- `3156b8b` - Refreshed investment table snapshots for Portfolio V2.
- `ad8a423` - Enhanced the portfolio directory workspace.
- Portfolio-code V2 investment routes and `LIVE`/`SIMULATION`/`MODEL` portfolio
  types exist before the current uncommitted compliance work.
