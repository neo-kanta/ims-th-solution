<script setup lang="ts">
import type { MarketQuoteView } from "../market-data.types";
import { useI18n } from "~/composables/useI18n";

const props = defineProps<{
  quote?: MarketQuoteView;
  loading?: boolean;
}>();
const { t } = useI18n();

function currencyPrefix(ccy?: string): string {
  if (!ccy) return "";
  switch (ccy.toUpperCase()) {
    case "THB": return "฿";
    case "USD": return "$";
    case "EUR": return "€";
    case "JPY": return "¥";
    case "GBP": return "£";
    default:    return "";
  }
}

function parseDecimal(value?: string | null): number | undefined {
  if (value == null || value === "") return undefined;
  const n = Number(value);
  return Number.isFinite(n) ? n : undefined;
}

const priceText = computed(() => {
  const n = parseDecimal(props.quote?.lastPrice);
  if (n == null) return "—";
  return `${currencyPrefix(props.quote?.currency)}${n.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 4 })}`;
});

const changeAmount = computed(() => parseDecimal(props.quote?.changeAmount));
const changePercent = computed(() => parseDecimal(props.quote?.changePercent));

const changeClass = computed(() => {
  const c = changeAmount.value ?? changePercent.value;
  if (c == null) return "md-text-muted";
  return c >= 0 ? "md-text-success" : "md-text-danger";
});

const changeText = computed(() => {
  const amt = changeAmount.value;
  const pct = changePercent.value;
  if (amt == null && pct == null) return "—";
  const parts: string[] = [];
  if (amt != null) {
    const sign = amt >= 0 ? "+" : "";
    parts.push(`${sign}${amt.toFixed(2)}`);
  }
  if (pct != null) {
    const sign = pct >= 0 ? "+" : "";
    parts.push(`${sign}${pct.toFixed(2)}%`);
  }
  return parts.join(" · ");
});

function formatVolume(v?: number): string {
  if (v == null || !Number.isFinite(v)) return "—";
  if (Math.abs(v) >= 1_000_000) return `${(v / 1_000_000).toFixed(2)}M`;
  if (Math.abs(v) >= 1_000)     return `${(v / 1_000).toFixed(2)}K`;
  return String(v);
}

function formatTimestamp(iso?: string): string {
  if (!iso) return "—";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toISOString().slice(0, 16).replace("T", " ") + " UTC";
}

const isEmpty = computed(() => !props.quote);
</script>

<template>
  <article class="card md-sec-card md-sec-quote">
    <header class="card-header">
      <div>
        <span class="card-title">{{ t("marketData.labels.latestQuote") }}</span>
        <div class="card-subtitle">
          <template v-if="quote?.provider">via {{ quote.provider }}</template>
          <template v-else>{{ t("marketData.messages.noQuoteLoaded") }}</template>
        </div>
      </div>
    </header>
    <div class="card-body md-sec-quote__body">
      <div class="md-sec-quote__primary">
        <div class="md-sec-quote__price md-num">{{ priceText }}</div>
        <div class="md-sec-quote__change md-num" :class="changeClass">{{ changeText }}</div>
      </div>

      <dl class="md-sec-quote__grid">
        <div>
          <dt>{{ t("marketData.labels.openPrice") }}</dt>
          <dd class="md-num">{{ quote?.open ?? "—" }}</dd>
        </div>
        <div>
          <dt>{{ t("marketData.labels.high") }}</dt>
          <dd class="md-num">{{ quote?.high ?? "—" }}</dd>
        </div>
        <div>
          <dt>{{ t("marketData.labels.low") }}</dt>
          <dd class="md-num">{{ quote?.low ?? "—" }}</dd>
        </div>
        <div>
          <dt>{{ t("marketData.labels.prevClose") }}</dt>
          <dd class="md-num">{{ quote?.previousClose ?? "—" }}</dd>
        </div>
        <div>
          <dt>{{ t("marketData.labels.volume") }}</dt>
          <dd class="md-num">{{ formatVolume(quote?.volume) }}</dd>
        </div>
        <div>
          <dt>{{ t("marketData.labels.asOf") }}</dt>
          <dd>{{ formatTimestamp(quote?.asOf) }}</dd>
        </div>
      </dl>

      <p
        v-if="isEmpty && !loading"
        class="md-sec-quote__hint"
      >
        {{ t("marketData.messages.pressSyncQuote") }}
      </p>
      <p
        v-else-if="loading"
        class="md-sec-quote__hint"
      >
        {{ t("marketData.messages.loadingLatestQuote") }}
      </p>
      <p
        v-else-if="quote?.stale"
        class="md-sec-quote__hint md-sec-quote__hint--warn"
      >
        {{ t("marketData.messages.quoteStale", { reason: quote.staleReason || t("marketData.messages.noFurtherDetail") }) }}
      </p>
    </div>
  </article>
</template>

<style scoped>
.md-sec-quote__body {
  display: grid;
  gap: 1rem;
}
.md-sec-quote__primary {
  display: flex;
  align-items: baseline;
  gap: 1rem;
  flex-wrap: wrap;
}
.md-sec-quote__price {
  font-size: 1.85rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}
.md-sec-quote__change {
  font-size: 1rem;
  font-weight: 600;
}
.md-sec-quote__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.7rem 1rem;
  margin: 0;
}
.md-sec-quote__grid div {
  display: grid;
  gap: 0.15rem;
}
.md-sec-quote__grid dt {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--md-text-muted);
}
.md-sec-quote__grid dd {
  margin: 0;
  font-size: 0.9rem;
  color: var(--md-text);
}
.md-sec-quote__hint {
  margin: 0;
  font-size: 0.85rem;
  color: var(--md-text-muted);
}
.md-sec-quote__hint--warn { color: var(--md-warning); }
.md-num { font-variant-numeric: tabular-nums; }
@media (max-width: 600px) {
  .md-sec-quote__grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
</style>
