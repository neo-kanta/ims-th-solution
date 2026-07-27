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
 *   GET  /api/v2/portfolios                                — list portfolios (accessible funds)
 *   POST /api/v2/portfolios                                — create a portfolio under a fund_code
 *   GET  /api/v2/portfolios/{portfolioCode}                 — portfolio detail by code
 *   PATCH /api/v2/portfolios/{portfolioCode}                — update mutable descriptive metadata
 *   GET  /api/v2/portfolios/{portfolioCode}/holdings        — current holdings
 *   GET  /api/v2/portfolios/{portfolioCode}/cash            — cash balances
 *   GET  /api/v2/portfolios/{portfolioCode}/transactions    — ledger history
 *   POST /api/v2/portfolios/{portfolioCode}/transactions/simulate — pre-trade preview
 *   POST /api/v2/portfolios/{portfolioCode}/transactions    — post to ledger (201) or, for a
 *       LIVE cash movement, stage it for approval (202, CashRequestResponse)
 *   GET  /api/v2/portfolios/{portfolioCode}/cash-requests    — submitter's LIVE cash-approval requests
 *   POST /api/v2/portfolios/{portfolioCode}/cash-requests/{id}/cancel — cancel a PENDING request
 *   GET  /api/v2/portfolios/{portfolioCode}/executions      — trade executions (list)
 *   GET  /api/v2/portfolios/{portfolioCode}/executions/{id} — trade execution (detail)
 *   GET  /api/v2/portfolios/{portfolioCode}/confirmations       — trade confirmations (list)
 *   GET  /api/v2/portfolios/{portfolioCode}/confirmations/{id}  — trade confirmation (detail)
 *   POST /api/v2/portfolios/{portfolioCode}/valuations/run  — trigger a valuation run
 */
import { unwrapOpenApiResponse, useOpenApiClientV2 } from "~/api/openapi";
import type { components } from "~/api/ims-api";

export type ApiPortfolioV2 = components["schemas"]["PortfolioResponse"];
export type ApiPortfolioListV2 = components["schemas"]["PortfolioListResponse"];
export type ApiCreatePortfolioV2Request =
  components["schemas"]["CreatePortfolioV2Request"];
export type ApiPatchPortfolioV2Request =
  components["schemas"]["PatchPortfolioV2Request"];
export type ApiHoldingV2 = components["schemas"]["HoldingResponse"];
export type ApiCashBalanceV2 = components["schemas"]["CashBalanceResponse"];
export type ApiTransactionV2 = components["schemas"]["TransactionResponse"];
export type ApiTransactionListV2 =
  components["schemas"]["TransactionListResponse"];
export type ApiValuationV2 = components["schemas"]["ValuationResponse"];
export type ApiValuationListV2 = components["schemas"]["ValuationListResponse"];
export type ApiPostTransactionRequestV2 =
  components["schemas"]["PostTransactionRequest"];
export type ApiTransactionSimulationV2 =
  components["schemas"]["TransactionSimulationResponse"];
export type ApiExecutionV2 = components["schemas"]["ExecutionResponse"];
export type ApiExecutionListV2 = components["schemas"]["ExecutionListResponse"];
export type ApiTradeConfirmationV2 =
  components["schemas"]["TradeConfirmationResponse"];
export type ApiTradeConfirmationListV2 =
  components["schemas"]["TradeConfirmationListResponse"];
export type ApiRunValuationRequest =
  components["schemas"]["RunValuationRequest"];
export type ApiResolveConfirmationRequest =
  components["schemas"]["ResolveConfirmationRequest"];
export type ApiCashRequestV2 = components["schemas"]["CashRequestResponse"];
export type ApiCashRequestListV2 =
  components["schemas"]["CashRequestListResponse"];
export type ApiCancelCashRequestV2Request =
  components["schemas"]["CancelCashRequestV2Request"];

// Re-exported for convenience so most call sites only need one import; the
// implementation lives in lib/cashRequestGuard.ts (no Nuxt-aliased imports)
// so it can also be imported directly in plain Vitest without mocking this
// module's `~/api/openapi` dependency.
export { isCashRequestResponse } from "../lib/cashRequestGuard";

export const portfolioApi = {
  /** List portfolios visible to the caller, optionally filtered by fund/status. */
  async list(
    query: {
      fund_code?: string;
      status?: string;
      page?: number;
      limit?: number;
    } = {},
  ): Promise<ApiPortfolioListV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET("/portfolios", { params: { query } });
    return unwrapOpenApiResponse<ApiPortfolioListV2>(response);
  },

  /**
   * Create a portfolio under a fund resolved by business `fund_code`. The
   * request body must never include `fund_id` — the backend resolves fund
   * scope from `fund_code` alone (Portfolio V2 identity rule).
   */
  async create(body: ApiCreatePortfolioV2Request): Promise<ApiPortfolioV2> {
    const client = useOpenApiClientV2();
    const response = await client.POST("/portfolios", { body });
    return unwrapOpenApiResponse<ApiPortfolioV2>(response);
  },

  /**
   * Update mutable descriptive metadata only (name/description/benchmark/
   * risk_profile/strategy_code/manager_user_id/style_id) using optimistic
   * concurrency via `expected_version`. Never changes fund association,
   * code, portfolio_type, or lifecycle status — the backend does not expose
   * those as patchable fields.
   */
  async update(
    portfolioCode: string,
    body: ApiPatchPortfolioV2Request,
  ): Promise<ApiPortfolioV2> {
    const client = useOpenApiClientV2();
    const response = await client.PATCH("/portfolios/{portfolioCode}", {
      params: { path: { portfolioCode } },
      body,
    });
    return unwrapOpenApiResponse<ApiPortfolioV2>(response);
  },

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

  /**
   * Preview the ledger/cash/position/compliance impact of a transaction
   * without posting it, resolved by business code (Portfolio V2).
   */
  async simulateTransaction(
    portfolioCode: string,
    body: ApiPostTransactionRequestV2,
  ): Promise<ApiTransactionSimulationV2> {
    const client = useOpenApiClientV2();
    const response = await client.POST(
      "/portfolios/{portfolioCode}/transactions/simulate",
      { params: { path: { portfolioCode } }, body },
    );
    return unwrapOpenApiResponse<ApiTransactionSimulationV2>(response);
  },

  /**
   * Post a transaction to the ledger, resolved by business code (Portfolio
   * V2). The backend re-runs pre-trade compliance using the actual posted
   * values before committing. For a LIVE cash movement (CASH_IN/CASH_OUT/
   * FEE/DIVIDEND) the backend does not post immediately — it stages the
   * movement for approval and returns a pending CashRequestResponse
   * instead (HTTP 202). Use `isCashRequestResponse` to distinguish the two
   * shapes at the call site.
   *
   * `idempotencyKey`, when supplied, is sent as the `Idempotency-Key` header
   * so a retried submission (e.g. a network error after the request actually
   * reached the backend) resolves to the original LIVE cash request instead
   * of creating a duplicate. Ignored server-side for immediate (non-gated)
   * posts.
   */
  async postTransaction(
    portfolioCode: string,
    body: ApiPostTransactionRequestV2,
    idempotencyKey?: string,
  ): Promise<ApiTransactionV2 | ApiCashRequestV2> {
    const client = useOpenApiClientV2();
    const response = await client.POST(
      "/portfolios/{portfolioCode}/transactions",
      {
        params: { path: { portfolioCode } },
        body,
        ...(idempotencyKey
          ? { headers: { "Idempotency-Key": idempotencyKey } }
          : {}),
      },
    );
    return unwrapOpenApiResponse<ApiTransactionV2 | ApiCashRequestV2>(response);
  },

  /**
   * The submitter's LIVE cash-approval requests for a portfolio, resolved
   * by business code. Optional status filter (PENDING/APPROVED/REJECTED/
   * CANCELLED).
   */
  async listCashRequests(
    portfolioCode: string,
    query: { status?: string; page?: number; limit?: number } = {},
  ): Promise<ApiCashRequestListV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/cash-requests",
      { params: { path: { portfolioCode }, query } },
    );
    return unwrapOpenApiResponse<ApiCashRequestListV2>(response);
  },

  /**
   * Cancel a PENDING LIVE cash-approval request. Only the submitter may
   * cancel; after cancel, approvers can no longer act on it.
   */
  async cancelCashRequest(
    portfolioCode: string,
    cashRequestId: string,
    body: ApiCancelCashRequestV2Request = {},
  ): Promise<ApiCashRequestV2> {
    const client = useOpenApiClientV2();
    const response = await client.POST(
      "/portfolios/{portfolioCode}/cash-requests/{cashRequestId}/cancel",
      { params: { path: { portfolioCode, cashRequestId } }, body },
    );
    return unwrapOpenApiResponse<ApiCashRequestV2>(response);
  },

  /** Trade executions for a portfolio, resolved by business code. */
  async listExecutions(
    portfolioCode: string,
    query: { status?: string; page?: number; limit?: number } = {},
  ): Promise<ApiExecutionListV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/executions",
      { params: { path: { portfolioCode }, query } },
    );
    return unwrapOpenApiResponse<ApiExecutionListV2>(response);
  },

  /** One trade execution that belongs to the resolved portfolio. */
  async getExecution(
    portfolioCode: string,
    executionId: string,
  ): Promise<ApiExecutionV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/executions/{executionId}",
      { params: { path: { portfolioCode, executionId } } },
    );
    return unwrapOpenApiResponse<ApiExecutionV2>(response);
  },

  /** Trade confirmations for a portfolio, resolved by business code. */
  async listConfirmations(
    portfolioCode: string,
    query: { status?: string; page?: number; limit?: number } = {},
  ): Promise<ApiTradeConfirmationListV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/confirmations",
      { params: { path: { portfolioCode }, query } },
    );
    return unwrapOpenApiResponse<ApiTradeConfirmationListV2>(response);
  },

  /** One trade confirmation that belongs to the resolved portfolio. */
  async getConfirmation(
    portfolioCode: string,
    confirmationId: string,
  ): Promise<ApiTradeConfirmationV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/confirmations/{confirmationId}",
      { params: { path: { portfolioCode, confirmationId } } },
    );
    return unwrapOpenApiResponse<ApiTradeConfirmationV2>(response);
  },

  /**
   * Resolve a PENDING_REVIEW/MISMATCHED trade confirmation to a target
   * status, resolved by business code. The backend records who resolved it
   * and when — never sent from the client.
   */
  async resolveConfirmation(
    portfolioCode: string,
    confirmationId: string,
    body: ApiResolveConfirmationRequest,
  ): Promise<ApiTradeConfirmationV2> {
    const client = useOpenApiClientV2();
    const response = await client.POST(
      "/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve",
      { params: { path: { portfolioCode, confirmationId } }, body },
    );
    return unwrapOpenApiResponse<ApiTradeConfirmationV2>(response);
  },

  /**
   * Manually trigger the valuation runner for a portfolio, resolved by
   * business code. Rejected by the backend for MODEL portfolios, which have
   * no official valuation.
   */
  async runValuation(
    portfolioCode: string,
    body: ApiRunValuationRequest,
  ): Promise<ApiValuationV2> {
    const client = useOpenApiClientV2();
    const response = await client.POST(
      "/portfolios/{portfolioCode}/valuations/run",
      { params: { path: { portfolioCode } }, body },
    );
    return unwrapOpenApiResponse<ApiValuationV2>(response);
  },
};
