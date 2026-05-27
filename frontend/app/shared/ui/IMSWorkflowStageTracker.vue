<script setup lang="ts">
import { computed } from "vue";
import AppBadge from "./AppBadge.vue";

export type WorkflowStage =
  | "DAY_START"
  | "MANAGER_APPROVAL"
  | "TRANSACTION_CLOSING"
  | "ACCOUNTING_CLOSING";

export type WorkflowStageStatus = "complete" | "active" | "blocked" | "upcoming";

interface Props {
  currentState: string;
  timestamps?: Partial<Record<WorkflowStage, string | null>>;
  actors?: Partial<Record<WorkflowStage, string | null>>;
  blockingReasons?: readonly string[];
  stageLabels?: Partial<Record<WorkflowStage, string>>;
  statusLabels?: Partial<Record<WorkflowStageStatus, string>>;
}

const props = withDefaults(defineProps<Props>(), {
  timestamps: () => ({}),
  actors: () => ({}),
  blockingReasons: () => [],
  stageLabels: () => ({}),
  statusLabels: () => ({}),
});

const defaultStageLabels: Record<WorkflowStage, string> = {
  DAY_START: "Day Start",
  MANAGER_APPROVAL: "Manager Approval",
  TRANSACTION_CLOSING: "Transaction Close",
  ACCOUNTING_CLOSING: "Accounting Close",
};

const defaultStatusLabels: Record<WorkflowStageStatus, string> = {
  complete: "Complete",
  active: "Active",
  blocked: "Blocked",
  upcoming: "Upcoming",
};

const stageOrder: WorkflowStage[] = [
  "DAY_START",
  "MANAGER_APPROVAL",
  "TRANSACTION_CLOSING",
  "ACCOUNTING_CLOSING",
];

const stateRanks: Record<string, number> = {
  NOT_STARTED: 0,
  DAY_STARTED: 1,
  APPROVED: 2,
  TRANSACTION_CLOSED: 3,
  ACCOUNTING_CLOSED: 4,
};

function getStageRank(stage: WorkflowStage): number {
  if (stage === "DAY_START") return 1;
  if (stage === "MANAGER_APPROVAL") return 2;
  if (stage === "TRANSACTION_CLOSING") return 3;
  if (stage === "ACCOUNTING_CLOSING") return 4;
  return 0;
}

const cards = computed(() => {
  const currentRank = stateRanks[props.currentState] ?? 0;
  const hasBlockers = props.blockingReasons.length > 0;

  return stageOrder.map((stage) => {
    const stageRank = getStageRank(stage);
    let status: WorkflowStageStatus = "upcoming";

    if (currentRank >= stageRank) {
      status = "complete";
    } else if (currentRank + 1 === stageRank) {
      status = hasBlockers ? "blocked" : "active";
    }

    return {
      key: stage,
      label: props.stageLabels[stage] || defaultStageLabels[stage],
      status,
      timestamp: props.timestamps[stage] ?? null,
      actor: props.actors[stage] ?? null,
    };
  });
});

const stageBadgeVariant: Record<WorkflowStageStatus, "success" | "info" | "warning" | "neutral"> = {
  complete: "success",
  active: "info",
  blocked: "warning",
  upcoming: "neutral",
};

function formatTimestamp(value: string | null): string {
  if (!value) return "—";
  try {
    return new Intl.DateTimeFormat("en-CA", {
      timeZone: "Asia/Bangkok",
      year: "numeric",
      month: "short",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      hourCycle: "h23",
    }).format(new Date(value));
  } catch {
    return value;
  }
}
</script>

<template>
  <div class="ims-workflow-tracker" data-testid="ims-workflow-tracker">
    <div
      v-for="card in cards"
      :key="card.key"
      class="ims-workflow-tracker__stage"
      :class="`is-${card.status}`"
    >
      <div class="ims-workflow-tracker__header">
        <span class="ims-workflow-tracker__title">{{ card.label }}</span>
        <AppBadge
          :variant="stageBadgeVariant[card.status]"
          size="sm"
          dot
        >
          {{ statusLabels[card.status] || defaultStatusLabels[card.status] }}
        </AppBadge>
      </div>
      <dl class="ims-workflow-tracker__body">
        <div class="ims-workflow-tracker__row">
          <dt class="ims-workflow-tracker__label">Time</dt>
          <dd class="ims-workflow-tracker__val">{{ formatTimestamp(card.timestamp) }}</dd>
        </div>
        <div class="ims-workflow-tracker__row">
          <dt class="ims-workflow-tracker__label">Actor</dt>
          <dd class="ims-workflow-tracker__val">{{ card.actor || "—" }}</dd>
        </div>
      </dl>
    </div>
  </div>
</template>

<style scoped>
.ims-workflow-tracker {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3, 12px);
  width: 100%;
}

.ims-workflow-tracker__stage {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 6px);
  padding: var(--space-3, 12px) var(--space-4, 16px);
  background: var(--bg-card, #ffffff);
  display: grid;
  gap: var(--space-2, 8px);
}

.ims-workflow-tracker__stage.is-complete {
  border-color: var(--state-success, #1f883d);
  background: rgba(31, 136, 61, 0.04);
}

.ims-workflow-tracker__stage.is-active {
  border-color: var(--action-primary, #0969da);
  background: rgba(9, 105, 218, 0.04);
}

.ims-workflow-tracker__stage.is-blocked {
  border-color: var(--state-warning, #9a6700);
  background: rgba(154, 103, 0, 0.06);
}

.ims-workflow-tracker__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2, 8px);
}

.ims-workflow-tracker__title {
  font-size: var(--font-size-sm, 14px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #1f2328);
}

.ims-workflow-tracker__body {
  margin: 0;
  display: grid;
  gap: var(--space-1, 4px);
}

.ims-workflow-tracker__row {
  display: grid;
  grid-template-columns: 3.5rem 1fr;
  gap: var(--space-2, 8px);
}

.ims-workflow-tracker__label {
  font-size: 10px;
  text-transform: uppercase;
  color: var(--text-tertiary, #6e7781);
  letter-spacing: 0.04em;
}

.ims-workflow-tracker__val {
  margin: 0;
  font-size: var(--font-size-xs, 12px);
  color: var(--text-secondary, #57606a);
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}

@media (max-width: 960px) {
  .ims-workflow-tracker {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .ims-workflow-tracker {
    grid-template-columns: 1fr;
  }
}
</style>
