/**
 * Pure discriminator between the two possible POST /transactions response
 * shapes (see portfolioApi.postTransaction). Kept in a dependency-free file
 * (no `~/` Nuxt-aliased imports) so it — and anything that only needs it —
 * can be imported directly in plain Vitest without needing to mock the
 * Nuxt runtime client that portfolioApi.ts depends on (same pattern as
 * dashboard/lib/valuationSummaryMapping.ts).
 */
import type { components } from "~/api/ims-api";

export type ApiTransactionV2 = components["schemas"]["TransactionResponse"];
export type ApiCashRequestV2 = components["schemas"]["CashRequestResponse"];

/**
 * POST /transactions returns either a posted TransactionResponse (201) or,
 * for a LIVE cash movement, a pending CashRequestResponse (202) — see
 * PostTransactionByCode's dual @Success annotation. `submitted_by` only
 * exists on CashRequestResponse, so it is a safe runtime discriminator
 * between the two response shapes.
 */
export function isCashRequestResponse(
  value: ApiTransactionV2 | ApiCashRequestV2,
): value is ApiCashRequestV2 {
  return (
    typeof value === "object" && value !== null && "submitted_by" in value
  );
}
