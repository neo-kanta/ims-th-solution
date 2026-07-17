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

export interface DecisionDraftErrors {
  instrument?: string;
  side?: string;
  quantity?: string;
  amount?: string;
  limitPrice?: string;
  currency?: string;
  businessDate?: string;
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
    errors.instrument = "Select an instrument.";
  } else if (
    draft.side === "SELL" &&
    draft.instrumentId &&
    options.isOwnedInstrument === false
  ) {
    errors.instrument = "This portfolio does not hold this instrument — select an owned position to sell.";
  }
  if (draft.side !== "BUY" && draft.side !== "SELL") {
    errors.side = "Choose BUY or SELL.";
  }
  if (!draft.currency) {
    errors.currency = "Currency is required.";
  } else if (!isCurrencyCode(draft.currency)) {
    errors.currency = "Use a 3-letter uppercase ISO currency code.";
  }
  if (!draft.businessDate) {
    errors.businessDate = "Business date is required.";
  } else if (!isIsoDate(draft.businessDate)) {
    errors.businessDate = "Use YYYY-MM-DD format.";
  }

  const hasQuantity = draft.quantity.trim() !== "";
  const hasAmount = draft.amount.trim() !== "";
  if (!hasQuantity && !hasAmount) {
    errors.quantity = "Enter a quantity or an amount.";
  }
  if (hasQuantity && !isStrictPositiveDecimal(draft.quantity)) {
    errors.quantity = "Quantity must be a positive number (no scientific notation).";
  }
  if (hasAmount && !isStrictPositiveDecimal(draft.amount)) {
    errors.amount = "Amount must be a positive number (no scientific notation).";
  }
  if (draft.limitPrice.trim() && !isStrictPositiveDecimal(draft.limitPrice)) {
    errors.limitPrice = "Limit price must be a positive number (no scientific notation).";
  }

  if (
    !errors.quantity &&
    draft.side === "SELL" &&
    hasQuantity &&
    typeof options.availableQuantity === "number"
  ) {
    const requested = Number(draft.quantity);
    if (requested > options.availableQuantity) {
      errors.quantity = `Exceeds available holding (${options.availableQuantity}).`;
    }
  }

  return errors;
}
