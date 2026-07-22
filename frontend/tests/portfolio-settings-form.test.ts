/**
 * Portfolio Settings update composable: success with the loaded
 * `expected_version`, duplicate-submission guard, missing-version
 * short-circuit, and status-classified error mapping (403/404/409/422/
 * unexpected) — mirrors `portfolio-create-form.test.ts`'s established
 * mocking pattern.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const updateMock = vi.fn();

vi.mock("../app/features/portfolio-workspace/services/portfolioApi", () => ({
  portfolioApi: {
    update: (...args: unknown[]) => updateMock(...args),
  },
}));

import {
  usePortfolioSettingsForm,
  valuesFromPortfolio,
} from "../app/features/portfolio-workspace/composables/usePortfolioSettingsForm";
import type { ApiPortfolioV2 } from "../app/features/portfolio-workspace/services/portfolioApi";

const portfolio: ApiPortfolioV2 = {
  id: "pf-1",
  code: "PF-001",
  name: "Growth Portfolio",
  description: "Existing description",
  benchmark: "SET50",
  risk_profile: "MODERATE",
  strategy_code: "EQUITY_GROWTH",
  version: 3,
};

beforeEach(() => {
  updateMock.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("valuesFromPortfolio", () => {
  it("hydrates editable fields from the loaded portfolio", () => {
    expect(valuesFromPortfolio(portfolio)).toEqual({
      name: "Growth Portfolio",
      description: "Existing description",
      benchmark: "SET50",
      risk_profile: "MODERATE",
      strategy_code: "EQUITY_GROWTH",
    });
  });

  it("returns empty values for a null portfolio", () => {
    expect(valuesFromPortfolio(null)).toEqual({
      name: "",
      description: "",
      benchmark: "",
      risk_profile: "",
      strategy_code: "",
    });
  });
});

describe("usePortfolioSettingsForm — success", () => {
  it("sends the loaded expected_version and updated fields", async () => {
    updateMock.mockResolvedValueOnce({ ...portfolio, name: "Renamed Portfolio", version: 4 });
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);
    form.setValues({ name: "Renamed Portfolio" });

    const result = await form.submit("PF-001", 3, "fallback");

    expect(updateMock).toHaveBeenCalledWith("PF-001", {
      expected_version: 3,
      name: "Renamed Portfolio",
      description: "Existing description",
      benchmark: "SET50",
      risk_profile: "MODERATE",
      strategy_code: "EQUITY_GROWTH",
    });
    expect(result?.version).toBe(4);
    expect(form.updated.value?.version).toBe(4);
    expect(form.busy.value).toBe(false);
  });
});

describe("usePortfolioSettingsForm — missing version", () => {
  it("never calls the API when expected_version is undefined", async () => {
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);

    const result = await form.submit("PF-001", undefined, "fallback message");

    expect(updateMock).not.toHaveBeenCalled();
    expect(result).toBeNull();
    expect(form.submitError.value?.message).toBe("fallback message");
  });
});

describe("usePortfolioSettingsForm — duplicate-submission guard", () => {
  it("a second submit() call while the first is in flight is a no-op", async () => {
    let resolveUpdate: (value: unknown) => void = () => undefined;
    updateMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveUpdate = resolve;
      }),
    );
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);

    const firstSubmit = form.submit("PF-001", 3, "fallback");
    expect(form.busy.value).toBe(true);

    const secondResult = await form.submit("PF-001", 3, "fallback");
    expect(secondResult).toBeNull();
    expect(updateMock).toHaveBeenCalledTimes(1);

    resolveUpdate({ ...portfolio, version: 4 });
    await firstSubmit;
  });
});

describe("usePortfolioSettingsForm — error classification", () => {
  it("preserves entered values and classifies a 409 as version_conflict", async () => {
    updateMock.mockRejectedValueOnce({ status: 409, data: { error: "version mismatch" } });
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);
    form.setValues({ name: "Attempted rename" });

    const result = await form.submit("PF-001", 3, "fallback");

    expect(result).toBeNull();
    expect(form.submitError.value?.kind).toBe("version_conflict");
    expect(form.submitError.value?.status).toBe(409);
    expect(form.values.value.name).toBe("Attempted rename");
  });

  it("classifies a 403 as forbidden", async () => {
    updateMock.mockRejectedValueOnce({ status: 403, data: { error: "no access" } });
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);

    await form.submit("PF-001", 3, "fallback");

    expect(form.submitError.value?.kind).toBe("forbidden");
  });

  it("classifies a 404 as not_found", async () => {
    updateMock.mockRejectedValueOnce({ status: 404, data: { error: "not found" } });
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);

    await form.submit("PF-001", 3, "fallback");

    expect(form.submitError.value?.kind).toBe("not_found");
  });

  it("classifies a 422 as business_rule", async () => {
    updateMock.mockRejectedValueOnce({ status: 422, data: { error: "invalid risk profile" } });
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);

    await form.submit("PF-001", 3, "fallback");

    expect(form.submitError.value?.kind).toBe("business_rule");
  });

  it("classifies a 400 as validation", async () => {
    updateMock.mockRejectedValueOnce({ status: 400, data: { error: "malformed" } });
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);

    await form.submit("PF-001", 3, "fallback");

    expect(form.submitError.value?.kind).toBe("validation");
  });

  it("classifies a 500 as unexpected", async () => {
    updateMock.mockRejectedValueOnce({ status: 500, data: { error: "internal error" } });
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);

    await form.submit("PF-001", 3, "fallback");

    expect(form.submitError.value?.kind).toBe("unexpected");
  });
});

describe("usePortfolioSettingsForm — hydrate", () => {
  it("clears submitError/updated and resets values when re-hydrated", async () => {
    updateMock.mockRejectedValueOnce({ status: 409 });
    const form = usePortfolioSettingsForm();
    form.hydrate(portfolio);
    await form.submit("PF-001", 3, "fallback");
    expect(form.submitError.value).not.toBeNull();

    form.hydrate({ ...portfolio, name: "Reloaded Name", version: 4 });

    expect(form.submitError.value).toBeNull();
    expect(form.updated.value).toBeNull();
    expect(form.values.value.name).toBe("Reloaded Name");
  });
});
