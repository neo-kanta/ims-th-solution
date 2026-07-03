<script setup lang="ts">
import PermissionRequestListScreen from "~/features/permissions/components/PermissionRequestListScreen.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.change_request.review",
});
</script>

<template>
  <PermissionRequestListScreen />
</template>
