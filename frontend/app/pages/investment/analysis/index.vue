<script setup lang="ts">
definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_VIEW",
});
</script>

<template>
  <AppRoutePlaceholder
    title-key="placeholders.analysisReports.title"
    description-key="placeholders.analysisReports.description"
  />
</template>
