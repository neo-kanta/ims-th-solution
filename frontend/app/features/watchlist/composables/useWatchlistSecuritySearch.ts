import { ref } from "vue";
import { unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { components } from "~/api/ims-api";

export type WatchlistSecurity = components["schemas"]["SecurityDTO"];

export function useWatchlistSecuritySearch() {
  const results = ref<WatchlistSecurity[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function search(query: string) {
    if (!query.trim()) {
      results.value = [];
      return;
    }
    loading.value = true;
    error.value = null;
    try {
      const client = useOpenApiClient();
      const response = await client.GET("/reference-data/securities/search", {
        params: { query: { query, status: "ACTIVE", limit: 50 } },
      });
      const data = unwrapOpenApiResponse<components["schemas"]["SearchResponse"]>(response);
      results.value = data.items ?? [];
    } catch (err: unknown) {
      error.value = err instanceof Error ? err.message : "Security search failed.";
      results.value = [];
    } finally {
      loading.value = false;
    }
  }

  function clear() {
    results.value = [];
  }

  return { results, loading, error, search, clear };
}
