/**
 * Market Data — typed API client.
 *
 * Live backend endpoints exposed under /api/v1/market-data:
 *   GET  /market-data/quote?symbol=&provider=
 *   GET  /market-data/history?symbol=&provider=&limit=
 *   POST /market-data/import        body: ImportMarketDataRequest
 *   GET  /market-data/provider-health
 *   POST /market-data/import-batches
 *   POST /market-data/import-batches/{batch_id}/run
 *   GET  /market-data/import-batches/{batch_id}
 *   GET  /market-data/import-batches/{batch_id}/errors
 *   GET  /market-data/screen/watchlist
 *   GET  /market-data/screen/search?query=
 *
 * All requests go through the generated OpenAPI client; do NOT add manual type
 * definitions or hand-rolled fetch calls here.
 */
import { unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { paths, components } from "~/api/ims-api";

export type ApiQuote = components["schemas"]["QuoteResponse"];
export type ApiPriceBar = components["schemas"]["PriceBarResponse"];
export type ApiProviderHealth = components["schemas"]["ProviderHealthResponse"];
export type ApiImportResult = components["schemas"]["ImportMarketDataResponse"];
export type ApiImportRequest = components["schemas"]["ImportMarketDataRequest"];

export type ApiScreenRow = components["schemas"]["MarketDataScreenRowDTO"];
export type ApiScreenSearchRow = components["schemas"]["MarketDataScreenSearchItemDTO"];
export type ApiProviderBadge = components["schemas"]["ProviderBadgeDTO"];
export type ApiScreenWatchlistResponse = components["schemas"]["ScreenWatchlistResponseDTO"];
export type ApiScreenSearchResponse = components["schemas"]["ScreenSearchResponseDTO"];

export type ApiCreateBatchRequest = components["schemas"]["CreateImportBatchRequest"];
export type ApiCreateBatchResponse = components["schemas"]["CreateImportBatchResponse"];
export type ApiRunBatchResponse = components["schemas"]["RunImportBatchResponse"];
export type ApiBatchStatusResponse = components["schemas"]["ImportBatchStatusResponseDTO"];
export type ApiImportChunkItem = components["schemas"]["ImportChunkItem"];
export type ApiImportChunkSummary = components["schemas"]["ImportChunkSummary"];

type HistoryQuery = NonNullable<
  paths["/market-data/history"]["get"]["parameters"]["query"]
>;
type QuoteQuery = NonNullable<
  paths["/market-data/quote"]["get"]["parameters"]["query"]
>;
type ScreenWatchlistQuery = NonNullable<
  paths["/market-data/screen/watchlist"]["get"]["parameters"]["query"]
>;
type ScreenSearchQuery = NonNullable<
  paths["/market-data/screen/search"]["get"]["parameters"]["query"]
>;

export const marketDataApi = {
  async getQuote(symbol: string, provider?: string): Promise<ApiQuote> {
    const client = useOpenApiClient();
    const query: QuoteQuery = { symbol };
    if (provider) query.provider = provider;
    const response = await client.GET("/market-data/quote", { params: { query } });
    return unwrapOpenApiResponse<ApiQuote>(response);
  },

  async getHistory(symbol: string, limit = 30, provider?: string): Promise<ApiPriceBar[]> {
    const client = useOpenApiClient();
    const query: HistoryQuery = { symbol, limit };
    if (provider) query.provider = provider;
    const response = await client.GET("/market-data/history", { params: { query } });
    return unwrapOpenApiResponse<ApiPriceBar[]>(response);
  },

  async importSymbol(req: ApiImportRequest): Promise<ApiImportResult> {
    const client = useOpenApiClient();
    const response = await client.POST("/market-data/import", { body: req });
    return unwrapOpenApiResponse<ApiImportResult>(response);
  },

  async providerHealth(): Promise<ApiProviderHealth[]> {
    const client = useOpenApiClient();
    const response = await client.GET("/market-data/provider-health");
    return unwrapOpenApiResponse<ApiProviderHealth[]>(response);
  },

  async screenWatchlist(limit?: number): Promise<ApiScreenWatchlistResponse> {
    const client = useOpenApiClient();
    const query: ScreenWatchlistQuery = {};
    if (limit) query.limit = limit;
    const response = await client.GET("/market-data/screen/watchlist", { params: { query } });
    return unwrapOpenApiResponse<ApiScreenWatchlistResponse>(response);
  },

  async screenSearch(queryText: string, limit?: number): Promise<ApiScreenSearchResponse> {
    const client = useOpenApiClient();
    const query: ScreenSearchQuery = { query: queryText };
    if (limit) query.limit = limit;
    const response = await client.GET("/market-data/screen/search", { params: { query } });
    return unwrapOpenApiResponse<ApiScreenSearchResponse>(response);
  },

  // ----- chunk-based import batches -----

  async createImportBatch(req: ApiCreateBatchRequest): Promise<ApiCreateBatchResponse> {
    const client = useOpenApiClient();
    const response = await client.POST("/market-data/import-batches", { body: req });
    return unwrapOpenApiResponse<ApiCreateBatchResponse>(response);
  },

  async runImportBatch(batchId: string): Promise<ApiRunBatchResponse> {
    const client = useOpenApiClient();
    const response = await client.POST("/market-data/import-batches/{batch_id}/run", {
      params: { path: { batch_id: batchId } },
    });
    return unwrapOpenApiResponse<ApiRunBatchResponse>(response);
  },

  async getImportBatch(batchId: string): Promise<ApiBatchStatusResponse> {
    const client = useOpenApiClient();
    const response = await client.GET("/market-data/import-batches/{batch_id}", {
      params: { path: { batch_id: batchId } },
    });
    return unwrapOpenApiResponse<ApiBatchStatusResponse>(response);
  },

  async getImportBatchErrors(batchId: string): Promise<ApiImportChunkItem[]> {
    const client = useOpenApiClient();
    const response = await client.GET("/market-data/import-batches/{batch_id}/errors", {
      params: { path: { batch_id: batchId } },
    });
    return unwrapOpenApiResponse<ApiImportChunkItem[]>(response);
  },
};
