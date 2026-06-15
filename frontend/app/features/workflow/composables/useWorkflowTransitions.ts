import { ref } from "vue";
import { workflowApi } from "../api/workflow.api";
import type { components } from "~/api/ims-api";
import { OpenApiRequestError } from "~/api/openapi";

export function useWorkflowTransitions() {
  const transitions = ref<components["schemas"]["DailyTimelineEntry"][]>([]);
  const page = ref(1);
  const pageSize = ref(50);
  const total = ref(0);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchTransitions(businessDate: string) {
    if (!businessDate) return;
    loading.value = true;
    error.value = null;
    try {
      const response = await workflowApi.getDailyTransitions(
        businessDate,
        page.value,
        pageSize.value,
      );
      transitions.value = response.transitions ?? [];
      total.value = response.total ?? 0;
    } catch (err: any) {
      transitions.value = [];
      total.value = 0;
      error.value = err instanceof OpenApiRequestError ? err.message : (err?.message || "Failed to load transition history");
    } finally {
      loading.value = false;
    }
  }

  return {
    transitions,
    page,
    pageSize,
    total,
    loading,
    error,
    fetchTransitions,
  };
}
