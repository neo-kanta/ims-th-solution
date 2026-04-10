<script setup lang="ts">
import { computed, ref } from "vue";

import {
  formatDashboardHeadlineDate,
  formatDashboardTime,
} from "../lib/dashboard";

interface Props {
  businessDate: string;
  dayStatus: string;
  lastRefresh?: string | null;
  onRefresh?: () => void | Promise<void>;
}

const props = withDefaults(defineProps<Props>(), {
  lastRefresh: null,
  onRefresh: undefined,
});

const { t, locale } = useI18n();
const isRefreshing = ref(false);
const refreshLabel = computed(() =>
  t("dashboardOverview.refreshData", "Refresh Data"),
);

const metadata = computed(
  () =>
    `${t("dashboardOverview.systemName", "Investment Management System")} - ${formatDashboardHeadlineDate(props.businessDate, locale.value)} · ${t("dashboardOverview.dayStatusLabel", "Day Status")}: ${props.dayStatus}`,
);

const refreshTitle = computed(() =>
  props.lastRefresh
    ? `${refreshLabel.value}. ${t("dashboardOverview.lastUpdatedUtc", {
        time: formatDashboardTime(props.lastRefresh, locale.value),
      })}`
    : refreshLabel.value,
);

const handleRefresh = async () => {
  if (isRefreshing.value || !props.onRefresh) {
    return;
  }

  isRefreshing.value = true;

  try {
    await props.onRefresh();
  } finally {
    isRefreshing.value = false;
  }
};
</script>

<template>
  <header class="dashboard-overview-header">
    <div class="dashboard-overview-header__copy">
      <h1 class="dashboard-overview-header__title">
        {{ t("dashboardOverview.title", "Dashboard Overview") }}
      </h1>
      <p class="dashboard-overview-header__meta">{{ metadata }}</p>
    </div>

    <AppButton
      class="dashboard-overview-header__action"
      variant="primary"
      size="sm"
      :loading="isRefreshing"
      :title="refreshTitle"
      @click="handleRefresh"
    >
      <AppIcon name="refresh" size="xs" />
      <span>{{ refreshLabel }}</span>
    </AppButton>
  </header>
</template>

<style scoped>
.dashboard-overview-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.dashboard-overview-header__copy {
  min-width: 0;
}

.dashboard-overview-header__title {
  margin: 0;
  color: var(--text-primary);
  font-size: clamp(1.75rem, 2vw, 2.125rem);
  font-weight: var(--font-weight-bold);
  letter-spacing: -0.03em;
}

.dashboard-overview-header__meta {
  margin: var(--space-2) 0 0;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

@media (max-width: 720px) {
  .dashboard-overview-header {
    flex-direction: column;
    align-items: stretch;
  }

  .dashboard-overview-header__action {
    width: 100%;
    justify-content: center;
  }
}
</style>
