<script setup lang="ts">
import { reactive } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";

import { GROUP_MEMBER_TYPES } from "../types";
import type { ApprovalGroupMember, GroupMemberInput } from "../types";

const props = defineProps<{
  members: ApprovalGroupMember[];
  busy?: boolean;
}>();

const emit = defineEmits<{
  add: [payload: GroupMemberInput];
  approve: [memberId: string];
  revoke: [memberId: string];
  reorder: [orderedIds: string[]];
}>();

const draft = reactive<{ user_id: string; member_type: string; priority_order: number }>({
  user_id: "",
  member_type: "MEMBER",
  priority_order: 1,
});

function addMember() {
  if (!draft.user_id.trim()) return;
  emit("add", {
    user_id: draft.user_id.trim(),
    member_type: draft.member_type,
    priority_order: Number(draft.priority_order) || 1,
    status: "APPROVED",
    is_active: true,
  });
  draft.user_id = "";
  draft.priority_order = props.members.length + 1;
}

function move(index: number, dir: -1 | 1) {
  const ids = props.members.map((m) => m.id ?? "");
  const target = index + dir;
  if (target < 0 || target >= ids.length) return;
  [ids[index], ids[target]] = [ids[target], ids[index]];
  emit("reorder", ids);
}

function statusKeyword(status?: string): string {
  if (status === "APPROVED") return "approved";
  if (status === "REVOKED") return "inactive";
  return "pending";
}
</script>

<template>
  <div class="member-table">
    <table class="member-table__table">
      <thead>
        <tr>
          <th>Priority</th>
          <th>User</th>
          <th>Type</th>
          <th>Status</th>
          <th class="member-table__right">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="members.length === 0">
          <td colspan="5" class="member-table__empty">No members yet. Add eligible approvers below.</td>
        </tr>
        <tr v-for="(m, i) in members" :key="m.id">
          <td>
            <div class="member-table__priority">
              <span>{{ m.priority_order }}</span>
              <div class="member-table__arrows">
                <button type="button" :disabled="i === 0 || busy" @click="move(i, -1)">▲</button>
                <button type="button" :disabled="i === members.length - 1 || busy" @click="move(i, 1)">▼</button>
              </div>
            </div>
          </td>
          <td>
            <div class="member-table__user">
              <span>{{ m.display_name || "—" }}</span>
              <span class="member-table__uid">{{ m.user_id }}</span>
            </div>
          </td>
          <td>{{ m.member_type }}</td>
          <td><AppStatusBadge :status="statusKeyword(m.status)" :label="m.status" /></td>
          <td class="member-table__right">
            <AppButton
              v-if="m.status !== 'APPROVED'"
              size="sm"
              variant="success"
              :disabled="busy"
              @click="emit('approve', m.id ?? '')"
            >
              Approve
            </AppButton>
            <AppButton
              v-if="m.status !== 'REVOKED'"
              size="sm"
              variant="danger"
              :disabled="busy"
              @click="emit('revoke', m.id ?? '')"
            >
              Revoke
            </AppButton>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="member-table__add">
      <AppInput v-model="draft.user_id" placeholder="User UUID" />
      <AppSelect v-model="draft.member_type" :options="[...GROUP_MEMBER_TYPES]" />
      <AppInput v-model="draft.priority_order" type="number" placeholder="Priority" />
      <AppButton variant="primary" :disabled="busy || !draft.user_id" @click="addMember">Add member</AppButton>
    </div>
    <p class="member-table__hint">
      Members must be APPROVED and active to be selected as approvers. Priority order is used by GROUP_PRIORITY stages.
    </p>
  </div>
</template>

<style scoped>
.member-table {
  display: grid;
  gap: var(--space-3, 12px);
}
.member-table__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm, 0.875rem);
}
.member-table__table th,
.member-table__table td {
  text-align: left;
  padding: var(--space-2, 8px);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}
.member-table__right {
  text-align: right;
  display: flex;
  gap: var(--space-2, 8px);
  justify-content: flex-end;
}
.member-table__empty {
  color: var(--text-tertiary, #6e7781);
  text-align: center;
}
.member-table__priority {
  display: flex;
  align-items: center;
  gap: var(--space-2, 8px);
}
.member-table__arrows {
  display: flex;
  flex-direction: column;
  font-size: 10px;
  line-height: 1;
}
.member-table__arrows button {
  border: none;
  background: none;
  cursor: pointer;
  color: var(--text-secondary, #57606a);
}
.member-table__arrows button:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}
.member-table__user {
  display: flex;
  flex-direction: column;
}
.member-table__uid {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 0.7rem);
  color: var(--text-tertiary, #6e7781);
}
.member-table__add {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr auto;
  gap: var(--space-2, 8px);
  align-items: end;
}
.member-table__hint {
  margin: 0;
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary, #6e7781);
}
</style>
