<script setup lang="ts">
import type {
  SecurityDataQuality,
  DetailFreshnessStatus,
  DetailMappingStatus,
} from "../market-data.types";

defineProps<{ quality: SecurityDataQuality }>();

function freshnessLabel(s: DetailFreshnessStatus): string {
  switch (s) {
    case "FRESH":        return "Fresh";
    case "STALE":        return "Stale";
    case "EXPIRED":      return "Expired";
    case "NOT_IMPORTED": return "Not imported";
    case "FAILED":       return "Failed";
    case "RATE_LIMITED": return "Rate limited";
    default:             return s;
  }
}
function freshnessClass(s: DetailFreshnessStatus): string {
  switch (s) {
    case "FRESH":        return "md-status-badge--fresh";
    case "STALE":        return "md-status-badge--stale";
    case "EXPIRED":      return "md-status-badge--failed";
    case "NOT_IMPORTED": return "md-sec-quality__pill--neutral";
    case "FAILED":       return "md-status-badge--failed";
    case "RATE_LIMITED": return "md-status-badge--rate-limited";
    default:             return "md-sec-quality__pill--neutral";
  }
}
function mappingLabel(s: DetailMappingStatus): string {
  switch (s) {
    case "MAPPED":          return "Mapped";
    case "UNMAPPED":        return "Unmapped";
    case "REVIEW_REQUIRED": return "Review required";
    case "CONFLICTED":      return "Conflicted";
    default:                return s;
  }
}
function mappingClass(s: DetailMappingStatus): string {
  switch (s) {
    case "MAPPED":          return "md-status-badge--fresh";
    case "UNMAPPED":        return "md-status-badge--failed";
    case "REVIEW_REQUIRED": return "md-status-badge--stale";
    case "CONFLICTED":      return "md-status-badge--failed";
    default:                return "md-sec-quality__pill--neutral";
  }
}
function formatTimestamp(iso?: string): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toISOString().slice(0, 16).replace("T", " ") + " UTC";
}
</script>

<template>
  <article class="card md-sec-card">
    <header class="card-header">
      <span class="card-title">Data quality</span>
    </header>
    <div class="card-body md-sec-quality">
      <div class="md-sec-quality__row">
        <span class="md-sec-quality__label">Freshness</span>
        <span
          class="md-status-badge"
          :class="freshnessClass(quality.freshness)"
          :aria-label="`Freshness ${freshnessLabel(quality.freshness)}`"
        >
          <span class="md-status-dot" aria-hidden="true" />
          {{ freshnessLabel(quality.freshness) }}
        </span>
      </div>
      <div class="md-sec-quality__row">
        <span class="md-sec-quality__label">Mapping</span>
        <span
          class="md-status-badge"
          :class="mappingClass(quality.mapping)"
          :aria-label="`Mapping ${mappingLabel(quality.mapping)}`"
        >
          <span class="md-status-dot" aria-hidden="true" />
          {{ mappingLabel(quality.mapping) }}
        </span>
      </div>
      <div class="md-sec-quality__row">
        <span class="md-sec-quality__label">Provider status</span>
        <span class="md-sec-quality__value">{{ quality.providerStatus }}</span>
      </div>
      <div class="md-sec-quality__row">
        <span class="md-sec-quality__label">Last sync</span>
        <span class="md-sec-quality__value">{{ formatTimestamp(quality.lastSyncedAt) }}</span>
      </div>
      <div v-if="quality.errorCode || quality.errorMessage" class="md-sec-quality__row md-sec-quality__row--alert">
        <span class="md-sec-quality__label">Last error</span>
        <span class="md-sec-quality__value">
          <code v-if="quality.errorCode">{{ quality.errorCode }}</code>
          <template v-if="quality.errorCode && quality.errorMessage"> · </template>
          <span v-if="quality.errorMessage">{{ quality.errorMessage }}</span>
        </span>
      </div>
      <div v-if="quality.rateLimited" class="md-sec-quality__rate">
        Provider is rate limited. Wait a moment before retrying.
      </div>
    </div>
  </article>
</template>

<style scoped>
.md-sec-quality {
  display: grid;
  gap: 0.6rem;
}
.md-sec-quality__row {
  display: grid;
  grid-template-columns: minmax(120px, 30%) 1fr;
  gap: 0.5rem;
  align-items: center;
}
.md-sec-quality__row--alert .md-sec-quality__value {
  color: var(--md-warning);
}
.md-sec-quality__label {
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--md-text-muted);
}
.md-sec-quality__value {
  font-size: 0.9rem;
  color: var(--md-text);
}
.md-sec-quality__value code {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.78rem;
  padding: 0.05rem 0.4rem;
  border-radius: 0.25rem;
  background: var(--md-surface-muted);
}
.md-sec-quality__rate {
  margin-top: 0.4rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.85rem;
  border-radius: 0.4rem;
  background: var(--md-warning-bg);
  color: var(--md-warning-text);
}
.md-sec-quality__pill--neutral {
  background: var(--md-neutral-bg);
  color: var(--md-neutral-text);
}
</style>
