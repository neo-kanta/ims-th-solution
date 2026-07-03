<script setup lang="ts">
import { computed } from "vue";

import { EmptyTabPanel, FundDetailLayout } from "~/features/my-funds";
import { useI18n } from "~/composables/useI18n";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_FUND_VIEW",
});

const route = useRoute();
const fundId = computed(() => String(route.params.fundId ?? ""));
const { t } = useI18n();
</script>

<template>
  <FundDetailLayout :fund-id="fundId" active-tab="reviewers">
    <template #default>
      <EmptyTabPanel
        :title="t('myFunds.detail.reviewers.title', 'Reviewers & approvers')"
        :body="
          t(
            'myFunds.detail.reviewers.body',
            'Approval flows for this fund will list here once the approval module exposes its routes. For now, workflow stage approvals are visible on the Workflow tab.',
          )
        "
        :backend-gap="
          t(
            'myFunds.detail.reviewers.backendGap',
            'backend/internal/approval is scaffold-only.',
          )
        "
      />
    </template>
  </FundDetailLayout>
</template>
