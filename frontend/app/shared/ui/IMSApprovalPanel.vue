<script setup lang="ts">
import { computed } from "vue";
import IMSApprovalStamp from "./IMSApprovalStamp.vue";
import AppStatusBadge from "./AppStatusBadge.vue";

interface Approver {
  name: string;
  role?: string;
  status?: "approved" | "rejected" | "pending" | string;
  timestamp?: string | null;
  delegated?: boolean;
}

interface Props {
  stageName: string;
  status?: string;
  approvers: Approver[];
  groupApproval?: boolean;
  teamName?: string;
}

const props = withDefaults(defineProps<Props>(), {
  status: "pending",
  groupApproval: false,
  teamName: "",
});

defineSlots<{
  actions(): unknown;
}>();

const approvedCount = computed(() => {
  return props.approvers.filter((a) => a.status?.toLowerCase() === "approved").length;
});
</script>

<template>
  <div class="ims-approval-panel">
    <div class="ims-approval-panel__header">
      <div class="ims-approval-panel__title-area">
        <h3 class="ims-approval-panel__title">{{ stageName }}</h3>
        <p v-if="teamName || groupApproval" class="ims-approval-panel__meta">
          <span v-if="teamName">Team: <strong>{{ teamName }}</strong></span>
          <span v-if="teamName && groupApproval" class="ims-approval-panel__divider">|</span>
          <span v-if="groupApproval" class="ims-approval-panel__group-tag">Requires Group Sign-off</span>
        </p>
      </div>
      <div class="ims-approval-panel__status-badge">
        <AppStatusBadge :status="status" />
      </div>
    </div>

    <!-- Stamps / Approvers directory -->
    <div class="ims-approval-panel__body">
      <div class="ims-approval-panel__sign-zone">
        <div
          v-for="(approver, idx) in approvers"
          :key="idx"
          class="ims-approval-panel__approver-card"
        >
          <template v-if="approver.status?.toLowerCase() === 'approved' || approver.status?.toLowerCase() === 'rejected'">
            <IMSApprovalStamp
              :name="approver.name"
              :role="approver.role"
              :timestamp="approver.timestamp"
              :delegated="approver.delegated"
              :status="approver.status"
            />
          </template>
          <div v-else class="ims-approval-panel__pending-stamp">
            <span class="ims-approval-panel__pending-role">{{ approver.role || "APPROVER" }}</span>
            <span class="ims-approval-panel__pending-name">{{ approver.name }}</span>
            <span class="ims-approval-panel__pending-status">PENDING</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Actions slot -->
    <div v-if="$slots.actions" class="ims-approval-panel__footer">
      <slot name="actions" />
    </div>
  </div>
</template>

<style scoped>
.ims-approval-panel {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 6px);
  background: var(--bg-card, #ffffff);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.ims-approval-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4, 16px);
  padding: var(--space-4, 16px) var(--space-5, 20px);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.ims-approval-panel__title {
  margin: 0;
  font-size: var(--font-size-md, 16px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #1f2328);
}

.ims-approval-panel__meta {
  margin: var(--space-1, 4px) 0 0;
  font-size: var(--font-size-xs, 12px);
  color: var(--text-secondary, #57606a);
}

.ims-approval-panel__divider {
  margin: 0 var(--space-2, 8px);
  color: var(--border-subtle, #d0d7de);
}

.ims-approval-panel__group-tag {
  color: var(--accent-orange, #f78166);
  font-weight: var(--font-weight-semibold, 600);
}

.ims-approval-panel__body {
  padding: var(--space-5, 20px);
  background: var(--bg-card-muted, #f6f8fa);
}

.ims-approval-panel__sign-zone {
  display: flex;
  gap: var(--space-5, 20px);
  flex-wrap: wrap;
}

.ims-approval-panel__approver-card {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100px;
}

.ims-approval-panel__pending-stamp {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 90px;
  height: 90px;
  border: 1px dashed var(--border-default, #d0d7de);
  border-radius: 50%;
  color: var(--text-tertiary, #6e7781);
  padding: 4px;
}

.ims-approval-panel__pending-role {
  font-size: 8px;
  font-weight: bold;
  border-bottom: 1px dashed var(--border-default, #d0d7de);
  padding-bottom: 2px;
  width: 80%;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ims-approval-panel__pending-name {
  font-size: 10px;
  font-weight: bold;
  padding: 3px 0;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 80px;
}

.ims-approval-panel__pending-status {
  font-size: 8px;
  font-weight: bold;
  letter-spacing: 0.05em;
  color: var(--text-tertiary, #6e7781);
}

.ims-approval-panel__footer {
  padding: var(--space-4, 16px) var(--space-5, 20px);
  border-top: 1px solid var(--border-subtle, #d0d7de);
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2, 8px);
}
</style>
