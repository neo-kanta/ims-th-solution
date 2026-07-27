/**
 * Tests for the LIVE cash-transaction ticket composable (Stage 2 of the
 * LIVE cash-transaction approval gate): payload mapping, validation, and
 * the simulate -> confirm -> post lifecycle, including the pending-vs-
 * posted outcome distinction (202 CashRequestResponse vs 201
 * TransactionResponse).
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const simulateMock = vi.fn();
const postMock = vi.fn();

// portfolioApi.ts depends on the Nuxt runtime client (~/api/openapi), which
// plain Vitest cannot resolve — mock the whole service module directly
// (same pattern as investment-ledger-order-ticket.test.ts's
// investmentLedgerApi mock) rather than trying to mock its transitive
// `~/api/openapi` import.
vi.mock("../app/features/portfolio-workspace/services/portfolioApi", () => ({
  portfolioApi: {
    simulateTransaction: (...args: unknown[]) => simulateMock(...args),
    postTransaction: (...args: unknown[]) => postMock(...args),
  },
}));

import {
  buildCashTransactionRequest,
  useCashTicket,
  validateCashDraft,
} from "../app/features/portfolio-workspace/composables/useCashTicket";
import { isCashRequestResponse } from "../app/features/portfolio-workspace/lib/cashRequestGuard";

beforeEach(() => {
  simulateMock.mockReset();
  postMock.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

function withFilledDraft(ticket: ReturnType<typeof useCashTicket>) {
  ticket.patchDraft({
    transaction_type: "CASH_IN",
    gross_amount: "1000",
    fees: "0",
    currency: "THB",
    business_date: "2026-07-24",
  });
}

describe("buildCashTransactionRequest", () => {
  it("maps the draft into the backend wire format", () => {
    const body = buildCashTransactionRequest({
      transaction_type: "CASH_IN",
      gross_amount: "5000",
      fees: "10",
      currency: "THB",
      business_date: "2026-07-24",
      reason: "capital call",
    });

    expect(body).toEqual({
      transaction_type: "CASH_IN",
      gross_amount: "5000",
      fees: "10",
      currency: "THB",
      business_date: "2026-07-24",
      reason: "capital call",
    });
  });

  it("drops empty optional fields rather than sending blank strings", () => {
    const body = buildCashTransactionRequest({
      transaction_type: "FEE",
      gross_amount: "100",
      fees: "",
      currency: "THB",
      business_date: "2026-07-24",
      reason: "",
    });
    expect(body.fees).toBeUndefined();
    expect(body.reason).toBeUndefined();
    expect(body.transaction_type).toBe("FEE");
  });
});

describe("validateCashDraft", () => {
  it("requires a positive amount, valid currency, and business date", () => {
    const errors = validateCashDraft({
      transaction_type: "CASH_IN",
      gross_amount: "",
      fees: "",
      currency: "",
      business_date: "",
      reason: "",
    });
    expect(errors.gross_amount).toBeDefined();
    expect(errors.currency).toBeDefined();
    expect(errors.business_date).toBeDefined();
  });

  it("rejects a non-positive amount", () => {
    const errors = validateCashDraft({
      transaction_type: "CASH_OUT",
      gross_amount: "-5",
      fees: "",
      currency: "THB",
      business_date: "2026-07-24",
      reason: "",
    });
    expect(errors.gross_amount).toMatch(/positive/i);
  });

  it("rejects a non-ISO currency code", () => {
    const errors = validateCashDraft({
      transaction_type: "DIVIDEND",
      gross_amount: "10",
      fees: "",
      currency: "thb",
      business_date: "2026-07-24",
      reason: "",
    });
    expect(errors.currency).toBeDefined();
  });

  it("accepts a fully populated draft", () => {
    expect(
      validateCashDraft({
        transaction_type: "CASH_IN",
        gross_amount: "1000",
        fees: "0",
        currency: "THB",
        business_date: "2026-07-24",
        reason: "",
      }),
    ).toEqual({});
  });
});

describe("useCashTicket lifecycle", () => {
  it("blocks post when verdict is BLOCK", async () => {
    const ticket = useCashTicket({ simulate: simulateMock, post: postMock });
    ticket.open("B14-CORE");
    withFilledDraft(ticket);

    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "BLOCK", breaches: [], rules_evaluated: 1 },
      cash: { current_balance: "1000", projected_balance: "2000", cash_impact: "1000", currency: "THB" },
    });

    await ticket.simulate();
    expect(ticket.isBlocked.value).toBe(true);
    expect(ticket.canPost.value).toBe(false);

    await ticket.post();
    expect(postMock).not.toHaveBeenCalled();
    expect(ticket.postError.value).toMatch(/BLOCK/);
  });

  it("moves to the 'pending' stage on a 202 CashRequestResponse (LIVE cash staged for approval)", async () => {
    const ticket = useCashTicket({ simulate: simulateMock, post: postMock });
    ticket.open("B14-CORE");
    withFilledDraft(ticket);

    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 1 },
      cash: { current_balance: "1000", projected_balance: "2000", cash_impact: "1000", currency: "THB" },
    });
    postMock.mockResolvedValueOnce({
      id: "cash-req-1",
      status: "PENDING",
      transaction_type: "CASH_IN",
      submitted_by: "user-1",
      approval_request_id: "approval-1",
    });

    await ticket.simulate();
    const result = await ticket.post();

    expect(result).toBeDefined();
    expect(isCashRequestResponse(result!)).toBe(true);
    expect(ticket.stage.value).toBe("pending");
    expect(ticket.lastCashRequest.value?.id).toBe("cash-req-1");
    expect(ticket.lastPostedTransaction.value).toBeNull();
  });

  it("moves to the 'posted' stage on a 201 TransactionResponse (SIMULATION cash posts immediately)", async () => {
    const ticket = useCashTicket({ simulate: simulateMock, post: postMock });
    ticket.open("TEST");
    withFilledDraft(ticket);

    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 1 },
      cash: { current_balance: "1000", projected_balance: "2000", cash_impact: "1000", currency: "THB" },
    });
    postMock.mockResolvedValueOnce({ id: "txn-1", status: "POSTED", transaction_type: "CASH_IN" });

    await ticket.simulate();
    const result = await ticket.post();

    expect(isCashRequestResponse(result!)).toBe(false);
    expect(ticket.stage.value).toBe("posted");
    expect(ticket.lastPostedTransaction.value?.id).toBe("txn-1");
    expect(ticket.lastCashRequest.value).toBeNull();
  });

  it("invalidates the simulation when the draft mutates afterwards", async () => {
    const ticket = useCashTicket({ simulate: simulateMock, post: postMock });
    ticket.open("B14-CORE");
    withFilledDraft(ticket);

    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 1 },
      cash: { current_balance: "1000", projected_balance: "2000", cash_impact: "1000", currency: "THB" },
    });
    await ticket.simulate();
    expect(ticket.isSimulationFresh.value).toBe(true);

    ticket.patchDraft({ gross_amount: "2000" });
    expect(ticket.isSimulationFresh.value).toBe(false);
    expect(ticket.canPost.value).toBe(false);
  });

  it("classifies a 403 from simulate as a permission error", async () => {
    const ticket = useCashTicket({ simulate: simulateMock, post: postMock });
    ticket.open("B14-CORE");
    withFilledDraft(ticket);

    simulateMock.mockRejectedValueOnce({ status: 403, data: { error: "forbidden" } });
    await ticket.simulate();
    expect(ticket.simulationError.value).toMatch(/INVESTMENT_LEDGER_SIMULATE/);
  });

  // ── G2: frontend must always send a stable Idempotency-Key ────────────────

  it("sends a non-empty idempotency key on post", async () => {
    const ticket = useCashTicket({ simulate: simulateMock, post: postMock });
    ticket.open("B14-CORE");
    withFilledDraft(ticket);

    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 1 },
      cash: { current_balance: "1000", projected_balance: "2000", cash_impact: "1000", currency: "THB" },
    });
    postMock.mockResolvedValueOnce({ id: "cr-1", status: "PENDING", submitted_by: "u1", transaction_type: "CASH_IN" });

    await ticket.simulate();
    await ticket.post();

    expect(postMock).toHaveBeenCalledTimes(1);
    const [, , key] = postMock.mock.calls[0]!;
    expect(typeof key).toBe("string");
    expect((key as string).length).toBeGreaterThan(0);
    expect(ticket.pendingIdempotencyKey.value).toBe(key);
  });

  it("reuses the same key across a retry of the same submission attempt", async () => {
    const ticket = useCashTicket({ simulate: simulateMock, post: postMock });
    ticket.open("B14-CORE");
    withFilledDraft(ticket);

    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 1 },
      cash: { current_balance: "1000", projected_balance: "2000", cash_impact: "1000", currency: "THB" },
    });
    await ticket.simulate();

    // First attempt fails (e.g. a transient network error) — the draft/
    // simulation are unchanged, so a retry must reuse the same key rather
    // than minting a new one (the backend needs the SAME key to recognise
    // the retry as the same request, not a new one).
    postMock.mockRejectedValueOnce({ status: 500, data: { error: "network blip" } });
    await ticket.post();
    const firstKey = ticket.pendingIdempotencyKey.value;
    expect(firstKey).toBeTruthy();

    postMock.mockResolvedValueOnce({ id: "cr-1", status: "PENDING", submitted_by: "u1", transaction_type: "CASH_IN" });
    await ticket.post();
    const secondKey = ticket.pendingIdempotencyKey.value;

    expect(secondKey).toBe(firstKey);
    expect(postMock.mock.calls[0]![2]).toBe(firstKey);
    expect(postMock.mock.calls[1]![2]).toBe(firstKey);
  });

  it("mints a new key once the draft changes after a failed attempt", async () => {
    const ticket = useCashTicket({ simulate: simulateMock, post: postMock });
    ticket.open("B14-CORE");
    withFilledDraft(ticket);

    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 1 },
      cash: { current_balance: "1000", projected_balance: "2000", cash_impact: "1000", currency: "THB" },
    });
    await ticket.simulate();
    postMock.mockRejectedValueOnce({ status: 500, data: { error: "network blip" } });
    await ticket.post();
    const firstKey = ticket.pendingIdempotencyKey.value;
    expect(firstKey).toBeTruthy();

    // A genuinely different movement (amount changed) must never reuse the
    // first attempt's key.
    ticket.patchDraft({ gross_amount: "2000" });
    expect(ticket.pendingIdempotencyKey.value).toBeNull();

    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 1 },
      cash: { current_balance: "1000", projected_balance: "3000", cash_impact: "2000", currency: "THB" },
    });
    await ticket.simulate();
    postMock.mockResolvedValueOnce({ id: "cr-2", status: "PENDING", submitted_by: "u1", transaction_type: "CASH_IN" });
    await ticket.post();

    expect(ticket.pendingIdempotencyKey.value).not.toBe(firstKey);
    expect(postMock.mock.calls[1]![2]).not.toBe(firstKey);
  });
});

describe("isCashRequestResponse", () => {
  it("returns true for a CashRequestResponse (has submitted_by)", () => {
    expect(
      isCashRequestResponse({ id: "x", status: "PENDING", submitted_by: "u1" }),
    ).toBe(true);
  });

  it("returns false for a TransactionResponse (no submitted_by)", () => {
    expect(isCashRequestResponse({ id: "x", status: "POSTED" })).toBe(false);
  });
});
