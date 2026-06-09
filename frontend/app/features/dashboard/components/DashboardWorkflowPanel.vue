<script setup lang="ts">
/**
 * DashboardWorkflowPanel
 * ------------------------------------------------------------------
 * Operational control panel for the contract workflow day-state machine.
 * Wired to the real backend module (`/api/v1/workflow/day-states/*`) via
 * the generated OpenAPI client; permission codes follow the workflow
 * module's catalog (WORKFLOW_VIEW, WORKFLOW_OPEN_DAY, …).
 *
 * Backend is the source of truth for `allowedActions` and `blockingReasons`.
 * The UI mirrors them — it never re-derives transitions client-side.
 *
 * Layout:
 *   - Contract picker (sourced from /integration/dashboard/me workflowStates)
 *   - Business date picker (YYYY-MM-DD, Asia/Bangkok)
 *   - Four stage cards: Day Start → Mgr Approval → Tx Closing → Acctg Closing
 *   - Operation form, gated by allowedActions + per-action permissions
 *   - Confirmation dialog (stronger wording for cancel/rollback)
 *   - Recent operation history (audit timeline)
 */
import { computed, ref, watch, type DeepReadonly } from "vue";

import AppBadge from "~/shared/ui/AppBadge.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";
import AppEmptyState from "~/shared/ui/AppEmptyState.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

import type { AppTranslationKey } from "~/shared/i18n/messages";
import { useAuthStore } from "~/stores/useAuthStore";

import {
  useDashboardWorkflow,
  type WorkflowExecuteRequest,
  type WorkflowStateResponse,
} from "../composables/useDashboardWorkflow";
import type { WorkflowStateDTO } from "../types";

const props = defineProps<{
  workflowStates: WorkflowStateDTO[];
}>();

const { t } = useI18n();
const authStore = useAuthStore();

const {
  contractId,
  businessDate,
  state,
  history,
  loadingState,
  loadingHistory,
  executing,
  error,
  lastResult,
  hasContract,
  setContract,
  setBusinessDate,
  refresh,
  execute,
  clearLastResult,
} = useDashboardWorkflow();

// ─────────────────────────────────────────────────────────────────────────
// Static config — labels, permission gates, stage→action mapping.
// Backend action codes (OPEN_DAY, APPROVE, …) are wire constants; UI labels
// localise them via i18n.
// ─────────────────────────────────────────────────────────────────────────

type StageKey = "DAY_OPEN" | "MANAGER_APPROVED" | "TRANSACTION_CLOSED" | "ACCOUNTING_CLOSED";

const STAGE_KEYS: StageKey[] = [
  "DAY_OPEN",
  "MANAGER_APPROVED",
  "TRANSACTION_CLOSED",
  "ACCOUNTING_CLOSED",
];

const STAGE_ORDER: Record<string, number> = {
  NOT_STARTED: 0,
  DAY_OPEN: 1,
  MANAGER_APPROVED: 2,
  TRANSACTION_CLOSED: 3,
  ACCOUNTING_CLOSED: 4,
};

interface StageDefinition {
  key: StageKey;
  labelKey: AppTranslationKey;
  labelFallback: string;
}

const STAGES: StageDefinition[] = [
  {
    key: "DAY_OPEN",
    labelKey: "dashboardWorkflow.stageDayStart",
    labelFallback: "Investment Day Start",
  },
  {
    key: "MANAGER_APPROVED",
    labelKey: "dashboardWorkflow.stageManagerApproval",
    labelFallback: "Manager Approval",
  },
  {
    key: "TRANSACTION_CLOSED",
    labelKey: "dashboardWorkflow.stageTransactionClosing",
    labelFallback: "Transaction Closing",
  },
  {
    key: "ACCOUNTING_CLOSED",
    labelKey: "dashboardWorkflow.stageAccountingClosing",
    labelFallback: "Accounting Closing",
  },
];

interface ActionOption {
  value: NonNullable<WorkflowExecuteRequest["action"]>;
  labelKey: AppTranslationKey;
  labelFallback: string;
  permission: string;
  /** `cancel`/`rollback` triggers stronger confirmation wording. */
  tone: "primary" | "warning" | "danger";
  /** When true the operator must supply a reason (matches backend rule). */
  requiresReason: boolean;
}

const ACTIONS: ActionOption[] = [
  {
    value: "OPEN_DAY",
    labelKey: "dashboardWorkflow.actionOpenDay",
    labelFallback: "Investment Day Start",
    permission: "WORKFLOW_OPEN_DAY",
    tone: "primary",
    requiresReason: false,
  },
  {
    value: "APPROVE",
    labelKey: "dashboardWorkflow.actionApprove",
    labelFallback: "Manager Approval",
    permission: "WORKFLOW_APPROVE",
    tone: "primary",
    requiresReason: false,
  },
  {
    value: "CLOSE_TRANSACTIONS",
    labelKey: "dashboardWorkflow.actionCloseTransactions",
    labelFallback: "Transaction Closing",
    permission: "WORKFLOW_CLOSE_TRANSACTIONS",
    tone: "primary",
    requiresReason: false,
  },
  {
    value: "CLOSE_ACCOUNTING",
    labelKey: "dashboardWorkflow.actionCloseAccounting",
    labelFallback: "Accounting Closing",
    permission: "WORKFLOW_CLOSE_ACCOUNTING",
    tone: "primary",
    requiresReason: false,
  },
  {
    value: "CANCEL_DAY_START",
    labelKey: "dashboardWorkflow.actionCancelDayStart",
    labelFallback: "Cancel Investment Day Start",
    permission: "WORKFLOW_CANCEL_DAY_START",
    tone: "warning",
    requiresReason: true,
  },
  {
    value: "CANCEL_APPROVAL",
    labelKey: "dashboardWorkflow.actionCancelApproval",
    labelFallback: "Cancel Manager Approval",
    permission: "WORKFLOW_CANCEL_APPROVAL",
    tone: "warning",
    requiresReason: true,
  },
  {
    value: "CANCEL_TRANSACTION_CLOSE",
    labelKey: "dashboardWorkflow.actionCancelTransactionClose",
    labelFallback: "Cancel Transaction Closing",
    permission: "WORKFLOW_CANCEL_TRANSACTION_CLOSE",
    tone: "warning",
    requiresReason: true,
  },
  {
    value: "ROLLBACK_ACCOUNTING_CLOSE",
    labelKey: "dashboardWorkflow.actionRollbackAccountingClose",
    labelFallback: "Rollback Accounting Closing",
    permission: "WORKFLOW_ROLLBACK_ACCOUNTING_CLOSE",
    tone: "danger",
    requiresReason: true,
  },
];

const ACTION_LOOKUP: Record<string, ActionOption> = Object.fromEntries(
  ACTIONS.map((a) => [a.value, a]),
);

const hasViewPermission = computed(() =>
  authStore.hasPermission("WORKFLOW_VIEW"),
);

// ─────────────────────────────────────────────────────────────────────────
// Contract picker — derived from the integration snapshot.
// ─────────────────────────────────────────────────────────────────────────

const contractSearch = ref("");

const contractOptions = computed(() => {
  const seen = new Map<string, WorkflowStateDTO>();
  for (const row of props.workflowStates) {
    if (row.contractId && !seen.has(row.contractId)) {
      seen.set(row.contractId, row);
    }
  }
  return Array.from(seen.values());
});

const filteredContracts = computed(() => {
  const needle = contractSearch.value.trim().toLowerCase();
  if (!needle) return contractOptions.value;
  return contractOptions.value.filter((c) =>
    c.contractId.toLowerCase().includes(needle),
  );
});

watch(
  contractOptions,
  (options) => {
    if (!contractId.value && options.length > 0) {
      const first = options[0];
      if (first?.contractId) {
        setContract(first.contractId);
        void refresh();
      }
    }
  },
  { immediate: true },
);

watch(
  () => contractId.value,
  (id) => {
    if (id) void refresh();
  },
);

watch(businessDate, () => {
  if (contractId.value) void refresh();
});

// ─────────────────────────────────────────────────────────────────────────
// Stage card data — derived from the current state response.
// ─────────────────────────────────────────────────────────────────────────

interface StageCardModel {
  key: StageKey;
  label: string;
  status: "complete" | "active" | "upcoming" | "blocked";
  timestamp: string | null;
  actor: string | null;
}

function stageStatusOf(
  stage: StageKey,
  current: string | undefined | null,
): StageCardModel["status"] {
  const currentRank = STAGE_ORDER[current ?? "NOT_STARTED"] ?? 0;
  const stageRank = STAGE_ORDER[stage] ?? 0;
  if (currentRank >= stageRank) return "complete";
  if (currentRank + 1 === stageRank) return "active";
  return "upcoming";
}

type StateLike = DeepReadonly<WorkflowStateResponse> | null;

function stageTimestamp(stage: StageKey, s: StateLike): string | null {
  if (!s) return null;
  switch (stage) {
    case "DAY_OPEN":
      return s.openedAt ?? null;
    case "MANAGER_APPROVED":
      return s.managerApprovedAt ?? null;
    case "TRANSACTION_CLOSED":
      return s.transactionClosedAt ?? null;
    case "ACCOUNTING_CLOSED":
      return s.accountingClosedAt ?? null;
    default:
      return null;
  }
}

function stageActor(stage: StageKey, s: StateLike): string | null {
  if (!s) return null;
  switch (stage) {
    case "DAY_OPEN":
      return s.openedBy ?? null;
    case "MANAGER_APPROVED":
      return s.managerApprovedBy ?? null;
    default:
      // Backend doesn't surface a separate actor for closing transitions yet —
      // they live in the history. Stage card stays empty; the timeline below
      // shows who closed/locked the day.
      return null;
  }
}

const stageCards = computed<StageCardModel[]>(() => {
  const current = state.value;
  return STAGES.map((def) => ({
    key: def.key,
    label: t(def.labelKey, def.labelFallback),
    status: stageStatusOf(def.key, current?.currentState ?? null),
    timestamp: stageTimestamp(def.key, current),
    actor: stageActor(def.key, current),
  }));
});

const stageBadgeVariant: Record<StageCardModel["status"], string> = {
  complete: "success",
  active: "info",
  upcoming: "neutral",
  blocked: "warning",
};

function statusLabel(status: StageCardModel["status"]): string {
  switch (status) {
    case "complete":
      return t("dashboardWorkflow.statusComplete", "Complete");
    case "active":
      return t("dashboardWorkflow.statusActive", "Active");
    case "blocked":
      return t("dashboardWorkflow.statusBlocked", "Blocked");
    case "upcoming":
    default:
      return t("dashboardWorkflow.statusUpcoming", "Upcoming");
  }
}

function formatTimestamp(value: string | null): string {
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

// ─────────────────────────────────────────────────────────────────────────
// Operation form — action choice, reason/notes, confirmation.
// ─────────────────────────────────────────────────────────────────────────

const allowedActionSet = computed(
  () => new Set<string>(state.value?.allowedActions ?? []),
);

const visibleActions = computed(() => ACTIONS.filter(
  (a) => allowedActionSet.value.has(a.value),
));

const selectedAction = ref<NonNullable<WorkflowExecuteRequest["action"]> | "">("");
const reasonInput = ref("");
const notesInput = ref("");
const zeroAttestationInput = ref(false);
const attestationReasonInput = ref("");

watch(visibleActions, (actions) => {
  if (
    selectedAction.value
    && !actions.some((a) => a.value === selectedAction.value)
  ) {
    selectedAction.value = "";
  }
  if (!selectedAction.value && actions.length === 1) {
    const first = actions[0];
    if (first) selectedAction.value = first.value;
  }
});

const selectedActionOption = computed(() =>
  selectedAction.value ? ACTION_LOOKUP[selectedAction.value] ?? null : null,
);

const hasActionPermission = computed(() => {
  const opt = selectedActionOption.value;
  if (!opt) return false;
  return authStore.hasPermission(opt.permission);
});

const reasonRequired = computed(
  () => Boolean(selectedActionOption.value?.requiresReason),
);

const reasonValid = computed(() => {
  if (!reasonRequired.value) return true;
  return reasonInput.value.trim().length >= 20;
});

const attestationValid = computed(() => {
  if (selectedAction.value !== "APPROVE") return true;
  if (!zeroAttestationInput.value) return true;
  return attestationReasonInput.value.trim().length >= 30;
});

const canSubmit = computed(() => {
  if (!hasContract.value) return false;
  if (!selectedAction.value) return false;
  if (!hasActionPermission.value) return false;
  if (!reasonValid.value) return false;
  if (!attestationValid.value) return false;
  if (executing.value) return false;
  return true;
});

const disabledReason = computed(() => {
  if (!hasContract.value) {
    return t(
      "dashboardWorkflow.disabledNoContract",
      "Select a contract to start.",
    );
  }
  if (visibleActions.value.length === 0) {
    const blockers = state.value?.blockingReasons ?? [];
    if (blockers.length > 0) {
      return blockers.map((b) => b.message).join(" · ");
    }
    return t(
      "dashboardWorkflow.disabledNoActions",
      "No transitions are currently allowed for this contract day.",
    );
  }
  if (!selectedAction.value) {
    return t(
      "dashboardWorkflow.disabledNoAction",
      "Choose an operation to execute.",
    );
  }
  if (!hasActionPermission.value) {
    return t(
      "dashboardWorkflow.disabledMissingPermission",
      "Your role doesn't include this workflow permission.",
    );
  }
  if (!reasonValid.value) {
    return t(
      "dashboardWorkflow.disabledReasonRequired",
      "A reason of at least 20 characters is required for cancel/rollback operations.",
    );
  }
  if (!attestationValid.value) {
    return t(
      "dashboardWorkflow.disabledAttestationRequired",
      "Zero-transaction attestation requires a reason of at least 30 characters.",
    );
  }
  return "";
});

// ─────────────────────────────────────────────────────────────────────────
// Confirmation dialog
// ─────────────────────────────────────────────────────────────────────────

const confirmOpen = ref(false);

function openConfirm() {
  if (!canSubmit.value) return;
  clearLastResult();
  confirmOpen.value = true;
}

function closeConfirm() {
  if (executing.value) return;
  confirmOpen.value = false;
}

type DialogTone = "danger" | "warning" | "neutral";

const confirmTone = computed<DialogTone>(() => {
  const tone = selectedActionOption.value?.tone;
  if (tone === "danger") return "danger";
  if (tone === "warning") return "warning";
  return "neutral";
});

const confirmTitle = computed(() => {
  const opt = selectedActionOption.value;
  if (!opt) return "";
  if (opt.tone === "danger") {
    return t(
      "dashboardWorkflow.confirmRollbackTitle",
      "Roll back the accounting close?",
    );
  }
  if (opt.tone === "warning") {
    return t(
      "dashboardWorkflow.confirmCancelTitle",
      "Cancel this workflow step?",
    );
  }
  return t("dashboardWorkflow.confirmExecuteTitle", "Execute workflow operation?");
});

const confirmDescription = computed(() => {
  const opt = selectedActionOption.value;
  if (!opt) return "";
  const label = t(opt.labelKey, opt.labelFallback);
  if (opt.tone === "danger") {
    return t(
      "dashboardWorkflow.confirmRollbackDesc",
      `Rolling back the accounting close on ${businessDate.value} will require re-execution. This is recorded in the audit log.`,
    );
  }
  if (opt.tone === "warning") {
    return t(
      "dashboardWorkflow.confirmCancelDesc",
      `${label} on ${businessDate.value} will revert this contract day to the prior state. The action and reason are recorded in the audit log.`,
    );
  }
  return t(
    "dashboardWorkflow.confirmExecuteDesc",
    `Apply ${label} to this contract on business date ${businessDate.value}? The action is recorded in the audit log.`,
  );
});

async function handleConfirm() {
  if (!selectedAction.value) return;
  const payload: WorkflowExecuteRequest = {
    action: selectedAction.value,
    businessDate: businessDate.value,
  };
  if (reasonInput.value.trim()) payload.reason = reasonInput.value.trim();
  if (notesInput.value.trim()) payload.notes = notesInput.value.trim();
  if (selectedAction.value === "APPROVE") {
    payload.zeroTransactionAttestation = zeroAttestationInput.value;
    if (zeroAttestationInput.value) {
      payload.attestationReason = attestationReasonInput.value.trim();
    }
  }

  const ok = await execute(payload);
  if (ok) {
    confirmOpen.value = false;
    reasonInput.value = "";
    notesInput.value = "";
    zeroAttestationInput.value = false;
    attestationReasonInput.value = "";
  }
}

// ─────────────────────────────────────────────────────────────────────────
// History timeline — most recent first.
// ─────────────────────────────────────────────────────────────────────────

const recentHistory = computed(() => {
  const items = [...history.value];
  items.sort((a, b) => {
    const ta = a.occurredAt ? Date.parse(a.occurredAt) : 0;
    const tb = b.occurredAt ? Date.parse(b.occurredAt) : 0;
    return tb - ta;
  });
  return items.slice(0, 10);
});

function actionLabel(action: string | undefined): string {
  if (!action) return "";
  const opt = ACTION_LOOKUP[action];
  if (!opt) return action;
  return t(opt.labelKey, opt.labelFallback);
}
</script>

<template>

</template>

<style scoped>
.workflow-panel__layout {
  display: grid;
  grid-template-columns: minmax(14rem, 18rem) minmax(0, 1fr);
  gap: var(--space-5);
}

.workflow-panel__sidebar {
  display: grid;
  gap: var(--space-3);
  align-content: start;
  border-right: 1px solid var(--border-subtle);
  padding-right: var(--space-4);
}

.workflow-panel__label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.workflow-panel__input,
.workflow-panel__input:is(textarea, select) {
  width: 100%;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  background: var(--bg-input, white);
  color: var(--text-primary);
  font-family: inherit;
}

.workflow-panel__input:focus {
  outline: 2px solid var(--action-primary);
  outline-offset: -1px;
  border-color: var(--action-primary);
}

.workflow-panel__input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.workflow-panel__contract-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-1);
  max-height: 22rem;
  overflow-y: auto;
}

.workflow-panel__contract-button {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: transparent;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  cursor: pointer;
  text-align: left;
}

.workflow-panel__contract-button:hover {
  background: var(--bg-row-hover);
}

.workflow-panel__contract-item.is-active .workflow-panel__contract-button {
  background: var(--bg-selected);
  border-color: var(--action-primary);
}

.workflow-panel__contract-id {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 12rem;
}

.workflow-panel__main {
  display: grid;
  gap: var(--space-4);
  min-width: 0;
}

.workflow-panel__main-header {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: flex-end;
  gap: var(--space-3);
}

.workflow-panel__contract-label {
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  color: var(--text-tertiary);
  letter-spacing: 0.04em;
}

.workflow-panel__contract-value {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.workflow-panel__date-control {
  display: grid;
  gap: var(--space-1);
  min-width: 12rem;
}

.workflow-panel__stages {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
}

.workflow-stage {
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--space-3) var(--space-4);
  background: var(--bg-card, white);
  display: grid;
  gap: var(--space-2);
}

.workflow-stage.is-complete {
  border-color: var(--state-success, #10b981);
  background: var(--status-approved-bg);
}

.workflow-stage.is-active {
  border-color: var(--action-primary, #2563eb);
  background: var(--status-executed-bg);
}

.workflow-stage__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

.workflow-stage__title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.workflow-stage__body {
  margin: 0;
  display: grid;
  gap: var(--space-1);
}

.workflow-stage__body > div {
  display: grid;
  grid-template-columns: 5rem 1fr;
  gap: var(--space-2);
}

.workflow-stage__body dt {
  font-size: 11px;
  text-transform: uppercase;
  color: var(--text-tertiary);
  letter-spacing: 0.04em;
}

.workflow-stage__body dd {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.workflow-panel__notice {
  display: flex;
  gap: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4);
  background: var(--bg-card-muted, #f9fafb);
}

.workflow-panel__notice.is-warning {
  border-color: var(--state-warning, #d97706);
  background: var(--alert-warning-bg);
}

.workflow-panel__notice.is-danger {
  border-color: var(--state-danger, #dc2626);
  background: var(--alert-danger-bg);
}

.workflow-panel__notice.is-success {
  border-color: var(--state-success, #059669);
  background: var(--alert-success-bg);
}

.workflow-panel__notice-title {
  font-weight: var(--font-weight-semibold);
}

.workflow-panel__notice-list {
  margin: var(--space-1) 0 0;
  padding-left: var(--space-4);
  font-size: var(--font-size-sm);
}

.workflow-panel__form {
  display: grid;
  gap: var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: var(--space-4);
  background: var(--bg-card, white);
}

.workflow-panel__field {
  display: grid;
  gap: var(--space-1);
}

.workflow-panel__checkbox {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-sm);
}

.workflow-panel__form-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.workflow-panel__help {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.workflow-panel__history {
  display: grid;
  gap: var(--space-2);
}

.workflow-panel__history-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.workflow-panel__history-title {
  margin: 0;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.workflow-panel__history-loading,
.workflow-panel__history-empty {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.workflow-panel__timeline {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-2);
}

.workflow-panel__timeline-item {
  display: grid;
  grid-template-columns: 10rem 1fr;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border-left: 2px solid var(--border-subtle);
}

.workflow-panel__timeline-time {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 12px;
  color: var(--text-tertiary);
}

.workflow-panel__timeline-headline {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  font-size: var(--font-size-sm);
}

.workflow-panel__timeline-arrow {
  font-family: var(--font-mono, ui-monospace, monospace);
  font-size: 12px;
  color: var(--text-tertiary);
}

.workflow-panel__timeline-meta {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

@media (max-width: 960px) {
  .workflow-panel__layout {
    grid-template-columns: 1fr;
  }

  .workflow-panel__sidebar {
    border-right: 0;
    border-bottom: 1px solid var(--border-subtle);
    padding-right: 0;
    padding-bottom: var(--space-4);
  }

  .workflow-panel__stages {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .workflow-panel__stages {
    grid-template-columns: 1fr;
  }

  .workflow-panel__timeline-item {
    grid-template-columns: 1fr;
  }
}
</style>
