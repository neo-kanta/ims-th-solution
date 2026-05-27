<script setup lang="ts">
import { computed } from "vue";

interface Props {
  amount?: number | string | null;
  currency?: string;
  precision?: number;
  loading?: boolean;
  accountingStyle?: boolean; // wraps negative numbers in parentheses e.g. (฿100.00)
}

const props = withDefaults(defineProps<Props>(), {
  amount: null,
  currency: "THB",
  precision: 2,
  loading: false,
  accountingStyle: false,
});

const formattedMoney = computed(() => {
  if (props.loading || props.amount === null || props.amount === undefined || props.amount === "") {
    return "—";
  }

  const num = typeof props.amount === "string" ? parseFloat(props.amount) : props.amount;
  if (Number.isNaN(num)) return "—";

  const absoluteValue = Math.abs(num);
  
  // Format absolute value
  const formatter = new Intl.NumberFormat(undefined, {
    style: "currency",
    currency: props.currency,
    minimumFractionDigits: props.precision,
    maximumFractionDigits: props.precision,
  });

  const formattedAbs = formatter.format(absoluteValue);

  if (num < 0) {
    return props.accountingStyle ? `(${formattedAbs})` : `-${formattedAbs}`;
  }
  return formattedAbs;
});

const isNegative = computed(() => {
  if (props.amount === null || props.amount === undefined || props.amount === "") return false;
  const num = typeof props.amount === "string" ? parseFloat(props.amount) : props.amount;
  return !Number.isNaN(num) && num < 0;
});
</script>

<template>
  <span
    class="app-money"
    :class="{
      'is-negative': isNegative,
      'is-loading': loading
    }"
  >
    {{ formattedMoney }}
  </span>
</template>

<style scoped>
.app-money {
  font-family: inherit;
  font-variant-numeric: tabular-nums;
  font-weight: var(--font-weight-medium, 500);
}

.is-negative {
  color: var(--state-danger, #cf222e);
}

.is-loading {
  opacity: 0.5;
}
</style>
