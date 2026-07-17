/**
 * Tests for the create/submit orchestration composable: save-draft,
 * create-then-submit, partial-failure retry (create succeeds, submit
 * fails), and duplicate-click prevention.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const createMock = vi.fn();
const submitMock = vi.fn();

vi.mock("../app/features/portfolio-decision/services/portfolioDecisionApi", () => ({
  portfolioDecisionApi: {
    create: (...args: unknown[]) => createMock(...args),
    submit: (...args: unknown[]) => submitMock(...args),
  },
}));

import { useDecisionLifecycle } from "../app/features/portfolio-decision/composables/useDecisionLifecycle";
import type { ApiCreateDecisionV2Request } from "../app/features/portfolio-decision/services/portfolioDecisionApi";

const body: ApiCreateDecisionV2Request = {
  instrument_code: "PTT",
  business_date: "2026-07-14",
  side: "BUY",
  currency: "THB",
  quantity: "1000",
};

beforeEach(() => {
  createMock.mockReset();
  submitMock.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("saveDraft", () => {
  it("creates the decision and stays in DRAFT — never calls submit", async () => {
    createMock.mockResolvedValueOnce({ id: "dec-1", status: "DRAFT" });
    const lifecycle = useDecisionLifecycle();

    const result = await lifecycle.saveDraft("TH-EQ-01", body);

    expect(createMock).toHaveBeenCalledWith("TH-EQ-01", body);
    expect(submitMock).not.toHaveBeenCalled();
    expect(result?.status).toBe("DRAFT");
    expect(lifecycle.phase.value).toBe("draft-saved");
    expect(lifecycle.decisionId.value).toBe("dec-1");
  });

  it("surfaces a create failure without setting a decisionId", async () => {
    createMock.mockRejectedValueOnce({ status: 400, data: { error: "invalid request" } });
    const lifecycle = useDecisionLifecycle();

    const result = await lifecycle.saveDraft("TH-EQ-01", body);

    expect(result).toBeNull();
    expect(lifecycle.phase.value).toBe("create-failed");
    expect(lifecycle.error.value).toMatch(/invalid request/i);
    expect(lifecycle.decisionId.value).toBeNull();
  });
});

describe("createThenSubmit", () => {
  it("creates then submits, reaching the submitted phase", async () => {
    createMock.mockResolvedValueOnce({ id: "dec-2", status: "DRAFT" });
    submitMock.mockResolvedValueOnce({ id: "dec-2", status: "PENDING_APPROVAL" });
    const lifecycle = useDecisionLifecycle();

    const result = await lifecycle.createThenSubmit("TH-EQ-01", body);

    expect(createMock).toHaveBeenCalledTimes(1);
    expect(submitMock).toHaveBeenCalledWith("TH-EQ-01", "dec-2");
    expect(result?.status).toBe("PENDING_APPROVAL");
    expect(lifecycle.phase.value).toBe("submitted");
  });

  it("never calls submit when create fails, and never creates a second draft", async () => {
    createMock.mockRejectedValueOnce({ status: 422, data: { error: "compliance blocked" } });
    const lifecycle = useDecisionLifecycle();

    const result = await lifecycle.createThenSubmit("TH-EQ-01", body);

    expect(result).toBeNull();
    expect(submitMock).not.toHaveBeenCalled();
    expect(lifecycle.phase.value).toBe("create-failed");
    expect(lifecycle.decisionId.value).toBeNull();
  });

  it("partial failure: create succeeds but submit fails — keeps the decisionId and tells the caller the draft was saved", async () => {
    createMock.mockResolvedValueOnce({ id: "dec-3", status: "DRAFT" });
    submitMock.mockRejectedValueOnce({ status: 409, data: { error: "workflow day not open" } });
    const lifecycle = useDecisionLifecycle();

    const result = await lifecycle.createThenSubmit("TH-EQ-01", body);

    expect(result).toBeNull();
    expect(lifecycle.phase.value).toBe("submit-failed");
    expect(lifecycle.decisionId.value).toBe("dec-3");
    expect(lifecycle.error.value).toMatch(/draft was saved/i);
    expect(lifecycle.error.value).toMatch(/workflow day not open/i);
  });

  it("retrySubmit after a partial failure calls submit only — create is never called again", async () => {
    createMock.mockResolvedValueOnce({ id: "dec-4", status: "DRAFT" });
    submitMock
      .mockRejectedValueOnce({ status: 409, data: { error: "workflow day not open" } })
      .mockResolvedValueOnce({ id: "dec-4", status: "PENDING_APPROVAL" });
    const lifecycle = useDecisionLifecycle();

    await lifecycle.createThenSubmit("TH-EQ-01", body);
    expect(lifecycle.phase.value).toBe("submit-failed");

    const retried = await lifecycle.retrySubmit("TH-EQ-01");

    expect(createMock).toHaveBeenCalledTimes(1); // still only once
    expect(submitMock).toHaveBeenCalledTimes(2);
    expect(submitMock).toHaveBeenLastCalledWith("TH-EQ-01", "dec-4");
    expect(retried?.status).toBe("PENDING_APPROVAL");
    expect(lifecycle.phase.value).toBe("submitted");
  });

  it("retrySubmit is a no-op without a prior decisionId", async () => {
    const lifecycle = useDecisionLifecycle();
    const result = await lifecycle.retrySubmit("TH-EQ-01");

    expect(result).toBeNull();
    expect(submitMock).not.toHaveBeenCalled();
  });
});

describe("duplicate-submission prevention", () => {
  it("ignores a second saveDraft click while the first is still in flight", async () => {
    let resolveCreate!: (value: { id: string; status: string }) => void;
    createMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveCreate = resolve;
      }),
    );
    const lifecycle = useDecisionLifecycle();

    const firstCall = lifecycle.saveDraft("TH-EQ-01", body);
    expect(lifecycle.busy.value).toBe(true);

    const secondCall = lifecycle.saveDraft("TH-EQ-01", body);
    const secondResult = await secondCall;
    expect(secondResult).toBeNull(); // dropped — a second click must not fire a second request

    resolveCreate({ id: "dec-5", status: "DRAFT" });
    const firstResult = await firstCall;

    expect(createMock).toHaveBeenCalledTimes(1);
    expect(firstResult?.id).toBe("dec-5");
  });

  it("ignores a second createThenSubmit click while the first is still in flight", async () => {
    let resolveCreate!: (value: { id: string; status: string }) => void;
    createMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveCreate = resolve;
      }),
    );
    submitMock.mockResolvedValueOnce({ id: "dec-6", status: "PENDING_APPROVAL" });
    const lifecycle = useDecisionLifecycle();

    const firstCall = lifecycle.createThenSubmit("TH-EQ-01", body);
    const secondResult = await lifecycle.createThenSubmit("TH-EQ-01", body);
    expect(secondResult).toBeNull();

    resolveCreate({ id: "dec-6", status: "DRAFT" });
    await firstCall;

    expect(createMock).toHaveBeenCalledTimes(1);
  });
});
