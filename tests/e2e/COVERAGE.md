# IAM E2E Coverage Report

Scope: end-to-end proof that IAM enforces authentication, RBAC/function permissions, and data permissions from the browser through the API to the database, against a dedicated `ims_e2e` database. All findings below came from actually running the suites against a real, freshly migrated+seeded database and a real running backend/frontend — nothing here is inferred from reading source alone.

## Covered, by scenario

**A. Authentication** (`specs/iam/auth.spec.ts`) — successful login, wrong password, disabled user, locked user, logout (real UI flow, session cleared, protected page blocked), session survives a page reload.

**B. Frontend route protection** (`specs/iam/route-protection.spec.ts`) — unauthenticated → `/auth/login`, no-permission → `/403`, admin succeeds, for both `/permissions/accounts` and `/administration/settings`.

**C. Backend authorization enforcement** (`specs/iam/api-authorization.spec.ts`, `backend/tests/e2e/iam_test.go`) — 401 unauthenticated, 403 authenticated-without-permission, 200 admin, disabled/locked login rejection, all asserted directly against the API independent of any frontend. The Go suite additionally proves data-permission scoping (Fund A/B) and that logout revokes the session's refresh token server-side.

**D. Permission-driven UI** (`specs/iam/permission-driven-ui.spec.ts`) — admin sees IAM/administration nav items, no-permission user doesn't; every "hidden menu" assertion is paired with a direct backend API call proving the hidden menu isn't the only thing enforcing access.

**E. Account management** (`specs/iam/account-management.spec.ts`) — admin creates an account via the real UI; disable/enable via UI with backend login rejection *and* proof that an already-issued token for the disabled user is rejected mid-session (403, not just failure at next login); lock/unlock via UI; reset-password via UI. Runs entirely against scratch users created fresh per test run — never touches the 7 fixed fixtures.

**F. Group / role / function permission** (`specs/iam/group-role-permission.spec.ts`) — the only mutation entry point that exists (`POST /permissions/users/{userId}/role-assignment-request`) is proven permission-gated (403 without `permission.change_request.create`, 201 for admin). See "Known gaps" below for what this does *not* cover.

**G. Data permission** (`specs/iam/data-permission.spec.ts`, `backend/tests/e2e/iam_test.go`) — `e2e_trader` (data-rights scoped to Fund A only, via `permissions_data_rights`) can read Fund A, gets 403 on Fund B, and the fund list endpoint filters accordingly. Enforced and asserted at the backend API layer.

**H. Audit / security trail** (`specs/iam/audit-trail.spec.ts`) — login success/failure, account creation, and account deactivation all produce queryable `GET /admin/audit` events; a no-permission user cannot read the audit log at all.

## Known gaps (documented, not faked)

1. **Activation gate** — the brief expects a newly created account to be unusable until activation/approval. Actual backend behavior (`backend/internal/iam/application/command/admin_user.go`) makes new accounts immediately active, only setting `force_password_change=true`. `account-management.spec.ts` asserts the real behavior and calls this out inline rather than testing a gate that doesn't exist.

2. **Logout doesn't revoke the server-side session from the frontend** — `frontend/app/stores/useAuthStore.ts`'s `logout()` only clears local cookies; the refresh token returned by login is never persisted client-side, so the store has nothing to call `POST /auth/logout` with. The *backend* endpoint itself correctly revokes sessions (proven directly in `backend/tests/e2e/iam_test.go`), but a real browser logout currently leaves the access token valid until its natural (short) expiry. Fixing this requires a token-storage design decision (where to persist a refresh token, ideally an httpOnly cookie set by the backend) that's out of scope for an E2E-test change — flagged here rather than patched under time pressure.

3. **SSR permission-hydration gap (found while writing this suite, reproduced with a real UI login)** — a *hard* navigation (typed URL, bookmark, page reload) to a permission-gated page evaluates the `permission` middleware during Nuxt's SSR pass, where `authStore.permissions` is still the empty default — `authStore.restoreSession()` (which populates it from the auth cookie via `GET /auth/me`) is explicitly client-only (`if (import.meta.client)` in `frontend/app/middleware/auth.ts`). Result: a fully authorized user hard-navigating to e.g. `/permissions/accounts` gets bounced to `/403`, not the page. Not a security hole (fails closed), but a real correctness/UX bug — any bookmark or hard refresh of a permission-gated page currently misbehaves for legitimate users. `route-protection.spec.ts` and `account-management.spec.ts` route around it by landing on `/` first and clicking through, which is also the more realistic user flow. Not fixed here; fixing it means deciding how permissions should be derived during SSR (e.g., decode the JWT server-side) — a design decision beyond this task's scope.

4. **Change-request submit → approve → merge not driven end-to-end** — `permission_role_assignments`/`approval_workflow_settings` encode risk-level routing, minimum-approvals, and self-approval rules that need dedicated backend research to exercise reliably. Driving the full cycle with guessed inputs risked a flaky, misleading test, so scenario F stops at proving the request-creation entry point is permission-gated and functional. "Permission change affects visible menu/API access after re-login" is proven instead by the static before/after contrast in `permission-driven-ui.spec.ts` and `data-permission.spec.ts`.

5. **MFA is entirely out of scope** — not one of the 7 required scenarios; no MFA-enrolled fixture exists.

6. **`TestE2E_InvestmentOversellEnvelope`** (pre-existing, in the same Go package as `iam_test.go`, unrelated to IAM) still has a bug (`ERROR: inconsistent types deduced for parameter $4` in its `seedOpeningPosition` helper) found as a side effect of being the first real run of `backend/tests/e2e`. Three other pre-existing bugs in that same test were fixed along the way (see "Bugs fixed" below) because they blocked `go build -tags e2e` and `go test -tags e2e` entirely for *this* package, including the new IAM tests. This fourth one is investment-module-scoped and out of bounds for an IAM task — `make test-e2e-backend` runs with `-run TestE2E_IAM` specifically to avoid shipping a permanently-red CI job for someone else's module.

## Bugs fixed along the way

Found and fixed because they directly blocked this suite from running or passing — verified with a real before/after run against a live database each time, not just reasoned about:

- `GET /admin/users` returned 500 whenever *any* user had a `NULL` email — always true once system actor rows (`system.market_data`, `system.scheduler`, seeded by every `make seed` run) exist. `entity.User.Email` is a non-nullable `string`, but the column is nullable; `backend/internal/iam/infrastructure/persistence/user_repository.go`'s scans now go through a nullable intermediate.
- `backend/tests/e2e/investment_test.go` didn't compile with `-tags e2e` (`marketdata.NewModule` gained a required `SecurityResolver` argument that this test never updated) — fixed by wiring the `reference_data` module into `bootTestServer`, matching `cmd/server/main.go`.
- Same file's `loginAdmin()` helper decoded the login response without unwrapping the `{"data": ...}` envelope (`httputil.OK` wraps all success responses) — `AccessToken` was silently always empty.
- Same file queried `investment__countries.code`, but the actual column is `iso_code`.
- Same file's admin login assertion (`force_password_change` must be false) fails against a freshly seeded database by design (the bootstrap admin seed sets it true) — fixed by resetting the flag in test setup rather than requiring a manual first-login step before the suite can run.
- `frontend/app/pages/administration/settings.vue` had no route-level permission guard (only `auth`), unlike every other `/permissions/*` admin page — fixed to match that pattern (see gap #3 above for the deeper SSR issue this doesn't cover).

## Flaky-risk areas

- **Nuxt dev-mode hydration timing**: filling the login form before client hydration completes silently no-ops (the DOM shows the typed value but the bound `ref` never updates, so the submit button stays disabled). `support/auth.ts`'s `loginViaUI` waits for `networkidle` before interacting; this held up under 2-worker parallel runs in this session but dev-mode compile time is inherently variable — a much slower CI runner could still be flaky here. A production build (`nuxt build` + serve) would likely not have this issue at all; not verified in this session for time reasons.
- **Login rate limiting**: `RATE_LIMIT_LOGIN_PER_USER` is bumped to 50/1m in `infra/env/.env.e2e.example` specifically so repeated wrong-password assertions across re-runs in the same minute don't 429 before asserting 401. Lowering it without updating the tests will cause intermittent failures on fast re-runs.
- **Sidebar nav clicks**: `route-protection.spec.ts` and `account-management.spec.ts` dispatch nav-link clicks via `locator.evaluate(el => el.click())` rather than Playwright's pointer click, working around a viewport-actionability false-negative in the sidebar's nested scroll container (confirmed the link is genuinely visible via a separate `toBeVisible()` assertion in `permission-driven-ui.spec.ts`). If the sidebar layout changes, this may need revisiting.
- **Scratch user accumulation**: `account-management.spec.ts` and one case in `audit-trail.spec.ts` create uniquely-named scratch users (`e2e_scratch_<timestamp>`) per run rather than reusing/resetting shared fixtures, to stay parallel-safe and avoid needing a database reset between runs. Harmless for correctness (CI always starts from a fresh database), but a long-lived local `ims_e2e` will accumulate these over many manual runs — `make e2e-db-setup` doesn't clean them up (it only re-asserts the 7 fixed fixtures); a full local reset requires dropping and recreating `ims_e2e`.
- **Session idle/absolute timeout and IP allowlist behavior** are not tested — the former is real-time dependent (would need clock manipulation), the latter is fail-open when unset (`ADMIN_IP_ALLOWLIST=` in the E2E env) and can't be positively tested for the "blocks a disallowed IP" case without a second, deliberately-restricted environment.
