<script setup lang="ts">
import { reactive, computed } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";

import { TEAM_MEMBER_TYPES } from "../types";
import type { ApprovalTeamMember, TeamMemberInput } from "../types";

const props = defineProps<{
  members: ApprovalTeamMember[];
  users?: Array<{ id: string; display_name?: string; username?: string }>;
  busy?: boolean;
}>();

const userOptions = computed(() =>
  (props.users ?? []).map((u) => ({ value: u.id, label: u.display_name || u.username || u.id })),
);

const emit = defineEmits<{
  add: [payload: TeamMemberInput];
  remove: [memberId: string];
}>();

const draft = reactive<{ user_id: string; member_type: string; priority_order: number }>({
  user_id: "",
  member_type: "REVIEWER_AGENT",
  priority_order: 1,
});

function addMember() {
  if (!draft.user_id.trim()) return;
  emit("add", {
    user_id: draft.user_id.trim(),
    member_type: draft.member_type,
    priority_order: Number(draft.priority_order) || 1,
    is_active: true,
  });
  draft.user_id = "";
}
</script>

<template>
  <div class="tm-table">
    <table class="tm-table__table">
      <thead>
        <tr>
          <th>Priority</th>
          <th>User</th>
          <th>Role</th>
          <th class="tm-table__right">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="members.length === 0">
          <td colspan="4" class="tm-table__empty">No team members yet.</td>
        </tr>
        <tr v-for="m in members" :key="m.id">
          <td>{{ m.priority_order }}</td>
          <td>
            <div class="tm-table__user">
              <span>{{ m.display_name || "—" }}</span>
            </div>
          </td>
          <td>{{ m.member_type }}</td>
          <td class="tm-table__right">
            <AppButton size="sm" variant="danger" :disabled="busy" @click="emit('remove', m.id ?? '')">Remove</AppButton>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="tm-table__add">
      <AppSelect v-if="users && users.length" v-model="draft.user_id" :options="userOptions" placeholder="Select a user…" />
      <AppInput v-else v-model="draft.user_id" placeholder="Enter user UUID…" />
      <AppSelect v-model="draft.member_type" :options="[...TEAM_MEMBER_TYPES]" />
      <AppInput v-model="draft.priority_order" type="number" placeholder="Priority" />
      <AppButton variant="primary" :disabled="busy || !draft.user_id" @click="addMember">Add member</AppButton>
    </div>
    <p class="tm-table__hint">
      Only REVIEWER_AGENT members can approve. ORDER_SUBMITTER members represent the order maker (excluded from approving).
    </p>
  </div>
</template>

<style scoped>
.tm-table {
  display: grid;
  gap: var(--space-3, 12px);
}
.tm-table__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm, 0.875rem);
}
.tm-table__table th,
.tm-table__table td {
  text-align: left;
  padding: var(--space-2, 8px);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}
.tm-table__right {
  text-align: right;
}
.tm-table__empty {
  color: var(--text-tertiary, #6e7781);
  text-align: center;
}
.tm-table__user {
  display: flex;
  flex-direction: column;
}
.tm-table__uid {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 0.7rem);
  color: var(--text-tertiary, #6e7781);
}
.tm-table__add {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr auto;
  gap: var(--space-2, 8px);
  align-items: end;
}
.tm-table__hint {
  margin: 0;
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary, #6e7781);
}
</style>
