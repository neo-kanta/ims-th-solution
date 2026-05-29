<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";

import type { MyFundsFilter, MyFundsSort } from "../types";

interface Counts {
  all: number;
  managed: number;
  breached: number;
  locked: number;
  stale: number;
}

interface Props {
  filter: MyFundsFilter;
  sort: MyFundsSort;
  search: string;
  counts: Counts;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "update:filter", value: MyFundsFilter): void;
  (e: "update:sort", value: MyFundsSort): void;
  (e: "update:search", value: string): void;
}>();

const { t } = useI18n();

const filters = computed<Array<{ key: MyFundsFilter; label: string; count: number }>>(() => [
  { key: "all", label: t("myFunds.filters.all", "All funds"), count: props.counts.all },
  { key: "managed", label: t("myFunds.filters.managed", "Managed"), count: props.counts.managed },
  { key: "breached", label: t("myFunds.filters.breached", "Breached"), count: props.counts.breached },
  { key: "locked", label: t("myFunds.filters.locked", "Locked"), count: props.counts.locked },
  { key: "stale", label: t("myFunds.filters.stale", "Stale NAV"), count: props.counts.stale },
]);

const sortOptions = computed<Array<{ key: MyFundsSort; label: string }>>(() => [
  { key: "aum", label: t("myFunds.sort.aum", "AUM") },
  { key: "breach", label: t("myFunds.sort.breach", "Breach severity") },
  { key: "updated", label: t("myFunds.sort.updated", "Last updated") },
  { key: "name", label: t("myFunds.sort.name", "Name") },
]);

function onSearch(event: Event) {
  const target = event.target as HTMLInputElement;
  emit("update:search", target.value);
}
</script>

<template>
  <div class="toolbar">
    <div class="toolbar__filters" role="tablist">
      <button
        v-for="f in filters"
        :key="f.key"
        type="button"
        :class="['toolbar__chip', { 'is-active': filter === f.key }]"
        :aria-pressed="filter === f.key"
        @click="emit('update:filter', f.key)"
      >
        <span>{{ f.label }}</span>
        <span class="toolbar__chip-count">{{ f.count }}</span>
      </button>
    </div>
    <div class="toolbar__right">
      <label class="toolbar__search">
        <span class="toolbar__search-icon" aria-hidden="true">⌕</span>
        <input
          type="search"
          :value="search"
          :placeholder="t('myFunds.search.placeholder', 'Filter by code, name, manager…')"
          @input="onSearch"
        />
      </label>
      <label class="toolbar__sort">
        <span class="toolbar__sort-label">{{ t("myFunds.sort.label", "Sort") }}</span>
        <select
          :value="sort"
          @change="emit('update:sort', (($event.target as HTMLSelectElement).value as MyFundsSort))"
        >
          <option v-for="opt in sortOptions" :key="opt.key" :value="opt.key">
            {{ opt.label }}
          </option>
        </select>
      </label>
    </div>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  flex-wrap: wrap;
  padding: var(--space-2) 0;
}

.toolbar__filters {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.toolbar__chip {
  font-family: inherit;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card, #ffffff);
  color: var(--text-secondary, #57606a);
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.toolbar__chip:hover {
  background: var(--bg-card-muted, #f6f8fa);
}

.toolbar__chip.is-active {
  background: var(--bg-card-strong, #0d1117);
  color: #ffffff;
  border-color: var(--bg-card-strong, #0d1117);
}

.toolbar__chip-count {
  font-size: 11px;
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--text-tertiary, #6e7781);
  padding: 1px 6px;
  border-radius: 999px;
  font-weight: 500;
}

.toolbar__chip.is-active .toolbar__chip-count {
  background: rgba(255, 255, 255, 0.18);
  color: #ffffff;
}

.toolbar__right {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.toolbar__search {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 5px);
  padding: 4px 10px;
  background: var(--bg-card, #ffffff);
  min-width: 240px;
}

.toolbar__search input {
  border: none;
  background: transparent;
  outline: none;
  font-family: inherit;
  font-size: 12px;
  color: var(--text-primary, #1f2328);
  flex: 1;
  min-width: 0;
}

.toolbar__search-icon {
  color: var(--text-tertiary, #6e7781);
  font-size: 13px;
}

.toolbar__sort {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 5px);
  padding: 2px 10px;
  background: var(--bg-card, #ffffff);
  font-size: 12px;
}

.toolbar__sort-label {
  color: var(--text-tertiary, #6e7781);
}

.toolbar__sort select {
  border: none;
  background: transparent;
  outline: none;
  font-family: inherit;
  font-size: 12px;
  color: var(--text-primary, #1f2328);
  cursor: pointer;
}
</style>
