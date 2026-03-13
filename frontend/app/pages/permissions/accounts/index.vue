<script setup lang="ts">
definePageMeta({
  layout: 'dashboard',
  middleware: ['auth', 'permission'],
  meta: { permission: 'PERMISSIONS_VIEW' },
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Account Management</h1>
    <div class="card">
      <p class="text-gray-500">Account Management — coming soon.</p>
    </div>
  </div>
</template>
