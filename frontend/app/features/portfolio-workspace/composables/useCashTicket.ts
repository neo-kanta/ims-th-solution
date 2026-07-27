import { computed, reactive, ref } from "vue";

import {
  portfolioApi,
  type ApiCashRequestV2,
  type ApiPostTransactionRequestV2,
  type ApiTransactionSimulationV2,
  type ApiTransactionV2,
} from "../services/portfolioApi";
import { isCashRequestResponse } from "../lib/cashRequestGuard";
import {
  isNonNegativeNumeric,
  isPositiveNumeric,
  todayIso,
} from "../../investment-ledger/lib/ledgerFormat";

export const CASH_TRANSACTION_TYPES = [
  "CASH_IN",
  "CASH_OUT",
  "FEE",
  "DIVIDEND",
] as const;
export type CashTransactionType = (typeof CASH_TRANSACTION_TYPES)[number];

export interface CashTicketDraft {
  transaction_type: CashTransactionType;
  gross_amount: string;
  fees: string;
  currency: string;
  business_date: string;
  reason: string;
}

export interface CashTicketErrors {
  gross_amount?: string;
  fees?: string;
  currency?: string;
  business_date?: string;
}

export interface CashTicketDefaults {
  currency?: string;
  business_date?: string;
}

function emptyDraft(defaults: CashTicketDefaults = {}): CashTicketDraft {
  return {
    transaction_type: "CASH_IN",
    gross_amount: "",
    fees: "",
    currency: defaults.currency ?? "",
    business_date: defaults.business_date ?? todayIso(),
    reason: "",
  };
}

function isIsoDate(value: string): boolean {
  return /^\d{4}-\d{2}-\d{2}$/.test(value);
}

// Mirrors useMarketDataSecurityDetail.ts's generateIdempotencyKey guard: some
// environments (older browsers; certain SSR/test contexts) may not expose a
// global crypto.randomUUID, so fall back to a non-cryptographic but still
// practically-unique key rather than throwing.
function generateIdempotencyKey(): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return `cash-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

function extractMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== "object") return fallback;
  const data = (err as { data?: { error?: unknown; message?: unknown } }).data;
  if (data && typeof data === "object") {
    if (typeof data.error === "string" && data.error.trim()) return data.error;
    if (typeof data.message === "string" && data.message.trim())
      return data.message;
  }
  const msg = (err as { message?: unknown }).message;
  if (typeof msg === "string" && msg.trim()) return msg;
  return fallback;
}

function statusOf(err: unknown): number | null {
  if (!err || typeof err !== "object") return null;
  const s = (err as { status?: unknown }).status;
  if (typeof s === "number") return s;
  const sc = (err as { statusCode?: unknown }).statusCode;
  if (typeof sc === "number") return sc;
  return null;
}

/**
 * Builds the request body the backend expects from the local cash draft.
 * Exported so unit tests can pin the wire shape.
 */
export function buildCashTransactionRequest(
  draft: CashTicketDraft,
): ApiPostTransactionRequestV2 {
  const body: ApiPostTransactionRequestV2 = {
    business_date: draft.business_date,
    currency: draft.currency,
    transaction_type: draft.transaction_type,
  };
  if (draft.gross_amount) body.gross_amount = draft.gross_amount;
  if (draft.fees) body.fees = draft.fees;
  if (draft.reason) body.reason = draft.reason;
  return body;
}

export function validateCashDraft(draft: CashTicketDraft): CashTicketErrors {
  const errors: CashTicketErrors = {};
  if (!draft.gross_amount) {
    errors.gross_amount = "Amount is required.";
  } else if (!isPositiveNumeric(draft.gross_amount)) {
    errors.gross_amount = "Amount must be a positive number.";
  }
  if (draft.fees && !isNonNegativeNumeric(draft.fees)) {
    errors.fees = "Fees must be zero or a positive number.";
  }
  if (!draft.currency) {
    errors.currency = "Currency is required.";
  } else if (!/^[A-Z]{3}$/.test(draft.currency)) {
    errors.currency = "Use a 3-letter ISO currency code.";
  }
  if (!draft.business_date) {
    errors.business_date = "Business date is required.";
  } else if (!isIsoDate(draft.business_date)) {
    errors.business_date = "Use YYYY-MM-DD format.";
  }
  return errors;
}

export type CashTicketStage =
  | "edit"
  | "simulated"
  | "confirming"
  | "posted"
  | "pending";

export interface CashTicketDeps {
  simulate: (
    portfolioCode: string,
    payload: ApiPostTransactionRequestV2,
  ) => Promise<ApiTransactionSimulationV2>;
  post: (
    portfolioCode: string,
    payload: ApiPostTransactionRequestV2,
    idempotencyKey: string,
  ) => Promise<ApiTransactionV2 | ApiCashRequestV2>;
}

const defaultDeps: CashTicketDeps = {
  simulate: (code, payload) => portfolioApi.simulateTransaction(code, payload),
  post: (code, payload, idempotencyKey) =>
    portfolioApi.postTransaction(code, payload, idempotencyKey),
};

/**
 * Mirrors useOrderTicket's simulate -> confirm -> post stage machine (same
 * simulate/post endpoints, same fresh-simulation and BLOCK guards) but for
 * cash movements (CASH_IN/CASH_OUT/FEE/DIVIDEND) rather than BUY/SELL. Kept
 * as a separate composable instead of extending useOrderTicket because the
 * draft shape (amount, no instrument/quantity/price) and the LIVE-cash
 * outcome (a 202 pending CashRequestResponse, not a posted TransactionResponse)
 * genuinely differ, and useOrderTicket is shared with the legacy V1 order
 * ticket which never sees a pending outcome.
 */
export function useCashTicket(deps: CashTicketDeps = defaultDeps) {
  const draft = reactive<CashTicketDraft>(emptyDraft());
  const stage = ref<CashTicketStage>("edit");
  const portfolioCode = ref<string | null>(null);

  const lastSimulation = ref<ApiTransactionSimulationV2 | null>(null);
  const lastSimulatedPayload = ref<ApiPostTransactionRequestV2 | null>(null);
  const lastPostedTransaction = ref<ApiTransactionV2 | null>(null);
  const lastCashRequest = ref<ApiCashRequestV2 | null>(null);
  // Generated once per simulated payload and reused across retries of the
  // SAME submission attempt (e.g. a network error after the request actually
  // reached the backend) so the backend's idempotency-key dedup can resolve
  // the retry to the original request instead of creating a duplicate. Reset
  // whenever the draft changes (a genuinely new submission gets a new key).
  const pendingIdempotencyKey = ref<string | null>(null);

  const simulating = ref(false);
  const posting = ref(false);

  const simulationError = ref<string | null>(null);
  const postError = ref<string | null>(null);
  const successMessage = ref<string | null>(null);

  const validationErrors = computed<CashTicketErrors>(() =>
    validateCashDraft(draft),
  );

  const isValid = computed<boolean>(
    () => Object.keys(validationErrors.value).length === 0,
  );

  const isSimulationFresh = computed<boolean>(() => {
    if (!lastSimulatedPayload.value || !lastSimulation.value) return false;
    if (stage.value !== "simulated" && stage.value !== "confirming") {
      return false;
    }
    return (
      JSON.stringify(lastSimulatedPayload.value) ===
      JSON.stringify(buildCashTransactionRequest(draft))
    );
  });

  const verdict = computed<string | null>(
    () => lastSimulation.value?.compliance?.verdict ?? null,
  );

  const isBlocked = computed<boolean>(() => verdict.value === "BLOCK");

  const canPost = computed<boolean>(
    () =>
      isValid.value &&
      isSimulationFresh.value &&
      !isBlocked.value &&
      !simulating.value &&
      !posting.value,
  );

  function open(activePortfolioCode: string, defaults: CashTicketDefaults = {}) {
    portfolioCode.value = activePortfolioCode;
    Object.assign(draft, emptyDraft(defaults));
    stage.value = "edit";
    lastSimulation.value = null;
    lastSimulatedPayload.value = null;
    lastPostedTransaction.value = null;
    lastCashRequest.value = null;
    pendingIdempotencyKey.value = null;
    simulationError.value = null;
    postError.value = null;
    successMessage.value = null;
  }

  function patchDraft(patch: Partial<CashTicketDraft>) {
    Object.assign(draft, patch);
    if (stage.value !== "edit") {
      stage.value = "edit";
    }
    // A changed draft is a genuinely new intended movement — never reuse a
    // key generated for the previous (now stale) payload.
    pendingIdempotencyKey.value = null;
  }

  function setTransactionType(type: CashTransactionType) {
    patchDraft({ transaction_type: type });
  }

  function classifySimulationError(err: unknown): string {
    const status = statusOf(err);
    const backend = extractMessage(err, "Simulation failed.");
    if (status === 403) {
      return "You do not have permission to simulate transactions (INVESTMENT_LEDGER_SIMULATE).";
    }
    if (status === 404) {
      return "Portfolio was not found.";
    }
    if (status === 422 || status === 409 || status === 400) {
      return backend;
    }
    return status ? `${status} · ${backend}` : backend;
  }

  function classifyPostError(err: unknown): string {
    const status = statusOf(err);
    const backend = extractMessage(err, "Post failed.");
    if (status === 403) {
      return "You do not have permission to post transactions (INVESTMENT_LEDGER_POST).";
    }
    if (status === 409) {
      return "The portfolio state changed since the last simulation. Re-run the simulation and try again.";
    }
    if (status === 422) {
      return backend;
    }
    return status ? `${status} · ${backend}` : backend;
  }

  async function simulate(): Promise<ApiTransactionSimulationV2 | null> {
    if (!portfolioCode.value) {
      simulationError.value = "Select a portfolio first.";
      return null;
    }
    if (!isValid.value) {
      simulationError.value = "Fix the highlighted fields before simulating.";
      return null;
    }
    simulating.value = true;
    simulationError.value = null;
    postError.value = null;
    successMessage.value = null;
    try {
      const payload = buildCashTransactionRequest(draft);
      const result = await deps.simulate(portfolioCode.value, payload);
      lastSimulation.value = result;
      lastSimulatedPayload.value = payload;
      stage.value = "simulated";
      return result;
    } catch (err) {
      simulationError.value = classifySimulationError(err);
      lastSimulation.value = null;
      lastSimulatedPayload.value = null;
      stage.value = "edit";
      return null;
    } finally {
      simulating.value = false;
    }
  }

  function requestConfirmation() {
    if (!canPost.value) return;
    stage.value = "confirming";
  }

  function cancelConfirmation() {
    if (stage.value === "confirming") {
      stage.value = "simulated";
    }
  }

  async function post(): Promise<ApiTransactionV2 | ApiCashRequestV2 | null> {
    if (!portfolioCode.value) {
      postError.value = "Select a portfolio first.";
      return null;
    }
    if (!isSimulationFresh.value || !lastSimulatedPayload.value) {
      postError.value =
        "Re-run the simulation — the order details changed since the last preview.";
      stage.value = "edit";
      return null;
    }
    if (isBlocked.value) {
      postError.value = "Compliance verdict is BLOCK. The order cannot be posted.";
      return null;
    }
    posting.value = true;
    postError.value = null;
    try {
      // Generate the key once per simulated payload and reuse it across
      // retries of this same attempt (see pendingIdempotencyKey above).
      if (!pendingIdempotencyKey.value) {
        pendingIdempotencyKey.value = generateIdempotencyKey();
      }
      const result = await deps.post(
        portfolioCode.value,
        lastSimulatedPayload.value,
        pendingIdempotencyKey.value,
      );
      if (isCashRequestResponse(result)) {
        lastCashRequest.value = result;
        lastPostedTransaction.value = null;
        stage.value = "pending";
      } else {
        lastPostedTransaction.value = result;
        lastCashRequest.value = null;
        stage.value = "posted";
      }
      return result;
    } catch (err) {
      postError.value = classifyPostError(err);
      stage.value = "simulated";
      return null;
    } finally {
      posting.value = false;
    }
  }

  return {
    draft,
    stage,
    portfolioCode,
    lastSimulation,
    lastSimulatedPayload,
    lastPostedTransaction,
    lastCashRequest,
    pendingIdempotencyKey,
    simulating,
    posting,
    simulationError,
    postError,
    successMessage,
    validationErrors,
    isValid,
    isSimulationFresh,
    verdict,
    isBlocked,
    canPost,
    open,
    patchDraft,
    setTransactionType,
    simulate,
    requestConfirmation,
    cancelConfirmation,
    post,
  };
}
