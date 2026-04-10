<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "NOTIFICATION_CONFIG",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.notificationSettings.title"
    description-key="placeholders.notificationSettings.description"
  />
</template>
