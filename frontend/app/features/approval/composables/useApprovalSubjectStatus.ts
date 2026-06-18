import { ref } from "vue";

import { approvalApi, approvalErrorMessage } from "../services/approvalApi";
import type { ApprovalSubjectStatus } from "../types";

/** Lightweight composable for fetching a subject's approval status panel data. */
export function useApprovalSubjectStatus() {
  const status = ref<ApprovalSubjectStatus | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetch(subjectType: string, subjectId: string) {
    if (!subjectId) return;
    loading.value = true;
    error.value = null;
    try {
      status.value = await approvalApi.subjectStatus(subjectType, subjectId);
    } catch (err) {
      error.value = approvalErrorMessage(err, "Failed to load approval status.");
    } finally {
      loading.value = false;
    }
  }

  return { status, loading, error, fetch };
}
