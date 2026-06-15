<script setup lang="ts">
import { onMounted, watch } from "vue";
import { useI18n } from "~/composables/useI18n";
import { useWorkflowTransitions } from "../composables/useWorkflowTransitions";
import WorkflowAuditTable from "./WorkflowAuditTable.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

const props = defineProps<{
  businessDate: string;
}>();

const { t } = useI18n();

const {
  transitions,
  page,
  pageSize,
  total,
  loading,
  error,
  fetchTransitions,
} = useWorkflowTransitions();

async function loadData() {
  await fetchTransitions(props.businessDate);
}

onMounted(loadData);

watch(
  () => props.businessDate,
  () => {
    page.value = 1;
    loadData();
  },
);

watch(page, loadData);

function prevPage() {
  if (page.value > 1) {
    page.value--;
  }
}

function nextPage() {
  const totalPages = Math.ceil(total.value / pageSize.value);
  if (page.value < totalPages) {
    page.value++;
  }
}
</script>

<template>
  <div class="workflow-audit-tab">
    <!-- Error State -->
    <div v-if="error" class="workflow-error-card">
      <div class="workflow-error-card__content">
        <h3 class="workflow-error-card__title">Failed to load audit trail</h3>
        <p class="workflow-error-card__text">{{ error }}</p>
        <button type="button" class="workflow-error-card__retry" @click="loadData">
          Retry Loading
        </button>
      </div>
    </div>

    <!-- Table Container -->
    <AppCard v-else>
      <div class="workflow-audit-header">
        <h3 class="workflow-audit-title">Transition History Trail</h3>
        <span class="workflow-audit-total">Total transitions: {{ total }}</span>
      </div>

      <div v-if="loading && transitions.length === 0" class="workflow-audit-loading">
        <div class="workflow-audit-loading__spinner"></div>
        <span>Loading history entries...</span>
      </div>

      <template v-else>
        <!-- Table view -->
        <WorkflowAuditTable :transitions="transitions" />

        <!-- Pagination -->
        <div v-if="total > pageSize" class="workflow-pagination">
          <span class="workflow-pagination__status">
            Showing Page {{ page }} of {{ Math.ceil(total / pageSize) }}
          </span>
          <div class="workflow-pagination__buttons">
            <AppButton
              variant="secondary"
              size="sm"
              :disabled="page <= 1 || loading"
              @click="prevPage"
            >
              <template #icon>
                <AppIcon name="chevron-down" size="xs" class="prev-arrow" />
              </template>
              Previous
            </AppButton>
            <AppButton
              variant="secondary"
              size="sm"
              :disabled="page >= Math.ceil(total / pageSize) || loading"
              @click="nextPage"
            >
              Next
              <template #icon>
                <AppIcon name="chevron-down" size="xs" class="next-arrow" />
              </template>
            </AppButton>
          </div>
        </div>
      </template>
    </AppCard>

  </div>
</template>

<style scoped>
.workflow-audit-tab {
  display: grid;
  gap: var(--space-6);
}

.workflow-audit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-3);
}

.workflow-audit-title {
  margin: 0;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.workflow-audit-total {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.workflow-audit-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-10) 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.workflow-audit-loading__spinner {
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

.workflow-pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--space-4);
  border-top: 1px solid var(--border-subtle);
  padding-top: var(--space-4);
}

.workflow-pagination__status {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.workflow-pagination__buttons {
  display: flex;
  gap: var(--space-2);
}

.prev-arrow {
  transform: rotate(90deg);
}

.next-arrow {
  transform: rotate(-90deg);
}

.workflow-audit-empty {
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-lg);
  padding: var(--space-10);
  text-align: center;
  color: var(--text-secondary);
  font-size: var(--font-size-md);
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
</style>
