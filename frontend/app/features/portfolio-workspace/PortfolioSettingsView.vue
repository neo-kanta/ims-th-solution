<script setup lang="ts">
/**
 * Portfolio-code-routed Settings page. Edits mutable descriptive metadata
 * (name/description/benchmark/risk_profile/strategy_code) via the real
 * `PATCH /portfolios/{portfolioCode}` endpoint, using optimistic
 * `expected_version` concurrency. Fund association, code, portfolio_type,
 * and lifecycle status are immutable through this screen by backend design
 * — shown read-only, with an explicit note that lifecycle changes are not
 * available here.
 */
import { computed, onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import IMSPermissionGuard from "~/shared/ui/IMSPermissionGuard.vue";
import { useI18n } from "~/composables/useI18n";
import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import { usePortfolioSettingsForm } from "./composables/usePortfolioSettingsForm";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.workspaceTabs.settings"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true },
);

const ctx = usePortfolioContext(() => props.portfolioCode);
const form = usePortfolioSettingsForm();
const successVisible = ref(false);

async function loadPortfolio() {
  await ctx.reload();
  form.hydrate(ctx.portfolio.value);
}

onMounted(loadPortfolio);
watch(() => props.portfolioCode, loadPortfolio);

function displayOrDash(value: string | undefined | null): string {
  return value && value.trim() ? value : t("portfolio.terminal.unavailable");
}

const errorSummary = computed(() => {
  if (!form.submitError.value) return null;
  switch (form.submitError.value.kind) {
    case "validation":
      return t("portfolio.settingsPage.errors.validation", { message: form.submitError.value.message });
    case "forbidden":
      return t("portfolio.settingsPage.errors.forbidden", { message: form.submitError.value.message });
    case "not_found":
      return t("portfolio.settingsPage.errors.notFound", { message: form.submitError.value.message });
    case "version_conflict":
      return t("portfolio.settingsPage.errors.versionConflict", { message: form.submitError.value.message });
    case "business_rule":
      return t("portfolio.settingsPage.errors.businessRule", { message: form.submitError.value.message });
    default:
      return t("portfolio.settingsPage.errors.unexpected", { message: form.submitError.value.message });
  }
});

async function onSubmit() {
  successVisible.value = false;
  const result = await form.submit(
    props.portfolioCode,
    ctx.portfolio.value?.version,
    t("portfolio.settingsPage.errors.unexpectedFallback"),
  );
  if (result) {
    successVisible.value = true;
    await loadPortfolio();
  }
}

async function reloadAfterConflict() {
  await loadPortfolio();
}
</script>

<template>
  <section class="portfolio-settings">
    <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

    <AppCard :title="t('portfolio.settingsPage.title')" :subtitle="t('portfolio.settingsPage.subtitle')">
      <div v-if="ctx.loading.value && !ctx.portfolio.value" class="portfolio-settings__notice" role="status">
        {{ t("portfolio.settingsPage.loading") }}
      </div>
      <div v-else-if="ctx.error.value" class="portfolio-settings__error" role="alert">
        <div>{{ t("portfolio.settingsPage.errorTitle") }}: {{ ctx.error.value }}</div>
        <button type="button" class="portfolio-settings__retry" @click="loadPortfolio">
          {{ t("portfolio.settingsPage.retry") }}
        </button>
      </div>

      <template v-else-if="ctx.portfolio.value">
        <dl class="portfolio-settings__readonly-fields">
          <div>
            <dt>{{ t("portfolio.settingsPage.fields.code") }}</dt>
            <dd>{{ displayOrDash(ctx.portfolio.value.code) }}</dd>
          </div>
          <div>
            <dt>{{ t("portfolio.settingsPage.fields.baseCurrency") }}</dt>
            <dd>{{ displayOrDash(ctx.portfolio.value.base_currency) }}</dd>
          </div>
          <div>
            <dt>{{ t("portfolio.settingsPage.fields.valuationCurrency") }}</dt>
            <dd>{{ displayOrDash(ctx.portfolio.value.valuation_currency) }}</dd>
          </div>
          <div>
            <dt>{{ t("portfolio.settingsPage.fields.portfolioType") }}</dt>
            <dd>{{ displayOrDash(ctx.portfolio.value.portfolio_type) }}</dd>
          </div>
          <div>
            <dt>{{ t("portfolio.settingsPage.fields.status") }}</dt>
            <dd>{{ displayOrDash(ctx.portfolio.value.status) }}</dd>
          </div>
        </dl>
        <p class="portfolio-settings__lifecycle-note">{{ t("portfolio.settingsPage.lifecycleNote") }}</p>

        <form class="portfolio-settings__form" novalidate @submit.prevent="onSubmit">
          <div
            v-if="errorSummary"
            class="portfolio-settings__error-summary"
            role="alert"
            tabindex="-1"
          >
            {{ errorSummary }}
            <button
              v-if="form.submitError.value?.kind === 'version_conflict'"
              type="button"
              class="portfolio-settings__retry"
              @click="reloadAfterConflict"
            >
              {{ t("portfolio.settingsPage.reload") }}
            </button>
          </div>
          <p v-if="successVisible" class="portfolio-settings__success" role="status">
            {{ t("portfolio.settingsPage.success") }}
          </p>

          <label class="portfolio-settings__field">
            <span>{{ t("portfolio.settingsPage.fields.name") }}</span>
            <input v-model="form.values.value.name" type="text" :disabled="form.busy.value" />
          </label>
          <label class="portfolio-settings__field">
            <span>{{ t("portfolio.settingsPage.fields.description") }}</span>
            <textarea v-model="form.values.value.description" rows="2" :disabled="form.busy.value" />
          </label>
          <label class="portfolio-settings__field">
            <span>{{ t("portfolio.settingsPage.fields.benchmark") }}</span>
            <input v-model="form.values.value.benchmark" type="text" :disabled="form.busy.value" />
          </label>
          <label class="portfolio-settings__field">
            <span>{{ t("portfolio.settingsPage.fields.riskProfile") }}</span>
            <select v-model="form.values.value.risk_profile" :disabled="form.busy.value">
              <option value="">{{ t("portfolio.create.fields.riskProfile.placeholder") }}</option>
              <option value="LOW">{{ t("portfolio.create.fields.riskProfile.options.low") }}</option>
              <option value="MEDIUM">{{ t("portfolio.create.fields.riskProfile.options.medium") }}</option>
              <option value="HIGH">{{ t("portfolio.create.fields.riskProfile.options.high") }}</option>
              <option value="SPECULATIVE">{{ t("portfolio.create.fields.riskProfile.options.speculative") }}</option>
            </select>
          </label>
          <label class="portfolio-settings__field">
            <span>{{ t("portfolio.settingsPage.fields.strategyCode") }}</span>
            <input v-model="form.values.value.strategy_code" type="text" :disabled="form.busy.value" />
          </label>

          <IMSPermissionGuard permission="INVESTMENT_PORTFOLIO_MANAGE" mode="disable">
            <AppButton type="submit" variant="primary" size="sm" :loading="form.busy.value" :disabled="form.busy.value">
              {{ t("portfolio.settingsPage.save") }}
            </AppButton>
          </IMSPermissionGuard>
        </form>
      </template>
    </AppCard>
  </section>
</template>

<style scoped>
.portfolio-settings {
  display: grid;
  gap: var(--space-4, 16px);
}

.portfolio-settings__notice,
.portfolio-settings__error {
  padding: var(--space-4, 16px);
  font-size: 13px;
  color: var(--text-secondary, #57606a);
}

.portfolio-settings__error {
  color: var(--alert-danger-text, #cf222e);
  display: grid;
  gap: 6px;
}

.portfolio-settings__retry {
  justify-self: start;
  font-family: inherit;
  font-size: 12px;
  font-weight: 600;
  border: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card, #ffffff);
  border-radius: var(--radius-md, 5px);
  padding: 4px 10px;
  cursor: pointer;
  margin-top: 6px;
}

.portfolio-settings__readonly-fields {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 14px;
  margin: 0 0 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.portfolio-settings__readonly-fields dt {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary, #6e7781);
}

.portfolio-settings__readonly-fields dd {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--text-primary, #1f2328);
}

.portfolio-settings__lifecycle-note {
  margin: 0 0 20px;
  padding: 10px 12px;
  font-size: 12px;
  color: var(--text-tertiary, #6e7781);
  border: 1px dashed var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
}

.portfolio-settings__form {
  display: grid;
  gap: 14px;
  max-width: 480px;
}

.portfolio-settings__field {
  display: grid;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary, #57606a);
}

.portfolio-settings__field input,
.portfolio-settings__field textarea {
  font: inherit;
  font-weight: 400;
  padding: 8px 10px;
  border: 1px solid var(--border-default, #d0d7de);
  border-radius: 6px;
  background: var(--bg-input, #fff);
  color: var(--text-primary, #1f2328);
}

.portfolio-settings__error-summary {
  padding: 10px 12px;
  background: var(--alert-danger-bg, #ffebe9);
  border: 1px solid var(--alert-danger-border, #cf222e);
  border-radius: var(--radius-sm, 4px);
  color: var(--alert-danger-text, #cf222e);
  font-size: 13px;
}

.portfolio-settings__success {
  padding: 10px 12px;
  background: rgba(26, 127, 55, 0.1);
  border: 1px solid var(--state-success, #1a7f37);
  border-radius: var(--radius-sm, 4px);
  color: var(--state-success, #1a7f37);
  font-size: 13px;
}
</style>
