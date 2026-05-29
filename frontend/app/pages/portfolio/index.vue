<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import CashBalancesCard from "~/features/investment-ledger/components/CashBalancesCard.vue";
import HoldingsTable from "~/features/investment-ledger/components/HoldingsTable.vue";
import OrderTicketDrawer from "~/features/investment-ledger/components/OrderTicketDrawer.vue";
import TransactionLedgerTable from "~/features/investment-ledger/components/TransactionLedgerTable.vue";
import { useInstrumentDirectory } from "~/features/investment-ledger/composables/useInstrumentDirectory";
import { useOrderTicket } from "~/features/investment-ledger/composables/useOrderTicket";
import { usePortfolioDirectory } from "~/features/investment-ledger/composables/usePortfolioDirectory";
import { usePortfolioLedger } from "~/features/investment-ledger/composables/usePortfolioLedger";
import { shortenId, todayIso } from "~/features/investment-ledger/lib/ledgerFormat";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_PORTFOLIO_VIEW",
});

const authStore = useAuthStore();
const canSimulate = computed(() => authStore.hasPermission("INVESTMENT_LEDGER_SIMULATE"));
const canPost = computed(() => authStore.hasPermission("INVESTMENT_LEDGER_POST"));

const portfolios = usePortfolioDirectory();
const ledger = usePortfolioLedger();
const instruments = useInstrumentDirectory();
const ticket = useOrderTicket();

const activeTab = ref<"holdings" | "ledger">("holdings");
const showTicket = ref(false);

const activePortfolioId = computed(() => portfolios.activePortfolioId.value);
const activePortfolio = computed(() => portfolios.activePortfolio.value);

const summaryCash = computed(() => ledger.cash.value);

onMounted(async () => {
  // The watcher below picks up the activePortfolioId change inside load()
  // and triggers refreshAll on its own — don't double-fetch here.
  await portfolios.load();
  // Kick instrument directory load in parallel — the ledger tables enrich
  // ticker labels from this list, and the ticket reuses it instantly.
  void instruments.ensureLoaded();
});

watch(activePortfolioId, async (next) => {
  if (!next) {
    ledger.reset();
    return;
  }
  await ledger.refreshAll(next);
});

function onPortfolioChange(event: Event) {
  const id = (event.target as HTMLSelectElement).value;
  if (id) portfolios.setActive(id);
}

function openTicket() {
  if (!activePortfolioId.value || !activePortfolio.value) return;
  ticket.open(activePortfolioId.value, {
    currency: activePortfolio.value.valuation_currency ?? activePortfolio.value.base_currency,
    business_date: todayIso(),
  });
  showTicket.value = true;
}

function closeTicket() {
  showTicket.value = false;
  ticket.close();
}

async function onTransactionPosted() {
  if (!activePortfolioId.value) return;
  // Refresh ledger sections immediately so the operator sees the new state
  // before they decide whether to dismiss the drawer.
  await ledger.refreshAll(activePortfolioId.value);
  activeTab.value = "ledger";
}

async function refreshLedger() {
  if (!activePortfolioId.value) return;
  await ledger.refreshAll(activePortfolioId.value);
}
</script>

<template>
  <section class="portfolio-page">
    <AppPageHeader
      title="Portfolio operations"
      description="Live holdings, cash, and transaction ledger backed by the investment service. Use Simulate to preview pre-trade compliance, cash impact, and position changes before posting."
    >
      <template #actions>
        <select
          v-if="portfolios.portfolios.value.length > 0"
          class="portfolio-page__select"
          :value="activePortfolioId ?? ''"
          aria-label="Select portfolio"
          @change="onPortfolioChange"
        >
          <option
            v-for="p in portfolios.portfolios.value"
            :key="p.id"
            :value="p.id ?? ''"
          >
            {{ p.code ?? shortenId(p.id) }}<span v-if="p.name"> — {{ p.name }}</span>
            <span v-if="p.base_currency"> · {{ p.base_currency }}</span>
          </option>
        </select>
        <AppButton variant="secondary" size="sm" :disabled="!activePortfolioId" @click="refreshLedger">
          Refresh
        </AppButton>
        <AppButton
          variant="primary"
          size="sm"
          :disabled="!activePortfolioId || !canSimulate"
          @click="openTicket"
        >
          New order
        </AppButton>
      </template>
    </AppPageHeader>

    <div v-if="portfolios.loading.value" class="portfolio-page__notice">
      Loading portfolios…
    </div>

    <div v-else-if="portfolios.error.value" class="portfolio-page__notice portfolio-page__notice--error" role="alert">
      {{ portfolios.error.value }}
    </div>

    <div v-else-if="portfolios.portfolios.value.length === 0" class="portfolio-page__notice">
      No portfolios are visible with your current permissions.
    </div>

    <template v-else>
      <div v-if="!canSimulate" class="portfolio-page__notice portfolio-page__notice--warn">
        You can view this portfolio but cannot simulate or post orders. Required permissions:
        <code>INVESTMENT_LEDGER_SIMULATE</code> and <code>INVESTMENT_LEDGER_POST</code>.
      </div>

      <div class="portfolio-page__grid">
        <AppCard
          v-if="activePortfolio"
          :title="activePortfolio.code ?? 'Portfolio'"
          :subtitle="`${activePortfolio.name ?? 'Unnamed portfolio'} · ${activePortfolio.base_currency ?? '—'} · ${activePortfolio.status ?? '—'}`"
        >
          <dl class="portfolio-page__meta">
            <div>
              <dt>Strategy</dt>
              <dd>{{ activePortfolio.strategy_code ?? "—" }}</dd>
            </div>
            <div>
              <dt>Tax lot method</dt>
              <dd>{{ activePortfolio.tax_lot_method ?? "—" }}</dd>
            </div>
            <div>
              <dt>Valuation ccy</dt>
              <dd>{{ activePortfolio.valuation_currency ?? activePortfolio.base_currency ?? "—" }}</dd>
            </div>
            <div>
              <dt>Inception</dt>
              <dd>{{ activePortfolio.inception_date ?? "—" }}</dd>
            </div>
            <div>
              <dt>Holdings</dt>
              <dd>{{ ledger.holdings.value.length }}</dd>
            </div>
            <div>
              <dt>Transactions</dt>
              <dd>{{ ledger.transactionsTotal.value }}</dd>
            </div>
          </dl>
        </AppCard>

        <AppCard title="Cash on hand" subtitle="Balances by settlement currency">
          <CashBalancesCard
            :balances="summaryCash"
            :loading="ledger.loadingCash.value"
            :error="ledger.cashError.value"
          />
        </AppCard>
      </div>

      <AppCard
        :title="activeTab === 'holdings' ? 'Holdings' : 'Transaction ledger'"
        :subtitle="
          activeTab === 'holdings'
            ? 'Current positions and average cost basis.'
            : 'Most recent posted transactions for this portfolio.'
        "
      >
        <template #header-actions>
          <div class="portfolio-page__tabs" role="tablist">
            <button
              type="button"
              class="portfolio-page__tab"
              :class="{ 'portfolio-page__tab--active': activeTab === 'holdings' }"
              role="tab"
              :aria-selected="activeTab === 'holdings'"
              @click="activeTab = 'holdings'"
            >
              Holdings
              <span class="portfolio-page__tab-count">{{ ledger.holdings.value.length }}</span>
            </button>
            <button
              type="button"
              class="portfolio-page__tab"
              :class="{ 'portfolio-page__tab--active': activeTab === 'ledger' }"
              role="tab"
              :aria-selected="activeTab === 'ledger'"
              @click="activeTab = 'ledger'"
            >
              Ledger
              <span class="portfolio-page__tab-count">{{ ledger.transactionsTotal.value }}</span>
            </button>
          </div>
        </template>

        <HoldingsTable
          v-if="activeTab === 'holdings'"
          :holdings="ledger.holdings.value"
          :instruments="instruments.items.value"
          :base-currency="activePortfolio?.base_currency ?? ''"
          :loading="ledger.loadingHoldings.value"
          :error="ledger.holdingsError.value"
        />
        <TransactionLedgerTable
          v-else
          :transactions="ledger.transactions.value"
          :instruments="instruments.items.value"
          :loading="ledger.loadingTransactions.value"
          :error="ledger.transactionsError.value"
        />
      </AppCard>
    </template>

    <OrderTicketDrawer
      :open="showTicket"
      :portfolio="activePortfolio"
      :ticket="ticket"
      :can-simulate="canSimulate"
      :can-post="canPost"
      @close="closeTicket"
      @posted="onTransactionPosted"
    />
  </section>
</template>

<style scoped>
.portfolio-page {
  display: grid;
  gap: var(--space-4);
}

.portfolio-page__select {
  height: 36px;
  min-width: 220px;
  padding: 0 var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  background: var(--bg-input);
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.portfolio-page__select:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.portfolio-page__notice {
  padding: var(--space-3) var(--space-4);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card-muted);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.portfolio-page__notice--error {
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border-color: var(--alert-danger-border);
}

.portfolio-page__notice--warn {
  background: var(--alert-warning-bg);
  color: var(--alert-warning-text);
  border-color: var(--alert-warning-border);
}

.portfolio-page__notice code {
  font-family: var(--font-family-mono);
  font-size: 0.85em;
  padding: 1px 4px;
  background: rgba(0, 0, 0, 0.04);
  border-radius: 3px;
}

.portfolio-page__grid {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr);
  gap: var(--space-4);
}

.portfolio-page__meta {
  margin: 0;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: var(--space-2);
}

.portfolio-page__meta > div {
  display: grid;
  gap: 2px;
  padding: var(--space-2) var(--space-3);
  background: var(--bg-card-muted);
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-subtle);
}

.portfolio-page__meta dt {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
}

.portfolio-page__meta dd {
  margin: 0;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-primary);
  font-variant-numeric: tabular-nums;
}

.portfolio-page__tabs {
  display: inline-flex;
  background: var(--bg-card-muted);
  padding: 3px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
}

.portfolio-page__tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border: none;
  background: transparent;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--text-secondary);
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.portfolio-page__tab--active {
  background: var(--bg-card);
  color: var(--text-primary);
  box-shadow: var(--shadow-xs);
}

.portfolio-page__tab-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 22px;
  height: 18px;
  padding: 0 6px;
  font-size: 10px;
  font-variant-numeric: tabular-nums;
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 9px;
  color: var(--text-tertiary);
}

@media (max-width: 960px) {
  .portfolio-page__grid {
    grid-template-columns: 1fr;
  }
}
</style>
