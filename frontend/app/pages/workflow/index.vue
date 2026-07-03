<script setup lang="ts">
import WorkflowPage from "~/features/workflow/components/WorkflowPage.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "WORKFLOW_VIEW",
});
</script>

<template>
  <WorkflowPage />
</template>

