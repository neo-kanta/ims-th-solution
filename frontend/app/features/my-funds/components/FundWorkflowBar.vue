<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";

import { activeStageIndex, WORKFLOW_STAGES, type WorkflowStageKey } from "../lib/derive";

interface Props {
  currentState: string;
  available: boolean;
  businessDate?: string | null;
}

const props = withDefaults(defineProps<Props>(), { businessDate: null });
const { t } = useI18n();

const stageIndex = computed(() => activeStageIndex(props.currentState));

const stages = computed(() =>
  WORKFLOW_STAGES.map((key, i) => ({
    key,
    label: stageLabel(key),
    state: stageState(i, stageIndex.value, props.available),
  })),
);

function stageLabel(key: WorkflowStageKey): string {
  const labels: Record<WorkflowStageKey, string> = {
    DAY_START: t("myFunds.workflow.stages.dayStart", "Day Start"),
    ANALYSIS: t("myFunds.workflow.stages.analysis", "Analysis"),
    DECISION: t("myFunds.workflow.stages.decision", "Decision"),
    MANAGER_APPROVAL: t("myFunds.workflow.stages.managerApproval", "Mgr Approval"),
    EXECUTION: t("myFunds.workflow.stages.execution", "Execution"),
    TRANSACTION_CLOSED: t("myFunds.workflow.stages.txClosing", "Tx Closing"),
    ACCOUNTING_CLOSED: t("myFunds.workflow.stages.acctgClosing", "Acctg Closing"),
  };
  return labels[key];
}

function stageState(
  i: number,
  active: number,
  available: boolean,
): "done" | "current" | "pending" | "unknown" {
  if (!available) return "unknown";
  if (active < 0) return "pending";
  if (i < active) return "done";
  if (i === active) return "current";
  return "pending";
}

const stageOfLabel = computed(() => {
  if (!props.available || stageIndex.value < 0) {
    return t("myFunds.workflow.notStarted", "Not started");
  }
  return t(
    "myFunds.workflow.stageOf",
    { current: stageIndex.value + 1, total: WORKFLOW_STAGES.length },
    `stage ${stageIndex.value + 1} of ${WORKFLOW_STAGES.length}`,
  );
});
</script>

<template>
  <div class="workflow-bar">
    <div class="workflow-bar__header">
      <span class="workflow-bar__title">{{
        t("myFunds.workflow.title", "Today's workflow")
      }}</span>
      <span class="workflow-bar__progress">{{ stageOfLabel }}</span>
    </div>
    <ol class="workflow-bar__list">
      <li
        v-for="stage in stages"
        :key="stage.key"
        class="workflow-bar__item"
        :data-state="stage.state"
      >
        <span class="workflow-bar__pip" />
        <span class="workflow-bar__label">{{ stage.label }}</span>
      </li>
    </ol>
  </div>
</template>

<style scoped>
.workflow-bar {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.workflow-bar__header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
}

.workflow-bar__title {
  letter-spacing: 0.04em;
  text-transform: uppercase;
  font-weight: 600;
}

.workflow-bar__list {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: 1fr;
  gap: 0;
  list-style: none;
  margin: 0;
  padding: 0;
}

.workflow-bar__item {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 4px;
  position: relative;
  font-size: 10px;
  text-align: left;
  color: var(--text-tertiary, #6e7781);
}

.workflow-bar__pip {
  display: block;
  height: 4px;
  border-radius: 2px;
  background: var(--bg-card-muted, #eaeef2);
}

.workflow-bar__item[data-state="done"] .workflow-bar__pip {
  background: var(--state-success, #1f883d);
}

.workflow-bar__item[data-state="current"] .workflow-bar__pip {
  background: var(--state-warning, #d97706);
  box-shadow: 0 0 0 2px rgba(217, 119, 6, 0.18);
}

.workflow-bar__item[data-state="done"] .workflow-bar__label,
.workflow-bar__item[data-state="current"] .workflow-bar__label {
  color: var(--text-primary, #1f2328);
}

.workflow-bar__item[data-state="current"] .workflow-bar__label {
  font-weight: 600;
}
</style>
