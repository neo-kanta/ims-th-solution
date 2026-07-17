import { describe, expect, it } from "vitest";

import {
  filterPortfolioPicks,
  toPortfolioPicks,
} from "../app/features/portfolio-decision/lib/portfolioPicker";

describe("toPortfolioPicks", () => {
  it("drops entries with no code", () => {
    const picks = toPortfolioPicks([{ code: null, name: "Orphan" }, { code: "OP-1024", name: "Growth" }]);
    expect(picks).toEqual([{ code: "OP-1024", name: "Growth", status: null }]);
  });

  it("falls back to the code as the name when name is blank", () => {
    const picks = toPortfolioPicks([{ code: "OP-1024", name: "   " }]);
    expect(picks[0]?.name).toBe("OP-1024");
  });

  it("carries the status through, defaulting to null", () => {
    const picks = toPortfolioPicks([{ code: "OP-1024", name: "Growth", status: "ACTIVE" }, { code: "OP-2", name: "X" }]);
    expect(picks[0]?.status).toBe("ACTIVE");
    expect(picks[1]?.status).toBeNull();
  });
});

describe("filterPortfolioPicks", () => {
  const picks = [
    { code: "OP-1024", name: "Growth Fund", status: "ACTIVE" },
    { code: "OP-2048", name: "Income Fund", status: "ACTIVE" },
  ];

  it("returns everything when the query is blank", () => {
    expect(filterPortfolioPicks(picks, "  ")).toEqual(picks);
  });

  it("matches by code, case-insensitively", () => {
    expect(filterPortfolioPicks(picks, "op-1024")).toEqual([picks[0]]);
  });

  it("matches by name, case-insensitively", () => {
    expect(filterPortfolioPicks(picks, "income")).toEqual([picks[1]]);
  });

  it("returns an empty list when nothing matches", () => {
    expect(filterPortfolioPicks(picks, "no-such-portfolio")).toEqual([]);
  });
});
