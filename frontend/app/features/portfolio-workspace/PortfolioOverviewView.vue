<script setup lang="ts">
import { onMounted, watch } from "vue";

import AppCard from "~/shared/ui/AppCard.vue";
import { useI18n } from "~/composables/useI18n";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();
const ctx = usePortfolioContext(() => props.portfolioCode);

onMounted(() => void ctx.reload());
watch(() => props.portfolioCode, () => void ctx.reload());
</script>

<template>
  <section class="portfolio-overview">
    <div v-if="ctx.loading.value && !ctx.portfolio.value" class="portfolio-overview__notice" role="status">
      {{ t("portfolio.overview.loading") }}
    </div>

    <div v-else-if="ctx.error.value" class="portfolio-overview__error" role="alert">
      <div class="portfolio-overview__error-title">{{ t("portfolio.overview.errorTitle") }}</div>
      <div>{{ ctx.error.value }}</div>
      <button type="button" class="portfolio-overview__retry" @click="ctx.reload">
        {{ t("portfolio.overview.retry") }}
      </button>
    </div>

    <template v-else-if="ctx.portfolio.value">
      <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

      <AppCard :title="t('portfolio.overview.title')">
        <dl class="portfolio-overview__grid">
          <div class="portfolio-overview__row">
            <dt>{{ t("portfolio.overview.code") }}</dt>
            <dd>{{ ctx.portfolio.value.code }}</dd>
          </div>
          <div class="portfolio-overview__row">
            <dt>{{ t("portfolio.overview.status") }}</dt>
            <dd>{{ ctx.portfolio.value.status }}</dd>
          </div>
          <div class="portfolio-overview__row">
            <dt>{{ t("portfolio.overview.baseCurrency") }}</dt>
            <dd>{{ ctx.portfolio.value.base_currency }}</dd>
          </div>
          <div class="portfolio-overview__row">
            <dt>{{ t("portfolio.overview.valuationCurrency") }}</dt>
            <dd>{{ ctx.portfolio.value.valuation_currency }}</dd>
          </div>
          <div v-if="ctx.portfolio.value.benchmark" class="portfolio-overview__row">
            <dt>{{ t("portfolio.overview.benchmark") }}</dt>
            <dd>{{ ctx.portfolio.value.benchmark }}</dd>
          </div>
          <div v-if="ctx.portfolio.value.risk_profile" class="portfolio-overview__row">
            <dt>{{ t("portfolio.overview.riskProfile") }}</dt>
            <dd>{{ ctx.portfolio.value.risk_profile }}</dd>
          </div>
        </dl>
      </AppCard>
    </template>

    <div v-else class="portfolio-overview__notice">
      {{ t("portfolio.overview.notFound") }}
    </div>
  </section>
</template>

<style scoped>
.portfolio-overview {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-overview__notice,
.portfolio-overview__error {
  padding: var(--space-4, 16px);
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-overview__error {
  background: var(--alert-danger-bg);
  border-color: var(--alert-danger-border);
  display: grid;
  gap: 6px;
}

.portfolio-overview__error-title {
  font-weight: 600;
  color: var(--alert-danger-text, #cf222e);
}

.portfolio-overview__retry {
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

.portfolio-overview__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: var(--space-3, 12px);
  margin: 0;
}

.portfolio-overview__row {
  display: grid;
  gap: 2px;
}

.portfolio-overview__row dt {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
}

.portfolio-overview__row dd {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
</style>
