# Stage 2 Handoff Prompt (paste into a fresh Claude Code session)

> Copy everything inside the code fence below into a new Claude Code session
> (any account) opened at the repo root
> `C:\Users\kanta\source\repos\ims-th-solution`. It is self-contained.

```text
Continue IMS-PORTFOLIO-FUND-OPTIONAL Stage 2 (LIVE cash-transaction approval gate).

READ FIRST, in this order:
1. docs/investment/live-cash-transaction-approval-design.md  ← the authoritative
   spec: owner-confirmed policy, verified codebase facts, exact file/migration
   pointers, backend design (§4), frontend design (§5), test matrix (§6), and
   hard constraints (§7). Do NOT re-derive or re-negotiate the policy.
2. docs/MANAGER/MEMORY.md, docs/MANAGER/HANDOFF.md, docs/MANAGER/TASKS.md
3. CLAUDE.md

OWNER-CONFIRMED POLICY (do not change): LIVE portfolios require a full approval
workflow before cash movements (CASH_IN, CASH_OUT, FEE, DIVIDEND) post.
SIMULATION posts immediately (unchanged). MODEL stays blocked (unchanged).
BUY/SELL and other types are OUT of scope. Submitter can SEE their pending
request and can CANCEL it before an approver acts.

HARD RULES (from owner memory — non-negotiable):
- NEVER run git commit / push / amend / rebase / reset. The owner (kanta) is the
  sole committer. Leave all work uncommitted for their review.
- Do NOT weaken the append-only triggers on investment__portfolio_transactions.
  The real transaction row is materialized ONLY after approval, inside a new
  ApprovalSubjectCallback.OnApprovalDecision (a separate pending-request table
  holds the request until then).
- Backend-authoritative permissions/data-scope; enforce via the handler's
  PermissionChecker (h.pc), never literal nil (that was the P0-A vuln).
- Every .up.sql needs a matching .down.sql. Regenerate Swagger + OpenAPI client
  (make swagger, make api-client) if the API changes; never hand-edit generated
  files or hand-write frontend API types.
- Attributable, immutable audit trail for submit/approve/reject/cancel/materialize.
- If you rebuild Docker images, run docker-compose build in the FOREGROUND
  (backgrounded builds were silently killed twice before).
- Preserve ALL uncommitted working-tree changes (Stage 1 fund-optional work,
  ~30 backend files + frontend + migrations 20260723000001/20260723000003, plus
  docs). Do not reset or discard anything.

CURRENT STATE (verify with `git status` and against the running containers first —
the local DB volume self-wiped twice in a prior session, so confirm, don't trust):
- Branch neo-develop. Stage 1 (fund-optional portfolios + fund-less ledger/trading)
  is verified and UNCOMMITTED — the owner commits it themselves.
- Migrations 20260723000001 (portfolio fund_id nullable) and 20260723000003
  (trading tables fund_id nullable) are applied; schema_migrations latest =
  20260723000003, not dirty. All five investment tables have nullable fund_id.
- Local stack: docker containers ims-postgres (psql -U ims_app -p 5437 -d ims_dev),
  ims-backend (:8080), ims-frontend (:3000), ims-redis, ims-mailpit. Users:
  admin/admin2/ben/green/neo. Fund-bound portfolios exist (e.g. B14-CORE LIVE);
  a fund-less SIMULATION portfolio "TEST" exists.
- The BACKEND for Stage 2 was dispatched to a subagent in the prior session.
  CHECK whether it completed: look for new files
  investment__portfolio_cash_requests migration + seed, a
  cash_request_approval_adapter.go, enum additions (PORTFOLIO_CASH_TRANSACTION /
  CASH_TRANSACTION), and gate logic in
  backend/internal/investment/application/command/post_transaction.go. Run
  `go build ./...` and `go test ./...` from backend/ to see if it's green.

YOUR JOB (pick up wherever the backend left off):
A. If backend is incomplete or untested: finish it per spec §4 and §6, then run
   go build/vet/test from backend/ and report exact results.
B. Regenerate Swagger + OpenAPI client (make swagger, make api-client) once the
   backend API is final.
C. Build the FRONTEND per spec §5 (only after B — it needs the regenerated
   ims-api.d.ts): ledger "pending approval" state for LIVE cash, a submitter
   pending-requests panel with Cancel, the CASH_IN/CASH_OUT/FEE/DIVIDEND inputs
   (the ledger form currently only emits BUY/SELL — confirm this sub-scope with
   the owner), EN/TH/ZH copy (zh = Traditional), Vitest coverage. Use the repo's
   frontier-frontend-engineer skill / ims-frontend-implementer conventions and
   the generated typed client only.
D. LIVE browser verification (this is the whole point — two Stage 1 bugs passed
   every automated gate and were only caught by clicking through the UI):
   - On a LIVE portfolio, submit a cash movement → confirm it becomes PENDING
     (no transaction posted yet, cash unchanged), check the BROWSER CONSOLE for
     errors (not just the network tab).
   - Approve it as the configured approver → confirm exactly ONE real
     investment__portfolio_transactions row is created and cash updates.
   - Reject another → confirm no transaction.
   - Cancel a pending one as the submitter → confirm CANCELLED, no transaction,
     approver can no longer act.
   - Confirm SIMULATION still posts immediately and MODEL is still blocked.
   - For any field mapping to a DB column, inspect the actual CHECK constraints
     via docker exec into ims-postgres before trusting the Go type.
E. Report: what worked, what failed, what was fixed, exact files changed, exact
   commands + pass/fail, and browser/API/DB evidence. Then STOP and ask the owner
   to review and commit. Do not commit yourself.
```
