<script setup lang="ts">
import { reactive, ref, watch } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppFormField from "~/shared/ui/AppFormField.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import AppCheckbox from "~/shared/ui/AppCheckbox.vue";

import ApprovalStageBuilder from "./ApprovalStageBuilder.vue";
import { CONTRACT_TYPES, PROCESS_TYPES } from "../types";
import type { ApprovalGroup, ProcessConfigInput, StageInput } from "../types";

const props = defineProps<{ groups: ApprovalGroup[]; submitting?: boolean }>();
const emit = defineEmits<{ submit: [payload: ProcessConfigInput]; cancel: [] }>();

const form = reactive<Omit<ProcessConfigInput, "stages">>({
  process_code: "",
  process_name: "",
  process_type: "INVESTMENT_ANALYSIS_REPORT",
  contract_type: "COMPANY",
  contract_id: "",
  effective_date: "",
  is_active: true,
  group_approval_enabled: true,
  require_team_approval: false,
});

const stages = ref<StageInput[]>([
  {
    stage_number: 1,
    stage_name: "Reviewer sign-off",
    approver_mode: "GROUP_ANY",
    approver_user_id: "",
    approval_group_id: "",
    required_approval_count: 1,
    is_final_stage: true,
  },
]);

const localError = ref<string | null>(null);

watch(
  () => props.groups,
  () => {
    // keep stages' selected group valid is left to the operator; no auto-mutation
  },
);

function save() {
  localError.value = null;
  if (!form.process_code?.trim() || !form.process_name?.trim()) {
    localError.value = "Process code and name are required.";
    return;
  }
  if (stages.value.length === 0) {
    localError.value = "At least one approval stage is required.";
    return;
  }
  for (const s of stages.value) {
    if ((s.approver_mode === "GROUP_ANY" || s.approver_mode === "GROUP_PRIORITY") && !s.approval_group_id) {
      localError.value = `Stage ${s.stage_number}: select an approval group.`;
      return;
    }
    if (s.approver_mode === "SINGLE_USER" && !s.approver_user_id) {
      localError.value = `Stage ${s.stage_number}: enter an approver user.`;
      return;
    }
  }
  emit("submit", { ...form, stages: stages.value });
}
</script>

<template>
  <form class="proc-form" @submit.prevent="save">
    <div class="proc-form__grid">
      <AppFormField label="Process code" required>
        <AppInput v-model="form.process_code" placeholder="PROC_ANALYSIS_DEFAULT" />
      </AppFormField>
      <AppFormField label="Process name" required>
        <AppInput v-model="form.process_name" placeholder="Analysis Report Approval" />
      </AppFormField>
      <AppFormField label="Process type" required>
        <AppSelect v-model="form.process_type" :options="[...PROCESS_TYPES]" />
      </AppFormField>
      <AppFormField label="Contract type">
        <AppSelect v-model="form.contract_type" :options="[...CONTRACT_TYPES]" />
      </AppFormField>
      <AppFormField label="Applicable Scope" hint="Leave blank for a global / company-wide process.">
        <AppInput :model-value="''" disabled placeholder="Scope selector source not available" />
      </AppFormField>
      <AppFormField label="Effective date" hint="YYYY-MM-DD; blank = today.">
        <AppInput v-model="form.effective_date" placeholder="2026-01-01" />
      </AppFormField>
    </div>

    <div class="proc-form__flags">
      <AppCheckbox v-model="form.group_approval_enabled" label="Group approval enabled" />
      <AppCheckbox v-model="form.require_team_approval" label="Require team approval (stage 1 uses contract team)" />
      <AppCheckbox v-model="form.is_active" label="Active" />
    </div>

    <h3 class="proc-form__section">Approval stages</h3>
    <ApprovalStageBuilder v-model="stages" :groups="groups" />

    <p v-if="localError" class="proc-form__error" role="alert">{{ localError }}</p>

    <div class="proc-form__actions">
      <AppButton type="submit" variant="primary" :loading="submitting">Create process</AppButton>
      <AppButton type="button" variant="ghost" @click="emit('cancel')">Cancel</AppButton>
    </div>
  </form>
</template>

<style scoped>
.proc-form {
  display: grid;
  gap: var(--space-4, 16px);
}
.proc-form__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-3, 12px);
}
.proc-form__flags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4, 16px);
}
.proc-form__section {
  margin: 0;
  font-size: var(--font-size-md, 1rem);
}
.proc-form__error {
  margin: 0;
  color: var(--alert-danger-text, #cf222e);
  font-size: var(--font-size-sm, 0.875rem);
}
.proc-form__actions {
  display: flex;
  gap: var(--space-2, 8px);
}
</style>
