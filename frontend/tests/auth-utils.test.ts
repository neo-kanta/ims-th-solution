import { describe, expect, it } from "vitest";
import {
  resolveAuthRedirectReasonKey,
  resolveLoginErrorTranslationKey,
  sanitizeAuthRedirectTarget,
} from "../app/features/auth/lib/auth";

describe("auth utils", () => {
  it("keeps valid internal redirects", () => {
    expect(sanitizeAuthRedirectTarget("/workflow?tab=queue#next")).toBe(
      "/workflow?tab=queue#next",
    );
  });

  it("rejects external absolute redirect targets", () => {
    expect(sanitizeAuthRedirectTarget("https://evil.example/phish")).toBe("/");
  });

  it("rejects protocol-relative redirect targets", () => {
    expect(sanitizeAuthRedirectTarget("//evil.example/phish")).toBe("/");
  });

  it("rejects malformed redirect targets", () => {
    expect(sanitizeAuthRedirectTarget("%E0%A4%A")).toBe("/");
    expect(sanitizeAuthRedirectTarget("/\\evil")).toBe("/");
  });

  it("maps known auth redirect reasons to translation keys", () => {
    expect(resolveAuthRedirectReasonKey("token_expired")).toBe(
      "auth.sessionExpired",
    );
    expect(resolveAuthRedirectReasonKey("session_restore_failed")).toBe(
      "auth.sessionExpired",
    );
    expect(resolveAuthRedirectReasonKey("unknown")).toBeNull();
  });

  it("maps common backend auth errors to translation keys", () => {
    expect(resolveLoginErrorTranslationKey("invalid credentials")).toBe(
      "auth.invalidCredentials",
    );
    expect(resolveLoginErrorTranslationKey("account locked")).toBe(
      "auth.accountLocked",
    );
    expect(resolveLoginErrorTranslationKey("session expired")).toBe(
      "auth.sessionExpired",
    );
    expect(resolveLoginErrorTranslationKey("unexpected backend failure")).toBeNull();
  });
});
