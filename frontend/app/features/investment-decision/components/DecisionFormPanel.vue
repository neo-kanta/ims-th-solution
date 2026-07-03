<script setup lang="ts">
import type { ApiCreateDecisionRequest } from "../services/decisionApi";

const props = defineProps<{
  modelValue: Partial<ApiCreateDecisionRequest>;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: Partial<ApiCreateDecisionRequest>): void;
}>();

function update(
  field: keyof ApiCreateDecisionRequest,
  value: string,
) {
  emit("update:modelValue", { ...props.modelValue, [field]: value || undefined });
}
</script>

<template>
  <div class="panel-header">New Decision</div>
  <div class="legacy-form">
    <div class="form-row">
      <div class="form-field form-field--double">
        <label>Instrument Code <span class="required">*</span></label>
        <input
          type="text"
          placeholder="e.g. PTT, AOT"
          :value="modelValue.instrument_code ?? ''"
          @input="update('instrument_code', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="form-field">
        <label>Side <span class="required">*</span></label>
        <select
          :value="modelValue.side ?? ''"
          @change="update('side', ($event.target as HTMLSelectElement).value)"
        >
          <option value="">— Select —</option>
          <option value="BUY">BUY</option>
          <option value="SELL">SELL</option>
        </select>
      </div>
    </div>

    <div class="form-row">
      <div class="form-field">
        <label>Quantity</label>
        <input
          type="text"
          placeholder="0"
          :value="modelValue.quantity ?? ''"
          @input="update('quantity', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="form-field">
        <label>Amount</label>
        <input
          type="text"
          placeholder="0.00"
          :value="modelValue.amount ?? ''"
          @input="update('amount', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="form-field">
        <label>Limit Price</label>
        <input
          type="text"
          placeholder="0.00"
          :value="modelValue.limit_price ?? ''"
          @input="update('limit_price', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </div>

    <div class="form-row">
      <div class="form-field">
        <label>Currency <span class="required">*</span></label>
        <input
          type="text"
          placeholder="THB"
          maxlength="3"
          :value="modelValue.currency ?? ''"
          @input="update('currency', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="form-field form-field--double">
        <label>Exchange</label>
        <input
          type="text"
          placeholder="SET"
          :value="modelValue.exchange ?? ''"
          @input="update('exchange', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </div>

    <div class="form-row">
      <div class="form-field">
        <label>Business Date <span class="required">*</span></label>
        <input
          type="date"
          :value="modelValue.business_date ?? ''"
          @input="update('business_date', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </div>

    <div class="form-row">
      <div class="form-field form-field--double">
        <label>Rationale</label>
        <textarea
          rows="3"
          placeholder="Investment thesis…"
          :value="modelValue.rationale ?? ''"
          @input="update('rationale', ($event.target as HTMLTextAreaElement).value)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel-header {
  background: var(--bg-card-hover);
  border-bottom: 1px solid var(--border-subtle);
  padding: 8px 12px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.legacy-form {
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow-y: auto;
  flex: 1;
}

.form-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
}

.form-field--double {
  flex: 2;
}

.form-field label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
}

.form-field input,
.form-field select,
.form-field textarea {
  padding: 5px 8px;
  font-size: 13px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm, 4px);
  background: var(--bg-card);
  color: var(--text-primary);
  width: 100%;
  box-sizing: border-box;
  resize: none;
}

.form-field input:focus,
.form-field select:focus,
.form-field textarea:focus {
  outline: none;
  border-color: var(--border-focus);
}

.required {
  color: var(--state-danger, #cf222e);
}
</style>
