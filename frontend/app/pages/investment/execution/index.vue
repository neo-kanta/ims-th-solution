<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppCard from "~/shared/ui/AppCard.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

/**
 * /investment/execution — global execution / order ticket index.
 *
 * The backend does not expose an order-management lifecycle. The supported
 * way to record a real transaction today is per-fund: the Operator page's
 * Operation workflow posts to `POST /investment/portfolios/{id}/transactions`,
 * with the pre-trade compliance gate enforced inside the same handler.
 *
 * Rather than fake a global execution surface, we route the user to that
 * real flow.
 */

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_VIEW",
});

const { t } = useI18n();
</script>

<template>
  <section class="execution-page">
    <AppPageHeader
      :title="t(
        'placeholders.executionOrders.title',
        'Execution & orders',
      )"
      :description="t(
        'placeholders.executionOrders.description',
        'Record trades and review execution status.',
      )"
    />

    <AppCard
      :title="t(
        'placeholders.executionOrders.notConfiguredTitle',
        'Execution is handled inside each fund',
      )"
    >
      <div class="execution-page__body">
        <p>
          {{
            t(
              "placeholders.executionOrders.notConfiguredCopy",
              "There is no global order-management endpoint in the backend yet. To post a real BUY or SELL transaction, open the Operator page and use its Operation workflow. The backend re-runs the pre-trade compliance check inside the post handler — a successful response means the trade was both compliant and recorded on the ledger.",
            )
          }}
        </p>
        <div class="execution-page__links">
          <NuxtLink to="/investment/operator" class="execution-page__link">
            {{ t("navigation.operatorPage", "Operator page") }}
          </NuxtLink>
          <NuxtLink to="/compliance/pre-trade" class="execution-page__link">
            {{ t("compliance.preTrade.title", "Pre-trade simulator") }}
          </NuxtLink>
        </div>
      </div>
    </AppCard>
  </section>
</template>

<style scoped>
.execution-page {
  display: grid;
  gap: var(--space-7);
}

.execution-page__body {
  display: grid;
  gap: var(--space-3);
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
  line-height: 1.55;
}

.execution-page__body p {
  margin: 0;
}

.execution-page__links {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
  margin-top: var(--space-2);
}

.execution-page__link {
  color: var(--text-link);
  text-decoration: none;
  font-weight: var(--font-weight-semibold);
}

.execution-page__link:hover {
  text-decoration: underline;
}
</style>
