<script setup lang="ts">
import type { ApiPortfolioV2 } from "~/features/portfolio-workspace/services/portfolioApi";
import { useI18n } from "~/composables/useI18n";
import { formatMoney, formatQuantity } from "../lib/decisionFormat";
import type { EnrichedHolding } from "../composables/usePortfolioHoldingsDirectory";

const props = defineProps<{
  portfolio: ApiPortfolioV2 | null;
  side: "BUY" | "SELL";
  currency: string;
  cashBalance: string | null;
  holding: EnrichedHolding | null;
}>();

const { t } = useI18n();
</script>

<template>
  <div class="context-panel">
    <div v-if="portfolio" class="context-panel__identity">
      <span class="context-panel__code">{{ portfolio.code }}</span>
      <div class="context-panel__name">{{ portfolio.name }}</div>
      <div class="context-panel__meta">
        <span class="context-panel__badge">{{ portfolio.portfolio_type }}</span>
        <span class="context-panel__badge">{{ portfolio.status }}</span>
      </div>
    </div>
    <div v-else class="context-panel__loading">{{ t("portfolio.decisionNew.loadingPortfolio") }}</div>

    <dl v-if="side === 'BUY'" class="context-panel__facts">
      <dt>{{ t("portfolio.decisionNew.availableCash") }}{{ currency ? ` (${currency})` : "" }}</dt>
      <dd v-if="cashBalance !== null">{{ formatMoney(cashBalance, currency) }}</dd>
      <dd v-else class="context-panel__muted">
        {{ currency ? t("portfolio.decisionNew.noCashForCurrency") : t("portfolio.decisionNew.selectCurrencyForCash") }}
      </dd>
    </dl>

    <dl v-else class="context-panel__facts">
      <dt>{{ t("portfolio.decisionNew.availableQuantity") }}</dt>
      <dd v-if="holding">{{ formatQuantity(holding.quantity) }}</dd>
      <dd v-else class="context-panel__muted">{{ t("portfolio.decisionNew.selectOwnedInstrument") }}</dd>

      <template v-if="holding">
        <dt>{{ t("portfolio.decisionNew.averageCost") }}</dt>
        <dd>{{ formatMoney(holding.averageCost, holding.instrument?.currency) }}</dd>
      </template>
    </dl>
  </div>
</template>

<style scoped>
.context-panel {
  display: grid;
  gap: 16px;
}

.context-panel__identity {
  display: grid;
  gap: 4px;
}

.context-panel__code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  font-weight: 700;
  color: var(--text-tertiary, #6e7781);
  width: fit-content;
  padding: 2px 8px;
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: 999px;
}

.context-panel__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.context-panel__meta {
  display: flex;
  gap: 6px;
}

.context-panel__badge {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 2px 7px;
  border-radius: 999px;
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  color: var(--text-secondary);
}

.context-panel__loading {
  font-size: 12px;
  color: var(--text-tertiary);
}

.context-panel__facts {
  margin: 0;
  padding-top: 12px;
  border-top: 1px solid var(--border-subtle, #d0d7de);
  display: grid;
  gap: 2px;
}

.context-panel__facts dt {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
  margin-top: 8px;
}

.context-panel__facts dt:first-child {
  margin-top: 0;
}

.context-panel__facts dd {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}

.context-panel__muted {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-tertiary);
}
</style>
