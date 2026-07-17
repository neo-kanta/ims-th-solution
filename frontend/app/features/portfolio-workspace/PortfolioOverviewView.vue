<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import { useI18n } from "~/composables/useI18n";
import {
  currencySymbol,
  formatMoneyCompact,
  formatPercent,
  parseDecimal,
} from "~/features/my-funds/lib/format";
import {
  myFundsApi,
  type ApiAssetClass,
  type ApiInstrument,
  type FundAllocation,
  type FundMarketDataStatus,
  type FundNavHistory,
  type FundNavSnapshot,
  type IntradayFundValuation,
} from "~/features/my-funds/services/myFundsApi";
import PortfolioAllocationDonut from "./components/PortfolioAllocationDonut.vue";
import PortfolioValuationHistory from "./components/PortfolioValuationHistory.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import type { AllocationDonutItem } from "./lib/allocationDonut";
import {
  portfolioApi,
  type ApiCashBalanceV2,
  type ApiHoldingV2,
  type ApiValuationV2,
} from "./services/portfolioApi";

const props = defineProps<{ portfolioCode: string }>();
const { t, locale } = useI18n();
const ctx = usePortfolioContext(() => props.portfolioCode);

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.workspaceTabs.overview"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true },
);

const latestValuation = ref<ApiValuationV2 | null>(null);
const valuationHistory = ref<ApiValuationV2[]>([]);
const latestFundNav = ref<FundNavSnapshot | null>(null);
const intradayFundValuation = ref<IntradayFundValuation | null>(null);
const fundMarketDataStatus = ref<FundMarketDataStatus | null>(null);
const fundAllocation = ref<FundAllocation | null>(null);
const fundNavHistory = ref<FundNavHistory | null>(null);
const cashBalances = ref<ApiCashBalanceV2[]>([]);
const holdings = ref<ApiHoldingV2[]>([]);
const instruments = ref<ApiInstrument[]>([]);
const assetClasses = ref<ApiAssetClass[]>([]);
const loadingData = ref(false);
const loadingHistory = ref(false);
const dataError = ref<string | null>(null);
type AllocationMode = "assetClass" | "sector" | "country" | "currency";
type HistoryRange = "1M" | "3M" | "6M" | "1Y";
const allocationMode = ref<AllocationMode>("assetClass");
const historyRange = ref<HistoryRange>("3M");

function oneYearAgo(): string {
  const date = new Date();
  date.setUTCFullYear(date.getUTCFullYear() - 1);
  return date.toISOString().slice(0, 10);
}

async function loadWorkspaceData() {
  if (!props.portfolioCode) return;
  loadingData.value = true;
  dataError.value = null;

  try {
    await ctx.reload();
    if (!ctx.portfolio.value) return;
    const fundId = ctx.portfolio.value.fund_id;

    const [
      valuation,
      cash,
      holdingRows,
      instrumentRows,
      classRows,
      history,
      fundNav,
      intraday,
      feedStatus,
      allocation,
      navHistory,
    ] = await Promise.all([
      portfolioApi.getLatestValuation(props.portfolioCode).catch(() => null),
      portfolioApi
        .getCash(props.portfolioCode)
        .catch(() => [] as ApiCashBalanceV2[]),
      portfolioApi
        .getHoldings(props.portfolioCode)
        .catch(() => [] as ApiHoldingV2[]),
      myFundsApi.listInstruments().catch(() => [] as ApiInstrument[]),
      myFundsApi.listAssetClasses().catch(() => [] as ApiAssetClass[]),
      portfolioApi
        .listValuations(props.portfolioCode, {
          from: oneYearAgo(),
          page: 1,
          limit: 200,
        })
        .catch(() => ({ items: [] })),
      fundId
        ? myFundsApi.getLatestFundNav(fundId).catch(() => null)
        : Promise.resolve(null),
      fundId
        ? myFundsApi.getFundIntradayValuation(fundId).catch(() => null)
        : Promise.resolve(null),
      fundId
        ? myFundsApi.getFundMarketDataStatus(fundId).catch(() => null)
        : Promise.resolve(null),
      fundId
        ? myFundsApi.getFundAllocation(fundId).catch(() => null)
        : Promise.resolve(null),
      fundId
        ? myFundsApi
            .getFundNavHistory(fundId, historyRange.value)
            .catch(() => null)
        : Promise.resolve(null),
    ]);

    latestValuation.value = valuation;
    cashBalances.value = cash;
    holdings.value = holdingRows;
    instruments.value = instrumentRows;
    assetClasses.value = classRows;
    valuationHistory.value = history.items ?? [];
    latestFundNav.value = fundNav;
    intradayFundValuation.value = intraday;
    fundMarketDataStatus.value = feedStatus;
    fundAllocation.value = allocation;
    fundNavHistory.value = navHistory;
  } catch (error) {
    dataError.value =
      error instanceof Error
        ? error.message
        : t("portfolio.terminal.loadError");
  } finally {
    loadingData.value = false;
  }
}

async function loadFundHistory(range: HistoryRange) {
  const fundId = ctx.portfolio.value?.fund_id;
  if (!fundId) return;
  loadingHistory.value = true;
  try {
    fundNavHistory.value = await myFundsApi.getFundNavHistory(fundId, range);
  } catch {
    fundNavHistory.value = null;
  } finally {
    loadingHistory.value = false;
  }
}

function handleHistoryRangeChange(range: HistoryRange) {
  historyRange.value = range;
  void loadFundHistory(range);
}

onMounted(() => {
  void loadWorkspaceData();
});
watch(
  () => props.portfolioCode,
  () => {
    void loadWorkspaceData();
  },
);

const valuationCurrency = computed(
  () =>
    latestFundNav.value?.valuation_ccy ||
    intradayFundValuation.value?.valuation_ccy ||
    latestValuation.value?.valuation_ccy ||
    ctx.portfolio.value?.valuation_currency ||
    ctx.portfolio.value?.base_currency ||
    "",
);
const portfolioOfficialAum = computed(() =>
  parseDecimal(latestValuation.value?.aum),
);
const officialAum = computed(() =>
  parseDecimal(
    latestFundNav.value?.aum ||
      intradayFundValuation.value?.official_aum ||
      latestValuation.value?.aum,
  ),
);
const estimatedAum = computed(() =>
  parseDecimal(intradayFundValuation.value?.estimated_aum),
);
const officialUnitNav = computed(
  () =>
    latestFundNav.value?.nav_per_unit ||
    intradayFundValuation.value?.official_nav_per_unit ||
    "",
);
const estimatedUnitNav = computed(
  () => intradayFundValuation.value?.estimated_nav_per_unit || "",
);
const unitsOutstanding = computed(
  () =>
    intradayFundValuation.value?.units_outstanding ||
    latestFundNav.value?.total_units ||
    "",
);
const officialBusinessDate = computed(
  () =>
    latestFundNav.value?.business_date ||
    intradayFundValuation.value?.official_as_of ||
    latestValuation.value?.business_date ||
    "",
);
const marketValue = computed(() =>
  parseDecimal(latestValuation.value?.market_value),
);
const costBasis = computed(() =>
  parseDecimal(latestValuation.value?.cost_basis),
);
const unrealisedPnl = computed(() =>
  parseDecimal(latestValuation.value?.unrealised_pnl),
);
const officialCash = computed(() =>
  parseDecimal(latestValuation.value?.cash_balance),
);
const isStale = computed(() =>
  Boolean(
    intradayFundValuation.value?.is_stale ||
    latestFundNav.value?.has_stale_inputs ||
    latestValuation.value?.has_stale_inputs,
  ),
);
const isIndicative = computed(() =>
  Boolean(
    intradayFundValuation.value?.units_indicative ||
    latestFundNav.value?.is_indicative ||
    latestValuation.value?.is_indicative,
  ),
);

const instrumentById = computed(
  () =>
    new Map(instruments.value.map((instrument) => [instrument.id, instrument])),
);
const assetClassById = computed(
  () =>
    new Map(
      assetClasses.value.map((assetClass) => [assetClass.id, assetClass]),
    ),
);
const holdingByInstrument = computed(
  () =>
    new Map(holdings.value.map((holding) => [holding.instrument_id, holding])),
);
const valuationLineByInstrument = computed(
  () => {
    const entries = (latestValuation.value?.holding_lines ?? []).flatMap((line) =>
      line.instrument_id ? [[line.instrument_id, line] as const] : [],
    );
    return new Map(entries);
  },
);

function instrumentLabel(
  instrumentId: string | undefined,
  index: number,
): string {
  const instrument = instrumentId
    ? instrumentById.value.get(instrumentId)
    : undefined;
  return (
    instrument?.primary_ticker?.trim() ||
    instrument?.name?.trim() ||
    t("portfolio.allocation.position", { index: index + 1 })
  );
}

function instrumentName(instrumentId: string | undefined): string {
  return (
    (instrumentId
      ? instrumentById.value.get(instrumentId)?.name?.trim()
      : "") || t("portfolio.terminal.positions.unnamed")
  );
}

function assetClassLabel(instrumentId: string | undefined): string {
  const instrument = instrumentId
    ? instrumentById.value.get(instrumentId)
    : undefined;
  const assetClass = instrument?.asset_class_id
    ? assetClassById.value.get(instrument.asset_class_id)
    : undefined;
  return (
    assetClass?.name?.trim() ||
    assetClass?.code?.trim() ||
    t("portfolio.terminal.positions.unclassified")
  );
}

const positionRows = computed(() => {
  const ids = new Set<string>();
  holdings.value.forEach((holding) => {
    if (holding.instrument_id) ids.add(holding.instrument_id);
  });
  (latestValuation.value?.holding_lines ?? []).forEach((line) => {
    if (line.instrument_id) ids.add(line.instrument_id);
  });

  return [...ids].map((instrumentId, index) => {
    const holding = holdingByInstrument.value.get(instrumentId);
    const line = valuationLineByInstrument.value.get(instrumentId);
    const instrument = instrumentById.value.get(instrumentId);
    const pnl = line?.unrealised_pnl ? parseDecimal(line.unrealised_pnl) : null;
    return {
      key: instrumentId,
      ticker: instrumentLabel(instrumentId, index),
      name: instrumentName(instrumentId),
      assetClass: assetClassLabel(instrumentId),
      quantity: line?.quantity ?? holding?.quantity,
      averageCost: holding?.average_cost,
      latestPrice: line?.price_in_quote_ccy,
      marketValue: line?.market_value,
      costBasis: line?.cost_basis ?? holding?.cost_basis,
      priceCurrency:
        line?.quote_currency || instrument?.currency || valuationCurrency.value,
      pnl,
      isStale: Boolean(line?.is_stale),
      hasSnapshot: Boolean(line),
    };
  });
});

const historyWithLatest = computed(() => {
  const snapshots = [...valuationHistory.value];
  const latest = latestValuation.value;
  if (
    latest?.business_date &&
    !snapshots.some((item) => item.business_date === latest.business_date)
  ) {
    snapshots.push(latest);
  }
  return snapshots;
});

function fundAllocationItems(mode: AllocationMode): AllocationDonutItem[] {
  const allocation = fundAllocation.value;
  if (!allocation) return [];
  const buckets =
    mode === "assetClass"
      ? allocation.by_asset_class
      : mode === "sector"
        ? allocation.by_sector
        : mode === "country"
          ? allocation.by_country
          : allocation.by_currency;

  return (buckets ?? [])
    .map((bucket, index) => ({
      key: bucket.key || `${mode}-${index + 1}`,
      label: bucket.label || t("portfolio.allocation.other"),
      value: parseDecimal(bucket.market_value),
    }))
    .filter((bucket) => bucket.value > 0)
    .sort((a, b) => b.value - a.value);
}

function groupAllocation(mode: AllocationMode): AllocationDonutItem[] {
  const authoritativeFundItems = fundAllocationItems(mode);
  if (authoritativeFundItems.length > 0) return authoritativeFundItems;
  if (mode === "sector" || mode === "country") return [];

  const lines = latestValuation.value?.holding_lines ?? [];
  if (lines.length === 0) {
    const aggregate: AllocationDonutItem[] = [];
    if (marketValue.value > 0) {
      aggregate.push({
        key: "__invested__",
        label: t("portfolio.allocation.investedAssets"),
        value: marketValue.value,
      });
    }
    if (officialCash.value > 0) {
      aggregate.push({
        key: "__cash__",
        label: t("portfolio.allocation.cash"),
        value: officialCash.value,
      });
    }
    return aggregate;
  }

  const grouped = new Map<string, number>();
  for (const line of lines) {
    const instrument = line.instrument_id
      ? instrumentById.value.get(line.instrument_id)
      : undefined;
    const label =
      mode === "assetClass"
        ? assetClassLabel(line.instrument_id)
        : instrument?.currency?.trim() ||
          line.quote_currency?.trim() ||
          t("portfolio.allocation.other");
    grouped.set(
      label,
      (grouped.get(label) ?? 0) + parseDecimal(line.market_value),
    );
  }

  if (officialCash.value > 0) {
    const cashLabel =
      mode === "assetClass"
        ? t("portfolio.allocation.cash")
        : valuationCurrency.value || t("portfolio.allocation.cash");
    grouped.set(cashLabel, (grouped.get(cashLabel) ?? 0) + officialCash.value);
  }

  return [...grouped.entries()]
    .map(([label, value]) => ({ key: `${mode}-${label}`, label, value }))
    .sort((a, b) => b.value - a.value);
}

const allocationItems = computed(() => groupAllocation(allocationMode.value));
const allocationUsesFundData = computed(
  () => fundAllocationItems(allocationMode.value).length > 0,
);
const allocationTotal = computed(() => {
  if (allocationUsesFundData.value) {
    return parseDecimal(fundAllocation.value?.total_nav);
  }
  return (
    portfolioOfficialAum.value ||
    allocationItems.value.reduce((sum, item) => sum + item.value, 0)
  );
});
const allocationSubtitle = computed(() =>
  allocationUsesFundData.value
    ? t("portfolio.terminal.allocation.fundScope", {
        count: fundAllocation.value?.portfolio_count ?? 0,
      })
    : t("portfolio.terminal.allocation.portfolioScope"),
);
const allocationTabs = computed(
  () =>
    [
      {
        key: "assetClass",
        label: t("portfolio.terminal.allocation.assetClass"),
        disabled: groupAllocation("assetClass").length === 0,
      },
      {
        key: "sector",
        label: t("portfolio.terminal.allocation.sector"),
        disabled: groupAllocation("sector").length === 0,
      },
      {
        key: "country",
        label: t("portfolio.terminal.allocation.country"),
        disabled: groupAllocation("country").length === 0,
      },
      {
        key: "currency",
        label: t("portfolio.terminal.allocation.currency"),
        disabled: groupAllocation("currency").length === 0,
      },
    ] as const,
);

const officialAumScopeHint = computed(() =>
  latestFundNav.value
    ? t("portfolio.terminal.fundAccountingScope", {
        count: latestFundNav.value.portfolio_count ?? 0,
      })
    : t("portfolio.terminal.portfolioAccountingScope"),
);
const liveEstimateHint = computed(() => {
  if (!intradayFundValuation.value)
    return t("portfolio.terminal.noLiveEstimate");
  const change = Number(intradayFundValuation.value.delta_pct_vs_last_close);
  return Number.isFinite(change)
    ? t("portfolio.terminal.liveVsClose", {
        value: formatPercent(change, 2, true),
      })
    : t("portfolio.terminal.linkedFundLiveScope");
});
const fundUnitScopeHint = computed(() =>
  unitsOutstanding.value
    ? t("portfolio.terminal.linkedFundUnitScope")
    : t("portfolio.terminal.notUnitised"),
);
const freshnessTimestamp = computed(
  () =>
    fundMarketDataStatus.value?.last_quote_at ||
    intradayFundValuation.value?.as_of ||
    latestValuation.value?.created_at,
);
const freshnessSource = computed(
  () =>
    fundMarketDataStatus.value?.primary_provider ||
    intradayFundValuation.value?.primary_provider ||
    t("portfolio.terminal.officialSnapshot"),
);
const historyIsUnitNav = computed(() =>
  Boolean(fundNavHistory.value?.has_units),
);
const historyTitle = computed(() =>
  historyIsUnitNav.value
    ? t("portfolio.terminal.history.unitNavTitle")
    : t("portfolio.terminal.history.title"),
);
const historySubtitle = computed(() =>
  historyIsUnitNav.value
    ? t("portfolio.terminal.history.unitNavSubtitle", {
        currency: valuationCurrency.value,
      })
    : t("portfolio.terminal.history.subtitle", {
        currency: valuationCurrency.value,
      }),
);

const freshness = computed(() => {
  if (fundMarketDataStatus.value) {
    const stalePositions = fundMarketDataStatus.value.stale_positions ?? 0;
    if (fundMarketDataStatus.value.healthy && stalePositions === 0) {
      return { label: t("portfolio.terminal.freshness.ok"), state: "ok" };
    }
    return { label: t("portfolio.terminal.freshness.stale"), state: "warning" };
  }
  if (
    !latestValuation.value &&
    !latestFundNav.value &&
    !intradayFundValuation.value
  ) {
    return {
      label: t("portfolio.terminal.freshness.noSnapshot"),
      state: "missing",
    };
  }
  if (isStale.value) {
    return { label: t("portfolio.terminal.freshness.stale"), state: "warning" };
  }
  if (isIndicative.value) {
    return {
      label: t("portfolio.terminal.freshness.indicative"),
      state: "warning",
    };
  }
  return { label: t("portfolio.terminal.freshness.ok"), state: "ok" };
});

function formatExact(
  value: number | string | null | undefined,
  currency: string | null | undefined,
  signed = false,
): string {
  if (value === null || value === undefined || value === "") return "—";
  const numeric = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(numeric)) return "—";
  const sign = signed && numeric > 0 ? "+" : numeric < 0 ? "−" : "";
  return `${sign}${currencySymbol(currency)}${Math.abs(numeric).toLocaleString(
    "en-US",
    {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    },
  )}`;
}

function formatQuantity(value: string | null | undefined): string {
  if (!value) return "—";
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) return "—";
  return numeric.toLocaleString("en-US", { maximumFractionDigits: 4 });
}

function formatUnitNav(value: string | null | undefined): string {
  if (!value) return "—";
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) return "—";
  return `${currencySymbol(valuationCurrency.value)}${numeric.toLocaleString(
    "en-US",
    {
      minimumFractionDigits: 4,
      maximumFractionDigits: 6,
    },
  )}`;
}

function formatUnits(value: string | null | undefined): string {
  if (!value) return "—";
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) return "—";
  return numeric.toLocaleString("en-US", { maximumFractionDigits: 4 });
}

function formatSnapshotTime(value: string | undefined): string {
  if (!value) return t("portfolio.terminal.unavailable");
  const parsed = Date.parse(value);
  if (!Number.isFinite(parsed)) return t("portfolio.terminal.unavailable");
  const localeCode =
    locale.value === "th" ? "th-TH" : locale.value === "zh" ? "zh-CN" : "en-GB";
  return new Intl.DateTimeFormat(localeCode, {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(parsed);
}

function setAllocationMode(key: string) {
  if (
    key === "assetClass" ||
    key === "sector" ||
    key === "country" ||
    key === "currency"
  ) {
    allocationMode.value = key;
  }
}
</script>

<template>
  <section class="portfolio-terminal">
    <div
      v-if="ctx.loading.value && !ctx.portfolio.value"
      class="portfolio-terminal__state"
      role="status"
    >
      <span class="portfolio-terminal__spinner" aria-hidden="true" />
      {{ t("portfolio.overview.loading") }}
    </div>

    <div
      v-else-if="ctx.error.value"
      class="portfolio-terminal__error"
      role="alert"
    >
      <div>
        <strong>{{ t("portfolio.overview.errorTitle") }}</strong>
        <span>{{ ctx.error.value }}</span>
      </div>
      <button type="button" @click="loadWorkspaceData">
        {{ t("portfolio.overview.retry") }}
      </button>
    </div>

    <template v-else-if="ctx.portfolio.value">
      <div v-if="dataError" class="portfolio-terminal__data-error" role="alert">
        {{ dataError }}
      </div>

      <section
        class="valuation-strip"
        :aria-label="t('portfolio.terminal.valuationStripLabel')"
      >
        <article class="valuation-strip__cell valuation-strip__cell--emphasis">
          <span>{{ t("portfolio.terminal.officialAum") }}</span>
          <strong>{{
            officialAum > 0 ? formatExact(officialAum, valuationCurrency) : "—"
          }}</strong>
          <small
            >{{ valuationCurrency || ctx.portfolio.value.base_currency }} ·
            {{ officialAumScopeHint }}</small
          >
        </article>
        <article class="valuation-strip__cell">
          <span>{{ t("portfolio.terminal.estimatedAum") }}</span>
          <strong>{{
            estimatedAum > 0
              ? formatExact(estimatedAum, valuationCurrency)
              : "—"
          }}</strong>
          <small
            :data-trend="
              Number(intradayFundValuation?.delta_pct_vs_last_close) > 0
                ? 'up'
                : Number(intradayFundValuation?.delta_pct_vs_last_close) < 0
                  ? 'down'
                  : 'flat'
            "
          >
            {{ liveEstimateHint }}
          </small>
        </article>
        <article class="valuation-strip__cell">
          <span>{{ t("portfolio.terminal.officialUnitNav") }}</span>
          <strong>{{ formatUnitNav(officialUnitNav) }}</strong>
          <small>{{ fundUnitScopeHint }}</small>
        </article>
        <article class="valuation-strip__cell">
          <span>{{ t("portfolio.terminal.estimatedUnitNav") }}</span>
          <strong>{{ formatUnitNav(estimatedUnitNav) }}</strong>
          <small>{{
            estimatedUnitNav
              ? t("portfolio.terminal.liveMarkedUnitNav")
              : fundUnitScopeHint
          }}</small>
        </article>
        <article class="valuation-strip__cell">
          <span>{{ t("portfolio.terminal.unitsOutstanding") }}</span>
          <strong>{{ formatUnits(unitsOutstanding) }}</strong>
          <small>{{
            unitsOutstanding
              ? t("portfolio.terminal.linkedFundUnits")
              : t("portfolio.terminal.noUnits")
          }}</small>
        </article>
        <article class="valuation-strip__cell">
          <span>{{ t("portfolio.terminal.lastSettled") }}</span>
          <strong class="valuation-strip__date">{{
            officialBusinessDate || "—"
          }}</strong>
          <small>{{ t("portfolio.terminal.businessDate") }}</small>
        </article>
        <article class="valuation-strip__cell">
          <span>{{ t("portfolio.terminal.dataFreshness") }}</span>
          <strong
            class="valuation-strip__freshness"
            :data-state="freshness.state"
          >
            <i aria-hidden="true" />{{ freshness.label }}
          </strong>
          <small
            >{{ freshnessSource }} ·
            {{ formatSnapshotTime(freshnessTimestamp) }}</small
          >
        </article>
      </section>

      <div class="portfolio-terminal__layout">
        <section class="terminal-panel positions-panel">
          <header class="terminal-panel__header positions-panel__header">
            <div>
              <h2>{{ t("portfolio.terminal.positions.title") }}</h2>
              <p>
                {{
                  t("portfolio.terminal.positions.subtitle", {
                    count: positionRows.length,
                    code: ctx.portfolio.value.code,
                  })
                }}
              </p>
            </div>
            <dl class="positions-panel__summary">
              <div>
                <dt>{{ t("portfolio.terminal.positions.marketValue") }}</dt>
                <dd>
                  {{
                    latestValuation
                      ? formatMoneyCompact(marketValue, valuationCurrency)
                      : "—"
                  }}
                </dd>
              </div>
              <div>
                <dt>{{ t("portfolio.terminal.positions.costBasis") }}</dt>
                <dd>
                  {{
                    latestValuation
                      ? formatMoneyCompact(costBasis, valuationCurrency)
                      : "—"
                  }}
                </dd>
              </div>
              <div>
                <dt>{{ t("portfolio.terminal.positions.unrealisedPnl") }}</dt>
                <dd
                  :data-trend="
                    unrealisedPnl > 0
                      ? 'up'
                      : unrealisedPnl < 0
                        ? 'down'
                        : 'flat'
                  "
                >
                  {{
                    latestValuation
                      ? formatMoneyCompact(
                          unrealisedPnl,
                          valuationCurrency,
                          true,
                        )
                      : "—"
                  }}
                </dd>
              </div>
            </dl>
          </header>

          <div class="positions-panel__table-wrap">
            <table class="positions-table">
              <thead>
                <tr class="positions-table__groups">
                  <th colspan="3">
                    {{ t("portfolio.terminal.positions.groups.instrument") }}
                  </th>
                  <th colspan="3">
                    {{ t("portfolio.terminal.positions.groups.position") }}
                  </th>
                  <th colspan="2">
                    {{ t("portfolio.terminal.positions.groups.valuation") }}
                  </th>
                  <th colspan="2">
                    {{ t("portfolio.terminal.positions.groups.pnlFeed") }}
                  </th>
                </tr>
                <tr>
                  <th>
                    {{ t("portfolio.terminal.positions.columns.ticker") }}
                  </th>
                  <th>{{ t("portfolio.terminal.positions.columns.name") }}</th>
                  <th>
                    {{ t("portfolio.terminal.positions.columns.assetClass") }}
                  </th>
                  <th class="num">
                    {{ t("portfolio.terminal.positions.columns.quantity") }}
                  </th>
                  <th class="num">
                    {{ t("portfolio.terminal.positions.columns.averageCost") }}
                  </th>
                  <th class="num">
                    {{ t("portfolio.terminal.positions.columns.latestPrice") }}
                  </th>
                  <th class="num">
                    {{ t("portfolio.terminal.positions.columns.marketValue") }}
                  </th>
                  <th class="num">
                    {{ t("portfolio.terminal.positions.columns.costBasis") }}
                  </th>
                  <th class="num">
                    {{
                      t("portfolio.terminal.positions.columns.unrealisedPnl")
                    }}
                  </th>
                  <th class="num">
                    {{ t("portfolio.terminal.positions.columns.feed") }}
                  </th>
                </tr>
              </thead>
              <tbody v-if="positionRows.length || cashBalances.length">
                <tr v-for="row in positionRows" :key="row.key">
                  <td>
                    <strong class="positions-table__ticker">{{
                      row.ticker
                    }}</strong>
                  </td>
                  <td class="positions-table__name" :title="row.name">
                    {{ row.name }}
                  </td>
                  <td>
                    <span class="positions-table__asset">{{
                      row.assetClass
                    }}</span>
                  </td>
                  <td class="num">{{ formatQuantity(row.quantity) }}</td>
                  <td class="num">
                    {{ formatExact(row.averageCost, row.priceCurrency) }}
                  </td>
                  <td class="num">
                    {{ formatExact(row.latestPrice, row.priceCurrency) }}
                  </td>
                  <td class="num positions-table__strong">
                    {{ formatExact(row.marketValue, valuationCurrency) }}
                  </td>
                  <td class="num">
                    {{ formatExact(row.costBasis, valuationCurrency) }}
                  </td>
                  <td
                    class="num"
                    :data-trend="
                      (row.pnl ?? 0) > 0
                        ? 'up'
                        : (row.pnl ?? 0) < 0
                          ? 'down'
                          : 'flat'
                    "
                  >
                    {{
                      row.pnl === null
                        ? "—"
                        : formatExact(row.pnl, valuationCurrency, true)
                    }}
                  </td>
                  <td class="num">
                    <span
                      class="feed-badge"
                      :data-state="
                        row.isStale
                          ? 'warning'
                          : row.hasSnapshot
                            ? 'ok'
                            : 'neutral'
                      "
                    >
                      <i aria-hidden="true" />
                      {{
                        row.isStale
                          ? t("portfolio.terminal.positions.feed.stale")
                          : row.hasSnapshot
                            ? t("portfolio.terminal.positions.feed.snapshot")
                            : t("portfolio.terminal.positions.feed.ledger")
                      }}
                    </span>
                  </td>
                </tr>
                <tr
                  v-for="cash in cashBalances"
                  :key="`cash-${cash.currency}`"
                  class="positions-table__cash-row"
                >
                  <td><strong class="positions-table__ticker">CASH</strong></td>
                  <td>
                    {{ t("portfolio.terminal.positions.cashEquivalents") }}
                  </td>
                  <td>
                    <span
                      class="positions-table__asset positions-table__asset--cash"
                      >{{ t("portfolio.allocation.cash") }}</span
                    >
                  </td>
                  <td class="num">—</td>
                  <td class="num">—</td>
                  <td class="num">—</td>
                  <td class="num positions-table__strong">
                    {{ formatExact(cash.balance, cash.currency) }}
                  </td>
                  <td class="num">
                    {{ formatExact(cash.balance, cash.currency) }}
                  </td>
                  <td class="num">—</td>
                  <td class="num">
                    <span class="feed-badge" data-state="ok"
                      ><i aria-hidden="true" />{{
                        t("portfolio.terminal.positions.feed.settled")
                      }}</span
                    >
                  </td>
                </tr>
              </tbody>
              <tbody v-else>
                <tr>
                  <td colspan="10" class="positions-table__empty">
                    {{
                      loadingData
                        ? t("portfolio.holdings.loading")
                        : t("portfolio.holdings.empty")
                    }}
                  </td>
                </tr>
              </tbody>
              <tfoot v-if="latestValuation">
                <tr>
                  <th colspan="6">
                    {{ t("portfolio.terminal.positions.totalIncludingCash") }}
                  </th>
                  <td class="num">
                    {{ formatExact(officialAum, valuationCurrency) }}
                  </td>
                  <td class="num">
                    {{ formatExact(costBasis, valuationCurrency) }}
                  </td>
                  <td
                    class="num"
                    :data-trend="
                      unrealisedPnl > 0
                        ? 'up'
                        : unrealisedPnl < 0
                          ? 'down'
                          : 'flat'
                    "
                  >
                    {{ formatExact(unrealisedPnl, valuationCurrency, true) }}
                  </td>
                  <td />
                </tr>
              </tfoot>
            </table>
          </div>
          <p class="positions-panel__footnote">
            {{ t("portfolio.terminal.positions.footnote") }}
          </p>
        </section>

        <aside class="portfolio-terminal__rail">
          <section class="terminal-panel allocation-panel">
            <header class="terminal-panel__header allocation-panel__header">
              <div>
                <h2>{{ t("portfolio.allocation.title") }}</h2>
                <p>{{ allocationSubtitle }}</p>
              </div>
              <div
                class="allocation-panel__tabs"
                role="tablist"
                :aria-label="t('portfolio.terminal.allocation.tabsLabel')"
              >
                <button
                  v-for="tab in allocationTabs"
                  :key="tab.key"
                  type="button"
                  role="tab"
                  :disabled="tab.disabled"
                  :title="
                    tab.disabled
                      ? t('portfolio.terminal.allocation.dimensionUnavailable')
                      : undefined
                  "
                  :aria-selected="allocationMode === tab.key"
                  :class="{ 'is-active': allocationMode === tab.key }"
                  @click="setAllocationMode(tab.key)"
                >
                  {{ tab.label }}
                </button>
              </div>
            </header>
            <div class="allocation-panel__body">
              <PortfolioAllocationDonut
                compact
                :items="allocationItems"
                :total-value="allocationTotal"
                :currency="valuationCurrency"
                :loading="loadingData"
              />
            </div>
          </section>

          <section class="terminal-panel ratios-panel">
            <header class="terminal-panel__header">
              <div>
                <h2>{{ t("portfolio.terminal.ratios.title") }}</h2>
                <p>{{ t("portfolio.terminal.ratios.subtitle") }}</p>
              </div>
            </header>
            <div class="ratios-panel__empty">
              <svg
                aria-hidden="true"
                viewBox="0 0 24 24"
                width="22"
                height="22"
              >
                <path
                  d="M6 4v3m0 10v3m12-16v3m0 10v3M3 7h6v10H3V7Zm12 0h6v10h-6V7ZM9 12h6"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                />
              </svg>
              <div>
                <strong>{{ t("portfolio.terminal.ratios.emptyTitle") }}</strong>
                <p>{{ t("portfolio.terminal.ratios.emptyBody") }}</p>
              </div>
            </div>
          </section>

          <section class="terminal-panel history-panel">
            <header class="terminal-panel__header">
              <div>
                <h2>{{ historyTitle }}</h2>
                <p>{{ historySubtitle }}</p>
              </div>
              <svg
                class="history-panel__icon"
                aria-hidden="true"
                viewBox="0 0 24 24"
                width="14"
                height="14"
              >
                <path
                  d="M8 3H3v5M16 3h5v5M8 21H3v-5m13 5h5v-5"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                />
              </svg>
            </header>
            <PortfolioValuationHistory
              :snapshots="historyWithLatest"
              :fund-history="fundNavHistory"
              :currency="valuationCurrency"
              :loading="loadingData || loadingHistory"
              :range="historyRange"
              @range-change="handleHistoryRangeChange"
            />
          </section>
        </aside>
      </div>
    </template>

    <div v-else class="portfolio-terminal__state">
      {{ t("portfolio.overview.notFound") }}
    </div>
  </section>
</template>

<style scoped>
.portfolio-terminal {
  --terminal-bg: color-mix(
    in srgb,
    var(--bg-page, #ffffff) 96%,
    var(--text-primary, #1f2328) 4%
  );
  --terminal-panel: var(--bg-card, #ffffff);
  --terminal-raised: color-mix(
    in srgb,
    var(--bg-card, #ffffff) 94%,
    var(--text-primary, #1f2328) 6%
  );
  --terminal-border: var(--border-default, #d0d7de);
  --terminal-border-soft: var(--border-subtle, #e1e8ed);
  --terminal-text: var(--text-primary, #1f2328);
  --terminal-muted: var(--text-secondary, #57606a);
  --terminal-dim: var(--text-tertiary, #6e7781);
  --terminal-blue: var(--action-primary, #2563eb);
  --terminal-green: var(--state-success, #1f883d);
  --terminal-red: var(--state-danger, #cf222e);
  --terminal-amber: var(--state-warning, #9a6700);
  --terminal-series-1: #2563eb;
  --terminal-series-2: #0d9488;
  --terminal-series-3: #d97706;
  --terminal-series-4: #7c3aed;
  --terminal-series-5: #db2777;
  --terminal-series-6: #4f46e5;
  --terminal-series-other: var(--text-tertiary, #64748b);
  --terminal-mono:
    ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;

  width: auto;
  min-height: calc(100dvh - var(--header-height, 56px) - 42px);
  display: grid;
  align-content: start;
  gap: 10px;
  padding: 10px;
  margin: calc(-1 * var(--space-8, 32px));
  border-radius: 0;
  background: var(--terminal-bg);
  color: var(--terminal-text);
  box-sizing: border-box;
}

:global(:root[data-theme="dark"]) .portfolio-terminal {
  --terminal-series-1: #5b7cfa;
  --terminal-series-2: #43c6b8;
  --terminal-series-3: #e7a93e;
  --terminal-series-4: #a78bfa;
  --terminal-series-5: #f472b6;
  --terminal-series-6: #818cf8;
}

.portfolio-terminal__state,
.portfolio-terminal__error {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 24px;
  border: 1px solid var(--terminal-border);
  border-radius: 7px;
  background: var(--terminal-panel);
  color: var(--terminal-muted);
  font-size: 12px;
}

.portfolio-terminal__spinner {
  width: 14px;
  height: 14px;
  border: 2px solid var(--terminal-border);
  border-top-color: var(--terminal-blue);
  border-radius: 50%;
  animation: terminal-spin 0.7s linear infinite;
}

.portfolio-terminal__error {
  flex-direction: column;
}
.portfolio-terminal__error div {
  display: grid;
  gap: 4px;
  text-align: center;
}
.portfolio-terminal__error strong {
  color: var(--terminal-red);
}
.portfolio-terminal__error button {
  padding: 6px 12px;
  border: 1px solid var(--terminal-border);
  border-radius: 5px;
  background: var(--terminal-raised);
  color: var(--terminal-text);
  cursor: pointer;
}

.portfolio-terminal__data-error {
  padding: 8px 11px;
  border: 1px solid
    color-mix(in srgb, var(--terminal-red) 45%, var(--terminal-border));
  border-radius: 5px;
  background: color-mix(in srgb, var(--terminal-red) 8%, var(--terminal-panel));
  color: var(--terminal-red);
  font-size: 11px;
}

.valuation-strip {
  display: grid;
  grid-template-columns: 1.35fr 1.35fr 0.85fr 0.85fr 0.72fr 0.72fr 0.9fr;
  overflow: hidden;
  border: 1px solid var(--terminal-border);
  border-radius: 7px;
  background: var(--terminal-panel);
}

.valuation-strip__cell {
  min-width: 0;
  min-height: 75px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 5px;
  padding: 10px 12px;
  border-right: 1px solid var(--terminal-border-soft);
}

.valuation-strip__cell:last-child {
  border-right: 0;
}
.valuation-strip__cell > span {
  color: var(--terminal-muted);
  font-size: 8px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.valuation-strip__cell strong {
  overflow: hidden;
  color: var(--terminal-text);
  font-family: var(--terminal-mono);
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  font-weight: 800;
  line-height: 1.15;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.valuation-strip__cell--emphasis strong {
  font-size: 15px;
}
.valuation-strip__cell small {
  overflow: hidden;
  color: var(--terminal-muted);
  font-size: 9px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.valuation-strip__cell small[data-trend="up"] {
  color: var(--terminal-green);
}
.valuation-strip__cell small[data-trend="down"] {
  color: var(--terminal-red);
}

.valuation-strip__date {
  font-size: 10px !important;
}
.valuation-strip__freshness {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-family: inherit !important;
  font-size: 10px !important;
}
.valuation-strip__freshness i,
.feed-badge i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}
.valuation-strip__freshness[data-state="ok"] {
  color: var(--terminal-green);
}
.valuation-strip__freshness[data-state="warning"] {
  color: var(--terminal-amber);
}
.valuation-strip__freshness[data-state="missing"] {
  color: var(--terminal-muted);
}

.portfolio-terminal__layout {
  min-width: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 310px;
  gap: 10px;
  align-items: start;
}

.terminal-panel {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--terminal-border);
  border-radius: 7px;
  background: var(--terminal-panel);
  box-shadow: 0 1px 2px color-mix(in srgb, var(--terminal-text) 7%, transparent);
}

.terminal-panel__header {
  min-height: 53px;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 11px 13px;
  border-bottom: 1px solid var(--terminal-border-soft);
}

.terminal-panel__header h2 {
  margin: 0;
  color: var(--terminal-text);
  font-size: 12px;
  font-weight: 750;
  line-height: 1.3;
}

.terminal-panel__header p {
  margin: 3px 0 0;
  color: var(--terminal-muted);
  font-size: 9px;
  line-height: 1.35;
}

.positions-panel__header {
  align-items: center;
}
.positions-panel__summary {
  display: flex;
  gap: 22px;
  margin: 0;
}

.positions-panel__summary div {
  display: grid;
  gap: 3px;
  text-align: right;
}
.positions-panel__summary dt {
  color: var(--terminal-muted);
  font-size: 7px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.positions-panel__summary dd {
  margin: 0;
  color: var(--terminal-text);
  font-family: var(--terminal-mono);
  font-size: 10px;
  font-weight: 800;
}

[data-trend="up"] {
  color: var(--terminal-green) !important;
}
[data-trend="down"] {
  color: var(--terminal-red) !important;
}

.positions-panel__table-wrap {
  width: 100%;
  overflow-x: auto;
}
.positions-table {
  width: 100%;
  min-width: 1040px;
  border-collapse: collapse;
  color: var(--terminal-text);
  font-size: 9px;
}

.positions-table th,
.positions-table td {
  padding: 8px 10px;
  border-bottom: 1px solid var(--terminal-border-soft);
  text-align: left;
  white-space: nowrap;
}

.positions-table thead th {
  color: var(--terminal-muted);
  background: var(--terminal-panel);
  font-size: 7px;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.positions-table__groups th {
  padding-top: 6px;
  padding-bottom: 6px;
  background: var(--terminal-raised) !important;
  color: var(--terminal-dim) !important;
  letter-spacing: 0.1em !important;
  text-transform: uppercase;
}

.positions-table tbody tr:hover td {
  background: color-mix(in srgb, var(--terminal-blue) 5%, transparent);
}
.positions-table .num {
  text-align: right;
  font-family: var(--terminal-mono);
  font-variant-numeric: tabular-nums;
}
.positions-table__ticker {
  color: var(--text-link, var(--terminal-blue));
  font-family: var(--terminal-mono);
  font-size: 9px;
}
.positions-table__name {
  max-width: 165px;
  overflow: hidden;
  color: var(--terminal-text);
  text-overflow: ellipsis;
}
.positions-table__asset {
  display: inline-flex;
  padding: 2px 7px;
  border: 1px solid color-mix(in srgb, var(--terminal-blue) 48%, transparent);
  border-radius: 4px;
  background: color-mix(in srgb, var(--terminal-blue) 15%, transparent);
  color: var(--text-link, var(--terminal-blue));
  font-size: 8px;
}

.positions-table__asset--cash {
  border-color: var(--terminal-border);
  background: transparent;
  color: var(--terminal-muted);
}
.positions-table__strong {
  color: var(--terminal-text);
  font-weight: 800;
}
.positions-table__cash-row td {
  color: var(--terminal-muted);
}
.positions-table__empty {
  height: 120px;
  color: var(--terminal-muted);
  text-align: center !important;
}

.feed-badge {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 6px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--terminal-muted) 10%, transparent);
  color: var(--terminal-muted);
  font-family: inherit;
  font-size: 8px;
  font-weight: 700;
}

.feed-badge i {
  width: 5px;
  height: 5px;
}
.feed-badge[data-state="ok"] {
  background: color-mix(in srgb, var(--terminal-green) 12%, transparent);
  color: var(--terminal-green);
}
.feed-badge[data-state="warning"] {
  background: color-mix(in srgb, var(--terminal-amber) 12%, transparent);
  color: var(--terminal-amber);
}
.positions-table tfoot th,
.positions-table tfoot td {
  border-bottom: 0;
  background: var(--terminal-raised);
  font-weight: 800;
}
.positions-table tfoot th {
  color: var(--terminal-muted);
  font-size: 8px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.positions-panel__footnote {
  margin: 0;
  padding: 11px 13px;
  color: var(--terminal-muted);
  font-size: 8px;
  line-height: 1.5;
}

.portfolio-terminal__rail {
  min-width: 0;
  display: grid;
  gap: 10px;
}
.allocation-panel__header {
  display: grid;
  grid-template-columns: 76px minmax(0, 1fr);
  align-items: start;
  gap: 8px;
}
.allocation-panel__tabs {
  min-width: 0;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  border: 1px solid var(--terminal-border);
  border-radius: 4px;
  overflow: hidden;
}
.allocation-panel__tabs button {
  min-width: 0;
  padding: 5px 3px;
  border: 0;
  border-right: 1px solid var(--terminal-border);
  background: transparent;
  color: var(--terminal-muted);
  font-size: 7px;
  cursor: pointer;
}
.allocation-panel__tabs button:last-child {
  border-right: 0;
}
.allocation-panel__tabs button:hover:not(:disabled),
.allocation-panel__tabs button:focus-visible {
  color: var(--terminal-text);
  outline: none;
}
.allocation-panel__tabs button.is-active {
  background: color-mix(in srgb, var(--terminal-blue) 12%, transparent);
  color: var(--text-link, var(--terminal-blue));
}
.allocation-panel__tabs button:disabled {
  color: var(--terminal-dim);
  cursor: not-allowed;
}
.allocation-panel__body {
  padding: 13px;
}

.allocation-panel :deep(.allocation-donut__track) {
  stroke: var(--terminal-raised);
}
.allocation-panel :deep(.allocation-donut__centre-label),
.allocation-panel :deep(.allocation-donut__centre-meta),
.allocation-panel :deep(.allocation-donut__legend-value) {
  color: var(--terminal-muted);
}
.allocation-panel :deep(.allocation-donut__centre-value),
.allocation-panel :deep(.allocation-donut__legend-label) {
  color: var(--terminal-text);
}
.allocation-panel :deep(.allocation-donut__legend-button:hover),
.allocation-panel :deep(.allocation-donut__legend-button.is-active) {
  background: var(--terminal-raised);
}
.allocation-panel :deep(.allocation-donut__state) {
  min-height: 130px;
  color: var(--terminal-muted);
}

.ratios-panel__empty {
  min-height: 104px;
  display: grid;
  grid-template-columns: 26px minmax(0, 1fr);
  gap: 10px;
  align-items: start;
  margin: 12px;
  padding: 15px;
  border: 1px dashed var(--terminal-border);
  border-radius: 5px;
  color: var(--terminal-muted);
}
.ratios-panel__empty strong {
  color: var(--terminal-text);
  font-size: 10px;
}
.ratios-panel__empty p {
  margin: 5px 0 0;
  font-size: 9px;
  line-height: 1.55;
}
.history-panel__icon {
  color: var(--terminal-muted);
}

@keyframes terminal-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1200px) {
  .portfolio-terminal {
    margin: calc(-1 * var(--space-6, 24px));
  }
}

@media (max-width: 1180px) {
  .valuation-strip {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
  .valuation-strip__cell {
    border-bottom: 1px solid var(--terminal-border-soft);
  }
  .valuation-strip__cell:nth-child(4n) {
    border-right: 0;
  }
  .valuation-strip__cell:nth-child(n + 5) {
    border-bottom: 0;
  }
  .portfolio-terminal__layout {
    grid-template-columns: 1fr;
  }
  .portfolio-terminal__rail {
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    align-items: start;
  }
}

@media (max-width: 720px) {
  .portfolio-terminal {
    padding: 7px;
  }
  .valuation-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .valuation-strip__cell:nth-child(4n) {
    border-right: 1px solid var(--terminal-border-soft);
  }
  .valuation-strip__cell:nth-child(2n) {
    border-right: 0;
  }
  .valuation-strip__cell:nth-child(n + 5) {
    border-bottom: 1px solid var(--terminal-border-soft);
  }
  .valuation-strip__cell:nth-child(n + 7) {
    border-bottom: 0;
  }
  .positions-panel__header {
    align-items: flex-start;
    flex-direction: column;
  }
  .positions-panel__summary {
    width: 100%;
    justify-content: space-between;
    gap: 10px;
  }
  .positions-panel__summary div {
    text-align: left;
  }
  .portfolio-terminal__rail {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 640px) {
  .portfolio-terminal {
    margin-top: calc(-1 * var(--space-6, 24px));
    margin-right: calc(-1 * var(--space-5, 20px));
    margin-bottom: calc(-1 * var(--space-6, 24px));
    margin-left: calc(-1 * var(--space-5, 20px));
  }
}

@media (max-width: 430px) {
  .valuation-strip {
    grid-template-columns: 1fr;
  }
  .valuation-strip__cell {
    min-height: 66px;
    border-right: 0 !important;
    border-bottom: 1px solid var(--terminal-border-soft) !important;
  }
  .valuation-strip__cell:last-child {
    border-bottom: 0 !important;
  }
  .positions-panel__summary {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
  .positions-panel__summary dd {
    font-size: 9px;
  }
  .allocation-panel__header {
    grid-template-columns: 1fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .portfolio-terminal__spinner {
    animation: none;
  }
}
</style>
