<script setup lang="ts">
import { computed } from "vue";

interface Props {
  title?: string;
  subtitle?: string;
  density?: "compact" | "comfortable";
  selected?: boolean;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  title: "",
  subtitle: "",
  density: "comfortable",
  selected: false,
  disabled: false,
});

const panelClass = computed(() => {
  return {
    "app-panel--compact": props.density === "compact",
    "app-panel--comfortable": props.density === "comfortable",
    "app-panel--selected": props.selected,
    "app-panel--disabled": props.disabled,
  };
});
</script>

<template>
  <div class="app-panel" :class="panelClass">
    <!-- Header -->
    <div v-if="title || subtitle || $slots['header-actions']" class="app-panel__header">
      <div class="app-panel__header-copy">
        <h3 v-if="title" class="app-panel__title">{{ title }}</h3>
        <p v-if="subtitle" class="app-panel__subtitle">{{ subtitle }}</p>
      </div>
      <div v-if="$slots['header-actions']" class="app-panel__header-actions">
        <slot name="header-actions" />
      </div>
    </div>

    <!-- Body -->
    <div class="app-panel__body">
      <slot />
    </div>

    <!-- Footer -->
    <div v-slot:footer v-if="$slots.footer" class="app-panel__footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<style scoped>
.app-panel {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card, #ffffff);
  overflow: hidden;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
  display: flex;
  flex-direction: column;
}

.app-panel--comfortable .app-panel__header,
.app-panel--comfortable .app-panel__body,
.app-panel--comfortable .app-panel__footer {
  padding: var(--space-5, 20px);
}

.app-panel--compact .app-panel__header,
.app-panel--compact .app-panel__body,
.app-panel--compact .app-panel__footer {
  padding: var(--space-3, 12px) var(--space-4, 16px);
}

.app-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4, 16px);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}

.app-panel__header-copy {
  min-width: 0;
}

.app-panel__title {
  margin: 0;
  font-size: var(--font-size-md, 16px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--text-primary, #1f2328);
}

.app-panel__subtitle {
  margin: var(--space-1, 4px) 0 0;
  font-size: var(--font-size-xs, 12px);
  color: var(--text-secondary, #57606a);
  line-height: 1.4;
}

.app-panel__header-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  flex-shrink: 0;
}

.app-panel__body {
  flex-grow: 1;
  min-width: 0;
}

.app-panel__footer {
  border-top: 1px solid var(--border-subtle, #d0d7de);
  background: var(--bg-card-muted, #f6f8fa);
}

.app-panel--selected {
  border-color: var(--border-focus, #0969da);
  box-shadow: 0 0 0 1px var(--border-focus, #0969da);
}

.app-panel--disabled {
  opacity: 0.6;
  cursor: not-allowed;
  pointer-events: none;
}
</style>
