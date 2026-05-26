/**
 * Composable powering the Market Data → Security Detail page.
 *
 * Backend has no aggregator endpoint yet, so we compose the view model from
 * four reads:
 *   1. GET /reference-data/securities/{id}          — identity + mappings
 *      (with a search fallback for IMS-symbol URLs)
 *   2. GET /market-data/quote?symbol=…              — latest quote
 *   3. GET /market-data/history?symbol=…&limit=250  — price history
 *   4. (TODO) GET /market-data/sync-activity?security_id=…
 *      — not implemented on the backend; we render an empty state.
 *
 * Sync actions go through the import-batches endpoints, exactly mirroring the
 * recipes the task spec lays out.
 *
 * Decimal values stay as strings end-to-end — the components handle
 * formatting and colouring.
 */
import { marketDataApi } from "../services/marketDataApi";
import type { ApiPriceBar, ApiQuote } from "../services/marketDataApi";
import { referenceDataApi } from "~/features/reference-data/services/referenceDataApi";
import type {
  ApiSecurity,
  ApiProviderMapping,
} from "~/features/reference-data/services/referenceDataApi";
import { useImportBatches } from "./useImportBatches";
import { deriveDataQuality } from "./dataQuality";
import type {
  MarketDataSecurityDetail,
  ProviderMappingView,
  MarketQuoteView,
  MarketPricePoint,
  SyncActivityItem,
} from "../market-data.types";

export { deriveDataQuality };

const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function looksLikeUuid(value: string): boolean {
  return UUID_PATTERN.test(value);
}

function toMappingView(m: ApiProviderMapping): ProviderMappingView {
  return {
    mappingId: m.mapping_id ?? "",
    providerCode: m.provider_code ?? "",
    providerSymbol: m.provider_symbol ?? "",
    providerExchange: m.provider_exchange,
    providerAssetType: m.provider_asset_type,
    providerCurrency: m.provider_currency,
    priority: m.priority,
    status: m.mapping_status ?? "ACTIVE",
    isPrimary: Boolean(m.is_primary),
  };
}

function toQuoteView(q: ApiQuote): MarketQuoteView {
  return {
    provider: q.provider,
    lastPrice: q.price != null ? String(q.price) : undefined,
    changeAmount: q.change != null ? String(q.change) : undefined,
    changePercent: q.change_percent != null ? String(q.change_percent) : undefined,
    open: q.open != null ? String(q.open) : undefined,
    high: q.high != null ? String(q.high) : undefined,
    low: q.low != null ? String(q.low) : undefined,
    previousClose: q.previous_close != null ? String(q.previous_close) : undefined,
    volume: q.volume,
    currency: q.currency,
    asOf: q.as_of,
    capturedAt: q.captured_at,
    stale: q.stale,
    staleReason: q.stale_reason,
  };
}

function toPricePoint(bar: ApiPriceBar): MarketPricePoint | null {
  if (bar.close == null || bar.date == null) return null;
  return {
    date: String(bar.date),
    close: Number(bar.close),
    adjustedClose: bar.adjusted_close != null ? Number(bar.adjusted_close) : undefined,
    open: bar.open != null ? Number(bar.open) : undefined,
    high: bar.high != null ? Number(bar.high) : undefined,
    low: bar.low != null ? Number(bar.low) : undefined,
    volume: bar.volume,
  };
}

function generateIdempotencyKey(prefix: string, key: string): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return `${prefix}-${key}-${crypto.randomUUID()}`;
  }
  return `${prefix}-${key}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

export interface SyncOutcome {
  status: string;
  acceptedRecords: number;
  rejectedRecords: number;
  warningRecords: number;
  unmapped: boolean;
  reviewRequired: boolean;
  rateLimited: boolean;
  errorMessage?: string;
}

export function useMarketDataSecurityDetail(securityIdRef: Ref<string>) {
  const detail = ref<MarketDataSecurityDetail | null>(null);
  const loading = ref(false);
  const loadingQuote = ref(false);
  const loadingHistory = ref(false);
  const error = ref<string | null>(null);
  const notFound = ref(false);

  const lastSync = ref<SyncOutcome | null>(null);
  const syncActivity = ref<SyncActivityItem[]>([]);
  const batches = useImportBatches();

  let activitySeq = 0;
  function pushActivity(item: Omit<SyncActivityItem, "id" | "timestamp"> & { timestamp?: string }) {
    activitySeq += 1;
    syncActivity.value.unshift({
      id: `act-${activitySeq}-${Date.now()}`,
      timestamp: item.timestamp ?? new Date().toISOString(),
      ...item,
    });
    if (syncActivity.value.length > 30) syncActivity.value.length = 30;
  }

  async function resolveSecurity(id: string): Promise<ApiSecurity | null> {
    try {
      if (looksLikeUuid(id)) {
        return await referenceDataApi.getSecurity(id);
      }
    } catch (err) {
      const status = (err as { status?: number } | null)?.status;
      if (status !== 404 && status != null) {
        throw err;
      }
    }
    // Fallback: route param is an IMS symbol or display symbol — search.
    try {
      const resp = await referenceDataApi.searchSecurities({ query: id, limit: 5 });
      const items = resp.items ?? [];
      const upper = id.toUpperCase();
      const exact = items.find(
        (s) => s.ims_symbol === upper || s.display_symbol === id || s.security_id === id,
      );
      return exact ?? items[0] ?? null;
    } catch {
      return null;
    }
  }

  async function load(symbolHint?: string) {
    const id = securityIdRef.value?.trim();
    if (!id) return;
    loading.value = true;
    error.value = null;
    notFound.value = false;
    try {
      const sec = await resolveSecurity(id);
      if (!sec) {
        notFound.value = true;
        detail.value = null;
        return;
      }
      const mappings = (sec.provider_mappings ?? []).map(toMappingView);
      const lookupSymbol = symbolHint ?? sec.display_symbol ?? sec.ims_symbol ?? id;

      // Best-effort quote — a 502/404 here is expected for first-time securities.
      loadingQuote.value = true;
      let quoteView: MarketQuoteView | undefined;
      try {
        if (lookupSymbol) {
          const q = await marketDataApi.getQuote(lookupSymbol);
          quoteView = toQuoteView(q);
        }
      } catch {
        quoteView = undefined;
      } finally {
        loadingQuote.value = false;
      }

      // Best-effort history.
      loadingHistory.value = true;
      const history: MarketPricePoint[] = [];
      try {
        if (lookupSymbol) {
          const bars = await marketDataApi.getHistory(lookupSymbol, 250);
          for (const bar of bars) {
            const point = toPricePoint(bar);
            if (point) history.push(point);
          }
          // Backend returns descending order; reverse to ascending for charts.
          history.reverse();
        }
      } catch {
        /* ignore — empty chart state */
      } finally {
        loadingHistory.value = false;
      }

      const dataQuality = deriveDataQuality(mappings, quoteView);

      detail.value = {
        securityId: sec.security_id ?? id,
        imsSymbol: sec.ims_symbol ?? "",
        displaySymbol: sec.display_symbol ?? "",
        name: sec.name ?? "",
        assetType: sec.asset_type ?? "UNKNOWN",
        currency: sec.currency,
        countryCode: sec.country_code,
        exchangeMic: sec.exchange_mic,
        isin: sec.isin,
        figi: sec.figi,
        cusip: sec.cusip,
        status: sec.status ?? "ACTIVE",
        providerMappings: mappings,
        latestQuote: quoteView,
        history,
        dataQuality,
        // TODO: backend GET /market-data/sync-activity?security_id=… is not
        // implemented; we surface in-session activity from sync actions only.
        syncActivity: syncActivity.value.slice(),
        watching: false,
        pinned: false,
      };
    } catch (err) {
      error.value = err instanceof Error ? err.message : "Failed to load security";
    } finally {
      loading.value = false;
    }
  }

  function symbolForImport(): string | null {
    if (!detail.value) return null;
    return detail.value.imsSymbol || detail.value.displaySymbol || null;
  }

  function summariseRunResult(result: Awaited<ReturnType<typeof batches.run>> | null): SyncOutcome {
    const errs = batches.errors.value;
    const unmapped = errs.some((e) => e.status === "UNMAPPED");
    const reviewRequired = errs.some((e) => e.status === "REVIEW_REQUIRED");
    const rateLimited = errs.some((e) => e.status === "RATE_LIMITED");
    return {
      status: result?.status ?? "UNKNOWN",
      acceptedRecords: result?.accepted_records ?? 0,
      rejectedRecords: result?.rejected_records ?? 0,
      warningRecords: result?.warning_records ?? 0,
      unmapped,
      reviewRequired,
      rateLimited,
      errorMessage: errs[0]?.error_message,
    };
  }

  async function syncQuote() {
    const symbol = symbolForImport();
    if (!symbol) return;
    try {
      const result = await batches.createAndRun({
        provider: "default",
        import_type: "QUOTE_SYNC",
        symbols: [symbol],
        chunk_size: 1,
        include_quote: true,
        include_history: false,
        idempotency_key: generateIdempotencyKey("sec-quote", symbol),
      });
      lastSync.value = summariseRunResult(result);
      pushActivity({
        provider: "default",
        operation: "QUOTE_SYNC",
        status: lastSync.value.rejectedRecords > 0 ? "warning" : "success",
        message: `Quote sync · accepted ${lastSync.value.acceptedRecords} · rejected ${lastSync.value.rejectedRecords}`,
      });
      await load(symbol);
    } catch (err) {
      pushActivity({
        provider: "default",
        operation: "QUOTE_SYNC",
        status: "error",
        message: err instanceof Error ? err.message : "Quote sync failed",
      });
    }
  }

  async function importQuoteAndHistory() {
    const symbol = symbolForImport();
    if (!symbol) return;
    try {
      const result = await batches.createAndRun({
        provider: "default",
        import_type: "QUOTE_AND_HISTORY_SYNC",
        symbols: [symbol],
        chunk_size: 1,
        include_quote: true,
        include_history: true,
        history_limit: 250,
        idempotency_key: generateIdempotencyKey("sec-history", symbol),
      });
      lastSync.value = summariseRunResult(result);
      pushActivity({
        provider: "default",
        operation: "QUOTE_AND_HISTORY_SYNC",
        status: lastSync.value.rejectedRecords > 0 ? "warning" : "success",
        message: `History sync · accepted ${lastSync.value.acceptedRecords} · rejected ${lastSync.value.rejectedRecords}`,
      });
      await load(symbol);
    } catch (err) {
      pushActivity({
        provider: "default",
        operation: "QUOTE_AND_HISTORY_SYNC",
        status: "error",
        message: err instanceof Error ? err.message : "History sync failed",
      });
    }
  }

  function toggleWatching() {
    if (detail.value) detail.value.watching = !detail.value.watching;
  }
  function togglePinned() {
    if (detail.value) detail.value.pinned = !detail.value.pinned;
  }

  // Keep history in sync with in-session activity items so they show up
  // immediately in the activity panel.
  watch(syncActivity, (items) => {
    if (detail.value) detail.value.syncActivity = items.slice();
  }, { deep: true });

  // Reload when the route id changes.
  watch(securityIdRef, () => { void load(); });

  return {
    detail,
    loading,
    loadingQuote,
    loadingHistory,
    error,
    notFound,
    lastSync,
    syncActivity,
    isRunning: batches.isRunning,

    // actions
    load,
    refresh: () => load(),
    syncQuote,
    importQuoteAndHistory,
    toggleWatching,
    togglePinned,
  };
}
