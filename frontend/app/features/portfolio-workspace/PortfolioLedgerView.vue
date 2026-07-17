<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import AppCard from "~/shared/ui/AppCard.vue";
import { useI18n } from "~/composables/useI18n";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import { portfolioApi, type ApiTransactionV2 } from "./services/portfolioApi";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.workspaceTabs.ledger", "Ledger"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true }
);
const ctx = usePortfolioContext(() => props.portfolioCode);

const transactions = ref<ApiTransactionV2[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

async function loadTransactions() {
  if (!props.portfolioCode) return;
  loading.value = true;
  error.value = null;
  try {
    const result = await portfolioApi.listTransactions(props.portfolioCode, { limit: 50 });
    transactions.value = result.items ?? [];
  } catch (err) {
    transactions.value = [];
    error.value = err instanceof Error ? err.message : "Failed to load transactions.";
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void ctx.reload();
  void loadTransactions();
});
watch(
  () => props.portfolioCode,
  () => {
    void ctx.reload();
    void loadTransactions();
  },
);
</script>

<template>
  <section class="portfolio-ledger">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <AppCard :title="t('portfolio.ledger.title')" :subtitle="t('portfolio.ledger.subtitle')">
      <div v-if="loading && transactions.length === 0" class="portfolio-ledger__notice" role="status">
        {{ t("portfolio.ledger.loading") }}
      </div>
      <div v-else-if="error" class="portfolio-ledger__error" role="alert">
        <div>{{ t("portfolio.ledger.errorTitle") }}: {{ error }}</div>
        <button type="button" class="portfolio-ledger__retry" @click="loadTransactions">
          {{ t("portfolio.ledger.retry") }}
        </button>
      </div>
      <div v-else-if="transactions.length === 0" class="portfolio-ledger__notice">
        {{ t("portfolio.ledger.empty") }}
      </div>
      <table v-else class="portfolio-ledger__table">
        <thead>
          <tr>
            <th>{{ t("portfolio.ledger.columns.date") }}</th>
            <th>{{ t("portfolio.ledger.columns.type") }}</th>
            <th>{{ t("portfolio.ledger.columns.instrument") }}</th>
            <th>{{ t("portfolio.ledger.columns.quantity") }}</th>
            <th>{{ t("portfolio.ledger.columns.price") }}</th>
            <th>{{ t("portfolio.ledger.columns.amount") }}</th>
            <th>{{ t("portfolio.ledger.columns.status") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="txn in transactions" :key="txn.id">
            <td>{{ txn.business_date }}</td>
            <td>{{ txn.transaction_type }}</td>
            <td class="portfolio-ledger__mono">{{ txn.instrument_id || "—" }}</td>
            <td>{{ txn.quantity || "—" }}</td>
            <td>{{ txn.price || "—" }}</td>
            <td>{{ txn.net_amount }}</td>
            <td>{{ txn.status }}</td>
          </tr>
        </tbody>
      </table>
    </AppCard>
  </section>
</template>

<style scoped>
.portfolio-ledger {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-ledger__notice,
.portfolio-ledger__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-ledger__error {
  color: var(--alert-danger-text, #cf222e);
  display: grid;
  gap: 6px;
}

.portfolio-ledger__retry {
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

.portfolio-ledger__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.portfolio-ledger__table th {
  text-align: left;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.portfolio-ledger__table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.portfolio-ledger__mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
}
</style>
