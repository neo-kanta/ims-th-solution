/**
 * Status-aware error classification for the Create Portfolio form.
 *
 * `OpenApiRequestError` (thrown by `unwrapOpenApiResponse` — see
 * `~/api/openapi`) is imported as a type only and detected by duck-typing at
 * runtime, matching `features/compliance/lib/errors.ts`'s established
 * convention: a real value import of `~/api/openapi` would make this module
 * (and its test) unloadable under plain `vitest run`, which has no `~` alias
 * configured.
 */
import type { OpenApiRequestError } from "~/api/openapi";

export type PortfolioCreateErrorKind =
  | "validation"
  | "forbidden"
  | "fund_not_found"
  | "duplicate_code"
  | "business_rule"
  | "unexpected";

export interface ClassifiedPortfolioCreateError {
  kind: PortfolioCreateErrorKind;
  status: number | null;
  message: string;
}

function isOpenApiRequestError(err: unknown): err is OpenApiRequestError {
  return (
    err instanceof Error &&
    err.name === "OpenApiRequestError" &&
    typeof (err as { status?: unknown }).status === "number"
  );
}

function statusOf(err: unknown): number | null {
  if (isOpenApiRequestError(err)) return err.status;
  if (!err || typeof err !== "object") return null;
  const status = (err as { status?: unknown; statusCode?: unknown }).status;
  if (typeof status === "number") return status;
  const statusCode = (err as { statusCode?: unknown }).statusCode;
  return typeof statusCode === "number" ? statusCode : null;
}

function messageOf(err: unknown, fallback: string): string {
  if (isOpenApiRequestError(err)) {
    const details = err.details as { error?: unknown; message?: unknown } | undefined;
    if (details && typeof details === "object") {
      if (typeof details.error === "string" && details.error.trim()) return details.error;
      if (typeof details.message === "string" && details.message.trim()) return details.message;
    }
    return err.message?.trim() || fallback;
  }
  if (!err || typeof err !== "object") return fallback;
  const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
  if (data && typeof data === "object") {
    if (typeof data.error === "string" && data.error.trim()) return data.error;
    if (typeof data.message === "string" && data.message.trim()) return data.message;
  }
  const message = (err as { message?: unknown }).message;
  return typeof message === "string" && message.trim() ? message : fallback;
}

/**
 * Maps a failed `portfolioApi.create` call to a distinct, user-facing error
 * kind by HTTP status: 400 validation, 403 forbidden (no fund access / no
 * manage permission), 404 fund not found, 409 duplicate code, 422 business
 * validation, anything else (or no status at all) unexpected/500.
 */
export function classifyPortfolioCreateError(
  err: unknown,
  fallback: string,
): ClassifiedPortfolioCreateError {
  const status = statusOf(err);
  const message = messageOf(err, fallback);

  switch (status) {
    case 400:
      return { kind: "validation", status, message };
    case 403:
      return { kind: "forbidden", status, message };
    case 404:
      return { kind: "fund_not_found", status, message };
    case 409:
      return { kind: "duplicate_code", status, message };
    case 422:
      return { kind: "business_rule", status, message };
    default:
      return { kind: "unexpected", status, message };
  }
}
