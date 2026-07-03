<script setup lang="ts">
import { computed, onMounted, reactive, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";

import { useCompliancePortfolioDirectory } from "../composables/useCompliancePortfolioDirectory";
import { isPositiveDecimal, isUuid } from "../lib/formatters";
import type {
  CompliancePortfolioOption,
  CompliancePreTradeRequest,
} from "../types";

interface Props {
  loading?: boolean;
  initial?: Partial<CompliancePreTradeRequest>;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  initial: () => ({}),
});

const emit = defineEmits<{
  submit: [payload: CompliancePreTradeRequest];
}>();

const { t } = useI18n();
const authStore = useAuthStore();
const portfolios = useCompliancePortfolioDirectory();

function todayIso(): string {
  return new Date().toISOString().slice(0, 10);
}

interface FormState {
  portfolio_id: string;
  contract_id: string;
  business_date: string;
  order_id: string;
  ticker: string;
  side: "BUY" | "SELL";
  quantity: string;
  price: string;
  currency: string;
  exchange: string;
}

const form = reactive<FormState>({
  portfolio_id: props.initial?.portfolio_id ?? "",
  contract_id: props.initial?.contract_id ?? "",
  business_date: props.initial?.business_date ?? todayIso(),
  order_id: props.initial?.order_id ?? "",
  ticker: props.initial?.ticker ?? "",
  side: (props.initial?.side as FormState["side"]) ?? "BUY",
  quantity: props.initial?.quantity ?? "",
  price: props.initial?.price ?? "",
  currency: props.initial?.currency ?? "",
  exchange: props.initial?.exchange ?? "",
});

const touched = reactive<Record<keyof FormState, boolean>>({
  portfolio_id: false,
  contract_id: false,
  business_date: false,
  order_id: false,
  ticker: false,
  side: false,
  quantity: false,
  price: false,
  currency: false,
  exchange: false,
});

function markTouched(key: keyof FormState) {
  touched[key] = true;
}

// Auth store exposes `permissions.contracts: string[]` — UUIDs the user is
// authorised against. We show those as a dropdown; if the list is empty we
// fall back to a free-text UUID input so a developer can still test.
const authorisedContractIds = computed<string[]>(() => {
  const list = authStore.permissions?.contracts;
  return Array.isArray(list) ? list : [];
});

const hasPortfolioOptions = computed(() => portfolios.items.value.length > 0);
const hasContractOptions = computed(
  () => authorisedContractIds.value.length > 0,
);

const selectedPortfolio = computed<CompliancePortfolioOption | null>(() => {
  if (!form.portfolio_id) return null;
  return portfolios.items.value.find((p) => p.id === form.portfolio_id) ?? null;
});

// Auto-fill currency from the selected portfolio's base currency, but only
// when the user has not manually typed one — never overwrite a deliberate
// override.
watch(selectedPortfolio, (next) => {
  if (next && !form.currency) {
    form.currency = next.base_currency;
  }
});

function formatPortfolioOption(p: CompliancePortfolioOption): string {
  const parts = [p.code, p.name].filter(Boolean);
  const head = parts.length > 0 ? parts.join(" — ") : p.id;
  return p.base_currency ? `${head} (${p.base_currency})` : head;
}

function shortenId(id: string): string {
  if (!id) return "";
  return id.length > 12 ? `${id.slice(0, 8)}…${id.slice(-4)}` : id;
}

const errors = computed(() => ({
  portfolio_id: !form.portfolio_id
    ? t("compliance.preTrade.form.required")
    : !isUuid(form.portfolio_id)
      ? t("compliance.preTrade.form.invalidUuid")
      : null,
  contract_id: !form.contract_id
    ? t("compliance.preTrade.form.required")
    : !isUuid(form.contract_id)
      ? t("compliance.preTrade.form.invalidUuid")
      : null,
  business_date: !form.business_date
    ? t("compliance.preTrade.form.required")
    : null,
  order_id: !form.order_id
    ? t("compliance.preTrade.form.required")
    : !isUuid(form.order_id)
      ? t("compliance.preTrade.form.invalidUuid")
      : null,
  ticker: !form.ticker ? t("compliance.preTrade.form.required") : null,
  quantity: !form.quantity
    ? t("compliance.preTrade.form.required")
    : !isPositiveDecimal(form.quantity)
      ? t("compliance.preTrade.form.invalidDecimal")
      : null,
  price: !form.price
    ? t("compliance.preTrade.form.required")
    : !isPositiveDecimal(form.price)
      ? t("compliance.preTrade.form.invalidDecimal")
      : null,
  currency: !form.currency ? t("compliance.preTrade.form.required") : null,
  exchange: !form.exchange ? t("compliance.preTrade.form.required") : null,
}));

const isValid = computed(() =>
  Object.values(errors.value).every((value) => value === null),
);

function generateOrderId() {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    form.order_id = crypto.randomUUID();
    touched.order_id = true;
  }
}

function submit() {
  (Object.keys(touched) as (keyof FormState)[]).forEach(
    (key) => (touched[key] = true),
  );
  if (!isValid.value) return;
  const payload: CompliancePreTradeRequest = { ...form };
  emit("submit", payload);
}

onMounted(() => {
  void portfolios.ensureLoaded();
});
</script>

<template>
  <AppCard
    :title="t('compliance.preTrade.title')"
    :subtitle="t('compliance.preTrade.description')"
  >
    <form class="pretrade-form" @submit.prevent="submit">
      <div class="pretrade-form__section">
        <h3 class="pretrade-form__heading">
          {{ t("compliance.preTrade.form.sectionContext") }}
        </h3>
        <div class="pretrade-form__grid">
          <label class="pretrade-form__field">
            <span class="pretrade-form__label">
              {{ t("compliance.preTrade.form.portfolio") }}
            </span>
            <select
              v-if="hasPortfolioOptions"
              v-model="form.portfolio_id"
              class="form-control"
              aria-describedby="err-portfolio-id"
              :disabled="portfolios.loading.value"
              @blur="markTouched('portfolio_id')"
            >
              <option value="" disabled>
                {{
                  portfolios.loading.value
                    ? t("compliance.preTrade.form.portfolioLoading")
                    : t("compliance.preTrade.form.portfolioPlaceholder")
                }}
              </option>
              <option
                v-for="p in portfolios.items.value"
                :key="p.id"
                :value="p.id"
              >
                {{ formatPortfolioOption(p) }}
              </option>
            </select>
            <input
              v-else
              v-model="form.portfolio_id"
              type="text"
              class="form-control"
              autocomplete="off"
              :placeholder="t('compliance.preTrade.form.portfolioIdFallback')"
              aria-describedby="err-portfolio-id"
              @blur="markTouched('portfolio_id')"
            />
            <span
              v-if="!portfolios.loading.value && !hasPortfolioOptions && !portfolios.error.value"
              class="pretrade-form__hint"
            >
              {{ t("compliance.preTrade.form.portfolioEmpty") }}
            </span>
            <span
              v-if="portfolios.error.value"
              class="pretrade-form__error"
              role="alert"
            >
              {{ t("compliance.preTrade.form.portfolioErrorPrefix") }}
              {{ portfolios.error.value }}
            </span>
            <span
              v-if="form.portfolio_id && hasPortfolioOptions"
              class="pretrade-form__hint"
            >
              <code>{{ shortenId(form.portfolio_id) }}</code>
            </span>
            <span
              v-if="touched.portfolio_id && errors.portfolio_id"
              id="err-portfolio-id"
              class="pretrade-form__error"
            >
              {{ errors.portfolio_id }}
            </span>
          </label>

          <label class="pretrade-form__field">
            <span class="pretrade-form__label">
              {{ t("compliance.preTrade.form.contract") }}
            </span>
            <select
              v-if="hasContractOptions"
              v-model="form.contract_id"
              class="form-control"
              aria-describedby="err-contract-id"
              @blur="markTouched('contract_id')"
            >
              <option value="" disabled>
                {{ t("compliance.preTrade.form.contractPlaceholder") }}
              </option>
              <option
                v-for="id in authorisedContractIds"
                :key="id"
                :value="id"
              >
                {{ shortenId(id) }}
              </option>
            </select>
            <input
              v-else
              v-model="form.contract_id"
              type="text"
              class="form-control"
              autocomplete="off"
              :placeholder="t('compliance.preTrade.form.contractIdFallback')"
              aria-describedby="err-contract-id"
              @blur="markTouched('contract_id')"
            />
            <span v-if="!hasContractOptions" class="pretrade-form__hint">
              {{ t("compliance.preTrade.form.contractEmpty") }}
            </span>
            <span
              v-if="form.contract_id && hasContractOptions"
              class="pretrade-form__hint"
            >
              <code>{{ form.contract_id }}</code>
            </span>
            <span
              v-if="touched.contract_id && errors.contract_id"
              id="err-contract-id"
              class="pretrade-form__error"
            >
              {{ errors.contract_id }}
            </span>
          </label>

          <label class="pretrade-form__field">
            <span class="pretrade-form__label">
              {{ t("compliance.preTrade.form.businessDate") }}
            </span>
            <input
              v-model="form.business_date"
              type="date"
              class="form-control"
              aria-describedby="err-business-date"
              @blur="markTouched('business_date')"
            />
            <span
              v-if="touched.business_date && errors.business_date"
              id="err-business-date"
              class="pretrade-form__error"
            >
              {{ errors.business_date }}
            </span>
          </label>
        </div>
      </div>

      <div class="pretrade-form__section">
        <h3 class="pretrade-form__heading">
          {{ t("compliance.preTrade.form.sectionOrder") }}
        </h3>
        <div class="pretrade-form__grid">
          <label class="pretrade-form__field">
            <span class="pretrade-form__label">
              {{ t("compliance.preTrade.form.orderId") }}
            </span>
            <div class="pretrade-form__inline">
              <input
                v-model="form.order_id"
                type="text"
                class="form-control"
                autocomplete="off"
                aria-describedby="err-order-id"
                @blur="markTouched('order_id')"
              />
              <AppButton
                type="button"
                variant="ghost"
                size="sm"
                @click="generateOrderId"
              >
                {{ t("compliance.preTrade.form.generate") }}
              </AppButton>
            </div>
            <span
              v-if="touched.order_id && errors.order_id"
              id="err-order-id"
              class="pretrade-form__error"
            >
              {{ errors.order_id }}
            </span>
          </label>

          <label class="pretrade-form__field">
            <span class="pretrade-form__label">
              {{ t("compliance.preTrade.form.ticker") }}
            </span>
            <input
              v-model="form.ticker"
              type="text"
              class="form-control"
              autocomplete="off"
              aria-describedby="err-ticker"
              @blur="markTouched('ticker')"
            />
            <span
              v-if="touched.ticker && errors.ticker"
              id="err-ticker"
              class="pretrade-form__error"
            >
              {{ errors.ticker }}
            </span>
          </label>

          <fieldset class="pretrade-form__field">
            <legend class="pretrade-form__label">
              {{ t("compliance.preTrade.form.side") }}
            </legend>
            <div class="pretrade-form__segmented">
              <button
                type="button"
                :class="['pretrade-form__segment-btn', { 'pretrade-form__segment-btn--active-buy': form.side === 'BUY' }]"
                @click="form.side = 'BUY'"
              >
                {{ t("compliance.preTrade.form.buy") }}
              </button>
              <button
                type="button"
                :class="['pretrade-form__segment-btn', { 'pretrade-form__segment-btn--active-sell': form.side === 'SELL' }]"
                @click="form.side = 'SELL'"
              >
                {{ t("compliance.preTrade.form.sell") }}
              </button>
            </div>
          </fieldset>

          <label class="pretrade-form__field">
            <span class="pretrade-form__label">
              {{ t("compliance.preTrade.form.quantity") }}
            </span>
            <input
              v-model="form.quantity"
              type="text"
              inputmode="decimal"
              class="form-control"
              autocomplete="off"
              aria-describedby="err-quantity"
              @blur="markTouched('quantity')"
            />
            <span
              v-if="touched.quantity && errors.quantity"
              id="err-quantity"
              class="pretrade-form__error"
            >
              {{ errors.quantity }}
            </span>
          </label>

          <label class="pretrade-form__field">
            <span class="pretrade-form__label">
              {{ t("compliance.preTrade.form.price") }}
            </span>
            <input
              v-model="form.price"
              type="text"
              inputmode="decimal"
              class="form-control"
              autocomplete="off"
              aria-describedby="err-price"
              @blur="markTouched('price')"
            />
            <span
              v-if="touched.price && errors.price"
              id="err-price"
              class="pretrade-form__error"
            >
              {{ errors.price }}
            </span>
          </label>

          <label class="pretrade-form__field">
            <span class="pretrade-form__label">
              {{ t("compliance.preTrade.form.currency") }}
            </span>
            <input
              v-model="form.currency"
              type="text"
              maxlength="3"
              class="form-control"
              autocomplete="off"
              aria-describedby="err-currency"
              @blur="markTouched('currency')"
            />
            <span
              v-if="selectedPortfolio?.base_currency && form.currency === selectedPortfolio.base_currency"
              class="pretrade-form__hint"
            >
              {{ t("compliance.preTrade.form.currencyAutofill") }}
            </span>
            <span
              v-if="touched.currency && errors.currency"
              id="err-currency"
              class="pretrade-form__error"
            >
              {{ errors.currency }}
            </span>
          </label>

          <label class="pretrade-form__field">
            <span class="pretrade-form__label">
              {{ t("compliance.preTrade.form.exchange") }}
            </span>
            <input
              v-model="form.exchange"
              type="text"
              class="form-control"
              autocomplete="off"
              aria-describedby="err-exchange"
              @blur="markTouched('exchange')"
            />
            <span
              v-if="touched.exchange && errors.exchange"
              id="err-exchange"
              class="pretrade-form__error"
            >
              {{ errors.exchange }}
            </span>
          </label>
        </div>
      </div>

      <div class="pretrade-form__footer">
        <AppButton
          variant="primary"
          size="md"
          :loading="loading"
          :disabled="loading"
          @click="submit"
        >
          {{ loading ? t("compliance.preTrade.form.running") : t("compliance.preTrade.form.submit") }}
        </AppButton>
        <span v-if="!isValid" class="pretrade-form__hint">
          {{ t("compliance.preTrade.form.invalid") }}
        </span>
      </div>
    </form>
  </AppCard>
</template>

<style scoped>
.pretrade-form {
  display: grid;
  gap: var(--space-7);
}

.pretrade-form__section {
  display: grid;
  gap: var(--space-4);
  border-bottom: 1px solid var(--border-subtle);
  padding-bottom: var(--space-6);
}

.pretrade-form__section:last-of-type {
  border-bottom: none;
  padding-bottom: 0;
}

.pretrade-form__heading {
  margin: 0;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
}

.pretrade-form__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--space-5);
}

.pretrade-form__field {
  display: grid;
  gap: var(--space-2);
  margin: 0;
  padding: 0;
  border: 0;
}

.pretrade-form__label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.pretrade-form__inline {
  display: flex;
  gap: var(--space-2);
}

.pretrade-form__inline .form-control {
  flex: 1;
}

.form-control {
  width: 100%;
  height: var(--size-control-md);
  padding: 0 var(--space-3);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: var(--font-size-sm);
  font-family: inherit;
  box-shadow: var(--shadow-xs);
  transition: all var(--transition-fast);
}

select.form-control {
  appearance: auto;
}

.form-control:hover {
  border-color: var(--border-strong);
}

.form-control:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.pretrade-form__segmented {
  display: flex;
  background: var(--bg-card-muted);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  padding: var(--space-1);
  height: var(--size-control-md);
  box-shadow: var(--shadow-xs);
}

.pretrade-form__segment-btn {
  flex: 1;
  text-align: center;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  color: var(--text-secondary);
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
  border: none;
  background: transparent;
  outline: none;
}

.pretrade-form__segment-btn:hover {
  color: var(--text-primary);
  background: var(--action-ghost-hover);
}

.pretrade-form__segment-btn--active-buy {
  background: var(--state-success) !important;
  color: var(--text-on-primary) !important;
  box-shadow: var(--shadow-sm);
}

.pretrade-form__segment-btn--active-sell {
  background: var(--state-danger) !important;
  color: var(--text-on-primary) !important;
  box-shadow: var(--shadow-sm);
}

.pretrade-form__hint {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.pretrade-form__hint code {
  font-family: var(--font-family-mono);
  background: var(--bg-card-muted);
  padding: 1px 3px;
  border-radius: var(--radius-xs);
}

.pretrade-form__error {
  color: var(--state-danger);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

.pretrade-form__footer {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  flex-wrap: wrap;
  padding-top: var(--space-4);
}
</style>
