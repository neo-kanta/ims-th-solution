import { ref } from "vue";

import { extractComplianceErrorMessage, isComplianceConflict } from "../lib/errors";
import { complianceApi } from "../services/complianceApi";
import type {
  ComplianceBreach,
  ComplianceBreachListFilters,
  ComplianceOverride,
  ComplianceOverrideRequest,
} from "../types";

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
  const loaded = ref(false);
  const error = ref<string | null>(null);
  let requestId = 0;

  async function fetchList(filters: ComplianceBreachListFilters = {}) {
    const currentRequest = ++requestId;
    loading.value = true;
    error.value = null;
    try {
      const payload = await complianceApi.listBreaches({
        offset: offset.value,
        limit: limit.value,
        ...filters,
      });
      if (currentRequest !== requestId) return;
      items.value = payload.breaches ?? [];
      total.value = payload.total ?? 0;
      offset.value = payload.offset ?? offset.value;
      limit.value = payload.limit ?? limit.value;
    } catch (err) {
      if (currentRequest !== requestId) return;
      error.value = extractComplianceErrorMessage(err, "Failed to load compliance breaches.");
      items.value = [];
      total.value = 0;
    } finally {
      if (currentRequest === requestId) {
        loading.value = false;
        loaded.value = true;
      }
    }
  }

  return {
    items,
    total,
    offset,
    limit,
    loading,
    loaded,
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
  const conflict = ref(false);
  const lastResult = ref<ComplianceOverride | null>(null);

  async function override(breachId: string, payload: ComplianceOverrideRequest) {
    if (submitting.value) {
      throw new Error("An override submission is already in progress.");
    }
    submitting.value = true;
    error.value = null;
    conflict.value = false;
    try {
      lastResult.value = await complianceApi.overrideBreach(breachId, payload);
      return lastResult.value;
    } catch (err) {
      conflict.value = isComplianceConflict(err);
      error.value = extractComplianceErrorMessage(err, "Failed to record override.");
      throw err;
    } finally {
      submitting.value = false;
    }
  }

  function reset() {
    error.value = null;
    conflict.value = false;
    lastResult.value = null;
  }

  return { submitting, error, conflict, lastResult, override, reset };
}
