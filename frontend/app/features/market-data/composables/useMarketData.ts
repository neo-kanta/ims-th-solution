/**
 * Market Data composable.
 *
 * Live wiring:
 *   provider cards  → GET  /market-data/provider-health
 *   watchlist rows  → GET  /market-data/screen/watchlist     (canonical DTOs)
 *   search rows     → GET  /market-data/screen/search        (canonical DTOs)
 *   sparklines      → GET  /market-data/history              (per row)
 *   sync (one)      → POST /market-data/import
 *   sync (all)      → POST /market-data/import-batches + run (see useImportBatches)
 *
 * The canonical screen/* endpoints already return frontend-safe rows — every
 * decimal value comes through as a string, no UI formatting, no provider raw
 * payload. This composable maps those into the existing UI types so the
 * components stay untouched.
 */
import { marketDataApi } from "../services/marketDataApi";
import type {
  ApiProviderHealth,
  ApiScreenRow,
  ApiScreenSearchRow,
} from "../services/marketDataApi";
import type {
  MarketDataProvider,
  ProviderStatus,
  WatchlistItem,
  SyncEvent,
  SyncSeverity,
  SecuritySearchResult,
  AssetClass,
  ProviderId,
  SecurityStatus,
} from "../market-data.types";

function providerIdFor(name?: string): ProviderId {
  if (name === "alpha_vantage") return "alpha-vantage";
  if (name === "yahoo") return "yahoo-finance";
  return "manual";
}

function providerLabel(id: ProviderId): string {
  switch (id) {
    case "alpha-vantage": return "Alpha Vantage";
    case "yahoo-finance": return "Yahoo Finance";
    case "manual":        return "Manual upload";
  }
}

function providerEndpoint(id: ProviderId): string {
  switch (id) {
    case "alpha-vantage": return "alphavantage.co · REST · intraday + fundamentals";
    case "yahoo-finance": return "query1.finance.yahoo.com · unofficial + delayed quotes";
    case "manual":        return "CSV / XLSX / ThaiBMA bond reference";
  }
}

function providerInitials(id: ProviderId): string {
  return id === "alpha-vantage" ? "AV" : id === "yahoo-finance" ? "YF" : "MU";
}

function mapProviderStatus(health: ApiProviderHealth): ProviderStatus {
  if (!health.configured) return "idle";
  if (!health.healthy) {
    if ((health.last_status ?? "").toLowerCase().includes("rate")) return "rate-limited";
    return "error";
  }
  return "connected";
}

function classifyAsset(assetType?: string, symbol?: string): AssetClass {
  const at = (assetType ?? "").toUpperCase();
  if (at === "ETF") return "ETF";
  if (at === "BOND") return "Bond";
  if (at === "FX")   return "FX";
  if (at === "EQUITY") return "Equity";
  if ((symbol ?? "").endsWith(".BK") && (symbol ?? "").includes("-")) return "NVDR";
  return "Equity";
}

function currencySymbol(currency?: string): string {
  if (!currency) return "";
  switch (currency.toUpperCase()) {
    case "THB": return "฿";
    case "USD": return "$";
    case "EUR": return "€";
    case "JPY": return "¥";
    case "GBP": return "£";
    default:    return "";
  }
}

function formatRelative(iso?: string): string {
  if (!iso) return "—";
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return "—";
  const diff = Math.max(0, Date.now() - t);
  const min = Math.floor(diff / 60_000);
  if (min < 1) return "just now";
  if (min < 60) return `${min} min ago`;
  const hrs = Math.floor(min / 60);
  if (hrs < 24) return `${hrs} h ago`;
  return new Date(iso).toISOString().slice(0, 16).replace("T", " ");
}

function compactNumber(n?: number): string | undefined {
  if (n == null || Number.isNaN(n)) return undefined;
  if (Math.abs(n) >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (Math.abs(n) >= 1_000)     return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

function statusFromFreshness(freshness?: string, dataQuality?: string): SecurityStatus {
  const f = (freshness ?? "").toUpperCase();
  const dq = (dataQuality ?? "").toUpperCase();
  if (f === "FRESH" && dq === "OK") return "fresh";
  if (dq === "MISSING" || f === "MISSING") return "failed";
  if (f === "EXPIRED") return "failed";
  return "stale";
}

function parseDecimal(value?: string | null): number | undefined {
  if (value == null || value === "") return undefined;
  const n = Number(value);
  return Number.isFinite(n) ? n : undefined;
}

function rowToWatchlistItem(row: ApiScreenRow, spark: number[] = []): WatchlistItem {
  const providers = (row.provider_badges ?? []).map((b) => providerIdFor(b.provider_code));
  const assetClass = classifyAsset(row.asset_type, row.display_symbol);
  const price = parseDecimal(row.last_price) ?? 0;
  const change = parseDecimal(row.change_percent);
  const ytm = parseDecimal(row.yield_to_maturity);
  return {
    id: row.security_id ?? row.ims_symbol ?? row.display_symbol ?? `row-${Math.random().toString(36).slice(2)}`,
    symbol: row.display_symbol ?? row.ims_symbol ?? "",
    name: row.name ?? row.display_symbol ?? "",
    assetClass,
    source: (row.provider_badges ?? []).map((b) => b.label ?? b.provider_code ?? "").filter(Boolean).join(" · ") || "—",
    providers: providers.length > 0 ? providers : ["manual"],
    price,
    currency: currencySymbol(row.currency),
    change,
    spark,
    volume: assetClass === "Bond" ? undefined : compactNumber(row.volume),
    ytm,
    lastUpdate: formatRelative(row.last_update_at),
    status: statusFromFreshness(row.freshness_status, row.data_quality_status),
    statusDetail: row.data_quality_status && row.data_quality_status !== "OK"
      ? row.data_quality_status
      : undefined,
    pinned: row.pinned,
  };
}

function searchRowToResult(row: ApiScreenSearchRow): SecuritySearchResult {
  const providers = (row.provider_badges ?? []).map((b) => providerIdFor(b.provider_code));
  const assetClass = classifyAsset(row.asset_type, row.display_symbol);
  return {
    id: row.security_id ?? row.ims_symbol ?? row.display_symbol ?? `sr-${Math.random().toString(36).slice(2)}`,
    symbol: row.display_symbol ?? row.ims_symbol ?? "",
    name: row.name ?? row.display_symbol ?? "",
    source: row.action_hint === "MAP_SYMBOL" ? "Needs mapping" : (row.mapping_status ?? "—"),
    assetClass,
    rating: undefined,
    price: parseDecimal(row.last_price) ?? 0,
    currency: currencySymbol(row.currency),
    change: parseDecimal(row.change_percent),
    ytm: parseDecimal(row.yield_to_maturity),
    providers: providers.length > 0 ? providers : ["manual"],
    watching: Boolean(row.watching),
  };
}

export interface UseMarketDataOptions {
  /** Override the page-size for the screen/watchlist endpoint. */
  watchlistLimit?: number;
}

export function useMarketData(options: UseMarketDataOptions = {}) {
  const providers = ref<MarketDataProvider[]>([]);
  const providersLoading = ref(false);
  const providersError = ref<string | null>(null);
  const lastHealth = ref<ApiProviderHealth[]>([]);

  const watchlist = ref<WatchlistItem[]>([]);
  const watchlistLoading = ref(false);
  const watchlistError = ref<string | null>(null);

  const searchResults = ref<SecuritySearchResult[]>([]);
  const searchLoading = ref(false);
  const searchError = ref<string | null>(null);

  const syncActivity = ref<SyncEvent[]>([]);
  const isSyncing = ref(false);

  let activitySeq = 0;
  function pushActivity(provider: string, message: string, severity: SyncSeverity, code?: string, duration?: string) {
    activitySeq += 1;
    syncActivity.value.unshift({
      id: `evt-${activitySeq}-${Date.now()}`,
      timestamp: new Date().toISOString().slice(11, 19),
      provider,
      message,
      severity,
      code,
      duration,
    });
    if (syncActivity.value.length > 60) syncActivity.value.length = 60;
  }

  // ---- providers ----------------------------------------------------------

  async function refreshProviders() {
    providersLoading.value = true;
    providersError.value = null;
    try {
      const health = await marketDataApi.providerHealth();
      lastHealth.value = health;
      providers.value = buildProviderCards(health, watchlist.value);
    } catch (err) {
      providersError.value = err instanceof Error ? err.message : "Failed to load providers";
      providers.value = buildProviderCards(lastHealth.value, watchlist.value);
    } finally {
      providersLoading.value = false;
    }
  }

  function buildProviderCards(
    health: ApiProviderHealth[],
    rows: WatchlistItem[],
  ): MarketDataProvider[] {
    const byProvider = new Map<ProviderId, ApiProviderHealth>();
    for (const h of health) {
      byProvider.set(providerIdFor(h.provider_name), h);
    }
    const symbolsPerProvider: Record<ProviderId, number> = {
      "alpha-vantage": 0,
      "yahoo-finance": 0,
      manual: 0,
    };
    for (const row of rows) {
      for (const p of row.providers) symbolsPerProvider[p] += 1;
    }
    const errorsPerProvider: Record<ProviderId, number> = {
      "alpha-vantage": 0,
      "yahoo-finance": 0,
      manual: 0,
    };
    for (const evt of syncActivity.value) {
      if (evt.severity === "danger") {
        const id = evt.provider === "Alpha Vantage"
          ? "alpha-vantage"
          : evt.provider === "Yahoo Finance"
            ? "yahoo-finance"
            : "manual";
        errorsPerProvider[id] += 1;
      }
    }

    const ids: ProviderId[] = ["alpha-vantage", "yahoo-finance", "manual"];
    return ids.map((id) => {
      const h = byProvider.get(id);
      const lastSync = symbolsPerProvider[id] > 0
        ? formatRelative(new Date().toISOString())
        : h?.last_checked
          ? formatRelative(h.last_checked)
          : undefined;
      const card: MarketDataProvider = {
        id,
        name: providerLabel(id),
        initials: providerInitials(id),
        endpoint: providerEndpoint(id),
        status: id === "manual"
          ? "idle"
          : h
            ? mapProviderStatus(h)
            : "idle",
        symbolsCount: id === "manual" ? undefined : symbolsPerProvider[id],
        recordsCount: id === "manual" ? symbolsPerProvider.manual : undefined,
        lastSync: id === "manual" ? undefined : lastSync,
        lastUpload: id === "manual" ? lastSync : undefined,
        errors24h: id === "manual" ? undefined : errorsPerProvider[id],
        description: id === "manual" ? "Required for instruments without API coverage" : undefined,
        supportedFiles: id === "manual" ? "CSV, XLSX, ThaiBMA" : undefined,
        usageLabel: h?.last_status,
        usageRatio: id === "manual" ? undefined : (h?.healthy ? 0.45 : 0.92),
      };
      return card;
    });
  }

  // ---- watchlist ----------------------------------------------------------

  async function refreshWatchlist() {
    watchlistLoading.value = true;
    watchlistError.value = null;
    try {
      const resp = await marketDataApi.screenWatchlist(options.watchlistLimit ?? 200);
      const rows = resp.items ?? [];
      const items: WatchlistItem[] = [];
      for (const row of rows) {
        items.push(rowToWatchlistItem(row));
      }
      watchlist.value = items;

      // Best-effort sparkline backfill — non-fatal.
      await Promise.allSettled(items.map(async (item) => {
        if (!item.symbol) return;
        try {
          const bars = await marketDataApi.getHistory(item.symbol, 30);
          const series = bars
            .map((b) => b.close ?? 0)
            .filter((v) => typeof v === "number" && Number.isFinite(v));
          if (series.length > 0) item.spark = series;
        } catch {
          /* sparkline is best-effort; ignore failures. */
        }
      }));

      providers.value = buildProviderCards(lastHealth.value, watchlist.value);
    } catch (err) {
      watchlistError.value = err instanceof Error ? err.message : "Failed to load watchlist";
    } finally {
      watchlistLoading.value = false;
    }
  }

  // ---- search -------------------------------------------------------------

  async function search(query: string) {
    const q = query.trim();
    if (q.length < 2) {
      searchResults.value = [];
      return;
    }
    searchLoading.value = true;
    searchError.value = null;
    try {
      const resp = await marketDataApi.screenSearch(q, 50);
      const rows = resp.items ?? [];
      searchResults.value = rows.map(searchRowToResult);
      if (searchResults.value.length === 0) {
        searchError.value = `No match for "${q}"`;
      }
    } catch (err) {
      searchError.value = err instanceof Error ? err.message : "Search failed";
      searchResults.value = [];
    } finally {
      searchLoading.value = false;
    }
  }

  // ---- single-symbol sync (legacy import endpoint, kept for "add" flow) ---

  async function importOne(symbol: string) {
    const started = Date.now();
    try {
      const result = await marketDataApi.importSymbol({
        symbol,
        include_quote: true,
        include_history: true,
        history_limit: 30,
      });
      pushActivity(
        result.quote_provider === "alpha_vantage" ? "Alpha Vantage" : "Yahoo Finance",
        `imported ${symbol}`,
        "success",
        "200 OK",
        `${((Date.now() - started) / 1000).toFixed(1)}s`,
      );
    } catch (err) {
      pushActivity("Provider", `failed ${symbol} · ${err instanceof Error ? err.message : "error"}`, "danger", "ERR");
    }
  }

  function addSymbol(symbol: string) {
    const sym = symbol.trim().toUpperCase();
    if (!sym) return;
    pushActivity("User", `requested import for ${sym}`, "info");
    void importOne(sym).then(() => refreshWatchlist());
  }

  // ---- batch sync ---------------------------------------------------------
  //
  // "Sync all" creates a chunk-import batch over the current watchlist
  // symbols, runs it synchronously, and shows a single aggregated activity
  // event. Per-symbol failures are visible in the dedicated batch viewer.

  async function syncAll() {
    if (isSyncing.value) return;
    if (watchlist.value.length === 0) return;
    isSyncing.value = true;
    const symbols = watchlist.value
      .map((r) => r.symbol)
      .filter((s): s is string => Boolean(s));
    pushActivity("Scheduler", `Batch sync triggered · ${symbols.length} symbols`, "info");
    const started = Date.now();
    try {
      const batch = await marketDataApi.createImportBatch({
        provider: "default",
        import_type: "QUOTE_SYNC",
        symbols,
        chunk_size: 25,
        include_quote: true,
        include_history: false,
      });
      if (!batch.batch_id) throw new Error("batch_id missing from response");
      const result = await marketDataApi.runImportBatch(batch.batch_id);
      pushActivity(
        "Scheduler",
        `Batch ${result.status ?? "complete"} · ${result.accepted_records ?? 0} accepted · ${result.rejected_records ?? 0} rejected`,
        (result.rejected_records ?? 0) > 0 ? "warning" : "success",
        result.status ?? "",
        `${((Date.now() - started) / 1000).toFixed(1)}s`,
      );
      await Promise.all([refreshProviders(), refreshWatchlist()]);
    } catch (err) {
      pushActivity("Scheduler", `Batch failed · ${err instanceof Error ? err.message : "error"}`, "danger", "ERR");
    } finally {
      isSyncing.value = false;
    }
  }

  async function init() {
    await Promise.all([refreshProviders(), refreshWatchlist()]);
  }

  return {
    // state
    providers,
    providersLoading,
    providersError,

    watchlist,
    watchlistLoading,
    watchlistError,

    searchResults,
    searchLoading,
    searchError,

    syncActivity,
    isSyncing,

    // actions
    init,
    refreshProviders,
    refreshWatchlist,
    addSymbol,
    search,
    syncAll,
    importOne,
  };
}
