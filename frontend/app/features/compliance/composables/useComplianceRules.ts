import { ref } from "vue";

import {
  complianceApi,
  type ComplianceCreateRuleInstanceRequest,
  type ComplianceCreateRuleInstanceResponse,
} from "../services/complianceApi";
import type {
  ComplianceRule,
  ComplianceRuleListFilters,
} from "../types";

function extractErrorMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
  if (data) {
    if (typeof data.error === "string" && data.error.trim()) return data.error;
    if (typeof data.message === "string" && data.message.trim()) return data.message;
  }
  const message = (err as { message?: unknown }).message;
  if (typeof message === "string" && message.trim()) return message;
  return fallback;
}

/**
 * Compliance rule listing composable — Phase 1.
 *
 * Owns its own loading/error/data state. Mirrors the pattern used by
 * `useResearchReportsList` so rule lists feel consistent across the app.
 */
export function useComplianceRulesList() {
  const items = ref<ComplianceRule[]>([]);
  const total = ref(0);
  const offset = ref(0);
  const limit = ref(50);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchList(filters: ComplianceRuleListFilters = {}) {
    loading.value = true;
    error.value = null;
    try {
      const payload = await complianceApi.listRules({
        offset: offset.value,
        limit: limit.value,
        ...filters,
      });
      items.value = payload.instances ?? [];
      total.value = payload.total ?? 0;
      offset.value = payload.offset ?? offset.value;
      limit.value = payload.limit ?? limit.value;
    } catch (err) {
      error.value = extractErrorMessage(err, "Failed to load compliance rules.");
      items.value = [];
      total.value = 0;
    } finally {
      loading.value = false;
    }
  }

  return {
    items,
    total,
    offset,
    limit,
    loading,
    error,
    fetchList,
  };
}

/**
 * Single-rule lookup composable — Phase 2.
 *
 * Until `GET /compliance/rules/{id}` ships (Missing API #1), we fetch the
 * paginated list and search client-side. The list query lets us filter by
 * rule_type_id but not by instance UUID, so we pull a wider page and find
 * the row in memory. This is acknowledged as slow and is the most-important
 * Phase 2 backend gap.
 */
export function useComplianceRuleDetail() {
  const rule = ref<ComplianceRule | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchById(id: string) {
    loading.value = true;
    error.value = null;
    try {
      // Walk pages until we find the rule or exhaust the listing.
      let offset = 0;
      const limit = 200;
      while (true) {
        const page = await complianceApi.listRules({ offset, limit });
        const hit = (page.instances ?? []).find((r) => r.id === id);
        if (hit) {
          rule.value = hit;
          return hit;
        }
        const fetched = (page.instances ?? []).length;
        offset += fetched;
        if (fetched < limit || offset >= (page.total ?? 0)) break;
      }
      rule.value = null;
      error.value =
        "Rule not found. The detail endpoint is missing — see Missing API #1.";
      return null;
    } catch (err) {
      rule.value = null;
      error.value = extractErrorMessage(err, "Failed to load compliance rule.");
      throw err;
    } finally {
      loading.value = false;
    }
  }

  return { rule, loading, error, fetchById };
}

/**
 * Create-only mutation composable — Phase 2.
 *
 * Maps to the real `POST /compliance/rules` endpoint. No update / submit /
 * approve / disable endpoints exist yet — the wizard UI keeps those buttons
 * disabled and references Missing API #3–#6.
 */
export function useComplianceRuleCreate() {
  const submitting = ref(false);
  const error = ref<string | null>(null);
  const lastResult = ref<ComplianceCreateRuleInstanceResponse | null>(null);

  async function create(payload: ComplianceCreateRuleInstanceRequest) {
    submitting.value = true;
    error.value = null;
    try {
      lastResult.value = await complianceApi.createRule(payload);
      return lastResult.value;
    } catch (err) {
      error.value = extractErrorMessage(err, "Failed to create compliance rule.");
      throw err;
    } finally {
      submitting.value = false;
    }
  }

  function reset() {
    error.value = null;
    lastResult.value = null;
  }

  return { submitting, error, lastResult, create, reset };
}
