<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";

import ComplianceRuleBuilder from "~/features/compliance/components/ComplianceRuleBuilder.vue";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "IRG_EDIT_RULE_INSTANCE",
});

const { t } = useI18n();
const router = useRouter();

function onCreated(id: string) {
  void router.push(`/compliance/rules/${id}`);
}
</script>

<template>
  <section class="new-rule-page">
    <AppPageHeader
      :title="t('compliance.builder.title')"
      :description="t('compliance.builder.description')"
    />
    <ComplianceRuleBuilder @created="onCreated" />
  </section>
</template>

<style scoped>
.new-rule-page {
  display: grid;
  gap: var(--space-7);
}
</style>
