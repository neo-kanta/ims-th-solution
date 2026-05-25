<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import ComplianceCheckResultPanel from "~/features/compliance/components/ComplianceCheckResultPanel.vue";
import ComplianceNoRulesEmptyState from "~/features/compliance/components/ComplianceNoRulesEmptyState.vue";
import ComplianceTestPanel from "~/features/compliance/components/ComplianceTestPanel.vue";
import { useComplianceChecks } from "~/features/compliance/composables/useComplianceChecks";
import { useComplianceRulesList } from "~/features/compliance/composables/useComplianceRules";
import { ruleLabel } from "~/features/compliance/lib/ruleTypeCatalog";
import type { CompliancePreTradeRequest } from "~/features/compliance/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_VIEW_RULES",
});

const { t } = useI18n();
const route = useRoute();
const router = useRouter();

const checks = useComplianceChecks();
const rules = useComplianceRulesList();

// Optional context from /compliance/rules/[id] → "Open in simulator".
const prefilledRuleTypeId = computed(() =>
  typeof route.query.rule_type_id === "string"
    ? route.query.rule_type_id
    : null,
);
const prefilledRuleLabel = computed(() =>
  prefilledRuleTypeId.value ? ruleLabel(prefilledRuleTypeId.value) : null,
);

const hasNoActiveRules = computed(
  () =>
    !rules.loading.value &&
    !rules.error.value &&
    rules.total.value === 0,
);

const checkPanelEl = ref<HTMLElement | null>(null);

function scrollToResult() {
  // Use requestAnimationFrame so the v-if has settled before scrolling.
  if (typeof window === "undefined") return;
  window.requestAnimationFrame(() => {
    checkPanelEl.value?.scrollIntoView({ behavior: "smooth", block: "start" });
  });
}

async function handleSubmit(payload: CompliancePreTradeRequest) {
  try {
    await checks.runPreTrade(payload);
  } catch {
    /* error surfaced via composable */
  }
  scrollToResult();
}

async function handleRecheck() {
  if (!checks.lastRequest.value) return;
  try {
    await checks.runPreTrade(checks.lastRequest.value);
  } catch {
    /* surfaced via composable */
  }
  scrollToResult();
}

function handleViewRule(ruleTypeId: string) {
  void router.push({
    path: "/compliance/rules",
    query: { highlight: ruleTypeId },
  });
}

onMounted(() => {
  rules.limit.value = 1;
  void rules.fetchList({ limit: 1, is_active: true });
});
</script>

<template>
  <section class="pretrade-page">
    <AppPageHeader
      :title="t('compliance.preTrade.title')"
      :description="t('compliance.preTrade.description')"
    />

    <ComplianceNoRulesEmptyState v-if="hasNoActiveRules" />

    <div
      v-if="prefilledRuleTypeId"
      class="pretrade-page__rule-hint"
      role="status"
    >
      <span>
        Exercising rule:
        <strong>{{ prefilledRuleLabel }}</strong>
        <code>{{ prefilledRuleTypeId }}</code>
      </span>
      <NuxtLink to="/compliance/pre-trade" class="pretrade-page__clear">
        Clear
      </NuxtLink>
    </div>

    <ComplianceTestPanel
      :loading="checks.loading.value"
      @submit="handleSubmit"
    />

    <div
      v-if="checks.error.value"
      class="pretrade-page__error"
      role="alert"
    >
      <strong>{{ t("compliance.common.backendError") }}</strong>
      <span>{{ checks.error.value }}</span>
    </div>

    <div
      v-if="checks.result.value"
      ref="checkPanelEl"
      class="pretrade-page__result"
    >
      <ComplianceCheckResultPanel
        :result="checks.result.value"
        :loading="checks.loading.value"
        :hydrated-breaches="checks.breachById.value"
        :hydration-error="checks.hydrationError.value"
        @recheck="handleRecheck"
        @view-rule="handleViewRule"
      />
    </div>
  </section>
</template>

<style scoped>
.pretrade-page {
  display: grid;
  gap: var(--space-7);
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.pretrade-page__error {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-4) var(--space-5);
  border: 1px solid var(--alert-danger-border);
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border-radius: var(--radius-md);
}

.pretrade-page__result {
  scroll-margin-top: var(--space-8);
}

.pretrade-page__rule-hint {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--alert-info-bg);
  color: var(--alert-info-text);
  border: 1px solid var(--alert-info-border);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
}

.pretrade-page__rule-hint code {
  margin-left: var(--space-2);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-xs);
  opacity: 0.7;
}

.pretrade-page__clear {
  font-size: var(--font-size-xs);
  color: var(--text-link);
  text-decoration: none;
}

.pretrade-page__clear:hover {
  text-decoration: underline;
}
</style>
