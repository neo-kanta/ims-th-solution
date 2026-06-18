<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "~/composables/useI18n";
import { permissionWorkflowApi } from "~/features/permissions/services/permissionWorkflowApi";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";
import type { components } from "~/api/ims-api";
import type { PermissionUserSummary } from "~/features/permissions/types";

const props = defineProps<{
  approvers: components["schemas"]["DailyApproverEntry"][];
}>();

const emit = defineEmits<{
  add: [approver: components["schemas"]["ApproverInput"]];
  remove: [accountCode: string];
}>();

const { t } = useI18n();

// Search state
const searchInput = ref("");
const searchResults = ref<PermissionUserSummary[]>([]);
const searchLoading = ref(false);
const showDropdown = ref(false);
const selectedUser = ref<PermissionUserSummary | null>(null);

// Role input
const roleInput = ref("");
const addError = ref<string | null>(null);

let searchTimer: ReturnType<typeof setTimeout> | null = null;

function onSearchInput() {
  selectedUser.value = null;
  const q = searchInput.value.trim();
  if (!q || q.length < 2) {
    searchResults.value = [];
    showDropdown.value = false;
    return;
  }
  if (searchTimer) clearTimeout(searchTimer);
  searchTimer = setTimeout(() => doSearch(q), 300);
}

async function doSearch(q: string) {
  searchLoading.value = true;
  try {
    const res = await permissionWorkflowApi.listUsers({ search: q, limit: 8 });
    searchResults.value = res.items;
    showDropdown.value = res.items.length > 0;
  } catch {
    searchResults.value = [];
    showDropdown.value = false;
  } finally {
    searchLoading.value = false;
  }
}

function selectUser(user: PermissionUserSummary) {
  selectedUser.value = user;
  searchInput.value = user.display_name || user.username;
  searchResults.value = [];
  showDropdown.value = false;
  addError.value = null;
}

function clearSelected() {
  selectedUser.value = null;
  searchInput.value = "";
  searchResults.value = [];
  showDropdown.value = false;
}

function handleBlur() {
  setTimeout(() => { showDropdown.value = false; }, 200);
}

function handleAdd() {
  addError.value = null;

  let accountCode: string;
  let displayName: string;

  if (selectedUser.value) {
    // accountCode must equal the JWT claims.Username so the backend matches it
    accountCode = selectedUser.value.username;
    displayName = selectedUser.value.display_name || selectedUser.value.username;
  } else {
    accountCode = searchInput.value.trim();
    displayName = accountCode;
  }

  if (!accountCode) {
    addError.value = "Account code is required";
    return;
  }

  if (props.approvers.some((app) => app.accountCode?.toLowerCase() === accountCode.toLowerCase())) {
    addError.value = "Approver already added";
    return;
  }

  emit("add", {
    accountCode,
    role: roleInput.value.trim() || undefined,
    username: displayName,
  });

  searchInput.value = "";
  selectedUser.value = null;
  roleInput.value = "";
  searchResults.value = [];
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
        <!-- User search with dropdown -->
        <div class="add-form-field add-form-field--search">
          <div class="search-wrapper">
            <input
              v-model="searchInput"
              type="text"
              class="add-form-field__input"
              :class="{ 'is-selected': selectedUser }"
              :placeholder="t('workflow.settings.accountCodePlaceholder', 'Search user (e.g. ben) or type account code')"
              :disabled="!!selectedUser"
              @input="onSearchInput"
              @blur="handleBlur"
              @keydown.enter.prevent="handleAdd"
            />
            <button
              v-if="selectedUser"
              type="button"
              class="search-clear-btn"
              aria-label="Clear selected user"
              @click="clearSelected"
            >
              <AppIcon name="close" size="xs" />
            </button>
            <span v-if="searchLoading" class="search-spinner" />
          </div>

          <!-- Dropdown suggestions -->
          <div v-if="showDropdown && searchResults.length > 0" class="search-dropdown">
            <button
              v-for="user in searchResults"
              :key="user.id"
              type="button"
              class="search-dropdown__item"
              @mousedown.prevent="selectUser(user)"
            >
              <span class="search-dropdown__name">{{ user.display_name || user.username }}</span>
              <span class="search-dropdown__code">{{ user.username }}</span>
            </button>
          </div>

          <!-- Selected user info -->
          <p v-if="selectedUser" class="search-selected-hint">
            Account code: <strong>{{ selectedUser.username }}</strong>
          </p>
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

.add-form-field--search {
  position: relative;
  min-width: 200px;
  flex-grow: 2;
}

.search-wrapper {
  position: relative;
  display: flex;
  align-items: center;
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

.add-form-field__input.is-selected {
  border-color: var(--color-primary-400);
  background: var(--color-primary-50);
  padding-right: var(--space-8);
}

.add-form-field__input:focus {
  outline: none;
  border-color: var(--color-primary-500);
}

.search-clear-btn {
  position: absolute;
  right: var(--space-2);
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 2px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.search-clear-btn:hover {
  background: var(--border-subtle);
}

.search-spinner {
  position: absolute;
  right: var(--space-2);
  width: 14px;
  height: 14px;
  border: 2px solid var(--border-subtle);
  border-top-color: var(--color-primary-500);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.search-dropdown {
  position: absolute;
  top: calc(100% + 2px);
  left: 0;
  right: 0;
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  z-index: 50;
  overflow: hidden;
}

.search-dropdown__item {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-2) var(--space-3);
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
  gap: var(--space-2);
}

.search-dropdown__item:hover {
  background: var(--bg-row-hover);
}

.search-dropdown__name {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--text-primary);
}

.search-dropdown__code {
  font-size: 11px;
  font-family: var(--font-mono, ui-monospace, monospace);
  color: var(--text-tertiary);
}

.search-selected-hint {
  margin: var(--space-1) 0 0 0;
  font-size: 11px;
  color: var(--text-secondary);
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
