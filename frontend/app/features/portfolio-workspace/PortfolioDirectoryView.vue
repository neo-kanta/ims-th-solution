<script setup lang="ts">
import { onMounted } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import { useI18n } from "~/composables/useI18n";
import { usePortfolioDirectory } from "~/features/investment-ledger/composables/usePortfolioDirectory";

const { t } = useI18n();
const { portfolios, loading, error, load } = usePortfolioDirectory();

onMounted(() => {
  void load();
});

function portfolioRoute(code: string): string {
  return `/portfolios/${encodeURIComponent(code)}/overview`;
}
</script>

<template>
  <section class="portfolio-directory">
    <AppPageHeader
      :title="t('portfolio.directory.title')"
      :description="t('portfolio.directory.subtitle')"
    />

    <div v-if="loading && portfolios.length === 0" class="portfolio-directory__notice" role="status">
      {{ t("portfolio.directory.loading") }}
    </div>

    <div v-else-if="error" class="portfolio-directory__error" role="alert">
      <div class="portfolio-directory__error-title">{{ t("portfolio.directory.errorTitle") }}</div>
      <div class="portfolio-directory__error-detail">{{ error }}</div>
      <button type="button" class="portfolio-directory__retry" @click="load">
        {{ t("portfolio.directory.retry") }}
      </button>
    </div>

    <div v-else-if="portfolios.length === 0" class="portfolio-directory__notice">
      {{ t("portfolio.directory.empty") }}
    </div>

    <AppCard v-else :flush="true">
      <table class="portfolio-directory__table">
        <thead>
          <tr>
            <th>{{ t("portfolio.directory.columns.code") }}</th>
            <th>{{ t("portfolio.directory.columns.name") }}</th>
            <th>{{ t("portfolio.directory.columns.status") }}</th>
            <th>{{ t("portfolio.directory.columns.currency") }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in portfolios" :key="p.id">
            <td class="portfolio-directory__code">{{ p.code }}</td>
            <td>{{ p.name }}</td>
            <td>{{ p.status }}</td>
            <td>{{ p.valuation_currency || p.base_currency }}</td>
            <td class="portfolio-directory__actions">
              <NuxtLink
                v-if="p.code"
                :to="portfolioRoute(p.code)"
                class="portfolio-directory__open"
              >
                {{ t("portfolio.directory.open") }}
              </NuxtLink>
            </td>
          </tr>
        </tbody>
      </table>
    </AppCard>
  </section>
</template>

<style scoped>
.portfolio-directory {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-directory__notice,
.portfolio-directory__error {
  padding: var(--space-4, 16px);
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-directory__error {
  background: var(--alert-danger-bg);
  border-color: var(--alert-danger-border);
  display: grid;
  gap: 6px;
}

.portfolio-directory__error-title {
  font-weight: 600;
  color: var(--alert-danger-text, #cf222e);
}

.portfolio-directory__retry {
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

.portfolio-directory__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.portfolio-directory__table th {
  text-align: left;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.portfolio-directory__table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.portfolio-directory__code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-weight: 600;
}

.portfolio-directory__actions {
  text-align: right;
}

.portfolio-directory__open {
  font-size: 12px;
  font-weight: 600;
  color: var(--state-info, #0969da);
  text-decoration: none;
}
.portfolio-directory__open:hover {
  text-decoration: underline;
}
</style>
