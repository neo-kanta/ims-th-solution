/**
 * Tests for the New Decision page's state resolver (lib/pageState.ts).
 * Kept as a pure function, separate from PortfolioDecisionNewView.vue,
 * because this project's Vitest setup has no Nuxt/component runtime to
 * mount an SFC against.
 */
import { describe, expect, it } from "vitest";

import {
  classifyPortfolioContextError,
  resolveDecisionNewPageState,
} from "../app/features/portfolio-decision/lib/pageState";

describe("classifyPortfolioContextError", () => {
  it("classifies the backend's literal 404 message", () => {
    expect(classifyPortfolioContextError("portfolio not found")).toBe("not-found");
  });

  it("classifies the backend's literal 403 message", () => {
    expect(classifyPortfolioContextError("no access to this portfolio")).toBe("permission-denied");
  });

  it("falls back to unknown for anything else, without guessing", () => {
    expect(classifyPortfolioContextError("internal server error")).toBe("unknown");
    expect(classifyPortfolioContextError(null)).toBe("unknown");
  });
});

describe("resolveDecisionNewPageState", () => {
  it("is 'loading' before any portfolio has loaded and no error is set", () => {
    const state = resolveDecisionNewPageState({
      portfolioCode: "TH-EQ-01",
      loadedPortfolioCode: null,
      error: null,
    });
    expect(state).toEqual({ kind: "loading" });
  });

  it("is 'ready' once the loaded descriptor's own code matches the route's portfolioCode", () => {
    const state = resolveDecisionNewPageState({
      portfolioCode: "TH-EQ-01",
      loadedPortfolioCode: "TH-EQ-01",
      error: null,
    });
    expect(state).toEqual({ kind: "ready" });
  });

  it("stays 'loading' — never 'ready' — when a stale response for a different portfolio is the only thing loaded", () => {
    // Simulates: user was on TH-EQ-01, switched to TH-EQ-02, but a slow
    // response for TH-EQ-01 is the last thing to land in ctx.portfolio.
    const state = resolveDecisionNewPageState({
      portfolioCode: "TH-EQ-02",
      loadedPortfolioCode: "TH-EQ-01",
      error: null,
    });
    expect(state).toEqual({ kind: "loading" });
  });

  it("is 'no-fund' once loaded when the portfolio has no fund bound", () => {
    const state = resolveDecisionNewPageState({
      portfolioCode: "PF-NOFUND",
      loadedPortfolioCode: "PF-NOFUND",
      error: null,
      hasFund: false,
    });
    expect(state).toEqual({ kind: "no-fund" });
  });

  it("defaults hasFund to true when omitted, so pre-existing callers stay 'ready'", () => {
    const state = resolveDecisionNewPageState({
      portfolioCode: "TH-EQ-01",
      loadedPortfolioCode: "TH-EQ-01",
      error: null,
    });
    expect(state).toEqual({ kind: "ready" });
  });

  it("maps a 404 error to 'not-found'", () => {
    const state = resolveDecisionNewPageState({
      portfolioCode: "TH-EQ-01",
      loadedPortfolioCode: null,
      error: "portfolio not found",
    });
    expect(state).toEqual({ kind: "not-found" });
  });

  it("maps a 403 error to 'permission-denied' and preserves the backend message", () => {
    const state = resolveDecisionNewPageState({
      portfolioCode: "TH-EQ-01",
      loadedPortfolioCode: null,
      error: "no access to this portfolio",
    });
    expect(state).toEqual({ kind: "permission-denied", message: "no access to this portfolio" });
  });

  it("maps any other error to a generic retryable 'error' state", () => {
    const state = resolveDecisionNewPageState({
      portfolioCode: "TH-EQ-01",
      loadedPortfolioCode: null,
      error: "network timeout",
    });
    expect(state).toEqual({ kind: "error", message: "network timeout" });
  });
});
