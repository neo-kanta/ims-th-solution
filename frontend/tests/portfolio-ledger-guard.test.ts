import { describe, expect, it } from "vitest";

import {
  canEnterLedgerTransaction,
  ledgerIntentFor,
} from "../app/features/portfolio-workspace/lib/ledgerGuard";

describe("canEnterLedgerTransaction", () => {
  it("blocks MODEL portfolios", () => {
    expect(canEnterLedgerTransaction("MODEL")).toBe(false);
  });

  it("allows LIVE portfolios", () => {
    expect(canEnterLedgerTransaction("LIVE")).toBe(true);
  });

  it("allows SIMULATION portfolios", () => {
    expect(canEnterLedgerTransaction("SIMULATION")).toBe(true);
  });

  it("defaults to allowed for an unrecognized type rather than silently blocking", () => {
    expect(canEnterLedgerTransaction("SOMETHING_NEW")).toBe(true);
  });
});

describe("ledgerIntentFor", () => {
  it("returns BLOCKED for MODEL", () => {
    expect(ledgerIntentFor("MODEL")).toBe("BLOCKED");
  });

  it("returns PAPER for SIMULATION so posting copy never implies a real trade", () => {
    expect(ledgerIntentFor("SIMULATION")).toBe("PAPER");
  });

  it("returns OFFICIAL for LIVE", () => {
    expect(ledgerIntentFor("LIVE")).toBe("OFFICIAL");
  });
});
