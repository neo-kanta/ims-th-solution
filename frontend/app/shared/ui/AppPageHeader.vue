<script setup lang="ts">
interface Props {
  title: string
  description?: string
}

defineProps<Props>()

defineSlots<{
  eyebrow(): unknown
  actions(): unknown
  tabs(): unknown
}>()
</script>

<template>
  <div>
    <div class="page-header-shell">
      <div class="page-header-main">
        <div v-if="$slots.eyebrow" class="page-header-eyebrow">
          <slot name="eyebrow" />
        </div>
        <h1 class="page-header-title">{{ title }}</h1>
        <p v-if="description" class="page-header-description">{{ description }}</p>
      </div>

      <div v-if="$slots.actions" class="page-header-actions">
        <slot name="actions" />
      </div>
    </div>
    <slot name="tabs" />
  </div>
</template>

<style scoped>
.page-header-shell {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-5);
  margin-bottom: var(--space-7);
}

.page-header-main {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

.page-header-eyebrow {
  min-width: 0;
}

.page-header-title {
  margin: 0;
  color: var(--text-primary);
  font-size: clamp(1.75rem, 2.4vw, 2.25rem);
  font-weight: var(--font-weight-semibold);
  line-height: 1.08;
  letter-spacing: -0.03em;
}

.page-header-description {
  margin: 0;
  max-width: 48rem;
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.page-header-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  flex-shrink: 0;
  flex-wrap: wrap;
}

@media (max-width: 1024px) {
  .page-header-shell {
    flex-direction: column;
    align-items: stretch;
  }
}

@media (max-width: 640px) {
  .page-header-actions {
    width: 100%;
  }
}
</style>
