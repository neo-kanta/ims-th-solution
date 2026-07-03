<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import { clampPct, formatPercent } from "../lib/holdingsFormat";
import type { RatioStatus, SpecialRatio } from "../types";

defineProps<{
  ratios: SpecialRatio[];
}>();

const { t } = useI18n();

// OK / WARNING / BLOCKER mapping. The backend ultimately owns this decision;
// the frontend only renders the status it's given. Mapping kept here only
// for the bar tint, not for re-classification.
function toneClass(status: RatioStatus): string {
  switch (status) {
    case "BLOCKER": return "is-blocker";
    case "WARNING": return "is-warning";
    default: return "is-ok";
  }
}
</script>

<template>
  <ul class="ht-ratios">
    <li
      v-for="r in ratios"
      :key="r.code"
      class="ht-ratios__row"
      :class="toneClass(r.status)"
    >
      <div class="ht-ratios__head">
        <span class="ht-ratios__label">{{ t(r.label_key as any, r.label_fallback) }}</span>
        <span class="ht-ratios__ceiling">{{ r.ceiling_label }}</span>
      </div>
      <div class="ht-ratios__bar-wrap" aria-hidden="true">
        <div
          class="ht-ratios__bar"
          :style="{ width: `${clampPct(((r.value) / (r.policy_ceiling ?? 100)) * 100)}%` }"
        />
      </div>
      <div class="ht-ratios__meta">
        <span class="ht-ratios__value">{{ formatPercent(r.value) }}</span>
        <span class="ht-ratios__status">
          {{ t(`holdings.ratios.status.${r.status}` as any, r.status) }}
        </span>
      </div>
      <p v-if="r.caption" class="ht-ratios__caption">{{ r.caption }}</p>
    </li>
  </ul>
</template>

<style scoped>
.ht-ratios {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-3);
}

.ht-ratios__row {
  display: grid;
  gap: 6px;
  padding: 10px 12px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
}

.ht-ratios__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.ht-ratios__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.ht-ratios__ceiling {
  font-size: 11px;
  color: var(--text-tertiary);
  font-variant-numeric: tabular-nums;
}

.ht-ratios__bar-wrap {
  position: relative;
  height: 6px;
  background: var(--border-subtle);
  border-radius: 3px;
  overflow: hidden;
}

.ht-ratios__bar {
  height: 100%;
  background: var(--color-primary-500, #2563eb);
  border-radius: 3px;
  transition: width 0.25s ease;
}

.is-ok .ht-ratios__bar {
  background: var(--color-success-500, #12b76a);
}
.is-warning .ht-ratios__bar {
  background: var(--color-warning-500, #f59e0b);
}
.is-blocker .ht-ratios__bar {
  background: var(--color-danger-500, #f04438);
}

.ht-ratios__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.ht-ratios__value {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
}

.ht-ratios__status {
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  padding: 2px 8px;
  border-radius: 10px;
}

.is-ok .ht-ratios__status {
  background: var(--status-approved-bg);
  color: var(--status-approved-text);
}
.is-warning .ht-ratios__status {
  background: var(--status-pending-bg);
  color: var(--status-pending-text);
}
.is-blocker .ht-ratios__status {
  background: var(--status-rejected-bg);
  color: var(--status-rejected-text);
}

.ht-ratios__caption {
  margin: 0;
  font-size: 11px;
  color: var(--text-tertiary);
}
</style>
