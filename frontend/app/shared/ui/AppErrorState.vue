<script setup lang="ts">
import { ref } from "vue";
import AppButton from "./AppButton.vue";
import AppIcon from "./AppIcon.vue";

interface Props {
  title?: string;
  message: string;
  retry?: boolean;
  details?: string;
}

withDefaults(defineProps<Props>(), {
  title: "An error occurred",
  message: "We encountered an error while processing your request.",
  retry: false,
  details: "",
});

const emit = defineEmits<{
  retry: [];
}>();

const showDetails = ref(false);
</script>

<template>
  <div class="app-error" role="alert">
    <div class="app-error__header">
      <div class="app-error__icon">
        <AppIcon name="warning" size="md" />
      </div>
      <div class="app-error__info">
        <h3 class="app-error__title">{{ title }}</h3>
        <p class="app-error__message">{{ message }}</p>
      </div>
    </div>

    <!-- Actions -->
    <div v-if="retry || details" class="app-error__actions">
      <AppButton
        v-if="retry"
        variant="secondary"
        size="sm"
        @click="emit('retry')"
      >
        Retry Action
      </AppButton>
      <button
        v-if="details"
        type="button"
        class="app-error__details-toggle"
        @click="showDetails = !showDetails"
      >
        {{ showDetails ? "Hide technical details" : "Show technical details" }}
      </button>
    </div>

    <!-- Collapsible Technical Details -->
    <Transition name="details-slide">
      <div v-if="details && showDetails" class="app-error__details">
        <pre><code>{{ details }}</code></pre>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.app-error {
  display: grid;
  gap: var(--space-4, 16px);
  padding: var(--space-5, 20px);
  border: 1px solid var(--alert-danger-border, #cf222e);
  border-radius: var(--radius-lg, 6px);
  background: var(--alert-danger-bg, #ffebe9);
  color: var(--alert-danger-text, #d1242f);
}

.app-error__header {
  display: flex;
  align-items: flex-start;
  gap: var(--space-4, 16px);
}

.app-error__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  background: rgba(207, 34, 46, 0.1);
  border-radius: var(--radius-md, 6px);
  color: currentColor;
}

.app-error__info {
  min-width: 0;
}

.app-error__title {
  margin: 0;
  font-size: var(--font-size-md, 16px);
  font-weight: var(--font-weight-semibold, 600);
}

.app-error__message {
  margin: var(--space-1, 4px) 0 0;
  font-size: var(--font-size-sm, 14px);
  opacity: 0.9;
  line-height: 1.4;
}

.app-error__actions {
  display: flex;
  align-items: center;
  gap: var(--space-4, 16px);
}

.app-error__details-toggle {
  font-family: inherit;
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-medium, 500);
  background: transparent;
  border: none;
  color: currentColor;
  text-decoration: underline;
  cursor: pointer;
  padding: var(--space-1, 4px) 0;
  opacity: 0.8;
}

.app-error__details-toggle:hover {
  opacity: 1;
}

.app-error__details {
  overflow-x: auto;
  padding: var(--space-3, 12px);
  border-radius: var(--radius-sm, 4px);
  background: rgba(0, 0, 0, 0.05);
  border: 1px solid rgba(0, 0, 0, 0.08);
  color: var(--text-primary, #1f2328);
  font-size: var(--font-size-xs, 12px);
  line-height: var(--line-height-relaxed, 1.6);
}

.app-error__details pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

/* Slide Transition */
.details-slide-enter-active,
.details-slide-leave-active {
  transition: max-height 0.2s ease, opacity 0.15s ease;
  overflow: hidden;
}
.details-slide-enter-from,
.details-slide-leave-to {
  opacity: 0;
  max-height: 0;
}
.details-slide-enter-to,
.details-slide-leave-from {
  opacity: 1;
  max-height: 200px;
}
</style>
