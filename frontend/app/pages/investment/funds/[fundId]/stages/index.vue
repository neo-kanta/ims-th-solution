<script setup lang="ts">
import { computed } from "vue";

import { FundDetailLayout, StagesPanel } from "~/features/my-funds";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const route = useRoute();
const fundId = computed(() => String(route.params.fundId ?? ""));
</script>

<template>
  <FundDetailLayout :fund-id="fundId" active-tab="stages">
    <template #default="{ card }">
      <StagesPanel :fund-id="card.fund_id" />
    </template>
  </FundDetailLayout>
</template>
