import { ref } from "vue";
import type { DashboardSnapshotDTO } from "../types";
import { dashboardApi } from "../services/dashboardApi";

export function useDashboardTasks() {
  const loading = ref(false);
  const error = ref<string | null>(null);
  const snapshot = ref<DashboardSnapshotDTO | null>(null);

  async function fetchTasks() {
    loading.value = true;
    error.value = null;

    try {
      snapshot.value = await dashboardApi.snapshot();
    } catch (err) {
      snapshot.value = null;
      error.value =
        err instanceof Error ? err.message : "Failed to load task feed";
    } finally {
      loading.value = false;
    }
  }

  return {
    snapshot,
    loading,
    error,
    fetchTasks,
  };
}
