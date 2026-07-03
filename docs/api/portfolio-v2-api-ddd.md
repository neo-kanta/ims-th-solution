# Portfolio V2 API - DDD Design

**Status:** target design, not current implementation  
**Primary bounded context:** Portfolio Management  
**Base path:** `/api/v2/portfolios`  
**Design decision:** public API routes use `portfolioCode`; `portfolio_id` remains the internal operational and accounting source of truth.

## 1. Purpose

Portfolio V2 simplifies the investment API around the thing the application actually manages: the portfolio.

In V2:

- A portfolio can exist without a fund.
- Public URLs identify a portfolio by `code`, not UUID.
- Operational records point to `portfolio_id`.
- `contract_id` is removed from investment APIs.
- `fund_id` is removed from operational write requests.
- Fund/product/unit issuance is a later bounded context, not part of portfolio ledger truth.

## 2. DDD Boundary

### Portfolio Management Bounded Context

Owns:

- Portfolio master data and lifecycle.
- Portfolio type: `LIVE`, `SIMULATION`, `MODEL`.
- Ledger transactions.
- Position and cash projections.
- Portfolio valuation snapshots.
- Portfolio decisions, executions, and trade confirmations.
- Portfolio-scoped compliance inputs and watchlist links.

Does not own:

- Fund unit issuance/redemption.
- Legal fund share-class/unit-class accounting.
- Official fund NAV per unit.
- Generic approval routing internals.
- Generic workflow state machine internals.

### Optional Fund/Product Bounded Context

Future scope. Owns:

- Fund wrapper metadata.
- Mapping one fund to one or many portfolios.
- Unit classes.
- Fund subscription/redemption ledger.
- Official fund NAV per unit.

V2 portfolio APIs must not depend on this context.

## 3. Core Aggregate Model

### Aggregate Roots

| Aggregate | Root ID | Responsibility |
|---|---|---|
| Portfolio | `portfolio_id` | Lifecycle, type, ownership, currencies, manager, policy metadata |
| PortfolioTransaction | `transaction_id` | Immutable financial event posted to one portfolio |
| InvestmentDecision | `decision_id` | Proposed order or rebalance for one portfolio |
| Execution | `execution_id` | Broker execution/fill record for one decision |
| TradeConfirmation | `confirmation_id` | Broker-confirmation review for one execution |
| ValuationSnapshot | `valuation_id` | Reproducible point-in-time portfolio valuation |

### Entities And Projections

| Entity / Projection | Key | Mutability |
|---|---|---|
| PortfolioPosition | `(portfolio_id, instrument_id)` | Mutable projection from immutable transactions |
| CashBalance | `(portfolio_id, currency)` | Mutable projection from immutable cash movements |
| CashMovement | `cash_movement_id` | Immutable |
| ValuationHoldingLine | `valuation_line_id` | Immutable child of valuation snapshot |

### Value Objects

| Value Object | Values / Notes |
|---|---|
| PortfolioType | `LIVE`, `SIMULATION`, `MODEL` |
| PortfolioStatus | `DRAFT`, `PENDING_APPROVAL`, `ACTIVE`, `PAUSED`, `SUSPENDED`, `CLOSED`, `REJECTED` |
| TransactionType | `BUY`, `SELL`, `CASH_IN`, `CASH_OUT`, `FEE`, `DIVIDEND`, `REVERSAL` |
| ValuationSource | `INTERNAL`, `PAM`, `EXTERNAL` |
| OrderSide | `BUY`, `SELL` |

`SUBSCRIPTION` and `REDEMPTION` are intentionally excluded from the portfolio transaction type set. A portfolio buying a mutual fund is `BUY` of a mutual-fund instrument. Own-fund issuance/redemption belongs to the future Fund/Product bounded context.

## 4. Portfolio Types

| Type | Meaning | Ledger Writes | Workflow | Valuation |
|---|---|---|---|---|
| `LIVE` | Real managed portfolio | Allowed | Required | Official portfolio accounting snapshot |
| `SIMULATION` | Paper / what-if portfolio | Allowed, non-official | Optional or bypassed by policy | Indicative only |
| `MODEL` | Target allocation template | No real ledger writes | Not required | No official valuation |

Rules:

- `LIVE` portfolios participate in normal approval, compliance, and close workflow.
- `SIMULATION` portfolios may run order previews and hypothetical ledgers, but must never feed official AUM/NAV reporting.
- `MODEL` portfolios store target weights or strategy lines, not real cash or positions.

## 5. Identity Rules

### Public API Rule

All V2 portfolio-scoped endpoints use `portfolioCode` as the route identity:

```text
/api/v2/portfolios/{portfolioCode}/...
```

`portfolioCode` is the human-readable business identifier users recognize, for example `TH-EQ-01`. The API layer resolves `portfolioCode` to internal `portfolio_id` before invoking application services.

Rules:

- `portfolioCode` must be unique among active/non-deleted portfolios.
- `portfolioCode` should be immutable after activation. If the business needs a code rename later, add a code-alias/redirect table instead of breaking old links.
- Application services, database FKs, projections, and domain events continue to use `portfolio_id`.
- API responses may include `id` for internal correlation, but frontend routes and visible labels must use `code` and `name`.

### Removed From V2 API Requests

| Field | V2 Decision |
|---|---|
| `contract_id` | Removed from investment/portfolio APIs |
| `fund_id` | Removed from operational write requests |

### Transitional Database Rule

During migration, existing V1 columns may remain, but new code must derive them from the portfolio row and never trust the client.

Pre-V2 hardening constraints:

```sql
ALTER TABLE investment__portfolios
ADD CONSTRAINT uq_inv_portfolios_id_fund UNIQUE (id, fund_id);

ALTER TABLE investment__portfolio_transactions
ADD CONSTRAINT fk_inv_txn_portfolio_fund
FOREIGN KEY (portfolio_id, fund_id)
REFERENCES investment__portfolios(id, fund_id);

ALTER TABLE investment__decisions
ADD CONSTRAINT chk_inv_decisions_contract_is_fund
CHECK (contract_id = fund_id);
```

Apply the same pattern to executions and trade confirmations until `fund_id` and `contract_id` are removed.

## 6. API Resources

### Endpoint Summary

| Method | Path | Purpose | Aggregate |
|---|---|---|---|
| `GET` | `/api/v2/portfolios` | List visible portfolios | Portfolio |
| `POST` | `/api/v2/portfolios` | Create portfolio | Portfolio |
| `GET` | `/api/v2/portfolios/{portfolioCode}` | Get portfolio detail | Portfolio |
| `PATCH` | `/api/v2/portfolios/{portfolioCode}` | Update portfolio metadata | Portfolio |
| `POST` | `/api/v2/portfolios/{portfolioCode}/activate` | Activate approved portfolio | Portfolio |
| `POST` | `/api/v2/portfolios/{portfolioCode}/pause` | Pause trading | Portfolio |
| `POST` | `/api/v2/portfolios/{portfolioCode}/close` | Close portfolio | Portfolio |
| `GET` | `/api/v2/portfolios/{portfolioCode}/holdings` | Current positions | Position projection |
| `GET` | `/api/v2/portfolios/{portfolioCode}/cash` | Current cash | Cash projection |
| `GET` | `/api/v2/portfolios/{portfolioCode}/transactions` | Ledger history | PortfolioTransaction |
| `POST` | `/api/v2/portfolios/{portfolioCode}/transactions` | Post ledger transaction | PortfolioTransaction |
| `POST` | `/api/v2/portfolios/{portfolioCode}/transactions/simulate` | Preview ledger impact | PortfolioTransaction |
| `POST` | `/api/v2/portfolios/{portfolioCode}/transactions/{transactionId}/reverse` | Reverse immutable post | PortfolioTransaction |
| `GET` | `/api/v2/portfolios/{portfolioCode}/valuations` | List valuation snapshots | ValuationSnapshot |
| `POST` | `/api/v2/portfolios/{portfolioCode}/valuations/run` | Run portfolio valuation | ValuationSnapshot |
| `GET` | `/api/v2/portfolios/{portfolioCode}/decisions` | List decisions | InvestmentDecision |
| `POST` | `/api/v2/portfolios/{portfolioCode}/decisions` | Create decision | InvestmentDecision |
| `GET` | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}` | Decision detail | InvestmentDecision |
| `PATCH` | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}` | Update draft decision | InvestmentDecision |
| `POST` | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/submit` | Submit for compliance and approval | InvestmentDecision |
| `POST` | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/cancel` | Cancel decision | InvestmentDecision |
| `POST` | `/api/v2/portfolios/{portfolioCode}/decisions/{decisionId}/executions` | Create execution | Execution |
| `POST` | `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/fill` | Fill execution | Execution |
| `POST` | `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/cancel` | Cancel execution | Execution |
| `POST` | `/api/v2/portfolios/{portfolioCode}/executions/{executionId}/confirmations` | Record broker confirmation | TradeConfirmation |
| `POST` | `/api/v2/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve` | Match/review confirmation | TradeConfirmation |

## 7. DTO Contracts

### PortfolioResponse

```json
{
  "id": "uuid",
  "portfolio_type": "LIVE",
  "code": "TH-EQ-01",
  "name": "Thailand Equity Portfolio",
  "base_currency": "THB",
  "valuation_currency": "THB",
  "manager_user": {
    "user_id": "uuid",
    "display_name": "Portfolio Manager"
  },
  "benchmark": "SET TRI",
  "risk_profile": "HIGH",
  "status": "ACTIVE",
  "version": 3,
  "created_at": "2026-07-03T00:00:00Z",
  "updated_at": "2026-07-03T00:00:00Z"
}
```

Route identity is `code`. The `id` field is returned for internal correlation and database-backed operations only.

### CreatePortfolioRequest

```json
{
  "portfolio_type": "LIVE",
  "code": "TH-EQ-01",
  "name": "Thailand Equity Portfolio",
  "base_currency": "THB",
  "valuation_currency": "THB",
  "manager_user_id": "uuid",
  "benchmark": "SET TRI",
  "risk_profile": "HIGH",
  "inception_date": "2026-07-03"
}
```

Validation:

- `portfolio_type` defaults to `LIVE`.
- `code` must be unique among non-deleted portfolios.
- `base_currency` and `valuation_currency` must be ISO 4217 uppercase codes.
- `MODEL` portfolios may omit manager and inception date only if product accepts template-style records.

### PostTransactionRequest

```json
{
  "transaction_type": "BUY",
  "side": "BUY",
  "instrument_id": "uuid",
  "quantity": "1000",
  "price": "12.50",
  "currency": "THB",
  "fees": "10.00",
  "gross_amount": "12500.00",
  "net_amount": "-12510.00",
  "fx_rate_to_base": "1",
  "business_date": "2026-07-03",
  "settlement_date": "2026-07-05",
  "external_ref": "BROKER-001",
  "reason": "Initial buy"
}
```

Validation:

- `BUY` / `SELL` require `instrument_id`, `quantity`, and `price`.
- `CASH_IN`, `CASH_OUT`, `FEE`, and `DIVIDEND` must not carry `instrument_id`, `quantity`, or `price`.
- `REVERSAL` is not posted through this endpoint.
- `SIMULATION` portfolios return `is_official=false` in the response.
- `MODEL` portfolios reject ledger posts.

### TransactionResponse

```json
{
  "id": "uuid",
  "portfolio_id": "uuid",
  "transaction_type": "BUY",
  "instrument_id": "uuid",
  "quantity": "1000",
  "price": "12.50",
  "currency": "THB",
  "gross_amount": "12500.00",
  "fees": "10.00",
  "net_amount": "-12510.00",
  "realised_pnl_base": "0",
  "business_date": "2026-07-03",
  "settlement_date": "2026-07-05",
  "status": "POSTED",
  "is_official": true,
  "created_at": "2026-07-03T00:00:00Z",
  "created_by": "uuid"
}
```

### CreateDecisionRequest

```json
{
  "decision_type": "SINGLE_ORDER",
  "instrument_id": "uuid",
  "instrument_code": "PTT",
  "side": "BUY",
  "quantity": "1000",
  "amount": null,
  "limit_price": "35.00",
  "currency": "THB",
  "business_date": "2026-07-03",
  "rationale": "Portfolio rebalance",
  "research_report_id": "uuid"
}
```

Rules:

- No `fund_id`.
- No `contract_id`.
- The server loads the portfolio and determines any external context needed by approval, compliance, or workflow.
- `BASKET_ORDER`, `REBALANCE`, and `SWITCH` use `lines`.

### RunValuationRequest

```json
{
  "business_date": "2026-07-03",
  "fx_rates": {
    "USD->THB": "36.50"
  },
  "mode": "OFFICIAL"
}
```

Rules:

- `mode=OFFICIAL` is valid only for `LIVE`.
- `mode=INDICATIVE` is valid for `LIVE` and `SIMULATION`.
- `MODEL` portfolios do not produce accounting valuation snapshots.
- No `total_units`; fund/unit NAV is not part of Portfolio V2.

## 8. Application Services

### Commands

Route handlers resolve `portfolioCode` to `portfolio_id` before creating commands. Command handlers should receive the internal UUID so domain logic and database constraints stay stable.

| Command | Input Anchor | Notes |
|---|---|---|
| CreatePortfolio | none | Creates Portfolio aggregate |
| UpdatePortfolio | `portfolio_id` | Optimistic version check |
| PostPortfolioTransaction | `portfolio_id` | Loads portfolio, applies type/status guards |
| SimulatePortfolioTransaction | `portfolio_id` | No mutation |
| ReversePortfolioTransaction | `portfolio_id`, `transaction_id` | Reversal row, never update original |
| RunPortfolioValuation | `portfolio_id` | Reproducible valuation snapshot |
| CreateDecision | `portfolio_id` | No fund/contract input |
| SubmitDecision | `portfolio_id`, `decision_id` | Calls compliance, workflow, approval ports |
| CreateExecution | `portfolio_id`, `decision_id` | Decision must belong to portfolio |
| ResolveConfirmation | `portfolio_id`, `confirmation_id` | Confirmation must belong to portfolio |

### Queries

| Query | Output |
|---|---|
| GetPortfolio | Portfolio descriptor and allowed actions |
| ListPortfolios | User-visible portfolio directory |
| GetHoldings | Current position projection |
| GetCash | Current cash projection |
| ListTransactions | Immutable transaction list |
| GetLatestValuation | Latest valuation by source/mode |
| ListPortfolioTimeline | Decisions, executions, confirmations, transactions |

## 9. Domain Events

Events are internal application events unless promoted to an event bus later.

| Event | Emitted When | Consumers |
|---|---|---|
| `PortfolioCreated` | Portfolio created | Audit, notification |
| `PortfolioActivated` | Portfolio becomes active | Workflow, audit |
| `PortfolioTransactionPosted` | Ledger post succeeds | Audit, compliance post-trade, projections |
| `PortfolioTransactionReversed` | Reversal succeeds | Audit, projections |
| `PortfolioValuationRun` | Valuation snapshot inserted | Audit, dashboards |
| `DecisionSubmitted` | Decision enters approval/compliance | Approval, notification |
| `ExecutionFilled` | Execution fill recorded | Confirmation workflow |
| `ConfirmationResolved` | Confirmation matched/reviewed | Workflow close gate |

## 10. Ports And Adapters

| Port | Direction | Purpose |
|---|---|---|
| `PortfolioPermissionChecker` | inbound guard | User can view/manage portfolio |
| `ComplianceChecker` | outbound | Evaluate proposed order using portfolio context |
| `ApprovalSubmitter` | outbound | Submit decision/portfolio subject |
| `WorkflowGate` | outbound | Check trade day state for portfolio |
| `MarketPriceProvider` | outbound | Read quotes for valuation/intraday |
| `AuditLogger` | outbound | Record immutable user action |

Approval and workflow should receive a portfolio subject where possible:

```json
{
  "subject_type": "PORTFOLIO_DECISION",
  "subject_id": "decision_uuid",
  "portfolio_id": "portfolio_uuid"
}
```

If a legacy external contract identifier is still required by another module, create an anti-corruption adapter that derives it from the portfolio. Do not expose it in Portfolio V2 API payloads.

## 11. Permissions

Recommended V2 permissions:

| Permission | Applies To |
|---|---|
| `PORTFOLIO_VIEW` | Portfolio list/detail, holdings, cash, valuations |
| `PORTFOLIO_MANAGE` | Portfolio create/update/lifecycle |
| `PORTFOLIO_LEDGER_POST` | Ledger posts |
| `PORTFOLIO_LEDGER_SIMULATE` | Simulations |
| `PORTFOLIO_LEDGER_REVERSE` | Reversals |
| `PORTFOLIO_VALUATION_RUN` | Official/indicative valuation run |
| `PORTFOLIO_DECISION_MANAGE` | Decision draft/update |
| `PORTFOLIO_DECISION_SUBMIT` | Submit decision |
| `PORTFOLIO_EXECUTION_MANAGE` | Execution create/fill/cancel |
| `PORTFOLIO_CONFIRMATION_MANAGE` | Confirmation record/resolve/import |

Data permission is portfolio-scoped. Fund/product permission can be added later as a separate scope.

## 12. Migration Checklist

### Phase 0 - Hardening

- Add preflight SQL to find drift between `portfolio_id`, `fund_id`, and `contract_id`.
- Add temporary constraints to prevent new drift.
- Update write paths to derive old columns from portfolio.

### Phase 1 - API V2 Additive

- Add V2 routes under `/api/v2/portfolios`.
- Add `portfolio_type` to portfolios.
- Add V2 DTOs with no `fund_id` or `contract_id`.
- Keep V1 routes available.
- Make V2 application services call existing repositories through portfolio-first request models.

### Phase 2 - Frontend Cutover

- Move screens from fund-first workspace to portfolio workspace.
- Replace fund selectors with portfolio selectors.
- Keep fund labels as optional context only.

### Phase 3 - Schema Simplification

- Drop `contract_id` from investment operational tables.
- Drop `fund_id` from investment operational tables once all reads derive context from portfolio.
- Remove `SUBSCRIPTION` and `REDEMPTION` from portfolio transaction type set.
- Remove or ignore portfolio `has_units` and `investment__nav_snapshots` until Fund/Product V2 exists.

## 13. Preflight SQL

Find operational rows whose portfolio belongs to another fund:

```sql
SELECT t.id, t.portfolio_id, t.fund_id, p.fund_id AS portfolio_fund_id
FROM investment__portfolio_transactions t
JOIN investment__portfolios p ON p.id = t.portfolio_id
WHERE t.fund_id <> p.fund_id;
```

Repeat the same pattern for decisions, executions, and trade confirmations:

```sql
SELECT d.id, d.portfolio_id, d.fund_id, d.contract_id, p.fund_id AS portfolio_fund_id
FROM investment__decisions d
JOIN investment__portfolios p ON p.id = d.portfolio_id
WHERE d.fund_id <> p.fund_id
   OR d.contract_id <> d.fund_id;
```

Find own-fund-style transaction types that need classification before removal:

```sql
SELECT transaction_type, COUNT(*)
FROM investment__portfolio_transactions
WHERE transaction_type IN ('SUBSCRIPTION', 'REDEMPTION')
GROUP BY transaction_type;
```

Find portfolio NAV/unit usage:

```sql
SELECT p.id, p.code, p.has_units, COUNT(n.id) AS nav_count
FROM investment__portfolios p
LEFT JOIN investment__nav_snapshots n ON n.portfolio_id = p.id
GROUP BY p.id, p.code, p.has_units
HAVING p.has_units = true OR COUNT(n.id) > 0;
```

## 14. Acceptance Criteria

- V2 write requests do not contain `fund_id` or `contract_id`.
- V2 route URLs use `portfolioCode`, not raw UUID.
- Every operational row can be resolved from `portfolio_id`.
- `LIVE`, `SIMULATION`, and `MODEL` portfolios have distinct behavior.
- Simulation output is never reported as official valuation.
- Fund/unit issuance is absent from the portfolio ledger.
- V1 compatibility remains available until frontend cutover is complete.

## Source References

- `docs/api/portfolio-api.md`
- `docs/api/investment-api.md`
- `database/migrations/20260428000002_investment__create_funds_portfolios.up.sql`
- `database/migrations/20260428000004_investment__create_ledger.up.sql`
- `database/migrations/20260428000005_investment__create_pricing_valuation.up.sql`
- `database/migrations/20260601000001_investment__create_decisions_executions_confirmations.up.sql`
- `backend/internal/investment/application`
- `backend/internal/investment/domain`
