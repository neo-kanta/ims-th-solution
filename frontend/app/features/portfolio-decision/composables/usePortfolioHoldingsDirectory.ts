import { computed, ref, shallowRef } from "vue";

import { portfolioApi } from "../../portfolio-workspace/services/portfolioApi";
import type { ApiCashBalanceV2, ApiHoldingV2 } from "../../portfolio-workspace/services/portfolioApi";
import {
  investmentLedgerApi,
  type ApiInstrument,
} from "../../investment-ledger/services/investmentLedgerApi";
import { extractMessage } from "../lib/decisionErrors";

export interface EnrichedHolding {
  instrumentId: string;
  quantity: string;
  quantityNumeric: number;
  averageCost: string;
  instrument: ApiInstrument | null;
}

/**
 * Loads a portfolio's current holdings and cash so the order ticket can show
 * "available cash" for BUY and "owned positions" / available quantity for
 * SELL. HoldingResponse only carries instrument_id (backend/internal/investment/transport/dto/response/responses.go)
 * so each holding is enriched with its instrument via a bounded set of
 * parallel GET /investment/instruments/{id} calls — the existing
 * PortfolioHoldingsView.vue renders the raw UUID instead; this composable
 * deliberately does not repeat that.
 */
export function usePortfolioHoldingsDirectory() {
  const holdings = shallowRef<EnrichedHolding[]>([]);
  const cash = shallowRef<ApiCashBalanceV2[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  // A portfolio switch can fire a new load() before an earlier portfolio's
  // request settles. Only the most recently issued call may write results,
  // and stale holdings/cash are cleared immediately so a switch never shows
  // the previous portfolio's positions while the new one is loading.
  let requestSeq = 0;

  async function load(portfolioCode: string) {
    if (!portfolioCode) return;
    const seq = ++requestSeq;
    holdings.value = [];
    cash.value = [];
    loading.value = true;
    error.value = null;
    try {
      const [rawHoldings, rawCash] = await Promise.all([
        portfolioApi.getHoldings(portfolioCode),
        portfolioApi.getCash(portfolioCode),
      ]);
      if (seq !== requestSeq) return;
      cash.value = rawCash ?? [];

      const withIds = (rawHoldings ?? []).filter(
        (h): h is ApiHoldingV2 & { instrument_id: string } => Boolean(h.instrument_id),
      );
      const instrumentEntries = await Promise.all(
        withIds.map(async (h) => {
          try {
            const inst = await investmentLedgerApi.getInstrument(h.instrument_id);
            return [h.instrument_id, inst] as const;
          } catch {
            return [h.instrument_id, null] as const;
          }
        }),
      );
      const instrumentById = new Map(instrumentEntries);

      if (seq !== requestSeq) return;
      holdings.value = withIds.map((h) => ({
        instrumentId: h.instrument_id,
        quantity: h.quantity ?? "0",
        quantityNumeric: Number(h.quantity ?? 0) || 0,
        averageCost: h.average_cost ?? "0",
        instrument: instrumentById.get(h.instrument_id) ?? null,
      }));
    } catch (err) {
      if (seq !== requestSeq) return;
      holdings.value = [];
      cash.value = [];
      error.value = extractMessage(err, "Failed to load portfolio holdings.");
    } finally {
      if (seq === requestSeq) loading.value = false;
    }
  }

  function holdingFor(instrumentId: string | null | undefined): EnrichedHolding | null {
    if (!instrumentId) return null;
    return holdings.value.find((h) => h.instrumentId === instrumentId) ?? null;
  }

  function cashFor(currency: string | null | undefined): ApiCashBalanceV2 | null {
    if (!currency) return null;
    return cash.value.find((c) => c.currency === currency) ?? null;
  }

  const ownedInstrumentIds = computed(() => new Set(holdings.value.map((h) => h.instrumentId)));

  return {
    holdings,
    cash,
    loading,
    error,
    load,
    holdingFor,
    cashFor,
    ownedInstrumentIds,
  };
}
