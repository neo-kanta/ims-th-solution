/**
 * OP-03 print-summary sections. Compliance and approval history are both
 * best-effort: a decision without a compliance_check_group_id / an
 * approval_request_id must render as "not recorded", never as an error or a
 * fabricated result, and both sections must reflect real data when the ids
 * are present.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const getCheckGroupMock = vi.fn();
const timelineMock = vi.fn();
const listExecutionsMock = vi.fn();
const listConfirmationsMock = vi.fn();

vi.mock("../app/features/compliance/services/complianceApi", () => ({
  complianceApi: {
    getCheckGroup: (...args: unknown[]) => getCheckGroupMock(...args),
  },
}));

vi.mock("../app/features/approval/services/approvalApi", () => ({
  approvalApi: {
    timeline: (...args: unknown[]) => timelineMock(...args),
  },
}));

vi.mock("../app/features/portfolio-workspace/services/portfolioApi", () => ({
  portfolioApi: {
    listExecutions: (...args: unknown[]) => listExecutionsMock(...args),
    listConfirmations: (...args: unknown[]) => listConfirmationsMock(...args),
  },
}));

import { usePrintSummarySections } from "../app/features/portfolio-decision/composables/usePrintSummarySections";
import type { ApiDecisionV2 } from "../app/features/portfolio-decision/services/portfolioDecisionApi";

function decision(overrides: Partial<ApiDecisionV2> = {}): ApiDecisionV2 {
  return {
    id: "dec-1",
    decision_number: "DEC-1",
    status: "APPROVED",
    ...overrides,
  };
}

beforeEach(() => {
  getCheckGroupMock.mockReset();
  timelineMock.mockReset();
  listExecutionsMock.mockReset();
  listConfirmationsMock.mockReset();
  listExecutionsMock.mockResolvedValue({ items: [] });
  listConfirmationsMock.mockResolvedValue({ items: [] });
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("usePrintSummarySections", () => {
  it("loads nothing and clears state for a null decision", async () => {
    const summary = usePrintSummarySections();
    await summary.load(null);

    expect(getCheckGroupMock).not.toHaveBeenCalled();
    expect(timelineMock).not.toHaveBeenCalled();
    expect(summary.compliance.result.value).toBeNull();
    expect(summary.approvalHistory.value).toEqual([]);
  });

  it("does not call the compliance or approval APIs when the decision carries neither id", async () => {
    const summary = usePrintSummarySections();
    await summary.load(decision());

    expect(getCheckGroupMock).not.toHaveBeenCalled();
    expect(timelineMock).not.toHaveBeenCalled();
  });

  it("fetches the compliance check group when compliance_check_group_id is present", async () => {
    getCheckGroupMock.mockResolvedValueOnce({
      check_group_id: "grp-1",
      records: [{ id: "rec-1", ruleTypeID: "concentration.single_issuer", finalVerdict: "PASS", message: "ok" }],
      breaches: [],
    });
    const summary = usePrintSummarySections();

    await summary.load(decision({ compliance_check_group_id: "grp-1" }));

    expect(getCheckGroupMock).toHaveBeenCalledWith("grp-1");
    expect(summary.compliance.result.value?.records).toHaveLength(1);
  });

  it("fetches the approval timeline when approval_request_id is present", async () => {
    timelineMock.mockResolvedValueOnce([
      { id: "evt-1", event_type: "APPROVED", actor_name: "Green", created_at: "2026-07-15T09:00:00Z" },
    ]);
    const summary = usePrintSummarySections();

    await summary.load(decision({ approval_request_id: "req-1" }));

    expect(timelineMock).toHaveBeenCalledWith("req-1");
    expect(summary.approvalHistory.value).toHaveLength(1);
    expect(summary.approvalHistory.value[0]?.actor_name).toBe("Green");
  });

  it("surfaces an approval-timeline fetch failure as an error, not a silent empty list", async () => {
    timelineMock.mockRejectedValueOnce(new Error("network down"));
    const summary = usePrintSummarySections();

    await summary.load(decision({ approval_request_id: "req-2" }));

    expect(summary.approvalHistoryError.value).toMatch(/network down/);
    expect(summary.approvalHistory.value).toEqual([]);
  });

  it("does not call the execution/confirmation endpoints when no portfolioCode is given", async () => {
    const summary = usePrintSummarySections();
    await summary.load(decision());

    expect(listExecutionsMock).not.toHaveBeenCalled();
    expect(summary.executions.value).toEqual([]);
  });

  it("loads and matches executions for this decision by decision_id, ignoring other decisions' rows", async () => {
    listExecutionsMock.mockResolvedValueOnce({
      items: [
        { id: "exec-1", decision_id: "dec-1", status: "EXECUTED", executed_quantity: "1000" },
        { id: "exec-2", decision_id: "dec-other", status: "PENDING" },
      ],
    });
    const summary = usePrintSummarySections();

    await summary.load(decision({ id: "dec-1" }), "TH-EQ-01");

    expect(listExecutionsMock).toHaveBeenCalledWith("TH-EQ-01", { limit: 200 });
    expect(summary.executions.value).toHaveLength(1);
    expect(summary.executions.value[0]?.id).toBe("exec-1");
  });

  it("matches a trade confirmation to its execution via execution_id", async () => {
    listExecutionsMock.mockResolvedValueOnce({
      items: [{ id: "exec-1", decision_id: "dec-1", status: "EXECUTED" }],
    });
    listConfirmationsMock.mockResolvedValueOnce({
      items: [
        { id: "conf-1", execution_id: "exec-1", status: "MATCHED" },
        { id: "conf-2", execution_id: "exec-other", status: "MISMATCHED" },
      ],
    });
    const summary = usePrintSummarySections();

    await summary.load(decision({ id: "dec-1" }), "TH-EQ-01");

    expect(summary.confirmationsByExecutionId.value["exec-1"]?.status).toBe("MATCHED");
    expect(summary.confirmationsByExecutionId.value["exec-other"]).toBeUndefined();
  });

  it("renders no executions as an empty (not error) state when the decision has none yet", async () => {
    listExecutionsMock.mockResolvedValueOnce({ items: [] });
    const summary = usePrintSummarySections();

    await summary.load(decision({ id: "dec-1" }), "TH-EQ-01");

    expect(summary.executions.value).toEqual([]);
    expect(summary.executionsError.value).toBeNull();
  });

  it("surfaces an execution-list fetch failure as an error, not a silent empty list", async () => {
    listExecutionsMock.mockRejectedValueOnce(new Error("execution service down"));
    const summary = usePrintSummarySections();

    await summary.load(decision({ id: "dec-1" }), "TH-EQ-01");

    expect(summary.executionsError.value).toMatch(/execution service down/);
    expect(summary.executions.value).toEqual([]);
  });

  it("a superseded load (second call before the first settles) does not overwrite the newer result", async () => {
    let resolveFirst: (value: unknown[]) => void = () => undefined;
    timelineMock.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveFirst = resolve;
      }),
    );
    const summary = usePrintSummarySections();

    const firstLoad = summary.load(decision({ id: "dec-a", approval_request_id: "req-a" }));

    timelineMock.mockResolvedValueOnce([{ id: "evt-b", event_type: "SUBMITTED" }]);
    await summary.load(decision({ id: "dec-b", approval_request_id: "req-b" }));

    resolveFirst([{ id: "evt-a", event_type: "APPROVED" }]);
    await firstLoad;

    expect(summary.approvalHistory.value).toEqual([{ id: "evt-b", event_type: "SUBMITTED" }]);
  });
});
