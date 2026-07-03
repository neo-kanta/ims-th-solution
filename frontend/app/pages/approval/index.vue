<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRouter } from "#imports";
import { useI18n } from "~/composables/useI18n";
import { useAuthStore } from "~/stores/useAuthStore";
import { useAppToast } from "~/composables/useAppToast";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import AppDateField from "~/shared/ui/AppDateField.vue";
import AppTabs from "~/shared/ui/AppTabs.vue";
import AppDataTable from "~/shared/ui/AppDataTable.vue";
import type { TableColumn } from "~/shared/ui/AppDataTable.vue";
import ApprovalStatusBadge from "~/features/approval/components/ApprovalStatusBadge.vue";
import { prettify } from "~/features/approval/lib/approvalStatus";
import { normalizeInboxItem, normalizeRequestItem, type NormalizedRequest } from "~/features/approval/lib/approvalNormalizers";
import { useApprovalInbox } from "~/features/approval/composables/useApprovalInbox";
import { approvalApi } from "~/features/approval/services/approvalApi";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "APPROVAL_VIEW_INBOX",
});

const { t } = useI18n();
const router = useRouter();
const authStore = useAuthStore();
const toast = useAppToast();

const currentUserId = computed(() => authStore.user?.id ?? "");

// Tabs configuration
const activeTab = ref("pending");
const tabsItems = computed(() => {
  const list = [
    { key: "pending", label: t("approval.tabs.pending" as any, "My Pending"), count: pendingInbox.total.value },
    { key: "submitted", label: t("approval.tabs.submitted" as any, "Submitted By Me"), count: submittedCount.value },
    { key: "completed", label: t("approval.tabs.completed" as any, "Completed"), count: completedCount.value },
  ];
  if (authStore.hasPermission("APPROVAL_VIEW_ALL")) {
    list.push({ key: "all", label: t("approval.tabs.all" as any, "All Approval Records"), count: allCount.value });
  }
  return list;
});

// Filters state
const filterModule = ref("");
const filterProcessType = ref("");
const filterStatus = ref("");
const filterDate = ref("");

const MODULE_OPTIONS = [
  { value: "", label: "All Modules" },
  { value: "FUND", label: "Fund" },
  { value: "CONTRACT", label: "Contract" },
  { value: "INVESTMENT", label: "Investment" },
  { value: "LEAVE", label: "Leave" },
  { value: "EXPENSE", label: "Expense" },
];

const PROCESS_TYPE_OPTIONS = [
  { value: "", label: "All Processes" },
  { value: "INVESTMENT_ANALYSIS_REPORT", label: "Investment Analysis Report" },
  { value: "INVESTMENT_DECISION", label: "Investment Decision" },
  { value: "INVESTMENT_CANCELLATION", label: "Investment Cancellation" },
  { value: "WORKFLOW_OPERATION", label: "Workflow Operation" },
  { value: "LEAVE_REQUEST", label: "Leave Request" },
  { value: "LEAVE_CANCELLATION", label: "Leave Cancellation" },
  { value: "DELEGATION_REQUEST", label: "Delegation Request" },
];

const STATUS_OPTIONS = [
  { value: "", label: "All Statuses" },
  { value: "SUBMITTED", label: "Submitted" },
  { value: "PENDING_APPROVAL", label: "Pending Approval" },
  { value: "APPROVED", label: "Approved" },
  { value: "REJECTED", label: "Rejected" },
  { value: "CANCELLED", label: "Cancelled" },
  { value: "REVOKED", label: "Revoked" },
];

// Inbox State ("My Pending")
const pendingInbox = useApprovalInbox();

// Request List States (for other tabs)
const requestList = ref<any[]>([]);
const requestListLoading = ref(false);
const requestListError = ref<string | null>(null);

// Tab Counts for Submitted, Completed, All
const submittedCount = ref<number | null>(null);
const completedCount = ref<number | null>(null);
const allCount = ref<number | null>(null);

// Fetch counts for tabs badges
async function refreshTabCounts() {
  try {
    const res = await approvalApi.requests({ limit: 1000 });
    const items = res.items ?? [];
    
    submittedCount.value = items.filter((x) => x.submitter_id === currentUserId.value).length;
    completedCount.value = items.filter(
      (x) =>
        x.status === "APPROVED" ||
        x.status === "REJECTED" ||
        x.status === "CANCELLED" ||
        x.status === "WITHDRAWN" ||
        x.status === "REVOKED"
    ).length;
    allCount.value = items.length;
  } catch {
    // Ignore counts fetch errors silently
  }
}

// Table Definition
const columns: TableColumn[] = [
  { key: "status", label: "Status", align: "center" },
  { key: "module", label: "Module" },
  { key: "processType", label: "Process Type" },
  { key: "recordType", label: "Record Type" },
  { key: "recordTitle", label: "Record Title" },
  { key: "submitter", label: "Submitter" },
  { key: "currentStage", label: "Current Stage", align: "center" },
  { key: "nextApprover", label: "Next Approver" },
  { key: "submittedAt", label: "Submitted At" },
  { key: "actions", label: "", align: "right" },
];

// Normalize inbox and request models
const normalizedItems = computed<NormalizedRequest[]>(() => {
  if (activeTab.value === "pending") {
    return pendingInbox.items.value.map(normalizeInboxItem);
  }

  // Filter and map requestList
  let filtered = [...requestList.value];

  // Apply API-level filtering that is computed locally
  if (activeTab.value === "submitted") {
    filtered = filtered.filter((x) => x.submitter_id === currentUserId.value);
  } else if (activeTab.value === "completed") {
    filtered = filtered.filter(
      (x) =>
        x.status === "APPROVED" ||
        x.status === "REJECTED" ||
        x.status === "CANCELLED" ||
        x.status === "WITHDRAWN" ||
        x.status === "REVOKED"
    );
  }

  // Apply UI Filters
  if (filterModule.value) {
    filtered = filtered.filter(
      (x) => x.contract_type === filterModule.value || x.contract_id === filterModule.value
    );
  }
  if (filterProcessType.value) {
    filtered = filtered.filter((x) => x.process_type === filterProcessType.value);
  }
  if (filterStatus.value && activeTab.value === "all") {
    filtered = filtered.filter((x) => x.status === filterStatus.value);
  }
  if (filterDate.value) {
    filtered = filtered.filter((x) => {
      const date = x.submitted_at || x.created_at;
      return date && date.startsWith(filterDate.value);
    });
  }

  return filtered.map(normalizeRequestItem);
});

// Main refresh orchestrator
async function refreshData() {
  if (activeTab.value === "pending") {
    // Apply process_type to inbox fetch if filtered
    await pendingInbox.fetchInbox("PENDING");
  } else {
    requestListLoading.value = true;
    requestListError.value = null;
    try {
      let statusParam: string | undefined = undefined;
      if (activeTab.value === "all" && filterStatus.value) {
        statusParam = filterStatus.value;
      }
      
      const payload = await approvalApi.requests({
        status: statusParam,
        process_type: filterProcessType.value || undefined,
        limit: 200,
      });
      requestList.value = payload.items ?? [];
    } catch (err: any) {
      requestListError.value = err.message || "Failed to load requests.";
      requestList.value = [];
    } finally {
      requestListLoading.value = false;
    }
  }
  void refreshTabCounts();
}

function open(requestId: string) {
  if (requestId) void router.push(`/approval/requests/${requestId}`);
}

function resetFilters() {
  filterModule.value = "";
  filterProcessType.value = "";
  filterStatus.value = "";
  filterDate.value = "";
  void refreshData();
}

onMounted(() => {
  void refreshData();
});

watch(activeTab, () => {
  void refreshData();
});

// Watch filters to trigger reload
watch([filterProcessType, filterStatus], () => {
  void refreshData();
});

function fmtDate(value?: string | null): string {
  if (!value) return "—";
  try {
    return new Intl.DateTimeFormat("en-CA", {
      timeZone: "Asia/Bangkok",
      year: "numeric",
      month: "short",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      hourCycle: "h23",
    }).format(new Date(value));
  } catch {
    return value;
  }
}
</script>

<template>
  <section class="approval-dashboard-page">
    <AppPageHeader
      :title="t('approval.inbox.title', 'Approval Center')"
      :description="t('approval.inbox.description', 'Manage and track cross-module approvals.')"
    >
      <template #actions>
        <AppButton
          variant="secondary"
          size="sm"
          :loading="pendingInbox.loading.value || requestListLoading"
          @click="refreshData"
        >
          {{ t('approval.actions.refresh', 'Refresh') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <!-- Tabs row -->
    <AppTabs :items="tabsItems" v-model="activeTab" />

    <!-- Filters Row -->
    <AppCard class="filters-card">
      <div class="filters-row">
        <div class="filter-item">
          <label class="filter-item__label">Module</label>
          <AppSelect v-model="filterModule" :options="MODULE_OPTIONS" />
        </div>
        
        <div class="filter-item">
          <label class="filter-item__label">Process Type</label>
          <AppSelect v-model="filterProcessType" :options="PROCESS_TYPE_OPTIONS" />
        </div>
        
        <div v-if="activeTab === 'all'" class="filter-item">
          <label class="filter-item__label">Status</label>
          <AppSelect v-model="filterStatus" :options="STATUS_OPTIONS" />
        </div>

        <div class="filter-item">
          <label class="filter-item__label">Submitted Date</label>
          <AppDateField v-model="filterDate" />
        </div>

        <div class="filter-actions">
          <AppButton variant="secondary" size="sm" @click="resetFilters">
            Reset
          </AppButton>
        </div>
      </div>
    </AppCard>

    <!-- Error/State display -->
    <AppCard v-if="pendingInbox.forbidden.value">
      <p class="approval-dashboard-page__state">
        {{ t('approval.errors.noPermission' as any, 'You do not have permission to view approval inbox.') }}
      </p>
    </AppCard>

    <!-- Main Table -->
    <template v-else>
      <AppDataTable
        :columns="columns"
        :items="normalizedItems"
        :loading="pendingInbox.loading.value || requestListLoading"
        :error="pendingInbox.error.value || requestListError"
        :empty-text="t('approval.inbox.empty' as any, 'No approval records found matching current criteria.')"
      >
        <template #[`cell(status)`]="{ item }">
          <ApprovalStatusBadge :status="(item as NormalizedRequest).status" />
        </template>
        
        <template #[`cell(module)`]="{ item }">
          <span class="module-label">{{ (item as NormalizedRequest).module }}</span>
        </template>

        <template #[`cell(processType)`]="{ item }">
          {{ prettify((item as NormalizedRequest).processType) }}
        </template>

        <template #[`cell(recordType)`]="{ item }">
          {{ prettify((item as NormalizedRequest).recordType) }}
        </template>

        <template #[`cell(recordTitle)`]="{ item }">
          <div class="title-cell">
            <span class="title-cell__main">{{ (item as NormalizedRequest).recordTitle }}</span>
            <span class="title-cell__sub">{{ (item as NormalizedRequest).requestNumber }}</span>
          </div>
        </template>

        <template #[`cell(submitter)`]="{ item }">
          {{ (item as NormalizedRequest).submitter }}
        </template>

        <template #[`cell(currentStage)`]="{ item }">
          {{ (item as NormalizedRequest).currentStage }}
        </template>

        <template #[`cell(nextApprover)`]="{ item }">
          {{ (item as NormalizedRequest).nextApprover }}
        </template>

        <template #[`cell(submittedAt)`]="{ item }">
          {{ fmtDate((item as NormalizedRequest).submittedAt) }}
        </template>

        <template #[`cell(actions)`]="{ item }">
          <AppButton size="sm" variant="primary" @click="open((item as NormalizedRequest).id)">
            Review
          </AppButton>
        </template>
      </AppDataTable>
    </template>
  </section>
</template>

<style scoped>
.approval-dashboard-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-4, 16px);
}

.filters-card {
  padding: var(--space-3, 12px) var(--space-4, 16px);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md, 6px);
}

.filters-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4, 16px);
  align-items: flex-end;
}

.filter-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-1, 4px);
  flex: 1 1 180px;
  min-width: 140px;
}

.filter-item__label {
  font-size: var(--font-size-xs, 0.75rem);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-secondary);
}

.filter-actions {
  display: flex;
  align-items: center;
  height: 32px;
}

.approval-dashboard-page__state {
  margin: 0;
  color: var(--text-secondary);
}

.module-label {
  font-weight: var(--font-weight-semibold, 600);
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 0.75rem);
  background: var(--bg-card-hover);
  color: var(--text-primary);
  padding: 2px 6px;
  border-radius: var(--radius-sm, 4px);
  border: 1px solid var(--border-subtle);
}

.title-cell {
  display: flex;
  flex-direction: column;
}

.title-cell__main {
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary);
}

.title-cell__sub {
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary);
  font-family: var(--font-mono, monospace);
}
</style>
