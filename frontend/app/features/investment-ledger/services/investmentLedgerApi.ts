/**
 * Investment Ledger — typed API client for the Simulated Order Service (S6).
 *
 * Live backend endpoints exposed under /api/v1/investment:
 *   GET  /investment/portfolios                                — list portfolios
 *   GET  /investment/portfolios/{id}                          — portfolio detail
 *   GET  /investment/portfolios/{id}/holdings                 — positions
 *   GET  /investment/portfolios/{id}/cash                     — cash by currency
 *   GET  /investment/portfolios/{id}/transactions             — ledger
 *   POST /investment/portfolios/{id}/transactions/simulate    — pre-trade preview
 *   POST /investment/portfolios/{id}/transactions             — post to ledger
 *   GET  /investment/instruments                              — instrument search
 *
 * All requests go through the generated OpenAPI client; do NOT add manual type
 * definitions or hand-rolled fetch calls here.
 */
import { unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { components, paths } from "~/api/ims-api";

export type ApiPortfolio = components["schemas"]["PortfolioResponse"];
export type ApiPortfolioList = components["schemas"]["PortfolioListResponse"];

export type ApiInstrument = components["schemas"]["InstrumentResponse"];
export type ApiInstrumentList = components["schemas"]["InstrumentListResponse"];

export type ApiHolding = components["schemas"]["HoldingResponse"];
export type ApiCashBalance = components["schemas"]["CashBalanceResponse"];

export type ApiTransaction = components["schemas"]["TransactionResponse"];
export type ApiTransactionList = components["schemas"]["TransactionListResponse"];

export type ApiPostTransactionRequest =
  components["schemas"]["PostTransactionRequest"];
export type ApiTransactionSimulation =
  components["schemas"]["TransactionSimulationResponse"];

export type ApiCashProjection = components["schemas"]["CashProjectionResponse"];
export type ApiPositionProjection =
  components["schemas"]["PositionProjectionResponse"];
export type ApiCompliancePreview =
  components["schemas"]["CompliancePreviewResponse"];
export type ApiComplianceBreachPreview =
  components["schemas"]["ComplianceBreachPreviewResponse"];

type PortfolioListQuery = NonNullable<
  paths["/investment/portfolios"]["get"]["parameters"]["query"]
>;

type TransactionListQuery = NonNullable<
  paths["/investment/portfolios/{id}/transactions"]["get"]["parameters"]["query"]
>;

type InstrumentListQuery = NonNullable<
  paths["/investment/instruments"]["get"]["parameters"]["query"]
>;

export interface PortfolioListFilters {
  fund_id?: string;
  status?: string;
  page?: number;
  limit?: number;
}

export interface TransactionListFilters {
  instrument_id?: string;
  from?: string;
  to?: string;
  page?: number;
  limit?: number;
}

export interface InstrumentListFilters {
  asset_class_id?: string;
  search?: string;
  page?: number;
  limit?: number;
}

function applyDefined<T extends Record<string, unknown>>(
  target: T,
  source: Partial<T>,
): T {
  for (const [key, value] of Object.entries(source)) {
    if (value === undefined || value === null) continue;
    if (typeof value === "string" && value.trim() === "") continue;
    (target as Record<string, unknown>)[key] = value;
  }
  return target;
}

export const investmentLedgerApi = {
  async listPortfolios(
    filters: PortfolioListFilters = {},
  ): Promise<ApiPortfolioList> {
    const client = useOpenApiClient();
    const query = applyDefined<PortfolioListQuery>({}, filters);
    const response = await client.GET("/investment/portfolios", {
      params: { query },
    });
    return unwrapOpenApiResponse<ApiPortfolioList>(response);
  },

  async getPortfolio(id: string): Promise<ApiPortfolio> {
    const client = useOpenApiClient();
    const response = await client.GET("/investment/portfolios/{id}", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse<ApiPortfolio>(response);
  },

  async listHoldings(portfolioId: string): Promise<ApiHolding[]> {
    const client = useOpenApiClient();
    const response = await client.GET(
      "/investment/portfolios/{id}/holdings",
      { params: { path: { id: portfolioId } } },
    );
    return unwrapOpenApiResponse<ApiHolding[]>(response);
  },

  async listCash(portfolioId: string): Promise<ApiCashBalance[]> {
    const client = useOpenApiClient();
    const response = await client.GET(
      "/investment/portfolios/{id}/cash",
      { params: { path: { id: portfolioId } } },
    );
    return unwrapOpenApiResponse<ApiCashBalance[]>(response);
  },

  async listTransactions(
    portfolioId: string,
    filters: TransactionListFilters = {},
  ): Promise<ApiTransactionList> {
    const client = useOpenApiClient();
    const query = applyDefined<TransactionListQuery>({}, filters);
    const response = await client.GET(
      "/investment/portfolios/{id}/transactions",
      { params: { path: { id: portfolioId }, query } },
    );
    return unwrapOpenApiResponse<ApiTransactionList>(response);
  },

  async simulateTransaction(
    portfolioId: string,
    payload: ApiPostTransactionRequest,
  ): Promise<ApiTransactionSimulation> {
    const client = useOpenApiClient();
    const response = await client.POST(
      "/investment/portfolios/{id}/transactions/simulate",
      { params: { path: { id: portfolioId } }, body: payload },
    );
    return unwrapOpenApiResponse<ApiTransactionSimulation>(response);
  },

  async postTransaction(
    portfolioId: string,
    payload: ApiPostTransactionRequest,
  ): Promise<ApiTransaction> {
    const client = useOpenApiClient();
    const response = await client.POST(
      "/investment/portfolios/{id}/transactions",
      { params: { path: { id: portfolioId } }, body: payload },
    );
    return unwrapOpenApiResponse<ApiTransaction>(response);
  },

  async listInstruments(
    filters: InstrumentListFilters = {},
  ): Promise<ApiInstrumentList> {
    const client = useOpenApiClient();
    const query = applyDefined<InstrumentListQuery>({}, filters);
    const response = await client.GET("/investment/instruments", {
      params: { query },
    });
    return unwrapOpenApiResponse<ApiInstrumentList>(response);
  },

  /** Fetch one instrument by its canonical ID (used to enrich holdings, which only carry instrument_id). */
  async getInstrument(id: string): Promise<ApiInstrument> {
    const client = useOpenApiClient();
    const response = await client.GET("/investment/instruments/{id}", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse<ApiInstrument>(response);
  },
};
