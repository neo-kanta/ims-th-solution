import { ref } from "vue";
import { unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { components } from "~/api/ims-api";

export type WatchlistPortfolio = components["schemas"]["PortfolioResponse"];

export function useWatchlistPortfolios() {
  const portfolios = ref<WatchlistPortfolio[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function load() {
    loading.value = true;
    error.value = null;
    try {
      const client = useOpenApiClient();
      const response = await client.GET("/investment/portfolios", {
        params: { query: { limit: 200 } },
      });
      const data = unwrapOpenApiResponse<components["schemas"]["PortfolioListResponse"]>(response);
      portfolios.value = data.items ?? [];
    } catch (err: unknown) {
      error.value = err instanceof Error ? err.message : "Failed to load portfolios.";
      portfolios.value = [];
      throw err;
    } finally {
      loading.value = false;
    }
  }

  return { portfolios, loading, error, load };
}
