import { describe, expect, it } from "vitest";

import {
  buildLockedWatchlistAlertQuery,
  buildLockedWatchlistItemQuery,
} from "../app/features/portfolio-workspace/lib/watchlistScope";

describe("buildLockedWatchlistItemQuery", () => {
  it("returns null when the portfolio id has not resolved yet", () => {
    expect(buildLockedWatchlistItemQuery(null)).toBeNull();
    expect(buildLockedWatchlistItemQuery(undefined)).toBeNull();
    expect(buildLockedWatchlistItemQuery("")).toBeNull();
  });

  it("locks scope_type to PORTFOLIO and carries the exact portfolio id, never a user choice", () => {
    expect(buildLockedWatchlistItemQuery("11111111-1111-1111-1111-111111111111")).toEqual({
      scope_type: "PORTFOLIO",
      portfolio_id: "11111111-1111-1111-1111-111111111111",
      include_thresholds: true,
    });
  });
});

describe("buildLockedWatchlistAlertQuery", () => {
  it("returns null without a portfolio id", () => {
    expect(buildLockedWatchlistAlertQuery(null)).toBeNull();
  });

  it("locks scope_type to PORTFOLIO with the resolved portfolio id", () => {
    expect(buildLockedWatchlistAlertQuery("22222222-2222-2222-2222-222222222222")).toEqual({
      scope_type: "PORTFOLIO",
      portfolio_id: "22222222-2222-2222-2222-222222222222",
    });
  });
});
