<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.roles.view",
});
</script>

<template>
  <RoleHierarchyScreen />
</template>
