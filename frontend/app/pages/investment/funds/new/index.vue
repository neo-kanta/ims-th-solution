<script setup lang="ts">
/**
 * /investment/funds/new — create a real fund.
 *
 * Calls POST /api/v1/investment/funds via the generated OpenAPI client.
 * The form collects every field the backend accepts in CreateFundRequest
 * plus the new per-fund `require_pretrade_preview` toggle that controls
 * whether the Operation-tab trade ticket must run a pre-trade simulation
 * before allowing a posted transaction.
 */
import { computed, onMounted, ref } from "vue";

import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import { useI18n } from "~/composables/useI18n";
import { myFundsApi } from "~/features/my-funds";
import type {
  ApiFundCategory,
  CreateFundPayload,
} from "~/features/my-funds/services/myFundsApi";
import { OpenApiRequestError } from "~/api/openapi";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_MANAGE",
});

const { t } = useI18n();
const router = useRouter();

// ─── Form state ──────────────────────────────────────────────────────────
const code = ref("");
const name = ref("");
const shortName = ref("");
const fundCategoryId = ref("");
const baseCurrency = ref("THB");
const inceptionDate = ref(new Date().toISOString().slice(0, 10));
const managerUserId = ref(""); // optional
const benchmark = ref("");
const riskProfile = ref<"" | "LOW" | "MEDIUM" | "HIGH" | "SPECULATIVE">("MEDIUM");
const hasUnits = ref(false);
const requirePretradePreview = ref(true);
const externalPamRef = ref("");

const categories = ref<ApiFundCategory[]>([]);
const categoriesLoading = ref(false);

const submitting = ref(false);
const submitError = ref<string | null>(null);
const fieldErrors = ref<Record<string, string>>({});

onMounted(async () => {
  categoriesLoading.value = true;
  try {
    categories.value = await myFundsApi.listFundCategories();
    // Default to the first category if available so the form is valid
    // by default and the user only has to override deliberately.
    const first = categories.value[0]?.id;
    if (first && !fundCategoryId.value) fundCategoryId.value = first;
  } finally {
    categoriesLoading.value = false;
  }
});

const baseCurrencyOptions = ["THB", "USD", "EUR", "GBP", "JPY", "SGD", "HKD", "CNY"];
const riskProfileOptions: { value: typeof riskProfile.value; label: string }[] = [
  { value: "LOW", label: t("funds.create.riskLow", "Low") },
  { value: "MEDIUM", label: t("funds.create.riskMedium", "Medium") },
  { value: "HIGH", label: t("funds.create.riskHigh", "High") },
  { value: "SPECULATIVE", label: t("funds.create.riskSpec", "Speculative") },
];

const canSubmit = computed(() =>
  !!code.value.trim() &&
  !!name.value.trim() &&
  !!fundCategoryId.value &&
  /^[A-Z]{3}$/.test(baseCurrency.value) &&
  !!inceptionDate.value &&
  !submitting.value,
);

function validate(): boolean {
  const errors: Record<string, string> = {};
  if (!code.value.trim()) errors.code = t("funds.create.errCode", "Code is required");
  if (code.value.length > 40) errors.code = t("funds.create.errCodeLen", "Code must be ≤ 40 chars");
  if (!name.value.trim()) errors.name = t("funds.create.errName", "Name is required");
  if (!fundCategoryId.value) errors.fundCategoryId = t("funds.create.errCategory", "Category is required");
  if (!/^[A-Z]{3}$/.test(baseCurrency.value)) {
    errors.baseCurrency = t("funds.create.errCcy", "Base currency must be 3 uppercase letters");
  }
  if (!inceptionDate.value) errors.inceptionDate = t("funds.create.errDate", "Inception date is required");
  fieldErrors.value = errors;
  return Object.keys(errors).length === 0;
}

async function onSubmit() {
  submitError.value = null;
  if (!validate()) return;
  submitting.value = true;
  try {
    const payload: CreateFundPayload = {
      code: code.value.trim(),
      name: name.value.trim(),
      short_name: shortName.value.trim() || undefined,
      fund_category_id: fundCategoryId.value,
      base_currency: baseCurrency.value,
      inception_date: inceptionDate.value,
      manager_user_id: managerUserId.value.trim() || undefined,
      benchmark: benchmark.value.trim() || undefined,
      risk_profile: riskProfile.value || undefined,
      has_units: hasUnits.value,
      // NB: the generated openapi types are slightly behind on the
      // require_pretrade_preview field for older builds — cast through
      // Record so we still send it. The backend accepts the field.
      ...(({ require_pretrade_preview: requirePretradePreview.value } as unknown) as Record<string, unknown>),
      external_pam_ref: externalPamRef.value.trim() || undefined,
    };
    const created = await myFundsApi.createFund(payload);
    // Land on the new fund's Holdings page so the operator can immediately
    // start setting up portfolios / trading.
    if (created?.id) {
      void router.push(`/investment/funds/${created.id}/holdings`);
    } else {
      void router.push("/investment/funds");
    }
  } catch (err) {
    if (err instanceof OpenApiRequestError) {
      submitError.value = `${err.status}: ${err.message}`;
    } else if (err instanceof Error) {
      submitError.value = err.message;
    } else {
      submitError.value = t("funds.create.errGeneric", "Failed to create fund");
    }
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <section class="cf-page">
    <AppPageHeader
      :title="t('funds.create.title', 'Create fund')"
      :description="
        t(
          'funds.create.subtitle',
          'Register a new fund / contract. The fund ID becomes the contract_id used across workflow, compliance and permissions.',
        )
      "
    >
      <template #actions>
        <NuxtLink to="/investment/funds" class="cf-page__back">
          {{ t("funds.create.cancel", "← Back to My funds") }}
        </NuxtLink>
      </template>
    </AppPageHeader>

    <form class="cf-form" @submit.prevent="onSubmit">
      <fieldset class="cf-form__section">
        <legend>{{ t("funds.create.sectionIdentity", "Identity") }}</legend>

        <div class="cf-form__row">
          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.code", "Code") }} <em>*</em></span>
            <input
              v-model.trim="code"
              class="cf-input"
              placeholder="e.g. ABC-EQUITY"
              maxlength="40"
              required
            />
            <small v-if="fieldErrors.code" class="cf-form__err">{{ fieldErrors.code }}</small>
            <small v-else class="cf-form__hint">
              {{ t("funds.create.codeHint", "Unique business code. Becomes the contract identifier.") }}
            </small>
          </label>

          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.name", "Name") }} <em>*</em></span>
            <input
              v-model.trim="name"
              class="cf-input"
              placeholder="e.g. ABC Large-cap Equity"
              maxlength="255"
              required
            />
            <small v-if="fieldErrors.name" class="cf-form__err">{{ fieldErrors.name }}</small>
          </label>
        </div>

        <div class="cf-form__row">
          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.shortName", "Short name") }}</span>
            <input
              v-model.trim="shortName"
              class="cf-input"
              placeholder="e.g. abc-equity"
              maxlength="80"
            />
            <small class="cf-form__hint">
              {{ t("funds.create.shortNameHint", "Slug used in breadcrumbs.") }}
            </small>
          </label>

          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.category", "Category") }} <em>*</em></span>
            <select v-model="fundCategoryId" class="cf-input" required>
              <option value="" disabled>
                {{
                  categoriesLoading
                    ? t("funds.create.loadingCategories", "Loading…")
                    : t("funds.create.pickCategory", "Pick a category")
                }}
              </option>
              <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
            <small v-if="fieldErrors.fundCategoryId" class="cf-form__err">
              {{ fieldErrors.fundCategoryId }}
            </small>
          </label>
        </div>
      </fieldset>

      <fieldset class="cf-form__section">
        <legend>{{ t("funds.create.sectionMoney", "Currency, manager and benchmark") }}</legend>

        <div class="cf-form__row">
          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.baseCcy", "Base currency") }} <em>*</em></span>
            <select v-model="baseCurrency" class="cf-input" required>
              <option v-for="c in baseCurrencyOptions" :key="c" :value="c">{{ c }}</option>
            </select>
            <small v-if="fieldErrors.baseCurrency" class="cf-form__err">
              {{ fieldErrors.baseCurrency }}
            </small>
          </label>

          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.inception", "Inception date") }} <em>*</em></span>
            <input v-model="inceptionDate" type="date" class="cf-input" required />
            <small v-if="fieldErrors.inceptionDate" class="cf-form__err">
              {{ fieldErrors.inceptionDate }}
            </small>
          </label>
        </div>

        <div class="cf-form__row">
          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.manager", "Manager user id") }}</span>
            <input
              v-model.trim="managerUserId"
              class="cf-input"
              placeholder="UUID — optional"
            />
            <small class="cf-form__hint">
              {{
                t(
                  "funds.create.managerHint",
                  "Optional. If left blank you can assign the manager later via the fund settings tab.",
                )
              }}
            </small>
          </label>

          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.benchmark", "Benchmark") }}</span>
            <input
              v-model.trim="benchmark"
              class="cf-input"
              placeholder="e.g. SET50 / NASDAQ-100"
              maxlength="120"
            />
          </label>
        </div>

        <div class="cf-form__row">
          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.risk", "Risk profile") }}</span>
            <select v-model="riskProfile" class="cf-input">
              <option v-for="opt in riskProfileOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </label>

          <label class="cf-form__field">
            <span class="cf-form__label">{{ t("funds.create.pamRef", "External PAM reference") }}</span>
            <input
              v-model.trim="externalPamRef"
              class="cf-input"
              placeholder="Optional PAM link"
              maxlength="80"
            />
          </label>
        </div>
      </fieldset>

      <fieldset class="cf-form__section">
        <legend>{{ t("funds.create.sectionFlags", "Behaviour") }}</legend>

        <div class="cf-form__check-row">
          <label class="cf-form__check">
            <input v-model="hasUnits" type="checkbox" />
            <span>
              <strong>{{ t("funds.create.hasUnits", "Unitised fund") }}</strong>
              <small>
                {{
                  t(
                    "funds.create.hasUnitsHint",
                    "NAV-per-unit accounting. Required when external investors hold units in the fund.",
                  )
                }}
              </small>
            </span>
          </label>

          <label class="cf-form__check">
            <input v-model="requirePretradePreview" type="checkbox" />
            <span>
              <strong>{{ t("funds.create.requirePretrade", "Require pre-trade preview") }}</strong>
              <small>
                {{
                  t(
                    "funds.create.requirePretradeHint",
                    "Operation tab will force a pre-trade simulation before any BUY/SELL post. Leave on for stricter funds; turn off if the server-side gates are enough.",
                  )
                }}
              </small>
            </span>
          </label>
        </div>
      </fieldset>

      <div v-if="submitError" class="cf-form__alert" role="alert">{{ submitError }}</div>

      <div class="cf-form__actions">
        <NuxtLink to="/investment/funds" class="cf-btn cf-btn--secondary">
          {{ t("funds.create.cancel", "Cancel") }}
        </NuxtLink>
        <button
          type="submit"
          class="cf-btn cf-btn--primary"
          :disabled="!canSubmit"
        >
          {{
            submitting
              ? t("funds.create.submitting", "Creating…")
              : t("funds.create.submit", "Create fund")
          }}
        </button>
      </div>
    </form>
  </section>
</template>

<style scoped>
.cf-page {
  display: grid;
  gap: var(--space-3);
  max-width: 980px;
}

.cf-page__back {
  color: var(--text-secondary);
  font-size: 13px;
  text-decoration: none;
}
.cf-page__back:hover { color: var(--text-primary); }

.cf-form {
  display: grid;
  gap: var(--space-3);
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: 8px;
  padding: 20px 24px;
}

.cf-form__section {
  border: 0;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 12px;
}

.cf-form__section legend {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-tertiary);
  font-weight: 600;
  padding-bottom: 4px;
  border-bottom: 1px solid var(--border-subtle);
  width: 100%;
}

.cf-form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.cf-form__field {
  display: grid;
  gap: 4px;
}

.cf-form__label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary);
}
.cf-form__label em {
  color: var(--state-danger, #cf222e);
  font-style: normal;
}

.cf-input {
  height: 32px;
  padding: 0 10px;
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-primary);
}
.cf-input:focus {
  outline: none;
  border-color: var(--border-focus, var(--color-primary-500, #1f6feb));
  box-shadow: 0 0 0 3px rgba(31, 111, 235, 0.18);
}
select.cf-input {
  appearance: none;
  background-image: linear-gradient(45deg, transparent 50%, currentColor 50%),
    linear-gradient(135deg, currentColor 50%, transparent 50%);
  background-position: calc(100% - 14px) 14px, calc(100% - 10px) 14px;
  background-size: 4px 4px, 4px 4px;
  background-repeat: no-repeat;
  padding-right: 26px;
  color: var(--text-primary);
}

.cf-form__hint {
  font-size: 11px;
  color: var(--text-tertiary);
}

.cf-form__err {
  font-size: 11px;
  color: var(--state-danger, #cf222e);
}

.cf-form__check-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.cf-form__check {
  display: grid;
  grid-template-columns: auto 1fr;
  align-items: start;
  gap: 10px;
  padding: 12px;
  background: var(--surface-1);
  border: 1px solid var(--border-default);
  border-radius: 6px;
  cursor: pointer;
}
.cf-form__check input[type="checkbox"] {
  margin-top: 2px;
}
.cf-form__check span {
  display: grid;
  gap: 2px;
}
.cf-form__check strong {
  font-size: 13px;
  color: var(--text-primary);
}
.cf-form__check small {
  font-size: 11px;
  color: var(--text-tertiary);
  line-height: 1.5;
}

.cf-form__alert {
  padding: 10px 12px;
  background: var(--alert-danger-bg);
  border: 1px solid var(--alert-danger-border);
  border-radius: 6px;
  color: var(--alert-danger-text);
  font-size: 12px;
}

.cf-form__actions {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  justify-content: flex-end;
  border-top: 1px solid var(--border-subtle);
  padding-top: 14px;
}

.cf-btn {
  display: inline-flex;
  align-items: center;
  height: 32px;
  padding: 0 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  text-decoration: none;
  border: 1px solid var(--border-default);
}
.cf-btn--secondary {
  background: var(--action-secondary);
  color: var(--text-primary);
}
.cf-btn--secondary:hover { background: var(--action-secondary-hover, var(--surface-1)); }

.cf-btn--primary {
  background: var(--state-success, #1a7f37);
  color: #fff;
  border-color: transparent;
}
.cf-btn--primary:hover { background: #1f8e3f; }
.cf-btn--primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 720px) {
  .cf-form__row,
  .cf-form__check-row {
    grid-template-columns: 1fr;
  }
}
</style>
