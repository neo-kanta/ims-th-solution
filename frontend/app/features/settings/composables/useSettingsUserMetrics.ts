import { ref } from "vue";

import { adminApi } from "../services/adminApi";

export interface UserMetricsState {
  total: number;
  active: number;
  locked: number;
}

export interface UseSettingsUserMetricsOptions {
  enabled: () => boolean;
  onError?: (error: unknown) => void;
}

/**
 * Loads the three user-status totals shown in the Settings KPI grid.
 *
 * NOTE: the backend has no `/admin/users/stats` endpoint yet, so this issues
 * three count-only `listUsers({ limit: 1 })` calls in parallel. Replace with
 * a single endpoint when one becomes available — see the backend track in
 * the Settings refactor plan.
 */
export function useSettingsUserMetrics(opts: UseSettingsUserMetricsOptions) {
  const userMetrics = ref<UserMetricsState>({ total: 0, active: 0, locked: 0 });
  const userMetricsLoading = ref(false);

  async function loadUserMetrics(): Promise<void> {
    if (!opts.enabled()) return;

    userMetricsLoading.value = true;

    try {
      const [total, active, locked] = await Promise.all([
        adminApi.listUsers({ offset: 0, limit: 1 }),
        adminApi.listUsers({ is_active: true, offset: 0, limit: 1 }),
        adminApi.listUsers({ is_locked: true, offset: 0, limit: 1 }),
      ]);

      userMetrics.value = {
        total: total.total,
        active: active.total,
        locked: locked.total,
      };
    } catch (error) {
      opts.onError?.(error);
    } finally {
      userMetricsLoading.value = false;
    }
  }

  return { userMetrics, userMetricsLoading, loadUserMetrics };
}
