<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "#imports";
import { useI18n } from "~/composables/useI18n";
import { useWorkflowDaily } from "../composables/useWorkflowDaily";
import WorkflowOverviewTab from "./WorkflowOverviewTab.vue";
import WorkflowAuditTab from "./WorkflowAuditTab.vue";
import WorkflowSettingsTab from "./WorkflowSettingsTab.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

const { t } = useI18n();
const route = useRoute();

const {
  businessDate,
  dailyState,
  loading,
  executing,
  error,
  fetchState,
  execute,
} = useWorkflowDaily();

const activeTab = computed(() => {
  if (route.hash === "#settings") return "settings";
  if (route.hash === "#audit") return "audit";
  return "overview";
});

onMounted(async () => {
  await fetchState();
});
</script>

<template>
  <div class="workflow-page-container">
    <!-- Header with actions / filters -->
    <div class="workflow-header-panel">
      <div class="workflow-header-panel__title-section">
        <h1 class="workflow-header-panel__title">
          {{ t("navigation.workflow", "Workflow") }}
        </h1>
        <p class="workflow-header-panel__subtitle">
          {{ t("workflow.subtitle", "Manage daily workflow operations and day-start procedures.") }}
        </p>
      </div>

      <div class="workflow-header-panel__controls">
        <!-- Business Date Selector -->
        <div class="workflow-control-group">
          <label for="business-date-selector" class="workflow-control-group__label">
            {{ t("holdings.page.businessDate" as any, "Business Date") }}
          </label>
          <input
            id="business-date-selector"
            type="date"
            v-model="businessDate"
            class="workflow-control-group__date-input"
          />
        </div>

        <!-- Refresh Button -->
        <AppButton
          variant="secondary"
          size="md"
          class="workflow-header-panel__refresh-btn"
          :loading="loading"
          @click="fetchState"
        >
          <template #icon>
            <AppIcon name="sync" size="sm" />
          </template>
          {{ t("common.actions.refresh" as any, "Refresh") }}
        </AppButton>
      </div>
    </div>

    <!-- Active Tab Display -->
    <div class="workflow-tab-content">
      <div v-if="loading && !dailyState" class="workflow-loading-placeholder">
        <div class="workflow-loading-placeholder__spinner"></div>
        <p class="workflow-loading-placeholder__text">Loading daily state details...</p>
      </div>

      <template v-else>
        <!-- Tab Views -->
        <WorkflowOverviewTab
          v-if="activeTab === 'overview'"
          :business-date="businessDate"
          :daily-state="dailyState"
          :loading="loading"
          :executing="executing"
          :error="error"
          @execute="execute"
          @refresh="fetchState"
        />

        <WorkflowAuditTab
          v-else-if="activeTab === 'audit'"
          :business-date="businessDate"
        />

        <WorkflowSettingsTab
          v-else-if="activeTab === 'settings'"
        />
      </template>
    </div>
  </div>
</template>

<style scoped>
.workflow-page-container {
  display: grid;
  gap: var(--space-6);
  padding-bottom: var(--space-8);
}

.workflow-header-panel {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  flex-wrap: wrap;
  gap: var(--space-5);
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-5);
}

.workflow-header-panel__title-section {
  display: grid;
  gap: var(--space-1);
}

.workflow-header-panel__title {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.workflow-header-panel__subtitle {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.workflow-header-panel__controls {
  display: flex;
  align-items: flex-end;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.workflow-control-group {
  display: grid;
  gap: var(--space-1);
}

.workflow-control-group__label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.workflow-control-group__select,
.workflow-control-group__date-input {
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  background: var(--bg-card);
  color: var(--text-primary);
  height: 38px;
  min-width: 160px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.workflow-control-group__select:focus,
.workflow-control-group__date-input:focus {
  outline: none;
  border-color: var(--color-primary-500);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.15);
}

.workflow-header-panel__refresh-btn {
  height: 38px;
  display: inline-flex;
  align-items: center;
}

.workflow-tab-content {
  min-height: 250px;
}

.workflow-loading-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-10) 0;
  gap: var(--space-4);
}

.workflow-loading-placeholder__spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-subtle);
  border-top-color: var(--color-primary-500);
  border-radius: 50%;
  animation: workflow-spin 1s linear infinite;
}

.workflow-loading-placeholder__text {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

@keyframes workflow-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 768px) {
  .workflow-header-panel {
    flex-direction: column;
    align-items: stretch;
  }
  .workflow-header-panel__controls {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }
  .workflow-header-panel__refresh-btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
