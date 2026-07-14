/**
 * Static mirror of the backend SPI rule registry — Phase 1.
 *
 * This is NOT a substitute for a real `GET /compliance/rule-types` endpoint
 * (Missing API #7). It exists so the Pre-trade panel can render a meaningful
 * human label and suggested correction for the rule_type_ids that today's
 * backend can actually return. The strings here are derived from each
 * self-registered rule package under `backend/internal/compliance/rules/`.
 *
 * Only stable technical metadata (rule_type_id, category, selectable) lives
 * in this file. Translated presentation copy (label/explanation/suggested
 * correction) lives in `app/shared/i18n/messages/{en,th,zh}/compliance.ts`
 * under `compliance.catalog.*` and is resolved through the caller's `t()`.
 *
 * If a backend response includes an unknown rule_type_id, callers fall back
 * to the raw id (and the backend's own message) rather than inventing copy.
 */
import type { AppTranslationKey } from "~/shared/i18n/messages";
import type { useI18n } from "~/composables/useI18n";

import type { ComplianceRuleCategory } from "../types";

/** The `t` function shape returned by `useI18n()`. */
type Translate = ReturnType<typeof useI18n>["t"];

export interface RuleCatalogEntry {
  typeId: string;
  category: ComplianceRuleCategory;
  labelKey: AppTranslationKey;
  explanationKey: AppTranslationKey;
  suggestedCorrectionKey: AppTranslationKey;
  /**
   * False for backend stub rule types (`credit_rating.minimum`,
   * `regulatory.thai_sec`) that always PASS/WARN without real evaluation.
   * Kept displayable for historical breach/audit records, but the rule
   * builder must not offer them as a production-ready choice.
   */
  selectable: boolean;
}

export const RULE_CATALOG: readonly RuleCatalogEntry[] = [
  {
    typeId: "allocation.asset_class_max",
    category: "MANDATE",
    labelKey: "compliance.catalog.allocation.asset_class_max.label",
    explanationKey: "compliance.catalog.allocation.asset_class_max.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.allocation.asset_class_max.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "allocation.asset_class_min",
    category: "MANDATE",
    labelKey: "compliance.catalog.allocation.asset_class_min.label",
    explanationKey: "compliance.catalog.allocation.asset_class_min.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.allocation.asset_class_min.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "amount.minimum_trade",
    category: "MANDATE",
    labelKey: "compliance.catalog.amount.minimum_trade.label",
    explanationKey: "compliance.catalog.amount.minimum_trade.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.amount.minimum_trade.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "cash.availability",
    category: "MANDATE",
    labelKey: "compliance.catalog.cash.availability.label",
    explanationKey: "compliance.catalog.cash.availability.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.cash.availability.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "concentration.single_issuer",
    category: "MANDATE",
    labelKey: "compliance.catalog.concentration.single_issuer.label",
    explanationKey:
      "compliance.catalog.concentration.single_issuer.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.concentration.single_issuer.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "credit.min_rating",
    category: "MANDATE",
    labelKey: "compliance.catalog.credit.min_rating.label",
    explanationKey: "compliance.catalog.credit.min_rating.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.credit.min_rating.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "credit_rating.minimum",
    category: "MANDATE",
    labelKey: "compliance.catalog.credit_rating.minimum.label",
    explanationKey: "compliance.catalog.credit_rating.minimum.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.credit_rating.minimum.suggestedCorrection",
    selectable: false,
  },
  {
    typeId: "exposure.max_order_percent_aum",
    category: "MANDATE",
    labelKey: "compliance.catalog.exposure.max_order_percent_aum.label",
    explanationKey:
      "compliance.catalog.exposure.max_order_percent_aum.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.exposure.max_order_percent_aum.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "quantity.min_trading_unit",
    category: "MANDATE",
    labelKey: "compliance.catalog.quantity.min_trading_unit.label",
    explanationKey:
      "compliance.catalog.quantity.min_trading_unit.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.quantity.min_trading_unit.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "quantity.sell_available",
    category: "MANDATE",
    labelKey: "compliance.catalog.quantity.sell_available.label",
    explanationKey: "compliance.catalog.quantity.sell_available.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.quantity.sell_available.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "ratio.sector_exposure",
    category: "MANDATE",
    labelKey: "compliance.catalog.ratio.sector_exposure.label",
    explanationKey: "compliance.catalog.ratio.sector_exposure.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.ratio.sector_exposure.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "regulatory.thai_sec",
    category: "REGULATORY",
    labelKey: "compliance.catalog.regulatory.thai_sec.label",
    explanationKey: "compliance.catalog.regulatory.thai_sec.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.regulatory.thai_sec.suggestedCorrection",
    selectable: false,
  },
  {
    typeId: "restriction.blacklist",
    category: "RESTRICTION",
    labelKey: "compliance.catalog.restriction.blacklist.label",
    explanationKey: "compliance.catalog.restriction.blacklist.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.restriction.blacklist.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "restriction.whitelist",
    category: "RESTRICTION",
    labelKey: "compliance.catalog.restriction.whitelist.label",
    explanationKey: "compliance.catalog.restriction.whitelist.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.restriction.whitelist.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "restriction.list_enforcement",
    category: "RESTRICTION",
    labelKey: "compliance.catalog.restriction.list_enforcement.label",
    explanationKey:
      "compliance.catalog.restriction.list_enforcement.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.restriction.list_enforcement.suggestedCorrection",
    selectable: true,
  },
  {
    typeId: "valuation.min_nav",
    category: "MANDATE",
    labelKey: "compliance.catalog.valuation.min_nav.label",
    explanationKey: "compliance.catalog.valuation.min_nav.explanation",
    suggestedCorrectionKey:
      "compliance.catalog.valuation.min_nav.suggestedCorrection",
    selectable: true,
  },
] as const;

const CATALOG_INDEX: Record<string, RuleCatalogEntry> = Object.fromEntries(
  RULE_CATALOG.map((entry) => [entry.typeId, entry]),
);

export function lookupRuleCatalog(typeId: string): RuleCatalogEntry | null {
  return CATALOG_INDEX[typeId] ?? null;
}

export function ruleLabel(typeId: string, t: Translate): string {
  const entry = lookupRuleCatalog(typeId);
  return entry ? t(entry.labelKey) : typeId;
}

export function ruleExplanation(
  typeId: string,
  fallback: string,
  t: Translate,
): string {
  const entry = lookupRuleCatalog(typeId);
  return entry ? t(entry.explanationKey) : fallback;
}

export function ruleSuggestedCorrection(
  typeId: string,
  t: Translate,
): string | null {
  const entry = lookupRuleCatalog(typeId);
  return entry ? t(entry.suggestedCorrectionKey) : null;
}
