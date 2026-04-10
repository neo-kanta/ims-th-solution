<script setup lang="ts">
import { computed, ref } from "vue";

import DashboardActivityPanel from "./DashboardActivityPanel.vue";
import DashboardApprovalPanel from "./DashboardApprovalPanel.vue";
import DashboardContractsTable from "./DashboardContractsTable.vue";
import DashboardMetricCard from "./DashboardMetricCard.vue";
import DashboardOverviewHeader from "./DashboardOverviewHeader.vue";
import DashboardWorkflowRail from "./DashboardWorkflowRail.vue";
import { useDashboardData } from "../composables/useDashboardData";

const { t } = useI18n();
const { payload, loading, error, fetchDashboardData } = useDashboardData();
const isRefreshing = ref(false);

const hasDashboardData = computed(() => payload.value !== null);

const refreshDashboard = async () => {
  if (isRefreshing.value) {
    return;
  }

  isRefreshing.value = true;

  try {
    await fetchDashboardData();
  } finally {
    isRefreshing.value = false;
  }
};

if (!payload.value) {
  await fetchDashboardData();
}
</script>

<template>
  <div class="dashboard-overview">
    <template v-if="hasDashboardData && payload">
      <DashboardOverviewHeader
        :business-date="payload.workflow.businessDate"
        :day-status="payload.workflow.dayStatus"
        :last-refresh="payload.lastRefresh"
        :on-refresh="refreshDashboard"
      />

      <DashboardWorkflowRail :stages="payload.workflow.stages" />

      <section class="dashboard-overview__metrics">
        <DashboardMetricCard
          v-for="metric in payload.metrics"
          :key="metric.id"
          :metric="metric"
        />
      </section>

      <section class="dashboard-overview__workspace">
        <DashboardContractsTable
          :contracts="payload.contracts"
          :can-create-contract="payload.canCreateContract"
          :create-contract-url="payload.createContractUrl"
        />

        <div class="dashboard-overview__sidebar">
          <DashboardActivityPanel :items="payload.activityFeed" />
          <DashboardApprovalPanel :items="payload.pendingApprovals" />
        </div>
      </section>
    </template>

    <AppCard v-else-if="loading" class="dashboard-state-card" no-border>
      <div class="dashboard-state-panel">
        <div class="dashboard-state-content">
          <AppIcon class="is-spinning" name="refresh" size="lg" />
          <div class="dashboard-state-title">
            {{ t("dashboardOverview.loadingTitle", "Loading dashboard") }}
          </div>
          <p class="dashboard-state-copy">
            {{
              t(
                "dashboardOverview.loadingCopy",
                "Pulling contract, workflow, and approval data for the current session.",
              )
            }}
          </p>
        </div>
      </div>
    </AppCard>

    <AppCard v-else-if="error" class="dashboard-state-card">
      <div class="dashboard-state-panel">
        <div class="dashboard-state-content">
          <AppIcon name="warning" size="lg" />
          <div class="dashboard-state-title">
            {{
              t(
                "dashboardOverview.unavailableTitle",
                "Dashboard unavailable",
              )
            }}
          </div>
          <p class="dashboard-state-copy is-danger">{{ error }}</p>
          <AppButton variant="secondary" size="sm" @click="fetchDashboardData">
            <AppIcon name="refresh" size="xs" />
            <span>{{ t("dashboardOverview.retry", "Retry") }}</span>
          </AppButton>
        </div>
      </div>
    </AppCard>

    <AppCard v-else class="dashboard-state-card">
      <div class="dashboard-state-panel">
        <div class="dashboard-state-content">
          <AppIcon name="info" size="lg" />
          <div class="dashboard-state-title">
            {{ t("dashboardOverview.noDataTitle", "No dashboard data") }}
          </div>
          <p class="dashboard-state-copy">
            {{
              t(
                "dashboardOverview.noDataCopy",
                "The dashboard has no session data to display yet.",
              )
            }}
          </p>
        </div>
      </div>
    </AppCard>
  </div>
</template>

<style scoped>
.dashboard-overview {
  display: grid;
  gap: var(--space-5);
  padding-bottom: var(--space-8);
}

.dashboard-overview__metrics {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
}

.dashboard-overview__workspace {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(18rem, 0.88fr);
  gap: var(--space-4);
  align-items: start;
}

.dashboard-overview__sidebar {
  display: grid;
  gap: var(--space-4);
}

.dashboard-state-card {
  min-height: 18rem;
}

.dashboard-state-panel {
  min-height: 18rem;
  display: grid;
  place-items: center;
  text-align: center;
}

.dashboard-state-content {
  display: grid;
  justify-items: center;
  gap: var(--space-4);
  max-width: 30rem;
}

.dashboard-state-title {
  color: var(--text-primary);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.dashboard-state-copy {
  margin: 0;
  color: var(--text-secondary);
}

.dashboard-state-copy.is-danger {
  color: var(--state-danger);
}

.is-spinning {
  animation: dashboard-spin 0.9s linear infinite;
}

@keyframes dashboard-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1100px) {
  .dashboard-overview__metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 960px) {
  .dashboard-overview__workspace {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .dashboard-overview {
    gap: var(--space-4);
  }
}

@media (max-width: 640px) {
  .dashboard-overview__metrics {
    grid-template-columns: 1fr;
  }
}
</style>
