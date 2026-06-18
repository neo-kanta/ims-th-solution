<script setup lang="ts">
import { computed, onMounted, ref, reactive } from "vue";

import { permissionWorkflowApi } from "../services/permissionWorkflowApi";
import { myFundsApi } from "~/features/my-funds/services/myFundsApi";
import type { EffectivePermissions, PermissionUserSummary } from "../types";

const props = defineProps<{
  initialUserId?: string;
}>();

const users = ref<PermissionUserSummary[]>([]);
const selectedUserId = ref(props.initialUserId || "");
const effective = ref<EffectivePermissions | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

// Modal states
const showManageModal = ref(false);
const activeModalTab = ref("roles"); // "roles" | "groups" | "functions" | "data"
const modalLoading = ref(false);
const modalError = ref<string | null>(null);

// Form data
const selectedRole = ref("");
const roleReason = ref("");

const selectedGroup = ref("");
const groupAction = ref("ADD");
const groupReason = ref("");

const selectedFunction = ref("");
const functionReason = ref("");
const rights = reactive({
  canView: false,
  canSearch: false,
  canAdd: false,
  canEdit: false,
  canDelete: false,
  canApprove: false,
  canRevokeApproval: false,
  canExport: false,
  canConfigure: false,
});

const selectedFund = ref("");
const dataAccessLevel = ref("READ");
const dataReason = ref("");

const workflowApproveReason = ref("");

// Dropdown lists
const rolesList = ref<any[]>([]);
const groupsList = ref<any[]>([]);
const functionsList = ref<any[]>([]);
const fundsList = ref<any[]>([]);

const selectedUser = computed(() => users.value.find((user) => user.id === selectedUserId.value));

async function loadUsers() {
  const response = await permissionWorkflowApi.listUsers({ page: 1, limit: 100 });
  users.value = response.items;
  if (!selectedUserId.value && users.value[0]) {
    selectedUserId.value = users.value[0].id;
  }
}

async function loadEffective() {
  if (!selectedUserId.value) return;
  loading.value = true;
  error.value = null;
  try {
    effective.value = await permissionWorkflowApi.getEffectivePermissions(selectedUserId.value);
  } catch (err: any) {
    error.value = err?.data?.error || err?.message || "Unable to load effective permissions";
  } finally {
    loading.value = false;
  }
}

async function bootstrap() {
  loading.value = true;
  try {
    await loadUsers();
    await loadEffective();
  } catch (err: any) {
    error.value = err?.data?.error || err?.message || "Unable to load users";
  } finally {
    loading.value = false;
  }
}

async function openManageModal() {
  showManageModal.value = true;
  modalLoading.value = true;
  modalError.value = null;
  try {
    const [rolesRes, groupsRes, functionsRes, fundsRes] = await Promise.all([
      permissionWorkflowApi.listRoles(),
      permissionWorkflowApi.listGroups(),
      permissionWorkflowApi.listFunctionDefinitions(),
      myFundsApi.listMyFunds(),
    ]);
    rolesList.value = rolesRes;
    groupsList.value = groupsRes.items;
    functionsList.value = functionsRes;
    fundsList.value = fundsRes;
  } catch (err: any) {
    modalError.value = err?.data?.error || err?.message || "Failed to load options";
  } finally {
    modalLoading.value = false;
  }
}

async function submitRoleRequest() {
  if (!selectedRole.value || !roleReason.value.trim()) return;
  modalLoading.value = true;
  modalError.value = null;
  try {
    const role = rolesList.value.find((r) => r.role_code === selectedRole.value || r.id === selectedRole.value);
    if (!role) throw new Error("Selected role not found");

    const res = await permissionWorkflowApi.requestRoleAssignment(selectedUserId.value, {
      role_code: role.role_code,
      role_id: role.id,
      reason: roleReason.value.trim(),
    });
    showManageModal.value = false;
    await navigateTo(`/permissions/change-requests/${res.request.id}`);
  } catch (err: any) {
    modalError.value = err?.data?.error || err?.message || "Failed to request role assignment";
  } finally {
    modalLoading.value = false;
  }
}

async function submitGroupRequest() {
  if (!selectedGroup.value || !groupReason.value.trim()) return;
  modalLoading.value = true;
  modalError.value = null;
  try {
    const cr = await permissionWorkflowApi.createRequest({
      title: `Group membership change for ${selectedUser.value?.display_name || "user"}`,
      description: groupReason.value.trim(),
      request_type: "PERMISSION_CHANGE",
      risk_level: "MEDIUM",
      target_entity_type: "USER",
      target_entity_id: selectedUserId.value,
    });
    await permissionWorkflowApi.addItem(cr.id, {
      item_type: "GROUP_MEMBERSHIP",
      target_table: "permissions_accounts_groups",
      target_id: selectedUserId.value,
      action_type: groupAction.value === "ADD" ? "ADD_GROUP_MEMBER" : "REMOVE_GROUP_MEMBER",
      after_json: {
        user_id: selectedUserId.value,
        group_id: selectedGroup.value,
      },
    });
    showManageModal.value = false;
    await navigateTo(`/permissions/change-requests/${cr.id}`);
  } catch (err: any) {
    modalError.value = err?.data?.error || err?.message || "Failed to request group membership change";
  } finally {
    modalLoading.value = false;
  }
}

async function submitFunctionRequest() {
  if (!selectedFunction.value || !functionReason.value.trim()) return;
  modalLoading.value = true;
  modalError.value = null;
  try {
    const cr = await permissionWorkflowApi.createRequest({
      title: `Direct function rights override for ${selectedUser.value?.display_name || "user"}`,
      description: functionReason.value.trim(),
      request_type: "PERMISSION_CHANGE",
      risk_level: "MEDIUM",
      target_entity_type: "USER",
      target_entity_id: selectedUserId.value,
    });
    await permissionWorkflowApi.addItem(cr.id, {
      item_type: "FUNCTION_RIGHT",
      target_table: "permission_function_rights",
      target_id: selectedUserId.value,
      action_type: "UPSERT_FUNCTION_RIGHT",
      after_json: {
        subject_type: "USER",
        subject_id: selectedUserId.value,
        permission_code: selectedFunction.value,
        can_view: rights.canView,
        can_search: rights.canSearch,
        can_add: rights.canAdd,
        can_edit: rights.canEdit,
        can_delete: rights.canDelete,
        can_approve: rights.canApprove,
        can_revoke_approval: rights.canRevokeApproval,
        can_export: rights.canExport,
        can_configure: rights.canConfigure,
      },
    });
    showManageModal.value = false;
    await navigateTo(`/permissions/change-requests/${cr.id}`);
  } catch (err: any) {
    modalError.value = err?.data?.error || err?.message || "Failed to request function permission change";
  } finally {
    modalLoading.value = false;
  }
}

async function submitDataRequest() {
  if (!selectedFund.value || !dataReason.value.trim()) return;
  modalLoading.value = true;
  modalError.value = null;
  try {
    const cr = await permissionWorkflowApi.createRequest({
      title: `Direct data scope override for ${selectedUser.value?.display_name || "user"}`,
      description: dataReason.value.trim(),
      request_type: "DATA_RIGHTS_CHANGE",
      risk_level: "MEDIUM",
      target_entity_type: "USER",
      target_entity_id: selectedUserId.value,
    });
    await permissionWorkflowApi.addItem(cr.id, {
      item_type: "DATA_RIGHT",
      target_table: "permission_data_rights",
      target_id: selectedUserId.value,
      action_type: "UPSERT_DATA_RIGHT",
      after_json: {
        subject_type: "USER",
        subject_id: selectedUserId.value,
        fund_id: selectedFund.value,
        access_level: dataAccessLevel.value,
      },
    });
    showManageModal.value = false;
    await navigateTo(`/permissions/change-requests/${cr.id}`);
  } catch (err: any) {
    modalError.value = err?.data?.error || err?.message || "Failed to request data rights change";
  } finally {
    modalLoading.value = false;
  }
}

async function submitWorkflowApproveRequest() {
  if (!workflowApproveReason.value.trim()) return;
  modalLoading.value = true;
  modalError.value = null;
  try {
    const cr = await permissionWorkflowApi.createRequest({
      title: `Grant WORKFLOW_APPROVE to ${selectedUser.value?.display_name || "user"}`,
      description: workflowApproveReason.value.trim(),
      request_type: "PERMISSION_CHANGE",
      risk_level: "HIGH",
      target_entity_type: "USER",
      target_entity_id: selectedUserId.value,
    });
    await permissionWorkflowApi.addItem(cr.id, {
      item_type: "FUNCTION_RIGHT",
      target_table: "permission_function_rights",
      target_id: selectedUserId.value,
      action_type: "UPSERT_FUNCTION_RIGHT",
      after_json: {
        subject_type: "USER",
        subject_id: selectedUserId.value,
        permission_code: "WORKFLOW_APPROVE",
        can_approve: true,
      },
    });
    showManageModal.value = false;
    await navigateTo(`/permissions/change-requests/${cr.id}`);
  } catch (err: any) {
    modalError.value = err?.data?.error || err?.message || "Failed to create WORKFLOW_APPROVE request";
  } finally {
    modalLoading.value = false;
  }
}

onMounted(bootstrap);
</script>

<template>
  <main class="effective-permissions">
    <header class="effective-permissions__header">
      <div>
        <p class="effective-permissions__eyebrow">Permission Management</p>
        <h1>Effective Permission Viewer</h1>
        <p>Inspect direct, group-derived, and role-derived permissions for a user.</p>
      </div>
      <form class="effective-permissions__picker" @submit.prevent="loadEffective">
        <select v-model="selectedUserId" class="select">
          <option v-for="user in users" :key="user.id" :value="user.id">
            {{ user.display_name || user.username }}
          </option>
        </select>
        <div class="picker-buttons">
          <button class="btn btn-primary btn-sm" :disabled="loading || !selectedUserId">Load</button>
          <button
            v-if="selectedUserId"
            type="button"
            class="btn btn-secondary btn-sm"
            @click="openManageModal"
          >
            Manage Permissions
          </button>
        </div>
      </form>
    </header>

    <p v-if="error" class="alert alert-danger">{{ error }}</p>

    <section v-if="selectedUser" class="effective-permissions__identity">
      <strong>{{ selectedUser.display_name }}</strong>
      <span>{{ selectedUser.username }} | {{ selectedUser.email }}</span>
      <span class="badge" :class="selectedUser.is_active && !selectedUser.is_locked ? 'badge-success' : 'badge-error'">
        {{ selectedUser.is_active && !selectedUser.is_locked ? "Active" : "Restricted" }}
      </span>
    </section>

    <section v-if="effective" class="effective-permissions__grid">
      <article>
        <h2>Final Function Permissions</h2>
        <div class="effective-permissions__chips">
          <span v-for="code in effective.final_function_permissions" :key="code">{{ code }}</span>
        </div>
      </article>

      <article>
        <h2>Final Data Permissions</h2>
        <div class="effective-permissions__chips">
          <span v-for="scope in effective.final_contract_permissions" :key="scope">{{ scope }}</span>
        </div>
      </article>

      <article>
        <h2>Role-Derived Permissions</h2>
        <p>{{ effective.role_function_permissions.length }} function grants | {{ effective.role_data_permissions.length }} data grants</p>
      </article>

      <article>
        <h2>Group-Derived Permissions</h2>
        <p>{{ effective.group_function_permissions.length }} function grants | {{ effective.group_data_permissions.length }} data grants</p>
      </article>

      <article>
        <h2>Direct User Permissions</h2>
        <p>{{ effective.direct_function_permissions.length }} function grants | {{ effective.direct_data_permissions.length }} data grants</p>
      </article>

      <article>
        <h2>Approved Roles</h2>
        <div class="effective-permissions__chips">
          <span v-for="role in effective.roles" :key="role.id">{{ role.role_code }}</span>
        </div>
      </article>
    </section>

    <!-- Modal Backdrop -->
    <div v-if="showManageModal" class="modal-backdrop" @click.self="showManageModal = false">
      <!-- Modal Container -->
      <div class="modal-container">
        <header class="modal-header">
          <h3>Manage Permissions — {{ selectedUser?.display_name }}</h3>
          <button class="modal-close" @click="showManageModal = false">&times;</button>
        </header>

        <div class="modal-body">
          <nav class="modal-tabs">
            <button
              class="modal-tab-btn"
              :class="{ 'is-active': activeModalTab === 'roles' }"
              @click="activeModalTab = 'roles'"
            >
              Assign Role
            </button>
            <button
              class="modal-tab-btn"
              :class="{ 'is-active': activeModalTab === 'groups' }"
              @click="activeModalTab = 'groups'"
            >
              Group Membership
            </button>
            <button
              class="modal-tab-btn"
              :class="{ 'is-active': activeModalTab === 'functions' }"
              @click="activeModalTab = 'functions'"
            >
              Direct Functions
            </button>
            <button
              class="modal-tab-btn"
              :class="{ 'is-active': activeModalTab === 'data' }"
              @click="activeModalTab = 'data'"
            >
              Direct Data
            </button>
            <button
              class="modal-tab-btn"
              :class="{ 'is-active': activeModalTab === 'workflow-approve' }"
              @click="activeModalTab = 'workflow-approve'"
            >
              Workflow Approve
            </button>
          </nav>

          <p v-if="modalError" class="alert alert-danger">{{ modalError }}</p>
          <p v-if="modalLoading" class="is-muted">Loading options...</p>

          <template v-if="!modalLoading">
            <!-- 1. Assign Role Tab -->
            <form v-if="activeModalTab === 'roles'" class="modal-form" @submit.prevent="submitRoleRequest">
              <div class="modal-form-group">
                <label for="role-select">Select Role</label>
                <select id="role-select" v-model="selectedRole" class="select" required>
                  <option value="" disabled>-- Select a Role --</option>
                  <option v-for="role in rolesList" :key="role.id" :value="role.role_code">
                    {{ role.role_name }} ({{ role.role_code }})
                  </option>
                </select>
              </div>
              <div class="modal-form-group">
                <label for="role-reason">Justification / Reason</label>
                <textarea id="role-reason" v-model="roleReason" class="textarea" rows="3" placeholder="Enter reason for role assignment..." required></textarea>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-secondary btn-sm" @click="showManageModal = false">Cancel</button>
                <button type="submit" class="btn btn-primary btn-sm" :disabled="!selectedRole || !roleReason.trim()">Create Request</button>
              </div>
            </form>

            <!-- 2. Group Membership Tab -->
            <form v-else-if="activeModalTab === 'groups'" class="modal-form" @submit.prevent="submitGroupRequest">
              <div class="modal-form-group">
                <label for="group-select">Select Group</label>
                <select id="group-select" v-model="selectedGroup" class="select" required>
                  <option value="" disabled>-- Select a Group --</option>
                  <option v-for="grp in groupsList" :key="grp.id" :value="grp.id">
                    {{ grp.name }}
                  </option>
                </select>
              </div>
              <div class="modal-form-group">
                <label>Action</label>
                <div style="display: flex; gap: var(--space-4);">
                  <label class="checkbox-label">
                    <input type="radio" v-model="groupAction" value="ADD" /> Add to Group
                  </label>
                  <label class="checkbox-label">
                    <input type="radio" v-model="groupAction" value="REMOVE" /> Remove from Group
                  </label>
                </div>
              </div>
              <div class="modal-form-group">
                <label for="group-reason">Justification / Reason</label>
                <textarea id="group-reason" v-model="groupReason" class="textarea" rows="3" placeholder="Enter reason..." required></textarea>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-secondary btn-sm" @click="showManageModal = false">Cancel</button>
                <button type="submit" class="btn btn-primary btn-sm" :disabled="!selectedGroup || !groupReason.trim()">Create Request</button>
              </div>
            </form>

            <!-- 3. Direct Functions Tab -->
            <form v-else-if="activeModalTab === 'functions'" class="modal-form" @submit.prevent="submitFunctionRequest">
              <div class="modal-form-group">
                <label for="func-select">Select Function Definition</label>
                <select id="func-select" v-model="selectedFunction" class="select" required>
                  <option value="" disabled>-- Select a Function --</option>
                  <option v-for="f in functionsList" :key="f.code" :value="f.code">
                    {{ f.name }} ({{ f.code }})
                  </option>
                </select>
              </div>
              <div class="modal-form-group">
                <label>Granted Operations</label>
                <div class="modal-grid-checkboxes">
                  <label class="checkbox-label"><input type="checkbox" v-model="rights.canView" /> can_view</label>
                  <label class="checkbox-label"><input type="checkbox" v-model="rights.canSearch" /> can_search</label>
                  <label class="checkbox-label"><input type="checkbox" v-model="rights.canAdd" /> can_add</label>
                  <label class="checkbox-label"><input type="checkbox" v-model="rights.canEdit" /> can_edit</label>
                  <label class="checkbox-label"><input type="checkbox" v-model="rights.canDelete" /> can_delete</label>
                  <label class="checkbox-label"><input type="checkbox" v-model="rights.canApprove" /> can_approve</label>
                  <label class="checkbox-label"><input type="checkbox" v-model="rights.canRevokeApproval" /> can_revoke_approval</label>
                  <label class="checkbox-label"><input type="checkbox" v-model="rights.canExport" /> can_export</label>
                  <label class="checkbox-label"><input type="checkbox" v-model="rights.canConfigure" /> can_configure</label>
                </div>
              </div>
              <div class="modal-form-group">
                <label for="func-reason">Justification / Reason</label>
                <textarea id="func-reason" v-model="functionReason" class="textarea" rows="3" placeholder="Enter reason..." required></textarea>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-secondary btn-sm" @click="showManageModal = false">Cancel</button>
                <button type="submit" class="btn btn-primary btn-sm" :disabled="!selectedFunction || !functionReason.trim()">Create Request</button>
              </div>
            </form>

            <!-- 5. Workflow Approve Tab -->
            <form v-else-if="activeModalTab === 'workflow-approve'" class="modal-form" @submit.prevent="submitWorkflowApproveRequest">
              <div class="modal-notice">
                <p>
                  This creates a permission change request granting
                  <strong>{{ selectedUser?.display_name || "this user" }}</strong>
                  the <code>WORKFLOW_APPROVE</code> permission code, which is required to perform
                  Manager Approval on the workflow day.
                </p>
                <p class="modal-notice__warning">
                  The permission becomes effective only after the request is submitted, approved
                  by a reviewer, and merged. The user must also log out and log back in
                  for the new permission to take effect.
                </p>
                <p>
                  After the request is created you will be taken to the request detail page
                  to submit, review, and merge it.
                </p>
              </div>
              <div class="modal-form-group">
                <label for="workflow-approve-reason">Justification / Reason</label>
                <textarea
                  id="workflow-approve-reason"
                  v-model="workflowApproveReason"
                  class="textarea"
                  rows="3"
                  placeholder="Why should this user be able to approve the workflow day? (e.g. Ben is the designated Manager and needs to approve days that someone else opened.)"
                  required
                ></textarea>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-secondary btn-sm" @click="showManageModal = false">Cancel</button>
                <button
                  type="submit"
                  class="btn btn-primary btn-sm"
                  :disabled="!workflowApproveReason.trim()"
                >
                  Create Permission Request
                </button>
              </div>
            </form>

            <!-- 4. Direct Data Tab -->
            <form v-else-if="activeModalTab === 'data'" class="modal-form" @submit.prevent="submitDataRequest">
              <div class="modal-form-group">
                <label for="fund-select">Select Fund</label>
                <select id="fund-select" v-model="selectedFund" class="select" required>
                  <option value="" disabled>-- Select a Fund --</option>
                  <option v-for="fund in fundsList" :key="fund.id" :value="fund.id">
                    {{ fund.name }} ({{ fund.code }})
                  </option>
                </select>
              </div>
              <div class="modal-form-group">
                <label for="access-level">Access Level</label>
                <select id="access-level" v-model="dataAccessLevel" class="select" required>
                  <option value="READ">READ</option>
                  <option value="WRITE">WRITE</option>
                  <option value="NONE">NONE</option>
                </select>
              </div>
              <div class="modal-form-group">
                <label for="data-reason">Justification / Reason</label>
                <textarea id="data-reason" v-model="dataReason" class="textarea" rows="3" placeholder="Enter reason..." required></textarea>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-secondary btn-sm" @click="showManageModal = false">Cancel</button>
                <button type="submit" class="btn btn-primary btn-sm" :disabled="!selectedFund || !dataReason.trim()">Create Request</button>
              </div>
            </form>
          </template>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.effective-permissions {
  display: grid;
  gap: var(--space-6);
}

.effective-permissions__header,
.effective-permissions__identity {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-6);
  align-items: end;
}

.effective-permissions__picker {
  display: grid;
  grid-template-columns: minmax(260px, 380px) auto;
  gap: var(--space-3);
}

.picker-buttons {
  display: flex;
  gap: var(--space-2);
}

.effective-permissions__eyebrow {
  margin: 0 0 var(--space-2);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
}

.effective-permissions__identity {
  grid-template-columns: auto minmax(0, 1fr) auto;
  justify-content: start;
  padding: var(--space-4) 0;
  border-top: 1px solid var(--border-subtle);
  border-bottom: 1px solid var(--border-subtle);
}

.effective-permissions__identity span {
  color: var(--text-secondary);
}

.effective-permissions__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-5);
}

.effective-permissions__grid article {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-5);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-lg);
}

.effective-permissions__grid h2 {
  margin: 0;
  font-size: var(--font-size-md);
}

.effective-permissions__grid p {
  margin: 0;
}

.effective-permissions__chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.effective-permissions__chips span {
  min-height: 24px;
  display: inline-flex;
  align-items: center;
  padding: 0 var(--space-3);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-pill);
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

/* Modal Styling */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(8px);
  display: grid;
  place-items: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease-out;
}

.modal-container {
  background: var(--bg-card, #ffffff);
  border: 1px solid var(--border-subtle, #e1e4e8);
  border-radius: var(--radius-xl, 16px);
  width: 90%;
  max-width: 650px;
  max-height: 85vh;
  overflow-y: auto;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  display: flex;
  flex-direction: column;
  animation: scaleIn 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-header {
  padding: var(--space-5) var(--space-6);
  border-bottom: 1px solid var(--border-subtle, #e1e4e8);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h3 {
  margin: 0;
  font-size: var(--font-size-lg, 18px);
  font-weight: var(--font-weight-bold, 700);
}

.modal-close {
  background: none;
  border: none;
  font-size: var(--font-size-lg, 20px);
  cursor: pointer;
  color: var(--text-secondary, #586069);
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-md, 6px);
  transition: background 0.15s;
}

.modal-close:hover {
  background: var(--bg-selected, #f3f4f6);
}

.modal-body {
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.modal-tabs {
  display: flex;
  border-bottom: 1px solid var(--border-subtle, #e1e4e8);
  margin-bottom: var(--space-4);
  gap: var(--space-2);
}

.modal-tab-btn {
  padding: var(--space-2) var(--space-4);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  color: var(--text-secondary, #586069);
  font-weight: var(--font-weight-medium, 500);
  font-size: var(--font-size-sm, 14px);
  transition: all 0.15s;
}

.modal-tab-btn:hover {
  color: var(--text-primary, #24292e);
  border-bottom-color: var(--border-muted, #d1d5da);
}

.modal-tab-btn.is-active {
  color: var(--text-link, #0969da);
  border-bottom-color: var(--text-link, #0969da);
  font-weight: var(--font-weight-semibold, 600);
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.modal-form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.modal-form-group label {
  font-weight: var(--font-weight-semibold, 600);
  font-size: var(--font-size-sm, 14px);
}

.modal-grid-checkboxes {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-3);
  padding: var(--space-3);
  background: var(--bg-card-muted, #f6f8fa);
  border-radius: var(--radius-lg, 8px);
  border: 1px solid var(--border-subtle, #e1e4e8);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-xs, 12px);
  cursor: pointer;
}

.modal-footer {
  padding: var(--space-4) var(--space-6);
  border-top: 1px solid var(--border-subtle, #e1e4e8);
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  background: var(--bg-card-muted, #f6f8fa);
  border-bottom-left-radius: var(--radius-xl, 16px);
  border-bottom-right-radius: var(--radius-xl, 16px);
}

.modal-notice {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-4);
  background: var(--bg-card-muted, #f6f8fa);
  border: 1px solid var(--border-subtle, #e1e4e8);
  border-radius: var(--radius-lg, 8px);
  font-size: var(--font-size-sm, 14px);
}

.modal-notice p {
  margin: 0;
  line-height: 1.5;
}

.modal-notice code {
  font-family: var(--font-mono, ui-monospace, monospace);
  background: var(--bg-selected, #eef1f4);
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 12px;
}

.modal-notice__warning {
  color: var(--color-warning-800, #92400e);
  background: var(--color-warning-50, #fffbeb);
  border: 1px solid var(--color-warning-200, #fde68a);
  border-radius: var(--radius-md, 6px);
  padding: var(--space-3);
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes scaleIn {
  from { transform: scale(0.95); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

@media (max-width: 900px) {
  .effective-permissions__header,
  .effective-permissions__identity,
  .effective-permissions__picker,
  .effective-permissions__grid {
    grid-template-columns: 1fr;
  }
}
</style>
