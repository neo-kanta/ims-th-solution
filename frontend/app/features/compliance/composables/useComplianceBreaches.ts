import { ref } from "vue";

import { complianceApi } from "../services/complianceApi";
import type {
  ComplianceBreach,
  ComplianceBreachListFilters,
  ComplianceOverride,
  ComplianceOverrideRequest,
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
 * Compliance breach listing — Phase 1.
 *
 * Used by the Dashboard's recent-failures panel. Phase 2 will reuse this
 * composable for the Post-trade Breach Inbox screen.
 */
export function useComplianceBreachesList() {
  const items = ref<ComplianceBreach[]>([]);
  const total = ref(0);
  const offset = ref(0);
  const limit = ref(50);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchList(filters: ComplianceBreachListFilters = {}) {
    loading.value = true;
    error.value = null;
    try {
      const payload = await complianceApi.listBreaches({
        offset: offset.value,
        limit: limit.value,
        ...filters,
      });
      items.value = payload.breaches ?? [];
      total.value = payload.total ?? 0;
      offset.value = payload.offset ?? offset.value;
      limit.value = payload.limit ?? limit.value;
    } catch (err) {
      error.value = extractErrorMessage(err, "Failed to load compliance breaches.");
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
 * Breach override mutation — used by the Post-trade Breach Inbox.
 * Phase 2: the only real backend write for the inbox. Status transitions
 * (Missing API #8) are intentionally absent here.
 */
export function useComplianceBreachOverride() {
  const submitting = ref(false);
  const error = ref<string | null>(null);
  const lastResult = ref<ComplianceOverride | null>(null);

  function extractErrorMessage(err: unknown, fallback: string): string {
    if (!err || typeof err !== "object") return fallback;
    const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
    if (data) {
      if (typeof data.error === "string" && data.error.trim()) return data.error;
      if (typeof data.message === "string" && data.message.trim()) return data.message;
    }
    const status = (err as { status?: unknown; statusCode?: unknown }).status;
    const code =
      typeof status === "number"
        ? status
        : typeof (err as { statusCode?: unknown }).statusCode === "number"
          ? ((err as { statusCode?: unknown }).statusCode as number)
          : null;
    const message = (err as { message?: unknown }).message;
    const base =
      typeof message === "string" && message.trim() ? message : fallback;
    return code ? `${code} · ${base}` : base;
  }

  async function override(breachId: string, payload: ComplianceOverrideRequest) {
    submitting.value = true;
    error.value = null;
    try {
      lastResult.value = await complianceApi.overrideBreach(breachId, payload);
      return lastResult.value;
    } catch (err) {
      error.value = extractErrorMessage(err, "Failed to record override.");
      throw err;
    } finally {
      submitting.value = false;
    }
  }

  function reset() {
    error.value = null;
    lastResult.value = null;
  }

  return { submitting, error, lastResult, override, reset };
}
