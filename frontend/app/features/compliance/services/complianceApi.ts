/**
 * Compliance HTTP service — Phase 1.
 *
 * Dashboard reads use the generated OpenAPI client. A small set of older
 * compliance mutations still uses `useApi()` because their generated schemas
 * currently erase JSON object fields as `Record<string, never>`; those are
 * intentionally kept separate from the dashboard's accepted read contract.
 *
 * The backend returns successful responses as `{ data, message }`; we unwrap
 * `.data` once here so callers always work in domain types.
 */
import { unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { components, paths } from "~/api/ims-api";
import { useApi } from "~/composables/useApi";

import {
  normalizeComplianceBreach,
  normalizeCompliancePortfolio,
  normalizeComplianceRule,
} from "../lib/formatters";
import {
  collectNumberedPages,
  collectOffsetPages,
} from "../lib/pagination";

import type {
  ComplianceBreachListFilters,
  ComplianceBreachListResponse,
  ComplianceCheckGroupResult,
  ComplianceOverride,
  ComplianceOverrideRequest,
  CompliancePortfolioListResponse,
  ComplianceRuleListFilters,
  ComplianceRuleListResponse,
  CompliancePreTradeRequest,
  CompliancePreTradeResponse,
} from "../types";

export interface ComplianceCreateRuleInstanceRequest {
  rule_type_id: string;
  name: string;
  description?: string;
  parameters: Record<string, unknown>;
  effective_from: string;
  effective_to?: string;
  is_active?: boolean;
  change_reason?: string;
}

export interface ComplianceCreateRuleInstanceResponse {
  instance: import("../types").ComplianceRule;
  version: {
    id: string;
    ruleInstanceID: string;
    versionNumber: number;
    parameters: Record<string, unknown>;
    changeReason?: string;
    createdAt: string;
    createdBy: string;
    approvedBy?: string;
  };
}

interface SuccessEnvelope<T> {
  data: T;
  message?: string;
}

type ApiBreachList = components["schemas"]["ListBreachesResult"];
type ApiPortfolioList = components["schemas"]["PortfolioListResponse"];
type ApiRuleList = components["schemas"]["ListRuleInstancesResult"];
type BreachListQuery = NonNullable<
  paths["/compliance/breaches"]["get"]["parameters"]["query"]
>;
type PortfolioListQuery = NonNullable<
  paths["/investment/portfolios"]["get"]["parameters"]["query"]
>;
type RuleListQuery = NonNullable<
  paths["/compliance/rules"]["get"]["parameters"]["query"]
>;

function paginationNumber(
  value: number | undefined,
  field: string,
): number {
  if (!Number.isSafeInteger(value) || (value ?? -1) < 0) {
    throw new Error(`The API response is missing valid ${field} pagination metadata.`);
  }
  return value as number;
}

function definedQuery<T extends Record<string, unknown>>(
  source: T,
): T {
  return Object.fromEntries(
    Object.entries(source).filter(([, value]) => {
      if (value === undefined || value === null) return false;
      return typeof value !== "string" || value.trim() !== "";
    }),
  ) as T;
}

function unwrap<T>(envelope: SuccessEnvelope<T>): T {
  return envelope.data;
}

function buildQuery(filters: Record<string, unknown>): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value === undefined || value === null) continue;
    const str =
      typeof value === "string"
        ? value.trim()
        : typeof value === "boolean"
          ? String(value)
          : String(value);
    if (str === "") continue;
    params.set(key, str);
  }
  const query = params.toString();
  return query ? `?${query}` : "";
}

export const complianceApi = {
  async listRules(
    filters: ComplianceRuleListFilters = {},
  ): Promise<ComplianceRuleListResponse> {
    const client = useOpenApiClient();
    const query = definedQuery<RuleListQuery>({ ...filters });
    const response = await client.GET("/compliance/rules", {
      params: { query },
    });
    const payload = unwrapOpenApiResponse<ApiRuleList>(response);
    return {
      instances: (payload.instances ?? []).map(normalizeComplianceRule),
      total: paginationNumber(payload.total, "total"),
      offset: paginationNumber(payload.offset, "offset"),
      limit: paginationNumber(payload.limit, "limit"),
    };
  },

  async listAllRules(): Promise<ComplianceRuleListResponse> {
    const result = await collectOffsetPages((offset, limit) =>
      complianceApi.listRules({ offset, limit }).then((page) => ({
        items: page.instances,
        total: page.total,
        offset: page.offset,
      })),
    );
    return {
      instances: result.items,
      total: result.total,
      offset: 0,
      limit: result.items.length,
    };
  },

  async listBreaches(
    filters: ComplianceBreachListFilters = {},
  ): Promise<ComplianceBreachListResponse> {
    const client = useOpenApiClient();
    const query = definedQuery<BreachListQuery>({ ...filters });
    const response = await client.GET("/compliance/breaches", {
      params: { query },
    });
    const payload = unwrapOpenApiResponse<ApiBreachList>(response);
    return {
      breaches: (payload.breaches ?? []).map(normalizeComplianceBreach),
      total: paginationNumber(payload.total, "total"),
      offset: paginationNumber(payload.offset, "offset"),
      limit: paginationNumber(payload.limit, "limit"),
    };
  },

  async runPreTradeCheck(
    payload: CompliancePreTradeRequest,
  ): Promise<CompliancePreTradeResponse> {
    const { apiFetch } = useApi();
    const response = await apiFetch<
      SuccessEnvelope<CompliancePreTradeResponse>
    >("/compliance/checks/pre-trade", {
      method: "POST",
      body: payload,
    });
    return unwrap(response);
  },

  /**
   * Lists portfolios visible to the authenticated user. Used by the simulator
   * to render a human-readable picker instead of asking for a raw UUID.
   * Endpoint owned by the investment module; we call it directly because no
   * `useApi` segregation exists in the project.
   */
  async listPortfolios(
    filters: { fund_id?: string; status?: string; page?: number; limit?: number } = {},
  ): Promise<CompliancePortfolioListResponse> {
    const client = useOpenApiClient();
    const query = definedQuery<PortfolioListQuery>({ ...filters });
    const response = await client.GET("/investment/portfolios", {
      params: { query },
    });
    const payload = unwrapOpenApiResponse<ApiPortfolioList>(response);
    return {
      items: (payload.items ?? []).map(normalizeCompliancePortfolio),
      total: paginationNumber(payload.total, "total"),
      page: paginationNumber(payload.page, "page"),
      limit: paginationNumber(payload.limit, "limit"),
    };
  },

  async listAllPortfolios(): Promise<CompliancePortfolioListResponse> {
    const result = await collectNumberedPages((page, limit) =>
      complianceApi.listPortfolios({ page, limit }).then((payload) => ({
        items: payload.items,
        total: payload.total,
        page: payload.page,
      })),
    );
    return {
      items: result.items,
      total: result.total,
      page: 1,
      limit: result.items.length,
    };
  },

  /**
   * Override an OPEN compliance breach. The endpoint demands a reason and
   * stamps the actor from the auth context — we never forward an actor id.
   */
  async overrideBreach(
    breachId: string,
    payload: ComplianceOverrideRequest,
  ): Promise<ComplianceOverride> {
    const { apiFetch } = useApi();
    const response = await apiFetch<SuccessEnvelope<ComplianceOverride>>(
      `/compliance/breaches/${encodeURIComponent(breachId)}/override`,
      { method: "POST", body: payload },
    );
    return unwrap(response);
  },

  /**
   * Fetch all records + breaches for one check group id. Used by the audit
   * trail page and the post-trade inbox detail drawer.
   */
  async getCheckGroup(groupId: string): Promise<ComplianceCheckGroupResult> {
    const { apiFetch } = useApi();
    const response = await apiFetch<SuccessEnvelope<ComplianceCheckGroupResult>>(
      `/compliance/checks/${encodeURIComponent(groupId)}`,
    );
    return unwrap(response);
  },

  /**
   * Create a compliance rule instance. The lifecycle (submit/approve/disable
   * /archive) endpoints do NOT exist yet — see Missing API Checklist #3-#6.
   * This endpoint creates and activates in one shot (or stages inactive when
   * `is_active=false`).
   */
  async createRule(
    payload: ComplianceCreateRuleInstanceRequest,
  ): Promise<ComplianceCreateRuleInstanceResponse> {
    const { apiFetch } = useApi();
    const response = await apiFetch<
      SuccessEnvelope<ComplianceCreateRuleInstanceResponse>
    >("/compliance/rules", { method: "POST", body: payload });
    return unwrap(response);
  },

  /**
   * Page through `/admin/users` to resolve UUIDs to display names. The
   * endpoint is gated by IAM_USER_VIEW — most compliance-only users will
   * get a 403 here, in which case callers should fall back to UUIDs.
   */
  async listUsers(
    filters: { offset?: number; limit?: number; search?: string } = {},
  ): Promise<{
    users: Array<{ id: string; display_name?: string; username?: string }>;
    total: number;
    offset: number;
    limit: number;
  }> {
    const { apiFetch } = useApi();
    const response = await apiFetch<
      SuccessEnvelope<{
        users: Array<{ id: string; display_name?: string; username?: string }>;
        total: number;
        offset: number;
        limit: number;
      }>
    >(`/admin/users${buildQuery(filters)}`);
    return unwrap(response);
  },
};
