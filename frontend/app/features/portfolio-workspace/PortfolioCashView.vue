<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import AppCard from "~/shared/ui/AppCard.vue";
import { useI18n } from "~/composables/useI18n";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import { portfolioApi, type ApiCashBalanceV2 } from "./services/portfolioApi";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.workspaceTabs.cash"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true }
);
const ctx = usePortfolioContext(() => props.portfolioCode);

const balances = ref<ApiCashBalanceV2[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

async function loadCash() {
  if (!props.portfolioCode) return;
  loading.value = true;
  error.value = null;
  try {
    balances.value = await portfolioApi.getCash(props.portfolioCode);
  } catch (err) {
    balances.value = [];
    error.value = err instanceof Error ? err.message : "Failed to load cash balances.";
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void ctx.reload();
  void loadCash();
});
watch(
  () => props.portfolioCode,
  () => {
    void ctx.reload();
    void loadCash();
  },
);
</script>

<template>
  <section class="portfolio-cash">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <AppCard :title="t('portfolio.cash.title')" :subtitle="t('portfolio.cash.subtitle')">
      <div v-if="loading && balances.length === 0" class="portfolio-cash__notice" role="status">
        {{ t("portfolio.cash.loading") }}
      </div>
      <div v-else-if="error" class="portfolio-cash__error" role="alert">
        <div>{{ t("portfolio.cash.errorTitle") }}: {{ error }}</div>
        <button type="button" class="portfolio-cash__retry" @click="loadCash">
          {{ t("portfolio.cash.retry") }}
        </button>
      </div>
      <div v-else-if="balances.length === 0" class="portfolio-cash__notice">
        {{ t("portfolio.cash.empty") }}
      </div>
      <table v-else class="portfolio-cash__table">
        <thead>
          <tr>
            <th>{{ t("portfolio.cash.columns.currency") }}</th>
            <th>{{ t("portfolio.cash.columns.balance") }}</th>
            <th>{{ t("portfolio.cash.columns.lastBusinessDate") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in balances" :key="b.currency">
            <td>{{ b.currency }}</td>
            <td>{{ b.balance }}</td>
            <td>{{ b.last_business_date }}</td>
          </tr>
        </tbody>
      </table>
    </AppCard>
  </section>
</template>

<style scoped>
.portfolio-cash {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-cash__notice,
.portfolio-cash__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-cash__error {
  color: var(--alert-danger-text, #cf222e);
  display: grid;
  gap: 6px;
}

.portfolio-cash__retry {
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

.portfolio-cash__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.portfolio-cash__table th {
  text-align: left;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.portfolio-cash__table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}
</style>
