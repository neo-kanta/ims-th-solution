<script setup lang="ts">
import { computed, onMounted } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import ResearchReportForm from "~/features/investment-research/components/ResearchReportForm.vue";
import {
  useResearchReportDetail,
  useResearchReportMutation,
} from "~/features/investment-research/composables/useResearchReports";
import type {
  CreateResearchReportInput,
  UpdateResearchReportInput,
} from "~/features/investment-research/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_RESEARCH_UPDATE",
});

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const reportId = computed(() => String(route.params.id ?? ""));
const { report, loading, error, fetch } = useResearchReportDetail();
const { saving, error: mutationError, update } = useResearchReportMutation();

async function handleSubmit(
  payload:
    | { mode: "create"; input: CreateResearchReportInput }
    | { mode: "edit"; input: UpdateResearchReportInput },
) {
  if (payload.mode !== "edit") return;
  try {
    await update(reportId.value, payload.input);
    void router.push(`/investment/analysis/${reportId.value}`);
  } catch {
    // surfaced via mutationError
  }
}

function cancel() {
  void router.push(`/investment/analysis/${reportId.value}`);
}

onMounted(() => {
  if (reportId.value) {
    void fetch(reportId.value);
  }
});
</script>

<template>
  <section class="analysis-edit-page">
    <AppPageHeader
      :title="t('investmentResearch.editReport')"
      :description="report ? report.report_no : ''"
    />

    <div v-if="loading" class="analysis-edit-page__state">
      {{ t('investmentResearch.loading') }}
    </div>
    <div
      v-else-if="error"
      class="analysis-edit-page__state analysis-edit-page__state--error"
    >
      {{ error }}
    </div>
    <ResearchReportForm
      v-else-if="report"
      mode="edit"
      :initial="report"
      :saving="saving"
      :error="mutationError"
      @submit="handleSubmit"
      @cancel="cancel"
    />
  </section>
</template>

<style scoped>
.analysis-edit-page {
  display: grid;
  gap: var(--space-5);
}

.analysis-edit-page__state {
  padding: var(--space-6);
  text-align: center;
  color: var(--text-secondary, #4b5563);
}

.analysis-edit-page__state--error {
  color: var(--state-error, #dc2626);
}
</style>
