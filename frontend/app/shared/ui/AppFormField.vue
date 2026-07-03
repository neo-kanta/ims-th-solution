<script setup lang="ts">
import { computed } from "vue";

interface Props {
  label?: string;
  id?: string;
  hint?: string;
  error?: string | null;
  required?: boolean;
  disabled?: boolean;
  horizontal?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  label: "",
  id: "",
  hint: "",
  error: null,
  required: false,
  disabled: false,
  horizontal: false,
});

const formFieldClass = computed(() => {
  return {
    "form-field--horizontal": props.horizontal,
    "form-field--error": !!props.error,
    "form-field--disabled": props.disabled,
  };
});
</script>

<template>
  <div class="form-field" :class="formFieldClass">
    <div v-if="label" class="form-field__label-zone">
      <label :for="id" class="label">
        <span>{{ label }}</span>
        <span v-if="required" class="form-field__required" aria-hidden="true">*</span>
      </label>
      <span v-if="hint && horizontal" class="form-field__hint">{{ hint }}</span>
    </div>

    <div class="form-field__control-zone">
      <slot :id="id" :disabled="disabled" :error="!!error" />
      
      <!-- Hint for vertical layout -->
      <span v-if="hint && !horizontal" class="form-field__hint">{{ hint }}</span>
      
      <!-- Error Message -->
      <Transition name="error-fade">
        <span v-if="error" class="form-field__error-msg" role="alert">
          {{ error }}
        </span>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.form-field {
  display: grid;
  gap: var(--space-2, 8px);
  width: 100%;
}

.form-field--horizontal {
  grid-template-columns: 12rem 1fr;
  align-items: start;
  gap: var(--space-4, 16px);
}

.form-field__label-zone {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.form-field__required {
  color: var(--state-danger, #cf222e);
  margin-left: var(--space-1, 4px);
  font-weight: bold;
}

.form-field__control-zone {
  display: flex;
  flex-direction: column;
  gap: var(--space-1, 4px);
}

.form-field__hint {
  font-size: var(--font-size-xs, 12px);
  color: var(--text-secondary, #57606a);
  line-height: 1.3;
}

.form-field__error-msg {
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-medium, 500);
  color: var(--state-danger, #cf222e);
  margin-top: 2px;
}

.form-field--disabled {
  opacity: 0.6;
}

/* Animations */
.error-fade-enter-active,
.error-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.error-fade-enter-from,
.error-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

@media (max-width: 640px) {
  .form-field--horizontal {
    grid-template-columns: 1fr;
    gap: var(--space-2, 8px);
  }
}
</style>
