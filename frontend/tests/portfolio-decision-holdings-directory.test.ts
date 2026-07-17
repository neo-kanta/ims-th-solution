/**
 * Tests for the holdings/cash directory composable that backs the SELL
 * "owned positions" panel. HoldingResponse only carries instrument_id
 * (no ticker/name), so this composable enriches each holding via
 * GET /investment/instruments/{id} — this test locks that behaviour so a
 * future change can't silently regress to showing raw UUIDs again.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const getHoldingsMock = vi.fn();
const getCashMock = vi.fn();
const getInstrumentMock = vi.fn();

vi.mock("../app/features/portfolio-workspace/services/portfolioApi", () => ({
  portfolioApi: {
    getHoldings: (...args: unknown[]) => getHoldingsMock(...args),
    getCash: (...args: unknown[]) => getCashMock(...args),
  },
}));

vi.mock("../app/features/investment-ledger/services/investmentLedgerApi", () => ({
  investmentLedgerApi: {
    getInstrument: (...args: unknown[]) => getInstrumentMock(...args),
  },
}));

import { usePortfolioHoldingsDirectory } from "../app/features/portfolio-decision/composables/usePortfolioHoldingsDirectory";

beforeEach(() => {
  getHoldingsMock.mockReset();
  getCashMock.mockReset();
  getInstrumentMock.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("usePortfolioHoldingsDirectory", () => {
  it("enriches each holding with its instrument (ticker/name), never exposing the raw instrument_id as the label", async () => {
    getHoldingsMock.mockResolvedValueOnce([
      { instrument_id: "11111111-1111-1111-1111-111111111111", quantity: "500", average_cost: "30.2" },
    ]);
    getCashMock.mockResolvedValueOnce([{ currency: "THB", balance: "1000000" }]);
    getInstrumentMock.mockResolvedValueOnce({
      id: "11111111-1111-1111-1111-111111111111",
      primary_ticker: "PTT",
      name: "PTT Public Company Limited",
      currency: "THB",
    });

    const dir = usePortfolioHoldingsDirectory();
    await dir.load("TH-EQ-01");

    expect(dir.holdings.value).toHaveLength(1);
    expect(dir.holdings.value[0]?.instrument?.primary_ticker).toBe("PTT");
    expect(dir.holdings.value[0]?.quantityNumeric).toBe(500);
  });

  it("holdingFor() and cashFor() look up by instrument id / currency", async () => {
    getHoldingsMock.mockResolvedValueOnce([
      { instrument_id: "id-1", quantity: "10", average_cost: "5" },
    ]);
    getCashMock.mockResolvedValueOnce([{ currency: "THB", balance: "500" }]);
    getInstrumentMock.mockResolvedValueOnce({ id: "id-1", primary_ticker: "AOT" });

    const dir = usePortfolioHoldingsDirectory();
    await dir.load("TH-EQ-01");

    expect(dir.holdingFor("id-1")?.instrument?.primary_ticker).toBe("AOT");
    expect(dir.holdingFor("missing")).toBeNull();
    expect(dir.cashFor("THB")?.balance).toBe("500");
    expect(dir.cashFor("USD")).toBeNull();
  });

  it("still shows the holding (with a null instrument) when instrument enrichment fails, instead of dropping the position", async () => {
    getHoldingsMock.mockResolvedValueOnce([
      { instrument_id: "id-2", quantity: "10", average_cost: "5" },
    ]);
    getCashMock.mockResolvedValueOnce([]);
    getInstrumentMock.mockRejectedValueOnce(new Error("not found"));

    const dir = usePortfolioHoldingsDirectory();
    await dir.load("TH-EQ-01");

    expect(dir.holdings.value).toHaveLength(1);
    expect(dir.holdings.value[0]?.instrument).toBeNull();
    expect(dir.error.value).toBeNull();
  });

  it("surfaces an error and clears state when the holdings/cash fetch itself fails", async () => {
    getHoldingsMock.mockRejectedValueOnce({ status: 403, data: { error: "forbidden" } });
    getCashMock.mockResolvedValueOnce([]);

    const dir = usePortfolioHoldingsDirectory();
    await dir.load("TH-EQ-01");

    expect(dir.holdings.value).toHaveLength(0);
    expect(dir.error.value).toMatch(/forbidden/i);
  });
});

describe("usePortfolioHoldingsDirectory portfolio-switch stale-response protection", () => {
  it("does not let a slow response for a previous portfolio overwrite a newer portfolio's holdings", async () => {
    let resolveFirst!: (value: unknown[]) => void;
    const firstHoldings = new Promise<unknown[]>((resolve) => {
      resolveFirst = resolve;
    });
    getHoldingsMock
      .mockReturnValueOnce(firstHoldings)
      .mockResolvedValueOnce([{ instrument_id: "id-2", quantity: "20", average_cost: "8" }]);
    getCashMock
      .mockResolvedValueOnce([{ currency: "THB", balance: "1000" }])
      .mockResolvedValueOnce([{ currency: "USD", balance: "2000" }]);
    getInstrumentMock.mockResolvedValue({ id: "id-2", primary_ticker: "AOT" });

    const dir = usePortfolioHoldingsDirectory();

    // Slow load for the portfolio the user is navigating away from.
    const firstLoad = dir.load("TH-EQ-01");
    // Fast load for the portfolio the user actually switched to.
    await dir.load("TH-EQ-02");

    expect(dir.cash.value).toEqual([{ currency: "USD", balance: "2000" }]);

    // The stale first response now resolves — it must not clobber TH-EQ-02's data.
    resolveFirst([{ instrument_id: "id-1", quantity: "500", average_cost: "30" }]);
    await firstLoad;

    expect(dir.cash.value).toEqual([{ currency: "USD", balance: "2000" }]);
    expect(dir.holdings.value.map((h) => h.instrumentId)).toEqual(["id-2"]);
  });

  it("clears the previous portfolio's holdings/cash immediately when a new load starts, instead of leaving stale data visible", async () => {
    getHoldingsMock.mockResolvedValueOnce([
      { instrument_id: "id-1", quantity: "500", average_cost: "30" },
    ]);
    getCashMock.mockResolvedValueOnce([{ currency: "THB", balance: "1000" }]);
    getInstrumentMock.mockResolvedValueOnce({ id: "id-1", primary_ticker: "PTT" });

    const dir = usePortfolioHoldingsDirectory();
    await dir.load("TH-EQ-01");
    expect(dir.holdings.value).toHaveLength(1);

    let resolveSecondHoldings!: (value: unknown[]) => void;
    let resolveSecondCash!: (value: unknown[]) => void;
    getHoldingsMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveSecondHoldings = resolve;
      }),
    );
    getCashMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveSecondCash = resolve;
      }),
    );

    const secondLoad = dir.load("TH-EQ-02");
    expect(dir.holdings.value).toHaveLength(0);
    expect(dir.cash.value).toHaveLength(0);

    resolveSecondHoldings([]);
    resolveSecondCash([]);
    await secondLoad;
  });
});
