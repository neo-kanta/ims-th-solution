<script setup lang="ts">
interface RadioOption {
  value: any;
  label: string;
  disabled?: boolean;
}

interface Props {
  options: RadioOption[];
  name: string;
  disabled?: boolean;
  error?: boolean;
  layout?: "horizontal" | "vertical";
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  error: false,
  layout: "vertical",
});

const modelValue = defineModel<any>();
</script>

<template>
  <div
    class="app-radio-group"
    :class="[
      `app-radio-group--${layout}`,
      { 'has-error': error }
    ]"
    role="radiogroup"
  >
    <div
      v-for="opt in options"
      :key="opt.value"
      class="app-radio-group__item"
      :class="{ 'is-disabled': disabled || opt.disabled }"
    >
      <input
        :id="`${name}-${opt.value}`"
        v-model="modelValue"
        type="radio"
        :name="name"
        :value="opt.value"
        :disabled="disabled || opt.disabled"
        class="radio"
      />
      <label :for="`${name}-${opt.value}`" class="app-radio-group__label">
        {{ opt.label }}
      </label>
    </div>
  </div>
</template>

<style scoped>
.app-radio-group {
  display: flex;
  gap: var(--space-3, 12px);
}

.app-radio-group--vertical {
  flex-direction: column;
}

.app-radio-group--horizontal {
  flex-direction: row;
  flex-wrap: wrap;
}

.app-radio-group__item {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  user-select: none;
}

.app-radio-group__label {
  font-size: var(--font-size-sm, 14px);
  color: var(--text-primary, #1f2328);
  cursor: pointer;
}

.is-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.is-disabled .app-radio-group__label {
  cursor: not-allowed;
}

.has-error .app-radio-group__label {
  color: var(--state-danger, #cf222e);
}
</style>
