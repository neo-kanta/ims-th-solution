<script setup lang="ts">
definePageMeta({
  layout: 'dashboard',
  middleware: ['auth', 'permission'],
  meta: { permission: 'NOTIFICATION_CONFIG' },
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Notification Settings</h1>
    <div class="card">
      <p class="text-gray-500">Notification Settings — coming soon.</p>
    </div>
  </div>
</template>
