<script setup lang="ts">
import { computed } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppFormField from "~/shared/ui/AppFormField.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import AppCheckbox from "~/shared/ui/AppCheckbox.vue";

import { APPROVER_MODES } from "../types";
import type { ApprovalGroup, StageInput } from "../types";

const props = defineProps<{ groups: ApprovalGroup[] }>();

const stages = defineModel<StageInput[]>({ default: () => [] });

const groupOptions = computed(() =>
  props.groups.map((g) => ({ value: g.id ?? "", label: g.group_name || g.group_code || "" })),
);

function addStage() {
  if (stages.value.length >= 3) return;
  stages.value = [
    ...stages.value,
    {
      stage_number: stages.value.length + 1,
      stage_name: "",
      approver_mode: "GROUP_ANY",
      approver_user_id: "",
      approval_group_id: "",
      required_approval_count: 1,
      is_final_stage: stages.value.length === 0,
    },
  ];
}

function removeStage(index: number) {
  const next = stages.value.filter((_, i) => i !== index);
  next.forEach((s, i) => (s.stage_number = i + 1));
  if (next.length > 0 && !next.some((s) => s.is_final_stage)) {
    next[next.length - 1].is_final_stage = true;
  }
  stages.value = next;
}

function needsGroup(mode?: string): boolean {
  return mode === "GROUP_ANY" || mode === "GROUP_PRIORITY";
}
function needsUser(mode?: string): boolean {
  return mode === "SINGLE_USER";
}
function needsCount(mode?: string): boolean {
  return mode === "GROUP_ANY" || mode === "TEAM_MINIMUM";
}
</script>

<template>
  <div class="stage-builder">
    <div v-for="(stage, i) in stages" :key="i" class="stage-builder__stage">
      <div class="stage-builder__head">
        <h4>Stage {{ i + 1 }}</h4>
        <AppButton size="sm" variant="ghost" @click="removeStage(i)">Remove</AppButton>
      </div>
      <div class="stage-builder__grid">
        <AppFormField label="Stage name">
          <AppInput v-model="stage.stage_name" placeholder="Reviewer sign-off" />
        </AppFormField>
        <AppFormField label="Approver mode">
          <AppSelect v-model="stage.approver_mode" :options="[...APPROVER_MODES]" />
        </AppFormField>
        <AppFormField v-if="needsGroup(stage.approver_mode)" label="Approval group">
          <AppSelect v-model="stage.approval_group_id" :options="groupOptions" placeholder="Select a group…" />
        </AppFormField>
        <AppFormField v-if="needsUser(stage.approver_mode)" label="Approver user UUID">
          <AppInput v-model="stage.approver_user_id" placeholder="User UUID" />
        </AppFormField>
        <AppFormField v-if="needsCount(stage.approver_mode)" label="Required approvals">
          <AppInput v-model="stage.required_approval_count" type="number" />
        </AppFormField>
      </div>
      <AppCheckbox v-model="stage.is_final_stage" label="Final stage (approval completes here)" />
    </div>

    <AppButton v-if="stages.length < 3" variant="secondary" size="sm" @click="addStage">
      + Add stage ({{ stages.length }}/3)
    </AppButton>
    <p v-if="stages.length === 0" class="stage-builder__hint">Add at least one approval stage (max 3).</p>
  </div>
</template>

<style scoped>
.stage-builder {
  display: grid;
  gap: var(--space-4, 16px);
}
.stage-builder__stage {
  border: 1px solid var(--border-subtle, #d0d7de);
  border-radius: var(--radius-md, 6px);
  padding: var(--space-4, 16px);
  display: grid;
  gap: var(--space-3, 12px);
}
.stage-builder__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.stage-builder__head h4 {
  margin: 0;
  font-size: var(--font-size-md, 1rem);
}
.stage-builder__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: var(--space-3, 12px);
}
.stage-builder__hint {
  margin: 0;
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary, #6e7781);
}
</style>
