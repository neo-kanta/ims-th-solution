/**
 * Pure filter-state helpers for the Breaches queue toolbar.
 *
 * The backend only supports server-side filtering by portfolio_id,
 * contract_id, status, rule_type_id, date_from, and date_to (see
 * `GET /compliance/breaches`). Severity, message search, and sorting are
 * intentionally absent here because the API has no equivalent parameter —
 * inventing a client-side version of them would silently lie about what the
 * "filter" is scoped to (current page vs. the full result set).
 *
 * Contract filtering is also intentionally absent: the only contract
 * identity available client-side today is a raw UUID list
 * (`authStore.permissions.contracts`), and the UX requirement is that a
 * filter must resolve to a human-readable label before it can be exposed.
 */
import type { ComplianceBreachListFilters, ComplianceBreachStatus } from "../types";

export type BreachStatusChoice = ComplianceBreachStatus | "ALL";

export interface BreachFilterState {
  status: BreachStatusChoice;
  ruleTypeId: string;
  portfolioId: string;
  dateFrom: string;
  dateTo: string;
}

/** "Open" is the operationally useful default — it is the attention queue. */
export const DEFAULT_BREACH_STATUS: BreachStatusChoice = "OPEN";

export function defaultBreachFilterState(): BreachFilterState {
  return {
    status: DEFAULT_BREACH_STATUS,
    ruleTypeId: "",
    portfolioId: "",
    dateFrom: "",
    dateTo: "",
  };
}

const VALID_STATUSES: BreachStatusChoice[] = ["OPEN", "OVERRIDDEN", "RESOLVED", "ALL"];

function queryString(value: unknown): string {
  if (Array.isArray(value)) return typeof value[0] === "string" ? value[0] : "";
  return typeof value === "string" ? value : "";
}

/**
 * Reads the current filter state from `route.query`. Missing/unknown status
 * falls back to the default (OPEN) rather than "ALL" so a bare `/post-trade`
 * visit lands on the attention queue, matching `defaultBreachFilterState`.
 */
export function filterStateFromQuery(
  query: Record<string, unknown>,
): BreachFilterState {
  const rawStatus = queryString(query.status).toUpperCase();
  const status = (VALID_STATUSES as string[]).includes(rawStatus)
    ? (rawStatus as BreachStatusChoice)
    : DEFAULT_BREACH_STATUS;

  return {
    status,
    ruleTypeId: queryString(query.rule_type_id),
    portfolioId: queryString(query.portfolio_id),
    dateFrom: queryString(query.date_from),
    dateTo: queryString(query.date_to),
  };
}

/** Builds a route query object. Empty optional fields are omitted entirely. */
export function queryFromFilterState(
  state: BreachFilterState,
): Record<string, string> {
  const query: Record<string, string> = { status: state.status };
  if (state.ruleTypeId.trim()) query.rule_type_id = state.ruleTypeId.trim();
  if (state.portfolioId.trim()) query.portfolio_id = state.portfolioId.trim();
  if (state.dateFrom) query.date_from = state.dateFrom;
  if (state.dateTo) query.date_to = state.dateTo;
  return query;
}

/** Maps toolbar state to the exact `GET /compliance/breaches` query filters. */
export function apiFiltersFromState(
  state: BreachFilterState,
): ComplianceBreachListFilters {
  const filters: ComplianceBreachListFilters = {};
  if (state.status !== "ALL") filters.status = state.status;
  if (state.ruleTypeId.trim()) filters.rule_type_id = state.ruleTypeId.trim();
  if (state.portfolioId.trim()) filters.portfolio_id = state.portfolioId.trim();
  if (state.dateFrom) filters.date_from = state.dateFrom;
  if (state.dateTo) filters.date_to = state.dateTo;
  return filters;
}

/** True when any filter departs from the default queue view. */
export function hasNonDefaultFilters(state: BreachFilterState): boolean {
  return (
    state.status !== DEFAULT_BREACH_STATUS ||
    state.ruleTypeId.trim() !== "" ||
    state.portfolioId.trim() !== "" ||
    state.dateFrom !== "" ||
    state.dateTo !== ""
  );
}

export function breachFilterStatesEqual(
  a: BreachFilterState,
  b: BreachFilterState,
): boolean {
  return (
    a.status === b.status &&
    a.ruleTypeId === b.ruleTypeId &&
    a.portfolioId === b.portfolioId &&
    a.dateFrom === b.dateFrom &&
    a.dateTo === b.dateTo
  );
}
