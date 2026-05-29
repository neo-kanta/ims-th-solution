<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import ResearchReportDetail from "~/features/investment-research/components/ResearchReportDetail.vue";
import {
  useResearchReportDetail,
  useResearchReportMutation,
} from "~/features/investment-research/composables/useResearchReports";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "INVESTMENT_RESEARCH_VIEW",
});

const route = useRoute();
const router = useRouter();
const { t } = useI18n();

const reportId = computed(() => String(route.params.id ?? ""));
const { report, loading, error, fetch } = useResearchReportDetail();
const { saving, error: mutationError, remove, submit, cancelSubmit } =
  useResearchReportMutation();

/**
 * Three confirm-dialog slots, all driven by AppConfirmDialog. Each
 * lifecycle action is gated by an explicit confirmation step — Delete is
 * destructive (soft-delete, but operator-visible as "gone"), Submit moves
 * the report into the review queue, and Cancel-submit pulls it back. For
 * an enterprise financial-workflow UI all three deserve the same friction.
 */
type ConfirmKind = "delete" | "submit" | "cancelSubmit";
const confirmKind = ref<ConfirmKind | null>(null);

const confirmConfig = computed(() => {
  switch (confirmKind.value) {
    case "delete":
      return {
        title: t("investmentResearch.confirm.delete.title"),
        description: t("investmentResearch.confirm.delete.description"),
        confirmLabel: t("investmentResearch.confirm.delete.confirmLabel"),
        tone: "danger" as const,
      };
    case "submit":
      return {
        title: t("investmentResearch.confirm.submit.title"),
        description: t("investmentResearch.confirm.submit.description"),
        confirmLabel: t("investmentResearch.confirm.submit.confirmLabel"),
        tone: "warning" as const,
      };
    case "cancelSubmit":
      return {
        title: t("investmentResearch.confirm.cancelSubmit.title"),
        description: t("investmentResearch.confirm.cancelSubmit.description"),
        confirmLabel: t("investmentResearch.confirm.cancelSubmit.confirmLabel"),
        tone: "warning" as const,
      };
    default:
      return null;
  }
});

async function refresh() {
  if (!reportId.value) return;
  await fetch(reportId.value);
}

function goEdit() {
  void router.push(`/investment/analysis/${reportId.value}/edit`);
}

function goBack() {
  void router.push("/investment/analysis");
}

function openConfirm(kind: ConfirmKind) {
  confirmKind.value = kind;
}

function dismissConfirm() {
  // Don't allow dismiss while the mutation is in-flight — AppConfirmDialog
  // already ignores ESC/backdrop in that case, but guard the button too.
  if (saving.value) return;
  confirmKind.value = null;
}

async function executeConfirm() {
  const kind = confirmKind.value;
  if (!kind) return;
  try {
    if (kind === "delete") {
      await remove(reportId.value);
      confirmKind.value = null;
      void router.push("/investment/analysis");
      return;
    }
    if (kind === "submit") {
      await submit(reportId.value);
    } else if (kind === "cancelSubmit") {
      await cancelSubmit(reportId.value);
    }
    confirmKind.value = null;
    await refresh();
  } catch {
    // Error is surfaced via mutationError; keep the dialog open so the
    // operator can retry or cancel.
  }
}

onMounted(() => {
  void refresh();
});
</script>

<template>
  <section class="analysis-detail-page">
    <AppPageHeader :title="t('investmentResearch.researchReport')" />

    <div v-if="loading" class="analysis-detail-page__state">
      {{ t('investmentResearch.loading') }}
    </div>
    <div
      v-else-if="error"
      class="analysis-detail-page__state analysis-detail-page__state--error"
    >
      {{ error }}
    </div>
    <ResearchReportDetail
      v-else-if="report"
      :report="report"
      :saving="saving"
      :error="mutationError"
      @edit="goEdit"
      @delete="() => openConfirm('delete')"
      @submit="() => openConfirm('submit')"
      @cancel-submit="() => openConfirm('cancelSubmit')"
      @back="goBack"
    />

    <AppConfirmDialog
      v-if="confirmConfig"
      :open="confirmKind !== null"
      :title="confirmConfig.title"
      :description="confirmConfig.description"
      :confirm-label="confirmConfig.confirmLabel"
      :tone="confirmConfig.tone"
      :loading="saving"
      @cancel="dismissConfirm"
      @confirm="executeConfirm"
    />
  </section>
</template>

<style scoped>
.analysis-detail-page {
  display: grid;
  gap: var(--space-5);
}

.analysis-detail-page__state {
  padding: var(--space-6);
  text-align: center;
  color: var(--text-secondary, #4b5563);
}

.analysis-detail-page__state--error {
  color: var(--state-error, #dc2626);
}
</style>
