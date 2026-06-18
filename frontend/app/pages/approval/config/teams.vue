<script setup lang="ts">
import { onMounted, ref } from "vue";

import { useI18n } from "~/composables/useI18n";
import AppPageHeader from "~/shared/ui/AppPageHeader.vue";
import AppCard from "~/shared/ui/AppCard.vue";
import AppButton from "~/shared/ui/AppButton.vue";
import AppInput from "~/shared/ui/AppInput.vue";
import AppStatusBadge from "~/shared/ui/AppStatusBadge.vue";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

import ApprovalTeamForm from "~/features/approval/components/ApprovalTeamForm.vue";
import ApprovalTeamMemberTable from "~/features/approval/components/ApprovalTeamMemberTable.vue";
import { useApprovalTeams } from "~/features/approval/composables/useApprovalConfig";
import { approvalApi, approvalErrorMessage } from "~/features/approval/services/approvalApi";
import type {
  ApprovalTeam,
  ApprovalTeamContract,
  ApprovalTeamMember,
  TeamInput,
  TeamMemberInput,
} from "~/features/approval/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "APPROVAL_CONFIG_VIEW",
});

const { t } = useI18n();
const { teams, loading, error, forbidden, fetchTeams } = useApprovalTeams();

const showForm = ref(false);
const editing = ref<ApprovalTeam | null>(null);
const submitting = ref(false);
const formError = ref<string | null>(null);

const selected = ref<ApprovalTeam | null>(null);
const members = ref<ApprovalTeamMember[]>([]);
const contracts = ref<ApprovalTeamContract[]>([]);
const detailLoading = ref(false);
const detailBusy = ref(false);
const detailError = ref<string | null>(null);
const newContractId = ref("");

async function reloadTeams() {
  await fetchTeams();
}

function startCreate() {
  editing.value = null;
  showForm.value = true;
  formError.value = null;
}
function startEdit(team: ApprovalTeam) {
  editing.value = team;
  showForm.value = true;
  formError.value = null;
}

async function saveTeam(payload: TeamInput) {
  submitting.value = true;
  formError.value = null;
  try {
    if (editing.value?.id) await approvalApi.updateTeam(editing.value.id, payload);
    else await approvalApi.createTeam(payload);
    showForm.value = false;
    await reloadTeams();
  } catch (err) {
    formError.value = approvalErrorMessage(err, "Failed to save team.");
  } finally {
    submitting.value = false;
  }
}

async function selectTeam(team: ApprovalTeam) {
  selected.value = team;
  await loadDetail();
}

async function loadDetail() {
  if (!selected.value?.id) return;
  detailLoading.value = true;
  detailError.value = null;
  try {
    members.value = await approvalApi.listTeamMembers(selected.value.id);
    contracts.value = await approvalApi.listTeamContracts(selected.value.id);
  } catch (err) {
    detailError.value = approvalErrorMessage(err, "Failed to load team detail.");
  } finally {
    detailLoading.value = false;
  }
}

async function runDetail(fn: () => Promise<unknown>, fallback: string) {
  detailBusy.value = true;
  detailError.value = null;
  try {
    await fn();
    await loadDetail();
  } catch (err) {
    detailError.value = approvalErrorMessage(err, fallback);
  } finally {
    detailBusy.value = false;
  }
}

function assignContract() {
  if (!selected.value?.id || !newContractId.value.trim()) return;
  const cid = newContractId.value.trim();
  void runDetail(
    () => approvalApi.assignTeamContract(selected.value!.id!, { contract_id: cid, effective_date: "" }),
    "Failed to assign contract.",
  ).then(() => (newContractId.value = ""));
}

function addMember(payload: TeamMemberInput) {
  if (!selected.value?.id) return;
  void runDetail(() => approvalApi.addTeamMember(selected.value!.id!, payload), "Failed to add member.");
}
function removeMember(memberId: string) {
  if (!selected.value?.id) return;
  void runDetail(() => approvalApi.removeTeamMember(selected.value!.id!, memberId), "Failed to remove member.");
}

onMounted(reloadTeams);
</script>

<template>
  <section class="teams-page">
    <AppPageHeader
      :title="t('approval.config.teams.title', 'Approval teams')"
      :description="t('approval.config.teams.description', 'Configure per contract/fund approval teams.')"
    >
      <template #actions>
        <AppButton variant="primary" size="sm" @click="startCreate">
          {{ t('approval.config.teams.new', 'New team') }}
        </AppButton>
      </template>
    </AppPageHeader>

    <AppCard v-if="forbidden">
      <p class="teams-page__state">{{ t('approval.errors.noPermission', 'You do not have permission to view approval configuration.') }}</p>
    </AppCard>

    <template v-else>
      <AppCard v-if="showForm" :title="editing ? 'Edit team' : 'New team'">
        <ApprovalTeamForm :team="editing" :submitting="submitting" @submit="saveTeam" @cancel="showForm = false" />
        <p v-if="formError" class="teams-page__error">{{ formError }}</p>
      </AppCard>

      <AppCard :title="t('approval.config.teams.listTitle', 'Teams')">
        <AppLoadingState v-if="loading" />
        <p v-else-if="error" class="teams-page__error">{{ error }}</p>
        <p v-else-if="teams.length === 0" class="teams-page__state">
          {{ t('approval.config.teams.empty', 'No approval teams configured yet.') }}
        </p>
        <table v-else class="teams-page__table">
          <thead>
            <tr><th>Code</th><th>Name</th><th>Min/Max</th><th>Status</th><th></th></tr>
          </thead>
          <tbody>
            <tr v-for="team in teams" :key="team.id" :class="{ 'is-selected': selected?.id === team.id }">
              <td class="teams-page__mono">{{ team.team_code }}</td>
              <td>{{ team.team_name }}</td>
              <td>{{ team.min_required_stamps }} / {{ team.max_allowed_stamps }}</td>
              <td><AppStatusBadge :status="team.is_active ? 'active' : 'inactive'" /></td>
              <td class="teams-page__right">
                <AppButton size="sm" variant="secondary" @click="selectTeam(team)">Manage</AppButton>
                <AppButton size="sm" variant="ghost" @click="startEdit(team)">Edit</AppButton>
              </td>
            </tr>
          </tbody>
        </table>
      </AppCard>

      <AppCard v-if="selected" :title="'Team — ' + (selected.team_name || '')">
        <AppLoadingState v-if="detailLoading" />
        <template v-else>
          <h4 class="teams-page__subhead">Contracts / funds</h4>
          <ul v-if="contracts.length" class="teams-page__contracts">
            <li v-for="c in contracts" :key="c.id">
              <span>Effective {{ c.effective_date || '—' }}</span>
              <AppStatusBadge :status="c.is_active ? 'active' : 'inactive'" size="sm" />
            </li>
          </ul>
          <p v-else class="teams-page__state">No contracts assigned to this team.</p>
          <div class="teams-page__assign">
            <AppInput v-model="newContractId" placeholder="Contract / fund identifier" />
            <AppButton variant="primary" :disabled="detailBusy || !newContractId" @click="assignContract">Assign contract</AppButton>
          </div>

          <h4 class="teams-page__subhead">Members</h4>
          <ApprovalTeamMemberTable :members="members" :busy="detailBusy" @add="addMember" @remove="removeMember" />
          <p v-if="detailError" class="teams-page__error">{{ detailError }}</p>
        </template>
      </AppCard>
    </template>
  </section>
</template>

<style scoped>
.teams-page {
  display: grid;
  gap: var(--space-5, 20px);
}
.teams-page__state {
  margin: 0;
  color: var(--text-secondary, #57606a);
}
.teams-page__error {
  margin: var(--space-2, 8px) 0 0;
  color: var(--alert-danger-text, #cf222e);
  font-size: var(--font-size-sm, 0.875rem);
}
.teams-page__table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm, 0.875rem);
}
.teams-page__table th,
.teams-page__table td {
  text-align: left;
  padding: var(--space-2, 8px);
  border-bottom: 1px solid var(--border-subtle, #d0d7de);
}
.teams-page__right {
  text-align: right;
  display: flex;
  gap: var(--space-2, 8px);
  justify-content: flex-end;
}
.teams-page__mono {
  font-family: var(--font-mono, monospace);
  font-size: var(--font-size-xs, 0.8rem);
}
.is-selected {
  background: var(--bg-selected, #ddf4ff);
}
.teams-page__subhead {
  margin: var(--space-4, 16px) 0 var(--space-2, 8px);
  font-size: var(--font-size-md, 1rem);
}
.teams-page__contracts {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: var(--space-2, 8px);
}
.teams-page__contracts li {
  display: flex;
  align-items: center;
  gap: var(--space-3, 12px);
}
.teams-page__assign {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: var(--space-2, 8px);
  margin-top: var(--space-3, 12px);
}
</style>
