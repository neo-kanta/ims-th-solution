import { describe, expect, it } from "vitest";

import { collectMessageKeys, resolveMessage } from "../app/shared/i18n/core";
import { messages } from "../app/shared/i18n/messages";

describe("i18n message registry", () => {
  it("keeps the locale key set aligned across all locales", () => {
    const englishKeys = collectMessageKeys(messages.en).sort();

    expect(collectMessageKeys(messages.th).sort()).toEqual(englishKeys);
    expect(collectMessageKeys(messages.zh).sort()).toEqual(englishKeys);
  });

  it("includes translated copy for the primary active surfaces", () => {
    expect(resolveMessage(messages.en, "auth.login")).toBeTruthy();
    expect(resolveMessage(messages.en, "dashboardOverview.title")).toBeTruthy();
    expect(resolveMessage(messages.en, "settings.title")).toBeTruthy();
    expect(resolveMessage(messages.en, "shell.workspaceLabel")).toBeTruthy();
    expect(resolveMessage(messages.en, "placeholders.approvalWorkflow.title")).toBeTruthy();
  });
});
