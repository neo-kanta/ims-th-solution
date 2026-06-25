<script setup lang="ts">
import type { components } from "~/api/ims-api";

type ApiPortfolio = components["schemas"]["PortfolioResponse"];

const props = defineProps<{
  portfolios: ApiPortfolio[];
  selectedId?: string;
  loading: boolean;
}>();

const emit = defineEmits<{
  (e: "select", portfolio: ApiPortfolio): void;
}>();
</script>

<template>
  <div class="panel-header">Portfolio</div>
  <div v-if="loading" class="panel-empty">Loading…</div>
  <div v-else class="portfolio-list">
    <div
      v-for="p in portfolios"
      :key="p.id"
      class="portfolio-item"
      :class="{ 'is-selected': selectedId === p.id }"
      @click="emit('select', p)"
    >
      <div class="portfolio-item__header">
        <span class="portfolio-item__code">{{ p.code ?? "—" }}</span>
        <span class="portfolio-item__status" :data-status="p.status">
          {{ p.status ?? "" }}
        </span>
      </div>
      <div class="portfolio-item__name">{{ p.name ?? "—" }}</div>
      <div class="portfolio-item__ccy">{{ p.base_currency ?? "" }}</div>
    </div>
    <div v-if="portfolios.length === 0" class="panel-empty">
      No portfolios found.
    </div>
  </div>
</template>

<style scoped>
.panel-header {
  background: var(--bg-card-hover);
  border-bottom: 1px solid var(--border-subtle);
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.portfolio-list {
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow-y: auto;
  flex: 1;
}

.portfolio-item {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md, 6px);
  padding: 10px;
  cursor: pointer;
  background: var(--bg-card);
  transition: border-color 0.15s, background-color 0.15s;
}

.portfolio-item:hover {
  border-color: var(--border-focus);
  background: var(--bg-card-hover);
}

.portfolio-item.is-selected {
  border-color: var(--state-info);
  background: var(--bg-card-hover);
  box-shadow: 0 0 0 1px var(--state-info);
}

.portfolio-item__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 3px;
}

.portfolio-item__code {
  font-weight: 700;
  color: var(--state-info);
  font-size: 12px;
  font-family: monospace;
}

.portfolio-item__status {
  font-size: 9px;
  font-weight: 700;
  text-transform: uppercase;
  padding: 2px 5px;
  border-radius: 8px;
  background: var(--bg-card-muted);
  color: var(--text-secondary);
}

.portfolio-item__status[data-status="ACTIVE"] {
  background: rgba(26, 127, 55, 0.15);
  color: var(--state-success, #1a7f37);
}

.portfolio-item__name {
  font-size: 12px;
  color: var(--text-primary);
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.portfolio-item__ccy {
  font-size: 11px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.panel-empty {
  padding: 16px 12px;
  font-size: 13px;
  color: var(--text-secondary);
  text-align: center;
}
</style>
