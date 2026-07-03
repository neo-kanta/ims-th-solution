/**
 * Typed wrappers for the intraday holdings valuation endpoints.
 *
 * Calls go through the generated OpenAPI client — see CLAUDE.md for the
 * no-manual-API-clients rule. Endpoints used:
 *
 *   GET  /investment/funds/{id}/holdings/valuation       — estimated NAV/AUM,
 *                                                          official NAV side-by-side,
 *                                                          positions with live price,
 *                                                          allocation
 *   POST /investment/funds/{id}/market-data/refresh      — force a provider
 *                                                          fetch for every
 *                                                          held instrument
 *   GET  /investment/funds/{id}/market-data/status       — feed health + stale
 *                                                          position count
 */
import { OpenApiRequestError, unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { components } from "~/api/ims-api";

/** Live holdings valuation response — the source of truth for the Holdings page KPIs. */
export type IntradayValuation = components["schemas"]["IntradayValuationResponse"];
export type IntradayPosition = components["schemas"]["IntradayPositionResponse"];
export type IntradayAllocationBucket = components["schemas"]["IntradayAllocationBucketResponse"];
export type IntradayCashRow = components["schemas"]["IntradayCashRowResponse"];

/** Refresh-result payload. */
export type MarketDataRefresh = components["schemas"]["MarketDataRefreshResponse"];

/** Provider health snapshot. */
export type MarketDataStatus = components["schemas"]["MarketDataStatusResponse"];

export const intradayValuationApi = {
  /**
   * Fetches the mark-to-market holdings valuation as of businessDate (YYYY-MM-DD).
   * Omitting businessDate defaults the backend to today (UTC). Returns null on
   * 404 (e.g. the fund exists but has no portfolios with positions) so the
   * page can render its empty state instead of an error banner.
   */
  async getFundValuation(fundId: string, businessDate?: string): Promise<IntradayValuation | null> {
    const client = useOpenApiClient();
    try {
      const response = await client.GET(
        "/investment/funds/{id}/holdings/valuation",
        {
          params: {
            path: { id: fundId },
            query: businessDate ? { business_date: businessDate } : undefined,
          },
        },
      );
      return unwrapOpenApiResponse<IntradayValuation>(response);
    } catch (err) {
      if (err instanceof OpenApiRequestError && err.status === 404) return null;
      throw err;
    }
  },

  /**
   * Triggers a synchronous provider fetch for every held instrument. The
   * backend records an audit event. Returns the per-symbol success / stale /
   * failure counts. Network errors propagate so the page can show an error
   * toast, but provider outages return a 200 with FailedSymbols populated —
   * the page treats that as "stale" rather than fatal.
   */
  async refreshFundQuotes(fundId: string): Promise<MarketDataRefresh> {
    const client = useOpenApiClient();
    const response = await client.POST(
      "/investment/funds/{id}/market-data/refresh",
      { params: { path: { id: fundId } } },
    );
    return unwrapOpenApiResponse<MarketDataRefresh>(response);
  },

  /**
   * Reads the live feed status: primary provider, total positions, how many
   * are stale, when the last good quote landed. Used by the "Feeds OK" badge.
   */
  async getFundFeedStatus(fundId: string): Promise<MarketDataStatus | null> {
    const client = useOpenApiClient();
    try {
      const response = await client.GET(
        "/investment/funds/{id}/market-data/status",
        { params: { path: { id: fundId } } },
      );
      return unwrapOpenApiResponse<MarketDataStatus>(response);
    } catch (err) {
      if (err instanceof OpenApiRequestError && err.status === 404) return null;
      throw err;
    }
  },
};
