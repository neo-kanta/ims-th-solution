<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import type { MyPortfoliosFilter, MyPortfoliosSort } from "../types";

interface Counts {
  all: number;
  managed: number;
  breached: number;
  locked: number;
  stale: number;
}

interface Props {
  filter: MyPortfoliosFilter;
  sort: MyPortfoliosSort;
  search: string;
  counts: Counts;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "update:filter", value: MyPortfoliosFilter): void;
  (e: "update:sort", value: MyPortfoliosSort): void;
  (e: "update:search", value: string): void;
}>();

const { t } = useI18n();

const filters = computed<Array<{ key: MyPortfoliosFilter; label: string; count: number }>>(() => [
  { key: "all", label: t("portfolio.filters.all", "All portfolios"), count: props.counts.all },
  { key: "managed", label: t("portfolio.filters.managed", "Managed"), count: props.counts.managed },
  { key: "breached", label: t("portfolio.filters.breached", "Breached"), count: props.counts.breached },
  { key: "locked", label: t("portfolio.filters.locked", "Closed"), count: props.counts.locked },
  { key: "stale", label: t("portfolio.filters.stale", "Stale NAV"), count: props.counts.stale },
]);

const sortOptions = computed<Array<{ key: MyPortfoliosSort; label: string }>>(() => [
  { key: "aum", label: t("portfolio.sort.aum", "AUM") },
  { key: "breach", label: t("portfolio.sort.breach", "Breach severity") },
  { key: "updated", label: t("portfolio.sort.updated", "Last updated") },
  { key: "name", label: t("portfolio.sort.name", "Name") },
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
        <span v-if="f.key === 'all'" class="toolbar__chip-icon-emoji">📊</span>
        <svg v-else-if="f.key === 'managed'" class="toolbar__chip-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
          <circle cx="12" cy="7" r="4"></circle>
        </svg>
        <svg v-else-if="f.key === 'breached'" class="toolbar__chip-icon color-breached" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        <svg v-else-if="f.key === 'locked'" class="toolbar__chip-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
          <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
        </svg>
        <svg v-else-if="f.key === 'stale'" class="toolbar__chip-icon color-stale" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <circle cx="12" cy="12" r="10"></circle>
          <polyline points="12 6 12 12 16 14"></polyline>
        </svg>
        <span>{{ f.label }}</span>
        <span class="toolbar__chip-count">{{ f.count }}</span>
      </button>
    </div>
    <div class="toolbar__right">
      <label class="toolbar__search">
        <span class="toolbar__search-icon" aria-hidden="true">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
        </span>
        <input
          type="search"
          :value="search"
          :placeholder="t('portfolio.search.placeholder', 'Filter by code, name, manager…')"
          @input="onSearch"
        />
      </label>
      <label class="toolbar__sort">
        <span class="toolbar__sort-label">{{ t("portfolio.sort.label", "Sort") }}</span>
        <select
          :value="sort"
          @change="emit('update:sort', (($event.target as HTMLSelectElement).value as MyPortfoliosSort))"
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
  gap: var(--space-3, 12px);
  flex-wrap: wrap;
  padding: var(--space-2, 8px) 0;
}

.toolbar__filters {
  display: flex;
  gap: 8px;
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
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.12s ease;
}

.toolbar__chip:hover {
  background: var(--bg-card-muted, #f6f8fa);
  border-color: #afb8c1;
}

.toolbar__chip.is-active {
  background: var(--text-primary, #0f172a);
  color: var(--bg-card, #ffffff);
  border-color: var(--text-primary, #0f172a);
}

.toolbar__chip-icon-emoji {
  font-size: 13px;
  margin-right: -1px;
}

.toolbar__chip-icon {
  flex-shrink: 0;
  color: var(--text-secondary, #57606a);
}

.toolbar__chip.is-active .toolbar__chip-icon {
  color: var(--bg-card, #ffffff);
}

.color-breached {
  color: var(--state-danger, #cf222e);
}
.toolbar__chip.is-active .color-breached {
  color: #ff858d;
}

.color-stale {
  color: var(--state-warning, #f08800);
}
.toolbar__chip.is-active .color-stale {
  color: #ffc470;
}

.toolbar__chip-count {
  font-size: 11px;
  background: var(--bg-card-muted, #f1f5f9);
  color: var(--text-secondary, #64748b);
  padding: 1px 7px;
  border-radius: 999px;
  font-weight: 600;
  margin-left: 2px;
}

.toolbar__chip.is-active .toolbar__chip-count {
  background: rgba(255, 255, 255, 0.2);
  color: var(--bg-card, #ffffff);
}

.toolbar__right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar__search {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 6px;
  padding: 4px 12px;
  background: var(--bg-card, #ffffff);
  min-width: 260px;
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
  color: var(--text-secondary, #8c959f);
  display: flex;
  align-items: center;
  justify-content: center;
}

.toolbar__sort {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 6px;
  padding: 4px 12px;
  background: var(--bg-card, #ffffff);
  font-size: 12px;
  font-weight: 500;
}

.toolbar__sort-label {
  color: var(--text-secondary, #6e7781);
}

.toolbar__sort select {
  border: none;
  background: transparent;
  outline: none;
  font-family: inherit;
  font-size: 12px;
  color: var(--text-primary, #1f2328);
  cursor: pointer;
  font-weight: 600;
  padding-right: 4px;
}
</style>
