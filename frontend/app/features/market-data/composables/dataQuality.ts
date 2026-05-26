/**
 * Pure data-quality derivation for the security-detail view. Lives in its own
 * file (no Nuxt imports) so vitest can exercise it directly without spinning
 * up the Nuxt runtime.
 */
import type {
  ProviderMappingView,
  MarketQuoteView,
  SecurityDataQuality,
  DetailFreshnessStatus,
  DetailMappingStatus,
} from "../market-data.types";

export function deriveMappingStatus(mappings: ProviderMappingView[]): DetailMappingStatus {
  if (mappings.length === 0) return "UNMAPPED";
  if (mappings.some((m) => m.status === "ACTIVE")) return "MAPPED";
  if (mappings.some((m) => m.status === "REVIEW_REQUIRED")) return "REVIEW_REQUIRED";
  if (mappings.some((m) => m.status === "CONFLICTED")) return "CONFLICTED";
  return "UNMAPPED";
}

export function deriveDataQuality(
  mappings: ProviderMappingView[],
  quote: MarketQuoteView | undefined,
): SecurityDataQuality {
  const mapping = deriveMappingStatus(mappings);
  if (!quote) {
    return {
      freshness: "NOT_IMPORTED",
      mapping,
      providerStatus: "—",
    };
  }
  let freshness: DetailFreshnessStatus = "FRESH";
  let providerStatus = quote.provider ?? "—";
  let errorCode: string | undefined;
  let errorMessage: string | undefined;
  let rateLimited = false;

  if (quote.stale) {
    const reason = (quote.staleReason ?? "").toLowerCase();
    if (reason.includes("rate") || reason.includes("429")) {
      freshness = "RATE_LIMITED";
      rateLimited = true;
      errorCode = "RATE_LIMITED";
      errorMessage = quote.staleReason ?? undefined;
    } else if (reason.includes("fail") || reason.includes("error")) {
      freshness = "FAILED";
      errorMessage = quote.staleReason ?? undefined;
    } else {
      freshness = "STALE";
    }
    providerStatus = `${providerStatus} · stale`;
  }
  return {
    freshness,
    mapping,
    providerStatus,
    lastSyncedAt: quote.capturedAt ?? quote.asOf,
    lastSyncResult: quote.stale ? "warning" : "success",
    errorCode,
    errorMessage,
    rateLimited,
  };
}
