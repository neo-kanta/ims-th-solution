<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import FundWorkspaceHeader from "~/features/investment-workspace/components/FundWorkspaceHeader.vue";
import HoldingsSubTabs from "~/features/investment-workspace/components/HoldingsSubTabs.vue";
import HoldingsToolbar from "~/features/investment-workspace/components/HoldingsToolbar.vue";
import AllocationDonutChart from "~/features/investment-workspace/components/AllocationDonutChart.vue";
import IntradayHoldingsCards from "~/features/investment-workspace/components/IntradayHoldingsCards.vue";
import LivePositionsTable from "~/features/investment-workspace/components/LivePositionsTable.vue";
import RealNavHistoryChart from "~/features/investment-workspace/components/RealNavHistoryChart.vue";
import { useFundWorkspace } from "~/features/investment-workspace/composables/useFundWorkspace";
import { useHoldings } from "~/features/investment-workspace/composables/useHoldings";
import { buildRealSummary } from "~/features/investment-workspace/lib/realSummary";
import {
  intradayValuationApi,
  type IntradayValuation,
  type MarketDataRefresh,
  type MarketDataStatus,
} from "~/features/investment-workspace/services/intradayValuationApi";
import { myFundsApi, useFundDetail } from "~/features/my-funds";
import type {
  FundAllocation,
  FundNavHistory,
  FundNavHistoryRange,
  FundNavSnapshot,
} from "~/features/my-funds/services/myFundsApi";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const fundIdParam = computed(() => String(route.params.fundId ?? ""));

// Fund switcher (real backend `GET /investment/funds`).
const { funds, loadFunds, setActiveFund } = useFundWorkspace();

// Local UI-state controller (date picker, sub-tab, export spinner).
const holdings = useHoldings();

// Real backplane — drives identity (breadcrumb / contract badge) and the
// "Last Settled" label for the actual selected fund.
const fundDetail = useFundDetail(() => fundIdParam.value);
const fundNav = ref<FundNavSnapshot | null>(null);

// Live (intraday) backplane — replaces the legacy mocked KPI values.
const intraday = ref<IntradayValuation | null>(null);
const intradayLoading = ref(false);
const feedStatus = ref<MarketDataStatus | null>(null);

// Real allocation + NAV history (server-computed from the ledger).
const allocation = ref<FundAllocation | null>(null);
const allocationLoading = ref(false);
const navHistory = ref<FundNavHistory | null>(null);
const navHistoryRange = ref<FundNavHistoryRange>("3M");
const navHistoryLoading = ref(false);

// Inline toast banner. We keep the toast in-page rather than introducing a
// global notification library — keeps the scope tight and avoids a new dep.
type ToastTone = "success" | "warning" | "error" | "info";
const toast = ref<{ tone: ToastTone; text: string } | null>(null);
let toastTimer: ReturnType<typeof setTimeout> | null = null;

function showToast(tone: ToastTone, text: string) {
  toast.value = { tone, text };
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = setTimeout(() => {
    toast.value = null;
  }, 5000);
}

async function refreshFundDetail() {
  if (!fundIdParam.value) return;
  await fundDetail.load();
  try {
    fundNav.value = await myFundsApi.getLatestFundNav(fundIdParam.value);
  } catch {
    fundNav.value = null;
  }
}

async function refreshIntradayValuation() {
  if (!fundIdParam.value) return;
  intradayLoading.value = true;
  try {
    intraday.value = await intradayValuationApi.getFundValuation(fundIdParam.value);
  } catch {
    intraday.value = null;
  } finally {
    intradayLoading.value = false;
  }
}

async function refreshFeedStatus() {
  if (!fundIdParam.value) return;
  try {
    feedStatus.value = await intradayValuationApi.getFundFeedStatus(fundIdParam.value);
  } catch {
    feedStatus.value = null;
  }
}

async function refreshAllocation() {
  if (!fundIdParam.value) return;
  allocationLoading.value = true;
  try {
    allocation.value = await myFundsApi.getFundAllocation(fundIdParam.value);
  } catch {
    allocation.value = null;
  } finally {
    allocationLoading.value = false;
  }
}

async function refreshNavHistory(next?: FundNavHistoryRange) {
  if (!fundIdParam.value) return;
  if (next) navHistoryRange.value = next;
  navHistoryLoading.value = true;
  try {
    navHistory.value = await myFundsApi.getFundNavHistory(
      fundIdParam.value,
      navHistoryRange.value,
    );
  } catch {
    navHistory.value = null;
  } finally {
    navHistoryLoading.value = false;
  }
}

onMounted(async () => {
  await Promise.allSettled([
    loadFunds(fundIdParam.value),
    refreshFundDetail(),
    refreshAllocation(),
    refreshNavHistory(),
    refreshFeedStatus(),
  ]);
  // Intraday valuation needs the fund detail loaded for the breadcrumb to
  // resolve — but the API itself is fund-id keyed, so we run it next.
  await refreshIntradayValuation();
});

watch(
  () => fundIdParam.value,
  async (next) => {
    setActiveFund(next);
    await Promise.allSettled([
      refreshFundDetail(),
      refreshAllocation(),
      refreshNavHistory(),
      refreshFeedStatus(),
    ]);
    await refreshIntradayValuation();
  },
);

// Identity / breadcrumb / contract badge come from the My Funds card. The
// intraday KPI strip now consumes its own backplane (the IntradayValuation),
// so realSummary only fuels the FundWorkspaceHeader.
const effectiveSummary = computed(() =>
  fundDetail.card.value
    ? buildRealSummary({
        card: fundDetail.card.value,
        portfolios: fundDetail.portfolios.value,
        nav: fundNav.value,
        asOf: holdings.asOf.value,
      })
    : null,
);

const fundShortName = computed(
  () => effectiveSummary.value?.fund.short_name ?? "",
);

const pageSubtitle = computed(() => {
  const name = fundShortName.value || t("holdings.page.unknownFund", "this fund");
  return t(
    "holdings.page.subtitleTemplate",
    { fund: name },
    `A daily snapshot of liquid, exposed, and allocated positions for ${name}. Numbers reflect the most recent Accounting Closing.`,
  );
});

function onChangeFund(slug: string) {
  void router.push(`/investment/funds/${slug}/holdings`);
}

async function onChangeAsOf(date: string) {
  await holdings.setAsOf(date);
  await refreshFundDetail();
}

// Refresh button: POST the live refresh endpoint, then reload every read-side
// payload. The toast mirrors the spec:
//   success → "Market data refreshed"
//   stale   → "Provider unavailable. Showing latest cached snapshot."
//   error   → "No market data available for this contract."
async function onRefreshAll() {
  if (!fundIdParam.value) {
    showToast("error", t("holdings.toast.error", "No market data available for this contract."));
    return;
  }
  let summary: MarketDataRefresh | null = null;
  try {
    summary = await intradayValuationApi.refreshFundQuotes(fundIdParam.value);
  } catch {
    summary = null;
  }

  await Promise.allSettled([
    refreshFundDetail(),
    refreshAllocation(),
    refreshNavHistory(),
    refreshFeedStatus(),
    refreshIntradayValuation(),
  ]);

  if (!summary) {
    showToast("error", t("holdings.toast.error", "No market data available for this contract."));
    return;
  }
  const failed = summary.failed_symbols ?? 0;
  const success = summary.success_symbols ?? 0;
  const stale = summary.stale_symbols ?? 0;
  const unmapped = summary.unmapped_symbols?.length ?? 0;
  if (success > 0 && failed + unmapped === 0 && stale === 0) {
    showToast("success", t("holdings.toast.success", "Market data refreshed"));
  } else if (success > 0 || stale > 0) {
    showToast(
      "warning",
      t("holdings.toast.stale", "Provider unavailable. Showing latest cached snapshot."),
    );
  } else {
    showToast("error", t("holdings.toast.error", "No market data available for this contract."));
  }
}

const positionsSubtitle = computed(() => {
  const portfolios = fundDetail.portfolios.value.length;
  const positions = intraday.value?.positions?.length ?? 0;
  return t(
    "holdings.positions.cardSubtitle",
    { positions, portfolios },
    `${positions} position${positions === 1 ? "" : "s"} across ${portfolios} portfolio${portfolios === 1 ? "" : "s"}.`,
  );
});

const valuationCcy = computed(
  () =>
    intraday.value?.valuation_ccy ??
    effectiveSummary.value?.kpis.today_nav.currency ??
    fundDetail.card.value?.valuation.valuation_ccy ??
    fundDetail.card.value?.base_currency ??
    "THB",
);

const lastRefreshAt = computed(
  () => intraday.value?.as_of || fundDetail.card.value?.updated_at || new Date().toISOString(),
);

const lastSettledDate = computed(
  () => intraday.value?.business_date || fundDetail.card.value?.valuation.business_date || "",
);

const workflowLabel = computed(() => {
  const state = fundDetail.card.value?.workflow.current_state ?? "";
  switch (state) {
    case "ACCOUNTING_CLOSED":
      return "acctg closing";
    case "TRANSACTION_CLOSED":
      return "tx closing";
    case "MANAGER_APPROVED":
      return "mgr approved";
    case "DAY_OPEN":
      return "day open";
    default:
      return "valuation";
  }
});

// Drives the Toolbar's "Feeds OK / Stale data" pill.
const toolbarFreshness = computed(() => {
  if (!intraday.value && !feedStatus.value) {
    return effectiveSummary.value?.freshness ?? null;
  }
  const stale =
    intraday.value?.is_stale ||
    (feedStatus.value ? !feedStatus.value.healthy : false);
  return {
    is_stale: stale,
    label: stale
      ? intraday.value?.stale_reason ||
        feedStatus.value?.note ||
        "Provider unavailable"
      : "Inputs fresh",
  };
});

// CSV export uses ONLY the rows already loaded from the backend — no mock
// fallback, no fabricated totals.
async function onExport() {
  if (!fundIdParam.value) return;
  const rows: string[][] = [];
  rows.push(["Fund", fundIdParam.value]);
  rows.push(["As of", lastSettledDate.value || holdings.asOf.value]);
  rows.push(["Official AUM", intraday.value?.official_aum ?? ""]);
  rows.push(["Estimated AUM", intraday.value?.estimated_aum ?? ""]);
  rows.push([]);
  rows.push([
    "Section",
    "Ticker",
    "Name",
    "Asset class",
    "Quantity",
    "Avg cost",
    "Latest price",
    "Market value",
    "Cost basis",
    "Unrealised P&L",
    "Provider",
  ]);
  for (const p of intraday.value?.positions ?? []) {
    rows.push([
      "HOLDING",
      p.ticker ?? "",
      p.name ?? "",
      p.asset_class_label ?? "",
      p.quantity ?? "",
      p.average_cost ?? "",
      p.latest_price ?? "",
      p.market_value ?? "",
      p.cost_basis ?? "",
      p.unrealised_pnl ?? "",
      p.provider ?? "",
    ]);
  }
  rows.push([]);
  rows.push(["Section", "Currency", "Balance"]);
  for (const c of intraday.value?.cash ?? []) {
    rows.push(["CASH", c.currency ?? "", c.balance ?? ""]);
  }
  await holdings.exportRows(
    rows,
    `holdings_${fundIdParam.value}_${lastSettledDate.value || holdings.asOf.value}.csv`,
  );
}

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("holdings.page.title", "Asset positions"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true }
);
</script>

<template>
  <section class="holdings-page">
    <FundWorkspaceHeader v-if="effectiveSummary" :header="effectiveSummary.fund" />

    <div v-if="fundDetail.error.value" class="holdings-page__alert" role="alert">
      {{ fundDetail.error.value }}
    </div>

    <div class="holdings-page__title-row">
      <p class="holdings-page__subtitle">{{ pageSubtitle }}</p>
    </div>

    <HoldingsToolbar
      :funds="funds"
      :active-fund-id="fundIdParam"
      :as-of="holdings.asOf.value"
      :loading="fundDetail.loading.value || intradayLoading"
      :exporting="holdings.exporting.value"
      :freshness="toolbarFreshness"
      @change-fund="onChangeFund"
      @change-as-of="onChangeAsOf"
      @refresh="onRefreshAll"
      @export="onExport"
    />

    <transition name="holdings-toast">
      <div
        v-if="toast"
        class="holdings-page__toast"
        :class="`holdings-page__toast--${toast.tone}`"
        role="status"
        aria-live="polite"
      >
        {{ toast.text }}
      </div>
    </transition>

    <div v-if="holdings.error.value" class="holdings-page__alert" role="alert">
      {{ holdings.error.value }}
    </div>
    <div v-if="holdings.exportError.value" class="holdings-page__alert" role="alert">
      {{ holdings.exportError.value }}
    </div>

    <HoldingsSubTabs
      :active="holdings.subTab.value"
      @select="holdings.setSubTab($event)"
    />

    <IntradayHoldingsCards
      :valuation="intraday"
      :status="feedStatus"
      :last-refresh-at="lastRefreshAt"
      :workflow-label="workflowLabel"
      :last-settled-date="lastSettledDate"
    />

    <div class="holdings-page__layout">
      <div class="holdings-page__main">
        <AppCard
          :title="t('holdings.positions.cardTitle', 'Positions')"
          :subtitle="positionsSubtitle"
        >
          <LivePositionsTable
            :positions="intraday?.positions ?? []"
            :cash-rows="intraday?.cash ?? []"
            :valuation-ccy="valuationCcy"
            :loading="intradayLoading"
          />
        </AppCard>
      </div>

      <aside class="holdings-page__rail">
        <AppCard
          :title="t('holdings.allocation.title', 'Allocation')"
          :subtitle="t('holdings.allocation.liveSubtitle', 'Asset class / sector / country / currency')"
        >
          <AllocationDonutChart
            :payload="allocation"
            :valuation="intraday"
            :loading="allocationLoading"
          />
        </AppCard>

        <AppCard
          :title="t('holdings.ratios.title', 'Special ratios')"
          :subtitle="t('holdings.ratios.subtitle', 'Policy floors and ceilings')"
        >
          <div class="holdings-page__placeholder">
            {{
              t(
                "holdings.ratios.placeholder",
                "Policy-ratio gauges are not configured yet — they will surface once the IRG policy-ratio endpoint is wired up.",
              )
            }}
          </div>
        </AppCard>

        <AppCard
          :title="t('holdings.navHistory.title', 'Unit NAV history')"
          :subtitle="t('holdings.navHistory.liveSubtitle', 'Daily NAV / AUM series')"
        >
          <RealNavHistoryChart
            :payload="navHistory"
            :range="navHistoryRange"
            :loading="navHistoryLoading"
            @change-range="refreshNavHistory($event)"
          />
        </AppCard>
      </aside>
    </div>

    <p class="holdings-page__disclaimer">
      {{
        t(
          "holdings.livePanelsNote",
          "Official accounting values come from closing snapshots. Estimated values are live market-data approximations and are not the authoritative NAV.",
        )
      }}
    </p>
  </section>
</template>

<style scoped>
.holdings-page {
  display: grid;
  gap: var(--space-3);
}

.holdings-page__title-row {
  display: grid;
  gap: 4px;
  padding-top: var(--space-2);
}

.holdings-page__title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.holdings-page__subtitle {
  margin: 0;
  font-size: 13px;
  color: var(--text-secondary);
  max-width: 64rem;
}

.holdings-page__alert {
  padding: 10px 12px;
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: 6px;
  color: var(--alert-danger-text);
  font-size: 13px;
}

.holdings-page__toast {
  padding: 10px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
}
.holdings-page__toast--success {
  background: var(--alert-success-bg);
  border: 1px solid var(--alert-success-border);
  color: var(--state-success);
}
.holdings-page__toast--warning {
  background: var(--alert-warning-bg);
  border: 1px solid var(--alert-warning-border);
  color: var(--state-warning, var(--color-warning-500, #f59e0b));
}
.holdings-page__toast--error {
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  color: var(--alert-danger-text);
}
.holdings-page__toast--info {
  background: var(--alert-info-bg, var(--surface-1));
  border: 1px solid var(--alert-info-border, var(--border-default));
  color: var(--text-primary);
}

.holdings-toast-enter-active,
.holdings-toast-leave-active {
  transition: opacity 0.18s ease;
}
.holdings-toast-enter-from,
.holdings-toast-leave-to {
  opacity: 0;
}

.holdings-page__layout {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr);
  gap: var(--space-3);
}

.holdings-page__main {
  display: grid;
  gap: var(--space-3);
  min-width: 0;
}

.holdings-page__rail {
  display: grid;
  gap: var(--space-3);
  min-width: 0;
  align-content: start;
}

.holdings-page__disclaimer {
  margin: 0;
  font-size: 11px;
  color: var(--text-tertiary);
  text-align: right;
}

.holdings-page__placeholder {
  font-size: 12px;
  color: var(--text-tertiary);
  line-height: 1.55;
  padding: 4px 0;
}

@media (max-width: 1024px) {
  .holdings-page__layout {
    grid-template-columns: 1fr;
  }
}
</style>
