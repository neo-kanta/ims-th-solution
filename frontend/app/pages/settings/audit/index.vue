<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "AUDIT_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.accountLog.title"
    description-key="placeholders.accountLog.description"
  />
</template>
