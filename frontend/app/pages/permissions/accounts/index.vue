<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IAM_USER_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.accountManagement.title"
    description-key="placeholders.accountManagement.description"
  />
</template>
