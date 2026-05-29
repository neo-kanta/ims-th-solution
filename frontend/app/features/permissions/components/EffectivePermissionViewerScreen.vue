<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import { permissionWorkflowApi } from "../services/permissionWorkflowApi";
import type { EffectivePermissions, PermissionUserSummary } from "../types";

const props = defineProps<{
  initialUserId?: string;
}>();

const users = ref<PermissionUserSummary[]>([]);
const selectedUserId = ref(props.initialUserId || "");
const effective = ref<EffectivePermissions | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

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
        <button class="btn btn-primary btn-sm" :disabled="loading || !selectedUserId">Load</button>
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

@media (max-width: 900px) {
  .effective-permissions__header,
  .effective-permissions__identity,
  .effective-permissions__picker,
  .effective-permissions__grid {
    grid-template-columns: 1fr;
  }
}
</style>
