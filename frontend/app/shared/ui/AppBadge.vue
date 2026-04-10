<script setup lang="ts">
interface Props {
  variant?: 'success' | 'warning' | 'error' | 'info' | 'neutral' | 'locked' | 'draft' | 'purple' | 'orange' | 'teal'
  dot?: boolean
  size?: 'sm' | 'md'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'neutral',
  dot: false,
  size: 'md',
})

const dotColorMap: Record<NonNullable<Props['variant']>, string> = {
  success: '#059669',
  warning: '#d97706',
  error:   '#dc2626',
  info:    '#2563eb',
  neutral: '#94a3b8',
  locked:  '#64748b',
  draft:   '#94a3b8',
  purple:  '#7c3aed',
  orange:  '#ea580c',
  teal:    '#0d9488',
}
</script>

<template>
  <span
    class="badge"
    :class="[`badge-${variant}`, { 'badge-sm': size === 'sm' }]"
  >
    <span
      v-if="dot"
      class="badge-dot"
      :style="{ backgroundColor: dotColorMap[variant] }"
    />
    <slot />
  </span>
</template>

<style scoped>
.badge-sm {
  min-height: 20px;
  padding-left: var(--space-2);
  padding-right: var(--space-2);
  font-size: 10px;
}
</style>
