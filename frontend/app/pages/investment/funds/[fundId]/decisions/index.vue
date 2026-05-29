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
  <FundDetailLayout :fund-id="fundId" active-tab="decisions">
    <template #default="{ card }">
      <EmptyTabPanel
        :title="t('myFunds.detail.decisions.title', 'Investment decisions')"
        :body="
          t(
            'myFunds.detail.decisions.body',
            'Investment Decisions live in this tab once the backend persistence is wired. The pre-trade compliance gate already runs through the existing /compliance/checks/pre-trade endpoint.',
          )
        "
        :backend-gap="
          t(
            'myFunds.detail.decisions.backendGap',
            'Decision persistence + list endpoints are pending in backend/internal/investment.',
          )
        "
        :cta-label="t('myFunds.detail.decisions.cta', 'Open compliance pre-trade')"
        :cta-path="`/compliance/pre-trade?contract_id=${card.fund_id}`"
      />
    </template>
  </FundDetailLayout>
</template>
