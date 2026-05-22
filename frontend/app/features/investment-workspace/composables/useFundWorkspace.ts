import { ref } from "vue";

import { holdingsMockApi } from "../services/holdingsMockApi";
import type { FundCard } from "../types";

/**
 * Lists funds the caller can open and resolves the active fund by slug.
 *
 * MOCK REPLACEMENT POINT — swap the holdingsMockApi import for the real
 * holdingsApi when the backend lands. The composable interface is stable.
 */
export function useFundWorkspace() {
  const funds = ref<FundCard[]>([]);
  const activeFund = ref<FundCard | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function loadFunds(activeSlug?: string) {
    loading.value = true;
    error.value = null;
    try {
      const list = await holdingsMockApi.listMyFunds();
      funds.value = list;
      activeFund.value =
        (activeSlug && list.find((f) => f.fund_id === activeSlug)) || list[0] || null;
    } catch (err) {
      error.value = extractMessage(err, "Failed to load funds.");
    } finally {
      loading.value = false;
    }
  }

  function setActiveFund(slug: string) {
    const next = funds.value.find((f) => f.fund_id === slug);
    if (next) activeFund.value = next;
  }

  return { funds, activeFund, loading, error, loadFunds, setActiveFund };
}

function extractMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const msg = (err as { message?: unknown }).message;
  return typeof msg === "string" && msg.trim() ? msg : fallback;
}
