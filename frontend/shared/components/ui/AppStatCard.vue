<script setup lang="ts">
interface Props {
  label: string
  value: string | number
  meta?: string
  trend?: 'up' | 'down' | 'flat'
  trendValue?: string
  variant?: 'default' | 'success' | 'warning' | 'error' | 'info'
}

withDefaults(defineProps<Props>(), {
  variant: 'default',
})

const variantBorderMap: Record<string, string> = {
  default: 'transparent',
  success: '#059669',
  warning: '#d97706',
  error:   '#dc2626',
  info:    '#2563eb',
}
</script>

<template>
  <div
    class="stat-card"
    :style="variant && variant !== 'default'
      ? `border-left: 3px solid ${variantBorderMap[variant]};`
      : undefined"
  >
    <div class="stat-card-label">{{ label }}</div>
    <div class="stat-card-value">{{ value }}</div>
    <div v-if="meta || trend" class="stat-card-meta">
      <template v-if="trend === 'up'">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="13"
          height="13"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#059669"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <polyline points="18 15 12 9 6 15" />
        </svg>
        <span v-if="trendValue" style="color: #059669; font-weight: 600;">{{ trendValue }}</span>
      </template>
      <template v-else-if="trend === 'down'">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="13"
          height="13"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#dc2626"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
        <span v-if="trendValue" style="color: #dc2626; font-weight: 600;">{{ trendValue }}</span>
      </template>
      <template v-else-if="trend === 'flat'">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="13"
          height="13"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#94a3b8"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
        <span v-if="trendValue" style="color: #94a3b8; font-weight: 600;">{{ trendValue }}</span>
      </template>
      <span v-if="meta">{{ meta }}</span>
    </div>
  </div>
</template>
