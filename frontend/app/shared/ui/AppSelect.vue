<script setup lang="ts">
import { computed } from "vue";

interface OptionObject {
  value: any;
  label: string;
  disabled?: boolean;
}

type SelectOption = OptionObject | string;

interface Props {
  options: SelectOption[];
  placeholder?: string;
  id?: string;
  disabled?: boolean;
  error?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: "Select an option...",
  id: "",
  disabled: false,
  error: false,
});

const modelValue = defineModel<any>({ default: "" });

const normalizedOptions = computed<OptionObject[]>(() => {
  return props.options.map((opt) => {
    if (typeof opt === "string") {
      return { value: opt, label: opt };
    }
    return opt;
  });
});
</script>

<template>
  <select
    :id="id"
    v-model="modelValue"
    :disabled="disabled"
    class="select"
    :class="{ 'is-invalid': error }"
  >
    <option v-if="placeholder" value="" disabled>{{ placeholder }}</option>
    <option
      v-for="opt in normalizedOptions"
      :key="opt.value"
      :value="opt.value"
      :disabled="opt.disabled"
    >
      {{ opt.label }}
    </option>
  </select>
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
