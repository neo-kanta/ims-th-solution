<script setup lang="ts">
import AppDateField from "./AppDateField.vue";

interface Props {
  disabled?: boolean;
  error?: boolean;
}

withDefaults(defineProps<Props>(), {
  disabled: false,
  error: false,
});

const start = defineModel<string>("start", { default: "" });
const end = defineModel<string>("end", { default: "" });
</script>

<template>
  <div class="app-date-range" :class="{ 'is-invalid': error }">
    <AppDateField
      v-model="start"
      :disabled="disabled"
      :error="error"
      aria-label="Start date"
    />
    <span class="app-date-range__separator">to</span>
    <AppDateField
      v-model="end"
      :disabled="disabled"
      :error="error"
      aria-label="End date"
    />
  </div>
</template>

<style scoped>
.app-date-range {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  width: 100%;
}

.app-date-range__separator {
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-medium, 500);
}
</style>
