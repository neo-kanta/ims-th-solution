<script setup lang="ts">
definePageMeta({
  layout: 'dashboard',
  middleware: ['auth', 'permission'],
  meta: { permission: 'APPROVAL_CONFIG' },
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Approval Configuration</h1>
    <div class="card">
      <p class="text-gray-500">Approval Configuration — coming soon.</p>
    </div>
  </div>
</template>
