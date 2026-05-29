import { ref } from "vue";

import { approvalApi, approvalErrorMessage, approvalErrorStatus } from "../services/approvalApi";
import type { ApprovalRequestDetail } from "../types";

/** Loads a single approval request detail (request + tasks + timeline + signatures). */
export function useApprovalRequest() {
  const detail = ref<ApprovalRequestDetail | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const forbidden = ref(false);
  const notFound = ref(false);

  async function fetchRequest(id: string) {
    loading.value = true;
    error.value = null;
    forbidden.value = false;
    notFound.value = false;
    try {
      detail.value = await approvalApi.request(id);
    } catch (err) {
      const status = approvalErrorStatus(err);
      if (status === 403) forbidden.value = true;
      if (status === 404) notFound.value = true;
      detail.value = null;
      error.value = approvalErrorMessage(err, "Failed to load approval request.");
    } finally {
      loading.value = false;
    }
  }

  return { detail, loading, error, forbidden, notFound, fetchRequest };
}
