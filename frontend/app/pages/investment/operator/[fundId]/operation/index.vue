<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import DecisionFilters from "~/features/investment-decision/components/DecisionFilters.vue";
import DecisionTable from "~/features/investment-decision/components/DecisionTable.vue";
import BatchActionBar from "~/features/investment-decision/components/BatchActionBar.vue";
import { useDecisionList } from "~/features/investment-decision/composables/useDecisionList";
import { useDecisionMutations } from "~/features/investment-decision/composables/useDecisionMutations";
import {
  DECISION_STATUSES,
  type DecisionListFilters,
  type DecisionStatus,
} from "~/features/investment-decision/services/decisionApi";
import type { ApiDecision } from "~/features/investment-decision/services/decisionApi";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const route = useRoute();
const router = useRouter();
const fundId = computed(() => route.params.fundId as string);

const { items, total, page, limit, loading, error, fetchList } =
  useDecisionList();
const mutations = useDecisionMutations();

const filters = ref<DecisionListFilters>({});
const selectedNos = ref<string[]>([]);

function firstString(raw: unknown): string | undefined {
  if (typeof raw === "string") return raw;
  if (Array.isArray(raw) && typeof raw[0] === "string") return raw[0];
  return undefined;
}

function parseIntInRange(
  raw: unknown,
  min: number,
  max: number,
): number | undefined {
  const s = firstString(raw);
  if (!s) return undefined;
  const n = Number.parseInt(s, 10);
  if (!Number.isFinite(n) || n < min || n > max) return undefined;
  return n;
}

function pickEnum<T extends string>(
  raw: unknown,
  allowed: readonly T[],
): T | undefined {
  if (typeof raw !== "string") return undefined;
  return (allowed as readonly string[]).includes(raw) ? (raw as T) : undefined;
}

function hydrateFromQuery() {
  const q = route.query;
  const next: DecisionListFilters = {};
  const search = firstString(q.search);
  if (search) next.search = search;
  const status = pickEnum<DecisionStatus>(
    firstString(q.status),
    DECISION_STATUSES,
  );
  if (status) next.status = status;
  const bd = firstString(q.business_date);
  if (bd) next.business_date = bd;
  filters.value = next;
  const pageFromUrl = parseIntInRange(q.page, 1, 10_000);
  if (pageFromUrl) page.value = pageFromUrl;
}

function writeQuery() {
  const q: Record<string, string> = {};
  if (filters.value.search) q.search = filters.value.search;
  if (filters.value.status) q.status = filters.value.status;
  if (filters.value.business_date) q.business_date = filters.value.business_date;
  if (page.value > 1) q.page = String(page.value);
  void router.replace({ query: q });
}

async function refresh() {
  await fetchList({ ...filters.value, fund_id: fundId.value });
  writeQuery();
}

function applyFilters(next: DecisionListFilters) {
  filters.value = next;
  page.value = 1;
  void refresh();
}

function onPageChange(p: number) {
  page.value = p;
  void refresh();
}

function openDecision(decision: ApiDecision) {
  void router.push(
    `/investment/operator/${fundId.value}/operation/${decision.id}`,
  );
}

const batchApproveDialog = ref(false);
const batchRejectDialog = ref(false);
const batchComment = ref("");
const batchReason = ref("");

async function confirmBatchApprove() {
  const result = await mutations.batchApprove({
    decision_nos: selectedNos.value,
    comment: batchComment.value,
  });
  if (result) {
    selectedNos.value = [];
    batchComment.value = "";
    batchApproveDialog.value = false;
    void refresh();
  }
}

async function confirmBatchReject() {
  if (!batchReason.value.trim()) return;
  const result = await mutations.batchReject({
    decision_nos: selectedNos.value,
    reason: batchReason.value,
  });
  if (result) {
    selectedNos.value = [];
    batchReason.value = "";
    batchRejectDialog.value = false;
    void refresh();
  }
}

watch(
  () => route.query,
  () => {
    hydrateFromQuery();
    void fetchList({ ...filters.value, fund_id: fundId.value });
  },
);

onMounted(() => {
  hydrateFromQuery();
  void refresh();
});
</script>

<template>
  <div>
    <AppPageHeader title="Operation — Decisions">
      <template #actions>
        <AppButton
          variant="primary"
          @click="router.push(`/investment/operator/${fundId}/operation/new`)"
        >
          New Decision
        </AppButton>
      </template>
    </AppPageHeader>

    <div class="toolbar">
      <DecisionFilters
        :model-value="filters"
        @update:model-value="(v) => (filters = v)"
        @change="() => applyFilters(filters)"
      />
    </div>

    <div v-if="error" class="error-banner">{{ error }}</div>

    <BatchActionBar
      :selected-nos="selectedNos"
      :loading="mutations.loading.value"
      @approve="batchApproveDialog = true"
      @reject="batchRejectDialog = true"
      @clear="selectedNos = []"
    />

    <DecisionTable
      :items="items"
      :loading="loading"
      :total="total"
      :page="page"
      :limit="limit"
      :selected-nos="selectedNos"
      @update:selected-nos="(v) => (selectedNos = v)"
      @update:page="onPageChange"
      @open="openDecision"
    />

    <div v-if="batchApproveDialog" class="dialog-overlay" @click.self="batchApproveDialog = false">
      <div class="dialog">
        <div class="dialog__title">Batch Approve {{ selectedNos.length }} Decision(s)</div>
        <div class="dialog__body">
          <label class="dialog__label">Comment (optional)</label>
          <textarea v-model="batchComment" rows="3" class="dialog__textarea" />
        </div>
        <div class="dialog__actions">
          <button class="dialog__btn" @click="batchApproveDialog = false">Cancel</button>
          <button
            class="dialog__btn dialog__btn--primary"
            :disabled="mutations.loading.value"
            @click="confirmBatchApprove"
          >
            Approve
          </button>
        </div>
      </div>
    </div>

    <div v-if="batchRejectDialog" class="dialog-overlay" @click.self="batchRejectDialog = false">
      <div class="dialog">
        <div class="dialog__title">Batch Reject {{ selectedNos.length }} Decision(s)</div>
        <div class="dialog__body">
          <label class="dialog__label">Reason <span class="required">*</span></label>
          <textarea v-model="batchReason" rows="3" class="dialog__textarea" />
        </div>
        <div class="dialog__actions">
          <button class="dialog__btn" @click="batchRejectDialog = false">Cancel</button>
          <button
            class="dialog__btn dialog__btn--danger"
            :disabled="mutations.loading.value || !batchReason.trim()"
            @click="confirmBatchReject"
          >
            Reject
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.error-banner {
  padding: 10px 14px;
  margin-bottom: 12px;
  background: rgba(207, 34, 46, 0.1);
  border: 1px solid var(--state-danger, #cf222e);
  border-radius: var(--radius-sm, 4px);
  color: var(--state-danger, #cf222e);
  font-size: 13px;
}

.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.dialog {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  padding: 20px 24px;
  min-width: 360px;
  max-width: 480px;
  box-shadow: var(--shadow-md);
}

.dialog__title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 14px;
}

.dialog__body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.dialog__label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}

.dialog__textarea {
  width: 100%;
  padding: 7px 10px;
  font-size: 13px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  color: var(--text-primary);
  box-sizing: border-box;
  resize: none;
}

.dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.dialog__btn {
  padding: 6px 16px;
  font-size: 13px;
  font-weight: 600;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  color: var(--text-primary);
  cursor: pointer;
}

.dialog__btn--primary {
  background: var(--state-success, #1a7f37);
  color: #fff;
  border-color: transparent;
}

.dialog__btn--danger {
  background: var(--state-danger, #cf222e);
  color: #fff;
  border-color: transparent;
}

.dialog__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.required {
  color: var(--state-danger, #cf222e);
}
</style>
