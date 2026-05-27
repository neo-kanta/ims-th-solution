<script setup lang="ts">
import { computed } from "vue";

interface Props {
  value?: number | string | null;
  precision?: number;
  signed?: boolean;
  colorize?: boolean;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  value: null,
  precision: 2,
  signed: false,
  colorize: true,
  loading: false,
});

const formattedPercent = computed(() => {
  if (props.loading || props.value === null || props.value === undefined || props.value === "") {
    return "—";
  }

  const num = typeof props.value === "string" ? parseFloat(props.value) : props.value;
  if (Number.isNaN(num)) return "—";

  const absoluteValue = Math.abs(num);
  const formattedVal = absoluteValue.toFixed(props.precision) + "%";

  if (num < 0) {
    return `-${formattedVal}`;
  }
  if (num > 0 && props.signed) {
    return `+${formattedVal}`;
  }
  return formattedVal;
});

const isNegative = computed(() => {
  if (props.value === null || props.value === undefined || props.value === "") return false;
  const num = typeof props.value === "string" ? parseFloat(props.value) : props.value;
  return !Number.isNaN(num) && num < 0;
});

const isPositive = computed(() => {
  if (props.value === null || props.value === undefined || props.value === "") return false;
  const num = typeof props.value === "string" ? parseFloat(props.value) : props.value;
  return !Number.isNaN(num) && num > 0;
});
</script>

<template>
  <span
    class="app-percent"
    :class="{
      'is-negative': colorize && isNegative,
      'is-positive': colorize && isPositive,
      'is-loading': loading
    }"
  >
    {{ formattedPercent }}
  </span>
</template>

<style scoped>
.app-percent {
  font-family: inherit;
  font-variant-numeric: tabular-nums;
  font-weight: var(--font-weight-medium, 500);
}

.is-negative {
  color: var(--state-danger, #cf222e);
}

.is-positive {
  color: var(--state-success, #1f883d);
}

.is-loading {
  opacity: 0.5;
}
</style>
