import { ref } from "vue";
import {
  decisionApi,
  type ApiDecision,
  type ApiCreateDecisionRequest,
  type ApiBatchApprovalRequest,
  type ApiBatchRejectionRequest,
  type ApiBatchApprovalResponse,
} from "../services/decisionApi";

export function useDecisionMutations() {
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function create(
    body: ApiCreateDecisionRequest,
  ): Promise<ApiDecision | null> {
    loading.value = true;
    error.value = null;
    try {
      return await decisionApi.createDecision(body);
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to create decision";
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function submit(id: string): Promise<ApiDecision | null> {
    loading.value = true;
    error.value = null;
    try {
      return await decisionApi.submitDecision(id);
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to submit decision";
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function cancel(
    id: string,
    reason: string,
  ): Promise<ApiDecision | null> {
    loading.value = true;
    error.value = null;
    try {
      return await decisionApi.cancelDecision(id, reason);
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to cancel decision";
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function batchApprove(
    body: ApiBatchApprovalRequest,
  ): Promise<ApiBatchApprovalResponse | null> {
    loading.value = true;
    error.value = null;
    try {
      return await decisionApi.batchApprove(body);
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to batch approve";
      return null;
    } finally {
      loading.value = false;
    }
  }

  async function batchReject(
    body: ApiBatchRejectionRequest,
  ): Promise<ApiBatchApprovalResponse | null> {
    loading.value = true;
    error.value = null;
    try {
      return await decisionApi.batchReject(body);
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to batch reject";
      return null;
    } finally {
      loading.value = false;
    }
  }

  return { loading, error, create, submit, cancel, batchApprove, batchReject };
}
