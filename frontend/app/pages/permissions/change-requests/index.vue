<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.change_request.review",
});
</script>

<template>
  <PermissionRequestListScreen />
</template>
