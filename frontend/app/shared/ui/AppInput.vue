<script setup lang="ts">
interface Props {
  type?: "text" | "search" | "number" | "email" | "password" | "tel" | "url";
  id?: string;
  placeholder?: string;
  disabled?: boolean;
  readonly?: boolean;
  error?: boolean;
}

withDefaults(defineProps<Props>(), {
  type: "text",
  id: "",
  placeholder: "",
  disabled: false,
  readonly: false,
  error: false,
});

const modelValue = defineModel<string | number>({ default: "" });
</script>

<template>
  <input
    :id="id"
    v-model="modelValue"
    :type="type"
    :placeholder="placeholder"
    :disabled="disabled"
    :readonly="readonly"
    class="input"
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
