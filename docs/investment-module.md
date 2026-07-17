# Investment Module — Function Description

---

## Document Control

| Field | Value |
|---|---|
| **Version** | 1.0.0 |
| **Date** | 2026-06-24 |
| **Author** | Systemweb / IMS Documentation |
| **Status** | Reviewed |
| **Source** | `backend/internal/investment/`, `frontend/app/features/investment/`, `database/migrations/20260601*`, `docs/handoff/investment-p0-foundation-fixes.md`, `docs/runbook/investment.md` |

---

## Table of Contents

1. [Module Overview](#1-module-overview)
2. [Business Purpose](#2-business-purpose)
3. [Functional Scope](#3-functional-scope)
4. [Out of Scope](#4-out-of-scope)
5. [User Roles and Permissions](#5-user-roles-and-permissions)
6. [Workflow / User Flow](#6-workflow--user-flow)
7. [Function Descriptions](#7-function-descriptions)
   - 7.1 [Research Report (Analysis)](#71-research-report-analysis)
   - 7.2 [Investment Decision](#72-investment-decision)
   - 7.3 [Decision Submission and Compliance Gate](#73-decision-submission-and-compliance-gate)
   - 7.4 [Decision Approval](#74-decision-approval)
   - 7.5 [Execution](#75-execution)
   - 7.6 [Trade Confirmation](#76-trade-confirmation)
   - 7.7 [Ledger Operations](#77-ledger-operations)
   - 7.8 [Valuation and AUM](#78-valuation-and-aum)
   - 7.9 [Fund Management](#79-fund-management)
   - 7.10 [Portfolio Management](#710-portfolio-management)
   - 7.11 [Instrument Management](#711-instrument-management)
   - 7.12 [Batch Approval](#712-batch-approval)
8. [API / Backend Behavior](#8-api--backend-behavior)
9. [Frontend Behavior](#9-frontend-behavior)
10. [Database / Data Model Summary](#10-database--data-model-summary)
11. [Validation Rules](#11-validation-rules)
12. [Approval and Permission Rules](#12-approval-and-permission-rules)
13. [Integration Points](#13-integration-points)
14. [Audit and Logging](#14-audit-and-logging)
15. [Notifications](#15-notifications)
16. [Error / Exception Handling](#16-error--exception-handling)
17. [Known Gaps / TODOs](#17-known-gaps--todos)
18. [Acceptance Checklist](#18-acceptance-checklist)

---

## 1. Module Overview

**Module name:** `investment`

**Package path:** `backend/internal/investment/`

**API base path:** `/api/v1/investment/`

**Frontend feature path:** `frontend/app/features/investment/`

The Investment module is the primary operational module for managing the full lifecycle of investment activities — from fund onboarding and instrument management through research analysis, investment decisions, trade execution, and settlement confirmation. It integrates with the Compliance module (via contract adapter) for pre-trade and post-trade checks, and with the Approval module for maker-checker workflows on research reports and investment decisions.

> **Note:** The module README at `backend/internal/investment/README.md` is outdated. It describes the module as "scaffolded only." The actual implementation is fully functional and production-quality across all four investment process stages.

---

## 2. Business Purpose

The Investment module provides a structured, auditable, and compliant investment process for Thai asset management operations. It ensures that:

- Every investment trade is preceded by a compliance check and an approval workflow.
- Research analysis can optionally be linked to investment decisions, with configurable requirements per fund.
- Portfolio positions and cash are maintained in an append-only ledger (immutable entries) for regulatory auditability.
- Net Asset Value (NAV) is computed from authoritative ledger positions and can be compared to live intraday quotes.
- Executions and trade confirmations provide end-to-end traceability from decision to settlement.
- All user actions on decisions, executions, and confirmations are captured in the audit trail.

---

## 3. Functional Scope

| Capability | Status |
|---|---|
| Fund management (create, configure, list, detail) | Active |
| Portfolio management (create, list, holdings, cash) | Active |
| Instrument management (create, update, price posting) | Active |
| Research report lifecycle (create, submit, invalidate) | Active |
| Investment decision lifecycle (DRAFT → EXECUTED) | Active |
| Pre-trade compliance check at decision submission | Active |
| PENDING_COMPLIANCE_RELEASE flow (BLOCK + overridable) | Active |
| Approval workflow for investment decisions | Active |
| Execution records (PENDING → EXECUTED) | Active |
| Trade confirmation records (import and review) | Active |
| Ledger posting, simulation, and force-post | Active |
| Locked-day reversal | Active |
| Valuation run (NAV computation) | Active |
| AUM computation | Active |
| Intraday valuation with live market quotes | Active |
| Batch approval for investment decisions | Active |
| Contract adapters exposing investment subjects to approval engine | Active |
| Fund dashboard tabs (compliance, holdings, operations, reviewers, settings, stages, audit) | Active |

---

## 4. Out of Scope

| Item | Notes |
|---|---|
| PORTFOLIO approval subject | Wired as subject type but `INVESTMENT_PORTFOLIO` approval flows are deferred to Phase 2 — registered but not triggered |
| Per-instrument-line compliance checks for BASKET_ORDER / REBALANCE / SWITCH | Only header-level compliance check is run at submission. Phase 2 item |
| Watchlist / threshold monitoring | Separate planned module; an API design document exists at `docs/api/watchlist-api.md` but no backend implementation |
| Automated intraday position refresh | Intraday valuation is on-demand only (triggered by UI); no scheduled auto-refresh |
| Multi-currency NAV | Current valuation is single-currency per fund |
| Real-time market data streaming | Quotes are point-in-time fetches from Alpha Vantage or equivalent provider |
| Leave and holiday calendar enforcement | `NopCalendarAdapter` always returns nil; no trading calendar blocking |
| Trade history-based rules | `NopTradeHistoryAdapter` always returns empty; rules depending on trade history cannot function |
| Instrument soft-delete recovery | Deleted instruments cannot be restored via API |

---

## 5. User Roles and Permissions

The investment module uses 28 permission codes. All are seeded into the `Admin` group by default.

### 5.1 Fund and Portfolio Permissions

| Permission Code | Description |
|---|---|
| `INVESTMENT_FUND_VIEW` | View fund list and detail |
| `INVESTMENT_FUND_MANAGE` | Create and configure funds |
| `INVESTMENT_PORTFOLIO_VIEW` | View portfolio list, holdings, and cash |
| `INVESTMENT_PORTFOLIO_MANAGE` | Create and update portfolios |
| `INVESTMENT_INSTRUMENT_VIEW` | View instrument catalog |
| `INVESTMENT_INSTRUMENT_MANAGE` | Create, update, and manage instruments |
| `INVESTMENT_REFERENCE_VIEW` | View reference data (classifications, currencies) |

### 5.2 Ledger and Valuation Permissions

| Permission Code | Description |
|---|---|
| `INVESTMENT_LEDGER_VIEW` | Read ledger entries and transaction history |
| `INVESTMENT_LEDGER_POST` | Post a standard ledger transaction |
| `INVESTMENT_LEDGER_SIMULATE` | Run a dry-run ledger post (no persistence) |
| `INVESTMENT_LEDGER_FORCE_POST` | Override workflow day gate and force-post on a locked day |
| `INVESTMENT_LEDGER_REVERSE` | Reverse a posted ledger entry on a locked day |
| `INVESTMENT_VALUATION_VIEW` | View NAV and valuation records |
| `INVESTMENT_VALUATION_RUN` | Trigger a valuation run |
| `INVESTMENT_PRICE_POST` | Post a price snapshot for an instrument |
| `INVESTMENT_FUND_AUM_COMPUTE` | Trigger AUM computation |

### 5.3 Research Report Permissions

| Permission Code | Description |
|---|---|
| `INVESTMENT_RESEARCH_VIEW` | View research reports |
| `INVESTMENT_RESEARCH_CREATE` | Create a new research report |
| `INVESTMENT_RESEARCH_UPDATE` | Update a draft or not-yet-submitted report |
| `INVESTMENT_RESEARCH_DELETE` | Soft-delete a draft report |
| `INVESTMENT_RESEARCH_SUBMIT` | Submit a research report for approval |
| `INVESTMENT_RESEARCH_CANCEL_SUBMIT` | Retract a submitted report back to DRAFT |
| `INVESTMENT_RESEARCH_INVALIDATE` | Invalidate a completed report |

### 5.4 Decision Permissions

| Permission Code | Description |
|---|---|
| `INVESTMENT_DECISION_VIEW` | View decision list and detail |
| `INVESTMENT_DECISION_MANAGE` | Create, update, and manage decisions |
| `INVESTMENT_DECISION_SUBMIT` | Submit a decision for compliance check and approval |
| `INVESTMENT_DECISION_CANCEL` | Cancel a decision in an eligible state |
| `INVESTMENT_DECISION_APPROVE` | Approve or reject an investment decision (approval engine role) |
| `INVESTMENT_COMPLIANCE_RELEASE_APPROVE` | Approve or reject a COMPLIANCE_RELEASE request (compliance officer role) |

### 5.5 Execution and Confirmation Permissions

| Permission Code | Description |
|---|---|
| `INVESTMENT_EXECUTION_VIEW` | View execution records |
| `INVESTMENT_EXECUTION_MANAGE` | Create and update executions |
| `INVESTMENT_CONFIRMATION_VIEW` | View trade confirmation records |
| `INVESTMENT_CONFIRMATION_MANAGE` | Update trade confirmation records |
| `INVESTMENT_CONFIRMATION_IMPORT` | Import trade confirmation files (CSV/XML) |

---

## 6. Workflow / User Flow

### 6.1 Four-Stage Investment Process

```
Stage 1: Analysis (Research)
  └── Research reports support decisions but are OPTIONAL by default.
      Mandatory only when fund.RequireResearchReportForDecision = true.

Stage 2: Decision
  └── Portfolio manager creates a decision, attaches it to a fund/portfolio,
      specifies instrument, side, quantity/amount, and optionally links a report.
      Decision is submitted → compliance check → approval workflow.

Stage 3: Execution
  └── Approved and ready decision is executed via the broker/exchange.
      Execution record captures fill price, quantity, fees, and broker reference.

Stage 4: Trade Confirmation / Review
  └── Settlement confirmation matched against execution.
      Confirmation status drives ledger posting and position settlement.
```

### 6.2 Research Report Lifecycle

```
(Draft Created)
  │
  ▼ DRAFT
  │  CanUpdate: true
  │  CanDelete: true
  │  CanSubmit: true
  │
  ▼ POST .../submit → SUBMITTED
  │  CanCancelSubmit: true
  │
  ├── Approval engine: RESEARCH_REPORT approval flow
  │   Approved → REVIEW_COMPLETED
  │   Rejected → DRAFT (or REJECTED if terminal)
  │
  └── INVALIDATED (from REVIEW_COMPLETED only — report superseded)
```

### 6.3 Decision Lifecycle (Full State Machine)

```
DRAFT
  │
  ▼ POST .../submit (requires INVESTMENT_DECISION_SUBMIT)
  │
  ├── [Workflow gate: IsTradeAllowed]  ──→ rejected if workflow day closed
  │
  ├── [Fund report gate]  ──→ rejected if fund.RequireResearchReportForDecision
  │                              and no valid report linked
  │
  ├── [Report reference policy]  ──→ validation of linked report scope,
  │                                   side, status, and date if report linked
  │
  ├── [Compliance check: IRG pre-trade]
  │   │
  │   ├── PASS / WARN  ──→  PENDING_APPROVAL
  │   │
  │   ├── BLOCK + all breaches overridable
  │   │   └── PENDING_COMPLIANCE_RELEASE
  │   │       (awaiting compliance officer approval of COMPLIANCE_RELEASE)
  │   │
  │   └── BLOCK + any breach non-overridable
  │       └── ErrComplianceRejected (HTTP 422, terminal)
  │
PENDING_COMPLIANCE_RELEASE
  │
  ├── Compliance officer approves COMPLIANCE_RELEASE
  │   ├── Re-validates workflow gate + report reference
  │   └── PENDING_APPROVAL
  │
  └── Compliance officer rejects COMPLIANCE_RELEASE
      └── CANCELLED

PENDING_APPROVAL
  │
  ├── Approver approves INVESTMENT_DECISION
  │   └── APPROVED
  │
  └── Approver rejects INVESTMENT_DECISION
      └── REJECTED (terminal)

APPROVED
  │
  └── Portfolio manager marks ready
      └── READY_FOR_EXECUTION

READY_FOR_EXECUTION
  │
  └── Execution posted against the decision
      └── EXECUTED (terminal)

CANCELLED (terminal)   — from DRAFT or PENDING_COMPLIANCE_RELEASE
REJECTED  (terminal)   — from PENDING_APPROVAL
```

### 6.4 Execution Lifecycle

```
(Execution created against READY_FOR_EXECUTION decision)
  │
  PENDING
  │
  ├── Full fill  ──→  EXECUTED
  ├── Partial fill  ──→  PARTIALLY_EXECUTED
  └── Cancel  ──→  CANCELLED
```

### 6.5 Trade Confirmation Lifecycle

```
(Confirmation created or imported)
  │
  PENDING_REVIEW
  │
  ├── Matched against execution  ──→  MATCHED
  ├── Mismatch detected  ──→  MISMATCHED
  └── Manual review completed  ──→  REVIEWED
```

### 6.6 Maker-Checker Constraint

The investment module enforces the maker-checker principle through the Approval module:

- The actor who submits a research report or investment decision **cannot** be an approver for that same record.
- This constraint is enforced by the Approval module's approver assignment rules and cannot be bypassed via the investment API.
- The `INVESTMENT_DECISION_APPROVE` and `INVESTMENT_COMPLIANCE_RELEASE_APPROVE` permissions are distinct and should be held by different personas (investment approver vs. compliance officer).

---

## 7. Function Descriptions

### 7.1 Research Report (Analysis)

**Function Name:** `CreateResearchReport` / `UpdateResearchReport` / `SubmitResearchReport` / `CancelSubmitResearchReport` / `InvalidateResearchReport` / `DeleteResearchReport`

**Function Summary:** Manage the full lifecycle of an investment research report. Reports provide analysis backing for investment decisions and can be mandatorily linked by fund configuration.

**Entry Conditions (Create):**
- `title` must be non-empty.
- `fund_id` must be a valid UUID referencing an existing fund.
- Actor must hold `INVESTMENT_RESEARCH_CREATE`.

**Entry Conditions (Submit):**
- Report must be in `DRAFT` status.
- Actor must hold `INVESTMENT_RESEARCH_SUBMIT`.
- Report must not be deleted or invalidated.

**Entry Conditions (Invalidate):**
- Report must be in `REVIEW_COMPLETED` status.
- Actor must hold `INVESTMENT_RESEARCH_INVALIDATE`.

**Exception Flow:**
- Report not found → HTTP 404 (`ErrResearchReportNotFound`).
- Invalid lifecycle transition → HTTP 409 (`ErrResearchReportLifecycle`).
- Unauthorized → HTTP 403.

**Post-Execution Status (Submit):**
- Report status → `SUBMITTED`.
- Approval request created under subject `RESEARCH_REPORT`.

**Paths:** Under `/investment/research/` (GET list, GET /{id}, POST create, PUT update, POST /{id}/submit, POST /{id}/cancel-submit, POST /{id}/invalidate, DELETE /{id})

---

### 7.2 Investment Decision

**Function Name:** `CreateDecision` / `UpdateDecision`

**Function Summary:** Create or update a draft investment decision specifying a trade on an instrument within a portfolio.

**Entry Conditions (Create):**
- `fund_id`, `portfolio_id`, `instrument_id` must be valid UUIDs.
- `side` must be `BUY` or `SELL`.
- At least one of `quantity` or `amount` must be provided.
- `business_date` must be provided.
- `decision_type` must be a valid enum value: `SINGLE_ORDER`, `BASKET_ORDER`, `REBALANCE`, `SWITCH`.
- Actor must hold `INVESTMENT_DECISION_MANAGE`.

**Exception Flow:**
- Invalid field values → HTTP 400.
- Fund or portfolio not found → HTTP 404.
- Decision not found → HTTP 404 (`ErrDecisionNotFound`).

**Post-Execution Status:**
- Decision record created with status `DRAFT`.

---

### 7.3 Decision Submission and Compliance Gate

**Function Name:** `SubmitDecision`

**Function Summary:** Submit a DRAFT decision for compliance check and approval. This is the principal transition step in the decision lifecycle.

**Entry Conditions:**
- Decision must be in `DRAFT` status.
- `business_date` must not be on a closed workflow day (unless forced).
- Actor must hold `INVESTMENT_DECISION_SUBMIT`.
- `CanSubmit()` must return true on the decision entity.

**Step Sequence (see also §6.3 for full state diagram):**

1. Workflow gate: `IsTradeAllowed` — day must be open.
2. Fund report gate: if `fund.RequireResearchReportForDecision = true`, a valid report must be linked.
3. Report reference policy: validate linked report's scope, side, status, and effective date.
4. IRG pre-trade check via `ComplianceChecker.CheckProposedOrder`.
5. Verdict routing:
   - PASS / WARN → proceed to INVESTMENT_DECISION approval.
   - BLOCK + all overridable → submit COMPLIANCE_RELEASE approval; decision → `PENDING_COMPLIANCE_RELEASE`.
   - BLOCK + any non-overridable → return `ErrComplianceRejected` (HTTP 422).
6. Submit to INVESTMENT_DECISION approval engine → decision → `PENDING_APPROVAL`.

**Exception Flow:**
- Workflow day closed → HTTP 409 (`ErrWorkflowDayClosed`).
- Report required but missing → HTTP 409 (`ErrDecisionLifecycle`).
- Compliance BLOCK + non-overridable → HTTP 422 (`ErrComplianceRejected`).
- Approval engine error → HTTP 500.

**Post-Execution Status:**
- Decision status: `PENDING_COMPLIANCE_RELEASE` or `PENDING_APPROVAL`.
- `ComplianceReleaseApprovalRequestID` or `ApprovalRequestID` stored on the decision entity.

**Path:** `POST /investment/decisions/{id}/submit`

---

### 7.4 Decision Approval

**Function Name:** `ApplyApprovalDecision` / `ApplyComplianceReleaseDecision`

**Function Summary:** Apply the outcome of an approval workflow action to an investment decision. Called internally via the Approval module's subject callback mechanism — not a direct HTTP endpoint.

**`ApplyApprovalDecision` (INVESTMENT_DECISION approval):**
- APPROVED → decision → `APPROVED`.
- REJECTED → decision → `REJECTED`.

**`ApplyComplianceReleaseDecision` (COMPLIANCE_RELEASE approval):**
- APPROVED:
  1. Re-validate workflow gate and report reference.
  2. Auto-submit to INVESTMENT_DECISION approval.
  3. Decision → `PENDING_APPROVAL`.
- REJECTED → decision → `CANCELLED`.

**Post-Execution Status:**
- Decision status updated.
- `ApprovalRequestID` stored when decision transitions to `PENDING_APPROVAL` from compliance release.

---

### 7.5 Execution

**Function Name:** `CreateExecution` / `UpdateExecution` / `MarkReadyForExecution`

**Function Summary:** Record trade execution details after a broker fill is received.

**Entry Conditions (MarkReadyForExecution):**
- Decision must be in `APPROVED` status.
- `CanMarkReadyForExecution()` must be true.
- Actor must hold `INVESTMENT_EXECUTION_MANAGE`.

**Entry Conditions (CreateExecution):**
- Decision must be in `READY_FOR_EXECUTION` status.
- Actor must hold `INVESTMENT_EXECUTION_MANAGE`.

**Exception Flow:**
- Decision not in correct state → HTTP 409 (`ErrDecisionLifecycle`).
- Execution not found → HTTP 404 (`ErrExecutionNotFound`).

**Post-Execution Status:**
- Execution record created with status `PENDING`.
- Decision status transitions to `EXECUTED` upon full execution.

---

### 7.6 Trade Confirmation

**Function Name:** `CreateConfirmation` / `UpdateConfirmation` / `ImportConfirmations`

**Function Summary:** Create, update, or import trade confirmation records. Match confirmations against executions and review exceptions.

**Entry Conditions (Create):**
- `execution_id` must reference a valid execution.
- Actor must hold `INVESTMENT_CONFIRMATION_MANAGE`.

**Entry Conditions (Import):**
- Actor must hold `INVESTMENT_CONFIRMATION_IMPORT`.
- File must be in accepted format (CSV or XML, broker-specific).

**Post-Execution Status:**
- Confirmation record created with status `PENDING_REVIEW`.

---

### 7.7 Ledger Operations

**Function Name:** `PostLedgerEntry` / `SimulateLedgerEntry` / `ForcePostLedgerEntry` / `ReverseLedgerEntry`

**Function Summary:** Post transactions to the portfolio ledger. The ledger is append-only; reversals are new entries with opposite signs.

**Entry Conditions (Standard Post):**
- Workflow day must be open.
- Actor must hold `INVESTMENT_LEDGER_POST`.
- `entry_type`, `portfolio_id`, `amount`, `currency`, and `business_date` required.

**Entry Conditions (Force Post):**
- Actor must hold `INVESTMENT_LEDGER_FORCE_POST`.
- Bypasses the workflow day gate. Used for corrections when day is locked.

**Entry Conditions (Reversal):**
- Actor must hold `INVESTMENT_LEDGER_REVERSE`.
- Workflow day for the entry being reversed must be locked (only locked-day reversals supported).
- New reversal entry created with opposite sign and same instrument/amount.

**Entry Conditions (Simulation):**
- Actor must hold `INVESTMENT_LEDGER_SIMULATE`.
- Dry run: computes projected balance; no records written.

**Exception Flow:**
- Workflow day open for reversal → HTTP 409 (`ErrCannotReverseOpenDay`).
- Insufficient cash/position → HTTP 409 (`ErrInsufficientBalance`).
- Optimistic concurrency conflict on position → HTTP 409 (`ErrConcurrentModification`); caller should retry.

---

### 7.8 Valuation and AUM

**Function Name:** `RunValuation` / `ComputeAUM` / `GetIntradayValuation`

**Function Summary:** Compute authoritative NAV for a fund or portfolio from ledger positions and price snapshots. Intraday valuation fetches live market quotes as a non-authoritative overlay.

**NAV Computation:**
- Reads holdings from `investment__positions`.
- Applies latest price snapshot from `investment__price_snapshots`.
- Stores result in `investment__valuations`.
- Does not read or write live market data tables.
- Actor must hold `INVESTMENT_VALUATION_RUN`.

**AUM Computation:**
- Aggregates NAV across all portfolios in a fund.
- Actor must hold `INVESTMENT_FUND_AUM_COMPUTE`.

**Intraday Valuation:**
- Fetches live quotes via `contract.MarketQuoteProvider` (backed by Alpha Vantage or equivalent).
- Applies live prices to current holdings for a real-time estimate.
- Result is **non-authoritative** — does not mutate `investment__valuations` or `investment__price_snapshots`.
- Used for intraday UI display only.

**Exception Flow:**
- Market quote provider unavailable → HTTP 503 (`ErrMarketDataUnavailable`). See `docs/runbook/investment.md` for Alpha Vantage outage procedures.

---

### 7.9 Fund Management

**Function Name:** `CreateFund` / `UpdateFund` / `GetFund` / `ListFunds`

**Function Summary:** Create and manage funds. Funds are the top-level container for portfolios and represent a named Thai investment fund with compliance, reporting, and permission boundaries.

**Key Fund Fields:**

| Field | Type | Description |
|---|---|---|
| `name` | string | Fund display name |
| `code` | string | Short fund code |
| `fund_category` | enum | Regulatory category (used in compliance scope) |
| `currency` | string | Base currency (ISO 4217) |
| `require_research_report_for_decision` | bool | If true, all decisions in this fund must link a REVIEW_COMPLETED research report. Default false. |
| `status` | enum | ACTIVE, INACTIVE |

---

### 7.10 Portfolio Management

**Function Name:** `CreatePortfolio` / `UpdatePortfolio` / `GetPortfolio` / `ListPortfolios` / `GetPortfolioHoldings` / `GetPortfolioCash`

**Function Summary:** Manage portfolios within a fund. Holdings and cash positions are read from the ledger projector.

**Position Read (Optimistic Locking):**
- Positions are stored in `investment__positions` with a `version` column.
- Updates use `UPDATE ... WHERE version = $n` to detect concurrent modifications.
- Concurrency conflicts return `ErrConcurrentModification`; caller retries.

---

### 7.11 Instrument Management

**Function Name:** `CreateInstrument` / `UpdateInstrument` / `PostPriceSnapshot` / `ListInstruments`

**Function Summary:** Manage the instrument catalog and maintain historical price snapshots used in NAV computation.

**Price Snapshot Post:**
- Actor must hold `INVESTMENT_PRICE_POST`.
- Snapshot stores ticker, price, currency, market, and business date.
- Price snapshot is authoritative for NAV computation; live quotes from Alpha Vantage are advisory.

---

### 7.12 Batch Approval

**Function Name:** `BatchApproveDecisions`

**Function Summary:** Approve or reject multiple investment decisions in a single operation. Uses the `ApprovalBatchActor` contract interface.

**Entry Conditions:**
- Each listed decision must be in `PENDING_APPROVAL` status.
- Actor must hold `INVESTMENT_DECISION_APPROVE`.
- Actor must not be the submitter of any included decision (maker-checker).

**Exception Flow:**
- Any decision in wrong state → operation skips that decision with a per-item error (partial success supported).
- Maker-checker violation → that item rejected with `ErrMakerCheckerViolation`.

---

## 8. API / Backend Behavior

### 8.1 API Route Summary

All routes are under `/api/v1/investment/` and require a valid JWT.

#### Research Reports

| Method | Path | Permission | Description |
|---|---|---|---|
| GET | `/investment/research` | `INVESTMENT_RESEARCH_VIEW` | List research reports |
| POST | `/investment/research` | `INVESTMENT_RESEARCH_CREATE` | Create a research report |
| GET | `/investment/research/{id}` | `INVESTMENT_RESEARCH_VIEW` | Get report detail |
| PUT | `/investment/research/{id}` | `INVESTMENT_RESEARCH_UPDATE` | Update a draft report |
| DELETE | `/investment/research/{id}` | `INVESTMENT_RESEARCH_DELETE` | Soft-delete a draft report |
| POST | `/investment/research/{id}/submit` | `INVESTMENT_RESEARCH_SUBMIT` | Submit for approval |
| POST | `/investment/research/{id}/cancel-submit` | `INVESTMENT_RESEARCH_CANCEL_SUBMIT` | Retract submission |
| POST | `/investment/research/{id}/invalidate` | `INVESTMENT_RESEARCH_INVALIDATE` | Invalidate a completed report |

#### Investment Decisions

| Method | Path | Permission | Description |
|---|---|---|---|
| GET | `/investment/decisions` | `INVESTMENT_DECISION_VIEW` | List decisions |
| POST | `/investment/decisions` | `INVESTMENT_DECISION_MANAGE` | Create a decision |
| GET | `/investment/decisions/{id}` | `INVESTMENT_DECISION_VIEW` | Get decision detail |
| PUT | `/investment/decisions/{id}` | `INVESTMENT_DECISION_MANAGE` | Update a draft decision |
| DELETE | `/investment/decisions/{id}` | `INVESTMENT_DECISION_CANCEL` | Cancel a decision |
| POST | `/investment/decisions/{id}/submit` | `INVESTMENT_DECISION_SUBMIT` | Submit for compliance + approval |
| POST | `/investment/decisions/{id}/mark-ready` | `INVESTMENT_EXECUTION_MANAGE` | Mark decision ready for execution |
| POST | `/investment/decisions/batch-approve` | `INVESTMENT_DECISION_APPROVE` | Batch approve decisions |

#### Executions

| Method | Path | Permission | Description |
|---|---|---|---|
| GET | `/investment/executions` | `INVESTMENT_EXECUTION_VIEW` | List executions |
| POST | `/investment/executions` | `INVESTMENT_EXECUTION_MANAGE` | Record an execution |
| GET | `/investment/executions/{id}` | `INVESTMENT_EXECUTION_VIEW` | Get execution detail |
| PUT | `/investment/executions/{id}` | `INVESTMENT_EXECUTION_MANAGE` | Update an execution |

#### Trade Confirmations

| Method | Path | Permission | Description |
|---|---|---|---|
| GET | `/investment/confirmations` | `INVESTMENT_CONFIRMATION_VIEW` | List confirmations |
| POST | `/investment/confirmations` | `INVESTMENT_CONFIRMATION_MANAGE` | Create a confirmation |
| GET | `/investment/confirmations/{id}` | `INVESTMENT_CONFIRMATION_VIEW` | Get confirmation detail |
| PUT | `/investment/confirmations/{id}` | `INVESTMENT_CONFIRMATION_MANAGE` | Update a confirmation |
| POST | `/investment/confirmations/import` | `INVESTMENT_CONFIRMATION_IMPORT` | Import confirmations from file |

#### Funds and Portfolios

| Method | Path | Permission | Description |
|---|---|---|---|
| GET | `/investment/funds` | `INVESTMENT_FUND_VIEW` | List funds |
| POST | `/investment/funds` | `INVESTMENT_FUND_MANAGE` | Create a fund |
| GET | `/investment/funds/{id}` | `INVESTMENT_FUND_VIEW` | Get fund detail |
| PUT | `/investment/funds/{id}` | `INVESTMENT_FUND_MANAGE` | Update a fund |
| GET | `/investment/portfolios` | `INVESTMENT_PORTFOLIO_VIEW` | List portfolios |
| POST | `/investment/portfolios` | `INVESTMENT_PORTFOLIO_MANAGE` | Create a portfolio |
| GET | `/investment/portfolios/{id}` | `INVESTMENT_PORTFOLIO_VIEW` | Get portfolio detail |
| GET | `/investment/portfolios/{id}/holdings` | `INVESTMENT_PORTFOLIO_VIEW` | Get current holdings |
| GET | `/investment/portfolios/{id}/cash` | `INVESTMENT_PORTFOLIO_VIEW` | Get cash position |

#### Instruments

| Method | Path | Permission | Description |
|---|---|---|---|
| GET | `/investment/instruments` | `INVESTMENT_INSTRUMENT_VIEW` | List instruments |
| POST | `/investment/instruments` | `INVESTMENT_INSTRUMENT_MANAGE` | Create an instrument |
| GET | `/investment/instruments/{id}` | `INVESTMENT_INSTRUMENT_VIEW` | Get instrument detail |
| PUT | `/investment/instruments/{id}` | `INVESTMENT_INSTRUMENT_MANAGE` | Update an instrument |
| POST | `/investment/instruments/{id}/price` | `INVESTMENT_PRICE_POST` | Post a price snapshot |

#### Ledger, Valuation, AUM

| Method | Path | Permission | Description |
|---|---|---|---|
| GET | `/investment/ledger` | `INVESTMENT_LEDGER_VIEW` | List ledger entries |
| POST | `/investment/ledger/post` | `INVESTMENT_LEDGER_POST` | Post a ledger transaction |
| POST | `/investment/ledger/simulate` | `INVESTMENT_LEDGER_SIMULATE` | Simulate a ledger post (dry-run) |
| POST | `/investment/ledger/force-post` | `INVESTMENT_LEDGER_FORCE_POST` | Force-post on a locked day |
| POST | `/investment/ledger/reverse` | `INVESTMENT_LEDGER_REVERSE` | Reverse a locked-day entry |
| POST | `/investment/valuation/run` | `INVESTMENT_VALUATION_RUN` | Trigger NAV computation |
| GET | `/investment/valuation/{portfolioId}` | `INVESTMENT_VALUATION_VIEW` | Get latest valuation |
| GET | `/investment/valuation/{portfolioId}/intraday` | `INVESTMENT_VALUATION_VIEW` | Get intraday valuation estimate |
| POST | `/investment/funds/{id}/aum` | `INVESTMENT_FUND_AUM_COMPUTE` | Compute fund AUM |

### 8.2 Cross-Module Setter Methods

The investment module exposes setter methods for receiving cross-module contract implementations (injected at startup):

| Setter | Contract | Provider |
|---|---|---|
| `SetApprovalSubmitter` | `ApprovalSubmitter` | Approval module |
| `SetApprovalCanceller` | `ApprovalCanceller` | Approval module |
| `SetApprovalStatusProvider` | `ApprovalStatusProvider` | Approval module |
| `SetApprovalBatchActor` | `ApprovalBatchActor` | Approval module |
| `SetComplianceChecker` | `ComplianceChecker` + `ComplianceSimulator` | Compliance module |
| `SetMarketQuoteProvider` | `MarketQuoteProvider` | Market data module |

Subject callbacks registered with the Approval module:

| Subject | Callback | Action on Approval | Action on Rejection |
|---|---|---|---|
| `RESEARCH_REPORT` | `researchReportApprovalCallback` | Status → `REVIEW_COMPLETED` | Status → `DRAFT` |
| `INVESTMENT_DECISION` | `investmentDecisionApprovalCallback` | Status → `APPROVED` | Status → `REJECTED` |
| `COMPLIANCE_RELEASE` | `complianceReleaseApprovalCallback` | Re-validate + submit to INVESTMENT_DECISION | Status → `CANCELLED` |

---

## 9. Frontend Behavior

### 9.1 Page Inventory

The investment module frontend is organized into four main areas reflecting the four process stages.

#### Analysis (Research) Pages

| Page (Nuxt route) | Description |
|---|---|
| `/investment/analysis` | List of research reports with status filters |
| `/investment/analysis/new` | Create a new research report |
| `/investment/analysis/[id]` | Research report detail view |
| `/investment/analysis/[id]/edit` | Edit a draft research report |

#### Decision Pages

| Page (Nuxt route) | Description |
|---|---|
| `/investment/decision` | Decision inbox — list with lifecycle status filters |
| `/investment/decision/new` | Create a new investment decision |

#### Execution Pages

| Page (Nuxt route) | Description |
|---|---|
| `/investment/execution` | Execution list — READY_FOR_EXECUTION decisions and execution records |

#### Trade Review Pages

| Page (Nuxt route) | Description |
|---|---|
| `/investment/review` | Trade confirmation review queue — PENDING_REVIEW and MISMATCHED confirmations |

#### Fund Management Pages

| Page (Nuxt route) | Description |
|---|---|
| `/investment/funds` | Fund list |
| `/investment/funds/new` | Create a new fund |
| `/investment/funds/[fundId]` | Fund overview dashboard |
| `/investment/funds/[fundId]/compliance` | Fund compliance status (IRG results for fund portfolios) |
| `/investment/funds/[fundId]/holdings` | Fund-level aggregate holdings |
| `/investment/funds/[fundId]/decisions` | Decisions scoped to this fund |
| `/investment/funds/[fundId]/operation` | Ledger operations panel |
| `/investment/funds/[fundId]/operation/new` | Post a new ledger entry |
| `/investment/funds/[fundId]/operation/[decisionId]` | Decision-linked operation detail |
| `/investment/funds/[fundId]/reviewers` | Approval reviewer assignment panel |
| `/investment/funds/[fundId]/settings` | Fund configuration settings |
| `/investment/funds/[fundId]/stages` | Workflow stage configuration for this fund |
| `/investment/funds/[fundId]/audit` | Fund-scoped audit trail |

### 9.2 Key Frontend Components

| Component | Purpose |
|---|---|
| `InvestmentDecisionCard` | Decision list item with lifecycle status badge |
| `InvestmentDecisionStatusBadge` | Color-coded badge for all decision statuses |
| `InvestmentDecisionForm` | Create/edit form: fund/portfolio/instrument selector, side, quantity, amount, date |
| `InvestmentDecisionSubmitModal` | Submit confirmation dialog with pre-submission compliance simulation |
| `InvestmentDecisionApprovalPanel` | Approval inbox panel for a decision |
| `InvestmentDecisionBatchApproveModal` | Multi-select decision batch approval UI |
| `InvestmentResearchReportForm` | Research report create/edit form |
| `InvestmentResearchStatusBadge` | Report status badge |
| `InvestmentExecutionForm` | Execution detail entry form |
| `InvestmentConfirmationTable` | Trade confirmation list with match status indicators |
| `InvestmentHoldingsTable` | Portfolio holdings table (tickers, quantities, market value) |
| `InvestmentValuationPanel` | NAV and intraday valuation display |
| `InvestmentFundDashboardLayout` | Fund-scoped navigation shell with tab switcher |

### 9.3 Key Frontend Composables

| Composable | Purpose |
|---|---|
| `useInvestmentDecisions` | Decision list and CRUD operations |
| `useInvestmentDecisionLifecycle` | Submit, cancel, mark-ready, batch-approve operations |
| `useInvestmentResearchReports` | Research report list and lifecycle |
| `useInvestmentExecutions` | Execution list and management |
| `useInvestmentConfirmations` | Confirmation list and review actions |
| `useInvestmentFunds` | Fund list and detail |
| `useInvestmentPortfolios` | Portfolio list, holdings, and cash |
| `useInvestmentInstruments` | Instrument catalog and price posting |
| `useInvestmentValuation` | NAV and intraday valuation queries |
| `useInvestmentLedger` | Ledger entry list and operations |
| `useInvestmentFundAum` | AUM computation trigger |

---

## 10. Database / Data Model Summary

All tables use prefix `investment__`. Migrations are in `database/migrations/20260501*` (funds, portfolios, instruments, ledger, positions, valuations) and `database/migrations/20260601*` (decisions, executions, confirmations).

### Core Reference Tables

| Table | Purpose |
|---|---|
| `investment__funds` | Fund master records |
| `investment__portfolios` | Portfolio master records linked to funds |
| `investment__instruments` | Instrument catalog (ticker, ISIN, asset class, currency, exchange) |

### Position and Cash Tables

| Table | Purpose | Mutability |
|---|---|---|
| `investment__positions` | Current holding quantity per portfolio-instrument | Mutable with optimistic locking (version column) |
| `investment__cash` | Current cash balance per portfolio-currency | Mutable with optimistic locking |
| `investment__ledger` | Transaction history (all posts) | Append-only |
| `investment__price_snapshots` | Historical price records per ticker-date | Append-only |
| `investment__valuations` | Computed NAV snapshots per portfolio-date | Mutable (recomputed on valuation run) |

### Investment Process Tables

| Table | Purpose | Key Constraint |
|---|---|---|
| `investment__research_reports` | Research report records | Soft-delete via `deleted_at`; `status` NOT NULL |
| `investment__decisions` | Decision records | `status` CHECK constraint; `side IN ('BUY','SELL')` |
| `investment__executions` | Execution records for decisions | FK to `investment__decisions`; `status` CHECK constraint |
| `investment__trade_confirmations` | Settlement confirmation records | FK to `investment__executions`; `status` CHECK constraint |

### Key Decision Table Columns

| Column | Type | Description |
|---|---|---|
| `status` | VARCHAR(50) | DRAFT, PENDING_APPROVAL, PENDING_COMPLIANCE_RELEASE, APPROVED, REJECTED, CANCELLED, READY_FOR_EXECUTION, EXECUTED |
| `side` | VARCHAR(10) | BUY or SELL |
| `decision_type` | VARCHAR(30) | SINGLE_ORDER, BASKET_ORDER, REBALANCE, SWITCH |
| `quantity` | NUMERIC nullable | Quantity in units; nullable if amount_based |
| `amount` | NUMERIC nullable | Amount in fund currency; nullable if quantity_based |
| `approval_request_id` | UUID nullable | FK to approval engine request (INVESTMENT_DECISION flow) |
| `compliance_release_approval_request_id` | UUID nullable | FK to approval engine request (COMPLIANCE_RELEASE flow); separate from `approval_request_id` |

> The `quantity` / `amount` constraint: at least one must be non-null (enforced by CHECK: `quantity IS NOT NULL OR amount IS NOT NULL`).

---

## 11. Validation Rules

| Rule | Enforcement Point |
|---|---|
| `fund_id`, `portfolio_id`, `instrument_id`, `decision_id` must be valid UUIDs | HTTP handler (400 on invalid) |
| `side` must be BUY or SELL | HTTP handler + DB CHECK constraint |
| At least one of `quantity` or `amount` must be present on a decision | HTTP handler + DB CHECK constraint |
| `quantity` must be positive if present | Application command validation |
| `amount` must be positive if present | Application command validation |
| `business_date` must be `YYYY-MM-DD` | HTTP handler (400 on invalid) |
| Research report must be in `REVIEW_COMPLETED` to be linked to a decision when a report is referenced | Application command policy (`CanReferenceResearchReport`) |
| Research report side must be compatible with decision side when linked | Report reference policy |
| Research report fund must match decision fund when linked | Report reference policy |
| Research report effective date must cover decision business date when linked | Report reference policy |
| Fund `require_research_report_for_decision = true` requires a linked valid report on submit | Application command gate |
| Workflow day must be open for standard decision submission | Application command gate |
| Override actor must come from JWT only — never from request body | HTTP handler (security invariant) |
| Maker-checker: submitter cannot approve | Enforced by Approval module |
| `quantity` in instrument's minimum trading unit must be a multiple of `min_trading_unit` | `quantity.min_trading_unit` IRG rule |
| SELL quantity must not exceed available position | `quantity.sell_available` IRG rule |
| Concentration limit compliance at pre-trade | `concentration.single_issuer` IRG rule |
| Cash availability at pre-trade | `cash.availability` IRG rule |
| Ledger reversal only allowed on locked-day entries | Application command gate |
| Force-post requires `INVESTMENT_LEDGER_FORCE_POST` | Router-level permission |
| Position update must pass optimistic lock check (version match) | DB-level: UPDATE WHERE version = $n |

---

## 12. Approval and Permission Rules

| Rule | Details |
|---|---|
| All investment routes require a valid JWT | Auth middleware applied globally |
| Permission checks are enforced at the router level | Route groups gate by permission code |
| Maker-checker on research reports | Submitter cannot be approver; enforced by Approval module |
| Maker-checker on investment decisions | Submitter cannot be approver; enforced by Approval module |
| COMPLIANCE_RELEASE approval requires `INVESTMENT_COMPLIANCE_RELEASE_APPROVE` | Distinct from `INVESTMENT_DECISION_APPROVE`; intended for compliance officer role |
| INVESTMENT_DECISION approval requires `INVESTMENT_DECISION_APPROVE` | Intended for investment committee / senior PM role |
| Batch approval inherits the same maker-checker constraint per decision | Applied per item; fails the item, not the batch |
| `PORTFOLIO` approval subject is registered but not triggered in current version | Phase 2 item |

---

## 13. Integration Points

| Integration | Direction | Method | Notes |
|---|---|---|---|
| **Compliance module** | Investment → Compliance | `ComplianceChecker.CheckProposedOrder` | Called at decision submission |
| **Compliance module (simulation)** | Investment → Compliance | `ComplianceSimulator.SimulateProposedOrder` | Dry-run simulation used by submission preview |
| **Approval module (RESEARCH_REPORT)** | Investment → Approval | `ApprovalSubmitter.SubmitForApproval` | On research report submission |
| **Approval module (INVESTMENT_DECISION)** | Investment → Approval | `ApprovalSubmitter.SubmitForApproval` | On decision submission after PASS/WARN verdict |
| **Approval module (COMPLIANCE_RELEASE)** | Investment → Approval | `ApprovalSubmitter.SubmitForApproval` | On decision submission when BLOCK + all overridable |
| **Approval module callbacks** | Approval → Investment | Subject callbacks registered at startup | `researchReportApprovalCallback`, `investmentDecisionApprovalCallback`, `complianceReleaseApprovalCallback` |
| **Market data module** | Investment → Market Data | `MarketQuoteProvider.GetQuote` | Used for intraday valuation overlay only |
| **Workflow module** | Investment → Workflow | `WorkflowStateProvider.IsTradeAllowed` | Workflow day gate at submission |
| **Workflow module (confirmation gate)** | Workflow → Investment | `ConfirmationGate.CheckConfirmation` | Used by workflow close-transactions transition |
| **Audit module** | Investment → Audit | `AuditLogger.LogAction` | Writes audit events for lifecycle transitions |
| **IAM module** | Investment → IAM | `PermissionChecker.HasFunctionPermission` | Route-level permission enforcement |

---

## 14. Audit and Logging

| Event | Trigger | Persistence |
|---|---|---|
| Decision lifecycle transitions | Submit, approve, reject, cancel, mark-ready, execute | Audit module (`audit_logs`) |
| Research report lifecycle transitions | Submit, cancel-submit, invalidate, delete | Audit module |
| Execution created/updated | Execution CRUD | Audit module |
| Confirmation status change | Match, mismatch, review | Audit module |
| Ledger post (all types) | Post, force-post, reversal, simulation | Audit module + `investment__ledger` (append-only) |
| Valuation run | Trigger NAV computation | Audit module |
| Price snapshot post | Post instrument price | Audit module |
| Fund and portfolio create/update | CRUD events | Audit module |

Structured `slog` output is also written for each significant operation with key fields: actor ID, entity type/ID, operation, and outcome.

---

## 15. Notifications

The investment module does not send notifications directly. Notifications are triggered by the Approval module as side-effects of approval workflow events:

| Approval Event | Notification |
|---|---|
| Research report submitted | Notification to configured approvers for `RESEARCH_REPORT` |
| Decision submitted for approval | Notification to configured approvers for `INVESTMENT_DECISION` |
| Compliance release approval pending | Notification to configured approvers for `COMPLIANCE_RELEASE` |
| Decision approved / rejected | Notification to submitter |

---

## 16. Error / Exception Handling

| Error Type | HTTP Status | When |
|---|---|---|
| `ErrDecisionNotFound` | 404 | Decision UUID not in database |
| `ErrResearchReportNotFound` | 404 | Research report UUID not in database |
| `ErrExecutionNotFound` | 404 | Execution UUID not in database |
| `ErrDecisionLifecycle` | 409 | Invalid lifecycle transition (e.g., submit from non-DRAFT status) |
| `ErrResearchReportLifecycle` | 409 | Invalid research report lifecycle transition |
| `ErrWorkflowDayClosed` | 409 | Submission attempted on a closed workflow day |
| `ErrReportRequired` | 409 | Fund requires a research report but none linked |
| `ErrComplianceRejected` | 422 | Pre-trade check returned BLOCK with non-overridable breach |
| `ErrInsufficientBalance` | 409 | Ledger post would result in negative cash or position |
| `ErrCannotReverseOpenDay` | 409 | Reversal attempted against an open workflow day entry |
| `ErrConcurrentModification` | 409 | Optimistic lock failure on position or cash update; client should retry |
| `ErrMarketDataUnavailable` | 503 | Live market quote fetch failed |
| `ErrMakerCheckerViolation` | 403 | Submitter attempting to approve their own decision |
| Infrastructure failure (DB, network) | 500 | Any unhandled error; internal details not leaked to client |

---

## 17. Known Gaps / TODOs

| Item | Priority | Notes |
|---|---|---|
| `PORTFOLIO` approval subject deferred | High | Approval subject `INVESTMENT_PORTFOLIO` is registered but never triggered. Portfolio-level approval workflow is Phase 2. |
| Per-line compliance checks for basket/rebalance/switch | High | `BASKET_ORDER`, `REBALANCE`, and `SWITCH` decisions run a single pre-trade check with an empty ticker. Per-instrument line checking is Phase 2. |
| Watchlist / threshold monitoring | High | `docs/api/watchlist-api.md` contains a complete API design but no backend implementation exists. |
| `NopCalendarAdapter` | Medium | No trading calendar enforcement. A real calendar adapter is Phase 2. |
| `NopTradeHistoryAdapter` | Medium | Trade history always empty — any rule depending on trade history will not function. |
| `NopLeaveChecker` | Low | Always returns false — no leave-based trade restriction. |
| Decision update while in PENDING_COMPLIANCE_RELEASE | Medium | Whether a PM can revise an in-flight compliance-release decision before the compliance officer acts is not defined in code. `CanEdit()` needs explicit check for this state. |
| Multi-currency NAV | Low | Single base currency per fund. Cross-currency portfolios are Phase 2. |
| Intraday quote refresh frequency | Low | No throttle or staleness indicator on intraday valuation — every UI open triggers a live fetch. |
| Confirmation import format | Low | Import endpoint exists but supported file formats (CSV schema, broker adapters) are not documented in code. |
| Negative quantity scenario | Medium | See `docs/runbook/investment.md`: negative quantity after data import can occur. Manual correction via force-post is the workaround. |

---

## 18. Acceptance Checklist

- [x] Decision cannot be submitted from a closed workflow day
- [x] Pre-trade compliance check runs at decision submission
- [x] BLOCK + all overridable → PENDING_COMPLIANCE_RELEASE
- [x] BLOCK + any non-overridable → HTTP 422 (`ErrComplianceRejected`)
- [x] PASS and WARN → decision proceeds to PENDING_APPROVAL
- [x] Compliance release approval re-validates workflow gate
- [x] Research report is optional by default; only required when `fund.RequireResearchReportForDecision = true`
- [x] Linked research report must be `REVIEW_COMPLETED`, same fund, compatible side and date
- [x] Maker-checker enforced: submitter cannot approve
- [x] Decision statuses: DRAFT, PENDING_COMPLIANCE_RELEASE, PENDING_APPROVAL, APPROVED, REJECTED, CANCELLED, READY_FOR_EXECUTION, EXECUTED
- [x] Execution and trade confirmation records track fill details
- [x] Ledger is append-only; reversals are new entries
- [x] Position updates use optimistic locking (version column)
- [x] Intraday valuation is non-authoritative and does not mutate NAV tables
- [x] All lifecycle events written to audit log
- [x] 28 permission codes documented and implemented
- [ ] Per-line compliance checks for basket/rebalance/switch decisions
- [ ] PORTFOLIO approval subject activated
- [ ] Watchlist / threshold monitoring backend implemented
