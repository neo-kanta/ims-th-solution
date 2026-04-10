<script setup lang="ts">
import type { DashboardOverviewContract } from "../types";

const { t } = useI18n();

const props = defineProps<{
  contracts: DashboardOverviewContract[];
  canCreateContract: boolean;
  createContractUrl: string;
}>();

const handleCreateContract = () => {
  navigateTo(props.createContractUrl);
};
</script>

<template>
  <section class="contracts-panel">
    <div class="contracts-panel__header">
      <h2 class="contracts-panel__title">
        {{ t("dashboardOverview.contractsTitle", "Investment Contracts") }}
      </h2>

      <AppButton
        v-if="canCreateContract"
        class="contracts-panel__action"
        variant="primary"
        size="xs"
        @click="handleCreateContract"
      >
        <AppIcon name="plus" size="xs" />
        <span>{{ t("dashboardOverview.newContract", "New Contract") }}</span>
      </AppButton>
    </div>

    <div class="contracts-panel__table-wrap">
      <table class="contracts-table">
        <thead>
          <tr>
            <th>{{ t("dashboardOverview.columns.contract", "Contract") }}</th>
            <th>{{ t("dashboardOverview.columns.asset", "Asset") }}</th>
            <th>{{ t("dashboardOverview.columns.value", "Value (B)") }}</th>
            <th>{{ t("dashboardOverview.columns.status", "Status") }}</th>
            <th>{{ t("dashboardOverview.columns.manager", "Manager") }}</th>
            <th>{{ t("dashboardOverview.columns.updated", "Upd.") }}</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="contract in contracts" :key="contract.id">
            <td class="contracts-table__contract">{{ contract.code }}</td>
            <td>{{ contract.assetType }}</td>
            <td class="contracts-table__value">{{ contract.valueLabel }}</td>
            <td>
              <AppBadge :variant="contract.statusTone" size="sm">
                {{ contract.statusLabel }}
              </AppBadge>
            </td>
            <td class="contracts-table__muted">{{ contract.manager }}</td>
            <td class="contracts-table__muted">{{ contract.updatedAt }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.contracts-panel {
  min-height: 25rem;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-xl);
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.contracts-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-5) var(--space-6) var(--space-4);
}

.contracts-panel__title {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
}

.contracts-panel__table-wrap {
  overflow-x: auto;
}

.contracts-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 42rem;
}

.contracts-table th,
.contracts-table td {
  padding: var(--space-4) var(--space-6);
  text-align: left;
  border-top: 1px solid var(--border-subtle);
}

.contracts-table th {
  color: var(--text-tertiary);
  font-size: 11px;
  font-weight: var(--font-weight-semibold);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.contracts-table td {
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  vertical-align: middle;
}

.contracts-table tbody tr:hover {
  background: var(--bg-row-hover);
}

.contracts-table__contract {
  color: var(--action-primary);
  font-weight: var(--font-weight-semibold);
}

.contracts-table__value {
  font-weight: var(--font-weight-medium);
}

.contracts-table__muted {
  color: var(--text-secondary);
}

@media (max-width: 640px) {
  .contracts-panel__header {
    flex-direction: column;
    align-items: stretch;
    padding: var(--space-4);
  }

  .contracts-panel__action {
    width: 100%;
    justify-content: center;
  }

  .contracts-table {
    min-width: 34rem;
  }

  .contracts-table th,
  .contracts-table td {
    padding: var(--space-3) var(--space-4);
  }
}
</style>
