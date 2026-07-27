import type { ApiCashRequestV2 } from "../services/portfolioApi";

/**
 * Whether the given user may cancel this cash request. The cash-requests
 * panel is visible to anyone with portfolio view access (approvers,
 * managers), not just the submitter, but only the submitter may cancel
 * (enforced backend-side too — see CancelCashRequestByCode ->
 * CancelCashRequest's "only the submitter may cancel" check). Kept as a pure
 * function, free of the Pinia auth store, so it is unit-testable without
 * mounting a component or mocking Nuxt runtime state.
 */
export function canCancelCashRequest(
  item: Pick<ApiCashRequestV2, "status" | "submitted_by">,
  currentUserId: string | null | undefined,
): boolean {
  return item.status === "PENDING" && !!currentUserId && item.submitted_by === currentUserId;
}
