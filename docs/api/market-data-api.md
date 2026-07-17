# Market Data API

## Purpose
Provides market quote, historical price, import, import-batch, provider health, and frontend market-data screen endpoints.

## Business Context
Market data supports investment valuation, intraday holdings views, and market-data refresh workflows. Investment uses this module through a contract quote provider for intraday valuation overlays.

## Authentication
All market-data endpoints are mounted inside authenticated `/api/v1`; Bearer JWT is required.

## Authorization / Permission
Permission rule not found in code. `backend/internal/market_data/permission/policies.go` declares no operator-grantable permission codes and the router does not apply `RequirePermission` middleware.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| GET | /market-data/history | Get Market Price History | JWT required | Permission rule not found in code. | Implemented |
| POST | /market-data/import | Import Market Data | JWT required | Permission rule not found in code. | Implemented |
| POST | /market-data/import-batches | Create Market Data Import Batch | JWT required | Permission rule not found in code. | Implemented |
| GET | /market-data/import-batches/{batch_id} | Get Market Data Import Batch | JWT required | Permission rule not found in code. | Implemented |
| GET | /market-data/import-batches/{batch_id}/errors | List Market Data Import Batch Errors | JWT required | Permission rule not found in code. | Implemented |
| POST | /market-data/import-batches/{batch_id}/run | Run Market Data Import Batch | JWT required | Permission rule not found in code. | Implemented |
| GET | /market-data/provider-health | Get Market Data Provider Health | JWT required | Permission rule not found in code. | Implemented |
| GET | /market-data/quote | Get Market Quote | JWT required | Permission rule not found in code. | Implemented |
| GET | /market-data/screen/search | Market Data screen - search | JWT required | Permission rule not found in code. | Implemented |
| GET | /market-data/screen/watchlist | Market Data screen - watchlist | JWT required | Permission rule not found in code. | Implemented |

## Endpoint Detail

### GET /market-data/history

#### Purpose
Get Market Price History Get daily price bars for a symbol from the configured provider chain or a requested provider.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| symbol | string | Yes | Market symbol |
| provider | string | No | Provider name (alpha_vantage or yahoo) |
| limit | integer | No | Maximum bars to return (default 250, max 1000) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[PriceBarResponse] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 502 | Bad Gateway | ErrorResponse |

#### Source
`backend/internal/market_data/transport/http/router.go`.

### POST /market-data/import

#### Purpose
Import Market Data Fetch and persist quote and/or daily price history for a symbol.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `ImportMarketDataRequest`.
```json
{
  "history_limit": 0,
  "include_history": false,
  "include_quote": false,
  "provider": "string",
  "symbol": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | ImportMarketDataResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 502 | Bad Gateway | ErrorResponse |

#### Source
`backend/internal/market_data/transport/http/router.go`.

### POST /market-data/import-batches

#### Purpose
Create Market Data Import Batch Create a chunked market data import batch for the given symbols.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `CreateImportBatchRequest`.
```json
{
  "chunk_size": 0,
  "history_limit": 0,
  "idempotency_key": "string",
  "import_type": "string",
  "include_history": false,
  "include_quote": false,
  "provider": "string",
  "symbols": [
    "string"
  ]
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | CreateImportBatchResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |

#### Source
`backend/internal/market_data/transport/http/router.go`.

### GET /market-data/import-batches/{batch_id}

#### Purpose
Get Market Data Import Batch Returns batch status with chunk summary and error rows.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| batch_id | string | Yes | Import batch id |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ImportBatchStatusResponseDTO |
| 400 | Bad Request | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/market_data/transport/http/router.go`.

### GET /market-data/import-batches/{batch_id}/errors

#### Purpose
List Market Data Import Batch Errors Returns rejected/failed/warning items for the batch.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| batch_id | string | Yes | Import batch id |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[ImportChunkItem] |
| 400 | Bad Request | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/market_data/transport/http/router.go`.

### POST /market-data/import-batches/{batch_id}/run

#### Purpose
Run Market Data Import Batch Runs all chunks for the batch synchronously.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| batch_id | string | Yes | Import batch id |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | RunImportBatchResponse |
| 400 | Bad Request | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 502 | Bad Gateway | ErrorResponse |

#### Source
`backend/internal/market_data/transport/http/router.go`.

### GET /market-data/provider-health

#### Purpose
Get Market Data Provider Health Get configured market data provider health and recent status information.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[ProviderHealthResponse] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/market_data/transport/http/router.go`.

### GET /market-data/quote

#### Purpose
Get Market Quote Get a latest quote for a symbol from the configured provider chain or a requested provider.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| symbol | string | Yes | Market symbol |
| provider | string | No | Provider name (alpha_vantage or yahoo) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | QuoteResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 502 | Bad Gateway | ErrorResponse |

#### Source
`backend/internal/market_data/transport/http/router.go`.

### GET /market-data/screen/search

#### Purpose
Market Data screen - search Search canonical securities with latest snapshot data and add-to-watchlist hint.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| query | string | No | Free-text query |
| limit | integer | No | Maximum rows (default 50, max 200) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ScreenSearchResponseDTO |

#### Source
`backend/internal/market_data/transport/http/router.go`.

### GET /market-data/screen/watchlist

#### Purpose
Market Data screen - watchlist Canonical securities joined with their latest snapshot data. Frontend-safe DTO.

#### Business Rule
Supplies quote and history data used by investment valuation and dashboards.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| limit | integer | No | Maximum rows (default 100, max 500) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ScreenWatchlistResponseDTO |
| 401 | Unauthorized | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/market_data/transport/http/router.go`.

## Source References
- `backend/internal/market_data/module.go`
- `backend/internal/market_data/transport/http/router.go`
- `backend/internal/market_data/transport/http/handler.go`
- `backend/internal/market_data/permission/policies.go`
- `backend/docs/swagger.json`
