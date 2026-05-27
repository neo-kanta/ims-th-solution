<script setup lang="ts">
import { computed } from "vue";
import AppLoadingState from "./AppLoadingState.vue";
import AppErrorState from "./AppErrorState.vue";

interface Props {
  title?: string;
  subtitle?: string;
  loading?: boolean;
  error?: string | null;
  width?: "full" | "contained" | "narrow";
}

const props = withDefaults(defineProps<Props>(), {
  title: "",
  subtitle: "",
  loading: false,
  error: null,
  width: "contained",
});

const emit = defineEmits<{
  retry: [];
}>();

const widthClass = computed(() => {
  return {
    "app-page--full": props.width === "full",
    "app-page--contained": props.width === "contained",
    "app-page--narrow": props.width === "narrow",
  };
});
</script>

<template>
  <div class="app-page" :class="widthClass">
    <!-- Header Block -->
    <header v-if="title || $slots.eyebrow || $slots.actions || $slots.tabs" class="app-page__header">
      <div class="app-page__header-top">
        <div class="app-page__header-main">
          <div v-if="$slots.eyebrow" class="app-page__eyebrow">
            <slot name="eyebrow" />
          </div>
          <h1 v-if="title" class="app-page__title">{{ title }}</h1>
          <p v-if="subtitle" class="app-page__subtitle">{{ subtitle }}</p>
        </div>
        <div v-if="$slots.actions" class="app-page__actions">
          <slot name="actions" />
        </div>
      </div>
      <div v-if="$slots.tabs" class="app-page__tabs">
        <slot name="tabs" />
      </div>
    </header>

    <!-- Content Block -->
    <div class="app-page__container">
      <div class="app-page__content">
        <!-- Loading State -->
        <AppLoadingState v-if="loading" />

        <!-- Error State -->
        <AppErrorState
          v-else-if="error"
          :message="error"
          retry
          @retry="emit('retry')"
        />

        <!-- Main Slot -->
        <slot v-else />
      </div>

      <!-- Optional Right Rail Sidebar -->
      <aside v-if="$slots['right-rail'] && !loading && !error" class="app-page__right-rail">
        <slot name="right-rail" />
      </aside>
    </div>
  </div>
</template>

<style scoped>
.app-page {
  width: 100%;
  margin-left: auto;
  margin-right: auto;
  display: grid;
  gap: var(--space-6, 24px);
  padding-bottom: var(--space-8, 32px);
}

.app-page--full {
  max-width: 100%;
}

.app-page--contained {
  max-width: 1280px;
}

.app-page--narrow {
  max-width: 800px;
}

.app-page__header {
  display: grid;
  gap: var(--space-3, 12px);
  border-bottom: 1px solid transparent;
}

.app-page__header-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-5, 20px);
}

.app-page__header-main {
  min-width: 0;
  display: grid;
  gap: var(--space-1, 4px);
}

.app-page__eyebrow {
  color: var(--text-tertiary, #6e7781);
  font-size: var(--font-size-xs, 12px);
  font-weight: var(--font-weight-semibold, 600);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.app-page__title {
  margin: 0;
  color: var(--text-primary, #1f2328);
  font-size: clamp(1.5rem, 2.2vw, 2rem);
  font-weight: var(--font-weight-semibold, 600);
  line-height: 1.15;
  letter-spacing: -0.02em;
}

.app-page__subtitle {
  margin: 0;
  color: var(--text-secondary, #57606a);
  font-size: var(--font-size-sm, 14px);
  line-height: 1.4;
}

.app-page__actions {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
  flex-shrink: 0;
  flex-wrap: wrap;
}

.app-page__tabs {
  margin-top: var(--space-2, 8px);
}

.app-page__container {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--space-6, 24px);
  align-items: start;
}

:has(.app-page__right-rail) > .app-page__container {
  grid-template-columns: minmax(0, 1.6fr) minmax(280px, 0.8fr);
}

.app-page__content {
  display: grid;
  gap: var(--space-5, 20px);
  min-width: 0;
}

.app-page__right-rail {
  display: grid;
  gap: var(--space-5, 20px);
  min-width: 0;
}

@media (max-width: 1024px) {
  .app-page__header-top {
    flex-direction: column;
    align-items: stretch;
  }
  .app-page__actions {
    width: 100%;
    margin-top: var(--space-2, 8px);
  }
  :has(.app-page__right-rail) > .app-page__container {
    grid-template-columns: 1fr;
  }
}
</style>
