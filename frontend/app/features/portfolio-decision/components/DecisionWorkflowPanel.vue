<script setup lang="ts">
import { computed } from "vue";
import { useI18n, type AppTranslationKey } from "~/composables/useI18n";
import { decisionWorkflowStages } from "../lib/decisionFormat";

const props = defineProps<{
  status?: string | null;
}>();

const { t } = useI18n();
const stages = computed(() => decisionWorkflowStages(props.status));
</script>

<template>
  <ol class="workflow-panel" :aria-label="t('portfolio.decisionNew.nextPanelTitle')">
    <li
      v-for="stage in stages"
      :key="stage.key"
      class="workflow-panel__stage"
      :class="`is-${stage.state}`"
    >
      <span class="workflow-panel__dot" aria-hidden="true" />
      <span class="workflow-panel__label">{{ t(stage.labelKey as AppTranslationKey, stage.label) }}</span>
    </li>
  </ol>
</template>

<style scoped>
.workflow-panel {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 10px;
}

.workflow-panel__stage {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary, #6e7781);
}

.workflow-panel__dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  border: 2px solid var(--border-default, #d0d7de);
  background: var(--bg-card, #fff);
  flex-shrink: 0;
}

.workflow-panel__stage.is-complete {
  color: var(--state-success, #1a7f37);
}
.workflow-panel__stage.is-complete .workflow-panel__dot {
  border-color: var(--state-success, #1a7f37);
  background: var(--state-success, #1a7f37);
}

.workflow-panel__stage.is-active {
  color: var(--action-primary, #0969da);
}
.workflow-panel__stage.is-active .workflow-panel__dot {
  border-color: var(--action-primary, #0969da);
  background: var(--action-primary, #0969da);
}

.workflow-panel__stage.is-blocked {
  color: var(--state-danger, #cf222e);
}
.workflow-panel__stage.is-blocked .workflow-panel__dot {
  border-color: var(--state-danger, #cf222e);
  background: var(--state-danger, #cf222e);
}
</style>
