<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import { OpenApiRequestError, unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import { useI18n } from "~/composables/useI18n";
import type { components } from "~/api/ims-api";

import { todayBangkokIso } from "../../lib/derive";
import { relativeTime } from "../../lib/format";
import type { ApiWorkflowState } from "../../types";

type HistoryResponse = components["schemas"]["HistoryResponse"];
type TransitionEntry = components["schemas"]["TransitionEntry"];
type BlockingReason = components["schemas"]["BlockingReason"];

interface Props {
  fundId: string;
}

const props = defineProps<Props>();
const { t } = useI18n();
const client = useOpenApiClient();

const businessDate = ref(todayBangkokIso());
const state = ref<ApiWorkflowState | null>(null);
const history = ref<TransitionEntry[]>([]);
const loadingState = ref(false);
const loadingHistory = ref(false);
const error = ref<string | null>(null);

const allowedActions = computed(() => state.value?.allowedActions ?? []);
const blockingReasons = computed<BlockingReason[]>(() => state.value?.blockingReasons ?? []);

const stageRows = computed(() => [
  {
    key: "opened",
    label: t("myFunds.detail.stagesPanel.dayOpen", "Day opened"),
    timestamp: state.value?.openedAt ?? null,
    actor: state.value?.openedBy ?? null,
  },
  {
    key: "manager-approved",
    label: t("myFunds.detail.stagesPanel.managerApproved", "Manager approved"),
    timestamp: state.value?.managerApprovedAt ?? null,
    actor: state.value?.managerApprovedBy ?? null,
  },
  {
    key: "tx-closed",
    label: t("myFunds.detail.stagesPanel.txClosed", "Transactions closed"),
    timestamp: state.value?.transactionClosedAt ?? null,
    actor: null,
  },
  {
    key: "acctg-closed",
    label: t("myFunds.detail.stagesPanel.acctgClosed", "Accounting closed"),
    timestamp: state.value?.accountingClosedAt ?? null,
    actor: null,
  },
]);

async function fetchState() {
  loadingState.value = true;
  error.value = null;
  try {
    const response = await client.GET("/workflow/day-states/{contractId}", {
      params: {
        path: { contractId: props.fundId },
        query: { businessDate: businessDate.value },
      },
    });
    if ("error" in response && response.error !== undefined) {
      throw new OpenApiRequestError(response.response, response.error);
    }
    state.value = unwrapOpenApiResponse<ApiWorkflowState>(response);
  } catch (err) {
    if (err instanceof OpenApiRequestError && err.status === 404) {
      state.value = null;
    } else {
      error.value = describeError(err, "Failed to load workflow state.");
    }
  } finally {
    loadingState.value = false;
  }
}

async function fetchHistory() {
  loadingHistory.value = true;
  try {
    const response = await client.GET(
      "/workflow/day-states/{contractId}/history",
      {
        params: {
          path: { contractId: props.fundId },
          query: { businessDate: businessDate.value },
        },
      },
    );
    if ("error" in response && response.error !== undefined) {
      history.value = [];
      return;
    }
    const body = unwrapOpenApiResponse<HistoryResponse>(response);
    history.value = body.transitions ?? [];
  } catch {
    history.value = [];
  } finally {
    loadingHistory.value = false;
  }
}

onMounted(async () => {
  await Promise.all([fetchState(), fetchHistory()]);
});

function describeError(err: unknown, fallback: string): string {
  if (err instanceof OpenApiRequestError) return err.message;
  if (err instanceof Error) return err.message;
  return fallback;
}
</script>

<template>
  <div class="stages-panel">
    <AppCard
      :title="t('myFunds.detail.stagesPanel.title', 'Workflow timeline')"
      :subtitle="
        t(
          'myFunds.detail.stagesPanel.subtitle',
          { date: businessDate },
          `Business date ${businessDate}`,
        )
      "
    >
      <div v-if="loadingState && !state" class="stages-panel__loading">
        {{ t("myFunds.detail.stagesPanel.loading", "Loading workflow…") }}
      </div>

      <div v-else-if="error" class="stages-panel__error" role="alert">
        {{ error }}
      </div>

      <div v-else class="stages-panel__body">
        <div class="stages-panel__current">
          <div class="stages-panel__current-label">
            {{ t("myFunds.detail.stagesPanel.currentState", "Current state") }}
          </div>
          <div class="stages-panel__current-value">
            {{ state?.currentState ?? "NOT_STARTED" }}
          </div>
          <div class="stages-panel__current-hint">
            {{
              state?.persisted
                ? t("myFunds.detail.stagesPanel.persisted", "Persisted")
                : t("myFunds.detail.stagesPanel.synthetic", "Synthetic — no row yet")
            }}
          </div>
        </div>

        <ol class="stages-panel__list">
          <li
            v-for="row in stageRows"
            :key="row.key"
            class="stages-panel__row"
            :data-done="!!row.timestamp"
          >
            <span class="stages-panel__row-label">{{ row.label }}</span>
            <span class="stages-panel__row-time">
              {{ row.timestamp ? relativeTime(row.timestamp) : t("myFunds.detail.stagesPanel.pending", "pending") }}
            </span>
          </li>
        </ol>

        <div v-if="allowedActions.length" class="stages-panel__actions">
          <div class="stages-panel__actions-label">
            {{ t("myFunds.detail.stagesPanel.allowedActions", "Allowed actions") }}
          </div>
          <ul>
            <li v-for="action in allowedActions" :key="action">{{ action }}</li>
          </ul>
          <p class="stages-panel__actions-hint">
            {{
              t(
                "myFunds.detail.stagesPanel.allowedActionsHint",
                "Actions are computed server-side. Use the Workflow page for execution.",
              )
            }}
          </p>
        </div>

        <div v-if="blockingReasons.length" class="stages-panel__blockers">
          <div class="stages-panel__blockers-label">
            {{ t("myFunds.detail.stagesPanel.blockers", "Blocking reasons") }}
          </div>
          <ul>
            <li v-for="(reason, i) in blockingReasons" :key="i">
              {{ reason.message ?? reason.code ?? "blocked" }}
            </li>
          </ul>
        </div>
      </div>
    </AppCard>

    <AppCard :title="t('myFunds.detail.stagesPanel.historyTitle', 'Transition history')">
      <div v-if="loadingHistory && history.length === 0" class="stages-panel__loading">
        {{ t("myFunds.detail.stagesPanel.historyLoading", "Loading history…") }}
      </div>
      <div v-else-if="history.length === 0" class="stages-panel__empty">
        {{ t("myFunds.detail.stagesPanel.historyEmpty", "No transitions recorded for this day yet.") }}
      </div>
      <ol v-else class="stages-panel__history">
        <li
          v-for="entry in history"
          :key="entry.id ?? `${entry.occurredAt}-${entry.action}`"
          class="stages-panel__history-row"
        >
          <span class="stages-panel__history-action">{{ entry.action ?? "—" }}</span>
          <span class="stages-panel__history-states">
            {{ entry.fromState ?? "—" }}
            <span aria-hidden="true">→</span>
            {{ entry.toState ?? "—" }}
          </span>
          <span class="stages-panel__history-actor">{{ entry.actorUsername || entry.actorId || "—" }}</span>
          <span class="stages-panel__history-time">{{ relativeTime(entry.occurredAt) }}</span>
        </li>
      </ol>
    </AppCard>
  </div>
</template>

<style scoped>
.stages-panel {
  display: grid;
  gap: var(--space-3);
}

.stages-panel__loading,
.stages-panel__empty {
  font-size: 12px;
  color: var(--text-tertiary, #6e7781);
}

.stages-panel__error {
  font-size: 12px;
  color: var(--state-danger, #cf222e);
}

.stages-panel__body {
  display: grid;
  gap: var(--space-3);
}

.stages-panel__current {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stages-panel__current-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  font-weight: 600;
}

.stages-panel__current-value {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary, #1f2328);
}

.stages-panel__current-hint {
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
}

.stages-panel__list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 4px;
}

.stages-panel__row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  padding: 6px 8px;
  border-radius: 4px;
  background: var(--bg-card-muted, #f6f8fa);
}

.stages-panel__row[data-done="true"] {
  background: rgba(31, 136, 61, 0.08);
}

.stages-panel__row-label {
  font-weight: 500;
}

.stages-panel__row-time {
  color: var(--text-tertiary, #6e7781);
  font-variant-numeric: tabular-nums;
}

.stages-panel__actions,
.stages-panel__blockers {
  font-size: 12px;
  background: var(--bg-card-muted, #f6f8fa);
  border-radius: 5px;
  padding: 8px 10px;
}

.stages-panel__actions-label,
.stages-panel__blockers-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  font-weight: 600;
  margin-bottom: 4px;
}

.stages-panel__actions ul,
.stages-panel__blockers ul {
  list-style: disc;
  padding-left: 16px;
  margin: 0;
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, monospace);
}

.stages-panel__actions-hint {
  margin: 6px 0 0;
  font-size: 11px;
  color: var(--text-tertiary, #6e7781);
  font-family: inherit;
}

.stages-panel__history {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 4px;
}

.stages-panel__history-row {
  display: grid;
  grid-template-columns: 100px minmax(0, 1fr) minmax(0, 1fr) auto;
  gap: var(--space-2);
  font-size: 12px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
  align-items: center;
}

.stages-panel__history-action {
  font-weight: 600;
  color: var(--state-info, #1f6feb);
}

.stages-panel__history-states {
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, monospace);
  color: var(--text-secondary, #57606a);
  font-size: 11px;
}

.stages-panel__history-actor {
  color: var(--text-secondary, #57606a);
}

.stages-panel__history-time {
  color: var(--text-tertiary, #6e7781);
  font-size: 11px;
  text-align: right;
}
</style>
