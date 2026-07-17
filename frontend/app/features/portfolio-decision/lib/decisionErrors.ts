/**
 * Error classification shared by the portfolio-decision composables.
 *
 * Decision endpoints answer with the legacy `{error, code?, details?}` shape
 * (backend/platform/httputil/response.go), not the newer request_id-bearing
 * envelope — so any request id we can show comes from the `X-Request-Id`
 * response header, when present.
 *
 * Deliberately duck-types instead of importing OpenApiRequestError from
 * ~/api/openapi (matches the existing useOrderTicket.ts convention) — it
 * keeps this module import-only-relative so it stays loadable in plain
 * Vitest, which has no `~` alias configured.
 */

interface ErrorLike {
  status?: unknown;
  statusCode?: unknown;
  message?: unknown;
  data?: { error?: unknown; message?: unknown; request_id?: unknown };
  details?: { error?: unknown; message?: unknown; request_id?: unknown };
  response?: { headers?: { get?: (name: string) => string | null } };
}

export function extractMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const e = err as ErrorLike;
  if (e.data && typeof e.data === "object") {
    if (typeof e.data.error === "string" && e.data.error.trim()) return e.data.error;
    if (typeof e.data.message === "string" && e.data.message.trim()) return e.data.message;
  }
  if (e.details && typeof e.details === "object") {
    if (typeof e.details.error === "string" && e.details.error.trim()) return e.details.error;
    if (typeof e.details.message === "string" && e.details.message.trim())
      return e.details.message;
  }
  if (typeof e.message === "string" && e.message.trim()) return e.message;
  return fallback;
}

export function statusOf(err: unknown): number | null {
  if (!err || typeof err !== "object") return null;
  const e = err as ErrorLike;
  if (typeof e.status === "number") return e.status;
  if (typeof e.statusCode === "number") return e.statusCode;
  return null;
}

export function requestIdOf(err: unknown): string | null {
  if (!err || typeof err !== "object") return null;
  const e = err as ErrorLike;
  const headerId = e.response?.headers?.get?.("x-request-id");
  if (headerId) return headerId;
  if (typeof e.data?.request_id === "string" && e.data.request_id) return e.data.request_id;
  if (typeof e.details?.request_id === "string" && e.details.request_id) return e.details.request_id;
  return null;
}

/**
 * Builds a user-facing message for a failed decision API call, appending the
 * backend request id (when the response carried one) so support can trace
 * the failure without hiding it from the operator.
 */
export function describeDecisionError(err: unknown, fallback: string): string {
  const message = extractMessage(err, fallback);
  const requestId = requestIdOf(err);
  return requestId ? `${message} (request ${requestId})` : message;
}
