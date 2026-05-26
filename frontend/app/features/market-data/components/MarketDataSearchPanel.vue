<script setup lang="ts">
import type { SecuritySearchResult, SearchFilter } from "../market-data.types";

const props = defineProps<{
  query: string;
  filter: SearchFilter;
  results: SecuritySearchResult[];
  loading?: boolean;
  error?: string | null;
}>();

const emit = defineEmits<{
  (e: "update:query", value: string): void;
  (e: "update:filter", value: SearchFilter): void;
  (e: "add", id: string): void;
  (e: "open", id: string): void;
}>();

function onOpen(id: string) {
  emit("open", id);
}

const filters: { value: SearchFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "stocks", label: "Stocks" },
  { value: "bonds", label: "Bonds" },
  { value: "etfs", label: "ETFs" },
  { value: "fx", label: "FX" },
  { value: "sources", label: "All sources" },
];

function onInput(e: Event) {
  emit("update:query", (e.target as HTMLInputElement).value);
}

function changeFilter(v: SearchFilter) {
  emit("update:filter", v);
}
</script>

<template>
  <div class="card md-search-panel">
    <div class="card-header">
      <div>
        <span class="card-title">Find &amp; import security</span>
        <div class="card-subtitle">Search reference data across providers</div>
      </div>
    </div>

    <div class="card-body md-search-panel__body">
      <div class="md-search-row">
        <div class="md-search-input">
          <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
            <circle cx="7" cy="7" r="5" />
            <path d="M11 11l3 3" />
          </svg>
          <input
            class="form-input"
            :value="props.query"
            placeholder="Search by ticker, ISIN/CUSIP, symbol, or name"
            aria-label="Search securities"
            @input="onInput"
          />
        </div>
        <div class="md-chip-row" role="tablist" aria-label="Source filter">
          <button
            v-for="f in filters"
            :key="f.value"
            type="button"
            class="md-chip"
            :class="{ 'md-chip--active': props.filter === f.value }"
            @click="changeFilter(f.value)"
          >
            {{ f.label }}
          </button>
        </div>
      </div>

      <div class="md-search-results">
        <div class="md-search-results__head">
          Top matches
          <span v-if="props.loading" class="md-text-muted"> · searching…</span>
        </div>
        <div v-if="props.error && props.results.length === 0" class="md-empty">
          {{ props.error }}
        </div>
        <div v-else-if="!props.loading && props.results.length === 0" class="md-empty">
          {{ props.query ? `No matches for "${props.query}"` : "Type a ticker, ISIN or symbol to search." }}
        </div>
        <div v-else class="md-search-list">
          <div
            v-for="row in props.results"
            :key="row.id"
            class="md-search-row-item md-search-row-item--clickable"
            role="button"
            tabindex="0"
            :aria-label="`Open detail for ${row.symbol}`"
            @click="onOpen(row.id)"
            @keydown.enter.prevent="onOpen(row.id)"
            @keydown.space.prevent="onOpen(row.id)"
          >
            <div class="md-search-row-item__main">
              <div class="md-search-row-item__symbol">{{ row.symbol }}</div>
              <div class="md-search-row-item__name">{{ row.name }}</div>
              <div class="md-search-row-item__source">{{ row.source }}</div>
            </div>
            <div class="md-search-row-item__price">
              <div class="md-num">{{ row.currency }}{{ row.price.toLocaleString(undefined, { minimumFractionDigits: 2 }) }}</div>
              <div v-if="row.ytm != null" class="md-text-muted md-num">YTM {{ row.ytm.toFixed(2) }}%</div>
              <div
                v-else-if="row.change != null"
                class="md-num"
                :class="row.change >= 0 ? 'md-text-success' : 'md-text-danger'"
              >
                {{ row.change >= 0 ? "+" : "" }}{{ row.change.toFixed(2) }}%
              </div>
            </div>
            <div class="md-search-row-item__providers">
              <span
                v-for="p in row.providers"
                :key="p"
                class="md-provider-pill"
                :data-provider="p"
              >{{ p === "alpha-vantage" ? "AV" : p === "yahoo-finance" ? "YF" : "MU" }}</span>
            </div>
            <div class="md-search-row-item__action" @click.stop>
              <button
                class="btn btn-sm btn-secondary"
                type="button"
                :aria-label="`Open ${row.symbol}`"
                @click.stop="onOpen(row.id)"
              >Open</button>
              <button
                class="btn btn-sm"
                :class="row.watching ? 'btn-secondary' : 'btn-primary'"
                type="button"
                :disabled="row.watching"
                @click.stop="$emit('add', row.id)"
              >
                {{ row.watching ? "Watching" : "Add" }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
