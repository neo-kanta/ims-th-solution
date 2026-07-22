<script setup lang="ts">
/**
 * Create Portfolio (Portfolio V2). Thin form view — real orchestration lives
 * in `usePortfolioCreateForm` + `lib/portfolioCreateValidation.ts` +
 * `lib/portfolioCreateErrors.ts`, all covered by pure-logic tests.
 *
 * Fields sent to the backend are exactly `CreatePortfolioV2Request`'s real
 * schema (see `ims-api.d.ts`) minus `style_id`/`manager_user_id`: no typed
 * style or manager-user directory exists in this frontend today (verified —
 * no `GET /investment/styles`-equivalent endpoint, and the only user
 * directory is `useComplianceUserDirectory`, a compliance-feature-owned,
 * non-generated-client wrapper around `GET /admin/users` — not a
 * general-purpose picker). Per this task's explicit rule, both optional
 * UUID fields are omitted rather than exposed as raw UUID text inputs.
 */
import { computed, onMounted, ref } from "vue";
import { useRouter } from "#imports";

import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import AppDateField from "~/shared/ui/AppDateField.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import { useI18n } from "~/composables/useI18n";
import { myFundsApi } from "~/features/my-funds/services/myFundsApi";
import type { ApiFund } from "~/features/my-funds/types";
import { usePortfolioCreateForm } from "./composables/usePortfolioCreateForm";
import { PORTFOLIO_TYPES } from "./lib/portfolioCreateValidation";
import type { PortfolioCreateFieldError } from "./lib/portfolioCreateValidation";
import type { PortfolioCreateErrorKind } from "./lib/portfolioCreateErrors";
import { portfolioOverviewPath } from "./lib/portfolioRoutes";

const { t } = useI18n();
const router = useRouter();

const form = usePortfolioCreateForm();

const funds = ref<ApiFund[]>([]);
const fundsLoading = ref(false);
const fundsError = ref<string | null>(null);

async function loadFunds() {
  fundsLoading.value = true;
  fundsError.value = null;
  try {
    const all = await myFundsApi.listMyFunds(200);
    funds.value = all.filter((f) => (f.status ?? "").toUpperCase() === "ACTIVE");
  } catch (err) {
    funds.value = [];
    fundsError.value = err instanceof Error ? err.message : t("portfolio.create.loadFundsError");
  } finally {
    fundsLoading.value = false;
  }
}

onMounted(() => {
  void loadFunds();
});

const fundOptions = computed(() =>
  funds.value.map((f) => ({ value: f.code ?? "", label: `${f.code ?? ""} — ${f.name ?? ""}` })),
);

function portfolioTypeLabel(type: (typeof PORTFOLIO_TYPES)[number]): string {
  switch (type) {
    case "LIVE":
      return t("portfolio.create.fields.portfolioType.options.live");
    case "SIMULATION":
      return t("portfolio.create.fields.portfolioType.options.simulation");
    case "MODEL":
      return t("portfolio.create.fields.portfolioType.options.model");
    default:
      return type;
  }
}

const portfolioTypeOptions = computed(() =>
  PORTFOLIO_TYPES.map((type) => ({
    value: type,
    label: portfolioTypeLabel(type),
  })),
);

function fieldErrorMessage(code: PortfolioCreateFieldError | undefined): string {
  switch (code) {
    case "required":
      return t("portfolio.create.validation.required");
    case "invalid_currency":
      return t("portfolio.create.validation.invalidCurrency");
    case "invalid_date":
      return t("portfolio.create.validation.invalidDate");
    case "invalid_portfolio_type":
      return t("portfolio.create.validation.invalidPortfolioType");
    default:
      return "";
  }
}

interface FieldSummaryItem {
  key: string;
  label: string;
  message: string;
}

function fieldLabel(key: string): string {
  switch (key) {
    case "fund_code":
      return t("portfolio.create.fields.fundCode.label");
    case "portfolio_type":
      return t("portfolio.create.fields.portfolioType.label");
    case "code":
      return t("portfolio.create.fields.code.label");
    case "name":
      return t("portfolio.create.fields.name.label");
    case "base_currency":
      return t("portfolio.create.fields.baseCurrency.label");
    case "valuation_currency":
      return t("portfolio.create.fields.valuationCurrency.label");
    case "inception_date":
      return t("portfolio.create.fields.inceptionDate.label");
    default:
      return key;
  }
}

const errorSummary = computed<FieldSummaryItem[]>(() =>
  Object.entries(form.fieldErrors.value)
    .filter(([, code]) => Boolean(code))
    .map(([key, code]) => ({
      key,
      label: fieldLabel(key),
      message: fieldErrorMessage(code),
    })),
);

function submitErrorMessage(kind: PortfolioCreateErrorKind, backendMessage: string): string {
  switch (kind) {
    case "validation":
      return t("portfolio.create.errors.validation", { message: backendMessage });
    case "forbidden":
      return t("portfolio.create.errors.forbidden", { message: backendMessage });
    case "fund_not_found":
      return t("portfolio.create.errors.fundNotFound", { message: backendMessage });
    case "duplicate_code":
      return t("portfolio.create.errors.duplicateCode", { message: backendMessage });
    case "business_rule":
      return t("portfolio.create.errors.businessRule", { message: backendMessage });
    default:
      return t("portfolio.create.errors.unexpected", { message: backendMessage });
  }
}

async function onSubmit() {
  const created = await form.submit(t("portfolio.create.errors.unexpectedFallback"));
  if (created?.code) {
    void router.push(portfolioOverviewPath(created.code));
  }
}

function onCancel() {
  void router.push("/portfolios");
}
</script>

<template>
  <section class="portfolio-create">
    <AppPageHeader
      :title="t('portfolio.create.title')"
      :description="t('portfolio.create.subtitle')"
    />

    <AppCard>
      <div
        v-if="errorSummary.length > 0"
        class="portfolio-create__summary"
        role="alert"
        aria-live="assertive"
        tabindex="-1"
      >
        <p class="portfolio-create__summary-title">{{ t("portfolio.create.validation.summaryTitle") }}</p>
        <ul>
          <li v-for="item in errorSummary" :key="item.key">
            {{ item.label }}: {{ item.message }}
          </li>
        </ul>
      </div>

      <div
        v-if="form.submitError.value"
        class="portfolio-create__summary portfolio-create__summary--submit"
        role="alert"
        aria-live="assertive"
      >
        {{ submitErrorMessage(form.submitError.value.kind, form.submitError.value.message) }}
      </div>

      <div v-if="form.phase.value === 'busy'" class="portfolio-create__notice" role="status" aria-live="polite">
        {{ t("portfolio.create.duplicateSubmitPrevented") }}
      </div>

      <form class="portfolio-create__form" @submit.prevent="onSubmit">
        <div class="portfolio-create__grid">
          <label class="portfolio-create__field" for="pc-fund-code">
            <span>{{ t("portfolio.create.fields.fundCode.label") }}</span>
            <AppSelect
              id="pc-fund-code"
              v-model="form.values.value.fund_code"
              :options="fundOptions"
              :placeholder="fundsLoading ? t('portfolio.create.loadingFunds') : t('portfolio.create.fields.fundCode.placeholder')"
              :disabled="fundsLoading || form.busy.value"
              :error="Boolean(form.fieldErrors.value.fund_code)"
            />
            <span v-if="fundsError" class="portfolio-create__field-error" role="alert">{{ fundsError }}</span>
            <span v-else-if="!fundsLoading && fundOptions.length === 0" class="portfolio-create__field-hint">
              {{ t("portfolio.create.noFunds") }}
            </span>
            <span v-if="form.fieldErrors.value.fund_code" class="portfolio-create__field-error" role="alert">
              {{ fieldErrorMessage(form.fieldErrors.value.fund_code) }}
            </span>
          </label>

          <label class="portfolio-create__field" for="pc-portfolio-type">
            <span>{{ t("portfolio.create.fields.portfolioType.label") }}</span>
            <AppSelect
              id="pc-portfolio-type"
              v-model="form.values.value.portfolio_type"
              :options="portfolioTypeOptions"
              :placeholder="t('portfolio.create.fields.portfolioType.placeholder')"
              :disabled="form.busy.value"
              :error="Boolean(form.fieldErrors.value.portfolio_type)"
            />
            <span v-if="form.fieldErrors.value.portfolio_type" class="portfolio-create__field-error" role="alert">
              {{ fieldErrorMessage(form.fieldErrors.value.portfolio_type) }}
            </span>
          </label>

          <label class="portfolio-create__field" for="pc-code">
            <span>{{ t("portfolio.create.fields.code.label") }}</span>
            <AppInput
              id="pc-code"
              v-model="form.values.value.code"
              :placeholder="t('portfolio.create.fields.code.placeholder')"
              :disabled="form.busy.value"
              :error="Boolean(form.fieldErrors.value.code)"
            />
            <span v-if="form.fieldErrors.value.code" class="portfolio-create__field-error" role="alert">
              {{ fieldErrorMessage(form.fieldErrors.value.code) }}
            </span>
          </label>

          <label class="portfolio-create__field" for="pc-name">
            <span>{{ t("portfolio.create.fields.name.label") }}</span>
            <AppInput
              id="pc-name"
              v-model="form.values.value.name"
              :placeholder="t('portfolio.create.fields.name.placeholder')"
              :disabled="form.busy.value"
              :error="Boolean(form.fieldErrors.value.name)"
            />
            <span v-if="form.fieldErrors.value.name" class="portfolio-create__field-error" role="alert">
              {{ fieldErrorMessage(form.fieldErrors.value.name) }}
            </span>
          </label>

          <label class="portfolio-create__field portfolio-create__field--wide" for="pc-description">
            <span>{{ t("portfolio.create.fields.description.label") }}</span>
            <AppInput
              id="pc-description"
              v-model="form.values.value.description"
              :placeholder="t('portfolio.create.fields.description.placeholder')"
              :disabled="form.busy.value"
            />
          </label>

          <label class="portfolio-create__field" for="pc-base-currency">
            <span>{{ t("portfolio.create.fields.baseCurrency.label") }}</span>
            <AppInput
              id="pc-base-currency"
              v-model="form.values.value.base_currency"
              :placeholder="t('portfolio.create.fields.baseCurrency.placeholder')"
              :disabled="form.busy.value"
              :error="Boolean(form.fieldErrors.value.base_currency)"
            />
            <span v-if="form.fieldErrors.value.base_currency" class="portfolio-create__field-error" role="alert">
              {{ fieldErrorMessage(form.fieldErrors.value.base_currency) }}
            </span>
          </label>

          <label class="portfolio-create__field" for="pc-valuation-currency">
            <span>{{ t("portfolio.create.fields.valuationCurrency.label") }}</span>
            <AppInput
              id="pc-valuation-currency"
              v-model="form.values.value.valuation_currency"
              :placeholder="t('portfolio.create.fields.valuationCurrency.placeholder')"
              :disabled="form.busy.value"
              :error="Boolean(form.fieldErrors.value.valuation_currency)"
            />
            <span v-if="form.fieldErrors.value.valuation_currency" class="portfolio-create__field-error" role="alert">
              {{ fieldErrorMessage(form.fieldErrors.value.valuation_currency) }}
            </span>
          </label>

          <label class="portfolio-create__field" for="pc-strategy-code">
            <span>{{ t("portfolio.create.fields.strategyCode.label") }}</span>
            <AppInput
              id="pc-strategy-code"
              v-model="form.values.value.strategy_code"
              :placeholder="t('portfolio.create.fields.strategyCode.placeholder')"
              :disabled="form.busy.value"
            />
          </label>

          <label class="portfolio-create__field" for="pc-benchmark">
            <span>{{ t("portfolio.create.fields.benchmark.label") }}</span>
            <AppInput
              id="pc-benchmark"
              v-model="form.values.value.benchmark"
              :placeholder="t('portfolio.create.fields.benchmark.placeholder')"
              :disabled="form.busy.value"
            />
          </label>

          <label class="portfolio-create__field" for="pc-risk-profile">
            <span>{{ t("portfolio.create.fields.riskProfile.label") }}</span>
            <AppInput
              id="pc-risk-profile"
              v-model="form.values.value.risk_profile"
              :placeholder="t('portfolio.create.fields.riskProfile.placeholder')"
              :disabled="form.busy.value"
            />
          </label>

          <label class="portfolio-create__field" for="pc-inception-date">
            <span>{{ t("portfolio.create.fields.inceptionDate.label") }}</span>
            <AppDateField
              id="pc-inception-date"
              v-model="form.values.value.inception_date"
              :disabled="form.busy.value"
              :error="Boolean(form.fieldErrors.value.inception_date)"
            />
            <span v-if="form.fieldErrors.value.inception_date" class="portfolio-create__field-error" role="alert">
              {{ fieldErrorMessage(form.fieldErrors.value.inception_date) }}
            </span>
          </label>
        </div>

        <div class="portfolio-create__actions">
          <AppButton variant="secondary" type="button" :disabled="form.busy.value" @click="onCancel">
            {{ t("portfolio.create.cancel") }}
          </AppButton>
          <AppButton variant="primary" type="submit" :loading="form.busy.value" :disabled="form.busy.value">
            {{ form.busy.value ? t("portfolio.create.submitting") : t("portfolio.create.submit") }}
          </AppButton>
        </div>
      </form>
    </AppCard>
  </section>
</template>

<style scoped>
.portfolio-create {
  display: grid;
  gap: var(--space-4, 16px);
  max-width: 960px;
}

.portfolio-create__summary {
  margin-bottom: var(--space-4, 16px);
  padding: var(--space-3, 12px) var(--space-4, 16px);
  border: 1px solid var(--alert-danger-border, #cf222e);
  border-radius: var(--radius-md, 6px);
  background: var(--alert-danger-bg, #ffebe9);
  color: var(--alert-danger-text, #cf222e);
  font-size: 13px;
}

.portfolio-create__summary ul {
  margin: 6px 0 0;
  padding-left: 18px;
}

.portfolio-create__summary-title {
  margin: 0;
  font-weight: 600;
}

.portfolio-create__notice {
  margin-bottom: var(--space-4, 16px);
  padding: var(--space-3, 12px) var(--space-4, 16px);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card-muted, #f6f8fa);
  color: var(--text-secondary, #57606a);
  font-size: 13px;
}

.portfolio-create__form {
  display: grid;
  gap: var(--space-5, 20px);
}

.portfolio-create__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-4, 16px);
}

.portfolio-create__field {
  display: grid;
  gap: 4px;
  min-width: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary, #57606a);
}

.portfolio-create__field--wide {
  grid-column: 1 / -1;
}

.portfolio-create__field-error {
  color: var(--alert-danger-text, #cf222e);
  font-size: 11px;
  font-weight: 500;
}

.portfolio-create__field-hint {
  color: var(--text-tertiary, #6e7781);
  font-size: 11px;
  font-weight: 500;
}

.portfolio-create__actions {
  display: flex;
  gap: var(--space-2, 8px);
  justify-content: flex-end;
}

@media (max-width: 720px) {
  .portfolio-create__grid {
    grid-template-columns: 1fr;
  }
}
</style>
