<script setup lang="ts">
/**
 * Operation form for executing a workflow transition.
 *
 * Gating layers (must all hold):
 *   1. backend `allowedActions[]` contains the selected action
 *   2. operator has the matching WORKFLOW_* permission (UX only — backend
 *      is final authority)
 *   3. cancel/rollback actions: reason ≥ 20 chars
 *   4. APPROVE + zeroTransactionAttestation: reason ≥ 30 chars
 *
 * Emits `submit` with a typed payload; the parent (the operations page)
 * opens the confirmation dialog and forwards to the store.
 */
import { computed, ref, watch } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

import { ACTION_CATALOG } from "../permissions";
import type { ActionOption } from "../permissions";
import type { WorkflowAction } from "../types";

export interface OperationFormPayload {
  action: WorkflowAction;
  reason: string;
  notes: string;
  zeroTransactionAttestation: boolean;
  attestationReason: string;
}

const props = defineProps<{
  /** Subset of ACTION_CATALOG already filtered by backend allowedActions. */
  visibleActions: readonly ActionOption[];
  hasContract: boolean;
  executing: boolean;
  /** Predicate from the parent — backed by useAuthStore.hasPermission. */
  hasPermission: (code: string) => boolean;
  labels: {
    operation: string;
    chooseOperation: string;
    reason: string;
    reasonPlaceholder: string;
    notes: string;
    zeroAttestation: string;
    attestationPlaceholder: string;
    execute: string;
    disabledNoContract: string;
    disabledNoActions: string;
    disabledNoAction: string;
    disabledMissingPermission: string;
    disabledReasonRequired: string;
    disabledAttestationRequired: string;
  };
}>();

const emit = defineEmits<{
  submit: [payload: OperationFormPayload];
}>();

const selectedAction = ref<WorkflowAction | "">("");
const reasonInput = ref("");
const notesInput = ref("");
const zeroAttestationInput = ref(false);
const attestationReasonInput = ref("");

watch(
  () => props.visibleActions,
  (actions) => {
    if (
      selectedAction.value
      && !actions.some((a) => a.value === selectedAction.value)
    ) {
      selectedAction.value = "";
    }
    if (!selectedAction.value && actions.length === 1) {
      const first = actions[0];
      if (first) selectedAction.value = first.value;
    }
  },
  { immediate: true },
);

const selectedOption = computed<ActionOption | null>(() => {
  if (!selectedAction.value) return null;
  return ACTION_CATALOG[selectedAction.value];
});

const hasActionPermission = computed(() => {
  const opt = selectedOption.value;
  if (!opt) return false;
  return props.hasPermission(opt.permission);
});

const reasonRequired = computed(
  () => Boolean(selectedOption.value?.requiresReason),
);

const reasonValid = computed(() => {
  if (!reasonRequired.value) return true;
  return reasonInput.value.trim().length >= 20;
});

const attestationValid = computed(() => {
  if (selectedAction.value !== "APPROVE") return true;
  if (!zeroAttestationInput.value) return true;
  return attestationReasonInput.value.trim().length >= 30;
});

const canSubmit = computed(() => {
  if (!props.hasContract) return false;
  if (!selectedAction.value) return false;
  if (!hasActionPermission.value) return false;
  if (!reasonValid.value) return false;
  if (!attestationValid.value) return false;
  if (props.executing) return false;
  return true;
});

const disabledReason = computed(() => {
  if (!props.hasContract) return props.labels.disabledNoContract;
  if (props.visibleActions.length === 0) return props.labels.disabledNoActions;
  if (!selectedAction.value) return props.labels.disabledNoAction;
  if (!hasActionPermission.value) return props.labels.disabledMissingPermission;
  if (!reasonValid.value) return props.labels.disabledReasonRequired;
  if (!attestationValid.value) return props.labels.disabledAttestationRequired;
  return "";
});

function resetInputs() {
  reasonInput.value = "";
  notesInput.value = "";
  zeroAttestationInput.value = false;
  attestationReasonInput.value = "";
}

function handleSubmit() {
  if (!canSubmit.value || !selectedAction.value) return;
  emit("submit", {
    action: selectedAction.value,
    reason: reasonInput.value,
    notes: notesInput.value,
    zeroTransactionAttestation: zeroAttestationInput.value,
    attestationReason: attestationReasonInput.value,
  });
}

function buttonVariant(): "primary" | "warning" | "danger" {
  const tone = selectedOption.value?.tone;
  if (tone === "danger") return "danger";
  if (tone === "warning") return "warning";
  return "primary";
}

defineExpose({
  canSubmit,
  disabledReason,
  reasonValid,
  attestationValid,
  hasActionPermission,
  resetInputs,
});
</script>

<template>
  <form
    class="workflow-form"
    data-testid="workflow-operation-form"
    @submit.prevent="handleSubmit"
  >
    <div class="workflow-form__field">
      <label class="workflow-form__label" for="workflow-action">
        {{ labels.operation }}
      </label>
      <select
        id="workflow-action"
        v-model="selectedAction"
        class="workflow-form__input"
        :disabled="!hasContract || visibleActions.length === 0"
        data-testid="workflow-action-select"
      >
        <option value="">{{ labels.chooseOperation }}</option>
        <option
          v-for="opt in visibleActions"
          :key="opt.value"
          :value="opt.value"
        >
          {{ opt.labelFallback }}
        </option>
      </select>
    </div>

    <div v-if="reasonRequired" class="workflow-form__field">
      <label class="workflow-form__label" for="workflow-reason">
        {{ labels.reason }}
      </label>
      <textarea
        id="workflow-reason"
        v-model="reasonInput"
        rows="2"
        class="workflow-form__input"
        :placeholder="labels.reasonPlaceholder"
        data-testid="workflow-reason-input"
      />
    </div>

    <div v-if="selectedAction === 'APPROVE'" class="workflow-form__field">
      <label class="workflow-form__checkbox">
        <input
          v-model="zeroAttestationInput"
          type="checkbox"
          data-testid="workflow-zero-attest"
        />
        <span>{{ labels.zeroAttestation }}</span>
      </label>
      <textarea
        v-if="zeroAttestationInput"
        v-model="attestationReasonInput"
        rows="2"
        class="workflow-form__input"
        :placeholder="labels.attestationPlaceholder"
        data-testid="workflow-attestation-input"
      />
    </div>

    <div v-if="selectedAction === 'APPROVE'" class="workflow-form__field">
      <label class="workflow-form__label" for="workflow-notes">
        {{ labels.notes }}
      </label>
      <textarea
        id="workflow-notes"
        v-model="notesInput"
        rows="2"
        class="workflow-form__input"
      />
    </div>

    <div class="workflow-form__actions">
      <div
        v-if="disabledReason"
        class="workflow-form__help"
        data-testid="workflow-disabled-reason"
      >
        <AppIcon name="info" size="xs" />
        <span>{{ disabledReason }}</span>
      </div>
      <AppButton
        type="submit"
        :variant="buttonVariant()"
        size="sm"
        :disabled="!canSubmit"
        :loading="executing"
        data-testid="workflow-submit"
      >
        <AppIcon name="check" size="xs" />
        <span>{{ labels.execute }}</span>
      </AppButton>
    </div>
  </form>
</template>

<style scoped>
.workflow-form {
  display: grid;
  gap: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--space-4);
  background: var(--bg-card, white);
}

.workflow-form__field {
  display: grid;
  gap: var(--space-1);
}

.workflow-form__label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.workflow-form__input {
  width: 100%;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  background: var(--bg-input, white);
  color: var(--text-primary);
  font-family: inherit;
}

.workflow-form__input:focus {
  outline: 2px solid var(--action-primary);
  outline-offset: -1px;
  border-color: var(--action-primary);
}

.workflow-form__input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.workflow-form__checkbox {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-sm);
}

.workflow-form__actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.workflow-form__help {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}
</style>
