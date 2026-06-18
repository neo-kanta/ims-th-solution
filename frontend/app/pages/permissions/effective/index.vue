<script setup lang="ts">
import EffectivePermissionViewerScreen from "~/features/permissions/components/EffectivePermissionViewerScreen.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.users.view",
});
</script>

<template>
  <EffectivePermissionViewerScreen />
</template>
