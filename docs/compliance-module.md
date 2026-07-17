# Compliance Module — Function Description

---

## Document Control

| Field | Value |
|---|---|
| **Version** | 1.0.0 |
| **Date** | 2026-06-24 |
| **Author** | Systemweb / IMS Documentation |
| **Status** | Reviewed |
| **Source** | `backend/internal/compliance/`, `frontend/app/features/compliance/`, `database/migrations/20260417*`, handoff docs |

---

## Table of Contents

1. [Module Overview](#1-module-overview)
2. [Business Purpose](#2-business-purpose)
3. [Functional Scope](#3-functional-scope)
4. [Out of Scope](#4-out-of-scope)
5. [User Roles and Permissions](#5-user-roles-and-permissions)
6. [Workflow / User Flow](#6-workflow--user-flow)
7. [Function Descriptions](#7-function-descriptions)
   - 7.1 [Pre-Trade Check](#71-pre-trade-check)
   - 7.2 [Post-Trade Check](#72-post-trade-check)
   - 7.3 [Breach Management](#73-breach-management)
   - 7.4 [Rule Instance Administration](#74-rule-instance-administration)
   - 7.5 [Check Group Query](#75-check-group-query)
8. [IRG Rule Engine Architecture](#8-irg-rule-engine-architecture)
9. [Registered Rule Types](#9-registered-rule-types)
10. [API / Backend Behavior](#10-api--backend-behavior)
11. [Frontend Behavior](#11-frontend-behavior)
12. [Database / Data Model Summary](#12-database--data-model-summary)
13. [Validation Rules](#13-validation-rules)
14. [Approval and Permission Rules](#14-approval-and-permission-rules)
15. [Integration Points](#15-integration-points)
16. [Audit and Logging](#16-audit-and-logging)
17. [Notifications](#17-notifications)
18. [Error / Exception Handling](#18-error--exception-handling)
19. [Known Gaps / TODOs](#19-known-gaps--todos)
20. [Acceptance Checklist](#20-acceptance-checklist)

---

## 1. Module Overview

**Module name:** `compliance` (also referred to as IRG — Investment Regulation Guard)

**Package path:** `backend/internal/compliance/`

**API base path:** `/api/v1/compliance/`

**Frontend feature path:** `frontend/app/features/compliance/`

The Compliance module is the backend boundary for all rule-based investment compliance checks within the IMS platform. It provides a plugin-based evaluation engine (the IRG pipeline) that evaluates configured rule instances against proposed or completed trades and portfolio states. The module persists all check results in an immutable audit trail and manages the lifecycle of compliance breaches.

> **Note:** The module README at `backend/internal/compliance/README.md` is outdated. It describes the module as "scaffolded only." The actual implementation is fully functional with a complete evaluation engine, persistence layer, HTTP API, frontend UI, and cross-module integration.

---

## 2. Business Purpose

The Compliance module enforces investment mandates, regulatory constraints, and internal house rules before and after trades are executed. It answers the question: "Is this proposed trade permitted?"

Key business objectives:

- Prevent portfolio managers from executing trades that breach concentration limits, cash availability constraints, credit rating requirements, sector exposure caps, or restricted-instrument lists.
- Provide a structured escalation path (via the Approval module) for borderline cases where breaches may be released by a compliance officer.
- Maintain an immutable, reproducible audit trail of every compliance evaluation.
- Allow compliance administrators to configure rules without code changes, by creating and binding `RuleInstance` records to scopes.

---

## 3. Functional Scope

The following capabilities are implemented:

| Capability | Status |
|---|---|
| Pre-trade compliance check (IRG pipeline) | Active |
| Pre-trade dry-run / simulation (no persistence) | Active |
| Post-trade compliance check | Active |
| Breach creation and lifecycle management | Active |
| Compliance officer override of overridable breaches | Active |
| Rule instance administration (create, list) | Active |
| Check group audit query | Active |
| Cross-module contract adapter (ComplianceChecker, ComplianceSimulator, PostTradeVerifier) | Active |
| COMPLIANCE_RELEASE approval flow (BLOCK + overridable breach escalation) | Active |
| Dashboard tabs and frontend compliance pages | Active |

---

## 4. Out of Scope

The following are not implemented in the current version:

- Rule binding management via HTTP API (bindings exist in the DB but have no CRUD endpoints)
- Rule set management via HTTP API
- `GET /compliance/rule-types` endpoint (rule types are discoverable in the Go registry only; the frontend uses a static catalog mirror)
- Second-level approval for override actions (`OverrideBreachRequest.ApprovedBy` is nil; see Known Gaps)
- Per-instrument-line compliance checks for BASKET_ORDER, REBALANCE, and SWITCH decisions (header-level checks only)
- Real Thai SEC regulatory checks (`regulatory.thai_sec` is a stub returning `VerdictWarn`)
- Real credit rating lookup (`credit_rating.minimum` stub; `NopCreditRatingAdapter` always returns empty)
- Leave checker integration (`NopLeaveChecker` always returns false)
- Trade history integration (`NopTradeHistoryAdapter` always returns empty)
- Calendar integration (`NopCalendarAdapter` always returns nil)
- Periodic scheduled compliance scans
- Watchlist / threshold-based monitoring (a separate planned module)

---

## 5. User Roles and Permissions

All compliance permissions are seeded into the `Admin` group by default. Other group assignments must be configured via the Permissions module.

| Permission Code | Name | Description |
|---|---|---|
| `IRG_VIEW_RULES` | View Rules | Read rule instances and check group results. Required for `GET /compliance/rules` and `GET /compliance/checks/{groupID}` and `GET /compliance/breaches`. |
| `IRG_EDIT_RULE_INSTANCE` | Edit Rule Instance | Create new rule instances. Required for `POST /compliance/rules`. |
| `IRG_EDIT_BINDING` | Edit Binding | Create or modify rule bindings (DB-only in current version; no HTTP API). |
| `IRG_OVERRIDE_BREACH` | Override Breach | Override an open, overridable breach. Required for `POST /compliance/breaches/{id}/override`. |
| `IRG_ADMIN_RULE_TYPE` | Admin Rule Type | Reserved for future rule type management. Not currently enforced on any route. |
| `WORKFLOW_EXECUTE` | Workflow Execute | Required to call the HTTP endpoints `POST /compliance/checks/pre-trade` and `POST /compliance/checks/post-trade` directly. |

> Pre-trade and post-trade checks invoked in-process via the `ComplianceContractAdapter` do not require `WORKFLOW_EXECUTE` — that permission gates the HTTP endpoint only.

---

## 6. Workflow / User Flow

### 6.1 Pre-Trade Flow (Decision Submission)

```
Investment Decision (DRAFT)
  │
  ▼ POST /investment/decisions/{id}/submit
  │
  ├── Workflow gate: IsTradeAllowed (workflow day must be open)
  │
  ├── Fund report gate: Fund.RequireResearchReportForDecision
  │   └── If true and no report → rejected (ErrDecisionLifecycle)
  │
  ├── Report reference policy: CanReferenceResearchReport
  │   └── If report linked → validate status, scope, side, date
  │
  ├── Pre-trade IRG check: ComplianceContractAdapter.CheckProposedOrder
  │   │
  │   ├── PASS or WARN → continue to INVESTMENT_DECISION approval
  │   │
  │   ├── BLOCK + all breaches Overridable=true
  │   │   └── Submit COMPLIANCE_RELEASE approval request
  │   │       Decision status → PENDING_COMPLIANCE_RELEASE
  │   │       (Awaiting compliance officer approval)
  │   │
  │   └── BLOCK + any breach Overridable=false
  │       └── Return ErrComplianceRejected (HTTP 422, terminal)
  │
  └── Submit to INVESTMENT_DECISION approval engine
      Decision status → PENDING_APPROVAL
```

### 6.2 Compliance Release Flow (BLOCK + Overridable)

```
PENDING_COMPLIANCE_RELEASE
  │
  ▼ Compliance officer approves COMPLIANCE_RELEASE approval request
  │
  ├── Re-validate workflow gate and report reference
  └── Auto-submit to INVESTMENT_DECISION approval engine
      Decision status → PENDING_APPROVAL

  OR

  ▼ Compliance officer rejects COMPLIANCE_RELEASE
  └── Decision status → CANCELLED
```

### 6.3 Post-Trade Flow (Workflow Closing Gate)

```
Workflow: CloseTransactions transition
  │
  ▼ PostTradeVerifier.RunPostTradeVerification
  │
  ├── PASS → allow TRANSACTION_CLOSED transition
  └── BLOCK verdict found → refuse closing (workflow gate)
```

### 6.4 Compliance Officer Override Flow

```
Breach (OPEN) detected in check results
  │
  ▼ POST /compliance/breaches/{id}/override
  │
  Actor (IRG_OVERRIDE_BREACH permission) supplies reason
  ├── Reason validated (non-empty)
  ├── Override record created (actor from JWT, not request body)
  └── Breach status → OVERRIDDEN
```

---

## 7. Function Descriptions

### 7.1 Pre-Trade Check

**Function Name:** `RunPreTradeCheck` / `SimulateProposedOrder`

**Function Summary:** Evaluate all applicable compliance rules against a proposed trade order before submission. Returns an aggregated verdict (PASS / WARN / BLOCK) and a list of breaches.

**Entry Conditions:**
- `portfolio_id` and `contract_id` must be valid UUIDs.
- `ticker` must be non-empty.
- `quantity` must be positive.
- `price` must be positive.
- `fees` must be non-negative.
- `business_date` must be provided.
- Actor must hold `WORKFLOW_EXECUTE` to call the HTTP endpoint directly.

**Exception Flow:**
- Infrastructure failure (DB, network) → HTTP 500.
- Invalid request fields → HTTP 400.
- Missing authentication → HTTP 401.
- Missing `WORKFLOW_EXECUTE` permission → HTTP 403.
- A BLOCK verdict is not an exception — it is returned in the HTTP 201 response body with `verdict: "BLOCK"` and a populated `breaches` array.

**Post-Execution Status:**
- `compliance_check_records` rows created for each evaluated rule (append-only).
- `compliance_breaches` rows created for BLOCK or WARN verdicts.
- Response includes `check_group_id` which links all records for this check.

**Special Requirements:**
- Dry-run mode (`HandleDryRun` / `SimulateProposedOrder`): evaluates the same rules but skips persistence. Used by investment simulation endpoints.
- All rules are evaluated — the engine does not short-circuit on the first BLOCK.

**Path:** `POST /api/v1/compliance/checks/pre-trade`

---

### 7.2 Post-Trade Check

**Function Name:** `RunPostTradeCheck`

**Function Summary:** Evaluate compliance rules at contract scope after trades have been captured, typically called during the workflow day-close process.

**Entry Conditions:**
- `portfolio_id`, `contract_id`, and `business_date` required.
- Actor must hold `WORKFLOW_EXECUTE` to call the HTTP endpoint directly.

**Exception Flow:**
- Same as pre-trade.

**Post-Execution Status:**
- Check records and breaches persisted.
- The `PostTradeVerifier.RunPostTradeVerification` variant (called by the workflow module) scans at contract + global scope only and does not require a `portfolio_id`.

**Path:** `POST /api/v1/compliance/checks/post-trade`

---

### 7.3 Breach Management

**Function Name:** `ListBreaches` / `OverrideBreach`

**Function Summary:** List open and historical breaches for audit and compliance officer review. Override an open, overridable breach with a written justification.

**Entry Conditions (Override):**
- `breach_id` must be a valid UUID referencing an existing breach.
- `reason` must be non-empty.
- Actor must hold `IRG_OVERRIDE_BREACH` permission.
- Breach status must be `OPEN`.
- Breach must not already have an active override.

**Exception Flow:**
- Breach not found → HTTP 404 (`ErrBreachNotFound`).
- Breach not `OPEN` → HTTP 409 (`ErrBreachNotOpen`).
- Override already exists → HTTP 409 (`ErrOverrideAlreadyExists`).
- Invalid request → HTTP 400 (`ErrInvalidOverrideRequest`).

**Post-Execution Status:**
- `compliance_overrides` row created (append-only, immutable).
- `compliance_breaches.status` updated to `OVERRIDDEN`.

**Field Description:**

| Field | Type | Required | Description |
|---|---|---|---|
| `reason` | string | Yes | Mandatory written justification for the override |
| `overridden_by` | UUID | System | Derived from JWT claims — never accepted from request body |

**Path:** `GET /api/v1/compliance/breaches`, `POST /api/v1/compliance/breaches/{breachID}/override`

---

### 7.4 Rule Instance Administration

**Function Name:** `CreateRuleInstance` / `ListRuleInstances`

**Function Summary:** Create a configured compliance rule instance linking a registered rule type to an effective window and initial parameters. List existing rule instances with optional filters.

**Entry Conditions (Create):**
- `rule_type_id` must match a registered `RuleEvaluator` in the SPI registry.
- `name` must be non-empty and unique.
- `parameters` must be valid JSON conforming to the rule type's `ParameterSchema`.
- `effective_from` must be a valid date.
- Actor must hold `IRG_EDIT_RULE_INSTANCE` permission.

**Exception Flow:**
- Unknown `rule_type_id` → HTTP 422 (`ErrRuleTypeNotFound`).
- Invalid parameters (JSON Schema violation) → HTTP 400 (`ErrParameterValidation`).
- Invalid request fields → HTTP 400 (`ErrInvalidCreateRuleRequest`).

**Post-Execution Status:**
- `compliance_rule_instances` row created.
- `compliance_rule_instance_versions` row created (version 1, immutable).

**Path:** `GET /api/v1/compliance/rules`, `POST /api/v1/compliance/rules`

---

### 7.5 Check Group Query

**Function Name:** `GetCheckGroup`

**Function Summary:** Retrieve all check records and associated breaches for a single compliance check group ID (e.g., the group produced by a pre-trade check on a specific order).

**Entry Conditions:**
- `groupID` must be a valid UUID.
- Actor must hold `IRG_VIEW_RULES` permission.

**Exception Flow:**
- Group not found → HTTP 404.

**Path:** `GET /api/v1/compliance/checks/{groupID}`

---

## 8. IRG Rule Engine Architecture

The IRG pipeline is a layered evaluation engine.

### 8.1 SPI (Service Provider Interface)

Each rule type is a Go package implementing the `spi.RuleEvaluator` interface:

```
RuleEvaluator
  ├── Metadata()          — stable TypeID, category, default severity, overridable flag
  ├── ParameterSchema()   — JSON Schema for parameter validation
  ├── DataDependencies()  — declares what data the rule reads
  ├── Evaluate()          — pure function: CheckInput + DataBundle + ParameterSet → EvalResult
  └── Explain()           — human-readable explanation of a result
```

Rule packages self-register by calling `spi.Register()` in their `init()` function. The module's blank-imports in `module.go` ensure all built-in packages register before the server starts.

### 8.2 Pipeline Execution Steps

1. **Resolve applicable rules** from `compliance_rule_bindings` based on the input scopes and business date.
2. **Sort** resolved bindings by scope specificity (PORTFOLIO > CONTRACT > FUND_CATEGORY > JURISDICTION > GLOBAL), then by priority (lower number = higher priority).
3. **Union data dependencies** from all applicable rules.
4. **Batch-fetch data** through ports (positions, market prices, classification, credit ratings, restriction lists). Fail-closed on fetch error.
5. **Evaluate each rule** — all rules run; no short-circuit. Evaluation errors result in `VerdictBlock` (fail-closed).
6. **Apply severity cap** from the binding: `MONITOR` caps to PASS, `WARN` caps BLOCK to WARN, `BLOCK` allows full BLOCK.
7. **Persist** check records (append-only) and breach records for BLOCK or WARN verdicts.
8. **Aggregate** final verdict: BLOCK > WARN > PASS (domination ranking).

### 8.3 Severity vs. Verdict

| Binding Severity | Effect |
|---|---|
| `BLOCK` | Rule raw verdict passes through unchanged. A BLOCK rule verdict → BLOCK final verdict. |
| `WARN` | Caps a BLOCK raw verdict to WARN. A WARN raw verdict stays WARN. A PASS stays PASS. |
| `REQUIRE_APPROVAL` | Equivalent to WARN for verdict capping. Semantic distinction for UI. |
| `MONITOR` | Always produces PASS final verdict regardless of raw verdict (log-only mode). |

### 8.4 Scope Specificity

Conflicts between two bindings for the same rule type targeting the same portfolio use scope specificity:

| Scope | Specificity |
|---|---|
| PORTFOLIO | 50 (most specific) |
| CONTRACT | 40 |
| ASSET_CLASS | 30 |
| INSTRUMENT_TYPE | 30 |
| FUND_CATEGORY | 20 |
| JURISDICTION | 10 |
| GLOBAL | 0 (least specific) |

Equal specificity is resolved by the `priority` field (lower = higher priority).

### 8.5 Overridable Flag

Each rule type declares `Overridable bool` in its metadata. This flag determines whether a BLOCK breach can be escalated to the compliance officer for release:

- `Overridable = true` → BLOCK breach may be overridden. If all BLOCK breaches on a check are overridable, the decision transitions to `PENDING_COMPLIANCE_RELEASE`.
- `Overridable = false` → BLOCK breach is terminal. The decision is rejected immediately with HTTP 422.

---

## 9. Registered Rule Types

The following rule types are registered in the global SPI registry at startup. The relevant Go package paths are `backend/internal/compliance/rules/*`.

| TypeID | Label | Category | Default Severity | Overridable | Status |
|---|---|---|---|---|---|
| `cash.availability` | Available cash check | MANDATE | BLOCK | true | Implemented |
| `concentration.single_issuer` | Maximum single-issuer exposure | MANDATE | BLOCK | true | Implemented |
| `credit.min_rating` | Minimum credit rating | RESTRICTION | BLOCK | false | Implemented (stub credit port) |
| `credit_rating.minimum` | Minimum credit rating (stub) | RESTRICTION | BLOCK | false | Stub — NopCreditRatingAdapter |
| `quantity.min_trading_unit` | Minimum trading unit | MANDATE | BLOCK | true | Implemented |
| `quantity.sell_available` | Available-to-sell quantity | MANDATE | BLOCK | false | Implemented |
| `ratio.sector_exposure` | Maximum sector exposure | RATIO | BLOCK | true | Implemented |
| `regulatory.thai_sec` | Thai SEC regulatory check | REGULATORY | WARN | false | **STUB — emits VerdictWarn with NOT_CONFIGURED evidence** |
| `restriction.blacklist` | Restricted security blacklist | RESTRICTION | BLOCK | false | Implemented |
| `restriction.whitelist` | Whitelist-only investment | RESTRICTION | BLOCK | false | Implemented |
| `restriction.list_enforcement` | Restricted-list enforcement | RESTRICTION | BLOCK | false | Implemented |

> **Important:** `regulatory.thai_sec` returns `VerdictWarn` with evidence `{"status":"NOT_CONFIGURED","stub":"true"}` in all cases. It does not implement actual Thai SEC regulatory logic. Real Thai SEC / BOT regulatory checks are Phase 2 work.

---

## 10. API / Backend Behavior

### 10.1 API Route Summary

All routes are under `/api/v1` and require a valid JWT.

| Method | Path | Permission Required | Handler |
|---|---|---|---|
| POST | `/compliance/checks/pre-trade` | `WORKFLOW_EXECUTE` | Run pre-trade IRG check |
| POST | `/compliance/checks/post-trade` | `WORKFLOW_EXECUTE` | Run post-trade IRG check |
| GET | `/compliance/checks/{groupID}` | `IRG_VIEW_RULES` | Get check group records and breaches |
| GET | `/compliance/breaches` | `IRG_VIEW_RULES` | List breaches with filters |
| POST | `/compliance/breaches/{breachID}/override` | `IRG_OVERRIDE_BREACH` | Override an open breach |
| GET | `/compliance/rules` | `IRG_VIEW_RULES` | List rule instances |
| POST | `/compliance/rules` | `IRG_EDIT_RULE_INSTANCE` | Create a rule instance |

### 10.2 PreTradeCheck Request Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `portfolio_id` | UUID string | Yes | Portfolio being checked |
| `contract_id` | UUID string | Yes | Fund/contract context |
| `business_date` | string (YYYY-MM-DD) | Yes | Trading date |
| `order_id` | UUID string | Yes | Order identifier (echoed in check records) |
| `ticker` | string | Yes | Instrument ticker symbol |
| `side` | enum: BUY, SELL | Yes | Order direction |
| `quantity` | decimal string | Yes | Order quantity (positive) |
| `price` | decimal string | Yes | Order price (positive) |
| `fees` | decimal string | No | Estimated fees (non-negative) |
| `currency` | string | Yes | ISO 4217 currency code |
| `exchange` | string | Yes | Exchange identifier |

### 10.3 PreTradeCheck Response Fields

| Field | Type | Description |
|---|---|---|
| `check_group_id` | UUID | Groups all check records for this request |
| `verdict` | enum: PASS, WARN, BLOCK | Aggregated final verdict |
| `rules_evaluated` | int | Number of rules evaluated |
| `total_duration_ms` | int64 | Pipeline execution time |
| `breaches` | array | Breach summaries (present when WARN or BLOCK) |

**Breach summary fields:**

| Field | Type | Description |
|---|---|---|
| `breach_id` | UUID | Reference to `compliance_breaches.id` |
| `rule_type_id` | string | SPI TypeID of the violated rule |
| `severity` | enum: BLOCK, WARN, REQUIRE_APPROVAL, MONITOR | Effective severity from binding |
| `verdict` | enum: PASS, WARN, BLOCK | Final verdict for this rule |
| `message` | string | Human-readable violation description |
| `overridable` | bool | Whether a compliance officer can release this breach |

### 10.4 Cross-Module Contract Adapter

The `ComplianceContractAdapter` implements three `pkg/contract` interfaces that other modules consume:

| Interface | Method | Consumer |
|---|---|---|
| `ComplianceChecker` | `CheckProposedOrder` | Investment module (decision submit path) |
| `ComplianceSimulator` | `SimulateProposedOrder` | Investment module (trade simulation) |
| `PostTradeVerifier` | `RunPostTradeVerification` | Workflow module (close-transactions gate) |

Consumers hold a reference to the contract adapter only — they never import `internal/compliance`.

---

## 11. Frontend Behavior

### 11.1 Page Inventory

| Page (Nuxt route) | Description |
|---|---|
| `/compliance` | Dashboard overview with KPI cards, recent failures, and high-risk items |
| `/compliance/rules` | List of configured rule instances with filters (active, rule type) |
| `/compliance/rules/new` | Create a new rule instance form |
| `/compliance/rules/[id]` | Rule instance detail view |
| `/compliance/pre-trade` | Pre-trade check simulator — run a manual pre-trade check against a portfolio |
| `/compliance/post-trade` | Post-trade check panel |
| `/compliance/exceptions` | Breach inbox — list and manage open breaches |
| `/compliance/audit` | Compliance audit timeline |
| `/compliance/permissions` | Compliance permission matrix |

### 11.2 Key Frontend Components

| Component | Purpose |
|---|---|
| `ComplianceDashboardCards` | KPI summary cards (open breaches count, etc.) |
| `ComplianceRuleTable` | Rule instance list with status badges and filters |
| `ComplianceRuleBuilder` | Create rule instance form with type selector and JSON parameters |
| `ComplianceBreachInbox` | Paginated breach list with status filters |
| `ComplianceBreachOverrideDialog` | Override form — accepts `reason` only; actor is from JWT |
| `ComplianceCheckResultPanel` | Displays verdict, breach list, and rule-type labels from check result |
| `ComplianceVerdictBadge` | Colored badge for PASS / WARN / BLOCK |
| `ComplianceSeverityBadge` | Colored badge for rule severity levels |
| `ComplianceTestPanel` | Pre-trade simulation panel |
| `ComplianceAuditTimeline` | Audit event timeline |

### 11.3 Frontend Composables

| Composable | Purpose |
|---|---|
| `useComplianceRules` | Rule instance list fetching and state |
| `useComplianceBreaches` | Breach list fetching with filters |
| `useComplianceChecks` | Pre/post-trade check invocation |
| `useCompliancePortfolioDirectory` | Portfolio picker for the pre-trade simulator |
| `useComplianceRuleDirectory` | Rule type catalog lookup |
| `useComplianceUserDirectory` | User directory for breach assignment display |

### 11.4 Rule Type Catalog

The frontend uses a static mirror of the SPI registry (`ruleTypeCatalog.ts`) to display human-readable labels and suggested corrections. This is not a live API endpoint — it is manually kept in sync with the backend registry. An unknown `rule_type_id` falls back to the raw TypeID.

---

## 12. Database / Data Model Summary

All tables use prefix `compliance_`. Migrations are in `database/migrations/20260417*.sql` and `database/migrations/20260424*.sql`.

| Table | Purpose | Mutability |
|---|---|---|
| `compliance_rule_instances` | Configured rule instance records | Mutable (status, active flag, name) |
| `compliance_rule_instance_versions` | Immutable versioned parameter snapshots | Append-only (DB rules prevent UPDATE/DELETE) |
| `compliance_rule_sets` | Named groupings of rule instances for bulk binding | Mutable |
| `compliance_rule_set_members` | Membership of a rule instance in a rule set | Mutable |
| `compliance_rule_bindings` | Scope → rule mapping with severity and effective dating | Mutable |
| `compliance_check_records` | Per-rule evaluation records from every check run | Append-only (REVOKE UPDATE/DELETE) |
| `compliance_breaches` | Breach records created from BLOCK or WARN final verdicts | Mutable (status field) |
| `compliance_overrides` | Compliance officer override actions for open breaches | Append-only (REVOKE UPDATE/DELETE) |
| `compliance_restriction_lists` | Blacklist / whitelist instrument entries | Mutable |

### Key Schema Details

**`compliance_rule_bindings`:**

| Column | Type | Description |
|---|---|---|
| `scope_type` | VARCHAR(50) | GLOBAL, JURISDICTION, FUND_CATEGORY, CONTRACT, PORTFOLIO, ASSET_CLASS, INSTRUMENT_TYPE |
| `scope_id` | UUID nullable | NULL for GLOBAL scope |
| `severity` | VARCHAR(30) | BLOCK, WARN, REQUIRE_APPROVAL, MONITOR |
| `priority` | INT | Lower = higher priority for tie-breaking |
| `effective_from` | DATE | Start of applicability |
| `effective_to` | DATE nullable | End of applicability; NULL = open-ended |

**`compliance_check_records`:**

| Column | Type | Description |
|---|---|---|
| `check_group_id` | UUID | Groups all records from one check request |
| `timing` | VARCHAR(30) | PRE_TRADE, POST_TRADE, PERIODIC |
| `parameter_snapshot` | JSONB | Frozen copy of parameters used at evaluation time |
| `verdict` | VARCHAR(10) | Raw verdict from rule: PASS, WARN, BLOCK |
| `effective_severity` | VARCHAR(30) | Severity from binding |
| `final_verdict` | VARCHAR(10) | After severity cap |
| `data_snapshot_hash` | VARCHAR(64) | SHA-256 of DataBundle for reproducibility |

**`compliance_breaches`:**

| Column | Type | Description |
|---|---|---|
| `status` | VARCHAR(30) | OPEN, OVERRIDDEN, RESOLVED |
| `evidence` | JSONB | Metrics and threshold details from evaluation |
| `message` | TEXT | Human-readable violation description |

---

## 13. Validation Rules

| Rule | Enforcement Point |
|---|---|
| `portfolio_id`, `contract_id`, `order_id` must be valid UUIDs | HTTP handler (400 on invalid) |
| `business_date` must be `YYYY-MM-DD` | HTTP handler (400 on invalid) |
| `quantity` must be positive | Application command validation |
| `price` must be positive | Application command validation |
| `fees` must be non-negative | Application command validation |
| `ticker` must be non-empty for pre-trade | Application command validation |
| `reason` must be non-empty for override | HTTP handler and application command |
| `rule_type_id` must exist in SPI registry | Application command (422 on not found) |
| `parameters` must conform to rule's JSON Schema | Application command (400 on schema violation) |
| `effective_from` must be a valid date | HTTP handler (400 on invalid) |
| Override actor is from JWT, not request body | HTTP handler (security invariant) |
| Breach must be `OPEN` to override | Application command (409 if not OPEN) |
| Breach must not have existing active override | Application command (409 if duplicate) |
| Unregistered rule type evaluates to VerdictBlock | Engine pipeline (fail-closed) |
| Data fetch failure results in VerdictBlock | Engine pipeline (fail-closed) |
| Evaluation error results in VerdictBlock | Engine pipeline (fail-closed) |

---

## 14. Approval and Permission Rules

| Rule | Details |
|---|---|
| All compliance routes require a valid JWT | Auth middleware applied globally |
| `IRG_OVERRIDE_BREACH` required to override a breach | Router-level permission group |
| Override actor is always derived from JWT `Subject` claim | `OverrideRequest.ApprovedBy` is intentionally not sourced from request body |
| `WORKFLOW_EXECUTE` required to call HTTP check endpoints | Router-level; in-process calls via ContractAdapter do not require this |
| `IRG_VIEW_RULES` required to read rules, breaches, check groups | Router-level |
| `IRG_EDIT_RULE_INSTANCE` required to create rule instances | Router-level |
| All compliance Admin group permissions are seeded in migration `20260417000001` | DB seed; modifiable via Permissions module |

---

## 15. Integration Points

| Integration | Direction | Method | Notes |
|---|---|---|---|
| **Investment module** | Compliance → Investment (via contract) | `ComplianceChecker.CheckProposedOrder` | Called in `DecisionCommandHandler.Submit` before the approval engine |
| **Investment module (simulation)** | Compliance → Investment | `ComplianceSimulator.SimulateProposedOrder` | Used for dry-run trade simulation; no breach rows written |
| **Workflow module** | Compliance → Workflow (via contract) | `PostTradeVerifier.RunPostTradeVerification` | Called by workflow `CloseTransactions`; BLOCK verdict prevents day closure |
| **Approval module** | Investment → Approval (via contract) | `ApprovalSubmitter.SubmitForApproval` | COMPLIANCE_RELEASE approval requests created by the investment module, not compliance directly |
| **Audit module** | Compliance → Audit | `AuditLogger.LogAction` | Audit records written for override events |
| **IAM module** | Compliance → IAM (via middleware) | `PermissionChecker.HasFunctionPermission` | Route-level permission enforcement |

---

## 16. Audit and Logging

| Event | Trigger | Persistence |
|---|---|---|
| Every compliance check record | Each rule evaluation in the IRG pipeline | `compliance_check_records` (append-only) |
| Breach creation | BLOCK or WARN final verdict | `compliance_breaches` |
| Override action | `POST /compliance/breaches/{id}/override` | `compliance_overrides` (append-only) |
| Structured server logs | Each rule evaluation result | `slog.Info` with check_group_id, rule_type_id, verdict, duration |

The `data_snapshot_hash` (SHA-256) field in `compliance_check_records` allows reproduction of the exact data state used during any historical evaluation.

---

## 17. Notifications

The compliance module does not directly send notifications. The `COMPLIANCE_RELEASE` approval request created by the investment module triggers the standard Approval module notification flow (notifying assigned compliance officers).

---

## 18. Error / Exception Handling

| Error Type | HTTP Status | When |
|---|---|---|
| `ErrInvalidOverrideRequest` | 400 | Missing or invalid request fields for override |
| `ErrBreachNotFound` | 404 | Breach UUID not in database |
| `ErrBreachNotOpen` | 409 | Breach status is not OPEN |
| `ErrOverrideAlreadyExists` | 409 | Active override already exists for this breach |
| `ErrRuleTypeNotFound` | 422 | `rule_type_id` not in SPI registry |
| `ErrParameterValidation` | 400 | Parameters fail rule type's JSON Schema |
| `ErrInvalidCreateRuleRequest` | 400 | Missing or invalid fields for rule instance creation |
| Infrastructure failure (DB, network) | 500 | Any unhandled error; internal details not leaked to client |
| BLOCK verdict | 201 (not an error) | Returned in response body with breach details |

> **Design note:** A BLOCK verdict on a pre-trade check is NOT an HTTP error code. It is communicated through the response body with `verdict: "BLOCK"`. The investment module translates BLOCK into an `ErrComplianceRejected` (HTTP 422) when appropriate.

---

## 18a. Portfolio Compliance V2 (additive, portfolio_id-primary)

Added 2026-07-07. Portfolio is the compliance identity; `fund_id` is optional.
All V2 routes are mounted under `/api/v2/portfolios/{portfolioCode}/compliance/*`
by the **investment** module (it owns `{portfolioCode}` resolution and reuses
the same ambiguity/permission handling as every other Portfolio V2 route).
Compliance never resolves portfolio codes itself; it exposes
`pkg/contract.PortfolioComplianceContract`, implemented by
`internal/compliance/transport/portfolio_contract_adapter.go` and wired into
investment via `Module.SetPortfolioComplianceAdmin` in `cmd/server/main.go`.

| Method | Path | Permission | Notes |
|---|---|---|---|
| POST | `/portfolios/{portfolioCode}/compliance/checks/pre-trade` | `WORKFLOW_EXECUTE` | fund_id passed through only if the portfolio has one |
| POST | `/portfolios/{portfolioCode}/compliance/checks/post-trade` | `WORKFLOW_EXECUTE` | portfolio-scoped scan; contract-only V1 `PostTradeVerifier` is unchanged |
| GET | `/portfolios/{portfolioCode}/compliance/rules` | `IRG_VIEW_RULES` | active rule catalog + this portfolio's binding + current parameters |
| POST | `/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings` | `IRG_EDIT_BINDING` | scope_type=PORTFOLIO, scope_id=portfolio_id |
| DELETE | `/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}` | `IRG_EDIT_BINDING` | 404s if the binding belongs to a different portfolio |
| GET | `/portfolios/{portfolioCode}/compliance/breaches` | `IRG_VIEW_RULES` | breaches filtered by portfolio_id |

Key differences from V1:

- `compliance_check_records.contract_id` and `compliance_breaches.contract_id`
  are now nullable (migration `20260707000001`) — a portfolio-only check
  persists `contract_id = NULL`, not a zero UUID.
- `RunPreTradeCheckHandler`/`RunPostTradeCheckHandler` no longer require
  `contract_id` at the command layer; `buildScopes` already omitted the
  CONTRACT scope when `contract_id` is `uuid.Nil`, so no engine change was
  needed there. V1's HTTP handler (`compliance_handler.go`) still requires
  `contract_id` in the request body — that validation is unchanged.
- New rule binding CRUD (`create_rule_binding.go`, `deactivate_rule_binding.go`)
  fills the "no HTTP API for bindings" gap noted in §4/§19, scoped to
  PORTFOLIO bindings only. A partial unique index
  (`uq_compliance_rb_active_portfolio`, migration `20260707000002`) plus an
  application-layer check prevent two active bindings for the same rule
  instance on the same portfolio.
- New rule types: `allocation.asset_class_max`, `allocation.asset_class_min`,
  `exposure.max_order_percent_aum`, `valuation.min_nav`
  (`rules/allocation`, `rules/exposure`, `rules/valuation`). AUM and NAV are
  the same figure at the portfolio level in this codebase
  (`investment__valuation_snapshots.aum`, read via the existing
  `InvestmentMarketDataAdapter`) — `cash.availability`'s `min_cash_buffer_pct`
  already covered "cash buffer min %" from the original ask, so it was reused
  rather than duplicated.

**Known limitation:** `investment__portfolios.fund_id` is still `NOT NULL` in
the schema, so a portfolio with no fund cannot exist yet in production data —
the "GLOBAL + PORTFOLIO scope only" path is implemented and unit-tested
against fakes, but is not reachable end-to-end until a future migration
relaxes that column (out of scope here: the task explicitly excludes removing
`fund_id` globally, and doing so touches fund-scoped permission checks,
`WorkflowStateProvider.IsTradeAllowed`, and investment decision creation).

**Known deviation from the ideal lifecycle:** the business rule describes
`decision draft → manager approval → pre-trade compliance → execution`.
The current, tested `DecisionCommandHandler.Submit` implementation
(§6.1 above) runs pre-trade compliance *before* the `INVESTMENT_DECISION`
manager-approval request is created, not after. This was deliberately left
unchanged — reordering it touches the COMPLIANCE_RELEASE approval flow and
its P0 test suite (`decision_compliance_test.go`), which was out of scope for
an additive change. Portfolio Compliance V2's standalone pre/post-trade
endpoints are available now for any flow that wants to run compliance at a
different point; wiring them into the decision lifecycle's post-approval step
is a follow-up.

## 19. Known Gaps / TODOs

| Item | Priority | Notes |
|---|---|---|
| `regulatory.thai_sec` is a stub | High | Returns `VerdictWarn` with `NOT_CONFIGURED` evidence. Real Thai SEC / BOT regulatory logic is Phase 2. |
| `credit_rating.minimum` stub | Medium | Uses `NopCreditRatingAdapter` — always returns empty credit data. Real credit rating provider is Phase 2. |
| No `GET /compliance/rule-types` endpoint | Medium | Rule types are discoverable only from the Go registry. The frontend uses a static `ruleTypeCatalog.ts` mirror. |
| No CRUD endpoints for rule bindings | High | Rule bindings can only be managed via direct SQL or migrations. There is no HTTP API for creating, reading, updating, or deleting bindings. |
| No CRUD endpoints for rule sets | Medium | Same as above. |
| `OverrideBreachRequest.ApprovedBy` is nil | Medium | Second-level approval for override requires the approval engine callback. Currently not integrated — overrides are single-actor. |
| Per-line compliance checks for basket/rebalance/switch | High | Pre-trade checks on BASKET_ORDER, REBALANCE, and SWITCH run only at the header level (empty ticker). Per-instrument line checking is Phase 2. |
| `NopTradeHistoryAdapter` | Low | Trade history data dependency always returns empty. Rules that depend on `TradeHistory` cannot function correctly. |
| `NopCalendarAdapter` | Low | Calendar data dependency always returns nil. Rules using trading-calendar information cannot function correctly. |
| Periodic scheduled compliance scans | Low | No scheduler job exists for periodic post-trade scans. Post-trade checks are triggered on-demand only (by workflow close gate or HTTP). |
| Breach `RESOLVED` status | Low | Breaches can be marked RESOLVED when a subsequent check passes, but no automated resolution job exists. |
| Frontend: override dialog shows no approver field | Low | Field was removed (P0-9 security fix). Second-level approver integration is deferred. |

---

## 20. Acceptance Checklist

- [x] Pre-trade compliance check runs before investment decision approval
- [x] BLOCK verdict prevents decision from reaching approval engine
- [x] BLOCK + all-overridable breaches route through COMPLIANCE_RELEASE approval
- [x] BLOCK + any non-overridable breach rejects decision with HTTP 422
- [x] PASS and WARN verdicts allow decision to proceed to approval
- [x] Post-trade check wired to workflow close gate
- [x] Override requires written reason; actor is from JWT only
- [x] `regulatory.thai_sec` stub documented as NOT_CONFIGURED (returns WARN, not PASS)
- [x] Check records are append-only (DB `REVOKE UPDATE DELETE`)
- [x] Override records are append-only (DB `REVOKE UPDATE DELETE`)
- [x] All compliance routes require valid JWT
- [x] Dry-run simulation mode available (no DB writes)
- [ ] Real Thai SEC regulatory rule implemented
- [ ] Rule binding CRUD endpoints
- [ ] Per-line compliance checks for basket orders
