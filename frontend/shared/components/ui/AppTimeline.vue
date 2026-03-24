<script setup lang="ts">
interface TimelineStep {
  key: string
  label: string
  status: 'done' | 'active' | 'pending' | 'error' | 'locked' | 'warn'
  timestamp?: string
  actor?: string
  note?: string
}

interface Props {
  steps: TimelineStep[]
}

defineProps<Props>()
</script>

<template>
  <div class="timeline">
    <div
      v-for="(step, index) in steps"
      :key="step.key"
      class="timeline-item"
    >
      <!-- Spine: dot + connector -->
      <div class="timeline-spine">
        <div class="timeline-dot" :class="`tl-${step.status}`">
          <!-- done: checkmark -->
          <svg
            v-if="step.status === 'done'"
            xmlns="http://www.w3.org/2000/svg"
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="#059669"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <polyline points="20 6 9 17 4 12" />
          </svg>

          <!-- active: filled blue circle -->
          <svg
            v-else-if="step.status === 'active'"
            xmlns="http://www.w3.org/2000/svg"
            width="10"
            height="10"
            viewBox="0 0 24 24"
          >
            <circle cx="12" cy="12" r="10" fill="#2563eb" />
          </svg>

          <!-- error: X -->
          <svg
            v-else-if="step.status === 'error'"
            xmlns="http://www.w3.org/2000/svg"
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="#dc2626"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>

          <!-- locked: padlock -->
          <svg
            v-else-if="step.status === 'locked'"
            xmlns="http://www.w3.org/2000/svg"
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="#64748b"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <rect x="5" y="11" width="14" height="10" rx="2" ry="2" />
            <path d="M8 11V7a4 4 0 0 1 8 0v4" />
          </svg>

          <!-- warn: exclamation -->
          <svg
            v-else-if="step.status === 'warn'"
            xmlns="http://www.w3.org/2000/svg"
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="#d97706"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <line x1="12" y1="8" x2="12" y2="13" />
            <line x1="12" y1="17" x2="12.01" y2="17" />
          </svg>

          <!-- pending: empty circle (no inner icon, dot provides border) -->
        </div>

        <!-- Connector line between items -->
        <div
          v-if="index < steps.length - 1"
          class="timeline-connector"
          :class="{ 'tl-done': step.status === 'done' }"
        />
      </div>

      <!-- Content -->
      <div class="timeline-content">
        <div style="display: flex; align-items: baseline; gap: 8px; flex-wrap: wrap;">
          <span style="font-size: 13px; font-weight: 600; color: #0f172a;">{{ step.label }}</span>
          <span v-if="step.timestamp" style="font-size: 11px; color: #94a3b8;">{{ step.timestamp }}</span>
          <span v-if="step.actor" style="font-size: 11px; color: #64748b;">by {{ step.actor }}</span>
        </div>
        <p v-if="step.note" style="font-size: 12px; color: #64748b; margin-top: 3px; margin-bottom: 0;">
          {{ step.note }}
        </p>
      </div>
    </div>
  </div>
</template>
