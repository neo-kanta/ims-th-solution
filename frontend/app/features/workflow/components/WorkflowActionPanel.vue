<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "~/composables/useI18n";
import { useAuthStore } from "~/stores/useAuthStore";
import { ACTION_CATALOG } from "../permissions";
import type { WorkflowAction } from "../types";
import WorkflowOperationConfirmDialog from "./WorkflowOperationConfirmDialog.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

const props = defineProps<{
  allowedOperations?: string[];
  blockedReasons?: Array<{ code?: string; message?: string }>;
  executing: boolean;
  criticalBlocker: boolean;
}>();

const emit = defineEmits<{
  execute: [
    operationType: string,
    remark?: string,
    zeroTransactionAttestation?: boolean,
    attestationReason?: string,
    notes?: string,
  ];
}>();

const { t } = useI18n();
const authStore = useAuthStore();

// Dialog state
const confirmOpen = ref(false);
const selectedAction = ref<WorkflowAction | null>(null);
const remarkInput = ref("");
const notesInput = ref("");
const zeroAttestation = ref(false);
const attestationReason = ref("");

// Clean dialog inputs when changed
watch(confirmOpen, (open) => {
  if (!open) {
    remarkInput.value = "";
    notesInput.value = "";
    zeroAttestation.value = false;
    attestationReason.value = "";
  }
});

// Map backend operation names to UI catalog
const allowedList = computed(() => {
  if (!props.allowedOperations) return [];
  return props.allowedOperations
    .map((op) => ACTION_CATALOG[op as WorkflowAction])
    .filter(Boolean);
});

// Map of all operations to see which ones are not in allowed list
const blockedList = computed(() => {
  const allowedSet = new Set(props.allowedOperations || []);
  return Object.values(ACTION_CATALOG).filter(
    (opt) => !allowedSet.has(opt.value),
  );
});

const activeOption = computed(() => {
  return selectedAction.value ? ACTION_CATALOG[selectedAction.value] : null;
});

// Validation rules
const isRemarkValid = computed(() => {
  if (!activeOption.value?.requiresReason) return true;
  return remarkInput.value.trim().length >= 20;
});

const isAttestationValid = computed(() => {
  if (selectedAction.value !== "MANAGER_APPROVE") return true;
  if (!zeroAttestation.value) return true;
  return attestationReason.value.trim().length >= 30;
});

const canConfirm = computed(() => {
  return isRemarkValid.value && isAttestationValid.value && !props.executing;
});

function openConfirm(action: WorkflowAction) {
  if (props.criticalBlocker) return;
  selectedAction.value = action;
  confirmOpen.value = true;
}

function handleConfirm() {
  if (!selectedAction.value || !canConfirm.value) return;
  emit(
    "execute",
    selectedAction.value,
    remarkInput.value,
    zeroAttestation.value,
    attestationReason.value,
    notesInput.value,
  );
  confirmOpen.value = false;
}
</script>

<template>
  <div class="workflow-action-panel">
    <!-- Critical blocker alert -->
    <div v-if="criticalBlocker" class="workflow-alert is-danger">
      <AppIcon name="warning" size="sm" class="workflow-alert__icon" />
      <div class="workflow-alert__content">
        <span class="workflow-alert__title">Operations Gated</span>
        <p class="workflow-alert__description">
          Critical core modules are not ready. All transition operations have been disabled for security.
        </p>
      </div>
    </div>

    <!-- Active Action Buttons -->
    <div class="workflow-actions-section">
      <h3 class="workflow-section-title">
        {{ t("workflow.overview.actions", "Allowed Operations") }}
      </h3>
      <div v-if="allowedList.length === 0" class="workflow-no-actions-card">
        <p class="workflow-no-actions-card__text">
          No workflow transition operations are permitted in the current state.
        </p>
      </div>
      <div v-else class="workflow-buttons-grid">
        <AppButton
          v-for="opt in allowedList"
          :key="opt.value"
          :variant="opt.tone === 'primary' ? 'primary' : 'secondary'"
          size="lg"
          class="workflow-action-button"
          :class="`is-tone-${opt.tone}`"
          :disabled="criticalBlocker || executing || !authStore.hasPermission(opt.permission)"
          @click="openConfirm(opt.value)"
        >
          {{ t(opt.labelKey, opt.labelFallback) }}
        </AppButton>
      </div>
    </div>

    <!-- Blocked Operations List -->
    <div class="workflow-blocked-section">
      <h3 class="workflow-section-title">
        {{ t("workflow.overview.blocked", "Unavailable Transitions") }}
      </h3>
      <div class="workflow-blocked-list">
        <div
          v-for="opt in blockedList"
          :key="opt.value"
          class="workflow-blocked-item"
        >
          <div class="workflow-blocked-item__left">
            <AppIcon name="close" size="xs" class="workflow-blocked-item__icon" />
            <span class="workflow-blocked-item__name">
              {{ t(opt.labelKey, opt.labelFallback) }}
            </span>
          </div>
          <span class="workflow-blocked-item__reason">
            {{ t("workflow.blocked.reasonText", "Not permitted in current state") }}
          </span>
        </div>
      </div>
    </div>

    <!-- Backend Blocked Reasons Banner -->
    <div v-if="blockedReasons && blockedReasons.length > 0" class="workflow-reasons-panel">
      <div
        v-for="(block, idx) in blockedReasons"
        :key="idx"
        class="workflow-alert is-warning"
      >
        <AppIcon name="warning" size="sm" class="workflow-alert__icon" />
        <div class="workflow-alert__content">
          <span v-if="block.code" class="workflow-alert__title">{{ block.code }}</span>
          <p class="workflow-alert__description">{{ block.message }}</p>
        </div>
      </div>
    </div>

    <!-- Confirmation Dialog -->
    <WorkflowOperationConfirmDialog
      :open="confirmOpen"
      :executing="executing"
      :tone="activeOption?.tone || 'primary'"
      :title="t(activeOption?.labelKey || '' as any, activeOption?.labelFallback || '')"
      :description="t('workflow.confirm.descriptionText' as any, 'Are you sure you want to execute this operation? This will change the daily workflow state and write to the audit trail.')"
      confirm-label="Confirm Action"
      @cancel="confirmOpen = false"
      @confirm="handleConfirm"
    >
      <template #inputs>
        <div class="workflow-dialog-inputs">
          <!-- Note/Notes for forward operations -->
          <div v-if="selectedAction === 'MANAGER_APPROVE'" class="workflow-dialog-field">
            <label class="workflow-dialog-field__label">
              {{ t("workflow.fields.notes", "Notes (Optional)") }}
            </label>
            <input
              v-model="notesInput"
              type="text"
              class="workflow-dialog-field__input"
              placeholder="e.g. Approved after daily holdings verification"
            />
          </div>

          <!-- Attestation for Manager Approval -->
          <div v-if="selectedAction === 'MANAGER_APPROVE'" class="workflow-dialog-attestation">
            <label class="workflow-dialog-attestation__checkbox-label">
              <input
                v-model="zeroAttestation"
                type="checkbox"
                class="workflow-dialog-attestation__checkbox"
              />
              <span>No transactions were executed today. I attest that this is correct.</span>
            </label>
            <div v-if="zeroAttestation" class="workflow-dialog-field">
              <label class="workflow-dialog-field__label">
                Attestation Reason (Required - Min 30 characters)
              </label>
              <textarea
                v-model="attestationReason"
                class="workflow-dialog-field__textarea"
                rows="3"
                placeholder="Please explain why no transactions were posted today (at least 30 characters)..."
              ></textarea>
              <span class="workflow-dialog-field__helper" :class="{ 'is-valid': attestationReason.trim().length >= 30 }">
                {{ attestationReason.trim().length }} / 30 characters
              </span>
            </div>
          </div>

          <!-- Reason for Cancel/Rollback -->
          <div v-if="activeOption?.requiresReason" class="workflow-dialog-field">
            <label class="workflow-dialog-field__label">
              Reason (Required - Min 20 characters)
            </label>
            <textarea
              v-model="remarkInput"
              class="workflow-dialog-field__textarea"
              rows="3"
              placeholder="Explain the reason for this rollback/cancellation..."
            ></textarea>
            <span class="workflow-dialog-field__helper" :class="{ 'is-valid': remarkInput.trim().length >= 20 }">
              {{ remarkInput.trim().length }} / 20 characters
            </span>
          </div>
        </div>
      </template>
    </WorkflowOperationConfirmDialog>
  </div>
</template>

<style scoped>
.workflow-action-panel {
  display: grid;
  gap: var(--space-5);
}

.workflow-section-title {
  margin: 0 0 var(--space-3) 0;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.workflow-no-actions-card {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  background: var(--bg-card-muted, #f9fafb);
  text-align: center;
}

.workflow-no-actions-card__text {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.workflow-buttons-grid {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.workflow-action-button {
  min-width: 180px;
}

.workflow-action-button.is-tone-warning {
  border-color: var(--color-warning-500);
  color: var(--color-warning-700);
}

.workflow-action-button.is-tone-danger {
  border-color: var(--color-danger-500);
  color: var(--color-danger-700);
}

.workflow-blocked-list {
  display: grid;
  gap: var(--space-2);
}

.workflow-blocked-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-2) var(--space-3);
  background: var(--bg-row-hover);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
}

.workflow-blocked-item__left {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.workflow-blocked-item__icon {
  color: var(--text-placeholder);
}

.workflow-blocked-item__name {
  color: var(--text-secondary);
  font-weight: var(--font-weight-medium);
}

.workflow-blocked-item__reason {
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
}

.workflow-alert {
  display: flex;
  gap: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4);
  align-items: flex-start;
}

.workflow-alert.is-warning {
  border-color: var(--color-warning-200);
  background: var(--color-warning-50);
  color: var(--color-warning-900);
}

.workflow-alert.is-danger {
  border-color: var(--color-danger-200);
  background: var(--color-danger-50);
  color: var(--color-danger-900);
}

.workflow-alert__icon {
  flex-shrink: 0;
  margin-top: 2px;
}

.workflow-alert__content {
  display: grid;
  gap: var(--space-1);
}

.workflow-alert__title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
}

.workflow-alert__description {
  margin: 0;
  font-size: var(--font-size-xs);
  line-height: 1.4;
}

.workflow-reasons-panel {
  display: grid;
  gap: var(--space-2);
}

.workflow-dialog-inputs {
  display: grid;
  gap: var(--space-4);
  margin-top: var(--space-4);
  text-align: left;
}

.workflow-dialog-field {
  display: grid;
  gap: var(--space-1);
}

.workflow-dialog-field__label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
}

.workflow-dialog-field__input {
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  background: var(--bg-card);
  color: var(--text-primary);
}

.workflow-dialog-field__textarea {
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  background: var(--bg-card);
  color: var(--text-primary);
  resize: vertical;
}

.workflow-dialog-field__helper {
  font-size: var(--font-size-2xs);
  color: var(--color-danger-500);
  text-align: right;
}

.workflow-dialog-field__helper.is-valid {
  color: var(--color-success-500);
}

.workflow-dialog-attestation {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-3);
  background: var(--color-neutral-50);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.workflow-dialog-attestation__checkbox-label {
  display: flex;
  align-items: flex-start;
  gap: var(--space-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
  cursor: pointer;
}

.workflow-dialog-attestation__checkbox {
  margin-top: 3px;
}
</style>
