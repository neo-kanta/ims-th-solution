<script setup lang="ts">
interface Props {
  id?: string;
  label?: string;
  disabled?: boolean;
  error?: boolean;
}

withDefaults(defineProps<Props>(), {
  id: "",
  label: "",
  disabled: false,
  error: false,
});

const modelValue = defineModel<boolean>({ default: false });
</script>

<template>
  <div class="app-checkbox" :class="{ 'is-disabled': disabled, 'has-error': error }">
    <input
      :id="id"
      v-model="modelValue"
      type="checkbox"
      :disabled="disabled"
      class="checkbox"
    />
    <label v-if="label" :for="id" class="app-checkbox__label">
      {{ label }}
    </label>
  </div>
</template>

<style scoped>
.app-checkbox {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  cursor: pointer;
  user-select: none;
}

.app-checkbox__label {
  font-size: var(--font-size-sm, 14px);
  color: var(--text-primary, #1f2328);
  cursor: inherit;
}

.is-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.has-error .app-checkbox__label {
  color: var(--state-danger, #cf222e);
}
</style>
