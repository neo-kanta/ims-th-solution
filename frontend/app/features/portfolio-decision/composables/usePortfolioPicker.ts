import { computed, ref } from "vue";

import { investmentLedgerApi } from "~/features/investment-ledger/services/investmentLedgerApi";

import { filterPortfolioPicks, toPortfolioPicks, type PortfolioPick } from "../lib/portfolioPicker";

function extractErrorMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
  if (data) {
    if (typeof data.error === "string" && data.error.trim()) return data.error;
    if (typeof data.message === "string" && data.message.trim()) return data.message;
  }
  const message = (err as { message?: unknown }).message;
  if (typeof message === "string" && message.trim()) return message;
  return fallback;
}

/**
 * Single-call, portfolio-first directory for the "Buy / Sell" entry point.
 * Loads every portfolio the caller can see (no fund_id filter) so the user
 * never has to pick a fund before reaching the decision form.
 */
export function usePortfolioPicker() {
  const portfolios = ref<PortfolioPick[]>([]);
  const search = ref("");
  const loading = ref(false);
  const error = ref<string | null>(null);

  const filtered = computed(() => filterPortfolioPicks(portfolios.value, search.value));

  async function load() {
    loading.value = true;
    error.value = null;
    try {
      const result = await investmentLedgerApi.listPortfolios({ limit: 200 });
      portfolios.value = toPortfolioPicks(result.items ?? []);
    } catch (err) {
      portfolios.value = [];
      error.value = extractErrorMessage(err, "Failed to load portfolios.");
    } finally {
      loading.value = false;
    }
  }

  return { portfolios, filtered, search, loading, error, load };
}
