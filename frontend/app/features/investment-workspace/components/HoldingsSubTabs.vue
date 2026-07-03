<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import type { HoldingsSubTab } from "../types";

const props = defineProps<{
  active: HoldingsSubTab;
  counts?: Partial<Record<HoldingsSubTab, number>>;
}>();

const emit = defineEmits<{
  select: [tab: HoldingsSubTab];
}>();

const { t } = useI18n();

interface Spec {
  key: HoldingsSubTab;
  label: string;
}

const tabs = computed<Spec[]>(() => [
  { key: "overview", label: t("holdings.subTabs.overview", "Positions overview") },
  { key: "positions", label: t("holdings.subTabs.positions", "Positions") },
  { key: "equities", label: t("holdings.subTabs.equities", "Equities") },
  { key: "fixed_income", label: t("holdings.subTabs.fixed_income", "Fixed income") },
  { key: "funds", label: t("holdings.subTabs.funds", "Funds") },
  { key: "futures", label: t("holdings.subTabs.futures", "Futures") },
  { key: "short_notes", label: t("holdings.subTabs.short_notes", "Short notes") },
  { key: "repos", label: t("holdings.subTabs.repos", "Repos") },
  { key: "fx_forwards", label: t("holdings.subTabs.fx_forwards", "FX forwards") },
]);
</script>

<template>
  <nav class="ht-subtabs" role="tablist">
    <button
      v-for="tab in tabs"
      :key="tab.key"
      type="button"
      role="tab"
      :aria-selected="active === tab.key"
      class="ht-subtabs__item"
      :class="{ 'is-active': active === tab.key }"
      @click="emit('select', tab.key)"
    >
      <span>{{ tab.label }}</span>
      <span v-if="props.counts?.[tab.key] !== undefined" class="ht-subtabs__count">
        {{ props.counts?.[tab.key] }}
      </span>
    </button>
  </nav>
</template>

<style scoped>
.ht-subtabs {
  display: flex;
  align-items: center;
  gap: 0;
  padding: 4px;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 6px;
  overflow-x: auto;
  scrollbar-width: thin;
}

.ht-subtabs__item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
  cursor: pointer;
  white-space: nowrap;
}

.ht-subtabs__item:hover {
  background: var(--bg-row-hover);
}

.ht-subtabs__item.is-active {
  background: var(--status-executed-bg);
  border-color: var(--border-focus);
  color: var(--status-executed-text);
}

.ht-subtabs__count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 16px;
  padding: 0 6px;
  background: var(--border-subtle);
  border-radius: 8px;
  font-size: 10px;
  font-weight: 600;
}
</style>
