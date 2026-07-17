<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";
import type { AppTranslationKey } from "~/shared/i18n/messages";

import type {
  ComplianceRule,
  ComplianceRuleCategory,
} from "../types";

interface Props {
  byCategory: Map<ComplianceRuleCategory | "UNKNOWN", ComplianceRule[]>;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), { loading: false });

const { t } = useI18n();

// Order + colour swatch per backend category. The colour is purely a visual
// affordance; the count is real, derived from /compliance/rules.
const CATEGORY_DEFS: Array<{
  key: ComplianceRuleCategory | "UNKNOWN";
  swatch: string;
}> = [
  { key: "MANDATE", swatch: "#0969da" },
  { key: "RATIO", swatch: "#8b5cf6" },
  { key: "RESTRICTION", swatch: "#d92d20" },
  { key: "REGULATORY", swatch: "#f59e0b" },
  { key: "HOUSE", swatch: "#12b76a" },
  { key: "CLIENT", swatch: "#0ea5e9" },
  { key: "TEMPORAL", swatch: "#7c3aed" },
  { key: "BEHAVIORAL", swatch: "#ec4899" },
  { key: "UNKNOWN", swatch: "#8c959f" },
];

function categoryLabel(key: ComplianceRuleCategory | "UNKNOWN"): string {
  return t(`compliance.dashboard.categories.items.${key}` as AppTranslationKey);
}

const rows = computed(() =>
  CATEGORY_DEFS.map((def) => ({
    ...def,
    label: categoryLabel(def.key),
    count: props.byCategory.get(def.key)?.length ?? 0,
  })).filter((r) => r.count > 0),
);

const total = computed(() => rows.value.reduce((sum, row) => sum + row.count, 0));
</script>

<template>
  <AppCard :title="t('compliance.dashboard.categories.title')">
    <div v-if="loading" class="cat-skeleton" aria-busy="true" aria-live="polite">
      <span v-for="index in 4" :key="index" />
    </div>
    <ul v-else-if="rows.length > 0" class="cat-list">
      <li v-for="r in rows" :key="r.key" class="cat-item">
        <span
          class="cat-item__swatch"
          aria-hidden="true"
          :style="{ background: r.swatch }"
        />
        <span class="cat-item__label">{{ r.label }}</span>
        <span class="cat-item__count">{{ r.count }}</span>
        <span class="cat-item__track" aria-hidden="true">
          <span
            class="cat-item__fill"
            :style="{
              background: r.swatch,
              width: `${Math.max(4, (r.count / total) * 100)}%`,
            }"
          />
        </span>
      </li>
    </ul>
    <p v-else class="cat-empty">
      {{ t("compliance.dashboard.categories.empty") }}
    </p>
  </AppCard>
</template>

<style scoped>
.cat-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-2);
}

.cat-item {
  display: grid;
  grid-template-columns: 12px minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.cat-item__track {
  grid-column: 2 / -1;
  height: 3px;
  overflow: hidden;
  border-radius: var(--radius-pill);
  background: var(--border-subtle);
}

.cat-item__fill {
  display: block;
  height: 100%;
  border-radius: inherit;
}

.cat-item__swatch {
  width: 12px;
  height: 12px;
  border-radius: 3px;
}

.cat-item__label {
  color: var(--text-primary);
}

.cat-item__count {
  color: var(--text-secondary);
  font-variant-numeric: tabular-nums;
  font-weight: var(--font-weight-semibold);
}

.cat-empty {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
}

.cat-skeleton {
  display: grid;
  gap: var(--space-3);
}

.cat-skeleton span {
  height: 18px;
  border-radius: var(--radius-sm);
  background: var(--bg-card-muted);
  animation: cat-pulse 1.4s ease-in-out infinite;
}

@keyframes cat-pulse {
  50% {
    opacity: 0.45;
  }
}

@media (prefers-reduced-motion: reduce) {
  .cat-skeleton span {
    animation: none;
  }
}
</style>
