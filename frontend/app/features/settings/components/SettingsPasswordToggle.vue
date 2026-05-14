<script setup lang="ts">
import { computed } from "vue";

interface Props {
  shown: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  toggle: [];
}>();

const { t } = useI18n();

const label = computed(() =>
  props.shown
    ? t("auth.hidePassword", "Hide password")
    : t("auth.showPassword", "Show password"),
);
</script>

<template>
  <button
    type="button"
    class="settings-password-toggle"
    :aria-label="label"
    :title="label"
    :aria-pressed="shown"
    @click="emit('toggle')"
  >
    <svg
      v-if="shown"
      xmlns="http://www.w3.org/2000/svg"
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
    >
      <path
        d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
      />
      <line x1="1" y1="1" x2="23" y2="23" />
    </svg>
    <svg
      v-else
      xmlns="http://www.w3.org/2000/svg"
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
    >
      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
  </button>
</template>

<style scoped>
.settings-password-toggle {
  position: absolute;
  top: 50%;
  right: var(--space-2);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  border-radius: var(--radius-sm);
  color: var(--text-tertiary);
  transform: translateY(-50%);
  background: transparent;
  border: 0;
  cursor: pointer;
}

.settings-password-toggle:hover {
  color: var(--text-primary);
  background: var(--action-ghost-hover);
}

.settings-password-toggle:focus-visible {
  outline: 2px solid var(--border-focus);
  outline-offset: 2px;
}
</style>
