<script setup lang="ts">
import { computed } from "vue";
import AppIcon from "./AppIcon.vue";
import AppLoadingState from "./AppLoadingState.vue";

interface Props {
  label: string;
  value: string | number;
  delta?: string | number | null;
  trend?: "up" | "down" | "flat" | null;
  loading?: boolean;
  helpText?: string;
}

const props = withDefaults(defineProps<Props>(), {
  delta: null,
  trend: null,
  loading: false,
  helpText: "",
});

const trendClass = computed(() => {
  if (props.trend === "up") return "trend--up";
  if (props.trend === "down") return "trend--down";
  return "trend--flat";
});

const trendIcon = computed(() => {
  if (props.trend === "up") return "trend-up";
  if (props.trend === "down") return "trend-down";
  return "trend-flat";
});
</script>

<template>
  <div class="app-metric-card" :class="{ 'is-loading': loading }">
    <div v-if="loading" class="app-metric-card__loader">
      <AppLoadingState message="" />
    </div>
    
    <template v-else>
      <div class="app-metric-card__header">
        <span class="app-metric-card__label" :title="helpText">{{ label }}</span>
        <span v-if="helpText" class="app-metric-card__help-icon" :title="helpText">?</span>
      </div>
      
      <div class="app-metric-card__value-group">
        <span class="app-metric-card__value">{{ value }}</span>
        
        <!-- Delta Badge -->
        <span
          v-if="delta !== null && delta !== undefined"
          class="app-metric-card__delta"
          :class="trendClass"
        >
          <AppIcon v-if="trend" :name="trendIcon" size="xs" class="app-metric-card__trend-icon" />
          <span>{{ delta }}</span>
        </span>
      </div>
    </template>
  </div>
</template>

<style scoped>
.app-metric-card {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 6px);
  padding: var(--space-4, 16px) var(--space-5, 20px);
  background: var(--bg-card, #ffffff);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: var(--space-2, 8px);
  min-height: 5.5rem;
  position: relative;
}

.app-metric-card__loader {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.app-metric-card__header {
  display: flex;
  align-items: center;
  gap: var(--space-1, 4px);
}

.app-metric-card__label {
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-secondary, #57606a);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.app-metric-card__help-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  color: var(--text-tertiary, #6e7781);
  font-size: 9px;
  font-weight: bold;
  cursor: help;
}

.app-metric-card__value-group {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-3, 12px);
  flex-wrap: wrap;
}

.app-metric-card__value {
  font-size: var(--font-size-xl, 24px);
  font-weight: var(--font-weight-bold, 700);
  color: var(--text-primary, #1f2328);
  font-variant-numeric: tabular-nums;
  line-height: 1.1;
}

.app-metric-card__delta {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1, 4px);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
  padding: 2px 6px;
  border-radius: var(--radius-sm, 4px);
}

.trend--up {
  color: var(--state-success, #1f883d);
  background: rgba(31, 136, 61, 0.08);
}

.trend--down {
  color: var(--state-danger, #cf222e);
  background: rgba(207, 34, 78, 0.08);
}

.trend--flat {
  color: var(--text-tertiary, #6e7781);
  background: rgba(110, 119, 129, 0.08);
}

.app-metric-card__trend-icon {
  flex-shrink: 0;
}
</style>
