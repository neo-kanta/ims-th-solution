/**
 * Tests for the Order Ticket composable: payload mapping, validation,
 * and the simulate-then-post lifecycle (including the BLOCK-disables-post
 * acceptance criterion).
 *
 * The API service module is mocked because it relies on Nuxt's runtime
 * client; this file exercises the composable's logic in isolation.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const simulateMock = vi.fn();
const postMock = vi.fn();

vi.mock("../app/api/openapi", () => ({
  useOpenApiClient: () => ({ GET: vi.fn(), POST: vi.fn() }),
  unwrapOpenApiResponse: <T,>(value: { data?: T }) => value.data as T,
}));

vi.mock("../app/features/investment-ledger/services/investmentLedgerApi", () => ({
  investmentLedgerApi: {
    simulateTransaction: (...args: unknown[]) => simulateMock(...args),
    postTransaction: (...args: unknown[]) => postMock(...args),
  },
}));

import {
  buildTransactionRequest,
  useOrderTicket,
  validateDraft,
} from "../app/features/investment-ledger/composables/useOrderTicket";

const instrumentId = "11111111-1111-1111-1111-111111111111";

function withFilledDraft(ticket: ReturnType<typeof useOrderTicket>) {
  ticket.patchDraft({
    instrument_id: instrumentId,
    quantity: "100",
    price: "50",
    fees: "10",
    currency: "THB",
    business_date: "2026-05-27",
  });
}

beforeEach(() => {
  simulateMock.mockReset();
  postMock.mockReset();
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("buildTransactionRequest", () => {
  it("maps the draft into the backend wire format with both transaction_type and side", () => {
    const body = buildTransactionRequest({
      side: "BUY",
      instrument_id: instrumentId,
      quantity: "10",
      price: "100.5",
      fees: "1.25",
      currency: "THB",
      business_date: "2026-05-27",
      settlement_date: "2026-05-29",
      external_ref: "ORD-1",
      reason: "rebalance",
    });

    expect(body).toEqual({
      transaction_type: "BUY",
      side: "BUY",
      instrument_id: instrumentId,
      quantity: "10",
      price: "100.5",
      fees: "1.25",
      currency: "THB",
      business_date: "2026-05-27",
      settlement_date: "2026-05-29",
      external_ref: "ORD-1",
      reason: "rebalance",
    });
  });

  it("drops empty optional fields rather than sending blank strings", () => {
    const body = buildTransactionRequest({
      side: "SELL",
      instrument_id: instrumentId,
      quantity: "10",
      price: "100",
      fees: "",
      currency: "THB",
      business_date: "2026-05-27",
      settlement_date: "",
      external_ref: "",
      reason: "",
    });

    expect(body.settlement_date).toBeUndefined();
    expect(body.fees).toBeUndefined();
    expect(body.external_ref).toBeUndefined();
    expect(body.reason).toBeUndefined();
    expect(body.transaction_type).toBe("SELL");
    expect(body.side).toBe("SELL");
  });
});

describe("validateDraft", () => {
  it("requires instrument, quantity, price, currency, and business date", () => {
    const errors = validateDraft({
      side: "BUY",
      instrument_id: "",
      quantity: "",
      price: "",
      fees: "",
      currency: "",
      business_date: "",
      settlement_date: "",
      external_ref: "",
      reason: "",
    });
    expect(errors.instrument_id).toBeDefined();
    expect(errors.quantity).toBeDefined();
    expect(errors.price).toBeDefined();
    expect(errors.currency).toBeDefined();
    expect(errors.business_date).toBeDefined();
  });

  it("rejects non-numeric / non-positive quantity and price", () => {
    const errors = validateDraft({
      side: "BUY",
      instrument_id: instrumentId,
      quantity: "abc",
      price: "-1",
      fees: "",
      currency: "THB",
      business_date: "2026-05-27",
      settlement_date: "",
      external_ref: "",
      reason: "",
    });
    expect(errors.quantity).toMatch(/positive/i);
    expect(errors.price).toMatch(/positive/i);
  });

  it("rejects a non-ISO currency code", () => {
    const errors = validateDraft({
      side: "BUY",
      instrument_id: instrumentId,
      quantity: "1",
      price: "1",
      fees: "",
      currency: "thb",
      business_date: "2026-05-27",
      settlement_date: "",
      external_ref: "",
      reason: "",
    });
    expect(errors.currency).toBeDefined();
  });

  it("accepts a fully populated draft", () => {
    expect(
      validateDraft({
        side: "BUY",
        instrument_id: instrumentId,
        quantity: "5",
        price: "10",
        fees: "0",
        currency: "THB",
        business_date: "2026-05-27",
        settlement_date: "2026-05-29",
        external_ref: "ORD-1",
        reason: "",
      }),
    ).toEqual({});
  });
});

describe("useOrderTicket lifecycle", () => {
  it("blocks post when verdict is BLOCK", async () => {
    simulateMock.mockResolvedValueOnce({
      compliance: {
        verdict: "BLOCK",
        breaches: [{ rule_type_id: "limit_eq", verdict: "BLOCK" }],
        rules_evaluated: 4,
      },
      cash: { current_balance: "1000", projected_balance: "500", cash_impact: "-500", currency: "THB" },
      position: {
        current_quantity: "0",
        projected_quantity: "100",
      },
    });

    const ticket = useOrderTicket();
    ticket.open("portfolio-1");
    withFilledDraft(ticket);

    expect(ticket.isValid.value).toBe(true);

    const result = await ticket.simulate();
    expect(result?.compliance?.verdict).toBe("BLOCK");
    expect(ticket.isBlocked.value).toBe(true);
    expect(ticket.canPost.value).toBe(false);

    // Attempting to post must not call the API and must surface an error.
    await ticket.post();
    expect(postMock).not.toHaveBeenCalled();
    expect(ticket.postError.value).toMatch(/BLOCK/);
  });

  it("allows post after a PASS simulation and refreshes success state", async () => {
    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 6 },
      cash: { current_balance: "1000", projected_balance: "500", cash_impact: "-500", currency: "THB" },
      position: { current_quantity: "0", projected_quantity: "100" },
    });
    postMock.mockResolvedValueOnce({ id: "txn-1", status: "POSTED" });

    const ticket = useOrderTicket();
    ticket.open("portfolio-1");
    withFilledDraft(ticket);

    await ticket.simulate();
    expect(ticket.verdict.value).toBe("PASS");
    expect(ticket.canPost.value).toBe(true);

    const result = await ticket.post();
    expect(postMock).toHaveBeenCalledTimes(1);
    expect(result?.id).toBe("txn-1");
    expect(ticket.stage.value).toBe("posted");
    expect(ticket.successMessage.value).toContain("posted");
  });

  it("invalidates the simulation when the draft mutates afterwards", async () => {
    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 2 },
      cash: { current_balance: "1000", projected_balance: "500", cash_impact: "-500", currency: "THB" },
      position: { current_quantity: "0", projected_quantity: "100" },
    });

    const ticket = useOrderTicket();
    ticket.open("portfolio-1");
    withFilledDraft(ticket);
    await ticket.simulate();
    expect(ticket.isSimulationFresh.value).toBe(true);

    ticket.patchDraft({ quantity: "200" });
    expect(ticket.isSimulationFresh.value).toBe(false);
    expect(ticket.canPost.value).toBe(false);
  });

  it("classifies a 403 from simulate as a permission error", async () => {
    simulateMock.mockRejectedValueOnce({ status: 403, data: { error: "forbidden" } });

    const ticket = useOrderTicket();
    ticket.open("portfolio-1");
    withFilledDraft(ticket);

    await ticket.simulate();
    expect(ticket.simulationError.value).toMatch(/INVESTMENT_LEDGER_SIMULATE/);
  });

  it("forces re-simulation when a 409 conflict is returned on post", async () => {
    simulateMock.mockResolvedValueOnce({
      compliance: { verdict: "PASS", breaches: [], rules_evaluated: 2 },
      cash: { current_balance: "1000", projected_balance: "500", cash_impact: "-500", currency: "THB" },
      position: { current_quantity: "0", projected_quantity: "100" },
    });
    postMock.mockRejectedValueOnce({ status: 409 });

    const ticket = useOrderTicket();
    ticket.open("portfolio-1");
    withFilledDraft(ticket);

    await ticket.simulate();
    await ticket.post();
    expect(ticket.postError.value).toMatch(/changed/i);
    expect(ticket.stage.value).toBe("simulated");
  });
});
