import { ref } from "vue";

import { approvalApi, approvalErrorMessage, approvalErrorStatus } from "../services/approvalApi";
import type { ApprovalInboxItem } from "../types";

/** Loads the authenticated user's approval inbox from the real backend API. */
export function useApprovalInbox() {
  const items = ref<ApprovalInboxItem[]>([]);
  const total = ref(0);
  const loading = ref(false);
  const error = ref<string | null>(null);
  const forbidden = ref(false);

  async function fetchInbox(status = "PENDING") {
    loading.value = true;
    error.value = null;
    forbidden.value = false;
    try {
      const payload = await approvalApi.inbox({ status, limit: 100 });
      items.value = payload.items ?? [];
      total.value = payload.total ?? 0;
    } catch (err) {
      if (approvalErrorStatus(err) === 403) forbidden.value = true;
      error.value = approvalErrorMessage(err, "Failed to load approval inbox.");
      items.value = [];
      total.value = 0;
    } finally {
      loading.value = false;
    }
  }

  return { items, total, loading, error, forbidden, fetchInbox };
}
