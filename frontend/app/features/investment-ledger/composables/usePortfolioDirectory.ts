import { computed, ref } from "vue";

import {
  investmentLedgerApi,
  type ApiPortfolio,
} from "../services/investmentLedgerApi";

function extractMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
  if (data && typeof data === "object") {
    if (typeof data.error === "string" && data.error.trim()) return data.error;
    if (typeof data.message === "string" && data.message.trim())
      return data.message;
  }
  const msg = (err as { message?: unknown }).message;
  if (typeof msg === "string" && msg.trim()) return msg;
  return fallback;
}

/**
 * Loads the portfolios the user can see and tracks the active one.
 *
 * The Order Ticket flow needs a portfolio context before any other API
 * call can happen, so this composable is the entry point for the page.
 */
export function usePortfolioDirectory() {
  const portfolios = ref<ApiPortfolio[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const activePortfolioId = ref<string | null>(null);

  const activePortfolio = computed<ApiPortfolio | null>(() => {
    if (!activePortfolioId.value) return null;
    return (
      portfolios.value.find((p) => p.id === activePortfolioId.value) ?? null
    );
  });

  async function load(preferredId?: string | null) {
    loading.value = true;
    error.value = null;
    try {
      const list = await investmentLedgerApi.listPortfolios({ limit: 200 });
      portfolios.value = list.items ?? [];
      if (preferredId && portfolios.value.some((p) => p.id === preferredId)) {
        activePortfolioId.value = preferredId;
      } else if (!activePortfolioId.value && portfolios.value[0]?.id) {
        activePortfolioId.value = portfolios.value[0].id;
      }
    } catch (err) {
      error.value = extractMessage(err, "Failed to load portfolios.");
      portfolios.value = [];
    } finally {
      loading.value = false;
    }
  }

  function setActive(id: string) {
    if (portfolios.value.some((p) => p.id === id)) {
      activePortfolioId.value = id;
    }
  }

  return {
    portfolios,
    activePortfolio,
    activePortfolioId,
    loading,
    error,
    load,
    setActive,
  };
}
