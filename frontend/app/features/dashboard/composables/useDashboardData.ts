import { readonly, ref, watch } from "vue";
import type { DashboardPayload } from "../types";
import { buildDashboardSnapshot } from "../lib/dashboard";

export function useDashboardData() {
  const authStore = useAuthStore();
  const { t, locale } = useI18n();
  const loading = ref(false);
  const error = ref<string | null>(null);
  const payload = ref<DashboardPayload | null>(null);

  async function fetchDashboardData() {
    loading.value = true;
    error.value = null;

    try {
      payload.value = buildDashboardSnapshot(
        Date.now(),
        authStore.permissions,
        t,
      );
    } catch (err) {
      payload.value = null;
      error.value =
        err instanceof Error ? err.message : "Failed to load dashboard";
    } finally {
      loading.value = false;
    }
  }

  watch(locale, () => {
    if (!payload.value) {
      return;
    }

    void fetchDashboardData();
  });

  return {
    payload: readonly(payload),
    loading: readonly(loading),
    error: readonly(error),
    fetchDashboardData,
  };
}
