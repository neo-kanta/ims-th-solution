<script setup lang="ts">
import { computed, ref, watch } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppButton from "~/shared/ui/AppButton.vue";
import AppDateField from "~/shared/ui/AppDateField.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import { cleanApprovalFilters } from "../lib/approvalLookup";
import type { DecisionApprovalFilters } from "../services/decisionApprovalApi";

const props = withDefaults(
  defineProps<{
    initialFilters?: DecisionApprovalFilters;
    loading?: boolean;
    /** Set when the last portfolio-code lookup failed (unknown/inaccessible code). */
    portfolioCodeError?: string | null;
  }>(),
  {
    initialFilters: () => ({}),
    loading: false,
    portfolioCodeError: null,
  },
);

const emit = defineEmits<{
  search: [filters: DecisionApprovalFilters];
  reset: [];
}>();

const { t } = useI18n();
const decisionNumber = ref("");
const search = ref("");
const processType = ref("");
const productType = ref("");
const dateFrom = ref("");
const dateTo = ref("");
const portfolioCode = ref("");

const processTypes = computed(() => [
  {
    value: "",
    label: t("approval.decisionWorkbench.lookup.allProcessTypes", "All process types"),
  },
  {
    value: "INVESTMENT_DECISION",
    label: t("approval.decisionWorkbench.lookup.investmentDecision", "Investment decision"),
  },
  {
    value: "ORDER_CANCEL",
    label: t("approval.decisionWorkbench.lookup.orderCancel", "Order cancellation"),
  },
  {
    value: "ORDER_AMEND",
    label: t("approval.decisionWorkbench.lookup.orderAmend", "Order amendment"),
  },
]);

const productTypes = computed(() => [
  {
    value: "",
    label: t("approval.decisionWorkbench.lookup.allProductTypes", "All product types"),
  },
  { value: "MUTUAL_FUND", label: t("approval.decisionWorkbench.lookup.mutualFund", "Mutual fund") },
  { value: "ETF", label: t("approval.decisionWorkbench.lookup.etf", "ETF") },
  { value: "STOCK", label: t("approval.decisionWorkbench.lookup.stock", "Stock") },
  { value: "BOND", label: t("approval.decisionWorkbench.lookup.bond", "Bond") },
  { value: "CASH", label: t("approval.decisionWorkbench.lookup.cash", "Cash") },
]);

function hydrate(filters: DecisionApprovalFilters) {
  decisionNumber.value = filters.decision_no ?? "";
  search.value = filters.search ?? "";
  processType.value = filters.process_type ?? "";
  productType.value = filters.product_type ?? "";
  dateFrom.value = filters.business_date_from ?? "";
  dateTo.value = filters.business_date_to ?? "";
  portfolioCode.value = filters.portfolio_code ?? "";
}

function onSearch() {
  emit(
    "search",
    cleanApprovalFilters({
      decision_no: decisionNumber.value,
      search: search.value,
      process_type: processType.value,
      product_type: productType.value,
      business_date_from: dateFrom.value,
      business_date_to: dateTo.value,
      portfolio_code: portfolioCode.value,
    }),
  );
}

function onReset() {
  hydrate({});
  emit("reset");
}

watch(
  () => props.initialFilters,
  (filters) => hydrate(filters),
  { deep: true, immediate: true },
);
</script>

<template>
  <form class="decision-filter" @submit.prevent="onSearch">
    <div class="decision-filter__primary">
      <label class="decision-filter__field decision-filter__field--decision" for="op02-decision-number">
        <span class="decision-filter__label">
          {{ t("approval.decisionWorkbench.lookup.decisionNumberLabel", "Decision number") }}
        </span>
        <AppInput
          id="op02-decision-number"
          v-model="decisionNumber"
          type="search"
          :disabled="loading"
          :placeholder="t(
            'approval.decisionWorkbench.lookup.decisionNumberPlaceholder',
            'e.g. DEC-20260715-0042',
          )"
        />
        <span class="decision-filter__help">
          {{
            t(
              "approval.decisionWorkbench.lookup.decisionNumberHelp",
              "Exact business decision number — no database UUID needed.",
            )
          }}
        </span>
      </label>

      <label class="decision-filter__field decision-filter__field--keyword" for="op02-keyword">
        <span class="decision-filter__label">
          {{ t("approval.decisionWorkbench.lookup.keywordLabel", "Instrument or research reference") }}
        </span>
        <AppInput
          id="op02-keyword"
          v-model="search"
          type="search"
          :disabled="loading"
          :placeholder="t(
            'approval.decisionWorkbench.lookup.keywordPlaceholder',
            'Instrument code or research number',
          )"
        />
      </label>

      <div class="decision-filter__buttons">
        <AppButton variant="primary" size="sm" type="submit" :loading="loading">
          {{ t("approval.decisionWorkbench.lookup.searchButton", "Query approvals") }}
        </AppButton>
        <AppButton variant="ghost" size="sm" type="button" :disabled="loading" @click="onReset">
          {{ t("approval.decisionWorkbench.lookup.resetButton", "Reset") }}
        </AppButton>
      </div>
    </div>

    <div class="decision-filter__secondary">
      <label class="decision-filter__field" for="op02-portfolio-code">
        <span class="decision-filter__label">
          {{ t("approval.decisionWorkbench.lookup.portfolioCodeLabel", "Portfolio code") }}
        </span>
        <AppInput
          id="op02-portfolio-code"
          v-model="portfolioCode"
          type="search"
          :disabled="loading"
          :placeholder="t(
            'approval.decisionWorkbench.lookup.portfolioCodePlaceholder',
            'e.g. PF-001',
          )"
        />
        <span v-if="portfolioCodeError" class="decision-filter__error" role="alert">
          {{ portfolioCodeError }}
        </span>
      </label>

      <label class="decision-filter__field" for="op02-process-type">
        <span class="decision-filter__label">
          {{ t("approval.decisionWorkbench.lookup.processTypeLabel", "Process") }}
        </span>
        <AppSelect
          id="op02-process-type"
          v-model="processType"
          :options="processTypes"
          :placeholder="''"
          :disabled="loading"
        />
      </label>

      <label class="decision-filter__field" for="op02-product-type">
        <span class="decision-filter__label">
          {{ t("approval.decisionWorkbench.lookup.productTypeLabel", "Product") }}
        </span>
        <AppSelect
          id="op02-product-type"
          v-model="productType"
          :options="productTypes"
          :placeholder="''"
          :disabled="loading"
        />
      </label>

      <label class="decision-filter__field" for="op02-date-from">
        <span class="decision-filter__label">
          {{ t("approval.decisionWorkbench.lookup.dateFromLabel", "Business date from") }}
        </span>
        <AppDateField id="op02-date-from" v-model="dateFrom" :disabled="loading" />
      </label>

      <label class="decision-filter__field" for="op02-date-to">
        <span class="decision-filter__label">
          {{ t("approval.decisionWorkbench.lookup.dateToLabel", "Business date to") }}
        </span>
        <AppDateField id="op02-date-to" v-model="dateTo" :disabled="loading" />
      </label>
    </div>
  </form>
</template>

<style scoped>
.decision-filter {
  display: grid;
  gap: var(--space-4, 16px);
}

.decision-filter__primary {
  display: grid;
  grid-template-columns: minmax(260px, 1.25fr) minmax(220px, 1fr) auto;
  gap: var(--space-3, 12px);
  align-items: end;
}

.decision-filter__secondary {
  display: grid;
  grid-template-columns: repeat(5, minmax(150px, 1fr));
  gap: var(--space-3, 12px);
  padding-top: var(--space-3, 12px);
  border-top: 1px solid var(--border-subtle, #d0d7de);
}

.decision-filter__error {
  color: var(--alert-danger-text, #cf222e);
  font-size: 11px;
  line-height: 1.35;
}

.decision-filter__field {
  display: grid;
  gap: var(--space-1, 4px);
  min-width: 0;
}

.decision-filter__label {
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
}

.decision-filter__help {
  color: var(--text-tertiary, #6e7781);
  font-size: 11px;
  line-height: 1.35;
}

.decision-filter__buttons {
  display: flex;
  gap: var(--space-2, 8px);
  align-items: center;
  padding-bottom: 17px;
}

@media (max-width: 1000px) {
  .decision-filter__primary {
    grid-template-columns: 1fr 1fr;
  }

  .decision-filter__buttons {
    grid-column: 1 / -1;
    padding-bottom: 0;
  }

  .decision-filter__secondary {
    grid-template-columns: repeat(2, minmax(150px, 1fr));
  }
}

@media (max-width: 640px) {
  .decision-filter__primary,
  .decision-filter__secondary {
    grid-template-columns: 1fr;
  }

  .decision-filter__buttons > * {
    flex: 1;
  }
}
</style>
