<script setup lang="ts">
interface Props {
  message?: string;
  skeleton?: boolean;
}

withDefaults(defineProps<Props>(), {
  message: "Loading data...",
  skeleton: false,
});
</script>

<template>
  <div class="app-loading" :class="{ 'app-loading--skeleton': skeleton }">
    <template v-if="skeleton">
      <div class="app-loading__skeleton-item app-loading__skeleton-title"></div>
      <div class="app-loading__skeleton-item app-loading__skeleton-text"></div>
      <div class="app-loading__skeleton-item app-loading__skeleton-text w-5/6"></div>
      <div class="app-loading__skeleton-item app-loading__skeleton-text w-2/3"></div>
    </template>
    <template v-else>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="32"
        height="32"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="app-loading__spinner"
      >
        <path d="M21 12a9 9 0 1 1-6.219-8.56" />
      </svg>
      <span v-if="message" class="app-loading__message">{{ message }}</span>
    </template>
  </div>
</template>

<style scoped>
.app-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4, 16px);
  padding: var(--space-8, 32px);
  color: var(--text-secondary, #57606a);
}

.app-loading__spinner {
  animation: spin 0.8s linear infinite;
  color: var(--action-primary, #0969da);
}

.app-loading__message {
  font-size: var(--font-size-sm, 14px);
  font-weight: var(--font-weight-medium, 500);
}

.app-loading--skeleton {
  align-items: stretch;
  gap: var(--space-3, 12px);
  padding: var(--space-4, 16px);
}

.app-loading__skeleton-item {
  background: var(--bg-card-muted, #f6f8fa);
  border-radius: var(--radius-sm, 4px);
  animation: pulse 1.5s ease-in-out infinite;
}

.app-loading__skeleton-title {
  height: 1.5rem;
  width: 35%;
  margin-bottom: var(--space-2, 8px);
}

.app-loading__skeleton-text {
  height: 1rem;
}

.w-5\/6 {
  width: 83.333333%;
}

.w-2\/3 {
  width: 66.666667%;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 0.6;
  }
  50% {
    opacity: 1;
  }
}
</style>
