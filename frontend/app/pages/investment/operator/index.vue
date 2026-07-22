<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "~/composables/useI18n";
import { useAuthStore } from "~/stores/useAuthStore";
import { useFundWorkspace } from "~/features/investment-workspace/composables/useFundWorkspace";
import { todayBangkokIso } from "~/features/my-funds/lib/derive";
import { myFundsApi } from "~/features/my-funds/services/myFundsApi";
import type { ApiWorkflowState } from "~/features/my-funds/types";
import {
  OPERATOR_CATALOG,
  deriveCapability,
  filterOperatorRows,
  type OperatorCatalogRow,
} from "~/features/operator/lib/operatorCatalog";

import AppCard from "~/shared/ui/AppCard.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppBadge from "~/shared/ui/AppBadge.vue";
import AppButton from "~/shared/ui/AppButton.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_VIEW",
});

const { t } = useI18n();
const router = useRouter();
const auth = useAuthStore();

// Fund workspace state
const { funds, activeFund, loadFunds, setActiveFund } = useFundWorkspace();
const activeFundId = ref<string>("");
const workflowState = ref<ApiWorkflowState | null>(null);
const workflowLoading = ref(false);

const businessDate = computed(() => workflowState.value?.businessDate || todayBangkokIso());

// Operations Catalog directory state
const catalogSearchQuery = ref("");

interface OperatorWorkflow extends OperatorCatalogRow {
  statusVariant: "success" | "warning" | "locked";
}

const catalogWorkflows = computed<OperatorWorkflow[]>(() =>
  OPERATOR_CATALOG.map((workflow) => {
    const capability = deriveCapability(workflow, (code) => auth.hasPermission(code));
    const statusLabel =
      capability === "denied"
        ? t("operator.directory.status.permissionRequired")
        : capability === "limited"
          ? t("operator.directory.status.limited")
          : t("operator.status.ready");
    const statusVariant: OperatorWorkflow["statusVariant"] =
      capability === "denied" ? "locked" : capability === "limited" ? "warning" : "success";

    return {
      ...workflow,
      title: t(workflow.titleKey),
      description: t(workflow.descriptionKey),
      capability,
      statusLabel,
      statusVariant,
    };
  }),
);

const filteredCatalog = computed(() => filterOperatorRows(catalogWorkflows.value, catalogSearchQuery.value));

function limitedReason(op: OperatorWorkflow): string {
  return op.limitedReasonKey ? t(op.limitedReasonKey) : "";
}

async function refreshWorkflowState() {
  if (!activeFundId.value) return;
  workflowLoading.value = true;
  try {
    workflowState.value = await myFundsApi.getWorkflowState(activeFundId.value, todayBangkokIso());
  } catch {
    workflowState.value = null;
  } finally {
    workflowLoading.value = false;
  }
}

function onFundChange(event: Event) {
  const target = event.target;
  if (target instanceof HTMLSelectElement && target.value) {
    activeFundId.value = target.value;
    setActiveFund(target.value);
  }
}

function openWorkflow(op: OperatorWorkflow) {
  if (op.capability === "denied") return;
  void router.push(op.routePath);
}

function onRowKeydown(event: KeyboardEvent, op: OperatorWorkflow) {
  if (event.key !== "Enter" && event.key !== " ") return;
  event.preventDefault();
  openWorkflow(op);
}

// Watchers
watch(activeFundId, async (newId) => {
  if (!newId) return;
  await refreshWorkflowState();
});

// Lifecycle hooks
onMounted(async () => {
  await loadFunds();
  if (funds.value.length > 0) {
    activeFundId.value = funds.value[0]?.fund_id || "";
    setActiveFund(activeFundId.value);
  }
});
</script>

<template>
  <section class="operator-page">
    <AppPageHeader
      :title="t('navigation.operatorPage')"
      :description="t('operator.page.description')"
    />

    <!-- Top Fund Selector Control Card -->
    <div class="operator-header">
      <div v-if="activeFund" class="operator-header__info">
        <div class="operator-header__metric">
          <span class="operator-header__metric-label">{{ t("operator.directory.businessDate") }}</span>
          <span class="operator-header__metric-value">{{ businessDate }}</span>
        </div>
        <div class="operator-header__metric">
          <span class="operator-header__metric-label">{{ t("operator.directory.fundStatus") }}</span>
          <span class="operator-header__metric-value">
            <AppBadge :variant="activeFund.status === 'ACTIVE' ? 'success' : 'neutral'">
              {{ activeFund.status === "ACTIVE" ? t("operator.status.active") : t("operator.status.inactive") }}
            </AppBadge>
          </span>
        </div>
      </div>
    </div>

    <!-- Operations Directory Catalog Table -->
    <div class="catalog-card mt">
      <AppCard :title="t('operator.directory.title')" :subtitle="t('operator.directory.subtitle')">
        <div class="controls-row">
          <div class="control-group search-group">
            <label class="control-label" for="operator-workflow-search">{{ t("operator.directory.filterLabel") }}</label>
            <input
              id="operator-workflow-search"
              v-model="catalogSearchQuery"
              type="text"
              :placeholder="t('operator.directory.searchPlaceholder')"
              class="control-input"
            />
          </div>
        </div>

        <div class="table-container mt">
          <table class="operator-table">
            <thead>
              <tr>
                <th width="80">{{ t("operator.directory.columns.code") }}</th>
                <th>{{ t("operator.directory.columns.workflow") }}</th>
                <th>{{ t("operator.directory.columns.description") }}</th>
                <th>{{ t("operator.directory.columns.status") }}</th>
                <th width="100" class="center">{{ t("operator.directory.columns.action") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="op in filteredCatalog"
                :key="op.code"
                class="clickable-row"
                :class="{ 'clickable-row--denied': op.capability === 'denied' }"
                tabindex="0"
                role="button"
                :aria-disabled="op.capability === 'denied'"
                :aria-label="`${op.code} ${op.title} — ${op.statusLabel}`"
                @click="openWorkflow(op)"
                @keydown="onRowKeydown($event, op)"
              >
                <td class="mono bold">{{ op.code }}</td>
                <td class="bold">{{ op.title }}</td>
                <td class="text-secondary">
                  {{ op.description }}
                  <p v-if="op.capability === 'limited'" class="limited-note">
                    {{ limitedReason(op) }}
                  </p>
                </td>
                <td>
                  <span
                    class="status-pill"
                    :class="`status-pill--${op.statusVariant}`"
                    :title="op.capability === 'denied' ? t('operator.directory.status.permissionRequiredHint') : ''"
                  >
                    {{ op.statusLabel }}
                  </span>
                </td>
                <td class="center" @click.stop>
                  <AppButton
                    variant="secondary"
                    size="xs"
                    :disabled="op.capability === 'denied'"
                    :title="op.capability === 'denied' ? t('operator.directory.status.permissionRequiredHint') : ''"
                    @click="openWorkflow(op)"
                  >
                    {{ t("operator.actions.open") }}
                  </AppButton>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </AppCard>
    </div>
  </section>
</template>

<style scoped>
.operator-page {
  display: grid;
  gap: var(--space-4);
  max-width: 100%;
}

/* Fund Selector Header Card */
.operator-header {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg, 8px);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  flex-wrap: wrap;
}

.operator-header__info {
  display: flex;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.operator-header__metric {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 110px;
}

.operator-header__metric-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-tertiary);
  letter-spacing: 0.04em;
}

.operator-header__metric-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

/* Operations directory table card */
.catalog-card {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.controls-row {
  display: flex;
  align-items: flex-end;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
  flex-wrap: wrap;
}

.control-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.control-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
}

.control-input {
  height: 32px;
  padding: 0 10px;
  font-size: 13px;
  border: 1px solid var(--border-default);
  border-radius: 6px;
  background: var(--bg-input);
  color: var(--text-primary);
}

.search-group {
  flex-grow: 1;
  max-width: 320px;
}

.control-input {
  width: 100%;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 12px;
}

.status-pill--success {
  background: rgba(26, 127, 55, 0.15);
  color: var(--state-success, #1a7f37);
}

.status-pill--warning {
  background: rgba(217, 119, 6, 0.15);
  color: var(--state-warning, #9a6700);
}

.status-pill--locked {
  background: rgba(100, 116, 139, 0.15);
  color: var(--text-tertiary, #64748b);
}

.limited-note {
  margin: 4px 0 0;
  font-size: 11px;
  color: var(--text-tertiary);
  line-height: 1.4;
}

/* Table styling */
.table-container {
  overflow-x: auto;
  border-radius: var(--radius-md);
}

.operator-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.operator-table th,
.operator-table td {
  padding: 8px 12px;
  text-align: left;
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}

.operator-table td.text-secondary {
  white-space: normal;
}

.operator-table th {
  background: var(--bg-card-muted);
  font-size: 11px;
  text-transform: uppercase;
  color: var(--text-tertiary);
  font-weight: 600;
}

.operator-table td.center,
.operator-table th.center {
  text-align: center;
}

.operator-table td.bold {
  font-weight: 600;
}

.clickable-row {
  cursor: pointer;
  transition: background-color 0.1s ease;
}

.clickable-row:hover {
  background: var(--bg-card-hover);
}

.clickable-row:focus-visible {
  outline: 2px solid var(--border-focus, #0969da);
  outline-offset: -2px;
}

.clickable-row--denied {
  cursor: not-allowed;
}

.clickable-row--denied:hover {
  background: none;
}

/* Helpers */
.bold { font-weight: 600; }
.mono { font-family: monospace; }
.mt { margin-top: 12px; }
</style>
