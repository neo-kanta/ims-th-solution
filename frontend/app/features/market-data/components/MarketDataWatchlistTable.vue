<script setup lang="ts">
import type { WatchlistItem, WatchlistFilter, ProviderId } from "../market-data.types";
import MarketDataSparkline from "./MarketDataSparkline.vue";

const props = defineProps<{
  items: WatchlistItem[];
  filter: WatchlistFilter;
  query: string;
  loading?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:filter", value: WatchlistFilter): void;
  (e: "update:query", value: string): void;
  (e: "togglePin", id: string): void;
  (e: "open", id: string): void;
  (e: "sync"): void;
}>();

const tabs: { value: WatchlistFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "stocks", label: "Stocks" },
  { value: "bonds", label: "Bonds" },
  { value: "fx", label: "FX" },
  { value: "stale", label: "Stale" },
  { value: "pinned", label: "Pinned" },
];

const filtered = computed(() => {
  const q = props.query.trim().toLowerCase();
  return props.items.filter((row) => {
    switch (props.filter) {
      case "stocks": if (row.assetClass !== "Equity" && row.assetClass !== "NVDR") return false; break;
      case "bonds":  if (row.assetClass !== "Bond") return false; break;
      case "fx":     if (row.assetClass !== "FX") return false; break;
      case "stale":  if (row.status !== "stale" && row.status !== "failed") return false; break;
      case "pinned": if (!row.pinned) return false; break;
    }
    if (q && !row.symbol.toLowerCase().includes(q) && !row.name.toLowerCase().includes(q)) {
      return false;
    }
    return true;
  });
});

const providerInitials = (p: ProviderId) =>
  p === "alpha-vantage" ? "AV" : p === "yahoo-finance" ? "YF" : "MU";

const statusLabel = (s: WatchlistItem["status"]) => {
  switch (s) {
    case "fresh":        return "Fresh";
    case "stale":        return "Stale";
    case "failed":       return "Failed";
    case "rate-limited": return "Rate-limited";
  }
};
</script>

<template>
  <div class="card md-watchlist">
    <div class="card-header md-watchlist__header">
      <div class="md-watchlist__title">
        <span class="card-title">Watchlist</span>
        <span class="md-count-badge">{{ filtered.length }} securities</span>
      </div>
      <div class="md-watchlist__controls">
        <div class="md-chip-row" role="tablist" aria-label="Watchlist filter">
          <button
            v-for="t in tabs"
            :key="t.value"
            type="button"
            class="md-chip"
            :class="{ 'md-chip--active': filter === t.value }"
            @click="emit('update:filter', t.value)"
          >
            {{ t.label }}
          </button>
        </div>
        <input
          class="form-input md-watchlist__search"
          :value="query"
          placeholder="Filter symbol or name"
          aria-label="Filter watchlist"
          @input="emit('update:query', ($event.target as HTMLInputElement).value)"
        />
        <button class="btn btn-secondary btn-sm" @click="emit('sync')">
          Sync
        </button>
      </div>
    </div>

    <div class="md-table-wrap">
      <table class="data-table md-table">
        <thead>
          <tr>
            <th class="md-col-pin" aria-label="Pin"></th>
            <th>Symbol</th>
            <th>Name</th>
            <th>Class</th>
            <th>Source</th>
            <th class="col-right">Last price</th>
            <th class="col-right">Change</th>
            <th>Trend</th>
            <th class="col-right">Volume / YTM</th>
            <th>Last update</th>
            <th>Status</th>
            <th class="col-action"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in filtered" :key="row.id">
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
                :aria-label="`Open detail for ${row.symbol}`"
                @click="emit('open', row.id)"
              >{{ row.symbol }}</button>
            </td>
            <td>
              <button
                type="button"
                class="md-cell-name__link"
                :aria-label="`Open detail for ${row.name}`"
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
              <span v-if="row.ytm != null">YTM {{ row.ytm.toFixed(2) }}%</span>
              <span v-else-if="row.volume">{{ row.volume }}</span>
              <span v-else class="md-text-muted">—</span>
            </td>
            <td class="md-text-muted md-cell-update">{{ row.lastUpdate }}</td>
            <td>
              <span
                class="md-status-badge"
                :class="`md-status-badge--${row.status}`"
                :aria-label="`Status: ${statusLabel(row.status)}`"
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
          <tr v-if="filtered.length === 0">
            <td colspan="12" class="md-empty md-empty--row">
              {{ props.loading ? "Loading market data…" : "No securities match the current filters." }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
