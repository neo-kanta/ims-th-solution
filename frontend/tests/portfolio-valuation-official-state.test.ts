import { describe, expect, it } from "vitest";

import {
  deriveValuationOfficialState,
  isOfficialValuationState,
} from "../app/features/portfolio-workspace/lib/valuationOfficialState";

describe("deriveValuationOfficialState", () => {
  it("is UNAVAILABLE for a MODEL portfolio even with a valuation-shaped payload", () => {
    expect(
      deriveValuationOfficialState({ is_indicative: false }, "MODEL"),
    ).toBe("UNAVAILABLE");
  });

  it("is UNAVAILABLE when there is no valuation at all", () => {
    expect(deriveValuationOfficialState(null, "LIVE")).toBe("UNAVAILABLE");
    expect(deriveValuationOfficialState(undefined, "LIVE")).toBe("UNAVAILABLE");
  });

  it("is OFFICIAL for a LIVE portfolio with is_indicative=false", () => {
    expect(
      deriveValuationOfficialState({ is_indicative: false }, "LIVE"),
    ).toBe("OFFICIAL");
  });

  it("is INDICATIVE for a LIVE portfolio whose snapshot itself is flagged indicative", () => {
    expect(
      deriveValuationOfficialState({ is_indicative: true }, "LIVE"),
    ).toBe("INDICATIVE");
  });

  it("is STALE_INDICATIVE when inputs are stale", () => {
    expect(
      deriveValuationOfficialState(
        { is_indicative: true, has_stale_inputs: true },
        "LIVE",
      ),
    ).toBe("STALE_INDICATIVE");
  });

  it("always treats SIMULATION as indicative even if is_indicative is false", () => {
    expect(
      deriveValuationOfficialState({ is_indicative: false }, "SIMULATION"),
    ).toBe("INDICATIVE");
  });

  it("SIMULATION with stale inputs is STALE_INDICATIVE", () => {
    expect(
      deriveValuationOfficialState(
        { is_indicative: false, has_stale_inputs: true },
        "SIMULATION",
      ),
    ).toBe("STALE_INDICATIVE");
  });
});

describe("isOfficialValuationState", () => {
  it("is true only for OFFICIAL", () => {
    expect(isOfficialValuationState("OFFICIAL")).toBe(true);
    expect(isOfficialValuationState("INDICATIVE")).toBe(false);
    expect(isOfficialValuationState("STALE_INDICATIVE")).toBe(false);
    expect(isOfficialValuationState("UNAVAILABLE")).toBe(false);
  });
});
