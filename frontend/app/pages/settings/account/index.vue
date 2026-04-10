<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "AUDIT_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.auditLog.title"
    description-key="placeholders.auditLog.description"
  />
</template>
