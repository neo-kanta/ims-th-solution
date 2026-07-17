/**
 * Tests for the portfolio-decision instrument search composable: debounce,
 * stale-response protection, and keyboard navigation. The service module is
 * mocked because it relies on Nuxt's runtime client — same pattern as
 * investment-ledger-order-ticket.test.ts.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const listInstrumentsMock = vi.fn();

vi.mock("../app/api/openapi", () => ({
  useOpenApiClient: () => ({ GET: vi.fn(), POST: vi.fn() }),
  unwrapOpenApiResponse: <T,>(value: { data?: T }) => value.data as T,
}));

vi.mock("../app/features/investment-ledger/services/investmentLedgerApi", () => ({
  investmentLedgerApi: {
    listInstruments: (...args: unknown[]) => listInstrumentsMock(...args),
  },
}));

import { useInstrumentSearch } from "../app/features/portfolio-decision/composables/useInstrumentSearch";

function instrument(id: string, ticker: string) {
  return { id, primary_ticker: ticker, name: `${ticker} Public Company`, currency: "THB" };
}

beforeEach(() => {
  listInstrumentsMock.mockReset();
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
  vi.clearAllMocks();
});

describe("useInstrumentSearch debounce", () => {
  it("does not call the API until ~280ms of silence after the last keystroke", async () => {
    listInstrumentsMock.mockResolvedValue({ items: [instrument("1", "PTT")] });
    const search = useInstrumentSearch();

    search.setQuery("P");
    await vi.advanceTimersByTimeAsync(100);
    search.setQuery("PT");
    await vi.advanceTimersByTimeAsync(100);
    search.setQuery("PTT");
    expect(listInstrumentsMock).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(280);
    expect(listInstrumentsMock).toHaveBeenCalledTimes(1);
    expect(listInstrumentsMock).toHaveBeenCalledWith(
      expect.objectContaining({ search: "PTT" }),
    );
  });

  it("populates items and resets the active index on a fresh result set", async () => {
    listInstrumentsMock.mockResolvedValue({
      items: [instrument("1", "PTT"), instrument("2", "PTTEP")],
    });
    const search = useInstrumentSearch();
    search.setQuery("PT");
    await vi.advanceTimersByTimeAsync(300);

    expect(search.items.value).toHaveLength(2);
    expect(search.activeIndex.value).toBe(0);
    expect(search.activeItem.value?.id).toBe("1");
  });
});

describe("useInstrumentSearch stale-response protection", () => {
  it("does not let an older, slower request overwrite a newer, faster one", async () => {
    let resolveFirst!: (value: { items: unknown[] }) => void;
    const firstPromise = new Promise((resolve) => {
      resolveFirst = resolve;
    });
    listInstrumentsMock
      .mockReturnValueOnce(firstPromise)
      .mockResolvedValueOnce({ items: [instrument("2", "AOT")] });

    const search = useInstrumentSearch();

    // First (slow) search request.
    search.setQuery("A");
    await vi.advanceTimersByTimeAsync(300);
    expect(listInstrumentsMock).toHaveBeenCalledTimes(1);

    // Second (fast) search request completes before the first resolves.
    search.setQuery("AO");
    await vi.advanceTimersByTimeAsync(300);
    expect(listInstrumentsMock).toHaveBeenCalledTimes(2);
    expect(search.items.value.map((i) => i.id)).toEqual(["2"]);

    // The slow first request now resolves — it must not clobber the newer result.
    resolveFirst({ items: [instrument("1", "PTT")] });
    await Promise.resolve();
    await Promise.resolve();

    expect(search.items.value.map((i) => i.id)).toEqual(["2"]);
  });
});

describe("useInstrumentSearch keyboard navigation", () => {
  it("moves the active index forward and wraps around with moveNext", async () => {
    listInstrumentsMock.mockResolvedValue({
      items: [instrument("1", "PTT"), instrument("2", "AOT"), instrument("3", "CPALL")],
    });
    const search = useInstrumentSearch();
    search.setQuery("T");
    await vi.advanceTimersByTimeAsync(300);

    expect(search.activeIndex.value).toBe(0);
    search.moveNext();
    expect(search.activeIndex.value).toBe(1);
    search.moveNext();
    expect(search.activeIndex.value).toBe(2);
    search.moveNext();
    expect(search.activeIndex.value).toBe(0); // wraps
  });

  it("moves the active index backward and wraps around with movePrev", async () => {
    listInstrumentsMock.mockResolvedValue({
      items: [instrument("1", "PTT"), instrument("2", "AOT")],
    });
    const search = useInstrumentSearch();
    search.setQuery("T");
    await vi.advanceTimersByTimeAsync(300);

    expect(search.activeIndex.value).toBe(0);
    search.movePrev();
    expect(search.activeIndex.value).toBe(1); // wraps backward
    search.movePrev();
    expect(search.activeIndex.value).toBe(0);
  });

  it("reset() clears query, items, and active index (called by the combobox on selection/clear)", async () => {
    listInstrumentsMock.mockResolvedValue({ items: [instrument("1", "PTT")] });
    const search = useInstrumentSearch();
    search.setQuery("PT");
    await vi.advanceTimersByTimeAsync(300);
    expect(search.items.value).toHaveLength(1);

    search.reset();
    expect(search.query.value).toBe("");
    expect(search.items.value).toHaveLength(0);
    expect(search.activeIndex.value).toBe(-1);
    expect(search.isOpen.value).toBe(false);
  });

  it("close() (the Escape handler) closes the listbox without discarding the result set", async () => {
    listInstrumentsMock.mockResolvedValue({ items: [instrument("1", "PTT")] });
    const search = useInstrumentSearch();
    search.setQuery("PT");
    await vi.advanceTimersByTimeAsync(300);
    expect(search.isOpen.value).toBe(true);

    search.close();
    expect(search.isOpen.value).toBe(false);
    expect(search.activeIndex.value).toBe(-1);
    // The query and results are preserved — Escape only dismisses the popup.
    expect(search.query.value).toBe("PT");
    expect(search.items.value).toHaveLength(1);
  });
});
