<script setup lang="ts">
import { onMounted, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

import ApprovalProcessConfigForm from "~/features/approval/components/ApprovalProcessConfigForm.vue";
import { useApprovalProcesses, useApprovalGroups } from "~/features/approval/composables/useApprovalConfig";
import { approvalApi, approvalErrorMessage } from "~/features/approval/services/approvalApi";
import { prettify } from "~/features/approval/lib/approvalStatus";
import { useComplianceUserDirectory } from "~/features/compliance/composables/useComplianceUserDirectory";
import type { ProcessConfigInput } from "~/features/approval/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "APPROVAL_CONFIG_VIEW",
});

const { t } = useI18n();
const { processes, loading, error, forbidden, fetchProcesses } = useApprovalProcesses();
const { groups, fetchGroups } = useApprovalGroups();
const userDirectory = useComplianceUserDirectory();

const showForm = ref(false);
const submitting = ref(false);
const formError = ref<string | null>(null);
const busy = ref(false);

async function reload() {
  await fetchProcesses();
}

function startCreate() {
  showForm.value = true;
  formError.value = null;
}

async function saveProcess(payload: ProcessConfigInput) {
  submitting.value = true;
  formError.value = null;
  try {
    await approvalApi.createProcess(payload);
    showForm.value = false;
    await reload();
  } catch (err) {
    formError.value = approvalErrorMessage(err, "Failed to create approval process.");
  } finally {
    submitting.value = false;
  }
}

async function toggleActive(id: string, active: boolean) {
  busy.value = true;
  try {
    if (active) await approvalApi.deactivateProcess(id);
    else await approvalApi.activateProcess(id);
    await reload();
  } catch (err) {
    formError.value = approvalErrorMessage(err, "Failed to update process state.");
  } finally {
    busy.value = false;
  }
}

onMounted(async () => {
  await Promise.all([reload(), fetchGroups(), userDirectory.ensureLoaded()]);
});
</script>

<template>
  <section class="proc-page">
    <AppPageHeader
      :title="t('approval.config.processes.title', 'Approval processes')"
      :description="t('approval.config.processes.description', 'Configure approval processes and stages by type and contract.')"
    >
      <template #actions>
        <AppButton variant="primary" size="sm" @click="startCreate">
          {{ t('approval.config.processes.new', 'New process') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="forbidden">
      <p class="proc-page__state">{{ t('approval.errors.noPermission', 'You do not have permission to view approval configuration.') }}</p>
    </AppCard>

    <template v-else>
      <AppCard v-if="showForm" title="New approval process">
        <ApprovalProcessConfigForm :groups="groups" :users="userDirectory.items.value" :submitting="submitting" @submit="saveProcess" @cancel="showForm = false" />
        <p v-if="formError" class="proc-page__error">{{ formError }}</p>
      </AppCard>

      <AppCard :title="t('approval.config.processes.listTitle', 'Processes')">
        <AppLoadingState v-if="loading" />
        <p v-else-if="error" class="proc-page__error">{{ error }}</p>
        <p v-else-if="processes.length === 0" class="proc-page__state">
          {{ t('approval.config.processes.empty', 'No approval processes configured yet. Submitting a subject will return "Approval process is not configured" until one exists.') }}
        </p>
        <div v-else class="proc-page__list">
          <div v-for="p in processes" :key="p.id" class="proc-page__item">
            <div class="proc-page__item-head">
              <div>
                <div class="proc-page__code">{{ p.process_code }}</div>
                <div class="proc-page__name">{{ p.process_name }}</div>
                <div class="proc-page__meta">
                  {{ prettify(p.process_type ?? '') }} · {{ p.contract_type }}
                  <span v-if="p.contract_id"> · contract-specific</span>
                </div>
              </div>
              <div class="proc-page__item-actions">
                <AppStatusBadge :status="p.is_active ? 'active' : 'inactive'" />
                <AppButton
                  size="sm"
                  :variant="p.is_active ? 'ghost' : 'success'"
                  :disabled="busy"
                  @click="toggleActive(p.id ?? '', !!p.is_active)"
                >
                  {{ p.is_active ? 'Deactivate' : 'Activate' }}
                </AppButton>
              </div>
            </div>
            <ol class="proc-page__stages">
              <li v-for="st in p.stages" :key="st.id">
                <strong>Stage {{ st.stage_number }}</strong>
                — {{ st.stage_name || prettify(st.approver_mode ?? '') }}
                <span class="proc-page__stage-mode">({{ st.approver_mode }}<span v-if="st.required_approval_count && st.required_approval_count > 1">, needs {{ st.required_approval_count }}</span>)</span>
                <span v-if="st.is_final_stage" class="proc-page__final">final</span>
              </li>
            </ol>
          </div>
        </div>
      </AppCard>
    </template>
  </section>
</template>

<style scoped>
.proc-page {
  display: grid;
  gap: var(--space-5, 20px);
}
.proc-page__state {
  margin: 0;
  color: var(--text-secondary, #57606a);
}
.proc-page__error {
  margin: var(--space-2, 8px) 0 0;
  color: var(--alert-danger-text, #cf222e);
  font-size: var(--font-size-sm, 0.875rem);
}
.proc-page__list {
  display: grid;
  gap: var(--space-4, 16px);
}
.proc-page__item {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  padding: var(--space-4, 16px);
}
.proc-page__item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4, 16px);
}
.proc-page__code {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary, #6e7781);
}
.proc-page__name {
  font-weight: var(--font-weight-semibold, 600);
}
.proc-page__meta {
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-secondary, #57606a);
  margin-top: 2px;
}
.proc-page__item-actions {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
}
.proc-page__stages {
  margin: var(--space-3, 12px) 0 0;
  padding-left: var(--space-5, 20px);
  display: grid;
  gap: var(--space-1, 4px);
  font-size: var(--font-size-sm, 0.875rem);
}
.proc-page__stage-mode {
  color: var(--text-tertiary, #6e7781);
}
.proc-page__final {
  margin-left: var(--space-2, 8px);
  font-size: var(--font-size-xs, 0.7rem);
  text-transform: uppercase;
  color: var(--accent-orange, #ea580c);
  font-weight: var(--font-weight-bold, 700);
}
</style>
