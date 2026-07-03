import { ref } from "vue";
import {
  decisionApi,
  type ApiDecision,
  type DecisionListFilters,
} from "../services/decisionApi";

export function useDecisionList() {
  const items = ref<ApiDecision[]>([]);
  const total = ref(0);
  const page = ref(1);
  const limit = ref(50);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchList(filters: DecisionListFilters = {}) {
    loading.value = true;
    error.value = null;
    try {
      const result = await decisionApi.listDecisions({
        ...filters,
        page: page.value,
        limit: limit.value,
      });
      items.value = result.items ?? [];
      total.value = result.total ?? 0;
    } catch (e) {
      error.value = e instanceof Error ? e.message : "Failed to load decisions";
      items.value = [];
      total.value = 0;
    } finally {
      loading.value = false;
    }
  }

  return { items, total, page, limit, loading, error, fetchList };
}
