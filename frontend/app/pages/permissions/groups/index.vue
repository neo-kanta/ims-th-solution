<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IAM_USER_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.groupManagement.title"
    description-key="placeholders.groupManagement.description"
  />
</template>
