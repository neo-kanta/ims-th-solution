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
  <div class="card" :style="noBorder ? 'border: none; box-shadow: none;' : undefined">
    <div
      v-if="title || $slots['header-actions']"
      class="card-header"
    >
      <div>
        <div v-if="title" class="card-title">{{ title }}</div>
        <div v-if="subtitle" class="card-subtitle">{{ subtitle }}</div>
      </div>
      <div v-if="$slots['header-actions']" style="flex-shrink: 0; display: flex; align-items: center; gap: 6px;">
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
