<script setup lang="ts">
import { computed } from "vue";

import AppIcon from "~/shared/ui/AppIcon.vue";

interface Props {
  icon?: string;
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
  icon: "",
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
      <span class="kpi__label-wrap">
        <span v-if="icon" class="kpi__icon" aria-hidden="true">
          <AppIcon :name="icon" size="xs" />
        </span>
        <span class="kpi__label">{{ label }}</span>
      </span>
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

.kpi__label-wrap {
  display: inline-flex;
  align-items: center;
  min-width: 0;
  gap: var(--space-2);
}

.kpi__icon {
  display: inline-flex;
  color: var(--text-tertiary);
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

</style>
