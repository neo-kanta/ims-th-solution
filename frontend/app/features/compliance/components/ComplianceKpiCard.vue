<script setup lang="ts">
import { computed } from "vue";

import AppIcon from "~/shared/ui/AppIcon.vue";

interface Props {
  label: string;
  /** null = backend call hasn't returned / errored; renders "—". */
  value: number | null;
  subtitle?: string;
  /** Optional info tooltip explaining missing-data or honest-state caveats. */
  hint?: string;
  /** Visual emphasis colour for the value. */
  tone?: "default" | "success" | "warning" | "danger" | "muted";
  /** Loading flag — keeps the card layout while the call is in flight. */
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  subtitle: "",
  hint: "",
  tone: "default",
  loading: false,
});

const displayValue = computed(() => {
  if (props.loading) return "…";
  if (props.value === null || props.value === undefined) return "—";
  return props.value.toLocaleString("en-US");
});
</script>

<template>
  <article class="kpi" :class="`kpi--${tone}`" :aria-busy="loading">
    <header class="kpi__head">
      <span class="kpi__label">{{ label }}</span>
      <span v-if="hint" class="kpi__hint" :title="hint">
        <AppIcon name="info" size="xs" />
      </span>
    </header>
    <div class="kpi__value">{{ displayValue }}</div>
    <p v-if="subtitle" class="kpi__sub">{{ subtitle }}</p>
  </article>
</template>

<style scoped>
.kpi {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  min-width: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.02), 0 4px 12px rgba(0, 0, 0, 0.01);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
}

.kpi::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 3px;
  background: transparent;
  transition: background-color 0.2s ease;
}

.kpi:hover {
  transform: translateY(-3px);
  border-color: var(--border-strong);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
}

.kpi--success::before { background: var(--state-success); }
.kpi--warning::before { background: var(--state-warning); }
.kpi--danger::before { background: var(--state-danger); }
.kpi--muted::before { background: var(--text-placeholder); }

.kpi__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-2);
}

.kpi__label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.kpi__hint {
  display: inline-flex;
  color: var(--text-tertiary);
  cursor: help;
  transition: color 0.15s ease;
}

.kpi__hint:hover {
  color: var(--text-primary);
}

.kpi__value {
  font-size: clamp(1.75rem, 2.2vw, 2.5rem);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
  line-height: 1.1;
  letter-spacing: -0.02em;
}

.kpi--success .kpi__value { color: var(--state-success); }
.kpi--warning .kpi__value { color: var(--state-warning); }
.kpi--danger .kpi__value { color: var(--state-danger); }
.kpi--muted .kpi__value { color: var(--text-tertiary); }

.kpi__sub {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  font-weight: var(--font-weight-medium);
}

:root[data-theme="dark"] .kpi {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2), 0 4px 12px rgba(0, 0, 0, 0.15);
}

:root[data-theme="dark"] .kpi:hover {
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.35);
  background: var(--bg-card-hover);
}
</style>
