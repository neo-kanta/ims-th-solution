/**
 * Status-aware error classification for the Portfolio Settings update form.
 * Mirrors `portfolioCreateErrors.ts`'s duck-typed `OpenApiRequestError`
 * detection so this module stays loadable under bare `vitest run` (no `~`
 * alias configured — see `features/compliance/lib/errors.ts`).
 */
import type { OpenApiRequestError } from "~/api/openapi";

export type PortfolioUpdateErrorKind =
  | "validation"
  | "forbidden"
  | "not_found"
  | "version_conflict"
  | "business_rule"
  | "unexpected";

export interface ClassifiedPortfolioUpdateError {
  kind: PortfolioUpdateErrorKind;
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
 * Maps a failed `portfolioApi.update` call to a distinct error kind: 400
 * validation, 403 forbidden (no data scope / no manage permission), 404 not
 * found, 409 optimistic-version conflict (the portfolio changed since it was
 * loaded — reload, do not silently overwrite), 422 business validation,
 * anything else unexpected.
 */
export function classifyPortfolioUpdateError(
  err: unknown,
  fallback: string,
): ClassifiedPortfolioUpdateError {
  const status = statusOf(err);
  const message = messageOf(err, fallback);

  switch (status) {
    case 400:
      return { kind: "validation", status, message };
    case 403:
      return { kind: "forbidden", status, message };
    case 404:
      return { kind: "not_found", status, message };
    case 409:
      return { kind: "version_conflict", status, message };
    case 422:
      return { kind: "business_rule", status, message };
    default:
      return { kind: "unexpected", status, message };
  }
}
