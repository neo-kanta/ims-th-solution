<script setup lang="ts">
interface Props {
  title: string
  description?: string
}

defineProps<Props>()

defineSlots<{
  actions(): unknown
  tabs(): unknown
}>()
</script>

<template>
  <div>
    <div class="page-header">
      <div>
        <h1 class="page-title">{{ title }}</h1>
        <p v-if="description" class="page-desc">{{ description }}</p>
      </div>
      <div
        v-if="$slots.actions"
        style="display: flex; gap: 8px; align-items: center; flex-shrink: 0;"
      >
        <slot name="actions" />
      </div>
    </div>
    <slot name="tabs" />
  </div>
</template>
