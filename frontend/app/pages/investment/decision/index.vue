<script setup lang="ts">
import { ref, computed } from "vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppErrorState from "~/shared/ui/AppErrorState.vue";
import AppToast from "~/shared/ui/AppToast.vue";
import DecisionSearchFilter from "~/features/investment-decision/components/DecisionSearchFilter.vue";
import DecisionApprovalGrid from "~/features/investment-decision/components/DecisionApprovalGrid.vue";
import DecisionDetailDrawer from "~/features/investment-decision/components/DecisionDetailDrawer.vue";
import {
  useDecisionApprovalList,
  useDecisionBatchAction,
} from "~/features/investment-decision/composables/useDecisionApproval";
import type { ApiDecision } from "~/features/investment-decision/services/decisionApprovalApi";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_DECISION_APPROVE",
});

const list = useDecisionApprovalList();
const batchAction = useDecisionBatchAction();

const selectedIds = ref<string[]>([]);
const activeFilters = ref<Record<string, string>>({});

const drawerOpen = ref(false);
const drawerDecisionId = ref<string | null>(null);

type ActionMode = "approve" | "reject" | null;
const actionMode = ref<ActionMode>(null);
const approveComment = ref("");
const rejectReason = ref("");

const toastVisible = ref(false);
const toastMessage = ref("");
const toastTone = ref<"success" | "danger" | "info">("success");

const selectedDecisionNos = computed<string[]>(() =>
  list.items.value
    .filter((d) => d.id && selectedIds.value.includes(d.id))
    .map((d) => d.decision_number ?? "")
    .filter(Boolean),
);

const canAct = computed(() => selectedIds.value.length > 0 && !batchAction.submitting.value);
const canConfirmReject = computed(() => rejectReason.value.trim().length >= 1);

async function loadList(extra: Record<string, string> = {}) {
  await list.fetchList({ ...activeFilters.value, ...extra });
}

function onSearch(filters: Record<string, string>) {
  activeFilters.value = filters;
  list.page.value = 1;
  loadList();
}

function onReset() {
  activeFilters.value = {};
  list.page.value = 1;
  loadList();
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

async function confirmAction() {
  const nos = selectedDecisionNos.value;
  try {
    if (actionMode.value === "approve") {
      await batchAction.approve(nos, approveComment.value);
    } else {
      await batchAction.reject(nos, rejectReason.value);
    }
    actionMode.value = null;

    if (batchAction.allFailed.value) {
      showToast("All decisions failed to process.", "danger");
    } else if (batchAction.hasPartialFailure.value) {
      showToast(
        `${batchAction.succeeded.value} succeeded, ${batchAction.failed.value} failed.`,
        "info",
      );
    } else {
      const verb = actionMode.value === null ? (batchAction.succeeded.value > 0 ? "processed" : "rejected") : actionMode.value === "approve" ? "approved" : "rejected";
      showToast(`${batchAction.succeeded.value} decision(s) processed successfully.`, "success");
    }

    selectedIds.value = [];
    await loadList();
  } catch {
    showToast(batchAction.error.value ?? "Action failed.", "danger");
  }
}

function showToast(message: string, tone: "success" | "danger" | "info") {
  toastMessage.value = message;
  toastTone.value = tone;
  toastVisible.value = true;
}

loadList();
</script>

<template>
  <section class="decision-page">
    <AppPageHeader
      title="Investment Decision Approval"
      description="Review and batch-approve or reject pending investment decisions before execution."
    />

    <AppCard>
      <template #default>
        <div class="decision-page__toolbar">
          <DecisionSearchFilter @search="onSearch" @reset="onReset" />
          <div class="decision-page__actions">
            <AppButton
              variant="primary"
              size="sm"
              :disabled="!canAct || actionMode !== null"
              @click="openApprove"
            >
              Approve ({{ selectedIds.length }})
            </AppButton>
            <AppButton
              variant="danger"
              size="sm"
              :disabled="!canAct || actionMode !== null"
              @click="openReject"
            >
              Reject ({{ selectedIds.length }})
            </AppButton>
          </div>
        </div>

        <Transition name="action-panel">
          <div v-if="actionMode" class="decision-page__action-panel">
            <div class="action-panel__inner">
              <p class="action-panel__prompt">
                <template v-if="actionMode === 'approve'">
                  Approve <strong>{{ selectedIds.length }}</strong> selected decision(s)?
                </template>
                <template v-else>
                  Reject <strong>{{ selectedIds.length }}</strong> selected decision(s)?
                </template>
              </p>

              <label v-if="actionMode === 'approve'" class="action-panel__field">
                <span>Comment (optional)</span>
                <textarea
                  v-model="approveComment"
                  class="action-panel__textarea"
                  rows="2"
                  placeholder="Add an approval comment..."
                />
              </label>

              <label v-if="actionMode === 'reject'" class="action-panel__field">
                <span>Reason <span class="action-panel__required">*</span></span>
                <textarea
                  v-model="rejectReason"
                  class="action-panel__textarea"
                  rows="2"
                  placeholder="Required — explain the rejection reason"
                />
              </label>

              <div class="action-panel__buttons">
                <AppButton variant="secondary" size="sm" @click="cancelAction">Cancel</AppButton>
                <AppButton
                  :variant="actionMode === 'approve' ? 'primary' : 'danger'"
                  size="sm"
                  :disabled="actionMode === 'reject' && !canConfirmReject"
                  :loading="batchAction.submitting.value"
                  @click="confirmAction"
                >
                  {{ actionMode === "approve" ? "Confirm Approval" : "Confirm Rejection" }}
                </AppButton>
              </div>
            </div>
          </div>
        </Transition>

        <AppErrorState
          v-if="list.error.value && !list.forbidden.value"
          :message="list.error.value"
        />
        <AppErrorState
          v-if="list.forbidden.value"
          message="You do not have permission to access the approval screen. Contact your system administrator."
        />

        <DecisionApprovalGrid
          :items="list.items.value"
          :loading="list.loading.value"
          :error="null"
          :selected-ids="selectedIds"
          @select-change="onSelectChange"
          @row-click="onRowClick"
        />

        <div v-if="list.total.value > list.limit.value" class="decision-page__pagination">
          <AppButton
            variant="ghost"
            size="sm"
            :disabled="list.page.value <= 1"
            @click="() => { list.page.value--; loadList(); }"
          >
            Previous
          </AppButton>
          <span class="decision-page__page-info">
            Page {{ list.page.value }} — {{ list.total.value }} total
          </span>
          <AppButton
            variant="ghost"
            size="sm"
            :disabled="list.page.value * list.limit.value >= list.total.value"
            @click="() => { list.page.value++; loadList(); }"
          >
            Next
          </AppButton>
        </div>
      </template>
    </AppCard>

    <DecisionDetailDrawer
      :open="drawerOpen"
      :decision-id="drawerDecisionId"
      @close="drawerOpen = false"
    />

    <AppToast
      v-model="toastVisible"
      :message="toastMessage"
      :tone="toastTone"
    />
  </section>
</template>

<style scoped>
.decision-page {
  display: grid;
  gap: var(--space-7);
}

.decision-page__toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: var(--space-5);
}

.decision-page__actions {
  display: flex;
  gap: var(--space-3);
  flex-shrink: 0;
}

.decision-page__action-panel {
  margin-bottom: var(--space-5);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  background: var(--surface-secondary, #f9fafb);
  overflow: hidden;
}

.action-panel__inner {
  padding: var(--space-4) var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.action-panel__prompt {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.action-panel__field {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-secondary);
}

.action-panel__required {
  color: var(--color-danger-600, #dc2626);
}

.action-panel__textarea {
  padding: var(--space-2) var(--space-3);
  border: 1px solid var(--border-primary);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  font-family: inherit;
  resize: vertical;
  background: var(--surface-primary, #fff);
  color: var(--text-primary);
}

.action-panel__textarea:focus {
  outline: 2px solid var(--color-primary-500, #3b82f6);
  outline-offset: -1px;
}

.action-panel__buttons {
  display: flex;
  gap: var(--space-3);
}

.decision-page__pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  margin-top: var(--space-5);
}

.decision-page__page-info {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.action-panel-enter-active,
.action-panel-leave-active {
  transition: opacity 0.2s ease, max-height 0.25s ease;
  max-height: 300px;
  overflow: hidden;
}

.action-panel-enter-from,
.action-panel-leave-to {
  opacity: 0;
  max-height: 0;
}
</style>
