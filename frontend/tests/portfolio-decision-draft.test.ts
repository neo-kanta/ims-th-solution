import { describe, expect, it } from "vitest";

import { useDecisionDraft } from "../app/features/portfolio-decision/composables/useDecisionDraft";
import type { ApiInstrument } from "../app/features/investment-ledger/services/investmentLedgerApi";

const ptt: ApiInstrument = {
  id: "11111111-1111-1111-1111-111111111111",
  primary_ticker: "PTT",
  name: "PTT Public Company Limited",
  primary_exchange: "SET",
  currency: "THB",
};

describe("useDecisionDraft instrument selection", () => {
  it("stores the canonical instrument id and derives instrument_code from primary_ticker", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);

    expect(draftCtl.draft.instrumentId).toBe(ptt.id);
    expect(draftCtl.draft.instrumentCode).toBe("PTT");
    expect(draftCtl.selectedInstrument.value).toEqual(ptt);
  });

  it("defaults currency and exchange from the selected instrument when unset", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);

    expect(draftCtl.draft.currency).toBe("THB");
    expect(draftCtl.draft.exchange).toBe("SET");
  });

  it("does not overwrite a currency the user already typed", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.patch({ currency: "USD" });
    draftCtl.selectInstrument(ptt);

    expect(draftCtl.draft.currency).toBe("USD");
  });

  it("clearInstrument() lets the user change/clear the selection", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);
    draftCtl.clearInstrument();

    expect(draftCtl.selectedInstrument.value).toBeNull();
    expect(draftCtl.draft.instrumentId).toBeNull();
    expect(draftCtl.draft.instrumentCode).toBe("");
  });

  it("never allows submission with unselected arbitrary text (typing alone doesn't set instrumentCode)", () => {
    const draftCtl = useDecisionDraft();
    // Simulates a user typing in the search box without picking a result —
    // the combobox's local search query is separate state and is never
    // patched into draft.instrumentCode.
    expect(draftCtl.draft.instrumentCode).toBe("");
    expect(draftCtl.errors.value.instrument).toBeDefined();
  });
});

describe("useDecisionDraft.buildCreateBody", () => {
  it("builds a BUY request body with quantity and limit price, no fund_id/portfolio_id/contract_id", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);
    draftCtl.setSide("BUY");
    draftCtl.patch({
      quantity: "1000",
      limitPrice: "35.00",
      businessDate: "2026-07-14",
      rationale: "Portfolio rebalance",
    });

    const body = draftCtl.buildCreateBody();

    expect(body).toEqual({
      instrument_code: "PTT",
      business_date: "2026-07-14",
      side: "BUY",
      currency: "THB",
      instrument_id: ptt.id,
      quantity: "1000",
      limit_price: "35.00",
      exchange: "SET",
      rationale: "Portfolio rebalance",
    });
    expect(body).not.toHaveProperty("fund_id");
    expect(body).not.toHaveProperty("portfolio_id");
    expect(body).not.toHaveProperty("contract_id");
    expect(body).not.toHaveProperty("amount");
  });

  it("builds a SELL request body and drops empty optional fields", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);
    draftCtl.setSide("SELL");
    draftCtl.patch({ quantity: "200", businessDate: "2026-07-14" });

    const body = draftCtl.buildCreateBody();

    expect(body.side).toBe("SELL");
    expect(body.quantity).toBe("200");
    expect(body.limit_price).toBeUndefined();
    expect(body.rationale).toBeUndefined();
    expect(body).not.toHaveProperty("fund_id");
    expect(body).not.toHaveProperty("portfolio_id");
    expect(body).not.toHaveProperty("contract_id");
  });

  it("supports an amount-only order (no quantity)", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);
    draftCtl.patch({ amount: "50000", businessDate: "2026-07-14" });

    const body = draftCtl.buildCreateBody();
    expect(body.amount).toBe("50000");
    expect(body.quantity).toBeUndefined();
  });

  it("sends quantity and amount without inventing a zero limit price", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);
    draftCtl.patch({ quantity: "1", amount: "11", businessDate: "2026-07-16" });

    const body = draftCtl.buildCreateBody();
    expect(body.quantity).toBe("1");
    expect(body.amount).toBe("11");
    expect(body.limit_price).toBeUndefined();
  });

  it("uppercases instrument_code and currency at the API boundary", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.patch({
      instrumentCode: "ptt",
      currency: "thb",
      quantity: "10",
      businessDate: "2026-07-14",
    });

    const body = draftCtl.buildCreateBody();
    expect(body.instrument_code).toBe("PTT");
    expect(body.currency).toBe("THB");
  });
});

describe("useDecisionDraft SELL availability guard", () => {
  it("flags an oversell once availableQuantity is set from the portfolio's holding", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);
    draftCtl.setSide("SELL");
    draftCtl.availableQuantity.value = 50;
    draftCtl.patch({ quantity: "100", businessDate: "2026-07-14" });

    expect(draftCtl.errors.value.quantity).toMatch(/exceeds available holding/i);
    expect(draftCtl.isValid.value).toBe(false);
  });
});

describe("useDecisionDraft SELL ownership guard", () => {
  it("blocks the draft once isOwnedInstrument is explicitly false, independent of quantity", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);
    draftCtl.setSide("SELL");
    draftCtl.isOwnedInstrument.value = false;
    draftCtl.patch({ quantity: "10", businessDate: "2026-07-14" });

    expect(draftCtl.errors.value.instrument).toMatch(/does not hold this instrument/i);
    expect(draftCtl.isValid.value).toBe(false);
  });

  it("clearInstrument() resets the ownership flag so a fresh pick isn't pre-blocked", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);
    draftCtl.setSide("SELL");
    draftCtl.isOwnedInstrument.value = false;

    draftCtl.clearInstrument();

    expect(draftCtl.isOwnedInstrument.value).toBeNull();
  });

  it("setSide('BUY') resets the ownership flag since the gate only applies to SELL", () => {
    const draftCtl = useDecisionDraft();
    draftCtl.selectInstrument(ptt);
    draftCtl.setSide("SELL");
    draftCtl.isOwnedInstrument.value = false;

    draftCtl.setSide("BUY");

    expect(draftCtl.isOwnedInstrument.value).toBeNull();
    expect(draftCtl.errors.value.instrument).toBeUndefined();
  });
});
