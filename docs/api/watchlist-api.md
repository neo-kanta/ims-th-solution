# Watchlist API Documentation

**Status:** original design/API contract; the Watchlist module is now implemented
**Current runtime contract:** [watchlist-current-api.md](watchlist-current-api.md)
**State:** This document preserves the design rationale and future-policy discussion. Use the current runtime document, mounted router, handlers, and Swagger annotations for client integration.

Primary architecture reference: [docs/architecture/watchlist-backend-design.md](../architecture/watchlist-backend-design.md)

## API Overview

The Watchlist API manages saved security monitoring lists and market-price threshold alerts. Watchlist is implemented as a standalone backend module at `backend/internal/watchlist`.

Supported in v1:

- `PERSONAL` watchlists owned by the authenticated user.
- `PORTFOLIO` watchlists attached to one portfolio.
- `MARKET_PRICE` threshold rules only.
- Crossing-based alerts with cooldown and re-arm behavior.
- Alert delivery through the notification module using event type `ALERT_THRESHOLD_BREACHED`.

Alert behavior summary:

- `ABOVE` rules breach when `observed_price >= threshold_value`.
- `BELOW` rules breach when `observed_price <= threshold_value`.
- First evaluation from `UNKNOWN` seeds state and does not notify, even if already breached.
- `NON_BREACHED` to `BREACHED` creates an alert when cooldown allows.
- `BREACHED` to `BREACHED` does not create repeated alerts.
- `BREACHED` to `NON_BREACHED` re-arms the rule.
- Default cooldown is 60 minutes.
- Stale quote alerting is allowed only when the stale quote age is within the configured max age. Default max age is 15 minutes.

Explicitly out of scope for v1:

- NAV thresholds.
- AUM thresholds.
- Portfolio valuation thresholds.
- Risk thresholds.
- Portfolio-performance thresholds.
- Portfolio recipient groups.
- Extending `/api/v1/market-data/screen/watchlist` into the real Watchlist domain API.
- Direct provider calls from Watchlist. Watchlist depends on the market-data quote provider port only.

## Common API Conventions

| Convention | Contract |
| --- | --- |
| Base path | `/api/v1/watchlists` |
| Authentication | Required for every endpoint through IAM auth middleware. |
| Function permission | Enforced with `middleware.RequirePermission` using the endpoint's Watchlist permission code. |
| Portfolio data permission | Enforced in the Watchlist application service after resolving request body, query, or loaded row scope. |
| Success envelope | `httputil.SuccessResponse`: `{ "data": ... }` with optional `"message"`. |
| Error envelope | `httputil.ErrorResponse`: `{ "error": "...", "code": "...", "details": ... }`. `code` and `details` are optional in the platform type, but Watchlist errors should populate `code`. |
| Timestamp format | RFC3339 UTC string, for example `2026-06-24T04:30:00Z`. |
| UUID format | Canonical UUID string, for example `9f357932-d590-41f6-9c3e-6ef8301a6f92`. |
| Decimal format | String decimal for prices, thresholds, previous close, and percentage values, for example `"184.25000000"`. Do not return JSON numbers for money or price fields. |
| Pagination | `limit` and `offset` query parameters. Response data includes `pagination` with `limit`, `offset`, and `total`. |
| Sorting | No public `sort` query parameter in v1. Item lists sort by `pinned desc`, `display_order asc`, `updated_at desc`, `id asc`. Alert lists sort by `created_at desc`, `id desc`. |
| Idempotency | Alert events and notifications use a generated idempotency key. Client-supplied idempotency keys are not part of v1 public API. |
| Soft delete | Deleting an item sets deleted metadata and stops all child rule evaluation. Deleted alert events are not supported. |

Watchlist handlers must use `httputil.JSON(status, httputil.ErrorResponse{Error: ..., Code: ..., Details: ...})` or future coded helper functions when returning stable Watchlist error codes. Do not use plain `httputil.BadRequest`, `httputil.Forbidden`, `httputil.NotFound`, `httputil.Conflict`, or similar helpers for Watchlist contract errors unless those helpers are extended to accept and emit `code` and `details`.

### Frontend Display Contract

- All `*_id` fields are transport identifiers. Top-level resource `id` fields are also transport identifiers.
- Frontend may store IDs in Pinia and use them for component keys, route params, mutations, cache normalization, and API calls.
- Frontend must not render raw UUIDs as user-facing labels.
- User-facing display should use descriptor objects and display fields, such as `security.display_symbol`, `security.name`, `portfolio.display_name`, and user descriptor `display_name`.
- Error `details` may contain UUIDs for debugging and support, but frontend should not show raw `details` directly to users. Show a localized friendly error message and keep raw details for logs or developer tooling.
- `owner_user_id`, `created_by_user_id`, and `acknowledged_by` are identifiers only. If UI needs to show creator or acknowledger labels, use `created_by_user` or `acknowledged_by_user`.
- If `created_by_user` or `acknowledged_by_user` is not populated in v1, frontend should omit that label rather than showing a UUID.

Recommended alert display guidance:

- Security display: use `security.display_symbol` and `security.name`.
- Portfolio display: use `portfolio.display_name`; never show `portfolio_id`.
- Price display: format `observed_price`, `threshold_value`, and `currency` together.
- Direction display: map `ABOVE` and `BELOW` to localized UI copy.
- Alert state display: use `acknowledgement_state`, `notification_status`, `stale`, and `stale_reason`.
- Never show `watchlist_item_id`, `threshold_rule_id`, `security_id`, `portfolio_id`, `created_by_user_id`, or `acknowledged_by` as labels.

## Enums

### `scope_type`

| Value | Meaning |
| --- | --- |
| `PERSONAL` | Watchlist item is owned by one user. `owner_user_id` and `created_by_user_id` are required. `portfolio_id` is null. |
| `PORTFOLIO` | Watchlist item belongs to one portfolio. `portfolio_id` is required. Access uses portfolio data permission. |

### `item_status`

| Value | Meaning |
| --- | --- |
| `ACTIVE` | Item is visible and eligible for rule evaluation. |
| `DISABLED` | Item remains stored but is not eligible for rule evaluation. |

### `metric_type`

| Value | Meaning |
| --- | --- |
| `MARKET_PRICE` | Threshold compares against market-data `MarketQuote.Price`. This is the only supported v1 metric. |

### `direction`

| Value | Meaning |
| --- | --- |
| `ABOVE` | Rule breaches when observed market price is greater than or equal to threshold. |
| `BELOW` | Rule breaches when observed market price is less than or equal to threshold. |

### `rule_status`

| Value | Meaning |
| --- | --- |
| `ENABLED` | Rule is eligible for evaluation. |
| `DISABLED` | Rule remains stored but is skipped by evaluation. |

### `rule_state`

| Value | Meaning |
| --- | --- |
| `UNKNOWN` | Rule has not been evaluated yet, or state has been reset. |
| `NON_BREACHED` | Latest usable quote is on the non-breached side. |
| `BREACHED` | Latest usable quote is on the breached side. |

### `notification_status`

| Value | Meaning |
| --- | --- |
| `PENDING` | Alert event exists and notification dispatch has not completed. |
| `CREATED` | Notification module accepted or created the notification. |
| `SUPPRESSED` | Duplicate alert or notification was idempotently suppressed. Use for notification suppression only when the Watchlist implementation can positively detect that outcome. |
| `FAILED` | Notification creation failed after the alert event was persisted. |
| `SKIPPED` | Evaluation did not require notification, such as dry run or operational skip. |

### `acknowledgement_state`

| Value | Meaning |
| --- | --- |
| `UNACKNOWLEDGED` | `acknowledged_at` is null. |
| `ACKNOWLEDGED` | `acknowledged_at` is populated. |

### `quote_status`

Used in manual evaluation result details when returned.

| Value | Meaning |
| --- | --- |
| `LIVE` | Provider returned a usable non-stale quote. |
| `STALE_ACCEPTED` | Market-data returned a stale fallback quote within the configured max age. |
| `STALE_SKIPPED` | Market-data returned a stale fallback quote older than the configured max age. No alert is created. |
| `UNAVAILABLE` | No live or accepted stale quote was available. No alert is created. |

## Permission Matrix

Endpoint permission codes:

| Permission | Purpose |
| --- | --- |
| `WATCHLIST_VIEW` | List visible watchlist items and visible alert events. |
| `WATCHLIST_MANAGE` | Create, update, disable, and soft-delete watchlist items and threshold rules. |
| `WATCHLIST_ALERT_ACK` | Acknowledge visible alert events. |
| `WATCHLIST_EVALUATE` | Manually trigger operational evaluation. Normal users must not receive this permission. |
| `WATCHLIST_ADMIN` | Optional override for approved admin/risk/audit use cases. Exact scope is an open decision. |

Access matrix:

| Actor/context | List own personal | Manage own personal | List non-owner personal | Manage non-owner personal | List portfolio with data permission | Manage portfolio with data permission | Portfolio without data permission | Manual evaluate |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Personal owner with `WATCHLIST_VIEW` | Allow | Deny without `WATCHLIST_MANAGE` | Deny | Deny | Depends on data permission | Deny without `WATCHLIST_MANAGE` | Deny | Deny |
| Personal owner with `WATCHLIST_MANAGE` | Allow if also has `WATCHLIST_VIEW` | Allow | Deny | Deny | Depends on data permission | Allow if data permission exists | Deny | Deny |
| Personal non-owner | Deny | Deny | Deny | Deny | Depends on data permission | Depends on `WATCHLIST_MANAGE` and data permission | Deny | Deny |
| Portfolio user with data permission | No personal access unless owner | No personal access unless owner | Deny | Deny | Allow with `WATCHLIST_VIEW` | Allow with `WATCHLIST_MANAGE` | Deny | Deny unless `WATCHLIST_EVALUATE` |
| Portfolio user without data permission | No personal access unless owner | No personal access unless owner | Deny | Deny | Deny | Deny | Deny | Deny |
| Admin/risk/audit with `WATCHLIST_ADMIN` | Open decision | Open decision | Open decision | Open decision | Should still require data permission unless global policy is approved | Should still require data permission unless global policy is approved | Deny unless global data policy is approved | Requires `WATCHLIST_EVALUATE` |
| Operational evaluator | Not a user-facing actor | Not a user-facing actor | Not a user-facing actor | Not a user-facing actor | May evaluate rules by system context | May evaluate rules by system context | Not applicable | Allow only through scheduler or `WATCHLIST_EVALUATE` |

Implementation notes:

- Route-level function permissions should use `middleware.RequirePermission`.
- Portfolio scope cannot rely only on `RequireDataPermission` because `portfolio_id` may be in JSON bodies or loaded database rows.
- Watchlist must resolve portfolio access through a narrow investment port before IAM data permission checks if IAM data scopes are fund/contract-based rather than raw portfolio IDs.
- Personal access must compare `owner_user_id` to the authenticated IAM subject.

## Endpoints

### A. List Watchlist Items

| Field | Value |
| --- | --- |
| Purpose | List watchlist items visible to the authenticated actor. |
| Method and path | `GET /api/v1/watchlists` |
| Required permission | `WATCHLIST_VIEW` |
| Success status | `200 OK` |

Scope rules:

- `PERSONAL`: return only items where `owner_user_id` equals the authenticated actor.
- `PORTFOLIO`: return only items whose portfolio passes data permission.
- If `portfolio_id` is supplied, the actor must have data permission for that portfolio's resolved data scope.
- `WATCHLIST_ADMIN` cross-owner personal behavior is not guaranteed in v1.

Query parameters:

| Name | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| `scope_type` | string | No | none | `PERSONAL` or `PORTFOLIO`. |
| `portfolio_id` | UUID string | No | none | Allowed only for `PORTFOLIO` queries. |
| `security_id` | UUID string | No | none | Must refer to canonical security when provided. |
| `include_disabled` | boolean | No | `false` | `true` or `false`. |
| `include_thresholds` | boolean | No | `true` | `true` or `false`. |
| `include_quote` | boolean | No | `true` | `true` or `false`. |
| `limit` | integer | No | `50` | Minimum `1`, maximum `200`. |
| `offset` | integer | No | `0` | Minimum `0`. |

Success response:

```json
{
  "data": {
    "items": [
      {
        "id": "0a7bb3e5-a258-4fb3-b2d9-e4b4f705b73d",
        "scope_type": "PERSONAL",
        "owner_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
        "created_by_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
        "created_by_user": {
          "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
          "display_name": "Alice Nakamura",
          "email": "alice.nakamura@example.com",
          "avatar_url": null
        },
        "portfolio_id": null,
        "portfolio": null,
        "security": {
          "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
          "ims_symbol": "AAPL.US",
          "display_symbol": "AAPL",
          "name": "Apple Inc.",
          "asset_type": "EQUITY",
          "currency": "USD",
          "exchange_mic": "XNAS"
        },
        "status": "ACTIVE",
        "pinned": true,
        "note": "Monitor upside breakout",
        "threshold_rules": [
          {
            "id": "485f4c94-28d8-4472-a960-a9c711c0dc01",
            "metric_type": "MARKET_PRICE",
            "direction": "ABOVE",
            "threshold_value": "190.00000000",
            "currency": "USD",
            "cooldown_minutes": 60,
            "status": "ENABLED",
            "last_state": "NON_BREACHED",
            "last_observed_price": "184.25000000",
            "last_observed_at": "2026-06-24T04:25:00Z",
            "last_evaluated_at": "2026-06-24T04:30:00Z",
            "last_state_changed_at": "2026-06-24T04:30:00Z",
            "last_alerted_at": null,
            "last_quote_stale": false,
            "last_stale_reason": null,
            "created_at": "2026-06-24T03:00:00Z",
            "updated_at": "2026-06-24T04:30:00Z"
          }
        ],
        "quote": {
          "symbol": "AAPL",
          "provider": "yahoo",
          "price": "184.25000000",
          "currency": "USD",
          "previous_close": "182.60000000",
          "change_percent": "0.90361446",
          "effective_at": "2026-06-24T04:25:00Z",
          "fetched_at": "2026-06-24T04:30:00Z",
          "market_status": "OPEN",
          "stale": false,
          "stale_reason": null
        },
        "created_at": "2026-06-24T03:00:00Z",
        "updated_at": "2026-06-24T04:30:00Z"
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "total": 1
    }
  }
}
```

Error examples:

```json
{
  "error": "invalid scope_type",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "scope_type",
    "allowed_values": ["PERSONAL", "PORTFOLIO"]
  }
}
```

```json
{
  "error": "portfolio scope is forbidden",
  "code": "WATCHLIST_FORBIDDEN_SCOPE",
  "details": {
    "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10"
  }
}
```

Backend notes:

- Apply function permission first.
- Apply personal owner and portfolio data-scope filters before returning rows.
- Do not call `/market-data/screen/watchlist`.
- `include_quote=true` should use the injected market-data quote provider or cached quote adapter selected for Watchlist.

Frontend notes:

- Treat all decimal fields as strings.
- `quote` may be null when `include_quote=false` or when no quote is available.
- `threshold_rules` is an array when `include_thresholds=true`; it is null when `include_thresholds=false`. An empty array means thresholds were included and no rules exist.

### B. Create Watchlist Item

| Field | Value |
| --- | --- |
| Purpose | Create a watchlist item and optional threshold rules. |
| Method and path | `POST /api/v1/watchlists/items` |
| Required permission | `WATCHLIST_MANAGE` |
| Success status | `201 Created` |

Scope rules:

- `PERSONAL`: server sets `owner_user_id` and `created_by_user_id` to the authenticated actor. Any client-supplied `owner_user_id` or `created_by_user_id` must be ignored.
- `PORTFOLIO`: `portfolio_id` is required and actor must have portfolio data permission.
- `PORTFOLIO`: server sets `created_by_user_id` to the authenticated actor. `owner_user_id` is personal-owner metadata and should be null in portfolio responses.
- v1 portfolio alert recipient is `created_by_user_id` only.

Request body:

```json
{
  "scope_type": "PERSONAL",
  "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
  "pinned": true,
  "note": "Monitor upside breakout",
  "threshold_rules": [
    {
      "metric_type": "MARKET_PRICE",
      "direction": "ABOVE",
      "threshold_value": "190.00000000",
      "currency": "USD",
      "cooldown_minutes": 60,
      "status": "ENABLED"
    }
  ]
}
```

Validation rules:

- `scope_type` is required.
- `security_id` is required and must reference an active canonical security.
- `portfolio_id` is required only when `scope_type = "PORTFOLIO"`.
- `portfolio_id` must be null or omitted when `scope_type = "PERSONAL"`.
- `threshold_rules` is optional.
- Every threshold rule must use `metric_type = "MARKET_PRICE"` when supplied.
- `threshold_value` must be a positive decimal string.
- `cooldown_minutes` defaults to `60` when omitted.
- `status` defaults to `ENABLED` when omitted.

Success response:

```json
{
  "data": {
    "id": "0a7bb3e5-a258-4fb3-b2d9-e4b4f705b73d",
    "scope_type": "PERSONAL",
    "owner_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "created_by_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "created_by_user": {
      "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
      "display_name": "Alice Nakamura",
      "email": "alice.nakamura@example.com",
      "avatar_url": null
    },
    "portfolio_id": null,
    "portfolio": null,
    "security": {
      "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
      "ims_symbol": "AAPL.US",
      "display_symbol": "AAPL",
      "name": "Apple Inc.",
      "asset_type": "EQUITY",
      "currency": "USD",
      "exchange_mic": "XNAS"
    },
    "status": "ACTIVE",
    "pinned": true,
    "note": "Monitor upside breakout",
    "threshold_rules": [
      {
        "id": "485f4c94-28d8-4472-a960-a9c711c0dc01",
        "metric_type": "MARKET_PRICE",
        "direction": "ABOVE",
        "threshold_value": "190.00000000",
        "currency": "USD",
        "cooldown_minutes": 60,
        "status": "ENABLED",
        "last_state": "UNKNOWN",
        "last_observed_price": null,
        "last_observed_at": null,
        "last_evaluated_at": null,
        "last_state_changed_at": null,
        "last_alerted_at": null,
        "last_quote_stale": false,
        "last_stale_reason": null,
        "created_at": "2026-06-24T04:00:00Z",
        "updated_at": "2026-06-24T04:00:00Z"
      }
    ],
    "quote": null,
    "created_at": "2026-06-24T04:00:00Z",
    "updated_at": "2026-06-24T04:00:00Z"
  }
}
```

Error examples:

```json
{
  "error": "watchlist item already exists for this scope and security",
  "code": "WATCHLIST_DUPLICATE_ITEM",
  "details": {
    "scope_type": "PERSONAL",
    "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10"
  }
}
```

```json
{
  "error": "invalid or inactive security",
  "code": "WATCHLIST_INVALID_SECURITY",
  "details": {
    "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10"
  }
}
```

Backend notes:

- Create item and rules in one transaction.
- Enforce uniqueness on active items only.
- Emit `WATCHLIST_ITEM_CREATED` and `WATCHLIST_THRESHOLD_CREATED` audit events.

### C. Update Watchlist Item

| Field | Value |
| --- | --- |
| Purpose | Update item metadata and threshold rules. |
| Method and path | `PATCH /api/v1/watchlists/items/{id}` |
| Required permission | `WATCHLIST_MANAGE` |
| Success status | `200 OK` |

Scope rules:

- Resolve item by ID first using a scope-safe query.
- `PERSONAL`: actor must own the item unless a future `WATCHLIST_ADMIN` policy allows otherwise.
- `PORTFOLIO`: actor must have portfolio data permission, even if the actor created the item.

Recommended v1 threshold rule behavior:

- If `threshold_rules` is omitted, existing rules are unchanged.
- If `threshold_rules` is present, it is the complete desired active rule set for the item.
- Rules with an `id` update the matching existing rule.
- Rules without an `id` create new rules.
- Existing active rules not included in `threshold_rules` are soft-disabled.
- Dedicated threshold-rule CRUD endpoints are future scope.

Request body:

```json
{
  "pinned": false,
  "note": "Use revised breakout level",
  "status": "ACTIVE",
  "threshold_rules": [
    {
      "id": "485f4c94-28d8-4472-a960-a9c711c0dc01",
      "metric_type": "MARKET_PRICE",
      "direction": "ABOVE",
      "threshold_value": "195.00000000",
      "currency": "USD",
      "cooldown_minutes": 60,
      "status": "ENABLED"
    }
  ]
}
```

Validation rules:

- At least one mutable field must be present.
- `status` must be `ACTIVE` or `DISABLED` when supplied.
- Rule IDs in the request must belong to the target item.
- Disabled item status stops evaluation of all child rules.
- Threshold rule validation matches create behavior.

Success response:

```json
{
  "data": {
    "id": "0a7bb3e5-a258-4fb3-b2d9-e4b4f705b73d",
    "scope_type": "PERSONAL",
    "owner_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "created_by_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "created_by_user": {
      "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
      "display_name": "Alice Nakamura",
      "email": "alice.nakamura@example.com",
      "avatar_url": null
    },
    "portfolio_id": null,
    "portfolio": null,
    "security": {
      "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
      "ims_symbol": "AAPL.US",
      "display_symbol": "AAPL",
      "name": "Apple Inc.",
      "asset_type": "EQUITY",
      "currency": "USD",
      "exchange_mic": "XNAS"
    },
    "status": "ACTIVE",
    "pinned": false,
    "note": "Use revised breakout level",
    "threshold_rules": [
      {
        "id": "485f4c94-28d8-4472-a960-a9c711c0dc01",
        "metric_type": "MARKET_PRICE",
        "direction": "ABOVE",
        "threshold_value": "195.00000000",
        "currency": "USD",
        "cooldown_minutes": 60,
        "status": "ENABLED",
        "last_state": "UNKNOWN",
        "last_observed_price": null,
        "last_observed_at": null,
        "last_evaluated_at": null,
        "last_state_changed_at": null,
        "last_alerted_at": null,
        "last_quote_stale": false,
        "last_stale_reason": null,
        "created_at": "2026-06-24T04:00:00Z",
        "updated_at": "2026-06-24T04:15:00Z"
      }
    ],
    "quote": null,
    "created_at": "2026-06-24T04:00:00Z",
    "updated_at": "2026-06-24T04:15:00Z"
  }
}
```

Error examples:

```json
{
  "error": "watchlist item not found",
  "code": "WATCHLIST_ITEM_NOT_FOUND",
  "details": {
    "id": "0a7bb3e5-a258-4fb3-b2d9-e4b4f705b73d"
  }
}
```

```json
{
  "error": "invalid threshold rule",
  "code": "WATCHLIST_INVALID_THRESHOLD",
  "details": {
    "field": "threshold_value",
    "reason": "must be greater than zero"
  }
}
```

Backend notes:

- Rule replacement should run in the same transaction as the item update.
- Reset a rule's `last_state` to `UNKNOWN` when `direction`, `threshold_value`, `currency`, or `metric_type` changes.
- Emit `WATCHLIST_ITEM_UPDATED` and `WATCHLIST_THRESHOLD_UPDATED` as applicable.

### D. Delete Watchlist Item

| Field | Value |
| --- | --- |
| Purpose | Soft-delete a watchlist item. |
| Method and path | `DELETE /api/v1/watchlists/items/{id}` |
| Required permission | `WATCHLIST_MANAGE` |
| Success status | `204 No Content` |

Scope rules:

- `PERSONAL`: actor must own the item.
- `PORTFOLIO`: actor must have portfolio data permission.

Success response:

```text
204 No Content
```

Rules stop evaluating immediately after the item is soft-deleted.

Error examples:

```json
{
  "error": "watchlist item not found",
  "code": "WATCHLIST_ITEM_NOT_FOUND",
  "details": {
    "id": "0a7bb3e5-a258-4fb3-b2d9-e4b4f705b73d"
  }
}
```

```json
{
  "error": "personal watchlist access is forbidden",
  "code": "WATCHLIST_FORBIDDEN_SCOPE",
  "details": {
    "scope_type": "PERSONAL"
  }
}
```

Backend notes:

- Set `deleted_at`, `deleted_by`, `updated_at`, and `status = "DISABLED"`.
- Child rules must be excluded from future evaluator queries.
- Emit `WATCHLIST_ITEM_DELETED`.

### E. List Alert Events

| Field | Value |
| --- | --- |
| Purpose | List visible alert events. |
| Method and path | `GET /api/v1/watchlists/alerts` |
| Required permission | `WATCHLIST_VIEW` |
| Success status | `200 OK` |

Scope rules:

- Personal alerts are visible only to the personal owner.
- Portfolio alerts require portfolio data permission.
- Admin/risk/audit access is controlled by future `WATCHLIST_ADMIN` policy.

Query parameters:

| Name | Type | Required | Default | Validation |
| --- | --- | --- | --- | --- |
| `scope_type` | string | No | none | `PERSONAL` or `PORTFOLIO`. |
| `portfolio_id` | UUID string | No | none | Requires portfolio data permission. |
| `security_id` | UUID string | No | none | Canonical security UUID. |
| `rule_id` | UUID string | No | none | Threshold rule UUID. |
| `acknowledged` | boolean | No | none | `true` returns acknowledged only; `false` returns unacknowledged only. |
| `created_from` | RFC3339 timestamp | No | none | Inclusive lower bound. |
| `created_to` | RFC3339 timestamp | No | none | Inclusive upper bound. |
| `limit` | integer | No | `50` | Minimum `1`, maximum `200`. |
| `offset` | integer | No | `0` | Minimum `0`. |

Success response:

```json
{
  "data": {
    "items": [
      {
        "id": "6f7c2c3a-81f5-4746-94aa-bd373d589013",
        "watchlist_item_id": "0a7bb3e5-a258-4fb3-b2d9-e4b4f705b73d",
        "threshold_rule_id": "485f4c94-28d8-4472-a960-a9c711c0dc01",
        "scope_type": "PERSONAL",
        "owner_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
        "created_by_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
        "created_by_user": {
          "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
          "display_name": "Alice Nakamura",
          "email": "alice.nakamura@example.com",
          "avatar_url": null
        },
        "portfolio_id": null,
        "portfolio": null,
        "security": {
          "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
          "ims_symbol": "AAPL.US",
          "display_symbol": "AAPL",
          "name": "Apple Inc.",
          "asset_type": "EQUITY",
          "currency": "USD",
          "exchange_mic": "XNAS"
        },
        "direction": "ABOVE",
        "previous_state": "NON_BREACHED",
        "current_state": "BREACHED",
        "observed_price": "195.50000000",
        "threshold_value": "195.00000000",
        "currency": "USD",
        "quote_provider": "yahoo",
        "observed_at": "2026-06-24T04:35:00Z",
        "evaluated_at": "2026-06-24T04:36:00Z",
        "stale": false,
        "stale_reason": null,
        "notification_status": "CREATED",
        "acknowledgement_state": "UNACKNOWLEDGED",
        "acknowledged_by": null,
        "acknowledged_by_user": null,
        "acknowledged_at": null,
        "acknowledgement_note": null,
        "created_at": "2026-06-24T04:36:00Z"
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "total": 1
    }
  }
}
```

Error examples:

```json
{
  "error": "invalid created_from timestamp",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "created_from",
    "expected_format": "RFC3339"
  }
}
```

```json
{
  "error": "portfolio scope is forbidden",
  "code": "WATCHLIST_FORBIDDEN_SCOPE",
  "details": {
    "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10"
  }
}
```

Backend notes:

- Scope-safe query filtering must be applied before joining alert detail.
- Alert events are immutable except acknowledgement fields.

### F. Acknowledge Alert

| Field | Value |
| --- | --- |
| Purpose | Acknowledge an alert event. |
| Method and path | `POST /api/v1/watchlists/alerts/{id}/acknowledge` |
| Required permission | `WATCHLIST_ALERT_ACK` |
| Success status | `200 OK` |

Scope rules:

- Actor must be able to view the alert using the same personal or portfolio rules used by alert listing.
- Admin acknowledgement across scopes requires a future explicit policy.

Request body:

```json
{
  "note": "Reviewed by portfolio manager"
}
```

Validation rules:

- `note` is optional.
- `note` should be trimmed.
- `note` maximum length should be 1000 characters.
- Already acknowledged alerts return `409`.

Success response:

```json
{
  "data": {
    "id": "6f7c2c3a-81f5-4746-94aa-bd373d589013",
    "watchlist_item_id": "0a7bb3e5-a258-4fb3-b2d9-e4b4f705b73d",
    "threshold_rule_id": "485f4c94-28d8-4472-a960-a9c711c0dc01",
    "scope_type": "PERSONAL",
    "owner_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "created_by_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "created_by_user": {
      "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
      "display_name": "Alice Nakamura",
      "email": "alice.nakamura@example.com",
      "avatar_url": null
    },
    "portfolio_id": null,
    "portfolio": null,
    "security": {
      "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
      "ims_symbol": "AAPL.US",
      "display_symbol": "AAPL",
      "name": "Apple Inc.",
      "asset_type": "EQUITY",
      "currency": "USD",
      "exchange_mic": "XNAS"
    },
    "direction": "ABOVE",
    "previous_state": "NON_BREACHED",
    "current_state": "BREACHED",
    "observed_price": "195.50000000",
    "threshold_value": "195.00000000",
    "currency": "USD",
    "quote_provider": "yahoo",
    "observed_at": "2026-06-24T04:35:00Z",
    "evaluated_at": "2026-06-24T04:36:00Z",
    "stale": false,
    "stale_reason": null,
    "notification_status": "CREATED",
    "acknowledgement_state": "ACKNOWLEDGED",
    "acknowledged_by": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "acknowledged_by_user": {
      "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
      "display_name": "Alice Nakamura",
      "email": "alice.nakamura@example.com",
      "avatar_url": null
    },
    "acknowledged_at": "2026-06-24T04:45:00Z",
    "acknowledgement_note": "Reviewed by portfolio manager",
    "created_at": "2026-06-24T04:36:00Z"
  }
}
```

Error examples:

```json
{
  "error": "alert event not found",
  "code": "WATCHLIST_ALERT_NOT_FOUND",
  "details": {
    "id": "6f7c2c3a-81f5-4746-94aa-bd373d589013"
  }
}
```

```json
{
  "error": "alert is already acknowledged",
  "code": "WATCHLIST_ALERT_ALREADY_ACKNOWLEDGED",
  "details": {
    "id": "6f7c2c3a-81f5-4746-94aa-bd373d589013",
    "acknowledged_at": "2026-06-24T04:45:00Z"
  }
}
```

Backend notes:

- Acknowledgement update should be atomic and conflict-safe.
- Emit `WATCHLIST_ALERT_ACKNOWLEDGED`.

### G. Manual Evaluation

| Field | Value |
| --- | --- |
| Purpose | Manually run operational evaluation for enabled rules. This is not the normal scheduler path. |
| Method and path | `POST /api/v1/watchlists/evaluate` |
| Required permission | `WATCHLIST_EVALUATE` |
| Success status | `200 OK` |

Scope rules:

- Normal users must not receive `WATCHLIST_EVALUATE`.
- Request may target one rule, one item, one security, one portfolio, one scope, or all visible operational scope.
- If `portfolio_id` is supplied, require portfolio data permission unless a future operational global data policy is approved.

Request body:

```json
{
  "scope_type": "PERSONAL",
  "portfolio_id": null,
  "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
  "item_id": null,
  "rule_id": null,
  "dry_run": true
}
```

Validation rules:

- All filter IDs must be valid UUIDs when supplied.
- `scope_type` must be `PERSONAL` or `PORTFOLIO` when supplied.
- `portfolio_id` should only be supplied for portfolio-scoped evaluation.
- `dry_run` defaults to `false`.
- If `rule_id` targets a disabled rule, return `409`.

Success response:

```json
{
  "data": {
    "dry_run": true,
    "rules_evaluated": 2,
    "alerts_created": 0,
    "alerts_suppressed": 1,
    "rules_skipped": 1,
    "provider_failures": 0,
    "results": [
      {
        "rule_id": "485f4c94-28d8-4472-a960-a9c711c0dc01",
        "watchlist_item_id": "0a7bb3e5-a258-4fb3-b2d9-e4b4f705b73d",
        "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
        "previous_state": "NON_BREACHED",
        "computed_state": "BREACHED",
        "would_create_alert": true,
        "notification_status": "SKIPPED",
        "quote_status": "LIVE",
        "observed_price": "195.50000000",
        "threshold_value": "195.00000000",
        "stale": false,
        "stale_reason": null
      },
      {
        "rule_id": "9dcc1fc5-4a61-4087-8941-e8c6c3cfdf02",
        "watchlist_item_id": "8b6b06e4-1051-44a2-9775-1a94e16a87cb",
        "security_id": "44bfb19e-9299-4381-8932-c2c7e4513ac1",
        "previous_state": "BREACHED",
        "computed_state": "BREACHED",
        "would_create_alert": false,
        "notification_status": "SKIPPED",
        "quote_status": "STALE_SKIPPED",
        "observed_price": null,
        "threshold_value": "70.00000000",
        "stale": true,
        "stale_reason": "cached quote is older than configured max age"
      }
    ]
  }
}
```

Error examples:

```json
{
  "error": "rule is disabled",
  "code": "WATCHLIST_RULE_DISABLED",
  "details": {
    "rule_id": "485f4c94-28d8-4472-a960-a9c711c0dc01"
  }
}
```

```json
{
  "error": "market data provider unavailable",
  "code": "WATCHLIST_PROVIDER_UNAVAILABLE",
  "details": {
    "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10"
  }
}
```

Backend notes:

- Manual evaluation must use the same domain evaluator as the scheduled worker.
- Provider failure for a batch should increment `provider_failures` and continue other rules.
- Provider failure for a single targeted rule may return `503`.
- Dry run must not persist alert events, rule state changes, notification calls, or audit alert-triggered events.

## Request and Response Schemas

### `WatchlistItem`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `id` | string UUID | Yes | No | none | Watchlist item ID. |
| `scope_type` | string | Yes | No | `PERSONAL`, `PORTFOLIO` | Defines ownership model. |
| `owner_user_id` | string UUID | Yes | Yes | none | Required and non-null for `PERSONAL`; null for `PORTFOLIO` because portfolio access is governed by data permission, not personal ownership. |
| `created_by_user_id` | string UUID | Yes | No | none | Required for both `PERSONAL` and `PORTFOLIO`. For `PORTFOLIO` v1 alerts, this is the notification recipient. |
| `created_by_user` | `UserDescriptor` | Yes | Yes | none | Nullable display descriptor for creator. Frontend must not render `created_by_user_id` directly. |
| `portfolio_id` | string UUID | Yes | Yes | none | Required for `PORTFOLIO`; null for `PERSONAL`. |
| `portfolio` | `PortfolioDescriptor` | Yes | Yes | none | Null for `PERSONAL`; should be populated for `PORTFOLIO` when possible. Frontend should show `portfolio.display_name`, not `portfolio_id`. |
| `security` | `SecurityDescriptor` | Yes | No | none | Canonical security snapshot. |
| `status` | string | Yes | No | `ACTIVE`, `DISABLED` | Disabled items do not evaluate. |
| `pinned` | boolean | Yes | No | none | Defaults to false. |
| `note` | string | Yes | Yes | none | Maximum 2000 characters recommended. |
| `threshold_rules` | array of `ThresholdRule` | Yes | Yes | none | Array when `include_thresholds=true`; empty array means included and no rules exist. Null when `include_thresholds=false`. |
| `quote` | `QuoteSnapshot` | Yes | Yes | none | Null when `include_quote=false` or no quote is available. |
| `created_at` | RFC3339 string | Yes | No | none | UTC. |
| `updated_at` | RFC3339 string | Yes | No | none | UTC. |

### `WatchlistItemCreateRequest`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `scope_type` | string | Yes | No | `PERSONAL`, `PORTFOLIO` | Required. |
| `portfolio_id` | string UUID | Conditional | Yes | none | Required for `PORTFOLIO`; omitted or null for `PERSONAL`. |
| `security_id` | string UUID | Yes | No | none | Must reference active canonical security. |
| `pinned` | boolean | No | No | none | Defaults to false. |
| `note` | string | No | Yes | none | Maximum 2000 characters recommended. |
| `threshold_rules` | array of `ThresholdRuleRequest` | No | No | none | Defaults to empty array. |

### `WatchlistItemUpdateRequest`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `pinned` | boolean | No | No | none | Updates item pinned state. |
| `note` | string | No | Yes | none | Null clears note. Maximum 2000 characters recommended. |
| `status` | string | No | No | `ACTIVE`, `DISABLED` | Disabled items stop evaluation. |
| `threshold_rules` | array of `ThresholdRuleRequest` | No | No | none | When present, complete desired active rule set for v1. |

### `ThresholdRule`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `id` | string UUID | Yes | No | none | Rule ID. |
| `metric_type` | string | Yes | No | `MARKET_PRICE` | Only supported v1 metric. |
| `direction` | string | Yes | No | `ABOVE`, `BELOW` | Breach predicate direction. |
| `threshold_value` | decimal string | Yes | No | none | Positive decimal. |
| `currency` | string | Yes | Yes | ISO 4217 | If set, evaluator should require matching quote currency. |
| `cooldown_minutes` | integer | Yes | No | none | Defaults to 60. Must be `>= 0`; recommended minimum is 1 unless product approves zero. |
| `status` | string | Yes | No | `ENABLED`, `DISABLED` | Disabled rules are skipped. |
| `last_state` | string | Yes | No | `UNKNOWN`, `NON_BREACHED`, `BREACHED` | Crossing state. |
| `last_observed_price` | decimal string | Yes | Yes | none | Last evaluated price. |
| `last_observed_at` | RFC3339 string | Yes | Yes | none | Quote effective time. |
| `last_evaluated_at` | RFC3339 string | Yes | Yes | none | Backend evaluation time. |
| `last_state_changed_at` | RFC3339 string | Yes | Yes | none | Last state transition time. |
| `last_alerted_at` | RFC3339 string | Yes | Yes | none | Last alert event time. |
| `last_quote_stale` | boolean | Yes | No | none | True when last accepted quote was stale. |
| `last_stale_reason` | string | Yes | Yes | none | Displayable reason from market-data fallback. |
| `created_at` | RFC3339 string | Yes | No | none | UTC. |
| `updated_at` | RFC3339 string | Yes | No | none | UTC. |

### `ThresholdRuleRequest`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `id` | string UUID | No | No | none | Required only when updating an existing rule through `PATCH`. |
| `metric_type` | string | No | No | `MARKET_PRICE` | Defaults to `MARKET_PRICE`; other values rejected. |
| `direction` | string | Yes | No | `ABOVE`, `BELOW` | Required for creates and replacement entries. |
| `threshold_value` | decimal string | Yes | No | none | Must be greater than zero. |
| `currency` | string | No | Yes | ISO 4217 | Optional expected quote currency. |
| `cooldown_minutes` | integer | No | No | none | Defaults to 60. |
| `status` | string | No | No | `ENABLED`, `DISABLED` | Defaults to `ENABLED`. |

### `AlertEvent`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `id` | string UUID | Yes | No | none | Alert event ID. |
| `watchlist_item_id` | string UUID | Yes | No | none | Source item ID. |
| `threshold_rule_id` | string UUID | Yes | No | none | Source rule ID. |
| `scope_type` | string | Yes | No | `PERSONAL`, `PORTFOLIO` | Snapshot from item. |
| `owner_user_id` | string UUID | Yes | Yes | none | Snapshot from item. Required and non-null for `PERSONAL`; null for `PORTFOLIO`. |
| `created_by_user_id` | string UUID | Yes | No | none | Snapshot from item creator. For `PORTFOLIO` v1 alerts, this is the notification recipient. |
| `created_by_user` | `UserDescriptor` | Yes | Yes | none | Nullable display descriptor for creator. Frontend must not render `created_by_user_id` directly. |
| `portfolio_id` | string UUID | Yes | Yes | none | Snapshot from item. |
| `portfolio` | `PortfolioDescriptor` | Yes | Yes | none | Null for personal alerts; should be populated for portfolio alerts when possible. |
| `security` | `SecurityDescriptor` | Yes | No | none | Canonical security snapshot. |
| `direction` | string | Yes | No | `ABOVE`, `BELOW` | Crossing direction. |
| `previous_state` | string | Yes | No | `NON_BREACHED` | Alerting transition starts from non-breached. |
| `current_state` | string | Yes | No | `BREACHED` | Alert event always represents breached state. |
| `observed_price` | decimal string | Yes | No | none | Price used for alert decision. |
| `threshold_value` | decimal string | Yes | No | none | Rule threshold snapshot. |
| `currency` | string | Yes | Yes | ISO 4217 | Quote currency. |
| `quote_provider` | string | Yes | Yes | none | Provider tag from market-data. |
| `observed_at` | RFC3339 string | Yes | No | none | Quote effective time. |
| `evaluated_at` | RFC3339 string | Yes | No | none | Evaluation time. |
| `stale` | boolean | Yes | No | none | True when alert used stale fallback within max age. |
| `stale_reason` | string | Yes | Yes | none | Required when `stale=true` if market-data provides a reason. |
| `notification_status` | string | Yes | No | `PENDING`, `CREATED`, `SUPPRESSED`, `FAILED`, `SKIPPED` | Notification outcome. `SUPPRESSED` is allowed only when Watchlist can positively detect suppression. |
| `acknowledgement_state` | string | Yes | No | `UNACKNOWLEDGED`, `ACKNOWLEDGED` | Derived from acknowledgement fields. |
| `acknowledged_by` | string UUID | Yes | Yes | none | Actor who acknowledged. |
| `acknowledged_by_user` | `UserDescriptor` | Yes | Yes | none | Nullable display descriptor for acknowledger. Frontend must not render `acknowledged_by` directly. |
| `acknowledged_at` | RFC3339 string | Yes | Yes | none | Acknowledgement time. |
| `acknowledgement_note` | string | Yes | Yes | none | Optional acknowledgement note. |
| `created_at` | RFC3339 string | Yes | No | none | UTC. |

### `AcknowledgeAlertRequest`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `note` | string | No | Yes | none | Maximum 1000 characters recommended. |

### `ManualEvaluateRequest`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `scope_type` | string | No | Yes | `PERSONAL`, `PORTFOLIO` | Optional filter. |
| `portfolio_id` | string UUID | No | Yes | none | Optional portfolio filter. Requires data permission. |
| `security_id` | string UUID | No | Yes | none | Optional canonical security filter. |
| `item_id` | string UUID | No | Yes | none | Optional item filter. |
| `rule_id` | string UUID | No | Yes | none | Optional rule filter. |
| `dry_run` | boolean | No | No | none | Defaults to false. |

### `ManualEvaluateResponse`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `dry_run` | boolean | Yes | No | none | Echoes request mode. |
| `rules_evaluated` | integer | Yes | No | none | Count of rules evaluated. |
| `alerts_created` | integer | Yes | No | none | Count of persisted alert events. |
| `alerts_suppressed` | integer | Yes | No | none | Count of cooldown or alert-event idempotency suppressions detected by Watchlist. |
| `rules_skipped` | integer | Yes | No | none | Count of disabled, deleted, stale-skipped, or ineligible rules. |
| `provider_failures` | integer | Yes | No | none | Count of provider failures without usable quote. |
| `results` | array | Yes | No | none | May be empty for non-dry-run responses. |

### `SecurityDescriptor`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `security_id` | string UUID | Yes | No | none | Canonical security ID from reference data. |
| `ims_symbol` | string | Yes | No | none | Canonical IMS symbol. |
| `display_symbol` | string | Yes | No | none | User-facing market symbol. |
| `name` | string | Yes | No | none | Security name. |
| `asset_type` | string | Yes | No | backend reference enum | From canonical security. |
| `currency` | string | Yes | Yes | ISO 4217 | Trading currency when known. |
| `exchange_mic` | string | Yes | Yes | ISO 10383 MIC | Exchange MIC when known. |

### `PortfolioDescriptor`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `portfolio_id` | string UUID | Yes | No | none | Transport identifier for state, routing, mutations, and API calls. Do not render as a label. |
| `portfolio_code` | string | Yes | Yes | none | User-facing when available. |
| `portfolio_name` | string | Yes | Yes | none | User-facing when available. |
| `display_name` | string | Yes | No | none | Required safe UI label. Frontend should show this instead of `portfolio_id`. |
| `fund_id` | string UUID | No | Yes | none | Optional transport identifier for the parent fund. Do not render as a label. |
| `fund_code` | string | No | Yes | none | Optional user-facing fund code. |
| `fund_name` | string | No | Yes | none | Optional user-facing fund name. |

### `UserDescriptor`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `user_id` | string UUID | Yes | No | none | Transport identifier. Do not render as a label. |
| `display_name` | string | Yes | No | none | Required safe UI label. |
| `email` | string | No | Yes | none | Optional user-facing/support context. |
| `avatar_url` | string | No | Yes | none | Optional image URL. |

### `QuoteSnapshot`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `symbol` | string | Yes | No | none | Provider input symbol. |
| `provider` | string | Yes | No | none | Provider that returned the quote. |
| `price` | decimal string | Yes | No | none | Market quote price. |
| `currency` | string | Yes | No | ISO 4217 | Quote currency. |
| `previous_close` | decimal string | Yes | Yes | none | Null when unavailable. |
| `change_percent` | decimal string | Yes | Yes | none | Signed percentage points when available. |
| `effective_at` | RFC3339 string | Yes | No | none | Provider observation time. |
| `fetched_at` | RFC3339 string | Yes | No | none | Platform fetch time. |
| `market_status` | string | Yes | Yes | `OPEN`, `CLOSED`, `UNKNOWN` | Null when provider does not supply status. |
| `stale` | boolean | Yes | No | none | Mirrors `MarketQuote.Stale`. |
| `stale_reason` | string | Yes | Yes | none | Mirrors `MarketQuote.StaleReason`. |

### `PaginationMeta`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `limit` | integer | Yes | No | none | Effective limit. |
| `offset` | integer | Yes | No | none | Effective offset. |
| `total` | integer | Yes | No | none | Total rows matching scope-safe filters. |

### `ErrorResponse`

| Field | Type | Required | Nullable | Enum | Validation and notes |
| --- | --- | --- | --- | --- | --- |
| `error` | string | Yes | No | none | Human-readable safe error message. |
| `code` | string | Recommended | No | error catalog | Stable machine-readable Watchlist/API error code. |
| `details` | object | No | Yes | none | Field-level or resource-level safe details. |

## Error Catalog

| Code | HTTP status | Meaning |
| --- | --- | --- |
| `WATCHLIST_INVALID_SECURITY` | `422` | Security is missing, inactive, or not canonical. |
| `WATCHLIST_DUPLICATE_ITEM` | `409` | Active item already exists for same scope and security. |
| `WATCHLIST_INVALID_THRESHOLD` | `400` | Threshold request is invalid. |
| `WATCHLIST_FORBIDDEN_SCOPE` | `403` | Actor cannot access personal owner or portfolio data scope. |
| `WATCHLIST_ITEM_NOT_FOUND` | `404` | Item does not exist, is deleted, or is hidden by scope-safe lookup. |
| `WATCHLIST_ALERT_NOT_FOUND` | `404` | Alert does not exist or is hidden by scope-safe lookup. |
| `WATCHLIST_STALE_QUOTE` | `409` | Quote is stale beyond allowed max age for targeted evaluation. |
| `WATCHLIST_PROVIDER_UNAVAILABLE` | `503` | No live or accepted stale quote is available for targeted evaluation. |
| `WATCHLIST_ALERT_ALREADY_ACKNOWLEDGED` | `409` | Alert acknowledgement already exists. |
| `WATCHLIST_RULE_DISABLED` | `409` | Targeted rule is disabled or ineligible for evaluation. |
| `VALIDATION_ERROR` | `400` | Malformed JSON, invalid query, invalid UUID, invalid timestamp, or missing required field. |
| `UNAUTHORIZED` | `401` | Authentication is missing or invalid. |
| `FORBIDDEN` | `403` | Function permission is missing. |
| `INTERNAL_ERROR` | `500` | Unexpected backend error. |

Error envelope example:

```json
{
  "error": "invalid threshold rule",
  "code": "WATCHLIST_INVALID_THRESHOLD",
  "details": {
    "field": "direction",
    "allowed_values": ["ABOVE", "BELOW"]
  }
}
```

## Alert Semantics

Predicate rules:

- `ABOVE` uses `observed_price >= threshold_value`.
- `BELOW` uses `observed_price <= threshold_value`.

State transition contract:

| Previous state | Computed state | Alert behavior |
| --- | --- | --- |
| `UNKNOWN` | `NON_BREACHED` | Seed state. No notification. |
| `UNKNOWN` | `BREACHED` | Seed breached state. No notification. |
| `NON_BREACHED` | `NON_BREACHED` | No notification. |
| `NON_BREACHED` | `BREACHED` | Crossing. Create alert if cooldown allows. |
| `BREACHED` | `BREACHED` | No repeated notification. |
| `BREACHED` | `NON_BREACHED` | Re-arm. No notification. |

Cooldown contract:

- Default cooldown is 60 minutes.
- Crossing inside cooldown is suppressed.
- Suppressed crossing should update rule state to `BREACHED` but should not create duplicate user notifications.
- Cooldown expiration alone does not create another alert. The rule must re-arm by returning to `NON_BREACHED`, then cross again.

Stale quote contract:

- Configurable stale quote max age defaults to 15 minutes.
- Stale quotes within max age may create alerts and must include `stale = true` and `stale_reason`.
- Stale quotes older than max age must be skipped and must not create alert events.
- Provider failure without a usable live or accepted stale quote creates no alert.
- In batch evaluation, provider failures are counted and evaluation continues for other rules.

Notification contract:

- Use notification event type `ALERT_THRESHOLD_BREACHED`.
- `PERSONAL` alert recipient is `owner_user_id`.
- `PORTFOLIO` v1 alert recipient is `created_by_user_id`.
- Portfolio recipient groups are future scope.

## Idempotency

Alert idempotency key format:

```text
watchlist:rule:<rule_id>:crossing:<direction>:bucket:<timestamp>
```

Rules:

- `rule_id` is the threshold rule UUID.
- `direction` is `ABOVE` or `BELOW`.
- `timestamp` is the UTC bucket start. The bucket should align to cooldown window or evaluator schedule bucket.
- `watchlist_alert_events.idempotency_key` must be unique.
- Public user-facing AlertEvent responses must not expose `idempotency_key`. If operational/debug APIs need it later, document it as admin/debug-only.
- The same key must be passed to the notification module.
- Duplicate alert event insertion or duplicate notification creation should be treated as success, not as a business failure.
- The current notification service returns `nil` for idempotent duplicates and does not expose created-vs-suppressed status to callers.
- `notification_status = "SUPPRESSED"` may be used only when the Watchlist implementation can positively detect suppression, such as from its own alert-event idempotency insert result or from a future structured notification adapter result.
- If API-visible notification suppression is required, the future Watchlist notification adapter must return a structured result such as `created`, `suppressed`, or `failed` instead of only `error`.

## API Examples

### Personal Item With One `ABOVE` Threshold

```json
{
  "scope_type": "PERSONAL",
  "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
  "pinned": true,
  "note": "Monitor upside breakout",
  "threshold_rules": [
    {
      "metric_type": "MARKET_PRICE",
      "direction": "ABOVE",
      "threshold_value": "190.00000000",
      "currency": "USD",
      "cooldown_minutes": 60,
      "status": "ENABLED"
    }
  ]
}
```

### Portfolio Item With One `BELOW` Threshold

```json
{
  "scope_type": "PORTFOLIO",
  "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10",
  "security_id": "44bfb19e-9299-4381-8932-c2c7e4513ac1",
  "pinned": false,
  "note": "Downside risk monitor",
  "threshold_rules": [
    {
      "metric_type": "MARKET_PRICE",
      "direction": "BELOW",
      "threshold_value": "70.00000000",
      "currency": "USD",
      "cooldown_minutes": 60,
      "status": "ENABLED"
    }
  ]
}
```

### List Response With Quote Snapshot

```json
{
  "data": {
    "items": [
      {
        "id": "8b6b06e4-1051-44a2-9775-1a94e16a87cb",
        "scope_type": "PORTFOLIO",
        "owner_user_id": null,
        "created_by_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
        "created_by_user": {
          "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
          "display_name": "Alice Nakamura",
          "email": "alice.nakamura@example.com",
          "avatar_url": null
        },
        "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10",
        "portfolio": {
          "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10",
          "portfolio_code": "GROWTH-01",
          "portfolio_name": "Thailand Growth Portfolio",
          "display_name": "GROWTH-01 - Thailand Growth Portfolio",
          "fund_id": "52a67f39-92b2-4b57-9b94-b78523ef2fd4",
          "fund_code": "TH-GROWTH",
          "fund_name": "Thailand Growth Fund"
        },
        "security": {
          "security_id": "44bfb19e-9299-4381-8932-c2c7e4513ac1",
          "ims_symbol": "TSLA.US",
          "display_symbol": "TSLA",
          "name": "Tesla Inc.",
          "asset_type": "EQUITY",
          "currency": "USD",
          "exchange_mic": "XNAS"
        },
        "status": "ACTIVE",
        "pinned": false,
        "note": "Downside risk monitor",
        "threshold_rules": [
          {
            "id": "9dcc1fc5-4a61-4087-8941-e8c6c3cfdf02",
            "metric_type": "MARKET_PRICE",
            "direction": "BELOW",
            "threshold_value": "70.00000000",
            "currency": "USD",
            "cooldown_minutes": 60,
            "status": "ENABLED",
            "last_state": "BREACHED",
            "last_observed_price": "68.50000000",
            "last_observed_at": "2026-06-24T04:18:00Z",
            "last_evaluated_at": "2026-06-24T04:20:00Z",
            "last_state_changed_at": "2026-06-24T04:20:00Z",
            "last_alerted_at": "2026-06-24T04:20:00Z",
            "last_quote_stale": true,
            "last_stale_reason": "provider unavailable; using cached quote",
            "created_at": "2026-06-24T03:50:00Z",
            "updated_at": "2026-06-24T04:20:00Z"
          }
        ],
        "quote": {
          "symbol": "TSLA",
          "provider": "yahoo",
          "price": "68.50000000",
          "currency": "USD",
          "previous_close": "72.10000000",
          "change_percent": "-4.99306519",
          "effective_at": "2026-06-24T04:18:00Z",
          "fetched_at": "2026-06-24T04:20:00Z",
          "market_status": "OPEN",
          "stale": true,
          "stale_reason": "provider unavailable; using cached quote"
        },
        "created_at": "2026-06-24T03:50:00Z",
        "updated_at": "2026-06-24T04:20:00Z"
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "total": 1
    }
  }
}
```

### Alert Event Response

```json
{
  "data": {
    "id": "f505f069-9b7c-478d-9488-3de57fbf602f",
    "watchlist_item_id": "8b6b06e4-1051-44a2-9775-1a94e16a87cb",
    "threshold_rule_id": "9dcc1fc5-4a61-4087-8941-e8c6c3cfdf02",
    "scope_type": "PORTFOLIO",
    "owner_user_id": null,
    "created_by_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "created_by_user": {
      "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
      "display_name": "Alice Nakamura",
      "email": "alice.nakamura@example.com",
      "avatar_url": null
    },
    "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10",
    "portfolio": {
      "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10",
      "portfolio_code": "GROWTH-01",
      "portfolio_name": "Thailand Growth Portfolio",
      "display_name": "GROWTH-01 - Thailand Growth Portfolio",
      "fund_id": "52a67f39-92b2-4b57-9b94-b78523ef2fd4",
      "fund_code": "TH-GROWTH",
      "fund_name": "Thailand Growth Fund"
    },
    "security": {
      "security_id": "44bfb19e-9299-4381-8932-c2c7e4513ac1",
      "ims_symbol": "TSLA.US",
      "display_symbol": "TSLA",
      "name": "Tesla Inc.",
      "asset_type": "EQUITY",
      "currency": "USD",
      "exchange_mic": "XNAS"
    },
    "direction": "BELOW",
    "previous_state": "NON_BREACHED",
    "current_state": "BREACHED",
    "observed_price": "68.50000000",
    "threshold_value": "70.00000000",
    "currency": "USD",
    "quote_provider": "yahoo",
    "observed_at": "2026-06-24T04:18:00Z",
    "evaluated_at": "2026-06-24T04:20:00Z",
    "stale": true,
    "stale_reason": "provider unavailable; using cached quote",
    "notification_status": "CREATED",
    "acknowledgement_state": "UNACKNOWLEDGED",
    "acknowledged_by": null,
    "acknowledged_by_user": null,
    "acknowledged_at": null,
    "acknowledgement_note": null,
    "created_at": "2026-06-24T04:20:00Z"
  }
}
```

### Acknowledgement Response

```json
{
  "data": {
    "id": "f505f069-9b7c-478d-9488-3de57fbf602f",
    "watchlist_item_id": "8b6b06e4-1051-44a2-9775-1a94e16a87cb",
    "threshold_rule_id": "9dcc1fc5-4a61-4087-8941-e8c6c3cfdf02",
    "scope_type": "PORTFOLIO",
    "owner_user_id": null,
    "created_by_user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "created_by_user": {
      "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
      "display_name": "Alice Nakamura",
      "email": "alice.nakamura@example.com",
      "avatar_url": null
    },
    "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10",
    "portfolio": {
      "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10",
      "portfolio_code": "GROWTH-01",
      "portfolio_name": "Thailand Growth Portfolio",
      "display_name": "GROWTH-01 - Thailand Growth Portfolio",
      "fund_id": "52a67f39-92b2-4b57-9b94-b78523ef2fd4",
      "fund_code": "TH-GROWTH",
      "fund_name": "Thailand Growth Fund"
    },
    "security": {
      "security_id": "44bfb19e-9299-4381-8932-c2c7e4513ac1",
      "ims_symbol": "TSLA.US",
      "display_symbol": "TSLA",
      "name": "Tesla Inc.",
      "asset_type": "EQUITY",
      "currency": "USD",
      "exchange_mic": "XNAS"
    },
    "direction": "BELOW",
    "previous_state": "NON_BREACHED",
    "current_state": "BREACHED",
    "observed_price": "68.50000000",
    "threshold_value": "70.00000000",
    "currency": "USD",
    "quote_provider": "yahoo",
    "observed_at": "2026-06-24T04:18:00Z",
    "evaluated_at": "2026-06-24T04:20:00Z",
    "stale": true,
    "stale_reason": "provider unavailable; using cached quote",
    "notification_status": "CREATED",
    "acknowledgement_state": "ACKNOWLEDGED",
    "acknowledged_by": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
    "acknowledged_by_user": {
      "user_id": "7a5c4138-4da1-4aa7-8cb9-bf4d9a28a001",
      "display_name": "Alice Nakamura",
      "email": "alice.nakamura@example.com",
      "avatar_url": null
    },
    "acknowledged_at": "2026-06-24T04:50:00Z",
    "acknowledgement_note": "Reviewed by portfolio manager",
    "created_at": "2026-06-24T04:20:00Z"
  }
}
```

### Manual Dry-Run Evaluation Response

```json
{
  "data": {
    "dry_run": true,
    "rules_evaluated": 1,
    "alerts_created": 0,
    "alerts_suppressed": 0,
    "rules_skipped": 0,
    "provider_failures": 0,
    "results": [
      {
        "rule_id": "485f4c94-28d8-4472-a960-a9c711c0dc01",
        "watchlist_item_id": "0a7bb3e5-a258-4fb3-b2d9-e4b4f705b73d",
        "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
        "previous_state": "NON_BREACHED",
        "computed_state": "BREACHED",
        "would_create_alert": true,
        "notification_status": "SKIPPED",
        "quote_status": "LIVE",
        "observed_price": "195.50000000",
        "threshold_value": "195.00000000",
        "stale": false,
        "stale_reason": null
      }
    ]
  }
}
```

### Validation Error

```json
{
  "error": "validation failed",
  "code": "VALIDATION_ERROR",
  "details": {
    "field": "threshold_rules[0].direction",
    "allowed_values": ["ABOVE", "BELOW"]
  }
}
```

### Forbidden Portfolio Scope

```json
{
  "error": "portfolio scope is forbidden",
  "code": "WATCHLIST_FORBIDDEN_SCOPE",
  "details": {
    "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10"
  }
}
```

### Duplicate Item Conflict

```json
{
  "error": "watchlist item already exists for this scope and security",
  "code": "WATCHLIST_DUPLICATE_ITEM",
  "details": {
    "scope_type": "PORTFOLIO",
    "portfolio_id": "4fa8e199-0b66-4a68-a654-19cf2878ce10",
    "security_id": "44bfb19e-9299-4381-8932-c2c7e4513ac1"
  }
}
```

### Provider Unavailable Response

```json
{
  "error": "market data provider unavailable",
  "code": "WATCHLIST_PROVIDER_UNAVAILABLE",
  "details": {
    "security_id": "2a02cf71-0b49-4e1a-80c6-d8dbb82c9f10",
    "provider_error": "market data unavailable"
  }
}
```

## Swagger Implementation Notes

Future implementation should map this Markdown contract to Swagger annotations as follows:

- Handler comments live in `backend/internal/watchlist/transport/http`.
- DTOs should use stable names such as `WatchlistItemResponse`, `WatchlistItemCreateRequest`, `WatchlistItemUpdateRequest`, `ThresholdRuleResponse`, `ThresholdRuleRequest`, `AlertEventResponse`, `AcknowledgeAlertRequest`, `ManualEvaluateRequest`, and `ManualEvaluateResponse`.
- DTOs should separate transport IDs from descriptor/display objects. Keep IDs for state and API calls, and include descriptor fields for UI labels.
- Successful responses should document `httputil.SuccessResponse{data=<DTO>}` or `httputil.SuccessResponse{data=<ListDTO>}`.
- Error responses should document `httputil.ErrorResponse`.
- Watchlist handlers should populate stable `code` values from the error catalog.
- Swagger examples should preserve decimal strings and RFC3339 timestamps.
- Swagger examples should include display descriptors for portfolio-scoped responses.
- Public DTOs should avoid exposing internal-only fields such as idempotency keys unless intentionally part of an admin/debug response.
- After implementation, regenerate Swagger with:

```bash
make swagger
```

- If frontend API types need regeneration, run:

```bash
make api-client
```

- Do not manually edit generated files under `backend/docs`.

## Open Decisions

- Exact scheduled evaluator cadence.
- Whether market-hours awareness is required.
- Whether threshold rules should later get dedicated CRUD endpoints.
- Whether `WATCHLIST_ADMIN` should allow cross-owner personal management.
- Future portfolio recipient group model.

## API Readiness Checklist

- [ ] Endpoint paths confirmed.
- [ ] Permissions confirmed.
- [ ] Schemas confirmed.
- [ ] Error codes confirmed.
- [ ] Stale quote policy confirmed.
- [ ] Portfolio resolver port required.
- [ ] Notification adapter required.
- [ ] Swagger generation required after implementation.
- [ ] API handler tests required.
- [ ] Permission tests required.
