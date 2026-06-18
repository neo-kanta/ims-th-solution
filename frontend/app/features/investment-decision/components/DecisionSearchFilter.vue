<script setup lang="ts">
import { ref } from "vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import AppDateField from "~/shared/ui/AppDateField.vue";

const emit = defineEmits<{
  search: [filters: Record<string, string>];
  reset: [];
}>();

const search = ref("");
const processType = ref("");
const productType = ref("");
const dateFrom = ref("");
const dateTo = ref("");

const PROCESS_TYPES = [
  { value: "", label: "All process types" },
  { value: "INVESTMENT_DECISION", label: "Investment Decision" },
  { value: "ORDER_CANCEL", label: "Order Cancel" },
  { value: "ORDER_AMEND", label: "Order Amend" },
];

const PRODUCT_TYPES = [
  { value: "", label: "All product types" },
  { value: "MUTUAL_FUND", label: "Mutual Fund" },
  { value: "ETF", label: "ETF" },
  { value: "STOCK", label: "Stock" },
  { value: "BOND", label: "Bond" },
  { value: "CASH", label: "Cash" },
];

function onSearch() {
  const filters: Record<string, string> = {};
  if (search.value.trim()) filters.search = search.value.trim();
  if (processType.value) filters.process_type = processType.value;
  if (productType.value) filters.product_type = productType.value;
  if (dateFrom.value) filters.business_date_from = dateFrom.value;
  if (dateTo.value) filters.business_date_to = dateTo.value;
  emit("search", filters);
}

function onReset() {
  search.value = "";
  processType.value = "";
  productType.value = "";
  dateFrom.value = "";
  dateTo.value = "";
  emit("reset");
}
</script>

<template>
  <form class="decision-filter" @submit.prevent="onSearch">
    <div class="decision-filter__row">
      <AppInput
        v-model="search"
        placeholder="Search by decision no, instrument, research no..."
        class="decision-filter__search"
      />
      <AppSelect v-model="processType" :options="PROCESS_TYPES" class="decision-filter__select" />
      <AppSelect v-model="productType" :options="PRODUCT_TYPES" class="decision-filter__select" />
      <AppDateField v-model="dateFrom" class="decision-filter__date" />
      <AppDateField v-model="dateTo" class="decision-filter__date" />
      <AppButton variant="primary" size="sm" type="submit">Search</AppButton>
      <AppButton variant="ghost" size="sm" type="button" @click="onReset">Reset</AppButton>
    </div>
  </form>
</template>

<style scoped>
.decision-filter__row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  align-items: flex-end;
}

.decision-filter__search {
  flex: 1 1 260px;
  min-width: 180px;
}

.decision-filter__select {
  flex: 0 0 180px;
}

.decision-filter__date {
  flex: 0 0 140px;
}
</style>
