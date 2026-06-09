<script setup lang="ts">
import { computed } from "vue";
import { useI18n, type AppTranslationKey } from "~/composables/useI18n";

import type { ToolCallView, ToolStatus, ValidationStatus } from "../composables/chatStreamReducer";

const props = defineProps<{
  toolCalls?: ToolCallView[];
  validation?: ValidationStatus;
}>();

const { t } = useI18n();

const calls = computed(() => props.toolCalls ?? []);
const validation = computed<ValidationStatus>(() => props.validation ?? "idle");

const show = computed(
  () => calls.value.length > 0 || validation.value !== "idle",
);

// Explicit key maps keep t() type-safe (dynamic dotted keys are not assignable
// to the generated AppTranslationKey union).
const toolKey: Record<ToolStatus, AppTranslationKey> = {
  running: "chat.tool.running",
  ok: "chat.tool.ok",
  error: "chat.tool.error",
  denied: "chat.tool.denied",
};
const validationKey: Record<Exclude<ValidationStatus, "idle">, AppTranslationKey> = {
  running: "chat.validation.running",
  passed: "chat.validation.passed",
  blocked: "chat.validation.blocked",
};

function statusLabel(status: ToolStatus): string {
  return t(toolKey[status]);
}

const validationText = computed(() =>
  validation.value === "idle" ? "" : t(validationKey[validation.value]),
);
</script>

<template>
  <div v-if="show" class="tool-trace" role="status" aria-live="polite">
    <span
      v-for="(tc, i) in calls"
      :key="`${tc.name}-${i}`"
      :class="['tool-chip', `tool-chip--${tc.status}`]"
      :title="statusLabel(tc.status)"
    >
      <span class="tool-chip__dot" aria-hidden="true" />
      <span class="tool-chip__name">{{ tc.name }}</span>
      <span class="tool-chip__status">{{ statusLabel(tc.status) }}</span>
    </span>

    <span
      v-if="validation !== 'idle'"
      :class="['tool-chip', `tool-chip--validation-${validation}`]"
    >
      {{ validationText }}
    </span>
  </div>
</template>

<style scoped>
.tool-trace {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}
.tool-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 500;
  border: 1px solid var(--border-subtle, #e2e8f0);
  background: var(--bg-canvas, #f8fafc);
  color: var(--text-secondary, #475569);
}
.tool-chip__dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}
.tool-chip__name {
  font-family: var(--font-mono, ui-monospace, monospace);
}
.tool-chip__status {
  opacity: 0.75;
}
.tool-chip--running {
  color: var(--text-accent, #3730a3);
  border-color: var(--border-accent, #c7d2fe);
}
.tool-chip--running .tool-chip__dot {
  animation: tool-pulse 1s ease-in-out infinite;
}
.tool-chip--ok {
  color: var(--ok-text, #15803d);
  border-color: var(--ok-border, #bbf7d0);
}
.tool-chip--error {
  color: var(--text-danger, #b91c1c);
  border-color: var(--border-danger, #fecaca);
}
.tool-chip--denied {
  color: var(--warn-text, #b45309);
  border-color: var(--warn-border, #fde68a);
}
.tool-chip--validation-running {
  color: var(--text-accent, #3730a3);
}
.tool-chip--validation-passed {
  color: var(--ok-text, #15803d);
  border-color: var(--ok-border, #bbf7d0);
}
.tool-chip--validation-blocked {
  color: var(--text-danger, #b91c1c);
  border-color: var(--border-danger, #fecaca);
}
@keyframes tool-pulse {
  50% {
    opacity: 0.3;
  }
}
</style>
