<script setup lang="ts">
import { reactive, watch } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppFormField from "~/shared/ui/AppFormField.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppTextarea from "~/shared/ui/AppTextarea.vue";
import AppCheckbox from "~/shared/ui/AppCheckbox.vue";

import type { ApprovalGroup, GroupInput } from "../types";

const props = defineProps<{
  group?: ApprovalGroup | null;
  submitting?: boolean;
}>();

const emit = defineEmits<{
  submit: [payload: GroupInput];
  cancel: [];
}>();

const form = reactive<GroupInput>({
  group_code: "",
  group_name: "",
  remarks: "",
  is_active: true,
});

watch(
  () => props.group,
  (g) => {
    form.group_code = g?.group_code ?? "";
    form.group_name = g?.group_name ?? "";
    form.remarks = g?.remarks ?? "";
    form.is_active = g?.is_active ?? true;
  },
  { immediate: true },
);

const isEdit = () => Boolean(props.group?.id);

function save() {
  emit("submit", { ...form });
}
</script>

<template>
  <form class="group-form" @submit.prevent="save">
    <AppFormField label="Group code" required>
      <AppInput v-model="form.group_code" :disabled="isEdit()" placeholder="FUND_MANAGER_REVIEWERS" />
    </AppFormField>
    <AppFormField label="Group name" required>
      <AppInput v-model="form.group_name" placeholder="Fund Manager Reviewers" />
    </AppFormField>
    <AppFormField label="Remarks">
      <AppTextarea v-model="form.remarks" :rows="2" />
    </AppFormField>
    <AppCheckbox v-model="form.is_active" label="Active" />
    <div class="group-form__actions">
      <AppButton type="submit" variant="primary" :loading="submitting">
        {{ isEdit() ? "Save changes" : "Create group" }}
      </AppButton>
      <AppButton type="button" variant="ghost" @click="emit('cancel')">Cancel</AppButton>
    </div>
  </form>
</template>

<style scoped>
.group-form {
  display: grid;
  gap: var(--space-3, 12px);
}
.group-form__actions {
  display: flex;
  gap: var(--space-2, 8px);
  margin-top: var(--space-2, 8px);
}
</style>
