<script setup lang="ts">
// Auth layout — used for login/register pages
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
    <div class="w-full max-w-md">
      <div class="mb-8 text-center">
        <h1 class="text-3xl font-bold text-gray-900">IMS Thailand</h1>
        <p class="mt-2 text-sm text-gray-600">Investment Management System</p>
      </div>
      <div class="card">
        <slot />
      </div>
    </div>
  </div>
</template>
