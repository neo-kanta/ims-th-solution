<script setup lang="ts">
import { computed, onMounted, reactive, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppCard from "~/shared/ui/AppCard.vue";

import { useCompliancePortfolioDirectory } from "../composables/useCompliancePortfolioDirectory";
import type { ComplianceBreachListFilters, ComplianceBreachStatus } from "../types";

interface Props {
  modelValue: ComplianceBreachListFilters;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  "update:modelValue": [value: ComplianceBreachListFilters];
  apply: [];
  reset: [];
}>();

const { t } = useI18n();
const authStore = useAuthStore();
const portfolios = useCompliancePortfolioDirectory();

type StatusChoice = "any" | ComplianceBreachStatus;

interface FormState {
  status: StatusChoice;
  ruleTypeId: string;
  dateFrom: string;
  dateTo: string;
  portfolioId: string;
  contractId: string;
}

function initialFrom(filters: ComplianceBreachListFilters): FormState {
  return {
    status: (filters.status as StatusChoice) ?? "any",
    ruleTypeId: filters.rule_type_id ?? "",
    dateFrom: filters.date_from ?? "",
    dateTo: filters.date_to ?? "",
    portfolioId: filters.portfolio_id ?? "",
    contractId: filters.contract_id ?? "",
  };
}

const form = reactive<FormState>(initialFrom(props.modelValue));

watch(
  () => props.modelValue,
  (next) => {
    const fresh = initialFrom(next);
    form.status = fresh.status;
    form.ruleTypeId = fresh.ruleTypeId;
    form.dateFrom = fresh.dateFrom;
    form.dateTo = fresh.dateTo;
    form.portfolioId = fresh.portfolioId;
    form.contractId = fresh.contractId;
  },
  { deep: true },
);

// Auth store exposes the contracts the user is authorised to see; we present
// them as a select so the filter never asks for a raw UUID when scoped.
const authorisedContractIds = computed<string[]>(() => {
  const list = authStore.permissions?.contracts;
  return Array.isArray(list) ? list : [];
});

const hasPortfolioOptions = computed(() => portfolios.items.value.length > 0);
const hasContractOptions = computed(
  () => authorisedContractIds.value.length > 0,
);

function shortenId(id: string, len = 8): string {
  if (!id) return "";
  return id.length > 12 ? `${id.slice(0, len)}…${id.slice(-4)}` : id;
}

function buildFilters(): ComplianceBreachListFilters {
  const next: ComplianceBreachListFilters = {};
  if (form.status !== "any") next.status = form.status;
  if (form.ruleTypeId.trim()) next.rule_type_id = form.ruleTypeId.trim();
  if (form.dateFrom) next.date_from = form.dateFrom;
  if (form.dateTo) next.date_to = form.dateTo;
  if (form.portfolioId.trim()) next.portfolio_id = form.portfolioId.trim();
  if (form.contractId.trim()) next.contract_id = form.contractId.trim();
  return next;
}

function apply() {
  emit("update:modelValue", buildFilters());
  emit("apply");
}

function reset() {
  form.status = "any";
  form.ruleTypeId = "";
  form.dateFrom = "";
  form.dateTo = "";
  form.portfolioId = "";
  form.contractId = "";
  emit("update:modelValue", {});
  emit("reset");
}

onMounted(() => {
  void portfolios.ensureLoaded();
});
</script>

<template>
  <AppCard :title="t('compliance.postTrade.filters.title')">
    <form class="breach-filters" @submit.prevent="apply">
      <label class="breach-filters__field">
        <span class="breach-filters__label">
          {{ t("compliance.postTrade.filters.status") }}
        </span>
        <select v-model="form.status" class="form-control">
          <option value="any">{{ t("compliance.postTrade.filters.any") }}</option>
          <option value="OPEN">{{ t("compliance.postTrade.filters.open") }}</option>
          <option value="OVERRIDDEN">
            {{ t("compliance.postTrade.filters.overridden") }}
          </option>
          <option value="RESOLVED">
            {{ t("compliance.postTrade.filters.resolved") }}
          </option>
        </select>
      </label>

      <label class="breach-filters__field">
        <span class="breach-filters__label">
          {{ t("compliance.postTrade.filters.ruleTypeId") }}
        </span>
        <input
          v-model="form.ruleTypeId"
          type="text"
          class="form-control"
          placeholder="e.g. concentration.single_issuer"
        />
      </label>

      <div class="breach-filters__row">
        <label class="breach-filters__field">
          <span class="breach-filters__label">
            {{ t("compliance.postTrade.filters.dateFrom") }}
          </span>
          <input v-model="form.dateFrom" type="date" class="form-control" />
        </label>
        <label class="breach-filters__field">
          <span class="breach-filters__label">
            {{ t("compliance.postTrade.filters.dateTo") }}
          </span>
          <input v-model="form.dateTo" type="date" class="form-control" />
        </label>
      </div>

      <label class="breach-filters__field">
        <span class="breach-filters__label">
          {{ t("compliance.postTrade.filters.portfolio") }}
        </span>
        <select
          v-if="hasPortfolioOptions"
          v-model="form.portfolioId"
          class="form-control"
        >
          <option value="">{{ t("compliance.postTrade.filters.portfolioAny") }}</option>
          <option v-for="p in portfolios.items.value" :key="p.id" :value="p.id">
            {{ p.code }} — {{ p.name }} ({{ p.base_currency }})
          </option>
        </select>
        <input
          v-else
          v-model="form.portfolioId"
          type="text"
          class="form-control"
          :placeholder="
            portfolios.loading.value
              ? t('compliance.postTrade.filters.portfolioLoading')
              : t('compliance.postTrade.filters.portfolioId')
          "
        />
        <span
          v-if="form.portfolioId && hasPortfolioOptions"
          class="breach-filters__hint"
        >
          <code>{{ shortenId(form.portfolioId) }}</code>
        </span>
      </label>

      <label class="breach-filters__field">
        <span class="breach-filters__label">
          {{ t("compliance.postTrade.filters.contract") }}
        </span>
        <select
          v-if="hasContractOptions"
          v-model="form.contractId"
          class="form-control"
        >
          <option value="">{{ t("compliance.postTrade.filters.contractAny") }}</option>
          <option v-for="id in authorisedContractIds" :key="id" :value="id">
            {{ shortenId(id) }}
          </option>
        </select>
        <input
          v-else
          v-model="form.contractId"
          type="text"
          class="form-control"
          :placeholder="t('compliance.postTrade.filters.contractId')"
        />
      </label>

      <div class="breach-filters__actions">
        <AppButton variant="primary" size="sm" :loading="loading" @click="apply">
          {{ t("compliance.postTrade.filters.apply") }}
        </AppButton>
        <AppButton variant="ghost" size="sm" @click="reset">
          {{ t("compliance.postTrade.filters.reset") }}
        </AppButton>
      </div>
    </form>
  </AppCard>
</template>

<style scoped>
.breach-filters {
  display: grid;
  gap: var(--space-4);
}

.breach-filters__field {
  display: grid;
  gap: var(--space-2);
  margin: 0;
}

.breach-filters__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-3);
}

.breach-filters__label {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
  letter-spacing: 0.04em;
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
}

.form-control:focus {
  outline: none;
  border-color: var(--border-focus);
  box-shadow: var(--shadow-focus);
}

.breach-filters__hint {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.breach-filters__hint code {
  font-family: var(--font-family-mono);
}

.breach-filters__actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}
</style>
