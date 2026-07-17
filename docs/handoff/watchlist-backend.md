# Watchlist Module — Handoff

**Status:** Backend complete, build clean, domain tests passing. Frontend not yet implemented.  
**Base path:** `/api/v1/watchlists`  
**Module path:** `backend/internal/watchlist`

---

## What it does

Allows users to track securities and receive threshold alerts when market prices cross a configured level.

- A **watchlist item** pins a security to either a personal or portfolio scope.
- Each item can carry one or more **threshold rules** (MARKET_PRICE, ABOVE or BELOW a decimal value).
- The **evaluator** runs those rules against live quotes, commits alert events, fires notifications, and records audit entries.
- Users see their **alert events** and can **acknowledge** them.

v1 supports **PERSONAL** and **PORTFOLIO** scopes only. MARKET_PRICE is the only metric type. No other thresholds or scopes exist in v1.

---

## Database

Three migration files (all under `database/migrations/`, timestamped `20260624`):

| File | Table |
|---|---|
| `20260624000001_watchlist__create_items` | `watchlist_items` |
| `20260624000002_watchlist__create_threshold_rules` | `watchlist_threshold_rules` |
| `20260624000003_watchlist__create_alert_events` | `watchlist_alert_events` |

### watchlist_items

Soft-delete table. `scope_type` is `PERSONAL` or `PORTFOLIO`.  
- PERSONAL items require `owner_user_id`.  
- PORTFOLIO items require `portfolio_id` (references `investment__portfolios.id`).  
- Unique active constraint per `(owner_user_id, security_id)` for personal and `(portfolio_id, security_id)` for portfolio.

### watchlist_threshold_rules

Soft-delete table. One rule per `(watchlist_item_id, metric_type, direction, threshold_value)` (unique index, partial on `deleted_at IS NULL`).  
- `metric_type` is locked to `'MARKET_PRICE'` by a CHECK constraint.  
- `last_state` tracks `UNKNOWN → NON_BREACHED ↔ BREACHED`.  
- `cooldown_minutes` (default 60) suppresses repeated alerts within the window.  
- `last_alerted_at`, `last_observed_price`, `last_observed_at`, `last_evaluated_at` are updated every evaluation cycle.

### watchlist_alert_events

Immutable once inserted. `current_state` is always `'BREACHED'` (enforced by CHECK).  
- `idempotency_key` has a UNIQUE constraint — insert conflicts are handled as `ErrAlertIdempotencyConflict`.  
- `notification_status` (`PENDING → CREATED | FAILED | SUPPRESSED | SKIPPED`) is updated post-commit as a best-effort non-transactional write.  
- `acknowledged_by` / `acknowledged_at` are set together (enforced by CHECK).

---

## API Endpoints

All routes require a valid bearer token. They are mounted by `module.go → RegisterRoutes` under the authenticated router.

| Method | Path | Permission | Handler |
|---|---|---|---|
| `GET` | `/watchlists` | `WATCHLIST_VIEW` | `ListItems` |
| `POST` | `/watchlists/items` | `WATCHLIST_MANAGE` | `CreateItem` |
| `PATCH` | `/watchlists/items/{id}` | `WATCHLIST_MANAGE` | `UpdateItem` |
| `DELETE` | `/watchlists/items/{id}` | `WATCHLIST_MANAGE` | `DeleteItem` |
| `GET` | `/watchlists/alerts` | `WATCHLIST_VIEW` | `ListAlerts` |
| `POST` | `/watchlists/alerts/{id}/acknowledge` | `WATCHLIST_ALERT_ACK` | `AcknowledgeAlert` |
| `POST` | `/watchlists/evaluate` | `WATCHLIST_EVALUATE` | `ManualEvaluate` |

### Permission codes (defined in `permission/policies.go`)

| Code | Purpose |
|---|---|
| `WATCHLIST_VIEW` | Read items and alert events |
| `WATCHLIST_MANAGE` | Create, update, soft-delete items and rules |
| `WATCHLIST_ALERT_ACK` | Acknowledge alert events |
| `WATCHLIST_EVALUATE` | Manually trigger evaluation (ops/scheduler only — do not grant to regular users) |
| `WATCHLIST_ADMIN` | Optional override for admin/risk/audit use cases |

Run `make seed` after deploying to upsert the permission catalog.

### Key request/response notes

**CreateItem / UpdateItem** (`threshold_rules` field):  
PATCH uses **full-replacement semantics** — the entire `threshold_rules` array replaces existing enabled rules. Omitting the field leaves existing rules untouched. Passing an empty array removes all rules.

**ListItems / ListAlerts query params:**

```
scope_type    PERSONAL | PORTFOLIO
portfolio_id  UUID — required when scope_type=PORTFOLIO (ListItems only enforces this implicitly; ListAlerts returns 400 if portfolio_id is provided without scope_type=PORTFOLIO)
security_id   UUID filter
include_disabled  bool (items only)
limit / offset
```

**ManualEvaluate body:**

```json
{
  "scope_type": "PERSONAL",
  "portfolio_id": "uuid or omit",
  "security_id": "uuid or omit",
  "item_id": "uuid or omit",
  "rule_id": "uuid or omit",
  "dry_run": false
}
```

Dry-run evaluates against the last-known snapshot without acquiring row locks or writing anything.

---

## Module Architecture

```
internal/watchlist/
  domain/
    entity/             WatchlistItem, ThresholdRule, AlertEvent
    service/            Evaluate() pure function — state machine logic
    repository/         repository interfaces + filter structs
    ports.go            SecurityPort, QuotePort, PortfolioScopePort,
                        UserLookupPort, WatchlistAlertNotifier, WatchlistAuditRecorder
  application/
    command/            CreateItem, UpdateItem, DeleteItem, AcknowledgeAlert
    query/              ListItems, ListAlerts
    service/            EvaluatorService
  infrastructure/
    persistence/        PostgresWatchlistItemRepository
                        PostgresThresholdRuleRepository
                        PostgresAlertEventRepository
    adapter/            SecurityAdapter, QuoteAdapter, PortfolioScopeAdapter,
                        AuditAdapter, PermissionCheckerAdapter,
                        PostgresUserLookupAdapter
  transport/http/       Handler, router, dto
  permission/           policies.go (permission catalog provider)
  module.go             dependency wiring
```

---

## Cross-Module Dependencies

The watchlist module does **not** import any other module's `internal/` package. All cross-module consumption goes through ports.

| Port / Interface | Provider | How resolved |
|---|---|---|
| `SecurityPort` | `reference_data` module | `SecurityAdapter` wraps `refdatadomain.SecurityResolver` |
| `QuotePort` | `market_data` module | `QuoteAdapter` wraps `contract.MarketQuoteProvider` |
| `PortfolioScopePort` | `investment` module | `PortfolioScopeAdapter` wraps `contract.PortfolioScopeResolver` |
| `WatchlistAlertNotifier` | `notification` module | `notification/infrastructure/adapter/watchlist_alert_notifier.go` |
| `WatchlistAuditRecorder` | `audit` module | `AuditAdapter` wraps `auditdomain.Recorder` |
| `PermissionChecker` | `iam` module | `PermissionCheckerAdapter` wraps `wladapter.IAMPermissionPort` |
| `UserLookupPort` | iam_users table (direct SQL) | `PostgresUserLookupAdapter` — same pattern as `approval/infrastructure/adapter/user_directory.go` |

`PostgresUserLookupAdapter` queries `iam_users` directly using `SELECT id, COALESCE(display_name, username, '') FROM iam_users WHERE id = ANY($1)`. This is intentional: it avoids an IAM module import while keeping the query read-only and projection-only.

---

## Evaluator Design

The evaluator (`application/service/evaluator_service.go`) is the most safety-critical component.

### Concurrency safety (SELECT FOR UPDATE)

The naive approach — read rule state, evaluate, write — has a TOCTOU race where two concurrent evaluators both see `NON_BREACHED` and both emit alerts for the same crossing. The evaluator prevents this:

1. **Fetch the quote outside any transaction** (avoids holding a DB row lock across a network call).
2. Open a transaction.
3. `GetByIDForUpdate(ctx, tx, ruleID)` — issues `SELECT ... FOR UPDATE`. Second evaluator blocks here until the first commits.
4. Re-evaluate against the freshly-locked row's `LastState`.
5. If `OutcomeAlertCreated`: insert the alert event (idempotency key provides a second safety net).
6. `UpdateState` within the same transaction.
7. Commit.
8. Post-commit (non-transactional): send notification, update `notification_status`, record audit.

Dry-run skips steps 2–8 entirely and uses the snapshot from `ListForEvaluation`.

### Idempotency key

`AlertIdempotencyKey(ruleID, direction, observedAt, cooldownMinutes)` truncates the quote timestamp to the cooldown bucket. Re-runs within the same cooldown window produce the same key. If the unique constraint fires on insert, `ErrAlertIdempotencyConflict` is returned and the caller marks the alert suppressed without erroring.

### Stale quote handling

`DefaultMaxStaleAge` is 15 minutes. A stale quote older than that returns `ErrStaleQuote` → the rule is skipped (`NotificationStatusSkipped`). A stale quote within the max age proceeds — the alert event records `stale=true` and `stale_reason`.

### Per-rule errors in batch mode

When `filter.RuleID` is nil (scheduler batch run), per-rule errors are swallowed, counted as `ProviderFailures`, and recorded as `WATCHLIST_EVALUATION_FAILED` audit events. When `filter.RuleID` is set (targeted single-rule run), errors propagate to the caller.

---

## Data-Scope Security

Two places enforce fund-based data permissions for PORTFOLIO-scoped operations.

**ListItems (query/list_items.go):**  
For PORTFOLIO scope without an explicit `portfolio_id`, the handler calls `iamChecker.GetAccessibleContracts(actorID)` to get the actor's fund UUIDs, then sets `filter.FundIDs`. The SQL translates this as:

```sql
EXISTS (
  SELECT 1 FROM investment__portfolios p
  WHERE p.id = watchlist_items.portfolio_id
    AND p.fund_id = ANY($N)
)
```

**ListAlerts (query/list_alerts.go):** Same pattern.

The outer table column is always qualified (`watchlist_items.portfolio_id`, `watchlist_alert_events.portfolio_id`) to prevent silent correlation if `investment__portfolios` ever gains a `portfolio_id` column.

Wildcard actors (those whose contracts contain `"*"`) bypass the fund filter and see all portfolio items.

---

## Audit Events

| Event type | Emitted when |
|---|---|
| `WATCHLIST_ITEM_CREATED` | Item created |
| `WATCHLIST_ITEM_UPDATED` | Item metadata updated |
| `WATCHLIST_ITEM_DELETED` | Item soft-deleted |
| `WATCHLIST_THRESHOLD_CREATED` | New threshold rule created (post-commit) |
| `WATCHLIST_THRESHOLD_DISABLED` | Existing rule disabled by full-replace PATCH (post-commit) |
| `WATCHLIST_ALERT_TRIGGERED` | Alert event inserted and notification attempted |
| `WATCHLIST_ALERT_ACKNOWLEDGED` | Alert acknowledged |
| `WATCHLIST_EVALUATION_FAILED` | Per-rule evaluator error in batch mode |

---

## Tests

`backend/internal/watchlist/domain/service/evaluator_test.go` — 16 table-driven unit tests covering the pure `Evaluate()` function:

- UNKNOWN → NON_BREACHED / BREACHED (state seeding, no alert)
- NON_BREACHED → BREACHED ABOVE and BELOW
- NON_BREACHED stays NON_BREACHED; BREACHED stays BREACHED
- BREACHED → NON_BREACHED (re-arm)
- Inside cooldown suppression; outside cooldown creates alert
- Stale beyond max age → `ErrStaleQuote`
- Stale within max age → accepted (alert created, `stale=true`)
- ABOVE/BELOW exact boundary conditions

`TestAlertIdempotencyKey` verifies same cooldown bucket → same key, different direction → different key.

No integration tests yet. The SQL filter paths (fund-scoped EXISTS subquery, FundIDs in alert filter) are not covered by automated tests.

---

## Known Gaps / v2 Candidates

- **N+1 queries in ListItems**: per-item security info and portfolio descriptor are fetched one-by-one. For large lists this is slow. Batch security/portfolio lookups would require adding batch methods to `SecurityPort` and `PortfolioScopePort`.
- **Scheduler integration not wired**: `EvaluatorService.Evaluate` is callable but no cron or scheduled job calls it automatically. A scheduler job (e.g., under `jobs/`) needs to be added and registered in `module.go` or `cmd/server/main.go`.
- **No Swagger annotations on evaluate endpoint**: `ManualEvaluate` is missing `@Param` / `@Success` annotations for Swagger generation.
- **WATCHLIST_ADMIN permission unused**: the code declares it but no route checks it. Reserved for a future admin override or bulk-disable endpoint.
- **P3-1 (nil scope_type defaults)**: ListAlerts with no `scope_type` currently returns items across both scopes, filtered to what the actor can see. This is safe but may surprise callers who expect a strict scope filter.

---

## Wiring Checklist (deploying for the first time)

1. Run `make migrate-up` — applies the three watchlist migration files.
2. Run `make seed` — upserts the five `WATCHLIST_*` permission codes into the permission catalog.
3. Assign permissions to roles in the permissions admin UI.
4. Wire a scheduler to call `POST /api/v1/watchlists/evaluate` (or call `EvaluatorService.Evaluate` internally from a job) on the desired cadence.
5. Ensure `notification` module is live — `WatchlistAlertNotifier` is implemented in `notification/infrastructure/adapter/watchlist_alert_notifier.go`.
