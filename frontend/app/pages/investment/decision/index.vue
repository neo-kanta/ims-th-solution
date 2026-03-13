<script setup lang="ts">
definePageMeta({
  layout: 'dashboard',
  middleware: ['auth', 'permission'],
  meta: { permission: 'INVESTMENT_VIEW' },
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Investment Decisions</h1>
    <div class="card">
      <p class="text-gray-500">Investment Decisions — coming soon.</p>
    </div>
  </div>
</template>
