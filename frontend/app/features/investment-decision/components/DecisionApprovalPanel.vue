<script setup lang="ts">
import type { ApiDecision } from "../services/decisionApi";

const props = defineProps<{ decision: ApiDecision }>();

const stages = computed(() => {
  const total = props.decision.approval_total_stages ?? 0;
  if (total === 0) return [];
  return Array.from({ length: total }, (_, i) => {
    const stageNo = i + 1;
    const current = props.decision.approval_stage ?? 0;
    let status: string;
    if (stageNo < current) status = "Done";
    else if (stageNo === current) status = "In Review";
    else status = "Pending";
    const approvers =
      stageNo === current ? props.decision.current_approvers ?? [] : [];
    return { stageNo, status, approvers };
  });
});
</script>

<template>
  <div class="panel-header">Approval Stages</div>
  <div class="approval-stages">
    <div v-if="stages.length === 0" class="panel-empty">
      No approval flow yet.
    </div>
    <div
      v-for="s in stages"
      :key="s.stageNo"
      class="stage-card"
    >
      <div class="stage-card__header">
        <span class="stage-badge">Stage {{ s.stageNo }}</span>
        <span class="stage-status" :data-status="s.status">{{ s.status }}</span>
      </div>
      <div v-if="s.approvers.length > 0" class="stage-card__approvers">
        <div
          v-for="name in s.approvers"
          :key="name"
          class="stage-card__name"
        >
          {{ name }}
        </div>
      </div>
    </div>

    <div v-if="decision.previous_approvers && decision.previous_approvers.length > 0" class="previous-section">
      <div class="section-label">Previous Approvers</div>
      <div
        v-for="name in decision.previous_approvers"
        :key="name"
        class="approver-chip"
      >
        {{ name }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel-header {
  background: var(--bg-card-hover);
  border-bottom: 1px solid var(--border-subtle);
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.approval-stages {
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow-y: auto;
  flex: 1;
}

.stage-card {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md, 6px);
  padding: 10px;
  background: var(--bg-card);
}

.stage-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.stage-badge {
  font-size: 10px;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
}

.stage-status {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 8px;
  background: var(--bg-card-muted);
  color: var(--text-secondary);
}

.stage-status[data-status="Done"] {
  background: rgba(26, 127, 55, 0.15);
  color: var(--state-success, #1a7f37);
}

.stage-status[data-status="In Review"] {
  background: rgba(154, 103, 0, 0.15);
  color: var(--state-warning, #9a6700);
}

.stage-status[data-status="Pending"] {
  background: rgba(207, 34, 46, 0.1);
  color: var(--state-danger, #cf222e);
}

.stage-card__approvers {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stage-card__name {
  font-size: 12px;
  color: var(--text-primary);
}

.previous-section {
  margin-top: 4px;
}

.section-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-tertiary);
  margin-bottom: 4px;
}

.approver-chip {
  font-size: 12px;
  color: var(--text-secondary);
  padding: 2px 0;
}

.panel-empty {
  padding: 16px 12px;
  font-size: 13px;
  color: var(--text-secondary);
  text-align: center;
}
</style>
