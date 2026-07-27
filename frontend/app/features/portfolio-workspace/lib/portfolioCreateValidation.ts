/**
 * Pure client-side validation/normalization for the Create Portfolio form
 * (`CreatePortfolioV2Request` — see `ims-api.d.ts`). No Nuxt/`~` imports so
 * this module (and its test) loads under bare `vitest run`.
 *
 * Client validation is a UX improvement only — the backend re-validates
 * everything and is the authoritative source of truth (see CLAUDE.md).
 */

export interface PortfolioCreateFormValues {
  /** Whether this portfolio is bound to a fund ("Bind with Fund: Y/N"). When
   * false, fund_code is never sent — the backend creates a fund-less
   * portfolio (fund_id NULL). */
  bindFund: boolean;
  fund_code: string;
  portfolio_type: string;
  code: string;
  name: string;
  description: string;
  base_currency: string;
  valuation_currency: string;
  strategy_code: string;
  benchmark: string;
  risk_profile: string;
  inception_date: string;
}

export const PORTFOLIO_TYPES = ["LIVE", "SIMULATION", "MODEL"] as const;
export type PortfolioTypeOption = (typeof PORTFOLIO_TYPES)[number];

/**
 * Fixed ISO 4217 choices offered in the Create Portfolio currency comboBoxes.
 * THB/USD are the only currencies present in real seed data today; the rest
 * are included as plausible near-term options so the field stays a picker
 * rather than free text. The backend accepts any 3-letter ISO code — this
 * list is a frontend UX narrowing only, not a backend-enforced allowlist.
 */
export const SUPPORTED_CURRENCIES = [
  "THB",
  "USD",
  "EUR",
  "GBP",
  "JPY",
  "SGD",
  "HKD",
  "CNY",
  "AUD",
] as const;

/**
 * risk_profile is NOT free text — the database enforces
 * `chk_inv_portfolios_risk_profile`: NULL or one of these four values only.
 * A number (or any other string) fails at INSERT time with a raw Postgres
 * constraint-violation error, which is why this must be a picker, not an
 * <input type="number"> (a mistake made and then fixed in this file).
 */
export const RISK_PROFILES = ["LOW", "MEDIUM", "HIGH", "SPECULATIVE"] as const;
export type RiskProfileOption = (typeof RISK_PROFILES)[number];

export type PortfolioCreateFieldError =
  | "required"
  | "invalid_currency"
  | "invalid_date"
  | "invalid_portfolio_type";

export type PortfolioCreateErrors = Partial<
  Record<keyof PortfolioCreateFormValues, PortfolioCreateFieldError>
>;

export interface PortfolioCreateValidationResult {
  valid: boolean;
  errors: PortfolioCreateErrors;
}

const ISO_CURRENCY_RE = /^[A-Za-z]{3}$/;
const ISO_DATE_RE = /^\d{4}-\d{2}-\d{2}$/;

/**
 * Trims every text field and uppercases `portfolio_type`/currency codes —
 * the backend normalizes/validates too, but this avoids sending obviously
 * wrong-cased input the user could have fixed for them.
 */
export function normalizePortfolioCreateValues(
  values: PortfolioCreateFormValues,
): PortfolioCreateFormValues {
  return {
    bindFund: values.bindFund,
    fund_code: values.fund_code.trim(),
    portfolio_type: values.portfolio_type.trim().toUpperCase(),
    code: values.code.trim(),
    name: values.name.trim(),
    description: values.description.trim(),
    base_currency: values.base_currency.trim().toUpperCase(),
    valuation_currency: values.valuation_currency.trim().toUpperCase(),
    strategy_code: values.strategy_code.trim(),
    benchmark: values.benchmark.trim(),
    risk_profile: values.risk_profile.trim().toUpperCase(),
    inception_date: values.inception_date.trim(),
  };
}

function isValidCalendarDate(value: string): boolean {
  if (!ISO_DATE_RE.test(value)) return false;
  const parsed = new Date(`${value}T00:00:00Z`);
  if (Number.isNaN(parsed.getTime())) return false;
  // Reject overflowed dates like 2026-02-30, which Date silently rolls
  // forward into March.
  return parsed.toISOString().slice(0, 10) === value;
}

export function validatePortfolioCreateForm(
  raw: PortfolioCreateFormValues,
): PortfolioCreateValidationResult {
  const values = normalizePortfolioCreateValues(raw);
  const errors: PortfolioCreateErrors = {};

  if (values.bindFund && !values.fund_code) errors.fund_code = "required";
  if (!values.portfolio_type) {
    errors.portfolio_type = "required";
  } else if (!(PORTFOLIO_TYPES as readonly string[]).includes(values.portfolio_type)) {
    errors.portfolio_type = "invalid_portfolio_type";
  }
  if (!values.code) errors.code = "required";
  if (!values.name) errors.name = "required";

  if (!values.base_currency) {
    errors.base_currency = "required";
  } else if (!ISO_CURRENCY_RE.test(values.base_currency)) {
    errors.base_currency = "invalid_currency";
  }

  if (!values.valuation_currency) {
    errors.valuation_currency = "required";
  } else if (!ISO_CURRENCY_RE.test(values.valuation_currency)) {
    errors.valuation_currency = "invalid_currency";
  }

  if (!values.inception_date) {
    errors.inception_date = "required";
  } else if (!isValidCalendarDate(values.inception_date)) {
    errors.inception_date = "invalid_date";
  }

  return { valid: Object.keys(errors).length === 0, errors };
}

/**
 * Shape matching `CreatePortfolioV2Request` — built from normalized form
 * values. Optional text fields are omitted entirely (not sent as empty
 * strings) when blank. `style_id`/`manager_user_id` are deliberately never
 * populated here: no typed style/manager directory exists in the frontend
 * today, so the form never asks the user to type either UUID (see the
 * work-package report for this gap).
 */
export interface PortfolioCreateRequestBody {
  fund_code?: string;
  portfolio_type: PortfolioTypeOption;
  code: string;
  name: string;
  description?: string;
  base_currency: string;
  valuation_currency: string;
  strategy_code?: string;
  benchmark?: string;
  risk_profile?: string;
  inception_date: string;
}

/**
 * Callers must run `validatePortfolioCreateForm` (or otherwise guarantee a
 * valid `portfolio_type`) before calling this — it assumes normalized values
 * already passed validation, matching `usePortfolioCreateForm.submit`'s
 * validate-then-build order.
 */
export function buildPortfolioCreateRequest(
  raw: PortfolioCreateFormValues,
): PortfolioCreateRequestBody {
  const values = normalizePortfolioCreateValues(raw);
  const body: PortfolioCreateRequestBody = {
    portfolio_type: values.portfolio_type as PortfolioTypeOption,
    code: values.code,
    name: values.name,
    base_currency: values.base_currency,
    valuation_currency: values.valuation_currency,
    inception_date: values.inception_date,
  };
  if (values.bindFund && values.fund_code) body.fund_code = values.fund_code;
  if (values.description) body.description = values.description;
  if (values.strategy_code) body.strategy_code = values.strategy_code;
  if (values.benchmark) body.benchmark = values.benchmark;
  if (values.risk_profile) body.risk_profile = values.risk_profile;
  return body;
}

export function emptyPortfolioCreateFormValues(): PortfolioCreateFormValues {
  return {
    bindFund: false,
    fund_code: "",
    portfolio_type: "",
    code: "",
    name: "",
    description: "",
    base_currency: "",
    valuation_currency: "",
    strategy_code: "",
    benchmark: "",
    risk_profile: "",
    inception_date: "",
  };
}
