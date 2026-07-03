<script setup lang="ts">
import { computed } from "vue";

interface Props {
  title?: string;
  description?: string;
  variant?: "bordered" | "plain" | "elevated";
}

const props = withDefaults(defineProps<Props>(), {
  title: "",
  description: "",
  variant: "bordered",
});

const variantClass = computed(() => {
  return `app-section--${props.variant}`;
});
</script>

<template>
  <section class="app-section" :class="variantClass">
    <div v-if="title || description || $slots.actions" class="app-section__header">
      <div class="app-section__header-copy">
        <h2 v-if="title" class="app-section__title">{{ title }}</h2>
        <p v-if="description" class="app-section__description">{{ description }}</p>
      </div>
      <div v-if="$slots.actions" class="app-section__actions">
        <slot name="actions" />
      </div>
    </div>
    
    <div class="app-section__body">
      <slot />
    </div>
  </section>
</template>

<style scoped>
.app-section {
  display: grid;
  gap: var(--space-4, 16px);
  padding: var(--space-5, 20px);
}

.app-section--bordered {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 6px);
  background: var(--bg-card, #ffffff);
}

.app-section--elevated {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-lg, 6px);
  background: var(--bg-card, #ffffff);
  box-shadow: var(--shadow-sm, 0 1px 3px rgba(0,0,0,0.05));
}

.app-section--plain {
  padding: 0;
  background: transparent;
}

.app-section__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4, 16px);
}

.app-section__header-copy {
  min-width: 0;
}

.app-section__title {
  margin: 0;
  font-size: var(--font-size-md, 16px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #1f2328);
}

.app-section__description {
  margin: var(--space-1, 4px) 0 0;
  font-size: var(--font-size-xs, 12px);
  color: var(--text-secondary, #57606a);
  line-height: 1.4;
}

.app-section__actions {
  display: flex;
  gap: var(--space-2, 8px);
  align-items: center;
  flex-shrink: 0;
}

.app-section__body {
  min-width: 0;
}
</style>
