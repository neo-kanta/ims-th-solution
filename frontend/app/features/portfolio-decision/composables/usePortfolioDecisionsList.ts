import { ref, shallowRef } from "vue";

import { portfolioDecisionApi, type ApiDecisionV2 } from "../services/portfolioDecisionApi";
import { extractMessage } from "../lib/decisionErrors";

export function usePortfolioDecisionsList() {
  const items = shallowRef<ApiDecisionV2[]>([]);
  const total = ref(0);
  const page = ref(1);
  const limit = ref(20);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function load(portfolioCode: string) {
    if (!portfolioCode) return;
    loading.value = true;
    error.value = null;
    try {
      const result = await portfolioDecisionApi.list(portfolioCode, {
        page: page.value,
        limit: limit.value,
      });
      items.value = result.items ?? [];
      total.value = result.total ?? items.value.length;
    } catch (err) {
      items.value = [];
      total.value = 0;
      error.value = extractMessage(err, "Failed to load decisions.");
    } finally {
      loading.value = false;
    }
  }

  function setPage(next: number, portfolioCode: string) {
    page.value = next;
    void load(portfolioCode);
  }

  return { items, total, page, limit, loading, error, load, setPage };
}
