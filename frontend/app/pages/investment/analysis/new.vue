<script setup lang="ts">
import { computed } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import ResearchReportForm from "~/features/investment-research/components/ResearchReportForm.vue";
import { useResearchReportMutation } from "~/features/investment-research/composables/useResearchReports";
import type { CreateResearchReportInput } from "~/features/investment-research/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_RESEARCH_CREATE",
});

const { t } = useI18n();
const router = useRouter();
const authStore = useAuthStore();
const { saving, error, create } = useResearchReportMutation();

/**
 * Pre-fill owner_user_id and author_user_id with the authenticated user's
 * UUID. The fields remain editable in the form so an admin can override.
 */
const defaultUserId = computed(() => authStore.user?.id ?? null);

async function handleSubmit(
  payload:
    | { mode: "create"; input: CreateResearchReportInput }
    | { mode: "edit"; input: unknown },
) {
  if (payload.mode !== "create") return;
  try {
    const report = await create(payload.input);
    void router.push(`/investment/analysis/${report.id}`);
  } catch {
    // error is exposed via the composable
  }
}

function cancel() {
  void router.push("/investment/analysis");
}
</script>

<template>
  <section class="analysis-new-page">
    <AppPageHeader
      :title="t('investmentResearch.newReport')"
      :description="t('investmentResearch.descriptionNew')"
    />
    <ResearchReportForm
      mode="create"
      :default-user-id="defaultUserId"
      :saving="saving"
      :error="error"
      @submit="handleSubmit"
      @cancel="cancel"
    />
  </section>
</template>

<style scoped>
.analysis-new-page {
  display: grid;
  gap: var(--space-5);
}
</style>
