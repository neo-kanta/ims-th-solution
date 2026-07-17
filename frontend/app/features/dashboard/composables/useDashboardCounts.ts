import { readonly, ref } from "vue";

import { useOpenApiClient, unwrapOpenApiResponse } from "~/api/openapi";

export function useDashboardCounts() {
  const activeContractsCount = ref(0);
  const pendingApprovalsCount = ref(0);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchCounts(): Promise<void> {
    loading.value = true;
    error.value = null;

    const client = useOpenApiClient();

    const [fundsResult, inboxResult] = await Promise.allSettled([
      client.GET("/investment/funds", { params: { query: { limit: 1 } } }),
      client.GET("/approvals/inbox", { params: { query: { limit: 1 } } }),
    ]);

    if (fundsResult.status === "fulfilled") {
      try {
        const data = unwrapOpenApiResponse(fundsResult.value);
        activeContractsCount.value = data?.total ?? 0;
      } catch {
        activeContractsCount.value = 0;
      }
    } else {
      activeContractsCount.value = 0;
    }

    if (inboxResult.status === "fulfilled") {
      try {
        const data = unwrapOpenApiResponse(inboxResult.value);
        pendingApprovalsCount.value = data?.total ?? 0;
      } catch {
        pendingApprovalsCount.value = 0;
      }
    } else {
      pendingApprovalsCount.value = 0;
    }

    const fundsFailed = fundsResult.status === "rejected";
    if (fundsFailed) {
      error.value = "Some count data could not be loaded";
    }

    loading.value = false;
  }

  return {
    activeContractsCount: readonly(activeContractsCount),
    pendingApprovalsCount: readonly(pendingApprovalsCount),
    loading: readonly(loading),
    error: readonly(error),
    fetchCounts,
  };
}
