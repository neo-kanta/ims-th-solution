<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "~/composables/useI18n";
import WorkflowStateBadge from "./WorkflowStateBadge.vue";
import WorkflowStageTracker from "./WorkflowStageTracker.vue";
import WorkflowActionPanel from "./WorkflowActionPanel.vue";
import WorkflowTimeline from "./WorkflowTimeline.vue";
import WorkflowModuleReadiness from "./WorkflowModuleReadiness.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import type { components } from "~/api/ims-api";
import type { WorkflowStage } from "../types";

const props = defineProps<{
  businessDate: string;
  dailyState: components["schemas"]["DailyWorkflowResponse"] | null;
  loading: boolean;
  executing: boolean;
  error: string | null;
}>();

const emit = defineEmits<{
  execute: [
    operationType: string,
    remark?: string,
    zeroTransactionAttestation?: boolean,
    attestationReason?: string,
    notes?: string,
  ];
  refresh: [];
}>();

const { t } = useI18n();

// Map dailyState timeline into tracker stages
const trackerTimestamps = computed(() => {
  const map: Partial<Record<WorkflowStage, string | null>> = {};
  if (!props.dailyState?.timeline) return map;
  for (const entry of props.dailyState.timeline) {
    if (entry.toState) {
      map[entry.toState as WorkflowStage] = entry.executedAt ?? null;
    }
  }
  return map;
});

const trackerActors = computed(() => {
  const map: Partial<Record<WorkflowStage, string | null>> = {};
  if (!props.dailyState?.timeline) return map;
  for (const entry of props.dailyState.timeline) {
    if (entry.toState) {
      map[entry.toState as WorkflowStage] =
        entry.executedByUsername || entry.executedByAccountCode || null;
    }
  }
  return map;
});

const criticalBlocker = computed(() => {
  const readiness = props.dailyState?.moduleReadiness;
  if (!readiness) return false;
  return Object.entries(readiness).some(([name, status]) => {
    return !status.ready && name !== "approval";
  });
});

const isToday = computed(() => !!props.dailyState?.isToday);
const currentState = computed(() => props.dailyState?.currentState ?? "NOT_STARTED");
</script>

<template>
  <div class="workflow-overview-tab">
    <!-- Error State -->
    <div v-if="error" class="workflow-error-card">
      <div class="workflow-error-card__content">
        <h3 class="workflow-error-card__title">Failed to load workflow data</h3>
        <p class="workflow-error-card__text">{{ error }}</p>
        <button type="button" class="workflow-error-card__retry" @click="emit('refresh')">
          Retry Loading
        </button>
      </div>
    </div>

    <!-- Active State Screen -->
    <div v-else-if="dailyState" class="workflow-overview-grid">
      <!-- Left Column: Status Tracker, Action Controls -->
      <div class="workflow-overview-main-col">
        <!-- Current State Info Card -->
        <AppCard class="workflow-status-card">
          <div class="workflow-status-card__header">
            <div class="workflow-status-card__status-row">
              <span class="workflow-status-card__label">Current Day Status:</span>
              <WorkflowStateBadge :state="currentState" />
              <span v-if="isToday" class="workflow-today-badge">Today</span>
            </div>
            <p v-if="dailyState.currentState === 'NOT_STARTED'" class="workflow-status-card__explanation">
              The daily workflow has not been started. Start the investment day to unlock trading operations.
            </p>
            <p v-else-if="dailyState.currentState === 'INVESTMENT_DAY_STARTED'" class="workflow-status-card__explanation">
              The investment day is open. Trading operations are active. Manager approval is required to freeze entries.
            </p>
            <p v-else-if="dailyState.currentState === 'MANAGER_APPROVED'" class="workflow-status-card__explanation">
              Manager has approved the day. Transaction ledger is locked. Proceed to close transactions.
            </p>
            <p v-else-if="dailyState.currentState === 'TRANSACTION_CLOSED'" class="workflow-status-card__explanation">
              Transactions are closed. Finalizing accounting records. Proceed to close accounting.
            </p>
            <p v-else-if="dailyState.currentState === 'ACCOUNTING_CLOSED'" class="workflow-status-card__explanation">
              All books are closed and archived for this business date.
            </p>
          </div>

          <!-- Stage Progress Tracker -->
          <div class="workflow-status-card__tracker-section">
            <WorkflowStageTracker
              :current-state="currentState"
              :timestamps="trackerTimestamps"
              :actors="trackerActors"
              :blocking-reasons="dailyState.blockedReasons"
              :stage-labels="{
                INVESTMENT_DAY_STARTED: t('workflow.stage.INVESTMENT_DAY_STARTED', 'Day Start'),
                MANAGER_APPROVED: t('workflow.stage.MANAGER_APPROVED', 'Manager Approved'),
                TRANSACTION_CLOSED: t('workflow.stage.TRANSACTION_CLOSED', 'Transaction Closed'),
                ACCOUNTING_CLOSED: t('workflow.stage.ACCOUNTING_CLOSED', 'Accounting Closed')
              }"
              :status-labels="{
                complete: t('workflow.status.complete', 'Complete'),
                active: t('workflow.status.active', 'Active'),
                upcoming: t('workflow.status.upcoming', 'Upcoming'),
                blocked: t('workflow.status.blocked', 'Blocked')
              }"
              :field-labels="{
                timestamp: t('workflow.fields.timestamp', 'Time'),
                personnel: t('workflow.fields.personnel', 'Actor')
              }"
            />
          </div>
        </AppCard>

        <!-- Actions Panel -->
        <AppCard>
          <WorkflowActionPanel
            :allowed-operations="dailyState.allowedOperations"
            :blocked-reasons="dailyState.blockedReasons"
            :executing="executing"
            :critical-blocker="criticalBlocker"
            @execute="(...args) => emit('execute', ...args)"
          />
        </AppCard>
      </div>

      <!-- Right Column: Timeline Summary, Subsystem Readiness -->
      <div class="workflow-overview-sidebar-col">
        <!-- Subsystems Readiness -->
        <AppCard :title="t('workflow.overview.readiness' as any, 'Subsystem Readiness')">
          <WorkflowModuleReadiness :readiness="dailyState.moduleReadiness" />
        </AppCard>

        <!-- Timeline Log Summary -->
        <AppCard>
          <WorkflowTimeline :timeline="dailyState.timeline" />
        </AppCard>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="workflow-empty-card">
      <p>Select a business date to display the daily workflow console.</p>
    </div>
  </div>
</template>

<style scoped>
.workflow-overview-tab {
  display: grid;
  gap: var(--space-6);
}

.workflow-overview-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: var(--space-6);
  align-items: start;
}

.workflow-overview-main-col {
  display: grid;
  gap: var(--space-6);
}

.workflow-overview-sidebar-col {
  display: grid;
  gap: var(--space-6);
}

.workflow-status-card__header {
  display: grid;
  gap: var(--space-2);
  margin-bottom: var(--space-5);
}

.workflow-status-card__status-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.workflow-status-card__label {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
}

.workflow-today-badge {
  font-size: var(--font-size-2xs);
  font-weight: var(--font-weight-bold);
  text-transform: uppercase;
  background: var(--color-primary-50);
  color: var(--color-primary-700);
  border: 1px solid var(--color-primary-200);
  padding: 1px var(--space-2);
  border-radius: var(--radius-sm);
}

.workflow-status-card__explanation {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  line-height: 1.55;
}

.workflow-status-card__tracker-section {
  border-top: 1px solid var(--border-subtle);
  padding-top: var(--space-5);
}

.workflow-error-card {
  border: 1px solid var(--color-danger-200);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  background: var(--color-danger-50);
  color: var(--color-danger-900);
  display: flex;
  justify-content: center;
  text-align: center;
}

.workflow-error-card__content {
  display: grid;
  gap: var(--space-3);
  justify-items: center;
}

.workflow-error-card__title {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
}

.workflow-error-card__text {
  margin: 0;
  font-size: var(--font-size-sm);
}

.workflow-error-card__retry {
  border: 1px solid var(--color-danger-500);
  background: transparent;
  color: var(--color-danger-700);
  font-weight: var(--font-weight-semibold);
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: var(--font-size-sm);
  transition: all 0.15s ease;
}

.workflow-error-card__retry:hover {
  background: var(--color-danger-100);
}

.workflow-empty-card {
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-lg);
  padding: var(--space-10);
  text-align: center;
  color: var(--text-secondary);
  font-size: var(--font-size-md);
}

@media (max-width: 1024px) {
  .workflow-overview-grid {
    grid-template-columns: 1fr;
  }
}
</style>
