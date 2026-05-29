<script setup lang="ts">
import { ref } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppTextarea from "~/shared/ui/AppTextarea.vue";
import AppFormField from "~/shared/ui/AppFormField.vue";

const props = defineProps<{
  canAct: boolean;
  submitting?: boolean;
  error?: string | null;
}>();

const emit = defineEmits<{
  approve: [comment: string];
  reject: [reason: string];
}>();

const comment = ref("");
const reason = ref("");
const mode = ref<"idle" | "reject">("idle");
const localError = ref<string | null>(null);

function doApprove() {
  localError.value = null;
  emit("approve", comment.value.trim());
}

function startReject() {
  mode.value = "reject";
}

function confirmReject() {
  if (!reason.value.trim()) {
    localError.value = "A rejection reason is required.";
    return;
  }
  localError.value = null;
  emit("reject", reason.value.trim());
}

function cancelReject() {
  mode.value = "idle";
  reason.value = "";
  localError.value = null;
}
</script>

<template>
  <div class="action-panel">
    <p v-if="!canAct" class="action-panel__hint">
      You have no pending approval task for the current stage of this request.
    </p>

    <template v-else>
      <div v-if="mode === 'idle'" class="action-panel__form">
        <AppFormField label="Comment (optional)" hint="Recorded on your approval stamp and the timeline.">
          <AppTextarea v-model="comment" :rows="2" placeholder="Add an optional comment…" />
        </AppFormField>
        <div class="action-panel__buttons">
          <AppButton variant="success" :loading="submitting" @click="doApprove">Approve</AppButton>
          <AppButton variant="danger" :disabled="submitting" @click="startReject">Reject</AppButton>
        </div>
      </div>

      <div v-else class="action-panel__form">
        <AppFormField
          label="Rejection reason"
          required
          :error="localError"
          hint="Rejecting stops the approval and returns the document to the submitter."
        >
          <AppTextarea v-model="reason" :rows="3" placeholder="Explain why this is rejected…" />
        </AppFormField>
        <div class="action-panel__buttons">
          <AppButton variant="danger" :loading="submitting" @click="confirmReject">Confirm rejection</AppButton>
          <AppButton variant="ghost" :disabled="submitting" @click="cancelReject">Cancel</AppButton>
        </div>
      </div>
    </template>

    <p v-if="error || localError" class="action-panel__error" role="alert">
      {{ error || localError }}
    </p>
  </div>
</template>

<style scoped>
.action-panel {
  display: grid;
  gap: var(--space-3, 12px);
}
.action-panel__hint {
  margin: 0;
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-sm, 0.875rem);
}
.action-panel__form {
  display: grid;
  gap: var(--space-3, 12px);
}
.action-panel__buttons {
  display: flex;
  gap: var(--space-2, 8px);
}
.action-panel__error {
  margin: 0;
  color: var(--alert-danger-text, #cf222e);
  font-size: var(--font-size-sm, 0.875rem);
}
</style>
