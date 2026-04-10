<script setup lang="ts">
interface Props {
  title?: string
  subtitle?: string
  flush?: boolean
  noBorder?: boolean
}

withDefaults(defineProps<Props>(), {
  flush: false,
  noBorder: false,
})

defineSlots<{
  'header-actions'(): unknown
  default(): unknown
  footer(): unknown
}>()
</script>

<template>
  <div class="card" :class="{ 'card--borderless': noBorder }">
    <div v-if="title || $slots['header-actions']" class="card-header">
      <div class="card-header-copy">
        <div v-if="title" class="card-title">{{ title }}</div>
        <div v-if="subtitle" class="card-subtitle">{{ subtitle }}</div>
      </div>
      <div v-if="$slots['header-actions']" class="card-header-actions">
        <slot name="header-actions" />
      </div>
    </div>

    <div :class="flush ? undefined : 'card-body'">
      <slot />
    </div>

    <div v-if="$slots.footer" class="card-footer">
      <slot name="footer" />
    </div>
  </div>
</template>
