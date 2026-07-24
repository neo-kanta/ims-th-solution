# Portfolio V2 API - Current Implementation

**Status:** Current implementation as of 2026-07-20
**Base path:** `/api/v2`
**Endpoint count:** 25

## Purpose

Provides the implemented portfolio-code API for portfolio reads, ledger operations, valuations, decisions, executions, confirmations, and portfolio-scoped compliance.

## Authentication

Bearer JWT is required for every endpoint under `/api/v2`.

## Authorization

Every route group applies the investment, workflow, or IRG permission shown in the endpoint table. Existing data-scope and aggregate-membership checks continue in the application and repository path.

## Runtime response conventions

Successful JSON handlers in this module use `{"data": ..., "message": ...}`; `message` is optional. A 204 response has no body. Error responses generally use `{"error": ..., "code": ..., "details": ...}`, although some newer typed-error paths use `{"error_code": ..., "message": ..., "request_id": ...}`. Clients should rely on the status and documented machine code when present, not exact human text.

## Endpoint summary

| Method | Path | Purpose | Authorization | Status |
| --- | --- | --- | --- | --- |
| GET | `/api/v2/portfolios/{portfolioCode}` | Get Portfolio By Code | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/cash` | Get Portfolio Cash Balances By Code | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/compliance/breaches` | List Compliance Breaches For Portfolio (V2) | IRG_VIEW_RULES | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/compliance/checks/post-trade` | Run Portfolio Post-Trade Compliance Check (V2) | WORKFLOW_EXECUTE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/compliance/checks/pre-trade` | Run Portfolio Pre-Trade Compliance Check (V2) | WORKFLOW_EXECUTE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/compliance/rules` | List Compliance Rules For Portfolio (V2) | IRG_VIEW_RULES | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings` | Bind Compliance Rule To Portfolio (V2) | IRG_EDIT_BINDING | Implemented |
| DELETE | `/api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}` | Deactivate Portfolio Compliance Rule Binding (V2) | IRG_EDIT_BINDING | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve` | Resolve Portfolio Trade Confirmation By Code | INVESTMENT_CONFIRMATION_MANAGE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/decisions` | List Portfolio Decisions By Code | INVESTMENT_DECISION_VIEW | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/decisions` | Create Portfolio Decision By Code | INVESTMENT_DECISION_MANAGE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}` | Get Portfolio Decision By Code | INVESTMENT_DECISION_VIEW | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/cancel` | Cancel Portfolio Decision By Code | INVESTMENT_DECISION_CANCEL | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/executions` | Create Portfolio Execution By Code | INVESTMENT_EXECUTION_MANAGE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/submit` | Submit Portfolio Decision By Code | INVESTMENT_DECISION_SUBMIT | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/cancel` | Cancel Portfolio Execution By Code | INVESTMENT_EXECUTION_MANAGE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/confirmations` | Record Portfolio Trade Confirmation By Code | INVESTMENT_CONFIRMATION_MANAGE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/fill` | Fill Portfolio Execution By Code | INVESTMENT_EXECUTION_MANAGE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/holdings` | Get Portfolio Holdings By Code | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/transactions` | List Portfolio Transactions By Code | INVESTMENT_PORTFOLIO_VIEW | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/transactions` | Post Portfolio Transaction By Code | INVESTMENT_LEDGER_POST | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/transactions/simulate` | Simulate Portfolio Transaction By Code | INVESTMENT_LEDGER_SIMULATE | Implemented |
| POST | `/api/v2/portfolios/{portfolioCode}/transactions/{transactionId}/reverse` | Reverse Portfolio Transaction By Code | INVESTMENT_LEDGER_REVERSE | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/valuations` | List Portfolio Valuations By Code | INVESTMENT_VALUATION_VIEW | Implemented |
| GET | `/api/v2/portfolios/{portfolioCode}/valuations/latest` | Get Latest Portfolio Valuation By Code | INVESTMENT_VALUATION_VIEW | Implemented |

## Endpoint details

### GET `/api/v2/portfolios/{portfolioCode}`

Retrieve one portfolio by its business code (Portfolio V2).

**Permission:** INVESTMENT_PORTFOLIO_VIEW

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<PortfolioResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### GET `/api/v2/portfolios/{portfolioCode}/cash`

List cash balances by currency for a portfolio, resolved by business code (Portfolio V2).

**Permission:** INVESTMENT_PORTFOLIO_VIEW

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<array[CashBalanceResponse]> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### GET `/api/v2/portfolios/{portfolioCode}/compliance/breaches`

List compliance breaches for a portfolio resolved by business code.

**Permission:** IRG_VIEW_RULES

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| status | query | string | No | Breach status |
| rule_type_id | query | string | No | Rule type ID |
| date_from | query | string | No | Start business date (YYYY-MM-DD) |
| date_to | query | string | No | End business date (YYYY-MM-DD) |
| page | query | integer | No | Page number (default 1) |
| limit | query | integer | No | Page size (default 50, max 200) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<array[PortfolioBreachView]> |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/compliance/checks/post-trade`

Evaluate compliance rules for a portfolio resolved by business code. fund_id is used only if the portfolio has one.

**Permission:** WORKFLOW_EXECUTE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

Schema: `PortfolioPostTradeRequest`.

```json
{
  "business_date": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<ProposedOrderResult> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/compliance/checks/pre-trade`

Evaluate compliance rules for a proposed order against a portfolio resolved by business code. fund_id is used only if the portfolio has one.

**Permission:** WORKFLOW_EXECUTE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

Schema: `PortfolioPreTradeRequest`.

```json
{
  "business_date": "string",
  "currency": "string",
  "exchange": "string",
  "fees": "string",
  "order_id": "string",
  "price": "string",
  "quantity": "string",
  "side": "string",
  "ticker": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<ProposedOrderResult> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### GET `/api/v2/portfolios/{portfolioCode}/compliance/rules`

List the active rule catalog, annotated with each rule's binding to this portfolio when one exists.

**Permission:** IRG_VIEW_RULES

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<array[PortfolioRuleCatalogEntry]> |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings`

Bind an existing rule instance to this portfolio (scope_type=PORTFOLIO).

**Permission:** IRG_EDIT_BINDING

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| ruleInstanceID | path | string | Yes | Rule instance UUID |

#### Request body

Schema: `BindRuleRequest`.

```json
{
  "effective_from": "string",
  "effective_to": "string",
  "priority": 0,
  "severity": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<PortfolioRuleBindingView> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### DELETE `/api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}`

Deactivate a rule binding on this portfolio. Bindings owned by another portfolio are reported as 404.

**Permission:** IRG_EDIT_BINDING

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| ruleInstanceID | path | string | Yes | Rule instance UUID (unused for lookup; kept for a stable REST shape) |
| bindingID | path | string | Yes | Binding UUID |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 204 | No Content |  |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve`

Matches or reviews a trade confirmation that belongs to the resolved portfolio (Portfolio V2).

**Permission:** INVESTMENT_CONFIRMATION_MANAGE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| confirmationId | path | string | Yes | Confirmation UUID |

#### Request body

Schema: `ResolveConfirmationRequest`.

```json
{
  "discrepancy_reason": "string",
  "target_status": "MATCHED"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<TradeConfirmationResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### GET `/api/v2/portfolios/{portfolioCode}/decisions`

List investment decisions for a portfolio, resolved by business code (Portfolio V2).

**Permission:** INVESTMENT_DECISION_VIEW

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| page | query | integer | No | Page number (default 1) |
| limit | query | integer | No | Page size (default 50, max 200) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<DecisionListResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/decisions`

Creates a new investment decision in DRAFT status for a portfolio, resolved by business code (Portfolio V2). Request body must not include fund_id, portfolio_id, or contract_id.

**Permission:** INVESTMENT_DECISION_MANAGE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

Schema: `CreateDecisionV2Request`.

```json
{
  "amount": "string",
  "business_date": "string",
  "currency": "string",
  "exchange": "string",
  "instrument_code": "string",
  "instrument_id": "string",
  "limit_price": "string",
  "quantity": "string",
  "rationale": "string",
  "research_report_id": "string",
  "side": "BUY"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<DecisionResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### GET `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}`

Retrieve one investment decision that belongs to the resolved portfolio (Portfolio V2).

**Permission:** INVESTMENT_DECISION_VIEW

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| decisionId | path | string | Yes | Decision UUID |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<DecisionResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/cancel`

Cancels a DRAFT or SUBMITTED decision that belongs to the resolved portfolio (Portfolio V2).

**Permission:** INVESTMENT_DECISION_CANCEL

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| decisionId | path | string | Yes | Decision UUID |

#### Request body

Schema: `CancelDecisionRequest`.

```json
{
  "reason": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<DecisionResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/executions`

Opens an execution against an APPROVED decision that belongs to the resolved portfolio (Portfolio V2).

**Permission:** INVESTMENT_EXECUTION_MANAGE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| decisionId | path | string | Yes | Decision UUID |

#### Request body

Schema: `CreateExecutionV2Request`.

```json
{
  "broker_reference": "string",
  "ordered_amount": "string",
  "ordered_quantity": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<ExecutionResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | COMPLIANCE_NOT_CONFIGURED, COMPLIANCE_UNAVAILABLE, or evaluated rule rejection | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/submit`

Submits a DRAFT decision that belongs to the resolved portfolio for approval (Portfolio V2).

**Permission:** INVESTMENT_DECISION_SUBMIT

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| decisionId | path | string | Yes | Decision UUID |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<DecisionResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | COMPLIANCE_NOT_CONFIGURED, COMPLIANCE_UNAVAILABLE, or evaluated rule rejection | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/cancel`

Cancels an execution that belongs to the resolved portfolio (Portfolio V2).

**Permission:** INVESTMENT_EXECUTION_MANAGE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| executionId | path | string | Yes | Execution UUID |

#### Request body

Schema: `CancelExecutionRequest`.

```json
{
  "reason": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<ExecutionResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/confirmations`

Records a broker confirmation against an execution that belongs to the resolved portfolio (Portfolio V2).

**Permission:** INVESTMENT_CONFIRMATION_MANAGE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| executionId | path | string | Yes | Execution UUID |

#### Request body

Schema: `RecordConfirmationV2Request`.

```json
{
  "broker_reference": "string",
  "confirmed_amount": "string",
  "confirmed_price": "string",
  "confirmed_quantity": "string",
  "import_batch_id": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<TradeConfirmationResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/fill`

Records a fill against an execution that belongs to the resolved portfolio (Portfolio V2).

**Permission:** INVESTMENT_EXECUTION_MANAGE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| executionId | path | string | Yes | Execution UUID |

#### Request body

Schema: `FillExecutionRequest`.

```json
{
  "broker_reference": "string",
  "executed_amount": "string",
  "executed_quantity": "string",
  "execution_price": "string",
  "status": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<ExecutionResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### GET `/api/v2/portfolios/{portfolioCode}/holdings`

List current holdings for a portfolio, resolved by business code (Portfolio V2).

**Permission:** INVESTMENT_PORTFOLIO_VIEW

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<array[HoldingResponse]> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### GET `/api/v2/portfolios/{portfolioCode}/transactions`

List transactions for a portfolio with optional instrument and date filters, resolved by business code (Portfolio V2).

**Permission:** INVESTMENT_PORTFOLIO_VIEW

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| instrument_id | query | string | No | Instrument UUID |
| from | query | string | No | Start business date (YYYY-MM-DD) |
| to | query | string | No | End business date (YYYY-MM-DD) |
| page | query | integer | No | Page number (default 1) |
| limit | query | integer | No | Page size (default 50, max 200) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<TransactionListResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/transactions`

Post a buy, sell, cash, or other portfolio transaction into the ledger, resolved by business code (Portfolio V2). Request body must not include fund_id or contract_id.

**Permission:** INVESTMENT_LEDGER_POST

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

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
  "side": "string",
  "source_decision_id": "string",
  "source_execution_id": "string",
  "transaction_type": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<TransactionResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/transactions/simulate`

Run the post preconditions and pre-trade compliance checks, then preview ledger cash and position impact without mutating investment tables, resolved by business code (Portfolio V2).

**Permission:** INVESTMENT_LEDGER_SIMULATE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

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
  "side": "string",
  "source_decision_id": "string",
  "source_execution_id": "string",
  "transaction_type": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<TransactionSimulationResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### POST `/api/v2/portfolios/{portfolioCode}/transactions/{transactionId}/reverse`

Post a reversal transaction for an existing portfolio transaction, resolved by business code (Portfolio V2).

**Permission:** INVESTMENT_LEDGER_REVERSE

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| transactionId | path | string | Yes | Transaction UUID |

#### Request body

Schema: `ReverseTransactionRequest`.

```json
{
  "business_date": "string",
  "force_post": false,
  "reason": "string"
}
```

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 201 | Created | SuccessResponse<TransactionResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 422 | Unprocessable Entity | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### GET `/api/v2/portfolios/{portfolioCode}/valuations`

List valuation snapshots for a portfolio, resolved by business code (Portfolio V2).

**Permission:** INVESTMENT_VALUATION_VIEW

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |
| from | query | string | No | Start date (YYYY-MM-DD) |
| to | query | string | No | End date (YYYY-MM-DD) |
| page | query | integer | No | Page number (default 1) |
| limit | query | integer | No | Page size (default 50, max 200) |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<ValuationListResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

### GET `/api/v2/portfolios/{portfolioCode}/valuations/latest`

Retrieve the latest internal valuation snapshot for a portfolio, resolved by business code (Portfolio V2).

**Permission:** INVESTMENT_VALUATION_VIEW

**Business rule:** The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again for nested decisions, executions, confirmations, transactions, compliance records, and valuations.

#### Parameters

| Name | Location | Type | Required | Description |
| --- | --- | --- | --- | --- |
| portfolioCode | path | string | Yes | Portfolio code |

#### Request body

No request body.

#### Responses

| HTTP status | Description | Runtime schema |
| --- | --- | --- |
| 200 | OK | SuccessResponse<ValuationResponse> |
| 400 | Bad Request | ErrorResponse |
| 401 | Unauthorized | ErrorResponse |
| 403 | Forbidden | ErrorResponse |
| 404 | Not Found | ErrorResponse |
| 409 | Conflict | ErrorResponse |
| 500 | Internal Server Error | ErrorResponse |

**Source:** `backend/internal/investment/module.go:RegisterRoutesV2`

## Source references

- `backend/cmd/server/main.go`
- `backend/internal/investment/module.go:RegisterRoutesV2`
- `backend/internal/investment/transport/handler/portfolio_v2_handler.go`
- `backend/internal/investment/transport/handler/portfolio_v2_ledger_handler.go`
- `backend/internal/investment/transport/handler/portfolio_v2_decision_handler.go`
- `backend/internal/investment/transport/handler/portfolio_v2_execution_handler.go`
- `backend/internal/investment/transport/handler/portfolio_v2_compliance_handler.go`
- `backend/docs/v2/v2_swagger.json`

## Notes

V2 is additive; V1 remains mounted. Only the 25 endpoints listed here are implemented. The broader `portfolio-v2-api-ddd.md` file is a design reference and includes routes that are still future scope.
