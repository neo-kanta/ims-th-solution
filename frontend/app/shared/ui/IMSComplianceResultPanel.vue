<script setup lang="ts">
import { ref, computed } from "vue";
import AppIcon from "./AppIcon.vue";
import AppRiskBadge from "./AppRiskBadge.vue";

interface Props {
  ruleName: string;
  result: "pass" | "warning" | "fail" | string;
  actualValue?: string | number | null;
  limitValue?: string | number | null;
  message?: string;
  details?: string;
}

const props = withDefaults(defineProps<Props>(), {
  actualValue: null,
  limitValue: null,
  message: "",
  details: "",
});

const isExpanded = ref(false);

const riskState = computed(() => {
  const r = String(props.result).toLowerCase();
  if (r === "pass") return "pass";
  if (r === "warning") return "warning";
  if (r === "fail") return "fail";
  return "unknown";
});

const outcomeClass = computed(() => {
  return `ims-compliance-result--${riskState.value}`;
});
</script>

<template>
  <div class="ims-compliance-result" :class="outcomeClass">
    <div class="ims-compliance-result__header" @click="details ? isExpanded = !isExpanded : undefined">
      <div class="ims-compliance-result__title-row">
        <h4 class="ims-compliance-result__title">{{ ruleName }}</h4>
        <AppRiskBadge :risk="riskState" size="sm" />
      </div>

      <!-- Values -->
      <div v-if="actualValue !== null || limitValue !== null" class="ims-compliance-result__values">
        <span v-if="actualValue !== null" class="ims-compliance-result__val">
          Actual: <strong>{{ actualValue }}</strong>
        </span>
        <span v-if="actualValue !== null && limitValue !== null" class="ims-compliance-result__sep">/</span>
        <span v-if="limitValue !== null" class="ims-compliance-result__val">
          Limit: <strong>{{ limitValue }}</strong>
        </span>
      </div>

      <!-- Error text / message -->
      <p v-if="message" class="ims-compliance-result__message">{{ message }}</p>

      <div v-if="details" class="ims-compliance-result__expand-btn">
        <AppIcon :name="isExpanded ? 'chevron-down' : 'chevron-right'" size="xs" />
        <span>{{ isExpanded ? "Hide Details" : "Show Details" }}</span>
      </div>
    </div>

    <!-- Details Box -->
    <Transition name="details">
      <div v-if="details && isExpanded" class="ims-compliance-result__details">
        <pre><code>{{ details }}</code></pre>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.ims-compliance-result {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card, #ffffff);
  overflow: hidden;
  transition: all 0.15s ease;
}

.ims-compliance-result--pass {
  border-left: 4px solid var(--state-success, #1f883d);
}

.ims-compliance-result--warning {
  border-left: 4px solid var(--state-warning, #9a6700);
}

.ims-compliance-result--fail {
  border-left: 4px solid var(--state-danger, #cf222e);
}

.ims-compliance-result__header {
  padding: var(--space-4, 16px);
  display: grid;
  gap: var(--space-2, 8px);
  cursor: pointer;
}

.ims-compliance-result:not(:has(.ims-compliance-result__expand-btn)) .ims-compliance-result__header {
  cursor: default;
}

.ims-compliance-result__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3, 12px);
}

.ims-compliance-result__title {
  margin: 0;
  font-size: var(--font-size-sm, 14px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #1f2328);
}

.ims-compliance-result__values {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  font-size: var(--font-size-xs, 12px);
  color: var(--text-secondary, #57606a);
}

.ims-compliance-result__sep {
  color: var(--border-subtle, #d0d7de);
}

.ims-compliance-result__message {
  margin: 0;
  font-size: var(--font-size-xs, 12px);
  line-height: 1.4;
  color: var(--text-secondary, #57606a);
}

.ims-compliance-result--fail .ims-compliance-result__message {
  color: var(--state-danger, #cf222e);
  font-weight: var(--font-weight-medium, 500);
}

.ims-compliance-result__expand-btn {
  display: flex;
  align-items: center;
  gap: var(--space-1, 4px);
  font-size: 11px;
  color: var(--text-link, #0969da);
  font-weight: var(--font-weight-semibold, 600);
}

.ims-compliance-result__expand-btn:hover {
  text-decoration: underline;
}

.ims-compliance-result__details {
  background: var(--bg-card-muted, #f6f8fa);
  border-top: 1px solid var(--border-subtle, #d0d7de);
  padding: var(--space-3, 12px) var(--space-4, 16px);
  font-size: var(--font-size-xs, 12px);
  color: var(--text-primary, #1f2328);
}

.ims-compliance-result__details pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

/* Animations */
.details-enter-active,
.details-leave-active {
  transition: max-height 0.2s ease, opacity 0.15s ease;
  overflow: hidden;
}
.details-enter-from,
.details-leave-to {
  opacity: 0;
  max-height: 0;
}
.details-enter-to,
.details-leave-from {
  opacity: 1;
  max-height: 500px;
}
</style>
