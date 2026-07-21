import { ref } from "vue";

import { extractComplianceErrorMessage } from "../lib/errors";
import { complianceApi } from "../services/complianceApi";
import type { ComplianceCheckGroupResult } from "../types";

/**
 * Loads one check group's records + breaches on demand — used by the breach
 * detail drawer's evidence section. Race-safe: a slow, superseded request
 * cannot overwrite a newer group's result (e.g. the operator opens a second
 * row before the first group finishes loading).
 */
export function useComplianceCheckGroup() {
  const result = ref<ComplianceCheckGroupResult | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);
  let requestId = 0;

  async function fetch(groupId: string) {
    const currentRequest = ++requestId;
    loading.value = true;
    error.value = null;
    try {
      const payload = await complianceApi.getCheckGroup(groupId);
      if (currentRequest !== requestId) return;
      result.value = payload;
    } catch (err) {
      if (currentRequest !== requestId) return;
      error.value = extractComplianceErrorMessage(err, "Failed to load the check group.");
      result.value = null;
    } finally {
      if (currentRequest === requestId) {
        loading.value = false;
      }
    }
  }

  function clear() {
    requestId += 1;
    result.value = null;
    error.value = null;
    loading.value = false;
  }

  return { result, loading, error, fetch, clear };
}
