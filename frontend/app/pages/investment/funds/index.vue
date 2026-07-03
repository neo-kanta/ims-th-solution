<script setup lang="ts">
import { onMounted } from "vue";

import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import { useI18n } from "~/composables/useI18n";
import {
  FundFilterToolbar,
  FundKpiStrip,
  FundSummaryCard,
  useMyFunds,
  useMyFundsKpis,
} from "~/features/my-funds";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const { t } = useI18n();
const router = useRouter();
const {
  cards,
  filtered,
  counts,
  filter,
  sort,
  search,
  businessDate,
  loading,
  decorating,
  error,
  loadAll,
  setFilter,
  setSort,
  setSearch,
} = useMyFunds();
const { kpis } = useMyFundsKpis(cards);

onMounted(() => {
  void loadAll();
});

function openFund(fundId: string) {
  void router.push(`/investment/funds/${fundId}/holdings`);
}

function runPreTrade(fundId: string) {
  void router.push({
    path: "/compliance/pre-trade",
    query: { contract_id: fundId },
  });
}

function viewBreaches(fundId: string) {
  void router.push({
    path: "/compliance/exceptions",
    query: { contract_id: fundId },
  });
}

function writeResearch(fundId: string) {
  void router.push({
    path: "/investment/analysis/new",
    query: { fund_id: fundId },
  });
}
</script>

<template>
  <section class="my-funds-page">
    <AppPageHeader
      :title="t('myFunds.page.title', 'My funds')"
      :description="
        t(
          'myFunds.page.subtitle',
          'Operational cockpit for the funds you can act on today — valuations, workflow, compliance and decisions in one place.',
        )
      "
    >
      <template #actions>
        <div class="my-funds-page__date" aria-live="polite">
          <span class="my-funds-page__date-label">{{ t("myFunds.page.asOf", "Business date") }}</span>
          <span class="my-funds-page__date-value">{{ businessDate }}</span>
        </div>
        <NuxtLink to="/investment/funds/new" class="my-funds-page__new">
          {{ t("myFunds.page.newFund", "+ New fund") }}
        </NuxtLink>
      </template>
    </AppPageHeader>

    <FundKpiStrip :kpis="kpis" :loading="loading && cards.length === 0" />

    <FundFilterToolbar
      :filter="filter"
      :sort="sort"
      :search="search"
      :counts="counts"
      @update:filter="setFilter"
      @update:sort="setSort"
      @update:search="setSearch"
    />

    <div v-if="decorating && !loading" class="my-funds-page__notice" role="status">
      {{ t("myFunds.page.decorating", "Refreshing valuation, workflow and compliance signals…") }}
    </div>

    <div v-if="loading && cards.length === 0" class="my-funds-page__notice" role="status">
      {{ t("myFunds.page.loading", "Loading funds…") }}
    </div>

    <div v-else-if="error" class="my-funds-page__error" role="alert">
      <div class="my-funds-page__error-title">
        {{ t("myFunds.page.errorTitle", "Failed to load funds") }}
      </div>
      <div class="my-funds-page__error-detail">{{ error }}</div>
      <button type="button" class="my-funds-page__retry" @click="loadAll">
        {{ t("myFunds.page.retry", "Retry") }}
      </button>
    </div>

    <div v-else-if="filtered.length === 0" class="my-funds-page__empty">
      {{
        cards.length === 0
          ? t("myFunds.page.empty", "You don't have access to any active funds yet.")
          : t("myFunds.page.noResults", "No funds match the current filter.")
      }}
    </div>

    <div v-else class="my-funds-page__grid">
      <FundSummaryCard
        v-for="card in filtered"
        :key="card.fund_id"
        :card="card"
        :loading="decorating && !card.valuation.available"
        @open="openFund"
        @pre-trade="runPreTrade"
        @view-breaches="viewBreaches"
        @write-research="writeResearch"
      />
    </div>
  </section>
</template>

<style scoped>
.my-funds-page {
  display: grid;
  gap: var(--space-4);
}

.my-funds-page__date {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  background: #ffffff;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  padding: 4px 10px;
  line-height: 1.2;
}

.my-funds-page__date-label {
  font-weight: 700;
  color: #8c959f;
  font-size: 9px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.my-funds-page__date-value {
  font-variant-numeric: tabular-nums;
  color: #0f172a;
  font-size: 13px;
  font-weight: 700;
}

.my-funds-page__notice,
.my-funds-page__empty {
  padding: var(--space-4);
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
  text-align: center;
}

.my-funds-page__error {
  padding: var(--space-4);
  background: var(--bg-danger-soft, rgba(207, 34, 46, 0.06));
  border: 1px solid rgba(207, 34, 46, 0.3);
  border-radius: var(--radius-md, 6px);
  display: grid;
  gap: 6px;
}

.my-funds-page__error-title {
  font-weight: 600;
  color: var(--state-danger, #cf222e);
}

.my-funds-page__error-detail {
  font-size: 12px;
  color: var(--text-secondary, #57606a);
}

.my-funds-page__retry {
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

.my-funds-page__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
  gap: var(--space-4);
}

.my-funds-page__new {
  display: inline-flex;
  align-items: center;
  height: 32px;
  padding: 0 14px;
  background: var(--state-success, #1a7f37);
  color: #fff;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  transition: background-color 0.15s ease, transform 0.05s ease;
}
.my-funds-page__new:hover { background: #1f8e3f; }
.my-funds-page__new:active { transform: translateY(1px); }

@media (max-width: 720px) {
  .my-funds-page__grid {
    grid-template-columns: 1fr;
  }
}
</style>
