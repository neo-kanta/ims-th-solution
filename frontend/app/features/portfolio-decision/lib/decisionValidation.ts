/**
 * Field-level validation for the portfolio decision order ticket.
 *
 * Mirrors the backend's actual acceptance rules
 * (backend/internal/investment/application/command/decision_lifecycle.go
 * validateCreateDecision + backend/internal/investment/transport/dto/request/requests_v2.go)
 * rather than the wider "everything optional" shape of the wire DTO:
 *   - instrument, side, currency (3-letter), business_date are required
 *   - quantity OR amount is required (limit_price is not)
 * Numeric fields are validated with a strict decimal regex — `Number()` alone
 * would accept scientific notation ("1e5"), which the spec forbids.
 */

import type {
  AppTranslationKey,
  TranslationParams,
} from "~/composables/useI18n";

const STRICT_POSITIVE_DECIMAL = /^\d+(\.\d+)?$/;

export function isStrictPositiveDecimal(value: string): boolean {
  if (!value || !STRICT_POSITIVE_DECIMAL.test(value.trim())) return false;
  return Number(value) > 0;
}

export function isIsoDate(value: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return false;
  const d = new Date(`${value}T00:00:00Z`);
  return !Number.isNaN(d.getTime()) && d.toISOString().slice(0, 10) === value;
}

export function isCurrencyCode(value: string): boolean {
  return /^[A-Z]{3}$/.test(value);
}

export interface DecisionDraftLike {
  side: "BUY" | "SELL";
  instrumentId: string | null;
  instrumentCode: string;
  quantity: string;
  amount: string;
  limitPrice: string;
  currency: string;
  businessDate: string;
}

export interface DecisionValidationError {
  key: AppTranslationKey;
  params?: TranslationParams;
}

export interface DecisionDraftErrors {
  instrument?: DecisionValidationError;
  side?: DecisionValidationError;
  quantity?: DecisionValidationError;
  amount?: DecisionValidationError;
  limitPrice?: DecisionValidationError;
  currency?: DecisionValidationError;
  businessDate?: DecisionValidationError;
}

export interface DecisionValidationOptions {
  /** Known available quantity for the selected instrument when side === "SELL". */
  availableQuantity?: number | null;
  /**
   * Whether the selected instrument is a position the portfolio actually
   * holds, known only when side === "SELL". `null`/`undefined` means
   * "unknown" (e.g. holdings haven't loaded yet) and is not treated as a
   * failure — only an explicit `false` blocks submission. This is a hard
   * gate, independent of the oversell-quantity check below, so a non-owned
   * instrument can never become a valid SELL selection even before a
   * quantity is typed.
   */
  isOwnedInstrument?: boolean | null;
}

export function validateDecisionDraft(
  draft: DecisionDraftLike,
  options: DecisionValidationOptions = {},
): DecisionDraftErrors {
  const errors: DecisionDraftErrors = {};

  if (!draft.instrumentCode.trim()) {
    errors.instrument = { key: "portfolio.decisionNew.validation.instrumentRequired" };
  } else if (
    draft.side === "SELL" &&
    draft.instrumentId &&
    options.isOwnedInstrument === false
  ) {
    errors.instrument = { key: "portfolio.decisionNew.validation.instrumentNotOwned" };
  }
  if (draft.side !== "BUY" && draft.side !== "SELL") {
    errors.side = { key: "portfolio.decisionNew.validation.sideRequired" };
  }
  if (!draft.currency) {
    errors.currency = { key: "portfolio.decisionNew.validation.currencyRequired" };
  } else if (!isCurrencyCode(draft.currency)) {
    errors.currency = { key: "portfolio.decisionNew.validation.currencyInvalid" };
  }
  if (!draft.businessDate) {
    errors.businessDate = { key: "portfolio.decisionNew.validation.businessDateRequired" };
  } else if (!isIsoDate(draft.businessDate)) {
    errors.businessDate = { key: "portfolio.decisionNew.validation.businessDateInvalid" };
  }

  const hasQuantity = draft.quantity.trim() !== "";
  const hasAmount = draft.amount.trim() !== "";
  if (!hasQuantity && !hasAmount) {
    errors.quantity = { key: "portfolio.decisionNew.validation.quantityOrAmount" };
  }
  if (hasQuantity && !isStrictPositiveDecimal(draft.quantity)) {
    errors.quantity = { key: "portfolio.decisionNew.validation.quantityInvalid" };
  }
  if (hasAmount && !isStrictPositiveDecimal(draft.amount)) {
    errors.amount = { key: "portfolio.decisionNew.validation.amountInvalid" };
  }
  if (draft.limitPrice.trim() && !isStrictPositiveDecimal(draft.limitPrice)) {
    errors.limitPrice = { key: "portfolio.decisionNew.validation.limitPriceInvalid" };
  }

  if (
    !errors.quantity &&
    draft.side === "SELL" &&
    hasQuantity &&
    typeof options.availableQuantity === "number"
  ) {
    const requested = Number(draft.quantity);
    if (requested > options.availableQuantity) {
      errors.quantity = {
        key: "portfolio.decisionNew.validation.exceedsHolding",
        params: { available: options.availableQuantity },
      };
    }
  }

  return errors;
}
