<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";

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
  label: string;
  swatch: string;
}> = [
  { key: "MANDATE", label: "Mandate", swatch: "#0969da" },
  { key: "RATIO", label: "Ratio / exposure", swatch: "#8b5cf6" },
  { key: "RESTRICTION", label: "Restriction lists", swatch: "#d92d20" },
  { key: "REGULATORY", label: "Regulatory", swatch: "#f59e0b" },
  { key: "HOUSE", label: "House rules", swatch: "#12b76a" },
  { key: "CLIENT", label: "Client mandate", swatch: "#0ea5e9" },
  { key: "TEMPORAL", label: "Temporal", swatch: "#7c3aed" },
  { key: "BEHAVIORAL", label: "Behavioural", swatch: "#ec4899" },
  { key: "UNKNOWN", label: "Uncategorised", swatch: "#8c959f" },
];

const rows = computed(() =>
  CATEGORY_DEFS.map((def) => ({
    ...def,
    count: props.byCategory.get(def.key)?.length ?? 0,
  })).filter((r) => r.count > 0),
);
</script>

<template>
  <AppCard :title="t('compliance.dashboard.categories.title')">
    <ul v-if="rows.length > 0" class="cat-list">
      <li v-for="r in rows" :key="r.key" class="cat-item">
        <span
          class="cat-item__swatch"
          aria-hidden="true"
          :style="{ background: r.swatch }"
        />
        <span class="cat-item__label">{{ r.label }}</span>
        <span class="cat-item__count">{{ r.count }}</span>
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
  grid-template-columns: 12px 1fr auto;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--text-primary);
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
</style>
