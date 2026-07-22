<script setup lang="ts">
/**
 * Portfolio-code-routed Valuations page: latest valuation + history, with an
 * explicit official-vs-indicative distinction (docs/MANAGER/MEMORY.md), and
 * a "Run valuation" action against the real, portfolio-code-routed
 * `POST /portfolios/{portfolioCode}/valuations/run` endpoint. The backend
 * itself rejects MODEL portfolios with a 422 (RunValuationByCode) — this
 * page additionally never offers the button for a MODEL portfolio, so the
 * user never even reaches that rejection.
 */
import { computed, onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import { useI18n } from "~/composables/useI18n";
import { formatMoneyCompact, formatMoneyExact, parseDecimalOrNull } from "~/features/my-funds/lib/format";
import { todayBangkokIso } from "~/features/my-funds/lib/derive";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import PortfolioValuationHistory from "./components/PortfolioValuationHistory.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import { portfolioApi, type ApiValuationV2 } from "./services/portfolioApi";
import { deriveValuationOfficialState } from "./lib/valuationOfficialState";

type HistoryRange = "1M" | "3M" | "6M" | "1Y";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.workspaceTabs.valuations"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true },
);

const ctx = usePortfolioContext(() => props.portfolioCode);

const latest = ref<ApiValuationV2 | null>(null);
const history = ref<ApiValuationV2[]>([]);
const historyRange = ref<HistoryRange>("3M");
const loading = ref(false);
const error = ref<string | null>(null);

const running = ref(false);
const runError = ref<string | null>(null);
const runSuccessVisible = ref(false);

function rangeDaysAgo(days: number): string {
  const d = new Date();
  d.setDate(d.getDate() - days);
  return d.toISOString().slice(0, 10);
}

const rangeDays: Record<HistoryRange, number> = { "1M": 30, "3M": 90, "6M": 180, "1Y": 365 };

async function loadValuations() {
  if (!props.portfolioCode) return;
  loading.value = true;
  error.value = null;
  try {
    const [latestResult, historyResult] = await Promise.all([
      portfolioApi.getLatestValuation(props.portfolioCode).catch(() => null),
      portfolioApi
        .listValuations(props.portfolioCode, {
          from: rangeDaysAgo(rangeDays[historyRange.value]),
          page: 1,
          limit: 200,
        })
        .catch(() => ({ items: [] })),
    ]);
    latest.value = latestResult;
    history.value = historyResult.items ?? [];
  } catch (err) {
    error.value = err instanceof Error ? err.message : t("portfolio.valuationsPage.errorTitle");
  } finally {
    loading.value = false;
  }
}

function handleRangeChange(range: HistoryRange) {
  historyRange.value = range;
  void loadValuations();
}

async function runValuation() {
  if (running.value || ctx.isModel.value) return;
  running.value = true;
  runError.value = null;
  runSuccessVisible.value = false;
  try {
    await portfolioApi.runValuation(props.portfolioCode, {
      business_date: todayBangkokIso(),
    });
    runSuccessVisible.value = true;
    await loadValuations();
  } catch (err) {
    runError.value = err instanceof Error ? err.message : t("portfolio.valuationsPage.runError");
  } finally {
    running.value = false;
  }
}

onMounted(() => {
  void ctx.reload();
  void loadValuations();
});
watch(
  () => props.portfolioCode,
  () => {
    void ctx.reload();
    void loadValuations();
  },
);

const officialState = computed(() =>
  deriveValuationOfficialState(latest.value, ctx.portfolioType.value),
);

const currency = computed(
  () => latest.value?.valuation_ccy || ctx.portfolio.value?.valuation_currency || ctx.portfolio.value?.base_currency || "",
);

function money(value: string | undefined, compact = false): string {
  const parsed = parseDecimalOrNull(value);
  if (parsed === null) return t("portfolio.terminal.unavailable");
  return compact ? formatMoneyCompact(parsed, currency.value) : formatMoneyExact(parsed, currency.value);
}
</script>

<template>
  <section class="portfolio-valuations">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <AppCard
      :title="t('portfolio.valuationsPage.latestTitle')"
      :subtitle="t('portfolio.valuationsPage.subtitle')"
    >
      <div v-if="loading && !latest" class="portfolio-valuations__notice" role="status">
        {{ t("portfolio.valuationsPage.loading") }}
      </div>
      <div v-else-if="error" class="portfolio-valuations__error" role="alert">
        <div>{{ t("portfolio.valuationsPage.errorTitle") }}: {{ error }}</div>
        <button type="button" class="portfolio-valuations__retry" @click="loadValuations">
          {{ t("portfolio.valuationsPage.retry") }}
        </button>
      </div>

      <template v-else>
        <div
          class="portfolio-valuations__state-banner"
          :data-state="officialState"
          role="status"
        >
          <template v-if="officialState === 'OFFICIAL'">
            {{ t("portfolio.valuationsPage.state.official") }}
          </template>
          <template v-else-if="officialState === 'INDICATIVE'">
            {{ t("portfolio.valuationsPage.state.indicative") }}
          </template>
          <template v-else-if="officialState === 'STALE_INDICATIVE'">
            {{ t("portfolio.valuationsPage.state.staleIndicative") }}
          </template>
          <template v-else-if="ctx.isModel.value">
            {{ t("portfolio.valuationsPage.state.modelUnavailable") }}
          </template>
          <template v-else>
            {{ t("portfolio.valuationsPage.state.unavailable") }}
          </template>
        </div>

        <div v-if="latest" class="portfolio-valuations__grid">
          <div class="portfolio-valuations__metric">
            <span>{{ t("portfolio.valuationsPage.aum") }}</span>
            <strong>{{ money(latest.aum) }}</strong>
          </div>
          <div class="portfolio-valuations__metric">
            <span>{{ t("portfolio.valuationsPage.cashBalance") }}</span>
            <strong>{{ money(latest.cash_balance) }}</strong>
          </div>
          <div class="portfolio-valuations__metric">
            <span>{{ t("portfolio.valuationsPage.marketValue") }}</span>
            <strong>{{ money(latest.market_value) }}</strong>
          </div>
          <div class="portfolio-valuations__metric">
            <span>{{ t("portfolio.valuationsPage.unrealisedPnl") }}</span>
            <strong>{{ money(latest.unrealised_pnl) }}</strong>
          </div>
          <div class="portfolio-valuations__metric">
            <span>{{ t("portfolio.valuationsPage.realisedPnl") }}</span>
            <strong>{{ money(latest.realised_pnl) }}</strong>
          </div>
          <div class="portfolio-valuations__metric">
            <span>{{ t("portfolio.valuationsPage.businessDate") }}</span>
            <strong>{{ latest.business_date || t("portfolio.terminal.unavailable") }}</strong>
          </div>
        </div>
        <div v-else class="portfolio-valuations__notice">
          {{ t("portfolio.valuationsPage.empty") }}
        </div>

        <p v-if="runError" class="portfolio-valuations__error" role="alert">{{ runError }}</p>
        <p v-if="runSuccessVisible" class="portfolio-valuations__success" role="status">
          {{ t("portfolio.valuationsPage.runSuccess") }}
        </p>

        <div v-if="ctx.isModel.value" class="portfolio-valuations__run-gap">
          {{ t("portfolio.valuationsPage.runUnavailableModel") }}
        </div>
        <IMSPermissionGuard v-else permission="INVESTMENT_VALUATION_RUN" mode="disable">
          <AppButton variant="secondary" size="sm" :loading="running" :disabled="running" @click="runValuation">
            {{ running ? t("portfolio.valuationsPage.running") : t("portfolio.valuationsPage.runValuation") }}
          </AppButton>
        </IMSPermissionGuard>
      </template>
    </AppCard>

    <AppCard
      :title="t('portfolio.terminal.history.title')"
      :subtitle="t('portfolio.terminal.history.subtitle', { currency })"
    >
      <PortfolioValuationHistory
        :snapshots="history"
        :currency="currency"
        :loading="loading"
        :range="historyRange"
        @range-change="handleRangeChange"
      />
    </AppCard>
  </section>
</template>

<style scoped>
.portfolio-valuations {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-valuations__notice,
.portfolio-valuations__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-valuations__error {
  color: var(--alert-danger-text, #cf222e);
  display: grid;
  gap: 6px;
}

.portfolio-valuations__success {
  padding: 10px 12px;
  font-size: 13px;
  color: var(--state-success, #1a7f37);
}

.portfolio-valuations__retry {
  justify-self: start;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card, #ffffff);
  border-radius: var(--radius-md, 5px);
  padding: 4px 10px;
  cursor: pointer;
}

.portfolio-valuations__state-banner {
  padding: 8px 12px;
  margin-bottom: 14px;
  font-size: 12px;
  font-weight: 600;
  border-radius: var(--radius-md, 6px);
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--text-secondary, #57606a);
}

.portfolio-valuations__state-banner[data-state="OFFICIAL"] {
  background: var(--alert-success-bg, #dafbe1);
  border-color: var(--alert-success-border, #1a7f37);
  color: var(--alert-success-text, #1a7f37);
}

.portfolio-valuations__state-banner[data-state="INDICATIVE"],
.portfolio-valuations__state-banner[data-state="STALE_INDICATIVE"] {
  background: var(--alert-warning-bg, #fff8c5);
  border-color: var(--alert-warning-border, #9a6700);
  color: var(--alert-warning-text, #9a6700);
}

.portfolio-valuations__state-banner[data-state="UNAVAILABLE"] {
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--text-tertiary, #6e7781);
}

.portfolio-valuations__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
  margin-bottom: 14px;
}

.portfolio-valuations__metric {
  display: grid;
  gap: 4px;
  padding: 10px 12px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card, #ffffff);
}

.portfolio-valuations__metric span {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
}

.portfolio-valuations__metric strong {
  font-size: 14px;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary, #1f2328);
}

.portfolio-valuations__run-gap {
  margin-top: 14px;
  padding: 10px 12px;
  font-size: 12px;
  color: var(--text-tertiary, #6e7781);
  border: 1px dashed var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
}
</style>
