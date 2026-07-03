import { ref } from "vue";
import type { AlertEvent, ListAlertsResult, AlertListQuery, AckAlertBody } from "../types";
import { watchlistApi } from "../services/watchlistApi";

export function useWatchlistAlerts() {
  const alerts = ref<AlertEvent[]>([]);
  const pagination = ref<ListAlertsResult["pagination"]>(undefined);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const acknowledging = ref(false);

  async function load(query: AlertListQuery = {}) {
    loading.value = true;
    error.value = null;
    try {
      const result = await watchlistApi.listAlerts(query);
      alerts.value = result.items ?? [];
      pagination.value = result.pagination;
    } catch (err: unknown) {
      error.value = err instanceof Error ? err.message : "Failed to load alerts.";
      alerts.value = [];
      throw err;
    } finally {
      loading.value = false;
    }
  }

  async function acknowledge(id: string, body: AckAlertBody = {}): Promise<AlertEvent> {
    acknowledging.value = true;
    try {
      return await watchlistApi.acknowledgeAlert(id, body);
    } finally {
      acknowledging.value = false;
    }
  }

  return {
    alerts,
    pagination,
    loading,
    error,
    acknowledging,
    load,
    acknowledge,
  };
}
