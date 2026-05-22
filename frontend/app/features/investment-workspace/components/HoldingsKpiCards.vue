<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import { formatBangkokTime, formatSignedPercent } from "../lib/holdingsFormat";
import type { HoldingsKpis } from "../types";

const props = defineProps<{
  kpis: HoldingsKpis;
}>();

const { t } = useI18n();

// Signed percent for the Today NAV tile drives a positive/negative tone
// purely client-side. The numeric magnitude itself stays as the API gave it.
const navDeltaTone = computed(() => {
  const v = props.kpis.today_nav.delta_pct_vs_last_close;
  if (v === 0) return "neutral";
  return v > 0 ? "positive" : "negative";
});

const unitNavDeltaTone = computed(() => {
  const v = props.kpis.today_unit_nav.delta_vs_yesterday;
  if (v === 0) return "neutral";
  return v > 0 ? "positive" : "negative";
});
</script>

<template>
  <section class="ht-kpis">
    <article class="ht-kpis__card ht-kpis__card--accent-blue">
      <header class="ht-kpis__label">
        <span class="ht-kpis__dot ht-kpis__dot--blue" aria-hidden="true" />
        {{ t("holdings.kpis.todayNav", "TODAY NAV") }}
      </header>
      <div class="ht-kpis__value">{{ kpis.today_nav.value }}</div>
      <div class="ht-kpis__meta" :class="`tone-${navDeltaTone}`">
        {{ formatSignedPercent(kpis.today_nav.delta_pct_vs_last_close) }}
        <span class="ht-kpis__meta-muted">{{ t("holdings.kpis.vsLastClose", "vs. last close") }}</span>
      </div>
    </article>

    <article class="ht-kpis__card">
      <header class="ht-kpis__label">{{ t("holdings.kpis.todayUnitNav", "TODAY UNIT NAV") }}</header>
      <div class="ht-kpis__value">{{ kpis.today_unit_nav.value }}</div>
      <div class="ht-kpis__meta" :class="`tone-${unitNavDeltaTone}`">
        {{ formatSignedPercent(kpis.today_unit_nav.delta_vs_yesterday, 3) }}
        <span class="ht-kpis__meta-muted">{{ t("holdings.kpis.vsYesterday", "vs. yesterday") }}</span>
      </div>
    </article>

    <article class="ht-kpis__card">
      <header class="ht-kpis__label">{{ t("holdings.kpis.todayUnitsOut", "TODAY UNITS OUT.") }}</header>
      <div class="ht-kpis__value">{{ kpis.today_units_outstanding.value }}</div>
      <div class="ht-kpis__meta ht-kpis__meta-muted">
        {{ kpis.today_units_outstanding.movement_label }}
      </div>
    </article>

    <article class="ht-kpis__card">
      <header class="ht-kpis__label">{{ t("holdings.kpis.netSubscription", "NET SUBSCRIPTION") }}</header>
      <div class="ht-kpis__value">฿ {{ kpis.net_subscription.value }}</div>
      <div class="ht-kpis__meta ht-kpis__meta-muted">{{ kpis.net_subscription.flow_label }}</div>
    </article>

    <article class="ht-kpis__card">
      <header class="ht-kpis__label">{{ t("holdings.kpis.lastSettled", "LAST SETTLED") }}</header>
      <div class="ht-kpis__value ht-kpis__value--date">{{ kpis.last_settled_date.date }}</div>
      <div class="ht-kpis__meta ht-kpis__meta-muted">{{ kpis.last_settled_date.label }}</div>
    </article>

    <article
      class="ht-kpis__card"
      :class="kpis.stale_warning.is_stale ? 'ht-kpis__card--warning' : 'ht-kpis__card--ok'"
    >
      <header class="ht-kpis__label">
        <span
          class="ht-kpis__dot"
          :class="kpis.stale_warning.is_stale ? 'ht-kpis__dot--warning' : 'ht-kpis__dot--success'"
          aria-hidden="true"
        />
        {{ t("holdings.kpis.staleWarning", "STALE DATA WARNING") }}
      </header>
      <div class="ht-kpis__value ht-kpis__value--sm">
        {{
          kpis.stale_warning.is_stale
            ? kpis.stale_warning.age_label
            : t("holdings.kpis.allFresh", "all fresh")
        }}
      </div>
      <div class="ht-kpis__meta ht-kpis__meta-muted">
        {{ t("holdings.kpis.lastRefresh", "last refresh") }}
        {{ formatBangkokTime(kpis.stale_warning.last_refresh_at) }}
      </div>
    </article>
  </section>
</template>

<style scoped>
.ht-kpis {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: var(--space-2);
}

.ht-kpis__card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 14px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  min-width: 0;
}

.ht-kpis__card--accent-blue {
  border-left: 3px solid var(--color-primary-500, #2563eb);
}

.ht-kpis__card--warning {
  background: var(--alert-warning-bg);
  border-color: var(--alert-warning-border);
}

.ht-kpis__card--ok {
  background: var(--alert-success-bg);
  border-color: var(--alert-success-border);
}

.ht-kpis__label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}

.ht-kpis__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.ht-kpis__dot--blue {
  background: var(--color-primary-500, #2563eb);
}
.ht-kpis__dot--warning {
  background: var(--color-warning-500, #f59e0b);
}
.ht-kpis__dot--success {
  background: var(--color-success-500, #12b76a);
}

.ht-kpis__value {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ht-kpis__value--date,
.ht-kpis__value--sm {
  font-size: 1.05rem;
}

.ht-kpis__meta {
  font-size: 11px;
  color: var(--text-primary);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.ht-kpis__meta-muted {
  color: var(--text-tertiary);
  font-weight: 400;
}

.tone-positive {
  color: var(--state-success);
}
.tone-negative {
  color: var(--state-danger);
}
.tone-neutral {
  color: var(--text-primary);
}

@media (max-width: 1280px) {
  .ht-kpis {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .ht-kpis {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 480px) {
  .ht-kpis {
    grid-template-columns: 1fr;
  }
}
</style>
