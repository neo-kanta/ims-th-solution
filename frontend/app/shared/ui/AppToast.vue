<script setup lang="ts">
import { onBeforeUnmount, watch } from "vue";
import AppIcon from "./AppIcon.vue";

interface Props {
  modelValue: boolean;
  message: string;
  tone?: "success" | "danger" | "info";
  duration?: number; // duration in ms, 0 means persist
}

const props = withDefaults(defineProps<Props>(), {
  tone: "success",
  duration: 5000,
});

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
  dismiss: [];
}>();

let timerId: number | null = null;

function startTimer() {
  clearTimer();
  if (props.duration > 0 && props.modelValue) {
    timerId = window.setTimeout(() => {
      dismissToast();
    }, props.duration);
  }
}

function clearTimer() {
  if (timerId !== null) {
    window.clearTimeout(timerId);
    timerId = null;
  }
}

function dismissToast() {
  clearTimer();
  emit("update:modelValue", false);
  emit("dismiss");
}

watch(
  () => props.modelValue,
  (val) => {
    if (val) {
      startTimer();
    } else {
      clearTimer();
    }
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  clearTimer();
});
</script>

<template>
  <Transition name="toast-slide">
    <div
      v-if="modelValue"
      class="app-toast"
      :class="`app-toast--${tone}`"
      role="alert"
      aria-live="assertive"
    >
      <div class="app-toast__content">
        <AppIcon
          :name="tone === 'danger' ? 'warning' : tone === 'success' ? 'check-circle' : 'info'"
          class="app-toast__icon"
        />
        <span class="app-toast__message">{{ message }}</span>
      </div>
      <button
        type="button"
        class="app-toast__btn"
        @click="dismissToast"
      >
        OK
      </button>
    </div>
  </Transition>
</template>

<style scoped>
.app-toast {
  position: fixed;
  bottom: var(--space-6);
  right: var(--space-6);
  z-index: var(--z-toast, 2000);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  min-width: 280px;
  max-width: 420px;
  padding: var(--space-3) var(--space-4);
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: 8px;
  box-shadow: var(--shadow-lg);
  pointer-events: auto;
}

.app-toast--success {
  border-color: var(--alert-success-border);
  background: var(--alert-success-bg);
  color: var(--alert-success-text);
}

.app-toast--danger {
  border-color: var(--alert-danger-border);
  background: var(--alert-danger-bg);
  color: var(--alert-danger-text);
}

.app-toast--info {
  border-color: var(--alert-info-border);
  background: var(--alert-info-bg);
  color: var(--alert-info-text);
}

.app-toast__content {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
}

.app-toast__icon {
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
}

.app-toast__message {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  line-height: var(--line-height-normal);
  word-break: break-word;
}

.app-toast__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: var(--space-9);
  padding: var(--space-1) var(--space-3);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
  border: 1px solid currentColor;
  border-radius: var(--radius-sm);
  background: transparent;
  color: currentColor;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.app-toast__btn:hover {
  opacity: 0.8;
}

.app-toast__btn:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: 2px;
}

/* Animations */
.toast-slide-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-slide-leave-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.toast-slide-enter-from {
  opacity: 0;
  transform: translateY(24px) scale(0.96);
}
.toast-slide-leave-to {
  opacity: 0;
  transform: translateY(16px) scale(0.96);
}
</style>
