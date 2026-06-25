import { ref } from "vue";
import { decisionApi, type ApiDecision } from "../services/decisionApi";

export function useDecisionDetail() {
  const decision = ref<ApiDecision | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetch(id: string) {
    loading.value = true;
    error.value = null;
    try {
      decision.value = await decisionApi.getDecisionDetail(id);
    } catch (e) {
      error.value = e instanceof Error ? e.message : "Failed to load decision";
      decision.value = null;
    } finally {
      loading.value = false;
    }
  }

  return { decision, loading, error, fetch };
}
