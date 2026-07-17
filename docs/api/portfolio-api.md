# Portfolio API

## Purpose
Provides portfolio master, holdings, cash balance, transaction ledger, simulation, reversal, and valuation APIs.

## Business Context
Portfolio APIs support the portfolio and ledger side of Stock Investment Management. Ledger posts and simulations support investment execution/review and are constrained by workflow day state and, where applicable, IRG checks.

## Authentication
All portfolio endpoints require Bearer JWT through `/api/v1`.

## Authorization / Permission
Portfolio read/manage routes use `INVESTMENT_PORTFOLIO_VIEW` and `INVESTMENT_PORTFOLIO_MANAGE`; ledger and valuation routes use `INVESTMENT_LEDGER_POST`, `INVESTMENT_LEDGER_SIMULATE`, `INVESTMENT_LEDGER_REVERSE`, `INVESTMENT_VALUATION_VIEW`, and `INVESTMENT_VALUATION_RUN`.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| GET | /investment/portfolios | List Portfolios | JWT required | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| POST | /investment/portfolios | Create Portfolio | JWT required | INVESTMENT_PORTFOLIO_MANAGE | Implemented |
| GET | /investment/portfolios/{id} | Get Portfolio | JWT required | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| PUT | /investment/portfolios/{id} | Update Portfolio | JWT required | INVESTMENT_PORTFOLIO_MANAGE | Implemented |
| DELETE | /investment/portfolios/{id} | Delete Portfolio | JWT required | INVESTMENT_PORTFOLIO_MANAGE | Implemented |
| GET | /investment/portfolios/{id}/cash | List Portfolio Cash Balances | JWT required | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | /investment/portfolios/{id}/holdings | List Portfolio Holdings | JWT required | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | /investment/portfolios/{id}/transactions | List Portfolio Transactions | JWT required | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| POST | /investment/portfolios/{id}/transactions | Post Portfolio Transaction | JWT required | INVESTMENT_LEDGER_POST | Implemented |
| POST | /investment/portfolios/{id}/transactions/simulate | Simulate Portfolio Transaction | JWT required | INVESTMENT_LEDGER_SIMULATE | Implemented |
| POST | /investment/portfolios/{id}/transactions/{txnId}/reverse | Reverse Portfolio Transaction | JWT required | INVESTMENT_LEDGER_REVERSE | Implemented |
| GET | /investment/portfolios/{id}/valuations | List Portfolio Valuations | JWT required | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | /investment/portfolios/{id}/valuations/latest | Get Latest Portfolio Valuation | JWT required | INVESTMENT_VALUATION_VIEW | Implemented |
| POST | /investment/portfolios/{id}/valuations/run | Run Portfolio Valuation | JWT required | INVESTMENT_VALUATION_RUN | Implemented |

## Endpoint Detail

### GET /investment/portfolios

#### Purpose
List Portfolios List portfolios visible to the authenticated user.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| fund_id | string | No | Fund UUID |
| status | string | No | Portfolio status |
| page | integer | No | Page number (default 1) |
| limit | integer | No | Page size (default 50, max 200) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | PortfolioListResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/portfolios

#### Purpose
Create Portfolio Create a portfolio under a fund.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

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
Schema: `CreatePortfolioRequest`.
```json
{
  "base_currency": "string",
  "benchmark": "string",
  "code": "string",
  "description": "string",
  "fund_id": "string",
  "has_units": false,
  "inception_date": "string",
  "manager_user_id": "string",
  "name": "string",
  "risk_profile": "string",
  "strategy_code": "string",
  "style_id": "string",
  "tax_lot_method": "string",
  "valuation_currency": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | PortfolioResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/portfolios/{id}

#### Purpose
Get Portfolio Retrieve one portfolio by ID.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | PortfolioResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### PUT /investment/portfolios/{id}

#### Purpose
Update Portfolio Update mutable fields on a portfolio using optimistic version control.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `UpdatePortfolioRequest`.
```json
{
  "benchmark": "string",
  "description": "string",
  "expected_version": 0,
  "manager_user_id": "string",
  "name": "string",
  "risk_profile": "string",
  "status": "string",
  "strategy_code": "string",
  "style_id": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | PortfolioResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### DELETE /investment/portfolios/{id}

#### Purpose
Delete Portfolio Soft-delete a portfolio using optimistic version control.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `DeletePortfolioRequest`.
```json
{
  "expected_version": 0
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/portfolios/{id}/cash

#### Purpose
List Portfolio Cash Balances List cash balances by currency for a portfolio.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[CashBalanceResponse] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/portfolios/{id}/holdings

#### Purpose
List Portfolio Holdings List current holdings for a portfolio.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[HoldingResponse] |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/portfolios/{id}/transactions

#### Purpose
List Portfolio Transactions List transactions for a portfolio with optional instrument and date filters.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| instrument_id | string | No | Instrument UUID |
| from | string | No | Start business date (YYYY-MM-DD) |
| to | string | No | End business date (YYYY-MM-DD) |
| page | integer | No | Page number (default 1) |
| limit | integer | No | Page size (default 50, max 200) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | TransactionListResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/portfolios/{id}/transactions

#### Purpose
Post Portfolio Transaction Post a buy, sell, cash, or other portfolio transaction into the ledger.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `PostTransactionRequest`.
```json
{
  "business_date": "string",
  "currency": "string",
  "external_ref": "string",
  "fees": "string",
  "force_post": false,
  "fx_rate_to_base": "string",
  "gross_amount": "string",
  "instrument_id": "string",
  "net_amount": "string",
  "price": "string",
  "quantity": "string",
  "reason": "string",
  "settlement_date": "string",
  "side": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | TransactionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/portfolios/{id}/transactions/simulate

#### Purpose
Simulate Portfolio Transaction Run the post preconditions and pre-trade compliance checks, then preview ledger cash and position impact without mutating investment tables.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `PostTransactionRequest`.
```json
{
  "business_date": "string",
  "currency": "string",
  "external_ref": "string",
  "fees": "string",
  "force_post": false,
  "fx_rate_to_base": "string",
  "gross_amount": "string",
  "instrument_id": "string",
  "net_amount": "string",
  "price": "string",
  "quantity": "string",
  "reason": "string",
  "settlement_date": "string",
  "side": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | TransactionSimulationResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/portfolios/{id}/transactions/{txnId}/reverse

#### Purpose
Reverse Portfolio Transaction Post a reversal transaction for an existing portfolio transaction.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |
| txnId | string | Yes | Transaction UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `ReverseTransactionRequest`.
```json
{
  "business_date": "string",
  "force_post": false,
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | TransactionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/portfolios/{id}/valuations

#### Purpose
List Portfolio Valuations List valuation snapshots for a portfolio with optional date filters.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| from | string | No | Start business date (YYYY-MM-DD) |
| to | string | No | End business date (YYYY-MM-DD) |
| page | integer | No | Page number (default 1) |
| limit | integer | No | Page size (default 50, max 200) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ValuationListResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/portfolios/{id}/valuations/latest

#### Purpose
Get Latest Portfolio Valuation Retrieve the latest internal valuation snapshot for a portfolio.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ValuationResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/portfolios/{id}/valuations/run

#### Purpose
Run Portfolio Valuation Run a valuation snapshot for a portfolio and business date.

#### Business Rule
Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Portfolio UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `RunValuationRequest`.
```json
{
  "business_date": "string",
  "fx_rates": {},
  "total_units": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | ValuationResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

## Source References
- `backend/internal/investment/module.go`
- `backend/internal/investment/transport/handler/investment_handler.go`
- `backend/internal/investment/application/command/post_transaction.go`
- `backend/internal/investment/application/command/reverse_transaction.go`
- `backend/internal/investment/application/service/valuation_runner.go`
- `backend/internal/investment/permission/policies.go`
- `backend/docs/swagger.json`
