import { computed, reactive, ref } from "vue";

import type { ApiInstrument } from "../../investment-ledger/services/investmentLedgerApi";
import { todayIso, type OrderSide } from "../../investment-ledger/lib/ledgerFormat";
import type { ApiCreateDecisionV2Request } from "../services/portfolioDecisionApi";
import {
  validateDecisionDraft,
  type DecisionDraftErrors,
} from "../lib/decisionValidation";

export interface DecisionDraftState {
  side: OrderSide;
  instrumentId: string | null;
  instrumentCode: string;
  quantity: string;
  amount: string;
  limitPrice: string;
  currency: string;
  exchange: string;
  businessDate: string;
  rationale: string;
}

function emptyDraft(): DecisionDraftState {
  return {
    side: "BUY",
    instrumentId: null,
    instrumentCode: "",
    quantity: "",
    amount: "",
    limitPrice: "",
    currency: "",
    exchange: "",
    businessDate: todayIso(),
    rationale: "",
  };
}

/**
 * Owns the order-ticket draft. Field requirements mirror the backend's
 * actual acceptance rules (validateCreateDecision in
 * decision_lifecycle.go): instrument, side, currency, business_date are
 * required; quantity OR amount is required; limit_price is optional.
 */
export function useDecisionDraft(defaultCurrency = "") {
  const draft = reactive<DecisionDraftState>({
    ...emptyDraft(),
    currency: defaultCurrency,
  });
  const selectedInstrument = ref<ApiInstrument | null>(null);
  const availableQuantity = ref<number | null>(null);
  /** Set by the caller from the portfolio's holdings directory whenever
   * side === "SELL" and an instrument is selected; see decisionValidation's
   * isOwnedInstrument gate. */
  const isOwnedInstrument = ref<boolean | null>(null);

  function patch(partial: Partial<DecisionDraftState>) {
    Object.assign(draft, partial);
  }

  function selectInstrument(instrument: ApiInstrument) {
    selectedInstrument.value = instrument;
    patch({
      instrumentId: instrument.id ?? null,
      instrumentCode: instrument.primary_ticker ?? draft.instrumentCode,
      currency: draft.currency || instrument.currency || "",
      exchange: draft.exchange || instrument.primary_exchange || "",
    });
  }

  function clearInstrument() {
    selectedInstrument.value = null;
    availableQuantity.value = null;
    isOwnedInstrument.value = null;
    patch({ instrumentId: null, instrumentCode: "" });
  }

  function setSide(side: OrderSide) {
    patch({ side });
    if (side === "BUY") {
      availableQuantity.value = null;
      isOwnedInstrument.value = null;
    }
  }

  const errors = computed<DecisionDraftErrors>(() =>
    validateDecisionDraft(
      {
        side: draft.side,
        instrumentId: draft.instrumentId,
        instrumentCode: draft.instrumentCode,
        quantity: draft.quantity,
        amount: draft.amount,
        limitPrice: draft.limitPrice,
        currency: draft.currency,
        businessDate: draft.businessDate,
      },
      {
        availableQuantity: availableQuantity.value,
        isOwnedInstrument: isOwnedInstrument.value,
      },
    ),
  );

  const isValid = computed(() => Object.keys(errors.value).length === 0);

  function buildCreateBody(): ApiCreateDecisionV2Request {
    const body: ApiCreateDecisionV2Request = {
      instrument_code: draft.instrumentCode.trim().toUpperCase(),
      business_date: draft.businessDate,
      side: draft.side,
      currency: draft.currency.trim().toUpperCase(),
    };
    if (draft.instrumentId) body.instrument_id = draft.instrumentId;
    if (draft.quantity.trim()) body.quantity = draft.quantity.trim();
    if (draft.amount.trim()) body.amount = draft.amount.trim();
    if (draft.limitPrice.trim()) body.limit_price = draft.limitPrice.trim();
    if (draft.exchange.trim()) body.exchange = draft.exchange.trim();
    if (draft.rationale.trim()) body.rationale = draft.rationale.trim();
    return body;
  }

  function reset(currency = defaultCurrency) {
    Object.assign(draft, emptyDraft(), { currency });
    selectedInstrument.value = null;
    availableQuantity.value = null;
    isOwnedInstrument.value = null;
  }

  return {
    draft,
    selectedInstrument,
    availableQuantity,
    isOwnedInstrument,
    errors,
    isValid,
    patch,
    selectInstrument,
    clearInstrument,
    setSide,
    buildCreateBody,
    reset,
  };
}
