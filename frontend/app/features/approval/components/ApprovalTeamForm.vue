<script setup lang="ts">
import { reactive, watch } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppFormField from "~/shared/ui/AppFormField.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppTextarea from "~/shared/ui/AppTextarea.vue";
import AppCheckbox from "~/shared/ui/AppCheckbox.vue";

import type { ApprovalTeam, TeamInput } from "../types";

const props = defineProps<{ team?: ApprovalTeam | null; submitting?: boolean }>();
const emit = defineEmits<{ submit: [payload: TeamInput]; cancel: [] }>();

const form = reactive<TeamInput>({
  team_code: "",
  team_name: "",
  remarks: "",
  has_co_manager: false,
  min_required_stamps: 1,
  max_allowed_stamps: 1,
  is_active: true,
});

watch(
  () => props.team,
  (t) => {
    form.team_code = t?.team_code ?? "";
    form.team_name = t?.team_name ?? "";
    form.remarks = t?.remarks ?? "";
    form.has_co_manager = t?.has_co_manager ?? false;
    form.min_required_stamps = t?.min_required_stamps ?? 1;
    form.max_allowed_stamps = t?.max_allowed_stamps ?? 1;
    form.is_active = t?.is_active ?? true;
  },
  { immediate: true },
);

const isEdit = () => Boolean(props.team?.id);

function save() {
  emit("submit", {
    ...form,
    min_required_stamps: Number(form.min_required_stamps) || 1,
    max_allowed_stamps: Number(form.max_allowed_stamps) || 1,
  });
}
</script>

<template>
  <form class="team-form" @submit.prevent="save">
    <AppFormField label="Team code" required>
      <AppInput v-model="form.team_code" :disabled="isEdit()" placeholder="SCB_FIXED_TEAM" />
    </AppFormField>
    <AppFormField label="Team name" required>
      <AppInput v-model="form.team_name" placeholder="SCB Fixed Income Team" />
    </AppFormField>
    <AppFormField label="Remarks">
      <AppTextarea v-model="form.remarks" :rows="2" />
    </AppFormField>
    <div class="team-form__row">
      <AppFormField label="Min required stamps">
        <AppInput v-model="form.min_required_stamps" type="number" />
      </AppFormField>
      <AppFormField label="Max allowed stamps">
        <AppInput v-model="form.max_allowed_stamps" type="number" />
      </AppFormField>
    </div>
    <AppCheckbox v-model="form.has_co_manager" label="Has co-manager" />
    <AppCheckbox v-model="form.is_active" label="Active" />
    <div class="team-form__actions">
      <AppButton type="submit" variant="primary" :loading="submitting">
        {{ isEdit() ? "Save changes" : "Create team" }}
      </AppButton>
      <AppButton type="button" variant="ghost" @click="emit('cancel')">Cancel</AppButton>
    </div>
  </form>
</template>

<style scoped>
.team-form {
  display: grid;
  gap: var(--space-3, 12px);
}
.team-form__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-3, 12px);
}
.team-form__actions {
  display: flex;
  gap: var(--space-2, 8px);
  margin-top: var(--space-2, 8px);
}
</style>
