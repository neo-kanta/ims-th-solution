import type {
  AuthUser,
  AuthUserPayload,
  AuthStoreResetState,
  UserPermissions,
} from "../types";
import type { AppTranslationKey } from "../../../shared/i18n/messages";

const INTERNAL_APP_ORIGIN = "https://ims.local";

function decodeBase64Url(value: string): string | null {
  try {
    const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");

    if (typeof atob === "function") {
      return atob(padded);
    }

    return Buffer.from(padded, "base64").toString("utf-8");
  } catch {
    return null;
  }
}

export function getTokenExpiry(token: string | null): string | null {
  if (!token) {
    return null;
  }

  const parts = token.split(".");
  if (parts.length < 2) {
    return null;
  }

  const payloadPart = parts[1];
  if (!payloadPart) {
    return null;
  }

  const payload = decodeBase64Url(payloadPart);
  if (!payload) {
    return null;
  }

  try {
    const parsed = JSON.parse(payload) as { exp?: number };
    return parsed.exp ? new Date(parsed.exp * 1000).toISOString() : null;
  } catch {
    return null;
  }
}

export function createEmptyPermissions(): UserPermissions {
  return {
    functions: [],
    contracts: [],
  };
}

export function createClearedAuthState(
  error: string | null = null,
): AuthStoreResetState {
  return {
    token: null,
    expiresAt: null,
    user: null,
    permissions: createEmptyPermissions(),
    error,
  };
}

export function mapAuthUser(user: AuthUserPayload): AuthUser {
  return {
    id: user.id,
    username: user.username,
    displayName: user.display_name,
    groups: user.groups || [],
  };
}

export function sanitizeAuthRedirectTarget(
  target: unknown,
  fallback = "/",
): string {
  if (typeof target !== "string") {
    return fallback;
  }

  const trimmedTarget = target.trim();
  if (!trimmedTarget) {
    return fallback;
  }

  if (/[\u0000-\u001F\u007F\\]/.test(trimmedTarget)) {
    return fallback;
  }

  let candidate = trimmedTarget;

  if (candidate.includes("%")) {
    try {
      candidate = decodeURIComponent(candidate);
    } catch {
      return fallback;
    }
  }

  if (!candidate.startsWith("/")) {
    return fallback;
  }

  try {
    const parsed = new URL(candidate, INTERNAL_APP_ORIGIN);

    if (parsed.origin !== INTERNAL_APP_ORIGIN || !parsed.pathname.startsWith("/")) {
      return fallback;
    }

    return `${parsed.pathname}${parsed.search}${parsed.hash}` || fallback;
  } catch {
    return fallback;
  }
}

type AuthFeedbackTranslationKey = Extract<
  AppTranslationKey,
  "auth.accountLocked" | "auth.invalidCredentials" | "auth.sessionExpired"
>;

export function resolveAuthRedirectReasonKey(
  reason: unknown,
): AuthFeedbackTranslationKey | null {
  if (reason === "token_expired" || reason === "session_restore_failed") {
    return "auth.sessionExpired";
  }

  return null;
}

export function resolveLoginErrorTranslationKey(
  error: unknown,
): AuthFeedbackTranslationKey | null {
  if (typeof error !== "string") {
    return null;
  }

  const normalizedError = error.trim().toLowerCase();
  if (!normalizedError) {
    return null;
  }

  if (normalizedError.includes("lock")) {
    return "auth.accountLocked";
  }

  if (normalizedError.includes("expired")) {
    return "auth.sessionExpired";
  }

  if (
    normalizedError.includes("invalid credential")
    || normalizedError.includes("invalid username")
    || normalizedError.includes("invalid password")
    || normalizedError.includes("unauthorized")
    || normalizedError.includes("unauthenticated")
    || normalizedError.includes("authentication failed")
  ) {
    return "auth.invalidCredentials";
  }

  return null;
}
