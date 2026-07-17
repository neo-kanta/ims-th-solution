/**
 * Compliance / IRG domain types — Phase 1.
 *
 * Aligned to the OpenAPI types in ~/api/ims-api.d.ts. The backend serialises
 * entity fields in camelCase (Go field names) and request/response wrappers
 * in snake_case — we preserve that distinction here so the wire shape and the
 * TS shape match without runtime renames.
 */

export type ComplianceVerdict = "PASS" | "WARN" | "BLOCK";

/** Action-level severity assigned at the binding. Mirrors backend value object. */
export type ComplianceBackendSeverity =
  | "BLOCK"
  | "WARN"
  | "REQUIRE_APPROVAL"
  | "MONITOR";

/** UI severity vocabulary used on badges, per the design spec. */
export type ComplianceUiSeverity = "INFO" | "WARNING" | "BLOCKER" | "CRITICAL";

export type ComplianceRuleCategory =
  | "REGULATORY"
  | "MANDATE"
  | "HOUSE"
  | "CLIENT"
  | "RESTRICTION"
  | "RATIO"
  | "TEMPORAL"
  | "BEHAVIORAL";

export type ComplianceCheckTiming = "PRE_TRADE" | "POST_TRADE" | "PERIODIC";

/** Derived from is_active + effective window. Not a real backend enum. */
export type ComplianceRuleDerivedStatus =
  | "ACTIVE"
  | "SCHEDULED"
  | "EXPIRED"
  | "DISABLED";

export type ComplianceBreachStatus = "OPEN" | "OVERRIDDEN" | "RESOLVED";

/**
 * Effective window — backend value object. The generated OpenAPI typing
 * is currently an opaque `Record<string, never>`; we render defensively
 * and treat missing dates as "—".
 */
export interface ComplianceEffectiveWindow {
  from?: string | null;
  to?: string | null;
}

/** Mirrors backend spi.RuleMetadata (snake_case JSON). */
export interface ComplianceRuleTypeMetadata {
  type_id: string;
  version: string;
  category: ComplianceRuleCategory;
  default_severity: ComplianceBackendSeverity;
  supported_timings: ComplianceCheckTiming[];
  supported_scopes: string[];
  overridable: boolean;
  description: string;
}

/** Mirrors backend query.RuleInstanceDetail (camelCase entity + snake_case metadata). */
export interface ComplianceRule {
  id: string;
  ruleTypeID: string;
  name: string;
  description: string;
  currentVersion: number;
  isActive: boolean;
  effectiveWindow?: ComplianceEffectiveWindow;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
  type_metadata?: ComplianceRuleTypeMetadata;
}

export interface ComplianceRuleListResponse {
  instances: ComplianceRule[];
  total: number;
  offset: number;
  limit: number;
}

export interface ComplianceRuleListFilters {
  rule_type_id?: string;
  is_active?: boolean;
  offset?: number;
  limit?: number;
}

/** Mirrors backend entity.Breach (camelCase). */
export interface ComplianceBreach {
  id: string;
  checkRecordID: string;
  checkGroupID: string;
  portfolioID: string;
  contractID?: string;
  ruleTypeID: string;
  ruleInstanceID: string;
  severity: ComplianceBackendSeverity;
  verdict: ComplianceVerdict;
  status: ComplianceBreachStatus;
  evidence?: Record<string, unknown>;
  message: string;
  businessDate: string;
  createdAt: string;
  resolvedAt?: string | null;
  resolvedBy?: string | null;
}

export interface ComplianceBreachListResponse {
  breaches: ComplianceBreach[];
  total: number;
  offset: number;
  limit: number;
}

export interface ComplianceBreachListFilters {
  portfolio_id?: string;
  contract_id?: string;
  status?: ComplianceBreachStatus;
  rule_type_id?: string;
  date_from?: string;
  date_to?: string;
  offset?: number;
  limit?: number;
}

/** Mirrors backend command.BreachSummary (snake_case). */
export interface ComplianceCheckBreachSummary {
  breach_id: string;
  rule_type_id: string;
  severity: ComplianceBackendSeverity;
  verdict: ComplianceVerdict;
  message: string;
  overridable: boolean;
}

/** Matches handler.PreTradeRequest. All strings; backend parses UUIDs/decimals. */
export interface CompliancePreTradeRequest {
  portfolio_id: string;
  contract_id: string;
  business_date: string;
  order_id: string;
  ticker: string;
  side: "BUY" | "SELL";
  quantity: string;
  price: string;
  currency: string;
  exchange: string;
}

/** Mirrors command.PreTradeCheckResponse. */
export interface CompliancePreTradeResponse {
  check_group_id: string;
  verdict: ComplianceVerdict;
  rules_evaluated: number;
  total_duration_ms: number;
  breaches?: ComplianceCheckBreachSummary[];
}

/**
 * Portfolio option used by the simulator's picker. We pull this from the real
 * `/investment/portfolios` endpoint rather than asking the user to type a UUID.
 */
export interface CompliancePortfolioOption {
  id: string;
  code: string;
  name: string;
  base_currency: string;
  fund_id: string;
  status?: string;
}

/** Response shape of `GET /investment/portfolios`. */
export interface CompliancePortfolioListResponse {
  items: CompliancePortfolioOption[];
  total: number;
  page: number;
  limit: number;
}

/**
 * Mirrors backend entity.Override (camelCase JSON). Returned when a compliance
 * officer overrides an OPEN breach via POST /compliance/breaches/{id}/override.
 */
export interface ComplianceOverride {
  id: string;
  breachID: string;
  reason: string;
  overriddenBy: string;
  delegatedFrom?: string | null;
  approvedBy?: string | null;
  createdAt: string;
}

/** Mirrors handler.OverrideRequest. Actor is derived from JWT on the backend — never sent in body. */
export interface ComplianceOverrideRequest {
  reason: string;
}

/**
 * Mirrors backend query.CheckGroupResult — the records + breaches captured
 * by one pre/post-trade check invocation.
 */
export interface ComplianceCheckRecord {
  id: string;
  checkGroupID: string;
  timing: string;
  orderID?: string | null;
  portfolioID: string;
  contractID: string;
  ticker?: string;
  ruleTypeID: string;
  ruleInstanceID: string;
  ruleInstanceVersion: number;
  parameterSnapshot?: Record<string, unknown>;
  verdict: ComplianceVerdict;
  effectiveSeverity: ComplianceBackendSeverity;
  finalVerdict: ComplianceVerdict;
  evidence?: Record<string, unknown>;
  message: string;
  dataSnapshotHash?: string;
  evalDurationMs?: number;
  checkedBy?: string;
  businessDate: string;
  checkedAt: string;
  createdAt: string;
}

export interface ComplianceCheckGroupResult {
  check_group_id: string;
  records: ComplianceCheckRecord[];
  breaches: ComplianceBreach[];
}
