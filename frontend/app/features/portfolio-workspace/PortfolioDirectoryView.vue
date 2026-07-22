<script setup lang="ts">
import { onMounted } from "vue";

import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import { useI18n } from "~/composables/useI18n";
import { useMyPortfolios } from "./composables/useMyPortfolios";
import PortfolioKpiStrip from "./components/PortfolioKpiStrip.vue";
import PortfolioFilterToolbar from "./components/PortfolioFilterToolbar.vue";
import PortfolioSummaryCard from "./components/PortfolioSummaryCard.vue";

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
  kpis,
  loadAll,
  setFilter,
  setSort,
  setSearch,
} = useMyPortfolios();

function reload() {
  void loadAll(t("portfolio.page.errorDetail"));
}

onMounted(reload);

function openPortfolio(code: string) {
  void router.push(`/portfolios/${encodeURIComponent(code)}/overview`);
}

function viewBreaches(portfolioCode: string) {
  void router.push(
    `/portfolios/${encodeURIComponent(portfolioCode)}/compliance`,
  );
}

function writeResearch(fundId: string) {
  void router.push({
    path: "/investment/analysis/new",
    query: { fund_id: fundId },
  });
}

function createPortfolio() {
  void router.push("/portfolios/new");
}
</script>

<template>
  <section class="portfolio-directory-page">
    <AppPageHeader
      :title="t('portfolio.directory.title')"
      :description="t('portfolio.directory.subtitle')"
    >
      <template #actions>
        <div class="portfolio-directory-page__date" aria-live="polite">
          <span class="portfolio-directory-page__date-label">{{ t("portfolio.page.asOf") }}</span>
          <span class="portfolio-directory-page__date-value">{{ businessDate }}</span>
        </div>
        <IMSPermissionGuard permission="INVESTMENT_PORTFOLIO_MANAGE">
          <AppButton variant="primary" size="sm" @click="createPortfolio">
            {{ t("portfolio.directory.createPortfolio") }}
          </AppButton>
        </IMSPermissionGuard>
      </template>
    </AppPageHeader>

    <PortfolioKpiStrip :kpis="kpis" :loading="loading && cards.length === 0" />

    <PortfolioFilterToolbar
      :filter="filter"
      :sort="sort"
      :search="search"
      :counts="counts"
      @update:filter="setFilter"
      @update:sort="setSort"
      @update:search="setSearch"
    />

    <div v-if="decorating && !loading" class="portfolio-directory-page__notice" role="status">
      {{ t("portfolio.page.decorating") }}
    </div>

    <div v-if="loading && cards.length === 0" class="portfolio-directory-page__notice" role="status">
      {{ t("portfolio.page.loading") }}
    </div>

    <div v-else-if="error" class="portfolio-directory-page__error" role="alert">
      <div class="portfolio-directory-page__error-title">
        {{ t("portfolio.page.errorTitle") }}
      </div>
      <div class="portfolio-directory-page__error-detail">{{ error }}</div>
      <button type="button" class="portfolio-directory-page__retry" @click="reload">
        {{ t("portfolio.page.retry") }}
      </button>
    </div>

    <div v-else-if="filtered.length === 0" class="portfolio-directory-page__empty">
      {{
        cards.length === 0
          ? t("portfolio.page.empty")
          : t("portfolio.page.noResults")
      }}
    </div>

    <div v-else class="portfolio-directory-page__grid">
      <PortfolioSummaryCard
        v-for="card in filtered"
        :key="card.portfolio_id"
        :card="card"
        :loading="decorating && !card.valuation.available"
        @open="openPortfolio"
        @view-breaches="viewBreaches"
        @write-research="writeResearch"
      />
    </div>
  </section>
</template>

<style scoped>
.portfolio-directory-page {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-directory-page__date {
  display: inline-flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
  background: var(--bg-card, #ffffff);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 6px;
  padding: 4px 10px;
  line-height: 1.2;
}

.portfolio-directory-page__date-label {
  font-weight: 700;
  color: var(--text-secondary, #8c959f);
  font-size: 9px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.portfolio-directory-page__date-value {
  font-variant-numeric: tabular-nums;
  color: var(--text-primary, #0f172a);
  font-size: 13px;
  font-weight: 700;
}

.portfolio-directory-page__notice,
.portfolio-directory-page__empty {
  padding: var(--space-4, 16px);
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
  text-align: center;
}

.portfolio-directory-page__error {
  padding: var(--space-4, 16px);
  background: var(--bg-danger-soft, rgba(207, 34, 46, 0.06));
  border: 1px solid rgba(207, 34, 46, 0.3);
  border-radius: var(--radius-md, 6px);
  display: grid;
  gap: 6px;
}

.portfolio-directory-page__error-title {
  font-weight: 600;
  color: var(--state-danger, #cf222e);
}

.portfolio-directory-page__error-detail {
  font-size: 12px;
  color: var(--text-secondary, #57606a);
}

.portfolio-directory-page__retry {
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

.portfolio-directory-page__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(420px, 1fr));
  gap: var(--space-4, 16px);
}

@media (max-width: 720px) {
  .portfolio-directory-page__grid {
    grid-template-columns: 1fr;
  }
}
</style>
