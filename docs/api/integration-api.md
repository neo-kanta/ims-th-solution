# Integration API

**Status:** Current implementation as of 2026-07-20
**Base path:** `/api/v1`
**Endpoint count:** 4

## Purpose

Provides authenticated dashboard snapshots, personal task feeds, task counts, and official company/mine AUM and P&L summaries.

## Authentication

Bearer JWT is required for every endpoint. The caller identity comes from validated JWT claims.

## Authorization

Dashboard and task endpoints require `INTEGRATION_DASHBOARD_VIEW` in the application layer. The valuation summary is available to every authenticated user under the explicit aggregate-only policy documented below.

## Runtime response conventions

Successful JSON handlers in this module use `{"data": ..., "message": ...}`; `message` is optional. A 204 response has no body. Error responses generally use `{"error": ..., "code": ..., "details": ...}`, although some newer typed-error paths use `{"error_code": ..., "message": ..., "request_id": ...}`. Clients should rely on the status and documented machine code when present, not exact human text.

## Endpoint summary

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/integration/dashboard/me` | Personal dashboard snapshot | INTEGRATION_DASHBOARD_VIEW (application check) | Implemented |
| GET | `/api/v1/integration/dashboard/valuation-summary` | Dashboard AUM / P&L summary | Authenticated; company aggregate is permission-independent, mine is data-scoped | Implemented |
| GET | `/api/v1/integration/tasks/my` | Personal task list | INTEGRATION_DASHBOARD_VIEW (application check) | Implemented |
| GET | `/api/v1/integration/tasks/my/summary` | Personal task summary | INTEGRATION_DASHBOARD_VIEW (application check) | Implemented |

## Endpoint details

### GET `/api/v1/integration/dashboard/me`

Returns the caller's full dashboard read model: tasks, workflow states, and aggregate counts.

**Permission:** INTEGRATION_DASHBOARD_VIEW (application check)

**Business rule:** Dashboard and task endpoints resolve the authenticated identity server-side and do not accept a user ID from the client.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| None |  |  |  |  |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<DashboardSnapshotDTO> |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/integration/transport/router.go`

### GET `/api/v1/integration/dashboard/valuation-summary`

Returns official LIVE-portfolio AUM and today's P&L converted to the configured reporting currency for every authenticated dashboard user. scope=company aggregates all active company funds independent of fund data scope and omits item-level fund/portfolio exclusions; scope=mine restricts to portfolios the authenticated caller manages inside funds they may access. status is AVAILABLE, NO_DATA, or INCOMPLETE. INCOMPLETE returns data_available=false and no usable numeric total; aggregate coverage reports included/excluded counts plus stable reasons for missing, stale, wrong-date, invalid, or currency-mismatched valuation/FX inputs. SIMULATION and MODEL portfolios are excluded as NON_OFFICIAL_PORTFOLIO.

**Permission:** Authenticated; company aggregate is permission-independent, mine is data-scoped

**Business rule:** `scope=company` is the default and returns a company-wide aggregate to any authenticated user without item-level exclusions. `scope=mine` includes portfolios managed by the caller inside accessible funds. Only LIVE portfolios contribute to official totals.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| scope | query | string | No | company (default) or mine |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<ValuationSummaryDTO> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/integration/transport/router.go`

### GET `/api/v1/integration/tasks/my`

Returns the caller's task feed, optionally filtered by module, priority, or status.

**Permission:** INTEGRATION_DASHBOARD_VIEW (application check)

**Business rule:** Dashboard and task endpoints resolve the authenticated identity server-side and do not accept a user ID from the client.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| module | query | string | No | Filter by module (investment\|workflow\|compliance) |
| priority | query | string | No | Filter by priority (HIGH\|MEDIUM\|LOW\|INFO) |
| status | query | string | No | Filter by status (PENDING\|IN_PROGRESS\|COMPLETED) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<TaskListDTO> |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/integration/transport/router.go`

### GET `/api/v1/integration/tasks/my/summary`

Returns aggregate task counts for the caller without the full task list.

**Permission:** INTEGRATION_DASHBOARD_VIEW (application check)

**Business rule:** Dashboard and task endpoints resolve the authenticated identity server-side and do not accept a user ID from the client.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| None |  |  |  |  |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<TaskSummaryDTO> |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/integration/transport/router.go`

## Source references

- `backend/internal/integration/transport/router.go`
- `backend/internal/integration/transport/handler/dashboard_handler.go`
- `backend/internal/integration/application/query`
- `backend/internal/integration/permission/policies.go`
- `backend/pkg/contract/valuation_summary.go`
- `backend/docs/swagger.json`

## Notes

Company valuation coverage is aggregate-only and must not expose excluded fund or portfolio identities. Drill-down APIs keep their existing data-scope rules.
