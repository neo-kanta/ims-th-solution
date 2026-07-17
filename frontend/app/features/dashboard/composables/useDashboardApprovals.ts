import { readonly, ref } from "vue";

import { approvalApi } from "~/features/approval/services/approvalApi";
import type { ApprovalInboxItem } from "~/features/approval/types";

import type { DashboardOverviewPendingApproval } from "../types";

/**
 * Returns a human-readable countdown string from now to `dueAt`.
 * Returns "" when dueAt is absent. Returns "Overdue" when the deadline has
 * passed. Otherwise returns "Xh Ym left" (or "Nm left" when under an hour).
 */
export function formatDueLabel(dueAt: string | undefined): string {
  if (!dueAt) return "";
  const diffMs = new Date(dueAt).getTime() - Date.now();
  if (diffMs <= 0) return "Overdue";
  const minutes = Math.floor(diffMs / 60_000);
  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  if (hours > 0) {
    return `${hours}h ${remainingMinutes}m left`;
  }
  return `${minutes}m left`;
}

/**
 * Returns true when the due date is in the future and within 2 hours of now.
 */
export function isUrgentDue(dueAt: string | undefined): boolean {
  if (!dueAt) return false;
  const diffMs = new Date(dueAt).getTime() - Date.now();
  return diffMs > 0 && diffMs <= 2 * 60 * 60 * 1_000;
}

function mapInboxItem(item: ApprovalInboxItem): DashboardOverviewPendingApproval {
  return {
    id: item.task?.id ?? "",
    contractCode:
      item.request?.subject?.display_label ?? item.request?.request_number ?? "",
    valueLabel: "",
    dueLabel: formatDueLabel(item.task?.due_at),
    isUrgent: isUrgentDue(item.task?.due_at),
  };
}

export function useDashboardApprovals() {
  const items = ref<DashboardOverviewPendingApproval[]>([]);
  const total = ref(0);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchApprovals(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const result = await approvalApi.inbox({ limit: 50 });
      items.value = (result.items ?? []).map(mapInboxItem);
      total.value = result.total ?? 0;
    } catch (err) {
      items.value = [];
      total.value = 0;
      error.value = null;
    } finally {
      loading.value = false;
    }
  }

  return {
    items: readonly(items),
    total: readonly(total),
    loading: readonly(loading),
    error: readonly(error),
    fetchApprovals,
  };
}
