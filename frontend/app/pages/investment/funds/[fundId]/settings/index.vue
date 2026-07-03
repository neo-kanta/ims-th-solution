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
  <FundDetailLayout :fund-id="fundId" active-tab="settings">
    <template #default>
      <EmptyTabPanel
        :title="t('myFunds.detail.settings.title', 'Fund settings')"
        :body="
          t(
            'myFunds.detail.settings.body',
            'Fund metadata, manager assignment and notification settings will live here. Only users with INVESTMENT_FUND_MANAGE can edit.',
          )
        "
        :backend-gap="
          t(
            'myFunds.detail.settings.backendGap',
            'Edit flow uses the existing PUT /investment/funds/{id} — wiring to come.',
          )
        "
      />
    </template>
  </FundDetailLayout>
</template>
