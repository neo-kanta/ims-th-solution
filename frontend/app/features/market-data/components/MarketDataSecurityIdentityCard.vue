<script setup lang="ts">
import type { MarketDataSecurityDetail } from "../market-data.types";

const props = defineProps<{ detail: MarketDataSecurityDetail }>();

interface Row {
  label: string;
  value?: string;
  mono?: boolean;
}

const rows = computed<Row[]>(() => [
  { label: "IMS symbol",      value: props.detail.imsSymbol,     mono: true },
  { label: "Display symbol",  value: props.detail.displaySymbol, mono: true },
  { label: "Name",            value: props.detail.name },
  { label: "Asset type",      value: props.detail.assetType },
  { label: "Currency",        value: props.detail.currency },
  { label: "Country",         value: props.detail.countryCode },
  { label: "Exchange MIC",    value: props.detail.exchangeMic },
  { label: "ISIN",            value: props.detail.isin,          mono: true },
  { label: "CUSIP",           value: props.detail.cusip,         mono: true },
  { label: "FIGI",            value: props.detail.figi,          mono: true },
  { label: "Status",          value: props.detail.status },
]);
</script>

<template>
  <article class="card md-sec-card">
    <header class="card-header">
      <span class="card-title">Identity &amp; reference data</span>
    </header>
    <div class="card-body">
      <dl class="md-sec-identity">
        <div v-for="row in rows" :key="row.label" class="md-sec-identity__row">
          <dt class="md-sec-identity__label">{{ row.label }}</dt>
          <dd class="md-sec-identity__value">
            <template v-if="row.value">
              <code v-if="row.mono">{{ row.value }}</code>
              <span v-else>{{ row.value }}</span>
            </template>
            <span v-else class="md-sec-identity__missing">
              <span class="md-status-dot" aria-hidden="true" />
              Missing
            </span>
          </dd>
        </div>
      </dl>
    </div>
  </article>
</template>

<style scoped>
.md-sec-identity {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65rem 1.5rem;
  margin: 0;
}
.md-sec-identity__row {
  display: grid;
  grid-template-columns: minmax(110px, 30%) 1fr;
  gap: 0.5rem;
  align-items: baseline;
}
.md-sec-identity__label {
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--md-text-muted);
  margin: 0;
}
.md-sec-identity__value {
  margin: 0;
  font-size: 0.9rem;
  color: var(--md-text);
}
.md-sec-identity__value code {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.82rem;
}
.md-sec-identity__missing {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  color: var(--md-warning, var(--md-text-muted));
  font-size: 0.78rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
@media (max-width: 768px) {
  .md-sec-identity { grid-template-columns: 1fr; }
}
</style>
