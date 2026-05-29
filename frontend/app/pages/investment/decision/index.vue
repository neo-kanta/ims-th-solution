<script setup lang="ts">
import { computed, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import ComplianceCheckResultPanel from "~/features/compliance/components/ComplianceCheckResultPanel.vue";
import ComplianceTestPanel from "~/features/compliance/components/ComplianceTestPanel.vue";
import { useComplianceChecks } from "~/features/compliance/composables/useComplianceChecks";
import type { CompliancePreTradeRequest } from "~/features/compliance/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_VIEW",
});

const { t } = useI18n();
const router = useRouter();

const checks = useComplianceChecks();
const checkPanelEl = ref<HTMLElement | null>(null);

const verdict = computed(() => checks.result.value?.verdict ?? null);
const isPass = computed(() => verdict.value === "PASS");
const isWarn = computed(() => verdict.value === "WARN");
const isBlock = computed(() => verdict.value === "BLOCK");

function scrollToResult() {
  if (typeof window === "undefined") return;
  window.requestAnimationFrame(() => {
    checkPanelEl.value?.scrollIntoView({ behavior: "smooth", block: "start" });
  });
}

async function handleSubmit(payload: CompliancePreTradeRequest) {
  try {
    await checks.runPreTrade(payload);
  } catch {
    /* surfaced via composable */
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

function handleViewMissingApi() {
  void router.push("/compliance#compliance-missing-api");
}
</script>

<template>
  <section class="decision-page">
    <AppPageHeader
      :title="t('compliance.decision.title')"
      :description="t('compliance.decision.description')"
    />

    <AppCard :title="t('compliance.decision.sectionGuide')">
      <ol class="decision-page__guide">
        <li>{{ t("compliance.decision.step1") }}</li>
        <li>{{ t("compliance.decision.step2") }}</li>
        <li>{{ t("compliance.decision.step3") }}</li>
      </ol>
    </AppCard>

    <ComplianceTestPanel
      :loading="checks.loading.value"
      @submit="handleSubmit"
    />

    <div v-if="checks.error.value" class="decision-page__error" role="alert">
      <strong>{{ t("compliance.common.backendError") }}</strong>
      <span>{{ checks.error.value }}</span>
    </div>

    <div
      v-if="checks.result.value"
      ref="checkPanelEl"
      class="decision-page__result"
    >
      <ComplianceCheckResultPanel
        :result="checks.result.value"
        :loading="checks.loading.value"
        :hydrated-breaches="checks.breachById.value"
        :hydration-error="checks.hydrationError.value"
        @recheck="handleRecheck"
        @view-rule="handleViewRule"
        @view-missing-api="handleViewMissingApi"
      />

      <!--
        Verdict-aware decision footer. We do NOT enable "Submit for approval"
        because the supporting backend lifecycle (Missing API #5) is not yet
        wired. The CTA is always visible (so the demo story shows the next
        step), but disabled with a tooltip until that endpoint exists.
      -->
      <div
        v-if="isPass"
        class="decision-page__banner decision-page__banner--pass"
        role="status"
      >
        <div>
          <strong>{{ t("compliance.decision.readyForApproval") }}</strong>
        </div>
        <AppButton
          variant="primary"
          size="sm"
          :disabled="true"
          :title="t('compliance.decision.submitForApprovalDisabled')"
        >
          {{ t("compliance.decision.submitForApproval") }}
        </AppButton>
      </div>

      <div
        v-else-if="isBlock"
        class="decision-page__banner decision-page__banner--block"
        role="alert"
      >
        <strong>{{ t("compliance.decision.blockedFromApproval") }}</strong>
      </div>

      <div
        v-else-if="isWarn"
        class="decision-page__banner decision-page__banner--warn"
        role="status"
      >
        <div>
          <strong>{{ t("compliance.decision.warnFromApproval") }}</strong>
        </div>
        <AppButton
          variant="primary"
          size="sm"
          :disabled="true"
          :title="t('compliance.decision.submitForApprovalDisabled')"
        >
          {{ t("compliance.decision.submitForApproval") }}
        </AppButton>
      </div>
    </div>
  </section>
</template>

<style scoped>
.decision-page {
  display: grid;
  gap: var(--space-7);
}

.decision-page__guide {
  margin: 0;
  padding: 0 0 0 var(--space-6);
  display: grid;
  gap: var(--space-2);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
}

.decision-page__error {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-4) var(--space-5);
  border: 1px solid var(--alert-danger-border);
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
  border-radius: var(--radius-md);
}

.decision-page__result {
  display: grid;
  gap: var(--space-5);
  scroll-margin-top: var(--space-8);
}

.decision-page__banner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  font-size: var(--font-size-sm);
}

.decision-page__banner--pass {
  background: var(--alert-success-bg);
  border-color: var(--alert-success-border);
  color: var(--alert-success-text);
}

.decision-page__banner--warn {
  background: var(--alert-warning-bg);
  border-color: var(--alert-warning-border);
  color: var(--alert-warning-text);
}

.decision-page__banner--block {
  background: var(--alert-danger-bg);
  border-color: var(--alert-danger-border);
  color: var(--alert-danger-text);
}
</style>
