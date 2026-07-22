import { computed, reactive, ref } from "vue";

import {
  investmentLedgerApi,
  type ApiPostTransactionRequest,
  type ApiTransaction,
  type ApiTransactionSimulation,
} from "../services/investmentLedgerApi";
import {
  isNonNegativeNumeric,
  isPositiveNumeric,
  todayIso,
  type OrderSide,
} from "../lib/ledgerFormat";

export interface OrderTicketDraft {
  side: OrderSide;
  instrument_id: string;
  quantity: string;
  price: string;
  fees: string;
  currency: string;
  business_date: string;
  settlement_date: string;
  external_ref: string;
  reason: string;
}

export interface OrderTicketErrors {
  instrument_id?: string;
  quantity?: string;
  price?: string;
  fees?: string;
  currency?: string;
  business_date?: string;
  settlement_date?: string;
}

export interface OrderTicketDefaults {
  currency?: string;
  business_date?: string;
}

function emptyDraft(defaults: OrderTicketDefaults = {}): OrderTicketDraft {
  return {
    side: "BUY",
    instrument_id: "",
    quantity: "",
    price: "",
    fees: "",
    currency: defaults.currency ?? "",
    business_date: defaults.business_date ?? todayIso(),
    settlement_date: "",
    external_ref: "",
    reason: "",
  };
}

function isIsoDate(value: string): boolean {
  return /^\d{4}-\d{2}-\d{2}$/.test(value);
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
  const s = (err as { status?: unknown; statusCode?: unknown }).status;
  if (typeof s === "number") return s;
  const sc = (err as { statusCode?: unknown }).statusCode;
  if (typeof sc === "number") return sc;
  return null;
}

/**
 * Builds the request body the backend expects from the local draft. Empty
 * strings are dropped so the backend can apply its own defaults / required
 * checks rather than seeing `""` for an unset optional field.
 *
 * Exported so unit tests can pin the wire shape.
 */
export function buildTransactionRequest(
  draft: OrderTicketDraft,
): ApiPostTransactionRequest {
  const body: ApiPostTransactionRequest = {
    business_date: draft.business_date,
    currency: draft.currency,
    transaction_type: draft.side,
    side: draft.side,
  };
  if (draft.instrument_id) body.instrument_id = draft.instrument_id;
  if (draft.quantity) body.quantity = draft.quantity;
  if (draft.price) body.price = draft.price;
  if (draft.fees) body.fees = draft.fees;
  if (draft.settlement_date) body.settlement_date = draft.settlement_date;
  if (draft.external_ref) body.external_ref = draft.external_ref;
  if (draft.reason) body.reason = draft.reason;
  return body;
}

export function validateDraft(draft: OrderTicketDraft): OrderTicketErrors {
  const errors: OrderTicketErrors = {};
  if (!draft.instrument_id) {
    errors.instrument_id = "Select an instrument.";
  }
  if (!draft.quantity) {
    errors.quantity = "Quantity is required.";
  } else if (!isPositiveNumeric(draft.quantity)) {
    errors.quantity = "Quantity must be a positive number.";
  }
  if (!draft.price) {
    errors.price = "Price is required.";
  } else if (!isPositiveNumeric(draft.price)) {
    errors.price = "Price must be a positive number.";
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
  if (draft.settlement_date && !isIsoDate(draft.settlement_date)) {
    errors.settlement_date = "Use YYYY-MM-DD format.";
  }
  return errors;
}

export type OrderTicketStage =
  | "edit"
  | "simulated"
  | "confirming"
  | "posted";

/**
 * Injectable simulate/post calls. Defaults to the legacy portfolio-UUID
 * routes (`investmentLedgerApi`). Callers on a portfolio-code-routed
 * workspace (Portfolio V2) should inject `portfolioApi.simulateTransaction`/
 * `portfolioApi.postTransaction` instead — the identifier passed to `open()`
 * then becomes a portfolio code rather than a UUID. The two are wire-
 * compatible: both routes decode/return the same generated
 * `PostTransactionRequest`/`TransactionSimulationResponse`/`TransactionResponse`
 * schemas, so no logic here needs to change, only which endpoint is called.
 */
export interface OrderTicketDeps {
  simulate: (
    portfolioIdentifier: string,
    payload: ApiPostTransactionRequest,
  ) => Promise<ApiTransactionSimulation>;
  post: (
    portfolioIdentifier: string,
    payload: ApiPostTransactionRequest,
  ) => Promise<ApiTransaction>;
}

const defaultDeps: OrderTicketDeps = {
  simulate: (portfolioId, payload) =>
    investmentLedgerApi.simulateTransaction(portfolioId, payload),
  post: (portfolioId, payload) =>
    investmentLedgerApi.postTransaction(portfolioId, payload),
};

export function useOrderTicket(deps: OrderTicketDeps = defaultDeps) {
  const draft = reactive<OrderTicketDraft>(emptyDraft());
  const stage = ref<OrderTicketStage>("edit");
  const portfolioId = ref<string | null>(null);

  const lastSimulation = ref<ApiTransactionSimulation | null>(null);
  const lastSimulatedPayload = ref<ApiPostTransactionRequest | null>(null);
  const lastPostedTransaction = ref<ApiTransaction | null>(null);

  const simulating = ref(false);
  const posting = ref(false);

  const simulationError = ref<string | null>(null);
  const postError = ref<string | null>(null);
  const successMessage = ref<string | null>(null);

  const validationErrors = computed<OrderTicketErrors>(() =>
    validateDraft(draft),
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
      JSON.stringify(buildTransactionRequest(draft))
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

  function open(activePortfolioId: string, defaults: OrderTicketDefaults = {}) {
    portfolioId.value = activePortfolioId;
    Object.assign(draft, emptyDraft(defaults));
    stage.value = "edit";
    lastSimulation.value = null;
    lastSimulatedPayload.value = null;
    lastPostedTransaction.value = null;
    simulationError.value = null;
    postError.value = null;
    successMessage.value = null;
  }

  function close() {
    stage.value = "edit";
    simulationError.value = null;
    postError.value = null;
    successMessage.value = null;
  }

  function patchDraft(patch: Partial<OrderTicketDraft>) {
    Object.assign(draft, patch);
    // Mutating the form invalidates the previous simulation result.
    if (stage.value !== "edit") {
      stage.value = "edit";
    }
  }

  function setSide(side: OrderSide) {
    patchDraft({ side });
  }

  function classifySimulationError(err: unknown): string {
    const status = statusOf(err);
    const backend = extractMessage(err, "Simulation failed.");
    if (status === 403) {
      return "You do not have permission to simulate transactions (INVESTMENT_LEDGER_SIMULATE).";
    }
    if (status === 404) {
      return "Portfolio or instrument was not found.";
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

  async function simulate(): Promise<ApiTransactionSimulation | null> {
    if (!portfolioId.value) {
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
      const payload = buildTransactionRequest(draft);
      const result = await deps.simulate(portfolioId.value, payload);
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

  async function post(): Promise<ApiTransaction | null> {
    if (!portfolioId.value) {
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
      const result = await deps.post(
        portfolioId.value,
        lastSimulatedPayload.value,
      );
      lastPostedTransaction.value = result;
      successMessage.value = `${draft.side} order posted (ref ${result.id ?? "n/a"}).`;
      stage.value = "posted";
      return result;
    } catch (err) {
      postError.value = classifyPostError(err);
      // Drop the confirmation modal so the user can adjust the form.
      stage.value = "simulated";
      return null;
    } finally {
      posting.value = false;
    }
  }

  function clearAfterPost() {
    lastSimulation.value = null;
    lastSimulatedPayload.value = null;
  }

  return {
    draft,
    stage,
    portfolioId,
    lastSimulation,
    lastSimulatedPayload,
    lastPostedTransaction,
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
    close,
    patchDraft,
    setSide,
    simulate,
    requestConfirmation,
    cancelConfirmation,
    post,
    clearAfterPost,
  };
}
