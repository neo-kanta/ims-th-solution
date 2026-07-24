# Watchlist API - Current Implementation

**Status:** Current implementation as of 2026-07-20
**Base path:** `/api/v1`
**Endpoint count:** 7

## Purpose

Manages personal and portfolio-scoped security watchlists, threshold rules, alert events, acknowledgement, and operator-triggered evaluation.

## Authentication

Bearer JWT is required for every endpoint.

## Authorization

Routes require `WATCHLIST_VIEW`, `WATCHLIST_MANAGE`, `WATCHLIST_ALERT_ACK`, or `WATCHLIST_EVALUATE`. Application services additionally enforce personal ownership or portfolio data scope.

## Runtime response conventions

Successful JSON handlers in this module use `{"data": ..., "message": ...}`; `message` is optional. A 204 response has no body. Error responses generally use `{"error": ..., "code": ..., "details": ...}`, although some newer typed-error paths use `{"error_code": ..., "message": ..., "request_id": ...}`. Clients should rely on the status and documented machine code when present, not exact human text.

## Endpoint summary

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/watchlists` | List watchlist items | WATCHLIST_VIEW + owner/portfolio data scope | Implemented |
| GET | `/api/v1/watchlists/alerts` | List alert events | WATCHLIST_VIEW + owner/portfolio data scope | Implemented |
| POST | `/api/v1/watchlists/alerts/{id}/acknowledge` | Acknowledge an alert event | WATCHLIST_ALERT_ACK + owner/portfolio data scope | Implemented |
| POST | `/api/v1/watchlists/evaluate` | Manually trigger rule evaluation | WATCHLIST_EVALUATE + optional portfolio data scope | Implemented |
| POST | `/api/v1/watchlists/items` | Create a watchlist item | WATCHLIST_MANAGE + owner/portfolio data scope | Implemented |
| PATCH | `/api/v1/watchlists/items/{id}` | Update a watchlist item | WATCHLIST_MANAGE + owner/portfolio data scope | Implemented |
| DELETE | `/api/v1/watchlists/items/{id}` | Delete a watchlist item | WATCHLIST_MANAGE + owner/portfolio data scope | Implemented |

## Endpoint details

### GET `/api/v1/watchlists`

List watchlist items

**Permission:** WATCHLIST_VIEW + owner/portfolio data scope

**Business rule:** Application services enforce personal ownership or portfolio data scope after the route-level Watchlist permission check.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| scope_type | query | string | No | PERSONAL or PORTFOLIO |
| portfolio_id | query | string | No | Portfolio UUID (PORTFOLIO scope only) |
| security_id | query | string | No | Canonical security UUID |
| include_disabled | query | boolean | No | Include disabled items (default false) |
| include_thresholds | query | boolean | No | Include threshold rules (default true) |
| include_quote | query | boolean | No | Include live quote (default true) |
| limit | query | integer | No | Page size (1-200, default 50) |
| offset | query | integer | No | Page offset (default 0) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<ListItemsResponseData> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |

**Source:** `backend/internal/watchlist/transport/http/router.go`

### GET `/api/v1/watchlists/alerts`

List alert events

**Permission:** WATCHLIST_VIEW + owner/portfolio data scope

**Business rule:** Application services enforce personal ownership or portfolio data scope after the route-level Watchlist permission check.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| scope_type | query | string | No | PERSONAL or PORTFOLIO |
| portfolio_id | query | string | No | Portfolio UUID |
| security_id | query | string | No | Canonical security UUID |
| rule_id | query | string | No | Threshold rule UUID |
| acknowledged | query | boolean | No | Filter by acknowledgement state |
| created_from | query | string | No | RFC3339 inclusive lower bound |
| created_to | query | string | No | RFC3339 inclusive upper bound |
| limit | query | integer | No | Page size (1-200, default 50) |
| offset | query | integer | No | Page offset (default 0) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<ListAlertsResponseData> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |

**Source:** `backend/internal/watchlist/transport/http/router.go`

### POST `/api/v1/watchlists/alerts/{id}/acknowledge`

Acknowledge an alert event

**Permission:** WATCHLIST_ALERT_ACK + owner/portfolio data scope

**Business rule:** Application services enforce personal ownership or portfolio data scope after the route-level Watchlist permission check.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| id | path | string | Yes | Alert event UUID |

#### Request body

Schema: `AcknowledgeAlertRequest`.

```json
{
  "note": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<AlertEventResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |

**Source:** `backend/internal/watchlist/transport/http/router.go`

### POST `/api/v1/watchlists/evaluate`

Manually trigger rule evaluation

**Permission:** WATCHLIST_EVALUATE + optional portfolio data scope

**Business rule:** Application services enforce personal ownership or portfolio data scope after the route-level Watchlist permission check.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| None |  |  |  |  |

#### Request body

Schema: `EvaluateRequest`.

```json
{
  "dry_run": false,
  "item_id": "string",
  "portfolio_id": "string",
  "rule_id": "string",
  "scope_type": "string",
  "security_id": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<EvaluateResponseData> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

**Source:** `backend/internal/watchlist/transport/http/router.go`

### POST `/api/v1/watchlists/items`

Create a watchlist item

**Permission:** WATCHLIST_MANAGE + owner/portfolio data scope

**Business rule:** Application services enforce personal ownership or portfolio data scope after the route-level Watchlist permission check.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| None |  |  |  |  |

#### Request body

Schema: `CreateItemRequest`.

```json
{
  "note": "string",
  "pinned": false,
  "portfolio_id": "string",
  "scope_type": "string",
  "security_id": "string",
  "threshold_rules": [
    "<ThresholdRuleRequest>"
  ]
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<WatchlistItemResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |

**Source:** `backend/internal/watchlist/transport/http/router.go`

### PATCH `/api/v1/watchlists/items/{id}`

Update a watchlist item

**Permission:** WATCHLIST_MANAGE + owner/portfolio data scope

**Business rule:** Application services enforce personal ownership or portfolio data scope after the route-level Watchlist permission check.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| id | path | string | Yes | Item UUID |

#### Request body

Schema: `UpdateItemRequest`.

```json
{
  "note": "string",
  "pinned": false,
  "status": "string",
  "threshold_rules": [
    "<ThresholdRuleRequest>"
  ]
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<WatchlistItemResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

**Source:** `backend/internal/watchlist/transport/http/router.go`

### DELETE `/api/v1/watchlists/items/{id}`

Delete a watchlist item

**Permission:** WATCHLIST_MANAGE + owner/portfolio data scope

**Business rule:** Application services enforce personal ownership or portfolio data scope after the route-level Watchlist permission check.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| id | path | string | Yes | Item UUID |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 204 | No Content |  |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

**Source:** `backend/internal/watchlist/transport/http/router.go`

## Source references

- `backend/internal/watchlist/transport/http/router.go`
- `backend/internal/watchlist/transport/http/handler.go`
- `backend/internal/watchlist/application`
- `backend/internal/watchlist/permission/policies.go`
- `backend/docs/swagger.json`

## Notes

This is the live route/DTO reference. `watchlist-api.md` remains the original design contract and contains future-policy discussion that is not automatically part of the implementation.
