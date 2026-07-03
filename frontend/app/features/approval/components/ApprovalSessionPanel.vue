<script setup lang="ts">
import { ref, watch } from "vue";
import { useApprovalSession } from "../composables/useApprovalSession";
import { useAppToast } from "~/composables/useAppToast";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";
import AppErrorState from "~/shared/ui/AppErrorState.vue";
import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";
import ApprovalStatusBadge from "./ApprovalStatusBadge.vue";
import ApprovalActionBar from "./ApprovalActionBar.vue";
import ApprovalStageCard from "./ApprovalStageCard.vue";
import ApprovalTimeline from "./ApprovalTimeline.vue";
import ApprovalRemarkDialog from "./ApprovalRemarkDialog.vue";
import ApprovalRejectDialog from "./ApprovalRejectDialog.vue";
import type { ApprovalTarget } from "../types";

const props = defineProps<{
  target: ApprovalTarget;
}>();

const emit = defineEmits<{
  (e: "action-completed"): void;
}>();

const {
  approval,
  rawTimeline,
  loading,
  error,
  actionPending,
  permissions,
  refresh,
  submit,
  cancelSubmission,
  approve,
  reject,
  revokeApproval,
} = useApprovalSession(props.target);

const toast = useAppToast();

watch(
  () => props.target,
  () => {
    void refresh();
  },
  { immediate: true, deep: true }
);

// Remark Dialog State
const remarkDialog = ref({
  open: false,
  action: "" as "SUBMIT" | "APPROVE" | "REVOKE_APPROVAL",
  title: "",
  description: "",
  confirmLabel: "",
});

function openRemarkDialog(action: "SUBMIT" | "APPROVE" | "REVOKE_APPROVAL") {
  remarkDialog.value.action = action;
  if (action === "SUBMIT") {
    remarkDialog.value.title = "Submit for Approval";
    remarkDialog.value.description = "Submit this record to start the approval process.";
    remarkDialog.value.confirmLabel = "Submit";
  } else if (action === "APPROVE") {
    remarkDialog.value.title = "Approve Request";
    remarkDialog.value.description = "Provide an optional comment for your signature stamp.";
    remarkDialog.value.confirmLabel = "Approve";
  } else if (action === "REVOKE_APPROVAL") {
    remarkDialog.value.title = "Revoke Approval";
    remarkDialog.value.description = "Explain why you are revoking the final approved state.";
    remarkDialog.value.confirmLabel = "Revoke";
  }
  remarkDialog.value.open = true;
}

async function handleRemarkConfirm(remark: string) {
  try {
    const act = remarkDialog.value.action;
    if (act === "SUBMIT") {
      await submit(remark);
      toast.showSuccess("Approval request submitted successfully.");
    } else if (act === "APPROVE") {
      await approve(remark);
      toast.showSuccess("Request approved successfully.");
    } else if (act === "REVOKE_APPROVAL") {
      await revokeApproval(remark);
      toast.showSuccess("Approval status revoked successfully.");
    }
    remarkDialog.value.open = false;
    emit("action-completed");
  } catch (err: any) {
    toast.showError(err.message || "Action failed.");
  }
}

// Reject Dialog State
const rejectDialogOpen = ref(false);
function openRejectDialog() {
  rejectDialogOpen.value = true;
}
async function handleRejectConfirm(reason: string) {
  try {
    await reject(reason);
    toast.showSuccess("Request rejected successfully.");
    rejectDialogOpen.value = false;
    emit("action-completed");
  } catch (err: any) {
    toast.showError(err.message || "Rejection failed.");
  }
}

// Cancel Confirm State
const cancelConfirmOpen = ref(false);
function confirmCancel() {
  cancelConfirmOpen.value = true;
}
async function handleCancelConfirm() {
  try {
    await cancelSubmission();
    toast.showSuccess("Submission cancelled successfully.");
    cancelConfirmOpen.value = false;
    emit("action-completed");
  } catch (err: any) {
    toast.showError(err.message || "Cancellation failed.");
  }
}
</script>

<template>
  <div class="approval-session-panel">
    <!-- Loading and error states -->
    <AppLoadingState v-if="loading" message="Loading approval session…" />
    <AppErrorState v-else-if="error" :message="error" />

    <template v-else-if="approval">
      <!-- Summary panel -->
      <div class="approval-session-panel__summary">
        <div class="summary-grid">
          <div class="summary-item">
            <span class="summary-item__label">Status</span>
            <div class="summary-item__value-row">
              <ApprovalStatusBadge :status="approval.status" />
              <span
                v-if="approval.status !== 'NOT_SUBMITTED' && approval.currentStageNo"
                class="summary-item__stage"
              >
                Stage {{ approval.currentStageNo }}
              </span>
            </div>
          </div>

          <div v-if="approval.nextApproverNames.length" class="summary-item">
            <span class="summary-item__label">Next Approver</span>
            <span class="summary-item__value">{{ approval.nextApproverNames.join(", ") }}</span>
          </div>
        </div>
      </div>

      <!-- Actions bar -->
      <ApprovalActionBar
        :permissions="permissions"
        :loading="actionPending"
        @submit="openRemarkDialog('SUBMIT')"
        @cancel="confirmCancel"
        @approve="openRemarkDialog('APPROVE')"
        @reject="openRejectDialog"
        @revoke="openRemarkDialog('REVOKE_APPROVAL')"
      />

      <!-- Stage Cards -->
      <div v-if="approval.stages.length" class="approval-session-panel__stages">
        <h5 class="panel-section-title">Approval Stages</h5>
        <div class="stages-list">
          <ApprovalStageCard
            v-for="stage in approval.stages"
            :key="stage.stageNo"
            :stage="stage"
          />
        </div>
      </div>

      <!-- History Timeline -->
      <div v-if="rawTimeline.length" class="approval-session-panel__history">
        <h5 class="panel-section-title">Approval Timeline</h5>
        <ApprovalTimeline :events="rawTimeline" />
      </div>
    </template>

    <!-- Dialogs -->
    <ApprovalRemarkDialog
      :open="remarkDialog.open"
      :title="remarkDialog.title"
      :description="remarkDialog.description"
      :confirm-label="remarkDialog.confirmLabel"
      :loading="actionPending"
      @cancel="remarkDialog.open = false"
      @confirm="handleRemarkConfirm"
    />

    <ApprovalRejectDialog
      :open="rejectDialogOpen"
      :loading="actionPending"
      @cancel="rejectDialogOpen = false"
      @confirm="handleRejectConfirm"
    />

    <AppConfirmDialog
      :open="cancelConfirmOpen"
      title="Cancel Approval Submission"
      description="Are you sure you want to cancel this approval submission? The request will be withdrawn."
      confirm-label="Confirm Cancel"
      tone="warning"
      :loading="actionPending"
      @cancel="cancelConfirmOpen = false"
      @confirm="handleCancelConfirm"
    />
  </div>
</template>

<style scoped>
.approval-session-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-4, 16px);
}

.approval-session-panel__summary {
  background: linear-gradient(135deg, var(--bg-card-hover), var(--bg-card));
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg, 8px);
  padding: var(--space-4, 16px) var(--space-5, 20px);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.03), 0 2px 4px -1px rgba(0, 0, 0, 0.02);
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: var(--space-4, 16px);
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-1, 4px);
}

.summary-item__label {
  font-size: 0.7rem;
  text-transform: uppercase;
  color: var(--text-tertiary);
  font-weight: 700;
  letter-spacing: 0.08em;
}

.summary-item__value-row {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
}

.summary-item__value {
  font-size: var(--font-size-sm, 0.875rem);
  font-weight: 600;
  color: var(--text-primary);
}

.summary-item__stage {
  font-size: 0.7rem;
  color: var(--color-primary-600);
  background: var(--color-primary-50);
  border: 1px solid rgba(37, 99, 235, 0.15);
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 700;
}

.panel-section-title {
  font-size: var(--font-size-sm, 0.875rem);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-secondary);
  margin: 0 0 var(--space-3, 12px) 0;
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-2, 8px);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.approval-session-panel__stages {
  margin-top: var(--space-3, 12px);
}

.stages-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4, 16px);
}

.approval-session-panel__history {
  margin-top: var(--space-3, 12px);
}
</style>
