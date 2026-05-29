<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.users.view",
});
</script>

<template>
  <EffectivePermissionViewerScreen />
</template>
