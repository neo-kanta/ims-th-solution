<script setup lang="ts">
/**
 * WorkflowDayRail — horizontal day-cycle rail driven by real backend state.
 *
 * Renders the 4 actionable workflow stages (Day Start → Manager Approval →
 * Transaction Closing → Accounting Closing) as a scroller. Stage status is
 * derived from the backend `currentState` (NOT_STARTED … ACCOUNTING_CLOSED)
 * via STATE_RANK — identical topology to WorkflowStageTracker, but laid out
 * as the compact rail used on the dashboard.
 *
 * The CSS mirrors DashboardWorkflowRail's `.workflow-rail__*` classes so the
 * two surfaces look identical; only the data source differs (this one is
 * backed by GET /workflow/day-states/{contractId}).
 */
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";

import { STATE_RANK, WORKFLOW_STAGES } from "../types";
import type { WorkflowStage } from "../types";

const props = defineProps<{
  /** Backend day state code; NOT_STARTED when no row persisted yet. */
  currentState: string;
  /** Per-stage ISO timestamps (UTC). Optional — empty renders no time label. */
  timestamps?: Partial<Record<WorkflowStage, string | null>>;
  /** Optional caption rendered above the rail (e.g. business date). */
  caption?: string;
}>();

const { t } = useI18n();

const stageLabels = computed<Record<WorkflowStage, string>>(() => ({
  DAY_OPEN: t("workflow.action.OPEN_DAY", "Day Start"),
  MANAGER_APPROVED: t("workflow.action.APPROVE", "Manager Approval"),
  TRANSACTION_CLOSED: t("workflow.action.CLOSE_TRANSACTIONS", "Transaction Closing"),
  ACCOUNTING_CLOSED: t("workflow.action.CLOSE_ACCOUNTING", "Accounting Closing"),
}));

function rank(state: string): number {
  return STATE_RANK[state as keyof typeof STATE_RANK] ?? 0;
}

function formatTime(value: string | null | undefined): string {
  if (!value) return "";
  try {
    return new Intl.DateTimeFormat("en-GB", {
      timeZone: "Asia/Bangkok",
      hour: "2-digit",
      minute: "2-digit",
      hourCycle: "h23",
    }).format(new Date(value));
  } catch {
    return "";
  }
}

interface RailStage {
  id: WorkflowStage;
  sequence: number;
  label: string;
  status: "complete" | "active" | "upcoming";
  timeLabel: string;
}

const stages = computed<RailStage[]>(() => {
  const currentRank = rank(props.currentState || "NOT_STARTED");
  return WORKFLOW_STAGES.map((stage, index) => {
    const stageRank = rank(stage);
    let status: RailStage["status"];
    if (currentRank >= stageRank) {
      status = "complete";
    } else if (currentRank + 1 === stageRank) {
      status = "active";
    } else {
      status = "upcoming";
    }
    return {
      id: stage,
      sequence: index + 1,
      label: stageLabels.value[stage],
      status,
      timeLabel: formatTime(props.timestamps?.[stage]),
    };
  });
});
</script>

<template>
  <section class="workflow-rail" aria-label="Workflow day cycle">
    <p v-if="caption" class="workflow-rail__caption">{{ caption }}</p>
    <div class="workflow-rail__scroll">
      <ol class="workflow-rail__list">
        <li
          v-for="(stage, index) in stages"
          :key="stage.id"
          class="workflow-rail__item"
          :data-status="stage.status"
        >
          <span
            v-if="index < stages.length - 1"
            class="workflow-rail__line"
            :class="{ 'is-complete': stage.status === 'complete' }"
          />

          <span
            class="workflow-rail__dot"
            :class="`is-${stage.status}`"
            aria-hidden="true"
          />

          <div class="workflow-rail__copy">
            <div class="workflow-rail__label">
              {{ stage.sequence }}. {{ stage.label }}
            </div>
            <div class="workflow-rail__time">{{ stage.timeLabel || "—" }}</div>
          </div>
        </li>
      </ol>
    </div>
  </section>
</template>

<style scoped>
.workflow-rail {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-xl);
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
}

.workflow-rail__caption {
  margin: 0;
  padding: var(--space-3) var(--space-6) 0;
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.workflow-rail__scroll {
  overflow-x: auto;
  padding: var(--space-5) var(--space-6);
}

.workflow-rail__list {
  display: grid;
  grid-template-columns: repeat(4, minmax(8rem, 1fr));
  gap: var(--space-4);
  min-width: 36rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.workflow-rail__item {
  position: relative;
  display: grid;
  justify-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.workflow-rail__line {
  position: absolute;
  top: 0.45rem;
  left: calc(50% + 0.6rem);
  width: calc(100% - 0.75rem);
  height: 2px;
  background: var(--border-subtle);
}

.workflow-rail__line.is-complete {
  background: var(--state-success);
}

.workflow-rail__dot {
  position: relative;
  z-index: 1;
  width: 14px;
  height: 14px;
  border-radius: 999px;
  background: var(--color-neutral-200);
}

.workflow-rail__dot.is-complete {
  background: var(--state-success);
}

.workflow-rail__dot.is-active {
  background: var(--action-primary);
  box-shadow: 0 0 0 4px var(--focus-ring);
}

.workflow-rail__dot.is-upcoming {
  background: var(--border-strong);
}

.workflow-rail__copy {
  display: grid;
  gap: 2px;
  text-align: center;
}

.workflow-rail__label {
  color: var(--text-primary);
  font-size: 11px;
  font-weight: var(--font-weight-medium);
  line-height: 1.3;
}

.workflow-rail__time {
  color: var(--text-tertiary);
  font-size: 10px;
  line-height: 1.2;
}

@media (max-width: 640px) {
  .workflow-rail__scroll {
    padding: var(--space-4);
  }

  .workflow-rail__list {
    min-width: 30rem;
    gap: var(--space-3);
  }
}
</style>
