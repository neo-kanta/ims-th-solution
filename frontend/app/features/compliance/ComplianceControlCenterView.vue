<script setup lang="ts">
import { computed, onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import { useAuthStore } from "~/stores/useAuthStore";
import AppButton from "~/shared/ui/AppButton.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import ComplianceBreachQueue from "./components/ComplianceBreachQueue.vue";
import ComplianceCategoryPanel from "./components/ComplianceCategoryPanel.vue";
import ComplianceKpiCard from "./components/ComplianceKpiCard.vue";
import ComplianceNoRulesEmptyState from "./components/ComplianceNoRulesEmptyState.vue";
import CompliancePortfolioFinder from "./components/CompliancePortfolioFinder.vue";
import { useComplianceBreachesList } from "./composables/useComplianceBreaches";
import { useCompliancePortfolioDirectory } from "./composables/useCompliancePortfolioDirectory";
import { useComplianceRuleDirectory } from "./composables/useComplianceRuleDirectory";
import { deviceLocalIsoDate } from "./lib/asOfDate";
import { sortBreachesForQueue } from "./lib/breachQueue";

const { t } = useI18n();
const authStore = useAuthStore();

const ruleDir = useComplianceRuleDirectory();
const openBreaches = useComplianceBreachesList();
const portfolios = useCompliancePortfolioDirectory();

const canCreateRule = computed(() =>
  authStore.hasPermission("IRG_EDIT_RULE_INSTANCE"),
);
const canViewPortfolios = computed(() =>
  authStore.hasPermission("INVESTMENT_PORTFOLIO_VIEW"),
);
const asOfDate = deviceLocalIsoDate();

const rulesPending = computed(
  () => !ruleDir.loaded.value || ruleDir.loading.value,
);
const breachesPending = computed(
  () => !openBreaches.loaded.value || openBreaches.loading.value,
);
const refreshing = computed(
  () =>
    ruleDir.loading.value ||
    openBreaches.loading.value ||
    portfolios.loading.value,
);

const noRules = computed(
  () =>
    ruleDir.loaded.value &&
    !ruleDir.error.value &&
    ruleDir.total.value === 0,
);

function ruleIndicator(count: number): number | null {
  if (rulesPending.value || ruleDir.error.value) return null;
  return count;
}

const activeRulesCount = computed(() =>
  ruleIndicator(ruleDir.byDerivedStatus.value.get("ACTIVE")?.length ?? 0),
);
const scheduledRulesCount = computed(() =>
  ruleIndicator(ruleDir.byDerivedStatus.value.get("SCHEDULED")?.length ?? 0),
);
const highRiskRulesCount = computed(() =>
  ruleIndicator(ruleDir.highRiskRules.value.length),
);
const openBreachCount = computed(() => {
  if (breachesPending.value || openBreaches.error.value) return null;
  return openBreaches.total.value;
});

const sortedOpenBreaches = computed(() =>
  sortBreachesForQueue(openBreaches.items.value),
);

function loadBreaches() {
  return openBreaches.fetchList({ status: "OPEN", limit: 50, offset: 0 });
}

function loadPortfolios() {
  if (!canViewPortfolios.value) return Promise.resolve();
  return portfolios.refresh();
}

async function refreshDashboard() {
  await Promise.allSettled([
    ruleDir.refresh(),
    loadBreaches(),
    loadPortfolios(),
  ]);
}

onMounted(() => {
  void Promise.allSettled([
    ruleDir.ensureLoaded(),
    loadBreaches(),
    canViewPortfolios.value
      ? portfolios.ensureLoaded()
      : Promise.resolve(),
  ]);
});
</script>

<template>
  <section class="control-center">
    <AppPageHeader
      :title="t('compliance.dashboard.title')"
      :description="t('compliance.dashboard.description')"
    >
      <template #actions>
        <span class="control-center__as-of">
          {{ t("compliance.dashboard.asOfLabel", { date: asOfDate }) }}
        </span>
        <AppButton
          variant="secondary"
          size="sm"
          :loading="refreshing"
          @click="refreshDashboard"
        >
          {{ t("compliance.dashboard.headerActions.refresh") }}
        </AppButton>
        <NuxtLink to="/compliance/post-trade">
          <AppButton variant="secondary" size="sm">
            {{ t("compliance.dashboard.headerActions.openBreaches") }}
          </AppButton>
        </NuxtLink>
        <NuxtLink v-if="canCreateRule" to="/compliance/rules/new">
          <AppButton variant="primary" size="sm">
            {{ t("compliance.dashboard.headerActions.newRule") }}
          </AppButton>
        </NuxtLink>
      </template>
    </AppPageHeader>

    <section
      class="control-center__indicators"
      :aria-label="t('compliance.dashboard.indicators.groupLabel')"
    >
      <ComplianceKpiCard
        icon="check"
        :label="t('compliance.dashboard.kpi.activeRules')"
        :value="activeRulesCount"
        :subtitle="t('compliance.dashboard.kpi.activeRulesSub')"
        tone="success"
        :loading="rulesPending"
      />
      <ComplianceKpiCard
        icon="pending"
        :label="t('compliance.dashboard.kpi.scheduledRules')"
        :value="scheduledRulesCount"
        :subtitle="t('compliance.dashboard.kpi.scheduledRulesSub')"
        tone="default"
        :loading="rulesPending"
      />
      <ComplianceKpiCard
        icon="warning"
        :label="t('compliance.dashboard.kpi.openBreaches')"
        :value="openBreachCount"
        :subtitle="t('compliance.dashboard.kpi.openBreachesSub')"
        tone="danger"
        :loading="breachesPending"
      />
      <ComplianceKpiCard
        icon="review"
        :label="t('compliance.dashboard.kpi.highRisk')"
        :value="highRiskRulesCount"
        :subtitle="t('compliance.dashboard.kpi.highRiskSub')"
        tone="warning"
        :loading="rulesPending"
      />
    </section>

    <div v-if="ruleDir.error.value" class="control-center__error" role="alert">
      <strong>{{ t("compliance.dashboard.errors.rulesTitle") }}</strong>
      {{ ruleDir.error.value }}
    </div>

    <ComplianceNoRulesEmptyState v-if="noRules" />

    <div class="control-center__workspace">
      <main class="control-center__primary">
        <ComplianceBreachQueue
          :items="sortedOpenBreaches"
          :total="openBreaches.total.value"
          :loading="breachesPending"
          :error="openBreaches.error.value"
          :portfolio-by-id="portfolios.byId.value"
          @retry="loadBreaches"
        />
      </main>

      <aside class="control-center__rail">
        <CompliancePortfolioFinder v-if="canViewPortfolios" />
        <section v-else class="control-center__permission-note" role="note">
          <strong>{{ t("compliance.dashboard.finder.title") }}</strong>
          <span>{{ t("compliance.dashboard.finder.permissionUnavailable") }}</span>
        </section>

        <ComplianceCategoryPanel
          :by-category="ruleDir.byCategory.value"
          :loading="rulesPending"
        />

        <NuxtLink class="control-center__library-link" to="/compliance/rules">
          <AppButton variant="secondary" size="sm" full-width>
            {{ t("compliance.dashboard.goToRules") }}
          </AppButton>
        </NuxtLink>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.control-center {
  display: grid;
  gap: var(--space-6);
}

.control-center__as-of {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 0 var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  white-space: nowrap;
}

.control-center__indicators {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-4);
}

.control-center__workspace {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(18rem, 0.8fr);
  gap: var(--space-5);
  align-items: start;
}

.control-center__primary,
.control-center__rail {
  min-width: 0;
}

.control-center__rail {
  display: grid;
  gap: var(--space-4);
}

.control-center__error,
.control-center__permission-note {
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
}

.control-center__error {
  border: 1px solid var(--alert-danger-border);
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
}

.control-center__permission-note {
  display: grid;
  gap: var(--space-2);
  border: 1px solid var(--border-subtle);
  background: var(--bg-card-muted);
  color: var(--text-secondary);
}

.control-center__permission-note strong {
  color: var(--text-primary);
}

.control-center__library-link {
  display: block;
}

@media (max-width: 1120px) {
  .control-center__indicators {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .control-center__workspace {
    grid-template-columns: minmax(0, 1fr) minmax(16rem, 0.65fr);
  }
}

@media (max-width: 840px) {
  .control-center__workspace {
    grid-template-columns: 1fr;
  }

  .control-center__rail {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .control-center__library-link {
    grid-column: 1 / -1;
  }
}

@media (max-width: 560px) {
  .control-center__indicators,
  .control-center__rail {
    grid-template-columns: 1fr;
  }

  .control-center__library-link {
    grid-column: auto;
  }

  .control-center__as-of {
    width: 100%;
    white-space: normal;
  }
}
</style>
