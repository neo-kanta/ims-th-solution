<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "APPROVAL_CONFIG",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.approvalConfiguration.title"
    description-key="placeholders.approvalConfiguration.description"
  />
</template>
