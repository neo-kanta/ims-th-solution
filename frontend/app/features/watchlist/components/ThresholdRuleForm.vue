<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import type { ThresholdRuleInput } from "../types";

interface Props {
  modelValue: ThresholdRuleInput;
  defaultCurrency?: string;
  index?: number;
}

const props = withDefaults(defineProps<Props>(), {
  defaultCurrency: "",
  index: 0,
});

const emit = defineEmits<{
  "update:modelValue": [value: ThresholdRuleInput];
}>();

const { t } = useI18n();

function update(field: keyof ThresholdRuleInput, value: unknown) {
  emit("update:modelValue", { ...props.modelValue, [field]: value });
}
</script>

<template>
  <div class="threshold-rule-form">
    <div class="form-grid-2">
      <div class="form-group">
        <label class="form-label form-label-req">{{ t('watchlist.rule.direction') }}</label>
        <select
          class="form-select"
          :value="modelValue.direction"
          @change="update('direction', ($event.target as HTMLSelectElement).value)"
        >
          <option value="ABOVE">{{ t('watchlist.direction.aboveDesc') }}</option>
          <option value="BELOW">{{ t('watchlist.direction.belowDesc') }}</option>
        </select>
      </div>
      <div class="form-group">
        <label class="form-label form-label-req">{{ t('watchlist.rule.threshold') }}</label>
        <input
          class="form-input"
          type="text"
          inputmode="decimal"
          placeholder="0.00"
          :value="modelValue.threshold_value"
          @input="update('threshold_value', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </div>
    <div class="form-grid-2">
      <div class="form-group">
        <label class="form-label">{{ t('watchlist.rule.currency') }}</label>
        <input
          class="form-input"
          type="text"
          placeholder="e.g. THB, USD"
          :value="modelValue.currency ?? defaultCurrency"
          @input="update('currency', ($event.target as HTMLInputElement).value || undefined)"
        />
      </div>
      <div class="form-group">
        <label class="form-label">{{ t('watchlist.rule.cooldown') }}</label>
        <input
          class="form-input"
          type="number"
          min="0"
          placeholder="60"
          :value="modelValue.cooldown_minutes ?? 60"
          @input="update('cooldown_minutes', Number(($event.target as HTMLInputElement).value))"
        />
      </div>
    </div>
    <div class="form-group">
      <label class="form-label">{{ t('watchlist.rule.status') }}</label>
      <select
        class="form-select"
        :value="modelValue.status ?? 'ENABLED'"
        @change="update('status', ($event.target as HTMLSelectElement).value)"
      >
        <option value="ENABLED">{{ t('watchlist.status.enabled') }}</option>
        <option value="DISABLED">{{ t('watchlist.status.disabled') }}</option>
      </select>
    </div>
  </div>
</template>

<style scoped>
.form-grid-2 {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}
@media (max-width: 640px) {
  .form-grid-2 {
    grid-template-columns: 1fr;
  }
}
</style>
