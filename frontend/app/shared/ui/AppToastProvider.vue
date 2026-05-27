<script setup lang="ts">
import { useAppToast } from "~/composables/useAppToast";
import AppIcon from "./AppIcon.vue";

const { toasts, dismiss } = useAppToast();
</script>

<template>
  <Teleport to="body">
    <div class="app-toast-container" aria-label="Notifications" role="log">
      <TransitionGroup name="toast-list">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="app-toast-item"
          :class="`app-toast-item--${toast.tone}`"
          :role="toast.tone === 'danger' ? 'alert' : 'status'"
          :aria-live="toast.tone === 'danger' ? 'assertive' : 'polite'"
        >
          <div class="app-toast-item__content">
            <AppIcon
              :name="
                toast.tone === 'success'
                  ? 'check'
                  : toast.tone === 'danger'
                  ? 'warning'
                  : toast.tone === 'warning'
                  ? 'warning'
                  : 'info'
              "
              size="sm"
              class="app-toast-item__icon"
            />
            <span class="app-toast-item__message">{{ toast.message }}</span>
          </div>
          <button
            type="button"
            class="app-toast-item__close"
            aria-label="Dismiss notification"
            @click="dismiss(toast.id)"
          >
            <AppIcon name="close" size="xs" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.app-toast-container {
  position: fixed;
  bottom: var(--space-6, 24px);
  right: var(--space-6, 24px);
  z-index: var(--z-toast, 2000);
  display: grid;
  gap: var(--space-3, 12px);
  width: 100%;
  max-width: 420px;
  pointer-events: none;
}

.app-toast-item {
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4, 16px);
  padding: var(--space-3, 12px) var(--space-4, 16px);
  background: var(--bg-card, #ffffff);
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  box-shadow: var(--shadow-lg, 0 8px 24px rgba(140, 149, 159, 0.2));
}

.app-toast-item--success {
  border-color: var(--alert-success-border, #1f883d);
  background: var(--alert-success-bg, #dafbe1);
  color: var(--alert-success-text, #1a7f37);
}

.app-toast-item--danger {
  border-color: var(--alert-danger-border, #cf222e);
  background: var(--alert-danger-bg, #ffebe9);
  color: var(--alert-danger-text, #d1242f);
}

.app-toast-item--warning {
  border-color: var(--alert-warning-border, #9a6700);
  background: var(--alert-warning-bg, #fff8c5);
  color: var(--alert-warning-text, #9a6700);
}

.app-toast-item--info {
  border-color: var(--alert-info-border, #0969da);
  background: var(--alert-info-bg, #ddf4ff);
  color: var(--alert-info-text, #0969da);
}

.app-toast-item__content {
  display: flex;
  align-items: center;
  gap: var(--space-3, 12px);
  min-width: 0;
}

.app-toast-item__icon {
  flex-shrink: 0;
}

.app-toast-item__message {
  font-size: var(--font-size-sm, 14px);
  font-weight: var(--font-weight-medium, 500);
  line-height: var(--line-height-normal, 1.4);
  word-break: break-word;
}

.app-toast-item__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  flex-shrink: 0;
  border-radius: var(--radius-sm, 4px);
  color: currentColor;
  background: transparent;
  border: none;
  cursor: pointer;
  opacity: 0.7;
  transition: opacity 0.15s ease, background-color 0.15s ease;
}

.app-toast-item__close:hover {
  opacity: 1;
  background: rgba(0, 0, 0, 0.06);
}

/* Animations */
.toast-list-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.toast-list-leave-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  position: absolute;
  width: 100%;
}

.toast-list-enter-from {
  opacity: 0;
  transform: translateY(24px) scale(0.96);
}

.toast-list-leave-to {
  opacity: 0;
  transform: translateY(16px) scale(0.96);
}

.toast-list-move {
  transition: transform 0.25s ease;
}
</style>
