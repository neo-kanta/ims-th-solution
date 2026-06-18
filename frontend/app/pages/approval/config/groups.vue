<script setup lang="ts">
import { onMounted, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

import ApprovalGroupForm from "~/features/approval/components/ApprovalGroupForm.vue";
import ApprovalGroupMemberTable from "~/features/approval/components/ApprovalGroupMemberTable.vue";
import { useApprovalGroups } from "~/features/approval/composables/useApprovalConfig";
import { approvalApi, approvalErrorMessage } from "~/features/approval/services/approvalApi";
import type { ApprovalGroup, ApprovalGroupMember, GroupInput, GroupMemberInput } from "~/features/approval/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "APPROVAL_CONFIG_VIEW",
});

const { t } = useI18n();
const { groups, loading, error, forbidden, fetchGroups } = useApprovalGroups();

const showForm = ref(false);
const editing = ref<ApprovalGroup | null>(null);
const submitting = ref(false);
const formError = ref<string | null>(null);

const selected = ref<ApprovalGroup | null>(null);
const members = ref<ApprovalGroupMember[]>([]);
const membersLoading = ref(false);
const memberBusy = ref(false);
const memberError = ref<string | null>(null);

async function reloadGroups() {
  await fetchGroups();
}

function startCreate() {
  editing.value = null;
  showForm.value = true;
  formError.value = null;
}

function startEdit(g: ApprovalGroup) {
  editing.value = g;
  showForm.value = true;
  formError.value = null;
}

async function saveGroup(payload: GroupInput) {
  submitting.value = true;
  formError.value = null;
  try {
    if (editing.value?.id) {
      await approvalApi.updateGroup(editing.value.id, payload);
    } else {
      await approvalApi.createGroup(payload);
    }
    showForm.value = false;
    await reloadGroups();
  } catch (err) {
    formError.value = approvalErrorMessage(err, "Failed to save group.");
  } finally {
    submitting.value = false;
  }
}

async function selectGroup(g: ApprovalGroup) {
  selected.value = g;
  await loadMembers();
}

async function loadMembers() {
  if (!selected.value?.id) return;
  membersLoading.value = true;
  memberError.value = null;
  try {
    members.value = await approvalApi.listGroupMembers(selected.value.id);
  } catch (err) {
    memberError.value = approvalErrorMessage(err, "Failed to load members.");
    members.value = [];
  } finally {
    membersLoading.value = false;
  }
}

async function addMember(payload: GroupMemberInput) {
  if (!selected.value?.id) return;
  await runMember(() => approvalApi.addGroupMember(selected.value!.id!, payload), "Failed to add member.");
}
async function approveMember(memberId: string) {
  if (!selected.value?.id) return;
  await runMember(() => approvalApi.approveGroupMember(selected.value!.id!, memberId), "Failed to approve member.");
}
async function revokeMember(memberId: string) {
  if (!selected.value?.id) return;
  await runMember(() => approvalApi.revokeGroupMember(selected.value!.id!, memberId), "Failed to revoke member.");
}
async function reorder(ids: string[]) {
  if (!selected.value?.id) return;
  await runMember(() => approvalApi.reorderGroupMembers(selected.value!.id!, { member_ids: ids }), "Failed to reorder members.");
}

async function runMember(fn: () => Promise<unknown>, fallback: string) {
  memberBusy.value = true;
  memberError.value = null;
  try {
    await fn();
    await loadMembers();
  } catch (err) {
    memberError.value = approvalErrorMessage(err, fallback);
  } finally {
    memberBusy.value = false;
  }
}

onMounted(reloadGroups);
</script>

<template>
  <section class="groups-page">
    <AppPageHeader
      :title="t('approval.config.groups.title', 'Approval groups')"
      :description="t('approval.config.groups.description', 'Manage reusable approver groups and members.')"
    >
      <template #actions>
        <AppButton variant="primary" size="sm" @click="startCreate">
          {{ t('approval.config.groups.new', 'New group') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="forbidden">
      <p class="groups-page__state">{{ t('approval.errors.noPermission', 'You do not have permission to view approval configuration.') }}</p>
    </AppCard>

    <template v-else>
      <!-- Form Card -->
      <AppCard v-if="showForm" :title="editing ? t('approval.config.groups.edit', 'Edit group') : t('approval.config.groups.new', 'New group')" class="groups-page__form-card">
        <ApprovalGroupForm :group="editing" :submitting="submitting" @submit="saveGroup" @cancel="showForm = false" />
        <p v-if="formError" class="groups-page__error">{{ formError }}</p>
      </AppCard>

      <!-- Two-Column Layout -->
      <div class="groups-page__layout">
        <!-- Left Column: Groups Sidebar -->
        <div class="groups-page__sidebar">
          <AppCard :title="t('approval.config.groups.listTitle', 'Groups')" class="groups-page__list-card">
            <AppLoadingState v-if="loading" />
            <p v-else-if="error" class="groups-page__error">{{ error }}</p>
            <p v-else-if="groups.length === 0" class="groups-page__state">
              {{ t('approval.config.groups.empty', 'No approval groups configured yet.') }}
            </p>
            <div v-else class="groups-list">
              <div
                v-for="g in groups"
                :key="g.id"
                class="group-item"
                :class="{ 'is-selected': selected?.id === g.id }"
                @click="selectGroup(g)"
              >
                <div class="group-item__content">
                  <div class="group-item__header">
                    <span class="group-item__name">{{ g.group_name }}</span>
                    <AppStatusBadge :status="g.is_active ? 'active' : 'inactive'" size="sm" />
                  </div>
                  <span class="group-item__code">{{ g.group_code }}</span>
                </div>
                <div class="group-item__actions" @click.stop>
                  <AppButton size="xs" variant="ghost" @click="startEdit(g)">Edit</AppButton>
                </div>
              </div>
            </div>
          </AppCard>
        </div>

        <!-- Right Column: Selected Group Detail & Members -->
        <div class="groups-page__detail">
          <AppCard
            v-if="selected"
            :title="t('approval.config.groups.membersTitle', 'Members — ') + (selected.group_name || '')"
            class="groups-page__detail-card"
          >
            <AppLoadingState v-if="membersLoading" />
            <template v-else>
              <ApprovalGroupMemberTable
                :members="members"
                :busy="memberBusy"
                @add="addMember"
                @approve="approveMember"
                @revoke="revokeMember"
                @reorder="reorder"
              />
              <p v-if="memberError" class="groups-page__error">{{ memberError }}</p>
            </template>
          </AppCard>

          <AppCard v-else class="groups-page__placeholder-card">
            <div class="groups-page__placeholder-content">
              <span class="groups-page__placeholder-icon">👥</span>
              <h3>No group selected</h3>
              <p>Select an approval group from the sidebar to manage its members and priorities.</p>
            </div>
          </AppCard>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.groups-page {
  display: grid;
  gap: var(--space-5, 20px);
}
.groups-page__state {
  margin: 0;
  color: var(--text-secondary, #57606a);
}
.groups-page__error {
  margin: var(--space-2, 8px) 0 0;
  color: var(--alert-danger-text, #cf222e);
  font-size: var(--font-size-sm, 0.875rem);
}
.groups-page__layout {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: var(--space-4, 16px);
  align-items: start;
}
@media (max-width: 992px) {
  .groups-page__layout {
    grid-template-columns: 1fr;
  }
}
.groups-page__form-card {
  margin-bottom: var(--space-2, 8px);
}
.groups-list {
  display: grid;
  gap: var(--space-2, 8px);
}
.group-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3, 12px) var(--space-4, 16px);
  border: 1px solid var(--border-subtle, #e1e4e8);
  border-radius: var(--radius-md, 6px);
  background: var(--bg-card, #ffffff);
  cursor: pointer;
  transition: all 0.15s ease;
  position: relative;
  overflow: hidden;
}
.group-item:hover {
  border-color: var(--color-primary-muted, #0969da);
  background: var(--bg-card-hover, #f6f8fa);
}
.group-item.is-selected {
  background: var(--bg-selected, #ddf4ff) !important;
  border-color: var(--color-primary, #0969da) !important;
}
.group-item.is-selected::before {
  content: "";
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  background: var(--color-primary, #0969da);
}
.group-item__content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
  padding-right: var(--space-2, 8px);
}
.group-item__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2, 8px);
}
.group-item__name {
  font-weight: 600;
  color: var(--text-primary, #24292f);
  font-size: var(--font-size-sm, 0.875rem);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.group-item__code {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 0.75rem);
  color: var(--text-secondary, #57606a);
}
.group-item__actions {
  opacity: 0.7;
}
.group-item:hover .group-item__actions {
  opacity: 1;
}
.groups-page__placeholder-card {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  border: 1px dashed var(--border-subtle, #e1e4e8);
  background: var(--bg-card-muted, #f8f9fa);
  color: var(--text-secondary, #57606a);
}
.groups-page__placeholder-content {
  text-align: center;
  max-width: 320px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2, 8px);
  padding: var(--space-5, 24px) 0;
}
.groups-page__placeholder-icon {
  font-size: var(--font-size-xl, 2rem);
  opacity: 0.6;
}
.groups-page__placeholder-content h3 {
  margin: 0;
  font-size: var(--font-size-md, 1rem);
  font-weight: 600;
  color: var(--text-primary, #24292f);
}
.groups-page__placeholder-content p {
  margin: 0;
  font-size: var(--font-size-xs, 0.75rem);
  line-height: 1.4;
}
</style>
