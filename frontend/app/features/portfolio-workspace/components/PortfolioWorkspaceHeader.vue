<script setup lang="ts">
import type { ApiPortfolioV2 } from "../services/portfolioApi";

defineProps<{
  portfolio: ApiPortfolioV2 | null;
}>();
</script>

<template>
  <div v-if="portfolio" class="portfolio-header">
    <div class="portfolio-header__identity">
      <span class="portfolio-header__code">{{ portfolio.code }}</span>
      <h1 class="portfolio-header__name">{{ portfolio.name }}</h1>
    </div>
    <div class="portfolio-header__meta">
      <span class="portfolio-header__badge">{{ portfolio.status }}</span>
      <span class="portfolio-header__currency">
        {{ portfolio.valuation_currency || portfolio.base_currency }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.portfolio-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-3, 12px);
  flex-wrap: wrap;
}

.portfolio-header__identity {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.portfolio-header__code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  font-weight: 700;
  color: var(--text-tertiary, #6e7781);
  padding: 2px 8px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 999px;
}

.portfolio-header__name {
  margin: 0;
  font-size: 1.375rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.portfolio-header__meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.portfolio-header__badge {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 3px 8px;
  border-radius: 999px;
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  color: var(--text-secondary, #57606a);
}

.portfolio-header__currency {
  font-size: 12px;
  color: var(--text-tertiary, #6e7781);
  font-weight: 600;
}
</style>
