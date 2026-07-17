<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useState } from "#imports";

import { useI18n } from "~/composables/useI18n";
import { useAuthStore } from "~/stores/useAuthStore";
import AppConfirmDialog from "~/shared/ui/AppConfirmDialog.vue";

import PortfolioAvailableRulesList from "~/features/compliance/components/PortfolioAvailableRulesList.vue";
import PortfolioBindRuleDrawer from "~/features/compliance/components/PortfolioBindRuleDrawer.vue";
import PortfolioBoundRulesTable from "~/features/compliance/components/PortfolioBoundRulesTable.vue";
import PortfolioBreachPanel from "~/features/compliance/components/PortfolioBreachPanel.vue";
import PortfolioCompliancePosture from "~/features/compliance/components/PortfolioCompliancePosture.vue";
import { deviceLocalIsoDate } from "~/features/compliance/lib/asOfDate";
import { derivePortfolioPosture } from "~/features/compliance/lib/portfolioPosture";
import { ruleLabel } from "~/features/compliance/lib/ruleTypeCatalog";

import PortfolioWorkspaceHeader from "./components/PortfolioWorkspaceHeader.vue";
import { usePortfolioContext } from "./composables/usePortfolioContext";
import { buildBindRulePayload, isSubmissionInFlight, type BindDraft } from "./lib/complianceBindingState";
import {
  portfolioComplianceApi,
  type ApiPortfolioBreachView,
  type ApiPortfolioRuleCatalogEntry,
} from "./services/portfolioComplianceApi";

const props = defineProps<{ portfolioCode: string }>();
const { t } = useI18n();
const authStore = useAuthStore();
const ctx = usePortfolioContext(() => props.portfolioCode);

const pageTitle = useState<string>("page-title", () => "");
watch(
  () => t("portfolio.workspaceTabs.compliance", "Compliance"),
  (newTitle) => {
    pageTitle.value = newTitle || "";
  },
  { immediate: true }
);

const canManage = computed(() => authStore.hasPermission("IRG_EDIT_BINDING"));
const todayIso = deviceLocalIsoDate();

const rules = ref<ApiPortfolioRuleCatalogEntry[]>([]);
const loadingRules = ref(false);
const rulesError = ref<string | null>(null);

const breaches = ref<ApiPortfolioBreachView[]>([]);
const loadingBreaches = ref(false);
const breachesError = ref<string | null>(null);

function extractMessage(err: unknown, fallback: string): string {
  if (err instanceof Error && err.message) return err.message;
  return fallback;
}

async function loadRules() {
  if (!props.portfolioCode) return;
  loadingRules.value = true;
  rulesError.value = null;
  try {
    rules.value = await portfolioComplianceApi.listRules(props.portfolioCode);
  } catch (err) {
    rulesError.value = extractMessage(err, t("portfolio.compliance.errors.rules"));
  } finally {
    loadingRules.value = false;
  }
}

async function loadBreaches() {
  if (!props.portfolioCode) return;
  loadingBreaches.value = true;
  breachesError.value = null;
  try {
    breaches.value = await portfolioComplianceApi.listBreaches(props.portfolioCode);
  } catch (err) {
    breachesError.value = extractMessage(err, t("portfolio.compliance.errors.breaches"));
  } finally {
    loadingBreaches.value = false;
  }
}

onMounted(() => {
  void ctx.reload();
  void loadRules();
  void loadBreaches();
});

watch(
  () => props.portfolioCode,
  () => {
    void ctx.reload();
    void loadRules();
    void loadBreaches();
  }
);

const boundRules = computed(() => rules.value.filter((r) => r.binding?.is_active));
const unboundRules = computed(() => rules.value.filter((r) => !r.binding?.is_active));

const posture = computed(() => derivePortfolioPosture(rules.value, breaches.value, todayIso));
const postureLoading = computed(() => loadingRules.value || loadingBreaches.value);

// --- Bind drawer (one focused panel, not many inline forms) ---
const bindingBusy = ref<Record<string, boolean>>({});
const selectedBindEntry = ref<ApiPortfolioRuleCatalogEntry | null>(null);
const bindDraft = ref<BindDraft>({ severity: "BLOCK", effectiveFrom: todayIso, effectiveTo: "" });
const bindSubmitting = ref(false);
const bindError = ref<string | null>(null);

function openBindDrawer(entry: ApiPortfolioRuleCatalogEntry) {
  selectedBindEntry.value = entry;
  bindDraft.value = { severity: "BLOCK", effectiveFrom: todayIso, effectiveTo: "" };
  bindError.value = null;
}

function closeBindDrawer() {
  if (bindSubmitting.value) return;
  selectedBindEntry.value = null;
}

function updateBindDraft(next: BindDraft) {
  bindDraft.value = next;
}

async function submitBind() {
  const entry = selectedBindEntry.value;
  const id = entry?.rule_instance_id ?? "";
  if (!entry || !id) return;
  if (isSubmissionInFlight(bindingBusy.value, id)) return;

  bindSubmitting.value = true;
  bindingBusy.value = { ...bindingBusy.value, [id]: true };
  bindError.value = null;
  try {
    await portfolioComplianceApi.bindRule(
      props.portfolioCode,
      id,
      buildBindRulePayload(bindDraft.value),
    );
    selectedBindEntry.value = null;
    await loadRules();
  } catch (err) {
    bindError.value = extractMessage(err, t("portfolio.compliance.errors.bind"));
  } finally {
    bindSubmitting.value = false;
    bindingBusy.value = { ...bindingBusy.value, [id]: false };
  }
}

// --- Deactivate confirmation ---
const deactivateTarget = ref<ApiPortfolioRuleCatalogEntry | null>(null);
const deactivateBusy = ref(false);
const deactivateError = ref<string | null>(null);

function requestDeactivate(entry: ApiPortfolioRuleCatalogEntry) {
  const id = entry.rule_instance_id ?? "";
  if (!id || isSubmissionInFlight(bindingBusy.value, id)) return;
  deactivateTarget.value = entry;
  deactivateError.value = null;
}

function cancelDeactivate() {
  if (deactivateBusy.value) return;
  deactivateTarget.value = null;
}

async function confirmDeactivate() {
  const entry = deactivateTarget.value;
  const id = entry?.rule_instance_id ?? "";
  const bindingId = entry?.binding?.binding_id;
  if (!entry || !id || !bindingId) return;

  deactivateBusy.value = true;
  bindingBusy.value = { ...bindingBusy.value, [id]: true };
  deactivateError.value = null;
  try {
    await portfolioComplianceApi.deactivateBinding(props.portfolioCode, id, bindingId);
    deactivateTarget.value = null;
    await loadRules();
  } catch (err) {
    deactivateError.value = extractMessage(err, t("portfolio.compliance.errors.deactivate"));
  } finally {
    deactivateBusy.value = false;
    bindingBusy.value = { ...bindingBusy.value, [id]: false };
  }
}

const deactivateDescription = computed(() => {
  const entry = deactivateTarget.value;
  if (!entry) return "";
  const label = ruleLabel(entry.rule_type_id ?? "", t);
  const base = t("portfolio.compliance.deactivateConfirm.description", { rule: label });
  return deactivateError.value ? `${base}\n\n${deactivateError.value}` : base;
});
</script>

<template>
  <section class="portfolio-compliance">
    <div v-if="ctx.loading.value && !ctx.portfolio.value" class="portfolio-compliance__notice" role="status">
      {{ t("portfolio.overview.loading") }}
    </div>

    <template v-else-if="ctx.portfolio.value">
      <PortfolioWorkspaceHeader :portfolio="ctx.portfolio.value" />

      <p class="portfolio-compliance__as-of">
        {{ t("portfolio.compliance.asOfLabel", { date: todayIso }) }}
      </p>

      <PortfolioCompliancePosture
        :open-breach-count="postureLoading ? null : posture.openBreachCount"
        :effective-count="postureLoading ? null : posture.effectiveCount"
        :scheduled-count="postureLoading ? null : posture.scheduledCount"
        :strongest-effective-severity="posture.strongestEffectiveSeverity"
        :loading="postureLoading"
      />

      <div v-if="!canManage" class="portfolio-compliance__readonly-notice" role="status">
        {{ t("portfolio.compliance.readOnlyNotice") }}
      </div>

      <PortfolioBreachPanel
        :breaches="breaches"
        :loading="loadingBreaches"
        :error="breachesError"
        @retry="loadBreaches"
      />

      <PortfolioBoundRulesTable
        :entries="boundRules"
        :today-iso="todayIso"
        :busy="bindingBusy"
        :can-manage="canManage"
        :loading="loadingRules"
        :error="rulesError"
        @deactivate="requestDeactivate"
      />

      <PortfolioAvailableRulesList
        :entries="unboundRules"
        :loading="loadingRules"
        :can-manage="canManage"
        @bind="openBindDrawer"
      />

      <PortfolioBindRuleDrawer
        :open="selectedBindEntry !== null"
        :entry="selectedBindEntry"
        :draft="bindDraft"
        :submitting="bindSubmitting"
        :error="bindError"
        @close="closeBindDrawer"
        @update:draft="updateBindDraft"
        @submit="submitBind"
      />

      <AppConfirmDialog
        :open="deactivateTarget !== null"
        tone="danger"
        :title="t('portfolio.compliance.deactivateConfirm.title')"
        :description="deactivateDescription"
        :confirm-label="t('portfolio.compliance.deactivate')"
        :cancel-label="t('portfolio.compliance.cancel')"
        :loading="deactivateBusy"
        @cancel="cancelDeactivate"
        @confirm="confirmDeactivate"
      />
    </template>

    <div v-else class="portfolio-compliance__notice">
      {{ t("portfolio.overview.notFound") }}
    </div>
  </section>
</template>

<style scoped>
.portfolio-compliance {
  display: grid;
  gap: var(--space-5, 20px);
}

.portfolio-compliance__notice {
  padding: var(--space-4, 16px);
  background: var(--bg-card-muted);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  text-align: center;
}

.portfolio-compliance__as-of {
  margin: 0;
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
}

.portfolio-compliance__readonly-notice {
  padding: var(--space-3) var(--space-4);
  background: var(--bg-card-muted);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}
</style>
