<script setup lang="ts">
import { computed } from "vue";
import type { WatchlistItem, WatchlistFilter, ProviderId } from "../market-data.types";
import type { ApiSecurity } from "../services/marketDataCatalogApi";
import MarketDataSparkline from "./MarketDataSparkline.vue";
import { useI18n, type AppTranslationKey } from "~/composables/useI18n";

const props = defineProps<{
  type: "watchlist" | "securities";
  items: any[];
  loading?: boolean;
  filter?: WatchlistFilter;
  query?: string;
}>();

const emit = defineEmits<{
  (e: "update:filter", value: WatchlistFilter): void;
  (e: "update:query", value: string): void;
  (e: "togglePin", id: string): void;
  (e: "open", id: string): void;
  (e: "sync"): void;
}>();
const { t } = useI18n();

// ---- Watchlist filtering & grouping ---------------------------------------

const watchlistTabs: { value: WatchlistFilter; labelKey: AppTranslationKey }[] = [
  { value: "all", labelKey: "marketData.filters.all" },
  { value: "stocks", labelKey: "marketData.filters.stocks" },
  { value: "bonds", labelKey: "marketData.filters.bonds" },
  { value: "fx", labelKey: "marketData.filters.fx" },
  { value: "stale", labelKey: "marketData.filters.stale" },
  { value: "pinned", labelKey: "marketData.filters.pinned" },
];

const watchlistItems = computed(() => {
  if (props.type !== "watchlist") return [];
  return props.items as WatchlistItem[];
});

const filteredWatchlist = computed(() => {
  const q = (props.query || "").trim().toLowerCase();
  return watchlistItems.value.filter((row) => {
    if (props.filter) {
      switch (props.filter) {
        case "stocks": if (row.assetClass !== "Equity" && row.assetClass !== "NVDR") return false; break;
        case "bonds":  if (row.assetClass !== "Bond") return false; break;
        case "fx":     if (row.assetClass !== "FX") return false; break;
        case "stale":  if (row.status !== "stale" && row.status !== "failed") return false; break;
        case "pinned": if (!row.pinned) return false; break;
      }
    }
    if (q && !row.symbol.toLowerCase().includes(q) && !row.name.toLowerCase().includes(q)) {
      return false;
    }
    return true;
  });
});

// Group watchlist items logically by asset class for top-tier UI grouping
const categorizedWatchlist = computed(() => {
  const list = filteredWatchlist.value;
  if (props.filter !== "all") {
    return [{ category: "", items: list }];
  }

  const assetClass = (row: WatchlistItem) => String(row.assetClass);
  const categories: { category: string; items: WatchlistItem[] }[] = [
    { category: t("marketData.groups.equities"), items: list.filter(r => assetClass(r) === "Equity" || assetClass(r) === "NVDR") },
    { category: t("marketData.groups.fixedIncome"), items: list.filter(r => assetClass(r) === "Bond") },
    { category: t("marketData.groups.funds"), items: list.filter(r => assetClass(r) === "ETF" || assetClass(r) === "Mutual Fund" || assetClass(r) === "MUTUAL_FUND") },
    { category: t("marketData.groups.fx"), items: list.filter(r => assetClass(r) === "FX") },
  ];

  const groupedIds = new Set(categories.flatMap(c => c.items.map(i => i.id)));
  const others = list.filter(r => !groupedIds.has(r.id));
  if (others.length > 0) {
    categories.push({ category: t("marketData.groups.otherAssets"), items: others });
  }

  return categories.filter(c => c.items.length > 0);
});

const providerInitials = (p: ProviderId) =>
  p === "alpha-vantage" ? "AV" : p === "yahoo-finance" ? "YF" : "MU";

const statusLabel = (s: WatchlistItem["status"]) => {
  switch (s) {
    case "fresh":        return t("marketData.statuses.fresh");
    case "stale":        return t("marketData.statuses.stale");
    case "failed":       return t("marketData.statuses.failed");
    case "rate-limited": return t("marketData.statuses.rateLimited");
  }
};

// ---- Securities Catalog helper methods -----------------------------------

const securitiesItems = computed(() => {
  if (props.type !== "securities") return [];
  return props.items as ApiSecurity[];
});

function badgeLabel(mapping: NonNullable<ApiSecurity["provider_mappings"]>[number]) {
  return mapping.provider_code ?? "—";
}

function statusVariant(status?: string): string {
  switch (status) {
    case "ACTIVE":    return "badge-success";
    case "INACTIVE":  return "badge-neutral";
    case "SUSPENDED": return "badge-warning";
    default:          return "badge-neutral";
  }
}
</script>

<template>
  <div class="card md-table-card">
    <!-- Header Area -->
    <div class="card-header md-table-header">
      <template v-if="type === 'watchlist'">
        <div class="md-table-title">
          <span class="card-title">{{ t("marketData.headings.watchlist") }}</span>
          <span class="md-count-badge">{{ filteredWatchlist.length }} {{ t("marketData.labels.symbols").toLowerCase() }}</span>
        </div>
        <div class="md-watchlist-controls">
          <div class="md-chip-row" role="tablist" :aria-label="t('marketData.headings.watchlist')">
            <button
              v-for="tab in watchlistTabs"
              :key="tab.value"
              type="button"
              class="md-chip"
              :class="{ 'md-chip--active': filter === tab.value }"
              @click="emit('update:filter', tab.value)"
            >
              {{ t(tab.labelKey) }}
            </button>
          </div>
          <input
            class="form-input md-watchlist-search"
            :value="query"
            :placeholder="t('marketData.placeholders.filterSymbolName')"
            :aria-label="t('marketData.placeholders.filterSymbolName')"
            @input="emit('update:query', ($event.target as HTMLInputElement).value)"
          />
          <button class="btn btn-secondary btn-sm" @click="emit('sync')">
            {{ t("marketData.actions.sync") }}
          </button>
        </div>
      </template>
      <template v-else>
        <div class="md-table-title">
          <span class="card-title">{{ t("marketData.headings.securitiesCatalog") }}</span>
          <span class="md-count-badge">{{ securitiesItems.length }}</span>
        </div>
      </template>
    </div>

    <!-- Standardised Table Wrap and Table Structure -->
    <div class="table-wrap">
      <table class="table">
        <!-- 1. Watchlist Table Mode -->
        <template v-if="type === 'watchlist'">
          <thead>
            <tr>
              <th class="md-col-pin" aria-label="Pin"></th>
              <th>{{ t("marketData.labels.symbols") }}</th>
              <th>{{ t("marketData.labels.name") }}</th>
              <th>{{ t("marketData.labels.class") }}</th>
              <th>{{ t("marketData.labels.source") }}</th>
              <th class="col-right">{{ t("marketData.labels.lastPrice") }}</th>
              <th class="col-right">{{ t("marketData.labels.change") }}</th>
              <th>{{ t("marketData.labels.trend") }}</th>
              <th class="col-right">{{ t("marketData.labels.volumeYtm") }}</th>
              <th>{{ t("marketData.labels.lastUpdate") }}</th>
              <th>{{ t("marketData.labels.status") }}</th>
              <th class="col-action"></th>
            </tr>
          </thead>
          <tbody>
            <template v-for="group in categorizedWatchlist" :key="group.category">
              <!-- Render Section Header Row if category grouping active -->
              <tr v-if="group.category" class="md-table-group-header">
                <td colspan="12">
                  <div class="md-table-group-title">
                    <svg class="octicon" viewBox="0 0 16 16" width="14" height="14" fill="currentColor" aria-hidden="true">
                      <path d="M2 1.75C2 .784 2.784 0 3.75 0h8.5C13.216 0 14 .784 14 1.75v5a.75.75 0 0 1-1.5 0v-5a.25.25 0 0 0-.25-.25h-8.5a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h8.5a.25.25 0 0 0 .25-.25v-2.5a.75.75 0 0 1 1.5 0v2.5A1.75 1.75 0 0 1 12.25 16h-8.5A1.75 1.75 0 0 1 2 14.25v-12.5z"/>
                      <path d="M11.5 7.25a.75.75 0 0 1 1.06 0l2.5 2.5a.75.75 0 0 1 0 1.06l-2.5 2.5a.75.75 0 0 1-1.06-1.06l1.22-1.22H7.75a.75.75 0 0 1 0-1.5h5.47l-1.22-1.22a.75.75 0 0 1 0-1.06z"/>
                    </svg>
                    <span>{{ group.category }}</span>
                    <span class="md-table-group-count">{{ group.items.length }}</span>
                  </div>
                </td>
              </tr>
              <!-- Render Rows -->
              <tr v-for="row in group.items" :key="row.id">
                <td class="md-col-pin">
                  <button
                    class="md-pin-btn"
                    :class="{ 'md-pin-btn--on': row.pinned }"
                    :aria-label="row.pinned ? 'Unpin' : 'Pin'"
                    @click="emit('togglePin', row.id)"
                  >
                    <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
                      <path d="M8 1l2.09 4.26 4.71.69-3.4 3.32.8 4.7L8 11.77l-4.2 2.2.8-4.7-3.4-3.32 4.71-.69L8 1z" />
                    </svg>
                  </button>
                </td>
                <td class="md-cell-symbol">
                  <button
                    type="button"
                    class="md-cell-symbol__link"
                    :aria-label="`${t('marketData.actions.open')} ${row.symbol}`"
                    @click="emit('open', row.id)"
                  >{{ row.symbol }}</button>
                </td>
                <td>
                  <button
                    type="button"
                    class="md-cell-name__link"
                    :aria-label="`${t('marketData.actions.open')} ${row.name}`"
                    @click="emit('open', row.id)"
                  >{{ row.name }}</button>
                </td>
                <td><span class="badge badge-neutral md-asset-pill">{{ row.assetClass }}</span></td>
                <td>
                  <div class="md-providers-cell">
                    <span class="md-source-label">{{ row.source }}</span>
                    <span
                      v-for="p in row.providers"
                      :key="p"
                      class="md-provider-pill md-provider-pill--xs"
                      :data-provider="p"
                      :title="p"
                    >{{ providerInitials(p) }}</span>
                  </div>
                </td>
                <td class="col-right md-num">
                  <span :class="{ 'md-text-muted': row.status === 'stale' || row.status === 'failed' }">
                    {{ row.currency }}{{ row.price.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}
                  </span>
                </td>
                <td class="col-right md-num">
                  <span
                    v-if="row.change != null"
                    :class="row.change >= 0 ? 'md-text-success' : 'md-text-danger'"
                  >
                    {{ row.change >= 0 ? "+" : "" }}{{ row.change.toFixed(2) }}%
                  </span>
                  <span v-else class="md-text-muted">—</span>
                </td>
                <td>
                  <MarketDataSparkline :data="row.spark" />
                </td>
                <td class="col-right md-num">
                  <span v-if="row.ytm != null">{{ t("marketData.labels.ytm") }} {{ row.ytm.toFixed(2) }}%</span>
                  <span v-else-if="row.volume">{{ row.volume }}</span>
                  <span v-else class="md-text-muted">—</span>
                </td>
                <td class="md-text-muted md-cell-update">{{ row.lastUpdate }}</td>
                <td>
                  <span
                    class="md-status-badge"
                    :class="`md-status-badge--${row.status}`"
                    :aria-label="`${t('marketData.labels.status')}: ${statusLabel(row.status)}`"
                  >
                    <span class="md-status-dot" aria-hidden="true" />
                    {{ statusLabel(row.status) }}<template v-if="row.statusDetail"> · {{ row.statusDetail }}</template>
                  </span>
                </td>
                <td class="col-action">
                  <button class="btn btn-ghost btn-icon-sm" aria-label="Row actions">
                    <svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8">
                      <circle cx="8" cy="8" r="1.4" /><circle cx="3" cy="8" r="1.4" /><circle cx="13" cy="8" r="1.4" />
                    </svg>
                  </button>
                </td>
              </tr>
            </template>
            <tr v-if="filteredWatchlist.length === 0">
              <td colspan="12" class="md-empty">
                {{ loading ? t("marketData.messages.loadingMarketData") : t("marketData.messages.filterNoWatchlist") }}
              </td>
            </tr>
          </tbody>
        </template>

        <!-- 2. Securities Catalog Table Mode -->
        <template v-else>
          <thead>
            <tr>
              <th>{{ t("marketData.labels.imsSymbol") }}</th>
              <th>{{ t("marketData.labels.display") }}</th>
              <th>{{ t("marketData.labels.name") }}</th>
              <th>{{ t("marketData.labels.assetType") }}</th>
              <th>{{ t("marketData.labels.ccy") }}</th>
              <th>{{ t("marketData.labels.mic") }}</th>
              <th>{{ t("marketData.labels.isin") }}</th>
              <th>{{ t("marketData.labels.status") }}</th>
              <th>{{ t("marketData.labels.mappings") }}</th>
              <th class="col-action"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sec in securitiesItems" :key="sec.security_id">
              <td class="mdc-mono">{{ sec.ims_symbol }}</td>
              <td>{{ sec.display_symbol }}</td>
              <td>{{ sec.name }}</td>
              <td><span class="badge badge-neutral">{{ sec.asset_type }}</span></td>
              <td>{{ sec.currency || "—" }}</td>
              <td>{{ sec.exchange_mic || "—" }}</td>
              <td class="mdc-mono">{{ sec.isin || "—" }}</td>
              <td><span class="badge" :class="statusVariant(sec.status)">{{ sec.status }}</span></td>
              <td>
                <div class="mdc-mapping-pills">
                  <span
                    v-for="m in sec.provider_mappings"
                    :key="m.mapping_id"
                    class="mdc-pill"
                    :class="{ 'mdc-pill--inactive': m.mapping_status !== 'ACTIVE' }"
                    :title="`${m.provider_code} → ${m.provider_symbol} (${m.mapping_status})`"
                  >
                    {{ badgeLabel(m) }}
                  </span>
                  <span v-if="!sec.provider_mappings || sec.provider_mappings.length === 0" class="mdc-muted">—</span>
                </div>
              </td>
              <td class="col-action">
                <NuxtLink
                  v-if="sec.security_id"
                  :to="`/market-data/securities/${sec.security_id}`"
                  class="btn btn-secondary btn-sm"
                >{{ t("marketData.actions.open") }}</NuxtLink>
              </td>
            </tr>
            <tr v-if="securitiesItems.length === 0">
              <td colspan="10" class="md-empty">
                {{ loading ? t("marketData.messages.loadingSecurities") : t("marketData.messages.filterNoSecurities") }}
              </td>
            </tr>
          </tbody>
        </template>
      </table>
    </div>
  </div>
</template>

<style scoped>
/* Card and Header styles */
.md-table-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}
.md-table-title {
  display: flex;
  align-items: center;
  gap: 8px;
}
.md-count-badge {
  background: var(--md-neutral-bg);
  color: var(--md-neutral-text);
  border: 1px solid var(--md-border);
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 999px;
}

/* Watchlist-specific sub-header controls */
.md-watchlist-controls {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  margin-left: auto;
}
.md-watchlist-search {
  width: 200px;
  height: 28px;
  font-size: 12px;
}
@media (max-width: 900px) {
  .md-watchlist-controls { width: 100%; }
  .md-watchlist-search { flex: 1; min-width: 0; }
}

/* Category Group headers for financial UI */
.md-table-group-header {
  background: var(--md-surface-muted);
}
.md-table-group-header td {
  padding: 0.5rem 0.75rem !important;
  border-bottom: 1px solid var(--md-border);
}
.md-table-group-title {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--md-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.md-table-group-title svg {
  color: var(--md-text-muted);
}
.md-table-group-count {
  background: var(--md-neutral-bg);
  color: var(--md-neutral-text);
  font-size: 0.7rem;
  padding: 0.05rem 0.4rem;
  border-radius: 999px;
  font-weight: 500;
}

/* Securities custom elements */
.mdc-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.78rem;
}
.mdc-mapping-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}
.mdc-pill {
  display: inline-block;
  padding: 0.1rem 0.5rem;
  border-radius: 999px;
  background: var(--md-surface-muted);
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--md-text-secondary);
}
.mdc-pill--inactive {
  opacity: 0.5;
  text-decoration: line-through;
}
.mdc-muted {
  color: var(--md-text-muted);
}
</style>
