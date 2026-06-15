<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "~/composables/useI18n";
import { useWorkflowSettings } from "../composables/useWorkflowSettings";
import { ACTION_CATALOG } from "../permissions";
import type { WorkflowAction } from "../types";
import WorkflowApproverSelector from "./WorkflowApproverSelector.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";
import type { components } from "~/api/ims-api";

const { t } = useI18n();

const {
  settings,
  loading,
  saving,
  error,
  fetchSettings,
  saveSettings,
} = useWorkflowSettings();

// Local editable copy of settings
const localSettings = ref<components["schemas"]["OperationSettingEntry"][]>([]);

onMounted(fetchSettings);

watch(
  settings,
  (newSettings) => {
    // Deep clone settings
    localSettings.value = JSON.parse(JSON.stringify(newSettings));
  },
  { immediate: true },
);

// Check if settings were modified compared to original fetched settings
const isModified = computed(() => {
  return JSON.stringify(localSettings.value) !== JSON.stringify(settings.value);
});

// Identify which specific operation settings were modified
const modifiedOperations = computed(() => {
  const list: components["schemas"]["OperationSettingEntry"][] = [];
  localSettings.value.forEach((localEntry) => {
    const originalEntry = settings.value.find(
      (s) => s.operationType === localEntry.operationType,
    );
    if (!originalEntry || JSON.stringify(localEntry) !== JSON.stringify(originalEntry)) {
      list.push(localEntry);
    }
  });
  return list;
});

function handleAddApprover(
  operationType: string,
  approver: components["schemas"]["ApproverInput"],
) {
  const entry = localSettings.value.find((s) => s.operationType === operationType);
  if (entry) {
    if (!entry.approvers) entry.approvers = [];
    entry.approvers.push(approver);
  }
}

function handleRemoveApprover(operationType: string, accountCode: string) {
  const entry = localSettings.value.find((s) => s.operationType === operationType);
  if (entry && entry.approvers) {
    entry.approvers = entry.approvers.filter(
      (app) => app.accountCode !== accountCode,
    );
  }
}

function handleCancel() {
  localSettings.value = JSON.parse(JSON.stringify(settings.value));
}

async function handleSave() {
  if (modifiedOperations.value.length === 0 || saving.value) return;

  // Show warning before saving settings
  const message = t(
    "workflow.settings.saveWarningText" as any,
    "Are you sure you want to save approval settings? Unsaved changes for modified operations will be written to the server.",
  );
  if (!confirm(message)) {
    return;
  }

  // Sequentially save each modified operation
  let success = true;
  for (const op of modifiedOperations.value) {
    if (op.operationType) {
      const ok = await saveSettings(
        op.operationType,
        (op.approvers || []).map((app) => ({
          accountCode: app.accountCode,
          role: app.role,
          username: app.username,
        })),
      );
      if (!ok) {
        success = false;
        break;
      }
    }
  }

  if (success) {
    alert(t("workflow.settings.saveSuccessText" as any, "Approval settings saved successfully."));
    await fetchSettings();
  }
}

function getOperationLabel(opType?: string): string {
  if (!opType) return "";
  const opt = ACTION_CATALOG[opType as WorkflowAction];
  return opt ? t(opt.labelKey, opt.labelFallback) : opType;
}
</script>

<template>
  <div class="workflow-settings-tab">
    <!-- Error State -->
    <div v-if="error" class="workflow-error-card">
      <div class="workflow-error-card__content">
        <h3 class="workflow-error-card__title">Failed to load settings</h3>
        <p class="workflow-error-card__text">{{ error }}</p>
        <button type="button" class="workflow-error-card__retry" @click="fetchSettings">
          Retry Loading
        </button>
      </div>
    </div>

    <!-- Loading state -->
    <div v-else-if="loading && localSettings.length === 0" class="workflow-settings-loading">
      <div class="workflow-settings-loading__spinner"></div>
      <span>Loading approval settings...</span>
    </div>

    <template v-else>
      <div class="workflow-settings-grid">
        <!-- Settings Card for each Operation Type -->
        <AppCard
          v-for="op in localSettings"
          :key="op.operationType"
          :title="getOperationLabel(op.operationType)"
          class="workflow-settings-card"
        >
          <div class="workflow-settings-card__body">
            <p class="workflow-settings-card__description">
              Configure personnel authorized to execute this transition.
            </p>
            <div class="workflow-settings-card__mode-row">
              <span class="mode-label">Approval Mode:</span>
              <span class="mode-value">{{ op.approvalMode || 'ANY_OF' }}</span>
            </div>

            <!-- Approver Selector Widget -->
            <WorkflowApproverSelector
              :approvers="op.approvers || []"
              @add="(approver) => handleAddApprover(op.operationType || '', approver)"
              @remove="(code) => handleRemoveApprover(op.operationType || '', code)"
            />
          </div>
        </AppCard>
      </div>

      <!-- Save / Cancel sticky footer -->
      <div v-if="isModified" class="workflow-settings-sticky-footer">
        <div class="footer-content">
          <span class="footer-status-text">
            {{ t("workflow.settings.modifiedCount" as any, "You have unsaved changes.") }}
          </span>
          <div class="footer-actions">
            <AppButton
              variant="secondary"
              size="md"
              :disabled="saving"
              @click="handleCancel"
            >
              Cancel
            </AppButton>
            <AppButton
              variant="primary"
              size="md"
              :loading="saving"
              @click="handleSave"
            >
              Save Changes
            </AppButton>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.workflow-settings-tab {
  display: grid;
  gap: var(--space-6);
  padding-bottom: var(--space-10);
}

.workflow-settings-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-10) 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.workflow-settings-loading__spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border-subtle);
  border-top-color: var(--color-primary-500);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.workflow-settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: var(--space-6);
}

.workflow-settings-card {
  height: 100%;
}

.workflow-settings-card__body {
  display: grid;
  gap: var(--space-4);
}

.workflow-settings-card__description {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
  line-height: 1.4;
}

.workflow-settings-card__mode-row {
  display: flex;
  gap: var(--space-2);
  font-size: var(--font-size-xs);
}

.mode-label {
  color: var(--text-tertiary);
  font-weight: var(--font-weight-medium);
}

.mode-value {
  color: var(--text-primary);
  font-weight: var(--font-weight-semibold);
  font-family: var(--font-mono, ui-monospace, monospace);
}

/* Sticky Action Footer */
.workflow-settings-sticky-footer {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: var(--bg-card);
  border-top: 1px solid var(--border-default);
  padding: var(--space-4) var(--space-6);
  display: flex;
  justify-content: center;
  z-index: 10;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.05);
}

.footer-content {
  width: min(80rem, 100%);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.footer-status-text {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-warning-700);
}

.footer-actions {
  display: flex;
  gap: var(--space-3);
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

@media (max-width: 640px) {
  .workflow-settings-grid {
    grid-template-columns: 1fr;
  }
}
</style>
