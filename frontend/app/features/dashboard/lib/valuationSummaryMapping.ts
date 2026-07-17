import type { components } from "~/api/ims-api";

import type { ValuationSummaryDTO } from "../types";

// Kept in its own module (no runtime `~/api/openapi` import) so it can be
// unit tested in plain Vitest — see tests/dashboard-valuation-summary.test.ts.
// dashboardApi.ts re-exports this for callers that need the wire mapper
// alongside the live client.
type GeneratedValuationSummary = components["schemas"]["ValuationSummaryDTO"];

export function normalizeValuationSummary(
  payload: GeneratedValuationSummary | undefined,
): ValuationSummaryDTO {
  return {
    scope: payload?.scope === "mine" ? "mine" : "company",
    username: payload?.username ?? null,
    businessDate: payload?.business_date ?? "",
    currency: payload?.currency ?? "",
    aumToday: payload?.aum_today ?? "",
    todayPnl: payload?.today_pnl ?? "",
    todayPnlPercent: payload?.today_pnl_percent ?? null,
    asOf: payload?.as_of ?? "",
    dataAvailable: Boolean(payload?.data_available),
  };
}
