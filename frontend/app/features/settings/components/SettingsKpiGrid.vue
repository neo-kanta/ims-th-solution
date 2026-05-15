<script setup lang="ts">
import type { SettingsKpiMetric } from "../ui.types";

defineProps<{
  metrics: SettingsKpiMetric[];
  loading?: boolean;
}>();
</script>

<template>
  <section class="settings-kpis" :aria-busy="loading">
    <article
      v-for="metric in metrics"
      :key="metric.id"
      class="settings-kpi"
      :class="`settings-kpi--${metric.tone}`"
    >
      <div class="settings-kpi__copy">
        <div class="settings-kpi__label">{{ metric.label }}</div>
        <div class="settings-kpi__value">
          <span v-if="loading" class="settings-kpi__skeleton" />
          <span v-else>{{ metric.value }}</span>
        </div>
        <div class="settings-kpi__helper">{{ metric.helper }}</div>
      </div>
      <div class="settings-kpi__icon" aria-hidden="true">
        <AppIcon :name="metric.icon" size="sm" />
      </div>
    </article>
  </section>
</template>

<style scoped>
.settings-kpis {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
}

.settings-kpi {
  --settings-kpi-accent: var(--action-primary);
  --settings-kpi-icon-bg: var(--status-executed-bg);
  --settings-kpi-icon-text: var(--status-executed-text);

  display: flex;
  justify-content: space-between;
  gap: var(--space-4);
  min-height: 8rem;
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-left: 4px solid var(--settings-kpi-accent);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  box-shadow: var(--shadow-xs);
}

.settings-kpi--success {
  --settings-kpi-accent: var(--state-success);
  --settings-kpi-icon-bg: var(--status-approved-bg);
  --settings-kpi-icon-text: var(--status-approved-text);
}

.settings-kpi--warning {
  --settings-kpi-accent: var(--state-warning);
  --settings-kpi-icon-bg: var(--status-pending-bg);
  --settings-kpi-icon-text: var(--status-pending-text);
}

.settings-kpi--danger {
  --settings-kpi-accent: var(--state-danger);
  --settings-kpi-icon-bg: var(--status-rejected-bg);
  --settings-kpi-icon-text: var(--status-rejected-text);
}

.settings-kpi--neutral {
  --settings-kpi-accent: var(--text-tertiary);
  --settings-kpi-icon-bg: var(--status-draft-bg);
  --settings-kpi-icon-text: var(--status-draft-text);
}

.settings-kpi__copy {
  display: grid;
  align-content: start;
  gap: var(--space-2);
  min-width: 0;
}

.settings-kpi__label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.settings-kpi__value {
  min-height: 2.25rem;
  color: var(--text-primary);
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  line-height: 1;
}

.settings-kpi__helper {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.settings-kpi__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.75rem;
  height: 2.75rem;
  flex: 0 0 auto;
  border-radius: var(--radius-md);
  background: var(--settings-kpi-icon-bg);
  color: var(--settings-kpi-icon-text);
}

.settings-kpi__skeleton {
  display: block;
  width: 4rem;
  height: 1.75rem;
  border-radius: var(--radius-sm);
  background: var(--bg-card-muted);
}

@media (max-width: 1180px) {
  .settings-kpis {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .settings-kpis {
    grid-template-columns: 1fr;
  }

  .settings-kpi {
    min-height: auto;
  }
}
</style>
