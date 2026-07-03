/**
 * Typed wrappers for the My Funds workspace. Every call goes through the
 * generated OpenAPI client — see CLAUDE.md for the no-manual-API-clients
 * rule.
 *
 * Endpoints used (live in backend today):
 *   GET  /investment/funds                           — accessible funds, server-scoped
 *   GET  /investment/portfolios?fund_id=<uuid>       — portfolios under one fund
 *   GET  /investment/portfolios/{id}/valuations/latest — fund-level NAV/AUM/P&L aggregator
 *   GET  /investment/portfolios/{id}/cash            — live cash buffer
 *   GET  /workflow/day-states/{contractId}?businessDate=YYYY-MM-DD
 *   GET  /compliance/breaches?contract_id=<uuid>&status=OPEN
 */
import { OpenApiRequestError, unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { components } from "~/api/ims-api";

import type {
  ApiBreach,
  ApiCashBalance,
  ApiFund,
  ApiPortfolio,
  ApiValuation,
  ApiWorkflowState,
} from "../types";

/** Generated fund-level NAV snapshot — same shape as backend's response.FundNAVResponse. */
export type FundNavSnapshot = components["schemas"]["FundNAVResponse"];
export type ApiHolding = components["schemas"]["HoldingResponse"];
export type ApiInstrument = components["schemas"]["InstrumentResponse"];
export type ApiAssetClass = components["schemas"]["AssetClass"];
export type FundAllocation = components["schemas"]["FundAllocationResponse"];
export type AllocationBucket = components["schemas"]["AllocationBucketResponse"];
export type FundNavHistory = components["schemas"]["FundNAVHistoryResponse"];
export type NavHistoryPoint = components["schemas"]["NAVHistoryPointResponse"];

export type FundNavHistoryRange = "1M" | "3M" | "6M" | "1Y" | "5Y" | "YTD";

export type CreateFundPayload = components["schemas"]["CreateFundRequest"];
export type ApiFundCategory = components["schemas"]["FundCategory"];
export type PostTransactionPayload = components["schemas"]["PostTransactionRequest"];
export type ApiTransaction = components["schemas"]["TransactionResponse"];
export type ApiPreTradeResponse = components["schemas"]["PreTradeCheckResponse"];
export type ApiTransactionSimulation = components["schemas"]["TransactionSimulationResponse"];
export type ApiCompliancePreview = components["schemas"]["CompliancePreviewResponse"];
export type ApiBreachSummary = components["schemas"]["BreachSummary"];

type FundsListResponse = {
  items?: ApiFund[];
  total?: number;
  page?: number;
  limit?: number;
};

type PortfolioListResponse = {
  items?: ApiPortfolio[];
  total?: number;
  page?: number;
  limit?: number;
};

type BreachListResponse = {
  breaches?: ApiBreach[];
  total?: number;
  limit?: number;
  offset?: number;
};

export const myFundsApi = {
  /**
   * Create a new fund. Requires INVESTMENT_FUND_MANAGE.
   * The caller is responsible for collecting all required fields including
   * `fund_category_id` (resolved via {@link listFundCategories}).
   */
  async createFund(payload: CreateFundPayload): Promise<ApiFund> {
    const client = useOpenApiClient();
    const response = await client.POST("/investment/funds", { body: payload });
    return unwrapOpenApiResponse<ApiFund>(response);
  },

  /** Active fund categories — used to populate the Create Fund category dropdown. */
  async listFundCategories(): Promise<ApiFundCategory[]> {
    const client = useOpenApiClient();
    const response = await client.GET("/investment/reference/fund-categories");
    const raw = unwrapOpenApiResponse<unknown[]>(response);
    return (raw ?? []).map((entry) => {
      const e = entry as Record<string, unknown>;
      return {
        id: (e.id ?? e.ID) as string | undefined,
        code: (e.code ?? e.Code) as string | undefined,
        name: (e.name ?? e.Name) as string | undefined,
        assetClassID: (e.assetClassID ?? e.AssetClassID ?? e.asset_class_id) as
          | string
          | undefined,
        isActive: (e.isActive ?? e.IsActive) as boolean | undefined,
        createdAt: (e.createdAt ?? e.CreatedAt) as string | undefined,
        updatedAt: (e.updatedAt ?? e.UpdatedAt) as string | undefined,
      } satisfies ApiFundCategory;
    });
  },

  /**
   * Run a pre-trade simulation against a portfolio. Hits
   * POST /investment/portfolios/{id}/transactions/simulate. Returns per-rule
   * verdicts without writing the ledger.
   */
  async simulateTransaction(
    portfolioId: string,
    payload: PostTransactionPayload,
  ): Promise<ApiTransactionSimulation> {
    const client = useOpenApiClient();
    const response = await client.POST(
      "/investment/portfolios/{id}/transactions/simulate",
      { params: { path: { id: portfolioId } }, body: payload },
    );
    return unwrapOpenApiResponse<ApiTransactionSimulation>(response);
  },

  /**
   * Post a transaction. Hits POST /investment/portfolios/{id}/transactions.
   * The backend re-runs the pre-trade gates inside the post — independent
   * of any UI-side simulation — so a successful response means both the
   * pre-trade gate and the ledger write succeeded.
   */
  async postTransaction(
    portfolioId: string,
    payload: PostTransactionPayload,
  ): Promise<ApiTransaction> {
    const client = useOpenApiClient();
    const response = await client.POST(
      "/investment/portfolios/{id}/transactions",
      { params: { path: { id: portfolioId } }, body: payload },
    );
    return unwrapOpenApiResponse<ApiTransaction>(response);
  },

  async listMyFunds(limit = 50): Promise<ApiFund[]> {
    const client = useOpenApiClient();
    const response = await client.GET("/investment/funds", {
      params: { query: { limit, page: 1 } },
    });
    const body = unwrapOpenApiResponse<FundsListResponse>(response);
    return body.items ?? [];
  },

  async listPortfoliosForFund(fundId: string): Promise<ApiPortfolio[]> {
    const client = useOpenApiClient();
    const response = await client.GET("/investment/portfolios", {
      params: { query: { fund_id: fundId, limit: 50, page: 1 } },
    });
    const body = unwrapOpenApiResponse<PortfolioListResponse>(response);
    return body.items ?? [];
  },

  /**
   * Returns the latest valuation snapshot for one portfolio, or null when
   * no valuation has been computed yet. 404 is treated as null instead of
   * an error since brand-new portfolios legitimately lack history.
   */
  async getLatestValuation(portfolioId: string): Promise<ApiValuation | null> {
    const client = useOpenApiClient();
    try {
      const response = await client.GET(
        "/investment/portfolios/{id}/valuations/latest",
        { params: { path: { id: portfolioId } } },
      );
      return unwrapOpenApiResponse<ApiValuation>(response);
    } catch (err) {
      if (err instanceof OpenApiRequestError && err.status === 404) return null;
      throw err;
    }
  },

  /**
   * Fetches the fund-level latest NAV snapshot served by Slice 3's
   * read-only aggregator. Returns null when no portfolio under the fund has
   * a valuation yet (the backend maps that to 404).
   */
  async getLatestFundNav(fundId: string): Promise<FundNavSnapshot | null> {
    const client = useOpenApiClient();
    try {
      const response = await client.GET("/investment/funds/{id}/nav/latest", {
        params: { path: { id: fundId } },
      });
      return unwrapOpenApiResponse<FundNavSnapshot>(response);
    } catch (err) {
      if (err instanceof OpenApiRequestError && err.status === 404) return null;
      throw err;
    }
  },

  /**
   * Latest aggregated position list (qty + avg cost + cost basis) for one
   * portfolio. The backend returns rows projected from the immutable ledger.
   * 404 is mapped to an empty array since a brand-new portfolio has none.
   */
  async listPortfolioHoldings(portfolioId: string): Promise<ApiHolding[]> {
    const client = useOpenApiClient();
    try {
      const response = await client.GET(
        "/investment/portfolios/{id}/holdings",
        { params: { path: { id: portfolioId } } },
      );
      return unwrapOpenApiResponse<ApiHolding[]>(response);
    } catch (err) {
      if (err instanceof OpenApiRequestError && err.status === 404) return [];
      throw err;
    }
  },

  /**
   * Instrument master rows — used to resolve tickers, names, asset class and
   * currency for holding entries. Paged; we typically fetch the whole list
   * once per workspace view because the master is small (≤ hundreds).
   */
  async listInstruments(limit = 200): Promise<ApiInstrument[]> {
    const client = useOpenApiClient();
    const response = await client.GET("/investment/instruments", {
      params: { query: { limit, page: 1 } },
    });
    const body = unwrapOpenApiResponse<{ items?: ApiInstrument[] }>(response);
    return body.items ?? [];
  },

  /**
   * Fund-level allocation breakdowns — by asset class, sector, country and
   * currency — computed from the same per-instrument mark-to-market values as
   * the holdings valuation (same business_date, same price-selection tier),
   * so the two pages never disagree. 404 (returned when the fund has no
   * positions or valuations) is mapped to null so the UI can render an
   * empty-state card.
   */
  async getFundAllocation(fundId: string, businessDate?: string): Promise<FundAllocation | null> {
    const client = useOpenApiClient();
    try {
      const response = await client.GET("/investment/funds/{id}/allocation", {
        params: {
          path: { id: fundId },
          query: businessDate ? { business_date: businessDate } : undefined,
        },
      });
      return unwrapOpenApiResponse<FundAllocation>(response);
    } catch (err) {
      if (err instanceof OpenApiRequestError && err.status === 404) return null;
      throw err;
    }
  },

  /**
   * Fund NAV / AUM time series across a logical range. For unitised funds
   * the series carries NAV-per-unit; otherwise it carries AUM.
   */
  async getFundNavHistory(
    fundId: string,
    range: FundNavHistoryRange = "3M",
  ): Promise<FundNavHistory | null> {
    const client = useOpenApiClient();
    try {
      const response = await client.GET("/investment/funds/{id}/nav-history", {
        params: { path: { id: fundId }, query: { range } },
      });
      return unwrapOpenApiResponse<FundNavHistory>(response);
    } catch (err) {
      if (err instanceof OpenApiRequestError && err.status === 404) return null;
      throw err;
    }
  },

  /**
   * Active asset classes — used to label holdings by category (Equity,
   * Fixed Income, Fund, …) without leaking UUIDs into the UI.
   *
   * The backend currently emits PascalCase keys (`ID`, `Code`, `Name`) for
   * this endpoint while the OpenAPI spec declares camelCase. We normalise
   * to the camelCase shape the typed client + UI consume so renaming the
   * server-side struct later won't break callers.
   */
  async listAssetClasses(): Promise<ApiAssetClass[]> {
    const client = useOpenApiClient();
    const response = await client.GET("/investment/reference/asset-classes");
    const raw = unwrapOpenApiResponse<unknown[]>(response);
    return (raw ?? []).map((entry) => {
      const e = entry as Record<string, unknown>;
      return {
        id: (e.id ?? e.ID) as string | undefined,
        code: (e.code ?? e.Code) as string | undefined,
        name: (e.name ?? e.Name) as string | undefined,
        description: (e.description ?? e.Description) as string | undefined,
        displayOrder: (e.displayOrder ?? e.DisplayOrder) as number | undefined,
        isActive: (e.isActive ?? e.IsActive) as boolean | undefined,
        createdAt: (e.createdAt ?? e.CreatedAt) as string | undefined,
        updatedAt: (e.updatedAt ?? e.UpdatedAt) as string | undefined,
      } satisfies ApiAssetClass;
    });
  },

  async listCash(portfolioId: string): Promise<ApiCashBalance[]> {
    const client = useOpenApiClient();
    const response = await client.GET(
      "/investment/portfolios/{id}/cash",
      { params: { path: { id: portfolioId } } },
    );
    return unwrapOpenApiResponse<ApiCashBalance[]>(response);
  },

  async getWorkflowState(
    contractId: string,
    businessDate: string,
  ): Promise<ApiWorkflowState | null> {
    const client = useOpenApiClient();
    try {
      const response = await client.GET("/workflow/day-states/{contractId}", {
        params: { path: { contractId }, query: { businessDate } },
      });
      return unwrapOpenApiResponse<ApiWorkflowState>(response);
    } catch (err) {
      if (err instanceof OpenApiRequestError && err.status === 404) return null;
      throw err;
    }
  },

  async listOpenBreaches(contractId: string, limit = 25): Promise<ApiBreach[]> {
    const client = useOpenApiClient();
    const response = await client.GET("/compliance/breaches", {
      params: { query: { contract_id: contractId, status: "OPEN", limit } },
    });
    const body = unwrapOpenApiResponse<BreachListResponse>(response);
    return body.breaches ?? [];
  },
};
