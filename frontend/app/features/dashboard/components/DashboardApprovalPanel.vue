<script setup lang="ts">
import type { DashboardOverviewPendingApproval } from "../types";

const { t } = useI18n();

defineProps<{
  items: DashboardOverviewPendingApproval[];
}>();
</script>

<template>
  <section class="dashboard-side-panel">
    <div class="dashboard-side-panel__header">
      <h2 class="dashboard-side-panel__title">
        {{ t("dashboardOverview.pendingApprovalsTitle", "Pending Approvals") }}
      </h2>
      <span class="dashboard-side-panel__count">{{ items.length }}</span>
    </div>

    <div class="approval-list">
      <article
        v-for="item in items"
        :key="item.id"
        class="approval-list__item"
        :class="{ 'is-urgent': item.isUrgent }"
      >
        <div class="approval-list__copy">
          <div class="approval-list__contract">{{ item.contractCode }}</div>
          <div class="approval-list__value">{{ item.valueLabel }}</div>
        </div>

        <div
          class="approval-list__due"
          :class="{ 'is-urgent': item.isUrgent }"
        >
          <AppIcon v-if="item.isUrgent" name="warning" size="xs" />
          <span>{{ item.dueLabel }}</span>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.dashboard-side-panel {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-xl);
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.dashboard-side-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-5) var(--space-5) var(--space-4);
}

.dashboard-side-panel__title {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  letter-spacing: -0.02em;
}

.dashboard-side-panel__count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.75rem;
  height: 1.75rem;
  padding: 0 var(--space-2);
  border-radius: 999px;
  background: var(--alert-danger-bg);
  color: var(--state-danger);
  font-size: 11px;
  font-weight: var(--font-weight-semibold);
}

.approval-list {
  display: grid;
  gap: var(--space-3);
  padding: 0 var(--space-5) var(--space-5);
}

.approval-list__item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-3);
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  background: var(--bg-card-muted);
}

.approval-list__item.is-urgent {
  background: var(--alert-warning-bg);
  border-color: var(--alert-warning-border);
}

.approval-list__copy {
  min-width: 0;
}

.approval-list__contract {
  color: var(--text-primary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.approval-list__value {
  margin-top: 2px;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.approval-list__due {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--text-tertiary);
  font-size: 11px;
  white-space: nowrap;
}

.approval-list__due.is-urgent {
  color: var(--state-danger);
}

@media (max-width: 640px) {
  .dashboard-side-panel__header {
    padding: var(--space-4) var(--space-4) var(--space-3);
  }

  .approval-list {
    padding: 0 var(--space-4) var(--space-4);
  }

  .approval-list__item {
    flex-direction: column;
    align-items: stretch;
    gap: var(--space-2);
  }

  .approval-list__due {
    justify-content: flex-start;
  }
}
</style>
