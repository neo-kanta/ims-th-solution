<script setup lang="ts">
import { computed } from "vue";
import AppStatusBadge from "./AppStatusBadge.vue";

interface Props {
  businessDate: string;
  currentState: string;
  executor?: string | null;
}

const props = withDefaults(defineProps<Props>(), {
  executor: null,
});

const formattedSystemTime = computed(() => {
  if (!import.meta.client) return "";
  return new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Bangkok",
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  }).format(new Date());
});
</script>

<template>
  <div class="ims-op-date-panel">
    <div class="ims-op-date-panel__title">Business Operation Date</div>
    
    <div class="ims-op-date-panel__body">
      <!-- Business Date Value -->
      <div class="ims-op-date-panel__value">{{ businessDate }}</div>

      <!-- Info Rows -->
      <dl class="ims-op-date-panel__details">
        <div class="ims-op-date-panel__row">
          <dt class="ims-op-date-panel__label">Workflow State</dt>
          <dd class="ims-op-date-panel__val">
            <AppStatusBadge :status="currentState" size="sm" />
          </dd>
        </div>
        <div v-if="executor" class="ims-op-date-panel__row">
          <dt class="ims-op-date-panel__label">Active Operator</dt>
          <dd class="ims-op-date-panel__val">
            <strong>{{ executor }}</strong>
          </dd>
        </div>
        <div v-if="formattedSystemTime" class="ims-op-date-panel__row">
          <dt class="ims-op-date-panel__label">System Time</dt>
          <dd class="ims-op-date-panel__val font-mono">{{ formattedSystemTime }}</dd>
        </div>
      </dl>
    </div>
  </div>
</template>

<style scoped>
.ims-op-date-panel {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 6px);
  background: var(--bg-card, #ffffff);
  padding: var(--space-4, 16px);
  display: flex;
  flex-direction: column;
  gap: var(--space-3, 12px);
}

.ims-op-date-panel__title {
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-tertiary, #6e7781);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.ims-op-date-panel__value {
  font-size: var(--font-size-xl, 24px);
  font-weight: var(--font-weight-bold, 700);
  color: var(--text-primary, #1f2328);
  font-variant-numeric: tabular-nums;
  margin-bottom: var(--space-3, 12px);
  border-bottom: 1px dashed var(--border-subtle, #d0d7de);
  padding-bottom: var(--space-2, 8px);
}

.ims-op-date-panel__details {
  margin: 0;
  display: grid;
  gap: var(--space-2, 8px);
}

.ims-op-date-panel__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3, 12px);
  font-size: var(--font-size-sm, 14px);
}

.ims-op-date-panel__label {
  color: var(--text-secondary, #57606a);
}

.ims-op-date-panel__val {
  margin: 0;
  color: var(--text-primary, #1f2328);
  text-align: right;
}

.font-mono {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: var(--font-size-xs, 12px);
}
</style>
