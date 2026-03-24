<script setup lang="ts">
interface Props {
  open: boolean
  title: string
  subtitle?: string
  size?: 'default' | 'lg' | 'xl'
}

withDefaults(defineProps<Props>(), {
  size: 'default',
})

const emit = defineEmits<{
  close: []
}>()

defineSlots<{
  default(): unknown
  footer(): unknown
}>()

function close() {
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <template v-if="open">
      <!-- Overlay -->
      <div class="drawer-overlay" @click="close" />

      <!-- Panel -->
      <div
        class="drawer"
        :class="{
          'drawer-lg': size === 'lg',
          'drawer-xl': size === 'xl',
        }"
        role="dialog"
        aria-modal="true"
        :aria-label="title"
      >
        <!-- Header -->
        <div class="drawer-header">
          <div>
            <div class="drawer-title">{{ title }}</div>
            <div v-if="subtitle" class="drawer-subtitle">{{ subtitle }}</div>
          </div>
          <button class="btn btn-ghost btn-icon" type="button" @click="close" aria-label="Close">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <!-- Body -->
        <div class="drawer-body">
          <slot />
        </div>

        <!-- Footer -->
        <div class="drawer-footer">
          <slot name="footer">
            <button class="btn btn-secondary btn-sm" type="button" @click="close">Close</button>
          </slot>
        </div>
      </div>
    </template>
  </Teleport>
</template>
