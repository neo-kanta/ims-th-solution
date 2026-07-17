/**
 * Typed wrapper for Portfolio V2 routes (docs/api/portfolio-v2-api-ddd.md,
 * docs/frontend/portfolio-v2-frontend-ddd.md). Every call goes through the
 * generated OpenAPI client — see CLAUDE.md for the no-manual-API-clients
 * rule.
 *
 * V2 identifies a portfolio by its business `code` in the URL, never by
 * UUID. More portfolio-scoped routes land here as the migration proceeds
 * (see the migration checklist in the DDD doc).
 *
 * Endpoints used:
 *   GET /api/v2/portfolios/{portfolioCode}              — portfolio detail by code
 *   GET /api/v2/portfolios/{portfolioCode}/holdings      — current holdings
 *   GET /api/v2/portfolios/{portfolioCode}/cash          — cash balances
 *   GET /api/v2/portfolios/{portfolioCode}/transactions  — ledger history
 */
import { unwrapOpenApiResponse, useOpenApiClientV2 } from "~/api/openapi";
import type { components } from "~/api/ims-api";

export type ApiPortfolioV2 = components["schemas"]["PortfolioResponse"];
export type ApiHoldingV2 = components["schemas"]["HoldingResponse"];
export type ApiCashBalanceV2 = components["schemas"]["CashBalanceResponse"];
export type ApiTransactionV2 = components["schemas"]["TransactionResponse"];
export type ApiTransactionListV2 =
  components["schemas"]["TransactionListResponse"];
export type ApiValuationV2 = components["schemas"]["ValuationResponse"];
export type ApiValuationListV2 = components["schemas"]["ValuationListResponse"];

export const portfolioApi = {
  /** Resolve a portfolio by its business code. Throws on 404. */
  async getByCode(portfolioCode: string): Promise<ApiPortfolioV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET("/portfolios/{portfolioCode}", {
      params: { path: { portfolioCode } },
    });
    return unwrapOpenApiResponse<ApiPortfolioV2>(response);
  },

  /** Current holdings for a portfolio, resolved by business code. */
  async getHoldings(portfolioCode: string): Promise<ApiHoldingV2[]> {
    const client = useOpenApiClientV2();
    const response = await client.GET("/portfolios/{portfolioCode}/holdings", {
      params: { path: { portfolioCode } },
    });
    return unwrapOpenApiResponse<ApiHoldingV2[]>(response);
  },

  /** Cash balances by currency for a portfolio, resolved by business code. */
  async getCash(portfolioCode: string): Promise<ApiCashBalanceV2[]> {
    const client = useOpenApiClientV2();
    const response = await client.GET("/portfolios/{portfolioCode}/cash", {
      params: { path: { portfolioCode } },
    });
    return unwrapOpenApiResponse<ApiCashBalanceV2[]>(response);
  },

  /** Latest authoritative valuation snapshot, resolved by business code. */
  async getLatestValuation(portfolioCode: string): Promise<ApiValuationV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/valuations/latest",
      { params: { path: { portfolioCode } } },
    );
    return unwrapOpenApiResponse<ApiValuationV2>(response);
  },

  /** Official valuation history used by the portfolio overview sparkline. */
  async listValuations(
    portfolioCode: string,
    query: {
      from?: string;
      to?: string;
      page?: number;
      limit?: number;
    } = {},
  ): Promise<ApiValuationListV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/valuations",
      { params: { path: { portfolioCode }, query } },
    );
    return unwrapOpenApiResponse<ApiValuationListV2>(response);
  },

  /** Ledger transactions for a portfolio, resolved by business code. */
  async listTransactions(
    portfolioCode: string,
    query: { page?: number; limit?: number } = {},
  ): Promise<ApiTransactionListV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/transactions",
      { params: { path: { portfolioCode }, query } },
    );
    return unwrapOpenApiResponse<ApiTransactionListV2>(response);
  },
};
