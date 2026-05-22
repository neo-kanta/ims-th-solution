<script setup lang="ts">
import { computed } from "vue";

interface Props {
  name: string;
  size?: "xs" | "sm" | "md" | "lg";
  decorative?: boolean;
  title?: string;
}

const props = withDefaults(defineProps<Props>(), {
  size: "md",
  decorative: true,
  title: undefined,
});

const titleId = computed(() => (
  props.title ? `app-icon-${props.name}-${props.size}` : undefined
));
</script>

<template>
  <svg
    class="app-icon"
    :class="`app-icon--${size}`"
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.9"
    stroke-linecap="round"
    stroke-linejoin="round"
    :aria-hidden="decorative ? 'true' : undefined"
    :aria-labelledby="!decorative && titleId ? titleId : undefined"
    :role="decorative ? undefined : 'img'"
  >
    <title v-if="!decorative && title && titleId" :id="titleId">{{ title }}</title>

    <template v-if="name === 'dashboard'">
      <rect x="3.5" y="3.5" width="7" height="7" rx="1.5" />
      <rect x="13.5" y="3.5" width="7" height="10" rx="1.5" />
      <rect x="3.5" y="13.5" width="7" height="7" rx="1.5" />
      <rect x="13.5" y="16.5" width="7" height="4" rx="1.5" />
    </template>

    <template v-else-if="name === 'workflow'">
      <path d="M7 7h10" />
      <path d="M13 4l4 3-4 3" />
      <path d="M17 17H7" />
      <path d="M11 20l-4-3 4-3" />
    </template>

    <template v-else-if="name === 'approval'">
      <path d="M9 3.75h6l4 4v12.5a1.75 1.75 0 0 1-1.75 1.75H6.75A1.75 1.75 0 0 1 5 20.25V5.5A1.75 1.75 0 0 1 6.75 3.75Z" />
      <path d="M9 11.5l2 2 4-4" />
    </template>

    <template v-else-if="name === 'leave'">
      <rect x="3.5" y="4.5" width="17" height="16" rx="2.5" />
      <path d="M8 2.75v3.5" />
      <path d="M16 2.75v3.5" />
      <path d="M3.5 9.5h17" />
      <path d="M8 13h3" />
      <path d="M13 16h3" />
    </template>

    <template v-else-if="name === 'analysis'">
      <path d="M4 19.5h16" />
      <path d="M7 15.5 10.5 12l3 2 4.5-6" />
      <path d="m17.5 8 1.5 0-0 1.5" />
    </template>

    <template v-else-if="name === 'decision'">
      <path d="M7 4.5h10a2 2 0 0 1 2 2v11a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2v-11a2 2 0 0 1 2-2Z" />
      <path d="M8.5 9h7" />
      <path d="M8.5 12h7" />
      <path d="M8.5 15h4.5" />
    </template>

    <template v-else-if="name === 'execution'">
      <path d="M5 12h14" />
      <path d="m13 6 6 6-6 6" />
    </template>

    <template v-else-if="name === 'review'">
      <path d="M12 3.5 5.5 6v5.5c0 4.2 2.4 7.3 6.5 9 4.1-1.7 6.5-4.8 6.5-9V6Z" />
      <path d="m9.5 12.5 1.75 1.75 3.25-3.75" />
    </template>

    <template v-else-if="name === 'accounts'">
      <path d="M12 12.5a3.75 3.75 0 1 0 0-7.5 3.75 3.75 0 0 0 0 7.5Z" />
      <path d="M5 19a7 7 0 0 1 14 0" />
    </template>

    <template v-else-if="name === 'groups'">
      <path d="M8 12a2.75 2.75 0 1 0 0-5.5A2.75 2.75 0 0 0 8 12Z" />
      <path d="M16.25 11.5a2.25 2.25 0 1 0 0-4.5 2.25 2.25 0 0 0 0 4.5Z" />
      <path d="M3.75 18.25a5 5 0 0 1 8.5-2.75" />
      <path d="M13 18.5a4.5 4.5 0 0 1 7.25-2.25" />
    </template>

    <template v-else-if="name === 'notifications'">
      <path d="M6.5 9a5.5 5.5 0 0 1 11 0v4.25l1.5 2.25H5l1.5-2.25Z" />
      <path d="M10 18a2 2 0 0 0 4 0" />
    </template>

    <template v-else-if="name === 'audit'">
      <path d="M8 4.5h8l3 3v11a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2v-12a2 2 0 0 1 2-2Z" />
      <path d="M9 10h6" />
      <path d="M9 13h6" />
      <path d="M9 16h4" />
    </template>

    <template v-else-if="name === 'menu'">
      <path d="M4 7h16" />
      <path d="M4 12h16" />
      <path d="M4 17h16" />
    </template>

    <template v-else-if="name === 'panel-open'">
      <rect x="3.5" y="4" width="17" height="16" rx="2" />
      <path d="M9 4v16" />
      <path d="m15 12-3-3" />
      <path d="m15 12-3 3" />
    </template>

    <template v-else-if="name === 'panel-close'">
      <rect x="3.5" y="4" width="17" height="16" rx="2" />
      <path d="M9 4v16" />
      <path d="m12 12 3-3" />
      <path d="m12 12 3 3" />
    </template>

    <template v-else-if="name === 'close'">
      <path d="m6 6 12 12" />
      <path d="M18 6 6 18" />
    </template>

    <template v-else-if="name === 'chevron-down'">
      <path d="m6 9 6 6 6-6" />
    </template>

    <template v-else-if="name === 'chevron-right'">
      <path d="m9 6 6 6-6 6" />
    </template>

    <template v-else-if="name === 'globe'">
      <circle cx="12" cy="12" r="8.5" />
      <path d="M3.5 12h17" />
      <path d="M12 3.5a12 12 0 0 1 0 17" />
      <path d="M12 3.5a12 12 0 0 0 0 17" />
    </template>

    <template v-else-if="name === 'sun'">
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2.75v2.5" />
      <path d="M12 18.75v2.5" />
      <path d="M4.75 12h2.5" />
      <path d="M16.75 12h2.5" />
      <path d="m5.85 5.85 1.8 1.8" />
      <path d="m16.35 16.35 1.8 1.8" />
      <path d="m18.15 5.85-1.8 1.8" />
      <path d="m7.65 16.35-1.8 1.8" />
    </template>

    <template v-else-if="name === 'moon'">
      <path d="M18.25 14.75A7.25 7.25 0 0 1 9.25 5.75 7.5 7.5 0 1 0 18.25 14.75Z" />
    </template>

    <template v-else-if="name === 'user'">
      <circle cx="12" cy="9" r="3.5" />
      <path d="M5 19a7 7 0 0 1 14 0" />
    </template>

    <template v-else-if="name === 'refresh'">
      <path d="M20 5.5v5.25h-5.25" />
      <path d="M4 18.5v-5.25h5.25" />
      <path d="M6.5 8.5a7.5 7.5 0 0 1 12.25-2.3" />
      <path d="M17.5 15.5A7.5 7.5 0 0 1 5.25 18" />
    </template>

    <template v-else-if="name === 'trend-up'">
      <path d="M5 16.5 10 11.5l3 3 6-7" />
      <path d="M14.5 7.5H19V12" />
    </template>

    <template v-else-if="name === 'trend-down'">
      <path d="M5 7.5 10 12.5l3-3 6 7" />
      <path d="M14.5 16.5H19V12" />
    </template>

    <template v-else-if="name === 'trend-flat'">
      <path d="M5 12h14" />
    </template>

    <template v-else-if="name === 'check'">
      <path d="m5 12 4 4 10-10" />
    </template>

    <template v-else-if="name === 'pending'">
      <circle cx="12" cy="12" r="8.5" />
      <path d="M12 8v4.5l3 1.5" />
    </template>

    <template v-else-if="name === 'lock'">
      <rect x="5" y="10.5" width="14" height="10" rx="2" />
      <path d="M8 10.5V8a4 4 0 0 1 8 0v2.5" />
    </template>

    <template v-else-if="name === 'warning'">
      <path d="M12 4.5 20 19H4l8-14.5Z" />
      <path d="M12 9v4" />
      <circle cx="12" cy="16.5" r=".8" fill="currentColor" stroke="none" />
    </template>

    <template v-else-if="name === 'info'">
      <circle cx="12" cy="12" r="8.5" />
      <path d="M12 10.5v5" />
      <circle cx="12" cy="7.5" r=".8" fill="currentColor" stroke="none" />
    </template>

    <template v-else-if="name === 'activity'">
      <path d="M4 13h4l2-5 4 10 2-5h4" />
    </template>

    <template v-else-if="name === 'portfolio'">
      <path d="M12 4.5a7.5 7.5 0 1 1-7.5 7.5" />
      <path d="M12 4.5v7.5l5.5 5.5" />
    </template>

    <template v-else-if="name === 'list'">
      <path d="M8.5 7h10" />
      <path d="M8.5 12h10" />
      <path d="M8.5 17h10" />
      <circle cx="5" cy="7" r="1" fill="currentColor" stroke="none" />
      <circle cx="5" cy="12" r="1" fill="currentColor" stroke="none" />
      <circle cx="5" cy="17" r="1" fill="currentColor" stroke="none" />
    </template>

    <template v-else-if="name === 'plus'">
      <path d="M12 5v14" />
      <path d="M5 12h14" />
    </template>

    <template v-else-if="name === 'thumbs-up'">
      <path d="M9.5 10V5.75A2.75 2.75 0 0 1 12.25 3L13 3.75V10" />
      <path d="M7.5 10H5.75A1.75 1.75 0 0 0 4 11.75v5.5A1.75 1.75 0 0 0 5.75 19h9.5a2 2 0 0 0 1.93-1.48l1.55-5.5A1.75 1.75 0 0 0 17.05 10H13" />
    </template>

    <template v-else-if="name === 'clock'">
      <circle cx="12" cy="12" r="8.5" />
      <path d="M12 8v4.5l3 2" />
    </template>

    <template v-else-if="name === 'system'">
      <path d="M12 6.5v11" />
      <path d="M7.25 9.25 16.75 14.75" />
      <path d="m7.25 14.75 9.5-5.5" />
      <circle cx="12" cy="12" r="8.5" />
    </template>

    <template v-else-if="name === 'shield'">
      <path d="M12 3.5 5.5 6v5.5c0 4.2 2.4 7.3 6.5 9 4.1-1.7 6.5-4.8 6.5-9V6Z" />
      <path d="M9.5 9.5h5" />
      <path d="M9.5 12.5h5" />
    </template>

    <template v-else-if="name === 'pull-request'">
      <circle cx="18" cy="18" r="3" />
      <circle cx="6" cy="6" r="3" />
      <circle cx="6" cy="18" r="3" />
      <path d="M18 15V9a4 4 0 0 0-4-4H9" />
      <path d="M6 9v6" />
    </template>

    <template v-else-if="name === 'issue'">
      <circle cx="12" cy="12" r="8.5" />
      <circle cx="12" cy="12" r="1.5" fill="currentColor" stroke="none" />
    </template>

    <template v-else-if="name === 'compliance'">
      <path d="M12 3.5 5.5 6v5.5c0 4.2 2.4 7.3 6.5 9 4.1-1.7 6.5-4.8 6.5-9V6Z" />
      <path d="m9 11.5 2 2 4-4" />
    </template>

    <template v-else-if="name === 'briefcase'">
      <rect x="3.5" y="7.5" width="17" height="11.5" rx="2" />
      <path d="M8.5 7.5v-1A2.5 2.5 0 0 1 11 4h2a2.5 2.5 0 0 1 2.5 2.5v1" />
      <path d="M3.5 12h17" />
    </template>

    <template v-else>
      <circle cx="12" cy="12" r="8.5" />
      <path d="M12 8v4" />
      <circle cx="12" cy="16.5" r=".9" fill="currentColor" stroke="none" />
    </template>
  </svg>
</template>
