<script setup lang="ts">
import type {
  MarketDataSecurityDetail,
  DetailFreshnessStatus,
} from "../market-data.types";
import { useI18n } from "~/composables/useI18n";

const props = defineProps<{
  detail: MarketDataSecurityDetail;
  isRunning?: boolean;
}>();

const emit = defineEmits<{
  (e: "sync-quote"): void;
  (e: "import-history"): void;
  (e: "open-mappings"): void;
  (e: "toggle-watching"): void;
  (e: "back"): void;
}>();
const { t } = useI18n();

function freshnessLabel(s: DetailFreshnessStatus): string {
  switch (s) {
    case "FRESH":        return t("marketData.statuses.fresh");
    case "STALE":        return t("marketData.statuses.stale");
    case "EXPIRED":      return t("marketData.statuses.expired");
    case "NOT_IMPORTED": return t("marketData.statuses.notImported");
    case "FAILED":       return t("marketData.statuses.failed");
    case "RATE_LIMITED": return t("marketData.statuses.rateLimited");
    default:             return s;
  }
}
function freshnessClass(s: DetailFreshnessStatus): string {
  switch (s) {
    case "FRESH":        return "md-status-badge--fresh";
    case "STALE":        return "md-status-badge--stale";
    case "EXPIRED":      return "md-status-badge--failed";
    case "NOT_IMPORTED": return "md-sec-header__pill--neutral";
    case "FAILED":       return "md-status-badge--failed";
    case "RATE_LIMITED": return "md-status-badge--rate-limited";
    default:             return "md-sec-header__pill--neutral";
  }
}

const mappingLabel = computed(() => {
  switch (props.detail.dataQuality.mapping) {
    case "MAPPED":          return t("marketData.statuses.mapped");
    case "UNMAPPED":        return t("marketData.statuses.unmapped");
    case "REVIEW_REQUIRED": return t("marketData.statuses.reviewRequired");
    case "CONFLICTED":      return t("marketData.statuses.conflicted");
    default:                return props.detail.dataQuality.mapping;
  }
});
const mappingClass = computed(() => {
  switch (props.detail.dataQuality.mapping) {
    case "MAPPED":          return "md-status-badge--fresh";
    case "UNMAPPED":        return "md-status-badge--failed";
    case "REVIEW_REQUIRED": return "md-status-badge--stale";
    case "CONFLICTED":      return "md-status-badge--failed";
    default:                return "md-sec-header__pill--neutral";
  }
});
</script>

<template>
  <header class="md-sec-header">
    <div class="md-sec-header__top">
      <button
        type="button"
        class="md-sec-header__back"
        :aria-label="t('marketData.actions.backToMarketData')"
        @click="emit('back')"
      >
        <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <path d="M10 2L4 8l6 6" />
        </svg>
        <span>{{ t("marketData.actions.backToMarketData") }}</span>
      </button>
    </div>

    <div class="md-sec-header__body">
      <div class="md-sec-header__identity">
        <div class="md-sec-header__symbol-row">
          <h1 class="md-sec-header__symbol">{{ detail.displaySymbol || detail.imsSymbol }}</h1>
          <span class="badge badge-neutral md-asset-pill">{{ detail.assetType }}</span>
          <span
            class="md-status-badge"
            :class="freshnessClass(detail.dataQuality.freshness)"
            :aria-label="`${t('marketData.labels.freshness')} ${freshnessLabel(detail.dataQuality.freshness)}`"
          >
            <span class="md-status-dot" aria-hidden="true" />
            {{ freshnessLabel(detail.dataQuality.freshness) }}
          </span>
          <span
            class="md-status-badge"
            :class="mappingClass"
            :aria-label="`${t('marketData.labels.mapping')} ${mappingLabel}`"
          >
            <span class="md-status-dot" aria-hidden="true" />
            {{ mappingLabel }}
          </span>
        </div>
        <p class="md-sec-header__name">{{ detail.name || "—" }}</p>
        <ul class="md-sec-header__meta" :aria-label="t('marketData.labels.identityMetadata')">
          <li v-if="detail.imsSymbol">
            <span class="md-sec-header__meta-label">IMS</span>
            <code>{{ detail.imsSymbol }}</code>
          </li>
          <li v-if="detail.exchangeMic">
            <span class="md-sec-header__meta-label">MIC</span>
            <span>{{ detail.exchangeMic }}</span>
          </li>
          <li v-if="detail.currency">
            <span class="md-sec-header__meta-label">CCY</span>
            <span>{{ detail.currency }}</span>
          </li>
          <li v-if="detail.countryCode">
            <span class="md-sec-header__meta-label">{{ t("marketData.labels.country") }}</span>
            <span>{{ detail.countryCode }}</span>
          </li>
          <li v-if="detail.isin">
            <span class="md-sec-header__meta-label">ISIN</span>
            <code>{{ detail.isin }}</code>
          </li>
        </ul>
      </div>

      <div class="md-sec-header__actions">
        <button
          class="btn btn-sm"
          :class="detail.watching ? 'btn-secondary' : 'btn-secondary'"
          type="button"
          @click="emit('toggle-watching')"
        >
          {{ detail.watching ? t("marketData.actions.watching") : t("marketData.actions.addToWatchlist") }}
        </button>
        <button
          class="btn btn-primary btn-sm"
          type="button"
          :disabled="isRunning"
          @click="emit('sync-quote')"
        >
          {{ isRunning ? t("marketData.actions.syncing") : t("marketData.actions.syncQuote") }}
        </button>
        <button
          class="btn btn-secondary btn-sm"
          type="button"
          :disabled="isRunning"
          @click="emit('import-history')"
        >
          {{ t("marketData.actions.importHistory") }}
        </button>
        <button
          class="btn btn-secondary btn-sm"
          type="button"
          @click="emit('open-mappings')"
        >
          {{ t("marketData.actions.configureMappings") }}
        </button>
      </div>
    </div>
  </header>
</template>

<style scoped>
.md-sec-header {
  background: var(--md-surface);
  border: 1px solid var(--md-border);
  border-radius: 0.75rem;
  padding: 1.25rem 1.5rem;
  display: grid;
  gap: 0.75rem;
  box-shadow: var(--md-shadow);
  color: var(--md-text);
}
.md-sec-header__top {
  display: flex;
}
.md-sec-header__back {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border: 0;
  background: transparent;
  color: var(--md-text-muted);
  cursor: pointer;
  padding: 0.25rem 0.4rem 0.25rem 0;
  font-size: 0.85rem;
}
.md-sec-header__back:hover { color: var(--md-text); }
.md-sec-header__body {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1.5rem;
  flex-wrap: wrap;
}
.md-sec-header__identity {
  display: grid;
  gap: 0.4rem;
  min-width: 260px;
}
.md-sec-header__symbol-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: center;
}
.md-sec-header__symbol {
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0;
  color: var(--md-text);
}
.md-sec-header__name {
  margin: 0;
  color: var(--md-text-muted);
  font-size: 1rem;
}
.md-sec-header__meta {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  font-size: 0.82rem;
  color: var(--md-text-muted);
}
.md-sec-header__meta li {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}
.md-sec-header__meta-label {
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: 600;
  font-size: 0.7rem;
  color: var(--md-text-muted);
}
.md-sec-header__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  align-items: center;
}
.md-sec-header__pill--neutral {
  background: var(--md-neutral-bg);
  color: var(--md-neutral-text);
}
</style>
