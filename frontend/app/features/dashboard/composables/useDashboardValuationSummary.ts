import { readonly, ref } from "vue";

import { dashboardApi } from "../services/dashboardApi";
import type { ValuationScope, ValuationSummaryDTO } from "../types";

export function useDashboardValuationSummary() {
  const summary = ref<ValuationSummaryDTO | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  // Guards against a slow, earlier request overwriting state after a newer
  // scope switch or refresh already resolved. Only the most recently issued
  // call is allowed to write to summary/loading/error.
  let latestRequestId = 0;

  async function fetchValuationSummary(scope: ValuationScope): Promise<void> {
    const requestId = ++latestRequestId;
    loading.value = true;
    error.value = null;
    try {
      const result = await dashboardApi.valuationSummary(scope);
      if (requestId !== latestRequestId) return;
      summary.value = result;
    } catch (err) {
      if (requestId !== latestRequestId) return;
      summary.value = null;
      error.value =
        err instanceof Error ? err.message : "Failed to load valuation summary";
    } finally {
      if (requestId === latestRequestId) {
        loading.value = false;
      }
    }
  }

  return {
    summary: readonly(summary),
    loading: readonly(loading),
    error: readonly(error),
    fetchValuationSummary,
  };
}
