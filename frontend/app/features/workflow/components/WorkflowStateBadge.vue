<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "~/composables/useI18n";
import AppBadge from "~/shared/ui/AppBadge.vue";

const props = defineProps<{
  state?: string;
}>();

const { t } = useI18n();

const stateConfig = computed(() => {
  switch (props.state) {
    case "INVESTMENT_DAY_STARTED":
      return {
        variant: "info" as const,
        label: t("workflow.state.INVESTMENT_DAY_STARTED", "Investment Day Started"),
      };
    case "MANAGER_APPROVED":
      return {
        variant: "success" as const,
        label: t("workflow.state.MANAGER_APPROVED", "Manager Approved"),
      };
    case "TRANSACTION_CLOSED":
      return {
        variant: "success" as const,
        label: t("workflow.state.TRANSACTION_CLOSED", "Transaction Closed"),
      };
    case "ACCOUNTING_CLOSED":
      return {
        variant: "neutral" as const,
        label: t("workflow.state.ACCOUNTING_CLOSED", "Accounting Closed"),
      };
    case "NOT_STARTED":
    default:
      return {
        variant: "neutral" as const,
        label: t("workflow.state.NOT_STARTED", "Not Started"),
      };
  }
});
</script>

<template>
  <AppBadge :variant="stateConfig.variant" size="md" dot>
    {{ stateConfig.label }}
  </AppBadge>
</template>
