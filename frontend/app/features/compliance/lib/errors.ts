/**
 * Shared error-message extraction for the compliance feature.
 *
 * Two distinct error shapes reach callers here: `OpenApiRequestError` thrown
 * by `unwrapOpenApiResponse`/`assertOpenApiResponse` (generated-client reads
 * and the breach override write), and the older ofetch-style error object
 * thrown by `useApi()` (still used by a few legacy compliance mutations).
 * Both carry a useful backend message; this helper reads whichever shape it
 * gets without assuming one over the other.
 *
 * `OpenApiRequestError` is imported as a type only (erased at compile time)
 * and detected by duck-typing at runtime — `~/api/openapi` has a top-level
 * `#imports` dependency that only resolves inside a running Nuxt app, and a
 * real value import here would make this module (and anything that imports
 * it) unloadable in plain Vitest unit tests.
 */
import type { OpenApiRequestError } from "~/api/openapi";

function isOpenApiRequestError(err: unknown): err is OpenApiRequestError {
  return (
    err instanceof Error &&
    err.name === "OpenApiRequestError" &&
    typeof (err as { status?: unknown }).status === "number"
  );
}

export function extractComplianceErrorMessage(
  err: unknown,
  fallback: string,
): string {
  if (isOpenApiRequestError(err)) {
    const message = err.message?.trim();
    return message ? `${err.status} · ${message}` : `${err.status} · ${fallback}`;
  }

  if (!err || typeof err !== "object") return fallback;

  const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
  if (data) {
    if (typeof data.error === "string" && data.error.trim()) return data.error;
    if (typeof data.message === "string" && data.message.trim()) return data.message;
  }

  const status = (err as { status?: unknown; statusCode?: unknown }).status;
  const code =
    typeof status === "number"
      ? status
      : typeof (err as { statusCode?: unknown }).statusCode === "number"
        ? ((err as { statusCode?: unknown }).statusCode as number)
        : null;

  const message = (err as { message?: unknown }).message;
  const base = typeof message === "string" && message.trim() ? message : fallback;
  return code ? `${code} · ${base}` : base;
}

export function complianceErrorStatus(err: unknown): number | null {
  if (isOpenApiRequestError(err)) return err.status;
  if (!err || typeof err !== "object") return null;
  const status = (err as { status?: unknown; statusCode?: unknown }).status;
  if (typeof status === "number") return status;
  const statusCode = (err as { statusCode?: unknown }).statusCode;
  return typeof statusCode === "number" ? statusCode : null;
}

export function isComplianceConflict(err: unknown): boolean {
  return complianceErrorStatus(err) === 409;
}
