<script setup lang="ts">
import type { SecuritySearchResult, SearchFilter } from "../market-data.types";
import { useI18n, type AppTranslationKey } from "~/composables/useI18n";

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
const { t } = useI18n();

function onOpen(id: string) {
  emit("open", id);
}

const filters: { value: SearchFilter; labelKey: AppTranslationKey }[] = [
  { value: "all", labelKey: "marketData.filters.all" },
  { value: "stocks", labelKey: "marketData.filters.stocks" },
  { value: "bonds", labelKey: "marketData.filters.bonds" },
  { value: "etfs", labelKey: "marketData.filters.etfs" },
  { value: "fx", labelKey: "marketData.filters.fx" },
  { value: "sources", labelKey: "marketData.filters.allSources" },
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
        <span class="card-title">{{ t("marketData.headings.findImportSecurity") }}</span>
        <div class="card-subtitle">{{ t("marketData.messages.searchPrompt") }}</div>
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
            :placeholder="t('marketData.placeholders.searchSecurities')"
            :aria-label="t('marketData.labels.search')"
            @input="onInput"
          />
        </div>
        <div class="md-chip-row" role="tablist" :aria-label="t('marketData.labels.source')">
          <button
            v-for="f in filters"
            :key="f.value"
            type="button"
            class="md-chip"
            :class="{ 'md-chip--active': props.filter === f.value }"
            @click="changeFilter(f.value)"
          >
            {{ t(f.labelKey) }}
          </button>
        </div>
      </div>

      <div class="md-search-results">
        <div class="md-search-results__head">
          {{ t("marketData.messages.topMatches") }}
          <span v-if="props.loading" class="md-text-muted"> &middot; {{ t("marketData.messages.searching") }}</span>
        </div>
        <div v-if="props.error && props.results.length === 0" class="md-empty">
          {{ props.error }}
        </div>
        <div v-else-if="!props.loading && props.results.length === 0" class="md-empty">
          {{ props.query ? t("marketData.messages.noMatches", { query: props.query }) : t("marketData.messages.searchPrompt") }}
        </div>
        <div v-else class="md-search-list">
          <div
            v-for="row in props.results"
            :key="row.id"
            class="md-search-row-item md-search-row-item--clickable"
            role="button"
            tabindex="0"
            :aria-label="`${t('marketData.actions.open')} ${row.symbol}`"
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
              <div v-if="row.ytm != null" class="md-text-muted md-num">{{ t("marketData.labels.ytm") }} {{ row.ytm.toFixed(2) }}%</div>
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
                :aria-label="`${t('marketData.actions.open')} ${row.symbol}`"
                @click.stop="onOpen(row.id)"
              >{{ t("marketData.actions.open") }}</button>
              <button
                class="btn btn-sm"
                :class="row.watching ? 'btn-secondary' : 'btn-primary'"
                type="button"
                :disabled="row.watching"
                @click.stop="$emit('add', row.id)"
              >
                {{ row.watching ? t("marketData.actions.watching") : t("marketData.actions.add") }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
