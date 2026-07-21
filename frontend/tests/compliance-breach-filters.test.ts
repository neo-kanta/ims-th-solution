import { describe, expect, it } from "vitest";

import {
  apiFiltersFromState,
  breachFilterStatesEqual,
  defaultBreachFilterState,
  filterStateFromQuery,
  hasNonDefaultFilters,
  queryFromFilterState,
  type BreachFilterState,
} from "../app/features/compliance/lib/breachFilters";

describe("defaultBreachFilterState", () => {
  it("defaults to the OPEN attention queue, not ALL", () => {
    expect(defaultBreachFilterState()).toEqual({
      status: "OPEN",
      ruleTypeId: "",
      portfolioId: "",
      dateFrom: "",
      dateTo: "",
    });
  });
});

describe("filterStateFromQuery", () => {
  it("falls back to OPEN when status is missing", () => {
    expect(filterStateFromQuery({})).toMatchObject({ status: "OPEN" });
  });

  it("falls back to OPEN when status is an unrecognized value", () => {
    expect(filterStateFromQuery({ status: "BOGUS" })).toMatchObject({ status: "OPEN" });
  });

  it("accepts each valid status, including ALL", () => {
    expect(filterStateFromQuery({ status: "OVERRIDDEN" }).status).toBe("OVERRIDDEN");
    expect(filterStateFromQuery({ status: "RESOLVED" }).status).toBe("RESOLVED");
    expect(filterStateFromQuery({ status: "ALL" }).status).toBe("ALL");
  });

  it("is case-insensitive on the status query value", () => {
    expect(filterStateFromQuery({ status: "open" }).status).toBe("OPEN");
  });

  it("reads rule/portfolio/date fields exactly by their wire names", () => {
    const state = filterStateFromQuery({
      status: "ALL",
      rule_type_id: "concentration.single_issuer",
      portfolio_id: "portfolio-1",
      date_from: "2026-07-01",
      date_to: "2026-07-31",
    });

    expect(state).toEqual({
      status: "ALL",
      ruleTypeId: "concentration.single_issuer",
      portfolioId: "portfolio-1",
      dateFrom: "2026-07-01",
      dateTo: "2026-07-31",
    });
  });

  it("takes the first value when vue-router repeats a query key as an array", () => {
    expect(filterStateFromQuery({ status: ["OVERRIDDEN", "RESOLVED"] }).status).toBe("OVERRIDDEN");
  });

  it("treats a null query value (bare query flag) as empty", () => {
    expect(filterStateFromQuery({ rule_type_id: null }).ruleTypeId).toBe("");
  });
});

describe("queryFromFilterState", () => {
  it("always includes status, even the OPEN default", () => {
    expect(queryFromFilterState(defaultBreachFilterState())).toEqual({ status: "OPEN" });
  });

  it("omits empty optional fields entirely rather than writing empty strings", () => {
    const query = queryFromFilterState({
      status: "ALL",
      ruleTypeId: "",
      portfolioId: "",
      dateFrom: "",
      dateTo: "",
    });
    expect(query).toEqual({ status: "ALL" });
  });

  it("includes every populated field under its exact wire key", () => {
    const query = queryFromFilterState({
      status: "OPEN",
      ruleTypeId: "cash.availability",
      portfolioId: "portfolio-9",
      dateFrom: "2026-01-01",
      dateTo: "2026-01-31",
    });
    expect(query).toEqual({
      status: "OPEN",
      rule_type_id: "cash.availability",
      portfolio_id: "portfolio-9",
      date_from: "2026-01-01",
      date_to: "2026-01-31",
    });
  });

  it("trims whitespace-only rule/portfolio ids down to omitted", () => {
    const query = queryFromFilterState({
      status: "OPEN",
      ruleTypeId: "   ",
      portfolioId: "  ",
      dateFrom: "",
      dateTo: "",
    });
    expect(query).toEqual({ status: "OPEN" });
  });
});

describe("apiFiltersFromState — exact GET /compliance/breaches serialization", () => {
  it("omits `status` entirely for ALL rather than sending status=ALL to the backend", () => {
    const filters = apiFiltersFromState({
      status: "ALL",
      ruleTypeId: "",
      portfolioId: "",
      dateFrom: "",
      dateTo: "",
    });
    expect(filters).toEqual({});
    expect(filters).not.toHaveProperty("status");
  });

  it("passes OPEN/OVERRIDDEN/RESOLVED through verbatim", () => {
    expect(apiFiltersFromState({ ...defaultBreachFilterState(), status: "OPEN" })).toEqual({
      status: "OPEN",
    });
    expect(apiFiltersFromState({ ...defaultBreachFilterState(), status: "OVERRIDDEN" })).toEqual({
      status: "OVERRIDDEN",
    });
  });

  it("maps every field to its exact backend query parameter name", () => {
    const state: BreachFilterState = {
      status: "OPEN",
      ruleTypeId: "ratio.sector_exposure",
      portfolioId: "22222222-2222-2222-2222-222222222222",
      dateFrom: "2026-07-01",
      dateTo: "2026-07-21",
    };

    expect(apiFiltersFromState(state)).toEqual({
      status: "OPEN",
      rule_type_id: "ratio.sector_exposure",
      portfolio_id: "22222222-2222-2222-2222-222222222222",
      date_from: "2026-07-01",
      date_to: "2026-07-21",
    });
  });

  it("never emits severity, message, contract, or sort keys — the API has no such parameters", () => {
    const filters = apiFiltersFromState({
      status: "OPEN",
      ruleTypeId: "cash.availability",
      portfolioId: "portfolio-1",
      dateFrom: "2026-07-01",
      dateTo: "2026-07-21",
    });
    const keys = Object.keys(filters);
    expect(keys).toEqual(
      expect.arrayContaining(["status", "rule_type_id", "portfolio_id", "date_from", "date_to"]),
    );
    expect(keys).toHaveLength(5);
  });
});

describe("hasNonDefaultFilters — drives the true-empty vs. filtered-empty distinction", () => {
  it("is false for the untouched default OPEN queue", () => {
    expect(hasNonDefaultFilters(defaultBreachFilterState())).toBe(false);
  });

  it("is false for OPEN with nothing else set (status alone is not a 'filter' from the user's view)", () => {
    expect(hasNonDefaultFilters({ ...defaultBreachFilterState(), status: "OPEN" })).toBe(false);
  });

  it("is true once the status departs from OPEN", () => {
    expect(hasNonDefaultFilters({ ...defaultBreachFilterState(), status: "ALL" })).toBe(true);
  });

  it("is true once a rule, portfolio, or date is set", () => {
    expect(hasNonDefaultFilters({ ...defaultBreachFilterState(), ruleTypeId: "cash.availability" })).toBe(
      true,
    );
    expect(hasNonDefaultFilters({ ...defaultBreachFilterState(), portfolioId: "p-1" })).toBe(true);
    expect(hasNonDefaultFilters({ ...defaultBreachFilterState(), dateFrom: "2026-07-01" })).toBe(true);
  });
});

describe("breachFilterStatesEqual", () => {
  it("is true for structurally identical states", () => {
    expect(breachFilterStatesEqual(defaultBreachFilterState(), defaultBreachFilterState())).toBe(true);
  });

  it("is false when exactly one field differs", () => {
    expect(
      breachFilterStatesEqual(defaultBreachFilterState(), {
        ...defaultBreachFilterState(),
        dateTo: "2026-07-21",
      }),
    ).toBe(false);
  });
});
