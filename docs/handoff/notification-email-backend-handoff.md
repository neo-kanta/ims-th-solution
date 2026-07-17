# Notification Email Backend — Handoff

**Date:** 2026-06-19  
**Spec:** `docs/api/notification-email-api.md`  
**Status:** Backend code complete; database migration and SMTP config required before demo.

---

## What Was Built

The notification module was extended from an in-app-only baseline to a full email outbox system:

| Area | Files changed / created |
|---|---|
| Shared enum | `backend/pkg/enum/notification_event.go` |
| Event catalog | `backend/internal/notification/domain/valueobject/event_catalog.go` |
| Email outbox entity | `backend/internal/notification/domain/entity/email_outbox.go` |
| Email sender interface | `backend/internal/notification/domain/email_sender.go` |
| Domain repository | `backend/internal/notification/domain/repository.go` (new interfaces + UserSummary) |
| Notification entity | `backend/internal/notification/domain/entity/notification.go` (new context fields) |
| Permission catalog | `backend/internal/notification/permission/policies.go` (5 new codes) |
| Config | `backend/platform/config/config.go` (SMTP + email worker fields) |
| Notification repo | `backend/internal/notification/infrastructure/persistence/notification_repository.go` |
| Email outbox repo | `backend/internal/notification/infrastructure/persistence/email_outbox_repository.go` |
| SMTP sender | `backend/internal/notification/infrastructure/email/smtp_sender.go` |
| Email template svc | `backend/internal/notification/application/service/email_template_service.go` |
| Notification service | `backend/internal/notification/application/service/notification_service.go` |
| Email outbox service | `backend/internal/notification/application/service/email_outbox_service.go` |
| Email outbox worker | `backend/internal/notification/jobs/email_outbox_worker.go` |
| Notification handler | `backend/internal/notification/transport/handler/notification_handler.go` |
| Email outbox handler | `backend/internal/notification/transport/handler/email_outbox_handler.go` |
| Router | `backend/internal/notification/transport/router.go` |
| Module wiring | `backend/internal/notification/module.go` |
| Server wiring | `backend/cmd/server/main.go` |
| Approval adapter | `backend/internal/notification/infrastructure/adapter/approval_notifier.go` |
| Workflow adapter | `backend/internal/notification/infrastructure/adapter/workflow_stuck_day_notifier.go` |
| Migrations | `database/migrations/20260619000001_*` and `20260619000002_*` |
| Permission seed | `database/seeds/016_notification_permission_seed.sql` |
| Docker compose | `infra/docker-compose.yml` (Mailpit service added) |

---

## Permission Codes

The notification module now owns six UPPERCASE permission codes (matching the existing `NOTIFICATION_CONFIG` pattern used throughout the permission system):

| Code | Purpose |
|---|---|
| `NOTIFICATION_CONFIG` | Existing: configure templates and rules |
| `NOTIFICATION_VIEW` | Read email outbox list and detail |
| `NOTIFICATION_EDIT` | General admin |
| `NOTIFICATION_RETRY` | Retry FAILED/DEAD outbox rows |
| `NOTIFICATION_TEST` | POST /email/test |
| `NOTIFICATION_HEALTH` | GET /email/health |

**Note:** The spec document uses dotted names (`permission.notification.view`). These are documentation aliases. The actual enforcement codes in `permissions_function_rights` are the UPPERCASE codes listed above.

---

## Operator Steps Required Before Demo

### 1. Run database migrations

```bash
make migrate-up
```

This runs:
- `20260619000001` — adds `event_type`, `idempotency_key`, `business_*`, `action_*` columns to `notification__notifications` and back-fills `event_type = category` for all existing rows.
- `20260619000002` — creates `notification__email_outbox` with all indexes.

### 2. Run the seeder

```bash
make seed
```

This upserts the new permission codes into `permissions_function_definitions` and grants them to the Admin group via `016_notification_permission_seed.sql`.

### 3. Configure environment

Copy the example variables into `backend/.env` (gitignored):

**Local Mailpit demo (zero cost, no real email):**

```env
NOTIFICATION_EMAIL_ENABLED=true
NOTIFICATION_EMAIL_WORKER_ENABLED=true
NOTIFICATION_EMAIL_WORKER_INTERVAL=10s
NOTIFICATION_EMAIL_WORKER_BATCH_SIZE=25
NOTIFICATION_EMAIL_MAX_ATTEMPTS=5
NOTIFICATION_EMAIL_STALE_SENDING_TIMEOUT=10m
NOTIFICATION_EMAIL_RETRY_POLICY=1m,5m,15m,1h
NOTIFICATION_EMAIL_ALLOW_RAW_EMAIL=false
NOTIFICATION_EMAIL_ALLOWED_DOMAINS=gmail.com,systemweb.com.th
NOTIFICATION_EMAIL_TEST_ENDPOINT_ENABLED=true
NOTIFICATION_EMAIL_SEND_REAL_EMAIL=false

SMTP_HOST=mailpit
SMTP_PORT=1025
SMTP_TLS_MODE=none
SMTP_FROM_ADDRESS=no-reply@ims-demo.local
SMTP_FROM_NAME=IMS Thailand Demo
SMTP_TIMEOUT=10s

APP_PUBLIC_BASE_URL=http://localhost:3000
```

**Gmail SMTP demo (real outbound email):**

```env
NOTIFICATION_EMAIL_SEND_REAL_EMAIL=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_TLS_MODE=starttls
SMTP_USERNAME=<your-gmail-address>
SMTP_PASSWORD=<gmail-app-password>    # NOT the account password
SMTP_FROM_ADDRESS=<your-gmail-address>
NOTIFICATION_EMAIL_ALLOWED_DOMAINS=gmail.com,systemweb.com.th
```

**Important:** Store `SMTP_PASSWORD` in a gitignored file or secret manager. It must never be committed.

### 4. Start Mailpit (local demo)

```bash
cd infra && docker compose up -d mailpit
```

Mailpit web UI: http://localhost:8025

### 5. Regenerate Swagger

```bash
make swagger
```

Do **not** run `make api-client` (frontend type regeneration is out of scope for this change).

---

## What Was Verified

- `go vet ./internal/notification/...` passes with zero errors.
- `go vet ./cmd/...` passes with zero errors (confirms module wiring in main.go compiles).
- `go build ./internal/notification/...` compiles successfully.

**Not verified in this environment** (requires running DB, SMTP, and worker):
- Actual migration execution. Verify migration idempotency key constraint with two back-to-back inserts sharing the same idempotency_key — the second must return no rows without error.
- `MarkFailed` / `RecoverStale` use `make_interval(secs => $N)` so Go durations become interval arithmetic correctly. Confirm with: `SELECT make_interval(secs => 60), make_interval(secs => 3600);` on the connected DB.
- Outbox worker claiming rows, sending email, and marking SENT.
- Mailpit receiving the test email.
- Gmail real send.

---

## Key Design Decisions

### Idempotency
- Approval notifications pass `idempotency_key = "APPROVAL_TASK_ASSIGNED:{request_id}:{recipient_id}"`.
- Email outbox rows get `idempotency_key = "email:{notification_idempotency_key}"` to keep namespaces separate.
- Both use `ON CONFLICT (idempotency_key) DO NOTHING RETURNING id`; when no row returns, the service treats it as already-queued success.

### Backward compatibility
- Existing notifications that have no `event_type` get back-filled with `category` via the migration.
- The API response includes both `notification_id` (spec) and `id` (legacy alias) to avoid breaking existing frontend code.
- Workflow stuck-day notifications now use `WORKFLOW_STUCK_DAY` as the canonical event type. The old `WORKFLOW_STUCK_{STATE}` category is still stored in `category` for historic rows; the event catalog falls back gracefully for unknown types.
- The `notification.NewModule` signature changed to accept `(pool, cfg, checker)`. Callers that pass `nil` for cfg get the in-app-only baseline.

### Permission enforcement
- User-facing notification center (`GET /notifications`, `POST /read`, `POST /read-all`) requires only authentication — no function permission needed.
- All email admin endpoints require the specific permission code above.

### Worker safety
- `FOR UPDATE SKIP LOCKED` prevents multiple instances from sending the same email.
- Stale `SENDING` recovery runs on worker startup and before each batch claim.
- `SMTP_PASSWORD` never appears in any API response.

### Permission OR-semantics (known deviation)
The spec describes `health → health OR view` and `retry → retry OR edit`. The current router enforces a single specific code per route because `RequirePermission` accepts one code. In practice, Admin has all codes so the demo is unaffected. A `VIEW`-only operator cannot hit `/email/health` under this implementation. Honoring the OR would require a new middleware variant — deferred.

### Email gate per event type
`createEmailOutbox` in `notification_service.go` sends email for any event that reaches `Service.Create` with a resolvable recipient — there is no per-event opt-out flag. This is correct for all current callers (approval and workflow stuck-day both want email). If a future caller wants in-app-only, it should either set `cfg.EmailEnabled = false` at the module level or pass an empty `EmailConfig` via `WithEmail`.

### Workflow stuck-day deduplication change
The stable idempotency key `WORKFLOW_STUCK_DAY:{day_id}:{state}:{recipient_id}` means each `(day, state, recipient)` triple receives exactly one notification for the lifetime of the row, replacing the old 23-hour re-reminder window. The `dedupeWindowHours` field on the notifier no longer has effect on this path. This matches the spec's stated key structure and is intentional.

---

## What Is NOT Done (deferred)

- Swagger generation (`make swagger`) — must be run by operator after starting the backend.
- Frontend API client regeneration (`make api-client`) — out of scope.
- `ALERT_THRESHOLD_BREACHED` and `PERMISSION_GRANTED` notification triggers — no alert/risk/permission business modules exist yet; stubs are in the event catalog for future wiring.
- Email templates are currently inline Go string formatting; a future iteration can move to embedded HTML templates.
