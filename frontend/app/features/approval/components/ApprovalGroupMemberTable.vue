<script setup lang="ts">
import { reactive, computed } from "vue";

import AppButton from "~/shared/ui/AppButton.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppSelect from "~/shared/ui/AppSelect.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import AppFormField from "~/shared/ui/AppFormField.vue";

import { GROUP_MEMBER_TYPES } from "../types";
import type { ApprovalGroupMember, GroupMemberInput } from "../types";

const props = defineProps<{
  members: ApprovalGroupMember[];
  users?: Array<{ id: string; display_name?: string; username?: string }>;
  busy?: boolean;
}>();

const userOptions = computed(() =>
  (props.users ?? []).map((u) => ({ value: u.id, label: u.display_name || u.username || u.id })),
);

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
  const itemA = ids[index];
  const itemB = ids[target];
  if (itemA !== undefined && itemB !== undefined) {
    ids[index] = itemB;
    ids[target] = itemA;
  }
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
    <div class="member-table__scroll">
      <table class="member-table__table">
        <thead>
          <tr>
            <th style="width: 100px;">Priority</th>
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
                <span class="member-table__priority-num">{{ m.priority_order }}</span>
                <div class="member-table__arrows">
                  <button
                    type="button"
                    class="member-table__arrow-btn"
                    title="Move up"
                    :disabled="i === 0 || busy"
                    @click="move(i, -1)"
                  >
                    ▲
                  </button>
                  <button
                    type="button"
                    class="member-table__arrow-btn"
                    title="Move down"
                    :disabled="i === members.length - 1 || busy"
                    @click="move(i, 1)"
                  >
                    ▼
                  </button>
                </div>
              </div>
            </td>
            <td>
              <div class="member-table__user">
                <span class="member-table__username">{{ m.display_name || "—" }}</span>
              </div>
            </td>
            <td>
              <span class="member-table__type-badge">{{ m.member_type }}</span>
            </td>
            <td><AppStatusBadge :status="statusKeyword(m.status)" :label="m.status" /></td>
            <td class="member-table__right">
              <AppButton
                v-if="m.status !== 'APPROVED'"
                size="xs"
                variant="success"
                :disabled="busy"
                @click="emit('approve', m.id ?? '')"
              >
                Approve
              </AppButton>
              <AppButton
                v-if="m.status !== 'REVOKED'"
                size="xs"
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
    </div>

    <div class="member-table__add-section">
      <h4 class="member-table__add-title">Add Group Member</h4>
      <div class="member-table__add-grid">
        <AppFormField label="User">
          <AppSelect v-if="users && users.length" v-model="draft.user_id" :options="userOptions" placeholder="Select a user…" />
          <AppInput v-else v-model="draft.user_id" placeholder="Enter user UUID…" />
        </AppFormField>
        <AppFormField label="Member Type">
          <AppSelect v-model="draft.member_type" :options="[...GROUP_MEMBER_TYPES]" />
        </AppFormField>
        <AppFormField label="Priority Order">
          <AppInput v-model="draft.priority_order" type="number" placeholder="Priority" />
        </AppFormField>
        <div class="member-table__add-action">
          <AppButton variant="primary" :disabled="busy || !draft.user_id" @click="addMember">Add member</AppButton>
        </div>
      </div>
    </div>
    
    <p class="member-table__hint">
      💡 Members must be APPROVED and active to be selected as approvers. Priority order is used by GROUP_PRIORITY stages.
    </p>
  </div>
</template>

<style scoped>
.member-table {
  display: grid;
  gap: var(--space-4, 16px);
}
.member-table__scroll {
  overflow-x: auto;
  border: 1px solid var(--border-subtle, #e1e4e8);
  border-radius: var(--radius-md, 6px);
}
.member-table__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm, 0.875rem);
}
.member-table__table th {
  background: var(--bg-table-header, #f6f8fa);
  color: var(--text-secondary, #57606a);
  font-weight: 600;
  padding: var(--space-3, 12px) var(--space-4, 16px);
  border-bottom: 1px solid var(--border-subtle, #e1e4e8);
  text-align: left;
}
.member-table__table td {
  padding: var(--space-3, 12px) var(--space-4, 16px);
  border-bottom: 1px solid var(--border-subtle, #e1e4e8);
  vertical-align: middle;
}
.member-table__table tr:last-child td {
  border-bottom: none;
}
.member-table__table tr:hover {
  background-color: var(--bg-row-hover, #fafbfc);
}
.member-table__right {
  text-align: right;
  display: flex;
  gap: var(--space-2, 8px);
  justify-content: flex-end;
  align-items: center;
}
.member-table__empty {
  color: var(--text-tertiary, #6e7781);
  text-align: center;
  padding: var(--space-5, 24px) !important;
}
.member-table__priority {
  display: flex;
  align-items: center;
  gap: var(--space-3, 12px);
}
.member-table__priority-num {
  font-weight: 600;
  font-size: var(--font-size-md, 1rem);
  color: var(--text-primary, #24292f);
}
.member-table__arrows {
  display: flex;
  gap: 2px;
}
.member-table__arrow-btn {
  border: 1px solid var(--border-subtle, #e1e4e8);
  background: var(--bg-card, #ffffff);
  border-radius: var(--radius-sm, 4px);
  cursor: pointer;
  color: var(--text-secondary, #57606a);
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 8px;
  transition: all 0.15s ease;
}
.member-table__arrow-btn:hover:not(:disabled) {
  border-color: var(--color-primary, #0969da);
  color: var(--color-primary, #0969da);
  background-color: var(--bg-selected, #ddf4ff);
}
.member-table__arrow-btn:disabled {
  opacity: 0.25;
  cursor: not-allowed;
}
.member-table__user {
  display: flex;
  flex-direction: column;
  line-height: 1.3;
}
.member-table__username {
  font-weight: 600;
  color: var(--text-primary, #24292f);
}
.member-table__uid {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-tertiary, #6e7781);
}
.member-table__type-badge {
  font-size: var(--font-size-xs, 0.75rem);
  font-weight: 500;
  color: var(--text-secondary, #57606a);
  background: var(--bg-card-hover, #f6f8fa);
  padding: 2px 6px;
  border-radius: var(--radius-sm, 4px);
  border: 1px solid var(--border-subtle, #e1e4e8);
}
.member-table__add-section {
  background: var(--bg-card-muted, #f8f9fa);
  border: 1px solid var(--border-subtle, #e1e4e8);
  border-radius: var(--radius-md, 6px);
  padding: var(--space-4, 16px);
}
.member-table__add-title {
  margin: 0 0 var(--space-3, 12px);
  font-size: var(--font-size-sm, 0.875rem);
  font-weight: 600;
  color: var(--text-primary, #24292f);
}
.member-table__add-grid {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr auto;
  gap: var(--space-3, 12px);
  align-items: flex-end;
}
@media (max-width: 768px) {
  .member-table__add-grid {
    grid-template-columns: 1fr;
    align-items: stretch;
  }
}
.member-table__add-action {
  margin-bottom: 2px;
}
.member-table__hint {
  margin: 0;
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-secondary, #57606a);
  line-height: 1.4;
}
</style>
