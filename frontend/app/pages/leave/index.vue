<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "LEAVE_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.leaveDelegation.title"
    description-key="placeholders.leaveDelegation.description"
  />
</template>
