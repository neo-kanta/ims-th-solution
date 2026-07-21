<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "#app";

import DashboardAICommandBar from "./DashboardAICommandBar.vue";
import DashboardLayerSidebar from "./DashboardLayerSidebar.vue";
import DashboardTaskFeed from "./DashboardTaskFeed.vue";
import DashboardOverviewHeader from "./DashboardOverviewHeader.vue";
import DashboardWorkflowRail from "./DashboardWorkflowRail.vue";
import DashboardMetricCard from "./DashboardMetricCard.vue";
import DashboardApprovalPanel from "./DashboardApprovalPanel.vue";
import DashboardActivityPanel from "./DashboardActivityPanel.vue";

import { useDashboardData } from "../composables/useDashboardData";
import { useDashboardTasks } from "../composables/useDashboardTasks";
import { useDashboardApprovals } from "../composables/useDashboardApprovals";
import { useDashboardCounts } from "../composables/useDashboardCounts";
import { useDashboardValuationSummary } from "../composables/useDashboardValuationSummary";
import { useWorkflowDaily } from "~/features/workflow/composables/useWorkflowDaily";
import {
  buildAumMetric,
  buildPnlMetric,
  buildScopeOptions,
  defaultAumScope,
  formatDashboardHeadlineDate,
} from "../lib/dashboard";
import type {
  DashboardOverviewMetric,
  DashboardTodoAction,
  DashboardTodoFilter,
  TaskDTO,
  ValuationScope,
} from "../types";

type DisplayMetric = DashboardOverviewMetric & { loading: boolean };

const { t, locale } = useI18n();
const { payload, fetchDashboardData } = useDashboardData();
const {
  snapshot,
  loading: tasksLoading,
  error: tasksError,
  fetchTasks,
} = useDashboardTasks();
const router = useRouter();

const {
  dailyState: workflowState,
  businessDate: workflowBusinessDate,
  fetchState: refreshWorkflowState,
} = useWorkflowDaily();

const {
  items: pendingApprovals,
  total: approvalsTotal,
  loading: approvalsLoading,
  error: approvalsError,
  fetchApprovals,
} = useDashboardApprovals();

const {
  activeContractsCount,
  pendingApprovalsCount,
  loading: countsLoading,
  fetchCounts,
} = useDashboardCounts();

const {
  summary: valuationSummary,
  loading: valuationLoading,
  error: valuationError,
  fetchValuationSummary,
} = useDashboardValuationSummary();

// Company AUM is an authenticated company-wide metric for every dashboard
// user. Fund/portfolio permissions still govern drill-down screens and the
// separate "My AUM" view; they do not hide or narrow this aggregate.
const aumScope = ref<ValuationScope>(defaultAumScope());

const scopeSelectOptions = computed(() =>
  buildScopeOptions(t).map((option) => ({
    value: option.key,
    label: option.label,
  })),
);

function onScopeChange(scope: string) {
  aumScope.value = scope === "company" ? "company" : "mine";
  void fetchValuationSummary(aumScope.value);
}

// Two-way adapter for AppSelect's v-model — keeps onScopeChange as the single
// place that also triggers the refetch, so selecting from the combo box and
// any future programmatic scope change both go through the same path.
const aumScopeSelectValue = computed<string>({
  get: () => aumScope.value,
  set: (value) => onScopeChange(value),
});

const isRefreshing = ref(false);
const toastVisible = ref(false);
const toastMessage = ref("");
const activeFilter = ref<DashboardTodoFilter>("my");
const isSidebarOpenOnMobile = ref(false);

watch(activeFilter, () => {
  isSidebarOpenOnMobile.value = false;
});

const businessDate = computed(
  () =>
    workflowBusinessDate.value ||
    payload.value?.workflow.businessDate ||
    new Date().toISOString().slice(0, 10),
);

const businessDateLabel = computed(() =>
  formatDashboardHeadlineDate(businessDate.value),
);

const lastRefreshSource = computed(
  () => snapshot.value?.lastRefreshed || payload.value?.lastRefresh || "",
);

const taskSourceLabel = computed(() => {
  if (tasksLoading.value) return t("dashboardOverview.taskSourceLoading");
  if (tasksError.value) return t("dashboardOverview.taskSourceUnavailable");
  if (snapshot.value) return t("dashboardOverview.taskSourceApi");
  return t("dashboardOverview.taskSourceWaiting");
});

const todoTotal = computed(() => snapshot.value?.summary.total ?? 0);

// AUM Today / Today's P&L: derivation (value formatting, sign/tone, and the
// explicit not-available/load-error states) lives in lib/dashboard.ts as
// pure functions so it can be unit tested without mounting this screen.
const aumMetric = computed(() => ({
  ...buildAumMetric(valuationSummary.value, Boolean(valuationError.value), t, locale.value),
  loading: valuationLoading.value,
}));
const pnlMetric = computed(() => ({
  ...buildPnlMetric(valuationSummary.value, Boolean(valuationError.value), t, locale.value),
  loading: valuationLoading.value,
}));

const displayMetrics = computed<DisplayMetric[]>(() => [
  aumMetric.value,
  {
    id: "contracts",
    loading: countsLoading.value,
    label: t("dashboardOverview.metricContractsLabel"),
    value: String(activeContractsCount.value),
    changeLabel: "",
    changeTone: "neutral" as const,
    helperText: t("dashboardOverview.metricContractsHelper", { count: activeContractsCount.value }),
    icon: "decision",
    tone: "info" as const,
  },
  {
    id: "approvals",
    loading: countsLoading.value,
    label: t("dashboardOverview.metricApprovalsLabel"),
    value: String(pendingApprovalsCount.value),
    changeLabel: pendingApprovalsCount.value > 0 ? t("dashboardOverview.actionRequired") : "",
    changeTone: pendingApprovalsCount.value > 0 ? ("danger" as const) : ("neutral" as const),
    helperText: t("dashboardOverview.metricApprovalsHelper", { count: pendingApprovalsCount.value }),
    icon: "approval",
    tone: pendingApprovalsCount.value > 0 ? ("danger" as const) : ("info" as const),
  },
  pnlMetric.value,
]);

const workflowTimestamps = computed(() => {
  const ws = workflowState.value;
  const result: Record<string, string | null> = {
    DAY_OPEN: null,
    MANAGER_APPROVED: null,
    TRANSACTION_CLOSED: null,
    ACCOUNTING_CLOSED: null,
  };
  if (ws?.timeline) {
    for (const entry of ws.timeline) {
      const toState = entry.toState;
      if (toState === "INVESTMENT_DAY_STARTED" || toState === "DAY_OPEN") {
        result.DAY_OPEN = entry.executedAt || null;
      } else if (toState === "MANAGER_APPROVED" || toState === "MANAGER_APPROVED_END_OF_DAY") {
        result.MANAGER_APPROVED = entry.executedAt || null;
      } else if (toState === "TRANSACTION_CLOSED") {
        result.TRANSACTION_CLOSED = entry.executedAt || null;
      } else if (toState === "ACCOUNTING_CLOSED") {
        result.ACCOUNTING_CLOSED = entry.executedAt || null;
      }
    }
  }
  return result;
});


async function refreshDashboard() {
  if (isRefreshing.value) return;

  isRefreshing.value = true;
  try {
    await Promise.all([
      fetchDashboardData(),
      fetchTasks(),
      refreshWorkflowState(),
      fetchApprovals(),
      fetchCounts(),
      fetchValuationSummary(aumScope.value),
    ]);
  } finally {
    isRefreshing.value = false;
  }
}

function showToast(message: string) {
  toastMessage.value = message;
  toastVisible.value = true;
}

function onAssistantPreviewSubmit(query: string) {
  const trimmed = (query ?? "").trim();
  void router.push(
    trimmed ? { path: "/chat", query: { q: trimmed } } : { path: "/chat" },
  );
}

function onTaskAction(action: DashboardTodoAction, _task: TaskDTO) {
  if (action === "more") {
    showToast(t("dashboardOverview.taskActionMenuNotConnected"));
    return;
  }

  showToast(t("dashboardOverview.taskActionsNotConnected"));
}

onMounted(() => {
  void refreshDashboard();
});
</script>

<template>
  <div
    class="dashboard-layered"
    :class="{ 'is-mobile-sidebar-open': isSidebarOpenOnMobile }"
  >
    <div
      v-if="isSidebarOpenOnMobile"
      class="dashboard-layered__sidebar-backdrop"
      @click="isSidebarOpenOnMobile = false"
    />

    <DashboardLayerSidebar
      id="dashboard-layer-sidebar"
      v-model:active-filter="activeFilter"
      class="dashboard-layered__sidebar"
      :snapshot="snapshot"
      :loading="tasksLoading"
      :error="tasksError"
    />

    <main class="dashboard-layered__center">
      <DashboardOverviewHeader
        :business-date="businessDate"
        :day-status="workflowState?.stateLabel || payload?.workflow?.dayStatus || 'Open'"
        :last-refresh="lastRefreshSource"
        :on-refresh="refreshDashboard"
      />

      <div class="dashboard-mobile-actions">
        <button
          class="dashboard-mobile-actions__toggle"
          type="button"
          aria-controls="dashboard-layer-sidebar"
          :aria-expanded="isSidebarOpenOnMobile"
          @click="isSidebarOpenOnMobile = true"
        >
          <AppIcon name="list" size="xs" />
          <span>{{
            t("dashboardOverview.showLayers")
          }}</span>
        </button>
      </div>

      <DashboardWorkflowRail
        :current-state="workflowState?.currentState || 'NOT_STARTED'"
        :timestamps="workflowTimestamps"
      />

      <div v-if="scopeSelectOptions.length > 1" class="dashboard-scope-row">
        <label class="dashboard-scope-row__label" for="dashboard-aum-scope-select">{{ t("dashboardOverview.scopeSelectorLabel") }}</label>
        <AppSelect
          id="dashboard-aum-scope-select"
          v-model="aumScopeSelectValue"
          :options="scopeSelectOptions"
          placeholder=""
          :aria-label="t('dashboardOverview.scopeSelectorLabel')"
          class="dashboard-scope-row__select"
        />
      </div>

      <section class="dashboard-metrics-grid">
        <DashboardMetricCard
          v-for="metric in displayMetrics"
          :key="metric.id"
          :metric="metric"
          :loading="metric.loading"
        />
      </section>

      <DashboardAICommandBar @preview-submit="onAssistantPreviewSubmit" />

      <DashboardTaskFeed
        :snapshot="snapshot"
        :loading="tasksLoading"
        :error="tasksError"
        :active-filter="activeFilter"
        @retry="fetchTasks"
        @task-action="onTaskAction"
      />

    </main>

    <aside class="dashboard-layered__rail" :aria-label="t('dashboardOverview.contextLabel')">
      <DashboardApprovalPanel
        :items="pendingApprovals"
        :loading="approvalsLoading"
        :error="approvalsError"
      />

      <DashboardActivityPanel :items="[]" :loading="false" />

      <section class="dashboard-rail__panel">
        <h2 class="dashboard-rail__title">{{ t("dashboardOverview.systemNotes") }}</h2>
        <ul class="dashboard-rail__notes">
          <li>
            <span class="dashboard-rail__note-dot is-preview" />
            <span>{{ t("dashboardOverview.aiAssistantPreview") }}</span>
          </li>
          <li>
            <span class="dashboard-rail__note-dot" />
            <span>{{ taskSourceLabel }}</span>
          </li>
          <li>
            <span class="dashboard-rail__note-dot is-muted" />
            <span>{{ t("dashboardOverview.taskRecordsCount", { count: todoTotal }) }}</span>
          </li>
        </ul>
      </section>
    </aside>

    <AppToast
      v-model="toastVisible"
      tone="info"
      :message="toastMessage"
      :duration="3500"
    />
  </div>
</template>

<style scoped>
:global(.app-content:has(.dashboard-layered)) {
  padding: 0 !important;
}

:global(.page-container:has(.dashboard-layered)) {
  max-width: 100% !important;
}

.dashboard-layered {
  display: grid;
  grid-template-columns: 320px minmax(0, 1fr) 340px;
  min-height: calc(100vh - var(--header-height));
  width: 100%;
  background: var(--bg-app);
}

.dashboard-layered__sidebar {
  grid-column: 1;
  grid-row: 1;
  position: sticky;
  top: 0;
  height: calc(100vh - var(--header-height));
  overflow-y: auto;
}

.dashboard-layered__center {
  grid-column: 2;
  grid-row: 1;
  padding: var(--space-8) var(--space-6);
  display: grid;
  gap: var(--space-6);
  min-width: 0;
}

.dashboard-layered__rail {
  grid-column: 3;
  grid-row: 1;
  padding: var(--space-8) var(--space-6);
  border-left: 1px solid var(--border-subtle);
  background: var(--bg-app);
  position: sticky;
  top: 0;
  height: calc(100vh - var(--header-height));
  overflow-y: auto;
  display: grid;
  gap: var(--space-5);
  align-content: start;
}

.dashboard-scope-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.dashboard-scope-row__label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  white-space: nowrap;
}

.dashboard-scope-row__select {
  width: auto;
  min-width: 12rem;
  max-width: 16rem;
}

.dashboard-metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
}

.dashboard-rail__panel {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.dashboard-rail__title {
  margin: 0;
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.dashboard-rail__notes {
  display: grid;
  gap: var(--space-3);
  margin: 0;
  padding: 0;
  list-style: none;
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.dashboard-rail__notes li {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-2);
  align-items: start;
}

.dashboard-rail__note-dot {
  width: 8px;
  height: 8px;
  margin-top: 6px;
  border-radius: 50%;
  background: var(--state-info);
}

.dashboard-rail__note-dot.is-preview {
  background: var(--state-warning);
}

.dashboard-rail__note-dot.is-muted {
  background: var(--text-tertiary);
}

@media (max-width: 1440px) {
  .dashboard-metrics-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .dashboard-layered {
    grid-template-columns: 280px minmax(0, 1fr);
  }

  .dashboard-layered__sidebar {
    grid-column: 1;
    grid-row: 1 / span 2;
    height: 100%;
  }

  .dashboard-layered__center {
    grid-column: 2;
    grid-row: 1;
  }

  .dashboard-layered__rail {
    grid-column: 2;
    grid-row: 2;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    position: static;
    border-left: 0;
    border-top: 1px solid var(--border-subtle);
    height: auto;
    padding: var(--space-6) var(--space-4);
  }
}

.dashboard-mobile-actions {
  display: none;
}

@media (max-width: 900px) {
  .dashboard-mobile-actions {
    display: flex;
    justify-content: flex-start;
  }

  .dashboard-mobile-actions__toggle {
    display: inline-flex;
    align-items: center;
    gap: var(--space-2);
    height: 32px;
    padding: 0 var(--space-3);
    color: var(--text-secondary);
    background: var(--action-secondary);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-semibold);
    transition:
      background var(--transition-fast),
      border-color var(--transition-fast);
  }

  .dashboard-mobile-actions__toggle:hover {
    color: var(--text-primary);
    background: var(--bg-row-hover);
    border-color: var(--border-default);
  }

  .dashboard-layered {
    grid-template-columns: 1fr;
    gap: var(--space-4);
  }

  .dashboard-layered__sidebar {
    position: fixed;
    top: var(--header-height);
    left: 0;
    bottom: 0;
    width: min(320px, calc(100vw - var(--space-6)));
    max-width: 100%;
    height: auto;
    z-index: var(--z-sidebar);
    background: var(--bg-sidebar);
    border-right: 1px solid var(--border-subtle);
    border-bottom: 0;
    transform: translateX(-100%);
    transition: transform var(--transition-base);
    box-shadow: var(--shadow-lg);
    overscroll-behavior: contain;
  }

  .is-mobile-sidebar-open .dashboard-layered__sidebar {
    transform: translateX(0);
  }

  .dashboard-layered__sidebar-backdrop {
    position: fixed;
    inset: var(--header-height) 0 0 0;
    background: var(--bg-overlay);
    z-index: calc(var(--z-sidebar) - 1);
  }

  .dashboard-layered__rail {
    grid-column: auto;
    position: static;
    height: auto;
    grid-template-columns: 1fr;
    border-top: 1px solid var(--border-subtle);
    padding: var(--space-6) var(--space-4);
  }

  .dashboard-layered__center {
    padding: var(--space-6) var(--space-4);
  }
}

@media (max-width: 640px) {
  .dashboard-metrics-grid {
    grid-template-columns: 1fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .dashboard-layered__sidebar {
    transition: none;
  }
}
</style>
