<script setup lang="ts">
import { onMounted, ref, watch } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import { useI18n } from "~/composables/useI18n";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import { portfolioApi, type ApiHoldingV2 } from "./services/portfolioApi";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();
const ctx = usePortfolioContext(() => props.portfolioCode);

const holdings = ref<ApiHoldingV2[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

async function loadHoldings() {
  if (!props.portfolioCode) return;
  loading.value = true;
  error.value = null;
  try {
    holdings.value = await portfolioApi.getHoldings(props.portfolioCode);
  } catch (err) {
    holdings.value = [];
    error.value = err instanceof Error ? err.message : "Failed to load holdings.";
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void ctx.reload();
  void loadHoldings();
});
watch(
  () => props.portfolioCode,
  () => {
    void ctx.reload();
    void loadHoldings();
  },
);
</script>

<template>
  <section class="portfolio-holdings">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <AppCard :title="t('portfolio.holdings.title')" :subtitle="t('portfolio.holdings.subtitle')">
      <div v-if="loading && holdings.length === 0" class="portfolio-holdings__notice" role="status">
        {{ t("portfolio.holdings.loading") }}
      </div>
      <div v-else-if="error" class="portfolio-holdings__error" role="alert">
        <div>{{ t("portfolio.holdings.errorTitle") }}: {{ error }}</div>
        <button type="button" class="portfolio-holdings__retry" @click="loadHoldings">
          {{ t("portfolio.holdings.retry") }}
        </button>
      </div>
      <div v-else-if="holdings.length === 0" class="portfolio-holdings__notice">
        {{ t("portfolio.holdings.empty") }}
      </div>
      <table v-else class="portfolio-holdings__table">
        <thead>
          <tr>
            <th>{{ t("portfolio.holdings.columns.instrument") }}</th>
            <th>{{ t("portfolio.holdings.columns.quantity") }}</th>
            <th>{{ t("portfolio.holdings.columns.avgCost") }}</th>
            <th>{{ t("portfolio.holdings.columns.costBasis") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="h in holdings" :key="h.instrument_id">
            <td class="portfolio-holdings__mono">{{ h.instrument_id }}</td>
            <td>{{ h.quantity }}</td>
            <td>{{ h.average_cost }}</td>
            <td>{{ h.cost_basis }}</td>
          </tr>
        </tbody>
      </table>
    </AppCard>
  </section>
</template>

<style scoped>
.portfolio-holdings {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-holdings__notice,
.portfolio-holdings__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-holdings__error {
  color: var(--alert-danger-text, #cf222e);
  display: grid;
  gap: 6px;
}

.portfolio-holdings__retry {
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

.portfolio-holdings__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.portfolio-holdings__table th {
  text-align: left;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.portfolio-holdings__table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.portfolio-holdings__mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}
</style>
