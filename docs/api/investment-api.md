# Investment API

## Purpose
Provides APIs for fund and instrument master data, research analysis reports, investment decisions, executions, trade confirmations, fund valuation, AUM, and live market-data overlays.

## Business Context
Investment APIs support the IMS investment process: Investment Research / Analysis Report, Investment Decision, Investment Execution, Trade Confirmation, and Investment Review. Decision submission integrates workflow gating, IRG pre-trade checks, approval workflows, and audit logging.

## Authentication
All investment endpoints require Bearer JWT through `/api/v1`.

## Authorization / Permission
Each route group is gated with an investment permission code such as `INVESTMENT_RESEARCH_SUBMIT`, `INVESTMENT_DECISION_SUBMIT`, `INVESTMENT_EXECUTION_MANAGE`, or `INVESTMENT_CONFIRMATION_IMPORT`.

## Endpoint Summary
| Method | Path | Purpose | Auth | Permission | Status |
| --- | --- | --- | --- | --- | --- |
| GET | /investment/decisions | List Investment Decisions | JWT required | INVESTMENT_DECISION_VIEW | Implemented |
| POST | /investment/decisions | Create Investment Decision | JWT required | INVESTMENT_DECISION_MANAGE | Implemented |
| GET | /investment/decisions/approval-items | List Investment Decision Approval Items | JWT required | INVESTMENT_DECISION_APPROVE | Implemented |
| POST | /investment/decisions/batch-approve | Batch Approve Investment Decisions | JWT required | INVESTMENT_DECISION_APPROVE | Implemented |
| POST | /investment/decisions/batch-reject | Batch Reject Investment Decisions | JWT required | INVESTMENT_DECISION_APPROVE | Implemented |
| GET | /investment/decisions/{id} | Get Investment Decision | JWT required | INVESTMENT_DECISION_VIEW | Implemented |
| PUT | /investment/decisions/{id} | Update Investment Decision | JWT required | INVESTMENT_DECISION_MANAGE | Implemented |
| POST | /investment/decisions/{id}/cancel | Cancel Investment Decision | JWT required | INVESTMENT_DECISION_CANCEL | Implemented |
| GET | /investment/decisions/{id}/details | Get Investment Decision With Lines | JWT required | INVESTMENT_DECISION_VIEW | Implemented |
| POST | /investment/decisions/{id}/submit | Submit Investment Decision | JWT required | INVESTMENT_DECISION_SUBMIT | Implemented |
| GET | /investment/executions | List Trade Executions | JWT required | INVESTMENT_EXECUTION_VIEW | Implemented |
| POST | /investment/executions | Create Trade Execution | JWT required | INVESTMENT_EXECUTION_MANAGE | Implemented |
| GET | /investment/executions/{id} | Get Trade Execution | JWT required | INVESTMENT_EXECUTION_VIEW | Implemented |
| POST | /investment/executions/{id}/cancel | Cancel Trade Execution | JWT required | INVESTMENT_EXECUTION_MANAGE | Implemented |
| POST | /investment/executions/{id}/fill | Fill Trade Execution | JWT required | INVESTMENT_EXECUTION_MANAGE | Implemented |
| GET | /investment/funds | List Funds | JWT required | INVESTMENT_FUND_VIEW | Implemented |
| POST | /investment/funds | Create Fund | JWT required | INVESTMENT_FUND_MANAGE | Implemented |
| GET | /investment/funds/{id} | Get Fund | JWT required | INVESTMENT_FUND_VIEW | Implemented |
| PUT | /investment/funds/{id} | Update Fund | JWT required | INVESTMENT_FUND_MANAGE | Implemented |
| DELETE | /investment/funds/{id} | Delete Fund | JWT required | INVESTMENT_FUND_MANAGE | Implemented |
| GET | /investment/funds/{id}/allocation | Get Fund Allocation | JWT required | INVESTMENT_VALUATION_VIEW | Implemented |
| POST | /investment/funds/{id}/aum/compute | Compute Fund AUM | JWT required | INVESTMENT_FUND_AUM_COMPUTE | Implemented |
| GET | /investment/funds/{id}/holdings/valuation | Get Intraday Holdings Valuation | JWT required | INVESTMENT_VALUATION_VIEW | Implemented |
| POST | /investment/funds/{id}/market-data/refresh | Refresh Fund Market Data | JWT required | INVESTMENT_VALUATION_RUN | Implemented |
| GET | /investment/funds/{id}/market-data/status | Get Fund Market Data Status | JWT required | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | /investment/funds/{id}/nav-history | Get Fund NAV History | JWT required | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | /investment/funds/{id}/nav/latest | Get Latest Fund NAV | JWT required | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | /investment/instruments | List Instruments | JWT required | INVESTMENT_INSTRUMENT_VIEW | Implemented |
| POST | /investment/instruments | Create Instrument | JWT required | INVESTMENT_INSTRUMENT_MANAGE | Implemented |
| GET | /investment/instruments/{id} | Get Instrument | JWT required | INVESTMENT_INSTRUMENT_VIEW | Implemented |
| PUT | /investment/instruments/{id} | Update Instrument | JWT required | INVESTMENT_INSTRUMENT_MANAGE | Implemented |
| POST | /investment/instruments/{id}/prices | Post Price Snapshot | JWT required | INVESTMENT_PRICE_POST | Implemented |
| GET | /investment/reference/asset-classes | List Asset Classes | JWT required | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | /investment/reference/asset-subtypes | List Asset Subtypes | JWT required | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | /investment/reference/countries | List Countries | JWT required | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | /investment/reference/fund-categories | List Fund Categories | JWT required | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | /investment/reference/investment-styles | List Investment Styles | JWT required | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | /investment/reference/regions | List Regions | JWT required | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | /investment/reference/sectors | List Sectors | JWT required | INVESTMENT_REFERENCE_VIEW | Implemented |
| GET | /investment/research-reports | List Investment Research Reports | JWT required | INVESTMENT_RESEARCH_VIEW | Implemented |
| POST | /investment/research-reports | Create Investment Research Report | JWT required | INVESTMENT_RESEARCH_CREATE | Implemented |
| GET | /investment/research-reports/{id} | Get Investment Research Report | JWT required | INVESTMENT_RESEARCH_VIEW | Implemented |
| PUT | /investment/research-reports/{id} | Update Investment Research Report | JWT required | INVESTMENT_RESEARCH_UPDATE | Implemented |
| DELETE | /investment/research-reports/{id} | Delete Investment Research Report | JWT required | INVESTMENT_RESEARCH_DELETE | Implemented |
| POST | /investment/research-reports/{id}/cancel-submit | Cancel Submission Of Investment Research Report | JWT required | INVESTMENT_RESEARCH_CANCEL_SUBMIT | Implemented |
| POST | /investment/research-reports/{id}/invalidate | Invalidate Investment Research Report | JWT required | INVESTMENT_RESEARCH_INVALIDATE | Implemented |
| POST | /investment/research-reports/{id}/submit | Submit Investment Research Report | JWT required | INVESTMENT_RESEARCH_SUBMIT | Implemented |
| GET | /investment/trade-confirmations | List Trade Confirmations | JWT required | INVESTMENT_CONFIRMATION_VIEW | Implemented |
| POST | /investment/trade-confirmations | Record Trade Confirmation | JWT required | INVESTMENT_CONFIRMATION_MANAGE | Implemented |
| POST | /investment/trade-confirmations/batch | Import Trade Confirmation Batch | JWT required | INVESTMENT_CONFIRMATION_IMPORT | Implemented |
| GET | /investment/trade-confirmations/{id} | Get Trade Confirmation | JWT required | INVESTMENT_CONFIRMATION_VIEW | Implemented |
| POST | /investment/trade-confirmations/{id}/resolve | Resolve Trade Confirmation | JWT required | INVESTMENT_CONFIRMATION_MANAGE | Implemented |

## Endpoint Detail

### GET /investment/decisions

#### Purpose
List Investment Decisions Returns a paginated list of investment decisions with optional filters.

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| page | integer | No | Page number (default 1) |
| limit | integer | No | Page size (default 50, max 200) |
| fund_id | string | No | Filter by fund UUID |
| portfolio_id | string | No | Filter by portfolio UUID |
| contract_id | string | No | Filter by contract UUID |
| business_date | string | No | Filter by business date YYYY-MM-DD |
| status | string | No | Filter by lifecycle status |
| instrument_code | string | No | Filter by instrument code |
| search | string | No | Substring search on decision_no / instrument_code |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DecisionListResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/decisions

#### Purpose
Create Investment Decision Creates a new investment decision in DRAFT status.

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

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
Schema: `CreateDecisionRequest`.
```json
{
  "amount": "string",
  "business_date": "string",
  "contract_id": "string",
  "currency": "string",
  "exchange": "string",
  "fund_id": "string",
  "instrument_code": "string",
  "instrument_id": "string",
  "limit_price": "string",
  "portfolio_id": "string",
  "quantity": "string",
  "rationale": "string",
  "research_report_id": "string",
  "side": "BUY"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | DecisionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/decisions/approval-items

#### Purpose
List Investment Decision Approval Items Returns decisions pending approval (default status=PENDING_APPROVAL), enriched with current and previous approver names and stage numbers.

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| page | integer | No | Page number (default 1) |
| limit | integer | No | Page size (default 50, max 200) |
| portfolio_id | string | No | Filter by portfolio UUID |
| fund_id | string | No | Filter by fund UUID |
| business_date_from | string | No | Inclusive lower bound YYYY-MM-DD |
| business_date_to | string | No | Inclusive upper bound YYYY-MM-DD |
| decision_no | string | No | Filter by exact decision number |
| process_type | string | No | INVESTMENT_DECISION \| ORDER_CANCEL \| ORDER_AMEND |
| product_type | string | No | MUTUAL_FUND \| ETF \| STOCK \| BOND \| CASH |
| research_no | string | No | Filter by research report number |
| status | string | No | Decision lifecycle status (default PENDING_APPROVAL) |
| search | string | No | Substring search on decision_no / instrument_code / research_report_no |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DecisionListResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/decisions/batch-approve

#### Purpose
Batch Approve Investment Decisions Approves multiple investment decision headers in a single call. Each decision must be PENDING_APPROVAL and have a pending approval task assigned to the authenticated user. Returns per-decision results - partial failures are reported without aborting the rest.

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

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
Schema: `BatchApprovalRequest`.
```json
{
  "comment": "string",
  "decision_nos": [
    "string"
  ]
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | BatchApprovalResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/decisions/batch-reject

#### Purpose
Batch Reject Investment Decisions Rejects multiple investment decision headers in a single call. A rejection reason is required and applied to all selected decisions. Returns per-decision results.

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

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
Schema: `BatchRejectionRequest`.
```json
{
  "decision_nos": [
    "string"
  ],
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | BatchApprovalResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/decisions/{id}

#### Purpose
Get Investment Decision

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DecisionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### PUT /investment/decisions/{id}

#### Purpose
Update Investment Decision

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `UpdateDecisionRequest`.
```json
{
  "fund_id": "uuid",
  "portfolio_id": "uuid",
  "instrument_id": "uuid",
  "side": "BUY",
  "quantity": "100.00",
  "amount": "100000.00",
  "business_date": "2026-06-26",
  "decision_type": "SINGLE_ORDER",
  "research_report_id": "uuid"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DecisionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/decisions/{id}/cancel

#### Purpose
Cancel Investment Decision Cancels a DRAFT or SUBMITTED decision. A cancellation reason is required.

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Decision UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `CancelDecisionRequest`.
```json
{
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DecisionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/decisions/{id}/details

#### Purpose
Get Investment Decision With Lines Returns a decision header with all child decision lines (for basket/rebalance/switch decisions) and approval stage info.

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Decision UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DecisionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/decisions/{id}/submit

#### Purpose
Submit Investment Decision Submits a DRAFT decision for approval. Transitions status to SUBMITTED.

#### Business Rule
Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Decision UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | DecisionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/executions

#### Purpose
List Trade Executions

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| decision_id | uuid | No | Decision UUID |
| contract_id | uuid | No | Contract/fund UUID |
| business_date | string | No | Business date; required with contract_id |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | Execution list object |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/executions

#### Purpose
Create Trade Execution

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

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
```json
{
  "decision_id": "uuid",
  "ordered_quantity": "100.00",
  "ordered_amount": "100000.00",
  "broker_reference": "BRK-001"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | ExecutionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/executions/{id}

#### Purpose
Get Trade Execution

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ExecutionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/executions/{id}/cancel

#### Purpose
Cancel Trade Execution

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ExecutionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/executions/{id}/fill

#### Purpose
Fill Trade Execution

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "executed_quantity": "100.00",
  "executed_amount": "100000.00",
  "execution_price": "1000.00",
  "status": "EXECUTED",
  "broker_reference": "BRK-001"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ExecutionResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/funds

#### Purpose
List Funds List funds visible to the authenticated user.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| page | integer | No | Page number (default 1) |
| limit | integer | No | Page size (default 50, max 200) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | FundListResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/funds

#### Purpose
Create Fund Create a fund record.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

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
Schema: `CreateFundRequest`.
```json
{
  "base_currency": "string",
  "benchmark": "string",
  "code": "string",
  "external_pam_ref": "string",
  "fund_category_id": "string",
  "has_units": false,
  "inception_date": "string",
  "manager_user_id": "string",
  "name": "string",
  "require_pretrade_preview": false,
  "risk_profile": "string",
  "short_name": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | FundResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/funds/{id}

#### Purpose
Get Fund Retrieve one fund by ID.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | FundResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### PUT /investment/funds/{id}

#### Purpose
Update Fund Update mutable fields on a fund using optimistic version control.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `UpdateFundRequest`.
```json
{
  "benchmark": "string",
  "expected_version": 0,
  "external_pam_ref": "string",
  "fund_category_id": "string",
  "manager_user_id": "string",
  "name": "string",
  "require_pretrade_preview": false,
  "risk_profile": "string",
  "short_name": "string",
  "status": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | FundResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### DELETE /investment/funds/{id}

#### Purpose
Delete Fund Soft-delete a fund using optimistic version control.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `DeleteFundRequest`.
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

### GET /investment/funds/{id}/allocation

#### Purpose
Get Fund Allocation Compute asset-class, sector, country and currency breakdowns for a fund (read-only).

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | FundAllocationResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/funds/{id}/aum/compute

#### Purpose
Compute Fund AUM Aggregate portfolio AUM snapshots into a fund-level snapshot for a business date.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `ComputeFundAUMRequest`.
```json
{
  "business_date": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ComputeFundAUMResponse |
| 201 | Created | ComputeFundAUMResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/funds/{id}/holdings/valuation

#### Purpose
Get Intraday Holdings Valuation Compute estimated NAV / AUM and per-position unrealised P&L from live market data, alongside the official accounting NAV.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | IntradayValuationResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/funds/{id}/market-data/refresh

#### Purpose
Refresh Fund Market Data Trigger a provider fetch for every instrument held by the fund. Records an audit event.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | MarketDataRefreshResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/funds/{id}/market-data/status

#### Purpose
Get Fund Market Data Status Report provider health and the number of stale positions for a fund.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | MarketDataStatusResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/funds/{id}/nav-history

#### Purpose
Get Fund NAV History Time series of NAV-per-unit (unitised funds) or AUM (non-unitised) over a range.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| range | string | No | Window: 1M, 3M, 6M, 1Y, 5Y, YTD (default 3M) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | FundNAVHistoryResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/funds/{id}/nav/latest

#### Purpose
Get Latest Fund NAV Aggregate latest per-portfolio valuations into a fund-level NAV view (read-only; no side effects).

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Fund UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | FundNAVResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/instruments

#### Purpose
List Instruments List tradable and reference instruments with optional filters.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| asset_class_id | string | No | Asset class UUID |
| search | string | No | Search text |
| page | integer | No | Page number (default 1) |
| limit | integer | No | Page size (default 50, max 200) |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | InstrumentListResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/instruments

#### Purpose
Create Instrument Create an instrument master record.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

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
Schema: `CreateInstrumentRequest`.
```json
{
  "asset_class_id": "string",
  "asset_subtype_id": "string",
  "attributes": {},
  "country_id": "string",
  "currency": "string",
  "fund_category_id": "string",
  "lot_size": 0,
  "name": "string",
  "primary_exchange": "string",
  "primary_ticker": "string",
  "region_id": "string",
  "sector_id": "string",
  "tick_size": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | InstrumentResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/instruments/{id}

#### Purpose
Get Instrument Retrieve one instrument by ID.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Instrument UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | InstrumentResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### PUT /investment/instruments/{id}

#### Purpose
Update Instrument Update mutable fields on an instrument master record.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Instrument UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `UpdateInstrumentRequest`.
```json
{
  "attributes": {},
  "fund_category_id": "string",
  "is_tradable": false,
  "lot_size": 0,
  "name": "string",
  "primary_exchange": "string",
  "sector_id": "string",
  "status": "string",
  "tick_size": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | InstrumentResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/instruments/{id}/prices

#### Purpose
Post Price Snapshot Capture a price snapshot for an instrument and business date.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Instrument UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `PostPriceSnapshotRequest`.
```json
{
  "business_date": "string",
  "currency": "string",
  "is_stale": false,
  "price": "string",
  "price_source": "string",
  "provider_ref": "string",
  "stale_reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | PriceResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/reference/asset-classes

#### Purpose
List Asset Classes List active asset classes.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

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
| 200 | OK | array[AssetClass] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/reference/asset-subtypes

#### Purpose
List Asset Subtypes List active asset subtypes, optionally filtered by asset class.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| asset_class_id | string | No | Asset class UUID |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | array[AssetSubtype] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/reference/countries

#### Purpose
List Countries List active countries.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

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
| 200 | OK | array[Country] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/reference/fund-categories

#### Purpose
List Fund Categories List active fund categories.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

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
| 200 | OK | array[FundCategory] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/reference/investment-styles

#### Purpose
List Investment Styles List active investment styles.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

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
| 200 | OK | array[InvestmentStyle] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/reference/regions

#### Purpose
List Regions List active regions.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

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
| 200 | OK | array[Region] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/reference/sectors

#### Purpose
List Sectors List investment sector taxonomy rows.

#### Business Rule
Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows.

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
| 200 | OK | array[Sector] |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/research-reports

#### Purpose
List Investment Research Reports Paginated list of research reports with optional filters.

#### Business Rule
Supports Investment Research / Analysis Report lifecycle and approval readiness for decisions.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| page | integer | No | Page number (default 1) |
| limit | integer | No | Page size (default 50, max 200) |
| report_status | string | No | DRAFT \| ACTIVE \| EXPIRED \| REJECTED |
| review_status | string | No | NOT_SUBMITTED \| SUBMITTED \| REVIEW_COMPLETED |
| recommendation | string | No | BUY \| SELL \| HOLD |
| instrument_code | string | No | Exact instrument code filter |
| owner_user_id | string | No | Owner user UUID |
| report_date_from | string | No | Inclusive lower bound YYYY-MM-DD |
| report_date_to | string | No | Inclusive upper bound YYYY-MM-DD |
| search | string | No | Substring search on report_no/instrument_code/report_title |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ResearchReportListResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/research-reports

#### Purpose
Create Investment Research Report Create a new DRAFT research report.

#### Business Rule
Supports Investment Research / Analysis Report lifecycle and approval readiness for decisions.

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
Schema: `CreateResearchReportRequest`.
```json
{
  "applicable_contract_id": "string",
  "author_user_id": "string",
  "company_outlook": "string",
  "company_overview": "string",
  "currency": "string",
  "effective_date": "string",
  "esg_comment": "string",
  "financial_status": "string",
  "instrument_code": "string",
  "instrument_name": "string",
  "instrument_type": "string",
  "investment_analysis": "string",
  "market": "string",
  "owner_user_id": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | ResearchReportResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/research-reports/{id}

#### Purpose
Get Investment Research Report Retrieve one research report by ID.

#### Business Rule
Supports Investment Research / Analysis Report lifecycle and approval readiness for decisions.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Research report UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ResearchReportResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### PUT /investment/research-reports/{id}

#### Purpose
Update Investment Research Report Apply partial updates to a research report. Refused when the report has been deleted or its review is completed.

#### Business Rule
Supports Investment Research / Analysis Report lifecycle and approval readiness for decisions.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Research report UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `UpdateResearchReportRequest`.
```json
{
  "applicable_contract_id": "string",
  "author_user_id": "string",
  "company_outlook": "string",
  "company_overview": "string",
  "currency": "string",
  "effective_date": "string",
  "esg_comment": "string",
  "financial_status": "string",
  "instrument_code": "string",
  "instrument_name": "string",
  "instrument_type": "string",
  "investment_analysis": "string",
  "market": "string",
  "owner_user_id": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ResearchReportResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### DELETE /investment/research-reports/{id}

#### Purpose
Delete Investment Research Report Soft-delete a research report. Only allowed while review_status = NOT_SUBMITTED.

#### Business Rule
Supports Investment Research / Analysis Report lifecycle and approval readiness for decisions.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Research report UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/research-reports/{id}/cancel-submit

#### Purpose
Cancel Submission Of Investment Research Report Move review_status from SUBMITTED back to NOT_SUBMITTED. Refused once review has been completed.

#### Business Rule
Supports Investment Research / Analysis Report lifecycle and approval readiness for decisions.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Research report UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ResearchReportResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/research-reports/{id}/invalidate

#### Purpose
Invalidate Investment Research Report One-way transition to INVALIDATED. An invalidated report cannot be referenced by a decision, edited, submitted, cancelled, or soft-deleted. The reason (>=20 characters) is stored and audited.

#### Business Rule
Supports Investment Research / Analysis Report lifecycle and approval readiness for decisions.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Research report UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
Schema: `InvalidateResearchReportRequest`.
```json
{
  "reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ResearchReportResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/research-reports/{id}/submit

#### Purpose
Submit Investment Research Report Move review_status from NOT_SUBMITTED to SUBMITTED. No real approval workflow is invoked.

#### Business Rule
Supports Investment Research / Analysis Report lifecycle and approval readiness for decisions.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Research report UUID |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | ResearchReportResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/trade-confirmations

#### Purpose
List Trade Confirmations

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| execution_id | uuid | No | Execution UUID |
| contract_id | uuid | No | Contract/fund UUID |
| business_date | string | No | Business date; required with contract_id |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | Trade confirmation list object |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/trade-confirmations

#### Purpose
Record Trade Confirmation

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

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
```json
{
  "execution_id": "uuid",
  "confirmed_quantity": "100.00",
  "confirmed_amount": "100000.00",
  "confirmed_price": "1000.00",
  "broker_reference": "BRK-001",
  "import_batch_id": "uuid"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | TradeConfirmationResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/trade-confirmations/batch

#### Purpose
Import Trade Confirmation Batch

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

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
```json
{
  "source_filename": "broker-confirmations.json",
  "rows": [
    {
      "execution_id": "uuid",
      "confirmed_quantity": "100.00",
      "confirmed_amount": "100000.00",
      "confirmed_price": "1000.00",
      "broker_reference": "BRK-001"
    }
  ]
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 201 | Created | ConfirmationBatchImportResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### GET /investment/trade-confirmations/{id}

#### Purpose
Get Trade Confirmation

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
No request body.

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | TradeConfirmationResponse |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

#### Source
`backend/internal/investment/module.go`.

### POST /investment/trade-confirmations/{id}/resolve

#### Purpose
Resolve Trade Confirmation

#### Business Rule
Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review.

#### Request

##### Path Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| id | string | Yes | Path parameter |

##### Query Parameters
| Name | Type | Required | Description |
| --- | --- | --- | --- |
| None |  |  |  |

##### Request Body
```json
{
  "target_status": "MATCHED",
  "discrepancy_reason": "string"
}
```

#### Response
| HTTP Status | Description | Schema |
| --- | --- | --- |
| 200 | OK | TradeConfirmationResponse |
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
- `backend/internal/investment/transport/handler`
- `backend/internal/investment/transport/dto/request/requests.go`
- `backend/internal/investment/transport/dto/response/responses.go`
- `backend/internal/investment/application/command`
- `backend/internal/investment/permission/policies.go`
- `backend/docs/swagger.json`
## Notes
Portfolio master, holdings, cash, ledger, and portfolio valuation endpoints are documented separately in `portfolio-api.md` even though they are mounted by the investment module.
