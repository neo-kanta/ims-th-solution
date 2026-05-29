import { ref } from "vue";

import { approvalApi, approvalErrorMessage } from "../services/approvalApi";

/** Approve / reject / withdraw / cancel actions against the real backend. */
export function useApprovalActions() {
  const submitting = ref(false);
  const error = ref<string | null>(null);

  async function run<T>(fn: () => Promise<T>, fallback: string): Promise<T> {
    submitting.value = true;
    error.value = null;
    try {
      return await fn();
    } catch (err) {
      error.value = approvalErrorMessage(err, fallback);
      throw err;
    } finally {
      submitting.value = false;
    }
  }

  return {
    submitting,
    error,
    approve: (taskId: string, comment: string) =>
      run(() => approvalApi.approve(taskId, { comment }), "Failed to approve task."),
    reject: (taskId: string, reason: string) =>
      run(() => approvalApi.reject(taskId, { reason }), "Failed to reject task."),
    withdraw: (requestId: string) =>
      run(() => approvalApi.withdraw(requestId), "Failed to withdraw request."),
    cancel: (requestId: string) =>
      run(() => approvalApi.cancel(requestId), "Failed to cancel request."),
  };
}
