<script setup lang="ts">
interface Props {
  variant?: 'primary' | 'secondary' | 'danger' | 'warning' | 'success' | 'ghost'
  size?: 'xs' | 'sm' | 'md' | 'lg'
  loading?: boolean
  disabled?: boolean
  icon?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'secondary',
  size: 'md',
  loading: false,
  disabled: false,
  icon: false,
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

function handleClick(event: MouseEvent) {
  if (props.loading || props.disabled) return
  emit('click', event)
}

const iconSizeMap: Record<NonNullable<Props['size']>, string> = {
  xs: 'btn-icon-xs',
  sm: 'btn-icon-sm',
  md: 'btn-icon',
  lg: 'btn-icon',
}
</script>

<template>
  <button
    class="btn"
    :class="[
      `btn-${variant}`,
      `btn-${size}`,
      icon ? iconSizeMap[size] : undefined,
    ]"
    :disabled="disabled || loading"
    @click="handleClick"
  >
    <svg
      v-if="loading"
      xmlns="http://www.w3.org/2000/svg"
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2.5"
      stroke-linecap="round"
      stroke-linejoin="round"
      style="animation: spin 0.75s linear infinite; flex-shrink: 0;"
    >
      <path d="M21 12a9 9 0 1 1-6.219-8.56" />
      <style>@keyframes spin { to { transform: rotate(360deg); } }</style>
    </svg>
    <slot />
  </button>
</template>
