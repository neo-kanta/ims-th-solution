<script setup lang="ts">
/**
 * Read-only trade-execution list for a portfolio (Portfolio V2). Reached
 * from the portfolio workspace tabs and from OP-02's approval workbench
 * ("Executions" action per decision row).
 *
 * This view is intentionally read-only: fill/cancel mutation actions are
 * out of this task's scope (only the existing approve/reject decision
 * endpoints are unchanged write paths). See the OP-02 work package report
 * for the exact reasoning.
 */
import { computed, onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import AppCard from "~/shared/ui/AppCard.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import { useI18n } from "~/composables/useI18n";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import { portfolioApi, type ApiExecutionV2 } from "./services/portfolioApi";
import { decisionDetailPath } from "~/features/portfolio-decision/lib/decisionRoutes";
import {
  deriveExecutionLifecycleState,
  executionStateTone,
} from "~/features/investment-decision/lib/executionState";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.executions.title"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true },
);

const ctx = usePortfolioContext(() => props.portfolioCode);

const executions = ref<ApiExecutionV2[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

function executionStateLabel(execution: ApiExecutionV2): string {
  const state = deriveExecutionLifecycleState(null, execution.status);
  switch (state) {
    case "execution_pending":
      return t("portfolio.executions.status.pending");
    case "filled":
      return t("portfolio.executions.status.filled");
    case "partially_filled":
      return t("portfolio.executions.status.partiallyFilled");
    case "execution_cancelled":
      return t("portfolio.executions.status.cancelled");
    default:
      return execution.status ?? t("common.notAvailable");
  }
}

function executionStateBadgeTone(execution: ApiExecutionV2): string {
  return executionStateTone(deriveExecutionLifecycleState(null, execution.status));
}

const hasRows = computed(() => executions.value.length > 0);

async function loadExecutions() {
  if (!props.portfolioCode) return;
  loading.value = true;
  error.value = null;
  try {
    const list = await portfolioApi.listExecutions(props.portfolioCode, { limit: 200 });
    executions.value = list.items ?? [];
  } catch (err) {
    executions.value = [];
    error.value = err instanceof Error ? err.message : t("portfolio.executions.errorTitle");
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void ctx.reload();
  void loadExecutions();
});
watch(
  () => props.portfolioCode,
  () => {
    void ctx.reload();
    void loadExecutions();
  },
);
</script>

<template>
  <section class="portfolio-executions">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <AppCard :title="t('portfolio.executions.title')" :subtitle="t('portfolio.executions.subtitle')">
      <div v-if="loading && executions.length === 0" class="portfolio-executions__notice" role="status">
        {{ t("portfolio.executions.loading") }}
      </div>
      <div v-else-if="error" class="portfolio-executions__error" role="alert">
        <div>{{ t("portfolio.executions.errorTitle") }}: {{ error }}</div>
        <button type="button" class="portfolio-executions__retry" @click="loadExecutions">
          {{ t("portfolio.executions.retry") }}
        </button>
      </div>
      <div v-else-if="!hasRows" class="portfolio-executions__notice">
        {{ t("portfolio.executions.empty") }}
      </div>
      <div v-else class="portfolio-executions__table-wrap">
        <table class="portfolio-executions__table">
          <thead>
            <tr>
              <th>{{ t("portfolio.executions.columns.instrument") }}</th>
              <th>{{ t("portfolio.executions.columns.side") }}</th>
              <th class="right">{{ t("portfolio.executions.columns.orderedQty") }}</th>
              <th class="right">{{ t("portfolio.executions.columns.executedQty") }}</th>
              <th class="right">{{ t("portfolio.executions.columns.executionPrice") }}</th>
              <th>{{ t("portfolio.executions.columns.status") }}</th>
              <th>{{ t("portfolio.executions.columns.businessDate") }}</th>
              <th>{{ t("portfolio.executions.columns.decision") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="execution in executions" :key="execution.id">
              <td>{{ execution.instrument_code ?? t("common.notAvailable") }}</td>
              <td>{{ execution.side ?? t("common.notAvailable") }}</td>
              <td class="right">{{ execution.ordered_quantity ?? "—" }}</td>
              <td class="right">{{ execution.executed_quantity ?? "—" }}</td>
              <td class="right">{{ execution.execution_price ?? "—" }}</td>
              <td>
                <AppStatusBadge
                  :status="executionStateBadgeTone(execution)"
                  :label="executionStateLabel(execution)"
                  size="sm"
                />
              </td>
              <td>{{ execution.business_date ?? "—" }}</td>
              <td>
                <NuxtLink
                  v-if="execution.decision_id"
                  :to="decisionDetailPath(props.portfolioCode, execution.decision_id)"
                >
                  {{ t("portfolio.executions.viewDecision") }}
                </NuxtLink>
                <span v-else>—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </AppCard>
  </section>
</template>

<style scoped>
.portfolio-executions {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-executions__notice,
.portfolio-executions__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-executions__error {
  color: var(--alert-danger-text, #cf222e);
  display: grid;
  gap: 6px;
}

.portfolio-executions__retry {
  justify-self: start;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card, #ffffff);
  border-radius: var(--radius-md, 5px);
  padding: 4px 10px;
  cursor: pointer;
}

.portfolio-executions__table-wrap {
  overflow-x: auto;
}

.portfolio-executions__table {
  width: 100%;
  min-width: 720px;
  border-collapse: collapse;
  font-size: 13px;
}

.portfolio-executions__table th {
  text-align: left;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  white-space: nowrap;
}

.portfolio-executions__table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  white-space: nowrap;
}

.portfolio-executions__table th.right,
.portfolio-executions__table td.right {
  text-align: right;
}
</style>
