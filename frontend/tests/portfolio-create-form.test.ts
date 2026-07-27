/**
 * Create Portfolio orchestration composable: success, duplicate-submission
 * guard, field-validation short-circuit (never calls the API), and
 * status-classified error mapping (403/404/409/422/unexpected) — mirrors
 * `portfolio-decision-lifecycle.test.ts`'s established mocking pattern.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const createMock = vi.fn();

vi.mock("../app/features/portfolio-workspace/services/portfolioApi", () => ({
  portfolioApi: {
    create: (...args: unknown[]) => createMock(...args),
  },
}));

import { usePortfolioCreateForm } from "../app/features/portfolio-workspace/composables/usePortfolioCreateForm";
import { portfolioOverviewPath } from "../app/features/portfolio-workspace/lib/portfolioRoutes";

const validValues = {
  bindFund: true,
  fund_code: "TH-FUND-01",
  portfolio_type: "LIVE",
  code: "PF-001",
  name: "Growth Portfolio",
  description: "",
  base_currency: "THB",
  valuation_currency: "THB",
  strategy_code: "",
  benchmark: "",
  risk_profile: "",
  inception_date: "2026-01-01",
};

beforeEach(() => {
  createMock.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("usePortfolioCreateForm — success", () => {
  it("creates the portfolio and reaches the success phase", async () => {
    createMock.mockResolvedValueOnce({ id: "pf-1", code: "PF-001", name: "Growth Portfolio" });
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    const result = await form.submit("fallback");

    expect(createMock).toHaveBeenCalledWith({
      fund_code: "TH-FUND-01",
      portfolio_type: "LIVE",
      code: "PF-001",
      name: "Growth Portfolio",
      base_currency: "THB",
      valuation_currency: "THB",
      inception_date: "2026-01-01",
    });
    expect(result?.code).toBe("PF-001");
    expect(form.phase.value).toBe("success");
    expect(form.busy.value).toBe(false);
  });

  it("builds the correct encodeURIComponent-safe router.push target on success", async () => {
    createMock.mockResolvedValueOnce({ id: "pf-1", code: "PF 001/ODD" });
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    const result = await form.submit("fallback");

    expect(portfolioOverviewPath(result!.code!)).toBe(
      `/portfolios/${encodeURIComponent("PF 001/ODD")}/overview`,
    );
    expect(portfolioOverviewPath(result!.code!)).not.toContain(" ");
  });
});

describe("usePortfolioCreateForm — validation short-circuit", () => {
  it("never calls the API when client-side validation fails, and preserves entered values", async () => {
    const form = usePortfolioCreateForm();
    form.setValues({ ...validValues, code: "" });

    const result = await form.submit("fallback");

    expect(createMock).not.toHaveBeenCalled();
    expect(result).toBeNull();
    expect(form.fieldErrors.value.code).toBe("required");
    expect(form.values.value.name).toBe("Growth Portfolio");
  });
});

describe("usePortfolioCreateForm — duplicate-submission guard", () => {
  it("a second submit() call while the first is in flight is a no-op", async () => {
    let resolveCreate: (value: unknown) => void = () => undefined;
    createMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveCreate = resolve;
      }),
    );
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    const firstSubmit = form.submit("fallback");
    expect(form.busy.value).toBe(true);

    const secondResult = await form.submit("fallback");
    expect(secondResult).toBeNull();
    expect(createMock).toHaveBeenCalledTimes(1);

    resolveCreate({ id: "pf-1", code: "PF-001" });
    await firstSubmit;
  });
});

describe("usePortfolioCreateForm — error classification", () => {
  it("preserves entered values and classifies a 403 as forbidden", async () => {
    createMock.mockRejectedValueOnce({ status: 403, data: { error: "no fund access" } });
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    const result = await form.submit("fallback");

    expect(result).toBeNull();
    expect(form.phase.value).toBe("failed");
    expect(form.submitError.value?.kind).toBe("forbidden");
    expect(form.submitError.value?.status).toBe(403);
    expect(form.values.value.code).toBe("PF-001");
  });

  it("classifies a 404 as fund_not_found", async () => {
    createMock.mockRejectedValueOnce({ status: 404, data: { error: "fund not found" } });
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    await form.submit("fallback");

    expect(form.submitError.value?.kind).toBe("fund_not_found");
  });

  it("classifies a 409 as duplicate_code", async () => {
    createMock.mockRejectedValueOnce({ status: 409, data: { error: "code already exists" } });
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    await form.submit("fallback");

    expect(form.submitError.value?.kind).toBe("duplicate_code");
  });

  it("classifies a 422 as business_rule", async () => {
    createMock.mockRejectedValueOnce({ status: 422, data: { error: "inception date in the future" } });
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    await form.submit("fallback");

    expect(form.submitError.value?.kind).toBe("business_rule");
  });

  it("classifies a 400 as validation", async () => {
    createMock.mockRejectedValueOnce({ status: 400, data: { error: "malformed request" } });
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    await form.submit("fallback");

    expect(form.submitError.value?.kind).toBe("validation");
  });

  it("classifies a 500 (and any unrecognized status) as unexpected", async () => {
    createMock.mockRejectedValueOnce({ status: 500, data: { error: "internal error" } });
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    await form.submit("fallback");

    expect(form.submitError.value?.kind).toBe("unexpected");
  });

  it("falls back to the caller-provided message when the error carries none", async () => {
    createMock.mockRejectedValueOnce({});
    const form = usePortfolioCreateForm();
    form.setValues(validValues);

    await form.submit("Failed to create the portfolio.");

    expect(form.submitError.value?.message).toBe("Failed to create the portfolio.");
    expect(form.submitError.value?.kind).toBe("unexpected");
  });
});

describe("usePortfolioCreateForm — reset", () => {
  it("clears values and state back to idle", async () => {
    createMock.mockRejectedValueOnce({ status: 409 });
    const form = usePortfolioCreateForm();
    form.setValues(validValues);
    await form.submit("fallback");

    form.reset();

    expect(form.values.value.code).toBe("");
    expect(form.phase.value).toBe("idle");
    expect(form.submitError.value).toBeNull();
  });
});
