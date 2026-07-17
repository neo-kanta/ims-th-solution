/**
 * Sample parameter scaffolds per backend rule package, verified against each
 * rule's `spi.ParameterSchema()` in `backend/internal/compliance/rules/**`.
 *
 * Used by the Rule Builder to prefill the parameters JSON editor when a rule
 * type is picked — the user can edit freely; the backend validates against
 * the real schema on submit. Kept in its own module (rather than inline in
 * the builder component) so `tests/compliance-rule-parameter-samples.test.ts`
 * can assert the sample keys against the backend schema without a Vue
 * component-rendering harness.
 *
 * Stub rule types (`credit_rating.minimum`, `regulatory.thai_sec`) have no
 * entry here — they are not selectable in the builder and must not look
 * production-ready.
 */
export const RULE_PARAMETER_SAMPLES: Record<string, string> = {
  "allocation.asset_class_max": JSON.stringify(
    { asset_class: "EQUITY", max_percent_nav: 60 },
    null,
    2,
  ),
  "allocation.asset_class_min": JSON.stringify(
    { asset_class: "FIXED_INCOME", min_percent_nav: 20 },
    null,
    2,
  ),
  "amount.minimum_trade": JSON.stringify(
    { min_amount: 100000, currency: "THB" },
    null,
    2,
  ),
  "cash.availability": JSON.stringify({ min_cash_buffer_pct: 0 }, null, 2),
  "concentration.single_issuer": JSON.stringify(
    { max_pct: 10, exempt_government: false },
    null,
    2,
  ),
  "credit.min_rating": JSON.stringify({ min_rating: "BBB-" }, null, 2),
  "exposure.max_order_percent_aum": JSON.stringify(
    { max_percent_aum: 10 },
    null,
    2,
  ),
  "quantity.min_trading_unit": JSON.stringify(
    { default_lot_size: 100 },
    null,
    2,
  ),
  "quantity.sell_available": JSON.stringify({}, null, 2),
  "ratio.sector_exposure": JSON.stringify(
    { sector: "FINANCIALS", max_pct: 30 },
    null,
    2,
  ),
  "restriction.blacklist": JSON.stringify({}, null, 2),
  "restriction.whitelist": JSON.stringify({}, null, 2),
  "restriction.list_enforcement": JSON.stringify(
    { enforced_types: ["BLACKLIST", "WHITELIST"] },
    null,
    2,
  ),
  "valuation.min_nav": JSON.stringify({ min_nav: 10000000 }, null, 2),
};
