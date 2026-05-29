/**
 * Market Data — catalog API client.
 *
 * Backed by the canonical reference-data REST endpoints under /api/v1. From
 * the frontend's point of view these are part of the Market Data feature —
 * the "catalog" is everything you've imported from Yahoo / Alpha Vantage.
 *
 * Endpoints consumed:
 *   GET    /reference-data/securities/search
 *   POST   /reference-data/securities
 *   GET    /reference-data/securities/{security_id}
 *   PATCH  /reference-data/securities/{security_id}
 *   GET    /reference-data/securities/{security_id}/mappings
 *   POST   /reference-data/securities/{security_id}/mappings
 *   DELETE /reference-data/securities/{security_id}/mappings/{mapping_id}
 *   GET    /reference-data/unmapped-candidates
 *   POST   /reference-data/unmapped-candidates/{candidate_id}/map
 *   POST   /reference-data/unmapped-candidates/{candidate_id}/reject
 *
 * Requests use the generated OpenAPI client. Do not hand-roll types here.
 */
import { unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { paths, components } from "~/api/ims-api";

export type ApiSecurity = components["schemas"]["SecurityDTO"];
export type ApiProviderMapping = components["schemas"]["ProviderMappingDTO"];
export type ApiUnmappedCandidate = components["schemas"]["UnmappedCandidateDTO"];
export type ApiSecuritySearchResponse = components["schemas"]["SearchResponse"];
export type ApiMappingsResponse = components["schemas"]["MappingsResponse"];
export type ApiCandidatesResponse = components["schemas"]["CandidatesResponse"];
export type ApiCreateSecurityRequest = components["schemas"]["CreateSecurityRequest"];
export type ApiUpdateSecurityRequest = components["schemas"]["UpdateSecurityRequest"];
export type ApiAddMappingRequest = components["schemas"]["AddProviderMappingRequest"];
export type ApiMapCandidateRequest = components["schemas"]["MapCandidateRequest"];
export type ApiRejectCandidateRequest = components["schemas"]["RejectCandidateRequest"];

type SecuritySearchQuery = NonNullable<
  paths["/reference-data/securities/search"]["get"]["parameters"]["query"]
>;
type UnmappedQuery = NonNullable<
  paths["/reference-data/unmapped-candidates"]["get"]["parameters"]["query"]
>;

export interface SecuritySearchParams {
  query?: string;
  assetType?: string;
  provider?: string;
  status?: string;
  limit?: number;
}

export interface UnmappedCandidateParams {
  status?: string;
  provider?: string;
  batchId?: string;
  limit?: number;
}

export const marketDataCatalogApi = {
  async searchSecurities(params: SecuritySearchParams = {}): Promise<ApiSecuritySearchResponse> {
    const client = useOpenApiClient();
    const query: SecuritySearchQuery = {};
    if (params.query) query.query = params.query;
    if (params.assetType) query.asset_type = params.assetType;
    if (params.provider) query.provider = params.provider;
    if (params.status) query.status = params.status;
    if (params.limit) query.limit = params.limit;
    const response = await client.GET("/reference-data/securities/search", {
      params: { query },
    });
    return unwrapOpenApiResponse<ApiSecuritySearchResponse>(response);
  },

  async createSecurity(req: ApiCreateSecurityRequest): Promise<ApiSecurity> {
    const client = useOpenApiClient();
    const response = await client.POST("/reference-data/securities", { body: req });
    return unwrapOpenApiResponse<ApiSecurity>(response);
  },

  async getSecurity(securityId: string): Promise<ApiSecurity> {
    const client = useOpenApiClient();
    const response = await client.GET("/reference-data/securities/{security_id}", {
      params: { path: { security_id: securityId } },
    });
    return unwrapOpenApiResponse<ApiSecurity>(response);
  },

  async updateSecurity(securityId: string, req: ApiUpdateSecurityRequest): Promise<ApiSecurity> {
    const client = useOpenApiClient();
    const response = await client.PATCH("/reference-data/securities/{security_id}", {
      params: { path: { security_id: securityId } },
      body: req,
    });
    return unwrapOpenApiResponse<ApiSecurity>(response);
  },

  async listMappings(securityId: string): Promise<ApiMappingsResponse> {
    const client = useOpenApiClient();
    const response = await client.GET("/reference-data/securities/{security_id}/mappings", {
      params: { path: { security_id: securityId } },
    });
    return unwrapOpenApiResponse<ApiMappingsResponse>(response);
  },

  async addMapping(securityId: string, req: ApiAddMappingRequest): Promise<ApiProviderMapping> {
    const client = useOpenApiClient();
    const response = await client.POST("/reference-data/securities/{security_id}/mappings", {
      params: { path: { security_id: securityId } },
      body: req,
    });
    return unwrapOpenApiResponse<ApiProviderMapping>(response);
  },

  async deleteMapping(securityId: string, mappingId: string): Promise<void> {
    const client = useOpenApiClient();
    await client.DELETE("/reference-data/securities/{security_id}/mappings/{mapping_id}", {
      params: { path: { security_id: securityId, mapping_id: mappingId } },
    });
  },

  async listUnmappedCandidates(params: UnmappedCandidateParams = {}): Promise<ApiCandidatesResponse> {
    const client = useOpenApiClient();
    const query: UnmappedQuery = {};
    if (params.status) query.status = params.status;
    if (params.provider) query.provider = params.provider;
    if (params.batchId) query.batch_id = params.batchId;
    if (params.limit) query.limit = params.limit;
    const response = await client.GET("/reference-data/unmapped-candidates", {
      params: { query },
    });
    return unwrapOpenApiResponse<ApiCandidatesResponse>(response);
  },

  async mapCandidate(candidateId: string, securityId: string): Promise<void> {
    const client = useOpenApiClient();
    await client.POST("/reference-data/unmapped-candidates/{candidate_id}/map", {
      params: { path: { candidate_id: candidateId } },
      body: { security_id: securityId } satisfies ApiMapCandidateRequest,
    });
  },

  async rejectCandidate(candidateId: string, reason: string): Promise<void> {
    const client = useOpenApiClient();
    await client.POST("/reference-data/unmapped-candidates/{candidate_id}/reject", {
      params: { path: { candidate_id: candidateId } },
      body: { reason } satisfies ApiRejectCandidateRequest,
    });
  },
};
