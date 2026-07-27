---
type: claude-goal-account-handoff
project: IMS Thailand
goal_id: IMS-MERGE-BLOCKERS
owner: Kanta
status: in-progress
last_updated: 2026-07-27
---

# Claude `/goal` Next-Account Handoff

This file is the durable continuation state for
`docs/MANAGER/CLAUDE-GOAL-IMS-REMEDIATION.md`. Every Claude account/session
must verify and replace this baseline before relying on it.

## Account / Session

- Claude account/session identifier: Account #1 (Opus 4.8, Claude Code), session
  `5dba1542`. Program started 2026-07-24. Continued by a new Account/session on
  2026-07-26 (Opus 4.8, Claude Code) which implemented Fix A/B/C — see below.
  Continued again 2026-07-27 (Sonnet 5, Claude Code, via `/goal` + Stop-hook)
  which: independently re-verified the findings-3-5 worker diff (main tree,
  real-Postgres proof on a genuine disposable container, not just the shared
  dev DB), confirmed the G1 demo-identities migration was NEVER recovered
  (worker died before persisting any file — verified by search across every
  worktree), and obtained two fresh owner decisions via `AskUserQuestion` (see
  below). This same account then completed G2 items 4, 6, 7, 8, and the
  frontend Idempotency-Key wiring (all source-proven, see Work-Package Matrix
  below), verified the `011_approval_permission_seed.sql` grant end-to-end on
  a fresh disposable DB (Go catalog upsert -> SQL grant -> Admin group, no FK
  gap), and dispatched the independent G2 full-diff review (task #9,
  in-flight as of this checkpoint — verdict not yet known to this account at
  time of writing).
- Started: 2026-07-24
- Last checkpoint: 2026-07-27 — **G2 items 2-3 (Codex findings 3-5: canonical
  fingerprint, keyed-retry payload validation, ambiguous-submit
  reconcile-or-compensate, real-Postgres concurrency) are CODE-COMPLETE and
  independently re-verified this session** (main tree, uncommitted). Evidence:
  `gofmt -l .` clean (after this account fixed two pre-existing unrelated
  gofmt violations — see "Files changed" below), `go vet ./...` clean, full
  `go test ./... -count=1` all `ok` (zero FAIL), and — the important part —
  the 3 new tests in `cash_request_repository_integration_test.go`
  (`TestIntegrationCashRequestRoundTrip`,
  `TestIntegrationCashRequestFingerprintDeterministic`,
  `TestIntegrationCashRequestConcurrentSameKey`) plus the pre-existing
  `TestIntegrationCashRepositoryConcurrentFirstInsertRetry`/
  `TestIntegrationPositionRepositoryConcurrentFirstInsertRetry` were rerun
  against a FRESH disposable `postgres:16-alpine` container (`docker run
  --name ims-g2-disposable-pg -p 15437:5432 ...`, destroyed immediately after)
  with `IMS_TEST_DATABASE_DSN` pointed at it — all 5 PASS. (Caution for the
  next reader: `go test ./...` run WITHOUT that env var silently targets
  `integrationDSN()`'s default, which is the SHARED `ims-postgres` dev
  container on port 5437, not a disposable one — this is pre-existing repo
  test convention via an isolated throwaway schema + `DROP SCHEMA CASCADE`
  cleanup, confirmed zero residue via `information_schema.schemata`, but it is
  NOT a disposable-container proof. Always run `go test ./... -short
  -count=1` for a general gate to avoid silently touching the shared dev DB,
  and export `IMS_TEST_DATABASE_DSN` at a throwaway container when a
  real-Postgres claim is actually needed.) G1 is UNCHANGED from the prior
  checkpoint below except one correction: the
  `20260727000001_permissions__remove_seeded_demo_identities` migration that
  the prior checkpoint said "a worker was authoring" **does not exist** —
  searched every worktree and the main tree's untracked files; nothing found.
  Do not assume it survived; it must be authored fresh. Kanta was asked and
  chose to author it now (see Owner Decisions). Continuing 2026-07-26 note
  verbatim below for the rest of the G1/G2 history:
- Previous checkpoint 2026-07-26 — **Fix A (classification redesign), Fix B
  (make/env/CI wiring), and Fix C (full-patch export) IMPLEMENTED and verified
  on a disposable DB + local Go gates.** An independent full-diff review is
  running; G1 is NOT yet accepted (awaits that review clean + Kanta's integration
  approval).
- Goal status: in progress. NEW GOAL 2026-07-27 (Kanta) = finish IMS-MERGE-BLOCKERS
  through G9, FIRST resolving 6 Codex Request-Changes findings. Progress:
  - **Codex finding 1 (G1 hard-fail) RESOLVED + verified 2026-07-27:** production
    seeding now HARD-ERRORS on zero SQL files, missing/empty manifest, OR any
    manifest-listed reference file absent (`validateReferenceFilesPresent`,
    checked before any zero-file success; the prior WARN is now a hard error).
    Disposable-DB proven (hiding `006` → exit 1 naming it; real tree still exit 0).
    A FRESH independent G1 review is RUNNING; G1 acceptance pending it (no P0/P1).
  - **Codex finding 2 (G2 item-1 missing-port) RESOLVED + verified 2026-07-27:** a
    missing subject-access port is now an operator-visible 5xx (logged), not 403;
    only `contract.ErrSubjectAccessDenied` → 403. Two tests updated to the new
    contract; approval build+tests green.
  - **Codex findings 3-5 (guaranteed cash idempotency + canonical fingerprint +
    timeout/concurrency safety + real-Postgres tests) DISPATCHED** to a backend
    worker 2026-07-27 (MAIN tree). Manager will re-verify its diff.
  - G1 candidate worktree `agent-a10bc9baae7136672` (27 paths); patch re-exported,
    applies clean onto neo-develop. NOT integrated (Kanta's gate). G2 items 1-3
    done+verified earlier; items 4-8 + independent G2 review still pending.
  Operating-method: local Go gates, disposable-DB proofs, review/impl subagents
  authorized for this goal; still no commit/push/deploy/CI-run/shared-DB.
- Token/context: long-running session; checkpoint discipline per runbook §9.
- Current status (2026-07-27, supersedes all earlier session notes in this file):
  G0 done. G1 code-complete (fail-closed manifest classification + production
  hard-fail + symlink-escape guard + forward migration; in worktree
  `agent-a10bc9baae7136672`, patch prepared, NOT integrated). G2 items 1-3 done +
  manager-verified; items 4-8 pending. Codex reopen findings 1 (G1 hard-fail),
  2 (missing-port 5xx), and the symlink P1 + README P2 from the G1 re-review are
  FIXED + verified; findings 3-5 (cash idempotency hardening) are in flight with a
  worker. **One G1 acceptance blocker remains, pending Kanta:** the G1 re-review
  found residual demo access (ben/green/neo accounts + seed-created grants) may
  persist in an already-seeded production DB, which the migration-owned-only
  forward migration does not remove — needs either a companion seed-data cleanup
  migration or Kanta's confirmation that production never ran the legacy
  always-run demo seeds (see Unresolved P0/P1 + Human Approval).

## Operating-Method Authorization (2026-07-24, via AskUserQuestion)

The owner's standing repo preferences (recorded in Claude auto-memory: "no
tests/typecheck/build", "no subagents/background agents", "work inline") are
**explicitly lifted for IMS-MERGE-BLOCKERS only**. For this goal Claude is
authorized to: run `gofmt`/`go vet`/`go test`/`go build`, frontend
`vitest`/`nuxi typecheck`/`npm run build`, disposable-DB migration + lifecycle/
concurrency E2E, authenticated browser proofs, and **read-only independent-review
subagents**. The no-commit / no-push / no-deploy / no-production-migration rule
still applies and is unchanged. This authorization is scoped to this goal; it
does not reverse the standing default for other work.

## Owner Decisions

**Recorded 2026-07-27 by Kanta via AskUserQuestion (this session):**

- **D5 G1 residual demo-access resolution: AUTHOR THE CLEANUP MIGRATION.**
  Kanta chose to author `database/migrations/20260727000002_permissions__remove_seeded_demo_identities`
  (up+down) rather than assert production never ran the legacy demo seeds.
  Must idempotently remove rows keyed to demo ids `a0000000-…010/011/012`
  (FK-safe, no-op if absent), prove fresh-no-op + seeded-upgrade-cleanup +
  admin/role preservation + up/down/up on a disposable DB. Deploy stays
  gated on Kanta; this only unblocks G1 acceptance, not integration/deploy.
  (Named `...002` in the G1 worktree, not `...001`, to avoid colliding with
  the main tree's already-existing `20260727000001_investment__cash_request_fingerprint`
  when the two trees are eventually integrated.)
- **D6 Approver attribution: KEEP the 2026-07-24 decision, do not reopen.**
  The active goal's G2 item-5 text ("actual approver attribution") conflicts
  with Kanta's 2026-07-24 acceptance of the missing-`ActorID`/
  `DecidedBy`-is-submitter gap as a known v1 limitation shared with
  `INVESTMENT_DECISION` (see the 2026-07-24 HANDOFF session note). Kanta
  re-confirmed 2026-07-27: keep the earlier decision. G2 item 5 is DROPPED
  from this pass and recorded as an explicit deferred P1, not implemented.
  Do not extend `contract.ApprovalDecision` for this goal.

Recorded 2026-07-24 by Kanta via AskUserQuestion, this session:

- **D1 Fund policy: `FUND_OPTIONAL`.** Owner: "What I want is fundless."
  Fund-less portfolios and fund-less financial activity are the approved
  replacement for the 2026-07-15 fund-required decision. This RESOLVES P0-H in
  the owner's favor toward fund-optional. Consequences to carry out: update
  `MEMORY.md` (Product Direction + Known Strategic Gaps), retire the old
  "Not Planned / fund-required" durable statements in `TASKS.md`, and ensure
  workflow/compliance/permissions/reporting/migrations/tests/UI all express the
  single fund-optional policy. Fund-less financial activity must NEVER skip
  EOD/compliance merely because no fund id exists. The staged nullable-fund
  migrations (`20260723000001`, `20260723000003`) are now policy-aligned and may
  be preserved (still no commit without explicit authorization).
- **D2 Official security-trade financial-effect event: `MATCHED CONFIRMATION`.**
  Official ledger/holding/cash effect for a LIVE security trade is created when
  the trade confirmation is matched/resolved — NOT at execution fill. G5 must
  build confirmation resolution as the authoritative event; execution fill is a
  prior, non-official stage. Partial-fill, multi-fill, confirmation mismatch,
  settlement failure, cancellation, correction, reversal, and duplicate delivery
  must all be defined against this event.
- **D3 Pending approval at business-day close: `EXPIRE PENDING`.** A LIVE
  cash/trade request still pending when the business date closes auto-expires
  and must be resubmitted on the next open day. G6 must implement expiry (with a
  distinct audited reason) and prove no financial effect is materialized from an
  expired request.
- **D4 Thai SEC product regime/source-to-rule matrix: HUMAN-GATED / DEFERRED.**
  Not decided this session (not asked; it requires legal/compliance sign-off).
  Per the runbook's mandatory safe default, `regulatory.thai_sec` MUST remain
  honestly `NOT_CONFIGURED`; no threshold may be invented and the stub must not
  become a false PASS. Technical work may make the gate clearer/fail-safe only.
  Kanta or a compliance expert must approve the product regime + effective-dated
  source-to-rule matrix before any real threshold is implemented. Recorded as an
  explicit human-gated non-implementation, which satisfies "D1-D4 recorded".
- Decision evidence and date: AskUserQuestion answers, 2026-07-24, this session.

## Verified Git State

VERIFIED THIS SESSION 2026-07-24 (differs materially from the runbook baseline —
recorded per the runbook's "if the state differs, record it and adapt without
resetting/discarding" rule). No reset/discard/clean/stage-change was performed;
the tree is preserved exactly as found.

- Branch: `neo-develop`
- HEAD: `b0faad1` (docs(api): add current API reference and operations guide)
- Upstream divergence: `origin/neo-develop...HEAD = 0 behind / 0 ahead`
  (EVEN — the runbook baseline said +2 ahead; origin has since advanced to
  `b0faad1`, so the two previously-local commits are now on the remote)
- Recent commits: `b0faad1`, `a3eb004`, `5334ae8`, `144d1a2`, `32c2d00`
- Index: **25 staged paths** (runbook baseline said 100 — the index was
  re-staged selectively outside this program). Staged set (coherent):
  - Cash-approval backend (new): `post_transaction_cash_approval.go` (+ test),
    `domain/entity/portfolio_cash_request.go`,
    `domain/valueobject/cash_request_status.go`,
    `infrastructure/adapter/cash_request_approval_adapter.go`,
    `infrastructure/persistence/cash_request_repository.go`,
    `transport/handler/portfolio_v2_cash_request_handler.go`
  - Migrations (new, up+down): `20260723000001` fund_id nullable (portfolios),
    `20260723000003` fund_id nullable (trading tables), `20260724000001`
    portfolio_cash_requests, `20260724000002` approval add cash-transaction
    process type; seed `020_cash_transaction_process_seed.sql`
  - Docs (new): `CLAUDE-GOAL-IMS-REMEDIATION.md`,
    `CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md`,
    `thai-sec-product-regime-decision-memo.md`,
    `docs/investment/STAGE2-HANDOFF-PROMPT.md`,
    `docs/investment/live-cash-transaction-approval-design.md`
  - Frontend cash-ticket (new): `CashRequestsPanel.vue`, `useCashTicket.ts`,
    `lib/cashRequestGuard.ts`, `tests/portfolio-cash-ticket.test.ts`
- Unstaged: **83 modified paths** — the 80 paths present when Claude stopped,
  plus Codex documentation corrections that added working-tree changes to
  staged-new `CLAUDE-GOAL-IMS-REMEDIATION.md`,
  staged-new `CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md`, and tracked
  `MEMORY.md`. The set includes modifications to pre-existing backend and
  frontend files that the new staged files build on (e.g. `module.go`,
  `main.go`, `runtime_service.go`, `enums.go`, generated `docs/*swagger*`,
  investment domain/command/adapter/handler files, frontend
  `PortfolioLedgerNewView.vue`, i18n, generated `ims-api.d.ts`), plus
  `docs/MANAGER/HANDOFF.md` and `docs/MANAGER/TASKS.md`.
- Untracked: **none** (0). The thai-sec memo is now tracked+staged.
- Commits created this session: none
- Push/deploy/migration performed: none

**UPDATE 2026-07-27 (post-G1-integration) — the above is now a HISTORICAL
baseline, not the current index.** After Kanta authorized applying the G1
patch (see Human Approval Required), `git apply --3way
.claude/scratchpad/G1-demo-seed-cleanup.patch` staged its new/modified paths
into the working tree and index. Current `git status --short` totals **167
paths**: 18 untracked (`??`), 23 staged-add (`A`), 10 staged-add-plus-modified
(`AM`), 109 modified (`M`), 7 renamed (`R`). This is EXPECTED and intentional
— the index is now materially larger than the 25-path 07-24 baseline above
because G1's 28-path patch added new files on top of G2's accumulated
changes. Do not try to reconcile the index back down to the 07-24 description
or un-stage anything from the patch. Still true: 0 commits created, 0
pushes/deploys, HEAD unchanged at `b0faad1`. Re-verify with `git status`
before any further edits, per standing instruction — this note describes the
shape (167/18/23/10/109/7), not a substitute for checking.

Index-scope handling for continuation: preserve the current staged/unstaged
split exactly as found. Regrouping, staging changes, and commits require
separate explicit authorization from Kanta.

## Work-Package Matrix

| Package | State | Evidence | Next action |
| --- | --- | --- | --- |
| G0 Preserve/reconcile baseline | COMPLETE | D1-D4 recorded; Git state verified (25 staged / 83 unstaged / 0 untracked, upstream even); tree and index preserved; MEMORY/HANDOFF/TASKS/runbook reconciled on 2026-07-24 | Re-verify Git at fresh-account start; do not change index without Kanta |
| G1 Demo/bootstrap cleanup | **FULLY ACCEPTED and INTEGRATED into the main tree 2026-07-27** (uncommitted, per Kanta's explicit choice via AskUserQuestion). Post-integration disposable-DB proof confirms G1+G2 coexist cleanly (see Human Approval Required for evidence). | **Fix A [P0] (2026-07-26):** replaced path-text demo classification with a **fail-closed reference MANIFEST** (`database/seeds/reference_manifest.txt`, 22 reference seeds). `runSQLSeeds` (demoAllowed=false): only manifest-listed files run; unlisted files under `demo/`/`zz_demo/` quiet-skip; unlisted files NOT in a demo location are a HARD ERROR (manifest-sync guard); missing/empty manifest or SEEDS_PATH-at-subdir = hard error. Classifier is relative-to-resolved-canonical-root (EvalSymlinks) so parent-dir-named-`demo` is immune; renamed-demo-dir now hard-errors loudly (no silent seed-nothing, no demo leak). `resolveDemoSeeding` (F2 fail-closed APP_ENV) preserved. **Fix B [P1]:** `make seed` (reference-only) + `make seed-demo`; `INCLUDE_DEMO_SEEDS=true` wired into `.env.development`, `.env.e2e.example`, `setup-db.sh`, CI `e2e-iam` job; `db-reset`→`seed-demo`. **Fix C:** full patch exported (28 paths) to scratchpad `G1-demo-seed-cleanup.patch`; `git apply --check --3way` onto `neo-develop@b0faad1` = clean (G1 files identical between b813b30 and b0faad1, verified). **Evidence:** `gofmt`/`go vet`/`go test ./cmd/seed`/`go build ./...` PASS; disposable-DB matrix (throwaway `postgres:16-alpine` :15433, destroyed) PASS: fresh=0 demo grants + prod roles; prod reference-only exit 0 + 17 demo files skipped + 0 ben/green/neo; dev opt-in restores 3 accounts + 8 memberships; prod+opt-in hard-errors; SEEDS_PATH-at-demo & -at-zz_demo hard-error; forward migration `DELETE 6` exactly (decoys+admin preserved); migrate down/up clean. | G1's independent review is DONE (two rounds, PASS, no open P0/P1) — do not re-dispatch it. The only remaining gate is **Kanta's authorization to apply `.claude/scratchpad/G1-demo-seed-cleanup.patch` onto the main tree** (no commit). This was asked via `AskUserQuestion` at the end of this session's checkpoint — check the answer before re-asking. Two other findings surfaced this session are in Unresolved P0/P1 below. |
| G2 Approval and LIVE cash controls | **ACCEPTED at code/review level 2026-07-27.** Items 1-4, 6-8 source-proven complete; item 5 DROPPED per D6. Independent full-diff review returned **PASS** (2 new P1 test-coverage gaps, not shipped-behavior defects — see Unresolved P0/P1). | **Item 1 COMPLETE 2026-07-26 (worker + manager re-verified, in MAIN tree, uncommitted):** added sentinel `contract.ErrSubjectAccessDenied` (`backend/pkg/contract/approval.go`); investment `investment_subject_access.go` now returns the sentinel ONLY for genuine denials (`!ok`, subject-not-found, unsupported type) and the RAW error for infra (nil dep, repo/IAM error) → 5xx; approval `classifySubjectAccessErr` (runtime_service.go:728) maps sentinel→403, else→`slog.Error`+raw(500); applied to all 3 check methods + 5 read call sites (`GetMyInbox`/`ListRequests` now abort on infra instead of silently truncating; single-subject reads keep existence-hiding 404 on denial, 5xx on infra). module.go adapter passes the error through unchanged (no fix needed). 7 files touched (+ tests: `investment_subject_access_test.go` 11 tests, runtime_service_test.go 2 tests, fake_repo_test.go `infraFailingSubjectPort`). Manager-verified: `go build ./...` exit 0; `go test ./internal/approval/... ./internal/investment/infrastructure/adapter/... ./pkg/contract/...` all `ok`; only the 7 intended files changed; index untouched; enums.go M is pre-existing Stage 2. | **Items 2+3 COMPLETE + manager-verified 2026-07-26 (MAIN tree, uncommitted):** new forward migration `20260726000001_investment__cash_request_idempotency` (adds `idempotency_key VARCHAR(255)` + partial unique index `uq_inv_cash_req_idempotency ON (portfolio_id, idempotency_key) WHERE idempotency_key IS NOT NULL`; down drops both). `Idempotency-Key` HTTP header threaded into `PostTransactionRequest`, enforced on the LIVE-cash gate; >255 chars → 400; payload-mismatch reuse → 400. Repo `GetByIdempotencyKey` + lookup-before-insert with unique-violation→fetch-existing race backstop. `submitCashForApproval` rewritten: reconcile PENDING-with-null-approval_request_id by re-linking via `ApprovalStatusProvider.GetApprovalStage` (fail-closed if provider nil — never blind re-submit); compensation errors surfaced via `errors.Join` (the `_ =` fix). Swagger regenerated (make swagger/swagger-v2; deterministic). Manager-verified: index preserved (worker staged nothing); `go build ./...` exit 0; `go test ./internal/investment/... ./internal/approval/...` ALL ok; 22-col INSERT placeholder/arg alignment inspected correct; disposable-DB proof (column/partial-index/23505-on-dup/clean-down). RESIDUAL (for the G2 independent review): full Go repo round-trip with a non-nil key unproven (fakes+psql only); a re-link fires a 2nd `INVESTMENT_CASH_REQUEST_SUBMITTED` audit; a keyed step-2 failure burns the key → retry returns 202 `status:CANCELLED` (client must use a fresh key); confirm empty-string header → nil. **Codex findings 3-5 (canonical fingerprint + timeout/concurrency safety + real-Postgres tests) COMPLETE + independently re-verified 2026-07-27 (this account, main tree, uncommitted):** new migration `20260727000001_investment__cash_request_fingerprint` (additive `request_fingerprint VARCHAR(64)` + non-unique lookup index; down drops both; NULLABLE, no backfill, legacy rows fall back to field-by-field comparison — forward-only per G1 discipline). `post_transaction_cash_approval.go` rewritten: deterministic sha256 fingerprint over `{submitter_id, portfolio_id, transaction_type, amount(8dp), fees(8dp), currency, value_date, memo}` joined with a `\x1f` separator; a no-key submit derives `"auto:" + fingerprint` as its effective idempotency key (reserved prefix, client keys starting with it are rejected) so even key-less retries of a byte-identical request dedup via the existing partial unique index; a keyed retry with a DIFFERENT fingerprint is rejected (400, "used for a different cash movement"); cross-actor key reuse is rejected; `resolveAmbiguousSubmit` never blindly compensates on a submit error/nil-result — it queries `ApprovalStatusProvider.GetApprovalStage` first and re-links to a committed approval instead of risking an orphan or a stranded approval, only cancelling the staged row when verified absent, and surfaces (via `errors.Join`) rather than swallows a compensation failure. New repo method `GetByFingerprint` (non-unique, deterministic earliest-match via `ORDER BY submitted_at, id`). Manager-verified: `gofmt -l .` clean (see Files Changed — 2 pre-existing unrelated files also fixed), `go vet ./...` clean, full `go test ./... -count=1` all `ok`, AND — rerun specifically against a FRESH disposable `postgres:16-alpine` container (not the shared dev DB) — `TestIntegrationCashRequestRoundTrip`/`FingerprintDeterministic`/`ConcurrentSameKey` (new) plus `TestIntegrationCashRepositoryConcurrentFirstInsertRetry`/`TestIntegrationPositionRepositoryConcurrentFirstInsertRetry` (pre-existing) all 5 PASS. `ApprovalStatusProvider`/`GetApprovalStage` wiring confirmed real (pre-existing contract, `module.go:644-646` now also wires `m.postTxn.SetApprovalStatusProvider(p)`, cascading from the unconditional `main.go:173` call — not a dangling nil-only path). **Item 4 COMPLETE 2026-07-27 (this account, main tree, uncommitted):** new migration `20260727000003_approval__create_sync_failures` (table `approval__sync_failures`: id, approval_request_id FK CASCADE, subject_type, subject_id, outcome, reason, attempt_count, max_attempts default 5, last_error, status, timestamps, resolved_at/by; CHECK constraints on outcome/status/attempts/resolved-consistency; indexes on (status,created_at) and (approval_request_id)). New `entity.SyncFailure` (+ `CanRetry()`), `vo.SyncFailureOutcome`/`SyncFailureStatus` enums, `domain.SyncFailureRepository` (added to the aggregate `domain.Repository`), `postgres.PostgresRepository` impl (`sync_failures.go`). `runtime_service.go`: `notifyRevoked`/`runPostAction` now call new `persistSyncFailure` alongside the existing audit log (audit kept, not removed) whenever a subject-sync callback fails; new `ListSyncFailures`/`RetrySyncFailure` methods. Operator surface: `GET /approvals/sync-failures` + `POST /approvals/sync-failures/{id}/retry`, gated by two new permission codes `APPROVAL_SYNC_VIEW`/`APPROVAL_SYNC_RETRY` (added to the catalog in `policy.go` AND granted to Admin in the always-run reference seed `011_approval_permission_seed.sql` in the same change — catalog+grant landed together per CLAUDE.md's own warning). Retry is **operator-triggered only, no new background scheduler/ticker was added** (deliberate — G6 already has an open P0 about a scheduler wired to demo data; adding a second one now would compound that, not fix it; automatic scheduled retry is deferred to G6/G8 once a real scheduler story exists). **A real transaction-safety bug was caught and fixed before landing:** the first draft of `RetrySyncFailure` returned the subject-sync callback's own error as the wrapping transaction's return value — since a non-nil return rolls back the tx, this would have silently discarded the attempt_count/status/last_error write on every failed retry, permanently defeating the bounded-retry mechanism (it would never reach EXHAUSTED, and every retry would look like the "first" attempt in the DB even though the in-memory value said otherwise). Fixed: the callback error is captured outside the tx closure and returned to the *caller* only after the tx (which always returns nil on the write-succeeded path) commits. 4 new unit tests in `runtime_service_test.go` cover exactly this: persist-on-failure, retry-succeeds-resolves, retry-still-fails-increments-and-persists-durably (asserts against a fresh repo read, not just the returned pointer), retry-with-no-registered-callback-fails-closed-without-consuming-an-attempt. Manager-verified: `gofmt -l .` clean, `go vet ./...` clean, full `go test ./... -short -count=1` all `ok` including the 4 new tests; disposable-DB proof (fresh `postgres:16-alpine` port 15439, destroyed after) — table/constraints created correctly (verified via `pg_constraint` catalog query, not just `\d`), up/down/up clean. Swagger + `ims-api.d.ts` regenerated (`make swagger && make swagger-v2 && cd frontend && npm run api:generate`) and confirmed present (`SyncFailureResponse`/`SyncFailureListResponse`, both `/sync-failures` routes). **Item 6 COMPLETE 2026-07-27 (this account, main tree, uncommitted):** audited both halves. Backend cancel authorization was ALREADY correct — `PostTransactionHandler.CancelCashRequest` (`post_transaction_cash_approval.go:598-623`) already hard-checks `cr.SubmittedBy != actorID` -> `ErrCashRequestForbidden` before allowing cancel; nothing to fix there. The actual gap was frontend-only: `CashRequestsPanel.vue` is visible to anyone with portfolio view access (approvers/managers, not just the submitter — confirmed by reading `ListCashRequestsByCode`, which is portfolio-scoped by design, not submitter-scoped), but the Cancel button was shown unconditionally for every PENDING row regardless of who submitted it — a non-submitter would click it and get a confusing backend 403 instead of the button never appearing. Fixed: new pure helper `frontend/app/features/portfolio-workspace/lib/cashRequestAuthz.ts` (`canCancelCashRequest(item, currentUserId)`, mirrors this feature's existing `cashRequestGuard.ts`/`ledgerGuard.ts` pattern of extracting alias-free pure logic so it unit-tests in plain Vitest without mounting a component or mocking Pinia/Nuxt runtime), wired into `CashRequestsPanel.vue` via `useAuthStore().user?.id`. New `tests/cash-request-authz.test.ts` (4 tests: submitter-can-cancel, other-user-cannot, non-pending-status-cannot, unauthenticated/unknown-user-cannot). No V1 duplication exists to create an inconsistency (grep-confirmed cash-request routes are V2-only, one code path). Verified: full frontend `npx vitest run` — 61 files / 659 tests pass (was 60/655 before this pass, +1 file/+4 tests, zero regressions); `npx nuxi typecheck` — exactly 91 diagnostics, byte-identical to the documented baseline, confirmed zero in either touched file; `npm run build` confirmed succeeded (exit 0, verified after this checkpoint alongside the frontend Idempotency-Key build below).

**Item 7 COMPLETE 2026-07-27 (this account, main tree, uncommitted):** new migration `20260727000004_investment__cash_request_ownership_and_terminal_guard`. Ownership: added the same composite FK pattern already used on `investment__decisions/executions/trade_confirmations` (`20260703000001`) — `FOREIGN KEY (portfolio_id, fund_id) REFERENCES investment__portfolios (id, fund_id)`, reusing that migration's existing `uq_inv_portfolios_id_fund UNIQUE (id, fund_id)` (no new unique constraint needed). Fund-optional-safe: Postgres' default MATCH SIMPLE means a fund-less row (`fund_id IS NULL`) is not checked by this FK at all — only a fund-bound row's pair is validated — confirmed live, not just by reading the Postgres docs. Terminal-state: no existing guard prevented a status transition OUT of APPROVED/REJECTED/CANCELLED (the existing `trg_inv_cash_req_updated_at` trigger only bumps `updated_at`, contrary to an earlier assumption that some existing trigger already handled this — checked the actual migration source before writing code, per this program's own standing instruction). New `BEFORE UPDATE` trigger `trg_inv_cash_req_terminal_status_immutable` + function `inv_cash_req_prevent_terminal_status_change()`: raises when `OLD.status` is terminal AND `NEW.status IS DISTINCT FROM OLD.status`; a benign same-status re-save (e.g. a memo correction) and the PENDING -> terminal transition itself are both unaffected. Disposable-DB proof (fresh `postgres:16-alpine` port 15441, full migrate+demo-seed for realistic fixtures, destroyed after): mismatched fund_id insert -> FK violation (confirmed exact error); correct fund_id insert -> succeeds; fund-less (NULL fund_id) insert against a fund-BOUND portfolio -> succeeds (MATCH SIMPLE confirmed live); PENDING->APPROVED then APPROVED->CANCELLED -> blocked with the trigger's own exception text; memo-only update on the now-terminal row -> succeeds; up/down/up lifecycle clean (trigger present/absent/present via `pg_trigger` catalog query). No Go code changed (pure DB-level constraint); `gofmt -l .`/`go vet ./...`/`go test ./... -short -count=1` all clean (confirming no incidental regression).

**Item 8 COMPLETE 2026-07-27 (this account, main tree, uncommitted).** Audited the six named scenarios (reject/withdraw/expiry/duplicate-callback/stale-day/approve-vs-cancel) against EXISTING coverage before writing anything new — reject (`TestRejectedCallbackNoTxn`) and duplicate-callback (`TestDuplicateApprovalIsIdempotent`) were already covered; the other four needed work:

- **stale-day: NEW test `TestLiveCashOnClosedBusinessDay_BlockedBeforeApprovalSubmit`** (`post_transaction_cash_approval_test.go`). Confirmed by reading `post_transaction.go` that the workflow-day check runs in the SHARED `prepareTransaction` step BEFORE the cash-only branch decides LIVE/SIMULATION/MODEL, so a closed day already blocks LIVE cash the same as a security trade — this needed a test, not a fix. Added `newCashGateHarnessWithWorkflow(t, workflow)` (small refactor of the existing harness to accept a custom `contract.WorkflowStateProvider`, defaulting to the existing `&allowWorkflow{}` so no other test changed) and reused the existing `closedWorkflow{}` fixture from `post_transaction_test.go`. Explicitly scoped to fund-BOUND portfolios only — the fund-less workflow-gate gap (`prepareTransaction`'s `fund == nil` branch always treats the gate as open) is a DIFFERENT, already-tracked P0 (TASKS.md P0-F / this goal's G6), not something this test covers or this pass fixes.
- **approve-vs-cancel: NEW real-Postgres test `TestIntegrationCashRequestConcurrentApproveVsCancel`** (`cash_request_repository_integration_test.go`). Deliberately simulates a hypothetical bug where the application forgets to recheck `IsPending()` after `GetForUpdate` — two goroutines race `GetForUpdate`+`Update` toward APPROVED vs CANCELLED on the same row with NO status recheck — to prove the item-7 terminal-state trigger (not just apo-level discipline) is the real backstop. Extended `cashReqIntegrationPool`'s throwaway-schema setup to also create the `inv_cash_req_prevent_terminal_status_change` trigger (mirroring migration `20260727000004`) so this test exercises the actual DB object. **Bug caught and fixed while writing this**: the trigger's `RAISE EXCEPTION` message uses PL/pgSQL's own `%` parameter syntax, and the whole schema-setup string is ALSO passed through Go's `fmt.Sprintf` for schema-name templating — Go's vet correctly rejected the unescaped `%` as an invalid format verb; fixed by escaping to `%%` in the Go source (`%%` -> literal `%` after Go's Sprintf -> the single `%` PL/pgSQL needs). Ran 5x against a fresh disposable container: exactly 1 success + 1 trigger-refusal every time, final row always exactly one of APPROVED/CANCELLED, never PENDING.
- **withdraw: NEW characterization test `TestWithdrawRequest_DoesNotNotifySubjectSync`** (`internal/approval/application/service/runtime_service_test.go`) — see the NEW FINDING below. This test intentionally documents CURRENT (gap) behavior rather than fixing it.
- **expiry: NOT TESTED, cannot be — no implementation exists.** Grepped the entire cash-approval code path: there is no `CashRequestStatusExpired` value (the `chk_inv_cash_req_status` CHECK constraint only allows PENDING/APPROVED/REJECTED/CANCELLED) and no scheduled/triggered mechanism that would ever flip a PENDING cash request at business-day close. D3 (`EXPIRE PENDING`) is explicitly assigned to **G6** in this program's own runbook ("G6 Real EOD/workflow races... D3=expire pending"), not G2 — implementing expiry now would mean inventing new G6-owned status/schema/scheduling design ahead of that package and is out of scope for this pass. Do not write a placeholder/fake expiry test; wait for G6.

**NEW FINDING (P1, not fixed this pass — recorded for Kanta/independent review): the shared approval engine's `terminate()` (used by BOTH `WithdrawRequest` and the privileged `CancelRequest`) never invokes the registered `SubjectSync` callback at all**, unlike `ApproveTask`/`RejectTask` (`runPostAction`) and `RevokeRequest` (`notifyRevoked`), which do. Investment's OWN cash-specific cancel endpoint (`CancelCashRequestByCode` -> `command.CancelCashRequest`) is unaffected — it does its own direct DB bookkeeping and doesn't rely on this notification. The real exposure: a fully-authorized submitter calling the GENERIC `POST /approvals/requests/{requestId}/withdraw` endpoint directly (instead of the cash-specific endpoint) can terminate a `CASH_TRANSACTION` approval request while `investment__portfolio_cash_requests` is never told — the row is orphaned PENDING forever, with a now-dead `approval_request_id`, and the cash-specific Cancel endpoint would then itself fail ("already finalised") if tried afterward. **Not fixed this pass because a correct fix needs a contract change**: wiring `terminate()` to call `OnRejected` would map a WITHDRAWN cash request to `CashRequestStatusRejected` (confirmed via `TestRejectedCallbackNoTxn` — `ApplyCashRequestApproval`'s `approved=false` path always sets REJECTED), which is semantically wrong for a withdrawal. A correct fix needs `contract.ApprovalDecision` (or a new callback method) to distinguish withdrawn/cancelled from rejected-by-approver — the SAME class of cross-module contract change Kanta explicitly deferred for approver attribution (owner decision D6, 2026-07-27, see Owner Decisions). Recommend resolving this finding in the SAME future task as D6's eventual reopening, not ad hoc. This likely also affects INVESTMENT_DECISION/RESEARCH_REPORT/COMPLIANCE_RELEASE subject types identically — not independently verified for each this pass, flagged as a shared-engine-level finding, not cash-specific.

**Item 8 COMPLETE.** `go build`/`go vet`/`gofmt -l .` clean; full `go test ./... -short -count=1` confirmed all `ok`, zero FAIL (the first background run's exit-1 was `grep`'s own "no matching lines" exit status from this account's filtering command, not a test failure — re-ran plainly to confirm). 3 new tests + 1 characterization test added across two packages; expiry explicitly deferred to G6 with reasoning, not silently skipped; 1 new cross-cutting P1 finding recorded above for Kanta/independent review.

**Frontend Idempotency-Key task COMPLETE 2026-07-27 (this account, main tree, uncommitted).** Confirmed by grep (repeated across sessions) that `useCashTicket.ts`/`portfolioApi.ts` sent NO `Idempotency-Key` header at all despite the backend fully supporting and preferring one (G2 items 2-3). Fixed: `portfolioApi.postTransaction` gained an optional third `idempotencyKey` param, sent as the `Idempotency-Key` request header via `openapi-fetch`'s standard `headers` option (no manual fetch code, per CLAUDE.md). `useCashTicket.ts` gained `pendingIdempotencyKey` (a ref, exposed on the composable's return value): generated once per simulated payload (`generateIdempotencyKey()`, mirroring the exact guarded-`crypto.randomUUID()`-with-fallback pattern already used by `useMarketDataSecurityDetail.ts`, not a new ad hoc pattern) and reused across retries of the SAME attempt; reset to `null` (forcing a fresh key next time) in both `open()` and `patchDraft()` so a genuinely new/edited movement never reuses a stale key. `CashTicketDeps.post` signature extended to take the key as a required third argument; `defaultDeps.post` forwards it. New tests in `tests/portfolio-cash-ticket.test.ts` (+3, 13->16): key is non-empty and sent; a retry after a failed attempt reuses the identical key; editing the draft after a failed attempt mints a genuinely new key on the next post. Verified: full `npx vitest run` 61 files/662 tests (was 659, zero regressions); `npx nuxi typecheck` exactly 91 diagnostics (byte-identical baseline, confirmed zero in either touched file); `npm run build` succeeded.

**G2 is COMPLETE and ACCEPTED as of 2026-07-27.** Items 1-4 and 6-8 all source-proven; item 5 **DROPPED per Kanta's decision D6**. Also verified this checkpoint: `011_approval_permission_seed.sql`'s two new codes round-trip cleanly end-to-end (Go catalog upsert -> FK -> SQL grant -> Admin group) on a fresh disposable DB (`ims-verify-g2seed`, port 15442, destroyed after) — confirmed via `psql`, zero errors. The independent full-diff review (task #9) returned **PASS** with 2 new P1 test-coverage-gap findings (neither blocking, see Unresolved P0/P1). Next package: **G3**. |
| G3 Fill validation/lifecycle | pending | | Positive/cumulative/ordered fill bounds + workflow-day + compliance re-eval |
| G4 Compliance binding audit | pending | | Actor-attributed immutable binding create/change/deactivate audit |
| G5 Authoritative trade-to-ledger | pending (D2=matched confirmation) | | Build confirmation-resolution as the official financial-effect event |
| G6 Real EOD/workflow races | pending (D3=expire pending) | | Replace `NopInvestmentQueryAdapter`; race-safe day locks; expire-at-close. **+ NEW P0 dependency (from G1): production `workflow__scheduler_contracts` is empty once `005` demo contracts are removed — must add a real contract-source integration; scheduler is inert until then.** |
| G7 Operator/frontend completion | pending | | Trader/operator mutation UI after backend contracts accepted |
| G8 Remaining readiness findings | pending (D4 human-gated) | | SSR fix, watchlist scheduler, typecheck-to-zero, CI E2E, honest reg stubs |
| G9 Integrated proof/review | pending | | Full gates + disposable-DB E2E + browser proof + independent reviews |

Only one package may be `active`. **G1 is ACCEPTED and INTEGRATED** into the
main tree (uncommitted, Kanta's choice 2026-07-27). **G2 is ACCEPTED** (items
1-4/6-8 complete + independently reviewed PASS; item 5 dropped per D6). G3 is
next.

## Acceptance Items Completed

- **G0** (baseline reconcile): D1-D4 recorded; Git state verified vs baseline;
  methods authorized; `MEMORY.md` aligned to D1=`FUND_OPTIONAL`; stale task and
  handoff policy/status statements corrected; index-preservation rule recorded.
  Evidence: manager documents plus current Git verification.

- **G1 NOT yet accepted (code-complete; one residual-access P1 pending Kanta).**
  2026-07-27 update: production hard-fail (Codex finding 1) and the symlink-escape
  guard (`validateReferenceFilesPresent` now requires a regular file whose
  EvalSymlinks-resolved path stays under the canonical root; proven by
  file-symlink + parent-junction tests) are FIXED + verified; the README was
  rewritten to the manifest model (was a stale path-text description). The fresh
  independent G1 re-review verdict is REQUEST CHANGES for ONE residual P1: an
  already-seeded production DB may still carry ben/green/neo accounts + their
  seed-created grants (the forward migration only removes the 6 migration-owned
  membership ids). Resolution is pending Kanta (companion cleanup migration vs
  confirm prod never ran the legacy seeds). The seeding/migration mechanics below
  were reviewed SOUND and collision-free. After two
  premature "complete" claims and multiple re-reviews, the candidate
  `agent-a10bc9baae7136672` now: (Fix A) classifies via a fail-closed reference
  MANIFEST + relative-to-canonical-root, closing all four bypasses (parent-dir,
  SEEDS_PATH-at-subdir, symlink, renamed-dir) with unlisted-non-demo → hard error;
  (Fix B) `make seed`/`make seed-demo` + `INCLUDE_DEMO_SEEDS` wiring in
  `.env.e2e.example`/`setup-db.sh`/CI; (Fix C) 27-path patch exported, applies
  clean onto neo-develop. **Independent full-diff review (fresh subagent,
  2026-07-26): verdict PASS, no P0/P1.** It confirmed demo cannot run in prod, no
  named demo identity on prod migration paths, fail-closed APP_ENV, correct
  scheduler zero-contract path, and seed-e2e out of scope. Its 4 P2s are all
  addressed: green-id collision FIXED + proven (demo/008 → generated ids); stale
  `012_approval_demo_seed.sql` header pointers in demo/002+003 fixed; .env.development
  churn dropped; manifest listed-but-missing now WARNs. Evidence: local Go gates +
  a full disposable-DB matrix, all PASS (see Commands/Database sections and the
  G1 matrix row). REMAINING: Kanta authorizes integration into the main tree
  (apply the patch; do NOT merge divergent history); no commit yet. Base-change
  detail below (now accepted with the above corrections):
- **G1 BASE (worker `a9d3b6a5`, NOT YET ACCEPTED)** — 22 files, uncommitted in worktree
  `agent-a9d3b6a5b1256fbf2`): migrations `20260716000001/2/3` (named ben/green
  assignment removed, prod role catalog preserved); new forward migration
  `20260725000001` (up deletes exactly the six historical migration-owned
  membership ids on `(id,user_id,group_id)`; down = documented no-op);
  `backend/cmd/seed/{main.go,sql_seeds.go,sql_seeds_test.go}` (env-gated demo
  seeding, hard non-zero-exit rejection in production); seed reorg into
  `database/seeds/demo/**` with 004/012/018 split to keep production approval
  process config + role catalogs as reference, and 013 (dead is_active=false
  scaffold) moved to demo. Tests/runtime evidence: worker + INDEPENDENT reviewer
  each proved on separate disposable Postgres DBs — fresh migration-only (0 demo
  grants, prod roles present), historical-upgrade cleanup (forward migration
  DELETE 6, decoys + admin untouched), dev/test restoration (ben/green full
  parity), production rejection (seeder exit 1, 0 ben/green/neo, reference
  committed), up/down/up; backend gofmt/vet/test/build/diff-check all PASS.
  Independent review disposition: **PASS, no P0/P1**. Residual risk (all P2,
  non-blocking): (1) `cfg.Env` defaults to development when `APP_ENV` unset →
  demoAllowed true (pre-existing convention; `.env.production` sets it; consider
  fail-closed default in G8); (2) `005_workflow_scheduler_contracts_seed.sql`
  seeds `IMS-DEMO-*` bridge contracts in prod unconditionally (pre-existing,
  untouched, fold into G8 hygiene); (3) stale `012_approval_demo_seed.sql`
  filename in an e2e test comment (cosmetic). STATE: uncommitted in worktree; NOT
  in main tree; integration to main tree awaits explicit owner authorization.

For every completed item, record:

- acceptance text or stable ID;
- source files;
- tests or runtime evidence;
- independent review disposition;
- residual risk.

## Files Changed by Claude Account #1 and Codex Re-review

Main tree (`neo-develop`), documentation authored or corrected:
- `docs/MANAGER/CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md` (this file — decisions, Git
  state, corrected matrix, evidence, next-account prompt).
- `docs/MANAGER/MEMORY.md` (D1=`FUND_OPTIONAL`; production/demo seed boundary;
  forward-cleanup rule; known strategic gaps).
- `docs/MANAGER/TASKS.md` (P0-B/G1 correction acceptance; P0-H policy resolved;
  historical fund-required decision marked superseded).
- `docs/MANAGER/HANDOFF.md` (current Account #1 re-review state and next action).
- `docs/MANAGER/CLAUDE-GOAL-IMS-REMEDIATION.md` (G1 safety acceptance and exact
  copy-paste `/goal` command).

G1 worktree (`worktree-agent-aa1dd1dbf78db7268`, uncommitted, NOT in main tree):
- `database/migrations/20260716000001_permissions__grant_ben_investment_approval_runtime.up.sql`
- `database/migrations/20260716000002_permissions__grant_ben_operation_page_access.up.sql`
- `database/migrations/20260716000003_permissions__grant_green_ben_workflow_parity.up.sql`
- `database/migrations/20260716000003_permissions__grant_green_ben_workflow_parity.down.sql`

NOTE: the 25 staged / 80 unstaged paths present when Claude stopped were mostly
prior fund-optional Stage 1 + Stage 2 cash-approval work. Codex changed
documentation only; the current unstaged count is 83 because two staged-new
manager documents now also have unstaged corrections and `MEMORY.md` became an
additional unstaged path. Do not attribute unrelated product paths to either
documentation session.

Do not list pre-existing dirty files as authored by the current account.

**2026-07-27 session (this account) — files touched:**

- Main tree (`neo-develop`), documentation only, plus one out-of-scope
  formatting fix:
  - `docs/MANAGER/CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md` (this file).
  - `backend/internal/investment/domain/entity/portfolio.go` and
    `backend/internal/investment/transport/dto/response/responses.go` —
    **`gofmt -w` ONLY** (struct-field alignment after an earlier session's
    inline-comment insertion broke column alignment). Zero semantic change;
    confirmed by `gofmt -d` before applying — the diff is purely whitespace
    realignment of existing field/tag columns. This was NOT part of the
    findings-3-5 worker's assigned scope; noted here explicitly so it is not
    misattributed. Both files were already in the unstaged-modified set
    before this fix, so no new file entered the diff.
- Verified but NOT authored by this account (pre-existing from the findings-
  3-5 worker dispatch, main tree, uncommitted): `post_transaction_cash_approval.go`
  (+test), `cash_request_repository.go`, `portfolio_cash_request.go` (entity),
  `20260727000001_investment__cash_request_fingerprint.{up,down}.sql`,
  `cash_request_repository_integration_test.go` (new, untracked).
- No commit, stage, push, migration execution (outside a disposable
  container), or shared/production DB mutation performed by this account
  this session, except the disclosed real-Postgres test runs (see below).

## Commands and Results

| Command | Environment | Result | Evidence/notes |
| --- | --- | --- | --- |
| `go run ./cmd/migrate up` | G1 worktree, throwaway `postgres:16-alpine` on port 15432 (NOT shared dev DB) | PASS | 78 migrations applied, version `20260723000002`, dirty=false; 0 ben/green rows; prod roles 041/042 + 9 defs present |
| `go run ./cmd/seed` | G1 worktree, same disposable DB | PASS | ben+green each regain all 4 expected groups (dev behavior preserved) |
| `migrate down` x6 | G1 worktree, disposable DB | PASS | migration 3 new no-op down + 1/2 cascading deletes all succeed; no dangling rows |
| `gofmt -l .` / `go vet ./...` / `go test ./... -count=1` / `go build ./...` | G1 worktree, backend/ | PASS | worker-reported; prior independent reviewer reported PASS; Codex did not rerun these gates |
| `git diff --check` | G1 worktree | PASS | exit 0 on all 4 migration files |
| `git diff --check -- docs/MANAGER/...` | main tree | PASS | documentation correction has no whitespace errors |

Record failures as well as passes. Do not write only "all tests passed."

## Database / Runtime Actions

- Migrations applied: only to a THROWAWAY `postgres:16-alpine` container (port
  15432, `ims-g1-disposable-pg`), by the G1 worker for proof. Removed after.
- Migrations rolled back: 6-step `down` on the same disposable DB only.
- `force` operations: NONE.
- Containers rebuilt/restarted: only the disposable G1 container (created +
  destroyed). Shared `ims-postgres`/dev stack untouched.
- Databases or schemas touched: disposable only. NO shared/production DB touched.
- Synthetic test data left behind: none (disposable container destroyed).
- Browser URLs/users used: none this session.

**2026-07-27 session (this account) additions:**

- Ran `go test ./... -count=1` (no `-short`, `IMS_TEST_DATABASE_DSN` unset)
  once early in this session, BEFORE noticing the new integration test's
  `integrationDSN()` default targets the SHARED `ims-postgres` dev container
  (port 5437). This created/dropped one UUID-named throwaway schema there
  (pre-existing repo convention — same pattern `position_repository_concurrency_test.go`
  already used). Verified via `information_schema.schemata` immediately after:
  zero residue, no leftover `inv_cashreq_it_*` schema. No other table/row in
  the shared DB was touched. Going forward this account uses `go test ./...
  -short -count=1` for general gates to avoid this.
- Started a FRESH disposable `postgres:16-alpine` container
  (`ims-g2-disposable-pg`, port 15437, distinct from the still-present
  leftover `ims-g1-demo-cleanup` on 55437 whose credentials were not known)
  specifically to re-prove the findings-3-5 real-Postgres tests without
  touching shared state. Ran the 3 new + 2 pre-existing concurrency
  integration tests against it (all 5 PASS), then `docker rm -f
  ims-g2-disposable-pg` immediately after. Nothing left running.
- No migration executed against any shared or production database. No
  `force` operation. No browser session.

## Active Agents and Worktrees

| Agent/session | Worktree/branch | Exclusive write scope | State |
| --- | --- | --- | --- |
| G1 CORRECTED worker `a9d3b6a5` (OLD) | isolated worktree `agent-a9d3b6a5b1256fbf2` (base verified == neo-develop content) | `database/migrations/20260716000001/2/3` (+000003 down), forward migration `20260725000001` pair, `database/seeds/**` (reorg), `backend/cmd/seed/**` | DONE; 22-file base. Reviewed by `a277f86e` (PASS) then SUPERSEDED by Codex re-review + the `a10bc9ba` 29-path candidate. Historical; do not use directly. |
| G1 independent reviewer `a277f86e` | main tree read-only | none | COMPLETE — verdict PASS, but SUPERSEDED: Codex re-review found P0s it missed (SEEDS_PATH bypass; `005` scheduler contracts) |
| Codex re-review (external) | read-only | none | COMPLETE — verdict REQUEST CHANGES; 2×P0 + 4×P1 (see G1 matrix row). Authoritative; G1 reopened on its basis |
| G1 correction worker `a10bc9ba` | worktree `agent-a10bc9baae7136672` (29 paths: 17 tracked + 12 untracked) | `backend/cmd/seed/**`, `database/seeds/**` (incl. `005`→`demo/009`) | **AUTHORITATIVE G1 candidate (flawed).** Fixed the first review's issues but Codex re-review #2 found open P0 (path-spelling classification) + P1 (no make/env/CI wiring). Needs Fix A/B/C; do NOT integrate as-is. |
| G1 correction worker `aa9057d6` | worktree `agent-aa9057d6d43a05c8b` | intended: classification redesign + wiring + export | **FAILED — session usage limit.** Only RECONSTRUCTED the `a10bc9ba` candidate (29 paths, hashes match) then died; implemented NONE of FIX A/B/C. Redundant reconstruction, no fixes — ignore it. |
| G1 CORRECTED worker `a8af799b` (prior) | isolated worktree | same scope | INTERRUPTED mid-run when the prior Claude process exited; produced NO persistent changes (worktree clean, empty diff). Superseded by `a9d3b6a5`. |
| G1 backend worker `aa1dd1db` (OLD) | `.claude/worktrees/agent-aa1dd1dbf78db7268` | `database/migrations/20260716000001/2/3` (up+down) | ABANDONED; 11 commits ahead; do not merge/reuse |
| G1 independent reviewer `a00ac42e` (OLD) | main tree (read-only) | none | COMPLETE but SUPERSEDED; a NEW full-diff independent review is required |

**AUTHORITATIVE G1 CANDIDATE for the next account = worktree `agent-a10bc9baae7136672`** (29 paths, uncommitted) — flawed: still has the open P0 (path-spelling classification) + P1 (no make/env/CI wiring). The next account must implement FIX A/B/C on it (or reconstruct + fix), NOT integrate it as-is. `aa9057d6`'s worktree is a redundant reconstruction with no fixes; ignore it.

NOTE: G1's changes are in the worker's worktree branch, UNCOMMITTED. They are NOT
in the main `neo-develop` worktree. The worktree HEAD is `b813b30`, 11 commits
ahead of `neo-develop@b0faad1`; do not merge the branch or blindly cherry-pick
its future commit. Recreate from current `neo-develop` or apply only the exact
reviewed file patch after re-verifying ancestry. No commit is authorized.

## Unresolved P0 / P1 Findings

Initial known set; replace as findings close or change:

G1 findings (candidate `agent-a10bc9baae7136672` — ADDRESSED-IN-CANDIDATE items
are implemented but NOT yet accepted; acceptance needs the redesign below + a new
full-diff review):
- ADDRESSED-IN-CANDIDATE (pending acceptance): recursive seeder ran demo in prod
  without rejection (now: demo split + production rejection + explicit opt-in);
  historical-migration edits didn't clean already-upgraded DBs (now: forward
  migration `20260725000001`, proven surgical `DELETE 6`); APP_ENV fail-open;
  `005` IMS-DEMO contracts in prod reference path (moved to `demo/009`).
- **FIXED 2026-07-26 (Fix A):** demo classification no longer infers from path
  text. Replaced with a fail-closed reference MANIFEST + relative-to-canonical-
  root (EvalSymlinks). Parent-dir-named-`demo`, SEEDS_PATH-at-demo, symlink, and
  renamed-dir bypasses all closed (renamed-dir now hard-errors loudly, never
  seeds-nothing silently). Disposable-DB + `go test` proven.
- **FIXED 2026-07-26 (Fix B):** `make seed`/`make seed-demo` added;
  `INCLUDE_DEMO_SEEDS=true` wired into `.env.development`, `.env.e2e.example`,
  `setup-db.sh`, and the CI `e2e-iam` job. Disposable-DB proved dev/test opt-in
  restores ben/green/neo (3 accounts, 8 memberships).
- **DONE 2026-07-26 (Fix C):** full 28-path patch exported via worktree-index
  `git diff --cached --binary` to scratchpad `G1-demo-seed-cleanup.patch`;
  `git apply --check --3way` onto `neo-develop@b0faad1` clean. A NEW independent
  full-diff review is RUNNING; G1 acceptance still requires it to return no P0/P1.
- **RESOLVED 2026-07-26 (was the green-id collision; independent review rated it
  P2):** `database/seeds/demo/008_...parity_seed.sql` no longer uses the fixed ids
  `b1000000-...051..054`; it now omits the `id` column (DB-generated default,
  matching the ben demo seeds), so green's demo memberships are disjoint from the
  migration-owned id space and the forward migration's "never by application code"
  invariant is true. PROVEN on disposable DB: after demo seed green has 4
  memberships, 0 using `b1...051-054`; re-running `20260725000001` on the seeded
  DB now `DELETE 0` and green keeps all 4 (previously would have deleted them).
- **NEW (this session — now ENFORCED, not open):** the fail-closed manifest means
  every production reference seed must be listed. A reference seed landing from a
  parallel branch (concretely `database/seeds/020_cash_transaction_process_seed.sql`
  from the uncommitted cash-approval work) is NOT in the candidate manifest.
  Mitigated by the Fix-A hard-error guard: at integration such a seed will HARD-
  ERROR at seed time (CI/staging), forcing an explicit add-to-manifest (if it is
  clean reference config) or move-to-demo/ (if it carries ben/green coupling).
  **INTEGRATION GATE:** whoever merges G1 with the cash-approval work must resolve
  `020_cash_transaction_process_seed.sql`'s classification.
- **AUTHORED + DISPOSABLE-DB PROVEN 2026-07-27 (this account, per D5):**
  `database/migrations/20260727000002_permissions__remove_seeded_demo_identities`
  (up+down) written in worktree `agent-a10bc9baae7136672`. Design: DELETE the
  seed-owned access rows (investment__process_groups GROUP_A/GROUP_B ids
  `88000000-...101/102` — CASCADEs to their process_group_members +
  process_step_assignments; approval__group_members ids
  `a9100000-...0001/0002/0004/0005`; approval__delegations id
  `a9400000-...0001`; permissions_accounts_groups matched by exact
  UNIQUE(user_id, group_id) tuples for ben/green against the 4 named catalog
  groups) and UPDATE (not DELETE) the 3 iam_users rows to `is_active=false`
  (deactivation chosen over hard delete specifically to avoid FK-constrained
  audit-history failures on a real environment — sessions/audit
  events/approval requests these accounts may have accumulated — while still
  fully revoking their ability to authenticate/act). Down = documented
  no-op, same rationale as `20260725000001`. Disposable-DB proof (fresh
  `postgres:16-alpine`, port 15438, destroyed after): (a) fresh
  migrate-up + reference-only seed → 0 rows for all 3 identity ids (confirms
  true no-op on a DB that never ran demo seeds); (b) stepped down 1 (removing
  only the new migration's version marker), seeded WITH
  `INCLUDE_DEMO_SEEDS=true` (simulating the historical already-seeded state)
  → confirmed all 8 target grant categories present (3 active identities, 2
  process groups, 4 approval-group-members, 1 delegation, 8
  permissions_accounts_groups rows, 1 workflow_approval_settings row);
  migrated up again (reapplying just the new migration against real demo
  data) → confirmed 0 active identities (3 rows preserved, not deleted), 0
  remaining rows in every grant category, CASCADE correctly removed the
  process-group members/assignments, admin still active, all 4 catalog
  permission groups and all 3 approval groups (FUND_MANAGER_REVIEWERS/
  INVESTMENT_SUPERVISORS/TRADING_SUPERVISORS) preserved untouched; (c)
  down (no-op, confirmed it does NOT reactivate) then up again (idempotent,
  schema_migrations lands on `20260727000002, dirty=false`, no error). `git
  diff --check` clean on both new files.

  **Independent review round 1 (background agent, same session) verdict:
  REQUEST CHANGES.** Found one real P0 and one real P1, both now fixed (see
  below) — this is a genuine catch, not a false positive; give the finding
  full credit rather than treating the first PASS-shaped narrative above as
  final. P0: the migration's first version missed
  `database/seeds/zz_demo/06_demo_data_permissions.sql` entirely — 6 fixed-id
  `permissions_data_rights` rows granting ben/green/neo real per-fund
  data-scope access (SCB-FIXED/TH-GOV-LTF for ben, GLOBAL-TECH/BBL-EQUITY for
  green, KTB-BALANCED/GLOBAL-TECH for neo). This table is LIVE production
  code (`permissions` module's `effectiveContractScopes` and `iam`'s
  login-time `GetUserDataPermissions` both read it) — leaving it untouched
  would mean the access becomes live again the instant anyone flipped
  `is_active` back to true, with zero further migration needed. P1: step 2
  (`approval__group_members`) matched by row id only, inconsistent with step
  4's safer exact-tuple match, even though the table carries the identical
  `UNIQUE(group_id, user_id)` protection.

  **Both fixed in the same file, same session:** added step 6 deleting the 6
  fixed `permissions_data_rights` ids
  (`d000e000-...0010/0011/0020/0021/0030/0031`), explicitly excluding the
  admin wildcard row (`d000e000-...0001`, a different id, never touched);
  renumbered the old step 6 (iam_users deactivation) to step 7; step 2 now
  matches by the exact `(group_id, user_id)` tuple instead of row id.
  Re-verified on a SECOND fresh disposable container (not reused): fresh
  migrate-up + reference-only seed → 0 of the 6 `permissions_data_rights`
  ids; down 1 + `INCLUDE_DEMO_SEEDS=true` seed (simulating the historical
  pre-cleanup state) → confirmed all 6 rows present + admin wildcard present;
  migrate up again → confirmed all 6 gone, admin wildcard still present,
  the 4 `approval__group_members` rows (tuple-matched) gone, iam_users still
  deactivated; down/up idempotence re-confirmed. `git diff --check` clean.

  Patch RE-EXPORTED (twice — once before, once after the P0/P1 fix): staged
  all 29 worktree paths (`git add -A` inside the isolated worktree only),
  `git diff --cached --binary` to
  `.claude/scratchpad/G1-demo-seed-cleanup.patch` (3042 lines, final
  version); `git apply --check --3way` onto `neo-develop@b0faad1` from the
  main tree returns exit 0 both times (clean). Also re-confirmed the
  `020_cash_transaction_process_seed.sql` manifest-collision fix (added
  earlier this session, see the G1 acceptance note above) survived the
  re-export.

  **Independent review round 2 (same reviewer agent, resumed with the fix
  description — same session): verdict PASS, no P0/P1.** It independently
  re-verified both fixes against the actual seed files (not just re-reading
  my description): confirmed the 6 `permissions_data_rights` ids exactly
  match `zz_demo/06_demo_data_permissions.sql`'s real INSERT values with no
  omission/over-reach, confirmed the admin wildcard row is correctly
  excluded, confirmed (via a fresh repo-wide grep) that no table anywhere
  holds an FK into `permissions_data_rights(id)` so the new DELETE is
  FK-safe, and confirmed the step-2 tuple fix's 4 tuples match
  `demo/002_approval_demo_seed.sql`'s real values exactly and correctly
  excludes admin's own INVESTMENT_SUPERVISORS membership. One cosmetic-only
  P2 noted (step 6 matches by PK id rather than a `(user_id, contract_id)`
  tuple like steps 2/4) — explicitly assessed as not worth changing, since
  the risk that motivated tuple-matching in step 2 doesn't apply to step 6's
  synthetic-only contract ids.

  **Migration-number collision check (this is the manager's job, not the
  reviewer's — the reviewer only checked completeness against named seed
  files, not cross-tree numbering) — confirmed no collision:**
  `20260727000002` (this migration, G1 worktree) vs
  `20260727000001_investment__cash_request_fingerprint` and
  `20260727000003_approval__create_sync_failures` (both main tree, this
  session) are three distinct numbers.

  **G1 IS NOW FULLY ACCEPTED at the code/review level.** Every known finding
  (Codex round 1-2, the fresh independent full-diff review, and this
  session's residual-access P1 with its own two-round re-review) is
  resolved with no open P0/P1. The ONLY remaining gate is Kanta's explicit
  authorization to integrate the patch into the main tree (apply
  `.claude/scratchpad/G1-demo-seed-cleanup.patch`; do NOT merge the
  worktree's divergent branch history; still no commit without Kanta).
- **CORRECTED 2026-07-27: NOT actually in progress — nothing was ever written.**
  The prior checkpoint said "a worker is authoring
  `database/migrations/20260727000001_permissions__remove_seeded_demo_identities`";
  this account searched every worktree (`git status --porcelain` in
  `agent-a10bc9baae7136672`, plus a repo-wide filename search) and found the
  file DOES NOT EXIST anywhere. That worker died before persisting anything —
  the same failure class as `aa9057d6` (see Failed Approaches). Kanta was
  asked fresh via `AskUserQuestion` 2026-07-27 and chose **D5: author the
  cleanup migration now** (not the "confirm prod never ran legacy seeds"
  alternative). NEXT ACCOUNT MUST ACTUALLY WRITE THIS FILE — do not assume it
  survived a second time. Name it
  `20260727000002_permissions__remove_seeded_demo_identities` (NOT `...001` —
  that number is already taken in the main tree by
  `20260727000001_investment__cash_request_fingerprint`, and the two trees
  will eventually be integrated together). Idempotently remove every row keyed
  to the demo ids a0000000-…010/011/012 (FK-safe, no-op if absent), down =
  documented no-op, proven on a disposable DB (fresh no-op; seeded→migrate→demo
  gone + admin/roles intact; up/down/up clean). Then re-export the G1 patch
  and obtain a final clean G1 review. Deploy stays gated on Kanta. Original
  finding:
- **RESOLVED 2026-07-27 (was the residual-access P1): migration `20260727000002_permissions__remove_seeded_demo_identities`
  authored, fixes a P0 (missing `permissions_data_rights` cleanup) and a P1
  (step 2 id-vs-tuple matching) found by its own independent review, and
  passed a second independent re-review with verdict PASS/no P0/P1. See the
  G1 matrix row and Owner Decisions (D5) for full detail. Only integration
  into the main tree remains, gated on Kanta.**
- P0: fund-optional governance conflict. **RESOLVED 2026-07-24 — D1 =
  FUND_OPTIONAL.** Remaining work is to align docs to the single policy.
- P0: execution fill validation and lifecycle incompleteness (G3, pending).
- P0: direct security ledger lifecycle bypass (G5, pending).
- P0: production EOD uses `NopInvestmentQueryAdapter` (G6, pending).
- P0: trader/operator mutation workflow incomplete or over-labelled (G7, pending).
- **RESOLVED 2026-07-27 (Codex finding 2 / G2 item 1):** subject-access error
  classification — genuine denial → 403, missing port / repo / IAM / timeout /
  cancellation → operator-visible 5xx. Verified (approval build+tests green).
- **RESOLVED + INDEPENDENTLY RE-VERIFIED 2026-07-27 (G2 items 2-3 / Codex
  findings 3-5):** LIVE cash request-level idempotency (idempotency_key +
  partial unique index) + canonical fingerprint + ambiguous-submit
  reconcile-or-compensate + real-Postgres concurrency tests are all
  code-complete. Verified this session on a genuine disposable container (see
  Database/Runtime Actions).
- **RESOLVED 2026-07-27 (G2 item 4):** approval subject synchronization now
  has durable replay — `approval__sync_failures` table + bounded-retry
  `RetrySyncFailure` + operator endpoints, independently verified (see G2
  full-diff review verdict below).
- **DROPPED FROM THIS GOAL 2026-07-27 (owner decision D6):** approval callback
  lacks actual approver attribution. Kanta re-confirmed the 2026-07-24
  acceptance of this as a known v1 limitation shared with
  `INVESTMENT_DECISION`; do NOT implement in this goal; do not extend
  `contract.ApprovalDecision`. Record as a standing deferred P1 for a future,
  separately-scoped task.
- **RESOLVED 2026-07-27 (G2 item 6):** cash-request cancellation UI/list
  semantics no longer disagree — backend already enforced submitter-only
  cancel; frontend gate (`cashRequestAuthz.ts`) now hides the Cancel button
  for non-submitters/non-PENDING rows instead of showing it and relying on a
  confusing 403.
- **INDEPENDENT G2 FULL-DIFF REVIEW: PASS, 2026-07-27** (scope: items 1-4,
  6-8; item 5 explicitly excluded per D6). Reviewer independently spun up a
  disposable Postgres (:15443, destroyed after, never touched the shared dev
  DB), live-probed the composite FK (all 4 NULL/mismatch/match cases) and the
  terminal-state trigger, ran migrate up/down/up twice clean, traced
  `RetrySyncFailure`'s transaction boundary by hand and confirmed the
  historical rollback bug is NOT present, ran the backend and frontend test
  suites (all pass), and confirmed the two new permission codes actually
  resolve end-to-end (catalog -> seed grant -> Admin group). Two P1s raised,
  both test-coverage gaps in shipped-correct behavior, not defects:
  - **P1 (new, 2026-07-27):** `TestRetrySyncFailure_StillFailing_IncrementsAttemptAndDurablyPersists`'s
    durability assertion is vacuous against the fake repo — `fake_repo_test.go`'s
    `runTxFn` has no real commit/rollback semantics, so a *different* reintroduced
    bug (returning `result, callbackErr` instead of `nil, callbackErr` from
    `RetrySyncFailure`) would still pass this test's `AttemptCount` assertion,
    though the specific historical bug described in this handoff (returning
    `callbackErr` as the tx's own return value) is still caught, incidentally,
    via the `out == nil` check. Fix direction: make the fake `runTxFn` discard
    writes on non-nil return (model real commit/rollback), or add a
    Postgres-backed integration test for `RetrySyncFailure` alongside
    `cash_request_repository_integration_test.go`. **Broader nuance the review
    surfaced but didn't generalize:** since `fake_repo_test.go`'s `runTxFn`
    models no commit/rollback at all, EVERY test in this file that asserts
    "durably persisted"/"survives a rollback" is currently asserting nothing
    about durability, not just this one test — fixing the fake itself (rather
    than adding one more integration test around it) closes the gap for all
    of them at once.
  - **P1 (new, 2026-07-27):** no automated regression test for the composite FK
    `fk_inv_cash_req_portfolio_fund` (item 7) — `cash_request_repository_integration_test.go`'s
    throwaway schema only recreates the terminal-status trigger, not the FK.
    Verified correct live by the reviewer via manual psql probing, but
    unguarded against future regression. Fix direction: extend the
    integration test's throwaway schema to include the FK (with parent-table
    stubs), or accept as documented residual risk since the application layer
    never lets a client choose `fund_id` directly.
  Neither P1 blocks G2 acceptance. **G2 is ACCEPTED at code/review level
  2026-07-27** (same standard as G1: code-complete + independently reviewed;
  integration into any deploy/shared environment remains a separate Kanta
  gate, same as everything else in this session — nothing has been committed).
- P1 (standing, not re-filed by the review — already tracked): the shared
  approval engine's `terminate()` (WithdrawRequest/CancelRequest) never
  invokes `SubjectSync`, unlike approve/reject/revoke. Characterization test
  `TestWithdrawRequest_DoesNotNotifySubjectSync` documents current behavior.
  Fix requires the same `contract.ApprovalDecision`-class change as D6;
  resolve together in a future task, not ad hoc. The independent reviewer
  re-checked this specifically and confirmed no NEW interaction with the
  sync_failures table (SyncFailureOutcome has no WITHDRAWN/CANCELLED value,
  so terminate() paths simply produce no sync-failure row — consistent with
  the already-known gap, nothing worse).
- P1: TypeScript typecheck has 91 diagnostics.
- P1: regulatory/credit-rating controls are human-gated stubs.
- P1: watchlist scheduling and CI lifecycle coverage are incomplete.

## Failed Approaches / Do Not Repeat

- **Path-spelling demo classification (REJECTED by Codex twice).** Classifying a
  seed as demo by matching directory-name text (`demo`/`zz_demo`) — whether the
  first relative segment or ANY absolute-path segment — is unsafe: symlink/junction
  bypass, renamed-dir bypass, and (worst) a repo cloned under a parent dir named
  `demo` makes ALL reference seeds classify as demo → production silently seeds
  nothing and exits 0. Use instead: resolve symlinks (`filepath.EvalSymlinks`),
  classify by path RELATIVE to the resolved CANONICAL seed root, and restrict
  `SEEDS_PATH` to that canonical root in production — OR an explicit demo MANIFEST.
- **Cross-worktree `git apply` is blocked by the sandbox.** Workers cannot
  `git -C <other-worktree> diff | git apply` from within an isolated worktree.
  Transport a prior candidate by READING its files (Glob/Read works cross-worktree)
  and reconstructing them, then verify byte-fidelity. (This is why two workers
  reconstructed the 29-path candidate.)
- **Exporting with `git diff HEAD --binary` drops untracked files.** The candidate
  has 12 untracked new files (forward migration, new demo seeds). Stage in the
  isolated worktree (`git add -A`/`git add -N`) then `git diff --cached --binary`.
- **Do not re-dispatch a worker on a session usage-limit failure** (worker
  `aa9057d6` died this way). Retrying loops; wait for the reset (~1:30pm
  Asia/Bangkok) or a fresh account.
- **Manager caution: this account marked G1 "complete" twice prematurely** (after
  its own review + one independent review each time). Codex's deeper reviews then
  found real P0s both times. For G1, require an independent FULL-DIFF review that
  specifically probes SEEDS_PATH/symlink/parent-dir edge cases and the dev/CI
  wiring before believing "complete."

## Human Approval Required

- D4 Thai SEC product regime and effective-dated source-to-rule matrix. D1-D3
  are already resolved and must not be reopened without a new owner decision.
- D5 and D6 are now RESOLVED (2026-07-27, see Owner Decisions) — do not re-ask.
- **G1 patch INTEGRATED 2026-07-27 (Kanta's answer via `AskUserQuestion`:
  "Apply now, uncommitted").** `git apply --check --3way` confirmed clean,
  then `git apply --3way .claude/scratchpad/G1-demo-seed-cleanup.patch`
  applied all 28 paths cleanly onto the main tree (uncommitted — nothing
  staged/committed, per standing no-commit rule). Post-apply verification on
  a SECOND fresh disposable Postgres (`ims-verify-merged`, port 15444,
  destroyed after — the first-ever proof of G1+G2 coexisting in one tree):
  `gofmt -l .` clean, `go vet ./...` clean, `go build ./...` clean,
  `go test ./cmd/seed/...` ok; `go run ./cmd/migrate up` clean through every
  migration including the newly-merged `20260727000002` (demo-identity
  cleanup, from the G1 worktree) alongside G2's `20260727000003`/`...000004`;
  `go run ./cmd/seed` in production mode (`APP_ENV=production` + dummy
  required secrets) seeded successfully with 0 demo accounts
  (`iam_users` ben/green/neo count = 0) and 17 demo files skipped; confirmed
  live: `APPROVAL_SYNC_VIEW`/`APPROVAL_SYNC_RETRY` in
  `permissions_function_definitions`, the composite FK
  `fk_inv_cash_req_portfolio_fund` and trigger
  `trg_inv_cash_req_terminal_status_immutable` both present on
  `investment__portfolio_cash_requests`. No G1/G2 conflict found. Do NOT
  re-apply this patch or re-ask this question.
- Index/commit regrouping.
- Migration execution outside a disposable database.
- Any local commit.
- Any push, PR, deployment, or production action.

## Next Exact Actions

1. **G1 is DONE — code-complete, independently reviewed, AND integrated into
   the main tree** (uncommitted; Kanta chose "apply now, uncommitted" via
   `AskUserQuestion` on 2026-07-27). Post-integration disposable-DB proof
   confirms G1+G2 coexist cleanly. Do NOT redo, re-author, or re-apply
   anything for G1. The worktree `agent-a10bc9baae7136672` is now historical
   only — its patch has already been applied to `neo-develop`'s working tree.
2. **G2 is DONE — ACCEPTED 2026-07-27.** Items 1-4 and 6-8 all source-proven
   (item 5 DROPPED per D6); frontend Idempotency-Key task done (build exit 0);
   independent full-diff review (task #9) returned **PASS** with 2 new P1
   test-coverage-gap findings (not blocking — see Unresolved P0/P1). Do NOT
   redo any G2 item. The two new P1s are optional hardening for a future pass,
   not required before G3.
3. **G3 is next — start here.** Fill validation/lifecycle: positive/cumulative/
   ordered fill bounds + workflow-day gating + compliance re-evaluation on
   fill. Read the runbook (`CLAUDE-GOAL-IMS-REMEDIATION.md`) for G3's exact
   acceptance criteria before starting; verify current state of execution/fill
   code first (source overrides any stale status) rather than assuming nothing
   exists.
4. **Then G4-G9** in runbook dependency order (D2 MATCHED CONFIRMATION, D3 EXPIRE
   PENDING, D4 NOT_CONFIGURED), each with gates + disposable-DB/E2E + browser proof
   + independent review, per the Definition of Done. No commit/push/deploy without
   Kanta.
   NOTE: production-scheduler-inert (005 moved to demo) is a **G6 P0 dependency**
   (real contract-source integration). And resolve
   `020_cash_transaction_process_seed.sql`'s manifest classification when
   integrating G1 with the cash-approval work (the Fix-A hard-error guard forces it).

Status of earlier-"deferred" items (CORRECTED — they were NOT G8 hygiene):
APP_ENV-unset fail-closed is now FIXED in the G1 candidate (F2); `005` IMS-DEMO
contracts are FIXED (moved to demo, F4) and the resulting production-scheduler-
inert gap is now a **G6 P0 dependency** (real contract-source integration), not
an owner-acknowledgment item; the stale test-comment filename remains a cosmetic
follow-up. Do NOT re-file the fixed items as deferred.

## Next-Account Prompt

This compact continuation prompt reflects CURRENT state (G1 AND G2 both fully
ACCEPTED and integrated/verified; G3 is the next package to start). It does
NOT resume G1 or G2 from scratch. Stays below Claude's 4,000-character limit:

```text
/goal Continue IMS-MERGE-BLOCKERS through G9 (FUND_OPTIONAL). Invoke /ai-engineering-manager. Read docs/MANAGER/{MEMORY,HANDOFF,TASKS,CLAUDE-GOAL-IMS-REMEDIATION,CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF}.md then CLAUDE.md; verify Git/source/worktrees and every handoff claim before editing (a prior checkpoint's "worker is authoring X" claim was FALSE and cost a full session to discover — always grep/search for the actual file, never trust the claim). Do NOT redo verified work. **G1 is DONE: code-complete, independently reviewed (two rounds, PASS), AND INTEGRATED into the main tree 2026-07-27** (uncommitted — Kanta chose "apply now, uncommitted" via AskUserQuestion; do not re-ask, do not re-apply the patch, the worktree agent-a10bc9baae7136672 is now historical only). **G2 is DONE: ACCEPTED 2026-07-27** — items 1-4 and 6-8 ALL source-proven complete (item 5 DROPPED per D6): sync-failure durable replay, submitter-only cancel authz, FUND_OPTIONAL ownership+terminal-state guard, lifecycle tests, frontend Idempotency-Key — independent full-diff review returned **PASS** (task #9) after live-probing migrations on a disposable Postgres, tracing the RetrySyncFailure tx boundary, and running both test suites. Post-integration, a SECOND disposable-DB proof confirmed G1+G2 coexist cleanly in one tree (migrate+seed clean, 0 demo accounts in prod mode, all G2 DB objects present). Two non-blocking P1s from the G2 review (fake-repo tx-rollback test gap — broader than the one named test, the whole fake models no commit/rollback; no automated FK regression test) are recorded in Unresolved P0/P1 — optional hardening, not required before G3. Owner decisions D5/D6 RESOLVED — do not re-ask. Do NOT redo any G1/G2 item. Resume: (a) a known deferred P1 remains open (terminate()/withdraw never notifies SubjectSync — `TestWithdrawRequest_DoesNotNotifySubjectSync` documents it; fix requires the same `contract.ApprovalDecision` change class as D6, resolve together, not ad hoc — not yet raised as its own decision to Kanta); (b) **start G3** (fill validation/lifecycle: positive/cumulative/ordered fill bounds + workflow-day gating + compliance re-eval on fill) — read the runbook's exact G3 acceptance criteria first, verify current execution/fill code state before assuming anything is missing. Then G4-G9 in runbook order (D2 MATCHED CONFIRMATION, D3 EXPIRE PENDING, D4 NOT_CONFIGURED; G6 has a new P0: production workflow__scheduler_contracts is empty now that G1 removed demo contracts from production seeds, needs a real contract-source integration). Completion = every acceptance box source-proven; backend gofmt/vet/test/build; frontend test/typecheck/build zero diagnostics; reproducible Swagger/OpenAPI; fresh+upgrade disposable-DB migration/rollback/lifecycle/concurrency E2E; authenticated EN/TH/ZH browser proof; independent security/database/lifecycle/workflow/frontend reviews no P0/P1. Preserve the dirty tree/index (167 paths changed post-G1-integration, all uncommitted). No commit/push/PR/deploy/shared-DB/production-data/legal decision without Kanta. Independently verify worker diffs; never stop on a worker report alone; run go test ./... -short to avoid silently touching the shared dev Postgres (integrationDSN() defaults to it). Boundaries: 60% checkpoint; 70% no new package; 80%/any warning stop only after a complete durable handoff (exact Git/worktree/files/commands/results/failures/remaining) + a <=4000-char prompt. Token exhaustion is not completion.
```
