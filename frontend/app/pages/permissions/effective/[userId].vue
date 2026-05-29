<script setup lang="ts">
const route = useRoute();

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.users.view",
});
</script>

<template>
  <EffectivePermissionViewerScreen :initial-user-id="String(route.params.userId)" />
</template>
