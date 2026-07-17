/**
 * Portfolio V2 investment decisions — typed API client
 * (docs/api/portfolio-v2-api-ddd.md, docs/frontend/portfolio-v2-frontend-ddd.md).
 *
 * All requests go through the generated OpenAPI client; do NOT add manual
 * type definitions or hand-rolled fetch calls here.
 *
 * Endpoints used (mounted under /api/v2, resolved by portfolioCode — never
 * a raw portfolio UUID):
 *   GET  /portfolios/{portfolioCode}/decisions
 *   POST /portfolios/{portfolioCode}/decisions
 *   GET  /portfolios/{portfolioCode}/decisions/{decisionId}
 *   POST /portfolios/{portfolioCode}/decisions/{decisionId}/submit
 *   POST /portfolios/{portfolioCode}/decisions/{decisionId}/cancel
 *
 * The create body intentionally carries no fund_id, portfolio_id, or
 * contract_id — portfolio identity comes exclusively from the portfolioCode
 * path segment; the backend derives the rest from the resolved portfolio.
 */
import { unwrapOpenApiResponse, useOpenApiClientV2 } from "~/api/openapi";
import type { components } from "~/api/ims-api";

export type ApiDecisionV2 = components["schemas"]["DecisionResponse"];
export type ApiDecisionListV2 = components["schemas"]["DecisionListResponse"];
export type ApiCreateDecisionV2Request =
  components["schemas"]["CreateDecisionV2Request"];
export type ApiCancelDecisionV2Request =
  components["schemas"]["CancelDecisionRequest"];

export interface DecisionListV2Filters {
  page?: number;
  limit?: number;
}

export const portfolioDecisionApi = {
  async list(
    portfolioCode: string,
    filters: DecisionListV2Filters = {},
  ): Promise<ApiDecisionListV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET("/portfolios/{portfolioCode}/decisions", {
      params: { path: { portfolioCode }, query: filters },
    });
    return unwrapOpenApiResponse<ApiDecisionListV2>(response);
  },

  async create(
    portfolioCode: string,
    body: ApiCreateDecisionV2Request,
  ): Promise<ApiDecisionV2> {
    const client = useOpenApiClientV2();
    const response = await client.POST("/portfolios/{portfolioCode}/decisions", {
      params: { path: { portfolioCode } },
      body,
    });
    return unwrapOpenApiResponse<ApiDecisionV2>(response);
  },

  async get(portfolioCode: string, decisionId: string): Promise<ApiDecisionV2> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/decisions/{decisionId}",
      { params: { path: { portfolioCode, decisionId } } },
    );
    return unwrapOpenApiResponse<ApiDecisionV2>(response);
  },

  async submit(portfolioCode: string, decisionId: string): Promise<ApiDecisionV2> {
    const client = useOpenApiClientV2();
    const response = await client.POST(
      "/portfolios/{portfolioCode}/decisions/{decisionId}/submit",
      { params: { path: { portfolioCode, decisionId } } },
    );
    return unwrapOpenApiResponse<ApiDecisionV2>(response);
  },

  async cancel(
    portfolioCode: string,
    decisionId: string,
    reason: string,
  ): Promise<ApiDecisionV2> {
    const client = useOpenApiClientV2();
    const response = await client.POST(
      "/portfolios/{portfolioCode}/decisions/{decisionId}/cancel",
      {
        params: { path: { portfolioCode, decisionId } },
        body: { reason },
      },
    );
    return unwrapOpenApiResponse<ApiDecisionV2>(response);
  },
};
