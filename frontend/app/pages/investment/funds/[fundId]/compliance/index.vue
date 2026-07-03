<script setup lang="ts">
import { computed } from "vue";

import { CompliancePanel, FundDetailLayout } from "~/features/my-funds";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const route = useRoute();
const fundId = computed(() => String(route.params.fundId ?? ""));
</script>

<template>
  <FundDetailLayout :fund-id="fundId" active-tab="compliance">
    <template #default="{ card }">
      <CompliancePanel :fund-id="card.fund_id" />
    </template>
  </FundDetailLayout>
</template>
