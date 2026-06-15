<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";
import type { components } from "~/api/ims-api";

const props = defineProps<{
  approvers: components["schemas"]["DailyApproverEntry"][];
}>();

const emit = defineEmits<{
  add: [approver: components["schemas"]["ApproverInput"]];
  remove: [accountCode: string];
}>();

const { t } = useI18n();

// Input states
const accountCodeInput = ref("");
const roleInput = ref("");
const addError = ref<string | null>(null);

function handleAdd() {
  addError.value = null;
  const code = accountCodeInput.value.trim();
  if (!code) {
    addError.value = "Account code is required";
    return;
  }

  // Prevent duplicates
  if (props.approvers.some((app) => app.accountCode?.toLowerCase() === code.toLowerCase())) {
    addError.value = "Approver already added";
    return;
  }

  emit("add", {
    accountCode: code,
    role: roleInput.value.trim() || undefined,
    username: code, // Fallback username to accountCode for display
  });

  accountCodeInput.value = "";
  roleInput.value = "";
}
</script>

<template>
  <div class="workflow-approver-selector">
    <!-- List of Configured Approvers (Chips) -->
    <div class="workflow-approvers-list">
      <div v-if="approvers.length === 0" class="workflow-approver-fallback">
        <AppIcon name="shield" size="xs" class="fallback-icon" />
        <span class="fallback-text">
          {{ t("workflow.settings.defaultAdminFallback", "No approvers configured. Default Admin group will act as fallback.") }}
        </span>
      </div>
      <div v-else class="workflow-chips-grid">
        <div
          v-for="app in approvers"
          :key="app.accountCode"
          class="workflow-approver-chip"
        >
          <div class="workflow-approver-chip__avatar">
            {{ (app.username || app.accountCode || "U").slice(0, 2).toUpperCase() }}
          </div>
          <div class="workflow-approver-chip__body">
            <span class="workflow-approver-chip__name" :title="app.username || app.accountCode">
              {{ app.username || app.accountCode }}
            </span>
            <span v-if="app.role" class="workflow-approver-chip__role">
              {{ app.role }}
            </span>
          </div>
          <button
            type="button"
            class="workflow-approver-chip__remove"
            aria-label="Remove approver"
            @click="emit('remove', app.accountCode || '')"
          >
            <AppIcon name="close" size="xs" />
          </button>
        </div>
      </div>
    </div>

    <!-- Add New Approver Panel -->
    <div class="workflow-approver-add-form">
      <div class="add-form-fields">
        <!-- Account Code manual input -->
        <div class="add-form-field">
          <input
            v-model="accountCodeInput"
            type="text"
            class="add-form-field__input"
            :placeholder="t('workflow.settings.accountCodePlaceholder', 'Account code (e.g. jsmith)')"
            @keydown.enter.prevent="handleAdd"
          />
        </div>

        <!-- Role optional input -->
        <div class="add-form-field">
          <input
            v-model="roleInput"
            type="text"
            class="add-form-field__input"
            :placeholder="t('workflow.settings.rolePlaceholder', 'Role/Label (Optional)')"
            @keydown.enter.prevent="handleAdd"
          />
        </div>

        <AppButton
          variant="secondary"
          size="sm"
          class="add-form-button"
          @click="handleAdd"
        >
          <template #icon>
            <AppIcon name="plus" size="xs" />
          </template>
          Add
        </AppButton>
      </div>

      <p v-if="addError" class="add-form-error">
        {{ addError }}
      </p>
    </div>
  </div>
</template>

<style scoped>
.workflow-approver-selector {
  display: grid;
  gap: var(--space-4);
}

.workflow-approvers-list {
  min-height: 48px;
  display: flex;
  align-items: center;
}

.workflow-approver-fallback {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  background: var(--color-warning-50);
  border: 1px solid var(--color-warning-100);
  border-radius: var(--radius-md);
  width: 100%;
}

.fallback-icon {
  color: var(--color-warning-600);
}

.fallback-text {
  font-size: var(--font-size-xs);
  color: var(--color-warning-800);
}

.workflow-chips-grid {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  width: 100%;
}

.workflow-approver-chip {
  display: flex;
  align-items: center;
  background: var(--bg-row-hover);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-pill);
  padding: 3px var(--space-1) 3px var(--space-3);
  gap: var(--space-2);
}

.workflow-approver-chip__avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--color-primary-100);
  color: var(--color-primary-700);
  font-size: 10px;
  font-weight: var(--font-weight-bold);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.workflow-approver-chip__body {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.workflow-approver-chip__name {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-approver-chip__role {
  font-size: 10px;
  color: var(--text-tertiary);
}

.workflow-approver-chip__remove {
  background: transparent;
  border: none;
  color: var(--text-placeholder);
  cursor: pointer;
  padding: 4px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.workflow-approver-chip__remove:hover {
  background: var(--border-subtle);
  color: var(--text-secondary);
}

.workflow-approver-add-form {
  display: grid;
  gap: var(--space-1);
}

.add-form-fields {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.add-form-field {
  flex-grow: 1;
  min-width: 150px;
}

.add-form-field__input {
  width: 100%;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-xs);
  background: var(--bg-card);
  color: var(--text-primary);
  height: 32px;
}

.add-form-field__input:focus {
  outline: none;
  border-color: var(--color-primary-500);
}

.add-form-button {
  height: 32px;
  display: inline-flex;
  align-items: center;
}

.add-form-error {
  margin: var(--space-1) 0 0 0;
  font-size: var(--font-size-2xs);
  color: var(--color-danger-500);
}
</style>
