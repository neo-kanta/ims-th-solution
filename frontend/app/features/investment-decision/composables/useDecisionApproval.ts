import { ref, computed } from "vue";

import {
  decisionApprovalApi,
  type ApiDecision,
  type ApiBatchApprovalResult,
  type DecisionApprovalFilters,
} from "../services/decisionApprovalApi";

function extractErrorMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const data = (err as { data?: { error?: unknown } }).data;
  if (data && typeof data.error === "string" && data.error.trim()) return data.error;
  const message = (err as { message?: unknown }).message;
  if (typeof message === "string" && message.trim()) return message;
  return fallback;
}

export function useDecisionApprovalList() {
  const items = ref<ApiDecision[]>([]);
  const total = ref(0);
  const page = ref(1);
  const limit = ref(50);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const forbidden = ref(false);

  async function fetchList(filters: DecisionApprovalFilters = {}) {
    loading.value = true;
    error.value = null;
    forbidden.value = false;
    try {
      const payload = await decisionApprovalApi.listApprovalItems({
        page: page.value,
        limit: limit.value,
        ...filters,
      });
      items.value = payload.items ?? [];
      total.value = payload.total ?? 0;
      page.value = payload.page ?? page.value;
      limit.value = payload.limit ?? limit.value;
    } catch (err: unknown) {
      const status = (err as { status?: number })?.status;
      if (status === 403) {
        forbidden.value = true;
        error.value = "You do not have permission to view approval items.";
      } else {
        error.value = extractErrorMessage(err, "Failed to load approval items.");
      }
      items.value = [];
      total.value = 0;
    } finally {
      loading.value = false;
    }
  }

  return { items, total, page, limit, loading, error, forbidden, fetchList };
}

export function useDecisionDetail() {
  const decision = ref<ApiDecision | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetch(id: string) {
    loading.value = true;
    error.value = null;
    try {
      decision.value = await decisionApprovalApi.getDetails(id);
    } catch (err: unknown) {
      decision.value = null;
      error.value = extractErrorMessage(err, "Failed to load decision details.");
    } finally {
      loading.value = false;
    }
  }

  function clear() {
    decision.value = null;
    error.value = null;
  }

  return { decision, loading, error, fetch, clear };
}

export function useDecisionBatchAction() {
  const submitting = ref(false);
  const error = ref<string | null>(null);
  const results = ref<ApiBatchApprovalResult[]>([]);
  const succeeded = ref(0);
  const failed = ref(0);

  const hasPartialFailure = computed(() => failed.value > 0 && succeeded.value > 0);
  const allFailed = computed(() => failed.value > 0 && succeeded.value === 0);

  async function approve(decisionNos: string[], comment = "") {
    submitting.value = true;
    error.value = null;
    results.value = [];
    succeeded.value = 0;
    failed.value = 0;
    try {
      const resp = await decisionApprovalApi.batchApprove({ decision_nos: decisionNos, comment });
      results.value = resp.results ?? [];
      succeeded.value = resp.succeeded ?? 0;
      failed.value = resp.failed ?? 0;
      return resp;
    } catch (err: unknown) {
      error.value = extractErrorMessage(err, "Batch approve failed.");
      throw err;
    } finally {
      submitting.value = false;
    }
  }

  async function reject(decisionNos: string[], reason: string) {
    submitting.value = true;
    error.value = null;
    results.value = [];
    succeeded.value = 0;
    failed.value = 0;
    try {
      const resp = await decisionApprovalApi.batchReject({ decision_nos: decisionNos, reason });
      results.value = resp.results ?? [];
      succeeded.value = resp.succeeded ?? 0;
      failed.value = resp.failed ?? 0;
      return resp;
    } catch (err: unknown) {
      error.value = extractErrorMessage(err, "Batch reject failed.");
      throw err;
    } finally {
      submitting.value = false;
    }
  }

  return { submitting, error, results, succeeded, failed, hasPartialFailure, allFailed, approve, reject };
}
