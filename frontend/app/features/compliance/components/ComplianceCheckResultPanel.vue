<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppIcon from "~/shared/ui/AppIcon.vue";

import { verdictLabel, verdictTone } from "../lib/formatters";
import {
  ruleExplanation,
  ruleLabel,
  ruleSuggestedCorrection,
} from "../lib/ruleTypeCatalog";
import type {
  ComplianceBreach,
  ComplianceCheckBreachSummary,
  CompliancePreTradeResponse,
} from "../types";
import ComplianceSeverityBadge from "./ComplianceSeverityBadge.vue";
import ComplianceVerdictBadge from "./ComplianceVerdictBadge.vue";
import ComplianceNoRulesEmptyState from "./ComplianceNoRulesEmptyState.vue";

interface Props {
  result: CompliancePreTradeResponse;
  loading?: boolean;
  /**
   * Persisted breach records keyed by breach_id, fetched via
   * `GET /compliance/checks/{groupID}` after the pre-trade run. Provides
   * the structured `evidence.threshold_breached` data the BreachSummary
   * does not include.
   */
  hydratedBreaches?: Map<string, ComplianceBreach>;
  /**
   * Error message from the hydration call, if any. When set, the panel
   * keeps showing "—" for evidence values and notes that the lookup failed
   * rather than silently rendering blanks.
   */
  hydrationError?: string | null;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  hydratedBreaches: () => new Map(),
  hydrationError: null,
});

const emit = defineEmits<{
  recheck: [];
  "view-rule": [ruleTypeId: string];
}>();

const { t } = useI18n();

const breaches = computed<ComplianceCheckBreachSummary[]>(() =>
  Array.isArray(props.result.breaches) ? props.result.breaches : [],
);

const blockingCount = computed(
  () => breaches.value.filter((b) => b.verdict === "BLOCK").length,
);

const warningCount = computed(
  () => breaches.value.filter((b) => b.verdict === "WARN").length,
);

const evaluatedZero = computed(
  () => (props.result.rules_evaluated ?? 0) === 0,
);

const headlineTone = computed(() => verdictTone(props.result.verdict));

const headlineCopy = computed(() => {
  if (props.result.verdict === "BLOCK") {
    return {
      headline: t("compliance.preTrade.result.block"),
      summary:
        blockingCount.value === 1
          ? t("compliance.preTrade.result.blockSummaryOne")
          : t("compliance.preTrade.result.blockSummaryMany", {
              count: blockingCount.value || breaches.value.length,
            }),
    };
  }
  if (props.result.verdict === "WARN") {
    return {
      headline: t("compliance.preTrade.result.warn"),
      summary:
        warningCount.value === 1
          ? t("compliance.preTrade.result.warnSummaryOne")
          : t("compliance.preTrade.result.warnSummaryMany", {
              count: warningCount.value || breaches.value.length,
            }),
    };
  }
  return {
    headline: t("compliance.preTrade.result.pass"),
    summary: t("compliance.preTrade.result.passSummary", {
      count: props.result.rules_evaluated,
    }),
  };
});

interface NumericEvidence {
  current: string | null;
  afterOrder: string | null;
  limit: string | null;
  difference: string | null;
  unit: string | null;
}

/** Best-effort number parse for diff math. Returns null on any failure. */
function parseDecimal(value: string | null | undefined): number | null {
  if (value === null || value === undefined || value === "") return null;
  const n = Number.parseFloat(String(value));
  return Number.isFinite(n) ? n : null;
}

function formatNumber(n: number, unit: string | null): string {
  // 2dp percents, 4dp other decimals; whole numbers stay whole.
  const fixed = unit === "percent" ? n.toFixed(2) : n.toFixed(4);
  // Strip trailing zeros for whole numbers / .5 etc., keep at most 4dp.
  return fixed.replace(/\.?0+$/, "") || "0";
}

function formatWithUnit(value: string | null, unit: string | null): string | null {
  if (value === null) return null;
  if (unit === "percent") return `${value}%`;
  return value;
}

/**
 * Read the persisted breach evidence and turn it into a 4-cell display row.
 * If a persisted breach is not available (or did not include a
 * `threshold_breached` payload) we return all-null so the UI renders "—"
 * rather than invent numbers.
 *
 * Field mapping comes directly from the backend rule packages:
 *   evidence.threshold_breached.{actual,limit,operator,unit}
 *   evidence.metrics.{nav, entity_market_value, available_cash, ...}
 *
 * We treat `threshold_breached.actual` as the post-order ("after order")
 * value because all current rules evaluate the proposed state. The
 * "current" cell is filled from a metric whose key starts with `current_`
 * when present; otherwise it stays "—" because the backend did not
 * compute it.
 */
function readEvidenceForBreach(
  breach: ComplianceCheckBreachSummary,
): NumericEvidence {
  const persisted = props.hydratedBreaches.get(breach.breach_id);
  const evidence = (persisted?.evidence ?? null) as
    | {
        metrics?: Record<string, string>;
        threshold_breached?: {
          actual?: string;
          limit?: string;
          operator?: string;
          unit?: string;
        };
      }
    | null;

  if (!evidence) {
    return { current: null, afterOrder: null, limit: null, difference: null, unit: null };
  }

  const tb = evidence.threshold_breached ?? null;
  const metrics = evidence.metrics ?? {};
  const unit = tb?.unit ?? null;

  // Honour what the backend actually returned. "After order" is the
  // proposed-state actual; "Limit" is the configured threshold.
  const afterOrder = tb?.actual ?? null;
  const limit = tb?.limit ?? null;

  // Pick up a "current" value if any metric is keyed `current_*`. We never
  // synthesize one — if the backend did not record a current value we
  // leave the cell blank.
  let current: string | null = null;
  for (const [key, value] of Object.entries(metrics)) {
    if (key.startsWith("current_") || key === "current") {
      current = value;
      break;
    }
  }

  // Difference: if both after-order and limit are numeric, compute |actual − limit|.
  // Otherwise leave blank.
  const a = parseDecimal(afterOrder);
  const l = parseDecimal(limit);
  const diffNum = a !== null && l !== null ? a - l : null;
  const difference =
    diffNum === null
      ? null
      : `${diffNum >= 0 ? "+" : "−"}${formatNumber(Math.abs(diffNum), unit)}`;

  return {
    current: formatWithUnit(current, unit),
    afterOrder: formatWithUnit(afterOrder, unit),
    limit: formatWithUnit(limit, unit),
    difference,
    unit,
  };
}

function bannerClass(): string {
  const map: Record<string, string> = {
    success: "result-banner result-banner--pass",
    warning: "result-banner result-banner--warn",
    error: "result-banner result-banner--block",
  };
  return map[headlineTone.value] || "result-banner";
}

/** Memoise the evidence lookup so each breach row reads it once. */
const evidenceByBreachId = computed(() => {
  const out = new Map<string, NumericEvidence>();
  for (const b of breaches.value) {
    out.set(b.breach_id, readEvidenceForBreach(b));
  }
  return out;
});

function evidenceFor(id: string): NumericEvidence {
  return (
    evidenceByBreachId.value.get(id) ?? {
      current: null,
      afterOrder: null,
      limit: null,
      difference: null,
      unit: null,
    }
  );
}

function hasAnyEvidence(id: string): boolean {
  const e = evidenceFor(id);
  return Boolean(e.afterOrder || e.limit || e.current);
}
</script>

<template>
  <section class="check-result" :aria-busy="loading" aria-live="polite">
    <ComplianceNoRulesEmptyState
      v-if="evaluatedZero"
      post-check
      :show-rules-link="true"
    />

    <template v-else>
      <header :class="bannerClass()" role="status">
        <div class="result-banner__icon" aria-hidden="true">
          <AppIcon
            :name="result.verdict === 'PASS' ? 'check' : 'warning'"
            size="lg"
          />
        </div>
        <div class="result-banner__copy">
          <div class="result-banner__eyebrow">
            <span class="result-banner__verdict-tag">
              {{ verdictLabel(result.verdict) }}
            </span>
            <span class="result-banner__meta">
              {{ t("compliance.preTrade.result.rulesEvaluated") }}:
              <strong>{{ result.rules_evaluated }}</strong>
              ·
              {{
                t("compliance.preTrade.result.durationLabel", {
                  ms: result.total_duration_ms ?? 0,
                })
              }}
            </span>
          </div>
          <h2 class="result-banner__headline">
            {{ headlineCopy.headline }}
          </h2>
          <p class="result-banner__summary">{{ headlineCopy.summary }}</p>
          <p v-if="result.check_group_id" class="result-banner__group">
            {{ t("compliance.preTrade.result.groupIdLabel") }}:
            <code>{{ result.check_group_id }}</code>
          </p>
        </div>
        <div class="result-banner__actions">
          <AppButton
            variant="ghost"
            size="sm"
            :loading="loading"
            @click="emit('recheck')"
          >
            <AppIcon name="refresh" size="xs" />
            {{ t("compliance.preTrade.result.recheck") }}
          </AppButton>
        </div>
      </header>

      <ul v-if="breaches.length > 0" class="result-breaches">
        <li
          v-for="breach in breaches"
          :key="breach.breach_id"
          :class="[
            'breach-card',
            breach.verdict === 'BLOCK'
              ? 'breach-card--block'
              : breach.verdict === 'WARN'
                ? 'breach-card--warn'
                : 'breach-card--neutral',
          ]"
        >
          <header class="breach-card__head">
            <div class="breach-card__title">
              {{ ruleLabel(breach.rule_type_id, t) }}
            </div>
            <div class="breach-card__badges">
              <ComplianceVerdictBadge :verdict="breach.verdict" />
              <ComplianceSeverityBadge :severity="breach.severity" />
            </div>
          </header>

          <p class="breach-card__explanation">
            <strong>{{ t("compliance.preTrade.result.explanationLabel") }}:</strong>
            {{ ruleExplanation(breach.rule_type_id, breach.message, t) }}
          </p>

          <div
            class="breach-card__grid"
            role="group"
            :aria-label="t('compliance.preTrade.result.currentValue')"
          >
            <div class="breach-card__cell">
              <span class="breach-card__cell-label">
                {{ t("compliance.preTrade.result.currentValue") }}
              </span>
              <span class="breach-card__cell-value">
                {{ evidenceFor(breach.breach_id).current ?? "—" }}
              </span>
            </div>
            <div class="breach-card__cell">
              <span class="breach-card__cell-label">
                {{ t("compliance.preTrade.result.afterOrder") }}
              </span>
              <span class="breach-card__cell-value">
                {{ evidenceFor(breach.breach_id).afterOrder ?? "—" }}
              </span>
            </div>
            <div class="breach-card__cell">
              <span class="breach-card__cell-label">
                {{ t("compliance.preTrade.result.ruleLimit") }}
              </span>
              <span class="breach-card__cell-value">
                {{ evidenceFor(breach.breach_id).limit ?? "—" }}
              </span>
            </div>
            <div class="breach-card__cell">
              <span class="breach-card__cell-label">
                {{ t("compliance.preTrade.result.difference") }}
              </span>
              <span class="breach-card__cell-value">
                {{ evidenceFor(breach.breach_id).difference ?? "—" }}
              </span>
            </div>
          </div>
          <p
            v-if="!hasAnyEvidence(breach.breach_id)"
            class="breach-card__data-note"
          >
            <template v-if="hydrationError">
              {{ t("compliance.preTrade.result.evidenceLookupFailed", { error: hydrationError }) }}
            </template>
            <template v-else>
              {{ t("compliance.preTrade.result.dataNotProvided") }}
            </template>
          </p>

          <p
            v-if="ruleSuggestedCorrection(breach.rule_type_id, t)"
            class="breach-card__correction"
          >
            <strong>
              {{ t("compliance.preTrade.result.suggestedCorrection") }}:
            </strong>
            {{ ruleSuggestedCorrection(breach.rule_type_id, t) }}
          </p>

          <footer class="breach-card__actions">
            <AppButton
              variant="primary"
              size="sm"
              :disabled="true"
              :title="t('compliance.preTrade.result.requestExceptionDisabled')"
            >
              {{ t("compliance.preTrade.result.requestException") }}
            </AppButton>
            <AppButton
              variant="ghost"
              size="sm"
              @click="emit('view-rule', breach.rule_type_id)"
            >
              {{ t("compliance.preTrade.result.viewRule") }}
            </AppButton>
            <span class="breach-card__id" :title="breach.breach_id">
              Breach <code>{{ breach.breach_id }}</code>
            </span>
          </footer>
        </li>
      </ul>
    </template>
  </section>
</template>

<style scoped>
.check-result {
  display: grid;
  gap: var(--space-6);
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.result-banner {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: var(--space-6);
  align-items: center;
  padding: var(--space-6) var(--space-7);
  border-radius: var(--radius-lg);
  border-width: 1px;
  border-style: solid;
  min-height: 140px;
  box-shadow: var(--shadow-md);
  position: relative;
  overflow: hidden;
  transition: all var(--transition-base);
}

.result-banner::after {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  width: 6px;
  bottom: 0;
}

/* Light / Dark themed gradients for PASS */
.result-banner--pass {
  background: linear-gradient(135deg, var(--color-success-50) 0%, rgba(18, 183, 106, 0.05) 100%);
  border-color: var(--color-success-200);
  color: var(--color-success-900);
}
.result-banner--pass::after {
  background: var(--state-success);
}
:root[data-theme="dark"] .result-banner--pass {
  background: linear-gradient(135deg, rgba(63, 185, 80, 0.1) 0%, rgba(63, 185, 80, 0.02) 100%);
  border-color: rgba(63, 185, 80, 0.3);
  color: #7ee787;
}

/* Light / Dark themed gradients for WARN */
.result-banner--warn {
  background: linear-gradient(135deg, var(--color-warning-50) 0%, rgba(245, 158, 11, 0.05) 100%);
  border-color: var(--color-warning-200);
  color: var(--color-warning-900);
}
.result-banner--warn::after {
  background: var(--state-warning);
}
:root[data-theme="dark"] .result-banner--warn {
  background: linear-gradient(135deg, rgba(210, 153, 34, 0.1) 0%, rgba(210, 153, 34, 0.02) 100%);
  border-color: rgba(210, 153, 34, 0.3);
  color: #f2cc60;
}

/* Light / Dark themed gradients for BLOCK */
.result-banner--block {
  background: linear-gradient(135deg, var(--color-danger-50) 0%, rgba(240, 68, 88, 0.05) 100%);
  border-color: var(--color-danger-200);
  color: var(--color-danger-900);
}
.result-banner--block::after {
  background: var(--state-danger);
}
:root[data-theme="dark"] .result-banner--block {
  background: linear-gradient(135deg, rgba(248, 81, 73, 0.1) 0%, rgba(248, 81, 73, 0.02) 100%);
  border-color: rgba(248, 81, 73, 0.3);
  color: #ffaba8;
}

.result-banner__icon {
  width: 56px;
  height: 56px;
  display: grid;
  place-items: center;
  border-radius: var(--radius-md);
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-subtle);
  transition: transform var(--transition-base);
}

.result-banner:hover .result-banner__icon {
  transform: scale(1.05);
}

.result-banner--block .result-banner__icon {
  color: var(--state-danger);
}

.result-banner--warn .result-banner__icon {
  color: var(--state-warning);
}

.result-banner--pass .result-banner__icon {
  color: var(--state-success);
}

.result-banner__copy {
  display: grid;
  gap: var(--space-1);
}

.result-banner__eyebrow {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-size-2xs);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
}

.result-banner__verdict-tag {
  display: inline-flex;
  align-items: center;
  height: 20px;
  padding: 0 var(--space-2);
  border-radius: var(--radius-pill);
  font-weight: var(--font-weight-bold);
  font-size: 10px;
  letter-spacing: 0.06em;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
}

.result-banner--block .result-banner__verdict-tag {
  color: var(--state-danger);
}

.result-banner--warn .result-banner__verdict-tag {
  color: var(--state-warning);
}

.result-banner--pass .result-banner__verdict-tag {
  color: var(--state-success);
}

.result-banner__headline {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  letter-spacing: -0.03em;
  color: var(--text-primary);
  line-height: var(--line-height-tight);
}

.result-banner__summary {
  margin: 0;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.result-banner__group {
  margin: 0;
  font-size: var(--font-size-2xs);
  color: var(--text-tertiary);
  margin-top: var(--space-1);
}

.result-banner__group code {
  font-family: var(--font-family-mono);
  background: rgba(0, 0, 0, 0.04);
  padding: 2px 4px;
  border-radius: var(--radius-xs);
}

:root[data-theme="dark"] .result-banner__group code {
  background: rgba(255, 255, 255, 0.06);
}

.result-banner__actions {
  align-self: center;
}

.result-breaches {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-5);
}

.breach-card {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5) var(--space-6);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  box-shadow: var(--shadow-sm);
  position: relative;
  transition: all var(--transition-base);
}

.breach-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  border-color: var(--border-strong);
}

.breach-card--block {
  border-left: 4px solid var(--state-danger);
}

.breach-card--warn {
  border-left: 4px solid var(--state-warning);
}

.breach-card--neutral {
  border-left: 4px solid var(--border-default);
}

.breach-card__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-4);
  flex-wrap: wrap;
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-3);
}

.breach-card__title {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
}

.breach-card__badges {
  display: inline-flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.breach-card__explanation {
  margin: 0;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  line-height: var(--line-height-normal);
}

.breach-card__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
  gap: var(--space-3);
  margin-top: var(--space-1);
}

.breach-card__cell {
  display: grid;
  gap: var(--space-1);
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
  transition: all var(--transition-fast);
}

.breach-card__cell:hover {
  background: var(--bg-card-hover);
  border-color: var(--border-strong);
}

.breach-card__cell-label {
  font-size: 10px;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: var(--font-weight-medium);
}

.breach-card__cell-value {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}

.breach-card__data-note {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  font-style: italic;
  background: var(--bg-card-muted);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  border: 1px dashed var(--border-subtle);
}

.breach-card__correction {
  margin: 0;
  padding: var(--space-3) var(--space-4);
  border-left: 3px solid var(--state-info);
  background: var(--alert-info-bg);
  color: var(--alert-info-text);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  line-height: var(--line-height-normal);
}

.breach-card__actions {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex-wrap: wrap;
  border-top: 1px solid var(--border-subtle);
  padding-top: var(--space-3);
  margin-top: var(--space-1);
}

.breach-card__id {
  margin-left: auto;
  color: var(--text-tertiary);
  font-size: 10px;
}

.breach-card__id code {
  font-family: var(--font-family-mono);
  background: var(--bg-card-muted);
  padding: 1px 3px;
  border-radius: var(--radius-xs);
}

@media (max-width: 720px) {
  .result-banner {
    grid-template-columns: auto 1fr;
  }
  .result-banner__actions {
    grid-column: 1 / -1;
    justify-self: flex-start;
  }
}
</style>
