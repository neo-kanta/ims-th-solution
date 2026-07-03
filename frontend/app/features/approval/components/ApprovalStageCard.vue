<script setup lang="ts">
import { computed } from "vue";
import type { ApprovalStage } from "../types";
import ApprovalStatusBadge from "./ApprovalStatusBadge.vue";

const props = defineProps<{
  stage: ApprovalStage;
}>();

const modeLabel = computed(() => {
  switch (props.stage.approvalMode) {
    case "SINGLE":
      return "Single Approver";
    case "GROUP_ANY":
      return "Group (Any member)";
    case "TEAM_STAMP":
      return "Team (Stamp count)";
    default:
      return props.stage.approvalMode;
  }
});

function fmtDate(value?: string | null): string {
  if (!value) return "";
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
  <div class="stage-card" :class="`stage-card--${stage.status.toLowerCase()}`">
    <div class="stage-card__header">
      <div class="stage-card__title-grp">
        <span class="stage-card__no">STAGE {{ stage.stageNo }}</span>
        <h4 class="stage-card__name">{{ stage.stageName }}</h4>
      </div>
      <div class="stage-card__meta">
        <span class="stage-card__mode">{{ modeLabel }}</span>
        <ApprovalStatusBadge :status="stage.status" />
      </div>
    </div>

    <!-- Team Stamp Counts -->
    <div v-if="stage.approvalMode === 'TEAM_STAMP'" class="stage-card__team-info">
      <div class="stamp-count">
        <span class="stamp-count__label">Required Stamps</span>
        <span class="stamp-count__value">{{ stage.requiredStampCount ?? 0 }}</span>
      </div>
      <div class="stamp-count">
        <span class="stamp-count__label">Approved Stamps</span>
        <span class="stamp-count__value stamp-count__value--approved">{{ stage.currentStampCount ?? 0 }}</span>
      </div>
      <div v-if="stage.status === 'PENDING'" class="stamp-count">
        <span class="stamp-count__label">Remaining Needed</span>
        <span class="stamp-count__value stamp-count__value--remaining">
          {{ Math.max(0, (stage.requiredStampCount ?? 0) - (stage.currentStampCount ?? 0)) }}
        </span>
      </div>
    </div>

    <!-- Approver candidates / signees -->
    <div class="stage-card__approvers">
      <div class="stage-card__section-title">Approver Roster</div>
      <div v-if="stage.approvers.length === 0" class="stage-card__empty-approvers">
        No approvers configured or assigned.
      </div>
      <ul v-else class="approver-list">
        <li
          v-for="approver in stage.approvers"
          :key="approver.userId"
          class="approver-item"
          :class="`approver-item--${approver.status.toLowerCase()}`"
        >
          <div class="approver-item__main">
            <span class="approver-item__dot" :class="`approver-item__dot--${approver.status.toLowerCase()}`"></span>
            <div class="approver-item__meta-info">
              <span class="approver-item__name">{{ approver.displayName }}</span>
              <span v-if="approver.isAgent" class="agent-tag" title="Signed as delegate/agent">Agent / 代</span>
              <span v-if="approver.principalUserName" class="principal-tag">
                for {{ approver.principalUserName }}
              </span>
            </div>
          </div>

          <div class="approver-item__info">
            <span v-if="approver.actedAt" class="approver-item__date">
              {{ fmtDate(approver.actedAt) }}
            </span>
            <span class="approver-item__status-badge" :class="`status-badge--${approver.status.toLowerCase()}`">
              {{ approver.status }}
            </span>
          </div>

          <div v-if="approver.remark" class="approver-item__remark">
            {{ approver.remark }}
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
.stage-card {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg, 8px);
  background: var(--bg-card);
  padding: var(--space-4, 16px);
  display: flex;
  flex-direction: column;
  gap: var(--space-4, 16px);
  position: relative;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.stage-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.stage-card--pending {
  border-left: 4px solid var(--color-primary-500);
}

.stage-card--approved {
  border-left: 4px solid var(--color-success-500);
  background: linear-gradient(to right, var(--bg-card), var(--bg-card-hover));
}

.stage-card--rejected {
  border-left: 4px solid var(--color-danger-500);
}

.stage-card--skipped {
  opacity: 0.65;
  background: var(--bg-card-hover);
}

.stage-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-3, 12px);
}

.stage-card__title-grp {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stage-card__no {
  font-family: var(--font-mono, monospace);
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--text-tertiary);
}

.stage-card__name {
  margin: 0;
  font-size: var(--font-size-md, 1rem);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary);
}

.stage-card__meta {
  display: flex;
  align-items: center;
  gap: var(--space-3, 12px);
}

.stage-card__mode {
  font-size: var(--font-size-xs, 0.75rem);
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--bg-card-hover);
  padding: 4px 8px;
  border-radius: var(--radius-sm, 4px);
  border: 1px solid var(--border-subtle);
}

.stage-card__team-info {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-5, 20px);
  background: var(--bg-card-hover);
  padding: var(--space-3, 12px);
  border-radius: var(--radius-md, 6px);
  font-size: var(--font-size-sm, 0.875rem);
  border: 1px solid var(--border-subtle);
}

.stamp-count {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stamp-count__label {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.02em;
  color: var(--text-tertiary);
}

.stamp-count__value {
  font-size: var(--font-size-md, 1rem);
  font-weight: 700;
  color: var(--text-primary);
}

.stamp-count__value--approved {
  color: var(--state-success);
}

.stamp-count__value--remaining {
  color: var(--state-warning);
}

.stage-card__approvers {
  display: flex;
  flex-direction: column;
  gap: var(--space-3, 12px);
}

.stage-card__section-title {
  font-size: 0.7rem;
  text-transform: uppercase;
  color: var(--text-tertiary);
  font-weight: var(--font-weight-bold, 700);
  letter-spacing: 0.08em;
}

.stage-card__empty-approvers {
  font-size: var(--font-size-sm, 0.875rem);
  color: var(--text-tertiary);
  font-style: italic;
}

.approver-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-2, 8px);
}

.approver-item {
  display: flex;
  flex-direction: column;
  padding: var(--space-3, 12px);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  transition: border-color 0.15s ease;
}

.approver-item:hover {
  border-color: var(--border-default);
}

.approver-item--approved {
  background: var(--status-approved-bg);
  border-color: rgba(18, 183, 106, 0.2);
}

.approver-item--rejected {
  background: var(--status-rejected-bg);
  border-color: rgba(240, 68, 56, 0.2);
}

.approver-item--waiting {
  background: var(--status-pending-bg);
  border-color: rgba(217, 119, 6, 0.2);
}

.approver-item__main {
  display: flex;
  align-items: center;
  gap: var(--space-3, 12px);
}

.approver-item__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--text-disabled);
  box-shadow: 0 0 0 2px var(--bg-card);
}

.approver-item__dot--approved {
  background: var(--state-success);
  box-shadow: 0 0 6px var(--state-success);
}

.approver-item__dot--rejected {
  background: var(--state-danger);
  box-shadow: 0 0 6px var(--state-danger);
}

.approver-item__dot--waiting {
  background: var(--state-warning);
  box-shadow: 0 0 6px var(--state-warning);
}

.approver-item__meta-info {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  flex-wrap: wrap;
}

.approver-item__name {
  font-size: var(--font-size-sm, 0.875rem);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary);
}

.agent-tag {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.02em;
  background: var(--status-delegated-bg);
  color: var(--status-delegated-text);
  padding: 1px 6px;
  border-radius: 10px;
  border: 1px solid rgba(109, 40, 217, 0.2);
}

.principal-tag {
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-secondary);
}

.approver-item__info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--space-2, 8px);
  padding-left: 20px;
  font-size: var(--font-size-xs, 0.75rem);
}

.approver-item__date {
  color: var(--text-secondary);
}

.approver-item__status-badge {
  font-size: 0.65rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 2px 6px;
  border-radius: 4px;
}

.status-badge--approved {
  color: var(--state-success);
  background: rgba(18, 183, 106, 0.1);
}

.status-badge--rejected {
  color: var(--state-danger);
  background: rgba(240, 68, 56, 0.1);
}

.status-badge--waiting {
  color: var(--state-warning);
  background: rgba(217, 119, 6, 0.1);
}

.approver-item__remark {
  margin-top: var(--space-2, 8px);
  margin-left: 20px;
  padding: var(--space-2, 8px);
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-secondary);
  background: var(--bg-card);
  border-radius: var(--radius-sm, 4px);
  border-left: 2px solid var(--border-default);
}
</style>
