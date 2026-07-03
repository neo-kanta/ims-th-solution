/**
 * Investment Decision — typed API client.
 *
 * All requests go through the generated OpenAPI client; do NOT add manual type
 * definitions or hand-rolled fetch calls here.
 */
import { unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { components, paths } from "~/api/ims-api";

export type ApiDecision = components["schemas"]["DecisionResponse"];
export type ApiDecisionList = components["schemas"]["DecisionListResponse"];
export type ApiCreateDecisionRequest =
  components["schemas"]["CreateDecisionRequest"];
export type ApiCancelDecisionRequest =
  components["schemas"]["CancelDecisionRequest"];
export type ApiBatchApprovalRequest =
  components["schemas"]["BatchApprovalRequest"];
export type ApiBatchRejectionRequest =
  components["schemas"]["BatchRejectionRequest"];
export type ApiBatchApprovalResponse =
  components["schemas"]["BatchApprovalResponse"];

type DecisionListQuery = NonNullable<
  paths["/investment/decisions"]["get"]["parameters"]["query"]
>;

export interface DecisionListFilters {
  fund_id?: string;
  portfolio_id?: string;
  business_date?: string;
  status?: string;
  search?: string;
  page?: number;
  limit?: number;
}

export const DECISION_STATUSES = [
  "DRAFT",
  "SUBMITTED",
  "PENDING_APPROVAL",
  "APPROVED",
  "BLOCKED",
  "CANCELLED",
  "READY_FOR_EXECUTION",
] as const;

export type DecisionStatus = (typeof DECISION_STATUSES)[number];

export const decisionApi = {
  async listDecisions(
    filters: DecisionListFilters = {},
  ): Promise<ApiDecisionList> {
    const client = useOpenApiClient();
    const query: DecisionListQuery = {};
    if (filters.fund_id) query.fund_id = filters.fund_id;
    if (filters.portfolio_id) query.portfolio_id = filters.portfolio_id;
    if (filters.business_date) query.business_date = filters.business_date;
    if (filters.status) query.status = filters.status;
    if (filters.search) query.search = filters.search;
    if (filters.page) query.page = filters.page;
    if (filters.limit) query.limit = filters.limit;
    const response = await client.GET("/investment/decisions", {
      params: { query },
    });
    return unwrapOpenApiResponse<ApiDecisionList>(response);
  },

  async getDecisionDetail(id: string): Promise<ApiDecision> {
    const client = useOpenApiClient();
    const response = await client.GET("/investment/decisions/{id}/details", {
      params: { path: { id } },
    });
    return unwrapOpenApiResponse<ApiDecision>(response);
  },

  async createDecision(
    body: ApiCreateDecisionRequest,
  ): Promise<ApiDecision> {
    const client = useOpenApiClient();
    const response = await client.POST("/investment/decisions", { body });
    return unwrapOpenApiResponse<ApiDecision>(response);
  },

  async submitDecision(id: string): Promise<ApiDecision> {
    const client = useOpenApiClient();
    const response = await client.POST(
      "/investment/decisions/{id}/submit",
      { params: { path: { id } } },
    );
    return unwrapOpenApiResponse<ApiDecision>(response);
  },

  async cancelDecision(id: string, reason: string): Promise<ApiDecision> {
    const client = useOpenApiClient();
    const response = await client.POST(
      "/investment/decisions/{id}/cancel",
      { params: { path: { id } }, body: { reason } },
    );
    return unwrapOpenApiResponse<ApiDecision>(response);
  },

  async batchApprove(
    body: ApiBatchApprovalRequest,
  ): Promise<ApiBatchApprovalResponse> {
    const client = useOpenApiClient();
    const response = await client.POST(
      "/investment/decisions/batch-approve",
      { body },
    );
    return unwrapOpenApiResponse<ApiBatchApprovalResponse>(response);
  },

  async batchReject(
    body: ApiBatchRejectionRequest,
  ): Promise<ApiBatchApprovalResponse> {
    const client = useOpenApiClient();
    const response = await client.POST(
      "/investment/decisions/batch-reject",
      { body },
    );
    return unwrapOpenApiResponse<ApiBatchApprovalResponse>(response);
  },
};
