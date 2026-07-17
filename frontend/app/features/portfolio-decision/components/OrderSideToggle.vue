<script setup lang="ts">
import type { OrderSide } from "~/features/investment-ledger/lib/ledgerFormat";
import { useI18n } from "~/composables/useI18n";

defineProps<{
  modelValue: OrderSide;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: OrderSide];
}>();

const { t } = useI18n();
</script>

<template>
  <div class="side-toggle" role="radiogroup" :aria-label="t('portfolio.decisionNew.side')">
    <button
      type="button"
      role="radio"
      :aria-checked="modelValue === 'BUY'"
      class="side-toggle__btn side-toggle__btn--buy"
      :class="{ 'is-active': modelValue === 'BUY' }"
      :disabled="disabled"
      @click="emit('update:modelValue', 'BUY')"
    >
      {{ t("portfolio.decisionNew.buy") }}
    </button>
    <button
      type="button"
      role="radio"
      :aria-checked="modelValue === 'SELL'"
      class="side-toggle__btn side-toggle__btn--sell"
      :class="{ 'is-active': modelValue === 'SELL' }"
      :disabled="disabled"
      @click="emit('update:modelValue', 'SELL')"
    >
      {{ t("portfolio.decisionNew.sell") }}
    </button>
  </div>
</template>

<style scoped>
.side-toggle {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  padding: 4px;
  max-width: 260px;
}

.side-toggle__btn {
  border: none;
  background: transparent;
  padding: 10px 12px;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--text-secondary, #57606a);
  cursor: pointer;
  border-radius: var(--radius-sm, 4px);
  transition: background 0.1s, color 0.1s;
}

.side-toggle__btn:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.side-toggle__btn--buy.is-active {
  background: var(--state-success, #1a7f37);
  color: #fff;
}

.side-toggle__btn--sell.is-active {
  background: var(--state-danger, #cf222e);
  color: #fff;
}
</style>
