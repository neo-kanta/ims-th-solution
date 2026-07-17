import { describe, expect, it } from "vitest";

import { collectMessageKeys, resolveMessage, translateMessage } from "../app/shared/i18n/core";
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
    expect(resolveMessage(messages.en, "portfolio.page.loading")).toBeTruthy();
    expect(resolveMessage(messages.en, "portfolio.kpis.todayPnl")).toBeTruthy();
    expect(resolveMessage(messages.en, "portfolio.kpis.summaryIncompleteCoverage")).toBeTruthy();
    expect(resolveMessage(messages.en, "dashboardOverview.metricIncompleteCoverage")).toBeTruthy();
    expect(resolveMessage(messages.en, "portfolio.decisionNew.validation.instrumentRequired")).toBeTruthy();
    expect(resolveMessage(messages.en, "operator.directory.workflows.op01.title")).toBeTruthy();
  });

  it("ships real Thai and Chinese copy for the portfolio/operator surfaces", () => {
    const keys = [
      "portfolio.page.loading",
      "portfolio.kpis.summaryNoData",
      "portfolio.kpis.summaryIncompleteCoverage",
      "dashboardOverview.metricIncompleteCoverage",
      "portfolio.decisionNew.validation.instrumentRequired",
      "operator.directory.title",
      "operator.review.previewEmpty",
    ];

    for (const key of keys) {
      const english = resolveMessage(messages.en, key);
      expect(resolveMessage(messages.th, key)).toBeTruthy();
      expect(resolveMessage(messages.zh, key)).toBeTruthy();
      expect(resolveMessage(messages.th, key)).not.toBe(english);
      expect(resolveMessage(messages.zh, key)).not.toBe(english);
    }
  });

  it("resolves interpolation without leaking raw translation keys", () => {
    const key = "operator.operation.batchApproveTitle";
    const rendered = translateMessage({
      key,
      locale: "th",
      messages,
      params: { count: 3 },
    });

    expect(rendered).toContain("3");
    expect(rendered).not.toBe(key);
    expect(rendered).not.toContain("{count}");
  });

  it("resolves AUM coverage copy without leaking placeholders or raw keys", () => {
    const key = "portfolio.kpis.summaryIncompleteCoverage";
    for (const locale of ["en", "th", "zh"] as const) {
      const rendered = translateMessage({
        key,
        locale,
        messages,
        params: {
          included: 2,
          total: 3,
          excluded: 1,
          currency: "THB",
        },
      });

      expect(rendered).toContain("2");
      expect(rendered).toContain("3");
      expect(rendered).toContain("THB");
      expect(rendered).not.toBe(key);
      expect(rendered).not.toMatch(/\{(?:included|total|excluded|currency)\}/);
    }
  });
});
