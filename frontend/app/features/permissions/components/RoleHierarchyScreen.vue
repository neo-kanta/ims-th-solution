<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import { permissionWorkflowApi } from "../services/permissionWorkflowApi";
import type { PermissionRole } from "../types";
import AppLoadingState from "~/shared/ui/AppLoadingState.vue";

const roles = ref<PermissionRole[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const search = ref("");

const filteredRoles = computed(() => {
  const term = search.value.trim().toLowerCase();
  if (!term) return roles.value;
  return roles.value.filter((role) => (
    role.role_code.toLowerCase().includes(term)
    || role.role_name.toLowerCase().includes(term)
    || role.assignment_scope.toLowerCase().includes(term)
  ));
});

async function loadRoles() {
  loading.value = true;
  error.value = null;
  try {
    roles.value = await permissionWorkflowApi.listRoles();
  } catch (err: any) {
    error.value = err?.data?.error || err?.message || "Unable to load role hierarchy";
  } finally {
    loading.value = false;
  }
}

function riskClass(role: PermissionRole) {
  return role.is_high_risk ? "badge-error" : "badge-neutral";
}

onMounted(loadRoles);
</script>

<template>
  <main class="role-hierarchy">
    <header class="role-hierarchy__header">
      <div>
        <p class="role-hierarchy__eyebrow">Permission Management</p>
        <h1>Role Hierarchy</h1>
        <p>Priority rank and assignment scope determine who may request role changes.</p>
      </div>
      <input v-model="search" class="input" placeholder="Search roles or scopes" />
    </header>

    <p v-if="error" class="alert alert-danger">{{ error }}</p>

    <section class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>Role code</th>
            <th>Role name</th>
            <th>Department</th>
            <th>Priority</th>
            <th>Scope</th>
            <th>Can request</th>
            <th>Can approve</th>
            <th>Risk</th>
            <th>Active</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="9" class="app-table__state-row">
              <AppLoadingState message="Loading role hierarchy..." />
            </td>
          </tr>
          <tr v-for="role in filteredRoles" :key="role.id">
            <td><strong>{{ role.role_code }}</strong></td>
            <td>{{ role.role_name }}</td>
            <td>{{ role.department }}</td>
            <td>{{ role.priority_rank }}</td>
            <td>{{ role.assignment_scope }}</td>
            <td><span class="badge" :class="role.can_request_role_assignment ? 'badge-success' : 'badge-neutral'">{{ role.can_request_role_assignment ? "Yes" : "No" }}</span></td>
            <td><span class="badge" :class="role.can_approve_role_assignment ? 'badge-success' : 'badge-neutral'">{{ role.can_approve_role_assignment ? "Yes" : "No" }}</span></td>
            <td><span class="badge" :class="riskClass(role)">{{ role.is_high_risk ? "High" : "Standard" }}</span></td>
            <td><span class="badge" :class="role.is_active ? 'badge-success' : 'badge-neutral'">{{ role.is_active ? "Active" : "Inactive" }}</span></td>
          </tr>
        </tbody>
      </table>
    </section>
  </main>
</template>

<style scoped>
.role-hierarchy {
  display: grid;
  gap: var(--space-6);
}

.role-hierarchy__header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(260px, 380px);
  gap: var(--space-6);
  align-items: end;
}

.role-hierarchy__eyebrow {
  margin: 0 0 var(--space-2);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
}

@media (max-width: 900px) {
  .role-hierarchy__header {
    grid-template-columns: 1fr;
  }
}

.app-table__state-row {
  text-align: center;
  background: transparent !important;
}
</style>
