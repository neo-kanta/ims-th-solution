<script setup lang="ts">
interface Props {
  id?: string;
  placeholder?: string;
  rows?: number;
  disabled?: boolean;
  readonly?: boolean;
  error?: boolean;
}

withDefaults(defineProps<Props>(), {
  id: "",
  placeholder: "",
  rows: 3,
  disabled: false,
  readonly: false,
  error: false,
});

const modelValue = defineModel<string>({ default: "" });
</script>

<template>
  <textarea
    :id="id"
    v-model="modelValue"
    :placeholder="placeholder"
    :rows="rows"
    :disabled="disabled"
    :readonly="readonly"
    class="textarea"
    :class="{ 'is-invalid': error }"
    autocomplete="off"
  />
</template>

<style scoped>
.is-invalid {
  border-color: var(--alert-danger-border, #cf222e) !important;
}

.is-invalid:focus {
  border-color: var(--alert-danger-border, #cf222e) !important;
  box-shadow: 0 0 0 3px rgba(207, 34, 46, 0.15) !important;
}
</style>
