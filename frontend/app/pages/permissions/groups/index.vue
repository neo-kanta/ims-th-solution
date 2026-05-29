<script setup lang="ts">
import { onMounted, ref } from "vue";

import { permissionWorkflowApi } from "~/features/permissions/services/permissionWorkflowApi";
import type { PermissionGroupSummary } from "~/features/permissions/types";

definePageMeta({
  layout: "dashboard",
  middleware: ["auth", "permission"],
  permission: "permission.groups.view",
});

const groups = ref<PermissionGroupSummary[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

async function loadGroups() {
  loading.value = true;
  error.value = null;
  try {
    const response = await permissionWorkflowApi.listGroups({ page: 1, limit: 100 });
    groups.value = response.items;
  } catch (err: any) {
    error.value = err?.data?.error || err?.message || "Unable to load groups";
  } finally {
    loading.value = false;
  }
}

onMounted(loadGroups);
</script>

<template>
  <main class="permission-directory">
    <AppPageHeader
      title="Groups"
      description="Permission groups used for account mapping and derived function grants."
    >
      <template #eyebrow>
        Permission Management
      </template>
    </AppPageHeader>
    <p v-if="error" class="alert alert-danger">{{ error }}</p>
    <section class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>Group</th>
            <th>Description</th>
            <th>Members</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="4" class="is-muted">Loading groups...</td>
          </tr>
          <tr v-for="group in groups" :key="group.id">
            <td><strong>{{ group.name }}</strong></td>
            <td>{{ group.description || "-" }}</td>
            <td>{{ group.members_count }}</td>
            <td><span class="badge" :class="group.is_active ? 'badge-success' : 'badge-neutral'">{{ group.is_active ? "Active" : "Inactive" }}</span></td>
          </tr>
        </tbody>
      </table>
    </section>
  </main>
</template>

<style scoped>
.permission-directory {
  display: grid;
  gap: var(--space-6);
}

.permission-directory__eyebrow {
  margin: 0 0 var(--space-2);
  color: var(--text-tertiary);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  text-transform: uppercase;
}
</style>
