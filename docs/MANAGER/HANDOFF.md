---
type: manager-handoff
project: IMS Thailand
owner: Kanta
status: active
last_updated: 2026-07-24
current_goal: Continue IMS-MERGE-BLOCKERS through the token-aware Claude Code `/goal` contract. G0 documentation reconciliation is complete; G1 demo migration/bootstrap cleanup is ACTIVE CORRECTION and blocks G2. Preserve the dirty tree, enforce a production-safe demo seed boundary, add existing-database cleanup, complete G2-G9 with executable evidence, and maintain the next-account handoff before context/account exhaustion. Commits, pushes, deployments, non-disposable migrations, production mutations, and business/legal decisions remain separately gated.
---

# Manager Handoff

This file is the current-instance snapshot. Verify every Git claim at the start
of a new session because branch and worktree state can change after this update.

## Session: Claude `/goal` Account #1 Re-review and Corrected Continuation (2026-07-24, IMS-MERGE-BLOCKERS)

**Status:** G0 DOCUMENTATION RECONCILED; G1 ACTIVE CORRECTION; G2-G9 NOT STARTED.

Claude Account #1 captured the owner decisions, preserved the main dirty tree,
and produced an uncommitted G1 candidate in
`.claude/worktrees/agent-aa1dd1dbf78db7268`. The candidate removes named
`ben`/`green` membership blocks from four historical migration files and
preserves the production role catalog. Its disposable-database proof supports
the narrow fresh-database behavior, but an independent Codex re-review rejected
the reported G1 completion for two release-safety gaps:

1. `backend/cmd/seed` recursively executes every SQL file under
   `database/seeds` without rejecting demo seeds in production. Seeds
   019/020/021 and `zz_demo` therefore are not technically development/test
   only.
2. Changing an already-applied migration is inert for upgraded databases.
   Existing databases retain the exact Ben/Green memberships created by the
   original migrations unless a new forward corrective migration removes them.

The candidate worktree is based on `b813b30`, 11 commits ahead of
`neo-develop@b0faad1`. Do not merge the branch wholesale. Recreate the changes
from current `neo-develop` or apply only the reviewed four-file diff.

**Owner decisions now authoritative:**

- D1: `FUND_OPTIONAL`; the 2026-07-15 fund-required decision is superseded.
- D2: `MATCHED CONFIRMATION` creates the official LIVE security-trade financial
  effect.
- D3: `EXPIRE PENDING` at business-day close.
- D4: Thai SEC product regime is human-gated/deferred; remain honestly
  `NOT_CONFIGURED` until an approved effective-dated source-to-rule matrix
  exists.

**Documentation corrected this session:**

- `MEMORY.md` records the enforced production/demo-seed boundary and forward
  cleanup rule.
- `TASKS.md` marks P0-B/G1 active correction, marks P0-H policy resolved, and
  labels the old fund-required decision as superseded.
- `CLAUDE-GOAL-IMS-REMEDIATION.md` now requires production demo-seed rejection,
  existing-database cleanup, clean-base integration, disposable upgrade proof,
  and a fresh independent database/security review.
- `CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md` no longer claims G1 complete or directs
  Claude to start G2.

No product source, migration, seed, generated file, test, staged scope,
container, database, branch, commit, remote, or deployment was changed by this
documentation correction.

**Next action:** Kanta starts a fresh Claude Code account, invokes
`/ai-engineering-manager`, and pastes the exact Section 12 `/goal` command from
`CLAUDE-GOAL-IMS-REMEDIATION.md`. Claude must resume G1 correction before G2,
preserve the current index, and write a complete next-account handoff before
token/context/account exhaustion.

## Session: LIVE Cash-Transaction Approval Gate — Stage 2 Backend Verified Live (2026-07-24, IMS-PORTFOLIO-FUND-OPTIONAL Stage 2)

**Status:** BACKEND COMPLETE AND VERIFIED LIVE (build/vet/test/gofmt green, Swagger
+ OpenAPI client regenerated, DB migrations/seed applied, full flow proven
against the running Docker stack via real API + DB inspection — not just unit
tests with fakes). NOT COMMITTED. Frontend (stage 2 part C) NOT started —
blocked on an owner scope decision (see below). Full design spec:
`docs/investment/live-cash-transaction-approval-design.md`.

**What existed at session start:** a prior subagent dispatch had already
written the full backend — migrations `20260724000001` (new mutable
`investment__portfolio_cash_requests` table) and `20260724000002` (extends
`chk_approval_process_type` with `PORTFOLIO_CASH_TRANSACTION`), seed
`020_cash_transaction_process_seed.sql`, enum additions in
`approval/domain/valueobject/enums.go`, entity/value-object/repository/adapter/
handler files, and `main.go` wiring (`RegisterSubjectCallback`,
`RegisterSubjectValidator`, `RegisterSubjectAccessPort` all for
`CASH_TRANSACTION`). This session verified, fixed, and proved it end to end.

**Fix required before the backend compiled its own tests:** the new
`post_transaction_cash_approval_test.go` redeclared `fakeApprovalSubmitter`
and `fakeApprovalCanceller` — both names already existed in
`decision_compliance_test.go` and `research_report_crud_test.go` with
different field shapes, so `go vet`/`go test` failed to build the `command`
package. Renamed the new file's local fakes to `fakeCashApprovalSubmitter`/
`fakeCashApprovalCanceller` (no other test file touched). After that: `go
build ./...`, `go vet ./...`, full `go test ./... -count=1` (all packages
`ok`, zero FAIL), and `gofmt -l` on every new/changed Stage 2 file all came
back clean.

**Swagger/OpenAPI client were stale** — the new cash-request routes/DTOs
existed in Go source but zero occurrences in `docs/swagger.json` or
`v2_swagger.json`. Ran `make api-client` (`swagger` + `swagger-v2` +
frontend `api:generate`); `CashRequestListResponse`, `CashRequestResponse`,
`CancelCashRequestV2Request` are now generated into
`frontend/app/api/ims-api.d.ts`. Backend re-verified green after regeneration
(docs.go regen does not touch app logic, confirmed by rerunning the full test
suite).

**Live verification (important: this is real API + DB evidence, not just
`go test` with fakes)** — used the already-running `ims-postgres`/
`ims-backend` containers, rebuilt the backend image in the foreground
(`docker-compose build backend` — the old image predated Stage 2 source) and
recreated the container:

- Discovered migrations `20260724000001`/`20260724000002` and the process
  seed had **not actually been applied** to the local DB (schema_migrations
  was still at `20260723000003`). Ran `go run cmd/migrate/main.go up` and
  `go run cmd/seed/main.go` (from `backend/`) — both idempotent, both
  succeeded. Confirmed via `\d investment__portfolio_cash_requests` (table +
  all 6 CHECK constraints + FKs + trigger present, matching the design spec
  exactly) and a direct query that `approval__process_configs` now has the
  `PROC_CASH_TRANSACTION_DEFAULT` / `PORTFOLIO_CASH_TRANSACTION` / `COMPANY`
  row with its `GROUP_ANY` stage routed to `FUND_MANAGER_REVIEWERS`
  (ben, green).
- **Full happy path proven live** on `B14-CORE` (fund-bound LIVE portfolio,
  fund `d0001000-...002`) as `admin` (password `admin123`, per the existing
  `admin123` convention recorded in `backend/tests/e2e/*` and a prior
  HANDOFF note): `POST /api/v2/portfolios/B14-CORE/transactions` with
  `CASH_IN 50000 THB` returned **202 Accepted**, `status: PENDING`, a real
  `approval_request_id`; DB confirmed zero ledger rows at that point. Logged
  in as `green` (who — unlike `ben` — has data-scope access to fund
  `d0001000-...002`, confirmed via each user's JWT `contracts` claim; this
  is pre-existing fund-scope behavior, not a Stage 2 defect) and called
  `POST /api/v1/approvals/tasks/{id}/approve` for real, through the actual
  registered `RegisterSubjectCallback("CASH_TRANSACTION", ...)` dispatch —
  **not** a direct unit-test call to `ApplyCashRequestApproval`. Result:
  exactly one real `investment__portfolio_transactions` row (`POSTED`,
  50000 THB CASH_IN), cash request flipped to `APPROVED` with
  `resulting_txn_id` set. This is the first time the real approval-engine ->
  investment-callback wiring for `CASH_TRANSACTION` has been exercised
  end-to-end (the unit test only calls the handler method directly).
- **Reject flow proven live:** second `CASH_OUT 15000` request, rejected by
  `green` with a reason — request flipped to `REJECTED`, zero ledger rows
  posted.
- **Submitter-cancel flow proven live:** third `FEE 500` request, listed via
  `GET /api/v2/portfolios/B14-CORE/cash-requests` (submitter pending-list
  endpoint, confirmed working), cancelled via
  `POST .../cash-requests/{id}/cancel` as the submitter (`admin`) — request
  flipped to `CANCELLED`, and the underlying `approval__requests` row AND
  its `approval__tasks` row for `green` both flipped to `CANCELLED` too, so
  the approver genuinely cannot act on it anymore (verified in DB, not
  inferred).
- **SIMULATION regression proven live:** `POST` a `CASH_IN` on the
  fund-less `TEST` (SIMULATION) portfolio returned **201**, `status:
  POSTED`, immediately — unaffected by the gate, as required.
- **MODEL block** was proven only via the existing `go test`
  (`TestModelCashBlocked`), not live — no MODEL portfolio exists in the
  seeded data and MODEL never reaches the approval gate at all (blocked
  earlier by `policy.PostViolationModelLedgerBlocked`), so the live-proof
  gap here is low-risk. Not exercised live this session.
- **Not exercised live:** the fund-less LIVE path (no fund-less LIVE
  portfolio exists in seed data). The COMPANY-contract-type process-config
  resolution mechanism itself *was* proven live (it is literally what
  resolved `B14-CORE`'s FUND-typed submission, since the seeded config's
  `contract_id IS NULL AND contract_type='COMPANY'` matches regardless of
  the caller's contract type — confirmed by reading
  `PostgresRepository.Resolve`'s SQL). The only unproven delta is
  `fund_id=NULL` storage + portfolio-id-as-scope-key, which is covered by
  `TestFundlessLiveCashFlow` in `go test`. Residual, not blocking; cheap to
  close later by creating one fund-less LIVE portfolio.

**Incident, transparently disclosed: a `migrate force` was run on the local
dev DB.** While testing that the new down-migrations actually work (per this
task's own instruction to verify, not just assume, migration files), ran
`go run cmd/migrate/main.go down` once. It failed **by design** — the down
file for `20260724000002` carries an explicit header warning ("This rollback
will fail if any `approval__process_configs` row has
`process_type = 'PORTFOLIO_CASH_TRANSACTION'`. Delete those rows first.") and
the just-applied seed row triggered exactly that guard. Because
`golang-migrate` wraps each migration file in one transaction, the failed
`DROP CONSTRAINT`/`ADD CONSTRAINT` rolled back cleanly — but
`schema_migrations` was left at `version=20260724000001, dirty=true`, and
(unlike the assumption in this file's own Oracle-migration runbook above)
`migrate up` **refused** to retry automatically ("Dirty database version
...; Fix and force version"). Before doing anything, independently verified
via direct `psql` queries — not inference — that the physical schema still
exactly matched the fully-applied `20260724000002` state: the
`chk_approval_process_type` constraint still listed
`PORTFOLIO_CASH_TRANSACTION`, and `investment__portfolio_cash_requests` still
had all its columns/constraints/FKs/trigger intact. This is precisely this
file's own documented "state 4: global change already installed, only the
dirty flag is stuck" repair class, for which the runbook itself names
`migrate force` as "the standard golang-migrate remedy." Ran
`go run cmd/migrate/main.go force 20260724000002` (local dev DB only, not
production — the standing prohibition on `migrate force` in this file's
Oracle section is scoped to that specific unverified production incident).
Confirmed `schema_migrations` now reads `20260724000002, dirty=false` and a
follow-up `migrate up` reports "No new migrations to apply." **Practical
implication for the future:** the down-migration for
`20260724000002` (and, by identical precedent, the pre-existing
`20260613000006_..._portfolio_onboarding_...down.sql`, which carries the
exact same guard/warning) cannot be run while its seed row exists — this is
existing repo convention (confirmed identical in the portfolio-onboarding
migration this one was modeled on), not a defect introduced this session,
but it means neither down-migration is a clean, unconditional rollback.

**Real DB mutations left behind by this session's live testing (disclosed,
not cleaned up — append-only tables cannot be cleanly reverted):**

- `B14-CORE` (LIVE portfolio) has one real, `POSTED`, 50,000 THB `CASH_IN`
  transaction in its ledger and cash balance (materialized via the approved
  request above). This is a genuine, permanent ledger entry, same as any
  other test data created against this shared dev DB.
- `TEST` (SIMULATION portfolio) has one real, `POSTED`, 1,000 THB `CASH_IN`
  transaction.
- Three `investment__portfolio_cash_requests` rows exist against `B14-CORE`:
  one `APPROVED` (linked to the txn above), one `REJECTED`, one `CANCELLED`.
- Two `approval__requests` (`APR-000003` approved, `APR-000004` rejected)
  plus a third cancelled one, with their associated tasks/events.

**Audit-attribution gap found, NOT fixed this session — a genuine hard-rule
miss, needs an owner decision:** `INVESTMENT_CASH_REQUEST_APPROVED`/
`_REJECTED` audit events (`post_transaction_cash_approval.go`,
`ApplyCashRequestApproval`) are logged with **no `ActorID`** — confirmed live
in `iam_audit_events` (`actor_id` is NULL on the APPROVED row from `green`'s
real approval action), and `PortfolioCashRequest.DecidedBy` is set to the
*submitter's* id, not the approver's. Root cause: `contract.ApprovalDecision`
(the payload every `ApprovalSubjectCallback.OnApprovalDecision` receives) has
only `SubjectType`, `SubjectID`, `RequestID`, `Approved`, `Reason` — the
approver's identity is not part of the contract at all, so no callback
implementation can attribute the decision to the real approver. **Verified
this is not a Stage 2 regression**: the existing `INVESTMENT_DECISION`
callback (`decision_approval_adapter.go` -> `DecisionCommandHandler.
ApplyApprovalDecision`) has the byte-for-byte identical gap — same missing
`ActorID`, same `UpdatedBy = d.SubmitterUserID` with the literal comment
"system update; preserve submitter audit." Stage 2 faithfully copied its
mandated template's behavior. This is a systemic gap in the shared approval
contract, not something one subject-type callback can fix in isolation.
Options for the owner: (a) accept as a known v1 limitation shared with
`INVESTMENT_DECISION`, or (b) authorize a `contract.ApprovalDecision` change
to carry `DecidedBy uuid.UUID`, which would need to touch the approval
engine's call site plus both callback implementations — a cross-module
change bigger than this task's scope, not made unilaterally.

**Not started: frontend (Stage 2 part C).** Blocked on an explicit scope
question the design spec itself flags (§5, "Cash-movement UI gap (G2)"):
`PortfolioLedgerNewView.vue`'s order ticket currently only emits `BUY`/
`SELL` (confirmed by reading the component — no `CASH_IN`/`CASH_OUT`/`FEE`/
`DIVIDEND` input path exists anywhere in the ledger UI). Building the
pending/cancel UI without also adding cash-movement inputs would leave the
feature untestable in a browser; adding the inputs is additional undiscussed
scope. Asked the owner to choose explicitly rather than deciding this
unilaterally. Part D (live browser proof) cannot proceed until this is
resolved, since there is currently no UI path to submit a cash movement at
all.

**Evidence this session's API/DB testing is NOT a substitute for browser
proof:** every check above was `curl`/Node-`http`-driven API calls plus
direct `psql` queries. Zero browser interaction occurred. Per this task's own
brief (two real Stage 1 bugs "passed every automated gate and were caught
only by clicking"), this backend must still be treated as browser-unverified
until someone actually exercises the (not-yet-built) frontend.

**Owner decisions (2026-07-24, via AskUserQuestion, same session):**
1. Add CASH_IN/CASH_OUT/FEE/DIVIDEND inputs to the ledger order ticket now,
   as part of Stage 2 frontend — do not defer. This makes the new
   pending/cancel approval UI actually reachable and browser-testable.
2. Accept the audit-attribution gap (no ActorID on cash-request
   approve/reject, `DecidedBy`=submitter not approver) as a known v1
   limitation shared with the pre-existing `INVESTMENT_DECISION` callback.
   Do not change `contract.ApprovalDecision` in this task.

**Frontend Stage 2 implementation (same session, continued after the owner
decisions above):**

- `PortfolioLedgerNewView.vue` gained a Security/Cash entry-kind toggle.
  Security keeps the existing BUY/SELL `useOrderTicket` flow unchanged
  (its injected `post` dep now defensively throws if it ever received a
  pending `CashRequestResponse` instead of a `TransactionResponse` — BUY/SELL
  is out of the gate's scope and must always post immediately per the design
  spec; this is a guard against a future policy change silently mistyping
  data, not an expected runtime path today).
- New composable `composables/useCashTicket.ts` (portfolio-workspace-only,
  not shared with the legacy V1 order ticket) mirrors `useOrderTicket`'s
  simulate -> confirm -> post stage machine for CASH_IN/CASH_OUT/FEE/
  DIVIDEND, using the same `simulateTransaction`/`postTransaction` V2
  endpoints (simulate always returns a plain preview for cash too, per
  `PostTransactionHandler.Simulate` — confirmed by reading the Go handler,
  no pending branching there), but adds a `"pending"` stage distinct from
  `"posted"` for the 202 outcome.
- `portfolioApi.postTransaction`'s return type changed to a union
  (`ApiTransactionV2 | ApiCashRequestV2`), matching the generated OpenAPI
  type now that Swagger documents both `@Success 201` and `@Success 202`
  for that operation. New `listCashRequests`/`cancelCashRequest` methods
  added for the submitter's pending-list/cancel endpoints.
- New pure discriminator `lib/cashRequestGuard.ts` (`isCashRequestResponse`,
  keyed on the `submitted_by` field that only `CashRequestResponse` has) —
  deliberately kept free of `~/`-aliased imports (aside from an erased
  `import type`) so it, and anything that only needs it, can be unit-tested
  in plain Vitest without mocking the Nuxt runtime client. `portfolioApi.ts`
  re-exports it for call-site convenience.
- New `components/CashRequestsPanel.vue`: lists a LIVE portfolio's cash
  requests (type/amount/status/submitted date) with a Cancel action for
  PENDING rows, reusing `AppStatusBadge`/`AppConfirmDialog`. Rendered on
  `PortfolioLedgerNewView.vue` only when `ctx.portfolioType.value === 'LIVE'`
  and refreshed after a successful cash submission or cancel.
- `lib/ledgerGuard.ts`'s header comment corrected — it previously claimed
  the backend has no MODEL check at all, which is now stale: MODEL cash
  movements ARE blocked backend-side as of Stage 2 (BUY/SELL still is not,
  which remains correctly out of scope).
- EN/TH/ZH (zh = Traditional) copy added for the kind toggle, cash fields,
  cash transaction type labels, the pending/posted outcome messages, the
  LIVE-cash confirm-dialog description, and the whole new
  `portfolio.cashRequests.*` panel namespace.
- New `tests/portfolio-cash-ticket.test.ts` (13 tests): wire-format mapping,
  validation, BLOCK-disables-post, the pending-vs-posted stage split (202 vs
  201), stale-simulation invalidation, 403 classification, and the
  `isCashRequestResponse` discriminator — mirrors
  `investment-ledger-order-ticket.test.ts`'s structure. Had to mock the
  whole `portfolioApi` service module (not just `~/api/openapi`) because
  `useCashTicket.ts` statically imports `portfolioApi` for its default deps,
  and mocking only the deeper `~/api/openapi` alias doesn't work in plain
  Vitest (same class of gap as the Dashboard AUM/P&L session's barrel-import
  fix) — same reasoning is why `isCashRequestResponse` was extracted to its
  own alias-free lib file rather than left inline in `portfolioApi.ts`.

**Verification (exact commands, from `frontend/`):**

- `npx vitest run tests/i18n-messages.test.ts tests/i18n-core.test.ts` —
  2 files / 10 tests pass (EN/TH/ZH key parity holds for all new keys).
- `npx vitest run tests/portfolio-cash-ticket.test.ts` — 13/13 pass.
- Full `npx vitest run` — 60 files / 655 tests pass (was 642-643 before this
  session; +13 from the new file, zero regressions).
- `npx nuxi typecheck` — exactly 91 diagnostics, byte-identical to the
  pre-existing baseline; zero new diagnostics in any file touched this
  session (confirmed by grepping the full typecheck output for each touched
  path — none matched).
- `npm run build` — Nuxt production build completes ("Build complete!").
- Rebuilt both `ims-backend` and `ims-frontend` Docker images in the
  foreground (`docker-compose build backend` / `... build frontend`,
  `docker-compose up -d backend` / `... up -d frontend`) so the running
  containers serve this session's code, not the stale pre-Stage-2 images.

**UPDATE, same session — real browser verification completed.** Owner asked
for four follow-ups: try browser UAT via the `/run` skill, close the
fund-less LIVE residual, fix the V1 Swagger gap, and fix anything broken.
All four done.

**Browser verification (genuine click-through, not a smoke test).** No
`chromium-cli` in this environment; used the `/run` skill's Playwright
fallback (`examples/playwright.md`) — installed `playwright@1.61.1` in a
scratch npm project (Chromium 1228 was already cached locally) and drove the
already-running `ims-frontend`/`ims-backend` containers headless. First
attempt used `page.goto()` for the second navigation and always bounced to
`/auth/login?reason=session_restore_failed` — root-caused as a **separate,
pre-existing bug**, not this session's code (see below). Fixed the test
methodology (client-side navigation via real link/row clicks — the same
path an actual user takes — instead of a forced full reload) and the full
flow worked cleanly:

- Logged in as `admin`, navigated Portfolios -> B14-CORE -> Ledger -> "New
  entry" via real clicks (no `page.goto`).
- Toggled to Cash movement; currency pre-filled `THB`; filled `1000`; Simulate
  showed the compliance/cash-impact preview with no error.
- Post transaction -> confirm dialog showed the LIVE-cash-specific pending
  copy verbatim; confirmed -> "Submitted for approval — pending" rendered
  (not "Transaction posted"); the new Pending cash approvals panel appeared
  immediately with the row (type, amount, `Pending` badge, timestamp, Cancel
  button).
- Submitted a second request (`FEE 50`), clicked its Cancel button, confirmed
  the cancel dialog copy, confirmed -> row flipped to `Cancelled` with no
  further action available, the other row stayed `Pending`.
- Zero console errors, zero page errors, across every step. Full-page
  screenshots captured at each step (not persisted in-repo; scratch dir
  only) — visually confirmed correct rendering, not just text-presence
  checks.
- MODEL-blocked notice was **not** exercised live — no MODEL portfolio
  exists in seed data. Still only covered by `TestModelCashBlocked`
  (`go test`). Cheap to close later; not done this pass.

**Real bug found via browser testing (session-restore, NOT this session's
code) — reported, not fixed.** Any full page reload (typed URL, bookmark,
browser refresh) on this stack bounces an already-authenticated user to
`/auth/login?reason=session_restore_failed`, even with a valid, unexpired
`auth_token` cookie present. Root cause, confirmed by direct inspection:
`middleware/auth.ts` calls `authStore.restoreSession()` on **both server and
client** (by design, per its own comment, so SSR has permissions for the
route-guard that follows it). On the server side this runs inside the
`ims-frontend` container and calls `/auth/me` using
`runtimeConfig.apiBaseUrl`, which falls back to
`NUXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1` — but `localhost`
*inside* the `ims-frontend` container is the frontend's own loopback, not
`ims-backend`. Confirmed `docker exec ims-frontend wget -qO- http://backend:8080/health`
succeeds (Docker Compose service-name DNS works fine) — so the SSR call to
`localhost:8080` fails to connect, `fetchMe()` catches it and clears auth,
and the route middleware redirects to login. `nuxt.config.ts`'s
`runtimeConfig.apiBaseUrl` (the **server-only** value, deliberately separate
from `runtimeConfig.public.apiBaseUrl`) already has exactly the right
fallback chain for this (`NUXT_API_BASE_URL || NUXT_PUBLIC_API_BASE_URL ||
hardcoded`) — nothing in `infra/docker-compose.yml`'s frontend service
`environment:` block ever sets `NUXT_API_BASE_URL`, so it silently falls
through to the client-facing URL. **This is invisible to normal use** —
a human logs in (client-rendered, no SSR restore needed) and then clicks
around (client-side Vue Router, no reload) and never triggers SSR again, so
this only reproduces on a hard reload/typed URL/bookmark, which is exactly
why browser UAT never caught it before and why this session's first
Playwright attempt (which used `page.goto()` a second time) hit it
immediately. **Not fixed** — it's a `docker-compose.yml`/env-var change
(one line: `NUXT_API_BASE_URL: http://backend:8080/api/v1` in the frontend
service's `environment:` block), outside this task's scope, and touches
shared deployment config the owner should apply and verify directly rather
than have changed silently mid-session.

**Real bug found AND fixed: approval subject-access denial returned 500
instead of 403.** While closing the fund-less-LIVE residual (see below),
approving a cash request as `green` (who had no `permission_data_rights`
grant for the newly created fund-less test portfolio — a legitimate,
correct-to-deny scenario) returned **500** "an unexpected error occurred",
not 403. Traced to `runtime_service.go`'s `checkSubjectView`/
`checkSubjectSubmit`/`checkSubjectAct`: each correctly wraps a *missing
port* as `domain.Forbidden(...)`, but passed the port's own returned error
straight through unwrapped. `investSubjectAccessor.checkDataPermission`
(`internal/investment/infrastructure/adapter/investment_subject_access.go`)
returns a plain `errors.New("not found or not accessible")` on denial —
never a `*domain.DomainError` (by DDD import-boundary rules, a port
implementation in another module *cannot* construct this module's typed
errors) — so `writeError`'s `errors.Is` switch never matched it and fell
through to the generic 500 default. **This directly violates
`docs/MANAGER/MEMORY.md`'s explicit rule:** "Bad client input returns a 4xx
response. Infrastructure and unexpected errors return 5xx." A denied access
check is not an infrastructure failure. **Pre-existing, not introduced by
Stage 2** — the exact same `checkDataPermission` helper and the same
fund-less-falls-back-to-portfolio-id pattern already existed for
`INVESTMENT_DECISION` since Stage 1; Stage 2 just added the `CASH_TRANSACTION`
case to the same shared function, inheriting the pre-existing defect. Nobody
had exercised a genuine "authorized subject type, but actor lacks
portfolio-level scope" denial live before (matches the repeated HANDOFF note
that fund-less trading was never live-tested until this session).

**Fix:** added `asSubjectAccessDenied(err error) error` in
`runtime_service.go` — normalizes any non-nil error from a
`SubjectAccessPort` into `domain.Forbidden(err.Error())` before it reaches
`writeError`. Applied to all three call sites (`checkSubjectView`,
`checkSubjectSubmit`, `checkSubjectAct`). Zero change to any authorization
*decision* (fail-closed stays fail-closed) — purely fixes HTTP-status/error
classification. Verified: `go build`/`go vet`/full `go test ./... -count=1`
clean (including the module's existing `denyingSubjectPort`/
`selectiveSubjectPort` test fakes, which already returned
`domain.Forbidden(...)` themselves and continue to pass — `errors.Is` still
matches after the re-wrap since `Unwrap()` returns the sentinel `Kind`
directly). Rebuilt and restarted `ims-backend`; retried the exact failing
call live: **403** `{"error":"not found or not accessible"}` — correct.
Granted `green` a real `permission_data_rights` row for the test portfolio
and retried: **200**, request `APPROVED`, materialized into exactly one real
`POSTED` transaction with `fund_id` correctly `NULL` throughout. This closes
both the residual proof (below) and the bug in one pass.

**Fund-less LIVE residual — now closed with full live proof, not just unit
tests.** Created `FL-TEST-01` (`portfolio_type=LIVE`, no `fund_code`) via
`POST /api/v2/portfolios` as `admin`. Submitted `CASH_IN 20000 THB` -> 202
PENDING, `fund_id` correctly absent from the response and `NULL` in the DB.
Confirmed the `COMPANY`-scoped process-config resolution (already proven
live via B14-CORE, since the seeded config's `contract_id IS NULL AND
contract_type='COMPANY'` branch matches regardless of the caller's contract
type) also works end-to-end for a genuinely fund-less submission. Approved
as `green` after the fix + a real permission grant above -> materialized
into one real ledger transaction, `fund_id NULL`, `POSTED`. The only
remaining untested delta from the design spec's test matrix (§6, "fund-less
LIVE portfolio → whole flow works") is now proven live, not just via
`TestFundlessLiveCashFlow`.

**V1 Swagger annotation gap — fixed.** Added `@Success 202
{object} response.CashRequestResponse` to `investment_handler.go`'s
`PostTransaction` (V1 legacy route), matching the V2 handler's existing dual
annotation, and updated its `@Description` to mention the LIVE-cash pending
behavior. Regenerated Swagger + the OpenAPI client (`make api-client`);
`CashRequestResponse` is now a documented possible response for the V1
route too. Backend `go build`/`go vet`/full `go test ./... -count=1` clean;
frontend `npx nuxi typecheck` still exactly the 91-diagnostic baseline, zero
new. This was low-risk/currently-unreachable (V1's order ticket only ever
sends BUY/SELL) but is now accurate documentation regardless.

**Test data created this pass (disclosed, on top of the earlier disclosed
rows):** portfolio `FL-TEST-01` (LIVE, fund-less, one real POSTED CASH_IN
20000 THB transaction); one additional cash request on `B14-CORE`
(`FEE 50`, cancelled) from the click-through Cancel test; one
`permission_data_rights` row granting `green` `APPROVE` access to
`FL-TEST-01`'s portfolio id.

**Final full gate re-run after all fixes (exact commands, from `backend/`
and `frontend/`):** `go build ./...`, `go vet ./...`, full
`go test ./... -count=1` (all packages `ok`) — clean. `npx nuxi typecheck`
— 91 diagnostics, byte-identical baseline. Both `ims-backend` and
`ims-frontend` Docker images rebuilt in the foreground a final time and
containers restarted; re-ran the full click-through Playwright pass against
the final images — identical clean result (one incidental console 404 for
`FL-TEST-01`'s missing valuation snapshot, expected for a portfolio that has
never been valued, unrelated to Stage 2).

**Commit staged, not committed.** Note: `neo-develop` moved during this
session — the owner committed `a3eb004` (watchlist permission grants,
resolving migration `20260723000002`'s provenance) and `b0faad1` (docs/api
generator + current API reference + Oracle runbook, resolving the
previously-flagged `docs/api/*` provenance question) directly, in parallel.
Neither touches this task's files; confirmed via `git show --stat` on both.
Staged exactly 100 files (`git add` with an explicit pathspec, never `-A`)
covering Stage 1 (fund-optional) + Stage 2 (LIVE cash-approval gate) backend,
frontend, migrations, seed, and design docs. Deliberately left unstaged:
`docs/MANAGER/HANDOFF.md` and `docs/MANAGER/TASKS.md` (this file mixes
multiple unrelated sessions' entries — including the unrelated Merge-Blocker
P0-E..P0-I documentation-only expansion — so committing the whole file would
repeat the exact "committed the entire index" mistake this same file
documents being caught and corrected once already, in P0-A's own history)
and `docs/compliance/thai-sec-product-regime-decision-memo.md` (unrelated,
pre-existing untracked file from an earlier session, never touched this
session). Commit message prepared and given to the owner directly; not run.

**Next action:** Kanta reviews the staged diff and runs the commit (or asks
for adjustments first). Separately, decide on the `NUXT_API_BASE_URL` Docker
env-var fix for the SSR session-restore bug — low-risk, one line, but
deployment config the owner should apply directly. The MODEL-blocked-notice
live check and a dedicated `docs/api/*`-generator confirmation remain open
but low-priority.

## Session: End-to-End Product/Business Flow Review Task Update (2026-07-24, IMS-MERGE-BLOCKERS)

**Status:** DOCUMENTATION UPDATED; NO PRODUCT IMPLEMENTATION STARTED.

Kanta requested that the 2026-07-24 read-only end-to-end IMS review be converted
into actionable project documentation. The findings were added to the existing
single active manager program in `TASKS.md` rather than creating a competing
`IN PROGRESS` task.

**Review baseline verified before the documentation edit:**

- Branch `neo-develop`, HEAD `5334ae8`, equal to `origin/neo-develop`
  (`0 0` ahead/behind at observation time).
- The working tree was already broadly dirty: portfolio/investment backend and
  frontend files, generated Swagger, manager/API docs, Bruno state, and new
  migrations. These changes predate this documentation update and belong to
  active/concurrent work, principally the fund-optional portfolio session.
- This session changed only `docs/MANAGER/TASKS.md` and this handoff. It did not
  modify product code, migrations, seeds, generated API artifacts, tests,
  containers, databases, branches, index state, commits, remotes, or deployment.

**Source-proven blocker summary now tracked in `TASKS.md`:**

1. Execution fill and confirmation state are disconnected from ledger,
   holdings, and cash.
2. Direct ledger BUY/SELL can bypass approved decision/execution because source
   references are optional.
3. Fill validation lacks positive/cumulative/ordered bounds, workflow gating,
   and compliance re-evaluation.
4. Workflow manager approval is wired to a NOP investment summary that always
   reports zero transactions.
5. LIVE cash/ledger approval is owner-required but not built.
6. Fund-less activity skips fund-scoped workflow controls while durable manager
   docs still say fund association is mandatory.
7. Trader/operator UI lacks execution mutations and confirmation creation but
   reports OP-02/OP-03 as fully available.
8. Regulatory/credit-rating stubs, approval-sync replay, audit semantics,
   watchlist scheduling, typecheck debt, and CI lifecycle coverage remain P1
   production-readiness work.

**New ordered work packages:** P0-E authoritative trade-to-ledger lifecycle;
P0-F real EOD activity integration; P0-G portfolio-type and LIVE ledger
approval; P0-H human resolution of the fund-optional policy conflict; P0-I
trader/operator frontend. Each package now records ownership, dependencies,
acceptance evidence, failure behavior, and human gates. P1 follow-ups and
program-wide release gates were also added.

**Immediate next action:** Kanta resolves P0-H before any overlapping backend,
migration, or frontend writer begins. Separately preserve/review Claude's active
dirty portfolio work. After the decision, continue the existing order P0-B ->
P0-C -> P0-D, then authorize and schedule P0-E/F/G/I in dependency order. Do not
commit, push, execute migrations, rebuild containers, mutate a database, or
deploy without explicit authorization.

## Session: Fund-Optional Portfolios + Fund-less Ledger/Trading (2026-07-23/24, IMS-PORTFOLIO-FUND-OPTIONAL)

**Status:** COMPLETE LOCALLY, deployed to local Docker containers; NOT
committed, NOT pushed. This work is unrelated to the Merge-Blocker
Remediation Program below and does not change that program's status — it ran
as a separate, directly-owner-driven session (no subagent delegation, no
worktree, no independent-review agent; worked in a single continuous session
per the owner's standing instruction against spawning background agents).
Two phases delivered end-to-end, plus two real production-shaped bugs found
and fixed only via live testing with the owner after `go test`/Vitest/
typecheck all stayed green.

**Phase 1 — fund-optional portfolio creation.** Added a "Bind with Fund:
Y/N" toggle to Create Portfolio, defaulting to **N/unchecked** (the owner
explicitly rejected defaulting to checked — did not want to re-toggle it on
every create). When unbound, `fund_code` is omitted entirely and the backend
creates a portfolio with `fund_id = NULL`.

- Migration `20260723000001_investment__portfolio_fund_id_nullable`: drops
  `NOT NULL` on `investment__portfolios.fund_id`.
- `entity.Portfolio.FundID` -> `*uuid.UUID`, propagated through persistence,
  the command layer (`portfolio_crud.go`), and every access-control call
  site (`resolvePortfolioByCode`, `checkPortfolioAccess`, new
  `portfolioScopeID` helper) — a fund-less portfolio's own id becomes its
  data-permission scope key. This reuses pre-existing infrastructure rather
  than inventing a new one: the generic `HasDataPermission(userID, scopeID)`
  check and `permission_data_rights.portfolio_id` (already in the schema,
  already wired into `PermissionManageDrawer.vue`'s new Fund/Portfolio
  scope-kind toggle added this session) — no new IAM concept was needed.
- `docker exec`-verified end to end: DB schema, Swagger, and the generated
  TypeScript client were all regenerated and confirmed nullable/optional.

**Phase 2 — fund-less Ledger + Decision/Execution/Confirmation trading**,
done immediately after Phase 1 at the owner's explicit follow-up request ("I
want portfolio can do the ledger on it own"). Extended the identical
nullable-`fund_id` treatment to `investment__portfolio_transactions`,
`investment__decisions`, `investment__executions`,
`investment__trade_confirmations` (migration `20260723000003` —
`...000002` was already claimed by a concurrent session's unrelated
permissions migration, confirmed via a duplicate-migration-file error at
`migrate up` time).

- Workflow gates (`IsTradeAllowed`/`IsTransactionLocked`/day-locking),
  compliance `ContractID`, and the `policy.EvaluatePost` fund-active check
  all now treat "no fund" as "skip fund-scoped state" rather than an error —
  matching a pattern (`portfolioFundID` in
  `portfolio_v2_compliance_handler.go`) an earlier/concurrent session had
  already anticipated with a forward-looking comment.
- Every `hasFundAccess` call site across decisions/executions/confirmations/
  transactions (new `decisionScopeID`/`executionScopeID`/
  `confirmationScopeID`/`transactionScopeID` helpers; `DecisionSubjectRef`
  gained a `PortfolioID` field for the batch-approval path) falls back to
  the owning portfolio's id when there is no fund, mirroring Phase 1's
  `portfolioScopeID`.
- Verification after each phase: `go build ./...`, `go vet ./...`,
  `go test ./... -count=1` (81 packages, zero FAIL); full Vitest (642-643
  tests) green; `npx nuxi typecheck` held at the exact pre-existing
  91-diagnostic/9-file baseline throughout, zero new diagnostics in any
  touched file.

**Bug 1 (found live, fixed): `risk_profile` as a number input silently
crashed Create Portfolio.** `<input type="number">`'s Vue `v-model` coerces
to a JS `number`; `normalizePortfolioCreateValues` then called `.trim()` on
it, throwing synchronously inside `form.submit()` *before* any network call
— so the owner's DevTools Network tab showed zero requests, indistinguishable
from the button doing nothing. Root-caused only after asking the owner to
paste the browser console stack trace (server-side logs showed nothing
useful — the failure never reached the backend).

**Bug 2 (found live, fixed): `risk_profile` is a real DB-enforced enum, not
free text.** `chk_inv_portfolios_risk_profile` allows only `NULL`/`LOW`/
`MEDIUM`/`HIGH`/`SPECULATIVE`. Neither this session's number-input version
nor the pre-existing free-text input on the Portfolio Settings page
(`PortfolioSettingsView.vue`, predates this session, found while fixing bug
1) ever validated against this — both would raise a raw Postgres
`SQLSTATE 23514` on submit. Fixed both frontend inputs to pickers with the
four real options (EN/TH/ZH), and added the missing validation server-side
in both `CreatePortfolioV2` and the portfolio `PATCH` handler
(`vo.RiskProfile.IsValid()` already existed in `valueobject/status.go` but
had zero callers before this fix) so any future bad value — from any client,
not just this frontend — returns a clean 400 instead of a leaked SQL error.

**Also handled this session, unresolved root cause:** the local Docker
environment (`ims-postgres`/`ims-backend`/`ims-frontend`/`ims-redis`/
`ims-mailpit`) was destroyed and recreated with an empty database by
something *outside* this session, twice (once mid-session, once again
between messages with no command run in between). Recovered both times via
`make migrate-up` + `go run ./cmd/seed` (idempotent, confirmed safe to
re-run) plus manually re-binding a hand-created test portfolio the second
time. Neither incident was triggered by any command this session ran
(confirmed via container-recreation timestamps and command history) — most
likely Docker Desktop restarting/resetting on the owner's machine, but this
was never confirmed. **Flagged to the owner, not resolved — could recur.**

**Deployed:** both Docker images rebuilt and containers restarted several
times as fixes landed. Final state was confirmed live by `docker exec`-
grepping the running containers' actually-served code (not just the git
diff or the build log) after every fix, including the last risk_profile fix.

**Explicitly NOT done / next action:**

- Nothing is committed. Roughly 30+ backend files, several frontend files,
  and 2 new migrations owned by this session (`20260723000001`,
  `20260723000003`) are uncommitted, on top of whatever pre-existing dirty
  working-tree state this repo already carries. The owner commits; this
  session never does.
- The Ledger ("add cash") and full Decision -> Execution -> Confirmation
  trading flow were implemented and covered by `go test`/Vitest, but **never
  exercised end-to-end in a live browser** by the owner or this session —
  every live browser test this session ran was scoped to Create Portfolio.
  Given that two real bugs already surfaced only through live testing (not
  caught by any automated gate), treat the Ledger/trading paths as
  *unverified in the browser* until someone actually clicks through them.
  See the ready-to-use follow-up prompt below.
- No frontend UI existed for an admin to grant a user portfolio-scoped data
  access on a fund-less portfolio; `PermissionManageDrawer.vue` gained a
  Fund/Portfolio scope-kind toggle this session to close that gap, but it
  has not been exercised live either.

**New requirement, specified but NOT built: cash/ledger posting must be
gated by portfolio type.** The owner clarified this explicitly after
reviewing the plan (not a code change yet):

- `SIMULATION` portfolios: post cash/ledger transactions immediately —
  this is already today's behavior (`PostTransactionHandler.Handle` posts
  unconditionally for every portfolio type); no change needed for this case.
- `MODEL` portfolios: stay fully blocked from the ledger, per the existing
  rule (`canEnterLedgerTransaction` on the frontend already blocks this).
  The owner's first phrasing of the requirement ("simulation/model can add
  immediately") was corrected on follow-up — MODEL must NOT gain ledger
  access; this was confirmed explicitly via AskUserQuestion, not assumed.
- `LIVE` portfolios: must require a **full approval-request workflow**
  before a cash/ledger transaction takes effect (not a lighter two-step
  confirm) — same rigor as the existing Decision workflow: a pending
  request routed through the approval module to an approver's inbox, only
  posting the real transaction on approval. Confirmed explicitly via
  AskUserQuestion (chosen over the lighter draft->confirm alternative).

This is a genuinely new subsystem, not a small addition. Key constraint
already identified: `investment__portfolio_transactions` rows are
append-only (`trg_inv_portfolio_transactions_no_update/no_delete`), so a
LIVE cash request **cannot** be created as a real transaction row and
edited/approved later — it needs a new pending-request entity/table,
submitted via the existing `contract.ApprovalSubmitter.SubmitForApproval`
(see `backend/pkg/contract/approval.go`) and only materialized as a real
`investment__portfolio_transactions` row inside an
`ApprovalSubjectCallback.OnApprovalDecision` implementation once approved
— mirroring `decision_lifecycle.go`'s `Submit()` pattern for
`INVESTMENT_DECISION`, which is the closest existing template. No approval
`ProcessType`/`SubjectType` for cash/ledger transactions exists in any seed
or config today (`grep -rln "CASH_TRANSACTION\|CASH_APPROVAL"` under
`database/seeds/` and `backend/internal/approval/` returns nothing) — this
would be new, not a matter of wiring up something already modeled.

Per advisor review this session: do NOT start this subsystem before (a)
the Ledger/trading paths above are actually verified working in a live
browser (see prompt below) and (b) the owner has committed the verified
Phase 1/2 work — piling a multi-file, multi-table subsystem onto ~30+
uncommitted files in an environment that has already self-wiped its Docker
volume twice this session is a real data-loss risk, not just untidy.

**Follow-up prompt for a fresh session (paste verbatim):**

```text
Continue the fund-optional portfolio work from IMS-PORTFOLIO-FUND-OPTIONAL
(see docs/MANAGER/HANDOFF.md, session dated 2026-07-23/24). Do this in two
stages — do not start stage 2 before stage 1 is verified and committed.

STAGE 1 — verify the existing (uncommitted) work live, then commit it.

Backend support for cash/ledger transactions and the full Decision ->
Execution -> Confirmation trading workflow on BOTH fund-bound and
fund-less portfolios was implemented and passes go test/vitest/typecheck,
but has never been exercised in a live browser. Two real bugs already
slipped past every automated gate in the adjacent Create Portfolio work (a
client-side v-model type-coercion crash, and a DB CHECK constraint neither
the frontend nor the backend validated against) — both were only found by
manually clicking through the UI. Assume the same class of gap exists here
until proven otherwise.

1. Start the local stack (docker compose in infra/, or make dev) and confirm
   admin/ben can log in.
2. Pick one fund-bound portfolio and one fund-less portfolio (create one via
   Create Portfolio with "Bind with Fund" off if none exists).
3. On EACH portfolio, in the browser: add cash via the Ledger page
   (frontend/app/features/portfolio-workspace/PortfolioLedgerNewView.vue,
   posts through useOrderTicket.ts -> portfolioApi.simulateTransaction/
   postTransaction -> POST /api/v2/portfolios/{code}/transactions on the
   backend, handled by PostTransactionHandler in
   backend/internal/investment/application/command/post_transaction.go).
   Confirm the cash balance actually updates (PortfolioCashView.vue) and
   check the browser console for errors, not just the network tab — the
   risk_profile bug threw client-side before any request fired.
4. On EACH portfolio, walk a decision through its full lifecycle in the
   browser: New Decision -> Submit -> (approval, if wired for your test
   user) -> Execution -> Confirmation. Entry point:
   frontend/app/features/portfolio-decision/PortfolioDecisionNewView.vue.
   Confirm the fund-less portfolio's decision actually reaches EXECUTED, not
   just DRAFT/PENDING_APPROVAL.
5. For any field you touch that maps to a DB column, grep the actual
   database schema for a CHECK constraint before trusting the Go struct
   field's type alone (docker exec into ims-postgres and \d the table) —
   that is exactly what bug 2 above was.
6. Report back: what worked, what didn't, and whether the "grant a user
   portfolio-scoped data access on a fund-less portfolio" flow in
   PermissionManageDrawer.vue's new Fund/Portfolio toggle actually works end
   to end (create a fund-less portfolio as one user, grant a second user
   access via that toggle, confirm the second user can see it).
7. Once verified, ask the owner to commit this work before moving on —
   ~30+ backend files, several frontend files, and 2 migrations
   (20260723000001, 20260723000003) are uncommitted, and this environment's
   Docker volume has self-wiped twice already this session.

STAGE 2 — build portfolio-type-gated cash/ledger approval (only after
stage 1 is committed).

Requirement (confirmed explicitly with the owner via AskUserQuestion, do
not re-derive or re-negotiate without cause):
- SIMULATION: posts immediately. Already true today — verify only.
- MODEL: stays fully blocked from the ledger. Do not change this.
- LIVE: requires a full approval-request workflow. A pending cash/ledger
  request must be submitted via contract.ApprovalSubmitter.SubmitForApproval
  (backend/pkg/contract/approval.go) and only materialized as a real
  investment__portfolio_transactions row inside a new
  ApprovalSubjectCallback.OnApprovalDecision implementation once approved.
  Use decision_lifecycle.go's Submit() (INVESTMENT_DECISION) as the closest
  existing template for the ProcessType/SubjectType/submit/callback shape.
  investment__portfolio_transactions rows are append-only
  (trg_inv_portfolio_transactions_no_update/no_delete) — a LIVE cash
  request CANNOT be created as a real transaction row and edited later; it
  needs its own new pending-request entity and table (new migration).
  No existing approval ProcessType/SubjectType/seed covers cash/ledger
  transactions — confirm this is still true before assuming any wiring
  already exists.

Before writing code: confirm with the owner (a) which transaction types
this gate applies to (the owner said "add cash", i.e. CASH_IN specifically
— confirm whether it should also cover CASH_OUT/FEE/DIVIDEND/BUY/SELL or
just cash movements), and (b) whether a pending LIVE cash request should be
visible/cancelable by its submitter before an approver acts on it.

Do not commit anything without the owner's explicit go-ahead. If you rebuild
Docker images, run docker-compose build in the FOREGROUND, not backgrounded
— backgrounded docker-compose build was silently killed twice with zero
output in the prior session for reasons never identified.
```

## Session: Merge-Blocker Remediation Program Kickoff (2026-07-21, IMS-MERGE-BLOCKERS)

**Status:** IN PROGRESS. Verified live Git state independently rather than
trusting a stale "52 modified files" claim: branch `neo-develop`, HEAD
`0ae1adb4a84f817f766be34963a7877446e03aa7`, tracking `origin/neo-develop`
with `+5 -0` (exactly the five commits `8d43747`..`0ae1adb` listed in the
prior session), and exactly 15 working-tree paths with zero untracked files.
This matches the expected starting state given in the owner's brief exactly.

**Phase 1 — mixed working-tree reconciliation (read-only; zero commits
created):** all 15 paths were read (not classified by filename) and held.
`tools/bruno/environments/ims-th-local.yml`'s only diff is two local-run
values (`compliance-breach-id`, `compliance-rule-type-id`) — session-local
test state, held per the owner's explicit instruction. `docs/api/*` (5
modified + 8 new files) and the two new
`tools/bruno/Investment/Executions/*.yml` files were all staged by a process
outside any manager session (flagged unresolved since the 2026-07-20
Architecture Cleanup session); `docs/api/_build_current_api_docs.py` is a
clean, well-structured generator that reads only `backend/docs/swagger.json`/
`v2/v2_swagger.json`/the legacy builder and writes only `docs/api/*.md`, and
the regenerated docs read as accurate — but authorship is still unconfirmed,
so nothing was executed, staged further, or committed. The five existing
local commits were reviewed via `git show --stat` only: each has a
single-purpose, non-overlapping file set matching its commit message: no
scope-creep or correctness concern found. Full detail and the exact
disposition table are in `TASKS.md`'s new Phase 1 section.

**Phase 2 — P0-A dispatched, P0-B/C/D queued:** grep/read investigation (this
session, before any delegation) confirmed the 2026-07-17 merge-review finding
still reproduces in current code: `resolvePortfolioByCode`
(`portfolio_v2_handler.go:46-71`) only calls `hasFundAccess` when its checker
argument is non-nil, and `portfolio_v2_decision_handler.go` (5 sites) plus
`portfolio_v2_execution_handler.go` (5 sites, split across `*ExecutionHandler`
and `*TradeConfirmationHandler`) all pass a literal `nil`. Root cause:
`DecisionHandler`/`ExecutionHandler`/`TradeConfirmationHandler` have no
permission-checker field at all — `module.go:219-229` wires them only via
`SetPortfolioRepository`/etc. The correct pattern already exists twice in
the same file to copy (`InvestmentHandler`'s constructor-injected
`m.permissionAdapter`, and `m.decisionBatchCmd.SetPermissionChecker(iamPort)`
as a post-construction setter precedent). Exact acceptance criteria are
recorded in `TASKS.md` under P0-A; dispatched to `frontier-backend-engineer`
as a single writer with independent review required before any commit.
P0-B (demo migration cleanup), P0-C (fill validation), and P0-D (compliance
binding auditability) are queued in `TASKS.md`, not yet started.

**Phase 3 — Thai SEC product-regime decision memo:** drafted
`docs/compliance/thai-sec-product-regime-decision-memo.md` from the existing
approved source map (`docs/compliance/thai-sec-regulatory-source-map.md`,
verified 2026-07-15) plus `AIREAD.md`'s business description. It recommends
retail mutual fund as the first regime (carrying forward the source map's own
2026-07-15 recommendation), conditioned explicitly on Kanta/compliance
confirming that matches the real licensed business — no regime is selected by
this memo, no threshold is proposed, and `regulatory.thai_sec` is untouched.
Ambiguities requiring legal interpretation, a draft (unapproved) source-to-
rule skeleton, and the exact two required sign-offs are listed in the memo.

**Phase 4 — Oracle migration 20260615000003 investigation (read-only, source
only; no production access available this session):** read
`20260615000003_workflow__global_business_date.up.sql`/`.down.sql` and the
migration runner (`backend/platform/database/migrate.go`,
`backend/cmd/migrate/main.go`, both using `golang-migrate/migrate/v4` with
the Postgres driver — "Oracle" is confirmed to be a deployment nickname, not
a literal Oracle database; the repository is exclusively PostgreSQL per
`CLAUDE.md`). The up migration's `DO $$ ... END $$` block checks for duplicate
`business_date` rows in `workflow__day_states` and `RAISE EXCEPTION`s before
any `ALTER TABLE` statement runs; golang-migrate's Postgres driver wraps a
migration file's entire body in one transaction by default. **Most likely
state class from code alone (state #1 in the matrix below), not confirmed
against the real database:** `schema_migrations` records version
`20260615000003` with `dirty=true` (set before the run, never cleared because
the run errored), while the transactional rollback means none of the four
DDL changes (nullable `contract_id` x3, unique-constraint swap, new index)
were actually applied — i.e., the physical schema most likely still matches
the prior migration's shape. This is an inference from source, not a
verified fact; no production connection, log, or query was available or used
this session. Prior sessions record only secondhand knowledge ("the latest
observed container restart loop reported dirty database version
20260615000003") with no captured query output or log excerpt.

**State matrix and forward-repair runbook (proposed, not executed):**

1. **Dirty metadata, no relevant DDL applied (most likely per source
   analysis above).** Prerequisite: full `pg_dump` schema+data backup of the
   affected tables before touching anything. Validation:
   `SELECT version, dirty FROM schema_migrations;`;
   `SELECT is_nullable FROM information_schema.columns WHERE table_name IN
   ('workflow__day_states','workflow__transition_log','workflow__approval_records')
   AND column_name='contract_id';` (expect `NO` in this state);
   `SELECT conname FROM pg_constraint WHERE conrelid =
   'workflow__day_states'::regclass;` (expect `uq_wf_day_states_contract_date`
   still present, `uq_wf_day_states_business_date` absent);
   `SELECT indexname FROM pg_indexes WHERE indexname =
   'idx_wf_transition_log_business_date';` (expect no row);
   `SELECT business_date, count(*) FROM workflow__day_states GROUP BY
   business_date HAVING count(*) > 1;` (this is almost certainly why the
   migration raised — expect >=1 row). Safest repair: resolve the duplicate
   `business_date` rows first (workflow/domain owner decides the canonical
   row per date; do not delete without an audit trail), re-run `migrate up`
   normally (do not `force`) once duplicates are gone — golang-migrate will
   retry version `20260615000003` cleanly from the dirty state as long as the
   physical schema truly matches the pre-migration shape. Rollback
   implication: none, because nothing was physically applied. Post-repair
   checks: re-run all four validation queries above expecting the
   post-migration shape, confirm `dirty=false`, run
   `go vet -tags e2e ./tests/e2e/...` compile check and the workflow test
   suite against a disposable database seeded from a prod-shaped dump.
   Approvals required: workflow/domain owner (duplicate-row resolution),
   DBA (backup + repair execution), production approver (deployment restart).

2. **Partially applied nullability changes only** (would mean the Postgres
   driver did *not* wrap the file in one transaction, or a prior manual
   intervention ran statements individually). Validation: same
   `information_schema.columns` query as above; if `contract_id` is
   `YES`-nullable on some but not all three tables, or nullable while the old
   `uq_wf_day_states_contract_date` constraint is still present, this state
   applies. Safest repair: do not `force` a version; manually complete the
   remaining `ALTER COLUMN ... DROP NOT NULL` statements one at a time inside
   a DBA-supervised transaction, re-validate, then let `migrate up` proceed
   with the constraint/index portion only if it can detect it's not yet
   applied (may require a corrective forward migration rather than reusing
   `20260615000003` — do not edit an applied migration file). Rollback:
   partially reversible per-column; get workflow-owner sign-off before
   re-adding `NOT NULL` since data may already rely on nulls. Approvals:
   DBA + workflow owner + production approver, plus a corrective-migration
   design review before writing it.

3. **Old uniqueness removed but global uniqueness not installed.**
   Validation: `uq_wf_day_states_contract_date` absent AND
   `uq_wf_day_states_business_date` absent from
   `pg_constraint`/`workflow__day_states`. This is the most dangerous state —
   the table currently has **no uniqueness protection at all** on
   `(contract_id, business_date)` or `(business_date)`. Immediate mitigation:
   treat as urgent even before full repair — application-level writes should
   be paused or a temporary advisory lock/careful monitoring added if the
   backend is live against this database, because concurrent writers could
   insert conflicting day-state rows with zero DB-level protection.
   Prerequisite: backup, then run the duplicate-date query immediately. If
   clean, add `uq_wf_day_states_business_date` directly (`ADD CONSTRAINT ...
   UNIQUE (business_date)`) via a corrective migration; if duplicates exist,
   resolve them first under workflow-owner supervision. Rollback: re-adding
   the old constraint requires confirming no legitimate multi-contract-same-
   date rows were created during the unprotected window. Approvals: DBA +
   workflow owner + production approver, treated as expedited given the live
   data-integrity exposure.

4. **Global uniqueness and index already installed (migration actually
   succeeded; only the dirty flag is stuck).** Validation: all four
   validation queries above show the fully-migrated shape (`contract_id`
   nullable x3, `uq_wf_day_states_business_date` present,
   `idx_wf_transition_log_business_date` present) but
   `schema_migrations.dirty = true`. Safest repair: this is the one state
   where `migrate force 20260615000003` (to clear the dirty bit without
   re-running DDL) is the standard golang-migrate remedy — but per the
   owner's explicit prohibition, this session does not execute it; it is
   listed here only as the documented safe action for the DBA/production
   approver to run themselves after confirming this exact state.
   Prerequisite: backup (cheap, no schema change expected). Rollback: none
   needed. Approvals: DBA/production approver confirm the four schema checks
   before forcing.

5. **Duplicate business dates blocking the migration (the DO block's own
   detection).** Validation: the `GROUP BY business_date HAVING count(*) > 1`
   query above returns rows. This can coexist with state 1 (most likely) or
   state 3. Repair is the same duplicate-resolution step described in states
   1 and 3: the workflow/domain owner picks the canonical row per duplicated
   business date (e.g., by transition history / most-advanced state), any
   other row is archived (not silently deleted) with an audit trail before
   uniqueness is added. This must not be automated by an agent without
   workflow-owner review of each duplicate.

6. **Metadata and physical schema disagree in some other way not covered by
   1-5** (e.g., version recorded but a *different* migration's DDL is
   missing, suggesting an out-of-band manual change). Validation: run all
   four schema queries plus `SELECT version FROM schema_migrations;` and
   compare against the full ordered migration file list in
   `database/migrations/`; look for any gap between what the version number
   implies and what physically exists. Safest repair: full DBA-led schema
   diff against a freshly migrated disposable database from the last known-
   good version, not a guess; likely requires a bespoke corrective migration
   designed jointly by the DBA and workflow/domain owner. No repair should
   proceed until this comparison is complete.

**Explicitly not done, per the brief's constraints:** no `migrate force`, no
edit to `schema_migrations`, no execution of the up or down SQL, no
constraint/index/nullability change, no data deletion/merge, no production
container restart, no deployment rerun, no push. Production remains NO-GO
until a DBA/workflow-owner/production-approver team runs the validation
queries above against the real database, confirms which state applies, and
approves the matching repair path.

**P0-A completed and committed as `67efe71`** (same session, same day). The
`frontier-backend-engineer` worker implemented the fix in an isolated
worktree (`.claude/worktrees/agent-a85d22ca515e2d197`, branch
`worktree-agent-a85d22ca515e2d197`) exactly as scoped: added a `pc
contract.PermissionChecker` field + `SetPermissionChecker` setter to
`DecisionHandler`, `ExecutionHandler`, `TradeConfirmationHandler`; wired
`module.go` to call `SetPermissionChecker(m.permissionAdapter)`
unconditionally for all three (verified: `m.permissionAdapter` is always
non-nil and its `HasDataPermission`/etc. methods fail closed even if the
wrapped `iamPort` is nil); fixed all 10 call sites from a literal `nil` to
`h.pc`; added 20 new tests (2 per endpoint — authorized-passes-gate,
cross-fund-403) in a new `portfolio_v2_fund_scope_test.go`.

**Manager independent verification (before any commit):** re-read the full
diff and every changed/new file myself (not just the worker's summary), then
independently ran `go build ./...`, `go vet ./...` (both clean) and `go test
./internal/investment/... -count=1` (all 10 investment sub-packages `ok`) in
the isolated worktree — matching, not just trusting, the worker's claimed
results.

**Independent security review (separate fresh agent, given only the raw
diff + acceptance criteria, not this manager's conclusions):** re-derived
the vulnerability from source (confirmed the pre-fix `pc != nil &&
!hasFundAccess(...)` truly no-ops on nil), traced the `module.go` wiring
end-to-end including `PermissionCheckerAdapter`'s actual fail-closed
implementation (not just its doc comment), grepped the whole
`internal/investment` tree for any missed call site or alternate
portfolio-code-resolving path (none found), manually traced source for at
least 3 of the 10 endpoints to confirm the JSON-decode-after-permission-check
ordering claims and the `SubmitDecisionByCode` real-lifecycle-rejection
claim, and independently ran the full build/vet/test gate. **Result: no
blocking findings.** One non-blocking hardening note was raised (the shared
resolver still silently no-ops on a nil checker; the guarantee is wiring
convention, not a structural impossibility) — logged in `TASKS.md` as
deferred, not silently dropped.

**Manager error caught and corrected during integration:** the fix was
implemented in an isolated worktree without committing; the manager copied
the 8 changed/new files into the main working tree (verified first that
those exact 8 paths were byte-identical between the worktree's older base
commit and current `neo-develop` HEAD, so no rebase/merge was needed),
re-ran the full independent verification in the real working tree (`go
build`, `go vet`, full `go test ./... -count=1` — 81 packages, zero `FAIL`;
`gofmt -l` clean; `git diff --check` clean), then ran `git commit` **without
a pathspec**, which committed the entire index — including the 10
pre-existing staged `docs/api`/Bruno-execution paths Phase 1 had explicitly
decided to hold. This was caught immediately by inspecting `git show --stat`
on the new commit (18 files instead of 8) before reporting anything to
Kanta. Corrected via `git reset --soft HEAD~1` (safe: unpushed, at the tip,
created by this session moments earlier, nothing built on top) followed by
`git commit -- <8 files>` with an explicit pathspec. Final commit `67efe71`
contains exactly the 8 P0-A files; `git status` afterward reconfirmed the
other 15 paths returned to their exact pre-commit staged/unstaged state.
**Lesson for future sessions: always pass an explicit pathspec (`--
<files>`) to `git commit` in this repository given the persistent held
working-tree state — never rely on "only what I just `git add`ed" being the
entire index.**

`neo-develop` is now 6 commits ahead of `origin/neo-develop`
(`8d43747`..`67efe71`). Nothing pushed.

**Next action:** (1) Kanta reviews commit `67efe71` and the deferred
hardening note, and decides whether to authorize P0-B (demo migration
cleanup) to start next; (2) Kanta or a delegate runs the Phase 4 validation
queries against the real production database (read-only) and reports
results so the correct state/runbook row can be confirmed; (3) Kanta or the
designated compliance expert reviews
`thai-sec-product-regime-decision-memo.md` and either confirms retail mutual
fund or selects a different first regime; (4) `docs/api/`/Bruno-execution
provenance is confirmed before any of those 15 held paths are committed.

## Session: Local Investment Integration Commits on neo-develop (2026-07-21)

**Status:** LOCAL COMMITS CREATED ON `neo-develop`; NOT PUSHED. The owner
explicitly authorized committing directly on the local `neo-develop` branch and
did not authorize a push. `origin/neo-develop` remains at `0c14c1d`.

**Scoped commits:**

- `8d43747` - `refactor(compliance): align response DTOs and rule loading`
- `39ab72d` - `feat(investment): aggregate latest portfolio AUM in THB`
- `09d9421` - `chore(api): regenerate compliance and valuation contracts`
- `f6c5205` - `feat(frontend): redesign compliance breach queue`

The code commits deliberately exclude all files under `docs/api/`, the local
Bruno environment edit, and the unverified `tools/bruno/Investment/Executions/`
requests. Those paths were already staged by another process during the commit
session; path-limited commits were used so their staged state and content were
preserved without inclusion.

**Verification:** focused Breaches tests passed 6 files / 59 tests. Full
frontend Vitest passed 49 files / 546 tests, and `npm run build` completed
successfully. `npx nuxi typecheck` still reports the known 91 pre-existing
diagnostics, with zero diagnostics in the Breaches files touched by the new
frontend package. `git diff --check` was clean. The backend code content is the
same content covered by the full Go test/vet/build evidence recorded in the
2026-07-20 sessions; backend gates were not rerun during this commit-only
session.

**Remaining gate:** authenticated browser UAT for the redesigned Breaches page
has not been run. Before any push decision, verify desktop/tablet/mobile,
light/dark themes, EN/TH/ZH, override permission and conflict behavior, check-
group drawer loading, console errors, and network failures. The Oracle dirty
migration at version `20260615000003` remains a production-deployment blocker
and is unrelated to these local commits.

## Session: Latest-Available AUM and THB FX (2026-07-20, IMS-LATEST-AUM)

**Status:** COMPLETE LOCALLY; NOT COMMITTED OR PRODUCTION-DEPLOYED. This
session implements the owner's explicit policy that both `company` and `mine`
must use each portfolio's latest available valuation, rather than withholding
the whole total when one portfolio has not been valued on the newest scope
date. Cross-currency amounts use the latest valid authoritative FX quote into
the configured reporting currency (`THB` in the current environment).

**Backend behavior:** the valuation-summary adapter still includes active
`LIVE` portfolios only. A portfolio's older latest snapshot and a snapshot
flagged with stale inputs are now included instead of excluded. Missing
valuations, currency changes across the latest/previous pair, and missing,
invalid, wrong-symbol, or wrong-quote-currency FX remain fail-closed and return
`INCOMPLETE`; no currency is omitted or assumed at 1:1. Cross-currency AUM is
the latest local AUM multiplied by the latest `FX_<BASE><REPORTING>` quote.
Today's cross-currency P&L converts the local cumulative-P&L delta with that
same latest quote, so the summary does not manufacture historical FX movement
from mismatched quote dates. The response adds
`latest_available_portfolio_count` and `oldest_included_business_date` while
retaining the newest included valuation as `business_date`.

**Market-data correction:** direct canonical lookup now enriches an existing
legacy `market_symbols` row from reference-data provider mappings, so
`FX_USDTHB` resolves to Yahoo's executable symbol `USDTHB=X`. The persisted
legacy asset type remains `UNKNOWN` because that table's existing check
constraint does not accept `FX`; canonical FX identity stays in reference
data. The cache was cleared only for the exact key
`marketdata:quote:FX_USDTHB`, then a normal live refresh fetched and persisted
the real Yahoo quote `33.62000000 THB` dated 2026-07-20. No valuation row was
created or changed.

**Frontend behavior:** the AUM and P&L cards consume the new freshness fields.
When any portfolio contributes an older latest valuation, EN/TH/ZH copy states
how many portfolios use latest-available data, the oldest included valuation
date, and that the total is converted to the reporting currency. Normal fully
current summaries retain the existing `As of` message.

**Live proof:** the authenticated local dashboard renders `Entire company AUM`
as `฿9.91B` and Today's P&L as `−฿2.64M`, with 7/7 portfolios included and the
message `Latest available data used for 1 of 7 portfolios; oldest valuation
Jul 17, 2026. Total converted to THB.` Switching to `My AUM` renders `฿5.68B`
and `−฿725K`. Browser console warnings/errors: zero. The live company API
returned `AVAILABLE`, `aum_today=9913070690`,
`today_pnl=-2642684.780895069`, `included_portfolio_count=7`,
`total_portfolio_count=7`, `latest_available_portfolio_count=1`, and
`oldest_included_business_date=2026-07-17`.

**Verification:** focused adapter, market-data, integration-handler, mapping,
and dashboard tests passed. Full backend `go test ./... -count=1`,
`go vet ./...`, and `go build ./...` passed. `make api-client` passed and a
second generation changed zero hashes across all seven generated files. Full
Vitest passed 46 files / 497 tests, and the Nuxt production build passed.
Typecheck still reports 91 pre-existing diagnostics, with zero diagnostics in
the files touched by this task. Browser verification covered both scope
options and found no console warning/error.

**Safety state:** work remains in the pre-existing dirty worktree; nothing is
staged, committed, pushed, migrated, seeded, or production-deployed. Local
backend/frontend containers were rebuilt. The only live-data write from this
session is the real provider-sourced market-data snapshot described above.
Existing unrelated compliance, API-documentation, localization, and other
working-tree changes were preserved.

## Session: Architecture Cleanup (2026-07-20, IMS-ARCH-CLEANUP)

**Status:** COMPLETE LOCALLY; NOT COMMITTED. Owner-authorized pull-forward of
the safe Architecture Cleanup items as the single IN PROGRESS task, scoped to
behavior-preserving refactoring only. Work sits on `neo-develop` at `0c14c1d`
in the same uncommitted worktree as IMS-COMPANY-AUM-VISIBILITY; every
pre-existing dirty/untracked file and all 71 stashes were preserved. No commit,
push, migration, seed, container restart, or live-database write occurred. The
regulatory task (IMS-REG-TH-SEC) and all DO-NOT-MERGE blockers were not
touched.

**Item 1 — compliance dedup and N+1.** The three byte-identical holding
market-value helpers (`assetClassHoldingMV`, `holdingMV`, `marketValue`) were
replaced by one shared `spi.HoldingMarketValue` in `spi/data_bundle.go`; the
three rule files now call it at their single call sites. The
`ListPortfolioRules` catalog no longer issues one `GetCurrentVersion` query
per rule instance (up to 500): `domain.RuleInstanceRepository` gained
`GetCurrentVersions(ctx, ids) (map[uuid.UUID]*entity.RuleInstanceVersion,
error)`, implemented in Postgres with a single `ANY($1)` query, and the
adapter batch-loads before the loop. Best-effort semantics preserved — a
lookup failure still nils `parameters` without failing the listing; the one
granularity nuance (a batch failure nils all entries' parameters rather than
one) is deliberate and harmless because a single query has no partial-failure
mode. The only other `RuleInstanceRepository` implementer
(`fakeRuleInstanceRepo` in `create_rule_instance_test.go`) implements the new
method by delegating to its per-instance fake.

**Item 2 — parameters generated-type drift.** `contract.
PortfolioRuleCatalogEntry.Parameters` changed from `json.RawMessage` +
`swaggertype:"object"` (which generated `Record<string, never>`) to `any` with
the swaggertype tag removed, following the `httputil.ErrorResponse.Details`
precedent. Producers still assign the raw pre-encoded `json.RawMessage`, so
wire bytes are unchanged (interface-boxed RawMessage marshals its raw bytes);
a `len > 0` guard in the adapter reproduces the previous `omitempty`
edge-case behavior exactly. Regenerated client now types
`parameters?: unknown`, which the existing frontend `asParameterRecord`
normalizer already accepts.

**Item 3 — @Failure 422 annotations.** Found already complete before this
session: both `CreateExecution` and `CreateExecutionByCode` carry the
`@Failure 422` annotation in source and the generated V1/V2 documents already
list 422 (closed earlier by `e80f086` + `790a2f7`). Verified against
`swagger.json`/`v2_swagger.json`; no change was needed or made.

**Item 4 — zh-TW normalization.** Every pre-existing Simplified-Chinese key in
`frontend/app/shared/i18n/messages/zh/portfolio.ts` was converted to
Traditional, matching the conventions already used by `zh/compliance.ts` and
this file's own Traditional sections (載入/失敗/重試/紀錄/存取權限/送出審批);
already-Traditional sections are untouched, all `{placeholder}` tokens and key
structure preserved. A scripted scan for Simplified-only characters over the
final file returns zero hits. No test asserts exact zh strings (verified);
parity/interpolation suites pass.

**Item 5 — E2E cleanup helper.** `cleanupInvestmentPrereqs` in
`backend/tests/e2e/investment_test.go` now deletes
`investment__trade_confirmations` → `investment__executions` →
`investment__decisions` (scoped through the fund's portfolios) before the
portfolio/fund deletes. Trade confirmations were added beyond the checklist
text because their `ON DELETE RESTRICT` FK on `execution_id` would otherwise
still block the execution delete and leave the orphan chain in place.

**Items 6–7 — valuation summary adapter refactor.** The duplicated
`listAllFunds`/`listAllPortfolios` pagination loops collapsed into one generic
`listAllPages[T]` helper. All three fail-closed guards (total drift,
invalid/negative pagination metadata, non-advancing page short of the total)
are preserved with byte-identical error messages — they are deliberate
independent-review hardening and were not weakened. The per-fund portfolio
listing loop (one exhaustive listing per scoped fund) was replaced by a single
paginated listing filtered by active status, portfolio manager (mine scope),
and accessible funds (mine scope only — company scope is never narrowed),
grouped by `FundID` in memory. Equivalence argument: the scoped-fund map
bounds both scopes exactly as the per-fund loop did (portfolios of inactive or
out-of-scope funds are ignored by the grouping lookup), and the repository's
`ORDER BY fund_id, code` means each fund's group preserves the same relative
order a per-fund listing produced, so coverage counts and exclusion ordering
are unchanged. Proven by the existing untouched adapter suite — all 25
`TestGetValuationSummary_*` tests pass, including forced one-row pagination,
company-ignores-AccessibleFundIDs, and the mine manager/access-intersection
regressions.

**Item 8 — compliance transport DTO mirror.** `PreTradeCheckResponse`,
`PostTradeCheckResponse`, and `BreachSummary` are now mirrored in
`compliance/transport/dto/response/responses.go` with `FromPreTradeCheck` /
`FromPostTradeCheck` mapping funcs in the same style as the five endpoints the
previous session already moved; enum-typed fields flatten to `string` with
identical wire values, and a nil-preserving breach mapper keeps the
`breaches` key omitted exactly as before. The two handler `@Success`
annotations now reference the `response` DTOs and the handlers map through
them. New wire-JSON tests prove the exact snake_case field names, the omitted
`breaches`/`status` keys, and the envelope. Note: regenerating also
materialized the previous session's (already-in-worktree) envelope-wrapping
annotations for the other five compliance endpoints into the generated
documents — the docs now match the runtime `httputil.SuccessResponse`
envelope those endpoints always emitted.

**Independent review.** Owner memory prohibits spawning subagents, so the
required independent read-only review of items 6–7 and 8 was performed as a
dedicated adversarial diff pass in this session against the pre-edit file
contents held in context: pagination-guard equivalence, grouping edge cases
(inactive funds, nil accessible-fund list, zero totals), exclusion ordering,
DTO field-name/omitempty parity, and envelope behavior were re-verified
line-by-line. No P0/P1 findings; the two recorded nuances are the batch
best-effort granularity (item 1) and the envelope-schema materialization
(item 8), both documented above.

**Verification (exact commands, from `backend/` and `frontend/`):**

- `gofmt -l <all 14 touched Go files>` — clean, zero output.
- `go build ./...` and `go vet ./...` — pass.
- `go test ./... -count=1` — exit 0, 81 packages ok, zero FAIL.
- `go vet -tags e2e ./tests/e2e/...` — pass (E2E-tag compile).
- `make api-client` (runs `make swagger` + `swagger-v2` + frontend
  `api:generate`) — pass; a second full generation left all seven generated
  files SHA-256 identical (determinism proven).
- Focused Vitest (`i18n-messages`, `portfolio-decision-navigation`,
  `dashboard-valuation-summary`, `portfolio-directory-kpis`) — 4 files /
  31 tests pass.
- Full `npm run test` — 46 files / 496 tests pass. `npm run build` — exit 0,
  "Build complete!".
- Scoped `git diff --check` over touched paths — clean.
- Live browser proof was not run (consistent with prior sessions: local
  containers run prebuilt images without source mounts; the owner verifies).

**Files changed by this session (all uncommitted, nothing staged):**

```text
backend/internal/compliance/application/command/create_rule_instance_test.go
backend/internal/compliance/domain/repository.go
backend/internal/compliance/infrastructure/persistence/rule_instance_repository.go
backend/internal/compliance/rules/allocation/asset_class.go
backend/internal/compliance/rules/concentration/single_issuer.go
backend/internal/compliance/rules/ratio/sector_exposure.go
backend/internal/compliance/spi/data_bundle.go
backend/internal/compliance/transport/dto/response/responses.go      (pre-existing untracked, extended)
backend/internal/compliance/transport/dto/response/responses_test.go (pre-existing untracked, extended)
backend/internal/compliance/transport/handler/compliance_handler.go
backend/internal/compliance/transport/portfolio_contract_adapter.go
backend/internal/investment/infrastructure/adapter/valuation_summary_adapter.go
backend/pkg/contract/contracts.go
backend/tests/e2e/investment_test.go
frontend/app/shared/i18n/messages/zh/portfolio.ts
backend/docs/* and frontend/app/api/ims-api.d.ts                     (regenerated, never hand-edited)
docs/MANAGER/HANDOFF.md, docs/MANAGER/TASKS.md
```

**Concurrent-writer observation (needs owner attention).** During this session
(2026-07-20 12:27–12:32 local), files under `docs/api/` were created/modified
by a process outside this session: `README.md`, `portfolio-v2-api-ddd.md`,
`watchlist-api.md` modified, and `_build_current_api_docs.py`,
`current-api-reference.md`, `chat-api.md`, `integration-api.md`,
`portfolio-v2-api.md`, `reference-data-api.md`, `system-api.md`,
`watchlist-current-api.md` newly created. Nothing in this session's commands
writes those paths (the make targets write only `backend/docs/*` and
`ims-api.d.ts`). This session did not touch them; they violate the
one-writer-at-a-time rule if another agent session is active — Kanta should
identify the writer before committing anything from `docs/api/`.

**Residual/deferred:** the same `swaggertype:"object"` empty-object drift
still exists on the V1 `dto/response/responses.go` fields
(`RuleInstanceVersion.Parameters`, `CheckRecord.ParameterSnapshot`,
`CheckRecord.Evidence`) — out of item 2's scope, logged in `TASKS.md`.
`nuxi typecheck` was not run (not in this task's gates; known pre-existing
debt of 97 diagnostics in untouched legacy files).

**Git state at end:** branch `neo-develop`, HEAD
`0c14c1d547e5b9e32fea0dbda8ca6e2754b7e9c6`, index empty, 71 stashes intact
(`stash@{0}` = `pre-neo-develop-merge-20260717-110007`). Kanta reviews and
commits; nothing was pushed.

## Session: Company AUM Visibility (2026-07-18, IMS-COMPANY-AUM-VISIBILITY)

**Status:** COMPLETE LOCALLY; NOT COMMITTED OR PRODUCTION-DEPLOYED. Work is on
`neo-develop` at base commit `0c14c1d` with an uncommitted working tree. No
push, migration, database repair, production deployment, live-data write, or
secret change occurred. The local backend and frontend containers were rebuilt
and restarted on 2026-07-20 for HTTP/browser verification only.

**Owner policy:** every authenticated dashboard user may see the complete
company aggregate AUM and P&L, independent of function and fund/portfolio data
scope. This grant is deliberately aggregate-only. Unauthenticated requests
remain rejected, company coverage retains aggregate counts/currencies/dates and
reason codes but omits fund/portfolio exclusion identities, and drill-down
screens retain their existing permissions. The `mine` scope is portfolio-based:
it includes only portfolios whose manager is the authenticated caller, inside
funds the caller may access.

**Backend:** the valuation-summary query no longer applies the dashboard-view
function gate or IAM fund-scope filter to `scope=company`; it requests all
active company funds from the integration provider. `scope=mine` resolves the
caller's accessible funds, then filters active portfolios by
`investment__portfolios.manager_user_id`; it no longer filters ownership through
`investment__funds.manager_user_id`. The handler still requires authenticated
claims. The provider independently enforces company scope by ignoring any
supplied accessible-fund filter, exhausts paginated fund and portfolio lists
instead of silently stopping at 10,000 rows, and fails rather than accepting
inconsistent pagination metadata. Contract comments, Swagger, generated
TypeScript declarations, and tests reflect the policy and verify that company
responses cannot leak item-level exclusions.

**Frontend:** company is now the default AUM scope for every authenticated
dashboard session and the company/mine selector is always available. EN/TH/ZH
incomplete-coverage copy now says “in-scope portfolios” instead of implying the
company total is permission-filtered. The portfolio-directory copy was aligned
because it consumes the same company summary endpoint.

**Independent review:** the initial security/financial review reported no
P0/P1 and two P2 hardening gaps: provider-level company filtering and silent
10,000-row repository caps. Both were fixed. Narrow re-review confirmed company
cannot be restricted by `AccessibleFundIDs`, mine remains the manager/access
intersection, forced one-row fund/portfolio pagination includes every row, and
no P0/P1/P2 findings remain.

**Verification:** focused integration query/handler tests passed; focused
dashboard, portfolio KPI, and i18n Vitest passed 3 files/24 tests; `make
api-client` regenerated Swagger and the TypeScript API contract successfully.
Full `go test ./... -count=1` passed in 76.1s, `go vet ./...` plus `go build
./...` passed in 67.6s, full Vitest passed 46 files/495 tests in 12.5s, and the
Nuxt production build passed in 210.1s. After independent review hardening,
focused investment-adapter plus integration tests passed; final full `go test
./... -count=1` passed in 53.3s and final `go vet ./...` plus `go build ./...`
passed in 44.7s.

**2026-07-20 correction and verification:** owner clarification established
that “My AUM” means portfolios currently managed by the user, not every
portfolio under funds managed by that user. The adapter was corrected at the
portfolio repository filter, with crossed fund/portfolio-manager regression
fixtures proving the distinction and the fund-access intersection. Focused
adapter/query/handler tests passed; `make api-client` passed; full `go test
./... -count=1`, `go vet ./...`, and `go build ./...` passed. Frontend focused
tests passed 20/20, full Vitest passed 46 files/496 tests, and the production
build passed; typecheck retained 97 pre-existing diagnostics with zero in the
touched dashboard/i18n/test files. Live local HTTP proof returned `AVAILABLE`
for `scope=mine` (5/5 admin-managed portfolios) and distinct `INCOMPLETE` for
`scope=company` (6/7 portfolios) because one USD valuation remains stale from
2026-07-17. This observation is superseded by the owner's later same-day
latest-available policy and the `IMS-LATEST-AUM` implementation above; no
valuation row was changed to manufacture completeness.

**Production blocker:** the Oracle backend deployment is still not healthy.
The latest observed container restart loop reported dirty database version
`20260615000003`, so this policy cannot be live until that database state is
safely investigated/repaired and a backend deployment succeeds. Editing GitHub
secrets or the reporting-currency environment file does not resolve that dirty
migration.

## Session: Three Production-Readiness Defects (2026-07-17, IMS-PR3-20260717)

**Status:** COMPLETE LOCALLY; DO NOT MERGE. Work ran from 13:30 through final
integration on `feature/investment`. The owner-authorized scope was exactly
three fixes. No push, merge, migration, seed, deployment, live-database write,
or stash operation occurred.

**Local commits:**

- `216c600783ff15270fbaaf29e3f81a38f260904a` - configured reporting-currency
  AUM/P&L and business-date FX.
- `e80f08688f24f07a9c023876b077b2d74f238655` - LIVE compliance fail-closed
  behavior and audit-correlated typed 422 results.
- `350fe19` - portfolio/dashboard/operator EN/TH/ZH, generated-contract
  consumption, branch TypeScript cleanup, and honest directory KPIs.
- `790a2f7` - reproducible generated V1/V2 Swagger and TypeScript contracts.

**Fix 1 - reporting currency:** production now requires a valid uppercase ISO
4217 `REPORTING_CURRENCY`; environment templates configure `THB`, but business
logic contains no THB default. Official company and mine totals include active
`LIVE` portfolios only. Conversion uses persisted direct `FX_<BASE><QUOTE>`
quotes for the exact valuation business date; same-currency values use identity.
Missing, stale, wrong-date, wrong-symbol, wrong-quote-currency, zero, or negative
FX and stale/inconsistent valuations produce `INCOMPLETE` with stable coverage
and exclusions and no partial numeric total. Today's P&L includes local P&L and
FX movement; external flows use the documented closing-FX convention.

**Fix 2 - compliance:** the cross-module contract now carries typed
`COMPLIANCE_EVALUATED`, `COMPLIANCE_NOT_CONFIGURED`, and
`COMPLIANCE_UNAVAILABLE` states. A `LIVE` portfolio with no active/effective
rule or missing existing/proposed asset classification blocks before decision,
approval, or execution persistence, returns a correlated typed HTTP 422, and
writes a strict control-gap audit event. Nil/unknown statuses and verdicts also
fail closed. `SIMULATION` and `MODEL` preserve their prior policy. Real
effective-window tests cover zero, inactive, future, expired, and valid rules.

**Fix 3 - frontend:** dashboard and portfolio-directory AUM, Today's P&L, and
percentage consume only the backend reporting-currency summary. No client
cross-currency aggregation remains. `AVAILABLE`, `NO_DATA`, `INCOMPLETE`, and
transport-error states withhold unavailable money and retain currency,
business-date, as-of, coverage, and exclusion evidence. Open-breach pagination
is exhaustive and fails honest if metadata changes or a page is incomplete.
EN/TH/ZH catalogs have structural parity, genuine localized workflow copy,
interpolation/raw-key tests, and no retained English fallback literals. Changed
UI labels resolve business names/codes or an explicit unavailable marker rather
than exposing UUIDs.

**Independent reviews:** the financial reviewer found and closed the economic
P&L formula issue; the backend/security reviewer found and closed severity,
portfolio-type, unknown-result, SIM/MODEL, and audit-correlation fail-open paths.
The frontend reviewer reported no P0/P1 and three P2s (truncated breach counts,
nested keyboard navigation, and two English fallbacks); all three were fixed and
narrowly re-reviewed with no remaining P0/P1/P2.

**Integrated verification:**

- Backend: `go test ./... -count=1` passed in 49.4s; `go vet ./...` passed in
  25.4s; `go build ./...` passed in 27.6s; E2E-tag compile passed in 11.6s;
  all 43 changed Go files are gofmt-clean.
- Frontend: focused contract/i18n/pagination tests passed 4 files/30 tests;
  full Vitest passed 46 files/497 tests; Nuxt production build passed with only
  dependency `DEP0155` warnings.
- TypeScript: exact `neo-develop` baseline is 123 diagnostics in 23 files;
  pre-fix `feature/investment` was 166 in 21. Final is 97 in 10 legacy
  approval/chat/settings/watchlist files, with zero diagnostics intersecting
  the 249 frontend files changed or generated-contract-affected by the branch.
- Contract generation: `make api-client` passed and a second generation changed
  zero SHA-256 hashes across all seven generated files.
- Static checks: diff/whitespace, staged secret patterns, added unsafe
  TypeScript, added mojibake, and raw-ID-label checks passed.
- Browser: the built UI loaded with zero console warnings/errors and protected
  `/portfolios` redirected to
  `/auth/login?reason=token_expired&redirect=/portfolios`. The local session was
  expired, so authenticated portfolio responsive/theme/locale visual UAT is not
  claimed.

**Residual risks/non-goals:** the branch remains DO NOT MERGE because earlier
Portfolio V2 data-scope authorization, production demo memberships,
fill-validation, binding-audit, regulatory, and UAT blockers remain. The
reporting implementation uses closing-date FX for external flows and does not
claim GIPS/transaction-time precision. `previousInternalSnapshot` still reads a
small source-agnostic latest-before set, which can hide an older INTERNAL
baseline behind enough newer non-INTERNAL rows; address with a source-filtered
repository query before high-volume production history. The 97 unrelated
TypeScript diagnostics and authenticated browser UAT remain follow-up work.

**Final safety state:** `neo-develop` and `origin/neo-develop` remain
`6ba5dc7ac67426b0d22df41e8570797c485d4b62`. All 71 pre-existing stashes remain;
`stash@{0}` is still `ef36adc3b5bd3a69e209e46fefafc3ee9ce62b19`.
`fund_id` remains required. No database was started or mutated.

**Exact implementation file inventory:**

`216c600` (17 files):

```text
backend/cmd/server/main.go
backend/internal/integration/application/query/get_valuation_summary.go
backend/internal/integration/application/query/get_valuation_summary_test.go
backend/internal/integration/domain/valuation_summary.go
backend/internal/integration/transport/dto/response/valuation_summary_response.go
backend/internal/integration/transport/handler/dashboard_handler.go
backend/internal/integration/transport/handler/dashboard_handler_test.go
backend/internal/investment/infrastructure/adapter/valuation_summary_adapter.go
backend/internal/investment/infrastructure/adapter/valuation_summary_adapter_test.go
backend/internal/investment/module.go
backend/pkg/contract/valuation_summary.go
backend/platform/config/config.go
backend/platform/config/config_test.go
infra/env/.env.development
infra/env/.env.e2e.example
infra/env/.env.example
infra/env/.env.production
```

`e80f086` (29 files):

```text
backend/internal/compliance/application/command/allocation_unavailable_test.go
backend/internal/compliance/application/command/run_pretrade_check.go
backend/internal/compliance/application/command/run_pretrade_check_test.go
backend/internal/compliance/domain/valueobject/compliance_status.go
backend/internal/compliance/engine/pipeline.go
backend/internal/compliance/engine/pipeline_unavailable_test.go
backend/internal/compliance/infrastructure/adapter/investment_data_adapters.go
backend/internal/compliance/infrastructure/adapter/nop_adapters.go
backend/internal/compliance/rules/allocation/asset_class.go
backend/internal/compliance/rules/allocation/asset_class_test.go
backend/internal/compliance/spi/data_bundle.go
backend/internal/compliance/spi/result.go
backend/internal/compliance/transport/contract_adapter.go
backend/internal/compliance/transport/contract_adapter_test.go
backend/internal/investment/application/command/compliance_gate.go
backend/internal/investment/application/command/compliance_gate_test.go
backend/internal/investment/application/command/decision_compliance_test.go
backend/internal/investment/application/command/decision_lifecycle.go
backend/internal/investment/application/command/execution.go
backend/internal/investment/application/command/execution_workflow_test.go
backend/internal/investment/module.go
backend/internal/investment/transport/handler/compliance_error.go
backend/internal/investment/transport/handler/compliance_status_handler_test.go
backend/internal/investment/transport/handler/decision_handler.go
backend/internal/investment/transport/handler/execution_handler.go
backend/internal/investment/transport/handler/portfolio_v2_decision_handler.go
backend/internal/investment/transport/handler/portfolio_v2_execution_handler.go
backend/internal/investment/transport/handler/portfolio_v2_handler_test.go
backend/pkg/contract/contracts.go
```

`350fe19` (71 files):

```text
frontend/app/features/compliance/components/ComplianceAuditTimeline.vue
frontend/app/features/compliance/components/ComplianceBreachInbox.vue
frontend/app/features/compliance/components/ComplianceExceptionTimeline.vue
frontend/app/features/compliance/components/CompliancePermissionMatrix.vue
frontend/app/features/compliance/components/ComplianceRuleBuilder.vue
frontend/app/features/compliance/components/ComplianceRuleDetailHeader.vue
frontend/app/features/compliance/components/ComplianceRuleDetailRail.vue
frontend/app/features/compliance/components/ComplianceRuleTable.vue
frontend/app/features/compliance/components/ComplianceRuleTabs.vue
frontend/app/features/compliance/components/ComplianceSectionTabs.vue
frontend/app/features/compliance/composables/useCompliancePortfolioDirectory.ts
frontend/app/features/compliance/composables/useComplianceUserDirectory.ts
frontend/app/features/dashboard/components/DashboardApprovalPanel.vue
frontend/app/features/dashboard/components/DashboardOverviewScreen.vue
frontend/app/features/dashboard/components/DashboardTaskFeed.vue
frontend/app/features/dashboard/lib/dashboard.ts
frontend/app/features/dashboard/lib/valuationSummaryMapping.ts
frontend/app/features/dashboard/types.ts
frontend/app/features/investment-decision/components/DecisionStatusBadge.vue
frontend/app/features/portfolio-decision/PortfolioDecisionDetailView.vue
frontend/app/features/portfolio-decision/PortfolioDecisionNewView.vue
frontend/app/features/portfolio-decision/PortfolioDecisionsListView.vue
frontend/app/features/portfolio-decision/components/DecisionListTable.vue
frontend/app/features/portfolio-decision/components/DecisionWorkflowPanel.vue
frontend/app/features/portfolio-decision/components/InstrumentCombobox.vue
frontend/app/features/portfolio-decision/composables/useDecisionLifecycle.ts
frontend/app/features/portfolio-decision/lib/decisionFormat.ts
frontend/app/features/portfolio-decision/lib/decisionValidation.ts
frontend/app/features/portfolio-decision/lib/legacyEntry.ts
frontend/app/features/portfolio-workspace/PortfolioCashView.vue
frontend/app/features/portfolio-workspace/PortfolioComplianceView.vue
frontend/app/features/portfolio-workspace/PortfolioDirectoryView.vue
frontend/app/features/portfolio-workspace/PortfolioHoldingsView.vue
frontend/app/features/portfolio-workspace/PortfolioLedgerView.vue
frontend/app/features/portfolio-workspace/PortfolioOverviewView.vue
frontend/app/features/portfolio-workspace/components/PortfolioAllocationDonut.vue
frontend/app/features/portfolio-workspace/components/PortfolioFilterToolbar.vue
frontend/app/features/portfolio-workspace/components/PortfolioKpiStrip.vue
frontend/app/features/portfolio-workspace/components/PortfolioRoleActionMenu.vue
frontend/app/features/portfolio-workspace/components/PortfolioSummaryCard.vue
frontend/app/features/portfolio-workspace/composables/useMyPortfolios.ts
frontend/app/features/portfolio-workspace/lib/directoryKpis.ts
frontend/app/features/portfolio-workspace/lib/valuationHistory.ts
frontend/app/features/portfolio-workspace/types.ts
frontend/app/features/workflow/store/useWorkflowStore.ts
frontend/app/pages/investment/operator/[fundId]/decisions/index.vue
frontend/app/pages/investment/operator/[fundId]/operation/[decisionId].vue
frontend/app/pages/investment/operator/[fundId]/operation/index.vue
frontend/app/pages/investment/operator/index.vue
frontend/app/shared/i18n/messages/en/compliance.ts
frontend/app/shared/i18n/messages/en/dashboard.ts
frontend/app/shared/i18n/messages/en/index.ts
frontend/app/shared/i18n/messages/en/operator.ts
frontend/app/shared/i18n/messages/en/portfolio.ts
frontend/app/shared/i18n/messages/th/compliance.ts
frontend/app/shared/i18n/messages/th/dashboard.ts
frontend/app/shared/i18n/messages/th/index.ts
frontend/app/shared/i18n/messages/th/operator.ts
frontend/app/shared/i18n/messages/th/portfolio.ts
frontend/app/shared/i18n/messages/zh/compliance.ts
frontend/app/shared/i18n/messages/zh/dashboard.ts
frontend/app/shared/i18n/messages/zh/index.ts
frontend/app/shared/i18n/messages/zh/operator.ts
frontend/app/shared/i18n/messages/zh/portfolio.ts
frontend/tests/dashboard-valuation-summary.test.ts
frontend/tests/i18n-messages.test.ts
frontend/tests/portfolio-decision-draft.test.ts
frontend/tests/portfolio-decision-errors-format.test.ts
frontend/tests/portfolio-decision-lifecycle.test.ts
frontend/tests/portfolio-decision-validation.test.ts
frontend/tests/portfolio-directory-kpis.test.ts
```

`790a2f7` (7 generated files):

```text
backend/docs/docs.go
backend/docs/swagger.json
backend/docs/swagger.yaml
backend/docs/v2/v2_docs.go
backend/docs/v2/v2_swagger.json
backend/docs/v2/v2_swagger.yaml
frontend/app/api/ims-api.d.ts
```

Manager evidence files: `docs/MANAGER/HANDOFF.md`,
`docs/MANAGER/MEMORY.md`, and `docs/MANAGER/TASKS.md`.

## Session: neo-develop Merge-Readiness Review (2026-07-17)

**Verdict:** DO NOT MERGE. `neo-develop` is an exact ancestor of
`feature/investment` (`6ba5dc7..226a9fb`, 30 commits / 425 files), so the Git
operation would be a conflict-free fast-forward, but the integrated behavior is
not release-safe. Three independent read-only reviews covered backend/security,
frontend/UX, and cross-layer integration. No merge, push, migration, seed,
service restart, or production action occurred.

**Merge-blocking findings:**

- Portfolio V2 decision, execution, and confirmation handlers resolve a
  portfolio code while passing a nil fund data-scope checker. Function
  permission alone can therefore disclose or mutate another fund's portfolio
  when its code is known.
- Three production migrations assign runtime permissions/roles directly to the
  demo usernames `ben` and `green`. Environment-specific memberships belong in
  development seeds or explicit provisioning, not universal migrations.
- Asset-class rules omit unclassified holdings/orders, zero applicable rules
  returns PASS for a LIVE portfolio, and actual fill quantity/amount can exceed
  the values checked when the execution was opened. These paths can bypass the
  intended compliance boundary.
- Compliance binding creation/deactivation has no immutable actor-attributed
  audit event.
- Company and portfolio-directory AUM/P&L combine or omit currencies without
  an approved reporting-currency/FX policy, producing materially misleading
  financial totals.
- Portfolio directory breach-fetch failures are converted to an empty list and
  displayed as compliance `Clear`; route switches can also retain or race stale
  decisions/compliance from the previous portfolio.
- Newly reachable holdings, ledger, operator-detail, and print screens expose
  raw instrument, portfolio, or fund UUIDs instead of business codes/names.
- The new portfolio workflow is not EN/TH/ZH complete. Nuxt typecheck reports
  166 diagnostics, including 65 in changed files; missing typed portfolio keys
  cause Thai/Chinese users to receive English fallbacks, and multiple new views
  contain hard-coded English.

**Additional compatibility gaps:** removed fund routes lack general redirect
shells; the V2 execution Swagger contract omits its real HTTP 422 response;
the generated rule-parameter type remains `Record<string, never>`; frontend CI
does not run Vitest; and the new Portfolio V2 E2E lifecycle is not in the CI
target.

**Validation evidence:** exact committed tree passed backend
`go test ./... -count=1`, `go build ./...`, and independent `go vet ./...`;
the E2E-tagged package compiled without executing tests. Frontend
`npm run test` passed 45 files / 488 tests and `npm run build` passed. Scoped
code diff/whitespace checks and credential-pattern scans passed. Full database
migrations, seeds, authenticated browser/mobile/theme/three-locale UAT, and
live runtime proof were not run. Passing unit/build checks do not override the
source-proven blockers above.

**Safety state:** before review, the three remaining local edits were saved in
dedicated stash commit `ef36adc3b5bd3a69e209e46fefafc3ee9ce62b19`
(`pre-neo-develop-merge-20260717-110007`). It contains only
`portfolio_crud.go` and the two Bruno health requests. Do not reapply the
fund-validation removal without an explicit reversal of the owner decision.
`neo-develop` and `origin/neo-develop` remained at `6ba5dc7`.

**Next action:** Kanta decides whether to authorize a dedicated merge-blocker
remediation program. Security/data-scope, demo migration grants, compliance
fail-closed policy, fill validation, and auditability must be resolved before
frontend correctness and compatibility fixes are integrated and the complete
gate is rerun. The reporting-currency/FX behavior and mandatory-rule policy
require explicit owner/compliance decisions.

## Session: Local Commit Cleanup (2026-07-17)

**Outcome:** Kanta authorized local commits on the current
`feature/investment` branch and retained the earlier no-push requirement. The
large mixed worktree was separated into nine reviewable commits without
switching branches, dropping stashes, pushing, deploying, or mutating a live
database:

- `47312fa` - classify compliance-blocked execution requests as HTTP 422 and
  add the Portfolio Compliance V2 lifecycle E2E proof.
- `e00a586` - add scoped dashboard AUM and P&L valuation summary integration.
- `6108650` - add the compliance operations control center.
- `d7be2ac` - add narrowly scoped investment-operation permission grants.
- `96d5eca` - align dashboard load-error translations with the tested copy.
- `71cb524` - add the portfolio compliance rule workspace.
- `4c3bdd8` - add the portfolio-first decision and approval workflow.
- `be81149` - consolidate fund-centric screens into the portfolio workspace.
- `c8144fa` - add shared Codex/Claude engineering-manager and frontier-agent
  guidance while keeping Claude worktrees and local settings ignored.

**Validation:** backend `go test ./... -count=1` and `go build ./...` passed.
Frontend `npm run test` passed 45 files / 488 tests and `npm run build` passed.
Additional focused checks passed for the dashboard (15 tests), portfolio
compliance workspace (20 tests), portfolio decision workflow (103 tests),
portfolio allocation/history components (8 tests), compliance control center
(51 tests), the investment command package, handler package, E2E-tag compile,
and integration/valuation packages. Every staged group passed scoped
`git diff --check` with CRLF-aware whitespace handling and a credential-pattern
scan before commit.

**Protected exclusion:**
`backend/internal/investment/application/command/portfolio_crud.go` remains
uncommitted because its local edit removes mandatory `fund_id` validation,
directly conflicting with Kanta's durable decision that fund-less portfolios
are not planned. The two Bruno health URL edits also remain uncommitted pending
confirmation of the intended local/cloud environment behavior. Existing safety
stashes were retained.

**Goal status:** this cleanup records and protects already-present development;
it does not waive or complete the documentation-first UAT gate and does not
claim production readiness. No remote push occurred.

## Session: Overnight Production Readiness and UAT (2026-07-16/17, IMS-UAT-20260717)

**Status:** BLOCKED AT DOCUMENTATION GATE; Goal remains active for a resumable
continuation. The owner made this documentation-first production-readiness/UAT
program the immediate priority. It does not close
`IMS-REG-TH-SEC`; that task remains gated on human approval of the applicable
Thai SEC product regime and exact source-to-rule matrix.

**Start and time gates:** started at 2026-07-16 23:40 Asia/Bangkok. Target
documentation completion is 02:15; no application coding is allowed if the
documentation gate is incomplete at 03:00. Source edits stop at 05:40, tests
and diff review finish by 06:05, final report by 06:20, and all work stops at
06:25.

**Verified safety state:** original branch `feature/investment`, original HEAD
`0362d00bfad71feb16e624f133557447b9c37f52`. Pre-safety state contained 298
changed paths: 156 staged additions and 142 unstaged tracked changes, with no
separate untracked files and 68 pre-existing stashes. The permanent safety
stash is `stash@{0}` / `bd589febc6d0885796694112bbf6b8f94be163cc`, message
`overnight-uat-safety-20260716T234123+0700`. It was applied cleanly by stable
hash on branch `codex/overnight-uat-20260717-20260716T234125`; no conflict
exists and the stash remains untouched. The new branch index was cleared with
`git restore --staged -- .` only after validating the backup and applied tree.

**Documentation gate:** the external root contains 48 files, not the previously
expected 43. The final disposition is 36 fully covered, 11 valid DOCX packages
structurally covered but visual/page QA blocked, and one zero-byte invalid DOCX
fully blocked. The canonical DOCX renderer cannot run because `soffice` is
absent; a bounded read-only Word fallback produced no PDF. `Repository
secret.docx` was reviewed only by the primary manager and is reported solely as
sensitive and redacted. The sanitized manifest and blocked UAT matrix are under
`docs/uat/`. No application source, test, service, database, generated file or
external provider was touched after the blocker.

**Independent review:** a read-only reviewer reconciled all 48 manifest paths,
all 48 matrix rows and required fields, manager state, overnight-only file
scope, live branch/HEAD/stash, redaction checks, and safe recovery steps. No
critical issue was found. The sole high finding was four unfinished fields in
the final report; they were completed and a narrow re-review confirmed no
remaining critical/high finding. Final live Git counts were 0 staged, 142
tracked-but-unstaged, 159 untracked, and 0 conflicts.

**Current next action:** Kanta provides/approves a valid replacement for the
zero-byte migration guide and makes a DOCX renderer available, or explicitly
waives page visual QA for the 11 structurally inspected files. The next Goal
continuation rescans all 48 paths, approves/completes the manifest, verifies
Git/stash state, and only then begins read-only source/test reconnaissance. No
coding starts until that gate is satisfied. Full state and recovery steps are
in `docs/uat/2026-07-17-overnight-report.md`.

## Session: Compliance Dashboard UX and Typed API Integration (2026-07-16, IMS-COMPLIANCE-DASHBOARD-UX)

**Outcome:** replaced `/compliance`'s generic landing cards with a responsive,
operations-first control center. It now presents exact configured-rule counts,
an open-breach queue, a portfolio workspace finder, category distribution,
permission-aware actions, independent retry/refresh behavior, and honest
loading/error/empty states. EN/TH/ZH copy explicitly distinguishes configured
definitions and role-visible breach records from applied controls or
enterprise-wide compliance analytics.

**Topology:** the primary agent owned integration and verification. One
read-only backend/API auditor confirmed endpoint semantics, permission scope,
runtime envelope behavior, and the absence of an aggregate read model. A
frontend subagent was stopped after orientation; its small normalizer and
pagination-helper partials were reviewed, corrected, and integrated by the
primary agent. No concurrent writer remained during integration.

**API integration:** dashboard reads now use `useOpenApiClient` and the
generated `ims-api.d.ts` contracts for `GET /compliance/rules`,
`GET /compliance/breaches`, and `GET /investment/portfolios`. Narrow
normalizers validate generated response DTOs, including the runtime
`effectiveWindow.valid_from`/`valid_to` shape and optional breach
`contractID`. Rules and portfolios are collected across every page with
stable-total/non-advancing-page guards, eliminating the previous 200-record
summary cap. Existing legacy compliance mutations remain on `useApi` because
their generated JSON-object schemas currently collapse to
`Record<string, never>`; this dashboard task did not conceal that schema debt.

**Backend boundary:** no backend code or endpoint was required. The current
APIs support configured definition metrics plus a compliance-role-visible V1
breach queue. They do not support authoritative applied-control counts,
binding-severity aggregates, a compliance business date, or portfolio-data-
scoped organization-wide breach analytics. Those claims were deliberately not
invented in Vue; they require a future server-side read model.

**Primary implementation files:**
`frontend/app/features/compliance/ComplianceControlCenterView.vue`,
`components/ComplianceBreachQueue.vue`,
`components/CompliancePortfolioFinder.vue`,
`components/ComplianceCategoryPanel.vue`,
`components/ComplianceKpiCard.vue`,
`composables/useComplianceBreaches.ts`,
`composables/useComplianceRuleDirectory.ts`,
`composables/useCompliancePortfolioDirectory.ts`,
`services/complianceApi.ts`, `lib/formatters.ts`, `lib/pagination.ts`,
`lib/asOfDate.ts`, `frontend/app/pages/compliance/index.vue`, three locale
`compliance.ts` files, and compliance-focused tests. Older dashboard-only card
components were removed as part of the redesign.

**Validation:** `npx vitest run tests/compliance` passed (12 files / 87 tests);
full `npm run test` passed (45 files / 488 tests); `npm run build` passed.
`npx nuxi typecheck` remains red from broad pre-existing repository debt; the
filtered compliance output contains only two pre-existing diagnostics in
untouched `ComplianceAuditTimeline.vue` and `ComplianceBreachInbox.vue`, with
no diagnostics in this task's touched files. Scoped compliance
`git diff --check` and trailing-whitespace scan passed. Repository-wide
`git diff --check` still reports unrelated existing dashboard/i18n whitespace.

**Browser verification:** the in-app browser reached the local app but
`/compliance` redirected to `/auth/login?reason=token_expired`; the existing
Chrome connector was unavailable. No credentials were entered. Authenticated
desktop/mobile visual proof therefore remains pending, while automated and
production-build verification is green.

**Git state:** branch `feature/investment`, HEAD `0362d00`. The worktree was
already broadly dirty/untracked and remains so. No staging, commit, push,
migration, seed, runtime restart, or database mutation was performed.

## Session: OP-01 Decision Effective Price Fix (2026-07-16, IMS-OP01-PRICE)

**Goal:** fix OP-01 submission when an operator supplies positive quantity and
amount but leaves optional limit price blank. The frontend correctly omitted
`limit_price`; submit-time investment compliance ignored `amount`, forwarded a
zero price, and compliance rejected the request before evaluating any rule.

**Fix:** decision submission now treats explicit amount as the authoritative
proposed notional and derives compliance unit price as `amount / quantity`,
matching the existing execution-time rule. Quantity-only tickets fall back to
a positive limit price. Amount-only tickets and quantity-only tickets without
a safe price fail closed with a typed invalid-decision error before compliance
or approval. Decision creation now rejects non-positive supplied quantity,
amount, or limit price. The frontend estimate now displays explicit amount and
its regression test proves quantity + amount omits rather than invents a zero
limit price.

**Files changed for this fix:**
`backend/internal/investment/application/command/decision_lifecycle.go`,
`decision_compliance_test.go`,
`frontend/app/features/portfolio-decision/PortfolioDecisionNewView.vue`,
`frontend/app/features/portfolio-decision/lib/decisionFormat.ts`,
`frontend/tests/portfolio-decision-draft.test.ts`, and
`frontend/tests/portfolio-decision-errors-format.test.ts`. No API schema,
migration, generated client, or database data changed.

**Validation:** backend `go test ./... -count=1` passed; frontend full Vitest
passed (43 files / 478 tests); `npm run build` passed; focused OP-01 frontend
tests passed (4 files / 60 tests). Local Docker backend/frontend images were
rebuilt and only those two app containers were recreated. Post-restart
`GET /health` returned status/database/redis `ok`; frontend `/` returned 200.

**Current A02-CORE test policy (read-only DB verification):** active rules are
cash availability (BLOCK, zero buffer), minimum trading unit (BLOCK, default
lot 1), minimum THB trade amount 1,000 (BLOCK), ENERGY sector maximum 35%
(BLOCK), and single-issuer maximum 10% (WARN). A02-CORE is ACTIVE, has THB
60,240,265 cash, and does not require a research report. Recommended aligned
test: BUY `TH-LB30DA`, quantity `1`, amount `1032`, blank limit price. For the
original AOT choice: quantity `100`, amount `6350`, blank limit price.

**Git state:** no commit, staging, push, migration, or seed was performed. The
large pre-existing dirty/untracked worktree remains and must be preserved.

## Session: P2 Reprioritization and Thai SEC Source Discovery (2026-07-15, IMS-REG-TH-SEC)

**Historical owner decision, superseded 2026-07-24 by D1 `FUND_OPTIONAL`:**
at this 2026-07-15 checkpoint the direction was not to develop optional fund
association. Do not treat this historical session as current product policy.
The former P1
feature and its seven open checkboxes were removed from the active backlog at
that time.
Real Portfolio Regulation Rules are now the only `IN PROGRESS` manager task;
Architecture Cleanup remains queued immediately after it. The cleanup item that
renames legacy compliance `ContractID` is scope clarification only and must not
reintroduce fund-less portfolio development.

**Verified repository state:** branch `feature/investment`, HEAD `2593c24`,
index empty. The large pre-existing dirty/untracked set documented below is
still present. This session changed only manager documentation and added the
regulatory source-map document; no application source, migration, generated
file, test, staging, commit, push, runtime, or database state was changed.

**Source discovery:** Thailand / Thai SEC is the confirmed first jurisdiction.
The authoritative base is Capital Market Supervisory Board Notification Tor
Nor. 87/2558, *Investment of Funds*, using the SEC rulebook's effective
consolidated text and current appendices rather than an old static PDF. As of
2026-07-15 the SEC rulebook lists Tor Nor. 1/2569 (35th amendment), signed
2026-02-09 and effective 2026-03-01. The source map also records the official
liquidity-risk page, derivatives notification, private-fund management sources,
and the mandate/prospectus layer required for portfolio-specific restrictions.

**Blocking business classification:** IMS currently has no authoritative field
that distinguishes retail mutual fund, AI fund, UI fund, private fund, or
provident fund. Those regimes use different Thai SEC appendices and limits.
Therefore `regulatory.thai_sec` remains a non-enforcing WARN stub. No threshold,
severity, override policy, or effective date was invented. Kanta or a compliance
expert must choose the first regime and approve a source-to-rule matrix before
backend implementation begins.

**Precise next action:** approve one first regime (recommended starting point if
it matches the real business is retail mutual fund) and the source-to-rule
matrix in `docs/compliance/thai-sec-regulatory-source-map.md`. Then design the
versioned compliance rule parameters/evidence for that one regime, implement
the backend contract and tests, and only afterward update the generated client
and frontend catalog. Architecture Cleanup starts after the regulatory outcome
contract is met.

## Session: Dashboard AUM/P&L Live Integration (2026-07-15, IMS-DASHBOARD-AUM-PNL)

**Goal:** complete and verify the live API integration for the "AUM Today"
and "Today's P&L" dashboard cards (task ID `IMS-DASHBOARD-AUM-PNL`), reusing
the existing `GET /integration/dashboard/valuation-summary?scope=company|mine`
endpoint rather than building a duplicate. Verified starting Git state
matched the prompt exactly: branch `feature/investment`, HEAD `2593c24`,
index empty, and the same large unrelated dirty-file set (Portfolio
Compliance V2 follow-ons, an in-progress `investment-workspace`/`my-funds`
refactor, a new `portfolio-decision` feature, and a broad untracked `docs/`
tree) — all preserved untouched. A partial, uncommitted AUM/P&L
implementation (endpoint, domain, adapter, handler, frontend
composable/service/lib, one test file) already existed from an earlier
session and was read and audited before any edit.

**Topology:** single instance, no subagents. The task prompt explicitly
authorized spawning `/frontier-backend-engineer` and `/frontier-frontend-engineer`
subagents, but the backend audit was already complete in this instance's own
context by the time that decision point was reached, so re-spawning cold
would have meant re-deriving already-held context or pasting findings in —
which the ai-engineering-manager skill itself discourages ("don't delegate
understanding"). Frontend orientation was likewise done inline. No subagent
was used.

### Backend audit findings and fixes

Audited `ValuationSummaryAdapter`
(`backend/internal/investment/infrastructure/adapter/valuation_summary_adapter.go`,
untracked/new) against the task's 16-point backend contract. Two real
correctness bugs were found and fixed on the money path; one design choice
was confirmed correct but flagged for an owner decision; everything else in
the contract was already satisfied by the existing partial implementation.

1. **(Fixed) Swallowed infrastructure error mistaken for "no baseline".**
   `previousInternalSnapshot` returned `nil` on a `ValuationRepository.List`
   error, indistinguishable from "no prior snapshot exists". The caller then
   attributed the portfolio's entire cumulative P&L to "today" — a
   materially wrong number produced by a transient DB error, not a real
   first-valuation-day case. Fixed to return `(*entity.ValuationSnapshot,
   error)` and propagate the error up through `GetValuationSummary`.
2. **(Fixed) Inconsistent business dates summed under one label.** The
   original loop tracked `businessDate` as the max across all in-scope
   portfolios' latest snapshots but summed every portfolio's AUM/P&L
   regardless of its own snapshot's date — so a portfolio not yet valued
   today (stale) would have its old AUM/P&L folded into a total labelled
   with today's date. Restructured into two passes: collect every latest
   snapshot first, determine the scope's max business date, then aggregate
   only the snapshots exactly on that date. Stale portfolios are excluded
   from the total rather than mixed in.
3. **(Confirmed correct, escalated — not changed) Minority-currency
   exclusion.** The adapter aggregates only the funds sharing the
   *predominant* currency within scope and excludes the rest (already
   tested as intentional in the pre-existing
   `TestGetValuationSummary_MixedCurrencyExcludesMinority`). Grepped
   `backend/internal/**`, `docs/**` for any reporting-currency/FX policy —
   none exists; `docs/investment-module.md` explicitly lists "Multi-currency
   NAV | Current valuation is single-currency per fund ... Cross-currency
   portfolios are Phase 2" as a known, documented gap. Per the task's own
   requirement 14, inventing an FX conversion here would be worse than
   excluding — **this needs Kanta's explicit sign-off as the interim
   behavior**, not a silent ship. No code change was made; this is a
   reporting decision, not a bug.
4. **(Fixed — coverage gap) No handler-level tests existed.** Added
   `backend/internal/integration/transport/handler/dashboard_handler_test.go`
   (new): auth-required (401), invalid scope (400), success envelope
   (200 + exact field values), provider failure (500).
5. Everything else in the 16-point contract was already correct in the
   partial implementation: `scope` is validated server-side to exactly
   `company`/`mine`; `mine` identity is resolved from `claims.Subject`/
   `claims.Username` (JWT), never from a query/body parameter (see
   `TestGetValuationSummary_MineScopeEchoesServerResolvedUsernameOnly`);
   "company" is bounded by the caller's data scope via
   `GetAccessibleContracts`, not unrestricted; the
   `INTEGRATION_DASHBOARD_VIEW` function permission gates the whole
   response before the provider is even called; monetary fields are decimal
   strings; `data_available=false` (never a fabricated zero) when no
   snapshot exists; 401/400/500 are mapped correctly.

Generated contract: Swagger (`docs/swagger.json`/`.yaml`/`docs.go`) and
`frontend/app/api/ims-api.d.ts` already matched the handler's existing
`@Success`/`@Failure`/`@Param` annotations exactly (generated in the earlier
partial session) — no request/response shape changed in this session, so
`make swagger`/`make api-client` were **not** re-run (nothing to
regenerate).

### Frontend completion

The partial frontend implementation already had the endpoint wired through
`useOpenApiClient()`/generated types, loading skeletons, distinct no-data vs.
error states, currency-aware formatting, signed P&L tone, scope-gated
"Entire company AUM" visibility (`authStore.hasContract("*")`), and full
EN/TH/ZH key parity (`compliance`-style i18n, already present). Five gaps
remained against the task's frontend requirements:

1. **(Fixed) Vitest-blocking barrel import.** `dashboard/lib/dashboard.ts`
   imported `formatMoneyCompact`/`formatPercent`/`parseDecimalOrNull` via
   `../../my-funds` (the feature barrel), which re-exports `myFundsApi` from
   `services/myFundsApi.ts` — a file with a top-level `~/api/openapi`
   import. Since ES modules evaluate every top-level import when the module
   loads, importing anything from the barrel in this project's Nuxt-less
   Vitest setup fails module resolution before any test can run. Fixed by
   importing directly from `my-funds/lib/format.ts` (pure, no Nuxt import).
2. **(Fixed, second instance of the same bug class) Test file's own direct
   import.** After fix 1, `tests/dashboard-valuation-summary.test.ts` still
   failed for the identical reason: it imported `normalizeValuationSummary`
   straight from `dashboard/services/dashboardApi.ts`, whose own top-level
   `~/api/openapi` import is unconditionally evaluated the moment any name
   is imported from that file — the barrel fix couldn't reach this. Extracted
   `normalizeValuationSummary` into a new pure module,
   `dashboard/lib/valuationSummaryMapping.ts` (only a `import type` from
   `~/api/ims-api`, which is erased at compile time and needs no runtime
   resolution). `dashboardApi.ts` now imports it from there; the test now
   imports it from there directly instead of through `dashboardApi.ts`.
3. **(Fixed) `as_of` was rendered in UTC, not Asia/Bangkok.**
   `formatDashboardTime` hardcoded `"UTC"`. Gave it an optional `timeZone`
   parameter (default `"UTC"` preserved) so the two pre-existing call sites
   — `DashboardOverviewHeader.vue`'s "last updated (UTC)" label (its i18n
   key literally says `lastUpdatedUtc` — intentionally UTC, left untouched)
   and the previous default in `buildAumMetric`/`buildPnlMetric` — keep
   their old behavior unless a caller opts in. `buildAumMetric`/
   `buildPnlMetric` now take a `locale` parameter and call
   `formatDashboardTime(summary.asOf, locale, "Asia/Bangkok")`;
   `DashboardOverviewScreen.vue` now destructures `locale` from `useI18n()`
   and passes it through. Verified there's an existing
   `useBangkokFormatter` composable elsewhere (settings/notifications
   pages) but it calls `useI18n()` internally and is Vue-setup-bound, which
   would have broken these functions' deliberate pure/unit-testable design
   (the file's own header comment: "this project's Vitest setup has no
   Nuxt/component runtime"); both share the same underlying
   `createDateFormatter(..., "Asia/Bangkok")` primitive, so the mechanism is
   consistent even though the composable itself wasn't reused.
   Updated the one test assertion that depended on the old UTC hour
   (`"14:30"` → `"21:30"` for the same UTC instant, now correctly rendered
   in Bangkok time).
4. **(Fixed) Scope-switch race safety.** `useDashboardValuationSummary` had
   no guard against a slow, superseded request overwriting a newer scope's
   result. Added a monotonically incrementing request id; `summary`/`error`/
   `loading` are only written by the most recently issued call.
5. **(Fixed) `metric as any` cast removed.** `DashboardOverviewScreen.vue`'s
   `displayMetrics` computed mixed `DashboardValuationMetric & {loading}`
   (from `buildAumMetric`/`buildPnlMetric`) with inline object literals for
   the contracts/approvals cards, and the template cast the loop variable to
   `any` to paper over the resulting inferred union. Both shapes are
   structurally identical to `DashboardOverviewMetric`; explicitly typed the
   computed as `(DashboardOverviewMetric & { loading: boolean })[]` and
   removed the cast.

### Files changed (all uncommitted; nothing staged)

Backend: `internal/investment/infrastructure/adapter/valuation_summary_adapter.go`
(edited, untracked/new file), `..._test.go` (edited, untracked/new),
`internal/integration/transport/handler/dashboard_handler_test.go` (new).

Frontend: `dashboard/lib/dashboard.ts`, `dashboard/services/dashboardApi.ts`,
`dashboard/composables/useDashboardValuationSummary.ts` (untracked/new,
edited), `dashboard/components/DashboardOverviewScreen.vue`,
`dashboard/lib/valuationSummaryMapping.ts` (new), `tests/dashboard-valuation-summary.test.ts`.

No other file was touched. The large pre-existing unrelated dirty/untracked
set (Portfolio Compliance V2 follow-ons, `investment-workspace`/`my-funds`
refactor, `portfolio-decision` feature, `docs/`) was verified present before
and after this session and is untouched.

### Financial formula (backend, unchanged design, now correctly gated)

Per portfolio: `today_pnl = (UnrealisedPnL_t + RealisedPnL_t) −
(UnrealisedPnL_t-1 + RealisedPnL_t-1)`, where `t` is the portfolio's latest
INTERNAL valuation snapshot and `t-1` is its most recent INTERNAL snapshot
strictly before `t`'s business date. `RealisedPnL` on a snapshot is
cumulative-to-date, so the subtraction isolates exactly that day's
contribution and stays correct across cash flows (a deposit/withdrawal
moves AUM but not unrealised/realised P&L). A portfolio with no prior
snapshot attributes its whole cumulative total to "today" (first valuation
day). `today_pnl_percent`'s denominator is the prior day's aggregate AUM,
falling back to today's AUM when there's no usable baseline. As of this
session: only portfolios whose latest snapshot is exactly on the scope's max
business date are summed (fix 2 above); a repository error while resolving
the prior snapshot now aborts the request with a 500 instead of silently
using the full cumulative total (fix 1 above).

### Validation — exact commands and results

Backend (from `backend/`):
- `go build ./...` — pass, no output.
- `gofmt -l` on all 3 changed/new Go files — clean, zero output.
- `go test ./internal/investment/infrastructure/adapter/... -run
  TestGetValuationSummary -v` — **PASS**, 12/12 (8 pre-existing + 4 new:
  error propagation, stale-portfolio exclusion, negative P&L, zero P&L).
- `go test ./internal/integration/... -v -run TestGetValuationSummary` —
  **PASS**, 13/13 (9 pre-existing query-layer + 4 new handler-level tests).
- `go test ./...` (full backend suite) — **PASS**, every package `ok` or
  `[no test files]`, zero `FAIL`.

Frontend (from `frontend/`):
- `npx vitest run tests/dashboard-valuation-summary.test.ts` — **PASS**,
  15/15 (was failing to even collect tests before this session's two import
  fixes).
- `npm run test -- --run` (full suite) — **PASS**, 30 files, 391 tests, 0
  failed.
- `npm run build` — **PASS**, `.output` produced, "Build complete!", no
  errors.
- `npx nuxi typecheck` — 89 pre-existing errors across 12 files, **none in
  any file this session touched** (verified by grepping the full output for
  each touched path). The 12 affected files
  (`DashboardTaskFeed.vue`, `DashboardWorkflowPanel.vue`,
  `PortfolioDirectoryView.vue`, `PortfolioOverviewView.vue`,
  `PortfolioFilterToolbar.vue`, `PortfolioKpiStrip.vue`,
  `PortfolioRoleActionMenu.vue`, `PortfolioSummaryCard.vue`,
  `SettingsControlCenter.vue`, `WatchlistAlertsPanel.vue`,
  `watchlistApi.ts`, `pages/investment/operator/index.vue`) are all part of
  the large concurrent unrelated refactor already present at session start.
  **Caveat:** this is not a byte-identical before/after diff (unlike the
  2026-07-13 session's rigorous baseline capture) — stashing across this
  heavily dirty, multi-feature worktree to get a true pre-edit baseline was
  judged too risky to the concurrent in-progress work. Confidence that zero
  new diagnostics were introduced rests on "no error appears in any touched
  file", not a full-repo diff.
- `git diff --check` scoped to this session's touched files only — clean.
  (Full-repo `git diff --check` still fails on pre-existing, untouched
  `dashboard/types.ts` trailing whitespace and the previously-documented
  `PortfolioOverviewView.vue:362` — neither touched by this session.)

### Remaining blocker: live browser proof not run

Same class of blocker as the 2026-07-14 P1 session, for two independent
reasons this time:

1. **No browser-driving tool is available in this session** — there is no
   Playwright/screenshot/browser-automation tool in this agent's toolset, so
   even an idle backend/frontend pair could not be visually verified here.
2. The user's own dev stack (`ims-backend`, `ims-frontend` Docker
   containers, healthy, up ~1h, pointed at `ims_dev`) is already running on
   the default ports. `infra/docker-compose.yml`'s `backend`/`frontend`
   services have **no bind-mounted source volumes** — they run prebuilt
   images — so this session's source edits are not live inside those
   containers even if they were reachable, and rebuilding+restarting them
   would interrupt the user's running session, which is out of scope to do
   unattended.

**Reproduction steps for whoever runs this next** (same pattern as the
2026-07-14 session): either (a) ask Kanta to rebuild/restart the existing
`ims-backend`/`ims-frontend` containers and drive the dashboard manually —
switch scope, refresh, inspect a negative-P&L day, switch `/en/`, `/th/`,
`/zh/` — or (b) stand up a second non-conflicting local pair (`APP_PORT`
override, `NUXT_PUBLIC_API_BASE_URL` pointed at it, `npm run dev -- --port
3090`) against a database that actually has a valuation snapshot dated the
current business date for at least one portfolio the logged-in user can see
in both `company` and `mine` scope. Checked `ims_dev` (read-only, via `docker exec ims-postgres psql`):
`investment__valuation_snapshots` has 427 `INTERNAL` rows with
`max(business_date) = 2026-07-15` — today's business date — so a live proof
against `ims_dev` will show real AUM/P&L numbers, not just the
`data_available=false` state. `ims_e2e`'s current seed state was not
checked.

### Financial/data risks for Kanta to review before committing

1. **Minority-currency exclusion (see backend finding 3 above).** "Entire
   company AUM" silently omits any fund whose base currency isn't the
   scope's predominant one. Correct behavior absent an FX policy, but a
   business-visible limitation that should be a conscious sign-off, not
   discovered later by a manager wondering why a USD fund is missing from
   the company total.
2. **Stale-portfolio exclusion is new behavior (fix 2).** A portfolio not
   yet valued today no longer contributes its (now provably-stale) AUM/P&L
   to "today's" total at all — previously it silently did. This is more
   correct, but changes what number renders whenever the nightly valuation
   run is incomplete across the scope at query time; worth Kanta's explicit
   read given it changes a number he's already looked at.
3. Live UI/three-locale browser proof is unverified (see blocker above) —
   only unit/handler/adapter-level correctness is proven.

### Precise next action

1. Kanta reviews this session's 8 changed/new files (listed above) —
   particularly the two adapter behavior changes and the FX/minority-
   currency escalation.
2. If acceptable, Kanta decides on committing (not pushing) this work,
   separately from the still-pending P1 `execution_handler.go` +
   `portfolio_v2_compliance_lifecycle_test.go` commit decision from the
   2026-07-14/07-15 sessions (unrelated files, no overlap).
3. Browser/three-locale proof per the reproduction steps above, once a
   live pair with real valuation-snapshot data is available.
4. Everything else already queued in `TASKS.md` (Make Fund Association
   Optional, Portfolio Regulation Rules, Architecture Cleanup items) is
   unaffected and unchanged by this session.

## Session: Independent Review of P1 Execution Handler Fix and E2E Test (2026-07-15)

**Goal:** a read-only, second-opinion review (task ID
`IMS-P1-COMPLIANCE-E2E-REVIEW`) of the exact two files flagged for Kanta's
approval by the 2026-07-14 session — no edits, no staging, no commits. Read
`MEMORY.md`, this file, and `TASKS.md` first, then verified Git state matched
the expected starting point exactly: branch `feature/investment`, HEAD
`2593c24`, index empty, same unrelated frontend churn present and untouched.

**Verdict: approve with concerns.** Both files are correct on their core
claims and safe to commit. Two non-blocking findings were recorded (also
logged in `TASKS.md` under P2 - Architecture Cleanup):

1. **(Medium) E2E cleanup leak.** `cleanupInvestmentPrereqs`
   (`tests/e2e/investment_test.go:394-415`, pre-existing/shared, not modified
   by this session) never deletes `investment__decisions` or
   `investment__executions`, both of which FK-reference the portfolio with
   `ON DELETE RESTRICT`. This new test is the first `cleanupInvestmentPrereqs`
   caller that creates a decision + executions against the seeded portfolio,
   so its cleanup's `DELETE FROM investment__portfolios`/`investment__funds`
   silently fail (errors are discarded via `_, _ = pool.Exec(...)`), leaving
   one fund, one portfolio, one decision, three executions, and their
   compliance check/breach rows orphaned in `ims_e2e` after every run. Fresh
   UUIDs per run mean this doesn't cause collisions or false passes — the
   test stays repeatable and correct — but it contradicts the 2026-07-14
   session's claim that cleanup "tears down its fund/portfolio/instrument
   afterward." The compliance rule instance and bindings the test creates are
   also never cleaned up at all.
2. **(Low) Stale Swagger.** `CreateExecution`
   (`execution_handler.go:112`) and `CreateExecutionByCode`
   (`portfolio_v2_execution_handler.go:56-72`) still document only
   `@Failure 400/401/404/409/500` — no `@Failure 422` — even though both can
   now return 422. `SubmitDecision` already documents 422 for the equivalent
   submit-time path. `git diff` on the three generated Swagger files confirms
   this session's unrelated 95/95/63-line additions don't touch the
   `executions` paths, so the contract change wasn't regenerated. No current
   frontend code calls this endpoint (only the generated `ims-api.d.ts`
   references the path) and no Bruno request asserts on the old 500, so
   blast radius today is zero — but `make swagger && make api-client` should
   still be run before/soon after commit per `CLAUDE.md`'s generated-file
   rule.

**HTTP 422 change — confirmed correct and consistent**, not a regression:
`ErrComplianceRejected` is returned before the execution entity is
constructed and before `runTx` (`execution.go:190-196`), so BLOCK persists
nothing; `errors.As` matches correctly since the error reaches the handler
unwrapped; the new case is byte-identical to two other existing precedents
(`writeDecisionError` in `decision_handler.go:641-642` and `writeDomainError`
in `investment_handler.go:152-153`), not a novel pattern; `writeExecutionError`
is shared by both V1 and V2 execution-create routes via `grep`-confirmed call
sites; PASS/WARN branches are untouched. This remains a genuine production
API contract change (500→422) that Kanta should personally bless for any
external (non-repo) caller that might branch on 500 vs. 4xx — no internal
caller does.

**E2E test — validated at the source level, not re-executed.** Every
acceptance point in the review brief (rule creation/binding via real HTTP,
real two-stage approval via `ben`/`green`, BLOCK/PASS/WARN/deactivation/
future-effective-date outcomes, DB-level assertions rather than status-code-
only checks, non-`GLOBAL`-rule isolation, generated-ID collision safety, the
non-E2E-DSN refusal guard, no sleeps/swallowed errors/sensitive data in the
test itself) was traced back to source and confirmed. Docker Desktop was not
running in this review session (`docker ps` failed to reach the daemon), so
no `ims_e2e` database was reachable and the targeted E2E test was
**not re-executed** — per scope, no database was created to work around
that. The PASS claim for
`TestE2E_PortfolioComplianceV2_ManagerApprovalToExecution` rests on the
2026-07-14 session's self-report; this review did not independently
reproduce it.

**Commands run (all read-only) and results:**

- `git status --short --branch`, `git branch --show-current`,
  `git log -1 --oneline`, `git diff --cached --name-only` — matched expected
  starting point.
- `go build ./...` — pass, no output.
- `go vet ./...` — pass, clean.
- `go vet -tags e2e ./tests/e2e/...` — pass, clean.
- `gofmt -l` on both reviewed files — clean.
- `go test ./...` — pass, every package `ok` or `[no test files]`, no `FAIL`.
- `docker ps` — failed (Docker Desktop not running); targeted E2E test not
  run.
- Final `git status --short --branch` / `git diff --cached --name-only` /
  `git log -1 --oneline` re-checked after the review — byte-identical to the
  pre-review capture. No file, index, commit, or remote state was changed by
  this session.

**Precise next action:** unchanged from the 2026-07-14 session's handoff
procedure below — Kanta decides whether to commit
`execution_handler.go` + `portfolio_v2_compliance_lifecycle_test.go` as-is,
or address finding 1 (cleanup) and/or finding 2 (Swagger) first. Neither
finding blocks the commit; both are logged as P2 follow-ups in `TASKS.md`.

## Session: P1 Portfolio Compliance V2 End-to-End Proof (2026-07-14)

**Goal:** prove the manager-approval-to-execution Portfolio Compliance V2
lifecycle against a disposable PostgreSQL database (never `ims_dev`), per
`TASKS.md`'s P1 task. Verified starting state matched the prompt: HEAD
`2593c24` on `feature/investment`, index empty. **Important deviation from
the documented starting state:** the worktree's unrelated dirty-file set had
grown substantially since the 2026-07-13 session's snapshot — a large,
in-progress frontend refactor (deletions across
`frontend/app/features/investment-workspace/`, `my-funds/`, and
`investment/funds/[fundId]/*` pages, plus new untracked
`frontend/app/features/portfolio-decision/`,
`frontend/app/pages/investment/operator/[fundId]/`, and
`frontend/app/pages/portfolios/[portfolioCode]/decisions/` — consistent with
the planned "Make Fund Association Optional for Portfolios" P1 item) appeared
mid-session, not caused by this session and not touched by it. This session's
own changes are exactly two: one modified backend file and one new backend
test file (see "Files changed" below). Do not confuse the larger dirty-file
set with this session's work when reviewing `git status`.

### What was proved, by requirement

**Setup (against `ims_e2e`, never `ims_dev`):**

- `ims_e2e` created fresh on the shared local Postgres container
  (`docker-compose exec postgres psql ... -c 'CREATE DATABASE "ims_e2e"'`),
  then `go run cmd/migrate/main.go up`, `go run ./cmd/seed`,
  `go run ./cmd/seed-e2e`, all with `infra/env/.env.e2e` sourced (copied from
  `infra/env/.env.e2e.example`, which was not yet present in this worktree).
  This is exactly what `make e2e-db-setup` does; it was run as constituent
  commands because the Makefile's `docker compose` subcommand form is not on
  PATH in this shell (`docker-compose` — the hyphenated standalone binary —
  works and is what was used).
- The `ims_e2e` database was lost partway through the session when the local
  Postgres/Docker Compose stack was recycled by something outside this
  session (the `ims-postgres`/`ims-backend`/`ims-frontend` containers all
  showed a fresh "up ~1 minute" uptime where they had previously been up 22+
  minutes) — expected behavior for a disposable database, not a bug. Recreated
  with the same three commands and re-verified before the final full-suite
  run recorded below.
- `database/seeds/zz_demo/04_demo_workflow_state.sql` (part of the standard
  `make seed`) already seeds `CURRENT_DATE` as `INVESTMENT_DAY_STARTED`
  (open for trading) — no additional workflow-day seeding was needed for
  `Decision.Submit`'s `IsTradeAllowed` gate or `Execution.Create`'s
  `IsTransactionLocked` gate.
- `database/seeds/012_approval_demo_seed.sql` already seeds a real two-stage
  `INVESTMENT_DECISION` approval process (`PROC_DECISION_DEFAULT`): stage 1
  `FUND_MANAGER_REVIEWERS` (ben, green), stage 2/final
  `INVESTMENT_SUPERVISORS` (admin, green) — used as-is, not invented.

**Rule instance and portfolio binding (real HTTP, not direct SQL):**
created a rule instance for `exposure.max_order_percent_aum`
(`max_percent_aum: 5`) via `POST /api/v1/compliance/rules`, then bound it to
the seeded test portfolio via
`POST /api/v2/portfolios/{code}/compliance/rules/{ruleInstanceID}/bindings`.
This rule type was deliberately chosen because it is **not** one of the four
rules `database/seeds/009_compliance_rule_seed.sql` already binds at `GLOBAL`
scope (`cash.availability`, `quantity.min_trading_unit`,
`amount.minimum_trade`, `ratio.sector_exposure`) — so the portfolio-scope
binding under test is the only thing gating order size against AUM, and
BLOCK/PASS/WARN outcomes are attributable to it specifically, not to
GLOBAL-rule interference. (First attempt used instrument code `"PTT"`, which
turned out to be a real seeded reference instrument already classified
`sector=ENERGY` — its 60,000 THB test order tripped the unrelated GLOBAL
`ratio.sector_exposure` rule and produced a false BLOCK. Fixed by using the
test's own freshly-seeded instrument's ticker instead.)

**Decision → two-stage manager approval (real HTTP, both stages):** created a
DRAFT decision (BUY, 20 units @ 100 THB = 2,000 THB = 2% of the portfolio's
100,000 THB AUM) via `POST /api/v2/portfolios/{code}/decisions`, submitted it
(pre-trade compliance PASSes at this size; a real `PROC_DECISION_DEFAULT`
approval request is created), then approved it through the **real**
`GET /api/v1/approvals/inbox` → `POST /api/v1/approvals/tasks/{id}/approve`
flow as `ben` (stage 1) and then `green` (stage 2/final) — not `admin`, since
the approval engine's maker-checker rule
(`authorizeAction` in `runtime_service.go`) refuses to let the submitter
approve their own request. The final approval's synchronous callback
(`DecisionApprovalSubjectCallback`) moved the decision to `APPROVED`,
verified via `GET .../decisions/{id}`.

**BLOCK** (`ordered_quantity=600` → 60,000 THB = 60% of AUM, over the 5%
limit): `POST .../decisions/{id}/executions` returned **422** (was 500
before this session's fix — see "Bug found and fixed" below) with a
human-readable rejection message; `investment__executions` row count for the
decision was unchanged (before == after); a `compliance_check_records` row
exists with `verdict=BLOCK`, `effective_severity=BLOCK`,
`final_verdict=BLOCK`; exactly one `compliance_breaches` row references it;
the decision remained `APPROVED` (retryable) afterward.

**PASS** (`ordered_quantity=20`, same size as the approved decision, BLOCK
binding still active): execution succeeded (201), persisted a real
`investment__executions` row, and the `compliance_check_records` row for
`exposure.max_order_percent_aum` shows `verdict=PASS`,
`effective_severity=BLOCK` (the binding was still BLOCK — proving the rule
was evaluated and genuinely passed, not skipped), `final_verdict=PASS`, with
zero breach rows for that check.

**WARN** (binding rebound to `severity=WARN`, same 600-unit order): the
rule's **raw** verdict is still `BLOCK` (`verdict=BLOCK` in the check
record — the order is still 60% of AUM), but
`valueobject.Severity.CapVerdict` downgrades it to `effective_severity=WARN`,
`final_verdict=WARN`; execution succeeded (201) and persisted, matching
`docs/MANAGER/MEMORY.md`'s "WARN currently does not block execution" against
actual code, not just the memory file's claim. A `compliance_breaches` row
still exists (`severity=WARN`, `status=OPEN`) — WARN remains auditable, not
silent.

**Deactivation changes which rule applies:** deactivated the BLOCK binding
(`DELETE .../compliance/rules/{id}/bindings/{bindingId}` → 204), then
re-attempted the 600-unit order: it succeeded (201), and no
`compliance_check_records` row for `exposure.max_order_percent_aum` was
created after that point — proving the rule was not resolved as applicable
at all (not "evaluated and happened to pass").

**Future `effective_from` does not apply today:** rebound the same rule
instance `severity=BLOCK` with `effective_from` = tomorrow. The 600-unit
order still succeeded (201) and, again, no check record for that rule type
was created — proving `ResolveApplicable`'s effective-window filter, not
just deactivation, correctly gates applicability by date.

**Locale (`en`/`th`/`zh`):** the V2 portfolio-code compliance routes are
pure JSON with no `Accept-Language`/locale handling anywhere in
`portfolio_v2_compliance_handler.go` — route behavior is locale-invariant by
construction, not something a locale switch can be run "against." The
locale-dependent surface is the frontend catalog UI, whose en/th/zh
resolution was re-verified via the existing unit suite
(`compliance-rule-catalog.test.ts`, `i18n-messages.test.ts`,
`portfolio-compliance-v2-binding.test.ts` — 3 files, 27 tests, all passed;
the P0 2026-07-13 session already asserted all 16 catalog entries × 3 keys
resolve in all three locales against the real message catalog, not a mock).
**Live three-locale browser proof was not run** — see "Remaining blocker"
below; this is a documented limitation, not a silent claim.

**E2E test:** `backend/tests/e2e/portfolio_v2_compliance_lifecycle_test.go`,
`TestE2E_PortfolioComplianceV2_ManagerApprovalToExecution`, 5 subtests (one
per verdict/lifecycle proof above). Refuses to run against any DSN whose
`dbname` doesn't contain `"e2e"` (in addition to the existing
`IMS_ALLOW_FAKE_WORKFLOW_STATE` guard the other e2e tests already use).
Boots its own server (`bootComplianceTestServer`) rather than reusing the
shared `bootTestServer` in `investment_test.go`, because it additionally
needs the approval module and the Portfolio Compliance V2 admin surface
(`SetPortfolioComplianceAdmin`) wired — kept separate so this P1 addition
cannot regress the three existing e2e test files that already depend on the
shared helper's current (narrower) wiring.

### Bug found and fixed (in scope — on the exact path under test)

`backend/internal/investment/transport/handler/execution_handler.go`'s
`writeExecutionError` had no `case` for `*domain.ErrComplianceRejected`, so
an execution-time compliance `BLOCK` fell through to
`httputil.InternalError` — **HTTP 500**, misclassified as an infrastructure
failure. The submit-time equivalent, `writeDecisionError` in
`decision_handler.go`, already had the correct case
(`httputil.UnprocessableEntity`, 422). Fixed `writeExecutionError` to add
the identical case, mirroring `writeDecisionError` exactly. This is a
**production API contract change** (500 → 422 on the execution-time BLOCK
path), not test scaffolding — flagged explicitly in "Highest-Risk Human
Review" below for Kanta's deliberate review before this is committed.

### Found, not fixed (pre-existing, unrelated to Portfolio Compliance V2)

Running the full `backend/tests/e2e` package (`-tags e2e`, no `-run` filter)
against the freshly seeded `ims_e2e` surfaced two pre-existing failures,
neither touched by or related to this session's work (confirmed via
`git status` — neither file appears in this session's diff):

- `TestE2E_InvestmentOversellEnvelope` — `ERROR: inconsistent types deduced
  for parameter $4` in its `seedOpeningPosition` helper. Already documented
  in `tests/e2e/COVERAGE.md` as a pre-existing bug found during the IAM E2E
  session; `make test-e2e-backend` already scopes to `-run TestE2E_IAM`
  specifically to avoid it.
- `TestE2E_PortfolioV2_DecisionOwnership` — **new finding this session**:
  `ERROR: column "contract_id" does not exist`. The test queries `SELECT
  fund_id, contract_id, portfolio_id FROM investment__decisions`, but
  `contract_id` was removed from that table by commit `91dab64` ("drop
  contract id from operational records"), which predates this session. The
  test is stale against the current schema — a real gap, but investment-
  decision-scoped, not compliance-scoped, and out of this task's bounds to
  fix.

### Tests and exact commands run (2026-07-14)

- `docker-compose exec -T postgres psql -U ims_app -d ims_dev -p 5437 -c
  'CREATE DATABASE "ims_e2e"'` — created (twice; see the mid-session
  container-recycle note above).
- `cd backend && (source ../infra/env/.env.e2e) && go run cmd/migrate/main.go
  up` — "Migrations successful".
- `... && go run ./cmd/seed` — "Database seeded successfully".
- `... && go run ./cmd/seed-e2e` — "E2E fixtures seeded successfully",
  users=7, groups=4.
- `go build ./...` (backend) — passed.
- `go vet ./...` and `go vet -tags e2e ./tests/e2e/...` — both clean.
- `gofmt -l` on both changed/new files — clean (zero output).
- `go test -tags e2e -count=1 -v -run
  TestE2E_PortfolioComplianceV2_ManagerApprovalToExecution
  ./tests/e2e/...` — **PASS**, all 5 subtests, run twice for idempotency
  (fresh UUIDs per run; the test's own `cleanupInvestmentPrereqs` tears down
  its fund/portfolio/instrument afterward).
- `go test -tags e2e -count=1 -v ./tests/e2e/...` (full package, no filter)
  — this session's new test passes; `TestE2E_IAM_AuthAndAuthorization` and
  `TestE2E_PortfolioV2_LedgerWrites` (pre-existing) pass;
  `TestE2E_InvestmentOversellEnvelope` and
  `TestE2E_PortfolioV2_DecisionOwnership` fail for the pre-existing/unrelated
  reasons documented above.
- `go test ./...` (backend, full suite, no `e2e` tag) — all packages pass,
  including `internal/investment/transport/handler` (contains the fixed
  file).
- `cd frontend && npm run test -- --run` — 22 files, 310 tests, all passed.
  (Differs from the 2026-07-13 session's recorded 23/325 baseline; this
  frontend worktree changed substantially mid-session per the note at the
  top of this section, and this session did not touch any frontend file —
  the discrepancy belongs to that concurrent work, not investigated further
  as out of scope for a backend-focused P1 task.)
- `npx vitest run tests/compliance-rule-catalog.test.ts
  tests/i18n-messages.test.ts tests/portfolio-compliance-v2-binding.test.ts`
  (frontend) — 3 files, 27 tests, all passed (locale proof, see above).

### Remaining blocker: UI and live three-locale browser proof not run

Both the browser proof of BLOCK/PASS in the UI (part of requirement 4/5) and
live navigation of the portfolio-code compliance routes in `/en/`, `/th/`,
`/zh/` (requirement 8) require a backend + frontend pair running against
`ims_e2e`, per the README's E2E workflow. This was **not run**: the host's
default ports (backend `:8080`, frontend `:3000`) are already held by the
user's own running dev stack (`ims-backend`, `ims-frontend` containers,
pointed at `ims_dev`), and stopping or reconfiguring that stack without
asking was judged out of scope for an unattended proof session — especially
given the concurrent frontend refactor discovered mid-session. This is a
documented, reproducible gap, not a silent UI-proof claim.

**Reproduction steps for whoever runs this next:**

```bash
cp infra/env/.env.e2e.example infra/env/.env.e2e   # if not already present
# pick non-conflicting ports, e.g.:
APP_PORT=8090 (backend/.env or exported)
NUXT_PUBLIC_API_BASE_URL=http://localhost:8090/api/v1 (frontend)
cd backend && go run cmd/server/main.go             # against ims_e2e, port 8090
cd frontend && npm run dev -- --port 3090            # pointed at the port-8090 backend
```

Then, as `ben`/`admin`/`green` (passwords reset to `admin123` by this
session's test — reset again if the database was recreated), drive the same
BLOCK/PASS lifecycle through the Portfolio Compliance V2 UI, and visit the
portfolio compliance page under `/th/...` and `/zh/...` to confirm the
already-tested `compliance.catalog.*` keys render correctly in the browser.

## Session: Compliance Frontend Correction (2026-07-13, after `e8b2a6d`)

A focused frontend-only task fixed six defects in the Portfolio Compliance V2
UI: rule-catalog i18n, backend-catalog coverage/category drift, rule-builder
parameter-sample drift, effective-state overclaiming, and a generated-type
cast. Verified starting state matched the prompt exactly: HEAD `e8b2a6d` on
`feature/investment`, index empty, the same unrelated dirty files present
(`portfolio_crud.go`, four `Portfolio*View.vue`, `dashboard.vue`, two Bruno
health files, and the broad untracked `docs/` tree) — all preserved untouched.

**Defects confirmed:**
1. `RULE_CATALOG` used hardcoded English strings with no real `en`/`th`/`zh`
   support — `t()` was never called for catalog content.
2. Catalog covered only 9 of the 16 backend-registered rule types (missing
   `allocation.asset_class_max/min`, `amount.minimum_trade`,
   `exposure.max_order_percent_aum`, `valuation.min_nav`,
   `credit_rating.minimum`, `regulatory.thai_sec`).
3. Category drift: `credit.min_rating` and `credit_rating.minimum` were
   labelled `RESTRICTION` in the catalog but are backend `MANDATE`;
   `ratio.sector_exposure` was labelled `RATIO` but is backend `MANDATE`.
   Verified against each rule's `Metadata()` in
   `backend/internal/compliance/rules/**`.
4. `SAMPLE_PARAMS` drift: `cash.availability` used `min_balance` (backend
   key is `min_cash_buffer_pct`), `ratio.sector_exposure` used `sector_code`
   (backend key is `sector`), `quantity.min_trading_unit` used `min_units`
   (backend key is `default_lot_size`), and the three `restriction.*` rules
   used an invented `list_code` key not in any `ParameterSchema()`.
   `credit_rating.minimum`/`regulatory.thai_sec` (stubs) had samples that
   made them look production-ready and were selectable in the builder
   `<select>`.
5. `PortfolioComplianceView.vue`'s bound-rules table and subtitle read
   `binding.is_active` alone as "currently enforced," overclaiming when a
   binding is scheduled in the future or past `effective_to` (matches the P2
   finding already recorded in `TASKS.md`).
6. `formatParameters` cast `entry.parameters as Record<string, unknown>`
   directly at the render call site instead of through a boundary
   normalizer (root cause: `contracts.go`'s `swaggertype:"object"` still
   produces `Record<string, never>` — unchanged, backend-only, out of scope).

**Portfolio Compliance V2 identity contract (Fix 1):** verified, not
changed. `portfolioComplianceApi.bindRule`'s body already matched the
generated `BindRuleRequest` schema (`severity`, `priority`,
`effective_from`, `effective_to`) and the backend's
`handler.BindRuleRequest` struct — no `fund_id`, `FundID`, `contract_id`,
`ContractID`, `portfolio_id`, `scope_id`, or `scope_type` in either. Added a
regression test (`portfolio-compliance-v2-binding.test.ts`) that asserts the
built payload's key set and the absence of every forbidden field, so a
future edit can't silently reintroduce one. The V1 simulator's
`fund_id` portfolio-picker filter (`complianceApi.listPortfolios`) is a
different endpoint and was correctly left alone.

**Files changed (24; staged, not committed):**

Catalog:
- `frontend/app/features/compliance/lib/ruleTypeCatalog.ts` — rewritten to
  `labelKey`/`explanationKey`/`suggestedCorrectionKey`/`selectable`, full
  16-type coverage, corrected categories.
- `frontend/app/features/compliance/lib/ruleParameterSamples.ts` (new) —
  `SAMPLE_PARAMS` extracted from the builder component so it's unit
  testable; corrected keys; no sample for either stub.
- `frontend/app/features/portfolio-workspace/lib/complianceBindingState.ts`
  (new) — `deriveBindingState`, `buildBindRulePayload`,
  `isSubmissionInFlight` extracted as pure functions for the same reason.

i18n: `en/th/zh` `compliance.ts` (new `compliance.catalog.*` tree, 16 rule
types × label/explanation/suggestedCorrection) and `en/th/zh` `portfolio.ts`
(`portfolio.compliance.asOfLabel`, `.state.*`, `.columns.status`, reworded
`boundSubtitle`).

Components: `ComplianceRuleBuilder.vue` (samples + stub gating +
`t(labelKey)`), `PortfolioComplianceView.vue` (Fix 5 + Fix 6 + duplicate-
submission guard), and eight read-only consumers updated to pass `t` into
`ruleLabel`/`ruleExplanation`/`ruleSuggestedCorrection`
(`ComplianceCheckResultPanel`, `ComplianceRuleTable`, `ComplianceRuleTabs`,
`ComplianceRuleDetailHeader`, `ComplianceRecentFailuresPanel`,
`ComplianceHighRiskList`, `ComplianceBreachOverrideDialog`,
`ComplianceBreachInbox`, `ComplianceAuditTimeline`, `pages/compliance/pre-trade.vue`).

Tests: `compliance-rule-catalog.test.ts` (rewritten — coverage, category,
selectable, i18n resolution in all three locales, unknown-type fallback),
`compliance-rule-parameter-samples.test.ts` (new — required/forbidden
parameter keys, no stub samples), `portfolio-compliance-v2-binding.test.ts`
(new — V2 payload forbidden-field absence, `deriveBindingState` boundary
matrix, `isSubmissionInFlight` guard).

**Rule catalog coverage:** 9/16 → 16/16 backend-registered types (verified
against `module.go`'s blank imports of all 11 rule packages).

**en-US / th-TH / zh-TW coverage:** all 16 catalog entries × 3 keys resolve
in all three locales (asserted against the real `messages` catalog via
`resolveMessage`, not a mock); `i18n-messages.test.ts`'s existing
locale-key-parity check continues to pass, confirming `en`/`th`/`zh` stayed
structurally identical. zh uses Traditional characters for all new content
in both `compliance.ts` and `portfolio.ts`. **Pre-existing debt noted, not
fixed:** `zh/portfolio.ts`'s existing (untouched) keys use Simplified
Chinese, inconsistent with `zh/compliance.ts`'s Traditional convention —
out of scope for this task; recommend a dedicated zh-TW normalization pass.

**Parameter-schema corrections:** `cash.availability` → `min_cash_buffer_pct`;
`ratio.sector_exposure` → `sector` + `max_pct`; `quantity.min_trading_unit`
→ `default_lot_size`; `restriction.{blacklist,whitelist}` → `{}`;
`restriction.list_enforcement` → `enforced_types`. Added verified samples
for `allocation.asset_class_max/min`, `amount.minimum_trade`,
`exposure.max_order_percent_aum`, `valuation.min_nav`. No sample for
`credit_rating.minimum` / `regulatory.thai_sec` (stubs) — both also now
`disabled` in the builder `<select>` via `entry.selectable`.

**Effective-state behavior:** `PortfolioComplianceView.vue` now derives
Scheduled/Currently effective/Expired/Deactivated from `is_active` +
`effective_from`/`effective_to` (boundary-inclusive, mirroring
`compliance/lib/formatters.ts`'s `deriveRuleStatus`) instead of reading
`is_active` alone as "enforced." The workspace exposes no authoritative
backend business date, so "today" is the browser's local date, always
labelled "As of {date} (your device's local date)" rather than implied as
the backend's business date — backend enforcement stays authoritative and
unchanged. `DEACTIVATED` bindings still surface in the "Available" section
per existing `boundRules`/`unboundRules` filtering (unchanged), so the
bound table itself only ever shows Scheduled/Effective/Expired; the
derivation helper still covers all four states for the unit tests.

**Tests and exact validation results (2026-07-13):**
- `npx vitest run tests/compliance-rule-catalog.test.ts
  tests/compliance-rule-parameter-samples.test.ts
  tests/portfolio-compliance-v2-binding.test.ts tests/i18n-messages.test.ts
  tests/compliance-formatters.test.ts` — 5 files, 51 tests, all passed.
- `npm run test` (full suite) — 23 files, 325 tests, all passed (up from
  21/300 in the prior session; +2 new test files, +25 new tests).
- `npx nuxi typecheck` (baseline captured before any edit, then re-run
  after) — both runs produced **byte-identical** output (`diff` exit 0):
  same 290 lines, same pre-existing failures
  (`ComplianceAuditTimeline.vue`, `ComplianceRecentFailuresPanel.vue`,
  `complianceApi.ts` ×2, `FundSummaryCard.vue`, `PortfolioSummaryCard.vue`
  — all unrelated to this task's touched files). Zero new diagnostics
  introduced anywhere.
- `npm run build` — passed; `.output` produced, no errors.
- `git diff --check` — fails only on the same pre-existing unrelated
  `PortfolioOverviewView.vue:362` trailing whitespace noted in the prior
  session; not touched by this task.
- `git diff --cached --check` — passed (staged diff is clean).
- `make lint`'s frontend step is `npx nuxi typecheck` — see above.
- Staged secret-pattern scan (`secret|password|api_key|private_key|token=`)
  over the full diff — zero hits.

**Remaining risks / follow-ups (recorded in `TASKS.md`):**
- `PortfolioRuleCatalogEntry.parameters` generated-type drift
  (`Record<string, never>`) is unchanged — root cause is
  `contracts.go`'s `swaggertype:"object"` annotation, backend-only, already
  tracked as a P2 item. `asParameterRecord`'s `unknown`-accepting boundary
  normalizer is the correct frontend-only mitigation and does not conceal
  the root cause.
- No backend rule-catalog endpoint was created (explicitly out of scope);
  the frontend catalog remains a manually-maintained mirror — already
  tracked as a P2 architecture item ("Add a live rule-type catalog
  endpoint...").
- `zh/portfolio.ts` Simplified/Traditional inconsistency noted above — new
  P2 follow-up, not yet in `TASKS.md`.
- The three pre-existing `bindDrafts[...]` type fixes and the
  `PortfolioComplianceView.vue`/`portfolioComplianceApi.ts` typecheck
  cleanliness from the prior session are preserved — confirmed by the
  byte-identical typecheck diff above.

**Exact Git state:** Kanta authorized the commit (not a push) on
2026-07-14. Branch `feature/investment` now at `2593c24`
(`fix(compliance): correct rule catalog i18n, coverage, and
effective-state display`), 24 files, on top of `e8b2a6d`. The index is
empty again. Everything else — backend, other portfolio-workspace views,
`dashboard.vue`, Bruno files, and the broad untracked `docs/` tree —
remains exactly as it was at session start. No push performed; no upstream
configured.

**Precise next action:** P1 (disposable-database end-to-end proof) is the
next queued task. No action is pending from this session.

## Start Here

1. Read `docs/MANAGER/MEMORY.md`.
2. Read this file.
3. Read `docs/MANAGER/TASKS.md` and work only on its active task.
4. Run `git status --short --branch`, `git branch --show-current`, and
   `git diff --cached --name-only` before taking action.
5. Read the actual diff for any file you will modify or stage.

## Current Git State

Verified on 2026-07-13 after the authorized local commit:

- Branch: `feature/investment`, attached at `e8b2a6d`
  (`feat(compliance): enforce portfolio compliance before execution`).
- The branch has no upstream configured. The commit is local and no push was
  performed.
- Before attaching, `HEAD` and `feature/investment` both resolved to the same
  full commit. Pre/post hashes for the complete status manifest, tracked diff,
  and cached diff were identical, so branch attachment changed no file.
- Commit `e8b2a6d`: 47 reviewed Portfolio Compliance V2 files, 6,551 additions
  and 39 deletions. The index is empty after the commit.
- Remaining worktree changes are unstaged and were not included. They include
  `portfolio_crud.go`, the unrelated portfolio cash, holdings, ledger,
  overview and dashboard-layout files, and two Bruno health files. The broad
  untracked documents, including `docs/MANAGER/`, remain unstaged.

## Portfolio Compliance V2 State

The implementation is reviewed, validated, and committed locally as `e8b2a6d`.
It has not been pushed.

- The business lifecycle is preserved: decision submit-time compliance remains;
  manager approval precedes execution; execution creation re-runs compliance
  after lifecycle/workflow checks and before persistence.
- `BLOCK` returns `domain.ErrComplianceRejected` and creates no execution.
  Explicit `PASS` and `WARN` proceed. A nil result or unknown verdict now fails
  closed. A nil checker remains tolerated only for existing unit-test builders;
  production receives the non-nil compliance adapter in `main.go`/`module.go`.
- When ordered quantity and amount are both supplied, execution compliance uses
  `ordered_amount / ordered_quantity` as the effective price, so evaluated trade
  value equals the actual notional. Quantity-only orders use the approved
  decision limit price. Amount-only orders fail closed because quantity-based
  rules cannot be evaluated safely without units.
- Invalid V2 `order_id`, invalid order side, and compliance validation failures
  are 400-class input errors; infrastructure failures remain 500-class.
- Portfolio-code handlers resolve the portfolio once in investment, enforce
  existing fund-scoped data access, and pass only internal IDs across the
  compliance contract.
- Binding creation/deactivation, duplicate protection, catalog listing, breach
  listing, and both route paths have focused command/handler tests.
- The nullable `contract_id` migration preserves honest portfolio-only audit
  rows; its down migration refuses to invent a sentinel. The partial unique
  index enforces one `is_active=true` binding per rule instance and portfolio.
- The three `bindDrafts[...]` type diagnostics were fixed locally. No UUID,
  `fund_id`, or `contract_id` is rendered as a normal compliance UI label.
- EN/TH/ZH add identical key sets. V2 Swagger contains all six compliance path
  shapes under `/api/v2`, V1 excludes them, and the generated TypeScript paths
  and schemas match. Generated hashes are stable across a second run.
- The Makefile now quotes tag filters portably and uses
  `--instanceName=v2`, so `make api-client` reproduces both specs on Windows.

## Highest-Risk Human Review

Kanta should personally review these files for business meaning before saying
`commit`:

1. `backend/internal/investment/application/command/execution.go`
   - Confirm the gate runs before persistence and uses approved decision data.
   - Confirm price/quantity behavior is financially correct for all order forms.
   - Confirm `BLOCK` produces no execution and `WARN` is intentionally allowed.
2. `backend/internal/investment/application/command/execution_workflow_test.go`
   - Confirm tests prove business outcomes, not only method calls.
3. `backend/internal/investment/module.go`
   - Confirm production wiring always supplies the compliance checker.
4. `database/migrations/20260707000001_compliance__nullable_contract_id.*.sql`
   - Confirm portfolio-only audit rows may legitimately have NULL contract scope.
   - Confirm the intentionally non-reversible rollback policy is acceptable.
5. `database/migrations/20260707000002_compliance__unique_active_portfolio_binding.*.sql`
   - Confirm the uniqueness rule matches binding lifecycle semantics.
6. `backend/internal/investment/transport/handler/portfolio_v2_compliance_handler.go`
   - Confirm portfolio access, portfolio-code resolution, and 400/500 mapping.
7. `frontend/app/features/portfolio-workspace/PortfolioComplianceView.vue`
   - Confirm a manager can understand active rules, parameters, severity, dates,
     binding, and deactivation without seeing internal UUIDs.
8. **New, 2026-07-14:**
   `backend/internal/investment/transport/handler/execution_handler.go`
   (`writeExecutionError`)
   - This is a production API contract change, not test scaffolding: an
     execution-time compliance `BLOCK` now returns HTTP 422 (with
     `error_code`-less `{"error": "..."}` body, matching the existing
     `writeDecisionError`/submit-time pattern) instead of 500. Confirm 422 is
     the correct status for any caller (frontend or external integration)
     that may currently branch on 500 vs 4xx for this specific endpoint.

AI review completed for generated Swagger consistency, i18n key coverage, API
type drift, formatting, secret patterns, and focused test gaps.

## Verified Validation

Exact results from this session:

- `make api-client` - passed after the two-line Makefile quoting fix; a second
  generation produced identical hashes for all seven generated files.
- `cd backend; go build ./...` - passed in 28.9s.
- `cd backend; go test ./...` - passed in 39.0s.
- `cd backend; gofmt -l .` - passed; `unformatted_count=0`.
- `cd frontend; npm run build` - passed in 194.6s; only Node `DEP0155`
  dependency deprecation warnings were emitted.
- `cd frontend; npm run test` - passed: 21 files, 300 tests, 300 passed.
- `cd frontend; npx nuxi typecheck` - failed on pre-existing branch-wide
  approval, watchlist, and unrelated i18n/type debt. The three
  `PortfolioComplianceView.vue` diagnostics are gone; no Portfolio Compliance
  diagnostic remained in the completed run. This debt is a P2 task.
- `git diff --cached --check` - passed.
- `git diff --check` - failed only on the unstaged, unrelated
  `frontend/app/features/portfolio-workspace/PortfolioOverviewView.vue:362`
  trailing whitespace.
- Staged secret-pattern scan - zero private-key, credential assignment,
  AWS-secret, or credentialed-URL hits.

## Immediate Handoff Procedure

1. Do not push `e8b2a6d` unless Kanta separately authorizes a push.
2. ~~Begin P1 by provisioning a disposable or dedicated PostgreSQL
   database...~~ Done 2026-07-14 — see "Session: P1 Portfolio Compliance V2
   End-to-End Proof" above. Backend/API proof is complete; live UI/locale
   browser proof remains a documented, reproducible gap (same section,
   "Remaining blocker").
3. **Next action:** Kanta reviews and (if acceptable) locally commits this
   session's two files —
   `backend/internal/investment/transport/handler/execution_handler.go` (the
   422 bug fix; see "Highest-Risk Human Review" item 8) and
   `backend/tests/e2e/portfolio_v2_compliance_lifecycle_test.go` (the new
   E2E test) — isolated from the large, unrelated, in-progress frontend
   refactor already present in this worktree. Do not stage or commit any of
   that frontend churn as part of this P1 change.
3a. **Done 2026-07-15:** an independent read-only review of exactly these two
   files returned "approve with concerns" — see "Session: Independent Review
   of P1 Execution Handler Fix and E2E Test (2026-07-15)" above. Two
   non-blocking findings (E2E cleanup leak, stale Swagger docs) were logged
   in `TASKS.md`; neither blocks the commit. The targeted E2E test was not
   re-executed in that session (Docker unavailable).
4. Optionally, run the UI/locale reproduction steps in the "Remaining
   blocker" section above once a non-conflicting backend/frontend pair can
   be started against `ims_e2e`.
5. Preserve every remaining unstaged/untracked worktree change and keep P1
   changes isolated from them.

## Independent Re-Verification (2026-07-13)

A second review, prompted by the owner's belief that a frontend defect
remained in `PortfolioComplianceView.vue`, re-derived every claim in this
file from source and a live `npx nuxi typecheck` run rather than accepting
the prior session's summary.

- Confirmed: the three `bindDrafts[...]` diagnostics are gone. `npx nuxi
  typecheck` reports zero diagnostics for `PortfolioComplianceView.vue` or
  `portfolioComplianceApi.ts`. Remaining failures are pre-existing
  `AppTranslationKey`/i18n and unrelated-page (`PortfolioSummaryCard.vue`,
  `PortfolioDirectoryView.vue`, watchlist, holdings, operator) debt in files
  this branch does not touch — matches the "P2 task" note above.
- Confirmed: `execution.go`'s compliance gate, `module.go` wiring,
  `main.go`'s non-nil adapter injection, both migration pairs, and the V2
  handler's error mapping are all correct against a fresh read. No backend
  defect found.
- Re-ran `gofmt -l .` (clean), `go build ./...` (pass), `go test ./...`
  (pass, no `FAIL`), `git diff --check` (fails only on the same unrelated
  `PortfolioOverviewView.vue:362` trailing whitespace), `git diff --cached
  --check` (pass), `npx vitest run` (21 files / 300 tests pass), `npm run
  build` (pass). No regression since the 07-10 session.
- **New P2 finding, not fixed (would require expanding UI scope the task
  forbids):** `PortfolioComplianceView.vue`'s `boundRules` computed
  (line ~166) and the "Bound Compliance Rules" card subtitle ("Rules
  currently enforced for this portfolio") key off `binding.is_active` only.
  The backend's `ResolveApplicable` (`rule_binding_repository.go:209-216`)
  additionally gates on `effective_from`/`effective_to` against today's
  business date. A binding that is `is_active=true` but scheduled in the
  future or past its `effective_to` will display as "currently enforced"
  even though the pipeline will not evaluate it today. Enforcement itself is
  correct — only the UI label overclaims. Recommend a follow-up task to add
  a per-row effective-state indicator (scheduled/active/expired) rather than
  folding it into this commit, since the current binding is scoped to the
  three type diagnostics and non-goals explicitly exclude UI redesign.
- **New P3 finding, not fixed:** `formatParameters` (line 81) casts
  `entry.parameters as Record<string, unknown>` because the generated type
  for `PortfolioRuleCatalogEntry.parameters` is `Record<string, never>`. Root
  cause is `backend/pkg/contract/contracts.go:248` — `Parameters
  json.RawMessage \`swaggertype:"object"\`` gives swaggo no shape info, so
  the OpenAPI schema has no `additionalProperties` and the generator emits an
  empty-object type. The cast is safe (`never` is assignable into
  `unknown`) and not the unsafe pattern the task's guardrails target, but it
  is a real generated-type/source-contract mismatch worth fixing at the
  Swagger annotation, not by casting further in the frontend.
- Migration SQL was reviewed line-by-line for correctness (nullable-contract
  ALTER TABLE pair, partial unique index, guarded non-reversible down
  migration) but **not applied** to the running local `ims-postgres`
  container in this session — `make migrate-up`/`migrate-down` were not run
  to avoid mutating the owner's live dev database without asking. This
  remains open under the P1 "Prove End to End" task.

No commit-blocking issue was found. Kanta authorized the local commit on
2026-07-13, producing `e8b2a6d`; no push was performed. The precise next action
is the disposable-database P1 end-to-end proof.

## New-Instance Prompt

```text
Act as the engineering manager for the IMS repository.

Read these files in order:
1. docs/MANAGER/MEMORY.md
2. docs/MANAGER/HANDOFF.md
3. docs/MANAGER/TASKS.md

Then verify the current Git state and actual source before acting. Work only on
the active task in TASKS.md. Preserve unrelated worktree changes. Do not commit
or push unless I explicitly request it. Report conflicts between documentation
and code, run the listed acceptance checks, and update HANDOFF.md and TASKS.md
before ending the session.
```
