<script setup lang="ts">
import AppButton from "~/shared/ui/AppButton.vue";

defineProps<{
  permissions: {
    canSubmit: boolean;
    canCancelSubmit: boolean;
    canApprove: boolean;
    canReject: boolean;
    canRevokeApproval: boolean;
  };
  loading?: boolean;
}>();

const emit = defineEmits<{
  submit: [];
  cancel: [];
  approve: [];
  reject: [];
  revoke: [];
}>();
</script>

<template>
  <div class="approval-action-bar">
    <div class="action-bar-inner">
      <!-- Submit -->
      <AppButton
        v-if="permissions.canSubmit"
        variant="primary"
        :loading="loading"
        class="action-btn"
        @click="emit('submit')"
      >
        Submit for Approval
      </AppButton>

      <!-- Approve & Reject Group -->
      <template v-if="permissions.canApprove || permissions.canReject">
        <AppButton
          v-if="permissions.canApprove"
          variant="success"
          :loading="loading"
          class="action-btn"
          @click="emit('approve')"
        >
          Approve
        </AppButton>

        <AppButton
          v-if="permissions.canReject"
          variant="danger"
          :disabled="loading"
          class="action-btn"
          @click="emit('reject')"
        >
          Reject
        </AppButton>
      </template>

      <!-- Cancel / Withdraw Submission -->
      <AppButton
        v-if="permissions.canCancelSubmit"
        variant="secondary"
        :loading="loading"
        class="action-btn"
        @click="emit('cancel')"
      >
        Cancel Submission
      </AppButton>

      <!-- Revoke -->
      <AppButton
        v-if="permissions.canRevokeApproval"
        variant="ghost"
        :loading="loading"
        class="action-btn btn-danger-ghost"
        @click="emit('revoke')"
      >
        Revoke Approval
      </AppButton>
    </div>
  </div>
</template>

<style scoped>
.approval-action-bar {
  background: var(--bg-card-hover);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg, 8px);
  padding: var(--space-3, 12px) var(--space-4, 16px);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.02);
}

.action-bar-inner {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3, 12px);
  align-items: center;
}

.action-btn {
  font-weight: 600;
  transition: all 0.15s cubic-bezier(0.4, 0, 0.2, 1);
}

.action-btn:hover {
  transform: translateY(-1px);
}

.action-btn:active {
  transform: translateY(0);
}

.btn-danger-ghost {
  color: var(--color-danger, #cf222e) !important;
}

.btn-danger-ghost:hover {
  background-color: var(--color-danger-subtle, #ffebe9) !important;
}
</style>
