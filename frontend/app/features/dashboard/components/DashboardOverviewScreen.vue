<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "#app";

import DashboardAICommandBar from "./DashboardAICommandBar.vue";
import DashboardLayerSidebar from "./DashboardLayerSidebar.vue";
import DashboardTaskFeed from "./DashboardTaskFeed.vue";
import DashboardOverviewHeader from "./DashboardOverviewHeader.vue";
import DashboardWorkflowRail from "./DashboardWorkflowRail.vue";
import WorkflowDayRail from "~/features/workflow/components/WorkflowDayRail.vue";
import DashboardMetricCard from "./DashboardMetricCard.vue";
import DashboardApprovalPanel from "./DashboardApprovalPanel.vue";
import DashboardActivityPanel from "./DashboardActivityPanel.vue";
import DashboardContractsTable from "./DashboardContractsTable.vue";

import { useDashboardData } from "../composables/useDashboardData";
import { useDashboardTasks } from "../composables/useDashboardTasks";
import { useWorkflowOperations } from "~/features/workflow/composables/useWorkflowOperations";
import type { WorkflowStage } from "~/features/workflow/types";
import type {
  DashboardTodoAction,
  DashboardTodoFilter,
  TaskDTO,
} from "../types";

const { t } = useI18n();
const { payload, fetchDashboardData } = useDashboardData();
const { snapshot, loading: tasksLoading, error: tasksError, fetchTasks } =
  useDashboardTasks();
const authStore = useAuthStore();
const router = useRouter();

const { store: workflowStore } = useWorkflowOperations();

const stageTimestamps = computed<
  Partial<Record<WorkflowStage, string | null>>
>(() => ({
  DAY_OPEN: workflowStore.state?.openedAt ?? null,
  MANAGER_APPROVED: workflowStore.state?.managerApprovedAt ?? null,
  TRANSACTION_CLOSED: workflowStore.state?.transactionClosedAt ?? null,
  ACCOUNTING_CLOSED: workflowStore.state?.accountingClosedAt ?? null,
}));

const isRefreshing = ref(false);
const toastVisible = ref(false);
const toastMessage = ref("");
const activeFilter = ref<DashboardTodoFilter>("my");
const isSidebarOpenOnMobile = ref(false);

watch(activeFilter, () => {
  isSidebarOpenOnMobile.value = false;
});

const businessDate = computed(
  () => payload.value?.workflow.businessDate || new Date().toISOString().slice(0, 10),
);

const lastRefreshSource = computed(
  () => snapshot.value?.lastRefreshed || payload.value?.lastRefresh || "",
);

const taskSourceLabel = computed(() => {
  if (tasksLoading.value) return "Tasks: loading integration source";
  if (tasksError.value) return "Tasks: API source unavailable";
  if (snapshot.value) return "Tasks: integration API source";
  return "Tasks: waiting for integration source";
});

const todoTotal = computed(() => snapshot.value?.summary.total ?? 0);

// High-fidelity operational metrics fallback
const displayMetrics = computed(() => {
  if (payload.value?.metrics?.length) {
    return payload.value.metrics;
  }
  return [
    { label: "AUM Today", value: "฿14.24B", changeLabel: "+2.4%", changeTone: "success", helperText: "Across 12 contracts", icon: "portfolio", tone: "primary" },
    { label: "Active Contracts", value: "12", changeLabel: "+1", changeTone: "success", helperText: "3 pending decisions", icon: "decision", tone: "info" },
    { label: "Pending Approvals", value: "2", changeLabel: "Due before 16:30", changeTone: "danger", helperText: "2 urgent approvals", icon: "approval", tone: "danger" },
    { label: "Today's P&L", value: "+฿5.82M", changeLabel: "+0.38%", changeTone: "success", helperText: "vs prior close", icon: "analysis", tone: "success" }
  ];
});

// Real workflow day state for the operator's primary contract, sourced from
// the integration dashboard snapshot. Drives the live WorkflowDayRail; when
// no contract state is available we fall back to the illustrative rail below.
const primaryWorkflowState = computed(() => {
  const states = snapshot.value?.workflowStates ?? [];
  return states.length > 0 ? states[0] : null;
});

// High-fidelity workflow stage fallback
const displayWorkflowStages = computed(() => {
  if (payload.value?.workflow?.stages?.length) {
    return payload.value.workflow.stages;
  }
  return [
    { id: "1", sequence: 1, label: "Day Start", status: "complete", timeLabel: "08:30" },
    { id: "2", sequence: 2, label: "Analysis", status: "complete", timeLabel: "09:15" },
    { id: "3", sequence: 3, label: "Decision", status: "complete", timeLabel: "10:30" },
    { id: "4", sequence: 4, label: "Mgr Approval", status: "active", timeLabel: "11:45" },
    { id: "5", sequence: 5, label: "Execution", status: "upcoming", timeLabel: "" },
    { id: "6", sequence: 6, label: "Tx Closing", status: "upcoming", timeLabel: "" },
    { id: "7", sequence: 7, label: "Acctg Closing", status: "upcoming", timeLabel: "" }
  ];
});

// High-fidelity pending approvals fallback
const displayPendingApprovals = computed(() => {
  if (payload.value?.pendingApprovals?.length) {
    return payload.value.pendingApprovals;
  }
  return [
    { id: "1", contractCode: "KBANK-EQ-001", valueLabel: "฿150M · Buy Order", dueLabel: "20m left", isUrgent: true },
    { id: "2", contractCode: "PTT-FI-002", valueLabel: "฿320M · Bond Issue", dueLabel: "1h 45m left", isUrgent: false },
    { id: "3", contractCode: "SCC-EQ-004", valueLabel: "฿75M · Sell Order", dueLabel: "2h left", isUrgent: false }
  ];
});

// High-fidelity activity log feed fallback
const displayActivityFeed = computed(() => {
  if (payload.value?.activityFeed?.length) {
    return payload.value.activityFeed;
  }
  return [
    { id: "1", actor: "IRG Ops", message: "cleared overnight controls", timeLabel: "10m ago", tone: "success" },
    { id: "2", actor: "neo-kanta", message: "approved analysis for KBANK-EQ-001", timeLabel: "25m ago", tone: "info" },
    { id: "3", actor: "System", message: "auto-escalated PTT-FI-002", timeLabel: "1h ago", tone: "warning" },
    { id: "4", actor: "admin", message: "executed trade for SCC-EQ-004", timeLabel: "2h ago", tone: "teal" }
  ];
});

// High-fidelity contracts list fallback
const displayContracts = computed(() => {
  if (payload.value?.contracts?.length) {
    return payload.value.contracts;
  }
  return [
    { id: "1", code: "KBANK-EQ-001", assetType: "Equity", valueLabel: "1.25B", statusLabel: "Active", statusTone: "success", manager: "neo-kanta", updatedAt: "10m ago" },
    { id: "2", code: "PTT-FI-002", assetType: "Fixed Income", valueLabel: "3.40B", statusLabel: "Review", statusTone: "warning", manager: "system", updatedAt: "1h ago" },
    { id: "3", code: "SCC-EQ-004", assetType: "Equity", valueLabel: "0.85B", statusLabel: "Active", statusTone: "success", manager: "admin", updatedAt: "2h ago" }
  ];
});

async function refreshDashboard() {
  if (isRefreshing.value) return;

  isRefreshing.value = true;
  try {
    await fetchDashboardData();
    await Promise.all([
      fetchTasks(),
      workflowStore.fetchStateByDate(businessDate.value),
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
  // Hand the typed question off to the wired chat page, which auto-sends the
  // `q` query param on load and streams a grounded answer.
  const trimmed = (query ?? "").trim();
  void router.push(trimmed ? { path: "/chat", query: { q: trimmed } } : { path: "/chat" });
}

function onTaskAction(action: DashboardTodoAction, _task: TaskDTO) {
  if (action === "more") {
    showToast("Task action menu is not connected yet.");
    return;
  }

  showToast("Task actions are not connected yet.");
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
      v-model:active-filter="activeFilter"
      class="dashboard-layered__sidebar"
      :snapshot="snapshot"
      :loading="tasksLoading"
      :error="tasksError"
    />

    <main class="dashboard-layered__center">
      <DashboardOverviewHeader
        :business-date="businessDate"
        :day-status="payload?.workflow?.dayStatus || 'Open'"
        :last-refresh="lastRefreshSource"
        :on-refresh="refreshDashboard"
      />

      <div class="dashboard-mobile-actions">
        <button
          class="dashboard-mobile-actions__toggle"
          type="button"
          @click="isSidebarOpenOnMobile = true"
        >
          <AppIcon name="list" size="xs" />
          <span>{{ t('dashboardOverview.showLayers', 'Task Layers & Contracts') }}</span>
        </button>
      </div>

      <WorkflowDayRail
        v-if="workflowStore.state"
        :current-state="workflowStore.currentStateCode"
        :timestamps="stageTimestamps"
        :caption="`${workflowStore.contractId || ''} · ${workflowStore.businessDate}`"
      />
      <WorkflowDayRail
        v-else-if="primaryWorkflowState"
        :current-state="primaryWorkflowState.currentState"
        :caption="`${primaryWorkflowState.contractId} · ${primaryWorkflowState.businessDate}`"
      />
      <DashboardWorkflowRail v-else :stages="displayWorkflowStages" />

      <section class="dashboard-metrics-grid">
        <DashboardMetricCard
          v-for="metric in displayMetrics"
          :key="metric.label"
          :metric="metric"
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

      <DashboardContractsTable
        :contracts="displayContracts"
        :can-create-contract="payload?.canCreateContract || false"
        :create-contract-url="payload?.createContractUrl || '/investment/funds'"
      />
    </main>

    <aside class="dashboard-layered__rail" aria-label="Dashboard context">
      <DashboardApprovalPanel :items="displayPendingApprovals" />

      <DashboardActivityPanel :items="displayActivityFeed" />

      <section class="dashboard-rail__panel">
        <h2 class="dashboard-rail__title">System notes</h2>
        <ul class="dashboard-rail__notes">
          <li>
            <span class="dashboard-rail__note-dot is-preview" />
            <span>AI assistant: Preview</span>
          </li>
          <li>
            <span class="dashboard-rail__note-dot" />
            <span>{{ taskSourceLabel }}</span>
          </li>
          <li>
            <span class="dashboard-rail__note-dot is-muted" />
            <span>{{ todoTotal }} task records in current source</span>
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
    width: 280px;
    z-index: var(--z-sidebar);
    background: var(--bg-sidebar);
    border-right: 1px solid var(--border-subtle);
    border-bottom: 0;
    transform: translateX(-100%);
    transition: transform var(--transition-base);
    box-shadow: var(--shadow-lg);
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
</style>
