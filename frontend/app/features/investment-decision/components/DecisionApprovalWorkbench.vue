<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "#imports";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppErrorState from "~/shared/ui/AppErrorState.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppToast from "~/shared/ui/AppToast.vue";
import {
  useDecisionApprovalList,
  useDecisionBatchAction,
} from "../composables/useDecisionApproval";
import {
  approvalFiltersFromQuery,
  approvalFiltersToQuery,
  cleanApprovalFilters,
} from "../lib/approvalLookup";
import type {
  ApiDecision,
  DecisionApprovalFilters,
} from "../services/decisionApprovalApi";
import DecisionApprovalGrid from "./DecisionApprovalGrid.vue";
import DecisionDetailDrawer from "./DecisionDetailDrawer.vue";
import DecisionSearchFilter from "./DecisionSearchFilter.vue";

type ActionMode = "approve" | "reject" | null;
type ToastTone = "success" | "danger" | "info";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const list = useDecisionApprovalList();
const batchAction = useDecisionBatchAction();

const activeFilters = ref<DecisionApprovalFilters>(
  approvalFiltersFromQuery(route.query),
);
const selectedIds = ref<string[]>([]);
const drawerOpen = ref(false);
const drawerDecisionId = ref<string | null>(null);
const actionMode = ref<ActionMode>(null);
const approveComment = ref("");
const rejectReason = ref("");
const toastVisible = ref(false);
const toastMessage = ref("");
const toastTone = ref<ToastTone>("success");

const selectedDecisionNos = computed<string[]>(() =>
  list.items.value
    .filter((decision) => decision.id && selectedIds.value.includes(decision.id))
    .map((decision) => decision.decision_number?.trim() ?? "")
    .filter(Boolean),
);

const canAct = computed(
  () => selectedDecisionNos.value.length > 0 && !batchAction.submitting.value,
);
const canConfirmReject = computed(() => rejectReason.value.trim().length > 0);
const activeLookup = computed(
  () =>
    activeFilters.value.decision_no ||
    t("approval.decisionWorkbench.queue.allPending", "All pending decisions"),
);

async function loadList() {
  await list.fetchList(activeFilters.value);
}

function writeQuery() {
  void router.replace({ query: approvalFiltersToQuery(activeFilters.value) });
}

function onSearch(filters: DecisionApprovalFilters) {
  activeFilters.value = cleanApprovalFilters(filters);
  list.page.value = 1;
  selectedIds.value = [];
  actionMode.value = null;
  writeQuery();
  void loadList();
}

function onReset() {
  activeFilters.value = {};
  list.page.value = 1;
  selectedIds.value = [];
  actionMode.value = null;
  writeQuery();
  void loadList();
}

function onSelectChange(ids: string[]) {
  selectedIds.value = ids;
}

function onRowClick(item: ApiDecision) {
  if (!item.id) return;
  drawerDecisionId.value = item.id;
  drawerOpen.value = true;
}

function openApprove() {
  actionMode.value = "approve";
  approveComment.value = "";
}

function openReject() {
  actionMode.value = "reject";
  rejectReason.value = "";
}

function cancelAction() {
  actionMode.value = null;
}

function showToast(message: string, tone: ToastTone) {
  toastMessage.value = message;
  toastTone.value = tone;
  toastVisible.value = true;
}

async function confirmAction() {
  const mode = actionMode.value;
  if (!mode || selectedDecisionNos.value.length === 0) return;

  try {
    if (mode === "approve") {
      await batchAction.approve(selectedDecisionNos.value, approveComment.value);
    } else {
      await batchAction.reject(selectedDecisionNos.value, rejectReason.value);
    }

    if (batchAction.allFailed.value) {
      showToast(
        t(
          "approval.decisionWorkbench.toast.allFailed",
          "None of the selected decisions could be processed.",
        ),
        "danger",
      );
    } else if (batchAction.hasPartialFailure.value) {
      showToast(
        t(
          "approval.decisionWorkbench.toast.partial",
          {
            succeeded: batchAction.succeeded.value,
            failed: batchAction.failed.value,
          },
          "{succeeded} succeeded and {failed} failed.",
        ),
        "info",
      );
    } else {
      showToast(
        t(
          mode === "approve"
            ? "approval.decisionWorkbench.toast.approved"
            : "approval.decisionWorkbench.toast.rejected",
          { count: batchAction.succeeded.value },
          mode === "approve"
            ? "Approved {count} decision(s)."
            : "Rejected {count} decision(s).",
        ),
        "success",
      );
    }

    actionMode.value = null;
    selectedIds.value = [];
    await loadList();
  } catch {
    showToast(
      batchAction.error.value ??
        t("approval.decisionWorkbench.toast.actionFailed", "Approval action failed."),
      "danger",
    );
  }
}

function previousPage() {
  if (list.page.value <= 1) return;
  list.page.value -= 1;
  selectedIds.value = [];
  void loadList();
}

function nextPage() {
  if (list.page.value * list.limit.value >= list.total.value) return;
  list.page.value += 1;
  selectedIds.value = [];
  void loadList();
}

onMounted(() => {
  void loadList();
});
</script>

<template>
  <section class="decision-workbench">
    <AppPageHeader
      :title="t('approval.decisionWorkbench.title', 'Execution & approval')"
      :description="t(
        'approval.decisionWorkbench.description',
        'Find an investment decision by its business number, inspect the approval route, and approve or reject it before execution.',
      )"
    >
      <template #eyebrow>
        <span class="decision-workbench__eyebrow">
          {{ t("approval.decisionWorkbench.eyebrow", "OP-02 · Investment operations") }}
        </span>
      </template>
    </AppPageHeader>

    <div
      class="decision-workbench__queue"
      role="group"
      :aria-label="t('approval.decisionWorkbench.queue.ariaLabel', 'Approval queue summary')"
    >
      <div class="queue-stat queue-stat--active">
        <span class="queue-stat__label">
          {{ t("approval.decisionWorkbench.queue.matches", "Matching approvals") }}
        </span>
        <strong class="queue-stat__value">{{ list.total.value }}</strong>
        <span class="queue-stat__detail">{{ activeLookup }}</span>
      </div>
      <div class="queue-stat">
        <span class="queue-stat__label">
          {{ t("approval.decisionWorkbench.queue.visible", "Visible rows") }}
        </span>
        <strong class="queue-stat__value">{{ list.items.value.length }}</strong>
        <span class="queue-stat__detail">
          {{ t("approval.decisionWorkbench.queue.currentPage", "Current page") }}
        </span>
      </div>
      <div class="queue-stat">
        <span class="queue-stat__label">
          {{ t("approval.decisionWorkbench.queue.selected", "Selected") }}
        </span>
        <strong class="queue-stat__value">{{ selectedDecisionNos.length }}</strong>
        <span class="queue-stat__detail">
          {{ t("approval.decisionWorkbench.queue.readyForAction", "Ready for action") }}
        </span>
      </div>
    </div>

    <AppCard
      :title="t('approval.decisionWorkbench.lookup.title', 'Decision lookup')"
      :subtitle="t(
        'approval.decisionWorkbench.lookup.subtitle',
        'Use the decision number shown on the investment decision. Internal database IDs are not required.',
      )"
    >
      <DecisionSearchFilter
        :initial-filters="activeFilters"
        :loading="list.loading.value"
        @search="onSearch"
        @reset="onReset"
      />
    </AppCard>

    <AppCard
      :title="t('approval.decisionWorkbench.table.title', 'Approval queue')"
      :subtitle="t(
        'approval.decisionWorkbench.table.subtitle',
        'Select one or more pending decisions. Open a row to review its stage, approvers, and order lines.',
      )"
    >
      <template #header-actions>
        <div class="decision-workbench__actions">
          <AppButton
            variant="primary"
            size="sm"
            :disabled="!canAct || actionMode !== null"
            @click="openApprove"
          >
            {{
              t(
                "approval.decisionWorkbench.actions.approveCount",
                { count: selectedDecisionNos.length },
                "Approve ({count})",
              )
            }}
          </AppButton>
          <AppButton
            variant="danger"
            size="sm"
            :disabled="!canAct || actionMode !== null"
            @click="openReject"
          >
            {{
              t(
                "approval.decisionWorkbench.actions.rejectCount",
                { count: selectedDecisionNos.length },
                "Reject ({count})",
              )
            }}
          </AppButton>
        </div>
      </template>

      <Transition name="action-panel">
        <div v-if="actionMode" class="decision-workbench__action-panel">
          <p class="action-panel__prompt">
            {{
              t(
                actionMode === "approve"
                  ? "approval.decisionWorkbench.actions.approvePrompt"
                  : "approval.decisionWorkbench.actions.rejectPrompt",
                { count: selectedDecisionNos.length },
                actionMode === "approve"
                  ? "Approve {count} selected decision(s)?"
                  : "Reject {count} selected decision(s)?",
              )
            }}
          </p>

          <label v-if="actionMode === 'approve'" class="action-panel__field">
            <span>
              {{ t("approval.decisionWorkbench.actions.commentLabel", "Comment (optional)") }}
            </span>
            <textarea
              v-model="approveComment"
              class="action-panel__textarea"
              rows="2"
              :placeholder="t(
                'approval.decisionWorkbench.actions.commentPlaceholder',
                'Add an approval comment…',
              )"
            />
          </label>

          <label v-else class="action-panel__field">
            <span>
              {{ t("approval.decisionWorkbench.actions.reasonLabel", "Reason") }}
              <span class="action-panel__required" aria-hidden="true">*</span>
            </span>
            <textarea
              v-model="rejectReason"
              class="action-panel__textarea"
              rows="2"
              required
              :placeholder="t(
                'approval.decisionWorkbench.actions.reasonPlaceholder',
                'Explain why this decision is being rejected.',
              )"
            />
          </label>

          <div class="action-panel__buttons">
            <AppButton variant="secondary" size="sm" @click="cancelAction">
              {{ t("approval.decisionWorkbench.actions.cancel", "Cancel") }}
            </AppButton>
            <AppButton
              :variant="actionMode === 'approve' ? 'primary' : 'danger'"
              size="sm"
              :disabled="actionMode === 'reject' && !canConfirmReject"
              :loading="batchAction.submitting.value"
              @click="confirmAction"
            >
              {{
                actionMode === "approve"
                  ? t("approval.decisionWorkbench.actions.confirmApproval", "Confirm approval")
                  : t("approval.decisionWorkbench.actions.confirmRejection", "Confirm rejection")
              }}
            </AppButton>
          </div>
        </div>
      </Transition>

      <AppErrorState
        v-if="list.forbidden.value"
        :title="t('approval.decisionWorkbench.errors.forbiddenTitle', 'Approval access required')"
        :message="t(
          'approval.decisionWorkbench.errors.forbiddenMessage',
          'You do not have permission to query or approve investment decisions.',
        )"
      />
      <AppErrorState
        v-else-if="list.error.value"
        :title="t('approval.decisionWorkbench.errors.loadTitle', 'Could not load approvals')"
        :message="list.error.value"
        retry
        @retry="loadList"
      />

      <DecisionApprovalGrid
        v-else
        :items="list.items.value"
        :loading="list.loading.value"
        :selected-ids="selectedIds"
        @select-change="onSelectChange"
        @row-click="onRowClick"
      />

      <div v-if="!list.forbidden.value && list.total.value > list.limit.value" class="decision-workbench__pagination">
        <AppButton
          variant="ghost"
          size="sm"
          :disabled="list.page.value <= 1 || list.loading.value"
          @click="previousPage"
        >
          {{ t("approval.decisionWorkbench.pagination.previous", "Previous") }}
        </AppButton>
        <span class="decision-workbench__page-info">
          {{
            t(
              "approval.decisionWorkbench.pagination.page",
              { page: list.page.value, total: list.total.value },
              "Page {page} · {total} total",
            )
          }}
        </span>
        <AppButton
          variant="ghost"
          size="sm"
          :disabled="list.page.value * list.limit.value >= list.total.value || list.loading.value"
          @click="nextPage"
        >
          {{ t("approval.decisionWorkbench.pagination.next", "Next") }}
        </AppButton>
      </div>
    </AppCard>

    <DecisionDetailDrawer
      :open="drawerOpen"
      :decision-id="drawerDecisionId"
      @close="drawerOpen = false"
    />

    <AppToast v-model="toastVisible" :message="toastMessage" :tone="toastTone" />
  </section>
</template>

<style scoped>
.decision-workbench {
  display: grid;
  gap: var(--space-5, 20px);
}

.decision-workbench__eyebrow {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 2px 9px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 999px;
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--text-secondary, #57606a);
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
}

.decision-workbench__queue {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  overflow: hidden;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 8px);
  background: var(--bg-card, #fff);
}

.queue-stat {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 2px var(--space-3, 12px);
  min-width: 0;
  padding: var(--space-3, 12px) var(--space-4, 16px);
  border-left: 1px solid var(--border-subtle, #d0d7de);
}

.queue-stat:first-child {
  border-left: 0;
}

.queue-stat--active {
  box-shadow: inset 0 3px 0 var(--action-primary, #0969da);
  background: color-mix(in srgb, var(--action-primary, #0969da) 5%, var(--bg-card, #fff));
}

.queue-stat__label {
  overflow: hidden;
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.queue-stat__value {
  grid-row: span 2;
  align-self: center;
  color: var(--text-primary, #1f2328);
  font-size: clamp(1.2rem, 2vw, 1.65rem);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.queue-stat__detail {
  overflow: hidden;
  color: var(--text-tertiary, #6e7781);
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.decision-workbench__actions,
.action-panel__buttons,
.decision-workbench__pagination {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  flex-wrap: wrap;
}

.decision-workbench__action-panel {
  display: grid;
  gap: var(--space-3, 12px);
  margin-bottom: var(--space-4, 16px);
  padding: var(--space-4, 16px);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card-muted, #f6f8fa);
}

.action-panel__prompt {
  margin: 0;
  color: var(--text-primary, #1f2328);
  font-size: var(--font-size-sm, 14px);
  font-weight: var(--font-weight-semibold, 600);
}

.action-panel__field {
  display: grid;
  gap: var(--space-1, 4px);
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-sm, 14px);
  font-weight: var(--font-weight-medium, 500);
}

.action-panel__required {
  color: var(--alert-danger-text, #cf222e);
}

.action-panel__textarea {
  width: 100%;
  min-height: 68px;
  padding: var(--space-2, 8px) var(--space-3, 12px);
  border: 1px solid var(--border-default, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-input, #fff);
  color: var(--text-primary, #1f2328);
  font: inherit;
  resize: vertical;
}

.action-panel__textarea:focus-visible {
  outline: 2px solid var(--border-focus, #0969da);
  outline-offset: 1px;
}

.decision-workbench__pagination {
  justify-content: center;
  margin-top: var(--space-4, 16px);
}

.decision-workbench__page-info {
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-sm, 14px);
  font-variant-numeric: tabular-nums;
}

.action-panel-enter-active,
.action-panel-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.action-panel-enter-from,
.action-panel-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

@media (prefers-reduced-motion: reduce) {
  .action-panel-enter-active,
  .action-panel-leave-active {
    transition: none;
  }
}

@media (max-width: 720px) {
  .decision-workbench__queue {
    grid-template-columns: 1fr;
  }

  .queue-stat,
  .queue-stat:first-child {
    border-top: 1px solid var(--border-subtle, #d0d7de);
    border-left: 0;
  }

  .queue-stat:first-child {
    border-top: 0;
  }

  .decision-workbench__actions {
    width: 100%;
  }

  .decision-workbench__actions > * {
    flex: 1;
  }
}
</style>
