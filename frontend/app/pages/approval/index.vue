<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "APPROVAL_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.approvalWorkflow.title"
    description-key="placeholders.approvalWorkflow.description"
  />
</template>
