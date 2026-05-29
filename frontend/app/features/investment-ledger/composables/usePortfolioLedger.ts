import { ref } from "vue";

import {
  investmentLedgerApi,
  type ApiCashBalance,
  type ApiHolding,
  type ApiTransaction,
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
 * Loads the live state of a portfolio: holdings, cash balances, and the most
 * recent transactions. Each section tracks its own loading flag so a refresh
 * after a successful post can update them in parallel without blocking the UI.
 */
export function usePortfolioLedger() {
  const holdings = ref<ApiHolding[]>([]);
  const cash = ref<ApiCashBalance[]>([]);
  const transactions = ref<ApiTransaction[]>([]);
  const transactionsTotal = ref<number>(0);

  const loadingHoldings = ref(false);
  const loadingCash = ref(false);
  const loadingTransactions = ref(false);

  const holdingsError = ref<string | null>(null);
  const cashError = ref<string | null>(null);
  const transactionsError = ref<string | null>(null);

  const transactionLimit = ref(50);

  async function loadHoldings(portfolioId: string) {
    loadingHoldings.value = true;
    holdingsError.value = null;
    try {
      holdings.value = await investmentLedgerApi.listHoldings(portfolioId);
    } catch (err) {
      holdingsError.value = extractMessage(err, "Failed to load holdings.");
      holdings.value = [];
    } finally {
      loadingHoldings.value = false;
    }
  }

  async function loadCash(portfolioId: string) {
    loadingCash.value = true;
    cashError.value = null;
    try {
      cash.value = await investmentLedgerApi.listCash(portfolioId);
    } catch (err) {
      cashError.value = extractMessage(err, "Failed to load cash balances.");
      cash.value = [];
    } finally {
      loadingCash.value = false;
    }
  }

  async function loadTransactions(portfolioId: string) {
    loadingTransactions.value = true;
    transactionsError.value = null;
    try {
      const list = await investmentLedgerApi.listTransactions(portfolioId, {
        limit: transactionLimit.value,
      });
      transactions.value = list.items ?? [];
      transactionsTotal.value = list.total ?? transactions.value.length;
    } catch (err) {
      transactionsError.value = extractMessage(
        err,
        "Failed to load transactions.",
      );
      transactions.value = [];
      transactionsTotal.value = 0;
    } finally {
      loadingTransactions.value = false;
    }
  }

  async function refreshAll(portfolioId: string) {
    await Promise.allSettled([
      loadHoldings(portfolioId),
      loadCash(portfolioId),
      loadTransactions(portfolioId),
    ]);
  }

  function reset() {
    holdings.value = [];
    cash.value = [];
    transactions.value = [];
    transactionsTotal.value = 0;
    holdingsError.value = null;
    cashError.value = null;
    transactionsError.value = null;
  }

  return {
    holdings,
    cash,
    transactions,
    transactionsTotal,
    loadingHoldings,
    loadingCash,
    loadingTransactions,
    holdingsError,
    cashError,
    transactionsError,
    transactionLimit,
    loadHoldings,
    loadCash,
    loadTransactions,
    refreshAll,
    reset,
  };
}
