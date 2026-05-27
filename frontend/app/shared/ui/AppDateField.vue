<script setup lang="ts">
interface Props {
  id?: string;
  disabled?: boolean;
  readonly?: boolean;
  error?: boolean;
}

withDefaults(defineProps<Props>(), {
  id: "",
  disabled: false,
  readonly: false,
  error: false,
});

const modelValue = defineModel<string>({ default: "" });
</script>

<template>
  <input
    :id="id"
    v-model="modelValue"
    type="date"
    :disabled="disabled"
    :readonly="readonly"
    class="input app-date-input"
    :class="{ 'is-invalid': error }"
  />
</template>

<style scoped>
.app-date-input {
  /* Webkit adjustments for tabular layout */
  font-variant-numeric: tabular-nums;
}

.is-invalid {
  border-color: var(--alert-danger-border, #cf222e) !important;
}

.is-invalid:focus {
  border-color: var(--alert-danger-border, #cf222e) !important;
  box-shadow: 0 0 0 3px rgba(207, 34, 46, 0.15) !important;
}
</style>
