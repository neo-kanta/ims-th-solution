<script setup lang="ts">
/**
 * WorkflowStageTracker — 4-stage status visualisation.
 *
 * Stage status semantics:
 *   - complete:   the day has already moved past this stage
 *   - active:     the current state, or the immediate next pending step
 *   - blocked:    the next pending step has blockingReasons from backend
 *   - upcoming:   further out, no decision yet
 *
 * The `blocked` branch is non-dead: when the operator's next step has a
 * backend-reported blocker (e.g. previous day not approved), the active
 * stage downgrades to blocked.
 */
import { computed } from "vue";

import AppBadge from "~/shared/ui/AppBadge.vue";

import { STATE_RANK, WORKFLOW_STAGES } from "../types";
import type {
  WorkflowBlockingReason,
  WorkflowStage,
  WorkflowStageStatus,
  WorkflowStateCode,
} from "../types";

interface StageProps {
  currentState: WorkflowStateCode | string;
  /** Timestamps keyed by stage (ISO, UTC). */
  timestamps: Partial<Record<WorkflowStage, string | null>>;
  /** Actor labels keyed by stage. May be backend username or "—". */
  actors: Partial<Record<WorkflowStage, string | null>>;
  blockingReasons?: readonly WorkflowBlockingReason[];
  stageLabels: Record<WorkflowStage, string>;
  statusLabels: Record<WorkflowStageStatus, string>;
  fieldLabels: { timestamp: string; personnel: string };
}

const props = defineProps<StageProps>();

const stageBadgeVariant: Record<WorkflowStageStatus, string> = {
  complete: "success",
  active: "info",
  upcoming: "neutral",
  blocked: "warning",
};

function rank(state: string): number {
  return STATE_RANK[(state as WorkflowStateCode)] ?? 0;
}

interface StageCard {
  key: WorkflowStage;
  label: string;
  status: WorkflowStageStatus;
  timestamp: string | null;
  actor: string | null;
}

const cards = computed<StageCard[]>(() => {
  const currentRank = rank(props.currentState ?? "NOT_STARTED");
  const hasBlockers = (props.blockingReasons?.length ?? 0) > 0;

  return WORKFLOW_STAGES.map((stage) => {
    const stageRank = rank(stage);
    let status: WorkflowStageStatus;
    if (currentRank >= stageRank) {
      status = "complete";
    } else if (currentRank + 1 === stageRank) {
      status = hasBlockers ? "blocked" : "active";
    } else {
      status = "upcoming";
    }
    return {
      key: stage,
      label: props.stageLabels[stage],
      status,
      timestamp: props.timestamps[stage] ?? null,
      actor: props.actors[stage] ?? null,
    };
  });
});

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

defineExpose({ cards });
</script>

<template>
  <div class="workflow-stages" data-testid="workflow-stage-tracker">
    <div
      v-for="card in cards"
      :key="card.key"
      class="workflow-stage"
      :class="`is-${card.status}`"
      :data-stage="card.key"
      :data-status="card.status"
    >
      <div class="workflow-stage__header">
        <span class="workflow-stage__title">{{ card.label }}</span>
        <AppBadge
          :variant="stageBadgeVariant[card.status] as never"
          size="sm"
          dot
        >
          {{ statusLabels[card.status] }}
        </AppBadge>
      </div>
      <dl class="workflow-stage__body">
        <div>
          <dt>{{ fieldLabels.timestamp }}</dt>
          <dd>{{ formatTimestamp(card.timestamp) }}</dd>
        </div>
        <div>
          <dt>{{ fieldLabels.personnel }}</dt>
          <dd>{{ card.actor ?? "—" }}</dd>
        </div>
      </dl>
    </div>
  </div>
</template>

<style scoped>
.workflow-stages {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
}

.workflow-stage {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--space-3) var(--space-4);
  background: var(--bg-card, white);
  display: grid;
  gap: var(--space-2);
}

.workflow-stage.is-complete {
  border-color: var(--state-success, #10b981);
  background: rgba(16, 185, 129, 0.04);
}

.workflow-stage.is-active {
  border-color: var(--action-primary, #2563eb);
  background: rgba(37, 99, 235, 0.04);
}

.workflow-stage.is-blocked {
  border-color: var(--state-warning, #d97706);
  background: rgba(217, 119, 6, 0.06);
}

.workflow-stage__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.workflow-stage__title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.workflow-stage__body {
  margin: 0;
  display: grid;
  gap: var(--space-1);
}

.workflow-stage__body > div {
  display: grid;
  grid-template-columns: 5rem 1fr;
  gap: var(--space-2);
}

.workflow-stage__body dt {
  font-size: 11px;
  text-transform: uppercase;
  color: var(--text-tertiary);
  letter-spacing: 0.04em;
}

.workflow-stage__body dd {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

@media (max-width: 960px) {
  .workflow-stages {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .workflow-stages {
    grid-template-columns: 1fr;
  }
}
</style>
