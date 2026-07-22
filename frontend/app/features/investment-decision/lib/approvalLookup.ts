import type { DecisionApprovalFilters } from "../services/decisionApprovalApi";

const ROUTE_FILTER_KEYS = [
  "decision_no",
  "search",
  "process_type",
  "product_type",
  "business_date_from",
  "business_date_to",
  "portfolio_code",
] as const satisfies readonly (keyof DecisionApprovalFilters)[];

function firstString(value: unknown): string {
  if (typeof value === "string") return value;
  if (Array.isArray(value) && typeof value[0] === "string") return value[0];
  return "";
}

/**
 * Keeps OP-02 on its approval-only contract and removes empty query values.
 * The user-facing decision number maps to `decision_no`, not the internal
 * decision UUID.
 */
export function cleanApprovalFilters(
  filters: DecisionApprovalFilters,
): DecisionApprovalFilters {
  const cleaned: DecisionApprovalFilters = {};

  for (const key of ROUTE_FILTER_KEYS) {
    const value = filters[key];
    if (typeof value !== "string") continue;
    const trimmed = value.trim();
    if (trimmed) cleaned[key] = trimmed;
  }

  return cleaned;
}

export function approvalFiltersFromQuery(
  query: Record<string, unknown>,
): DecisionApprovalFilters {
  const filters: DecisionApprovalFilters = {};

  for (const key of ROUTE_FILTER_KEYS) {
    filters[key] = firstString(query[key]);
  }

  return cleanApprovalFilters(filters);
}

export function approvalFiltersToQuery(
  filters: DecisionApprovalFilters,
): Record<string, string> {
  return cleanApprovalFilters(filters) as Record<string, string>;
}
