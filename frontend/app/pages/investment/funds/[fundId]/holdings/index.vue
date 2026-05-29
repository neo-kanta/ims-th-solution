<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import FundWorkspaceHeader from "~/features/investment-workspace/components/FundWorkspaceHeader.vue";
import HoldingsKpiCards from "~/features/investment-workspace/components/HoldingsKpiCards.vue";
import HoldingsSubTabs from "~/features/investment-workspace/components/HoldingsSubTabs.vue";
import HoldingsToolbar from "~/features/investment-workspace/components/HoldingsToolbar.vue";
import RealAllocationPanel from "~/features/investment-workspace/components/RealAllocationPanel.vue";
import RealNavHistoryChart from "~/features/investment-workspace/components/RealNavHistoryChart.vue";
import RealPositionsTable from "~/features/investment-workspace/components/RealPositionsTable.vue";
import { useFundWorkspace } from "~/features/investment-workspace/composables/useFundWorkspace";
import { useHoldings } from "~/features/investment-workspace/composables/useHoldings";
import { buildRealSummary } from "~/features/investment-workspace/lib/realSummary";
import { myFundsApi, useFundDetail } from "~/features/my-funds";
import type {
  ApiAssetClass,
  ApiHolding,
  ApiInstrument,
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

const { t, locale } = useI18n();
const route = useRoute();
const router = useRouter();

const fundIdParam = computed(() => String(route.params.fundId ?? ""));

// Mock backplane — still feeds the asset table, allocation, ratios, and the
// NAV history chart until those endpoints land.
const { funds, loadFunds, setActiveFund } = useFundWorkspace();
const holdings = useHoldings(fundIdParam.value, locale.value);

// Real backplane — drives identity (breadcrumb / contract badge) and KPIs
// (NAV, AUM, stale flag, settlement label) for the actual selected fund.
const fundDetail = useFundDetail(() => fundIdParam.value);
const fundNav = ref<FundNavSnapshot | null>(null);

// Real holdings + instrument + asset-class lookups for the positions table.
const realHoldings = ref<ApiHolding[]>([]);
const realInstruments = ref<ApiInstrument[]>([]);
const realAssetClasses = ref<ApiAssetClass[]>([]);
const realCashRows = ref<{ currency: string; balance: string }[]>([]);
const positionsLoading = ref(false);

// Real allocation + NAV history (server-computed from the seed).
const allocation = ref<FundAllocation | null>(null);
const allocationLoading = ref(false);
const navHistory = ref<FundNavHistory | null>(null);
const navHistoryRange = ref<FundNavHistoryRange>("3M");
const navHistoryLoading = ref(false);

async function refreshRealSummary() {
  if (!fundIdParam.value) return;
  await fundDetail.load();
  try {
    fundNav.value = await myFundsApi.getLatestFundNav(fundIdParam.value);
  } catch {
    // NAV is optional — non-unitised funds simply don't have one.
    fundNav.value = null;
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

async function refreshPositions() {
  if (!fundIdParam.value) return;
  positionsLoading.value = true;
  try {
    // The fund detail composable already pulled the portfolios; if it hasn't
    // run yet, wait for it so we know which portfolio(s) to query holdings on.
    if (fundDetail.portfolios.value.length === 0) {
      await fundDetail.load();
    }
    const portfolios = fundDetail.portfolios.value.filter((p) => !!p.id);
    const [classes, instruments, ...holdingsBatches] = await Promise.all([
      myFundsApi.listAssetClasses().catch(() => [] as ApiAssetClass[]),
      myFundsApi.listInstruments(500).catch(() => [] as ApiInstrument[]),
      ...portfolios.map((p) =>
        myFundsApi.listPortfolioHoldings(p.id as string).catch(
          () => [] as ApiHolding[],
        ),
      ),
    ]);
    realAssetClasses.value = classes;
    realInstruments.value = instruments;
    // Flatten holdings from every portfolio under this fund.
    realHoldings.value = (holdingsBatches as ApiHolding[][]).flat();

    // Cash balances per portfolio, flattened with currency stable.
    const cashLists = await Promise.all(
      portfolios.map((p) =>
        myFundsApi.listCash(p.id as string).catch(() => []),
      ),
    );
    realCashRows.value = cashLists
      .flat()
      .map((c) => ({ currency: c.currency ?? "", balance: c.balance ?? "0" }));
  } finally {
    positionsLoading.value = false;
  }
}

onMounted(async () => {
  await Promise.allSettled([
    loadFunds(fundIdParam.value),
    refreshRealSummary(),
    holdings.refreshAll(),
    refreshAllocation(),
    refreshNavHistory(),
  ]);
  // Positions need the portfolios from refreshRealSummary first; run after.
  await refreshPositions();
});

watch(
  () => fundIdParam.value,
  async (next) => {
    setActiveFund(next);
    await Promise.allSettled([
      refreshRealSummary(),
      refreshAllocation(),
      refreshNavHistory(),
    ]);
    await refreshPositions();
  },
);

// `realSummary` is preferred for the header + KPI strip whenever we have a
// usable card; the mock summary remains as a fallback for the rare case where
// the live fetch is still pending so the layout doesn't flash empty.
const realSummary = computed(() =>
  fundDetail.card.value
    ? buildRealSummary({
        card: fundDetail.card.value,
        portfolios: fundDetail.portfolios.value,
        nav: fundNav.value,
        asOf: holdings.asOf.value,
      })
    : null,
);

const effectiveSummary = computed(
  () => realSummary.value ?? holdings.summary.value ?? null,
);

const fundShortName = computed(
  () =>
    realSummary.value?.fund.short_name ??
    holdings.summary.value?.fund.short_name ??
    "",
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
  // Navigate so the URL reflects the active fund (deep-link friendly).
  void router.push(`/investment/funds/${slug}/holdings`);
}

async function onChangeAsOf(date: string) {
  await holdings.setAsOf(date);
  await refreshRealSummary();
}

async function onRefreshAll() {
  await Promise.allSettled([
    holdings.refreshAll(),
    refreshRealSummary(),
    refreshPositions(),
    refreshAllocation(),
    refreshNavHistory(),
  ]);
}

const positionsSubtitle = computed(() => {
  const portfolios = fundDetail.portfolios.value.length;
  const positions = realHoldings.value.length;
  return t(
    "holdings.positions.cardSubtitle",
    { positions, portfolios },
    `${positions} position${positions === 1 ? "" : "s"} across ${portfolios} portfolio${portfolios === 1 ? "" : "s"}.`,
  );
});

const valuationCcy = computed(
  () =>
    realSummary.value?.kpis.today_nav.currency ??
    fundDetail.card.value?.valuation.valuation_ccy ??
    fundDetail.card.value?.base_currency ??
    "THB",
);

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
      <!-- Title is omitted here as it is rendered in layout breadcrumbs instead -->
      <p class="holdings-page__subtitle">{{ pageSubtitle }}</p>
    </div>

    <HoldingsToolbar
      :funds="funds"
      :active-fund-id="fundIdParam"
      :as-of="holdings.asOf.value"
      :loading="holdings.loading.value || fundDetail.loading.value"
      :exporting="holdings.exporting.value"
      :freshness="effectiveSummary?.freshness ?? null"
      @change-fund="onChangeFund"
      @change-as-of="onChangeAsOf"
      @refresh="onRefreshAll"
      @export="holdings.download()"
    />

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

    <HoldingsKpiCards v-if="effectiveSummary" :kpis="effectiveSummary.kpis" />

    <div class="holdings-page__layout">
      <div class="holdings-page__main">
        <AppCard
          :title="t('holdings.positions.cardTitle', 'Positions')"
          :subtitle="positionsSubtitle"
        >
          <RealPositionsTable
            :holdings="realHoldings"
            :instruments="realInstruments"
            :asset-classes="realAssetClasses"
            :cash-rows="realCashRows"
            :valuation-ccy="valuationCcy"
            :loading="positionsLoading"
          />
        </AppCard>
      </div>

      <aside class="holdings-page__rail">
        <AppCard
          :title="t('holdings.allocation.title', 'Allocation')"
          :subtitle="t('holdings.allocation.liveSubtitle', 'Asset class / sector / country / currency')"
        >
          <RealAllocationPanel :payload="allocation" :loading="allocationLoading" />
        </AppCard>

        <AppCard
          :title="t('holdings.ratios.title', 'Special ratios')"
          :subtitle="t('holdings.ratios.subtitle', 'Policy floors and ceilings')"
        >
          <div class="holdings-page__placeholder">
            {{
              t(
                "holdings.ratios.placeholder",
                "IRG-derived ratio gauges will surface here once the policy ratios endpoint is connected to the live ruleset.",
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
          "Header, KPIs, freshness flag, positions table, allocation breakdowns and the NAV history chart all reflect live database state. Only special ratios remain a placeholder until the IRG policy-ratio endpoint is wired up.",
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

.holdings-page__nav-card {
  padding: 12px;
}

.holdings-page__pnl-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 8px;
  background: var(--status-in-review-bg);
  border: 1px solid var(--alert-info-border);
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--status-in-review-text);
}

.holdings-page__pnl-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--state-info);
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
