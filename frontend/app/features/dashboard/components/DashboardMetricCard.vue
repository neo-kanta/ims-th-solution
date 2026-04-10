<script setup lang="ts">
import type { DashboardOverviewMetric } from "../types";

defineProps<{
  metric: DashboardOverviewMetric;
}>();
</script>

<template>
  <article class="dashboard-metric-card" :class="`tone-${metric.tone}`">
    <div class="dashboard-metric-card__content">
      <div class="dashboard-metric-card__copy">
        <div class="dashboard-metric-card__value">{{ metric.value }}</div>
        <div class="dashboard-metric-card__label">{{ metric.label }}</div>

        <div class="dashboard-metric-card__meta">
          <span
            class="dashboard-metric-card__change"
            :class="`is-${metric.changeTone}`"
          >
            {{ metric.changeLabel }}
          </span>
          <span class="dashboard-metric-card__helper">{{ metric.helperText }}</span>
        </div>
      </div>

      <div class="dashboard-metric-card__icon-shell">
        <AppIcon :name="metric.icon" size="sm" />
      </div>
    </div>
  </article>
</template>

<style scoped>
.dashboard-metric-card {
  --metric-accent: var(--action-primary);
  --metric-icon-bg: var(--color-primary-50);
  --metric-icon-color: var(--color-primary-700);

  min-height: 8.25rem;
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-left: 4px solid var(--metric-accent);
  border-radius: var(--radius-xl);
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  transition:
    transform var(--transition-fast),
    box-shadow var(--transition-fast);
}

.dashboard-metric-card:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.dashboard-metric-card.tone-primary {
  --metric-accent: var(--color-primary-700);
  --metric-icon-bg: var(--color-primary-50);
  --metric-icon-color: var(--color-primary-700);
}

.dashboard-metric-card.tone-info {
  --metric-accent: var(--color-info-600);
  --metric-icon-bg: var(--color-info-50);
  --metric-icon-color: var(--color-info-700);
}

.dashboard-metric-card.tone-danger {
  --metric-accent: var(--color-danger-500);
  --metric-icon-bg: var(--color-danger-50);
  --metric-icon-color: var(--color-danger-700);
}

.dashboard-metric-card.tone-success {
  --metric-accent: var(--color-primary-700);
  --metric-icon-bg: var(--color-primary-50);
  --metric-icon-color: var(--color-primary-700);
}

.dashboard-metric-card__content {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.dashboard-metric-card__copy {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.dashboard-metric-card__value {
  color: var(--text-primary);
  font-size: clamp(1.625rem, 1.5vw, 1.875rem);
  font-weight: var(--font-weight-bold);
  letter-spacing: -0.03em;
  line-height: 1.05;
}

.dashboard-metric-card__label {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.dashboard-metric-card__meta {
  display: grid;
  gap: 2px;
  margin-top: var(--space-2);
}

.dashboard-metric-card__change {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.dashboard-metric-card__change.is-success {
  color: var(--state-success);
}

.dashboard-metric-card__change.is-warning {
  color: var(--state-warning);
}

.dashboard-metric-card__change.is-danger {
  color: var(--state-danger);
}

.dashboard-metric-card__change.is-neutral {
  color: var(--text-secondary);
}

.dashboard-metric-card__helper {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.dashboard-metric-card__icon-shell {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.75rem;
  height: 2.75rem;
  border-radius: var(--radius-lg);
  background: var(--metric-icon-bg);
  color: var(--metric-icon-color);
  flex-shrink: 0;
}

@media (max-width: 640px) {
  .dashboard-metric-card {
    min-height: auto;
    padding: var(--space-4);
  }

  .dashboard-metric-card__value {
    font-size: 1.5rem;
  }
}
</style>
