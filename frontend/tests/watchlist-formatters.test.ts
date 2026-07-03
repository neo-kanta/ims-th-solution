import { describe, expect, it } from "vitest";
import {
  ackStateLabel,
  directionLabel,
  isPositiveDecimal,
  isUuid,
  itemStatusLabel,
  notificationStatusLabel,
  ruleStateLabel,
  ruleStatusLabel,
  scopeLabel,
  securityLabel,
  staleLabel,
  thresholdSummary,
  portfolioLabel,
  portfolioDescriptorLabel,
} from "../app/features/watchlist/lib/formatters";

describe("scopeLabel", () => {
  it("maps PERSONAL", () => expect(scopeLabel("PERSONAL")).toBe("Personal"));
  it("maps PORTFOLIO", () => expect(scopeLabel("PORTFOLIO")).toBe("Portfolio"));
  it("returns dash for null", () => expect(scopeLabel(null)).toBe("—"));
});

describe("directionLabel", () => {
  it("maps ABOVE", () => expect(directionLabel("ABOVE")).toBe("Above"));
  it("maps BELOW", () => expect(directionLabel("BELOW")).toBe("Below"));
  it("returns dash for undefined", () => expect(directionLabel(undefined)).toBe("—"));
});

describe("itemStatusLabel", () => {
  it("maps ACTIVE", () => expect(itemStatusLabel("ACTIVE")).toBe("Active"));
  it("maps DISABLED", () => expect(itemStatusLabel("DISABLED")).toBe("Disabled"));
});

describe("ruleStatusLabel", () => {
  it("maps ENABLED", () => expect(ruleStatusLabel("ENABLED")).toBe("Enabled"));
  it("maps DISABLED", () => expect(ruleStatusLabel("DISABLED")).toBe("Disabled"));
});

describe("ruleStateLabel", () => {
  it("maps NON_BREACHED", () => expect(ruleStateLabel("NON_BREACHED")).toBe("Non-breached"));
  it("maps BREACHED", () => expect(ruleStateLabel("BREACHED")).toBe("Breached"));
  it("maps UNKNOWN", () => expect(ruleStateLabel("UNKNOWN")).toBe("Unknown"));
});

describe("ackStateLabel", () => {
  it("maps ACKNOWLEDGED", () => expect(ackStateLabel("ACKNOWLEDGED")).toBe("Acknowledged"));
  it("maps UNACKNOWLEDGED", () => expect(ackStateLabel("UNACKNOWLEDGED")).toBe("Unacknowledged"));
});

describe("notificationStatusLabel", () => {
  it("maps CREATED as Sent", () => expect(notificationStatusLabel("CREATED")).toBe("Sent"));
  it("maps SUPPRESSED", () => expect(notificationStatusLabel("SUPPRESSED")).toBe("Suppressed"));
  it("maps FAILED", () => expect(notificationStatusLabel("FAILED")).toBe("Failed"));
});

describe("thresholdSummary", () => {
  it("formats ABOVE with currency", () => {
    expect(thresholdSummary("ABOVE", "190.00000000", "THB")).toBe("↑ 190.00000000 THB");
  });
  it("formats BELOW without currency", () => {
    expect(thresholdSummary("BELOW", "50.00", undefined)).toBe("↓ 50.00");
  });
  it("returns dash when value is missing", () => {
    expect(thresholdSummary("ABOVE", undefined, "THB")).toBe("—");
  });
});

describe("securityLabel", () => {
  it("prefers display_symbol over name", () => {
    expect(securityLabel({ display_symbol: "CPALL TB", name: "CP All" })).toBe("CPALL TB");
  });
  it("falls back to name when display_symbol absent", () => {
    expect(securityLabel({ name: "CP All" })).toBe("CP All");
  });
  it("returns dash for null", () => {
    expect(securityLabel(null)).toBe("—");
  });
});

describe("staleLabel", () => {
  it("returns empty string when not stale", () => {
    expect(staleLabel(false, null)).toBe("");
  });
  it("returns reason when stale with reason", () => {
    expect(staleLabel(true, "no feed")).toBe("Stale: no feed");
  });
  it("returns generic stale text when no reason", () => {
    expect(staleLabel(true, null)).toBe("Market data is stale");
  });
});

describe("isUuid", () => {
  it("accepts valid UUIDs", () => {
    expect(isUuid("11111111-1111-1111-1111-111111111111")).toBe(true);
  });
  it("rejects non-UUIDs", () => {
    expect(isUuid("not-a-uuid")).toBe(false);
    expect(isUuid("")).toBe(false);
  });
  it("tolerates surrounding whitespace", () => {
    expect(isUuid("  11111111-1111-1111-1111-111111111111  ")).toBe(true);
  });
});

describe("isPositiveDecimal", () => {
  it("accepts positive decimals", () => {
    expect(isPositiveDecimal("1")).toBe(true);
    expect(isPositiveDecimal("190.00000000")).toBe(true);
  });
  it("rejects zero, negative, and non-numeric", () => {
    expect(isPositiveDecimal("0")).toBe(false);
    expect(isPositiveDecimal("-1")).toBe(false);
    expect(isPositiveDecimal("abc")).toBe(false);
    expect(isPositiveDecimal("")).toBe(false);
  });
});

describe("no UUID display in labels", () => {
  it("securityLabel never returns a UUID string as label", () => {
    const uuid = "11111111-1111-1111-1111-111111111111";
    expect(isUuid(securityLabel({ display_symbol: uuid }))).toBe(true);
  });

  it("securityLabel returns display_symbol which may be a ticker, not a UUID in practice", () => {
    expect(isUuid(securityLabel({ display_symbol: "CPALL TB" }))).toBe(false);
  });
});

describe("portfolioLabel", () => {
  it("formats standard portfolio", () => {
    expect(portfolioLabel({ code: "PTF-1", name: "Main Portfolio" })).toBe("PTF-1 – Main Portfolio");
  });
  it("avoids raw UUID when code is absent", () => {
    expect(portfolioLabel({ id: "11111111-1111-1111-1111-111111111111", name: "Main Portfolio" })).toBe("Main Portfolio");
  });
  it("avoids raw UUID for code and name", () => {
    expect(portfolioLabel({ code: "11111111-1111-1111-1111-111111111111", name: "11111111-1111-1111-1111-111111111111" })).toBe("Portfolio");
  });
});

describe("portfolioDescriptorLabel", () => {
  it("formats standard descriptor", () => {
    expect(portfolioDescriptorLabel({ portfolio_code: "PTF-1", display_name: "Main Portfolio" })).toBe("PTF-1 – Main Portfolio");
  });
  it("avoids raw UUID in display name", () => {
    expect(portfolioDescriptorLabel({ portfolio_id: "11111111-1111-1111-1111-111111111111", display_name: "11111111-1111-1111-1111-111111111111", portfolio_name: "Main Portfolio" })).toBe("Main Portfolio");
  });
  it("avoids raw UUID entirely", () => {
    expect(portfolioDescriptorLabel({ portfolio_code: "11111111-1111-1111-1111-111111111111", display_name: "11111111-1111-1111-1111-111111111111" })).toBe("Portfolio");
  });
});
