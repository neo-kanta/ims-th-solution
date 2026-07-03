<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppIcon from "~/shared/ui/AppIcon.vue";

interface Props {
  totalRules: number | null;
  activeRules: number | null;
  inactiveRules: number | null;
  recentBreaches: number | null;
  loading?: boolean;
}

withDefaults(defineProps<Props>(), {
  loading: false,
});

const { t } = useI18n();

function format(value: number | null): string {
  if (value === null || value === undefined) return "—";
  return value.toLocaleString("en-US");
}
</script>

<template>
  <section class="kpi-grid" aria-label="Compliance KPIs">
    <article class="kpi-card" :aria-busy="loading">
      <header class="kpi-card__header">
        <span class="kpi-card__icon kpi-card__icon--neutral">
          <AppIcon name="shield" size="sm" />
        </span>
        <span class="kpi-card__label">{{ t("compliance.dashboard.kpi.totalRules") }}</span>
      </header>
      <div class="kpi-card__value">{{ loading ? "…" : format(totalRules) }}</div>
    </article>

    <article class="kpi-card" :aria-busy="loading">
      <header class="kpi-card__header">
        <span class="kpi-card__icon kpi-card__icon--success">
          <AppIcon name="check" size="sm" />
        </span>
        <span class="kpi-card__label">{{ t("compliance.dashboard.kpi.activeRules") }}</span>
      </header>
      <div class="kpi-card__value">{{ loading ? "…" : format(activeRules) }}</div>
    </article>

    <article class="kpi-card" :aria-busy="loading">
      <header class="kpi-card__header">
        <span class="kpi-card__icon kpi-card__icon--neutral">
          <AppIcon name="pending" size="sm" />
        </span>
        <span class="kpi-card__label">{{ t("compliance.dashboard.kpi.inactiveRules") }}</span>
      </header>
      <div class="kpi-card__value">{{ loading ? "…" : format(inactiveRules) }}</div>
    </article>

    <article class="kpi-card" :aria-busy="loading">
      <header class="kpi-card__header">
        <span class="kpi-card__icon kpi-card__icon--danger">
          <AppIcon name="warning" size="sm" />
        </span>
        <span class="kpi-card__label">{{ t("compliance.dashboard.kpi.recentBreaches") }}</span>
      </header>
      <div class="kpi-card__value">{{ loading ? "…" : format(recentBreaches) }}</div>
    </article>
  </section>
</template>

<style scoped>
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: var(--space-5);
}

.kpi-card {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
}

.kpi-card__header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.kpi-card__icon {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
}

.kpi-card__icon--success {
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
}

.kpi-card__icon--danger {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
}

.kpi-card__icon--neutral {
  background: var(--bg-card-muted);
  color: var(--text-secondary);
}

.kpi-card__label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.kpi-card__value {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
}
</style>
