/**
 * Static mirror of the backend SPI rule registry — Phase 1.
 *
 * This is NOT a substitute for a real `GET /compliance/rule-types` endpoint
 * (Missing API #7). It exists so the Pre-trade panel can render a meaningful
 * human label and suggested correction for the rule_type_ids that today's
 * backend can actually return. The strings here are derived from each
 * self-registered rule package under `backend/internal/compliance/rules/`.
 *
 * If a backend response includes an unknown rule_type_id, the panel falls
 * back to the raw id rather than inventing copy.
 */
import type { ComplianceRuleCategory } from "../types";

export interface RuleCatalogEntry {
  typeId: string;
  label: string;
  category: ComplianceRuleCategory;
  /** Plain-English reason a portfolio manager can understand. */
  explanation: string;
  /** A safe, generic correction that does not invent specific numbers. */
  suggestedCorrection: string;
}

export const RULE_CATALOG: readonly RuleCatalogEntry[] = [
  {
    typeId: "cash.availability",
    label: "Available cash check",
    category: "MANDATE",
    explanation:
      "Buy order requires more cash than is available in the portfolio for the business date.",
    suggestedCorrection:
      "Reduce order quantity, raise cash (sell other positions), or stage the order for after settlement.",
  },
  {
    typeId: "concentration.single_issuer",
    label: "Maximum single-issuer exposure",
    category: "MANDATE",
    explanation:
      "Combined exposure to the issuer (and any grouped parent entity) exceeds the configured % of NAV.",
    suggestedCorrection:
      "Reduce the order so total issuer exposure stays below the configured cap, or rebalance other positions first.",
  },
  {
    typeId: "credit.min_rating",
    label: "Minimum credit rating",
    category: "RESTRICTION",
    explanation:
      "Instrument's credit rating is below the minimum permitted for this portfolio or mandate.",
    suggestedCorrection:
      "Choose an instrument that meets the minimum rating, or request a written mandate exception before execution.",
  },
  {
    typeId: "credit_rating.minimum",
    label: "Minimum credit rating (stub)",
    category: "RESTRICTION",
    explanation:
      "Stub credit-rating rule — instrument rating does not satisfy the configured threshold.",
    suggestedCorrection:
      "Pick a rating-compliant instrument or request an exception with risk acknowledgement.",
  },
  {
    typeId: "quantity.min_trading_unit",
    label: "Minimum trading unit",
    category: "MANDATE",
    explanation:
      "Order quantity is below the venue's minimum lot or the configured trading-unit floor.",
    suggestedCorrection:
      "Increase the order quantity to a multiple of the minimum lot size for this market.",
  },
  {
    typeId: "quantity.sell_available",
    label: "Available-to-sell quantity",
    category: "MANDATE",
    explanation:
      "Sell order exceeds the position currently available to sell (after pending sells / settlement holds).",
    suggestedCorrection:
      "Lower the sell quantity to the available-to-sell figure, or wait for pending sells to settle.",
  },
  {
    typeId: "ratio.sector_exposure",
    label: "Maximum sector exposure",
    category: "RATIO",
    explanation:
      "Sector exposure after the order would exceed the configured percentage of NAV.",
    suggestedCorrection:
      "Reduce the order, switch to a different sector, or rebalance prior holdings to free up sector capacity.",
  },
  {
    typeId: "regulatory.thai_sec",
    label: "Thai SEC regulatory check (stub)",
    category: "REGULATORY",
    explanation:
      "Stub regulatory check — order violates a configured Thai SEC parameter.",
    suggestedCorrection:
      "Review the SEC parameter list with compliance before re-submitting.",
  },
  {
    typeId: "restriction.blacklist",
    label: "Restricted security blacklist",
    category: "RESTRICTION",
    explanation:
      "Ticker is on the active blacklist (sanctions, banned issuers, internal blocks).",
    suggestedCorrection:
      "Select an unrestricted ticker. Blacklist entries are not overridable from the trading desk.",
  },
  {
    typeId: "restriction.whitelist",
    label: "Whitelist-only investment",
    category: "RESTRICTION",
    explanation:
      "Mandate permits only whitelisted securities, and the proposed ticker is not on the list.",
    suggestedCorrection:
      "Pick a ticker from the mandate whitelist, or request an exception to add it.",
  },
  {
    typeId: "restriction.list_enforcement",
    label: "Restricted-list enforcement",
    category: "RESTRICTION",
    explanation:
      "Combined restricted-list rules (whitelist + blacklist) flagged this order.",
    suggestedCorrection:
      "Check both the whitelist and blacklist; pick a permitted ticker or seek compliance exception.",
  },
] as const;

const CATALOG_INDEX: Record<string, RuleCatalogEntry> = Object.fromEntries(
  RULE_CATALOG.map((entry) => [entry.typeId, entry]),
);

export function lookupRuleCatalog(typeId: string): RuleCatalogEntry | null {
  return CATALOG_INDEX[typeId] ?? null;
}

export function ruleLabel(typeId: string): string {
  return lookupRuleCatalog(typeId)?.label ?? typeId;
}

export function ruleExplanation(typeId: string, fallback: string): string {
  const entry = lookupRuleCatalog(typeId);
  return entry?.explanation ?? fallback;
}

export function ruleSuggestedCorrection(typeId: string): string | null {
  return lookupRuleCatalog(typeId)?.suggestedCorrection ?? null;
}
