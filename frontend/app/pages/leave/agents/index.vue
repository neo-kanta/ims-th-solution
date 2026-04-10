<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "LEAVE_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.agentManagement.title"
    description-key="placeholders.agentManagement.description"
  />
</template>
