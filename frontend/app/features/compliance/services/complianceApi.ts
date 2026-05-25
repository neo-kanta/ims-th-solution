/**
 * Compliance HTTP service — Phase 1.
 *
 * Wraps the real backend `/compliance/*` endpoints through the project's
 * standard `useApi().apiFetch` composable so auth headers, base URL, and
 * 401 cleanup all behave identically to other features.
 *
 * The backend returns successful responses as `{ data, message }`; we unwrap
 * `.data` once here so callers always work in domain types.
 */
import { useApi } from "~/composables/useApi";

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
    const { apiFetch } = useApi();
    const response = await apiFetch<SuccessEnvelope<ComplianceRuleListResponse>>(
      `/compliance/rules${buildQuery(filters)}`,
    );
    return unwrap(response);
  },

  async listBreaches(
    filters: ComplianceBreachListFilters = {},
  ): Promise<ComplianceBreachListResponse> {
    const { apiFetch } = useApi();
    const response = await apiFetch<
      SuccessEnvelope<ComplianceBreachListResponse>
    >(`/compliance/breaches${buildQuery(filters)}`);
    return unwrap(response);
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
    const { apiFetch } = useApi();
    const response = await apiFetch<
      SuccessEnvelope<CompliancePortfolioListResponse>
    >(`/investment/portfolios${buildQuery(filters)}`);
    return unwrap(response);
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
