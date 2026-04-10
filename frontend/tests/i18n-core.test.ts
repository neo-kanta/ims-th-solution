import { describe, expect, it } from "vitest";
import {
  interpolateMessage,
  normalizeTranslateArgs,
  resolveMessage,
  translateMessage,
  type MessageCatalog,
} from "../app/shared/i18n/core";

const messages: MessageCatalog = {
  en: {
    auth: {
      failedAttempts: "Failed sign-in attempts: {count}",
      welcome: "Welcome",
    },
  },
  th: {
    auth: {
      welcome: "Sawatdee",
    },
  },
  zh: {},
};

describe("i18n core", () => {
  it("resolves nested messages by dot key", () => {
    expect(resolveMessage(messages.en, "auth.welcome")).toBe("Welcome");
    expect(resolveMessage(messages.en, "auth.missing")).toBeUndefined();
  });

  it("interpolates placeholders with provided params", () => {
    expect(
      interpolateMessage("Failed sign-in attempts: {count}", { count: 3 }),
    ).toBe("Failed sign-in attempts: 3");
  });

  it("normalizes params and fallback arguments", () => {
    expect(normalizeTranslateArgs("Fallback copy")).toEqual({
      fallback: "Fallback copy",
      params: undefined,
    });

    expect(
      normalizeTranslateArgs({ count: 2 }, "Failed sign-in attempts: {count}"),
    ).toEqual({
      fallback: "Failed sign-in attempts: {count}",
      params: { count: 2 },
    });
  });

  it("falls back to English when the active locale is missing a key", () => {
    expect(
      translateMessage({
        key: "auth.failedAttempts",
        locale: "th",
        messages,
        params: { count: 4 },
      }),
    ).toBe("Failed sign-in attempts: 4");
  });

  it("falls back to the provided string when all locale messages miss the key", () => {
    expect(
      translateMessage({
        fallback: "Default copy",
        key: "auth.unknown",
        locale: "zh",
        messages,
      }),
    ).toBe("Default copy");
  });
});
