<script setup lang="ts">
import DecisionApprovalWorkbench from "~/features/investment-decision/components/DecisionApprovalWorkbench.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_DECISION_APPROVE",
});
</script>

<template>
  <DecisionApprovalWorkbench />
</template>
