---
type: manager-task-board
project: IMS Thailand
owner: Kanta
status: active
last_updated: 2026-07-24
---

# Manager Tasks

Only one task may be `IN PROGRESS`. A new AI instance must verify the repository
before changing a status. Completed work belongs in the completion log, not in
the active queue.

## P0 - IN PROGRESS - Merge-Blocker Remediation Program (IMS-MERGE-BLOCKERS)

Goal: close the source-proven `neo-develop` -> merge-readiness blockers and
restore one authoritative end-to-end investment lifecycle. The original order
P0-A (Portfolio V2 data-scope security) -> P0-B (demo migration cleanup) ->
P0-C (fill validation) -> P0-D (compliance binding auditability) remains
preserved. The 2026-07-24 end-to-end product/business-flow review expanded this
same program with P0-E through P0-I below; it did not create a second active
task. Started 2026-07-21. This supersedes the Breaches-UAT task below as the
single active manager task; Breaches-UAT is not abandoned, it is queued (see
its own section) because only one task may be IN PROGRESS.

### Claude `/goal` execution contract (2026-07-24)

Kanta requested a durable, token-aware Claude Code `/goal` package to execute
this complete remediation program across Claude accounts without relying on
transcript memory:

- `docs/MANAGER/CLAUDE-GOAL-IMS-REMEDIATION.md` is the ordered outcome
  contract, work-package backlog, acceptance matrix, validation gate, and
  copy-paste `/goal` command.
- `docs/MANAGER/CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md` is the mandatory
  per-account continuation record. It must be updated before context/account
  exhaustion and consumed by the next Claude account.

The runbook consolidates P0-B through P0-I plus the latest independent review
findings: denial-versus-infrastructure error classification, LIVE cash request
idempotency, durable approval-sync replay, actual approver attribution,
cash-request cancellation UX, SSR session restore, watchlist scheduling,
typecheck/CI debt, regulatory human gates, integrated database/browser proof,
and documentation reconciliation.

This documentation authorizes Claude to investigate and implement the
owner-approved remediation scope through `/goal`; it does **not** itself
authorize a commit, push, deployment, production migration/data mutation, or
an unrecorded business/legal decision. D1 is resolved as `FUND_OPTIONAL`, D2 is
resolved as `MATCHED CONFIRMATION`, and D3 is resolved as `EXPIRE PENDING`.
D4 remains human-gated/deferred: the Thai SEC control must stay honestly
`NOT_CONFIGURED` until a compliance/legal owner approves the product regime
and effective-dated source-to-rule matrix.

Verified live Git state at start (2026-07-21): branch `neo-develop`, HEAD
`0ae1adb`, tracking `origin/neo-develop`, `+5 -0` (five unpushed local
commits, matching `8d43747`..`0ae1adb`). Working tree: exactly 15 paths — 5
tracked-modified (`docs/api/README.md`, `docs/api/compliance-api.md`,
`docs/api/portfolio-v2-api-ddd.md`, `docs/api/watchlist-api.md`,
`tools/bruno/environments/ims-th-local.yml`) and 10 staged-new (`docs/api/`
generated-doc additions plus two Bruno execution request files under
`tools/bruno/Investment/Executions/`). No untracked files. Nothing pushed.

**Phase 1 disposition (2026-07-21, read-only reconciliation — zero commits
created):**

- All 15 working-tree paths are held, not staged into any new commit. This is
  a correct Phase 1 outcome, not an omission: `docs/api/*` provenance is an
  external concurrent writer (flagged in the 2026-07-20 Architecture Cleanup
  session, still unresolved) and the two Bruno execution `.yml` files and the
  local-environment edit were likewise never authored by a manager session.
- `tools/bruno/environments/ims-th-local.yml`'s only diff is two local
  session values (`compliance-breach-id` and `compliance-rule-type-id`
  populated from a local test run) — machine/session-local state per the
  owner's explicit instruction not to commit that file when it carries local
  runtime values. Held.
- `docs/api/_build_current_api_docs.py` is a clean, self-contained generator
  (reads `backend/docs/swagger.json` + `v2/v2_swagger.json` + the legacy
  builder, writes only `docs/api/*.md`) — plausibly reproducible, but
  authorship/ownership is still unconfirmed (external concurrent writer, per
  HANDOFF), so it is held pending Kanta's confirmation rather than executed
  or committed by this session.
- The generated `docs/api/*.md` diffs read as coherent, accurate
  documentation-quality improvements (e.g. correcting `portfolio-v2-api-ddd.md`
  and `watchlist-api.md` from "design/proposal" to "implemented, see current
  doc"), not a correctness or security concern — but ownership, not quality,
  is the reason they are held.
- The two new `tools/bruno/Investment/Executions/*.yml` files are untested,
  unreviewed Bruno request definitions of unconfirmed origin; held as a
  separate scope per the owner's instruction, not merged into any commit.
- The five existing local commits (`8d43747`, `39ab72d`, `09d9421`, `f6c5205`,
  `0ae1adb`) were reviewed read-only via `git show --stat`: each has a
  coherent, single-purpose file set matching its message (compliance DTO
  refactor; investment AUM+FX with matching frontend/i18n/tests/Bruno;
  generated-contract regeneration only; frontend breach-queue redesign; and
  manager-doc updates). No scope creep or cross-boundary file found. This
  reconfirms, without re-running the full gate, the extensive verification
  already recorded in HANDOFF's "Session: Local Investment Integration
  Commits" and predecessor sessions.

- [x] Verify live Git state independently (do not trust stale "52 modified
      files" claim) — confirmed exactly 15 paths, `+5 -0`, no untracked files.
- [x] Classify all 15 remaining working-tree paths — outcome: hold all 15,
      create zero new Phase 1 commits (see disposition above).
- [x] Review the five existing local commits for scope/correctness concerns —
      no material finding; each commit is single-purpose and evidenced.
- [ ] Kanta confirms provenance of the `docs/api/*` generator and generated
      docs, and of the two Bruno execution `.yml` files, before either is
      committed.

### P0-A - COMPLETE - Portfolio V2 data-scope security

**Committed as `67efe71`** on `neo-develop` (2026-07-21). Not pushed.
Implemented by `frontier-backend-engineer` in an isolated worktree, verified
independently by the manager (own `go build`/`go vet`/`go test ./... -count=1`
run: 81 packages `ok`, zero `FAIL`; `gofmt -l` clean; `git diff --check`
clean), and reviewed by a separate independent security reviewer given only
the raw diff and acceptance criteria (not the manager's or implementer's
conclusions) — **no blocking findings**. One non-blocking hardening
suggestion was logged (see below) and deliberately deferred, not silently
dropped.

**Manager process error, caught and corrected before reporting completion:**
the first `git commit` for this task used no pathspec and therefore committed
the *entire index* — sweeping in the pre-existing staged `docs/api/*` and
`tools/bruno/Investment/Executions/*` files that Phase 1 had explicitly
decided to hold. This was caught immediately (`git show --stat` on the new
commit showed 18 files, not 8), corrected with `git reset --soft HEAD~1`
(safe: the commit was unpushed, at the tip, and created by this session
seconds earlier — nothing else was built on top of it), and redone with an
explicit `--  <8 files>` pathspec. Final commit contains exactly the 8 files
in this task's write scope; `git status` afterward reconfirmed the other 15
paths are back to their exact pre-commit staged/unstaged state.

- [x] `DecisionHandler`, `ExecutionHandler`, `TradeConfirmationHandler` each
      gained a `pc contract.PermissionChecker` field + `SetPermissionChecker`
      setter; `module.go` wires `m.permissionAdapter` unconditionally to all
      three (verified: `permissionAdapter` fails closed even if `iamPort` is
      nil, per the adapter's own `HasDataPermission`/etc. guard clauses).
- [x] All 10 `resolvePortfolioByCode(..., nil)` call sites now pass `h.pc`;
      grep-confirmed zero remaining literal-`nil` invocations anywhere in
      `internal/investment`.
- [x] 20 new tests (2 per endpoint: authorized-passes-gate,
      cross-fund-403) — traced by the independent reviewer against actual
      source (not just comments) to confirm each assertion could only be
      reached after the fund-scope check passed.
- [x] Existing handler tests, full `internal/investment/...` suite, and the
      full repository `go test ./... -count=1` (81 packages) all still pass.
- [x] No V1 routes, permission-catalog entries, or generated Swagger/OpenAPI
      files touched; audit attribution unchanged.
- [x] Independent review completed with no blocking findings.

**Deferred hardening (logged, not required for this fix):** the shared
`resolvePortfolioByCode` still silently no-ops when `pc == nil` rather than
erroring — the "never nil in production" guarantee rests entirely on
`module.go` always calling the setter, the same class of bug (forgotten
wiring) that caused the original vulnerability. Both the implementer and
independent reviewer judged this an acceptable trade-off given the adapter's
own fail-closed behavior and the now-unconditional wiring, but flagged a
stricter fail-closed resolver variant (error/500 on nil for write-capable
handlers) as a good follow-up. Not scheduled; add to Architecture Cleanup if
Kanta wants it tracked.

### P0-A original finding (preserved for record; superseded by COMPLETE section above)

**Finding (confirmed in current code, 2026-07-21):** `resolvePortfolioByCode`
(`backend/internal/investment/transport/handler/portfolio_v2_handler.go:46-71`)
only enforces fund data-scope when its `pc contract.PermissionChecker`
parameter is non-nil (`if pc != nil && !hasFundAccess(...)`). The correct
call site is `portfolio_v2_handler.go:31` (`h.resolvePortfolioCode`), which
passes the handler's real `h.pc`. But **10 call sites across the three other
V2 handler structs pass a literal `nil`**, permanently disabling the check:

- `portfolio_v2_decision_handler.go:70,106,181,209,248` (`*DecisionHandler`)
- `portfolio_v2_execution_handler.go:75,146,214` (`*ExecutionHandler`,
  `CreateExecutionByCode`/`FillExecutionByCode`/`CancelExecutionByCode`)
- `portfolio_v2_execution_handler.go:261,339` (`*TradeConfirmationHandler`,
  `RecordConfirmationByCode`/`ResolveConfirmationByCode`)

Root cause confirmed by struct inspection: `DecisionHandler`
(`decision_handler.go:21-28`), `ExecutionHandler` (`execution_handler.go:16-21`),
and `TradeConfirmationHandler` have **no `pc`/permission-checker field at
all** — they are wired post-construction only via
`SetPortfolioRepository`/`SetDecisionRepository`/etc. in
`backend/internal/investment/module.go:219-229`. This is not a call-site typo;
the dependency was never plumbed through. The correct pattern already exists
twice in the same codebase to copy: `InvestmentHandler` receives its checker
(`m.permissionAdapter`) through its constructor
(`module.go:205-217`), and `m.decisionBatchCmd.SetPermissionChecker(iamPort)`
(`module.go:170`) is an existing post-construction setter for the same kind of
dependency on a different type.

**Acceptance criteria (all satisfied — see COMPLETE section above for how):**

- [x] `DecisionHandler`, `ExecutionHandler`, and `TradeConfirmationHandler`
      each gain a `pc contract.PermissionChecker` field and a
      `SetPermissionChecker` setter (mirroring `decisionBatchCmd`'s existing
      setter), or an equivalent constructor-injected fix — manager has no
      preference on setter-vs-constructor as long as production wiring in
      `module.go` cannot leave `pc` nil.
- [x] `module.go` wires the real checker (the same `m.permissionAdapter` used
      by `InvestmentHandler`, or `iamPort` directly — confirm which is
      correct against `NewPermissionCheckerAdapter`'s semantics) into all
      three handlers.
- [x] All 10 `resolvePortfolioByCode(w, r, h.portfolios, nil)` call sites
      become `resolvePortfolioByCode(w, r, h.portfolios, h.pc)`.
- [x] Fail closed: confirm (by test, not inspection) that a nil/unwired
      checker still denies rather than silently passing — i.e. do not
      "fix" this by leaving `pc != nil` as an opt-in bypass; a production
      handler must never be constructible with a nil checker in the wiring
      path, or the resolver must deny when the checker is nil for these
      write-capable handlers specifically (recheck whether the existing
      `pc != nil` skip in `resolvePortfolioByCode` itself needs to become
      fail-closed for these callers, since read-only `InvestmentHandler`
      routes may have a different historical reason for the nil-skip
      behavior — investigate before assuming the shared function's
      contract can change safely for all callers). Resolved by keeping the
      nil-tolerant resolver contract and relying on unconditional,
      fail-closed wiring instead — see the Deferred hardening note above.
- [x] Test authorized same-fund access succeeds and cross-fund access is
      denied (403, not 404, matching `resolvePortfolioByCode`'s existing
      `httputil.Forbidden` path) for every one of the 10 endpoints: list/get
      decisions, create/submit/cancel decision, create/fill/cancel execution,
      record/resolve confirmation.
- [x] Cover repository/application layers too if fund-scope is supposed to be
      enforced beneath the handler as defense in depth — confirm with
      `hasFundAccess`'s existing definition and existing V1 equivalents
      before adding redundant checks that could diverge from them. Existing
      `hasFundAccess`/V1 semantics were sufficient; no redundant check added.
- [x] Preserve audit attribution — do not change what gets audited, only
      close the authorization gap.
- [x] Full backend `go test ./...`, `go vet ./...`, `go build ./...`.
- [x] Independent read-only review of the diff against this exact finding
      (not the manager's fix, the raw diff) before a local commit.

**Non-goals:** do not touch V1 (`/api/v1/investment/...`) routes, do not
change `hasFundAccess`'s allocation-rule semantics, do not add a new
permission catalog entry unless the existing `INVESTMENT_*`
permission set is insufficient (investigate first).

**Dispatch:** `frontier-backend-engineer`, single writer, no isolated
worktree needed (P0-A is the only active writer this session). Independent
review before commit per the skill's high-risk-change rule (security/data-scope).

### P0-B - ACTIVE CORRECTION - Demo migration/bootstrap cleanup

Claude Account #1 produced an uncommitted four-file candidate patch in
`.claude/worktrees/agent-aa1dd1dbf78db7268` that removes named `ben`/`green`
membership blocks from historical migrations `20260716000001/2/3` while
preserving the production role catalog. A second review rejected the G1
completion claim because two release-safety requirements remain unmet:

- `backend/cmd/seed` recursively runs every SQL seed under `database/seeds`
  without an enforced production-environment rejection, so seeds 019/020/021
  and `zz_demo` are still runnable against production.
- Editing already-applied historical migrations is inert. Databases that ran
  the original versions retain their migration-owned demo memberships unless a
  new forward corrective migration removes the exact owned rows.

The candidate worktree is based on `b813b30`, 11 commits ahead of
`neo-develop@b0faad1`. Do not merge that branch wholesale. Recreate or apply
only the reviewed four-file diff from a clean worktree based on the current
`neo-develop`.

Acceptance criteria:

- [ ] Production migration/bootstrap contains no named demo-user privilege or
      data-scope assignment.
- [ ] The seed entry point technically rejects demo seeds in production; a
      folder name, operator convention, or comment is insufficient.
- [ ] Reference/catalog seeds and explicit development/test demo seeds have
      separate, documented execution paths.
- [ ] A new forward migration removes only the exact historical
      migration-owned Ben/Green memberships from already-upgraded databases,
      without deleting production roles, function rights, legitimate
      administrator-created memberships, or data scopes.
- [ ] A fresh disposable database proves migration-only state has zero named
      demo grants while required production role catalogs remain.
- [ ] An upgrade fixture containing the historical memberships proves the
      forward migration removes exactly those rows.
- [ ] Explicit development/test seeding restores the intended synthetic demo
      behavior idempotently, including the anti-data-scope-escalation guard.
- [ ] A production-configured seeder invocation fails before executing demo
      SQL, with an automated test.
- [ ] Up/down/up lifecycle, `git diff --check`, backend formatting/tests/build,
      and an independent database/security review pass with no P0/P1.
- [ ] Integration is performed from current `neo-develop` without merging or
      cherry-picking unrelated history from the old G1 worktree.

### P0-C - QUEUED - Fill validation

Not started. Starts after P0-A/P0-B integrate cleanly (fill validation may
touch the same execution handlers/commands P0-A touches — sequence to avoid
concurrent-writer conflicts on `execution_handler.go` and
`portfolio_v2_execution_handler.go`).

### P0-D - QUEUED - Compliance binding auditability

Not started. Independent of P0-A/B/C's files (touches compliance binding
create/deactivate, not investment V2 handlers); could run in parallel with
P0-B/C in a separate worktree once P0-A is integrated, if Kanta authorizes
parallel writers.

### 2026-07-24 review expansion and execution order

**Review verdict:** NOT READY - CONTROL AND PRODUCT BLOCKERS. The repository
contains credible portfolio, approval, permissions, compliance-pipeline,
ledger-projection, and workflow components, but there is no single
authoritative route from approved decision -> execution -> confirmation ->
ledger -> holdings/cash -> valuation -> end-of-day close.

**Current-source evidence that must remain reproducible until fixed:**

- `ExecutionCommandHandler.Fill` updates only the execution row and does not
  enforce positive/cumulative/ordered bounds, rerun compliance, check the
  workflow day, transition the decision to `EXECUTED`, or post the ledger.
- `TradeConfirmationCommandHandler.Record/Resolve` updates confirmation state
  without settlement or ledger effects.
- `PostTransactionRequest.SourceDecisionID` and `SourceExecutionID` are
  optional, so a direct BUY/SELL ledger post can bypass decision approval,
  execution, and confirmation.
- Workflow production wiring still constructs
  `NopInvestmentQueryAdapter`, which always reports zero transactions to
  manager approval.
- Fund-less decisions, executions, ledger posts, and reversals skip the
  fund-scoped workflow gate in the current dirty implementation.
- The frontend has no create/fill/cancel execution action and no record-
  confirmation action, while OP-02/OP-03 are marked `fullyAvailable: true`.
- `regulatory.thai_sec` remains a warning stub and
  `credit_rating.minimum` still passes by default.
- Approval subject synchronization happens after approval commit and records
  callback failure as retryable, but no replay path was found.
- The server exposes manual watchlist evaluation but starts no watchlist
  evaluation scheduler.
- CI typechecks/builds the frontend but does not run Vitest, and its database
  E2E command selects IAM tests rather than the Portfolio V2 lifecycle suites.

**Safety gate before P0-E through P0-I:** do not layer these changes onto
Claude's active fund-optional dirty tree. First preserve and review that work,
resolve the fund-optional policy conflict in P0-H, and obtain Kanta's explicit
authorization for any commit, migration execution, container/database change,
push, or deployment. Documentation-only planning is authorized; implementation
is not authorized by this task-board update.

### P0-E - QUEUED - Authoritative trade-to-ledger lifecycle

**Objective:** remove the two competing financial paths. An approved investment
must have one idempotent, attributable lifecycle from decision through actual
financial effect.

**Owner:** investment backend/domain first; DBA review for constraints; frontend
consumer only after the backend contract is accepted.

**Dependencies:** P0-B, P0-C, and P0-H. P0-D may proceed independently if file
ownership is isolated.

**Acceptance criteria:**

- [ ] Record an explicit owner-approved lifecycle policy stating which event
      creates official financial effect: execution fill, matched confirmation,
      settlement, or another named event. Do not infer this in code.
- [ ] Require LIVE security BUY/SELL ledger posts to reference and validate an
      approved decision and eligible execution; reject guessed, missing,
      cross-portfolio, cancelled, overfilled, or already-posted sources.
- [ ] Post ledger, holdings, and cash atomically with the authoritative event,
      or use a durable outbox/idempotent consumer with observable retry status.
- [ ] Add database FKs/business keys and idempotency for source decision,
      execution, confirmation, and external broker identity as appropriate.
- [ ] Define partial-fill, multi-fill, cancellation, correction, reversal, and
      duplicate-request behavior.
- [ ] Transition decision/execution/confirmation status consistently and expose
      one correlated portfolio activity trail.
- [ ] Prove retry after a timeout or audit/notification failure cannot create a
      duplicate execution or duplicate ledger effect.
- [ ] Add Go domain/application/repository tests plus live-Postgres E2E for
      decision -> approval -> execution/fill -> confirmation/settlement ->
      ledger/holdings/cash.

**Failure and rollback behavior:** failed compliance, workflow, validation,
idempotency, persistence, or source-correlation checks create no partial
financial effect. Reversal is append-only, attributable, and cannot be used to
bypass a closed day without an explicit audited override permission.

**Non-goals:** do not use the frontend to compensate for missing backend
invariants; do not keep direct ledger posting as an undocumented alternative
trade workflow; do not hand-edit generated Swagger or TypeScript contracts.

### P0-F - QUEUED - Real EOD activity integration and workflow enforcement

**Objective:** make the global business-date workflow reflect real investment
activity and close every mutation path when the day is not open.

**Owner:** workflow backend with an investment-owned query adapter; independent
workflow/investment reviewer required.

**Dependencies:** lifecycle policy from P0-E and fund-scope decision from P0-H.

**Acceptance criteria:**

- [ ] Replace `NopInvestmentQueryAdapter` in production wiring with a real
      transaction/execution/confirmation summary for the global business date.
- [ ] Manager approval cannot accept a zero-transaction attestation when real
      transactions exist and cannot proceed while required review,
      confirmation, or reconciliation work is pending.
- [ ] Execution create/fill/cancel, confirmation record/resolve, ledger post,
      and reversal all use a positive `IsTradeAllowed` check and a race-safe
      day-row lock where financial state changes.
- [ ] Missing workflow day is fail-closed, not interpreted as merely
      "unlocked."
- [ ] Replace or explicitly gate the weekend-only holiday adapter before
      production; Thai business-calendar behavior requires owner/operations
      approval.
- [ ] Add concurrency tests proving manager approval/close cannot race a fill,
      confirmation, ledger post, or reversal.
- [ ] Add E2E proof for normal open/approve/close/accounting-close paths and
      every reverse/cancel transition.

### P0-G - QUEUED - Portfolio-type ledger policy and LIVE approval

**Objective:** implement the owner-confirmed portfolio-type policy without
mutating append-only ledger rows before approval.

**Human decisions required before implementation:**

- [ ] Kanta confirms whether LIVE approval applies only to `CASH_IN`, all cash
      movements (`CASH_IN`, `CASH_OUT`, fee, dividend), or also security
      BUY/SELL. This decision must align with P0-E and must not create a second
      trade-approval path accidentally.
- [ ] Kanta confirms whether a submitter may edit/cancel a pending LIVE ledger
      request and what happens when the business day closes while it is
      pending.

**Acceptance criteria after those decisions:**

- [ ] `SIMULATION` posts eligible paper transactions immediately and marks all
      outputs non-official.
- [ ] `MODEL` is rejected by the backend as well as hidden/disabled in the UI;
      it never receives official ledger, cash, execution, or valuation effects.
- [ ] `LIVE` uses a separate pending-request entity/table and a configured
      approval subject/process; no `investment__portfolio_transactions` row or
      holding/cash effect exists before final approval.
- [ ] Approval callback is idempotent and replayable; approval committed but
      subject-sync failed is visible and recoverable.
- [ ] Rejection, withdrawal, expiry, cancellation, duplicate approval callback,
      stale workflow day, compliance failure, and permission loss are covered.
- [ ] Backend/API/database work lands first; generated contracts are then
      regenerated; frontend implements submit/status/cancel/recovery UX in
      EN/TH/ZH; integrated browser proof follows.

### P0-H - POLICY RESOLVED; IMPLEMENTATION IN PROGRESS - Fund-optional governance

**Owner decision D1 (Kanta, 2026-07-24):** `FUND_OPTIONAL` supersedes the
2026-07-15 fund-required decision. Portfolios and their financial activity may
exist without a fund association. The portfolio is the scope key when no fund
exists; absence of `fund_id` never bypasses workflow, EOD, compliance,
approval, valuation, reporting, permissions, or audit.

- [x] Record D1 and mark the earlier fund-required decision superseded.
- [x] Align durable manager memory with `FUND_OPTIONAL`.
- [ ] Ensure migrations, backend invariants, permissions, frontend behavior,
      generated API documentation, and tests all express the same policy.
- [ ] Prove fund-less LIVE financial activity receives the same mandatory
      control gates as fund-bound activity.

### P0-I - QUEUED - Complete and honestly label the trader/operator frontend

**Owner:** frontend after P0-E/F/G backend contracts are accepted and generated.

**Acceptance criteria:**

- [ ] Authorized traders can create, fill/partially fill, and cancel executions
      through real typed endpoints with validation/conflict/retry states.
- [ ] Authorized operations users can record/import and resolve confirmations.
- [ ] Portfolio decision detail shows the complete correlated lifecycle and
      official versus non-official financial effect.
- [ ] OP-02/OP-03 cannot display `ready` while required mutations or backend
      contracts are missing; limited capability explains the exact gap.
- [ ] Rejection, approval-sync failure, compliance unavailable, closed day,
      stale data, duplicate request, partial fill, mismatch, and empty states
      have explicit user guidance.
- [ ] EN/TH/ZH, permission-denied, desktop/tablet/mobile, light/dark,
      keyboard/accessibility, console, and network behavior receive browser
      evidence.

### P1 follow-ups required before production readiness

- [ ] Implement durable approval-subject sync retry/replay and operator
      visibility; notification alone is insufficient.
- [ ] Standardize audit transaction semantics. Do not return a generic failure
      after a mutation committed, and do not silently accept missing audit for
      financial/compliance changes.
- [ ] Pass actor identity into compliance binding deactivation and persist an
      immutable create/change/deactivate audit trail.
- [ ] Complete Thai SEC/BOT and credit-rating controls only after the existing
      human compliance/legal source-to-rule gate is approved.
- [ ] Add scheduled watchlist evaluation and periodic/post-trade compliance
      evaluation with idempotency and stale-provider behavior.
- [ ] Add frontend Vitest and Portfolio V2 financial lifecycle/database E2E to
      CI; fix the current Nuxt typecheck baseline instead of excluding it.
- [ ] Add a unified auditor-facing portfolio timeline across workflow,
      research, decision, approval, compliance, execution, confirmation,
      ledger, reversal, valuation, and EOD.
- [ ] Move application-layer raw SQL into query/repository adapters where it
      blocks deterministic testing, especially valuation/allocation queries.

### Program-wide verification and release gates

- [ ] Every child task has an exclusive write scope, focused tests, broad tests,
      `gofmt`/lint/typecheck/build as applicable, `git diff --check`, and a
      completion report listing exact commands and residual risk.
- [ ] Database/security/lifecycle changes receive independent review against
      the raw acceptance contract and integrated diff.
- [ ] One automated real-Postgres scenario and one authenticated browser
      scenario prove the full happy path and selected failure paths without
      manual DB correction.
- [ ] Generated Swagger and frontend API types are regenerated from source and
      deterministic.
- [ ] No unresolved P0 or P1 control finding is relabelled as documentation,
      demo limitation, or frontend-only risk.
- [ ] Kanta explicitly approves business policy, migrations, commit scope,
      push, deployment, and any production action.

Acceptance for the whole program: every priority-ordered fix is evidenced by
a diff, a passing focused-then-full test gate, and an independent read-only
review before a local commit; no push; excluded `docs/api/`/Bruno files
remain preserved untouched throughout.

## P0 - QUEUED - Investment Integration Review and Breaches Browser UAT

Goal: finish review of the local `neo-develop` investment integration without
pushing or claiming production readiness. Code is separated into scoped
commits; the remaining user-visible gate is authenticated browser proof of the
redesigned Compliance Breaches workflow.

- [x] Commit compliance backend cleanup and response DTO alignment as
      `8d43747`.
- [x] Commit latest-available portfolio AUM, THB FX, dashboard, localization,
      tests, and the valuation Bruno request as `39ab72d`.
- [x] Commit regenerated compliance/valuation Swagger and TypeScript contracts
      plus response-envelope handling as `09d9421`.
- [x] Commit the API-backed Compliance Breaches frontend redesign, EN/TH/ZH
      copy, and tests as `f6c5205`.
- [x] Pass focused Breaches tests (59), full frontend Vitest (546), production
      build, touched-file typecheck comparison, and `git diff --check`.
- [ ] Run authenticated desktop/tablet/mobile browser UAT in EN/TH/ZH and
      light/dark themes; verify override permission/conflict, detail drawer,
      console, and network behavior.
- [ ] Resolve ownership and review of the excluded `docs/api/` changes before
      any commit containing them.
- [ ] Review the excluded Bruno execution requests and local environment values
      as a separate scope.
- [ ] Obtain explicit owner authorization before any push. No push is currently
      authorized.

Acceptance: local `neo-develop` remains reviewable; every pushed file, if later
authorized, belongs to an evidenced commit; browser claims are made only after
observed proof; excluded concurrent work is preserved.

## P0 - COMPLETED LOCALLY - Latest-Available AUM and THB FX (IMS-LATEST-AUM)

Goal: apply the owner's 2026-07-20 policy that both company and managed-
portfolio totals use every portfolio's latest available valuation and convert
all currencies with the latest valid authoritative FX quote into configured
THB, while disclosing valuation freshness rather than hiding the total.

- [x] Include older latest valuations and stale-input snapshots in company and
      mine aggregates; keep missing valuation and missing/invalid FX fail-
      closed.
- [x] Resolve canonical `FX_USDTHB` through the reference-data Yahoo mapping
      (`USDTHB=X`) even when a legacy symbol row already exists, and persist the
      provider mapping without violating the legacy asset-type constraint.
- [x] Convert cross-currency AUM and local P&L delta with the same latest FX
      quote; do not invent historical FX movement from mismatched quote dates.
- [x] Add `latest_available_portfolio_count` and
      `oldest_included_business_date` to the contract, DTO, Swagger, generated
      TypeScript API, and frontend normalization.
- [x] Render EN/TH/ZH freshness disclosure on AUM/P&L cards when older latest
      valuations are included.
- [x] Prove the live authenticated dashboard shows Entire Company AUM
      (`฿9.91B`, 7/7, oldest Jul 17) and My AUM (`฿5.68B`) in THB with zero
      browser warnings/errors.
- [x] Pass focused tests, full backend test/vet/build, deterministic API-client
      generation, full 46-file/497-test Vitest, and Nuxt production build;
      touched-file typecheck diagnostics remain zero.

Safety and delivery state: uncommitted on `neo-develop` at base `0c14c1d`; no
push, migration, seed, valuation mutation, or production deployment occurred.
Local containers were rebuilt. A normal authorized refresh persisted one real
Yahoo `USDTHB=X` market-data snapshot (`33.62000000 THB`); no fabricated data
was inserted. Production remains blocked by the Oracle dirty migration at
version `20260615000003`.

## P1 - COMPLETED LOCALLY - Architecture Cleanup (IMS-ARCH-CLEANUP)

Goal: owner-authorized (2026-07-20) pull-forward of the safe, behavior-
preserving items from the Architecture Cleanup checklist plus three items from
the 2026-07-20 code review. Zero wire-contract, SQL-result, or error-semantics
change except where an item explicitly authorizes a generated-type fix. The
regulatory task (IMS-REG-TH-SEC) remains gated and untouched. Full evidence in
`HANDOFF.md`'s "Session: Architecture Cleanup (2026-07-20)".

- [x] 1. Deduplicate exposure math: the three identical holding market-value
      helpers in `rules/allocation`, `rules/ratio`, and `rules/concentration`
      now call one shared `spi.HoldingMarketValue`. Fixed the N+1 rule-version
      loading in `ListPortfolioRules` with a batched
      `RuleInstanceRepository.GetCurrentVersions` (one `ANY($1)` query).
- [x] 2. `PortfolioRuleCatalogEntry.Parameters` is now `any` (annotation-only
      drift fix): generated TS changed `Record<string, never>` → `unknown`;
      wire bytes unchanged because producers still assign raw pre-encoded JSON.
- [x] 3. `@Failure 422` on CreateExecution/CreateExecutionByCode: found already
      complete in source and generated docs (closed earlier by `e80f086` +
      `790a2f7`); verified, no change needed.
- [x] 4. zh-TW normalization of `zh/portfolio.ts`: all pre-existing Simplified
      keys converted to Traditional matching `zh/compliance.ts` conventions
      (載入/失敗/重試/紀錄/存取權限); placeholders preserved; scripted scan
      finds zero remaining Simplified-only characters.
- [x] 5. `cleanupInvestmentPrereqs` now deletes `investment__trade_confirmations`
      → `investment__executions` → `investment__decisions` before the
      portfolio/fund deletes (confirmations added because their RESTRICT FK on
      execution_id would otherwise still block the chain).
- [x] 6. `valuation_summary_adapter.go`: `listAllFunds`/`listAllPortfolios`
      collapsed into one generic `listAllPages` helper; total-drift,
      invalid-metadata, and exhaustive-paging fail-closed guards preserved with
      byte-identical error messages.
- [x] 7. Per-fund portfolio N+1 replaced with one paginated listing filtered by
      Status/ManagerUserID/(mine-only) AccessibleFundIDs, grouped by FundID in
      memory; all 25 existing adapter tests (coverage counts, exclusion
      ordering, mine-scope intersection, forced one-row paging) pass unchanged.
- [x] 8. `command.PreTradeCheckResponse`/`PostTradeCheckResponse`/`BreachSummary`
      mirrored into `compliance/transport/dto/response` with mapping funcs like
      the other five endpoints; JSON field names proven identical by new wire-
      JSON tests; Swagger annotations updated; contracts regenerated
      deterministically (second run byte-identical).

Found-but-deferred (logged, not acted on):

- The same `swaggertype:"object"` drift still exists on
  `dto/response/responses.go` fields (`RuleInstanceVersion.Parameters`,
  `CheckRecord.ParameterSnapshot`/`Evidence`) — V1-only, out of item 2's scope.
- Concurrent-writer observation: files under `docs/api/` (README, new
  `current-api-reference.md`, `_build_current_api_docs.py`, etc.) were
  created/modified at 12:27–12:32 on 2026-07-20 by a process outside this
  session; left untouched, flagged to the owner.

## P0 - COMPLETED LOCALLY - Company AUM Visibility (IMS-COMPANY-AUM-VISIBILITY)

Goal: implement the owner's 2026-07-18 policy that every authenticated
dashboard user can see the complete company aggregate AUM/P&L, without granting
item-level fund/portfolio access or weakening existing drill-down permissions.

Manager checklist:

- [x] Confirm that dashboard routes still require authentication.
- [x] Remove function-permission and IAM fund-scope filtering from only the
      `company` valuation-summary scope.
- [x] Filter `mine` by portfolios managed by the authenticated user, intersected
      with accessible funds; do not substitute fund-manager ownership.
- [x] Remove item-level fund/portfolio exclusions from company responses while
      preserving aggregate coverage evidence.
- [x] Enforce company-wide scope again at the valuation provider boundary and
      exhaust all fund/portfolio repository pages without silent row caps.
- [x] Make company the default and always offer company/mine scope choices to
      authenticated dashboard users.
- [x] Align EN/TH/ZH coverage copy and regenerate Swagger/TypeScript contracts.
- [x] Add focused backend and frontend regression tests.
- [x] Record the final full Go and frontend verification results in
      `HANDOFF.md` after the currently running gates complete.

Safety and delivery state: changes are uncommitted on `neo-develop` at base
`0c14c1d`; no push or production deployment occurred. Local backend/frontend
containers were rebuilt for verification without database writes. Production
remains blocked by the Oracle database dirty migration at version
`20260615000003`. The environment and GitHub secrets are not the cause of that
failure.

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

Merge-readiness remediation checklist (remaining unchecked items require owner
approval):

- [ ] Enforce fund data scope on every Portfolio V2 decision, execution, and
      confirmation read/write route, fail closed, and prove cross-scope denial.
- [ ] Move `ben`/`green` memberships out of production migrations into
      environment-specific seeds or explicit provisioning.
- [ ] Fail closed for missing asset classification and missing mandatory LIVE
      portfolio rule configuration; validate/recheck actual fills.
- [ ] Add actor-attributed immutable audit events for compliance binding create
      and deactivate operations.
- [x] Approve and implement reporting-currency/FX policy: company/mine use each
      portfolio's latest available valuation and latest valid FX into configured
      THB, with freshness disclosure and no omitted currency subtotal.
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

## Superseded Owner Decision - Historical Record

### Fund-Optional Portfolio Development

On 2026-07-15 Kanta said not to develop optional fund association and to keep
fund association required. **Superseded by Kanta's later explicit D1 decision
on 2026-07-24: `FUND_OPTIONAL` ("what I want is fundless").** Preserve this
section only as history; it is not current product direction and must not be
used to block or reverse fund-less portfolio work.

## P2 - GATED - Regulation and Restriction

### Build Real Portfolio Regulation Rules (IMS-REG-TH-SEC)

Status: **GATED; NOT COMPLETE**. Thailand / Thai SEC is the confirmed first
jurisdiction. Official sources are mapped in
`docs/compliance/thai-sec-regulatory-source-map.md`. The remaining human gate
is selection of the first applicable product regime and approval of its exact
source-to-rule matrix; no legal threshold may be invented or enforced before
that approval. The regulatory program remains separately gated; the current
integration review must report this gap honestly and must not close it.

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
- [x] Resolve duplicated exposure math and N+1 rule-version loading. Done
      2026-07-20 (IMS-ARCH-CLEANUP item 1): shared `spi.HoldingMarketValue`
      plus batched `GetCurrentVersions`.
- [ ] Address branch-wide TypeScript typecheck debt as a dedicated quality task.
- [ ] Review stale module/runbook claims after Portfolio Compliance V2 commits.
- [x] Add a per-row effective-state indicator (scheduled/effective/expired)
      to `PortfolioComplianceView.vue`'s bound-rules table so `is_active=true`
      is not read as "currently enforced" when `effective_from`/`effective_to`
      place the binding outside today's reference date. Found 2026-07-13;
      fixed 2026-07-13 in the frontend-correction session (uses the
      browser's local date, explicitly labelled "as of", since the
      workspace exposes no authoritative backend business date).
- [x] Fix the `PortfolioRuleCatalogEntry.parameters` generated-type drift:
      `backend/pkg/contract/contracts.go`'s `swaggertype:"object"` annotation
      produces `Record<string, never>` in `ims-api.d.ts`. Found 2026-07-13.
      2026-07-13: frontend now routes through a safe `unknown`-accepting
      `asParameterRecord` boundary normalizer in `PortfolioComplianceView.vue`.
      Closed 2026-07-20 (IMS-ARCH-CLEANUP item 2): the field is now `any`
      (producers still assign raw JSON, wire bytes unchanged) and the
      regenerated client types `parameters?: unknown`.
- [x] Done 2026-07-20 (IMS-ARCH-CLEANUP item 4).
      `zh/portfolio.ts` uses Simplified Chinese for its pre-existing keys,
      inconsistent with `zh/compliance.ts`'s Traditional (zh-TW) convention
      used everywhere else. Found 2026-07-13 while adding new zh-TW content
      to `portfolio.ts`; the new keys added that session use Traditional
      characters, but the surrounding pre-existing keys were left as-is
      (out of scope). Recommend a dedicated zh-TW normalization pass over
      `portfolio.ts`.
- [x] Done 2026-07-20 (IMS-ARCH-CLEANUP item 5, plus trade confirmations).
      `tests/e2e/investment_test.go`'s shared `cleanupInvestmentPrereqs`
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
- [x] Verified already closed by `e80f086`/`790a2f7`; re-confirmed 2026-07-20
      (IMS-ARCH-CLEANUP item 3). `CreateExecution` (`execution_handler.go:112`) and
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
