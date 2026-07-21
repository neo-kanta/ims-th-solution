/**
 * Composable-level tests for the Breaches page's list/override/check-group
 * composables. `complianceApi` is mocked so these run as plain unit tests —
 * consistent with the rest of this codebase's composable test style — and so
 * assertions cover the exact request/response contract rather than transport
 * plumbing.
 */
import { describe, expect, it, vi } from "vitest";

const listBreachesMock = vi.fn();
const overrideBreachMock = vi.fn();
const getCheckGroupMock = vi.fn();

vi.mock("../app/features/compliance/services/complianceApi", () => ({
  complianceApi: {
    listBreaches: (...args: unknown[]) => listBreachesMock(...args),
    overrideBreach: (...args: unknown[]) => overrideBreachMock(...args),
    getCheckGroup: (...args: unknown[]) => getCheckGroupMock(...args),
  },
}));

import {
  useComplianceBreachesList,
  useComplianceBreachOverride,
} from "../app/features/compliance/composables/useComplianceBreaches";
import { useComplianceCheckGroup } from "../app/features/compliance/composables/useComplianceCheckGroup";

function flushMicrotasks() {
  return new Promise((resolve) => setTimeout(resolve, 0));
}

/**
 * Mirrors the shape `OpenApiRequestError` (app/api/openapi.ts) produces —
 * built without importing that module, since it has a top-level `#imports`
 * dependency that only resolves inside a running Nuxt app.
 */
function fakeOpenApiRequestError(status: number, message: string): Error {
  const err = new Error(message);
  err.name = "OpenApiRequestError";
  Object.assign(err, { status });
  return err;
}

describe("useComplianceBreachesList", () => {
  it("passes the current offset/limit plus caller filters to listBreaches", async () => {
    listBreachesMock.mockResolvedValueOnce({ breaches: [], total: 0, offset: 0, limit: 50 });
    const list = useComplianceBreachesList();
    list.offset.value = 20;
    list.limit.value = 25;

    await list.fetchList({ status: "OPEN" });

    expect(listBreachesMock).toHaveBeenCalledWith({ offset: 20, limit: 25, status: "OPEN" });
  });

  it("keeps items from a superseded slow request from overwriting a newer result (stale-request protection)", async () => {
    listBreachesMock.mockReset();
    let resolveFirst!: (value: unknown) => void;
    listBreachesMock.mockImplementationOnce(
      () => new Promise((resolve) => (resolveFirst = resolve)),
    );
    listBreachesMock.mockResolvedValueOnce({
      breaches: [{ id: "second" }],
      total: 1,
      offset: 0,
      limit: 50,
    });

    const list = useComplianceBreachesList();
    const firstCall = list.fetchList({ status: "OPEN" });
    const secondCall = list.fetchList({ status: "ALL" });

    await secondCall;
    expect(list.items.value).toEqual([{ id: "second" }]);

    // The slow first request resolves after the second — it must not clobber it.
    resolveFirst({ breaches: [{ id: "first" }], total: 1, offset: 0, limit: 50 });
    await firstCall;
    await flushMicrotasks();

    expect(list.items.value).toEqual([{ id: "second" }]);
  });

  it("clears items and surfaces a message on failure", async () => {
    listBreachesMock.mockReset();
    listBreachesMock.mockRejectedValueOnce(new Error("boom"));
    const list = useComplianceBreachesList();

    await list.fetchList();

    expect(list.items.value).toEqual([]);
    expect(list.error.value).toBe("boom");
    expect(list.loading.value).toBe(false);
  });
});

describe("useComplianceBreachOverride", () => {
  it("submits only { reason } — no breach id, actor, or extra fields in the body", async () => {
    overrideBreachMock.mockReset();
    overrideBreachMock.mockResolvedValueOnce({
      id: "override-1",
      breachID: "breach-1",
      reason: "Manager pre-approved.",
      overriddenBy: "user-1",
      createdAt: "2026-07-20T10:20:00Z",
    });

    const mutation = useComplianceBreachOverride();
    await mutation.override("breach-1", { reason: "Manager pre-approved." });

    expect(overrideBreachMock).toHaveBeenCalledWith("breach-1", { reason: "Manager pre-approved." });
    const [, payload] = overrideBreachMock.mock.calls[0] as [string, Record<string, unknown>];
    expect(Object.keys(payload)).toEqual(["reason"]);
  });

  it("rejects a second concurrent submission instead of firing a duplicate request", async () => {
    overrideBreachMock.mockReset();
    let resolveFirst!: (value: unknown) => void;
    overrideBreachMock.mockImplementationOnce(
      () => new Promise((resolve) => (resolveFirst = resolve)),
    );

    const mutation = useComplianceBreachOverride();
    const first = mutation.override("breach-1", { reason: "First attempt." });
    expect(mutation.submitting.value).toBe(true);

    await expect(mutation.override("breach-1", { reason: "Second attempt." })).rejects.toThrow(
      /already in progress/i,
    );
    expect(overrideBreachMock).toHaveBeenCalledTimes(1);

    resolveFirst({ id: "override-1", breachID: "breach-1", reason: "First attempt.", overriddenBy: "u", createdAt: "2026-07-20T10:20:00Z" });
    await first;
  });

  it("flags a 409 response as a conflict without leaking the raw status text as the primary message", async () => {
    overrideBreachMock.mockReset();
    overrideBreachMock.mockRejectedValueOnce(
      fakeOpenApiRequestError(409, "breach already overridden"),
    );

    const mutation = useComplianceBreachOverride();
    await expect(mutation.override("breach-1", { reason: "Too late." })).rejects.toThrow();

    expect(mutation.conflict.value).toBe(true);
    expect(mutation.submitting.value).toBe(false);
  });

  it("does not flag non-409 errors as a conflict", async () => {
    overrideBreachMock.mockReset();
    overrideBreachMock.mockRejectedValueOnce(fakeOpenApiRequestError(403, "forbidden"));

    const mutation = useComplianceBreachOverride();
    await expect(mutation.override("breach-1", { reason: "No permission." })).rejects.toThrow();

    expect(mutation.conflict.value).toBe(false);
    expect(mutation.error.value).toBeTruthy();
  });

  it("reset() clears error, conflict, and the last result", async () => {
    overrideBreachMock.mockReset();
    overrideBreachMock.mockRejectedValueOnce(fakeOpenApiRequestError(409, "conflict"));

    const mutation = useComplianceBreachOverride();
    await expect(mutation.override("breach-1", { reason: "x" })).rejects.toThrow();
    expect(mutation.conflict.value).toBe(true);

    mutation.reset();
    expect(mutation.conflict.value).toBe(false);
    expect(mutation.error.value).toBeNull();
    expect(mutation.lastResult.value).toBeNull();
  });

  it("succeeds and clears any prior error state on a fresh attempt", async () => {
    overrideBreachMock.mockReset();
    overrideBreachMock.mockResolvedValueOnce({
      id: "override-2",
      breachID: "breach-2",
      reason: "ok",
      overriddenBy: "user-1",
      createdAt: "2026-07-20T10:20:00Z",
    });

    const mutation = useComplianceBreachOverride();
    const result = await mutation.override("breach-2", { reason: "ok" });

    expect(result.id).toBe("override-2");
    expect(mutation.error.value).toBeNull();
    expect(mutation.submitting.value).toBe(false);
  });
});

describe("useComplianceCheckGroup", () => {
  it("loads records and breaches for the requested group", async () => {
    getCheckGroupMock.mockReset();
    getCheckGroupMock.mockResolvedValueOnce({
      check_group_id: "group-1",
      records: [{ id: "record-1" }],
      breaches: [],
    });

    const checkGroup = useComplianceCheckGroup();
    await checkGroup.fetch("group-1");

    expect(getCheckGroupMock).toHaveBeenCalledWith("group-1");
    expect(checkGroup.result.value?.check_group_id).toBe("group-1");
    expect(checkGroup.loading.value).toBe(false);
  });

  it("keeps a slow, superseded group fetch from overwriting a newer one", async () => {
    getCheckGroupMock.mockReset();
    let resolveFirst!: (value: unknown) => void;
    getCheckGroupMock.mockImplementationOnce(
      () => new Promise((resolve) => (resolveFirst = resolve)),
    );
    getCheckGroupMock.mockResolvedValueOnce({
      check_group_id: "group-2",
      records: [],
      breaches: [],
    });

    const checkGroup = useComplianceCheckGroup();
    const first = checkGroup.fetch("group-1");
    const second = checkGroup.fetch("group-2");

    await second;
    expect(checkGroup.result.value?.check_group_id).toBe("group-2");

    resolveFirst({ check_group_id: "group-1", records: [], breaches: [] });
    await first;
    await flushMicrotasks();

    expect(checkGroup.result.value?.check_group_id).toBe("group-2");
  });

  it("surfaces a load failure and clears any previous result", async () => {
    getCheckGroupMock.mockReset();
    getCheckGroupMock.mockRejectedValueOnce(new Error("group not found"));

    const checkGroup = useComplianceCheckGroup();
    await checkGroup.fetch("missing-group");

    expect(checkGroup.error.value).toBe("group not found");
    expect(checkGroup.result.value).toBeNull();
  });

  it("clear() resets result, error, and loading", async () => {
    getCheckGroupMock.mockReset();
    getCheckGroupMock.mockResolvedValueOnce({ check_group_id: "group-1", records: [], breaches: [] });

    const checkGroup = useComplianceCheckGroup();
    await checkGroup.fetch("group-1");
    checkGroup.clear();

    expect(checkGroup.result.value).toBeNull();
    expect(checkGroup.error.value).toBeNull();
    expect(checkGroup.loading.value).toBe(false);
  });
});
