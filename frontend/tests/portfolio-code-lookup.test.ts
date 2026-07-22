/**
 * usePortfolioCodeLookup resolves a decision's internal portfolio_id (never
 * shown to the user) to its business code for OP-02 routing/labels. Each
 * lookup must go through the caller's own permission-scoped API call — an id
 * the caller cannot access must fail to resolve, not silently guess a code.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const getPortfolioMock = vi.fn();

vi.mock("../app/features/investment-ledger/services/investmentLedgerApi", () => ({
  investmentLedgerApi: {
    getPortfolio: (...args: unknown[]) => getPortfolioMock(...args),
  },
}));

import { usePortfolioCodeLookup } from "../app/features/investment-decision/composables/usePortfolioCodeLookup";

beforeEach(() => {
  getPortfolioMock.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("usePortfolioCodeLookup", () => {
  it("returns null and does not call the API for a missing id", async () => {
    const lookup = usePortfolioCodeLookup();
    const result = await lookup.resolve(undefined);
    expect(result).toBeNull();
    expect(getPortfolioMock).not.toHaveBeenCalled();
  });

  it("resolves a portfolio id to its business code", async () => {
    getPortfolioMock.mockResolvedValueOnce({ id: "uuid-1", code: "PF-001" });
    const lookup = usePortfolioCodeLookup();

    const result = await lookup.resolve("uuid-1");

    expect(result).toBe("PF-001");
    expect(lookup.codesById.value["uuid-1"]).toBe("PF-001");
  });

  it("fails closed (returns null) when the API call fails — never guesses a code", async () => {
    getPortfolioMock.mockRejectedValueOnce({ status: 403 });
    const lookup = usePortfolioCodeLookup();

    const result = await lookup.resolve("uuid-forbidden");

    expect(result).toBeNull();
    expect(lookup.codesById.value["uuid-forbidden"]).toBeUndefined();
  });

  it("resolveMany resolves every unique id and skips blanks", async () => {
    getPortfolioMock.mockImplementation(async (id: string) => ({ id, code: `CODE-${id}` }));
    const lookup = usePortfolioCodeLookup();

    await lookup.resolveMany(["uuid-a", "uuid-b", undefined, null, "uuid-a"]);

    expect(getPortfolioMock).toHaveBeenCalledTimes(2);
    expect(lookup.codesById.value["uuid-a"]).toBe("CODE-uuid-a");
    expect(lookup.codesById.value["uuid-b"]).toBe("CODE-uuid-b");
  });
});
