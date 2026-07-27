---
type: claude-goal-runbook
project: IMS Thailand
goal_id: IMS-MERGE-BLOCKERS
owner: Kanta
status: ready-for-owner-decisions
created: 2026-07-24
---

# Claude `/goal` Runbook: IMS End-to-End Remediation

This is the execution contract for finishing the source-proven IMS product,
financial-control, security, workflow, audit, frontend, documentation, and
release-readiness findings recorded in:

1. `docs/MANAGER/MEMORY.md`
2. `docs/MANAGER/HANDOFF.md`
3. `docs/MANAGER/TASKS.md`
4. This runbook

When those sources disagree, apply the source-of-truth order in `MEMORY.md`.
Executable source and tests outrank stale handoff prose, but source must not be
used to invent a business decision that belongs to Kanta.

## 1. How to Use This Runbook

Claude Code `/goal` keeps working across turns until an evaluator decides the
completion condition is satisfied. The evaluator does not inspect files or run
commands itself. Claude must therefore surface a concise acceptance matrix,
exact commands, exit results, Git state, and residual risks in the conversation
after every work package.

Before starting:

1. Open Claude Code at the repository root.
2. Trust only this repository. Do not use production credentials or customer
   data.
3. Optionally configure a visible context/token status line:

   ```text
   /statusline show model name, context used percentage, session token usage, and git branch
   ```

4. Run `/context` and `/usage` once to establish the initial context and account
   budget.
5. Invoke `/ai-engineering-manager`.
6. Paste the `/goal` command from Section 12.

Use `/goal` with no argument to inspect goal turns and token spend. Use
`/context` to inspect context-window usage and `/usage` to inspect account or
plan usage. `/goal clear` is allowed only when Kanta cancels the program, a
human decision blocks all safe progress, or the token-limit handoff in Section
10 is complete.

## 2. Owner Decisions and Remaining Human Gate

The following decisions were recorded from Kanta on 2026-07-24 and are
authoritative for this goal. Claude must verify that they remain recorded near
the top of `docs/MANAGER/CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md`; do not ask Kanta
to choose D1-D3 again unless Kanta explicitly changes direction.

### D1. Fund policy

**Resolved: `FUND_OPTIONAL`.** Fund-less portfolios and financial activity
replace the 2026-07-15 fund-required decision. Workflow, compliance,
permissions, reporting, migrations, tests, and UI must use the portfolio as the
scope when no fund exists; absence of a fund never skips a mandatory control.

### D2. Official security-trade financial-effect event

**Resolved: `MATCHED CONFIRMATION`.** The matched/resolved confirmation is the
one authoritative event that creates official ledger, holding, and cash effects
for LIVE security trades. Execution fill is non-official. The implementation
must define partial fills, multi-fill orders, confirmation mismatches,
settlement failure, cancellation, correction, reversal, and duplicate delivery
against that event.

### D3. Pending approval at business-day close

**Resolved: `EXPIRE PENDING`.** A LIVE cash request or security-trade request
still pending at business-day close expires with a distinct audited reason and
must be resubmitted on the next open day. Expiry must create no financial
effect; cancellation/override permissions remain enforced by the owning
workflow.

### D4. Thai SEC product regime

**Human-gated/deferred.** Regulatory thresholds require Kanta or a compliance
expert to approve a product regime and exact source-to-rule matrix. Until approved,
`regulatory.thai_sec` must remain honestly `NOT_CONFIGURED`; Claude must not
invent a threshold or convert the stub into a false PASS. Technical work may
make the gate clearer and fail-safe, but legal meaning remains human-gated.

## 3. Current Baseline to Verify, Not Trust

At the time this runbook was created:

- Branch: `neo-develop`
- HEAD: `b0faad1`
- Tracking: two commits ahead of `origin/neo-develop`
- Index: 100 staged paths containing fund-optional Stage 1 plus LIVE cash
  approval Stage 2
- Unstaged manager files: `docs/MANAGER/HANDOFF.md` and
  `docs/MANAGER/TASKS.md`
- Untracked regulatory memo:
  `docs/compliance/thai-sec-product-regime-decision-memo.md`
- Latest verified checks:
  - targeted approval/investment/workflow Go tests passed;
  - 60 frontend test files / 655 tests passed;
  - Nuxt typecheck failed with 91 diagnostics;
  - backend and frontend production builds passed in the preceding review;
  - staged and unstaged `git diff --check` passed.

Claude must immediately rerun:

```powershell
git status --short --branch
git branch --show-current
git log -6 --oneline --decorate
git rev-list --left-right --count origin/neo-develop...HEAD
git diff --cached --name-only
git diff --name-only
```

If the state differs, record the new state in the next-account handoff and
adapt the work packages without resetting or discarding user work.

## 4. Global Safety and Ownership Rules

- Preserve all existing changes. Never run `git reset --hard`, destructive
  checkout, broad clean, or recursive deletion.
- Do not commit, push, open a PR, deploy, run production migrations, or mutate
  production data without Kanta's explicit authorization for that action.
- Do not bulk-stage a dirty tree. Use explicit pathspecs and review staged
  scope.
- Do not stack parallel writers in the same worktree. Use one writer per
  worktree with exclusive file ownership.
- Use `/frontier-backend-engineer` for Go, API, database, permissions, audit,
  workflow, and migrations.
- Use `/frontier-frontend-engineer` for Nuxt, generated-client consumption,
  frontend state, i18n, accessibility, and browser work.
- Backend contracts land and are accepted before generated clients and
  frontend consumers.
- Generated Swagger and `frontend/app/api/ims-api.d.ts` must be regenerated
  from authoritative source. Never hand-edit generated contracts.
- Every database change needs an up/down strategy, fresh disposable-database
  proof, concurrency analysis, and independent database review.
- Financial mutations must fail closed and leave no partial effect.
- A test is not allowed to pass by weakening assertions, validation,
  permissions, auditability, or error handling.
- Keep real secrets, customer data, and production financial data out of
  prompts, tests, screenshots, and handoffs.

## 5. Ordered Work Packages

Only one package may be `ACTIVE` in the manager documents. Dependency-free
read-only review may run in parallel; write packages follow this order.

### G0. Preserve and reconcile the dirty baseline

Acceptance:

- [ ] Verify every staged and unstaged path and its provenance.
- [ ] Record D1-D4 decisions or mark the exact human blocker.
- [ ] Record the current index exactly. Preserve its staged/unstaged split
      unless Kanta explicitly authorizes regrouping; documentation work alone
      never changes the index.
- [ ] Update stale `HANDOFF.md` front matter and `TASKS.md` status so they no
      longer claim implemented/browser-verified work is merely queued.
- [ ] Create a current acceptance matrix in the next-account handoff.

### G1. Remove production demo/bootstrap access

Scope: existing P0-B.

Acceptance:

- [ ] Locate migrations/bootstrap paths granting demo users such as `ben` or
      `green` production permissions or data scope.
- [ ] Move demo-only grants into an explicitly selected development/test seed
      path that is technically rejected in production. A directory name,
      comment, README, or operator convention is not an environment gate.
- [ ] Keep production reference/catalog seeding separate from synthetic demo
      users, memberships, grants, scopes, and datasets.
- [ ] Production migration/bootstrap contains no named demo-user privilege.
- [ ] Add a new forward corrective migration for already-upgraded databases.
      It removes only exact migration-owned demo memberships and preserves
      production role catalogs, function rights, legitimate administrator
      assignments, and data scopes.
- [ ] Do not treat editing an already-applied migration as upgrade remediation;
      historical-file cleanup protects fresh databases only.
- [ ] From a clean worktree based on current `neo-develop`, prove separately:
      fresh migration-only state; historical upgrade cleanup; explicit
      development/test demo restoration; production-configured demo-seed
      rejection; and migration up/down/up behavior.
- [ ] Secure bootstrap, migration, seed-loader, and permission tests prove the
      production and development behaviors. Obtain an independent raw-diff
      database/security review with no unresolved P0/P1.

Current correction note (2026-07-24): the uncommitted candidate in
`.claude/worktrees/agent-aa1dd1dbf78db7268` removes the historical named-user
blocks but does not satisfy the production seed gate or existing-database
cleanup requirements. Its branch is based on history ahead of `neo-develop`;
do not merge it wholesale. Apply only the reviewed file diff or recreate it
from a clean current `neo-develop` worktree.

### G2. Correct approval and LIVE cash controls

This package closes the findings in the staged Stage 2 implementation.

Acceptance:

- [ ] Replace blanket `asSubjectAccessDenied` coercion with a cross-module
      contract that distinguishes a genuine access denial from repository,
      IAM, timeout, cancellation, and unexpected failures.
- [ ] Prove denial returns 403 while IAM/database/infrastructure failure
      remains fail-closed and returns 5xx with operator-visible evidence.
- [ ] Add request-level idempotency to LIVE cash submission. A response/update
      timeout followed by retry cannot create a second cash request, approval
      request, ledger row, or financial effect.
- [ ] Reconcile or atomically coordinate cash-request creation, approval
      submission, and approval-request correlation. No orphan approval or
      permanently ambiguous pending row is silent.
- [ ] Implement durable approval-subject synchronization retry/replay with
      persisted status, bounded retry, idempotency, and operator visibility.
      Merely writing `retryable: true` to audit is insufficient.
- [ ] Extend the approval decision contract so approve/reject callbacks receive
      the actual deciding actor. Persist and audit the approver, not the
      submitter, while preserving submitter attribution for the eventual ledger
      request.
- [ ] Make list/cancel semantics consistent: either list only the submitter's
      requests or expose authorization so Cancel renders only for the
      submitter. Backend authorization remains authoritative.
- [ ] Enforce portfolio/fund ownership consistency for cash requests using the
      D1 policy; add state constraints for terminal/materialized fields and
      approval correlation where cross-schema ownership permits it.
- [ ] Define and test rejection, withdrawal, expiry, cancellation, duplicate
      callback, stale business date, compliance failure, permission loss, and
      concurrent approve-versus-cancel behavior.
- [ ] MODEL is backend-blocked for every official ledger/execution path;
      SIMULATION outputs are explicitly non-official; LIVE cash remains
      approval-gated.

### G3. Validate execution fills and lifecycle transitions

Scope: existing P0-C plus the fill portion of P0-E.

Acceptance:

- [ ] Require positive executed quantity, amount, and price as applicable.
- [ ] Prevent cumulative overfill beyond ordered/approved quantity and amount.
- [ ] Define and enforce partial-fill and multi-fill state transitions.
- [ ] Validate execution status transitions and terminal-state behavior.
- [ ] Re-run compliance against actual fill values where policy requires it.
- [ ] Enforce workflow/business-day state for create, fill, and cancel with
      race-safe locking.
- [ ] Reject cross-portfolio, cross-fund, guessed, cancelled, stale, or
      unauthorized decision/execution sources.
- [ ] Add domain, command, repository, concurrency, handler, and real-Postgres
      tests proving invalid fills create no partial state.

### G4. Make compliance binding changes attributable

Scope: existing P0-D.

Acceptance:

- [ ] Binding create/change/deactivate receives the authenticated actor.
- [ ] Persist immutable before/after audit evidence with business date and
      reason.
- [ ] A successful mutation cannot silently omit required audit.
- [ ] Failure semantics distinguish committed mutation, audit failure, and
      retry/recovery status without returning misleading generic errors.
- [ ] Add permission, lifecycle, audit, and concurrency tests.

### G5. Implement one authoritative trade-to-ledger lifecycle

Scope: existing P0-E and owner decision D2.

Acceptance:

- [ ] LIVE security BUY/SELL/SUBSCRIPTION/REDEMPTION cannot post directly
      without valid correlated lifecycle provenance.
- [ ] The D2 event creates ledger, holdings, and cash atomically, or through a
      durable outbox/idempotent consumer with observable retry.
- [ ] Decision, approval, compliance, execution, confirmation, settlement,
      ledger, and reversal states remain correlated and internally consistent.
- [ ] Database foreign keys/business keys enforce valid source relationships.
- [ ] External broker identity and delivery retries are idempotent.
- [ ] Partial fill, multi-fill, mismatch, cancellation, correction, settlement
      failure, reversal, and duplicate behavior are explicitly modeled.
- [ ] Confirmation resolution is connected to the chosen settlement/financial
      policy instead of only updating a confirmation row.
- [ ] Direct ledger endpoints reject attempts that bypass the lifecycle.
- [ ] Audit records the real actor and correlation IDs for every transition.
- [ ] A live-Postgres E2E proves decision -> approval -> execution/fill ->
      confirmation/settlement -> ledger -> holdings/cash.

### G6. Replace fake EOD activity and close mutation races

Scope: existing P0-F.

Acceptance:

- [ ] Replace production `NopInvestmentQueryAdapter` with an investment-owned
      summary of real transactions, executions, confirmations, settlement, and
      reconciliation for the global business date.
- [ ] Manager approval cannot attest zero activity when real activity exists.
- [ ] Required pending review, confirmation, settlement, or reconciliation
      blocks close.
- [ ] Missing workflow day/dependency fails closed.
- [ ] Create/fill/cancel, confirmation, ledger, cash approval materialization,
      and reversal all honor the same business-day policy with race-safe locks.
- [ ] Define D3 expiry/carry-forward behavior.
- [ ] Replace or explicitly production-gate the weekend-only holiday adapter;
      Thai calendar policy requires owner/operations approval.
- [ ] Concurrency tests prove close cannot race a financial mutation.
- [ ] E2E proves open, approval, transaction close, accounting close, and every
      supported reverse/cancel path.

### G7. Complete and honestly label operator/frontend workflows

Scope: existing P0-I and all accepted backend contracts.

Acceptance:

- [ ] Authorized traders can create, partially fill, fill, and cancel
      executions through generated typed APIs.
- [ ] Operations users can record/import, match, and resolve confirmations.
- [ ] Decision detail shows one correlated lifecycle and distinguishes official
      from non-official financial effects.
- [ ] OP-02/OP-03 are `ready` only when required mutations and backend
      contracts are genuinely complete; otherwise they are `limited` with an
      accurate reason.
- [ ] Cash Cancel renders only when the current actor may cancel.
- [ ] Approval sync failure/retry, compliance unavailable, closed day, stale
      data, duplicate request, partial fill, mismatch, and empty states have
      actionable guidance.
- [ ] EN/TH/ZH keys remain structurally identical and user-facing labels never
      expose raw UUID/fund/contract identities.
- [ ] Desktop/tablet/mobile, light/dark, keyboard, accessibility, console, and
      network behavior receive browser evidence.

### G8. Close remaining production-readiness findings

Acceptance:

- [ ] Fix the Docker SSR session-restore configuration so hard reload/bookmark
      uses a server-reachable backend URL and retains a valid session.
- [ ] Add scheduled watchlist evaluation with idempotency, stale-price/provider
      behavior, and notification proof.
- [ ] Add periodic/post-trade compliance evaluation where required.
- [ ] Reduce Nuxt typecheck from the current 91 diagnostics to zero without
      weakening TypeScript settings or excluding files.
- [ ] CI runs frontend Vitest, typecheck, build, backend tests, migration
      verification, and Portfolio V2 financial-lifecycle E2E.
- [ ] Regulatory and credit-rating stubs remain fail-safe and honestly labelled
      until D4 is approved; after approval, implement only the signed
      source-to-rule matrix with effective dates and boundary tests.
- [ ] Add an auditor-facing correlated portfolio timeline or an accepted
      equivalent query proving the full lifecycle.
- [ ] Resolve stale manager/design comments and generated API documentation.

### G9. Integrated verification and independent review

Completion requires all applicable commands to pass against the final
integrated source:

```powershell
Set-Location backend
gofmt -l .
go vet ./...
go test ./... -count=1
go build ./...

Set-Location ..\frontend
npm run test
npx nuxi typecheck
npm run build

Set-Location ..
git diff --check
git diff --cached --check
```

Also required:

- [ ] Regenerate Swagger and frontend API types twice; the second run produces
      no diff.
- [ ] Apply all migrations to a fresh disposable PostgreSQL database.
- [ ] Exercise safe rollback/forward behavior and document intentionally
      irreversible migrations.
- [ ] Run the real-Postgres lifecycle and concurrency E2E suites.
- [ ] Run authenticated browser proof for the full happy path and selected
      failure paths in EN/TH/ZH.
- [ ] Obtain independent read-only reviews of security/permissions,
      financial lifecycle/idempotency, migrations/database constraints,
      workflow/EOD races, frontend contract/UX, and test adequacy.
- [ ] Resolve every reproduced P0/P1 finding. A documented limitation is not a
      closure unless Kanta explicitly accepts it as a release exception.
- [ ] Update `MEMORY.md`, `HANDOFF.md`, and `TASKS.md` to match verified
      reality.

## 6. Required Per-Turn Evidence

At the end of every `/goal` turn, Claude must print:

```text
GOAL STATUS
- Active package:
- Acceptance completed this turn:
- Files changed:
- Commands and exact pass/fail:
- Git branch/HEAD/index/worktree:
- New or residual P0/P1 findings:
- Human decision needed:
- Context/token status:
- Next exact action:
- Handoff updated: yes/no
```

The `/goal` evaluator can only judge what appears in the transcript. Do not
hide proof only in a file or say "tests pass" without command and result.

## 7. Failure and Retry Rules

- After two materially identical failed attempts, stop retrying. Record the
  reproduction, commands, output summary, suspected missing dependency, and
  safest next experiment in the handoff.
- Never bypass a failing control to make tests green.
- Never use a migration `force` operation without Kanta's explicit approval
  after independently verifying the physical schema and recording the exact
  target database.
- Tests and browser proof use sanitized/synthetic data in a disposable or
  explicitly approved local environment.
- If a mutation may have committed before an error, inspect authoritative state
  before retrying.

## 8. Commit and Integration Policy

Documentation creation does not authorize commits.

When Kanta authorizes local commits:

- Split work by coherent package and dependency order.
- Inspect explicit staged paths and `git diff --cached`.
- Never combine manager-history churn, local runtime values, unrelated
  regulatory drafts, generated output of uncertain provenance, or separate
  feature packages accidentally.
- Record commit hashes and tests in the handoff.
- No push or deployment without separate explicit authorization.

## 9. Token and Context Awareness

Token/context continuity is part of correctness, not an optional courtesy.

At session start and each package boundary:

1. Check `/goal` for goal turns and token spend.
2. Check `/context` for context-window usage.
3. Check `/usage` for account/plan limits.
4. Update the acceptance matrix and next action in
   `docs/MANAGER/CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md`.

Operational thresholds:

- Below 60% context: normal scoped work.
- At or above 60%: finish the current bounded change and update the handoff
  before opening another large source area.
- At or above 70%: do not begin another work package. Run focused validation,
  record the exact diff and next action, then use `/compact` only if the same
  account/session can safely continue.
- At or above 80%, on a context warning, or on an account/session usage warning:
  stop implementation after leaving the tree buildable where practical; write
  the complete handoff in Section 10.

If Claude cannot read an exact percentage, it must treat any compaction,
context-left, rate-limit, or usage warning as the 80% trigger. It may ask Kanta
to run `/context`, `/usage`, or `/goal`, but must not wait to start recording
the handoff.

Use focused subagents for noisy investigation and independent review so raw
logs do not fill the manager context. Subagents return concise findings and
must not write the same files concurrently.

## 10. Mandatory Next-Account Handoff

Before token/context/account exhaustion, Claude must update:

`docs/MANAGER/CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md`

The handoff is complete only when it contains:

- exact branch, HEAD, upstream divergence, staged paths, unstaged paths, and
  untracked paths;
- D1-D4 decisions and who made them;
- active/completed/blocked work packages and acceptance boxes;
- every file changed during the current account/session;
- exact commands and pass/fail results;
- migrations, containers, databases, and test data touched;
- unresolved P0/P1 findings and failed approaches;
- the next three exact commands or edits;
- any agent/worktree identifiers and file ownership;
- whether a commit/push/deploy was authorized or performed;
- a copy-paste prompt for the next Claude account.

Do not rely on Claude transcript memory. The next account must be able to
continue from Git and repository documents alone.

## 11. Definition of Done

The goal is complete only when:

- D1-D4 are resolved or explicitly recorded as human-gated non-implementation;
- all authorized P0/P1 work packages above are evidenced complete;
- there is one authoritative security-trade financial lifecycle;
- LIVE/SIMULATION/MODEL behavior is enforced by the backend;
- approval, EOD, audit, permission, retry, and idempotency controls are
  operational and observable;
- backend and frontend quality gates pass, including zero typecheck diagnostics;
- disposable-database E2E and authenticated browser proof pass;
- generated contracts are reproducible;
- independent reviews have no unresolved P0/P1;
- manager documents match source and tests;
- Kanta has reviewed required business, security, migration, and release gates.

Token exhaustion is not product completion. It is only a successful
account/session terminal condition when the mandatory handoff is complete and
the next account can resume without guessing.

## 12. Copy-Paste `/goal` Command

The owner has resolved D1 as `FUND_OPTIONAL`. The detailed contract stays in
this runbook and the next-account handoff so the Goal condition remains below
Claude's 4,000-character limit. Paste this command as written:

```text
/goal Continue IMS-MERGE-BLOCKERS with FUND_POLICY=FUND_OPTIONAL. Invoke /ai-engineering-manager. Read docs/MANAGER/MEMORY.md, docs/MANAGER/HANDOFF.md, docs/MANAGER/TASKS.md, docs/MANAGER/CLAUDE-GOAL-IMS-REMEDIATION.md, docs/MANAGER/CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md, then CLAUDE.md; verify Git, source, worktrees, and every handoff claim before editing. Treat verified source plus the latest next-account handoff as authoritative for current package state; do not redo superseded work. G0 is complete. G1 is ACTIVE CORRECTION and blocks G2: resume the first incomplete G1 item under "Next Exact Actions" using the authoritative candidate recorded there; complete Fix A/B/C, required disposable-DB and code gates, complete-patch export, and a new independent full-diff review with no P0/P1. Do not integrate G1 without Kanta's approval or start G2 before G1 passes. Then execute G2-G9 in runbook dependency order using D1 FUND_OPTIONAL, D2 MATCHED CONFIRMATION, D3 EXPIRE PENDING, D4 HUMAN-GATED/NOT_CONFIGURED. Preserve unrelated changes and the main index. No commit, push, PR, deploy, shared/production data, non-disposable migration, or legal/business decision without Kanta's explicit approval. At every turn print the required GOAL STATUS evidence and update the handoff. Check /goal, /context, /usage at startup and package boundaries: at 60% finish and checkpoint; at 70% start no new package; at 80% or any compaction/context/account/session warning, stop implementation and write the complete durable handoff. Do not retry usage-limit failures. Token exhaustion is not completion.
```

## 13. Prompt for the Next Claude Account

Use this only after the previous account has written the mandatory handoff:

```text
Continue IMS-MERGE-BLOCKERS; do not restart the investigation.

Read, in order:
1. docs/MANAGER/MEMORY.md
2. docs/MANAGER/HANDOFF.md
3. docs/MANAGER/TASKS.md
4. docs/MANAGER/CLAUDE-GOAL-IMS-REMEDIATION.md
5. docs/MANAGER/CLAUDE-GOAL-NEXT-ACCOUNT-HANDOFF.md
6. CLAUDE.md

Verify the handoff against Git and source before editing. Resume the first
incomplete acceptance item in the active work package. Preserve unrelated
changes and existing index state. Do not repeat completed work. Invoke the
appropriate project skill for backend or frontend work. Surface exact evidence
at the end of every turn, keep the handoff current, and apply the token/context
checkpoint rules. Do not commit, push, deploy, run production migrations, or
use production data without Kanta's explicit permission.

After verification, start the compact `/goal` condition from Section 12. D1 is
already `FUND_OPTIONAL`; do not ask Kanta to choose it again. Resume the exact
first incomplete G1 item from the latest next-account handoff, not G2.
```
