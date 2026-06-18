import { unwrapEnvelope, unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { components, paths } from "~/api/ims-api";

export type ApiDecision = components["schemas"]["DecisionResponse"];
export type ApiDecisionList = components["schemas"]["DecisionListResponse"];
export type ApiDecisionLine = components["schemas"]["DecisionLineResponse"];
export type ApiBatchApprovalRequest = components["schemas"]["BatchApprovalRequest"];
export type ApiBatchRejectionRequest = components["schemas"]["BatchRejectionRequest"];
export type ApiBatchApprovalResponse = components["schemas"]["BatchApprovalResponse"];
export type ApiBatchApprovalResult = components["schemas"]["BatchApprovalResultResponse"];

type ApprovalItemsQuery = NonNullable<
  paths["/investment/decisions/approval-items"]["get"]["parameters"]["query"]
>;

export interface DecisionApprovalFilters {
  page?: number;
  limit?: number;
  portfolio_id?: string;
  fund_id?: string;
  business_date_from?: string;
  business_date_to?: string;
  decision_no?: string;
  process_type?: string;
  product_type?: string;
  research_no?: string;
  status?: string;
  search?: string;
}

export const decisionApprovalApi = {
  async listApprovalItems(filters: DecisionApprovalFilters = {}): Promise<ApiDecisionList> {
    const client = useOpenApiClient();
    const query: ApprovalItemsQuery = {};
    if (filters.page !== undefined) query.page = filters.page;
    if (filters.limit !== undefined) query.limit = filters.limit;
    if (filters.portfolio_id) query.portfolio_id = filters.portfolio_id;
    if (filters.fund_id) query.fund_id = filters.fund_id;
    if (filters.business_date_from) query.business_date_from = filters.business_date_from;
    if (filters.business_date_to) query.business_date_to = filters.business_date_to;
    if (filters.decision_no) query.decision_no = filters.decision_no;
    if (filters.process_type) query.process_type = filters.process_type;
    if (filters.product_type) query.product_type = filters.product_type;
    if (filters.research_no) query.research_no = filters.research_no;
    if (filters.status) query.status = filters.status;
    if (filters.search) query.search = filters.search;

    const result = await client.GET("/investment/decisions/approval-items", { params: { query } });
    const raw = unwrapOpenApiResponse(result);
    return (unwrapEnvelope<ApiDecisionList>(raw) ?? raw) as ApiDecisionList;
  },

  async getDetails(id: string): Promise<ApiDecision> {
    const client = useOpenApiClient();
    const result = await client.GET("/investment/decisions/{id}/details", {
      params: { path: { id } },
    });
    const raw = unwrapOpenApiResponse(result);
    return (unwrapEnvelope<ApiDecision>(raw) ?? raw) as ApiDecision;
  },

  async batchApprove(payload: ApiBatchApprovalRequest): Promise<ApiBatchApprovalResponse> {
    const client = useOpenApiClient();
    const result = await client.POST("/investment/decisions/batch-approve", { body: payload });
    const raw = unwrapOpenApiResponse(result);
    return (unwrapEnvelope<ApiBatchApprovalResponse>(raw) ?? raw) as ApiBatchApprovalResponse;
  },

  async batchReject(payload: ApiBatchRejectionRequest): Promise<ApiBatchApprovalResponse> {
    const client = useOpenApiClient();
    const result = await client.POST("/investment/decisions/batch-reject", { body: payload });
    const raw = unwrapOpenApiResponse(result);
    return (unwrapEnvelope<ApiBatchApprovalResponse>(raw) ?? raw) as ApiBatchApprovalResponse;
  },
};
