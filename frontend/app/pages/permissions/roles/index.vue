<script setup lang="ts">
import RoleHierarchyScreen from "~/features/permissions/components/RoleHierarchyScreen.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.roles.view",
});
</script>

<template>
  <RoleHierarchyScreen />
</template>
